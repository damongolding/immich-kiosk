package routes

import (
	"fmt"
	"image"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/fogleman/gg"
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

		if personAssetCount == 0 {
			log.Error("No assets found for", "person", person)
			continue
		}

		peopleAndAlbums = append(peopleAndAlbums, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourcePerson, ID: person},
			Weight: personAssetCount,
		})
	}

	for _, album := range requestConfig.Album {

		albumAssetCount, err := immichImage.AlbumImageCount(album, requestID)
		if err != nil {
			return nil, fmt.Errorf("getting album asset count: %w", err)
		}

		if albumAssetCount == 0 {
			log.Error("No assets found for", "album", album)
			continue
		}

		peopleAndAlbums = append(peopleAndAlbums, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceAlbums, ID: album},
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
func retrieveImage(immichImage *immich.ImmichAsset, pickedAsset utils.WeightedAsset, excludedAlbums []string, requestID, kioskDeviceID string, isPrefetch bool) error {

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	switch pickedAsset.Type {
	case kiosk.SourceAlbums:
		switch pickedAsset.ID {
		case kiosk.AlbumKeywordAll:
			pickedAlbumID, err := immichImage.RandomAlbumFromAllAlbums(requestID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordShared:
			pickedAlbumID, err := immichImage.RandomAlbumFromSharedAlbums(requestID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordFavourites, kiosk.AlbumKeywordFavorites:
			return immichImage.RandomImageFromFavourites(requestID, kioskDeviceID, isPrefetch)
		}
		return immichImage.RandomImageFromAlbum(pickedAsset.ID, requestID, kioskDeviceID, isPrefetch)
	case kiosk.SourcePerson:
		return immichImage.RandomImageOfPerson(pickedAsset.ID, requestID, kioskDeviceID, isPrefetch)
	default:
		return immichImage.RandomImage(requestID, kioskDeviceID, isPrefetch)
	}
}

// fetchImagePreview retrieves the preview of an image and logs the time taken.
// It returns the image bytes and an error if any occurs.
func fetchImagePreview(immichImage *immich.ImmichAsset, requestID, kioskDeviceID string, isPrefetch bool) (image.Image, error) {
	imageGet := time.Now()

	imgBytes, err := immichImage.ImagePreview()
	if err != nil {
		return nil, fmt.Errorf("getting image preview: %w", err)
	}

	img, err := utils.BytesToImage(imgBytes)
	if err != nil {
		return nil, err
	}

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestID, "Got image in", time.Since(imageGet).Seconds())
	}

	img = utils.ApplyExifOrientation(img, immichImage.IsLandscape, immichImage.ExifInfo.Orientation)

	return img, nil
}

// processImage handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processImage(immichImage *immich.ImmichAsset, requestConfig config.Config, requestID string, kioskDeviceID string, isPrefetch bool) (image.Image, error) {

	peopleAndAlbums, err := gatherPeopleAndAlbums(immichImage, requestConfig, requestID)
	if err != nil {
		return nil, err
	}

	pickedImage := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, peopleAndAlbums)

	if err := retrieveImage(immichImage, pickedImage, requestConfig.ExcludedAlbums, requestID, kioskDeviceID, isPrefetch); err != nil {
		return nil, err
	}

	immichImage.KioskSource = pickedImage.Type

	return fetchImagePreview(immichImage, requestID, kioskDeviceID, isPrefetch)
}

// imageToBase64 converts image bytes to a base64 string and logs the processing time.
// It returns the base64 string and an error if conversion fails.
func imageToBase64(img image.Image, config config.Config, requestID, kioskDeviceID string, action string, isPrefetch bool) (string, error) {
	startTime := time.Now()

	imgBytes, err := utils.ImageToBase64(img)
	if err != nil {
		return "", fmt.Errorf("converting image to base64: %w", err)
	}

	logImageProcessing(config, requestID, kioskDeviceID, isPrefetch, action, startTime)
	return imgBytes, nil
}

