package sorter

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
)

const defaultTopLimit = 10

func (s *Server) GetSummaryHandler(w http.ResponseWriter, r *http.Request) {
	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	summary, err := store.GetSummary(context.Background())
	if err != nil {
		s.logger.Error("Failed to get summary", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	result := SummaryRes{
		TotalCaptures:        summary.TotalCaptures,
		CapturesWithStats:    summary.CapturesWithStats,
		TotalFileSize:        int64(getFloatValue(summary.TotalFileSize)),
		TotalPackets:         int64(getFloatValue(summary.TotalPackets)),
		AvgPacketRate:        getFloatValue(summary.AvgPacketRate),
		AvgPacketSize:        getFloatValue(summary.AvgPacketSize),
		TotalDurationSeconds: int64(getFloatValue(summary.TotalDurationSeconds)),
		UniqueHostnames:      summary.UniqueHostnames,
		UniqueScenarios:      summary.UniqueScenarios,
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) GetFileStatsHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	captureID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		s.logger.Error("Invalid capture ID", "error", err, "id", idParam)
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	statsRow, err := store.GetCaptureStatsByID(context.Background(), captureID)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonResponse(w, http.StatusNotFound, StatusRes{Status: "error"})
		} else {
			s.logger.Error("Failed to get capture stats", "error", err, "id", captureID)
			jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		}
		return
	}

	if statsRow.Hostname == "" {
		jsonResponse(w, http.StatusNotFound, StatusRes{Status: "error"})
		return
	}

	result := FileStatsRes{
		CaptureID:       captureID,
		Hostname:        statsRow.Hostname,
		Scenario:        statsRow.Scenario,
		CaptureDatetime: statsRow.CaptureDatetime.Format(time.RFC3339),
		FilePath:        statsRow.FilePath,
	}

	if statsRow.PacketCount.Valid {
		result.PacketCount = &statsRow.PacketCount.Int64
	}
	if statsRow.PacketRate.Valid {
		result.PacketRate = &statsRow.PacketRate.Float64
	}
	if statsRow.AvgPacketSize.Valid {
		result.AvgPacketSize = &statsRow.AvgPacketSize.Float64
	}
	if statsRow.DurationSeconds.Valid {
		result.DurationSeconds = &statsRow.DurationSeconds.Int64
	}
	if statsRow.FirstPacketTime.Valid {
		ft := statsRow.FirstPacketTime.Time.Format(time.RFC3339)
		result.FirstPacketTime = &ft
	}
	if statsRow.LastPacketTime.Valid {
		lt := statsRow.LastPacketTime.Time.Format(time.RFC3339)
		result.LastPacketTime = &lt
	}

	if statsRow.ProtocolDistribution.Valid && statsRow.ProtocolDistribution.String != "" {
		var protocolDist map[string]int
		if err := json.Unmarshal([]byte(statsRow.ProtocolDistribution.String), &protocolDist); err == nil {
			result.ProtocolDistribution = protocolDist
		}
	}

	if statsRow.TopSrcIps.Valid && statsRow.TopSrcIps.String != "" {
		var topSrcIps map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopSrcIps.String), &topSrcIps); err == nil {
			result.TopSrcIps = topSrcIps
		}
	}

	if statsRow.TopDstIps.Valid && statsRow.TopDstIps.String != "" {
		var topDstIps map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopDstIps.String), &topDstIps); err == nil {
			result.TopDstIps = topDstIps
		}
	}

	if statsRow.TopTcpSrcPorts.Valid && statsRow.TopTcpSrcPorts.String != "" {
		var topTcpSrcPorts map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopTcpSrcPorts.String), &topTcpSrcPorts); err == nil {
			result.TopTcpSrcPorts = topTcpSrcPorts
		}
	}

	if statsRow.TopTcpDstPorts.Valid && statsRow.TopTcpDstPorts.String != "" {
		var topTcpDstPorts map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopTcpDstPorts.String), &topTcpDstPorts); err == nil {
			result.TopTcpDstPorts = topTcpDstPorts
		}
	}

	if statsRow.TopUdpSrcPorts.Valid && statsRow.TopUdpSrcPorts.String != "" {
		var topUdpSrcPorts map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopUdpSrcPorts.String), &topUdpSrcPorts); err == nil {
			result.TopUdpSrcPorts = topUdpSrcPorts
		}
	}

	if statsRow.TopUdpDstPorts.Valid && statsRow.TopUdpDstPorts.String != "" {
		var topUdpDstPorts map[string]int
		if err := json.Unmarshal([]byte(statsRow.TopUdpDstPorts.String), &topUdpDstPorts); err == nil {
			result.TopUdpDstPorts = topUdpDstPorts
		}
	}

	topLimit := resolveTopLimit(r)
	result.TopSrcIps = limitStatMap(result.TopSrcIps, topLimit)
	result.TopDstIps = limitStatMap(result.TopDstIps, topLimit)
	result.TopTcpSrcPorts = limitStatMap(result.TopTcpSrcPorts, topLimit)
	result.TopTcpDstPorts = limitStatMap(result.TopTcpDstPorts, topLimit)
	result.TopUdpSrcPorts = limitStatMap(result.TopUdpSrcPorts, topLimit)
	result.TopUdpDstPorts = limitStatMap(result.TopUdpDstPorts, topLimit)

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) GetStatsByHostnameHandler(w http.ResponseWriter, r *http.Request) {
	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	rows, err := store.GetStatsByHostname(context.Background())
	if err != nil {
		s.logger.Error("Failed to get stats by hostname", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	results := make([]StatsByHostnameRes, 0, len(rows))
	for _, row := range rows {
		results = append(results, StatsByHostnameRes{
			Hostname:             row.Hostname,
			CaptureCount:         row.CaptureCount,
			TotalFileSize:        int64(getFloatValue(row.TotalFileSize)),
			TotalPackets:         int64(getFloatValue(row.TotalPackets)),
			AvgPacketRate:        getFloatValue(row.AvgPacketRate),
			AvgPacketSize:        getFloatValue(row.AvgPacketSize),
			TotalDurationSeconds: int64(getFloatValue(row.TotalDurationSeconds)),
		})
	}

	jsonResponse(w, http.StatusOK, results)
}

func (s *Server) GetStatsByScenarioHandler(w http.ResponseWriter, r *http.Request) {
	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	rows, err := store.GetStatsByScenario(context.Background())
	if err != nil {
		s.logger.Error("Failed to get stats by scenario", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	results := make([]StatsByScenarioRes, 0, len(rows))
	for _, row := range rows {
		results = append(results, StatsByScenarioRes{
			Scenario:             row.Scenario,
			CaptureCount:         row.CaptureCount,
			TotalFileSize:        int64(getFloatValue(row.TotalFileSize)),
			TotalPackets:         int64(getFloatValue(row.TotalPackets)),
			AvgPacketRate:        getFloatValue(row.AvgPacketRate),
			AvgPacketSize:        getFloatValue(row.AvgPacketSize),
			TotalDurationSeconds: int64(getFloatValue(row.TotalDurationSeconds)),
		})
	}

	jsonResponse(w, http.StatusOK, results)
}

func getFloatValue(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0
}

func resolveTopLimit(r *http.Request) int {
	limitParam := r.URL.Query().Get("top_limit")
	if limitParam == "" {
		return defaultTopLimit
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		return defaultTopLimit
	}

	return limit
}

func limitStatMap(src map[string]int, limit int) map[string]int {
	if src == nil || limit <= 0 || len(src) <= limit {
		return src
	}

	type entry struct {
		key   string
		value int
	}

	entries := make([]entry, 0, len(src))
	for k, v := range src {
		entries = append(entries, entry{key: k, value: v})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].value == entries[j].value {
			return entries[i].key < entries[j].key
		}
		return entries[i].value > entries[j].value
	})

	trimmed := make(map[string]int, limit)
	for i := 0; i < limit && i < len(entries); i++ {
		trimmed[entries[i].key] = entries[i].value
	}

	return trimmed
}
