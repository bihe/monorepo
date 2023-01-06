package main

import (
	"github.com/playwright-community/playwright-go"
)

const bookmarksURL = "/bookmarks"
const mydmsURL = "/mydms"
const sitesURL = "/sites"

type navigationTest struct {
	page playwright.Page
}

func (n *navigationTest) validate() {
	// bookmarks
	lnk, _ := n.page.Locator("#link-bookmarks")
	href, _ := lnk.GetAttribute("href")
	assertEqual(bookmarksURL, href)

	// mydms
	lnk, _ = n.page.Locator("#link-mydms")
	href, _ = lnk.GetAttribute("href")
	assertEqual(mydmsURL, href)

	// sites
	lnk, _ = n.page.Locator("#link-sites")
	href, _ = lnk.GetAttribute("href")
	assertEqual(sitesURL, href)
}
