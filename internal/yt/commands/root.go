package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/mkozhukh/youtrack/internal/yt/commands/tickets"
)

var (
	cfgFile string
	verbose bool
	output  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yt",
	Short: "YouTrack CLI - Command-line interface for YouTrack",
	Long: `yt is a command-line interface (CLI) tool for interacting with a remote 
YouTrack instance. It allows users to perform common YouTrack operations 
directly from their terminal.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set log level based on verbose flag
		if verbose {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(tickets.TicketsCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(completionCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/yt/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "output format (text, json)")
}

// Helper function to output in the requested format
func outputResult(data interface{}, formatAsText func(interface{}) error) error {
	switch output {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	case "text":
		return formatAsText(data)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
