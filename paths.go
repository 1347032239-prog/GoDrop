package main

import (
	"errors"
	"path"
	"path/filepath"
	"strings"
)

// validateIndexPath checks the platform-independent path stored in an index.
// Index paths always use forward slashes; conversion to a native filesystem
// path happens only immediately before local filesystem access.
func validateIndexPath(indexPath string) error {
	if indexPath == "" {
		return errors.New("索引路径不能为空")
	}
	if indexPath == "." {
		return errors.New("索引路径不能是当前目录")
	}
	if strings.Contains(indexPath, `\`) {
		return errors.New("索引路径必须使用正斜杠")
	}
	if path.IsAbs(indexPath) {
		return errors.New("索引路径不能是绝对路径")
	}

	cleaned := path.Clean(indexPath)
	if cleaned != indexPath {
		return errors.New("索引路径不是规范路径")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return errors.New("索引路径不能逃逸共享目录")
	}

	if !filepath.IsLocal(filepath.FromSlash(indexPath)) {
		return errors.New("索引路径不是本地相对路径")
	}

	return nil
}
