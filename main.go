package main

import (
	"cli-aio/cmd"
	"os"
)

// main is the entrypoint of the application.
// It delegates all CLI setup and execution to the cmd package,
// keeping this file minimal for fast startup.
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

