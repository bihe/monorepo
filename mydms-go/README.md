# mydms-go

Simple application to upload, store, search documents and meta-data.

[![codecov](https://codecov.io/gh/bihe/mydms-go/branch/master/graph/badge.svg)](https://codecov.io/gh/bihe/mydms-go)
[![Sonarcloud Status](https://sonarcloud.io/api/project_badges/measure?project=bihe_mydms-go&metric=alert_status)](https://sonarcloud.io/dashboard?id=bihe_mydms-go)
[![Build Status](https://dev.azure.com/henrikbinggl/mydms-go/_apis/build/status/bihe.mydms-go?branchName=master)](https://dev.azure.com/henrikbinggl/mydms-go/_build/latest?definitionId=6&branchName=master)

## Structure

The basic structure of 'mydms' is a REST backend by [echo](https://github.com/labstack/echo) using [golang](https://golang.org/), meta-data is kept in [mariadb](https://mariadb.org/), documents stored in [S3](https://aws.amazon.com/s3/) and the frontend provided via [angular](https://angular.io/).

## Technology

* REST backend: labstack/echo (v4.x), golang (1.1x)
* frontend angular (8.x.x)
* mariadb: 10.x

## Build

The REST Api and the UI can be built separately.

### UI

`npm run build -- --prod --base-href /ui/`

### Api

`go build`

## Why

I needed something to keep track of my scanned invoices. Being a software nerd, I created a solution for this purpose. The added benefit for me is, that I have a technology playground to try out new things.

There are different versions/iterations available.

* [mydms-node](https://github.com/bihe/myDMS-node) - very early adventures in node.js
* [mydms-java (dropwizard)](https://github.com/bihe/mydms-java/tree/dropwizard) - use dropwizard as the REST backend and documents were stored in Google Drive
* [mydms-java (spring-boot)](https://github.com/bihe/mydms-java) - use spring-boot/kotlin as the REST backend and documents were stored in AWS S3
