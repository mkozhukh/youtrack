package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

var (
	// List command flags
	projectID string
	userID    string
	query     string
	limit     int

	// Create command flags
	createTitle       string
	createDescription string
	createAssignee    string
	createFields      []string

	// Update command flags
	updateStatus   string
	updateAssignee string
	updateFields   []string

	// Comment command flags
	commentMessage string

	// Worklog command flags
	worklogDuration    string
	worklogDescription string

	// Link command flags
	linkType string

	// Global output flag from parent
	output string
)

// TicketsCmd represents the tickets command
var TicketsCmd = &cobra.Command{
	Use:   "tickets",
	Short: "Manage tickets (issues)",
	Long:  `List, show, create, and manage YouTrack tickets.`,
	RunE:  listTickets, // Default to list when no subcommand is given
}

// listTicketsCmd represents the list command
var listTicketsCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows the latest tickets in a project",
	Long:  `Shows the latest tickets in a project with optional filtering.`,
	RunE:  listTickets,
}

// showTicketCmd represents the show command
var showTicketCmd = &cobra.Command{
	Use:   "show <ticket_id>",
	Short: "Shows detailed information for a specific ticket",
	Long:  `Shows detailed information for a specific ticket including description, assignee, tags, etc.`,
	Args:  cobra.ExactArgs(1),
	RunE:  showTicket,
}

// createTicketCmd represents the create command
var createTicketCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new ticket in a project",
	Long:  `Creates a new ticket with title, description, assignee, and custom fields.`,
	RunE:  createTicket,
}

// updateTicketCmd represents the update command
var updateTicketCmd = &cobra.Command{
	Use:   "update <ticket_id>",
	Short: "Updates fields of a specific ticket",
	Long:  `Updates status, assignee, and custom fields of a ticket. Only specified fields are updated (partial updates).`,
	Args:  cobra.ExactArgs(1),
	RunE:  updateTicket,
}

// tagTicketCmd represents the tag command
var tagTicketCmd = &cobra.Command{
	Use:   "tag <ticket_id> <tag_name...>",
	Short: "Adds one or more tags to a ticket",
	Long:  `Adds one or more tags to a ticket. If a tag doesn't exist, it will be created automatically.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  tagTicket,
}

// untagTicketCmd represents the untag command
var untagTicketCmd = &cobra.Command{
	Use:   "untag <ticket_id> <tag_name...>",
	Short: "Removes one or more tags from a ticket",
	Long:  `Removes one or more tags from a ticket. Non-existent tags are handled gracefully.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  untagTicket,
}

// commentsCmd represents the comments command
var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage comments on a ticket",
	Long:  `List and add comments to tickets.`,
}

// attachmentsCmd represents the attachments command
var attachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage attachments on a ticket",
	Long:  `List and add attachments to tickets.`,
}

// worklogsCmd represents the worklogs command
var worklogsCmd = &cobra.Command{
	Use:   "worklogs",
	Short: "Manage worklogs on a ticket",
	Long:  `List and add worklogs to tickets.`,
}

// linksCmd represents the links command
var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage links between tickets",
	Long:  `Create and manage links between tickets.`,
}

// listCommentsCmd represents the comments list command
var listCommentsCmd = &cobra.Command{
	Use:   "list <ticket_id>",
	Short: "Lists all comments for a specific ticket",
	Long:  `Lists all comments for a specific ticket with author and date information.`,
	Args:  cobra.ExactArgs(1),
	RunE:  listComments,
}

// addCommentCmd represents the comments add command
var addCommentCmd = &cobra.Command{
	Use:   "add <ticket_id>",
	Short: "Adds a comment to a ticket",
	Long:  `Adds a new comment to a ticket with the specified message.`,
	Args:  cobra.ExactArgs(1),
	RunE:  addComment,
}

// listAttachmentsCmd represents the attachments list command
var listAttachmentsCmd = &cobra.Command{
	Use:   "list <ticket_id>",
	Short: "Lists all attachments for a specific ticket",
	Long:  `Lists all attachments for a specific ticket with name, size, and author information.`,
	Args:  cobra.ExactArgs(1),
	RunE:  listAttachments,
}

// addAttachmentCmd represents the attachments add command
var addAttachmentCmd = &cobra.Command{
	Use:   "add <ticket_id> <file_path>",
	Short: "Attaches a file to a ticket",
	Long:  `Uploads a file as an attachment to the specified ticket.`,
	Args:  cobra.ExactArgs(2),
	RunE:  addAttachment,
}

