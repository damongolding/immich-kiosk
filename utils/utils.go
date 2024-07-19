package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/disintegration/imaging"
)

// ImageToBase64 converts image bytes into a base64 string
func ImageToBase64(imgBtyes []byte) (string, error) {

	var base64Encoding string

	mimeType := http.DetectContentType(imgBtyes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBtyes)

	return base64Encoding, nil
}

// BlurImage converts image bytes into a blurred base64 string
func BlurImage(imgBytes []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return buf.Bytes(), err
	}

	blurredImg := imaging.Blur(img, 20)
	blurredImg = imaging.AdjustBrightness(blurredImg, -20)

	err = jpeg.Encode(buf, blurredImg, nil)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
