package favicon_test

import (
	"bytes"
	_ "embed"
	"image"
	"testing"

	"golang.binggl.net/monorepo/internal/bookmarks/app/favicon"
)

//go:embed favicon.png
var pngFavicon []byte

//go:embed favicon.jpg
var jpegFavicon []byte

//go:embed favicon.ico
var icoFavicon []byte

func Test_ResizeImage(t *testing.T) {
	resized, err := favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, 50, 50)

	if err != nil {
		t.Errorf("could not resize imag: %v", err)
	}

	img, _, err := image.Decode(bytes.NewBuffer(resized.Payload))
	if err != nil {
		t.Errorf("could not decoed image; %v", err)
	}

	if img.Bounds().Dx() != 50 && img.Bounds().Dy() != 50 {
		t.Errorf("could not resize image to 50/50")
	}

	// resize a JPEG
	// ------------------------------------------------------------------
	img, _, err = image.Decode(bytes.NewBuffer(jpegFavicon))
	if err != nil {
		t.Errorf("could not decoed image; %v", err)
	}
	if img.Bounds().Dx() != 172 && img.Bounds().Dy() != 178 {
		t.Errorf("could not resize image to 50/50")
	}

	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  jpegFavicon,
		FileName: "favicon.jpg",
		MimeType: "image/jpeg",
	}, 50, 50)

	if err != nil {
		t.Errorf("could not resize imag: %v", err)
	}

	img, _, err = image.Decode(bytes.NewBuffer(resized.Payload))
	if err != nil {
		t.Errorf("could not decoed image; %v", err)
	}

	if img.Bounds().Dx() != 50 && img.Bounds().Dy() != 50 {
		t.Errorf("could not resize image to 50/50")
	}

	// no resize because of unsupported filetype
	// ------------------------------------------------------------------
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  icoFavicon,
		FileName: "favicon.ico",
		MimeType: "image/x-icon",
	}, 50, 50)

	if err != nil {
		t.Errorf("could not resize imag: %v", err)
	}

	if len(icoFavicon) != len(resized.Payload) {
		t.Errorf("the returned payload is not the same")
	}
}
