package tickets

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"mkozhukh/youtrack/pkg/youtrack"
)

// formatTicketsList formats tickets list for text output
func formatTicketsList(data interface{}) error {
	tickets := data.([]*youtrack.Issue)

	if len(tickets) == 0 {
		fmt.Println("No tickets found")
		return nil
	}

	cellStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("246"))
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			return cellStyle
		}).
		Headers("ID", "SUMMARY", "ASSIGNEE", "UPDATED", "TAGS")

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

		t.Row(
			ticket.ID,
			summary,
			assignee,
			updated,
			tags,
		)
	}

	fmt.Println(t)
	return nil
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

// formatCommentsList formats comments list for text output
func formatCommentsList(data interface{}) error {
	comments := data.([]*youtrack.IssueComment)

	if len(comments) == 0 {
		fmt.Println("No comments found")
		return nil
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
			}
		}).
		Headers("ID", "AUTHOR", "CREATED", "TEXT")

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

		t.Row(
			comment.ID,
			author,
			created,
			text,
		)
	}

	fmt.Println(t)
	return nil
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

// formatAttachmentsList formats attachments list for text output
func formatAttachmentsList(data interface{}) error {
	attachments := data.([]*youtrack.Attachment)

	if len(attachments) == 0 {
		fmt.Println("No attachments found")
		return nil
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
			}
		}).
		Headers("ID", "NAME", "SIZE", "AUTHOR", "CREATED")

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

		t.Row(
			attachment.ID,
			attachment.Name,
			size,
			author,
			created,
		)
	}

	fmt.Println(t)
	return nil
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

// formatWorklogsList formats worklogs list for text output
func formatWorklogsList(data interface{}) error {
	worklogs := data.([]*youtrack.WorkItem)

	if len(worklogs) == 0 {
		fmt.Println("No worklogs found")
		return nil
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
			}
		}).
		Headers("ID", "AUTHOR", "DATE", "DURATION", "DESCRIPTION")

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

		t.Row(
			worklog.ID,
			author,
			date,
			duration,
			description,
		)
	}

	fmt.Println(t)
	return nil
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