// listWorklogsCmd represents the worklogs list command
var listWorklogsCmd = &cobra.Command{
	Use:   "list <ticket_id>",
	Short: "Lists all worklog entries for a specific ticket",
	Long:  `Lists all worklog entries for a specific ticket with author, date, duration, and description information.`,
	Args:  cobra.ExactArgs(1),
	RunE:  listWorklogs,
}

// addWorklogCmd represents the worklogs add command
var addWorklogCmd = &cobra.Command{
	Use:   "add <ticket_id>",
	Short: "Adds a worklog entry to a ticket",
	Long:  `Adds a new worklog entry to a ticket with the specified duration and optional description.`,
	Args:  cobra.ExactArgs(1),
	RunE:  addWorklog,
}

// addLinkCmd represents the links add command
var addLinkCmd = &cobra.Command{
	Use:   "add <ticket_id> <other_ticket_id>",
	Short: "Links two tickets together",
	Long:  `Links two tickets together with the specified relationship type.`,
	Args:  cobra.ExactArgs(2),
	RunE:  addLink,
}

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history <ticket_id>",
	Short: "Shows the activity stream for a ticket",
	Long:  `Shows the activity stream for a ticket including field changes, comments, attachments, and other activities in chronological order.`,
	Args:  cobra.ExactArgs(1),
	RunE:  showHistory,
}

func init() {
	// Add subcommands
	TicketsCmd.AddCommand(listTicketsCmd)
	TicketsCmd.AddCommand(showTicketCmd)
	TicketsCmd.AddCommand(createTicketCmd)
	TicketsCmd.AddCommand(updateTicketCmd)
	TicketsCmd.AddCommand(tagTicketCmd)
	TicketsCmd.AddCommand(untagTicketCmd)
	TicketsCmd.AddCommand(commentsCmd)
	TicketsCmd.AddCommand(attachmentsCmd)
	TicketsCmd.AddCommand(worklogsCmd)
	TicketsCmd.AddCommand(linksCmd)
	TicketsCmd.AddCommand(historyCmd)

	// Add comments subcommands
	commentsCmd.AddCommand(listCommentsCmd)
	commentsCmd.AddCommand(addCommentCmd)

	// Add attachments subcommands
	attachmentsCmd.AddCommand(listAttachmentsCmd)
	attachmentsCmd.AddCommand(addAttachmentCmd)

	// Add worklogs subcommands
	worklogsCmd.AddCommand(listWorklogsCmd)
	worklogsCmd.AddCommand(addWorklogCmd)

	// Add links subcommands
	linksCmd.AddCommand(addLinkCmd)

	// Add flags to both tickets and tickets list commands for the alias to work
	TicketsCmd.Flags().StringVarP(&projectID, "project", "p", "", "The project ID (uses default from config if not provided)")
	TicketsCmd.Flags().StringVarP(&userID, "user", "u", "", "Filter tickets by assignee (defaults to current user from config)")
	TicketsCmd.Flags().StringVarP(&query, "query", "q", "", "Filter tickets with a YouTrack search query")
	TicketsCmd.Flags().IntVar(&limit, "limit", 20, "Number of tickets to show")

	listTicketsCmd.Flags().StringVarP(&projectID, "project", "p", "", "The project ID (uses default from config if not provided)")
	listTicketsCmd.Flags().StringVarP(&userID, "user", "u", "", "Filter tickets by assignee (defaults to current user from config)")
	listTicketsCmd.Flags().StringVarP(&query, "query", "q", "", "Filter tickets with a YouTrack search query")
	listTicketsCmd.Flags().IntVar(&limit, "limit", 20, "Number of tickets to show")

	// Add flags for create command
	createTicketCmd.Flags().StringVarP(&projectID, "project", "p", "", "The project ID (uses default from config if not provided)")
	createTicketCmd.Flags().StringVarP(&createTitle, "title", "t", "", "The title of the new ticket (required)")
	createTicketCmd.Flags().StringVarP(&createDescription, "description", "d", "", "The description for the ticket")
	createTicketCmd.Flags().StringVar(&createAssignee, "assignee", "", "Assign the ticket to a user")
	createTicketCmd.Flags().StringSliceVar(&createFields, "field", []string{}, "Set a custom field (key=value format). Can be specified multiple times")
	createTicketCmd.MarkFlagRequired("title")

	// Add flags for update command
	updateTicketCmd.Flags().StringVar(&updateStatus, "status", "", "Change the ticket's status")
	updateTicketCmd.Flags().StringVar(&updateAssignee, "assignee", "", "Change the assignee")
	updateTicketCmd.Flags().StringSliceVar(&updateFields, "field", []string{}, "Set a custom field (key=value format). Can be specified multiple times")

	// Add flags for comment add command
	addCommentCmd.Flags().StringVarP(&commentMessage, "message", "m", "", "The comment message (required)")
	addCommentCmd.MarkFlagRequired("message")

	// Add flags for worklog add command
	addWorklogCmd.Flags().StringVar(&worklogDuration, "duration", "", "The duration of the work (e.g., '1h 30m') (required)")
	addWorklogCmd.Flags().StringVar(&worklogDescription, "description", "", "An optional description for the worklog entry")
	addWorklogCmd.MarkFlagRequired("duration")

	// Add flags for link add command
	addLinkCmd.Flags().StringVar(&linkType, "type", "relates to", "The relationship type (e.g., 'relates to', 'is duplicated by')")
}

