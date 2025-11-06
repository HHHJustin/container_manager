// +build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func main() {
	// 讀取 CSV
	csvLines := 0
	if f, err := os.Open("/workspace/a.csv"); err == nil {
		r := csv.NewReader(f)
		for {
			_, err := r.Read()
			if err == io.EOF {
				break
			}
			csvLines++
		}
		f.Close()
	}

	// 讀取 JSON
	jsonBytes := int64(0)
	if info, err := os.Stat("/workspace/b.json"); err == nil {
		jsonBytes = info.Size()
	}

	// 輸出結果
	result, _ := os.Create("/workspace/result.txt")
	fmt.Fprintf(result, "CSV lines: %d\n", csvLines)
	fmt.Fprintf(result, "JSON bytes: %d\n", jsonBytes)
	fmt.Fprintln(result, "done")
	result.Close()

	// 列出所有檔案
	list, _ := os.Create("/workspace/list.txt")
	entries, _ := os.ReadDir("/workspace")
	for _, e := range entries {
		fmt.Fprintln(list, e.Name())
	}
	list.Close()

	fmt.Println("Go program completed")
}

