package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"

	"github.com/pelletier/go-toml/v2"
)

type ConfigSource int

const (
	ConfigSourceFile ConfigSource = iota + 1
	ConfigSourceDB
)

type ConfigDiff struct {
	Field     string
	FileValue any
	DBValue   any
}

func WriteConfig(cfg Config, path string) error {
	dat, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	writeErr := os.WriteFile(path, dat, 0644)
	if writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}
	return nil
}

func ReadConfigFromFile(path string) (Config, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(dat, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %w", err)
	}
	return cfg, nil
}

func ReadConfigFromDB(q *sqlc.Queries) (Config, error) {
	var cfg Config
	dbCfg, err := q.GetConfig(context.Background())
	if err != nil {
		return Config{}, fmt.Errorf("failed to get config from db: %w", err)
	}
	return cfg.FromDB(dbCfg), nil
}

func LoadAndCheckConfig(lg logger.Logger) (Config, error) {
	cfg, readFileErr := ReadConfigFromFile("config.toml")
	if readFileErr != nil {
		lg.Fatal("Failed to read config from file", "error", readFileErr)
	}
	s, err := db.InitIfNeeded()
	if err != nil {
		lg.Fatal("Failed to initialize database", "error", err)
	}
	dbCfg, readErr := ReadConfigFromDB(s.Queries)
	if readErr != nil {
		lg.Fatal("Failed to read config from db", "error", readErr)
	}

	diffs := diffConfigs(cfg, dbCfg)
	if len(diffs) == 0 {
		return cfg, nil
	}

	fmt.Printf("Found %d config difference(s)\n", len(diffs))
	fmt.Println("Diffs marked in \033[1;31mbold red\033[0m")
	cfg, source := chooseConfig(cfg, dbCfg, diffs)
	switch source {
	case ConfigSourceFile:
		updateErr := s.Queries.UpdateConfig(context.Background(), cfg.ToUpdateParams())
		if updateErr != nil {
			return Config{}, fmt.Errorf("failed to update config in db: %w", updateErr)
		}
	case ConfigSourceDB:
		WriteConfig(dbCfg, "config.toml")
	}

	return cfg, nil
}

func diffConfigs(fileCfg, dbCfg Config) []ConfigDiff {
	var diffs []ConfigDiff

	fileVal := reflect.ValueOf(fileCfg)
	dbVal := reflect.ValueOf(dbCfg)
	fileType := reflect.TypeOf(fileCfg)

	for i := 0; i < fileType.NumField(); i++ {
		field := fileType.Field(i)
		fileField := fileVal.Field(i).Interface()
		dbField := dbVal.Field(i).Interface()

		if !reflect.DeepEqual(fileField, dbField) {
			diffs = append(diffs, ConfigDiff{
				Field:     field.Name,
				FileValue: fileField,
				DBValue:   dbField,
			})
		}
	}

	return diffs
}

func chooseConfig(cfg Config, dbCfg Config, diffs []ConfigDiff) (Config, ConfigSource) {
	diffFields := make(map[string]bool)
	fieldLabelMap := map[string]string{
		"WatchDir":           "Watch Dir",
		"OrganizedDir":       "Organized Dir",
		"ArchiveDir":         "Archive Dir",
		"Port":               "Port",
		"ExposeService":      "Service Exposed",
		"CompressionEnabled": "Compression Enabled",
		"ArchiveDays":        "Archive Days",
		"MaxRetentionDays":   "Max Retention Days",
		"CleanupIntervalHrs": "Cleanup Interval Hours",
		"BatchSize":          "Batch Size",
		"LogLevel":           "Log Level",
		"UpdatedAt":          "Updated At",
	}

	for _, diff := range diffs {
		diffFields[fieldLabelMap[diff.Field]] = true
	}

	for {
		fmt.Println("\n=== Config Selection ===")
		fmt.Println("\n[f] File Config:")
		printConfigFormatted(cfg, diffFields)

		fmt.Println("\n[d] Database Config:")
		printConfigFormatted(dbCfg, diffFields)

		fmt.Println("\n[q] Exit")
		fmt.Print("\nEnter your choice (f,d,q): ")

		var choice string
		_, err := fmt.Scanf("%s\n", &choice)
		if err != nil {
			fmt.Println("Invalid input, please try again")
			continue
		}

		choice = strings.TrimSpace(choice)

		switch choice {
		case "f":
			return cfg, ConfigSourceFile
		case "d":
			return dbCfg, ConfigSourceDB
		case "q":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid choice, please select f, d, or q")
		}
	}
}

func printConfigFormatted(cfg Config, diffFields map[string]bool) {
	if reflect.DeepEqual(cfg, Config{}) {
		fmt.Println("Config is empty")
		return
	}

	fields := []struct {
		label string
		value any
	}{
		{"Watch Dir", cfg.WatchDir},
		{"Organized Dir", cfg.OrganizedDir},
		{"Archive Dir", cfg.ArchiveDir},
		{"Service Exposed", cfg.ExposeService},
		{"Port", cfg.Port},
		{"Compression Enabled", cfg.CompressionEnabled},
		{"Archive Days", cfg.ArchiveDays},
		{"Max Retention Days", cfg.MaxRetentionDays},
		{"Log Level", cfg.LogLevel},
	}

	for _, f := range fields {
		if diffFields[f.label] {
			fmt.Printf("  %s: \033[1;31m%v\033[0m\n", f.label, f.value)
		} else {
			fmt.Printf("  %s: %v\n", f.label, f.value)
		}
	}
}
