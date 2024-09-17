package routes

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// processImage processes an image based on the given configuration and returns the image bytes.
// It selects an image from either albums or people based on the configuration and asset weighting.
//
// Parameters:
//   - immichImage: A pointer to the ImmichAsset to process.
//   - requestConfig: The configuration for the current request.
//   - requestId: A unique identifier for the request.
//   - kioskDeviceId: The ID of the kiosk device.
//   - isPrefetch: A boolean indicating if this is a prefetch operation.
//
// Returns:
//   - []byte: The processed image bytes.
//   - error: An error if any occurred during processing.
func processImage(immichImage *immich.ImmichAsset, requestConfig config.Config, requestId string, kioskDeviceId string, isPrefetch bool) ([]byte, error) {
	var imgBytes []byte

	peopleAndAlbums := []immich.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		personAssetCount, err := immichImage.PersonImageCount(person, requestId)
		if err != nil {
			return imgBytes, fmt.Errorf("getting person image count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, immich.AssetWithWeighting{
			Asset:  immich.WeightedAsset{Type: "PERSON", ID: person},
			Weight: personAssetCount,
		})
	}

	for _, album := range requestConfig.Album {
		albumAssetCount, err := immichImage.AlbumImageCount(album, requestId)
		if err != nil {
			return imgBytes, fmt.Errorf("getting album asset count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, immich.AssetWithWeighting{
			Asset:  immich.WeightedAsset{Type: "ALBUM", ID: album},
			Weight: albumAssetCount,
		})
	}

	var pickedImage immich.WeightedAsset

	if requestConfig.Kiosk.AssetWeighting {
		pickedImage = utils.WeightedRandomItem(peopleAndAlbums)
	} else {
		var assetsOnly []immich.WeightedAsset
		for _, item := range peopleAndAlbums {
			assetsOnly = append(assetsOnly, item.Asset)
		}

		pickedImage = utils.RandomItem(assetsOnly)
	}

	log.Debug("picker", "pool", peopleAndAlbums, "picked", pickedImage)

	var err error
	switch pickedImage.Type {
	case "ALBUM":
		err = immichImage.RandomImageFromAlbum(pickedImage.ID, requestId, kioskDeviceId, isPrefetch)
	case "PERSON":
		err = immichImage.RandomImageOfPerson(pickedImage.ID, requestId, kioskDeviceId, isPrefetch)
	default:
		err = immichImage.RandomImage(requestId, kioskDeviceId, isPrefetch)
	}

	if err != nil {
		return imgBytes, fmt.Errorf("getting image: %w", err)
	}

	imageGet := time.Now()
	imgBytes, err = immichImage.ImagePreview()
	if err != nil {
		return imgBytes, fmt.Errorf("getting image preview: %w", err)
	}

	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceId, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestId, "Got image in", time.Since(imageGet).Seconds())
	}

	return imgBytes, err
}

// processPageData processes the page data for an image request.
// It handles image conversion, blurring (if configured), and prepares the PageData struct.
//
// Parameters:
//   - requestConfig: The configuration for the current request.
//   - c: The echo.Context for the current request.
//   - isPrefetch: A boolean indicating if this is a prefetch operation.
//
// Returns:
//   - views.PageData: The processed page data.
//   - error: An error if any occurred during processing.
func processPageData(requestConfig config.Config, c echo.Context, isPrefetch bool) (views.PageData, error) {
	requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceId := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)

	imgBytes, err := processImage(&immichImage, requestConfig, requestId, kioskDeviceId, isPrefetch)
	if err != nil {
		return views.PageData{}, fmt.Errorf("converting image to base64: %w", err)
	}

	imageConvertTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return views.PageData{}, fmt.Errorf("converting image to base64: %w", err)
	}
	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceId, "Converted image in", time.Since(imageConvertTime).Seconds())
	} else {
		log.Debug(requestId, "Converted image in", time.Since(imageConvertTime).Seconds())
	}

	var imgBlur string
	if requestConfig.BackgroundBlur && strings.ToLower(requestConfig.ImageFit) != "cover" {
		imageBlurTime := time.Now()
		imgBlurBytes, err := utils.BlurImage(imgBytes)
		if err != nil {
			return views.PageData{}, fmt.Errorf("blurring image: %w", err)
		}
		imgBlur, err = utils.ImageToBase64(imgBlurBytes)
		if err != nil {
			return views.PageData{}, fmt.Errorf("converting blurred image to base64: %w", err)
		}
		if isPrefetch {
			log.Debug(requestId, "PREFETCH", kioskDeviceId, "Blurred image in", time.Since(imageBlurTime).Seconds())
		} else {
			log.Debug(requestId, "Blurred image in", time.Since(imageBlurTime).Seconds())
		}
	}

	if len(requestConfig.History) > 10 {
		requestConfig.History = requestConfig.History[len(requestConfig.History)-10:]
	}

	return views.PageData{
		ImmichImage:   immichImage,
		ImageData:     img,
		ImageBlurData: imgBlur,
		Config:        requestConfig,
	}, nil
}

