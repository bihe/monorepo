// Package favicon fetches favicons from URLs
package favicon

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

const DefaultFaviconName = "favicon.ico"

type FetchType string

const (
	FetchImage FetchType = "image"
	FetchAll   FetchType = "*"
)

// Content defines the content of a URI response
type Content struct {
	Payload  []byte
	FileName string
	MimeType string
}

// GetFaviconFromURL tries to find and fetch the favicon from the given URL
func GetFaviconFromURL(url string) (content Content, err error) {
	var (
		scheme  string
		baseURL string
		pageURL string
		iconURL string
	)

	if scheme, baseURL, pageURL, err = parseURL(url); err != nil {
		return
	}

	if iconURL, _, err = parseHtmlPageForFavicon(url); err != nil {
		// no favicon found on page
		// fall back to the standard to get the favicon from the base-path
		iconURL = fmt.Sprintf("%s/%s", baseURL, DefaultFaviconName)
		if content, err = FetchURL(iconURL, FetchImage); err != nil {
			err = fmt.Errorf("could not fetch favicon '%s': %v", iconURL, err)
			return
		}
		return Content{
			Payload:  content.Payload,
			FileName: content.FileName,
			MimeType: content.MimeType,
		}, nil
	}

	// we have parsed the favicon from the html
	// now ensure that the parsed url is downloadable html pages use some kind of tricks:
	// a) missing base-url href=/assets/abc/favicon.png
	// b) missing scheme //cdn.com/abc/favicon.png

	if strings.HasPrefix(iconURL, "//") {
		iconURL = scheme + ":" + iconURL
	} else if strings.HasPrefix(iconURL, "/") {
		iconURL = baseURL + iconURL
	} else if strings.HasPrefix(iconURL, "./") {
		// local to the page-URL
		iconURL = strings.ReplaceAll(iconURL, "./", "/")
		iconURL = pageURL + iconURL
	} else if !strings.HasPrefix(iconURL, "http") {
		// if a file without anything is specified "favicon.png"
		// then use the pageurl
		iconURL = pageURL + iconURL
	}

	if content, err = FetchURL(iconURL, FetchImage); err != nil {
		err = fmt.Errorf("could not fetch favicon '%s': %v", iconURL, err)
		return
	}
	return
}

// FetchURL retrieves the payload of the specified URL
func FetchURL(uri string, what FetchType) (Content, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return Content{}, fmt.Errorf("could not parse the supplied uri: %v", err)
	}
	fileName := path.Base(u.Path)
	// determine if fileName has an extension or is just the base-name
	if path.Ext(fileName) == "" {
		fileName = ""
	}

	resp, err := http.Get(uri)
	if err != nil {
		return Content{}, fmt.Errorf("could not fetch page: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Content{}, fmt.Errorf("got status %d", resp.StatusCode)
	}
	mimeType := resp.Header.Get("Content-Type")
	if what == FetchImage {
		if !strings.HasPrefix(mimeType, "image/") {
			return Content{}, fmt.Errorf("the payload needs to be an image mime-type; got '%s'", mimeType)
		}
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return Content{}, fmt.Errorf("could not read content body: %v", err)
	}
	if fileName == "" {
		// try to parse the filename from a content-disposition header
		dispHeader := resp.Header.Get("Content-Disposition")
		if dispHeader != "" {
			var params map[string]string
			if _, params, err = mime.ParseMediaType(dispHeader); err == nil {
				fileName = params["filename"]
			}
		}
		if fileName == "" {
			fileName = DefaultFaviconName
		}
	}

	return Content{
		FileName: fileName,
		Payload:  content,
		MimeType: mimeType,
	}, nil
}

func parseURL(uri string) (scheme, baseURL, pageURL string, err error) {
	var u *url.URL
	u, err = url.Parse(uri)
	if err != nil {
		return "", "", "", fmt.Errorf("could not parse the supplied uri: %v", err)
	}
	path := u.Path
	if strings.HasSuffix(path, "index.html") || strings.HasSuffix(path, "index.htm") {
		path = strings.ReplaceAll(path, "index.html", "")
		path = strings.ReplaceAll(path, "index.htm", "")
	}

	return u.Scheme, fmt.Sprintf("%s://%s", u.Scheme, u.Host), fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path), nil
}

func parseHtmlPageForFavicon(url string) (iconUrl, fileName string, err error) {
	var (
		page Content
	)

	if page, err = FetchURL(url, FetchAll); err != nil {
		return
	}
	if iconUrl, err = tryFaviconDefinitions(page.Payload); err != nil {
		return
	}

	parts := strings.Split(iconUrl, "/")
	fileName = parts[len(parts)-1]
	return iconUrl, fileName, nil
}

func tryFaviconDefinitions(page []byte) (string, error) {
	var (
		iconUrl string
		err     error
	)
	iconUrl, err = parsePageForFavicon(page, "shortcut icon")
	if err != nil {
		iconUrl, err = parsePageForFavicon(page, "icon")
	}
	return iconUrl, err
}

func parsePageForFavicon(page []byte, faviconDef string) (string, error) {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("could not parse page: %v", err)
	}
	icon, err := faviconParser(doc, faviconDef)
	if err != nil || icon == "" {
		return "", fmt.Errorf("could not find a favicon definition on page")
	}
	return icon, nil
}

func faviconParser(doc *html.Node, faviconDef string) (string, error) {
	var link *html.Node
	var crawler func(*html.Node)

	extractFavIconURL := func(node *html.Node, faviconDef string) string {
		foundFavDev := false
		favHrf := ""
		for _, a := range node.Attr {

			switch a.Key {
			case "rel":
				foundFavDev = a.Val == faviconDef
			case "href":
				favHrf = a.Val
			}
		}
		if foundFavDev && favHrf != "" {
			return favHrf
		}
		return ""
	}

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "link" {
			link = node
			if extractFavIconURL(link, faviconDef) != "" {
				return
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				crawler(child)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if link != nil {
		return extractFavIconURL(link, faviconDef), nil
	}
	return "", fmt.Errorf("could not find a favicon definition on page")
}
