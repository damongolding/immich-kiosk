package routes

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
)

// PreviousHistoryAsset handles requests to show previously viewed assets in the navigation history.
// It processes both images and videos in parallel, retrieving asset info and generating
// regular and blurred preview images.
//
// For each previous asset:
// - Fetches asset info and album details if configured
// - Generates regular and blurred preview images
// - Returns video component for videos when ShowTime is enabled, image component otherwise
//
// Parameters:
// - baseConfig: Application configuration
// - com: Common functionality and context
//
// Returns:
// - echo.HandlerFunc that processes the request
// - 204 No Content if sleep mode is active (and not disabled) or history has < 2 entries
// - Error if asset processing fails
//
// Triggers webhook on successful render.
func PreviousHistoryAsset(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {
		return historyAsset(baseConfig, com, c, false)
	}
}

// NextHistoryAsset handles requests to show the next asset in the navigation history.
// It delegates to historyAsset with useNextImage=true to handle displaying the next asset.
func NextHistoryAsset(baseConfig *config.Config, com *common.Common, c echo.Context) error {
	return historyAsset(baseConfig, com, c, true)
}

// historyAsset handles the core logic for showing previous/next assets from navigation history.
// It retrieves the requested asset(s), processes images and metadata in parallel, and renders
// the appropriate view component.
//
// Parameters:
// - baseConfig: Application configuration
// - com: Common functionality and context
// - c: Echo context for the HTTP request
// - useNextImage: If true, shows next asset, if false shows previous
//
// Returns error if asset processing fails.
func historyAsset(baseConfig *config.Config, com *common.Common, c echo.Context, useNextImage bool) error {
	requestData, err := InitializeRequestData(c, baseConfig)
	if err != nil {
		return err
	}

	if requestData == nil {
		log.Info("Refreshing clients")
		return nil
	}

	requestConfig := requestData.RequestConfig
	requestID := requestData.RequestID
	deviceID := requestData.DeviceID

	log.Debug(
		requestID,
		"method", c.Request().Method,
		"deviceID", requestData.DeviceID,
		"path", c.Request().URL.String(),
		"requestConfig", requestConfig.String(),
	)

	historyLen := len(requestConfig.History)

	if (!requestConfig.DisableSleep && isSleepMode(requestConfig)) || historyLen < 2 {
		return c.NoContent(http.StatusNoContent)
	}

	historyEntry, entryIndex := findHistoryEntry(requestConfig.History, useNextImage)

	if historyEntry == "" {
		if useNextImage {
			historyEntry = requestConfig.History[historyLen-1]
			entryIndex = historyLen - 1
		} else {
			historyEntry = requestConfig.History[historyLen-2]
			entryIndex = historyLen - 2
		}
	}

	wantedAssets := strings.Split(historyEntry, ",")
	if len(wantedAssets) == 0 || (len(wantedAssets) == 1 && wantedAssets[0] == "") {
		return fmt.Errorf("no valid assets found in history entry: %s", historyEntry)
	}

	requestConfig.History[entryIndex] = kiosk.HistoryIndicator + requestConfig.History[entryIndex]

	if requestConfig.UseOfflineMode && requestConfig.OfflineMode.Enabled {
		return historyAssetOffline(c, requestID, deviceID, wantedAssets, requestConfig.History, com.Secret())
	}

	viewData := common.ViewData{
		KioskVersion: KioskVersion,
		RequestID:    requestID,
		DeviceID:     deviceID,
		Assets:       make([]common.ViewImageData, len(wantedAssets)),
		Queries:      c.QueryParams(),
		Config:       requestConfig,
	}

	g, _ := errgroup.WithContext(c.Request().Context())

	for i, assetID := range wantedAssets {

		parts := strings.Split(assetID, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid history entry format: %s", assetID)
		}

		prevAssetsID, currentAssetID, selectedUser := i, parts[0], parts[1]

		g.Go(func(prevAssetsID int, currentAssetID string) func() error {
			return func() error {
				requestConfig.SelectedUser = selectedUser

				asset := immich.New(com.Context(), requestConfig)
				asset.ID = currentAssetID

				var wg sync.WaitGroup
				wg.Add(1)

				go func(asset *immich.Asset, requestID string, wg *sync.WaitGroup) {
					defer wg.Done()
					var processingErr error

					if assetInfoErr := asset.AssetInfo(requestID, deviceID); assetInfoErr != nil {
						processingErr = fmt.Errorf("failed to get asset info: %w", assetInfoErr)
						log.Error(processingErr)
					}

					asset.AddRatio()

					if requestConfig.ShowAlbumName {
						asset.AlbumsThatContainAsset(requestID, deviceID)
					}

				}(&asset, requestID, &wg)

				var imgString, imgBlurString string

				defer func() {
					viewData.Assets[prevAssetsID] = common.ViewImageData{
						ImmichAsset:   asset,
						ImageData:     imgString,
						ImageBlurData: imgBlurString,
						User:          selectedUser,
					}
				}()

				// Image processing isn't required for video, audio, or other types
				// So if this fails, we can still proceed with the asset view
				imgBytes, previewErr := asset.ImagePreview()
				if previewErr != nil {
					switch asset.Type {
					case immich.ImageType:
						return fmt.Errorf("retrieving asset: %w", previewErr)

					case immich.VideoType, immich.AudioType, immich.OtherType:
						wg.Wait()
						return nil
					}
				}

				img, byteErr := utils.BytesToImage(imgBytes)
				if byteErr != nil {
					return byteErr
				}

				imgString, base64Err := imageToBase64(img, requestConfig, requestID, deviceID, "Converted", false)
				if base64Err != nil {
					return fmt.Errorf("converting image to base64: %w", base64Err)
				}

				imgBlurString, blurErr := processBlurredImage(img, asset.Type, requestConfig, requestID, deviceID, false)
				if blurErr != nil {
					return fmt.Errorf("converting blurred image to base64: %w", blurErr)
				}

				wg.Wait()

				return nil
			}
		}(prevAssetsID, currentAssetID))
	}

	// Wait for all goroutines to complete and check for errors
	errGroupWait := g.Wait()
	if errGroupWait != nil {
		return RenderError(c, errGroupWait, "processing images", requestConfig.Refresh)
	}

	webhookEvent := webhooks.PreviousHistoryAsset
	if useNextImage {
		webhookEvent = webhooks.NextHistoryAsset
	}

	go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhookEvent, viewData)

	if len(viewData.Assets) > 0 && requestConfig.ShowTime && viewData.Assets[0].ImmichAsset.Type == immich.VideoType {
		return Render(c, http.StatusOK, videoComponent.Video(viewData, com.Secret()))
	}

	return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))
}

