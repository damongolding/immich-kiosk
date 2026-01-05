// Package utils provides a collection of helper functions for various tasks.
//
// This package includes utilities for UUID generation, date formatting,
// image processing, URL query manipulation, random selection, color operations,
// and request ID colorization. It's designed to support common operations
// across different parts of the application.
package utils

import (
	"bytes"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"math/rand/v2"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/webp"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/disintegration/imaging"

	"github.com/google/uuid"

	"github.com/skip2/go-qrcode"
)

const (

	// sigmaConstant is used to normalise the blur effect across different image sizes.
	// The value 1300.0 was chosen as it provides consistent blur effects for typical screen resolutions.
	sigmaConstant          float64 = 1300.0
	blurredImageBrightness float64 = -20

	// minMemoryWeight is the minimum weight allowed for memory assets.
	minMemoryWeight float64 = 0.0001
)

// WeightedAsset represents an asset with a type and ID
type WeightedAsset struct {
	Type kiosk.Source
	ID   string
	Name string
}

// AssetWithWeighting represents a WeightedAsset with an associated weight value
type AssetWithWeighting struct {
	Asset   WeightedAsset
	Weight  int     // base weight
	Penalty float64 // penalty for weight. 1.0 = no penalty, 0.5 = 50% penalty.
}

// GenerateUUID generates a new random UUID string
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

// DateToJavascriptLayout converts a date format string from Go layout to JavaScript format
func DateToJavascriptLayout(input string) string {
	replacer := strings.NewReplacer(
		"YYYY", "yyyy",
		"YY", "yy",
		"DDDD", "eeee",
		"DDD", "eee",
		"DD", "dd",
		"D", "d",
	)
	return replacer.Replace(input)
}

// ImageToBytes converts an image.Image to a byte slice in JPEG format.
// It takes an image.Image as input and returns the encoded bytes and any error encountered.
// The bytes can be used for further processing, transmission, or storage.
func ImageToBytes(img image.Image) ([]byte, error) {

	buf := new(bytes.Buffer)

	err := imaging.Encode(buf, img, imaging.JPEG)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// BytesToImage converts a byte slice to an image.Image.
// It takes a byte slice as input and returns an image.Image and any error encountered.
// It handles both WebP and other common image formats (JPEG, PNG, GIF) automatically
// by detecting the MIME type and using the appropriate decoder.
func BytesToImage(imgBytes []byte, isOriginal bool) (image.Image, error) {

	var img image.Image
	var err error

	imageMime := ImageMimeType(bytes.NewReader(imgBytes))

	switch imageMime {
	case "image/webp":
		img, err = webp.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			log.Error("could not decode image", "image mime type", imageMime, "err", err)
			return nil, err
		}
	default:
		img, err = imaging.Decode(bytes.NewReader(imgBytes), imaging.AutoOrientation(isOriginal))
		if err != nil {
			log.Error("could not decode image", "image mime type", imageMime, "err", err)
			return nil, err
		}
	}

	return img, nil
}

