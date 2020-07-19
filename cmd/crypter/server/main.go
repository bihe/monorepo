package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/crypter/server"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
)

func main() {
	if err := server.Run(Version, Build); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}
