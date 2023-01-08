import random
import string

from playwright.sync_api import Page, expect


# validate the basic features of the bookmarks logic
def validate(baseURL: str, page: Page):
    prefix = ''.join(random.SystemRandom().choice(string.ascii_letters + string.digits) for _ in range(10))

    # we start by creating a new folder
    page.get_by_test_id('link-add-bookmark').click()
    page.locator('id=rd-bookmark-folder').click()
    page.locator('id=txt-bookmark-displaName').fill(prefix+'__FOLDER')
    page.locator('id=btn-bookmark-save').click()
    page.expect_request('**/api/v1/bookmarks**')

    # find the newly created item
    expect(page.get_by_text(prefix+'__FOLDER')).to_be_visible()
    page.get_by_text(prefix+'__FOLDER').click()

    # create a new bookmark within this folder
    page.get_by_test_id('link-add-bookmark').click()
    page.locator('id=txt-bookmark-displaName').fill(prefix+'__BOOKMARK')
    page.locator('id=txt-bookmark-url').fill('http://www.orf.at')
    page.locator('id=btn-bookmark-save').click()
    page.wait_for_timeout(500)

    # find the newly created item
    expect(page.get_by_text(prefix+'__BOOKMARK')).to_be_visible()
    expect(page.locator('.mat-menu-trigger')).to_be_visible()

    # edit the bookmark
    page.locator('.mat-menu-trigger').click()
    page.locator('id=btn-bookmark-edit').click()
    page.locator('id=txt-bookmark-displaName').fill(prefix+'__BOOKMARK-UPDATE')
    page.locator('id=btn-bookmark-save').click()
    page.wait_for_timeout(500)

    # again find the updated item in the list
    expect(page.get_by_text(prefix+'__BOOKMARK-UPDATE')).to_be_visible()
    expect(page.locator('.mat-menu-trigger')).to_be_visible()

    # delete the created bookmark
    page.locator('.mat-menu-trigger').click()
    page.locator('id=btn-bookmark-delete').click()
    page.locator('id=btn-confirm').click()
    page.wait_for_timeout(500)

    # empty list
    expect(page.get_by_text('no entries available')).to_be_visible()

    # got outside and search for the created folder
    page.locator('.rootroot').click()
    page.wait_for_url(baseURL + '/bookmarks')
    page.wait_for_timeout(500)
    # find the create item
    page.locator('.bookmark_item').filter(has_text=prefix+'__FOLDER').locator('.mat-menu-trigger').click()
    page.locator('id=btn-bookmark-delete').click()
    page.locator('id=btn-confirm').click()
    page.wait_for_timeout(500)