// ApplyExifOrientation adjusts an image's orientation based on EXIF data.
// It takes an image and an EXIF orientation string.
// The EXIF orientation values follow the standard specification:
//
//	1 = Normal
//	2 = Flipped horizontally
//	3 = Rotated 180 degrees
//	4 = Flipped vertically
//	5 = Transposed (flipped horizontally and rotated 90° CW)
//	6 = Rotated 90° CCW
//	7 = Transverse (flipped horizontally and rotated 90° CCW)
//	8 = Rotated 90° CW
//
// Returns the properly oriented image.
func ApplyExifOrientation(img image.Image, exifOrientation string) image.Image {

	if img == nil {
		return nil
	}

	o, err := strconv.Atoi(exifOrientation)
	if err != nil {
		return img
	}

	switch o {
	case 1:
		return img
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

// ImageToBase64 converts an image.Image to a base64 encoded data URI string with appropriate MIME type
func ImageToBase64(img image.Image) (string, error) {

	var buf bytes.Buffer

	err := imaging.Encode(&buf, img, imaging.JPEG)
	if err != nil {
		return "", err
	}

	var base64Encoding string

	mimeType := http.DetectContentType(buf.Bytes())

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(buf.Bytes())

	return base64Encoding, nil
}

// BytesToBase64 converts a byte slice to a base64 encoded string with MIME type prefix.
// It takes a byte slice representing an image and returns a data URI string suitable
// for use in HTML/CSS, such as "data:image/jpeg;base64,/9j/4AAQSkZJ...".
// The function detects the MIME type of the image automatically.
func BytesToBase64(imgBytes []byte) (string, error) {
	var base64Encoding string

	mimeType := http.DetectContentType(imgBytes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBytes)

	return base64Encoding, nil
}

// imageFormat retrieves the format name from the image decode config
func imageFormat(r io.Reader) (string, error) {
	_, format, err := image.DecodeConfig(r)
	return format, err
}

// ImageMimeType returns the MIME type (gif/jpeg/png/webp) for an image reader
func ImageMimeType(r io.Reader) string {
	format, err := imageFormat(r)
	if err != nil || format == "" {
		log.Error("getting mime", "err", err)
		return ""
	}

	return mime.TypeByExtension("." + format)
}

// BlurImage applies a Gaussian blur to an image with normalized sigma based on image dimensions.
// It can optionally resize the image first based on client data dimensions.
func BlurImage(img image.Image, blurrAmount int, isOptimized bool, clientWidth, clientHeight int) (image.Image, error) {

	blurredImage := img

	if clientWidth != 0 && clientHeight != 0 && !isOptimized {
		blurredImage = imaging.Fit(blurredImage, clientWidth, clientHeight, imaging.Lanczos)
	}

	sigma := calculateNormalizedSigma(blurrAmount, blurredImage.Bounds().Dx(), blurredImage.Bounds().Dy(), sigmaConstant)

	blurredImage = imaging.Blur(blurredImage, sigma)
	blurredImage = imaging.AdjustBrightness(blurredImage, blurredImageBrightness)

	return blurredImage, nil
}

// CombineQueries combines URL.Query() and Referer() queries into a single url.Values.
// Referer query parameters will overwrite URL query parameters with the same names.
func CombineQueries(urlQueries url.Values, refererURL string) (url.Values, error) {

	queries := urlQueries

	referer, err := url.Parse(refererURL)
	if err != nil {
		log.Error("parsing URL", "url", refererURL, "err", err)
		return queries, errors.New("could not read URL. Is it formatted correctly?")
	}

	// Combine referer into values
	for key, vals := range referer.Query() {
		for _, val := range vals {
			queries.Add(key, val)
		}
	}

	return queries, nil
}

// MergeQueries combines two url.Values objects into a single new url.Values object.
// Values from both input objects are preserved while avoiding duplicate values.
// The function processes values from urlQueriesB first, followed by values from urlQueriesA.
// For each key-value pair, duplicates are detected and only unique values are included in the result.
// A value is considered a duplicate if the same key-value combination already exists in the merged result.
// The function uses a nested map structure to efficiently track seen values and prevent duplicates.
// Returns a new url.Values object containing the combined unique values from both inputs.
func MergeQueries(urlQueriesA, urlQueriesB url.Values) url.Values {
	merged := make(url.Values)
	seen := make(map[string]map[string]bool)

	// Helper function to process values
	processValues := func(values url.Values) {
		for key, vals := range values {
			if seen[key] == nil {
				seen[key] = make(map[string]bool)
			}
			for _, value := range vals {
				if !seen[key][value] {
					merged[key] = append(merged[key], value)
					seen[key][value] = true
				}
			}
		}
	}

	processValues(urlQueriesB)
	processValues(urlQueriesA)

	return merged
}

// RandomItem returns a random item from the given slice.
// Returns the zero value of type T if the slice is empty.
func RandomItem[T any](s []T) T {
	var out T

	if len(s) == 0 {
		return out
	}

	copySlice := slices.Clone(s)

	rand.Shuffle(len(copySlice), func(i, j int) {
		copySlice[i], copySlice[j] = copySlice[j], copySlice[i]
	})

	return copySlice[0]
}

func assetWeight(a AssetWithWeighting) float64 {

	weight := max(0, a.Weight)

	// Base logarithmic weight
	base := math.Log(float64(weight) + 1)

	// Default penalty
	penalty := a.Penalty
	if penalty <= 0 {
		penalty = 1.0
	}

	final := base * penalty

	// Never allow zero or negative weight
	final = max(final, minMemoryWeight)

	return final
}

func calculateTotalWeight(assets []AssetWithWeighting) float64 {
	total := 0.0
	for _, asset := range assets {
		total += assetWeight(asset)
	}
	return total
}

func WeightedRandomItem(assets []AssetWithWeighting) WeightedAsset {
	switch len(assets) {
	case 0:
		return WeightedAsset{}
	case 1:
		return assets[0].Asset
	}

	totalWeight := calculateTotalWeight(assets)
	r := rand.Float64() * totalWeight

	for _, asset := range assets {
		w := assetWeight(asset)
		if r < w {
			return asset.Asset
		}
		r -= w
	}

	// Should never happen, but keep a safe fallback
	return assets[0].Asset
}

// Color represents an RGB color with string representations
type Color struct {
	RGB string
	Hex string
	R   int
	G   int
	B   int
}

// StringToColor takes any string and returns a Color struct with deterministic RGB values.
// Identical input strings will always return identical color values.
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

// ColorizeRequestID takes a request ID string and returns a colorized string representation.
// It generates a color based on the input string, determines the best contrasting text color,
// and applies styling using lipgloss to create a visually distinct, colored representation of the request ID.
func ColorizeRequestID(requestID string) string {

	c := StringToColor(requestID)

	textWhite := calculateContrastRatio(Color{R: 255, G: 255, B: 255}, c)
	textBlack := calculateContrastRatio(Color{R: 0, G: 0, B: 0}, c)

	textColor := lipgloss.Color("#000000")
	if textWhite > textBlack {
		textColor = lipgloss.Color("#ffffff")
	}

	requestID = strings.ToUpper(requestID)

	if len(requestID) > 4 {
		return lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(textColor).Background(lipgloss.Color(c.Hex)).Render(requestID[len(requestID)-4:])
	}

	return lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(textColor).Background(lipgloss.Color(c.Hex)).Render(requestID)
}

// calculateContrastRatio computes the contrast ratio between two RGB colors according to WCAG 2.0.
func calculateContrastRatio(color1, color2 Color) float64 {
	lum1 := calculateLuminance(color1)
	lum2 := calculateLuminance(color2)

	if lum1 > lum2 {
		return (lum1 + 0.05) / (lum2 + 0.05)
	}
	return (lum2 + 0.05) / (lum1 + 0.05)
}

// calculateLuminance calculates the relative luminance of an RGB color according to WCAG 2.0.
func calculateLuminance(color Color) float64 {
	r := linearize(float64(color.R) / 255.0)
	g := linearize(float64(color.G) / 255.0)
	b := linearize(float64(color.B) / 255.0)

	return 0.2126*r + 0.7152*g + 0.0722*b
}

// linearize converts an sRGB component to a linear value according to the sRGB color space specification.
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

// parseTimeString parses a time string in various formats and returns a time.Time value.
// It accepts formats like "1", "12", "130", "1430" and converts them to hours and minutes.
func parseTimeString(timeStr string) (time.Time, error) {

	// Trim whitespace and validate
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return time.Time{}, errors.New("invalid time format: empty or whitespace-only input")
	}

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

// IsSleepTime checks if the current time falls within a sleep period defined by start and end times.
// It handles periods that cross midnight by adjusting the times accordingly.
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

// FileExists checks if a file exists at the specified path and returns true if it does
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CreateQrCode generates a QR code for the given link and returns it as a base64 encoded string.
// Returns an empty string and logs an error if generation fails.
func CreateQrCode(link string) string {

	if link == "" {
		log.Error("QR code generation failed: empty link provided")
		return ""
	}

	if _, err := url.Parse(link); err != nil {
		log.Error("QR code generation failed: invalid URL", "link", link, "err", err)
		return ""
	}

	png, err := qrcode.Encode(link, qrcode.Low, 128)
	if err != nil {
		log.Error("QR code generation failed", "link", link, "err", err)
		return ""
	}

	i, err := BytesToBase64(png)
	if err != nil {
		log.Error("QR code base64 encoding failed", "link", link, "err", err)
		return ""
	}

	return i
}

// GenerateSharedSecret generates a random 256-bit (32-byte) secret and returns it as a hex string.
func GenerateSharedSecret() (string, error) {
	secret := make([]byte, 32)
	_, err := crand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return hex.EncodeToString(secret), nil
}

// CalculateSignature generates an HMAC-SHA256 signature for the given secret and timestamp
func CalculateSignature(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// IsValidSignature performs a constant-time comparison of two signatures to prevent timing attacks
func IsValidSignature(receivedSignature, calculatedSignature string) bool {
	received, err := hex.DecodeString(receivedSignature)
	if err != nil {
		return false
	}
	calculated, err := hex.DecodeString(calculatedSignature)
	if err != nil {
		return false
	}
	return hmac.Equal(received, calculated)
}

// IsValidTimestamp validates if a timestamp is within the acceptable tolerance window
func IsValidTimestamp(receivedTimestamp string, toleranceSeconds int) bool {
	ts, err := strconv.ParseInt(receivedTimestamp, 10, 64)
	if err != nil {
		return false
	}
	currentTime := time.Now().Unix()
	return abs(currentTime-ts) <= int64(toleranceSeconds)
}

// abs returns the absolute value of an int64
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// OptimizeImage resizes an image to the specified dimensions while maintaining aspect ratio.
// If width or height is 0, the image is returned unmodified.
func OptimizeImage(img image.Image, width, height int) (image.Image, error) {

	optimizedImage := img

	if width != 0 && height != 0 {
		optimizedImage = imaging.Fit(img, width, height, imaging.Lanczos)
	}

	return optimizedImage, nil
}

// calculateNormalizedSigma calculates a normalized sigma value for Gaussian blur based on image dimensions.
// The formula uses the diagonal length of the image (sqrt(width² + height²)) to adjust the blur intensity,
// ensuring consistent visual effects across different image sizes. The constant value helps maintain
// a balanced blur effect for typical screen resolutions.
//
// The formula is: sigma = baseSigma * sqrt(width² + height²) / constant
func calculateNormalizedSigma(baseSigma int, width, height int, constant float64) float64 {
	diagonal := math.Sqrt(float64(width*width + height*height))
	return float64(baseSigma) * diagonal / constant
}

// SystemLanguage detects the system's primary language by examining environment
// variables LANG, LC_ALL, and LC_MESSAGES in order of precedence.
//
// The function parses environment variable formats like:
// - "en_US.UTF-8"
// - "fr_FR.ISO8859-1"
// - "de_DE"
//
// For each environment variable, it:
// 1. Extracts the language code before any '.' separator
// 2. Splits on '_' to separate language and region codes
// 3. Normalizes to lowercase language code and uppercase region code
//
// Example outputs:
// - "en_US" (from "en_US.UTF-8")
// - "fr_FR" (from "fr_FR.ISO8859-1")
//
// Returns:
// - Normalized "language_REGION" code if valid language found
// - "en_GB" as fallback if no valid system language detected
func SystemLanguage() string {
	for _, envVar := range []string{"LANG", "LC_ALL", "LC_MESSAGES"} {
		if lang := os.Getenv(envVar); lang != "" {
			if parts := strings.Split(lang, ".")[0]; parts != "" {
				if code := strings.Split(parts, "_"); len(code) == 2 {
					return strings.ToLower(code[0]) + "_" + strings.ToUpper(code[1])
				}
			}
		}
	}

	return "en_GB"
}

// TrimHistory ensures that the history slice doesn't exceed the specified maximum length.
func TrimHistory(history *[]string, maxLength int) {
	if len(*history) > maxLength {
		*history = (*history)[len(*history)-maxLength:]
	}
}

// IsTimeBetween checks if a given time falls between a start and end time, inclusive.
// checkTime: The time to check
// startTime: The beginning of the time range
// endTime: The end of the time range
// Returns true if checkTime is equal to or after startTime AND equal to or before endTime
func IsTimeBetween(checkTime, startTime, endTime time.Time) bool {
	return (checkTime.Equal(startTime) || checkTime.After(startTime)) &&
		(checkTime.Equal(endTime) || checkTime.Before(endTime))
}

// Helper function to get days in a month
func DaysInMonth(date time.Time) int {
	return time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ParseSize converts a human-readable size string (e.g., "10MB", "1GB") to bytes
// using binary prefixes (1KB = 1024B, 1MB = 1024KB, etc.)
func ParseSize(sizeStr string) (int64, error) {

	if sizeStr == "0" {
		return 0, nil
	}

	re := regexp.MustCompile(`^(\d+)\s*([BKMGbkmg][Bb]?)$`)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	bytes, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	unit := strings.ToUpper(matches[2])
	switch unit {
	case "B":
		return bytes, nil
	case "KB", "K":
		return bytes * 1024, nil
	case "MB", "M":
		return bytes * 1024 * 1024, nil
	case "GB", "G":
		return bytes * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}
}

// CleanDirectory removes all files and directories within the specified directory path.
// It reads the directory contents and recursively removes each entry using os.RemoveAll.
// Returns an error if the directory cannot be read or if any removal operation fails.
func CleanDirectory(dirPath string) error {
	if dirPath == "" || dirPath == "/" {
		return fmt.Errorf("refusing to clean unsafe directory %q", dirPath)
	}

	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, d := range dir {
		err = os.RemoveAll(filepath.Join(dirPath, d.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveDuplicatesInPlace modifies the first slice by removing any strings that are present in the second slice.
// It uses a map for O(1) lookups of strings in the second slice and performs the removal in-place.
// The operation preserves the relative order of remaining elements in the first slice.
// The first slice is modified directly and truncated to the new length containing only unique elements.
func RemoveDuplicatesInPlace(slice1 *[]string, slice2 []string) {
	lookup := make(map[string]struct{})
	for _, s := range slice2 {
		lookup[s] = struct{}{}
	}

	j := 0
	for i := range *slice1 {
		if _, exists := lookup[(*slice1)[i]]; !exists {
			(*slice1)[j] = (*slice1)[i]
			j++
		}
	}
	*slice1 = (*slice1)[:j]
}

// Converts RGB (0–255) to HSL
func rgbToHsl(r, g, b uint8) (float64, float64, float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	maximum := math.Max(rf, math.Max(gf, bf))
	minimum := math.Min(rf, math.Min(gf, bf))

	var h, s, l float64

	l = (maximum + minimum) / 2

	if maximum == minimum {
		h, s = 0, 0 // achromatic
	} else {
		d := maximum - minimum
		if l > 0.5 {
			s = d / (2.0 - maximum - minimum)
		} else {
			s = d / (maximum + minimum)
		}
		switch maximum {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}
	return h, s, l
}

// Converts HSL to RGB (0–255)
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	var rF, gF, bF float64

	if s == 0 {
		return uint8(l * 255), uint8(l * 255), uint8(l * 255)
	}

	var hueToRgb = func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6 {
			return p + (q-p)*6*t
		}
		if t < 1.0/2 {
			return q
		}
		if t < 2.0/3 {
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}

	q := l * (1 + s)
	if l >= 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q
	rF = hueToRgb(p, q, h+1.0/3)
	gF = hueToRgb(p, q, h)
	bF = hueToRgb(p, q, h-1.0/3)

	return uint8(rF * 255), uint8(gF * 255), uint8(bF * 255)
}

// Adjust lightness toward dark
func darkenColor(c color.Color, minL float64) color.RGBA {
	r, g, b, a := c.RGBA()
	h, s, l := rgbToHsl(uint8(r>>8), uint8(g>>8), uint8(b>>8))

	if l > minL {
		l = minL // Clamp to darker value (e.g., 0.3)
	}

	r8, g8, b8 := hslToRgb(h, s, l)
	return color.RGBA{r8, g8, b8, uint8(a >> 8)}
}

func ExtractDominantColor(img image.Image) (color.RGBA, error) {
	colours, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping, img)
	if err != nil {
		return color.RGBA{}, err
	}

	if len(colours) == 0 {
		return color.RGBA{}, errors.New("no prominent colors found")
	}

	return darkenColor(color.RGBA{
		R: uint8(colours[0].Color.R),
		G: uint8(colours[0].Color.G),
		B: uint8(colours[0].Color.B),
		A: 255,
	}, 0.3), nil
}
