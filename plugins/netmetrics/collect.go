package netmetrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type receivedPayload struct {
	AgentID   string  `json:"agent_id"`
	Timestamp float64 `json:"timestamp"`
	Interval  float64 `json:"interval"`
	Metrics   struct {
		PacketCapture struct {
			Window struct {
				Packets   int            `json:"packets"`
				Bytes     int            `json:"bytes"`
				Protocols map[string]int `json:"protocols"`
			} `json:"window"`
			Lifetime struct {
				Packets int `json:"packets"`
				Bytes   int `json:"bytes"`
			} `json:"lifetime"`
		} `json:"packet_capture"`
	} `json:"metrics"`
}

func newCollectCommand() *cobra.Command {
	var (
		bindHost     string
		bindPort     int
		slackWebhook string
		slackMinPkts int
	)

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Receive UDP metrics and print them; optionally forward to Slack",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := net.UDPAddr{IP: net.ParseIP(bindHost), Port: bindPort}
			conn, err := net.ListenUDP("udp", &addr)
			if err != nil {
				return fmt.Errorf("UDP dinlenemedi: %w", err)
			}
			defer conn.Close()

			fmt.Printf("[collector] %s:%d dinleniyor", bindHost, bindPort)
			if slackWebhook != "" {
				fmt.Printf(" (Slack yonlendirme acik)")
			}
			fmt.Println()

			ctx, stop := signal.NotifyContext(
				context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			go func() {
				<-ctx.Done()
				conn.SetReadDeadline(time.Now())
			}()

			buf := make([]byte, 65535)
			client := &http.Client{Timeout: 5 * time.Second}

			for {
				n, src, err := conn.ReadFromUDP(buf)
				if err != nil {
					if ctx.Err() != nil {
						fmt.Println("[collector] durduruldu.")
						return nil
					}
					fmt.Printf("[collector] okuma hatasi: %v\n", err)
					continue
				}

				var p receivedPayload
				if err := json.Unmarshal(buf[:n], &p); err != nil {
					continue
				}

				w := p.Metrics.PacketCapture.Window
				fmt.Printf("[collector] %s (%s) -> %d paket / %d bayt | protokoller: %s\n",
					src.IP, p.AgentID, w.Packets, w.Bytes, formatProtocols(w.Protocols))

				if slackWebhook != "" && w.Packets >= slackMinPkts {
					if err := postToSlack(client, slackWebhook, src.IP.String(), &p); err != nil {
						fmt.Printf("[collector] Slack gonderim hatasi: %v\n", err)
					}
				}
			}
		},
	}

	cmd.Flags().StringVar(&bindHost, "bind-host", "0.0.0.0", "Bind address")
	cmd.Flags().IntVar(&bindPort, "bind-port", 9999, "UDP port to listen on")
	cmd.Flags().StringVar(&slackWebhook, "slack-webhook", os.Getenv("SLACK_WEBHOOK_URL"),
		"Slack Incoming Webhook URL (env: SLACK_WEBHOOK_URL)")
	cmd.Flags().IntVar(&slackMinPkts, "slack-min-packets", 1,
		"Only post to Slack when window packets >= this value")

	return cmd
}

func formatProtocols(m map[string]int) string {
	if len(m) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(parts, " ")
}

func postToSlack(client *http.Client, webhook, srcIP string, p *receivedPayload) error {
	w := p.Metrics.PacketCapture.Window
	life := p.Metrics.PacketCapture.Lifetime
	text := fmt.Sprintf(
		":satellite: *netmetrics* — agent `%s` (%s)\n"+
			"• son %.0fs: *%d paket* / %d bayt\n"+
			"• protokoller: %s\n"+
			"• toplam: %d paket / %d bayt",
		p.AgentID, srcIP, p.Interval, w.Packets, w.Bytes,
		formatProtocols(w.Protocols), life.Packets, life.Bytes,
	)

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack HTTP %d", resp.StatusCode)
	}
	return nil
}
