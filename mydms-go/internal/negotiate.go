package internal

import (
	"github.com/labstack/echo/v4"
	"github.com/markusthoemmes/goautoneg"
)

// Content defines a content-type
type Content int

const (
	// TEXT content-type requested by client
	TEXT Content = iota
	// JSON content-type requested by client
	JSON
	// HTML content-type requested by cleint
	HTML
)

// NegotiateContent returns a suitable content-type based on the Accept header
func NegotiateContent(c echo.Context) Content {
	header := c.Request().Header.Get("Accept")
	if header == "" {
		return JSON // default
	}

	accept := goautoneg.ParseAccept(header)
	if len(accept) == 0 {
		return JSON // default
	}

	// use the first element, because this has the highest priority
	switch accept[0].SubType {
	case "html":
		return HTML
	case "json":
		return JSON
	case "plain":
		return TEXT
	default:
		return JSON
	}
}
