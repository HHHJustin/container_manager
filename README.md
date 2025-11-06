# 容器管理系統

本專案以 Go + Gin 實作容器管理系統的 RESTful API。重點包含：清楚分層、Middleware、抽象介面（Provider：mock/docker）、PostgreSQL 紀錄容器狀態、檔案上傳與 OpenAPI 文件、JWT 驗證。

## 專案結構

```
容器管理系統/
├─ cmd/main.go
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
export PROVIDER=mock  # mock | docker（啟用 docker 時需本機 Docker 可用）
export JWT_SECRET=devsecret
export AUTH_USER=admin
export AUTH_PASS=admin
```

3. 安裝依賴並啟動：

```bash
go mod tidy
go run ./cmd
```

### 方式二：Docker Compose（容器化）

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

## OpenAPI 與範例

檔案：`api/openapi.yaml`

### 快速測試 OpenAPI
- Swagger Editor：`https://editor.swagger.io` → Import 本檔
- 本地 Swagger UI：
  ```bash
  docker run --rm -p 8082:8080 \
    -e SWAGGER_JSON=/spec/openapi.yaml \
    -v "$(pwd)/api/openapi.yaml:/spec/openapi.yaml" \
    swaggerapi/swagger-ui
  ```
  打開 `http://localhost:8082`，先呼叫 `POST /login` 取得 JWT，點 Authorize 帶入 `Bearer <token>`。

### 範例 Request/Response
- 登入（POST /login）
  ```json
  { "username": "admin", "password": "admin" }
  ```
  回應：
  ```json
  { "token": "<JWT>" }
  ```
- 上傳（POST /v1/uploads）回應：
  ```json
  { "userId": "u123", "dir": "./data/u123/2025...", "files": ["..."] }
  ```
- 執行作業（POST /v1/jobs）請求（顯式指定）：
  ```json
  {
    "image": "python:3.11-slim",
    "hostDir": "<上一步回傳的 dir>",
    "containerDir": "/workspace",
    "cmd": ["python", "/workspace/app.py"]
  }
  ```
  回應：
  ```json
  { "exitCode": 0, "logs": "..." }
  ```
  若未提供 `image/cmd`，系統會自動偵測（app 可執行 > run.sh > app.py > app.go）。

## API 規格

可將 `api/openapi.yaml` 匯入 Swagger UI / Postman / Insomnia 直接操作。

## 測試

執行 `go test ./...` 可進行單元與整合測試；Mock Provider 使得在 CI（無 Docker）也能順利測試。

## 注意事項
- 驗證採 JWT：請先 `POST /login` 取得 Token，再於受保護路由以 `Authorization: Bearer <token>` 呼叫。
- Docker Compose 時請務必設定 `HOST_DATA_DIR` 為宿主機的**絕對路徑**，並與 `DATA_DIR=/app/data` 映射一致。
