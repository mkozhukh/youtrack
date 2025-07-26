package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

var (
	usersProject    string
	worklogsProject string
	startDate       string
	endDate         string
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `List and show worklog information for YouTrack users.`,
	RunE:  listUsers, // Default to list when no subcommand is given
}

// listUsersCmd represents the list command
var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows all users associated with a project",
	Long:  `Shows all users associated with a project (the project team).`,
	RunE:  listUsers,
}

// userWorklogsCmd represents the worklogs command
var userWorklogsCmd = &cobra.Command{
	Use:   "worklogs <user>",
	Short: "Lists all worklog entries for a specific user",
	Long: `Lists all worklog entries for a specific user. The user can be specified by username or email.
Supports partial matching for user lookups.`,
	Args: cobra.ExactArgs(1),
	RunE: getUserWorklogs,
}

func init() {
	usersCmd.AddCommand(listUsersCmd)
	usersCmd.AddCommand(userWorklogsCmd)

	// Add project flag to both users and users list commands
	usersCmd.Flags().StringVarP(&usersProject, "project", "p", "", "The project ID")
	listUsersCmd.Flags().StringVarP(&usersProject, "project", "p", "", "The project ID")

	// Add flags for worklogs command
	userWorklogsCmd.Flags().StringVarP(&worklogsProject, "project", "p", "", "Filter worklogs by project")
	userWorklogsCmd.Flags().StringVar(&startDate, "since", "", "Show worklogs since a specific date (e.g., '2025-07-01')")
	userWorklogsCmd.Flags().StringVar(&endDate, "until", "", "Show worklogs until a specific date")
}

func listUsers(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile, cmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Determine project ID to use
	projectID := usersProject
	if projectID == "" {
		projectID = cfg.Defaults.Project
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required (use --project flag or set default project in config)")
	}

	// Create client and context
	client := youtrack.NewClient(cfg.Server.URL)
	ctx := youtrack.NewYouTrackContext(context.Background(), cfg.Server.Token)

	// Fetch all project users
	users, err := fetchAllProjectUsers(client, ctx, projectID)
	if err != nil {
		log.Error("Failed to fetch project users", "error", err)
		return fmt.Errorf("failed to fetch project users: %w", err)
	}

	// Output results
	return outputResult(users, func(data interface{}) error {
		return formatUsersList(data.([]*youtrack.User))
	})
}

func getUserWorklogs(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile, cmd.Flags())
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

	// Find the user (with partial matching)
	user, err := findUser(client, ctx, username, worklogsProject, cfg.Defaults.Project)
	if err != nil {
		log.Error("Failed to find user", "error", err)
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Validate date formats if provided
	if startDate != "" {
		if _, err := parseDate(startDate); err != nil {
			return fmt.Errorf("invalid start date format: %s (use YYYY-MM-DD)", startDate)
		}
	}
	if endDate != "" {
		if _, err := parseDate(endDate); err != nil {
			return fmt.Errorf("invalid end date format: %s (use YYYY-MM-DD)", endDate)
		}
	}

	// Fetch user worklogs
	workItems, err := fetchAllUserWorklogs(client, ctx, user.ID, worklogsProject, startDate, endDate)
	if err != nil {
		log.Error("Failed to fetch user worklogs", "error", err)
		return fmt.Errorf("failed to fetch user worklogs: %w", err)
	}

	// Create a structure that includes user info for JSON output
	worklogsResponse := struct {
		User      *youtrack.User       `json:"user"`
		WorkItems []*youtrack.WorkItem `json:"workItems"`
	}{
		User:      user,
		WorkItems: workItems,
	}

	// Output results
	return outputResult(worklogsResponse, func(data interface{}) error {
		resp := data.(struct {
			User      *youtrack.User       `json:"user"`
			WorkItems []*youtrack.WorkItem `json:"workItems"`
		})
		return formatUserWorklogs(resp.User, resp.WorkItems)
	})
}

