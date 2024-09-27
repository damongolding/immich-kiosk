package routes

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

// gatherPeopleAndAlbums collects asset weightings for people and albums.
// It returns a slice of AssetWithWeighting and an error if any occurs during the process.
func gatherPeopleAndAlbums(immichImage *immich.ImmichAsset, requestConfig config.Config, requestID string) ([]utils.AssetWithWeighting, error) {
	peopleAndAlbums := []utils.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		personAssetCount, err := immichImage.PersonImageCount(person, requestID)
		if err != nil {
			return nil, fmt.Errorf("getting person image count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: "PERSON", ID: person},
			Weight: personAssetCount,
		})
	}

	for _, album := range requestConfig.Album {
		albumAssetCount, err := immichImage.AlbumImageCount(album, requestID)
		if err != nil {
			return nil, fmt.Errorf("getting album asset count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: "ALBUM", ID: album},
			Weight: albumAssetCount,
		})
	}

	return peopleAndAlbums, nil
}

func isSleepMode(requestConfig config.Config) bool {
	if requestConfig.SleepStart == "" || requestConfig.SleepEnd == "" {
		return false
	}

	if isSleepTime, _ := utils.IsSleepTime(requestConfig.SleepStart, requestConfig.SleepEnd, time.Now()); isSleepTime {
		return isSleepTime
	}

	return false
}

// retrieveImage fetches a random image based on the picked image type.
// It returns an error if the image retrieval fails.
func retrieveImage(immichImage *immich.ImmichAsset, pickedAsset utils.WeightedAsset, requestID, kioskDeviceID string, isPrefetch bool) error {
	pageDataCacheMutex.Lock()
	defer pageDataCacheMutex.Unlock()

	switch pickedAsset.Type {
	case "ALBUM":
		switch pickedAsset.ID {
		case "all":
			pickedAlbumID, err := immichImage.RandomAlbumFromAllAlbums(requestID)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case "shared":
			pickedAlbumID, err := immichImage.RandomAlbumFromSharedAlbums(requestID)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		}
		return immichImage.RandomImageFromAlbum(pickedAsset.ID, requestID, kioskDeviceID, isPrefetch)
	case "PERSON":
		return immichImage.RandomImageOfPerson(pickedAsset.ID, requestID, kioskDeviceID, isPrefetch)
	default:
		return immichImage.RandomImage(requestID, kioskDeviceID, isPrefetch)
	}
}

// fetchImagePreview retrieves the preview of an image and logs the time taken.
// It returns the image bytes and an error if any occurs.
func fetchImagePreview(immichImage *immich.ImmichAsset, requestID, kioskDeviceID string, isPrefetch bool) ([]byte, error) {
	imageGet := time.Now()
	imgBytes, err := immichImage.ImagePreview()
	if err != nil {
		return nil, fmt.Errorf("getting image preview: %w", err)
	}

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestID, "Got image in", time.Since(imageGet).Seconds())
	}

	return imgBytes, nil
}

// processImage handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processImage(immichImage *immich.ImmichAsset, requestConfig config.Config, requestID string, kioskDeviceID string, isPrefetch bool) ([]byte, error) {

	peopleAndAlbums, err := gatherPeopleAndAlbums(immichImage, requestConfig, requestID)
	if err != nil {
		return nil, err
	}

	pickedImage := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, peopleAndAlbums)

	if err := retrieveImage(immichImage, pickedImage, requestID, kioskDeviceID, isPrefetch); err != nil {
		return nil, err
	}

	return fetchImagePreview(immichImage, requestID, kioskDeviceID, isPrefetch)
}

// imageToBase64 converts image bytes to a base64 string and logs the processing time.
// It returns the base64 string and an error if conversion fails.
func imageToBase64(imgBytes []byte, config config.Config, requestID, kioskDeviceID string, action string, isPrefetch bool) (string, error) {
	startTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return "", fmt.Errorf("converting image to base64: %w", err)
	}

	logImageProcessing(config, requestID, kioskDeviceID, isPrefetch, action, startTime)
	return img, nil
}

// processBlurredImage applies a blur effect to the image if required by the configuration.
// It returns the blurred image as a base64 string and an error if any occurs.
func processBlurredImage(imgBytes []byte, config config.Config, requestID, kioskDeviceID string, isPrefetch bool) (string, error) {
	if !config.BackgroundBlur || strings.EqualFold(config.ImageFit, "cover") {
		return "", nil
	}

	startTime := time.Now()
	imgBlurBytes, err := utils.BlurImage(imgBytes)
	if err != nil {
		return "", fmt.Errorf("blurring image: %w", err)
	}

	logImageProcessing(config, requestID, kioskDeviceID, isPrefetch, "Blurred", startTime)

	return imageToBase64(imgBlurBytes, config, requestID, kioskDeviceID, "Coverted blurred", isPrefetch)
}

