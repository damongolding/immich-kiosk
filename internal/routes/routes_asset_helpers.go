package routes

import (
	"context"
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
func gatherAssetBuckets(immichAsset *immich.Asset, requestConfig config.Config, requestID, deviceID string) ([]utils.AssetWithWeighting, error) {

	assets := []utils.AssetWithWeighting{}

	for _, person := range requestConfig.Person {
		if person == "" || strings.EqualFold(person, "none") {
			continue
		}

		personAssetCount, personCountErr := immichAsset.PersonImageCount(person, requestID, deviceID)
		if personCountErr != nil {
			if requestConfig.SelectedUser != "" {
				return nil, fmt.Errorf("user '<b>%s</b>' has no Person '%s'. error='%w'", requestConfig.SelectedUser, person, personCountErr)
			}
			return nil, fmt.Errorf("getting person image count: %w", personCountErr)
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

		albumAssetCount, albumCountErr := immichAsset.AlbumImageCount(album, requestID, deviceID)
		if albumCountErr != nil {
			if requestConfig.SelectedUser != "" {
				return nil, fmt.Errorf("user '<b>%s</b>' has no Album '%s'. error='%w'", requestConfig.SelectedUser, album, albumCountErr)
			}
			return nil, fmt.Errorf("getting album asset count: %w", albumCountErr)
		}

		if albumAssetCount == 0 {
			log.Error("No assets found for", "album", album)
			continue
		}

		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceAlbum, ID: album},
			Weight: albumAssetCount,
		})
	}

	for _, tag := range requestConfig.Tag {
		if tag == "" || strings.EqualFold(tag, "none") {
			continue
		}

		tags, _, tagsErr := immichAsset.AllTags(requestID, deviceID)
		if tagsErr != nil {
			log.Error("getting tags", "err", tagsErr)
			continue
		}

		tagData, tagErr := tags.Get(tag)
		if tagErr != nil {
			log.Error("getting tag from tags", "tag", tag, "err", tagErr)
			continue
		}

		taggedAssetsCount, tagCountErr := immichAsset.AssetsWithTagCount(tagData.ID, requestID, deviceID)
		if tagCountErr != nil {
			if requestConfig.SelectedUser != "" {
				return nil, fmt.Errorf("user '<b>%s</b>' has no assets with tag '%s'. error='%w'", requestConfig.SelectedUser, tagData.Value, tagCountErr)
			}
			return nil, fmt.Errorf("getting tagged asset count: %w", tagCountErr)
		}

		if taggedAssetsCount == 0 {
			log.Error("No assets found with", "tag", tagData.Value)
			continue
		}

		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceTag, ID: tagData.ID},
			Weight: taggedAssetsCount,
		})
	}

	for _, date := range requestConfig.Date {
		if date == "" || strings.EqualFold(date, "none") {
			continue
		}

		// use FetchedAssetsSize as a weighting for date ranges
		assets = append(assets, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceDateRange, ID: date},
			Weight: requestConfig.Kiosk.FetchedAssetsSize,
		})
	}

	if requestConfig.Memories {
		memories := immichAsset.MemoriesAssetsCount(requestID, deviceID)
		if memories == 0 {
			log.Warn("No assets found for memories")
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
func retrieveImage(immichAsset *immich.Asset, pickedAsset utils.WeightedAsset, albumOrder string, excludedAlbums []string, allowedAssetType []immich.AssetType, requestID, deviceID string, isPrefetch bool) error {

	switch pickedAsset.Type {
	case kiosk.SourceAlbum:
		switch pickedAsset.ID {
		case kiosk.AlbumKeywordAll:
			pickedAlbumID, err := immichAsset.RandomAlbumFromAllAlbums(requestID, deviceID, excludedAlbums)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case kiosk.AlbumKeywordOwned:
			pickedAlbumID, err := immichAsset.RandomAlbumFromOwnedAlbums(requestID, deviceID, excludedAlbums)
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
			return immichAsset.AssetFromAlbum(pickedAsset.ID, immich.Desc, requestID, deviceID)
		case config.AlbumOrderAscending, config.AlbumOrderAsc, config.AlbumOrderOldest:
			return immichAsset.AssetFromAlbum(pickedAsset.ID, immich.Asc, requestID, deviceID)
		default:
			return immichAsset.AssetFromAlbum(pickedAsset.ID, immich.Rand, requestID, deviceID)
		}

	case kiosk.SourceDateRange:
		return immichAsset.RandomImageInDateRange(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourcePerson:
		if pickedAsset.ID == kiosk.PersonKeywordAll {
			pickedPersonID, err := immichAsset.RandomPersonFromAllPeople(requestID, deviceID, true)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedPersonID
		}

		return immichAsset.RandomImageOfPerson(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourceMemories:
		return immichAsset.RandomMemoryAsset(requestID, deviceID)

	case kiosk.SourceTag:
		return immichAsset.RandomAssetWithTag(pickedAsset.ID, requestID, deviceID, isPrefetch)

	case kiosk.SourceRandom:
		fallthrough

	default:
		return immichAsset.RandomImage(requestID, deviceID, isPrefetch)
	}

}

// fetchImagePreview retrieves and processes the preview of an image from the Immich asset.
// Parameters:
//   - immichAsset: The Immich asset to retrieve the preview from
//   - isOriginal: Whether to preserve the original image orientation
//   - requestID: ID of the current request for logging
//   - deviceID: ID of the device making the request
//   - isPrefetch: Whether this is a prefetch request
//
// Returns:
//   - The processed image preview
//   - Any error that occurred during retrieval or processing
func fetchImagePreview(immichAsset *immich.Asset, isOriginal bool, requestID, deviceID string, isPrefetch bool) (image.Image, error) {
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

	if isOriginal {
		img = utils.ApplyExifOrientation(img, immichAsset.IsLandscape, immichAsset.ExifInfo.Orientation)
	}

	return img, nil
}

// processAsset handles the entire process of selecting and retrieving an image.
// It returns the image bytes and an error if any step fails.
func processAsset(immichAsset *immich.Asset, allowedAssetTypes []immich.AssetType, requestConfig config.Config, requestID string, deviceID string, requestURL string, isPrefetch bool) (image.Image, error) {

	var err error

	assets, assetsErr := gatherAssetBuckets(immichAsset, requestConfig, requestID, deviceID)
	if assetsErr != nil {
		return nil, assetsErr
	}

	for range maxProcessAssetRetries {

		pickedAsset := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, assets)

		err = retrieveImage(immichAsset, pickedAsset, requestConfig.AlbumOrder, requestConfig.ExcludedAlbums, allowedAssetTypes, requestID, deviceID, isPrefetch)
		if err != nil {
			continue
		}

		//  At this point immichAsset could be a video or an image
		if requestConfig.ExperimentalAlbumVideo && immichAsset.Type == immich.VideoType {
			return processVideo(immichAsset, requestConfig, requestID, deviceID, requestURL, isPrefetch)
		}

		return processImage(immichAsset, requestConfig.UseOriginalImage, requestID, deviceID, isPrefetch)
	}

	return nil, fmt.Errorf("%w: max retries exceeded", err)

}

// processVideo handles retrieving and processing video assets.
// It downloads videos if needed and returns a preview image.
func processVideo(immichAsset *immich.Asset, requestConfig config.Config, requestID string, deviceID string, requestURL string, isPrefetch bool) (image.Image, error) {
	// We need to see if the video has been downloaded
	// if so, return nil
	// if it hasn't been downloaded, download it and return a image

	// Video is available
	if VideoManager.IsDownloaded(immichAsset.ID) {
		return fetchImagePreview(immichAsset, requestConfig.UseOriginalImage, requestID, deviceID, isPrefetch)
	}

	//  video is not available, is video downloading?
	if !VideoManager.IsDownloading(immichAsset.ID) {
		go VideoManager.DownloadVideo(*immichAsset, requestConfig, deviceID, requestURL)
	}

	// if the video is not available, run processAsset again to get a new asset
	return processAsset(immichAsset, immich.AllAssetTypes, requestConfig, requestID, deviceID, requestURL, isPrefetch)
}

// processImage prepares an image asset for display by setting its source type and retrieving a preview
func processImage(immichAsset *immich.Asset, isOriginal bool, requestID string, deviceID string, isPrefetch bool) (image.Image, error) {
	return fetchImagePreview(immichAsset, isOriginal, requestID, deviceID, isPrefetch)
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
func processBlurredImage(img image.Image, assetType immich.AssetType, config config.Config, requestID, deviceID string, isPrefetch bool) (string, error) {
	isImage := assetType == immich.ImageType
	shouldSkipBlur := !config.BackgroundBlur ||
		strings.EqualFold(config.ImageFit, "cover") ||
		(config.ImageEffect != "" && config.ImageEffect != "none" && config.Layout != "single")

	if isImage && shouldSkipBlur {
		return "", nil
	}

	startTime := time.Now()
	imgBlur, err := utils.BlurImage(img, config.BackgroundBlurAmount, config.OptimizeImages, config.ClientData.Width, config.ClientData.Height)
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
func DrawFaceOnImage(img image.Image, i *immich.Asset) image.Image {

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

// processViewImageData processes an image request and returns view data for display.
// It handles the complete workflow from selecting an image to preparing it for display,
// including face detection, optimization, and format conversion.
//
// Parameters:
//   - requestConfig: Configuration settings for the request
//   - c: Copy of the request context
//   - isPrefetch: Whether this is a prefetch request
//   - options: Additional options for image processing
//
// Returns:
//   - ViewImageData containing the processed image and metadata
//   - Error if any step fails
func processViewImageData(requestConfig config.Config, c common.ContextCopy, isPrefetch bool, options common.ViewImageDataOptions) (common.ViewImageData, error) {
	// Initialize request metadata
	metadata := requestMetadata{
		requestID: utils.ColorizeRequestID(c.ResponseHeader.Get(echo.HeaderXRequestID)),
		deviceID:  c.RequestHeader.Get("kiosk-device-id"),
		urlString: c.URL.String(),
	}

	// Set up configuration
	setupRequestConfig(&requestConfig)
	immichAsset := setupImmichAsset(requestConfig, options.ImageOrientation)
	allowedAssetTypes := determineAllowedAssetTypes(requestConfig, isPrefetch)

	// Handle relative asset configuration if needed
	if options.RelativeAssetWanted {
		handleRelativeAssetConfig(&requestConfig, options)
	}

	// Process image
	img, err := processAsset(&immichAsset, allowedAssetTypes, requestConfig, metadata.requestID, metadata.deviceID, metadata.urlString, isPrefetch)
	if err != nil {
		return common.ViewImageData{}, fmt.Errorf("selecting asset: %w", err)
	}

	// Handle face detection and smart zoom
	img = handleFaceProcessing(img, &immichAsset, requestConfig, metadata)

	// Optimize image if needed
	if requestConfig.OptimizeImages {
		img, err = utils.OptimizeImage(img, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
		if err != nil {
			return common.ViewImageData{}, err
		}
	}

	// Convert images to required formats
	imgString, imgBlurString, err := convertImages(img, immichAsset.Type, requestConfig, metadata, isPrefetch)
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

// setupRequestConfig configures the selected user for the request by picking a random
// user from the config if multiple users are provided, otherwise sets to empty string
func setupRequestConfig(config *config.Config) {
	if len(config.User) > 0 {
		randomIndex := rand.IntN(len(config.User))
		config.SelectedUser = config.User[randomIndex]
	} else {
		config.SelectedUser = ""
	}
}

// setupImmichAsset creates and configures a new ImmichAsset based on the provided config
// and orientation settings
func setupImmichAsset(config config.Config, orientation immich.ImageOrientation) immich.Asset {
	asset := immich.New(context.Background(), config)
	if orientation == immich.PortraitOrientation || orientation == immich.LandscapeOrientation {
		asset.RatioWanted = orientation
	}
	return asset
}

// determineAllowedAssetTypes returns the allowed asset types based on config settings
// Returns AllAssetTypes if experimental video is enabled and isPrefetch is true,
// otherwise returns ImageOnlyAssetTypes
func determineAllowedAssetTypes(config config.Config, isPrefetch bool) []immich.AssetType {
	if config.ExperimentalAlbumVideo && isPrefetch {
		return immich.AllAssetTypes
	}
	return immich.ImageOnlyAssetTypes
}

// handleRelativeAssetConfig updates the config buckets based on the relative asset options.
// Resets existing buckets and configures the appropriate bucket based on the asset source type.
func handleRelativeAssetConfig(config *config.Config, options common.ViewImageDataOptions) {
	config.ResetBuckets()
	config.Memories = false

	switch options.RelativeAssetBucket {
	case kiosk.SourceAlbum:
		config.Album = append(config.Album, options.RelativeAssetBucketID)
	case kiosk.SourcePerson:
		config.Person = append(config.Person, options.RelativeAssetBucketID)
	case kiosk.SourceDateRange:
		config.Date = append(config.Date, options.RelativeAssetBucketID)
	case kiosk.SourceTag:
		config.Tag = append(config.Tag, options.RelativeAssetBucketID)
	case kiosk.SourceMemories:
		config.Memories = true
	case kiosk.SourceRandom:
	}
}

// handleFaceProcessing processes face detection and drawing for an image.
// Checks for faces if smart-zoom is enabled and draws faces if configured.
// Returns the processed image.
func handleFaceProcessing(img image.Image, asset *immich.Asset, config config.Config, metadata requestMetadata) image.Image {
	if strings.EqualFold(config.ImageEffect, "smart-zoom") && len(asset.People)+len(asset.UnassignedFaces) == 0 {
		asset.CheckForFaces(metadata.requestID, metadata.deviceID)
	}

	if ShouldDrawFacesOnImages() {
		log.Debug("Drawing faces")
		return DrawFaceOnImage(img, asset)
	}
	return img
}

// convertImages converts the provided image to base64 strings for both normal and blurred versions.
// Returns the base64 encoded normal image, blurred image, and any error that occurred.
func convertImages(img image.Image, assetType immich.AssetType, config config.Config, metadata requestMetadata, isPrefetch bool) (string, string, error) {
	imgString, err := imageToBase64(img, config, metadata.requestID, metadata.deviceID, "Converted", isPrefetch)
	if err != nil {
		return "", "", err
	}

	imgBlurString, err := processBlurredImage(img, assetType, config, metadata.requestID, metadata.deviceID, isPrefetch)
	if err != nil {
		return "", "", err
	}

	return imgString, imgBlurString, nil
}

// ProcessViewImageData processes view data for an image without orientation constraints
func ProcessViewImageData(requestConfig config.Config, c common.ContextCopy, isPrefetch bool) (common.ViewImageData, error) {
	return processViewImageData(requestConfig, c, isPrefetch, common.ViewImageDataOptions{})
}

func ProcessViewImageDataWithOptions(requestConfig config.Config, c common.ContextCopy, isPrefetch bool, options common.ViewImageDataOptions) (common.ViewImageData, error) {
	return processViewImageData(requestConfig, c, isPrefetch, options)
}

// assetToCache stores view data in the cache and triggers prefetch webhooks
func assetToCache(ctx context.Context, viewDataToAdd common.ViewData, requestConfig *config.Config, deviceID string, requestData *common.RouteRequestData, c common.ContextCopy) {

	cache.AssetToCache(viewDataToAdd, requestConfig, deviceID, c.URL.String())

	go webhooks.Trigger(ctx, requestData, KioskVersion, webhooks.PrefetchAsset, viewDataToAdd)
}

// assetPreFetch handles prefetching assets for the current request
func assetPreFetch(common *common.Common, requestData *common.RouteRequestData, c common.ContextCopy) {

	requestConfig := requestData.RequestConfig
	requestID := requestData.RequestID
	deviceID := requestData.DeviceID

	viewDataToAdd, err := generateViewData(requestConfig, c, requestID, deviceID, true)
	if err != nil {
		log.Error("generateViewData", "prefetch", true, "err", err)
		return
	}

	assetToCache(common.Context(), viewDataToAdd, &requestConfig, deviceID, requestData, c)
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
func renderCachedViewData(c echo.Context, cachedViewData []common.ViewData, requestConfig *config.Config, requestID, deviceID, secret string) error {

	log.Debug(requestID, "deviceID", deviceID, "cache hit for new image", true)

	cacheKey := cache.ViewCacheKey(c.Request().URL.String(), deviceID)

	viewDataToRender := cachedViewData[0]
	cache.Set(cacheKey, cachedViewData[1:])

	// Update history which will be outdated in cache
	utils.TrimHistory(&requestConfig.History, kiosk.HistoryLimit)
	viewDataToRender.History = requestConfig.History

	if requestConfig.ExperimentalAlbumVideo && viewDataToRender.Assets[0].ImmichAsset.Type == immich.VideoType {
		return Render(c, http.StatusOK, videoComponent.Video(viewDataToRender, secret))
	}

	return Render(c, http.StatusOK, imageComponent.Image(viewDataToRender, secret))
}

// fetchSecondSplitViewAsset attempts to retrieve a second asset for split view layouts that is different from the first asset.
// It tries up to three times to obtain a unique asset and appends it to the provided ViewData if successful.
// Returns an error if asset retrieval fails.
func fetchSecondSplitViewAsset(viewData *common.ViewData, viewDataSplitView common.ViewImageData, requestConfig config.Config, c common.ContextCopy, isPrefetch bool, options common.ViewImageDataOptions) error {
	const maxImageRetrievalAttempts = 3

	for range maxImageRetrievalAttempts {
		viewDataSplitViewSecond, err := ProcessViewImageDataWithOptions(requestConfig, c, isPrefetch, options)
		if err != nil {
			return err
		}

		if viewDataSplitView.ImmichAsset.ID != viewDataSplitViewSecond.ImmichAsset.ID {
			viewData.Assets = append(viewData.Assets, viewDataSplitViewSecond)
			return nil
		}
	}
	return nil
}

// determineLayoutMode returns the appropriate layout mode based on the requested layout and client display dimensions.
// If the requested layout is split view and the client height exceeds the width, it switches to landscape split view mode.
func determineLayoutMode(layout string, clientHeight, clientWidth int) string {
	if layout == kiosk.LayoutSplitview && clientHeight > clientWidth {
		return kiosk.LayoutSplitviewLandscape
	}
	return layout
}

// generateViewData prepares view data for a kiosk page request based on the specified layout and client display dimensions.
// It selects and processes one or two assets as needed for the layout, handling orientation and split view logic, and returns the resulting ViewData or an error.
func generateViewData(requestConfig config.Config, c common.ContextCopy, requestID, deviceID string, isPrefetch bool) (common.ViewData, error) {

	viewData := common.ViewData{
		RequestID: requestID,
		DeviceID:  deviceID,
		Config:    requestConfig,
	}

	requestConfig.Layout = determineLayoutMode(requestConfig.Layout, requestConfig.ClientData.Height, requestConfig.ClientData.Width)

	switch requestConfig.Layout {
	case kiosk.LayoutLandscape, kiosk.LayoutPortrait:
		options := common.ViewImageDataOptions{
			ImageOrientation: immich.LandscapeOrientation,
		}
		if requestConfig.Layout == kiosk.LayoutPortrait {
			options.ImageOrientation = immich.PortraitOrientation
		}
		viewDataSingle, err := ProcessViewImageDataWithOptions(requestConfig, c, isPrefetch, options)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSingle)

	case kiosk.LayoutSplitview:
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSplitView)

		if viewDataSplitView.ImmichAsset.Type == immich.VideoType || viewDataSplitView.ImmichAsset.IsLandscape {
			return viewData, nil
		}

		options := common.ViewImageDataOptions{
			RelativeAssetWanted:   true,
			RelativeAssetBucket:   viewDataSplitView.ImmichAsset.Bucket,
			RelativeAssetBucketID: viewDataSplitView.ImmichAsset.BucketID,
			ImageOrientation:      immich.PortraitOrientation,
		}

		// Second image
		if secondAssetErr := fetchSecondSplitViewAsset(&viewData, viewDataSplitView, requestConfig, c, isPrefetch, options); secondAssetErr != nil {
			return viewData, secondAssetErr
		}

	case kiosk.LayoutSplitviewLandscape:
		viewDataSplitView, err := ProcessViewImageData(requestConfig, c, isPrefetch)
		if err != nil {
			return viewData, err
		}
		viewData.Assets = append(viewData.Assets, viewDataSplitView)

		if viewDataSplitView.ImmichAsset.IsPortrait {
			return viewData, nil
		}

		options := common.ViewImageDataOptions{
			RelativeAssetWanted:   true,
			RelativeAssetBucket:   viewDataSplitView.ImmichAsset.Bucket,
			RelativeAssetBucketID: viewDataSplitView.ImmichAsset.BucketID,
			ImageOrientation:      immich.LandscapeOrientation,
		}

		// Second image
		if secondAssetErr := fetchSecondSplitViewAsset(&viewData, viewDataSplitView, requestConfig, c, isPrefetch, options); secondAssetErr != nil {
			return viewData, secondAssetErr
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
