package tickets

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

// listWorklogs handles the list worklogs command
func listWorklogs(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	// Validate ticket ID format
	if !isValidTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected format: PRJ-123)", ticketID)
	}

	// Load configuration
	cfg, err := config.Load("", cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	log.Info("Fetching worklogs for ticket", "ticketID", ticketID)

	// Get worklogs for the ticket
	worklogs, err := client.GetIssueWorklogs(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch worklogs", "error", err)
		return fmt.Errorf("failed to fetch worklogs: %w", err)
	}

	// Output results
	return outputResult(cmd, worklogs, formatWorklogsList)
}

// addWorklog handles the add worklog command
func addWorklog(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	// Validate ticket ID format
	if !isValidTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected format: PRJ-123)", ticketID)
	}

	// Validate required parameters
	if worklogDuration == "" {
		return fmt.Errorf("duration is required (use --duration flag)")
	}

	// Parse duration into minutes
	durationMinutes, err := parseDuration(worklogDuration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	// Load configuration
	cfg, err := config.Load("", cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Create worklog request
	req := &youtrack.CreateWorklogRequest{
		Duration:    durationMinutes,
		Description: worklogDescription,
	}

	log.Info("Adding worklog to ticket", "ticketID", ticketID, "duration", durationMinutes)

	// Add the worklog
	worklog, err := client.AddIssueWorklog(ctx, ticketID, req)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to add worklog", "error", err)
		return fmt.Errorf("failed to add worklog: %w", err)
	}

	// Output results
	return outputResult(cmd, worklog, formatWorklogAdded)
}