// logImageProcessing logs the time taken for image processing if debug verbose is enabled.
func logImageProcessing(config config.Config, requestID, kioskDeviceID string, isPrefetch bool, action string, startTime time.Time) {
	if !config.Kiosk.DebugVerbose {
		return
	}

	duration := time.Since(startTime).Seconds()
	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, action+" image in", duration)
	} else {
		log.Debug(requestID, action+" image in", duration)
	}
}

// trimHistory ensures that the history slice doesn't exceed the specified maximum length.
func trimHistory(history *[]string, maxLength int) {
	if len(*history) > maxLength {
		*history = (*history)[len(*history)-maxLength:]
	}
}

// processPageData handles the entire process of preparing page data including image processing.
// It returns the PageData and an error if any step fails.
func processPageData(requestConfig config.Config, c echo.Context, isPrefetch bool) (views.PageData, error) {
	requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceID := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)

	imgBytes, err := processImage(&immichImage, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.PageData{}, fmt.Errorf("selecting image: %w", err)
	}

	img, err := imageToBase64(imgBytes, requestConfig, requestID, kioskDeviceID, "Converted", isPrefetch)
	if err != nil {
		return views.PageData{}, err
	}

	imgBlur, err := processBlurredImage(imgBytes, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.PageData{}, err
	}

	trimHistory(&requestConfig.History, 10)

	return views.PageData{
		ImmichImage:   immichImage,
		ImageData:     img,
		ImageBlurData: imgBlur,
		Config:        requestConfig,
	}, nil
}

// imagePreFetch pre-fetches a specified number of images and caches them.
func imagePreFetch(numberOfImages int, requestConfig config.Config, c echo.Context, kioskDeviceID string) {

	var wg sync.WaitGroup

	wg.Add(numberOfImages)

	cacheKey := c.Request().URL.String() + kioskDeviceID

	for i := 0; i < numberOfImages; i++ {

		go func() {

			defer wg.Done()

			pageData, err := processPageData(requestConfig, c, true)
			if err != nil {
				log.Error("prefetch", "err", err)
				return
			}

			pageDataCacheMutex.Lock()
			defer pageDataCacheMutex.Unlock()

			cachedPageData := []views.PageData{}

			if data, found := pageDataCache.Get(cacheKey); found {
				cachedPageData = data.([]views.PageData)
			}

			cachedPageData = append(cachedPageData, pageData)

			pageDataCache.Set(cacheKey, cachedPageData, cache.DefaultExpiration)
		}()

	}

	wg.Wait()
}

// fromCache retrieves cached page data for a given request and device ID.
func fromCache(c echo.Context, kioskDeviceID string) []views.PageData {
	pageDataCacheMutex.Lock()
	defer pageDataCacheMutex.Unlock()

	cacheKey := c.Request().URL.String() + kioskDeviceID
	if data, found := pageDataCache.Get(cacheKey); found {
		cachedPageData := data.([]views.PageData)
		if len(cachedPageData) >= 2 {
			return cachedPageData
		}
		pageDataCache.Delete(cacheKey)
	}
	return nil
}

// renderCachedPageData renders cached page data and updates the cache.
func renderCachedPageData(c echo.Context, cachedPageData []views.PageData, requestConfig *config.Config, requestID string, kioskDeviceID string) error {
	pageDataCacheMutex.Lock()
	defer pageDataCacheMutex.Unlock()

	log.Debug(requestID, "deviceID", kioskDeviceID, "cache hit for new image", true)

	cacheKey := c.Request().URL.String() + kioskDeviceID

	var pageDataToRender []views.PageData
	if strings.EqualFold(requestConfig.Layout, "splitview") {
		pageDataToRender = cachedPageData[:2]
		pageDataCache.Set(cacheKey, cachedPageData[2:], cache.DefaultExpiration)
		go imagePreFetch(2, *requestConfig, c, kioskDeviceID)
	} else {
		pageDataToRender = []views.PageData{cachedPageData[0]}
		pageDataCache.Set(cacheKey, cachedPageData[1:], cache.DefaultExpiration)
		go imagePreFetch(1, *requestConfig, c, kioskDeviceID)
	}

	// Update history which will be outdated in cache
	trimHistory(&requestConfig.History, 10)
	pageDataToRender[0].History = requestConfig.History

	return Render(c, http.StatusOK, views.Image(pageDataToRender...))
}

// generatePageData generates page data for the current request.
func generatePageData(requestConfig config.Config, c echo.Context) ([]views.PageData, error) {

	imagesNeeded := 1
	if strings.EqualFold(requestConfig.Layout, "splitview") {
		imagesNeeded = 2
	}

	pageData := make([]views.PageData, imagesNeeded)
	for i := range pageData {
		pageDataSingle, err := processPageData(requestConfig, c, false)
		if err != nil {
			return nil, err
		}
		pageData[i] = pageDataSingle
	}
	return pageData, nil
}
