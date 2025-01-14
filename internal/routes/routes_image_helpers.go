package routes

import (
	"fmt"
	"image"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/fogleman/gg"
	"github.com/labstack/echo/v4"
)

// gatherAssetBuckets collects asset weightings for people, albums and date ranges.
// For each person, it gets the count of images containing that person.
// For each album, it gets the total count of images in the album.
// For date ranges, it currently assigns a fixed weighting of 1000.
// These weightings are used to determine the probability of selecting images from each source.
//
// Parameters:
//   - immichImage: The Immich asset used to query image counts
//   - requestConfig: Configuration containing people, albums and dates to gather assets for
//   - requestID: Identifier for the current request for logging
//
// Returns:
//   - A slice of AssetWithWeighting containing the weightings for each asset source
//   - An error if any database queries fail
func gatherAssetBuckets(immichImage *immich.ImmichAsset, requestConfig config.Config, requestID, deviceID string) ([]utils.AssetWithWeighting, error) {

	assets := []utils.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		personAssetCount, err := immichImage.PersonImageCount(person, requestID, deviceID)
		if err != nil {
			return nil, fmt.Errorf("getting person image count: %w", err)
		}

		if personAssetCount == 0 {
			log.Error("No assets found for", "person", person)
			continue
		}

		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourcePerson, ID: person},
			Weight: personAssetCount,
		})
	}

	for _, album := range requestConfig.Album {

		albumAssetCount, err := immichImage.AlbumImageCount(album, requestID, deviceID)
		if err != nil {
			return nil, fmt.Errorf("getting album asset count: %w", err)
		}

		if albumAssetCount == 0 {
			log.Error("No assets found for", "album", album)
			continue
		}

		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceAlbums, ID: album},
			Weight: albumAssetCount,
		})
	}

	for _, date := range requestConfig.Date {

		// use FetchedAssetsSize as a weighting for date ranges
		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceDateRangeAlbum, ID: date},
			Weight: requestConfig.Kiosk.FetchedAssetsSize,
		})
	}

	if requestConfig.Memories {
		memories := immichImage.MemoryLaneAssetsCount(requestID, deviceID)
		if memories == 0 {
			log.Error("No assets found for memories")
		} else {
			assets = append(assets, utils.AssetWithWeighting{
				Asset:  utils.WeightedAsset{Type: kiosk.SourceMemories, ID: "memories"},
				Weight: memories,
			})
		}
	}

	return assets, nil
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
func retrieveImage(immichImage *immich.ImmichAsset, pickedAsset utils.WeightedAsset, albumOrder string, excludedAlbums []string, requestID, deviceID string, isPrefetch bool) error {

	switch pickedAsset.Type {
	case kiosk.SourceAlbums:
		switch pickedAsset.ID {
		case kiosk.AlbumKeywordAll:
			pickedAlbumID, err := immichImage.RandomAlbumFromAllAlbums(requestID, deviceID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordShared:
			pickedAlbumID, err := immichImage.RandomAlbumFromSharedAlbums(requestID, deviceID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordFavourites, kiosk.AlbumKeywordFavorites:
			return immichImage.RandomImageFromFavourites(requestID, deviceID, isPrefetch)
		}

		switch strings.ToLower(albumOrder) {
		case config.AlbumOrderDescending, config.AlbumOrderDesc, config.AlbumOrderNewest:
			return immichImage.ImageFromAlbum(pickedAsset.ID, immich.Desc, requestID, deviceID, isPrefetch)
		case config.AlbumOrderAscending, config.AlbumOrderAsc, config.AlbumOrderOldest:
			return immichImage.ImageFromAlbum(pickedAsset.ID, immich.Asc, requestID, deviceID, isPrefetch)
		default:
			return immichImage.ImageFromAlbum(pickedAsset.ID, immich.Rand, requestID, deviceID, isPrefetch)
		}

	case kiosk.SourceDateRangeAlbum:
		return immichImage.RandomImageInDateRange(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourcePerson:
		return immichImage.RandomImageOfPerson(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourceMemories:
		return immichImage.RandomMemoryLaneImage(requestID, deviceID, isPrefetch)

	default:
		return immichImage.RandomImage(requestID, deviceID, isPrefetch)
	}
}

// fetchImagePreview retrieves the preview of an image and logs the time taken.
// It returns the image bytes and an error if any occurs.
func fetchImagePreview(immichImage *immich.ImmichAsset, requestID, deviceID string, isPrefetch bool) (image.Image, error) {
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
		log.Debug(requestID, "PREFETCH", deviceID, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestID, "Got image in", time.Since(imageGet).Seconds())
	}

	img = utils.ApplyExifOrientation(img, immichImage.IsLandscape, immichImage.ExifInfo.Orientation)

	return img, nil
}

// processImage handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processImage(immichImage *immich.ImmichAsset, requestConfig config.Config, requestID string, deviceID string, isPrefetch bool) (image.Image, error) {

	assets, err := gatherAssetBuckets(immichImage, requestConfig, requestID, deviceID)
	if err != nil {
		return nil, err
	}

	pickedAsset := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, assets)

	if err := retrieveImage(immichImage, pickedAsset, requestConfig.AlbumOrder, requestConfig.ExcludedAlbums, requestID, deviceID, isPrefetch); err != nil {
		return nil, err
	}

	immichImage.KioskSource = pickedAsset.Type

	return fetchImagePreview(immichImage, requestID, deviceID, isPrefetch)
}

// imageToBase64 converts image bytes to a base64 string and logs the processing time.
// It returns the base64 string and an error if conversion fails.
func imageToBase64(img image.Image, config config.Config, requestID, deviceID string, action string, isPrefetch bool) (string, error) {
	startTime := time.Now()

	imgBytes, err := utils.ImageToBase64(img)
	if err != nil {
		return "", fmt.Errorf("converting image to base64: %w", err)
	}

	logImageProcessing(config, requestID, deviceID, isPrefetch, action, startTime)
	return imgBytes, nil
}

// processBlurredImage applies a blur effect to the image if required by the configuration.
// It returns the blurred image as a base64 string and an error if any occurs.
func processBlurredImage(img image.Image, config config.Config, requestID, deviceID string, isPrefetch bool) (string, error) {
	if !config.BackgroundBlur || strings.EqualFold(config.ImageFit, "cover") || (config.ImageEffect != "" && config.ImageEffect != "none") {
		return "", nil
	}

	startTime := time.Now()
	imgBlur, err := utils.BlurImage(img, config.OptimizeImages, config.ClientData.Width, config.ClientData.Height)
	if err != nil {
		return "", fmt.Errorf("blurring image: %w", err)
	}

	logImageProcessing(config, requestID, deviceID, isPrefetch, "Blurred", startTime)

	return imageToBase64(imgBlur, config, requestID, deviceID, "Coverted blurred", isPrefetch)
}

// logImageProcessing logs the time taken for image processing if debug verbose is enabled.
func logImageProcessing(config config.Config, requestID, deviceID string, isPrefetch bool, action string, startTime time.Time) {
	if !config.Kiosk.DebugVerbose {
		return
	}

	duration := time.Since(startTime).Seconds()
	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, action+" image in", duration)
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
	deviceID := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)

	switch imageOrientation {
	case immich.PortraitOrientation:
		immichImage.RatioWanted = imageOrientation
	case immich.LandscapeOrientation:
		immichImage.RatioWanted = imageOrientation
	}

	img, err := processImage(&immichImage, requestConfig, requestID, deviceID, isPrefetch)
	if err != nil {
		return common.ViewImageData{}, fmt.Errorf("selecting image: %w", err)
	}

	if strings.EqualFold(requestConfig.ImageEffect, "smart-zoom") && len(immichImage.People)+len(immichImage.UnassignedFaces) == 0 {
		immichImage.CheckForFaces(requestID, deviceID)
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

	imgString, err := imageToBase64(img, requestConfig, requestID, deviceID, "Converted", isPrefetch)
	if err != nil {
		return common.ViewImageData{}, err
	}

	imgBlurString, err := processBlurredImage(img, requestConfig, requestID, deviceID, isPrefetch)
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

	viewCacheKey := cache.ViewCacheKey(c.Request().URL.String(), deviceID)

	if data, found := cache.Get(viewCacheKey); found {
		cachedViewData = data.([]common.ViewData)
	}

	cachedViewData = append(cachedViewData, viewDataToAdd)

	cache.Set(viewCacheKey, cachedViewData)

	go webhooks.Trigger(requestData, KioskVersion, webhooks.PrefetchAsset, viewDataToAdd)

}

