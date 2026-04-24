package main

import (
	"log-analyzer/internal/cli"

	_ "log-analyzer/plugins"
)

func main() {
	cli.Execute()
}
