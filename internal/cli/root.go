package cli

import (
	"fmt"
	"os"

	"log-analyzer/internal/plugin"

	"github.com/spf13/cobra"
)

var version = "dev"

var verbose bool

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log-analyzer",
		Short: "Modular log analysis and threat detection tool",
		Long: `log-analyzer is a plugin-based security tool. The core system
manages command-line arguments and automatically detects modules
added to the plugins/ directory, integrating them into the CLI.

To see available plugins:
  log-analyzer list`,
		Version: version,

		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"enable verbose output")

	return cmd
}

func bindPlugins(root *cobra.Command) {
	for _, p := range plugin.All() {
		root.AddCommand(p.Command())
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered plugins",
		Run: func(cmd *cobra.Command, args []string) {
			plugins := plugin.All()
			if len(plugins) == 0 {
				fmt.Println("No plugins registered.")
				return
			}
			fmt.Printf("Registered plugins: %d\n\n", len(plugins))
			for _, p := range plugins {
				fmt.Printf("  - %-20s %s\n", p.Name(), p.Description())
			}
		},
	}
}

func Execute() {
	root := newRootCommand()
	root.AddCommand(newListCommand())
	bindPlugins(root)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func IsVerbose() bool {
	return verbose
}
