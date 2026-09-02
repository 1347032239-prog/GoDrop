package main

import (
	"io/fs"
	"path/filepath"
)

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type ScanResult struct {
	Files     []FileEntry `json:"files"`
	TotalSize int64       `json:"total_size"`
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

			fileEntryTemp := FileEntry{
				Path: relPath,
				Size: sz,
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
