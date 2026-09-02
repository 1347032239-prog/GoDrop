package main

import (
	"encoding/json"
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
