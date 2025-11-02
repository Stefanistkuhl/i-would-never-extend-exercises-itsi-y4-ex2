package sorter

import (
	"fmt"
	"os"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
)

func loadRootDirs(cfg config.Config) error {
	dirs := []string{
		cfg.WatchDir,
		cfg.OrganizedDir,
		cfg.ArchiveDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", dir, err)
			}
		}
	}
	return nil
}

func checkIfStructureExists(cfg config.Config) error {
	dirs := map[string]string{
		"watch":     cfg.WatchDir,
		"organized": cfg.OrganizedDir,
		"archive":   cfg.ArchiveDir,
	}

	for name, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create %s dir (%s): %w", name, dir, err)
			}
		}
	}
	return nil
}

func ValidateConfigPaths(cfg config.Config) error {
	if cfg.WatchDir == "" {
		return fmt.Errorf("watch_dir is empty")
	}
	if cfg.OrganizedDir == "" {
		return fmt.Errorf("organized_dir is empty")
	}
	if cfg.ArchiveDir == "" {
		return fmt.Errorf("archive_dir is empty")
	}

	if err := checkIfStructureExists(cfg); err != nil {
		return err
	}

	return nil
}

func InitSorter(cfg config.Config) error {
	if err := ValidateConfigPaths(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := loadRootDirs(cfg); err != nil {
		return fmt.Errorf("failed to load root dirs: %w", err)
	}

	return nil
}
