package config

import (
	"database/sql"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
)

type Config struct {
	WatchDir           string `toml:"watch_dir"`
	OrganizedDir       string `toml:"organized_dir"`
	ArchiveDir         string `toml:"archive_dir"`
	CompressionEnabled bool   `toml:"compression_enabled"`
	ArchiveDays        int    `toml:"archive_days"`
	MaxRetentionDays   int    `toml:"max_retention_days"`
	LogLevel           string `toml:"log_level"`
}

func (c Config) FromDB(dbCfg sqlc.Config) Config {
	return Config{
		WatchDir:           dbCfg.WatchDir.String,
		OrganizedDir:       dbCfg.OrganizedDir.String,
		ArchiveDir:         dbCfg.ArchiveDir.String,
		CompressionEnabled: dbCfg.CompressionEnabled.Bool,
		ArchiveDays:        int(dbCfg.ArchiveDays.Int64),
		MaxRetentionDays:   int(dbCfg.MaxRetentionDays.Int64),
		LogLevel:           dbCfg.LogLevel.String,
	}
}

func (c Config) ToUpdateParams() sqlc.UpdateConfigParams {
	return sqlc.UpdateConfigParams{
		WatchDir:           sql.NullString{String: c.WatchDir, Valid: c.WatchDir != ""},
		OrganizedDir:       sql.NullString{String: c.OrganizedDir, Valid: c.OrganizedDir != ""},
		ArchiveDir:         sql.NullString{String: c.ArchiveDir, Valid: c.ArchiveDir != ""},
		CompressionEnabled: sql.NullBool{Bool: c.CompressionEnabled, Valid: true},
		ArchiveDays:        sql.NullInt64{Int64: int64(c.ArchiveDays), Valid: c.ArchiveDays > 0},
		MaxRetentionDays:   sql.NullInt64{Int64: int64(c.MaxRetentionDays), Valid: c.MaxRetentionDays > 0},
		LogLevel:           sql.NullString{String: c.LogLevel, Valid: c.LogLevel != ""},
	}
}
