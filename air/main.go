package main

import (
	"fmt"
	"os"

	"github.com/cloud-byte/air/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "air:", err)
		os.Exit(1)
	}
}
