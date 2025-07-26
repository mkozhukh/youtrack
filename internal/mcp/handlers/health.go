package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mkozhukh/youtrack/pkg/youtrack"
)

// HealthHandlers manages health check operations
type HealthHandlers struct {
	ytClient     HealthClient
	toolLogger   func(string, map[string]interface{})
	errorHandler *ErrorHandler
	startTime    time.Time
}

// HealthClient defines the interface for YouTrack client operations needed for health checks
type HealthClient interface {
	ListProjects(ctx context.Context, skip, top int) ([]*youtrack.Project, error)
}

// NewHealthHandlers creates a new instance of HealthHandlers
func NewHealthHandlers(ytClient HealthClient, toolLogger func(string, map[string]interface{}), startTime time.Time) *HealthHandlers {
	return &HealthHandlers{
		ytClient:     ytClient,
		toolLogger:   toolLogger,
		errorHandler: NewErrorHandler(),
		startTime:    startTime,
	}
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Overall        string
	ServerUptime   time.Duration
	YouTrackStatus string
	YouTrackError  error
	Projects       []string
	Timestamp      time.Time
}

// performHealthChecks runs various health checks
func (h *HealthHandlers) performHealthChecks(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		ServerUptime: time.Since(h.startTime),
		Timestamp:    time.Now(),
	}

	// Test YouTrack connectivity
	projects, err := h.ytClient.ListProjects(ctx, 0, 100)
	if err != nil {
		status.YouTrackStatus = "UNHEALTHY"
		status.YouTrackError = err
		status.Overall = "UNHEALTHY"
	} else {
		status.YouTrackStatus = "HEALTHY"
		for _, project := range projects {
			status.Projects = append(status.Projects, fmt.Sprintf("[%s] %s", project.ID, project.ShortName))
		}
		status.Overall = "HEALTHY"
	}

	return status
}

// formatHealthReportPlainText formats the health check report as plain text (no emojis)
func (h *HealthHandlers) formatHealthReportPlainText(status *HealthStatus) string {
	// Create header
	header := fmt.Sprintf("YouTrack MCP Server Health Report\n")
	header += fmt.Sprintf("Health check completed at: %s\n", h.getCurrentTimestamp())
	header += strings.Repeat("-", 80) + "\n\n"

	// Server status
	response := header
	response += fmt.Sprintf("Server Status: %s\n", status.Overall)
	response += fmt.Sprintf("Server Uptime: %s\n", h.formatDuration(status.ServerUptime))
	response += fmt.Sprintf("Current Time: %s\n\n", status.Timestamp.Format("2006-01-02 15:04:05 MST"))

	// YouTrack status
	response += fmt.Sprintf("YouTrack Status: %s\n", status.YouTrackStatus)

	if status.YouTrackError != nil {
		response += fmt.Sprintf("YouTrack Error: %s\n", status.YouTrackError.Error())
	} else {
		response += fmt.Sprintf("Projects Accessible: %s\n", strings.Join(status.Projects, ", "))
	}

	// Add footer
	footer := fmt.Sprintf("\n" + strings.Repeat("-", 80) + "\n")
	footer += fmt.Sprintf("Overall Status: %s | Checked at: %s\n", status.Overall, h.getCurrentTimestamp())

	return response + footer
}

// HealthCheckHTTPHandler handles HTTP health check requests
func (h *HealthHandlers) HealthCheckHTTPHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Perform health checks
	healthStatus := h.performHealthChecks(ctx)

	// Set content type to plain text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Set status code based on health
	if healthStatus.Overall != "HEALTHY" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Write the plain text response
	response := h.formatHealthReportPlainText(healthStatus)
	w.Write([]byte(response))
}

// formatDuration formats a duration in a human-readable format
func (h *HealthHandlers) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - hours*60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) - days*24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

// getCurrentTimestamp returns the current timestamp in a readable format
func (h *HealthHandlers) getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}
