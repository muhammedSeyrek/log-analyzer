package plugin

import "github.com/spf13/cobra"

type Plugin interface {
	Name() string

	Description() string

	Command() *cobra.Command
}
