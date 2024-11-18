package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/disintegration/imaging"
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
			Asset:  utils.WeightedAsset{Type: "PERSON", ID: person},
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

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	switch pickedAsset.Type {
	case "ALBUM":
		switch pickedAsset.ID {
		case immich.AlbumKeywordAll:
			pickedAlbumID, err := immichImage.RandomAlbumFromAllAlbums(requestID)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case immich.AlbumKeywordShared:
			pickedAlbumID, err := immichImage.RandomAlbumFromSharedAlbums(requestID)
			if err != nil {
				return err
			}
			pickedAsset.ID = pickedAlbumID
		case immich.AlbumKeywordFavourites, immich.AlbumKeywordFavorites:
			return immichImage.RandomImageFromFavourites(requestID, kioskDeviceID, isPrefetch)
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
	if !config.BackgroundBlur || strings.EqualFold(config.ImageFit, "cover") || (config.ImageEffect != "" && config.ImageEffect != "none") {
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

func DrawFaceOnImage(imgBytes []byte, i *immich.ImmichAsset) []byte {

	if len(i.People) == 0 && len(i.UnassignedFaces) == 0 {
		log.Debug("no people found")
		return imgBytes
	}

	img, err := imaging.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Error("could not decode image", "err", err)
		return imgBytes
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

	out := dc.Image()

	buf := new(bytes.Buffer)

	err = imaging.Encode(buf, out, imaging.JPEG)
	if err != nil {
		log.Error("Error encodeing image:", err)
		return imgBytes
	}

	return buf.Bytes()
}

// processViewImageData handles the entire process of preparing page data including image processing.
// It returns the ImageData and an error if any step fails.
func processViewImageData(imageOrientation immich.ImageOrientation, requestConfig config.Config, c echo.Context, isPrefetch bool) (views.ImageData, error) {
	requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceID := c.Request().Header.Get("kiosk-device-id")

	immichImage := immich.NewImage(requestConfig)

	switch imageOrientation {
	case immich.PortraitOrientation:
		immichImage.RatioWanted = imageOrientation
	case immich.LandscapeOrientation:
		immichImage.RatioWanted = imageOrientation
	}

	imgBytes, err := processImage(&immichImage, requestConfig, requestID, kioskDeviceID, isPrefetch)
	if err != nil {
		return views.ImageData{}, fmt.Errorf("selecting image: %w", err)
	}

	if strings.EqualFold(requestConfig.ImageEffect, "smart-zoom") && len(immichImage.People)+len(immichImage.UnassignedFaces) == 0 {
		immichImage.CheckForFaces(requestID)
	}

	if ShouldDrawFacesOnImages() {
		log.Debug("Drawing faces")
		imgBytes = DrawFaceOnImage(imgBytes, &immichImage)
	}

	img, err := imageToBase64(imgBytes, requestConfig, requestID, kioskDeviceID, "Converted", isPrefetch)
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

func ProcessViewImageDataWithRatio(imageOrientation immich.ImageOrientation, requestConfig config.Config, c echo.Context, isPrefetch bool) (views.ImageData, error) {
	return processViewImageData(imageOrientation, requestConfig, c, isPrefetch)
}

func imagePreFetch(requestConfig config.Config, c echo.Context, kioskDeviceID string) {

	viewDataToAdd, err := generateViewData(requestConfig, c, kioskDeviceID, true)
	if err != nil {
		log.Error("prefetch", "err", err)
		return
	}

	trimHistory(&requestConfig.History, 10)

	cachedViewData := []views.ViewData{}

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	cacheKey := c.Request().URL.String() + kioskDeviceID

	if data, found := ViewDataCache.Get(cacheKey); found {
		cachedViewData = data.([]views.ViewData)
	}

	cachedViewData = append(cachedViewData, viewDataToAdd)

	ViewDataCache.Set(cacheKey, cachedViewData, cache.DefaultExpiration)

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
	if data, found := ViewDataCache.Get(cacheKey); found {
		cachedPageData := data.([]views.ViewData)
		if len(cachedPageData) > 0 {
			return cachedPageData
		}
		ViewDataCache.Delete(cacheKey)
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
	ViewDataCache.Set(cacheKey, cachedViewData[1:], cache.DefaultExpiration)

	// Update history which will be outdated in cache
	trimHistory(&requestConfig.History, 10)
	viewDataToRender.History = requestConfig.History

	return Render(c, http.StatusOK, views.Image(viewDataToRender))
}

// generateViewData generates page data for the current request.
func generateViewData(requestConfig config.Config, c echo.Context, kioskDeviceID string, isPrefetch bool) (views.ViewData, error) {

	const maxImageRetrievalAttepmts = 3

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

		log.Info("in landscape mode 1st image", "IsPortrait", viewDataSplitView.ImmichImage.IsPortrait)

		if viewDataSplitView.ImmichImage.IsPortrait {
			return viewData, nil
		}

		log.Info("in landscape mode 2nd image", "IsPortrait", viewDataSplitView.ImmichImage.IsPortrait)

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
