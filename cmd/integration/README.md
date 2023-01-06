# Integration Tests

We are use [playwright](https://playwright.dev/) from Microsoft to perform integration tests. Those tests are implemented as a separate program and do not run as "normal" unit-tests. This is also the case because the integration tests need all the services running, therefor relying on a working docker-compose setup.

There is a community implementation of playwright to use [#golang for creating tests](https://github.com/playwright-community/playwright-go). This is one option, which is enough to cover basic needs. If more requirements arise the tests will possibly moved to one of the fully supported languages (js/ts, python, dotnet).

## Install
The `playwright-go` dependency is defined in the `go.mod` file, but this is not enough. In addition, the headless browser bits need to be installed as well to actually execute playwright.

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps

# output of the command above
2023/01/06 12:51:57 Downloading driver to /Users/henrik/Library/Caches/ms-playwright-go/1.20.0-beta-1647057403000
2023/01/06 12:52:00 Downloaded driver successfully
Downloading Playwright build of chromium v978106 - 118.4 Mb [====================] 100% 0.0s
Playwright build of chromium v978106 downloaded to /Users/henrik/Library/Caches/ms-playwright/chromium-978106
Downloading Playwright build of ffmpeg v1007 - 1 Mb [====================] 100% 0.0s
Playwright build of ffmpeg v1007 downloaded to /Users/henrik/Library/Caches/ms-playwright/ffmpeg-1007
Downloading Playwright build of firefox v1319 - 69.2 Mb [====================] 100% 0.0s
Playwright build of firefox v1319 downloaded to /Users/henrik/Library/Caches/ms-playwright/firefox-1319
Downloading Playwright build of webkit v1616 - 54.6 Mb [====================] 100% 0.0s
Playwright build of webkit v1616 downloaded to /Users/henrik/Library/Caches/ms-playwright/webkit-1616
````

