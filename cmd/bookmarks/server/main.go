package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/bookmarks"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "bookmarks"
)

func main() {
	if err := bookmarks.Run(Version, Build, AppName); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}