// imagePreFetch prefetches a specified number of images and caches them for future use.
// This function improves performance by preparing images in advance, reducing load times for subsequent requests.
//
// Parameters:
//   - numberOfImages: The number of images to prefetch and cache.
//   - requestConfig: Configuration for the current request.
//   - c: The echo.Context for the current request.
//   - kioskDeviceId: The unique identifier for the kiosk device.
//
// The function creates a worker pool to concurrently process and cache the specified number of images.
// Cached images are stored with a key that combines the request URL and kiosk device ID.
func imagePreFetch(numberOfImages int, requestConfig config.Config, c echo.Context, kioskDeviceId string) {

	var wg sync.WaitGroup

	wg.Add(numberOfImages)

	for range make([]struct{}, numberOfImages) {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			pageData, err := processPageData(requestConfig, c, true)
			if err != nil {
				log.Error("prefetch", "err", err)
				return
			}

			cacheKey := c.Request().URL.String() + kioskDeviceId

			cachedPageData := []views.PageData{}

			if data, found := pageDataCache.Get(cacheKey); found {
				cachedPageData = data.([]views.PageData)
			}

			cachedPageData = append(cachedPageData, pageData)

			pageDataCache.Set(cacheKey, cachedPageData, cache.DefaultExpiration)
		}(&wg)

	}

	wg.Wait()
}

// NewImage returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
//
// Parameters:
//   - baseConfig: A pointer to the base configuration.
//
// Returns:
//   - echo.HandlerFunc: A function that handles the image request.
func NewImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceId := c.Request().Header.Get("kiosk-device-id")
		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.String(http.StatusTemporaryRedirect, "")
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"deviceID", kioskDeviceId,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			cacheKey := c.Request().URL.String() + kioskDeviceId
			if data, found := pageDataCache.Get(cacheKey); found {
				log.Debug(
					requestId,
					"deviceID", kioskDeviceId,
					"cache hit for new image", true,
				)
				cachedPageData := data.([]views.PageData)
				if len(cachedPageData) != 0 {
					log.Debug("number of images in cache", "items", len(cachedPageData))
					nextPageData := cachedPageData[0]
					pageDataCache.Set(cacheKey, cachedPageData[1:], cache.DefaultExpiration)
					go imagePreFetch(1, requestConfig, c, kioskDeviceId)
					return Render(c, http.StatusOK, views.Image(nextPageData))
				}
			}
			log.Debug(
				requestId,
				"deviceID", kioskDeviceId,
				"cache miss for new image", false,
			)
		}

		pageData, err := processPageData(requestConfig, c, false)
		if err != nil {
			log.Error("processing image", "err", err)
			return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "Error processing image", Message: err.Error()}))
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(1, requestConfig, c, kioskDeviceId)
		}

		return Render(c, http.StatusOK, views.Image(pageData))
	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
//
// Parameters:
//   - baseConfig: A pointer to the base configuration.
//
// Returns:
//   - echo.HandlerFunc: A function that handles the raw image request.
func NewRawImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichImage := immich.NewImage(requestConfig)

		imgBytes, err := processImage(&immichImage, requestConfig, requestId, "", false)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}
}
