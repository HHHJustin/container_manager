package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"container-manager/internal/containers"
)

var Svc = containers.NewService()

type createContainerDTO struct {
	Name  string `json:"name" binding:"omitempty"`
	Image string `json:"image" binding:"required"`
}

func CreateContainer(c *gin.Context) {
	var dto createContainerDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := Svc.Create(containers.CreateOptions{Name: dto.Name, Image: dto.Image})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := Svc.Start(id); err != nil {
		status := http.StatusInternalServerError
		if err == containers.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func StopContainer(c *gin.Context) {
	id := c.Param("id")
	if err := Svc.Stop(id); err != nil {
		status := http.StatusInternalServerError
		if err == containers.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func DeleteContainer(c *gin.Context) {
	id := c.Param("id")
	if err := Svc.Delete(id); err != nil {
		status := http.StatusInternalServerError
		if err == containers.ErrNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Job API ----
type runJobDTO struct {
	Image        string   `json:"image"`                           // 可省略，將自動偵測
	HostDir      string   `json:"hostDir" binding:"required"`      // 來自 /v1/uploads 回傳的 dir
	ContainerDir string   `json:"containerDir" binding:"required"` // 例如 /workspace
	Cmd          []string `json:"cmd"`                             // 可省略，將自動偵測
}

func RunJob(c *gin.Context) {
	var dto runJobDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Ensure absolute host path for Docker bind mount
	hostDir := dto.HostDir
	if !filepath.IsAbs(hostDir) {
		abs, err := filepath.Abs(hostDir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hostDir"})
			return
		}
		hostDir = abs
	}
	// 若服務在容器中運行，hostDir 目前是容器內路徑，需轉換為宿主機路徑讓 Docker 進行 bind mount。
	// 使用 DATA_DIR 與 HOST_DATA_DIR 的對應做轉換。
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	absDataDir, _ := filepath.Abs(dataDir)
	hostDataDir := os.Getenv("HOST_DATA_DIR")
	if hostDataDir == "" {
		// 明確要求透過環境變數提供宿主機絕對路徑，避免在容器內推測失敗
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HOST_DATA_DIR not set"})
		return
	}
	if strings.HasPrefix(hostDir, absDataDir) {
		rel := strings.TrimPrefix(hostDir, absDataDir)
		if len(rel) > 0 && (rel[0] == '/' || rel[0] == '\\') {
			rel = rel[1:]
		}
		hostDir = filepath.Join(hostDataDir, rel)
		// 再次確保結果是絕對路徑
		if !filepath.IsAbs(hostDir) {
			abs, _ := filepath.Abs(hostDir)
			if abs != "" {
				hostDir = abs
			}
		}
	}
	// 注意：此處無法在容器內 stat 宿主機路徑，交由 Docker 在 bind mount 時檢查

	image := strings.TrimSpace(dto.Image)
	cmd := dto.Cmd
	if len(cmd) == 0 || image == "" {
		autoImg, autoCmd := detectProgram(hostDir, dto.ContainerDir)
		if image == "" {
			image = autoImg
		}
		if len(cmd) == 0 {
			cmd = autoCmd
		}
	}

	code, logs, err := Svc.RunJob(containers.JobOptions{Image: image, HostDir: hostDir, ContainerDir: dto.ContainerDir, Cmd: cmd})
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exitCode": code, "logs": logs})
}

// detectProgram 依序檢查可執行檔與常見腳本，回傳對應的映像與命令。
// 優先序：binary(app 可執行) > run.sh > app.py > app.go
func detectProgram(hostDir, containerDir string) (string, []string) {
	// binary
	appPath := filepath.Join(hostDir, "app")
	if fi, err := os.Stat(appPath); err == nil && fi.Mode()&0o111 != 0 {
		return "alpine:3.20", []string{"sh", "-c", "chmod +x " + filepath.Join(containerDir, "app") + " && " + filepath.Join(containerDir, "app")}
	}
	// run.sh
	shPath := filepath.Join(hostDir, "run.sh")
	if _, err := os.Stat(shPath); err == nil {
		return "alpine:3.20", []string{"sh", "-c", "chmod +x " + filepath.Join(containerDir, "run.sh") + " && " + filepath.Join(containerDir, "run.sh")}
	}
	// app.py
	pyPath := filepath.Join(hostDir, "app.py")
	if _, err := os.Stat(pyPath); err == nil {
		return "python:3.11-slim", []string{"python", filepath.Join(containerDir, "app.py")}
	}
	// app.go
	goPath := filepath.Join(hostDir, "app.go")
	if _, err := os.Stat(goPath); err == nil {
		return "golang:1.21-alpine", []string{"sh", "-c", "cd " + containerDir + " && go run app.go"}
	}
	// fallback: 列目錄
	return "alpine:3.20", []string{"sh", "-c", "ls -la " + containerDir}
}
