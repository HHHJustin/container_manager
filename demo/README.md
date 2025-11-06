# Demo 程式

一鍵測試 檔案上傳→容器啟動→輸出結果。

## 執行步驟

1. 啟動 API（預設 8081）
```bash
cd ../
PORT=8081 go run ./cmd
```

2. 在另一個終端執行 demo
```bash
cd demo
# 若要改 API 位址：export BASE_URL=http://localhost:8081
# 若要用固定 Token：export AUTH_TOKEN=devtoken
go run .
```

流程說明
- 無 AUTH_TOKEN 時會自動對 `/login` 以 admin/admin 換 JWT
- 上傳 `demo/files/a.csv` 與 `demo/files/b.json` 到伺服器
- 根據 `PROG_TYPE` 選擇上傳並執行對應的程式檔案
- 執行一次性作業：把上傳資料夾掛到容器 `/workspace`，執行對應程式
- 顯示 exitCode 與可能的靜態檔案路徑供檢視

## 支援的程式類型

### Shell 腳本（預設）
```bash
export PROG_TYPE=sh
go run .
```
- 上傳：`files/run.sh`
- 映像：`alpine:3.20`
- 執行：`sh run.sh`

### Python 腳本
```bash
export PROG_TYPE=py
go run .
```
- 上傳：`files/app.py`
- 映像：`python:3.11-slim`
- 執行：`python app.py`

### Go 程式
```bash
export PROG_TYPE=go
go run .
```
- 上傳：`files/app.go`
- 映像：`golang:1.21-alpine`
- 執行：`go run app.go`

### 已編譯執行檔
```bash
export PROG_TYPE=binary
go run .
```
- 上傳：`files/app`（需先編譯好）
- 映像：`alpine:3.20`
- 執行：`./app`

## 參數
- `BASE_URL`：API 位置（預設 `http://localhost:8081`）
- `AUTH_TOKEN`：固定 Bearer Token（存在則不呼叫 /login）
- `AUTH_USER`、`AUTH_PASS`：登入帳密（預設 admin/admin）
- `PROG_TYPE`：程式類型（預設 `sh`，可選：`sh`, `py`, `go`, `binary`）
- `JOB_IMAGE`：作業容器映像（會根據 PROG_TYPE 自動選擇，可手動覆蓋）


