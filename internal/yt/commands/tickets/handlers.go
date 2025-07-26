package tickets

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

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
	if userID == "" {
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
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch ticket", "error", err)
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	// Output results
	return outputResult(cmd, ticket, formatTicketDetails)
}

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
	}

	// Validate required parameters
	if projectID == "" {
		return fmt.Errorf("project ID is required (use --project flag or set default in config)")
	}

	if createTitle == "" {
		return fmt.Errorf("title is required (use --title flag)")
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Parse custom fields
	customFields, err := parseCustomFields(createFields)
	if err != nil {
		return fmt.Errorf("failed to parse custom fields: %w", err)
	}

	// Handle assignee if specified
	if createAssignee != "" {
		// Try to resolve assignee by username/email
		assigneeUser, err := client.SuggestUserByProject(ctx, projectID, createAssignee)
		if err != nil {
			log.Error("Failed to find assignee", "assignee", createAssignee, "error", err)
			return fmt.Errorf("failed to find assignee '%s' in project '%s': %w", createAssignee, projectID, err)
		}

		// Add assignee to custom fields as YouTrack expects it
		if customFields == nil {
			customFields = make(map[string]interface{})
		}
		customFields["assignee"] = map[string]interface{}{
			"id": assigneeUser.ID,
		}
	}

	// Create the issue request
	req := &youtrack.CreateIssueRequest{
		Project: youtrack.ProjectRef{
			ID: projectID,
		},
		Summary:     createTitle,
		Description: createDescription,
		Fields:      customFields,
	}

	log.Info("Creating ticket", "project", projectID, "title", createTitle)

	// Create the ticket
	ticket, err := client.CreateIssue(ctx, req)
	if err != nil {
		log.Error("Failed to create ticket", "error", err)
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	// Output results
	return outputResult(cmd, ticket, formatTicketCreated)
}

func updateTicket(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	// Validate ticket ID format
	if !isValidTicketID(ticketID) {
		return fmt.Errorf("invalid ticket ID format: %s (expected format: PRJ-123)", ticketID)
	}

	// Check if at least one update flag is provided
	if updateStatus == "" && updateAssignee == "" && len(updateFields) == 0 {
		return fmt.Errorf("at least one update flag must be specified (--status, --assignee, or --field)")
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

	// First, get the ticket to determine its project for assignee resolution
	originalTicket, err := client.GetIssue(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch ticket", "error", err)
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	// Extract project ID from ticket ID (e.g., "PRJ-123" -> "PRJ")
	projectID := extractProjectFromTicketID(ticketID)

	// Build update request
	req := &youtrack.UpdateIssueRequest{}

	// Parse custom fields
	customFields, err := parseCustomFields(updateFields)
	if err != nil {
		return fmt.Errorf("failed to parse custom fields: %w", err)
	}

	// Handle status update
	if updateStatus != "" {
		if customFields == nil {
			customFields = make(map[string]interface{})
		}
		customFields["State"] = updateStatus
	}

	// Handle assignee update
	if updateAssignee != "" {
		// Try to resolve assignee by username/email in the project
		assigneeUser, err := client.SuggestUserByProject(ctx, projectID, updateAssignee)
		if err != nil {
			log.Error("Failed to find assignee", "assignee", updateAssignee, "error", err)
			return fmt.Errorf("failed to find assignee '%s' in project '%s': %w", updateAssignee, projectID, err)
		}

		req.Assignee = &assigneeUser.ID
	}

	req.Fields = customFields

	log.Info("Updating ticket", "ticketID", ticketID)

	// Update the ticket
	updatedTicket, err := client.UpdateIssue(ctx, ticketID, req)
	if err != nil {
		log.Error("Failed to update ticket", "error", err)
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// Create a summary of changes for confirmation
	changesSummary := &UpdateSummary{
		TicketID:        ticketID,
		OriginalTicket:  originalTicket,
		UpdatedTicket:   updatedTicket,
		StatusChanged:   updateStatus,
		AssigneeChanged: updateAssignee,
		FieldsChanged:   updateFields,
	}

	// Output results
	return outputResult(cmd, changesSummary, formatTicketUpdated)
}

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

	// First, verify the ticket exists
	_, err = client.GetIssue(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch ticket", "error", err)
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	// Track results for each tag
	var results []TagOperationResult
	var hasErrors bool

	// Add each tag
	for _, tagName := range tagNames {
		result := TagOperationResult{
			TagName: tagName,
			Success: true,
		}

		// Try to add the tag
		err := client.AddIssueTag(ctx, ticketID, tagName)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			hasErrors = true
			log.Warn("Failed to add tag", "tag", tagName, "error", err)
		}

		results = append(results, result)
	}

	// Create summary for output
	summary := &TagOperationSummary{
		TicketID:  ticketID,
		Operation: "add",
		Results:   results,
		HasErrors: hasErrors,
	}

	// Output results
	return outputResult(cmd, summary, formatTagOperationSummary)
}

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

	// First, verify the ticket exists
	_, err = client.GetIssue(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch ticket", "error", err)
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	// Track results for each tag
	var results []TagOperationResult
	var hasErrors bool

	// Remove each tag
	for _, tagName := range tagNames {
		result := TagOperationResult{
			TagName: tagName,
			Success: true,
		}

		// Try to remove the tag
		err := client.RemoveIssueTag(ctx, ticketID, tagName)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			// For untag operation, non-existent tags are handled gracefully
			// We only log as warning, not count as error
			log.Warn("Failed to remove tag", "tag", tagName, "error", err)

			// Check if it's a "not found" type error, which we handle gracefully
			if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
				result.Error = "tag not found on ticket (already removed or never existed)"
			} else {
				hasErrors = true
			}
		}

		results = append(results, result)
	}

	// Create summary for output
	summary := &TagOperationSummary{
		TicketID:  ticketID,
		Operation: "remove",
		Results:   results,
		HasErrors: hasErrors,
	}

	// Output results
	return outputResult(cmd, summary, formatTagOperationSummary)
}

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

	log.Info("Fetching activity history for ticket", "ticketID", ticketID)

	// Get activities for the ticket
	activities, err := client.GetIssueActivities(ctx, ticketID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("ticket not found: %s", ticketID)
		}
		log.Error("Failed to fetch activity history", "error", err)
		return fmt.Errorf("failed to fetch activity history: %w", err)
	}

	// Create history summary for output
	summary := &HistorySummary{
		TicketID:   ticketID,
		Activities: activities,
	}

	// Output results
	return outputResult(cmd, summary, formatHistorySummary)
}

// buildSearchQuery constructs a YouTrack search query based on the provided filters
func buildSearchQuery(projectID, userID, customQuery string) string {
	var parts []string

	// Add project filter if specified
	if projectID != "" {
		parts = append(parts, fmt.Sprintf("project: %s", projectID))
	}

	// Add user filter if specified
	if userID != "" {
		parts = append(parts, fmt.Sprintf("assignee: %s", userID))
	}

	// Add custom query if specified
	if customQuery != "" {
		parts = append(parts, customQuery)
	}

	// If no filters are specified, return a query that shows all tickets
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// extractProjectFromTicketID extracts the project part from a ticket ID (e.g., "PRJ-123" -> "PRJ")
func extractProjectFromTicketID(ticketID string) string {
	parts := strings.Split(ticketID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