// findHistoryEntry searches through the history slice to find and return the appropriate
// history entry and its index based on navigation direction.
//
// Parameters:
// - history: Slice of history entries to search through
// - useNextImage: If true, looks for next entry, if false looks for previous entry
//
// Returns:
// - string: The found history entry, or empty string if none found
// - int: The index of the found entry
func findHistoryEntry(history []string, useNextImage bool) (string, int) {

	historyLen := len(history)
	entry := ""
	entryIndex := 0

	for i, h := range history {
		if strings.HasPrefix(h, kiosk.HistoryIndicator) {
			switch useNextImage {
			case true:
				if i+1 >= historyLen {
					continue
				}
				entry = history[i+1]
				entryIndex = i + 1
			case false:
				if i == 0 {
					continue
				}
				entry = history[i-1]
				entryIndex = i - 1
			}
			history[i] = strings.Replace(h, kiosk.HistoryIndicator, "", 1)
		}
	}

	return entry, entryIndex
}

// historyAssetOffline handles displaying assets when in offline mode by loading
// cached data from the filesystem.
//
// Parameters:
// - c: Echo context for the HTTP request
// - requestID: Unique identifier for the request
// - deviceID: Device identifier
// - wantedAssets: Slice of asset IDs to display
// - history: Navigation history
// - secret: Secret key for rendering
//
// Returns:
// - error if loading or rendering cached data fails
func historyAssetOffline(c echo.Context, requestID, deviceID string, wantedAssets, history []string, secret string) error {
	replacer := strings.NewReplacer(
		kiosk.HistoryIndicator, "",
		":", "",
		",", "",
	)

	var filename string
	for _, wa := range wantedAssets {
		filename += replacer.Replace(wa)
	}

	filename = generateCacheFilename(filename)

	filename = path.Join(OfflineAssetsPath, filename)

	viewData, loadMsgpackErr := loadMsgpackZstd(filename)
	if loadMsgpackErr != nil {
		log.Error("OfflineMode: loadMsgpackZstd", "picked", filename, "err", loadMsgpackErr)
		return loadMsgpackErr
	}

	viewData.KioskVersion = KioskVersion
	viewData.RequestID = requestID
	viewData.DeviceID = deviceID
	viewData.History = history

	return Render(c, http.StatusOK, imageComponent.Image(viewData, secret))
}
