package sorter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/capture"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
)

type FilenameValidationResult struct {
	IsValid         bool
	Hostname        string
	Scenario        string
	CaptureDateTime time.Time
	Error           string
}

func StartSorter() {
	logger := log.New(os.Stdout, "[pcap-sorter] ", log.LstdFlags)
	cfg, cfgErr := config.LoadAndCheckConfig()
	if cfgErr != nil {
		logger.Fatal("ERROR: Failed to load config:", cfgErr)
	}
	if err := InitSorter(cfg); err != nil {
		logger.Fatal("ERROR: Failed to initialize sorter:", err)
	}
	Watcher(cfg, logger)
}

func ValidateFilename(filePath string, cfg config.Config, logger *log.Logger) FilenameValidationResult {
	result := FilenameValidationResult{IsValid: false}
	filename := filepath.Base(filePath)

	// Remove .pcap or .pcap.gz extension
	base := strings.TrimSuffix(
		strings.TrimSuffix(filename, ".gz"),
		".pcap",
	)

	// Regex: {hostname}_{scenario}_{YYYYMMDD_HHmmss}
	re := regexp.MustCompile(`^\{([a-zA-Z0-9_-]+)\}_\{([a-zA-Z0-9_-]+)\}_\{(\d{8})_(\d{6})\}$`)
	matches := re.FindStringSubmatch(base)

	if len(matches) != 5 {
		newPath := filePath + ".INCORRECT"
		err := os.Rename(filePath, newPath)
		if err != nil {
			result.Error = "Failed to rename file: " + err.Error()
			logger.Printf("ERROR: Failed to rename %s to %s: %v", filePath, newPath, err)
			return result
		}

		result.Error = "Invalid filename format"
		if cfg.LogLevel == "info" {
			logger.Printf("INFO: Renamed invalid filename %s to %s", filePath, newPath)
		}
		return result
	}

	result.Hostname = matches[1]
	result.Scenario = matches[2]
	dateStr := matches[3] // YYYYMMDD
	timeStr := matches[4] // HHmmss

	// Parse datetime: YYYYMMDD + HHmmss
	datetimeStr := dateStr + timeStr
	captureDateTime, err := time.Parse("20060102150405", datetimeStr)
	if err != nil {
		newPath := filePath + ".INCORRECT"
		os.Rename(filePath, newPath)
		result.Error = "Invalid datetime: " + err.Error()
		if cfg.LogLevel == "info" {
			logger.Printf("INFO: Renamed invalid datetime %s to %s", filePath, newPath)
		}
		return result
	}

	result.IsValid = true
	result.CaptureDateTime = captureDateTime
	if cfg.LogLevel == "info" {
		logger.Printf("INFO: Valid filename - hostname=%s, scenario=%s, datetime=%s",
			result.Hostname, result.Scenario, result.CaptureDateTime.Format(time.RFC3339))
	}

	return result
}

