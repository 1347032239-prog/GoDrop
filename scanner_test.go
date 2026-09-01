package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    map[string]string
		expectedCount int
		expectedSize  int64
		expectedError bool
	}{
		{
			name: "包含子目录文件",
			setupFiles: map[string]string{
				"a.txt":       "abc",
				"bbb":         "",
				"sub/subFile": "12345",
			},
			expectedCount: 3,
			expectedSize:  8,
			expectedError: false,
		},
		{
			name:          "空目录",
			setupFiles:    map[string]string{},
			expectedCount: 0,
			expectedSize:  0,
			expectedError: false,
		},
		{
			name:          "不存在的目录",
			expectedError: true,
		},
	}
	for i := 0; i < 2; i++ {
		t.Run(tests[i].name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for relPath, data := range tests[i].setupFiles {
				fullPath := filepath.Join(tmpDir, relPath)
				MkdirErr := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if MkdirErr != nil {
					t.Fatalf("创建目录失败")
				}

				WriteFileErr := os.WriteFile(fullPath, []byte(data), 0644)
				if WriteFileErr != nil {
					t.Fatalf("写文件失败")
				}
			}
			result, err := scanDirectory(tmpDir)
			if (err != nil) != tests[i].expectedError {
				t.Fatalf("错误状态不匹配, 期望(err != nil)为%v, 实际为:%v", tests[i].expectedError, err)
			}

			if tests[i].expectedCount != len(result.Files) {
				t.Errorf("文件数量错误, 期望%v, 实际为%v", tests[i].expectedCount, len(result.Files))
			}

			if tests[i].expectedSize != result.TotalSize {
				t.Errorf("文件大小错误, 期望%v, 实际为%v", tests[i].expectedSize, result.TotalSize)
			}

		})
	}

	t.Run(tests[2].name, func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := scanDirectory(filepath.Join(tmpDir, "missing"))

		if (err != nil) != tests[2].expectedError {
			t.Fatalf("错误状态不匹配, 期望(err != nil)为%v, 实际为:%v", tests[2].expectedError, err)
		}
	})

}
