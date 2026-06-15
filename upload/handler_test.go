package upload

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"testing"

	"golang.org/x/image/bmp"
)

func TestReadCoverConvertsSupportedFormatsToPNG(t *testing.T) {
	img := testCoverImage(8, 6)
	tests := []struct {
		name     string
		filename string
		data     []byte
	}{
		{name: "png", filename: "cover.png", data: encodePNG(t, img)},
		{name: "jpeg", filename: "cover.jpg", data: encodeJPEG(t, img)},
		{name: "webp", filename: "cover.webp", data: decodeBase64(t, "UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEADsD+JaQAA3AAAAAA")},
		{name: "bmp", filename: "cover.bmp", data: encodeBMP(t, img)},
		{name: "gif", filename: "cover.gif", data: encodeGIF(t, img)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, cleanup := multipartFileHeader(t, tt.filename, tt.data)
			defer cleanup()

			data, err := readCover(header)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) > maxCoverSize {
				t.Fatalf("cover size=%d, want <= %d", len(data), maxCoverSize)
			}
			if _, format, err := image.Decode(bytes.NewReader(data)); err != nil || format != "png" {
				t.Fatalf("decoded format=%q err=%v, want png", format, err)
			}
		})
	}
}

func TestReadCoverRejectsNonImage(t *testing.T) {
	header, cleanup := multipartFileHeader(t, "cover.txt", []byte("this is not an image file"))
	defer cleanup()

	_, err := readCover(header)
	if err == nil {
		t.Fatal("readCover succeeded, want error")
	}
	var uploadErr uploadError
	if !errors.As(err, &uploadErr) {
		t.Fatalf("err=%T, want uploadError", err)
	}
	if uploadErr.description != "invalid cover format" {
		t.Fatalf("description=%q, want invalid cover format", uploadErr.description)
	}
}

func TestReadCoverShrinksLargePNGToMaxCoverSize(t *testing.T) {
	img := noisyCoverImage(2048, 2048)
	source := encodePNG(t, img)
	if len(source) <= maxCoverSize {
		t.Fatalf("test source size=%d, want > %d", len(source), maxCoverSize)
	}
	header, cleanup := multipartFileHeader(t, "cover.png", source)
	defer cleanup()

	data, err := readCover(header)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > maxCoverSize {
		t.Fatalf("cover size=%d, want <= %d", len(data), maxCoverSize)
	}
	decoded, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if format != "png" {
		t.Fatalf("format=%q, want png", format)
	}
	if decoded.Bounds().Dx() >= img.Bounds().Dx() || decoded.Bounds().Dy() >= img.Bounds().Dy() {
		t.Fatalf("decoded bounds=%v, want smaller than %v", decoded.Bounds(), img.Bounds())
	}
}

func multipartFileHeader(t *testing.T, filename string, data []byte) (*multipart.FileHeader, func()) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("cover", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data) + 1024))
	if err != nil {
		t.Fatal(err)
	}
	files := form.File["cover"]
	if len(files) != 1 {
		t.Fatalf("cover files=%d, want 1", len(files))
	}
	return files[0], func() { _ = form.RemoveAll() }
}

func testCoverImage(width int, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(20 + x*17),
				G: uint8(40 + y*19),
				B: uint8(80 + x*y),
				A: 255,
			})
		}
	}
	return img
}

func noisyCoverImage(width int, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	var value uint32 = 1
	for y := range height {
		for x := range width {
			value = value*1664525 + 1013904223
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(value >> 24),
				G: uint8(value >> 16),
				B: uint8(value >> 8),
				A: 255,
			})
		}
	}
	return img
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func encodeBMP(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func encodeGIF(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := gif.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func decodeBase64(t *testing.T, value string) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
