package tickets

import "github.com/mkozhukh/youtrack/pkg/youtrack"

// UpdateSummary contains information about what was changed in a ticket update
type UpdateSummary struct {
	TicketID        string
	OriginalTicket  *youtrack.Issue
	UpdatedTicket   *youtrack.Issue
	StatusChanged   string
	AssigneeChanged string
	FieldsChanged   []string
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

// LinkOperationSummary contains information about a link operation
type LinkOperationSummary struct {
	SourceTicketID string
	TargetTicketID string
	LinkType       string
	Success        bool
	Error          string
}

// HistorySummary contains the ticket history information
type HistorySummary struct {
	TicketID   string
	Activities []*youtrack.ActivityItem
}