// GetOutputFlag gets the output flag from the command hierarchy
func getOutputFlag(cmd *cobra.Command) string {
	// Walk up the command hierarchy to find the output flag
	current := cmd
	for current != nil {
		if flag := current.Flag("output"); flag != nil {
			return flag.Value.String()
		}
		current = current.Parent()
	}
	return "text" // default
}

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

// isValidTicketID validates the ticket ID format (e.g., "PRJ-123")
func isValidTicketID(ticketID string) bool {
	// Pattern: one or more alphanumeric characters, dash, one or more digits
	pattern := `^[A-Za-z0-9]+-\d+$`
	matched, err := regexp.MatchString(pattern, ticketID)
	return err == nil && matched
}

// outputResult outputs data in the requested format (text or JSON)
func outputResult(cmd *cobra.Command, data interface{}, formatAsText func(interface{}) error) error {
	outputFlag := getOutputFlag(cmd)

	switch outputFlag {
	case "json":
		return outputJSON(data)
	case "text":
		return formatAsText(data)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlag)
	}
}

// outputJSON outputs data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatTicketsList formats tickets list for text output
func formatTicketsList(data interface{}) error {
	tickets := data.([]*youtrack.Issue)

	if len(tickets) == 0 {
		fmt.Println("No tickets found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSUMMARY\tASSIGNEE\tUPDATED\tTAGS")
	fmt.Fprintln(w, "──\t───────\t────────\t───────\t────")

	for _, ticket := range tickets {
		assignee := "Unassigned"
		if ticket.Assignee != nil {
			assignee = ticket.Assignee.FullName
			if assignee == "" {
				assignee = ticket.Assignee.Login
			}
		}

		// Format updated time
		updated := ticket.Updated.Time.Format("2006-01-02")

		// Format tags
		var tagNames []string
		for _, tag := range ticket.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tags := strings.Join(tagNames, ", ")
		if tags == "" {
			tags = "-"
		}

		// Truncate summary if too long
		summary := ticket.Summary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ticket.ID,
			summary,
			assignee,
			updated,
			tags)
	}

	return w.Flush()
}

// formatTicketDetails formats ticket details for text output
func formatTicketDetails(data interface{}) error {
	ticket := data.(*youtrack.Issue)

	fmt.Printf("Ticket Details\n")
	fmt.Printf("==============\n\n")

	fmt.Printf("ID:          %s\n", ticket.ID)
	fmt.Printf("Summary:     %s\n", ticket.Summary)

	if ticket.Description != "" {
		fmt.Printf("Description: %s\n", ticket.Description)
	}

	if ticket.Assignee != nil {
		assignee := ticket.Assignee.FullName
		if assignee == "" {
			assignee = ticket.Assignee.Login
		}
		fmt.Printf("Assignee:    %s\n", assignee)
	} else {
		fmt.Printf("Assignee:    Unassigned\n")
	}

	if ticket.Reporter != nil {
		reporter := ticket.Reporter.FullName
		if reporter == "" {
			reporter = ticket.Reporter.Login
		}
		fmt.Printf("Reporter:    %s\n", reporter)
	}

	fmt.Printf("Created:     %s\n", ticket.Created.Time.Format(time.RFC3339))
	fmt.Printf("Updated:     %s\n", ticket.Updated.Time.Format(time.RFC3339))

	if ticket.Resolved != nil {
		fmt.Printf("Resolved:    %s\n", ticket.Resolved.Time.Format(time.RFC3339))
	}

	// Display tags if available
	if len(ticket.Tags) > 0 {
		fmt.Printf("\nTags\n")
		fmt.Printf("────\n")
		for _, tag := range ticket.Tags {
			fmt.Printf("- %s\n", tag.Name)
		}
	}

	return nil
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

// parseCustomFields parses key=value pairs from the --field flags
func parseCustomFields(fields []string) (map[string]interface{}, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	customFields := make(map[string]interface{})
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format: %s (expected key=value)", field)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty field key in: %s", field)
		}

		// For now, treat all custom field values as strings
		// More complex field types can be handled later
		customFields[key] = value
	}

	return customFields, nil
}

