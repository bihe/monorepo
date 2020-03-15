package cookies // import "golang.binggl.net/commons/cookies"

// Settings defines parameters for cookies used for HTML-based errors
type Settings struct {
	Path   string
	Domain string
	Secure bool
	Prefix string
}
