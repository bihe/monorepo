package server

import "fmt"

// PrintServerBanner put some nice emojis on the console
func PrintServerBanner(name, version, build, env, addr string) {
	fmt.Printf("%s Starting server '%s'\n", "🚀", name)
	fmt.Printf("%s Version: '%s-%s'\n", "🔖", version, build)
	fmt.Printf("%s Environment: '%s'\n", "🌍", env)
	fmt.Printf("%s Listening on '%s'\n", "💻", addr)
	fmt.Printf("%s Ready!\n", "🏁")
}
