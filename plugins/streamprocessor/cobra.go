package streamprocessor

import "github.com/spf13/cobra"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "streamprocessor",
		Short: "Real-time, low-memory stream processing module",
		Long: `Parses and filters a high-volume data stream (logs / dummy traffic)
in real time, without buffering it all into RAM.

Reads from stdin by default:
  cat dummy.log | log-analyzer streamprocessor run`,
	}
	cmd.AddCommand(newRunCommand())
	return cmd
}

func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Process a stream from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("streamprocessor: skeleton OK (logic comes in step 2)")
			return nil
		},
	}
}
