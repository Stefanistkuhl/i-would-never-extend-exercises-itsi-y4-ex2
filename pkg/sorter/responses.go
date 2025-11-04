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

// ============================================================================
// Configuration Types
// ============================================================================

// ============================================================================
// Statistics & Summary Types
// ============================================================================

// ============================================================================
// Query & Search Types
// ============================================================================

// ============================================================================
// Compression Types
// ============================================================================

// ============================================================================
// Cleanup Types
// ============================================================================

// ============================================================================
// Helper Types
// ============================================================================

func jsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
