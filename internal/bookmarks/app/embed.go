package app

import _ "embed"

//go:embed bookmark.svg
var DefaultFavicon []byte

//go:embed folder.svg
var DefaultIconFolder []byte

//go:embed file.svg
var DefaultIconFile []byte
