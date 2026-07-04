package netmetrics

import (
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "netmetrics",
		Short: "Network metrics capture (Scapy) and UDP collector",
		Long: `netmetrics captures network packets using Scapy and pushes the
collected metrics to a remote server over UDP.

Subcommands:
  capture   Sniff packets with Scapy and push metrics over UDP (agent side)
  collect   Receive UDP metrics and print them; optionally forward to Slack`,
	}

	cmd.AddCommand(newCaptureCommand())
	cmd.AddCommand(newCollectCommand())

	return cmd
}
