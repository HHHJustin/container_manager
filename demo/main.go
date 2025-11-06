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
	//    這裡一次上傳多種範例檔，API 端會自動偵測可執行的類型
	dir := upload(base+"/v1/uploads", token, map[string]string{"userId": "u123"},
		[]string{"files/a.csv", "files/b.json", "files/app.py"})
	fmt.Println("✓ 上傳完成:", dir)

	// 3) 不指定 image/cmd，交由 API 端自動偵測可執行內容
	jobResp := runJob(base+"/v1/jobs", token, map[string]any{
		"hostDir":      dir,
		"containerDir": "/workspace",
	})
	fmt.Printf("✓ Job 完成，exitCode=%v\n", jobResp.ExitCode)

	// 4) 提示可以透過靜態路徑檢視結果
	//    http://<host>/static/<userId>/<batch>/list.txt
	if u, err := url.Parse(base); err == nil {
		fmt.Printf("結果檔案（可能的網址）：%s/static/%s/list.txt\n", u.Host, trimStaticRoot(dir))
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
