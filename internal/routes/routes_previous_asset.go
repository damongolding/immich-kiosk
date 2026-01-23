package routes

import (
	"fmt"
	"image/color"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
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

// historyAsset processes and displays either the previous or next asset(s) from the navigation history, handling both online and offline modes.
// It retrieves the relevant history entry, fetches asset metadata and image previews concurrently, and prepares view data for rendering. If offline mode is enabled, it loads cached asset data instead. The function triggers a webhook event corresponding to the navigation direction and renders either an image or video component based on the asset type.
// Returns an error if asset retrieval, image processing, or view rendering fails.
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
		webhookEvent := webhooks.PreviousHistoryOfflineAsset
		if useNextImage {
			webhookEvent = webhooks.NextHistoryOfflineAsset
		}

		return historyAssetOffline(c, requestData, wantedAssets, requestConfig, com, webhookEvent)
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

		g.Go(getHistoryAsset(requestConfig, com, requestID, deviceID, selectedUser, &viewData, prevAssetsID, currentAssetID))
	}

	// Wait for all goroutines to complete and check for errors
	errGroupWait := g.Wait()
	if errGroupWait != nil {
		t := i18n.T()
		return RenderError(c, errGroupWait, t("processing_images"), requestConfig.Duration)
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

// getHistoryAsset returns a function that processes a single asset from the navigation history.
// It fetches asset metadata, album details (if configured), and generates regular and blurred preview images.
// The function is intended to be run as a goroutine (via errgroup) for each asset in the history entry.
func getHistoryAsset(requestConfig config.Config, com *common.Common, requestID, deviceID, selectedUser string, viewData *common.ViewData, prevAssetsID int, currentAssetID string) func() error {
	return func() error {
		requestConfig.SelectedUser = selectedUser

		asset := immich.New(com.Context(), requestConfig)
		asset.ID = currentAssetID
		if requestConfig.Memories {
			if ok, memory, assetIndex := asset.IsMemory(); ok {
				asset.Bucket = kiosk.SourceMemories
				asset.MemoryTitle = humanize.Time(memory.Assets[assetIndex].LocalDateTime)
			}
		}

		var wg sync.WaitGroup
		wg.Add(1)

		// Fetch asset info and album details in a goroutine.
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
		var dominantColor color.RGBA
		var err error

		// Populate the viewData.Assets entry for this asset after processing.
		defer func() {
			viewData.Assets[prevAssetsID] = common.ViewImageData{
				ImmichAsset:        asset,
				ImageData:          imgString,
				ImageBlurData:      imgBlurString,
				ImageDominantColor: dominantColor,
				User:               selectedUser,
			}
		}()

		// Image processing isn't required for video, audio, or other types.
		// If preview fails for an image, return error; for other types, proceed.
		imgBytes, _, previewErr := asset.ImagePreview()
		if previewErr != nil {
			switch asset.Type {
			case immich.ImageType:
				return fmt.Errorf("retrieving asset: %w", previewErr)

			case immich.VideoType, immich.AudioType, immich.OtherType:
				wg.Wait()
				return nil
			}
		}

		img, byteErr := utils.BytesToImage(imgBytes, requestConfig.UseOriginalImage)
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

		if requestConfig.Theme == kiosk.ThemeBubble {
			dominantColor, err = utils.ExtractDominantColor(img)
			if err != nil {
				return fmt.Errorf("extracting dominant colour: %w", err)
			}
		}

		wg.Wait()

		return nil
	}
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
func historyAssetOffline(c echo.Context, requestData *common.RouteRequestData, wantedAssets []string, requestConfig config.Config, com *common.Common, webhookEvent webhooks.WebhookEvent) error {
	replacer := strings.NewReplacer(
		kiosk.HistoryIndicator, "",
		":", "",
		",", "",
	)

	var sb strings.Builder
	for _, wa := range wantedAssets {
		sb.WriteString(replacer.Replace(wa))
	}

	filename := sb.String()

	filename = generateCacheFilename(filename)

	filename = path.Join(OfflineAssetsPath, filename)

	viewData, loadMsgpackErr := loadMsgpackZstd(filename)
	if loadMsgpackErr != nil {
		log.Error("OfflineMode: loadMsgpackZstd", "picked", filename, "err", loadMsgpackErr)
		return loadMsgpackErr
	}

	viewData.KioskVersion = KioskVersion
	viewData.RequestID = requestData.RequestID
	viewData.DeviceID = requestData.DeviceID
	viewData.History = requestConfig.History
	viewData.Theme = requestConfig.Theme
	viewData.Kiosk.DemoMode = requestConfig.Kiosk.DemoMode

	go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhookEvent, viewData)

	return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))
}
