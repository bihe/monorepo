from playwright.sync_api import Page, expect


# validate the principal navigation feature of the application
def validate_basic_navigation(baseURL: str, page: Page):
    page.get_by_role("link", name="Sites").click()
    expect(page).to_have_url(baseURL + "/sites")

    page.get_by_text(text="Bookmarks", exact=True).click()
    expect(page).to_have_url(baseURL + "/bm")

    page.get_by_text(text="Documents", exact=True).click()
    expect(page).to_have_url(baseURL + "/mydms")