// fetchAllProjectUsers retrieves all users for a project
func fetchAllProjectUsers(client *youtrack.Client, ctx *youtrack.YouTrackContext, projectID string) ([]*youtrack.User, error) {
	var allUsers []*youtrack.User
	skip := 0
	top := 100

	for {
		users, err := client.GetProjectUsers(ctx, projectID, skip, top)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)

		// If we got fewer users than requested, we've reached the end
		if len(users) < top {
			break
		}

		skip += top
	}

	return allUsers, nil
}

// fetchAllUserWorklogs retrieves all worklog entries for a user
func fetchAllUserWorklogs(client *youtrack.Client, ctx *youtrack.YouTrackContext, userID, projectID, startDate, endDate string) ([]*youtrack.WorkItem, error) {
	var allWorkItems []*youtrack.WorkItem
	skip := 0
	top := 100

	for {
		workItems, err := client.GetUserWorklogs(ctx, userID, projectID, startDate, endDate, skip, top)
		if err != nil {
			return nil, err
		}

		allWorkItems = append(allWorkItems, workItems...)

		// If we got fewer work items than requested, we've reached the end
		if len(workItems) < top {
			break
		}

		skip += top
	}

	return allWorkItems, nil
}

// findUser finds a user by username with partial matching
func findUser(client *youtrack.Client, ctx *youtrack.YouTrackContext, username, projectID, defaultProject string) (*youtrack.User, error) {
	// First try exact match by login
	if user, err := client.GetUserByLogin(ctx, username); err == nil {
		return user, nil
	}

	// If that fails, try partial matching within a project scope
	searchProject := projectID
	if searchProject == "" {
		searchProject = defaultProject
	}

	if searchProject != "" {
		if user, err := client.SuggestUserByProject(ctx, searchProject, username); err == nil {
			return user, nil
		}
	}

	// If no project scope or project-scoped search failed, search globally
	users, err := client.SearchUsers(ctx, username, 0, 10)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found matching '%s'", username)
	}

	// Look for exact matches first (case insensitive)
	lowerUsername := strings.ToLower(username)
	for _, user := range users {
		if strings.ToLower(user.Login) == lowerUsername ||
			strings.ToLower(user.Email) == lowerUsername {
			return user, nil
		}
	}

	// Look for partial matches
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Login), lowerUsername) ||
			strings.Contains(strings.ToLower(user.FullName), lowerUsername) ||
			strings.Contains(strings.ToLower(user.Email), lowerUsername) {
			return user, nil
		}
	}

	return users[0], nil // Return first match if no better match found
}

// parseDate parses a date in YYYY-MM-DD format
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// formatUsersList formats users list for text output
func formatUsersList(users []*youtrack.User) error {
	if len(users) == 0 {
		fmt.Println("No users found in project")
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
		Headers("LOGIN", "NAME", "EMAIL")

	for _, user := range users {
		t.Row(user.Login, user.FullName, user.Email)
	}

	fmt.Println(t)
	return nil
}

// formatUserWorklogs formats user worklogs for text output
func formatUserWorklogs(user *youtrack.User, workItems []*youtrack.WorkItem) error {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	fmt.Printf("%s\n", headerStyle.Render(fmt.Sprintf("Worklogs for %s (%s)", user.FullName, user.Login)))
	fmt.Printf("%s\n\n", headerStyle.Render("=================================="))

	if len(workItems) == 0 {
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
		Headers("DATE", "DURATION", "ISSUE", "DESCRIPTION")

	for _, item := range workItems {
		// Format duration from minutes to human readable
		duration := formatDuration(item.Duration)

		// Get issue ID if available
		issueID := ""
		if item.Issue != nil {
			issueID = item.Issue.ID
		}

		// Truncate description if too long
		description := item.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		t.Row(
			item.Date.Format("2006-01-02"),
			duration,
			issueID,
			description,
		)
	}

	fmt.Println(t)
	return nil
}

// formatDuration converts minutes to human readable format
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
