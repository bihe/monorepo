# Integration Tests
We are use [playwright](https://playwright.dev/) from Microsoft to perform integration tests. Those tests are implemented as a separate program and do not run as "normal" unit-tests. This is also the case because the integration tests need all the services running, therefor relying on a working docker-compose setup.


## Install
We use the `python` variant of playwright and according to the website [playwright-python](https://playwright.dev/python/docs/intro) the installation is as follows.

```bash
# python3 and pip need to be available!
# on mac os those versions are called python3 and pip3
pip install pytest-playwright
playwright install
```
