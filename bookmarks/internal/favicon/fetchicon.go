// Package favicon fetches favicons from URLs
package favicon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const DefaultFaviconName = "favicon.ico"

// GetFaviconFromURL tries to find and fetch the favicon from the given URL
func GetFaviconFromURL(url string) (fileName string, payload []byte, err error) {
	var (
		scheme  string
		baseURL string
		pageURL string
		iconURL string
	)

	if scheme, baseURL, pageURL, err = parseURL(url); err != nil {
		return
	}

	if iconURL, fileName, err = parseHtmlPageForFavicon(url); err != nil {
		// no favicon found on page
		// fall back to the standard to get the favicon from the base-path
		iconURL = fmt.Sprintf("%s/%s", baseURL, DefaultFaviconName)
		if payload, err = fetchURL(iconURL); err != nil {
			err = fmt.Errorf("could not fetch favicon '%s': %v", iconURL, err)
			return
		}
		return DefaultFaviconName, payload, nil
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

	if payload, err = fetchURL(iconURL); err != nil {
		err = fmt.Errorf("could not fetch favicon '%s': %v", iconURL, err)
		return
	}
	return
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
		page []byte
	)

	if page, err = fetchURL(url); err != nil {
		return
	}
	if iconUrl, err = tryFaviconDefinitions(page); err != nil {
		return
	}

	parts := strings.Split(iconUrl, "/")
	fileName = parts[len(parts)-1]
	return iconUrl, fileName, nil
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch page: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read content body: %v", err)
	}
	return content, nil
}

func tryFaviconDefinitions(page []byte) (string, error) {
	var (
		iconUrl string
		err     error
	)
	iconUrl, err = parseFavicon(page, "icon")
	if err != nil {
		iconUrl, err = parseFavicon(page, "shortcut icon")
	}
	return iconUrl, err
}

func parseFavicon(page []byte, faviconDef string) (string, error) {
	var (
		iconUrl string
		err     error
		doc     *goquery.Document
		ok      bool
	)

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("could not parse page: %v", err)
	}

	doc.Find(fmt.Sprintf(`link[rel="%s"]`, faviconDef)).Each(func(i int, s *goquery.Selection) {
		iconUrl, ok = s.Attr("href")
	})
	if iconUrl == "" || !ok {
		return "", fmt.Errorf("could not find a favicon definition on page")
	}
	return iconUrl, nil
}
