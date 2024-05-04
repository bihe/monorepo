package favicon

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path"

	"golang.org/x/image/draw"
)

type imageType int

const (
	PNG imageType = iota
	JPEG
	OTHER
)

// ResizeImage uses the image payload and resizes the image.
// If the payload is not supported the same content is resturned without any resize operation
func ResizeImage(content Content, x, y int) (Content, error) {
	imageType := determineImage(content)
	if imageType == OTHER {
		return content, nil
	}

	src, err := decode(content.Payload, imageType)
	if err != nil {
		return Content{}, err
	}
	// half of the original size
	// dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))
	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err = encode(dst, &buf, imageType); err != nil {
		return Content{}, err
	}

	return Content{
		FileName: content.FileName,
		MimeType: content.MimeType,
		Payload:  buf.Bytes(),
	}, nil
}

func decode(payload []byte, imageType imageType) (image.Image, error) {
	var (
		img image.Image
		err error
	)
	switch imageType {
	case PNG:
		buf := bytes.NewBuffer(payload)
		if img, err = png.Decode(buf); err != nil {
			return nil, fmt.Errorf("could not decoded png file; %w", err)
		}
		return img, nil
	case JPEG:
		buf := bytes.NewBuffer(payload)
		if img, err = jpeg.Decode(buf); err != nil {
			return nil, fmt.Errorf("could not decoded png file; %w", err)
		}
		return img, nil
	}
	return nil, fmt.Errorf("decode only supports PNG/JPEG")
}

func determineImage(content Content) imageType {
	var t imageType = OTHER
	switch content.MimeType {
	case "image/png":
		return PNG
	case "image/jpeg":
		return JPEG
	}

	if t == OTHER {
		// try to determine the type based on the file-extension
		ext := path.Ext(content.FileName)
		switch ext {
		case ".png":
			return PNG
		case ".jpg":
			return JPEG
		case ".jpeg":
			return JPEG
		}
	}
	return OTHER
}

func encode(resized *image.RGBA, w io.Writer, imageType imageType) error {
	switch imageType {
	case PNG:
		return png.Encode(w, resized)
	case JPEG:
		return jpeg.Encode(w, resized, &jpeg.Options{Quality: jpeg.DefaultQuality})
	}
	return fmt.Errorf("encode only supports PNG/JPEG")
}
