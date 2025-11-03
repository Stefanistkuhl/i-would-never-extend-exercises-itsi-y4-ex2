package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func GetSubdirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var subdirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
		}
	}
	return subdirs, nil
}

func StripCommonPrefix(p1, p2 string) string {
	parts1 := strings.Split(filepath.Clean(p1), string(filepath.Separator))
	parts2 := strings.Split(filepath.Clean(p2), string(filepath.Separator))

	commonLen := 0
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] == parts2[i] {
			commonLen++
		} else {
			break
		}
	}

	remaining := parts1[commonLen:]
	return filepath.Join(remaining...)

}
