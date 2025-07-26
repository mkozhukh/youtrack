package tickets

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

// listAttachments handles the list attachments command
func listAttachments(cmd *cobra.Command, args []string) error {
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

	log.Info("Fetching attachments for ticket", "ticketID", ticketID)

	// Get attachments for the ticket
	attachments, err := client.GetIssueAttachments(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch attachments", "error", err)
		return fmt.Errorf("failed to fetch attachments: %w", err)
	}

	// Output results
	return outputResult(cmd, attachments, formatAttachmentsList)
}

// addAttachment handles the add attachment command
func addAttachment(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	filePath := args[1]

	// Validate ticket ID format
	if !isValidTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected format: PRJ-123)", ticketID)
	}

	// Validate file existence
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	} else if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
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

	log.Info("Adding attachment to ticket", "ticketID", ticketID, "filePath", filePath)

	// Add the attachment
	attachment, err := client.AddIssueAttachment(ctx, ticketID, filePath)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to add attachment", "error", err)
		return fmt.Errorf("failed to add attachment: %w", err)
	}

	// Output results
	return outputResult(cmd, attachment, formatAttachmentAdded)
}
