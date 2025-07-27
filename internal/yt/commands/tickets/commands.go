package tickets

import (
	"github.com/spf13/cobra"
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
