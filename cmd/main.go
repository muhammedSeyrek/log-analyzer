package main

import (
	"bufio"
	"fmt"
	"log"
	"log-analyzer/internal/analyzer"
	"log-analyzer/internal/config" // We are importing the package we wrote ourselves.
	"log-analyzer/internal/reporter"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	fmt.Println("Is Started log analysis project.")

	cfg, err := config.LoadConfig("config/rules.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("1 - Start Static Analysis (File Based)\n")
		fmt.Print("2 - Start Live Monitoring (Stream Based or Tailing)\n")
		fmt.Print("3 - Exit\n")
		fmt.Print("Choose an option: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			runStaticAnalysis(cfg)
		case "2":
			runLiveMonitoring(cfg)
		case "3":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

func runStaticAnalysis(cfg *config.Config) {

	fmt.Println("\n Analysis is being initiated.")

	var allResults []analyzer.Result

	for i := 0; i < len(cfg.Files); i++ {

		if cfg.Files[i] == "SECURITY_LOGS" {
			fmt.Println("\n>>> SYSTEM SCAN: Reading Windows Event Logs (via API/PowerShell)...")
		} else {
			fmt.Printf("\nScanning file: %s\n", cfg.Files[i])
		}

		results, err := analyzer.AnalyzeFile(cfg.Files[i], cfg.Rules)
		if err != nil {
			log.Printf("Error analyzing file %s: %v", cfg.Files[i], err)
			continue
		}

		if len(results) > 0 {
			fmt.Printf("%d Threats found !!!\n", len(results))
			allResults = append(allResults, results...)
			for i := 0; i < len(results); i++ {
				fmt.Printf("[%s] %s -> %s\n", results[i].RuleType, results[i].RuleName, results[i].Line)
			}
		} else {
			fmt.Println("No threats found in this file.")
		}
	}

	if len(allResults) > 0 {
		fmt.Println("\nDo you want to save results to 'report.csv'? (y/n):")

		var choice string
		fmt.Scanln(&choice)

		if strings.ToLower(choice) == "y" {

			reportDir := "reports"

			if err := os.MkdirAll(reportDir, 0755); err != nil {
				fmt.Printf("Error creating report directory: %v\n", err)
				return
			}

			Timestamp := time.Now().Format("2006-01-02_15-04-05")
			fileName := fmt.Sprintf("report_%s.csv", Timestamp)
			fullPath := filepath.Join(reportDir, fileName)
			err := reporter.GenerateCSVReport(allResults, fullPath)
			if err != nil {
				fmt.Printf("Error generating CSV report: %v\n", err)
			} else {
				fmt.Printf("Report saved to %s\n", fullPath)
			}
		}
	}

	fmt.Println("Static analysis completed.")
	fmt.Println("Press enter to return to the main menu.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}

func runLiveMonitoring(cfg *config.Config) {

	if len(cfg.Files) == 0 {
		fmt.Println("No files specified for live monitoring in the configuration.")
		return
	}

	targetFile := cfg.Files[0]
	fmt.Printf("\n -> Starting Live Monitor on: %s\n", targetFile)

	err := analyzer.WatchFile(targetFile, cfg.Rules)
	if err != nil {
		fmt.Printf("Error in live monitoring: %v\n", err)
	}

}
