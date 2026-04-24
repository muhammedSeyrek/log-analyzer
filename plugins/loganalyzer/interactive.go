package loganalyzer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"log-analyzer/internal/config"

	"github.com/spf13/cobra"
)

func newInteractiveCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Run in interactive menu mode (classic UX)",
		Long: `Starts a menu-driven interface where the user chooses between
static analysis and live monitoring. Useful for manual, exploratory
usage. For scripted/automated runs, prefer 'static' or 'live'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return runInteractive(cfg)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "config/rules.yaml",
		"path to rules config file")

	return cmd
}

func runInteractive(cfg *config.Config) error {
	fmt.Println("Log analysis project started (interactive mode).")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n1 - Start Static Analysis (File Based)\n")
		fmt.Print("2 - Start Live Monitoring (Stream Based or Tailing)\n")
		fmt.Print("3 - Exit\n")
		fmt.Print("Choose an option: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			if err := interactiveStatic(cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "2":
			if err := interactiveLive(cfg); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "3":
			fmt.Println("Exiting...")
			return nil
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

func interactiveStatic(cfg *config.Config) error {
	results, err := runStatic(cfg.Files, cfg)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("Nothing to save.")
	} else {
		fmt.Println("\nDo you want to save results to a CSV report? (y/n):")
		var choice string
		fmt.Scanln(&choice)

		if strings.ToLower(strings.TrimSpace(choice)) == "y" {
			if err := writeCSVReport(results, "reports"); err != nil {
				return err
			}
		}
	}

	fmt.Println("Press enter to return to the main menu.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	return nil
}

func interactiveLive(cfg *config.Config) error {
	if len(cfg.Files) == 0 {
		return fmt.Errorf("no files specified for live monitoring in the config")
	}
	return runLive(cfg.Files[0], cfg)
}
