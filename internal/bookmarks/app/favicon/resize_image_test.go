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
	// force a resize
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
		t.Errorf("could not decode image; %v", err)
	}

	if img.Bounds().Dx() != 50 {
		t.Errorf("could not resize image to x:50/50")
	}
	if img.Bounds().Dy() != 50 {
		t.Errorf("could not resize image to 50/y:50")
	}

	// resize with aspect-ratio honoring // x
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, 50, 0)

	if err != nil {
		t.Errorf("could not resize imag: %v", err)
	}

	img, _, err = image.Decode(bytes.NewBuffer(resized.Payload))
	if err != nil {
		t.Errorf("could not decode image; %v", err)
	}

	if img.Bounds().Dx() != 50 {
		t.Errorf("could not resize image to x:50/41")
	}
	if img.Bounds().Dy() != 41 {
		t.Errorf("could not resize image to 50/y:41")
	}

	// resize with aspect-ratio honoring // y
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, 0, 50)

	if err != nil {
		t.Errorf("could not resize imag: %v", err)
	}

	img, _, err = image.Decode(bytes.NewBuffer(resized.Payload))
	if err != nil {
		t.Errorf("could not decode image; %v", err)
	}

	if img.Bounds().Dx() != 60 {
		t.Errorf("could not resize image to x:60/50")
	}
	if img.Bounds().Dy() != 50 {
		t.Errorf("could not resize image to 60/y:50")
	}

	// resize a JPEG
	// ------------------------------------------------------------------
	img, _, err = image.Decode(bytes.NewBuffer(jpegFavicon))
	if err != nil {
		t.Errorf("could not decode image; %v", err)
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
		t.Errorf("could not decode image; %v", err)
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

	// error: x/y = 0/0
	// ------------------------------------------------------------------
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, 0, 0)

	if err == nil {
		t.Error("error expected")
	}

	// error: x/y = -1/-1
	// ------------------------------------------------------------------
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, -1, -1)

	if err == nil {
		t.Error("error expected")
	}

	// error: x/y = -1/50
	// ------------------------------------------------------------------
	resized, err = favicon.ResizeImage(favicon.Content{
		Payload:  pngFavicon,
		FileName: "favicon.png",
		MimeType: "image/png",
	}, -1, 50)

	if err == nil {
		t.Error("error expected")
	}
}
