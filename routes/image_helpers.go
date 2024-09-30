package routes

import (
	"fmt"
	"net/http"
	"strings"
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
func gatherPeopleAndAlbums(immichImage *immich.ImmichAsset, requestConfig config.Config, requestId string) ([]immich.AssetWithWeighting, error) {

	peopleAndAlbums := []immich.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		personAssetCount, err := immichImage.PersonImageCount(person, requestId)
		if err != nil {
			return nil, fmt.Errorf("getting person image count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, immich.AssetWithWeighting{
			Asset:  immich.WeightedAsset{Type: "PERSON", ID: person},
			Weight: personAssetCount,
		})
	}

	for _, album := range requestConfig.Album {
		albumAssetCount, err := immichImage.AlbumImageCount(album, requestId)
		if err != nil {
			return nil, fmt.Errorf("getting album asset count: %w", err)
		}
		peopleAndAlbums = append(peopleAndAlbums, immich.AssetWithWeighting{
			Asset:  immich.WeightedAsset{Type: "ALBUM", ID: album},
			Weight: albumAssetCount,
		})
	}

	return peopleAndAlbums, nil
}

// pickRandomImageType selects a random image type based on the given configuration and weightings.
// It returns a WeightedAsset representing the picked image type.
func pickRandomImageType(requestConfig config.Config, peopleAndAlbums []immich.AssetWithWeighting) immich.WeightedAsset {

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

	return pickedImage
}

// retrieveImage fetches a random image based on the picked image type.
// It returns an error if the image retrieval fails.
func retrieveImage(immichImage *immich.ImmichAsset, pickedImage immich.WeightedAsset, requestID, kioskDeviceID string, isPrefetch bool) error {

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	switch pickedImage.Type {
	case "ALBUM":
		return immichImage.RandomImageFromAlbum(pickedImage.ID, requestID, kioskDeviceID, isPrefetch)
	case "PERSON":
		return immichImage.RandomImageOfPerson(pickedImage.ID, requestID, kioskDeviceID, isPrefetch)
	default:
		return immichImage.RandomImage(requestID, kioskDeviceID, isPrefetch)
	}
}

// fetchImagePreview retrieves the preview of an image and logs the time taken.
// It returns the image bytes and an error if any occurs.
func fetchImagePreview(immichImage *immich.ImmichAsset, requestId, kioskDeviceID string, isPrefetch bool) ([]byte, error) {

	imageGet := time.Now()
	imgBytes, err := immichImage.ImagePreview()
	if err != nil {
		return nil, fmt.Errorf("getting image preview: %w", err)
	}

	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceID, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestId, "Got image in", time.Since(imageGet).Seconds())
	}

	return imgBytes, nil
}

// processImage handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processImage(immichImage *immich.ImmichAsset, requestConfig config.Config, requestId string, kioskDeviceID string, isPrefetch bool) ([]byte, error) {

	peopleAndAlbums, err := gatherPeopleAndAlbums(immichImage, requestConfig, requestId)
	if err != nil {
		return nil, err
	}

	pickedImage := pickRandomImageType(requestConfig, peopleAndAlbums)

	if err := retrieveImage(immichImage, pickedImage, requestId, kioskDeviceID, isPrefetch); err != nil {
		return nil, err
	}

	return fetchImagePreview(immichImage, requestId, kioskDeviceID, isPrefetch)
}

// imageToBase64 converts image bytes to a base64 string and logs the processing time.
// It returns the base64 string and an error if conversion fails.
func imageToBase64(imgBytes []byte, config config.Config, requestId, kioskDeviceID string, isPrefetch bool) (string, error) {
	startTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return "", fmt.Errorf("converting image to base64: %w", err)
	}

	logImageProcessing(config, requestId, kioskDeviceID, isPrefetch, "Converted", startTime)
	return img, nil
}

// processBlurredImage applies a blur effect to the image if required by the configuration.
// It returns the blurred image as a base64 string and an error if any occurs.
func processBlurredImage(imgBytes []byte, config config.Config, requestId, kioskDeviceID string, isPrefetch bool) (string, error) {
	if !config.BackgroundBlur || strings.EqualFold(config.ImageFit, "cover") {
		return "", nil
	}

	startTime := time.Now()
	imgBlurBytes, err := utils.BlurImage(imgBytes)
	if err != nil {
		return "", fmt.Errorf("blurring image: %w", err)
	}

	logImageProcessing(config, requestId, kioskDeviceID, isPrefetch, "Blurred", startTime)

	return imageToBase64(imgBlurBytes, config, requestId, kioskDeviceID, isPrefetch)
}

// logImageProcessing logs the time taken for image processing if debug verbose is enabled.
func logImageProcessing(config config.Config, requestId, kioskDeviceID string, isPrefetch bool, action string, startTime time.Time) {
	if !config.Kiosk.DebugVerbose {
		return
	}

	duration := time.Since(startTime).Seconds()
	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceID, action, "image in", duration)
	} else {
		log.Debug(requestId, action, "image in", duration)
	}
}

// trimHistory ensures that the history slice doesn't exceed the specified maximum length.
func trimHistory(history *[]string, maxLength int) {
	if len(*history) > maxLength {
		*history = (*history)[len(*history)-maxLength:]
	}
}

