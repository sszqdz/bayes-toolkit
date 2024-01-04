package dirr

import (
	"errors"
	"os"
	"path/filepath"
)

var sep = string(os.PathSeparator)

func FindFileInParentDirs(filename, absStartDir string, deepth int) (string, error) {
	if !filepath.IsAbs(absStartDir) {
		return "", errors.New("must be absolute path")
	}
	if deepth <= 0 {
		deepth = 99
	}
	// 从指定目录开始逐层向上查找
	dir := absStartDir
	for ; deepth > 0; deepth -= 1 {
		// 构建完整的文件路径
		filePath := filepath.Join(dir, filename)

		// 检查文件是否存在
		if _, err := os.Stat(filePath); err == nil { // 文件存在，返回文件的完整路径
			return filePath, nil
		} else if !os.IsNotExist(err) { // 遇到不是 ErrNotExist 的错误，返回该错误
			return "", err
		}

		// 如果没有找到文件，并且已经到了根目录，停止查找
		if dir == sep || dir == "." || dir == "/" || dir == filepath.Dir(dir) {
			break
		}

		// 向上移动到父目录，继续查找
		dir = filepath.Dir(dir)
	}

	// 文件不存在于任何父目录中
	return "", os.ErrNotExist
}
