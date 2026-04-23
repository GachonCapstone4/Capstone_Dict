package main

import (
	"os"

	"capstone_network_test/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