// formatTicketCreated formats the created ticket for text output
func formatTicketCreated(data interface{}) error {
	ticket := data.(*youtrack.Issue)

	fmt.Printf("Ticket created successfully!\n\n")
	fmt.Printf("ID:      %s\n", ticket.ID)
	fmt.Printf("Summary: %s\n", ticket.Summary)

	if ticket.Description != "" {
		fmt.Printf("Description: %s\n", ticket.Description)
	}

	if ticket.Assignee != nil {
		assignee := ticket.Assignee.FullName
		if assignee == "" {
			assignee = ticket.Assignee.Login
		}
		fmt.Printf("Assignee: %s\n", assignee)
	}

	fmt.Printf("Created: %s\n", ticket.Created.Time.Format(time.RFC3339))

	return nil
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

// extractProjectFromTicketID extracts the project part from a ticket ID (e.g., "PRJ-123" -> "PRJ")
func extractProjectFromTicketID(ticketID string) string {
	parts := strings.Split(ticketID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// UpdateSummary contains information about what was changed in a ticket update
type UpdateSummary struct {
	TicketID        string
	OriginalTicket  *youtrack.Issue
	UpdatedTicket   *youtrack.Issue
	StatusChanged   string
	AssigneeChanged string
	FieldsChanged   []string
}

// formatTicketUpdated formats the update confirmation for text output
func formatTicketUpdated(data interface{}) error {
	summary := data.(*UpdateSummary)

	fmt.Printf("Ticket updated successfully!\n\n")
	fmt.Printf("Ticket: %s\n", summary.TicketID)
	fmt.Printf("Summary: %s\n\n", summary.UpdatedTicket.Summary)

	fmt.Printf("Changes made:\n")
	fmt.Printf("─────────────\n")

	hasChanges := false

	// Show status change
	if summary.StatusChanged != "" {
		fmt.Printf("• Status: %s\n", summary.StatusChanged)
		hasChanges = true
	}

	// Show assignee change
	if summary.AssigneeChanged != "" {
		assignee := "Unassigned"
		if summary.UpdatedTicket.Assignee != nil {
			assignee = summary.UpdatedTicket.Assignee.FullName
			if assignee == "" {
				assignee = summary.UpdatedTicket.Assignee.Login
			}
		}
		fmt.Printf("• Assignee: %s\n", assignee)
		hasChanges = true
	}

	// Show custom field changes
	if len(summary.FieldsChanged) > 0 {
		for _, field := range summary.FieldsChanged {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				fmt.Printf("• %s: %s\n", parts[0], parts[1])
				hasChanges = true
			}
		}
	}

	if !hasChanges {
		fmt.Printf("• No changes specified\n")
	}

	fmt.Printf("\nUpdated: %s\n", summary.UpdatedTicket.Updated.Time.Format(time.RFC3339))

	return nil
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

// TagOperationResult represents the result of a single tag operation
type TagOperationResult struct {
	TagName string
	Success bool
	Error   string
}

// TagOperationSummary contains information about tag operations performed
type TagOperationSummary struct {
	TicketID  string
	Operation string // "add" or "remove"
	Results   []TagOperationResult
	HasErrors bool
}

// formatTagOperationSummary formats the tag operation results for text output
func formatTagOperationSummary(data interface{}) error {
	summary := data.(*TagOperationSummary)

	if summary.Operation == "add" {
		fmt.Printf("Tag operation completed for ticket: %s\n\n", summary.TicketID)
	} else {
		fmt.Printf("Untag operation completed for ticket: %s\n\n", summary.TicketID)
	}

	// Count successful and failed operations
	successCount := 0
	failCount := 0
	for _, result := range summary.Results {
		if result.Success {
			successCount++
		} else {
			failCount++
		}
	}

	// Show successful operations
	if successCount > 0 {
		if summary.Operation == "add" {
			fmt.Printf("Successfully added tags:\n")
		} else {
			fmt.Printf("Successfully removed tags:\n")
		}
		fmt.Printf("────────────────────────\n")
		for _, result := range summary.Results {
			if result.Success {
				fmt.Printf("✓ %s\n", result.TagName)
			}
		}
		fmt.Printf("\n")
	}

	// Show failed operations
	if failCount > 0 {
		if summary.Operation == "add" {
			fmt.Printf("Failed to add tags:\n")
		} else {
			fmt.Printf("Failed to remove tags:\n")
		}
		fmt.Printf("───────────────────\n")
		for _, result := range summary.Results {
			if !result.Success {
				fmt.Printf("✗ %s: %s\n", result.TagName, result.Error)
			}
		}
		fmt.Printf("\n")
	}

	// Summary line
	if summary.Operation == "add" {
		fmt.Printf("Summary: %d tag(s) added successfully", successCount)
	} else {
		fmt.Printf("Summary: %d tag(s) removed successfully", successCount)
	}

	if failCount > 0 {
		fmt.Printf(", %d failed", failCount)
	}
	fmt.Printf("\n")

	return nil
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

// formatCommentsList formats comments list for text output
func formatCommentsList(data interface{}) error {
	comments := data.([]*youtrack.IssueComment)

	if len(comments) == 0 {
		fmt.Println("No comments found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAUTHOR\tCREATED\tTEXT")
	fmt.Fprintln(w, "──\t──────\t───────\t────")

	for _, comment := range comments {
		author := "Unknown"
		if comment.Author != nil {
			author = comment.Author.FullName
			if author == "" {
				author = comment.Author.Login
			}
		}

		// Format created time
		created := comment.Created.Time.Format("2006-01-02 15:04")

		// Truncate text if too long for table display
		text := comment.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		// Replace newlines with spaces for table display
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			comment.ID,
			author,
			created,
			text)
	}

	return w.Flush()
}

// formatCommentAdded formats the added comment for text output
func formatCommentAdded(data interface{}) error {
	comment := data.(*youtrack.IssueComment)

	fmt.Printf("Comment added successfully!\n\n")
	fmt.Printf("ID:      %s\n", comment.ID)

	if comment.Author != nil {
		author := comment.Author.FullName
		if author == "" {
			author = comment.Author.Login
		}
		fmt.Printf("Author:  %s\n", author)
	}

	fmt.Printf("Created: %s\n", comment.Created.Time.Format(time.RFC3339))
	fmt.Printf("Text:    %s\n", comment.Text)

	return nil
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

// formatAttachmentsList formats attachments list for text output
func formatAttachmentsList(data interface{}) error {
	attachments := data.([]*youtrack.Attachment)

	if len(attachments) == 0 {
		fmt.Println("No attachments found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSIZE\tAUTHOR\tCREATED")
	fmt.Fprintln(w, "──\t────\t────\t──────\t───────")

	for _, attachment := range attachments {
		author := "Unknown"
		if attachment.Author != nil {
			author = attachment.Author.FullName
			if author == "" {
				author = attachment.Author.Login
			}
		}

		// Format created time
		created := attachment.Created.Time.Format("2006-01-02 15:04")

		// Format file size in a human-readable way
		size := formatFileSize(attachment.Size)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			attachment.ID,
			attachment.Name,
			size,
			author,
			created)
	}

	return w.Flush()
}

// formatAttachmentAdded formats the added attachment for text output
func formatAttachmentAdded(data interface{}) error {
	attachment := data.(*youtrack.Attachment)

	fmt.Printf("Attachment added successfully!\n\n")
	fmt.Printf("ID:      %s\n", attachment.ID)
	fmt.Printf("Name:    %s\n", attachment.Name)
	fmt.Printf("Size:    %s\n", formatFileSize(attachment.Size))

	if attachment.Author != nil {
		author := attachment.Author.FullName
		if author == "" {
			author = attachment.Author.Login
		}
		fmt.Printf("Author:  %s\n", author)
	}

	fmt.Printf("Created: %s\n", attachment.Created.Time.Format(time.RFC3339))

	if attachment.MimeType != "" {
		fmt.Printf("Type:    %s\n", attachment.MimeType)
	}

	return nil
}

// formatFileSize formats file size in a human-readable way
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
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

// parseDuration parses a duration string like "1h 30m" or "90m" into minutes
func parseDuration(durationStr string) (int, error) {
	durationStr = strings.TrimSpace(durationStr)
	if durationStr == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	// Regular expression to match time units
	re := regexp.MustCompile(`(\d+)\s*([hm]?)`)
	matches := re.FindAllStringSubmatch(durationStr, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format (examples: '1h', '30m', '1h 30m', '90m')")
	}

	totalMinutes := 0
	usedUnits := make(map[string]bool)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		valueStr := match[1]
		unit := match[2]

		// Default to minutes if no unit specified
		if unit == "" {
			unit = "m"
		}

		// Check for duplicate units
		if usedUnits[unit] {
			return 0, fmt.Errorf("duplicate time unit '%s' in duration", unit)
		}
		usedUnits[unit] = true

		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number '%s' in duration", valueStr)
		}

		if value < 0 {
			return 0, fmt.Errorf("negative values not allowed in duration")
		}

		switch unit {
		case "h":
			totalMinutes += value * 60
		case "m":
			totalMinutes += value
		default:
			return 0, fmt.Errorf("invalid time unit '%s' (use 'h' for hours or 'm' for minutes)", unit)
		}
	}

	if totalMinutes <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return totalMinutes, nil
}

// formatDuration formats minutes into a human-readable duration
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

// formatWorklogsList formats worklogs list for text output
func formatWorklogsList(data interface{}) error {
	worklogs := data.([]*youtrack.WorkItem)

	if len(worklogs) == 0 {
		fmt.Println("No worklogs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAUTHOR\tDATE\tDURATION\tDESCRIPTION")
	fmt.Fprintln(w, "──\t──────\t────\t────────\t───────────")

	for _, worklog := range worklogs {
		author := "Unknown"
		if worklog.Author != nil {
			author = worklog.Author.FullName
			if author == "" {
				author = worklog.Author.Login
			}
		}

		// Format date
		date := worklog.Date.Time.Format("2006-01-02")

		// Format duration
		duration := formatDuration(worklog.Duration)

		// Truncate description if too long for table display
		description := worklog.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}
		// Replace newlines with spaces for table display
		description = strings.ReplaceAll(description, "\n", " ")
		description = strings.ReplaceAll(description, "\r", " ")
		if description == "" {
			description = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			worklog.ID,
			author,
			date,
			duration,
			description)
	}

	return w.Flush()
}

// formatWorklogAdded formats the added worklog for text output
func formatWorklogAdded(data interface{}) error {
	worklog := data.(*youtrack.WorkItem)

	fmt.Printf("Worklog added successfully!\n\n")
	fmt.Printf("ID:       %s\n", worklog.ID)

	if worklog.Author != nil {
		author := worklog.Author.FullName
		if author == "" {
			author = worklog.Author.Login
		}
		fmt.Printf("Author:   %s\n", author)
	}

	fmt.Printf("Date:     %s\n", worklog.Date.Time.Format("2006-01-02"))
	fmt.Printf("Duration: %s\n", formatDuration(worklog.Duration))

	if worklog.Description != "" {
		fmt.Printf("Description: %s\n", worklog.Description)
	}

	if worklog.Type != nil {
		fmt.Printf("Type:     %s\n", worklog.Type.Name)
	}

	return nil
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

// LinkOperationSummary contains information about a link operation
type LinkOperationSummary struct {
	SourceTicketID string
	TargetTicketID string
	LinkType       string
	Success        bool
	Error          string
}

// formatLinkOperationSummary formats the link operation results for text output
func formatLinkOperationSummary(data interface{}) error {
	summary := data.(*LinkOperationSummary)

	if summary.Success {
		fmt.Printf("Link created successfully!\n\n")
		fmt.Printf("Source:   %s\n", summary.SourceTicketID)
		fmt.Printf("Target:   %s\n", summary.TargetTicketID)
		fmt.Printf("Type:     %s\n", summary.LinkType)
		fmt.Printf("\nThe tickets are now linked with the '%s' relationship.\n", summary.LinkType)
	} else {
		fmt.Printf("Failed to create link!\n\n")
		fmt.Printf("Source:   %s\n", summary.SourceTicketID)
		fmt.Printf("Target:   %s\n", summary.TargetTicketID)
		fmt.Printf("Type:     %s\n", summary.LinkType)
		fmt.Printf("Error:    %s\n", summary.Error)
	}

	return nil
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

// HistorySummary contains the ticket history information
type HistorySummary struct {
	TicketID   string
	Activities []*youtrack.ActivityItem
}

// formatHistorySummary formats the activity history for text output
func formatHistorySummary(data interface{}) error {
	summary := data.(*HistorySummary)

	if len(summary.Activities) == 0 {
		fmt.Printf("No activity history found for ticket: %s\n", summary.TicketID)
		return nil
	}

	fmt.Printf("Activity History for Ticket: %s\n", summary.TicketID)
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n\n")

	for _, activity := range summary.Activities {
		// Format timestamp
		timestamp := activity.Timestamp.Time.Format("2006-01-02 15:04:05")

		// Format author
		author := "System"
		if activity.Author != nil {
			author = activity.Author.FullName
			if author == "" {
				author = activity.Author.Login
			}
		}

		// Format activity description based on category
		description := formatActivityDescription(activity)

		fmt.Printf("┌─ %s by %s\n", timestamp, author)
		fmt.Printf("│  %s\n", description)

		// Show field changes if available
		if activity.Field != nil && (activity.Added != nil || activity.Removed != nil || len(activity.AddedValues) > 0 || len(activity.RemovedValues) > 0) {
			fieldName := activity.Field.Name
			if fieldName == "" {
				fieldName = activity.Field.ID
			}

			// Handle single field changes
			if activity.Removed != nil || activity.Added != nil {
				fmt.Printf("│  Field: %s\n", fieldName)
				if activity.Removed != nil {
					oldValue := formatFieldValue(activity.Removed)
					fmt.Printf("│    From: %s\n", oldValue)
				}
				if activity.Added != nil {
					newValue := formatFieldValue(activity.Added)
					fmt.Printf("│    To:   %s\n", newValue)
				}
			}

			// Handle multiple field changes (arrays)
			if len(activity.RemovedValues) > 0 {
				fmt.Printf("│  Field: %s\n", fieldName)
				fmt.Printf("│    Removed: ")
				for i, val := range activity.RemovedValues {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", formatFieldValue(val))
				}
				fmt.Printf("\n")
			}

			if len(activity.AddedValues) > 0 {
				if len(activity.RemovedValues) == 0 {
					fmt.Printf("│  Field: %s\n", fieldName)
				}
				fmt.Printf("│    Added: ")
				for i, val := range activity.AddedValues {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", formatFieldValue(val))
				}
				fmt.Printf("\n")
			}
		}

		fmt.Printf("└─\n\n")
	}

	fmt.Printf("Total activities: %d\n", len(summary.Activities))

	return nil
}

// formatActivityDescription formats the activity description based on category
func formatActivityDescription(activity *youtrack.ActivityItem) string {
	categoryID := activity.Category.ID

	switch categoryID {
	case "IssueCreatedCategory":
		return "Issue created"
	case "CommentCategory":
		return "Comment added"
	case "AttachmentCategory":
		return "Attachment added"
	case "WorkItemCategory":
		return "Work item added"
	case "LinkCategory":
		return "Issue link created"
	case "IssueResolvedCategory":
		return "Issue resolved"
	case "CustomFieldCategory":
		if activity.Field != nil {
			fieldName := activity.Field.Name
			if fieldName == "" {
				fieldName = activity.Field.ID
			}
			return fmt.Sprintf("Field '%s' changed", fieldName)
		}
		return "Custom field changed"
	case "SummaryCategory":
		return "Summary changed"
	case "DescriptionCategory":
		return "Description changed"
	case "TagCategory":
		return "Tags changed"
	case "AssigneeCategory":
		return "Assignee changed"
	default:
		return fmt.Sprintf("Activity: %s", categoryID)
	}
}

// formatFieldValue formats a field value for display
func formatFieldValue(value *youtrack.FieldValue) string {
	if value == nil {
		return "(none)"
	}

	// Try different value fields in order of preference
	if value.Text != "" {
		return value.Text
	}
	if value.Name != "" {
		return value.Name
	}
	if value.FullName != "" {
		return value.FullName
	}
	if value.Login != "" {
		return value.Login
	}
	if value.Markdown != "" {
		return value.Markdown
	}
	if value.ID != "" {
		return value.ID
	}

	return "(empty)"
}
