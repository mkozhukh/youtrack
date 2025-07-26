package tickets

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

// listComments handles the list comments command
func listComments(cmd *cobra.Command, args []string) error {
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

	log.Info("Fetching comments for ticket", "ticketID", ticketID)

	// Get comments for the ticket
	comments, err := client.GetIssueComments(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch comments", "error", err)
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Output results
	return outputResult(cmd, comments, formatCommentsList)
}

// addComment handles the add comment command
func addComment(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	// Validate ticket ID format
	if !isValidTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected format: PRJ-123)", ticketID)
	}

	// Validate required parameters
	if commentMessage == "" {
		return fmt.Errorf("comment message is required (use --message flag)")
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

	log.Info("Adding comment to ticket", "ticketID", ticketID)

	// Add the comment
	comment, err := client.AddIssueComment(ctx, ticketID, commentMessage)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to add comment", "error", err)
		return fmt.Errorf("failed to add comment: %w", err)
	}

	// Output results
	return outputResult(cmd, comment, formatCommentAdded)
}
