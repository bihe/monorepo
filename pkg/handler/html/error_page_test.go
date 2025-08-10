package html_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler/html"
)

func Test404ErrorPage(t *testing.T) {
	var outBuffer bytes.Buffer
	page := html.ErrorPage404("--basepath--", config.Development, "dev")
	if err := page.Render(&outBuffer); err != nil {
		t.Error(err)
	}
	output := outBuffer.String()

	if !strings.Contains(output, "Content is not available") {
		t.Errorf("the page did not contain the expected h2 text '%s'", "Content is not available")
	}
}

func Test403ErrorPage(t *testing.T) {
	var outBuffer bytes.Buffer
	page := html.ErrorPage403("--basepath--", config.Development, "dev")
	if err := page.Render(&outBuffer); err != nil {
		t.Error(err)
	}
	output := outBuffer.String()

	if !strings.Contains(output, "Access denied") {
		t.Errorf("the page did not contain the expected h2 text '%s'", "Access denied")
	}
	if !strings.Contains(output, "/gettoken") {
		t.Errorf("the Development env should point to the url '%s'", "/gettoken")
	}
	if strings.Contains(output, "Generate development token") {
		t.Errorf("the Development env should have a button with the text '%s'", "Generate development token")
	}

	// try with different environment
	outBuffer.Reset()
	html.ErrorPage403("--basepath--", config.Integration, "dev").Render(&outBuffer)
	output = outBuffer.String()
	if !strings.Contains(output, "Generate development token") {
		t.Errorf("the Development env should have a button with the text '%s'", "Generate development token")
	}
}

func TestErrorPage(t *testing.T) {
	var outBuffer bytes.Buffer
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Error(err)
	}

	page := html.ErrorApplication("--basepath--", config.Development, "dev", "HOME", req, "ERROR")
	if err := page.Render(&outBuffer); err != nil {
		t.Error(err)
	}

	output := outBuffer.String()
	if !strings.Contains(output, "The current request could not be processed and led to an error!") {
		t.Errorf("the page did not contain the text '%s'", "The current request could not be processed and led to an error!")
	}
	if !strings.Contains(output, "Request-Data") {
		t.Errorf("the page did not contain the text '%s'", "Request-Data")
	}
	if !strings.Contains(output, "Method: <strong>GET</strong>") {
		t.Errorf("the page did not contain the text '%s'", "Method: <strong>GET</strong>")
	}

}
