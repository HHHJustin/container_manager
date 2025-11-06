package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadRequest 包含可選擇的 userId，若未提供將自動產生。
type UploadRequest struct {
	UserID string `form:"userId"` // 可從表單欄位帶入
}

func Upload(c *gin.Context) {
	var req UploadRequest
	_ = c.ShouldBind(&req)
	if req.UserID == "" {
		req.UserID = uuid.NewString()
	}

	root := os.Getenv("DATA_DIR")
	if root == "" {
		root = "./data"
	}

	batch := time.Now().UTC().Format("20060102T150405Z")
	destDir := filepath.Join(root, req.UserID, batch)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded (field: files)"})
		return
	}

	stored := make([]string, 0, len(files))
	for _, f := range files {
		name := filepath.Base(f.Filename)
		path := filepath.Join(destDir, name)
		if err := c.SaveUploadedFile(f, path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save %s failed: %v", name, err)})
			return
		}
		stored = append(stored, path)
	}

	c.JSON(http.StatusOK, gin.H{
		"userId": req.UserID,
		"dir":    destDir,
		"files":  stored,
	})
}
