package loganalyzer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"log-analyzer/internal/analyzer"
	"log-analyzer/internal/config"
	"log-analyzer/internal/reporter"

	"github.com/spf13/cobra"
)

func newStaticCommand() *cobra.Command {
	var (
		configPath string
		singleFile string
		report     bool
		outputDir  string
	)

	cmd := &cobra.Command{
		Use:   "static",
		Short: "Run static analysis on log files (batch mode)",
		Long: `Scans historical log files for threats using the rules defined in
the config file. Designed to be scriptable — no interactive prompts.

Examples:
  log-analyzer loganalyzer static
  log-analyzer loganalyzer static --config config/rules.yaml --report
  log-analyzer loganalyzer static --file /var/log/auth.log --report`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// --file overrides the config's files list
			files := cfg.Files
			if singleFile != "" {
				files = []string{singleFile}
			}

			results, err := runStatic(files, cfg)
			if err != nil {
				return err
			}

			if report && len(results) > 0 {
				if err := writeCSVReport(results, outputDir); err != nil {
					return fmt.Errorf("report generation failed: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "config/rules.yaml",
		"path to rules config file")
	cmd.Flags().StringVarP(&singleFile, "file", "f", "",
		"analyze a single file (overrides config)")
	cmd.Flags().BoolVarP(&report, "report", "r", false,
		"generate a CSV report of the findings")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "reports",
		"directory for CSV reports")

	return cmd
}

func runStatic(files []string, cfg *config.Config) ([]analyzer.Result, error) {
	fmt.Println("Starting static analysis...")

	var allResults []analyzer.Result

	for _, file := range files {
		if file == "SECURITY_LOGS" || file == "SYSTEM" {
			fmt.Println("\n>>> SYSTEM SCAN: Reading system event logs...")
		} else {
			fmt.Printf("\nScanning file: %s\n", file)
		}

		results, err := analyzer.AnalyzeFile(file, cfg.Rules)
		if err != nil {
			log.Printf("Error analyzing %s: %v", file, err)
			continue
		}

		if len(results) > 0 {
			fmt.Printf("%d threats found!\n", len(results))
			allResults = append(allResults, results...)
			for _, r := range results {
				fmt.Printf("[%s] %s -> %s\n", r.RuleType, r.RuleName, r.Line)
			}
		} else {
			fmt.Println("No threats found in this file.")
		}
	}

	fmt.Println("\nStatic analysis completed.")
	return allResults, nil
}

func writeCSVReport(results []analyzer.Result, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("report_%s.csv", timestamp)
	fullPath := filepath.Join(outputDir, fileName)

	if err := reporter.GenerateCSVReport(results, fullPath); err != nil {
		return err
	}

	fmt.Printf("Report saved to %s\n", fullPath)
	return nil
}
