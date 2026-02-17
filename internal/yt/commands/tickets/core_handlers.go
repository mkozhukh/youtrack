package tickets

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/mkozhukh/youtrack/internal/yt/config"
	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// listTickets handles the list tickets command
func listTickets(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load("", cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Use default project if not specified
	if projectID == "" {
		projectID = cfg.Defaults.Project
	}

	// Use default user if not specified
	if userID == ":me" {
		userID = cfg.Defaults.UserID
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Build the search query
	searchQuery := buildSearchQuery(projectID, userID, query)

	log.Info("Searching tickets", "query", searchQuery, "limit", limit)

	// Search for tickets
	tickets, err := client.SearchIssues(ctx, searchQuery, 0, limit)
	if err != nil {
		log.Error("Failed to search tickets", "error", err)
		return fmt.Errorf("failed to search tickets: %w", err)
	}

	// Output results
	return outputResult(cmd, tickets, formatTicketsList)
}

// showTicket handles the show ticket command
func showTicket(cmd *cobra.Command, args []string) error {
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

	log.Info("Fetching ticket details", "ticketID", ticketID)

	// Get ticket details
	ticket, err := client.GetIssue(ctx, ticketID)
	if err != nil {
		log.Error("Failed to get ticket", "ticketID", ticketID, "error", err)
		return fmt.Errorf("failed to get ticket %s: %w", ticketID, err)
	}

	// Output results
	return outputResult(cmd, ticket, formatTicketDetails)
}

// createTicket handles the create ticket command
func createTicket(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load("", cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Use default project if not specified
	if projectID == "" {
		projectID = cfg.Defaults.Project
		if projectID == "" {
			return fmt.Errorf("project ID is required (use --project flag or set default in config)")
		}
	}

	// Parse custom fields
	customFields, err := parseCustomFields(createFields)
	if err != nil {
		return fmt.Errorf("failed to parse custom fields: %w", err)
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Build create request
	req := &youtrack.CreateIssueRequest{
		Project:     youtrack.ProjectRef{ID: projectID},
		Summary:     createTitle,
		Description: createDescription,
	}

	// Set custom fields if any
	if customFields != nil {
		req.Fields = customFields
	}

	log.Info("Creating ticket", "project", projectID, "title", createTitle)

	// Create the ticket
	ticket, err := client.CreateIssue(ctx, req)
	if err != nil {
		log.Error("Failed to create ticket", "error", err)
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	log.Info("Ticket created successfully", "ticketID", ticket.ID)

	// Output results
	return outputResult(cmd, ticket, formatTicketCreated)
}

// updateTicket handles the update ticket command
func updateTicket(cmd *cobra.Command, args []string) error {
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

	// Check if at least one update field is provided
	if len(updateFields) == 0 {
		return fmt.Errorf("at least one update field must be specified (--field)")
	}

	// Parse custom fields
	customFields, err := parseCustomFields(updateFields)
	if err != nil {
		return fmt.Errorf("failed to parse custom fields: %w", err)
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Get original ticket for comparison
	originalTicket, err := client.GetIssue(ctx, ticketID)
	if err != nil {
		log.Error("Failed to get original ticket", "ticketID", ticketID, "error", err)
		return fmt.Errorf("failed to get original ticket %s: %w", ticketID, err)
	}

	// Build update request
	req := &youtrack.UpdateIssueRequest{}

	// Set custom fields if any
	if customFields != nil {
		req.Fields = customFields
	}

	log.Info("Updating ticket", "ticketID", ticketID)

	// Update the ticket
	updatedTicket, err := client.UpdateIssue(ctx, ticketID, req)
	if err != nil {
		log.Error("Failed to update ticket", "ticketID", ticketID, "error", err)
		return fmt.Errorf("failed to update ticket %s: %w", ticketID, err)
	}

	log.Info("Ticket updated successfully", "ticketID", ticketID)

	// Create update summary
	summary := &UpdateSummary{
		TicketID:       ticketID,
		OriginalTicket: originalTicket,
		UpdatedTicket:  updatedTicket,
	}

	// Track what was changed
	if len(updateFields) > 0 {
		summary.FieldsChanged = append(summary.FieldsChanged, updateFields...)
	}

	// Output results
	return outputResult(cmd, summary, formatTicketUpdated)
}

// tagTicket handles the tag ticket command
func tagTicket(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	tagNames := args[1:]

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

	log.Info("Adding tags to ticket", "ticketID", ticketID, "tags", tagNames)

	// Track operation results
	summary := &TagOperationSummary{
		TicketID:  ticketID,
		Operation: "add",
		Results:   make([]TagOperationResult, len(tagNames)),
	}

	// Add each tag
	for i, tagName := range tagNames {
		tagID, err := client.EnsureTag(ctx, tagName, "")
		if err != nil {
			summary.Results[i] = TagOperationResult{
				TagName: tagName,
				Success: false,
				Error:   err.Error(),
			}
			summary.HasErrors = true
			log.Error("Failed to ensure tag", "ticketID", ticketID, "tag", tagName, "error", err)
			continue
		}
		err = client.AddIssueTag(ctx, ticketID, tagID)
		result := TagOperationResult{
			TagName: tagName,
			Success: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
			summary.HasErrors = true
			log.Error("Failed to add tag", "ticketID", ticketID, "tag", tagName, "error", err)
		} else {
			log.Info("Tag added successfully", "ticketID", ticketID, "tag", tagName)
		}
		summary.Results[i] = result
	}

	// Output results
	return outputResult(cmd, summary, formatTagOperationSummary)
}

// untagTicket handles the untag ticket command
func untagTicket(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	tagNames := args[1:]

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

	log.Info("Removing tags from ticket", "ticketID", ticketID, "tags", tagNames)

	// Track operation results
	summary := &TagOperationSummary{
		TicketID:  ticketID,
		Operation: "remove",
		Results:   make([]TagOperationResult, len(tagNames)),
	}

	// Remove each tag
	for i, tagName := range tagNames {
		tag, err := client.GetTagByName(ctx, tagName)
		if err != nil {
			summary.Results[i] = TagOperationResult{
				TagName: tagName,
				Success: false,
				Error:   err.Error(),
			}
			summary.HasErrors = true
			log.Error("Failed to find tag", "ticketID", ticketID, "tag", tagName, "error", err)
			continue
		}
		err = client.RemoveIssueTag(ctx, ticketID, tag.ID)
		result := TagOperationResult{
			TagName: tagName,
			Success: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
			summary.HasErrors = true
			log.Error("Failed to remove tag", "ticketID", ticketID, "tag", tagName, "error", err)
		} else {
			log.Info("Tag removed successfully", "ticketID", ticketID, "tag", tagName)
		}
		summary.Results[i] = result
	}

	// Output results
	return outputResult(cmd, summary, formatTagOperationSummary)
}

// showHistory handles the history command
func showHistory(cmd *cobra.Command, args []string) error {
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

	log.Info("Fetching ticket history", "ticketID", ticketID)

	// Get ticket activities/history
	activities, err := client.GetIssueActivities(ctx, ticketID)
	if err != nil {
		log.Error("Failed to get ticket activities", "ticketID", ticketID, "error", err)
		return fmt.Errorf("failed to get ticket activities for %s: %w", ticketID, err)
	}

	// Create history summary
	summary := &HistorySummary{
		TicketID:   ticketID,
		Activities: activities,
	}

	// Output results
	return outputResult(cmd, summary, formatHistorySummary)
}

// buildSearchQuery builds a YouTrack search query from the provided parameters
func buildSearchQuery(projectID, userID, customQuery string) string {
	var parts []string

	// Add project filter if specified
	if projectID != "" {
		parts = append(parts, fmt.Sprintf("project: %s", projectID))
	}

	// Add assignee filter if specified
	if userID != "" {
		parts = append(parts, fmt.Sprintf("assignee: %s", userID))
	}

	// Add custom query if specified
	if customQuery != "" {
		parts = append(parts, customQuery)
	}

	// If no filters specified, return a default query
	if len(parts) == 0 {
		return "#Unresolved"
	}

	return strings.Join(parts, " ")
}

// extractProjectFromTicketID extracts the project part from a ticket ID (e.g., "PRJ" from "PRJ-123")
func extractProjectFromTicketID(ticketID string) string {
	parts := strings.Split(ticketID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