func processFile(cfg config.Config, path string, logger *log.Logger) {
	if cfg.LogLevel == "info" {
		logger.Println("INFO: Processing file:", path)
	}
	result := ValidateFilename(path, cfg, logger)
	if !result.IsValid {
		if cfg.LogLevel == "info" {
			logger.Println("WARNING: Filename is invalid:", result.Error)
		}
		return
	}
	organizedPath := filepath.Join(cfg.OrganizedDir, result.Hostname, result.CaptureDateTime.UTC().Format(time.RFC3339))
	if _, err := os.Stat(organizedPath); os.IsNotExist(err) {
		if err := os.MkdirAll(organizedPath, os.ModePerm); err != nil {
			logger.Println("ERROR: Failed to create organized dir:", err)
			return
		}
	}
	organizedFilePath := filepath.Join(organizedPath, fmt.Sprintf("%s-%s-%s", result.Hostname, result.Scenario, result.CaptureDateTime.UTC().Format(time.RFC3339)))
	renameErr := os.Rename(path, organizedFilePath)
	if renameErr != nil {
		logger.Println("ERROR: Failed to rename file:", renameErr)
		return
	}

	info, infoErr := os.Stat(organizedFilePath)
	if infoErr != nil {
		logger.Println("ERROR: Failed to get file info:", infoErr)
		return
	}

	s, getQeuryErr := db.InitIfNeeded()
	if getQeuryErr != nil {
		logger.Println("ERROR: Failed to get db queries:", getQeuryErr)
		return
	}

	caputureParams := sqlc.InsertCaptureParams{
		Hostname:        result.Hostname,
		Scenario:        result.Scenario,
		CaptureDatetime: result.CaptureDateTime,
		FilePath:        organizedFilePath,
		FileSize:        info.Size(),
		Compressed:      sql.NullBool{Bool: false, Valid: true},
		Archived:        sql.NullBool{Bool: false, Valid: true},
		CreatedAt:       sql.NullTime{Time: result.CaptureDateTime, Valid: true},
		UpdatedAt:       sql.NullTime{Time: result.CaptureDateTime, Valid: true},
	}

	res, parseErr := capture.AnalyzeCaptureFile(cfg, organizedFilePath)
	if parseErr != nil {
		logger.Println("ERROR: Failed to parse file:", parseErr)
		return
	}

	protDist, marshallProtDistErr := json.Marshal(res.ProtocolDistribution)
	if marshallProtDistErr != nil {
		logger.Println("ERROR: Failed to marshal protocol distribution:", marshallProtDistErr)
		return
	}
	topSrcIps, marshallTopSrcIpsErr := json.Marshal(res.TopSrcIPs)
	if marshallTopSrcIpsErr != nil {
		logger.Println("ERROR: Failed to marshal top src ips:", marshallTopSrcIpsErr)
		return
	}
	topDstIps, marshallTopDstIpsErr := json.Marshal(res.TopDstIPs)
	if marshallTopDstIpsErr != nil {
		logger.Println("ERROR: Failed to marshal top dst ips:", marshallTopDstIpsErr)
		return
	}
	topPorts, marshallTopPortsErr := json.Marshal(res.TopPorts)
	if marshallTopPortsErr != nil {
		logger.Println("ERROR: Failed to marshal top ports:", marshallTopPortsErr)
		return
	}
	tlsVersions, marshallTlsVersionsErr := json.Marshal(res.TLSVersions)
	if marshallTlsVersionsErr != nil {
		logger.Println("ERROR: Failed to marshal tls versions:", marshallTlsVersionsErr)
		return
	}

	statParams := sqlc.InsertCaptureStatsParams{
		PacketCount:          sql.NullInt64{Int64: int64(res.TotalPackets), Valid: true},
		ProtocolDistribution: sql.NullString{String: string(protDist), Valid: true},
		TopSrcIps:            sql.NullString{String: string(topSrcIps), Valid: true},
		TopDstIps:            sql.NullString{String: string(topDstIps), Valid: true},
		TopPorts:             sql.NullString{String: string(topPorts), Valid: true},
		PacketRate:           sql.NullFloat64{Float64: res.PacketRate, Valid: true},
		AvgPacketSize:        sql.NullFloat64{Float64: res.AvgPacketSize, Valid: true},
		TlsVersions:          sql.NullString{String: string(tlsVersions), Valid: true},
		DnsQueries:           sql.NullInt64{Int64: int64(res.DNSQueries), Valid: true},
		DurationSeconds:      sql.NullInt64{Int64: res.DurationSeconds, Valid: true},
		FirstPacketTime:      sql.NullTime{Time: res.FirstPacketTime, Valid: true},
		LastPacketTime:       sql.NullTime{Time: res.LastPacketTime, Valid: true},
	}

	_, insertErr := s.InsertCaptureWithStats(context.Background(), caputureParams, statParams)
	if insertErr != nil {
		logger.Println("ERROR: Failed to insert capture stats:", insertErr)
		return
	}
}
