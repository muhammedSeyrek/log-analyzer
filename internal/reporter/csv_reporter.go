package reporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"log-analyzer/internal/analyzer"
)

func GenerateCSVReport(results []analyzer.Result, outputPath string) error {

	// Create or truncate the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Timestamp", "RuleName", "RuleType", "Line"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for i := 0; i < len(results); i++ {
		record := []string{
			results[i].Timestamp.Format(time.RFC3339),
			results[i].RuleName,
			results[i].RuleType,
			results[i].Line,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	fmt.Printf("CSV report generated at %s\n", outputPath)
	return nil
}
