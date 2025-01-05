from bookmarks import validate_bookmarks
from mydms import validate_mydms
from navigation import validate_basic_navigation
from playwright.sync_api import sync_playwright

# the URL for local testing.
# self-signed certs and the hostname needs to be available locally
STARTUPURL = "https://dev.binggl.net"

# we want to see the browser in action
WORKHEADLESS = False
# optionally enable devtools in the browser instance
USEDEVTOOLS = False
# we work with a self-signed-cert, therefor this is needed
IGNORE_HTTPS_ERRORS = True


def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=WORKHEADLESS, devtools=USEDEVTOOLS)
        try:
            page = browser.new_page(ignore_https_errors=IGNORE_HTTPS_ERRORS)
            page.goto(STARTUPURL)

            # generate the development/test token to authenticate
            page.get_by_role("link", name="Generate development token").click()

            page.wait_for_url(STARTUPURL + "/bm")

            # start specific validation
            validate_basic_navigation(STARTUPURL, page)

            page.goto(STARTUPURL + "/bm")
            page.wait_for_timeout(500)
            validate_bookmarks(STARTUPURL, page)
            page.goto(STARTUPURL + "/mydms")
            page.wait_for_timeout(500)
            validate_mydms(STARTUPURL, page)

            page.wait_for_timeout(1000)
        except Exception as e:
            print("got an error during execution -> " + str(e))

        browser.close()


if __name__ == "__main__":
    main()
