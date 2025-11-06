#!/bin/sh
set -e

# 列出工作目錄內容
ls -la /workspace > /workspace/list.txt

# 簡單讀取示例檔案並輸出報告
CSV_LINES=$(wc -l < /workspace/a.csv || echo 0)
JSON_BYTES=$(wc -c < /workspace/b.json || echo 0)
{
  echo "CSV lines: $CSV_LINES"
  echo "JSON bytes: $JSON_BYTES"
  echo "done"
} > /workspace/result.txt


