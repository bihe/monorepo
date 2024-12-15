import os
import pathlib
import random
import string

from playwright.sync_api import Page, expect


def validate_mydms(baseURL: str, page: Page):
    prefix = "".join(random.SystemRandom().choice(string.ascii_letters + string.digits) for _ in range(10))

    title = prefix + "__Item"

    # we start by creating a new folder
    page.get_by_test_id("link-add-document").click()
    page.locator("id=document_title").fill(title)
    page.locator("id=document_amount").fill("10.1")
    page.locator("id=document_number").fill("123456")

    # get the current path
    path = pathlib.Path(__file__).parent.resolve()

    # select a file to upload
    with page.expect_file_chooser() as fc_info:
        page.locator("id=documentFileUpload").click()
        file_chooser = fc_info.value
        file_chooser.set_files(os.path.join(path, "../unencrypted.pdf"))
    # perform the upload
    page.locator("id=btn-doc-fileupload").click()
    page.wait_for_timeout(500)

    # search for the uploaded file
    expect(page.locator("id=document_download_link")).to_be_visible()

    # fill tags and senders
    page.get_by_placeholder("Choose a tag...").fill("Apple")
    page.locator("id=tags-menu-1-0").click()
    page.get_by_placeholder("Choose a sender...").fill("A1")
    page.locator("id=tags-menu-2-0").click()

    # save the new document
    page.locator("id=btn-document-save").click()
    page.wait_for_timeout(500)

    expect(page.get_by_test_id("edit-document").first).to_contain_text(title)

    # open dialog
    page.get_by_test_id("edit-document").first.click()
    # locate the close button
    page.locator("id=document_edit_dialog").get_by_text("Close").click()

    # locate the card again
    page.locator(".dropdown-toggle").first.click()
    page.locator("id=btn-document-delete").first.click()

    # confirm deletion
    page.locator("id=btn-confirm").click()
    page.wait_for_timeout(500)
    # check the toast message
    expect(page.locator("id=toast_messsage_text-success")).to_be_visible()
    expect(page.locator("id=toast_messsage_text-success")).to_contain_text(" was removed.")
    expect(page.get_by_test_id("edit-document").first).not_to_contain_text(title)
