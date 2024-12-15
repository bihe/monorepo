import random
import string

from playwright.sync_api import Page, expect

class_trigger_menu = ".dropdown-toggle"


# validate the basic features of the bookmarks logic
def validate_bookmarks(baseURL: str, page: Page):
    prefix = "".join(random.SystemRandom().choice(string.ascii_letters + string.digits) for _ in range(10))

    # we start by creating a new folder
    page.get_by_test_id("link-add-bookmark").click()
    page.locator("id=type_Folder").click()
    page.locator("id=bookmark_DisplayName").fill(prefix + "__FOLDER")
    page.locator("id=btn-bookmark-save").click()

    # find the newly created item
    expect(page.get_by_role("link", name=prefix + "__FOLDER")).to_be_visible()
    page.get_by_role("link", name=prefix + "__FOLDER").click()

    # create a new bookmark within this folder
    page.get_by_test_id("link-add-bookmark").click()
    page.locator("id=bookmark_DisplayName").fill(prefix + "__BOOKMARK")
    page.locator("id=bookmark_URL").fill("http://www.orf.at")
    page.locator("id=btn-bookmark-save").click()
    page.wait_for_timeout(500)

    # find the newly created item
    expect(page.get_by_role("link", name=prefix + "__BOOKMARK")).to_be_visible()
    expect(page.locator(class_trigger_menu)).to_be_visible()

    # edit the bookmark
    page.locator(class_trigger_menu).click()
    page.locator("id=btn-bookmark-edit").click()
    page.locator("id=bookmark_DisplayName").fill(prefix + "__BOOKMARK-UPDATE")
    page.locator("id=btn-bookmark-save").click()
    page.wait_for_timeout(500)

    # again find the updated item in the list
    expect(page.get_by_role("link", name=prefix + "__BOOKMARK-UPDATE")).to_be_visible()
    expect(page.locator(class_trigger_menu)).to_be_visible()

    # delete the created bookmark
    page.locator(class_trigger_menu).click()
    page.locator("id=btn-bookmark-delete").click()
    page.locator("id=btn-confirm").click()
    page.wait_for_timeout(500)

    # empty list
    expect(page.get_by_text("no entries available")).to_be_visible()

    # got outside and search for the created folder
    page.locator(".rootroot").click()
    page.wait_for_url(baseURL + "/bm/~/")
    page.wait_for_timeout(500)

    # find the folder where the stuff was created into
    expect(page.locator(".bookmark_item").filter(has_text=prefix + "__FOLDER")).to_be_visible()

    page.locator(".bookmark_item").filter(has_text=prefix + "__FOLDER").locator(class_trigger_menu).click()
    page.locator(".bookmark_item").filter(has_text=prefix + "__FOLDER").locator("#btn-bookmark-delete").click()
    page.locator("id=btn-confirm").click()
    page.wait_for_timeout(500)
