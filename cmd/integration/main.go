package main

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/playwright-community/playwright-go"
)

// the URL for local testing.
// note: it uses self-signed certs and the hostname needs to be available locally
const startupURL = "https://dev.binggl.net"

// we want to see the browser in action
const workHeadless = false

func main() {
	rand.Seed(time.Now().Unix())
	pw, browser, page := startUp()
	navigation := &navigationTest{page: page}
	bookmarks := bookmarksTest{page: page}

	_, err := page.Goto(startupURL)
	assertError("could not goto: %v", err)

	lnk, err := page.Locator("#link-gettoken")
	assertError("could not find the gettoken link during development: %v", err)

	txt, err := lnk.InnerText()
	assertError("could not get the inner text of the link: %v", err)
	assertEqual("Generate development token", txt)

	// generate a token and follow the redirect
	err = page.Click("#link-gettoken")
	assertError("could not follow the link to generate the jwt; %v", err)

	_, err = page.WaitForNavigation()
	assertError("could not navigate to app starting URL; %v", err)
	log.Printf("page URL: %s", page.URL())

	page.WaitForTimeout(1500) // I want to see something

	// start the individual tests
	navigation.validate()
	bookmarks.validate()

	// make clean house and close everything
	shutDown(pw, browser)
}

func assertError(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func assertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("%v does not equal %v", actual, expected))
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func startUp() (*playwright.Playwright, playwright.Browser, playwright.Page) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(workHeadless),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	return pw, browser, page
}

func shutDown(pw *playwright.Playwright, browser playwright.Browser) {
	var err error

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
