package streamprocessor

import (
	"log-analyzer/internal/plugin"

	"github.com/spf13/cobra"
)

type StreamProcessorPlugin struct{}

func init() {
	plugin.Register(&StreamProcessorPlugin{})
}

func (p *StreamProcessorPlugin) Name() string {
	return "streamprocessor"
}

func (p *StreamProcessorPlugin) Description() string {
	return "Real-time stream parser/filter with bounded memory"
}

func (p *StreamProcessorPlugin) Command() *cobra.Command {
	return newRootCommand()
}