// fromCache retrieves cached page data for a given request and device ID.
func fromCache(urlString string, deviceID string) []common.ViewData {
	cacheKey := cache.ViewCacheKey(urlString, deviceID)
	if data, found := cache.Get(cacheKey); found {
		cachedPageData, ok := data.([]common.ViewData)
		if !ok {
			log.Error("cache: invalid data type", "key", cacheKey)
			cache.Delete(cacheKey)
			return nil
		}
		if len(cachedPageData) > 0 {
			return cachedPageData
		}
		cache.Delete(cacheKey)
	}
	return nil
}

// renderCachedViewData renders cached page data and updates the cache.
func renderCachedViewData(c echo.Context, cachedViewData []common.ViewData, requestConfig *config.Config, requestID string, deviceID string) error {

	log.Debug(requestID, "deviceID", deviceID, "cache hit for new image", true)

	cacheKey := cache.ViewCacheKey(c.Request().URL.String(), deviceID)

	viewDataToRender := cachedViewData[0]
	cache.Set(cacheKey, cachedViewData[1:])

	// Update history which will be outdated in cache
	trimHistory(&requestConfig.History, 10)
	viewDataToRender.History = requestConfig.History

	return Render(c, http.StatusOK, imageComponent.Image(viewDataToRender))
}

// generateViewData generates page data for the current request.
func generateViewData(requestConfig config.Config, c echo.Context, deviceID string, isPrefetch bool) (common.ViewData, error) {

	const maxImageRetrievalAttepmts = 3

	viewData := common.ViewData{
		DeviceID: deviceID,
		Config:   requestConfig,
	}

	switch requestConfig.Layout {
	case "landscape", "portrait":
		orientation := immich.LandscapeOrientation
		if requestConfig.Layout == "portrait" {
			orientation = immich.PortraitOrientation
		}
		viewDataSingle, err := ProcessViewImageDataWithRatio(orientation, requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Images = append(viewData.Images, viewDataSingle)

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
