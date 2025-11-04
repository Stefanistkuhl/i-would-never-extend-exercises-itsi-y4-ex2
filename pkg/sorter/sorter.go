package sorter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/capture"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"
)

type FilenameValidationResult struct {
	IsValid         bool
	Hostname        string
	Scenario        string
	CaptureDateTime time.Time
	Error           string
}

func StartSorter() {
	lg, err := logger.New("[pcap-sorter]", "./logs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	cfg, cfgErr := config.LoadAndCheckConfig(lg)
	if cfgErr != nil {
		lg.Fatal("Failed to load config", "error", cfgErr)
	}
	if err := InitSorter(cfg); err != nil {
		lg.Fatal("Failed to initialize sorter", "error", err)
	}
	am := NewArchiveManager(cfg, lg)

	if err := am.InitialCheck(); err != nil {
		lg.Fatal("Failed to initially check for pending archive tasks", "error", err)
	}

	go startHTTPServer(lg, cfg)
	go am.StartPeriodicCheck()
	Watcher(cfg, lg)
}

func ValidateFilename(filePath string, cfg config.Config, logger logger.Logger) FilenameValidationResult {
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
			logger.Error("Failed to rename file", "from", filePath, "to", newPath, "error", err)
			return result
		}

		result.Error = "Invalid filename format"
		if cfg.LogLevel == "info" {
			logger.Info("Renamed invalid filename", "from", filePath, "to", newPath)
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
			logger.Info("Renamed invalid datetime", "from", filePath, "to", newPath)
		}
		return result
	}

	result.IsValid = true
	result.CaptureDateTime = captureDateTime
	if cfg.LogLevel == "info" {
		logger.Info("Valid filename", "hostname", result.Hostname, "scenario", result.Scenario, "datetime", result.CaptureDateTime.Format(time.RFC3339))
	}

	return result
}

func processFile(cfg config.Config, path string, logger logger.Logger) {
	if cfg.LogLevel == "info" {
		logger.Info("Processing file", "path", path)
	}
	result := ValidateFilename(path, cfg, logger)
	if !result.IsValid {
		if cfg.LogLevel == "info" {
			logger.Warn("Filename is invalid", "error", result.Error)
		}
		return
	}
	organizedPath := filepath.Join(cfg.OrganizedDir, result.Hostname, result.CaptureDateTime.UTC().Format(time.RFC3339))
	if _, err := os.Stat(organizedPath); os.IsNotExist(err) {
		if err := os.MkdirAll(organizedPath, os.ModePerm); err != nil {
			logger.Error("Failed to create organized directory", "path", organizedPath, "error", err)
			return
		}
	}
	organizedFilePath := filepath.Join(organizedPath, fmt.Sprintf("%s-%s-%s.pcap", result.Hostname, result.Scenario, result.CaptureDateTime.UTC().Format(time.RFC3339)))
	renameErr := os.Rename(path, organizedFilePath)
	if renameErr != nil {
		logger.Error("Failed to rename file", "from", path, "to", organizedFilePath, "error", renameErr)
		return
	}

	info, infoErr := os.Stat(organizedFilePath)
	if infoErr != nil {
		logger.Error("Failed to get file info", "path", organizedFilePath, "error", infoErr)
		return
	}

	s, getQeuryErr := db.InitIfNeeded()
	if getQeuryErr != nil {
		logger.Error("Failed to get database queries", "error", getQeuryErr)
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
		logger.Error("Failed to parse capture file", "path", organizedFilePath, "error", parseErr)
		return
	}

	protDist, marshallProtDistErr := json.Marshal(res.ProtocolDistribution)
	if marshallProtDistErr != nil {
		logger.Error("Failed to marshal protocol distribution", "error", marshallProtDistErr)
		return
	}
	topSrcIps, marshallTopSrcIpsErr := json.Marshal(res.TopSrcIPs)
	if marshallTopSrcIpsErr != nil {
		logger.Error("Failed to marshal top source IPs", "error", marshallTopSrcIpsErr)
		return
	}
	topDstIps, marshallTopDstIpsErr := json.Marshal(res.TopDstIPs)
	if marshallTopDstIpsErr != nil {
		logger.Error("Failed to marshal top destination IPs", "error", marshallTopDstIpsErr)
		return
	}

	topTcpSrcPorts, marshalTopTcpSrcErr := json.Marshal(res.TopTCPSrcPorts)
	if marshalTopTcpSrcErr != nil {
		logger.Error("Failed to marshal top TCP source ports", "error", marshalTopTcpSrcErr)
		return
	}
	topTcpDstPorts, marshalTopTcpDstErr := json.Marshal(res.TopTCPDstPorts)
	if marshalTopTcpDstErr != nil {
		logger.Error("Failed to marshal top TCP destination ports", "error", marshalTopTcpDstErr)
		return
	}
	topUdpSrcPorts, marshalTopUdpSrcErr := json.Marshal(res.TopUDPSrcPorts)
	if marshalTopUdpSrcErr != nil {
		logger.Error("Failed to marshal top UDP source ports", "error", marshalTopUdpSrcErr)
		return
	}
	topUdpDstPorts, marshalTopUdpDstErr := json.Marshal(res.TopUDPDstPorts)
	if marshalTopUdpDstErr != nil {
		logger.Error("Failed to marshal top UDP destination ports", "error", marshalTopUdpDstErr)
		return
	}

	// Assume you already have captureID available in scope.
	statParams := sqlc.InsertCaptureStatsParams{
		PacketCount:          sql.NullInt64{Int64: int64(res.TotalPackets), Valid: true},
		ProtocolDistribution: sql.NullString{String: string(protDist), Valid: true},
		TopSrcIps:            sql.NullString{String: string(topSrcIps), Valid: true},
		TopDstIps:            sql.NullString{String: string(topDstIps), Valid: true},
		TopTcpSrcPorts:       sql.NullString{String: string(topTcpSrcPorts), Valid: true},
		TopTcpDstPorts:       sql.NullString{String: string(topTcpDstPorts), Valid: true},
		TopUdpSrcPorts:       sql.NullString{String: string(topUdpSrcPorts), Valid: true},
		TopUdpDstPorts:       sql.NullString{String: string(topUdpDstPorts), Valid: true},
		PacketRate:           sql.NullFloat64{Float64: res.PacketRate, Valid: true},
		AvgPacketSize:        sql.NullFloat64{Float64: res.AvgPacketSize, Valid: true},
		DurationSeconds:      sql.NullInt64{Int64: res.DurationSeconds, Valid: true},
		FirstPacketTime:      sql.NullTime{Time: res.FirstPacketTime, Valid: true},
		LastPacketTime:       sql.NullTime{Time: res.LastPacketTime, Valid: true},
	}

	_, insertErr := s.InsertCaptureWithStats(context.Background(), caputureParams, statParams)
	if insertErr != nil {
		logger.Error("Failed to insert capture stats", "error", insertErr)
		return
	}

	if cfg.LogLevel == "info" {
		logger.Info("Successfully processed file", "path", organizedFilePath)
	}
}
