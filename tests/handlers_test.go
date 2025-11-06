package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "regexp"
    "testing"

    sqlmock "github.com/DATA-DOG/go-sqlmock"
    "github.com/gin-gonic/gin"

    "container-manager/internal/api/handlers"
    "container-manager/internal/containers"
    "container-manager/internal/storage"
)

func setupTestService(t *testing.T) {
    t.Helper()
    db, mock, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    repo := storage.NewContainerRepository(db)
    handlers.Svc = containers.NewServiceWith(containers.NewMockProvider(), repo)
    mock.ExpectExec(regexp.QuoteMeta("INSERT INTO containers(id,name,image,status,created_at) VALUES($1,$2,$3,$4,$5)")).WithArgs(sqlmock.AnyArg(), "demo", "alpine:3.20", "created", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
}

func TestCreateContainer_Handler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    setupTestService(t)
    r := gin.New()
    r.POST("/v1/containers", handlers.CreateContainer)

    body, _ := json.Marshal(map[string]string{"name": "demo", "image": "alpine:3.20"})
    req := httptest.NewRequest(http.MethodPost, "/v1/containers", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
    }
}


