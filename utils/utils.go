// Package utils provides a collection of helper functions for various tasks.
//
// This package includes utilities for UUID generation, date formatting,
// image processing, URL query manipulation, random selection, color operations,
// and request ID colorization. It's designed to support common operations
// across different parts of the application.
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
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/webp"
	_ "golang.org/x/image/webp"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/disintegration/imaging"

	"github.com/google/uuid"
)

type WeightedAsset struct {
	Type string
	ID   string
}

type AssetWithWeighting struct {
	Asset  WeightedAsset
	Weight int
}

// GenerateUUID generates as UUID
func GenerateUUID() string {
	return uuid.New().String()
}

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
		log.Error("parsing URL", "url", refererURL, "err", err)
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

// calculateTotalWeight calculates the sum of logarithmic weights for all assets in the given slice.
// It uses natural logarithm (base e) and adds 1 to avoid log(0).
func calculateTotalWeight(assets []AssetWithWeighting) int {
	total := 0
	for _, asset := range assets {
		logWeight := int(math.Log(float64(asset.Weight) + 1))
		if logWeight == 0 {
			logWeight = 1
		}
		total += logWeight
	}
	return total
}

// WeightedRandomItem selects a random asset from the given slice of WeightedAsset(s)
// based on their logarithmic weights. It uses a weighted random selection algorithm.
func WeightedRandomItem(assets []AssetWithWeighting) WeightedAsset {

	// guards
	switch len(assets) {
	case 0:
		return WeightedAsset{}
	case 1:
		return assets[0].Asset
	}

	totalWeight := calculateTotalWeight(assets)
	randomWeight := rand.IntN(totalWeight) + 1

	for _, asset := range assets {
		logWeight := int(math.Log(float64(asset.Weight) + 1))
		if randomWeight <= logWeight {
			return asset.Asset
		}
		randomWeight -= logWeight
	}

	// WeightedRandomItem sometimes returns an empty WeightedAsset
	// when the random selection process fails to pick an item.
	// This is a fallback to ensure we always return a valid asset.
	if len(assets) > 0 {
		return assets[0].Asset
	}
	return WeightedAsset{}
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

// ColorizeRequestId takes a request ID string and returns a colorized string representation.
// It generates a color based on the input string, determines the best contrasting text color,
// and applies styling using lipgloss to create a visually distinct, colored representation of the request ID.
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

// PickRandomImageType selects a random image type based on the given configuration and weightings.
// It returns a WeightedAsset representing the picked image type.
func PickRandomImageType(useWeighting bool, peopleAndAlbums []AssetWithWeighting) WeightedAsset {

	var pickedImage WeightedAsset

	if useWeighting {
		pickedImage = WeightedRandomItem(peopleAndAlbums)
	} else {
		var assetsOnly []WeightedAsset
		for _, item := range peopleAndAlbums {
			assetsOnly = append(assetsOnly, item.Asset)
		}
		pickedImage = RandomItem(assetsOnly)
	}

	return pickedImage
}

func parseTimeString(timeStr string) (time.Time, error) {
	// Extract only the digits
	digits := regexp.MustCompile(`\d`).FindAllString(timeStr, -1)

	if len(digits) == 0 {
		return time.Time{}, fmt.Errorf("invalid time format: no digits found in %s", timeStr)
	}

	// Join the digits
	timeStr = strings.Join(digits, "")

	var hours, minutes int
	var err error

	switch len(timeStr) {
	case 1, 2:
		// Interpret as hours
		hours, err = strconv.Atoi(timeStr)
		if err != nil || hours >= 24 {
			return time.Time{}, fmt.Errorf("invalid hours: %s", timeStr)
		}
	case 3:
		// Interpret as 1 digit hour and 2 digit minute
		hours, err = strconv.Atoi(timeStr[:1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid hours: %s", timeStr[:1])
		}
		minutes, err = strconv.Atoi(timeStr[1:])
		if err != nil || minutes >= 60 {
			return time.Time{}, fmt.Errorf("invalid minutes: %s", timeStr[1:])
		}
	case 4:
		// Interpret as 2 digit hour and 2 digit minute
		hours, err = strconv.Atoi(timeStr[:2])
		if err != nil || hours >= 24 {
			return time.Time{}, fmt.Errorf("invalid hours: %s", timeStr[:2])
		}
		minutes, err = strconv.Atoi(timeStr[2:])
		if err != nil || minutes >= 60 {
			return time.Time{}, fmt.Errorf("invalid minutes: %s", timeStr[2:])
		}
	default:
		// Truncate to 4 digits if longer
		hours, err = strconv.Atoi(timeStr[:2])
		if err != nil || hours >= 24 {
			return time.Time{}, fmt.Errorf("invalid hours: %s", timeStr[:2])
		}
		minutes, err = strconv.Atoi(timeStr[2:4])
		if err != nil || minutes >= 60 {
			return time.Time{}, fmt.Errorf("invalid minutes: %s", timeStr[2:4])
		}
	}

	// Create time.Time object
	return time.Date(0, 1, 1, hours, minutes, 0, 0, time.UTC), nil
}

func IsSleepTime(sleepStartTime, sleepEndTime string, currentTime time.Time) (bool, error) {
	// Parse start and end times
	startTime, err := parseTimeString(sleepStartTime)
	if err != nil {
		log.Error("parsing sleep start time:", err)
		return false, err
	}

	endTime, err := parseTimeString(sleepEndTime)
	if err != nil {
		log.Error("parsing sleep end time:", err)
		return false, err
	}

	// Set the date of startTime and endTime to the same as currentTime
	year, month, day := currentTime.Date()
	startTime = time.Date(year, month, day, startTime.Hour(), startTime.Minute(), 0, 0, currentTime.Location())
	endTime = time.Date(year, month, day, endTime.Hour(), endTime.Minute(), 0, 0, currentTime.Location())

	// If end time is before start time, it means the period crosses midnight
	if endTime.Before(startTime) {
		endTime = endTime.Add(24 * time.Hour)
	}

	// Check if current time is between start and end times
	if currentTime.Before(startTime) {
		currentTime = currentTime.Add(24 * time.Hour)
	}

	return (currentTime.After(startTime) || currentTime.Equal(startTime)) &&
		currentTime.Before(endTime), nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
