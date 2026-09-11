package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FileEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type ScanResult struct {
	Files     []FileEntry `json:"files"`
	TotalSize int64       `json:"total_size"`
}

func calculateFileSHA256(path string) (string, error) {

	var ret string

	hasher := sha256.New()

	file, err := os.Open(path)

	if err != nil {

		return ret, fmt.Errorf("打开文件失败, 错误信息为:%w", err)

	}

	defer file.Close()

	_, copyErr := io.Copy(hasher, file)

	if copyErr != nil {

		return ret, fmt.Errorf("流式读取失败, 错误信息为:%w", copyErr)

	}

	digest := hasher.Sum(nil)

	ret = hex.EncodeToString(digest)

	return ret, nil

}

func scanDirectory(root string) (ScanResult, error) {

	result := ScanResult{
		Files: make([]FileEntry, 0),
	}

	walkDirErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		sz := fileInfo.Size()

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if fileInfo.Mode().IsRegular() {

			tmp, err := calculateFileSHA256(path)

			if err != nil {

				return fmt.Errorf("服务端获取文件校验码失败, 错误信息:%w", err)

			}

			fileEntryTemp := FileEntry{
				Path:   filepath.ToSlash(relPath),
				Size:   sz,
				SHA256: tmp,
			}

			result.Files = append(result.Files, fileEntryTemp)
			result.TotalSize += sz
		}
		return nil
	})

	if walkDirErr != nil {

		return ScanResult{}, walkDirErr
	}

	return result, nil
}
