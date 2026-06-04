package streamprocessor

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log-analyzer/internal/config"
)

// Stats holds the results of a processing run.
//
// Lines/Matched/Bytes are plain int64 counters. The two maps are breakdowns
// shown only in --verbose mode: how many times each named rule fired, and how
// many matches fell into each rule type (security/critical/warning/...).
//
// IMPORTANT: in stream mode several workers run concurrently. We never write
// to these shared maps from multiple goroutines (that would panic). Instead
// each worker keeps a LOCAL counter set and we merge them once at the end.
type Stats struct {
	Lines   int64
	Matched int64
	Bytes   int64

	PerRule map[string]int64 // rule name -> match count
	PerType map[string]int64 // rule type -> match count
}

func newStats() *Stats {
	return &Stats{
		PerRule: make(map[string]int64),
		PerType: make(map[string]int64),
	}
}

// localCount is a worker's private tally — no locks needed because only that
// one goroutine touches it.
type localCount struct {
	matched int64
	perRule map[string]int64
	perType map[string]int64
}

func newLocalCount() *localCount {
	return &localCount{
		perRule: make(map[string]int64),
		perType: make(map[string]int64),
	}
}

// runStream is the streaming pipeline.
//
//	reader goroutine  ->  bounded channel (cap = bufSize)  ->  N worker goroutines
//
// Memory stays flat because:
//  1. bufio.Scanner hands us one line at a time; the full stream is never in RAM.
//  2. The channel is bounded, so if workers fall behind the reader BLOCKS
//     (backpressure) instead of letting an unbounded queue grow.
//  3. Matches are written to `out` immediately and discarded; nothing is kept.
func runStream(
	r io.Reader,
	out io.Writer,
	rules []config.Rule,
	extraFilter string,
	workers int,
	bufSize int,
	quiet bool,
) (*Stats, time.Duration, error) {

	if workers < 1 {
		workers = 1
	}
	if bufSize < 1 {
		bufSize = 1000
	}

	stats := newStats()
	lines := make(chan string, bufSize) // <- the bounded buffer
	var outMu sync.Mutex                // serialize writes to `out`
	var wg sync.WaitGroup

	locals := make([]*localCount, workers) // one tally per worker, merged later

	start := time.Now()

	// --- worker pool: parse + filter, write matches immediately ---
	for w := 0; w < workers; w++ {
		lc := newLocalCount()
		locals[w] = lc
		wg.Add(1)
		go func(lc *localCount) {
			defer wg.Done()
			for line := range lines {
				if name, rtype, ok := matchLine(line, rules, extraFilter); ok {
					lc.matched++
					lc.perRule[name]++
					lc.perType[rtype]++
					if !quiet {
						outMu.Lock()
						fmt.Fprintf(out, "[MATCH:%s] %s\n", name, line)
						outMu.Unlock()
					}
				}
			}
		}(lc)
	}

	// --- reader: one line at a time, feed the channel ---
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // up to 1MB lines
	var scanErr error
	for scanner.Scan() {
		line := scanner.Text()
		atomic.AddInt64(&stats.Lines, 1)
		atomic.AddInt64(&stats.Bytes, int64(len(line))+1)
		lines <- line // blocks here when the buffer is full -> backpressure
	}
	scanErr = scanner.Err()

	close(lines) // no more input; workers drain remaining items then exit
	wg.Wait()

	// merge all worker-local tallies into the shared Stats (single goroutine now)
	for _, lc := range locals {
		stats.Matched += lc.matched
		for k, v := range lc.perRule {
			stats.PerRule[k] += v
		}
		for k, v := range lc.perType {
			stats.PerType[k] += v
		}
	}

	return stats, time.Since(start), scanErr
}

// matchLine returns the name and type of the first matching rule (or
// "filter"/"filter" for the ad-hoc --filter term) and whether anything matched.
// Same case-insensitive substring logic the existing analyzer uses.
func matchLine(line string, rules []config.Rule, extraFilter string) (name, rtype string, ok bool) {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return "", "", false
	}
	if extraFilter != "" && strings.Contains(lower, strings.ToLower(extraFilter)) {
		return "filter", "filter", true
	}
	for i := range rules {
		if strings.Contains(lower, strings.ToLower(rules[i].Pattern)) {
			return rules[i].Name, rules[i].Type, true
		}
	}
	return "", "", false
}

// sortedKeys returns map keys sorted by descending count, then name, so the
// verbose breakdown is stable and readable.
func sortedKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys
}

// heapAllocMB reports current heap allocation in MB, for the proof readout.
func heapAllocMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1024 * 1024)
}
