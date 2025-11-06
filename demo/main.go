package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	base := getenv("BASE_URL", "http://localhost:8081")

	// 1) 取得 Token：統一使用 JWT（/login）
	token, err := login(base, getenv("AUTH_USER", "admin"), getenv("AUTH_PASS", "admin"))
	check(err)
	fmt.Println("✓ 取得 JWT")

	// 2) 上傳檔案（使用 demo/files 下的範例）
	//    回傳值中的 dir 將被用來掛載到容器
	dir := upload(base+"/v1/uploads", token, map[string]string{"userId": "u123"},
		[]string{"files/a.csv", "files/b.json", "files/app.py"})
	fmt.Println("✓ 上傳完成:", dir)

	// 3) 建立並啟動容器（Python），並掛載上傳的目錄
	//    注意：dir 是容器內路徑（如 /app/data/...），需要轉換為宿主機路徑
	//    這裡假設宿主機路徑與容器內路徑相同（在 Docker Compose 中已映射）
	//    實際使用時，需要根據 HOST_DATA_DIR 轉換
	cid := createContainer(base, token, map[string]any{
		"image": "python:3.11-slim",
		"name":  "demo",
		"mounts": map[string]string{
			dir: "/workspace", // 掛載上傳目錄到容器的 /workspace
		},
	})
	fmt.Printf("✓ 容器建立完成，ID: %s\n", cid)
	doEmpty(base+"/v1/containers/"+cid+"/start", token, "POST")
	fmt.Println("✓ 容器啟動完成")

	// 4) 在既有容器內執行 app.py（使用掛載的目錄）
	er := runExec(base+"/v1/containers/"+cid+"/exec", token, map[string]any{
		"cmd": []string{"python", "/workspace/app.py"},
	})
	fmt.Printf("✓ Exec 完成，exitCode=%v\n", er.ExitCode)

	// 4) 提示可以透過靜態路徑檢視結果
	//    http://<host>/static/<userId>/<batch>/list.txt
	if u, err := url.Parse(base); err == nil {
		fmt.Printf("結果檔案（可能的網址）：%s/static/%s/list.txt\n", u.Host, trimStaticRoot(dir))
	}
}

func createContainer(base, token string, payload map[string]any) string {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", base+"/v1/containers", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		bs, _ := io.ReadAll(res.Body)
		panic(fmt.Sprintf("create status %d: %s", res.StatusCode, bs))
	}
	var out struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out.ID
}

func doEmpty(urlStr, token, method string) {
	req, _ := http.NewRequest(method, urlStr, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	check(err)
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		panic(fmt.Sprintf("status %d: %s", res.StatusCode, b))
	}
}

// --- HTTP helpers ---

func login(base, user, pass string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": user, "password": pass})
	res, err := http.Post(base+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("login status %d: %s", res.StatusCode, b)
	}
	var out struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out.Token, nil
}

type uploadResp struct {
	UserID string `json:"userId"`
	Dir    string `json:"dir"`
}

func upload(urlStr, token string, fields map[string]string, files []string) string {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	for _, p := range files {
		f, err := os.Open(p)
		check(err)
		w, _ := mw.CreateFormFile("files", filepath.Base(p))
		_, _ = io.Copy(w, f)
		f.Close()
	}
	_ = mw.Close()
	req, _ := http.NewRequest("POST", urlStr, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		panic(fmt.Sprintf("upload status %d: %s", res.StatusCode, b))
	}
	var out uploadResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out.Dir
}

type jobResp struct {
	ExitCode int64  `json:"exitCode"`
	Logs     string `json:"logs"`
}

func runJob(urlStr, token string, payload map[string]any) jobResp {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", urlStr, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		panic(fmt.Sprintf("job status %d: %s", res.StatusCode, b))
	}
	var out jobResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out
}

type execResp struct {
	ExitCode int    `json:"exitCode"`
	Logs     string `json:"logs"`
	TaskID   string `json:"taskId"`
}

func runExec(urlStr, token string, payload map[string]any) execResp {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", urlStr, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	check(err)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		panic(fmt.Sprintf("exec status %d: %s", res.StatusCode, b))
	}
	var out execResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out
}

// 將 ./data/<uid>/<batch> 轉換成 <uid>/<batch> 方便拼接靜態網址
func trimStaticRoot(dir string) string {
	// 嘗試移除開頭的 ./ 或 / 符號
	d := dir
	for len(d) > 0 && (d[0] == '.' || d[0] == '/') {
		d = d[1:]
	}
	// 去掉開頭的 data/
	const prefix = "data/"
	if len(d) >= len(prefix) && d[:len(prefix)] == prefix {
		return d[len(prefix):]
	}
	return d
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}
