package netmetrics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed sniffer.py
var snifferPy []byte

func newCaptureCommand() *cobra.Command {
	var (
		iface      string
		bpfFilter  string
		remoteHost string
		remotePort int
		interval   float64
		agentID    string
		topN       int
		python     string
		scriptPath string
	)

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Sniff packets with Scapy and push metrics over UDP",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := scriptPath
			if path == "" {
				tmp, err := os.CreateTemp("", "netmetrics-sniffer-*.py")
				if err != nil {
					return fmt.Errorf("gecici betik olusturulamadi: %w", err)
				}
				defer os.Remove(tmp.Name())
				if _, err := tmp.Write(snifferPy); err != nil {
					tmp.Close()
					return fmt.Errorf("betik yazilamadi: %w", err)
				}
				tmp.Close()
				path = tmp.Name()
			}

			pyArgs := []string{
				path,
				"--remote-host", remoteHost,
				"--remote-port", strconv.Itoa(remotePort),
				"--interval", strconv.FormatFloat(interval, 'f', -1, 64),
				"--agent-id", agentID,
				"--top-n", strconv.Itoa(topN),
			}
			if iface != "" {
				pyArgs = append(pyArgs, "--iface", iface)
			}
			if bpfFilter != "" {
				pyArgs = append(pyArgs, "--filter", bpfFilter)
			}

			ctx, stop := signal.NotifyContext(
				context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			proc := exec.CommandContext(ctx, python, pyArgs...)
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr
			proc.Stdin = os.Stdin

			if err := proc.Run(); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("scapy ajani basarisiz: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&iface, "iface", "", "Interface to sniff (empty = auto)")
	cmd.Flags().StringVar(&bpfFilter, "filter", "", "BPF filter, e.g. 'tcp or udp'")
	cmd.Flags().StringVar(&remoteHost, "remote-host", "127.0.0.1", "Remote collector host")
	cmd.Flags().IntVar(&remotePort, "remote-port", 9999, "Remote collector UDP port")
	cmd.Flags().Float64Var(&interval, "interval", 5, "Push interval in seconds")
	cmd.Flags().StringVar(&agentID, "agent-id", hostnameOr("sniffer"), "Agent identifier")
	cmd.Flags().IntVar(&topN, "top-n", 5, "Top-talker list size")
	cmd.Flags().StringVar(&python, "python", pythonDefault(), "Python interpreter to use (env: NETMETRICS_PYTHON)")
	cmd.Flags().StringVar(&scriptPath, "script", "", "Override path to sniffer.py (default: embedded)")

	return cmd
}

func pythonDefault() string {
	if v := os.Getenv("NETMETRICS_PYTHON"); v != "" {
		return v
	}
	return "python3"
}

func hostnameOr(fallback string) string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return fallback
}
