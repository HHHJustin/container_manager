# 容器管理系統（AIDMS 實作）

本專案以 Go + Gin 實作容器管理系統的 RESTful API。重點包含清楚分層、Middleware、抽象介面（Provider）、SQLite 資料持久化、檔案上傳與 OpenAPI 文件。

## 專案結構

```
容器管理系統/
├─ cmd/server/main.go
├─ internal/
│  ├─ server/router.go
│  ├─ middleware/{auth,logger,recover}.go
│  ├─ api/handlers/{health,upload,containers}.go
│  ├─ containers/{provider.go,mock_provider.go,docker_provider.go,service.go}
│  ├─ storage/{db.go,container_repository.go}
│  └─ users/model.go
├─ api/openapi.yaml
├─ go.mod
└─ README.md
```

## 建置與執行

### 方式一：直接執行（開發用）

1. 安裝 Go 1.22+。
2. （可選）設定環境變數：

```bash
export PORT=8080
export DATA_DIR=./data
export PROVIDER=mock  # mock | docker
export JWT_SECRET=devsecret
export AUTH_USER=admin
export AUTH_PASS=admin
```

3. 安裝依賴並啟動：

```bash
go mod tidy
go run ./cmd
```

### 方式二：Docker Compose（生產用）

1. 設定環境變數（在 `container_manager` 目錄下）：

```bash
# 設定宿主機 data 目錄的絕對路徑（用於 Docker bind mount）
export HOST_DATA_DIR=$(pwd)/data

# 或手動指定絕對路徑
# export HOST_DATA_DIR=/path/to/container_manager/data
```

2. 啟動服務：

```bash
docker compose up --build
```

服務將在 `http://localhost:8081` 啟動。

**注意**：使用 Docker Compose 時，`HOST_DATA_DIR` 必須是**絕對路徑**，因為容器內的程式需要知道宿主機上 data 目錄的實際位置，才能正確進行 Docker bind mount。

## 測試請求範例

- 健康檢查：

```
curl http://localhost:8080/healthz
```

- 上傳檔案：

```
curl -H "Authorization: Bearer devtoken" \
     -F "files=@/path/to/a.csv" -F "files=@/path/to/b.json" \
     -F "userId=u123" \
     http://localhost:8080/v1/uploads
```

- 建立 / 啟動 / 停止 / 刪除容器（Mock Provider）：

```
# 建立
curl -H "Authorization: Bearer devtoken" -H "Content-Type: application/json" \
     -d '{"image":"alpine:latest","name":"demo"}' \
     http://localhost:8080/v1/containers

# 假設回傳 {"id":"<CID>"...}
CID=<回傳的 id>

curl -H "Authorization: Bearer devtoken" -X POST http://localhost:8080/v1/containers/$CID/start
curl -H "Authorization: Bearer devtoken" -X POST http://localhost:8080/v1/containers/$CID/stop
curl -H "Authorization: Bearer devtoken" -X DELETE http://localhost:8080/v1/containers/$CID
```

## API 規格

OpenAPI 位於 `api/openapi.yaml`，可匯入 Swagger UI 或 Postman 使用。

## 測試

執行 `go test ./...` 可進行單元與整合測試；Mock Provider 使得在 CI 環境（無 Docker）也能順利測試。
