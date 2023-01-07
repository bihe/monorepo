package main

import (
	"log"

	"github.com/playwright-community/playwright-go"
)

const bookmarksURL = "/bookmarks"
const mydmsURL = "/mydms"
const sitesURL = "/sites"

// validate the basic navigation features of the page
type navigationTest struct {
	page playwright.Page
}

func (n *navigationTest) validate() {
	// bookmarks
	lnk, err := n.page.Locator("#link-bookmarks")
	assertError("could not get #link-bookmarks, %v", err)
	href, _ := lnk.GetAttribute("href")
	assertEqual(bookmarksURL, href)

	// mydms
	lnk, err = n.page.Locator("#link-mydms")
	assertError("could not get #link-mydms, %v", err)
	href, _ = lnk.GetAttribute("href")
	assertEqual(mydmsURL, href)

	// sites
	lnk, _ = n.page.Locator("#link-sites")
	assertError("could not get #link-sites, %v", err)
	href, _ = lnk.GetAttribute("href")
	assertEqual(sitesURL, href)

	// test the navigation
	err = n.page.Click("#link-mydms")
	assertError("could not navigate to mydms; %v", err)
	n.page.WaitForURL(startupURL + mydmsURL)
	log.Printf("page URL: %s", n.page.URL())

	err = n.page.Click("#link-sites")
	assertError("could not navigate to sites; %v", err)
	n.page.WaitForURL(startupURL + mydmsURL)
	log.Printf("page URL: %s", n.page.URL())

	err = n.page.Click("#link-bookmarks")
	assertError("could not navigate to bookmarks; %v", err)
	n.page.WaitForURL(startupURL + mydmsURL)
	log.Printf("back to bookmarks URL: %s", n.page.URL())
}
