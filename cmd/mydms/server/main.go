package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/mydms"
)

var (
	// Version exports the application version
	Version = "3.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "mydms"
)

func main() {
	if err := mydms.Run(Version, Build, AppName); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}
