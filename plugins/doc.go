// Package plugins is the aggregation point for all plugin modules.
// Each plugin lives in its own subdirectory and registers itself via
// init(); this package imports them all via blank imports so their
// init() functions run when the binary starts.
//
// The list of imports is generated automatically by tools/gen. Do not
// edit plugins_gen.go by hand.
package plugins

//go:generate go run ../tools/gen/main.go
