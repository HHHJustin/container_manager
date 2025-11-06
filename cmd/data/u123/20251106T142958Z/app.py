#!/usr/bin/env python3
import json
import csv
import os

# 讀取 CSV
csv_lines = 0
if os.path.exists('/workspace/a.csv'):
    with open('/workspace/a.csv', 'r') as f:
        csv_lines = sum(1 for _ in f)

# 讀取 JSON
json_bytes = 0
if os.path.exists('/workspace/b.json'):
    with open('/workspace/b.json', 'rb') as f:
        json_bytes = len(f.read())

# 輸出結果
with open('/workspace/result.txt', 'w') as f:
    f.write(f"CSV lines: {csv_lines}\n")
    f.write(f"JSON bytes: {json_bytes}\n")
    f.write("done\n")

# 列出所有檔案
with open('/workspace/list.txt', 'w') as f:
    for item in os.listdir('/workspace'):
        f.write(f"{item}\n")

print("Python script completed")

