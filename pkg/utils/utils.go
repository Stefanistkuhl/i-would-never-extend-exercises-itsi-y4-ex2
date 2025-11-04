package utils

import (
	"os"
	"path/filepath"
	"runtime"
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

func GetDiskRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		volume := filepath.VolumeName(absPath)
		if volume == "" {
			return "\\", nil
		}
		return volume + "\\", nil
	}

	return absPath, nil
}

func GetDiskPathFromConfig(archiveDir, organizedDir, watchDir string) (string, string, error) {
	dirs := []struct {
		path   string
		source string
	}{
		{archiveDir, "archive_dir"},
		{organizedDir, "organized_dir"},
		{watchDir, "watch_dir"},
	}

	var lastErr error
	for _, dir := range dirs {
		if dir.path == "" {
			continue
		}
		diskRoot, err := GetDiskRoot(dir.path)
		if err == nil {
			return diskRoot, dir.source, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", "", lastErr
	}
	return "/", "default", nil
}

type DiskInfo struct {
	Path        string
	SourceDirs  []string
	ConfigPaths []string
}

func GetAllDisksFromConfig(archiveDir, organizedDir, watchDir string) ([]DiskInfo, error) {
	dirs := []struct {
		path   string
		source string
	}{
		{archiveDir, "archive_dir"},
		{organizedDir, "organized_dir"},
		{watchDir, "watch_dir"},
	}

	diskMap := make(map[string]*DiskInfo)

	for _, dir := range dirs {
		if dir.path == "" {
			continue
		}
		diskRoot, err := GetDiskRoot(dir.path)
		if err != nil {
			continue
		}

		normalizedPath := filepath.Clean(diskRoot)
		if runtime.GOOS == "windows" {
			normalizedPath = strings.ToUpper(normalizedPath)
		}

		if _, exists := diskMap[normalizedPath]; !exists {
			diskMap[normalizedPath] = &DiskInfo{
				Path:        diskRoot,
				SourceDirs:  []string{dir.source},
				ConfigPaths: []string{dir.path},
			}
		} else {
			// Multiple directories on the same disk - add this one too
			diskInfo := diskMap[normalizedPath]
			diskInfo.SourceDirs = append(diskInfo.SourceDirs, dir.source)
			diskInfo.ConfigPaths = append(diskInfo.ConfigPaths, dir.path)
		}
	}

	result := make([]DiskInfo, 0, len(diskMap))
	for _, disk := range diskMap {
		result = append(result, *disk)
	}
	if len(result) == 0 {
		return []DiskInfo{
			{
				Path:        "/",
				SourceDirs:  []string{"default"},
				ConfigPaths: []string{""},
			},
		}, nil
	}

	return result, nil
}
