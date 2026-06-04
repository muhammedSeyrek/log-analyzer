package streamprocessor

import (
	"fmt"
	"os"
	"time"

	"log-analyzer/internal/config"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "streamprocessor",
		Short: "Real-time, low-memory stream processing module",
		Long: `Parses and filters a high-volume data stream (logs / dummy traffic)
in real time, without buffering it all into RAM.

Reads from stdin by default:
  cat dummy.log | log-analyzer streamprocessor run
  ./traffic-gen | log-analyzer streamprocessor run --mode stream`,
	}
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newGenCommand())
	return cmd
}

func newRunCommand() *cobra.Command {
	var (
		configPath  string
		mode        string
		extraFilter string
		workers     int
		bufSize     int
		quiet       bool
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Process a stream from stdin",
		Long: `Reads a data stream from stdin, parses and filters each line against
the rules in the config file (plus an optional --filter term).

By default it prints matching lines and a small summary. Use --verbose for the
full report: throughput, live memory, and a breakdown of how many times each
rule and each rule type fired.

Modes:
  stream    bounded channel + worker pool, constant memory (default)
  buffered  accumulates all matches in RAM until the end (for comparison)

Examples:
  cat dummy.log | log-analyzer streamprocessor run
  cat dummy.log | log-analyzer streamprocessor run --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var (
				stats   *Stats
				elapsed time.Duration
			)

			switch mode {
			case "stream":
				stats, elapsed, err = runStream(os.Stdin, os.Stdout, cfg.Rules, extraFilter, workers, bufSize, quiet)
			case "buffered":
				stats, elapsed, err = runBuffered(os.Stdin, os.Stdout, cfg.Rules, extraFilter, quiet)
			default:
				return fmt.Errorf("unknown mode %q (use 'stream' or 'buffered')", mode)
			}
			if err != nil {
				return fmt.Errorf("processing error: %w", err)
			}

			if verbose {
				printVerbose(cmd, mode, workers, bufSize, stats, elapsed)
			} else {
				printSimple(cmd, stats)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "config/rules.yaml", "path to rules config file")
	cmd.Flags().StringVarP(&mode, "mode", "m", "stream", "processing mode: stream | buffered")
	cmd.Flags().StringVarP(&extraFilter, "filter", "f", "", "extra substring to match (optional)")
	cmd.Flags().IntVarP(&workers, "workers", "w", 1, "number of filter workers (stream mode)")
	cmd.Flags().IntVarP(&bufSize, "buffer", "b", 1000, "bounded channel capacity (stream mode)")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "suppress per-match output")
	cmd.Flags().BoolVarP(&verbose, "verbose", "V", false, "show full report (memory, throughput, rule/type breakdown)")

	return cmd
}

func newGenCommand() *cobra.Command {
	var (
		count      int64
		rate       int
		matchRatio float64
	)

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate synthetic log traffic to stdout",
		Long: `Emits synthetic, mixed log traffic (rule-matching lines + harmless noise)
to stdout. Pipe it straight into 'run' to demo real-time processing.

Examples:
  log-analyzer streamprocessor gen --count 1000000 | log-analyzer streamprocessor run -q -V
  log-analyzer streamprocessor gen --rate 5 | log-analyzer streamprocessor run
  log-analyzer streamprocessor gen --count 0 | log-analyzer streamprocessor run -q`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if matchRatio < 0 || matchRatio > 1 {
				return fmt.Errorf("--match-ratio must be between 0 and 1")
			}
			emitted, elapsed, err := runGenerator(os.Stdout, count, rate, matchRatio)
			if err != nil {
				return fmt.Errorf("generator error: %w", err)
			}
			fmt.Fprintf(os.Stderr, "[gen] emitted %d lines in %.3fs\n", emitted, elapsed.Seconds())
			return nil
		},
	}

	cmd.Flags().Int64VarP(&count, "count", "n", 100000, "lines to emit (0 = infinite)")
	cmd.Flags().IntVarP(&rate, "rate", "r", 0, "target lines/second (0 = full speed)")
	cmd.Flags().Float64Var(&matchRatio, "match-ratio", 0.3, "fraction of lines that match a rule (0..1)")

	return cmd
}

// printSimple is the default: just the bottom-line counts.
func printSimple(cmd *cobra.Command, s *Stats) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\nlines read: %d   matched: %d\n", s.Lines, s.Matched)
}

// printVerbose is the full diagnostic report behind --verbose.
func printVerbose(cmd *cobra.Command, mode string, workers, bufSize int, s *Stats, elapsed time.Duration) {
	secs := elapsed.Seconds()
	var rate float64
	if secs > 0 {
		rate = float64(s.Lines) / secs
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "\n════════════ verbose report ════════════")
	fmt.Fprintf(out, "mode:         %s\n", mode)
	if mode == "stream" {
		fmt.Fprintf(out, "workers:      %d\n", workers)
		fmt.Fprintf(out, "buffer cap:   %d\n", bufSize)
	}
	fmt.Fprintf(out, "lines read:   %d\n", s.Lines)
	fmt.Fprintf(out, "matched:      %d\n", s.Matched)
	fmt.Fprintf(out, "bytes read:   %.2f MB\n", float64(s.Bytes)/(1024*1024))
	fmt.Fprintf(out, "elapsed:      %.3f s\n", secs)
	fmt.Fprintf(out, "throughput:   %.0f lines/s\n", rate)
	fmt.Fprintf(out, "heap in use:  %.2f MB\n", heapAllocMB())

	fmt.Fprintln(out, "\n── matches by rule ──")
	if len(s.PerRule) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, name := range sortedKeys(s.PerRule) {
			fmt.Fprintf(out, "  %-32s %d\n", name, s.PerRule[name])
		}
	}

	fmt.Fprintln(out, "\n── matches by type ──")
	if len(s.PerType) == 0 {
		fmt.Fprintln(out, "  (none)")
	} else {
		for _, t := range sortedKeys(s.PerType) {
			fmt.Fprintf(out, "  %-32s %d\n", t, s.PerType[t])
		}
	}
	fmt.Fprintln(out, "═════════════════════════════════════════")
}
