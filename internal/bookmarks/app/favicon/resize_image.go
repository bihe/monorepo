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
// If the payload is not supported the same content is returned without any resize operation.
// The resize function honors the aspect-ratio of the original image if one of the values x/y is 0.
// Aspect is not considered if both values for x and y are set (!= 0)
func ResizeImage(content Content, x, y int) (Content, error) {
	if x == 0 && y == 0 {
		return Content{}, fmt.Errorf("x and y of the resize operation are 0")
	}
	if x < 0 || y < 0 {
		return Content{}, fmt.Errorf("x or y of the resize operation are less than 0")
	}

	imageType := determineImage(content)
	if imageType == OTHER {
		return content, nil
	}

	src, err := decode(content.Payload, imageType)
	if err != nil {
		return Content{}, err
	}

	aspect_x := float32(x)
	aspect_y := float32(y)

	if aspect_x == 0 || aspect_y == 0 {
		// stay true to the original aspect of the image
		aspect := float32(src.Bounds().Dx()) / float32(src.Bounds().Dy())

		if aspect_x == 0 {
			// y is set
			aspect_x = float32(y) * aspect
		}
		if aspect_y == 0 {
			// x is set
			aspect_y = float32(x) / aspect
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(aspect_x), int(aspect_y)))
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
