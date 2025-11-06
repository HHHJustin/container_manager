package tests

import (
    "bytes"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"

    "github.com/gin-gonic/gin"

    "container-manager/internal/api/handlers"
    "container-manager/internal/middleware"
)

// Minimal integration: build routes like server.Run and verify /uploads flow with Auth.
func TestAPI_Upload_WithAuth(t *testing.T) {
    gin.SetMode(gin.TestMode)
    os.Setenv("JWT_SECRET", "devsecret")

    r := gin.New()
    r.Use(middleware.Logger())
    r.Use(middleware.Recover())
    r.Use(middleware.Cors())
    r.GET("/healthz", handlers.Health)
    r.POST("/login", handlers.Login)
    v1 := r.Group("/v1")
    v1.Use(middleware.Auth())
    v1.POST("/uploads", handlers.Upload)

    // create a temp file to upload
    dir := t.TempDir()
    p := filepath.Join(dir, "a.txt")
    if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil { t.Fatalf("write temp: %v", err) }

    var body bytes.Buffer
    mw := multipart.NewWriter(&body)
    fw, _ := mw.CreateFormFile("files", filepath.Base(p))
    f, _ := os.Open(p)
    _, _ = io.Copy(fw, f)
    f.Close()
    _ = mw.WriteField("userId", "u123")
    _ = mw.Close()

    // 先取得 JWT
    loginBody, _ := json.Marshal(map[string]string{"username":"admin","password":"admin"})
    reqLogin := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
    reqLogin.Header.Set("Content-Type", "application/json")
    wLogin := httptest.NewRecorder()
    r.ServeHTTP(wLogin, reqLogin)
    if wLogin.Code != http.StatusOK { t.Fatalf("login status %d", wLogin.Code) }
    var out struct{ Token string `json:"token"` }
    _ = json.Unmarshal(wLogin.Body.Bytes(), &out)

    req := httptest.NewRequest(http.MethodPost, "/v1/uploads", &body)
    req.Header.Set("Authorization", "Bearer "+out.Token)
    req.Header.Set("Content-Type", mw.FormDataContentType())
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK { t.Fatalf("upload status %d body=%s", w.Code, w.Body.String()) }
}


