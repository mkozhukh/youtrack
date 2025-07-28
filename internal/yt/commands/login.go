package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mkozhukh/youtrack/internal/yt/config"
	"github.com/mkozhukh/youtrack/pkg/youtrack"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactively configure YouTrack authentication",
	Long: `Interactively prompts the user for the YouTrack URL and a permanent token, 
then saves them to the configuration file. It will also attempt to automatically 
determine and save the user's own YouTrack user ID, which enables commands to 
default to the current user.`,
	RunE: runLogin,
}

func init() {
	// Command initialization
}

func runLogin(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for YouTrack URL
	fmt.Print("YouTrack URL (e.g., https://youtrack.example.com): ")
	url, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read URL: %w", err)
	}
	url = strings.TrimSpace(url)

	// Validate URL
	if url == "" {
		return fmt.Errorf("YouTrack URL cannot be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("YouTrack URL must start with http:// or https://")
	}

	// Prompt for permanent token
	fmt.Print("Permanent token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	fmt.Println() // Add newline after password input
	token := strings.TrimSpace(string(tokenBytes))

	if token == "" {
		return fmt.Errorf("permanent token cannot be empty")
	}

	// Test the connection
	fmt.Println("Testing connection...")
	client := youtrack.NewClient(url)
	ctx := youtrack.NewYouTrackContext(context.Background(), token)

	// Try to get the current user
	userID, err := getCurrentUserID(client, ctx)
	if err != nil {
		log.Error("Failed to connect to YouTrack", "error", err)
		return fmt.Errorf("failed to verify connection: %w", err)
	}

	// Prompt for default project (optional)
	fmt.Print("Default project (optional, press Enter to skip): ")
	project, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read project: %w", err)
	}
	project = strings.TrimSpace(project)

	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			URL:   url,
			Token: token,
		},
		Defaults: config.DefaultsConfig{
			UserID:  userID,
			Project: project,
		},
	}

	// Save config
	configPath := cfgFile
	if configPath == "" {
		configPath = config.GetConfigPath()
	}

	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Successfully logged in!\n")
	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Printf("Your user ID: %s\n", userID)

	return nil
}

// getCurrentUserID fetches the current user's ID from YouTrack
func getCurrentUserID(client *youtrack.Client, ctx *youtrack.YouTrackContext) (string, error) {
	// YouTrack API supports getting current user with "me" as the user ID
	user, err := client.GetUser(ctx, "me")
	if err != nil {
		return "", err
	}

	log.Info("Successfully authenticated", "user", user.Login, "id", user.ID)
	return user.ID, nil
}
