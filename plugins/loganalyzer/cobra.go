package loganalyzer

import (
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loganalyzer",
		Short: "Log analysis and threat detection tool",
		Long: `Analyzes system logs (Windows Event Logs, Linux Syslog/Journald,
macOS Logs) to detect potential security threats.

Supports three modes:
  static       Analyze log files in batch (non-interactive)
  live         Monitor logs in real time
  interactive  Classic menu-driven mode`,
	}

	cmd.AddCommand(newStaticCommand())
	cmd.AddCommand(newLiveCommand())
	cmd.AddCommand(newInteractiveCommand())

	return cmd
}
