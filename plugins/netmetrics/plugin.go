package netmetrics

import (
	"log-analyzer/internal/plugin"

	"github.com/spf13/cobra"
)

type NetMetricsPlugin struct{}

func init() {
	plugin.Register(&NetMetricsPlugin{})
}

func (p *NetMetricsPlugin) Name() string {
	return "netmetrics"
}

func (p *NetMetricsPlugin) Description() string {
	return "Captures network packets with Scapy and pushes metrics over UDP"
}

func (p *NetMetricsPlugin) Command() *cobra.Command {
	return newRootCommand()
}
