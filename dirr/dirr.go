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
	// Starting from the absStartDir, search upwards one level at a time
	dir := absStartDir
	for ; deepth > 0; deepth -= 1 {
		// Build full file path
		filePath := filepath.Join(dir, filename)

		// Check if the file exists
		if _, err := os.Stat(filePath); err == nil { // If file exists, return the full path
			return filePath, nil
		} else if !os.IsNotExist(err) { // Return the error that's not ErrNotExist
			return "", err
		}

		// If the file is not found and the root directory has been reached, stop searching
		if dir == sep || dir == "." || dir == "/" || dir == filepath.Dir(dir) {
			break
		}

		// Move up to the parent directory and continue searching
		dir = filepath.Dir(dir)
	}

	// The file does not exist in any parent directory
	return "", os.ErrNotExist
}
