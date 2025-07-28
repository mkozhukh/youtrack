package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"mkozhukh/youtrack/internal/yt/config"
	"mkozhukh/youtrack/pkg/youtrack"
)

var (
	projectsQuery string
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  `List and describe YouTrack projects.`,
	RunE:  listProjects, // Default to list when no subcommand is given
}

// listProjectsCmd represents the list command
var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows a list of all available projects",
	Long:  `Shows a list of all available projects with optional query filter.`,
	RunE:  listProjects,
}

// describeProjectCmd represents the describe command
var describeProjectCmd = &cobra.Command{
	Use:   "describe <project_id>",
	Short: "Shows detailed information for a specific project",
	Long: `Shows detailed information for a specific project, including 
available custom fields, statuses, and types.`,
	Args: cobra.ExactArgs(1),
	RunE: describeProject,
}

func init() {
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(describeProjectCmd)

	// Add query flag to both projects and projects list commands
	projectsCmd.Flags().StringVarP(&projectsQuery, "query", "q", "", "Filter projects by a search query")
	listProjectsCmd.Flags().StringVarP(&projectsQuery, "query", "q", "", "Filter projects by a search query")
}

func listProjects(cmd *cobra.Command, args []string) error {
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

	// Fetch projects
	projects, err := fetchAllProjects(client, ctx)
	if err != nil {
		log.Error("Failed to fetch projects", "error", err)
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Filter projects if query is provided
	if projectsQuery != "" {
		projects = filterProjects(projects, projectsQuery)
	}

	// Output results
	return outputResult(projects, func(data interface{}) error {
		return formatProjectsList(data.([]*youtrack.Project))
	})
}

func describeProject(cmd *cobra.Command, args []string) error {
	projectID := args[0]

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

	// Fetch project details
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		if apiErr, ok := err.(*youtrack.APIError); ok && apiErr.StatusCode == 404 {
			return fmt.Errorf("project not found: %s", projectID)
		}
		log.Error("Failed to fetch project", "error", err)
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	// Fetch project custom fields
	customFields, err := fetchProjectCustomFields(client, ctx, projectID)
	if err != nil {
		log.Warn("Failed to fetch custom fields", "error", err)
		// Don't fail the command if we can't get custom fields
		customFields = nil
	}

	// Create a detailed project structure
	detailedProject := struct {
		*youtrack.Project
		CustomFields interface{} `json:"customFields,omitempty"`
	}{
		Project:      project,
		CustomFields: customFields,
	}
	return outputResult(detailedProject, formatProjectDetails)
}

// fetchAllProjects retrieves all projects from YouTrack
func fetchAllProjects(client *youtrack.Client, ctx *youtrack.YouTrackContext) ([]*youtrack.Project, error) {
	var allProjects []*youtrack.Project
	skip := 0
	top := 100 // Fetch 100 projects at a time

	for {
		projects, err := client.ListProjects(ctx, skip, top)
		if err != nil {
			return nil, err
		}

		allProjects = append(allProjects, projects...)

		// If we got fewer projects than requested, we've reached the end
		if len(projects) < top {
			break
		}

		skip += top
	}

	return allProjects, nil
}

// fetchProjectCustomFields retrieves custom fields for a specific project
func fetchProjectCustomFields(client *youtrack.Client, ctx *youtrack.YouTrackContext, projectID string) (interface{}, error) {
	// For now, we'll make a direct API call to get custom fields
	// This can be enhanced later with proper typed responses
	path := fmt.Sprintf("/api/admin/projects/%s/customFields", projectID)
	query := url.Values{}
	query.Add("fields", "id,field(id,name,fieldType(id)),canBeEmpty,emptyFieldText")

	resp, err := client.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var customFields interface{}
	if err := json.NewDecoder(resp.Body).Decode(&customFields); err != nil {
		return nil, fmt.Errorf("failed to decode custom fields: %w", err)
	}

	return customFields, nil
}

// filterProjects filters projects based on a search query
func filterProjects(projects []*youtrack.Project, query string) []*youtrack.Project {
	var filtered []*youtrack.Project
	queryLower := strings.ToLower(query)

	for _, project := range projects {
		// Search in ID, name, short name, and description
		if strings.Contains(strings.ToLower(project.ID), queryLower) ||
			strings.Contains(strings.ToLower(project.Name), queryLower) ||
			strings.Contains(strings.ToLower(project.ShortName), queryLower) ||
			strings.Contains(strings.ToLower(project.Description), queryLower) {
			filtered = append(filtered, project)
		}
	}

	return filtered
}

// formatProjectsList formats projects list for text output
func formatProjectsList(projects []*youtrack.Project) error {
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	cellStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("246"))
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			return cellStyle
		}).
		Headers("ID", "NAME", "SHORT NAME", "DESCRIPTION")

	for _, project := range projects {
		description := project.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}
		t.Row(
			project.ID,
			project.Name,
			project.ShortName,
			description,
		)
	}

	fmt.Println(t)
	return nil
}

// formatProjectDetails formats project details for text output
func formatProjectDetails(data interface{}) error {
	// Type assertion to access the detailed project structure
	detailedProject := data.(struct {
		*youtrack.Project
		CustomFields interface{} `json:"customFields,omitempty"`
	})

	fmt.Printf("Project Details\n")
	fmt.Printf("===============\n\n")

	fmt.Printf("ID:          %s\n", detailedProject.ID)
	fmt.Printf("Name:        %s\n", detailedProject.Name)
	fmt.Printf("Short Name:  %s\n", detailedProject.ShortName)

	if detailedProject.Description != "" {
		fmt.Printf("Description: %s\n", detailedProject.Description)
	}

	// Display custom fields if available
	if detailedProject.CustomFields != nil {
		fmt.Printf("\nCustom Fields\n")
		fmt.Printf("─────────────\n")

		// Handle the custom fields as a slice of interfaces
		if fields, ok := detailedProject.CustomFields.([]interface{}); ok {
			for _, field := range fields {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if fieldInfo, ok := fieldMap["field"].(map[string]interface{}); ok {
						fieldName := fieldInfo["name"]
						fieldType := "unknown"
						if ft, ok := fieldInfo["fieldType"].(map[string]interface{}); ok {
							if id, ok := ft["id"].(string); ok {
								fieldType = id
							}
						}
						fmt.Printf("- %v (%s)\n", fieldName, fieldType)
					}
				}
			}
		} else {
			// If we can't parse the structure, just print the raw JSON
			jsonBytes, _ := json.MarshalIndent(detailedProject.CustomFields, "", "  ")
			fmt.Printf("%s\n", jsonBytes)
		}
	}

	return nil
}
