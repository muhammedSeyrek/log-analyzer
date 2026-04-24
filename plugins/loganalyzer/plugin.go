package loganalyzer

import (
	"log-analyzer/internal/plugin"

	"github.com/spf13/cobra"
)

type LogAnalyzerPlugin struct{}

func init() {
	plugin.Register(&LogAnalyzerPlugin{})
}

func (p *LogAnalyzerPlugin) Name() string {
	return "loganalyzer"
}

func (p *LogAnalyzerPlugin) Description() string {
	return "Analyzes system logs and detects security threats"
}

func (p *LogAnalyzerPlugin) Command() *cobra.Command {
	return newRootCommand()
}
