package streamprocessor

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"time"
)

// genFunc builds one synthetic line given a timestamp and a RNG. Each template
// owns its own formatting, so there's never an arg/verb mismatch.
type genFunc func(ts string, rng *rand.Rand) string

// matching builders produce lines that trigger real rules in config/rules.yaml.
var matchingBuilders = []genFunc{
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server sshd[%d]: Failed password for invalid user admin from 10.0.%d.%d port 22",
			ts, r.Intn(9000)+1000, r.Intn(255), r.Intn(255))
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server sshd[%d]: Invalid user test from 192.168.%d.%d",
			ts, r.Intn(9000)+1000, r.Intn(255), r.Intn(255))
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server nginx: GET /admin.php 404", ts)
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server sudo:  user : TTY=pts/0 ; USER=root ; COMMAND=/bin/ls", ts)
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server kernel: UFW BLOCK IN=eth0 SRC=10.0.%d.%d", ts, r.Intn(255), r.Intn(255))
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server app: critical failure in module %d", ts, r.Intn(50))
	},
}

// noise builders produce harmless lines that match nothing — the bulk of traffic.
var noiseBuilders = []genFunc{
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server nginx: GET /index.html 200", ts)
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server systemd[%d]: Started Session %d of user deploy", ts, r.Intn(9000)+1000, r.Intn(500))
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server cron[%d]: job finished cleanly", ts, r.Intn(9000)+1000)
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server app: request handled in %dms", ts, r.Intn(500))
	},
	func(ts string, r *rand.Rand) string {
		return fmt.Sprintf("%s server kernel: usb device %d connected", ts, r.Intn(20))
	},
}

// runGenerator emits synthetic log traffic to `out`.
//
//	count       total lines to emit (0 = infinite, until interrupted)
//	rate        target lines/second (0 = unbounded, full speed)
//	matchRatio  fraction [0..1] of lines that should match a rule
//
// Output is buffered; the generator holds only one line at a time, so it
// streams at high volume without growing in RAM — just like the reader.
func runGenerator(out io.Writer, count int64, rate int, matchRatio float64) (int64, time.Duration, error) {
	w := bufio.NewWriterSize(out, 64*1024)
	defer w.Flush()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var interval time.Duration
	if rate > 0 {
		interval = time.Second / time.Duration(rate)
	}

	start := time.Now()
	var emitted int64

	for count <= 0 || emitted < count {
		ts := time.Now().Format("Jan 02 15:04:05")

		var line string
		if rng.Float64() < matchRatio {
			line = matchingBuilders[rng.Intn(len(matchingBuilders))](ts, rng)
		} else {
			line = noiseBuilders[rng.Intn(len(noiseBuilders))](ts, rng)
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			// downstream closed the pipe (e.g. `head`); stop cleanly
			return emitted, time.Since(start), nil
		}
		emitted++

		if interval > 0 {
			w.Flush() // flush so consumers see lines in real time when rate-limited
			time.Sleep(interval)
		}
	}

	return emitted, time.Since(start), nil
}
