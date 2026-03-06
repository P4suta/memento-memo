package handler

import (
	"net/http"
	"time"

	"github.com/sakashita/memento-memo/internal/service"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) Daily(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	report, err := h.reportService.GetDaily(r.Context(), date)
	if err != nil {
		writeError(w, err)
		return
	}

	if report == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"report_date": dateStr,
			"message":     "No report available for this date",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":                   report.ID,
		"report_date":          report.ReportDate.Time.Format("2006-01-02"),
		"session_count":        report.SessionCount,
		"total_memos":          report.TotalMemos,
		"total_chars":          report.TotalChars,
		"chars_per_minute":     report.CharsPerMinute,
		"active_minutes":       report.ActiveMinutes,
		"hourly_distribution":  report.HourlyDistribution,
	})
}

func (h *ReportHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reportService.GetStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
