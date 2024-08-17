package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"math/rand/v2"
	"mime"
	"net/http"
	"net/url"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/webp"
	_ "golang.org/x/image/webp"

	"github.com/charmbracelet/log"
	"github.com/disintegration/imaging"
)

// DateToLayout takes a string and replaces normal date layouts to GO layouts
func DateToLayout(input string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"M", "1",
		"DDDD", "Monday",
		"DDD", "Mon",
		"DD", "02",
		"D", "2",
	)
	return replacer.Replace(input)
}

// ImageToBase64 converts image bytes into a base64 string
func ImageToBase64(imgBtyes []byte) (string, error) {

	var base64Encoding string

	mimeType := http.DetectContentType(imgBtyes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBtyes)

	return base64Encoding, nil
}

// getImageFormat retrieve format a.k.a name from decode config
func getImageFormat(r io.Reader) (string, error) {
	_, format, err := image.DecodeConfig(r)
	return format, err
}

// getImageMimeType Get image mime type (gif/jpeg/png/webp)
func getImageMimeType(r io.Reader) string {
	format, _ := getImageFormat(r)
	if format == "" {
		return ""
	}
	return mime.TypeByExtension("." + format)
}

// BlurImage converts image bytes into a blurred base64 string
func BlurImage(imgBytes []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	var img image.Image
	var err error

	imageMime := getImageMimeType(bytes.NewReader(imgBytes))

	switch imageMime {
	case "image/webp":
		img, err = webp.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			log.Error("could not decode image", "image mime type", imageMime, "err", err)
			return buf.Bytes(), err
		}
	default:
		img, err = imaging.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			log.Error("could not decode image", "image mime type", imageMime, "err", err)
			return buf.Bytes(), err
		}
	}

	blurredImg := imaging.Blur(img, 20)
	blurredImg = imaging.AdjustBrightness(blurredImg, -20)

	err = imaging.Encode(buf, blurredImg, imaging.JPEG)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// CombineQueries combine URL.Query() and Referer() queries
// NOTE: Referer queries will overwrite URL queries
func CombineQueries(urlQueries url.Values, refererURL string) (url.Values, error) {

	queries := urlQueries

	referer, err := url.Parse(refererURL)
	if err != nil {
		log.Error("Error parsing URL", "url", refererURL, "err", err)
		return queries, fmt.Errorf("Could not read URL. Is it formatted correctly?")
	}

	// Combine referer into values
	for key, vals := range referer.Query() {
		for _, val := range vals {
			queries.Add(key, val)
		}
	}

	return queries, nil

}

// RandomItem returns a random item from given slice
func RandomItem[T any](s []T) T {

	var out T

	if len(s) == 0 {
		return out
	}

	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	return s[0]
}


type Color struct {
	R   int
	G   int
	B   int
	RGB string
	Hex string
}

// StringToColor takes any string and returns a Color struct. 
// Identical strings should return identical values
func StringToColor(inputString string) Color {
	sum := 0
	for _, char := range inputString {
		sum += int(char)
	}

	// Helper function to calculate a color component
	calcColor := func(offset int) int {
		return int(math.Floor((math.Sin(float64(sum+offset)) + 1) * 128))
	}

	r := calcColor(1)
	g := calcColor(2)
	b := calcColor(3)

	rgb := fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
	hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)

	return Color{R: r, G: g, B: b, RGB: rgb, Hex: hex}
}


