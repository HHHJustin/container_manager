package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"

    "container-manager/internal/api/handlers"
    "container-manager/internal/containers"
)

type execProviderMock struct{ containers.Provider }
func (execProviderMock) Exec(id string, cmd []string) (int, string, error) { return 0, "ok", nil }

func TestExec_Handler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    handlers.Svc = containers.NewServiceWith(execProviderMock{}, nil)

    r := gin.New()
    r.POST("/v1/containers/:id/exec", handlers.ExecInContainer)

    body, _ := json.Marshal(map[string]any{"cmd": []string{"echo","hi"}})
    req := httptest.NewRequest(http.MethodPost, "/v1/containers/cid/exec", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
}


