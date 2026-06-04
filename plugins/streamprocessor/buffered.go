package streamprocessor

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"log-analyzer/internal/config"
)

// runBuffered is the DELIBERATELY naive version, kept for comparison.
//
// It mirrors the original loganalyzer approach: every matching line is
// appended to an in-memory slice and only written out at the very end.
// Memory grows linearly with the number of matches — on a large or infinite
// stream this is exactly the RAM-bloat the task asks us to avoid.
//
// Run the same input through `--mode buffered` and `--mode stream` and compare
// the reported HeapAlloc: buffered climbs, stream stays flat.
//
// Single goroutine here, so it writes the breakdown maps directly.
func runBuffered(
	r io.Reader,
	out io.Writer,
	rules []config.Rule,
	extraFilter string,
	quiet bool,
) (*Stats, time.Duration, error) {

	stats := newStats()
	var collected []string // <- this is what grows without bound

	start := time.Now()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		stats.Lines++
		stats.Bytes += int64(len(line)) + 1

		if name, rtype, ok := matchLine(line, rules, extraFilter); ok {
			stats.Matched++
			stats.PerRule[name]++
			stats.PerType[rtype]++
			// Keep everything in memory until the stream ends.
			collected = append(collected, fmt.Sprintf("[MATCH:%s] %s", name, line))
		}
	}
	if err := scanner.Err(); err != nil {
		return stats, time.Since(start), err
	}

	// Only now do we flush — after holding all matches in RAM.
	if !quiet {
		for _, c := range collected {
			fmt.Fprintln(out, c)
		}
	}

	return stats, time.Since(start), nil
}