// processViewImageData handles the entire process of preparing page data including image processing.
// It returns the ImageData and an error if any step fails.
func processViewImageData(ratio string, requestConfig config.Config, c echo.Context, isPrefetch bool) (views.ImageData, error) {
	requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceID := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)
	immichImage.RatioWanted = ratio

	imgBytes, err := processImage(&immichImage, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.ImageData{}, fmt.Errorf("selecting image: %w", err)
	}

	img, err := imageToBase64(imgBytes, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.ImageData{}, err
	}

	imgBlur, err := processBlurredImage(imgBytes, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.ImageData{}, err
	}

	return views.ImageData{
		ImmichImage:   immichImage,
		ImageData:     img,
		ImageBlurData: imgBlur,
	}, nil
}

func ProcessViewImageData(requestConfig config.Config, c echo.Context, isPrefetch bool) (views.ImageData, error) {
	return processViewImageData("", requestConfig, c, isPrefetch)
}

func ProcessViewImageDataWithRatio(ratio string, requestConfig config.Config, c echo.Context, isPrefetch bool) (views.ImageData, error) {
	return processViewImageData(ratio, requestConfig, c, isPrefetch)
}

func imagePreFetch(numberOfImages int, requestConfig config.Config, c echo.Context, kioskDeviceID string) {

	viewDataToAdd, err := generateViewData(numberOfImages, requestConfig, c, kioskDeviceID, true)
	if err != nil {
		log.Error("prefetch", "err", err)
		return
	}

	trimHistory(&requestConfig.History, 10)

	cachedViewData := []views.ViewData{}

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	cacheKey := c.Request().URL.String() + kioskDeviceID

	if data, found := viewDataCache.Get(cacheKey); found {
		cachedViewData = data.([]views.ViewData)
	}

	cachedViewData = append(cachedViewData, viewDataToAdd)

	viewDataCache.Set(cacheKey, cachedViewData, cache.DefaultExpiration)

}

// imagePreFetch pre-fetches a specified number of images and caches them.
// func imagePreFetchOld(numberOfImages int, requestConfig config.Config, c echo.Context, kioskDeviceID string) {

// 	var wg sync.WaitGroup

// 	wg.Add(numberOfImages)

// 	cacheKey := c.Request().URL.String() + kioskDeviceID

// 	for i := 0; i < numberOfImages; i++ {

// 		go func() {

// 			defer wg.Done()

// 			viewImageData, err := processViewImageData(requestConfig, c, true)
// 			if err != nil {
// 				log.Error("prefetch", "err", err)
// 				return
// 			}

// 			viewDataCacheMutex.Lock()
// 			defer viewDataCacheMutex.Unlock()

// 			cachedViewData := []views.ViewData{}

// 			if data, found := viewDataCache.Get(cacheKey); found {
// 				cachedViewData = data.([]views.ViewData)
// 			}

// 			cachedViewData = append(cachedViewData, viewImageData)

// 			viewDataCache.Set(cacheKey, cachedViewData, cache.DefaultExpiration)
// 		}()

// 	}

// 	wg.Wait()
// }

// fromCache retrieves cached page data for a given request and device ID.
func fromCache(c echo.Context, kioskDeviceID string) []views.ViewData {

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	cacheKey := c.Request().URL.String() + kioskDeviceID
	if data, found := viewDataCache.Get(cacheKey); found {
		cachedPageData := data.([]views.ViewData)
		if len(cachedPageData) > 0 {
			return cachedPageData
		}
		viewDataCache.Delete(cacheKey)
	}
	return nil
}

// renderCachedViewData renders cached page data and updates the cache.
func renderCachedViewData(c echo.Context, cachedViewData []views.ViewData, requestConfig *config.Config, requestID string, kioskDeviceID string) error {
	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	log.Debug(requestID, "deviceID", kioskDeviceID, "cache hit for new image", true)

	cacheKey := c.Request().URL.String() + kioskDeviceID

	viewDataToRender := cachedViewData[0]
	viewDataCache.Set(cacheKey, cachedViewData[1:], cache.DefaultExpiration)

	// Update history which will be outdated in cache
	trimHistory(&requestConfig.History, 10)
	viewDataToRender.History = requestConfig.History

	return Render(c, http.StatusOK, views.Image(viewDataToRender))
}

// generateViewData generates page data for the current request.
func generateViewData(numberOfImages int, requestConfig config.Config, c echo.Context, kioskDeviceID string, isPrefetch bool) (views.ViewData, error) {

	viewData := views.ViewData{
		DeviceID: kioskDeviceID,
		Config:   requestConfig,
	}

	switch requestConfig.Layout {
	case "splitview":
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSplitView)

		if viewDataSplitView.ImmichImage.ExifInfo.Ratio == immich.Landscape {
			log.Info("Got a landscape image")
			return viewData, nil
		}

		log.Info("want a portrait image")
		viewDataSplitView, err = ProcessViewImageDataWithRatio(immich.Portrait, requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSplitView)

	default:
		viewDataSingle, err := processViewImageData("", requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSingle)
	}

	return viewData, nil
}
