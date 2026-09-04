package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeIndex(path string, result ScanResult) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close() //这里一开始写成了os.Close(file)

	encoder := json.NewEncoder(file)

	encoder.SetIndent("", "  ") //indent缩进

	encodeErr := encoder.Encode(result)
	if encodeErr != nil {
		return encodeErr
	}

	return nil
}

func readIndex(path string) (ScanResult, error) {
	var result ScanResult
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	derr := decoder.Decode(&result)
	if derr != nil {
		return result, derr
	}

	return result, nil

}

func printScanResult(result ScanResult) {
	length := len(result.Files)
	for i := 0; i < length; i++ {
		fmt.Printf("%v %v bytes\n", result.Files[i].Path, result.Files[i].Size)
	}
	fmt.Printf("文件数量: %v\n", length)
	fmt.Printf("总大小: %v bytes\n", result.TotalSize)
}
