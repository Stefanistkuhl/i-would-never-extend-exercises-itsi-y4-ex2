package sorter

import (
	"encoding/json"
	"net/http"
)

// ============================================================================
// Health & Status Operations
// ============================================================================

type HealthRes struct {
	Status string `json:"status"`
}

type StatusRes struct {
	Status string `json:"status"`
}

type VersionRes struct {
	Version string `json:"version"`
}

type HWStatusRes struct {
	CpuPercent float64       `json:"cpu"`
	Memory     MinMaxPercent `json:"memory"`
	Disks      []DiskStatus  `json:"disks"`
}

type DiskStatus struct {
	Path        string        `json:"path"`
	SourceDirs  []string      `json:"source_dirs"`
	ConfigPaths []string      `json:"config_paths"`
	Usage       MinMaxPercent `json:"usage"`
}

type MinMaxPercent struct {
	Used    uint64  `json:"used"`
	Max     uint64  `json:"max"`
	Percent float64 `json:"percent"`
}

// ============================================================================
// File Types
// ============================================================================

// ============================================================================
// Archiving Types
// ============================================================================

type ArchiveStatus struct {
	Files    []ArchivedFile `json:"files"`
	NumFiles int            `json:"num_files"`
	Size     int64          `json:"size"`
}

type ArchivedFile struct {
	ID        int64  `json:"id"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ============================================================================
// Configuration Types
// ============================================================================

// ============================================================================
// Statistics & Summary Types
// ============================================================================

type SummaryRes struct {
	TotalCaptures        int64   `json:"total_captures"`
	CapturesWithStats    int64   `json:"captures_with_stats"`
	TotalFileSize        int64   `json:"total_file_size"`
	TotalPackets         int64   `json:"total_packets"`
	AvgPacketRate        float64 `json:"avg_packet_rate"`
	AvgPacketSize        float64 `json:"avg_packet_size"`
	TotalDurationSeconds int64   `json:"total_duration_seconds"`
	UniqueHostnames      int64   `json:"unique_hostnames"`
	UniqueScenarios      int64   `json:"unique_scenarios"`
}

type FileStatsRes struct {
	CaptureID            int64          `json:"capture_id"`
	Hostname             string         `json:"hostname"`
	Scenario             string         `json:"scenario"`
	CaptureDatetime      string         `json:"capture_datetime"`
	FilePath             string         `json:"file_path"`
	PacketCount          *int64         `json:"packet_count,omitempty"`
	ProtocolDistribution map[string]int `json:"protocol_distribution,omitempty"`
	TopSrcIps            []string       `json:"top_src_ips,omitempty"`
	TopDstIps            []string       `json:"top_dst_ips,omitempty"`
	TopTcpSrcPorts       map[string]int `json:"top_tcp_src_ports,omitempty"`
	TopTcpDstPorts       map[string]int `json:"top_tcp_dst_ports,omitempty"`
	TopUdpSrcPorts       map[string]int `json:"top_udp_src_ports,omitempty"`
	TopUdpDstPorts       map[string]int `json:"top_udp_dst_ports,omitempty"`
	PacketRate           *float64       `json:"packet_rate,omitempty"`
	AvgPacketSize        *float64       `json:"avg_packet_size,omitempty"`
	DurationSeconds      *int64         `json:"duration_seconds,omitempty"`
	FirstPacketTime      *string        `json:"first_packet_time,omitempty"`
	LastPacketTime       *string        `json:"last_packet_time,omitempty"`
}

type StatsByHostnameRes struct {
	Hostname             string  `json:"hostname"`
	CaptureCount         int64   `json:"capture_count"`
	TotalFileSize        int64   `json:"total_file_size"`
	TotalPackets         int64   `json:"total_packets"`
	AvgPacketRate        float64 `json:"avg_packet_rate"`
	AvgPacketSize        float64 `json:"avg_packet_size"`
	TotalDurationSeconds int64   `json:"total_duration_seconds"`
}

type StatsByScenarioRes struct {
	Scenario             string  `json:"scenario"`
	CaptureCount         int64   `json:"capture_count"`
	TotalFileSize        int64   `json:"total_file_size"`
	TotalPackets         int64   `json:"total_packets"`
	AvgPacketRate        float64 `json:"avg_packet_rate"`
	AvgPacketSize        float64 `json:"avg_packet_size"`
	TotalDurationSeconds int64   `json:"total_duration_seconds"`
}

// ============================================================================
// Query & Search Types
// ============================================================================

type SearchRes struct {
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
}

type SearchResult struct {
	ID              int64  `json:"id"`
	Hostname        string `json:"hostname"`
	Scenario        string `json:"scenario"`
	CaptureDatetime string `json:"capture_datetime"`
	FilePath        string `json:"file_path"`
	FileSize        int64  `json:"file_size"`
	Compressed      bool   `json:"compressed"`
	Archived        bool   `json:"archived"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

type SQLQueryReq struct {
	Query string `json:"query"`
}

type SQLQueryRes struct {
	Success bool             `json:"success"`
	Columns []string         `json:"columns,omitempty"`
	Rows    []map[string]any `json:"rows,omitempty"`
	Count   int              `json:"count,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// ============================================================================
// Compression Types
// ============================================================================

type CompressionTriggerRes struct {
	Processed int      `json:"processed"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// ============================================================================
// Cleanup Types
// ============================================================================

type CleanupCandidatesRes struct {
	OldArchivedFiles []OldArchivedFile `json:"old_archived_files"`
	EmptyDirectories []string          `json:"empty_directories"`
	UntrackedFiles   []UntrackedFile   `json:"untracked_files"`
	Summary          CleanupSummary    `json:"summary"`
}

type OldArchivedFile struct {
	ID       int64  `json:"id"`
	FilePath string `json:"file_path"`
}

type UntrackedFile struct {
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
}

type CleanupSummary struct {
	OldArchivedFilesCount int   `json:"old_archived_files_count"`
	EmptyDirectoriesCount int   `json:"empty_directories_count"`
	UntrackedFilesCount   int   `json:"untracked_files_count"`
	TotalSizeToFree       int64 `json:"total_size_to_free"`
}

type CleanupExecuteRes struct {
	DeletedOldArchivedFiles int            `json:"deleted_old_archived_files"`
	DeletedEmptyDirectories int            `json:"deleted_empty_directories"`
	DeletedUntrackedFiles   int            `json:"deleted_untracked_files"`
	Errors                  []string       `json:"errors,omitempty"`
	Summary                 CleanupSummary `json:"summary"`
}

type CleanupEmptyRes struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ============================================================================
// Helper Types
// ============================================================================

func jsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
