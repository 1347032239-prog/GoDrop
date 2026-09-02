package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteIndex(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     map[string]string
		writeDir       string
		expectedCount  int
		expectedSize   int64
		expectedError1 bool
		expectedError2 bool
		flag           int
	}{
		{
			name: "previous",
			setupFiles: map[string]string{
				"a.txt":       "abc",
				"bbb":         "",
				"sub/subFile": "12345",
			},
			writeDir:       "",
			expectedCount:  3,
			expectedSize:   8,
			expectedError1: false,
			flag:           0,
		},
		{
			name: "newNormal",
			setupFiles: map[string]string{
				"new1.txt": "12",
				"new2.txt": "1234",
			},
			writeDir:       "/tmp/godrop-index.json",
			expectedCount:  2,
			expectedSize:   6,
			expectedError1: false,
			expectedError2: false,
			flag:           0,
		},
		{
			name: "outDirNoExisting",
			setupFiles: map[string]string{
				"new1.txt": "12",
				"new2.txt": "1234",
			},
			writeDir:       "index.json",
			expectedCount:  2,
			expectedSize:   6,
			expectedError1: false,
			expectedError2: true,
			flag:           1,
		},
	}

	for _, TT := range tests {
		t.Run(TT.name, func(t *testing.T) {

			tmpDir := t.TempDir()

			for relPath, data := range TT.setupFiles {
				fullPath := filepath.Join(tmpDir, relPath)
				MkdirErr := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if MkdirErr != nil {
					t.Fatalf("创建辅助目录失败, 错误信息:%v", MkdirErr)
				}

				WriteFileErr := os.WriteFile(fullPath, []byte(data), 0644)
				if WriteFileErr != nil {
					t.Fatalf("将数据写入辅助文件失败, 错误信息:%v", WriteFileErr)
				}

			}

			result, err := scanDirectory(tmpDir)

			if (err != nil) != TT.expectedError1 {
				t.Fatalf("错误状态不匹配, 期望(err != nil)为%v, 实际为:%v", TT.expectedError1, err)
			}

			if len(result.Files) != TT.expectedCount {
				t.Errorf("文件数量错误, 期望%v, 实际为%v", TT.expectedCount, len(result.Files))
			}

			if result.TotalSize != TT.expectedSize {
				t.Errorf("文件大小错误, 期望%v, 实际为%v", TT.expectedSize, result.TotalSize)
			}

			if TT.writeDir != "" {
				jsonDir := TT.writeDir
				if TT.flag == 1 {
					jsonDir = filepath.Join(tmpDir, "missing", TT.writeDir)
				}
				werr := writeIndex(jsonDir, result)

				if (werr != nil) != TT.expectedError2 {
					t.Fatalf("错误状态不匹配, 期望(werr != nil)为%v, 实际为:%v", TT.expectedError2, werr)
				}

				if TT.expectedError2 {
					return
				}

				data, err := os.ReadFile(jsonDir)
				if err != nil {
					t.Fatalf("读取json文件失败")
				}

				var decoded ScanResult
				rerr := json.Unmarshal(data, &decoded)
				if rerr != nil {
					t.Fatalf("反序列化失败")
				}
				if len(decoded.Files) != TT.expectedCount {
					t.Errorf("文件数量错误, 期望%v, 实际为%v", TT.expectedCount, len(decoded.Files))
				}

				if decoded.TotalSize != TT.expectedSize {
					t.Errorf("文件大小错误, 期望%v, 实际为%v", TT.expectedSize, decoded.TotalSize)
				}

			}

		})
	}
}
