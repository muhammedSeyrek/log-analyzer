package loganalyzer

import (
	"fmt"

	"log-analyzer/internal/analyzer"
	"log-analyzer/internal/config"

	"github.com/spf13/cobra"
)

func newLiveCommand() *cobra.Command {
	var (
		configPath string
		targetFile string
	)

	cmd := &cobra.Command{
		Use:   "live",
		Short: "Monitor logs in real time (stream/tailing mode)",
		Long: `Watches a log file or system log stream in real time and emits
alerts as rule matches are detected. Runs until interrupted (Ctrl+C).

Examples:
  log-analyzer loganalyzer live
  log-analyzer loganalyzer live --file /var/log/auth.log`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			file := targetFile
			if file == "" {
				if len(cfg.Files) == 0 {
					return fmt.Errorf("no files specified in config and --file not provided")
				}
				file = cfg.Files[0]
			}

			return runLive(file, cfg)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "config/rules.yaml",
		"path to rules config file")
	cmd.Flags().StringVarP(&targetFile, "file", "f", "",
		"file to watch (overrides config)")

	return cmd
}

func runLive(file string, cfg *config.Config) error {
	fmt.Printf("-> Starting live monitor on: %s\n", file)
	fmt.Println("Press Ctrl+C to stop.")

	if err := analyzer.WatchFile(file, cfg.Rules); err != nil {
		return fmt.Errorf("live monitoring error: %w", err)
	}
	return nil
}
