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
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/webp"
	_ "golang.org/x/image/webp"

	"github.com/charmbracelet/log"
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

	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	return s[0]
}

// Function to get the name of a function
func getFunctionName(i interface{}) string {
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(fullFuncName, ".")
	return strings.Replace(parts[len(parts)-1], "-fm", "", -1)
}

// Decorator function to measure execution time
func TimeTrackDecorator(fn interface{}, requestId string) interface{} {
	originalFunc := reflect.ValueOf(fn)
	if originalFunc.Kind() != reflect.Func {
		panic("Decorator can only be applied to functions")
	}

	originalType := originalFunc.Type()
	funcName := getFunctionName(fn)

	decoratedFunc := reflect.MakeFunc(originalType, func(args []reflect.Value) []reflect.Value {
		log.Debug("Profiling: " + funcName + "()")
		cpuprofile1 := "./pprof/cpu-" + funcName + ".pprof"
		f, err := os.Create(cpuprofile1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			panic(err)
		}

		start := time.Now()

		// Call the original function
		results := originalFunc.Call(args)

		pprof.StopCPUProfile()

		elapsed := time.Since(start)
		log.Debug(requestId+" "+funcName, "done in", elapsed.Seconds())

		return results
	})

	return decoratedFunc.Interface()
}
