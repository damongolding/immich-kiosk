package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"math"
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

	"github.com/charmbracelet/lipgloss"
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

func ColorizeRequestId(requestId string) string {

	c := StringToColor(requestId)

	textWhite := calculateContrastRatio(Color{R: 255, G: 255, B: 255}, c)
	textBlack := calculateContrastRatio(Color{R: 0, G: 0, B: 0}, c)

	textColor := lipgloss.Color("#000000")
	if textWhite > textBlack {
		textColor = lipgloss.Color("#ffffff")
	}

	return lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(textColor).Background(lipgloss.Color(c.Hex)).Render(requestId)
}

// calculateContrastRatio computes the contrast ratio between two RGB colors.
func calculateContrastRatio(color1, color2 Color) float64 {
	lum1 := calculateLuminance(color1)
	lum2 := calculateLuminance(color2)

	if lum1 > lum2 {
		return (lum1 + 0.05) / (lum2 + 0.05)
	}
	return (lum2 + 0.05) / (lum1 + 0.05)
}

// calculateLuminance calculates the relative luminance of an RGB color.
func calculateLuminance(color Color) float64 {
	r := linearize(float64(color.R) / 255.0)
	g := linearize(float64(color.G) / 255.0)
	b := linearize(float64(color.B) / 255.0)

	return 0.2126*r + 0.7152*g + 0.0722*b
}

// linearize converts an sRGB component to a linear value.
func linearize(value float64) float64 {
	if value <= 0.03928 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}
