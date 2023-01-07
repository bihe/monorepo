package main

import (
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// validate the basic logic of the bookmarks page
// execute CRUD operations and check the results
type bookmarksTest struct {
	page playwright.Page
}

func (b *bookmarksTest) validate() {
	prefix := randStr(7)
	b.addEditDeleteBookmark(prefix)
}

func (b *bookmarksTest) addEditDeleteBookmark(p string) {
	err := b.page.Click("#link-add-bookmark")
	assertError("could not click link to add a bookmark; %v", err)

	// create a folder
	b.page.Click("#rd-bookmark-folder")
	b.page.Fill("#txt-bookmark-displaName", p+"__FOLDER")

	err = b.page.Click("#btn-bookmark-save")
	assertError("could not create a new folder; %v", err)

	// wait for the dialog to settle
	b.page.WaitForTimeout(1000)

	// go to the created folder
	item := locateItem(b.page, ".class-bookmark-item", p+"__FOLDER")
	err = item.Click()
	assertError("could not click the link, %v", err)

	b.page.WaitForTimeout(500)

	// ---- create a bookmark ----

	// create a new bookmark item
	err = b.page.Click("#link-add-bookmark")
	assertError("could not click link to add a bookmark; %v", err)

	// create a folder
	b.page.Fill("#txt-bookmark-displaName", p+"__BOOKMARK")
	b.page.Fill("#txt-bookmark-url", "http://www.orf.at")

	err = b.page.Click("#btn-bookmark-save")
	assertError("could not create a new folder; %v", err)

	// wait for the dialog to settle
	b.page.WaitForTimeout(1000)

	// find created bookmark
	locateItem(b.page, ".class-bookmark-item", p+"__BOOKMARK")

	// ---- edit the bookmark ----

	// open the context menu
	locateAndClick(b.page, ".mat-menu-trigger")

	// click the delete button
	locateAndClick(b.page, "#btn-bookmark-edit")

	b.page.Fill("#txt-bookmark-displaName", p+"__BOOKMARK-UPDATE")
	locateAndClick(b.page, "#btn-bookmark-save")

	// settle dialog
	b.page.WaitForTimeout(1000)

	locateItem(b.page, ".class-bookmark-item", p+"__BOOKMARK-UPDATE")

	// ---- delete the bookmark ----

	// open the context menu
	locateAndClick(b.page, ".mat-menu-trigger")

	// click the delete button
	locateAndClick(b.page, "#btn-bookmark-delete")

	// confirm the delete dialog
	locateAndClick(b.page, "#btn-confirm")

	// wait for the dialog to settle
	b.page.WaitForTimeout(1000)

	locateItem(b.page, ".bookmark_item", "no entries available")

	// ---- go outside and search folder ----

	item = locateItem(b.page, ".rootroot", "root")
	item.Click()

	b.page.WaitForTimeout(1000)

	loc, err := b.page.Locator(".bookmark_item")
	assertError("could not locate '.bookmark_item'", err)
	items, _ := loc.ElementHandles()
	for _, item := range items {
		txt, _ := item.InnerText()
		if strings.Contains(txt, p+"__FOLDER") {
			locItem, _ := item.QuerySelector(".mat-menu-trigger")
			locItem.Click()

			// click the delete button
			locateAndClick(b.page, "#btn-bookmark-delete")

			// confirm the delete dialog
			locateAndClick(b.page, "#btn-confirm")
			break
		}
	}

	b.page.WaitForTimeout(1000)

}

func locateAndClick(page playwright.Page, selector string) {
	loc, err := page.Locator(selector)
	assertError("could not get the locator for '"+selector+"'; %v", err)
	loc.Click()
}

func locateItem(page playwright.Page, selector, itemText string) playwright.ElementHandle {
	var (
		itemDisplayName string
		item            playwright.ElementHandle
	)

	loc, err := page.Locator(selector)
	assertError("could not get the locator for class '.class-bookmark-item'; %v", err)
	items, err := loc.ElementHandles()
	assertError("could not get the list of items; %v", err)

	for _, entry := range items {
		name, _ := entry.TextContent()
		log.Printf("found item: %s", name)
		if name == itemText {
			itemDisplayName = name
			item = entry
			break
		}
	}
	assertEqual(itemText, itemDisplayName)

	return item
}
