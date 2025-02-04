package routes

import (
	"fmt"
	"image"
	"math/rand/v2"
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
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
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
//   - immichAsset: The Immich asset used to query image counts
//   - requestConfig: Configuration containing people, albums and dates to gather assets for
//   - requestID: Identifier for the current request for logging
//
// Returns:
//   - A slice of AssetWithWeighting containing the weightings for each asset source
//   - An error if any database queries fail
func gatherAssetBuckets(immichAsset *immich.ImmichAsset, requestConfig config.Config, requestID, deviceID string) ([]utils.AssetWithWeighting, error) {

	assets := []utils.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		if person == "" || strings.EqualFold(person, "none") {
			continue
		}

		personAssetCount, err := immichAsset.PersonImageCount(person, requestID, deviceID)
		if err != nil {
			if requestConfig.SelectedUser != "" {
				return nil, fmt.Errorf("user '<b>%s</b>' has no Person '%s'. error='%w'", requestConfig.SelectedUser, person, err)
			}
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
		if album == "" || strings.EqualFold(album, "none") {
			continue
		}

		albumAssetCount, err := immichAsset.AlbumImageCount(album, requestID, deviceID)
		if err != nil {
			if requestConfig.SelectedUser != "" {
				return nil, fmt.Errorf("user '<b>%s</b>' has no Album '%s'. error='%w'", requestConfig.SelectedUser, album, err)
			}
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
		if date == "" || strings.EqualFold(date, "none") {
			continue
		}

		// use FetchedAssetsSize as a weighting for date ranges
		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceDateRangeAlbum, ID: date},
			Weight: requestConfig.Kiosk.FetchedAssetsSize,
		})
	}

	if requestConfig.Memories {
		memories := immichAsset.MemoryLaneAssetsCount(requestID, deviceID)
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

// isSleepMode checks if the kiosk should currently be in sleep mode based on configured sleep times
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
func retrieveImage(immichAsset *immich.ImmichAsset, pickedAsset utils.WeightedAsset, albumOrder string, excludedAlbums []string, allowedAssetType []immich.ImmichAssetType, requestID, deviceID string, isPrefetch bool) error {

	switch pickedAsset.Type {
	case kiosk.SourceAlbums:
		switch pickedAsset.ID {
		case kiosk.AlbumKeywordAll:
			pickedAlbumID, err := immichAsset.RandomAlbumFromAllAlbums(requestID, deviceID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordShared:
			pickedAlbumID, err := immichAsset.RandomAlbumFromSharedAlbums(requestID, deviceID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordFavourites, kiosk.AlbumKeywordFavorites:
			return immichAsset.RandomImageFromFavourites(requestID, deviceID, allowedAssetType, isPrefetch)
		}

		switch strings.ToLower(albumOrder) {
		case config.AlbumOrderDescending, config.AlbumOrderDesc, config.AlbumOrderNewest:
			return immichAsset.ImageFromAlbum(pickedAsset.ID, immich.Desc, requestID, deviceID, isPrefetch)
		case config.AlbumOrderAscending, config.AlbumOrderAsc, config.AlbumOrderOldest:
			return immichAsset.ImageFromAlbum(pickedAsset.ID, immich.Asc, requestID, deviceID, isPrefetch)
		default:
			return immichAsset.ImageFromAlbum(pickedAsset.ID, immich.Rand, requestID, deviceID, isPrefetch)
		}

	case kiosk.SourceDateRangeAlbum:
		return immichAsset.RandomImageInDateRange(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourcePerson:
		return immichAsset.RandomImageOfPerson(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourceMemories:
		return immichAsset.RandomMemoryLaneImage(requestID, deviceID, isPrefetch)

	default:
		return immichAsset.RandomImage(requestID, deviceID, isPrefetch)
	}
}

// fetchImagePreview retrieves the preview of an image and logs the time taken.
// It returns the image bytes and an error if any occurs.
func fetchImagePreview(immichAsset *immich.ImmichAsset, requestID, deviceID string, isPrefetch bool) (image.Image, error) {
	imageGet := time.Now()

	imgBytes, err := immichAsset.ImagePreview()
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

	img = utils.ApplyExifOrientation(img, immichAsset.IsLandscape, immichAsset.ExifInfo.Orientation)

	return img, nil
}

// processAsset handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processAsset(immichAsset *immich.ImmichAsset, allowedAssetTypes []immich.ImmichAssetType, requestConfig config.Config, requestID string, deviceID string, requestUrl string, isPrefetch bool) (image.Image, error) {

	assets, err := gatherAssetBuckets(immichAsset, requestConfig, requestID, deviceID)
	if err != nil {
		return nil, err
	}

	pickedAsset := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, assets)

	if err := retrieveImage(immichAsset, pickedAsset, requestConfig.AlbumOrder, requestConfig.ExcludedAlbums, allowedAssetTypes, requestID, deviceID, isPrefetch); err != nil {
		return nil, err
	}

	if requestConfig.ShowAlbumName {
		go immichAsset.AlbumsThatContainAsset(requestID, deviceID)
	}

	//  At this point immichAsset could be a video or an image
	if requestConfig.ExperimentalAlbumVideo && immichAsset.Type == immich.VideoType {
		return processVideo(immichAsset, requestConfig, requestID, deviceID, requestUrl, isPrefetch)
	}

	return processImage(immichAsset, requestID, deviceID, isPrefetch)
}

// processVideo handles retrieving and processing video assets.
// It downloads videos if needed and returns a preview image.
func processVideo(immichAsset *immich.ImmichAsset, requestConfig config.Config, requestID string, deviceID string, requestUrl string, isPrefetch bool) (image.Image, error) {
	// We need to see if the video has been downloaded
	// if so, return nil
	// if it hasn't been downloaded, download it and return a image

	// Video is available
	if VideoManager.IsDownloaded(immichAsset.ID) {
		return fetchImagePreview(immichAsset, requestID, deviceID, isPrefetch)
	}

	//  video is not available, is video downloading?
	if !VideoManager.IsDownloading(immichAsset.ID) {
		go VideoManager.DownloadVideo(*immichAsset, requestConfig, deviceID, requestUrl)
	}

	// if the video is not available, run processAsset again to get a new asset
	return processAsset(immichAsset, immich.AllAssetTypes, requestConfig, requestID, deviceID, requestUrl, isPrefetch)
}

// processImage prepares an image asset for display by setting its source type and retrieving a preview
func processImage(immichAsset *immich.ImmichAsset, requestID string, deviceID string, isPrefetch bool) (image.Image, error) {
	return fetchImagePreview(immichAsset, requestID, deviceID, isPrefetch)
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
func processBlurredImage(img image.Image, assetType immich.ImmichAssetType, config config.Config, requestID, deviceID string, isPrefetch bool) (string, error) {
	isImage := assetType == immich.ImageType
	shouldSkipBlur := !config.BackgroundBlur ||
		strings.EqualFold(config.ImageFit, "cover") ||
		(config.ImageEffect != "" && config.ImageEffect != "none")

	if isImage && shouldSkipBlur {
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

// DrawFaceOnImage draws bounding boxes around detected faces in an image
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

	// if multiple users are given via the url pick a random one
	if len(requestConfig.User) > 0 {
		randomIndex := rand.IntN(len(requestConfig.User))
		requestConfig.SelectedUser = requestConfig.User[randomIndex]
	} else {
		requestConfig.SelectedUser = ""
	}

	immichAsset := immich.NewAsset(requestConfig)

	switch imageOrientation {
	case immich.PortraitOrientation:
		immichAsset.RatioWanted = imageOrientation
	case immich.LandscapeOrientation:
		immichAsset.RatioWanted = imageOrientation
	}

	allowedAssetTypes := immich.ImageOnlyAssetTypes

	if requestConfig.ExperimentalAlbumVideo && isPrefetch {
		allowedAssetTypes = immich.AllAssetTypes
	}

	img, err := processAsset(&immichAsset, allowedAssetTypes, requestConfig, requestID, deviceID, c.Request().URL.String(), isPrefetch)
	if err != nil {
		return common.ViewImageData{}, fmt.Errorf("selecting image: %w", err)
	}

	if strings.EqualFold(requestConfig.ImageEffect, "smart-zoom") && len(immichAsset.People)+len(immichAsset.UnassignedFaces) == 0 {
		immichAsset.CheckForFaces(requestID, deviceID)
	}

	if ShouldDrawFacesOnImages() {
		log.Debug("Drawing faces")
		img = DrawFaceOnImage(img, &immichAsset)
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

	imgBlurString, err := processBlurredImage(img, immichAsset.Type, requestConfig, requestID, deviceID, isPrefetch)
	if err != nil {
		return common.ViewImageData{}, err
	}

	return common.ViewImageData{
		ImmichAsset:   immichAsset,
		ImageData:     imgString,
		ImageBlurData: imgBlurString,
		User:          requestConfig.SelectedUser,
	}, nil
}

// ProcessViewImageData processes view data for an image without orientation constraints
func ProcessViewImageData(requestConfig config.Config, c echo.Context, isPrefetch bool) (common.ViewImageData, error) {
	return processViewImageData("", requestConfig, c, isPrefetch)
}

// ProcessViewImageDataWithRatio processes view data for an image with the specified orientation
func ProcessViewImageDataWithRatio(imageOrientation immich.ImageOrientation, requestConfig config.Config, c echo.Context, isPrefetch bool) (common.ViewImageData, error) {
	return processViewImageData(imageOrientation, requestConfig, c, isPrefetch)
}

// assetToCache stores view data in the cache and triggers prefetch webhooks
func assetToCache(viewDataToAdd common.ViewData, requestConfig *config.Config, deviceID string, requestData *common.RouteRequestData, c echo.Context) {

	cache.AssetToCache(viewDataToAdd, requestConfig, deviceID, c.Request().URL.String())

	go webhooks.Trigger(requestData, KioskVersion, webhooks.PrefetchAsset, viewDataToAdd)
}

// assetPreFetch handles prefetching assets for the current request
func assetPreFetch(requestData *common.RouteRequestData, c echo.Context) {

	requestConfig := requestData.RequestConfig
	requestID := requestData.RequestID
	deviceID := requestData.DeviceID

	viewDataToAdd, err := generateViewData(requestConfig, c, requestID, true)
	if err != nil {
		log.Error("generateViewData", "prefetch", true, "err", err)
		return
	}

	assetToCache(viewDataToAdd, &requestConfig, deviceID, requestData, c)
}

// fromCache retrieves cached page data for a given request and device ID.
func fromCache(urlString string, deviceID string) []common.ViewData {
	cacheKey := cache.ViewCacheKey(urlString, deviceID)
	if data, found := cache.Get(cacheKey); found {
		cachedPageData, ok := data.([]common.ViewData)
		if !ok {
			log.Error("cache: invalid data type", "type", fmt.Sprintf("%T", data), "key", cacheKey)
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
	utils.TrimHistory(&requestConfig.History, 10)
	viewDataToRender.History = requestConfig.History

	if requestConfig.ExperimentalAlbumVideo && viewDataToRender.Assets[0].ImmichAsset.Type == immich.VideoType {
		return Render(c, http.StatusOK, videoComponent.Video(viewDataToRender))
	}

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
		viewData.Assets = append(viewData.Assets, viewDataSingle)

	case "splitview":
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSplitView)

		if viewDataSplitView.ImmichAsset.Type == immich.VideoType || viewDataSplitView.ImmichAsset.IsLandscape {
			return viewData, nil
		}

		// Second image
		for i := 0; i < maxImageRetrievalAttepmts; i++ {
			viewDataSplitViewSecond, err := ProcessViewImageDataWithRatio(immich.PortraitOrientation, requestConfig, c, isPrefetch)
			if err != nil {
				return viewData, err
			}

			if viewDataSplitView.ImmichAsset.ID != viewDataSplitViewSecond.ImmichAsset.ID {
				viewData.Assets = append(viewData.Assets, viewDataSplitViewSecond)
				break
			}
		}

	case "splitview-landscape":
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSplitView)

		if viewDataSplitView.ImmichAsset.IsPortrait {
			return viewData, nil
		}

		// Second image
		for i := 0; i < maxImageRetrievalAttepmts; i++ {
			viewDataSplitViewSecond, err := ProcessViewImageDataWithRatio(immich.LandscapeOrientation, requestConfig, c, isPrefetch)
			if err != nil {
				return viewData, err
			}

			if viewDataSplitView.ImmichAsset.ID != viewDataSplitViewSecond.ImmichAsset.ID {
				viewData.Assets = append(viewData.Assets, viewDataSplitViewSecond)
				break
			}
		}

	default:
		viewDataSingle, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSingle)
	}

	return viewData, nil
}
