package tickets

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

// addLink handles the add link command
func addLink(cmd *cobra.Command, args []string) error {
	sourceTicketID := args[0]
	targetTicketID := args[1]

	// Validate ticket ID formats
	if !isValidTicketID(sourceTicketID) {
		return fmt.Errorf("invalid source ticket ID format: %s (expected format: PRJ-123)", sourceTicketID)
	}
	if !isValidTicketID(targetTicketID) {
		return fmt.Errorf("invalid target ticket ID format: %s (expected format: PRJ-123)", targetTicketID)
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

	log.Info("Creating link between tickets", "source", sourceTicketID, "target", targetTicketID, "type", linkType)

	// First, verify both tickets exist
	_, err = client.GetIssue(ctx, sourceTicketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("source ticket not found: %s", sourceTicketID)
		}
		log.Error("Failed to fetch source ticket", "error", err)
		return fmt.Errorf("failed to fetch source ticket: %w", err)
	}

	_, err = client.GetIssue(ctx, targetTicketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("target ticket not found: %s", targetTicketID)
		}
		log.Error("Failed to fetch target ticket", "error", err)
		return fmt.Errorf("failed to fetch target ticket: %w", err)
	}

	// Create the link
	err = client.CreateIssueLink(ctx, sourceTicketID, targetTicketID, linkType)
	if err != nil {
		log.Error("Failed to create link", "error", err)
		return fmt.Errorf("failed to create link: %w", err)
	}

	// Create summary for output
	summary := &LinkOperationSummary{
		SourceTicketID: sourceTicketID,
		TargetTicketID: targetTicketID,
		LinkType:       linkType,
		Success:        true,
	}

	// Output results
	return outputResult(cmd, summary, formatLinkOperationSummary)
}
