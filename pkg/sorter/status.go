package sorter

import (
	"net/http"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/utils"
)

// Health & Status Operations

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if s.cfg.LogLevel == "info" {
		s.logger.Info("Health check requested")
	}
	jsonResponse(w, http.StatusOK, HealthRes{Status: "healthy"})
}

func (s *Server) StatusHandler(w http.ResponseWriter, r *http.Request) {
	if s.cfg.LogLevel == "info" {
		s.logger.Info("Status check requested")
	}
	percentages, cpuErr := cpu.Percent(0, true)
	if cpuErr != nil {
		s.logger.Error("Failed to get CPU stats", "error", cpuErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	virtualMemory, memErr := mem.VirtualMemory()
	if memErr != nil {
		s.logger.Error("Failed to get memory stats", "error", memErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	diskInfos, pathErr := utils.GetAllDisksFromConfig(s.cfg.ArchiveDir, s.cfg.OrganizedDir, s.cfg.WatchDir)
	if pathErr != nil {
		s.logger.Error("Failed to determine disk paths from config", "error", pathErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	diskStatuses := make([]DiskStatus, 0, len(diskInfos))
	for _, diskInfo := range diskInfos {
		diskUsage, diskErr := disk.Usage(diskInfo.Path)
		if diskErr != nil {
			s.logger.Warn("Failed to get disk stats", "error", diskErr, "disk_path", diskInfo.Path, "source_dirs", diskInfo.SourceDirs)
			continue
		}

		diskStatuses = append(diskStatuses, DiskStatus{
			Path:        diskInfo.Path,
			SourceDirs:  diskInfo.SourceDirs,
			ConfigPaths: diskInfo.ConfigPaths,
			Usage: MinMaxPercent{
				Used:    diskUsage.Used,
				Max:     diskUsage.Total,
				Percent: diskUsage.UsedPercent,
			},
		})
	}

	if len(diskStatuses) == 0 {
		s.logger.Error("Failed to get disk stats for any configured disk")
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, HWStatusRes{
		CpuPercent: percentages[0],
		Memory: MinMaxPercent{
			Used:    virtualMemory.Used,
			Max:     virtualMemory.Total,
			Percent: virtualMemory.UsedPercent,
		},
		Disks: diskStatuses,
	})
}

func (s *Server) VersionHandler(w http.ResponseWriter, r *http.Request) {
	if s.cfg.LogLevel == "info" {
		s.logger.Info("Version check requested")
	}
	jsonResponse(w, http.StatusOK, VersionRes{Version: pkg.Version})
}