// processBlurredImage applies a blur effect to the image if required by the configuration.
//   - An alternative image effect is specified
func processBlurredImage(img image.Image, config config.Config, requestID, kioskDeviceID string, isPrefetch bool) (string, error) {
	if !config.BackgroundBlur || strings.EqualFold(config.ImageFit, "cover") || (config.ImageEffect != "" && config.ImageEffect != "none") {
		return "", nil
	}

	startTime := time.Now()
	imgBlur, err := utils.BlurImage(img, config.OptimizeImages, config.ClientData.Width, config.ClientData.Height)
	if err != nil {
		return "", fmt.Errorf("blurring image: %w", err)
	}

	logImageProcessing(config, requestID, kioskDeviceID, isPrefetch, "Blurred", startTime)

	return imageToBase64(imgBlur, config, requestID, kioskDeviceID, "Coverted blurred", isPrefetch)
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

func DrawFaceOnImage(img image.Image, i *immich.ImmichAsset) image.Image {

	if len(i.People) == 0 && len(i.UnassignedFaces) == 0 {
		log.Debug("no people found")
		return img
	}

	dc := gg.NewContext(img.Bounds().Dx(), img.Bounds().Dy())

	dc.DrawImage(img, 0, 0)

	for _, person := range i.People {
		for _, face := range person.Faces {
			width := face.BoundingBoxX2 - face.BoundingBoxX1
			height := face.BoundingBoxY2 - face.BoundingBoxY1

			dc.DrawRectangle(float64(face.BoundingBoxX1), float64(face.BoundingBoxY1), float64(width), float64(height))
			dc.SetHexColor("#990000")
			dc.Fill()
		}
	}

	for _, face := range i.UnassignedFaces {
		width := face.BoundingBoxX2 - face.BoundingBoxX1
		height := face.BoundingBoxY2 - face.BoundingBoxY1

		dc.DrawRectangle(float64(face.BoundingBoxX1), float64(face.BoundingBoxY1), float64(width), float64(height))
		dc.SetHexColor("#000099")
		dc.Fill()
	}

	facesBoundX, facesBoundY := i.FacesCenterPointPX()
	dc.DrawRectangle(facesBoundX-10, facesBoundY-10, 20, 20)
	dc.SetHexColor("#889900")
	dc.Fill()

	return dc.Image()

}

// processViewImageData handles the entire process of preparing page data including image processing.
// It returns the ImageData and an error if any step fails.
func processViewImageData(imageOrientation immich.ImageOrientation, requestConfig config.Config, c echo.Context, isPrefetch bool) (common.ViewImageData, error) {
	requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceID := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)

	switch imageOrientation {
	case immich.PortraitOrientation:
		immichImage.RatioWanted = imageOrientation
	case immich.LandscapeOrientation:
		immichImage.RatioWanted = imageOrientation
	}

	img, err := processImage(&immichImage, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return common.ViewImageData{}, fmt.Errorf("selecting image: %w", err)
	}

	if strings.EqualFold(requestConfig.ImageEffect, "smart-zoom") && len(immichImage.People)+len(immichImage.UnassignedFaces) == 0 {
		immichImage.CheckForFaces(requestID)
	}

	if ShouldDrawFacesOnImages() {
		log.Debug("Drawing faces")
		img = DrawFaceOnImage(img, &immichImage)
	}

	if requestConfig.OptimizeImages {
		img, err = utils.OptimizeImage(img, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
		if err != nil {
			return common.ViewImageData{}, err
		}
	}

	imgString, err := imageToBase64(img, requestConfig, requestID, kioskDeviceID, "Converted", isPrefetch)
	if err != nil {
		return common.ViewImageData{}, err
	}

	imgBlurString, err := processBlurredImage(img, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return common.ViewImageData{}, err
	}

	return common.ViewImageData{
		ImmichImage:   immichImage,
		ImageData:     imgString,
		ImageBlurData: imgBlurString,
	}, nil
}

func ProcessViewImageData(requestConfig config.Config, c echo.Context, isPrefetch bool) (common.ViewImageData, error) {
	return processViewImageData("", requestConfig, c, isPrefetch)
}

func ProcessViewImageDataWithRatio(imageOrientation immich.ImageOrientation, requestConfig config.Config, c echo.Context, isPrefetch bool) (common.ViewImageData, error) {
	return processViewImageData(imageOrientation, requestConfig, c, isPrefetch)
}

func imagePreFetch(requestData *common.RouteRequestData, c echo.Context) {

	requestConfig := requestData.RequestConfig
	requestID := requestData.RequestID
	deviceID := requestData.DeviceID

	viewDataToAdd, err := generateViewData(requestConfig, c, requestID, true)
	if err != nil {
		log.Error("generateViewData", "prefetch", true, "err", err)
		return
	}

	trimHistory(&requestConfig.History, 10)

	cachedViewData := []common.ViewData{}

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	cacheKey := c.Request().URL.String() + deviceID

	if data, found := ViewDataCache.Get(cacheKey); found {
		cachedViewData = data.([]common.ViewData)
	}

	cachedViewData = append(cachedViewData, viewDataToAdd)

	ViewDataCache.Set(cacheKey, cachedViewData, cache.DefaultExpiration)

	go webhooks.Trigger(requestData, KioskVersion, webhooks.PrefetchAsset, viewDataToAdd)

}

// fromCache retrieves cached page data for a given request and device ID.
func fromCache(urlString string, kioskDeviceID string) []common.ViewData {

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	cacheKey := urlString + kioskDeviceID
	if data, found := ViewDataCache.Get(cacheKey); found {
		cachedPageData := data.([]common.ViewData)
		if len(cachedPageData) > 0 {
			return cachedPageData
		}
		ViewDataCache.Delete(cacheKey)
	}
	return nil
}

// renderCachedViewData renders cached page data and updates the cache.
func renderCachedViewData(c echo.Context, cachedViewData []common.ViewData, requestConfig *config.Config, requestID string, kioskDeviceID string) error {
	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	log.Debug(requestID, "deviceID", kioskDeviceID, "cache hit for new image", true)

	cacheKey := c.Request().URL.String() + kioskDeviceID

	viewDataToRender := cachedViewData[0]
	ViewDataCache.Set(cacheKey, cachedViewData[1:], cache.DefaultExpiration)

	// Update history which will be outdated in cache
	trimHistory(&requestConfig.History, 10)
	viewDataToRender.History = requestConfig.History

	return Render(c, http.StatusOK, imageComponent.Image(viewDataToRender))
}

// generateViewData generates page data for the current request.
func generateViewData(requestConfig config.Config, c echo.Context, kioskDeviceID string, isPrefetch bool) (common.ViewData, error) {

	const maxImageRetrievalAttepmts = 3

	viewData := common.ViewData{
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

		if viewDataSplitView.ImmichImage.IsLandscape {
			return viewData, nil
		}

		// Second image
		for i := 0; i < maxImageRetrievalAttepmts; i++ {
			viewDataSplitViewSecond, err := ProcessViewImageDataWithRatio(immich.PortraitOrientation, requestConfig, c, isPrefetch)
			if err != nil {
				return viewData, err
			}

			if viewDataSplitView.ImmichImage.ID != viewDataSplitViewSecond.ImmichImage.ID {
				viewData.Images = append(viewData.Images, viewDataSplitViewSecond)
				break
			}
		}

	case "splitview-landscape":
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSplitView)

		if viewDataSplitView.ImmichImage.IsPortrait {
			return viewData, nil
		}

		// Second image
		for i := 0; i < maxImageRetrievalAttepmts; i++ {
			viewDataSplitViewSecond, err := ProcessViewImageDataWithRatio(immich.LandscapeOrientation, requestConfig, c, isPrefetch)
			if err != nil {
				return viewData, err
			}

			if viewDataSplitView.ImmichImage.ID != viewDataSplitViewSecond.ImmichImage.ID {
				viewData.Images = append(viewData.Images, viewDataSplitViewSecond)
				break
			}
		}

	default:
		viewDataSingle, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSingle)
	}

	return viewData, nil
}
