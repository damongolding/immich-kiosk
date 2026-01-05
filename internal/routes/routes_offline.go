package routes

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/zstd"
	"github.com/labstack/echo/v4"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/errgroup"
)

const (
	OfflineAssetsPath         = "./offline-assets"
	OfflineExpirationFilename = "_expiration"
)

var ErrMaxStorageReached = errors.New("max offline storage size reached")

// OfflineMode returns an HTTP handler that serves offline assets for the kiosk application.
// It attempts to load and render a cached offline asset, handling asset expiration, cache directory creation, and duplicate removal.
// If no valid offline assets are available or assets have expired, it initiates an asynchronous download and displays a status page.
func OfflineMode(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestID := requestData.RequestID
		deviceID := requestData.DeviceID

		requestConfig := *baseConfig
		requestConfig.History = requestData.RequestConfig.History
		requestConfig.Memories = false
		requestConfig.ShowVideos = false
		requestConfig.Theme = requestData.RequestConfig.Theme

		if len(requestConfig.History) > 1 && !strings.HasPrefix(requestConfig.History[len(requestConfig.History)-1], "*") {
			return NextHistoryAsset(baseConfig, com, c)
		}

		replacer := strings.NewReplacer(
			kiosk.HistoryIndicator, "",
			":", "",
			",", "",
		)
		historyAsFilenames := make([]string, len(requestConfig.History))
		for i, h := range requestConfig.History {
			historyAsFilenames[i] = generateCacheFilename(replacer.Replace(h))
		}

		if _, err = os.Stat(OfflineAssetsPath); os.IsNotExist(err) {
			log.Warn("creating offline assets directory - NOTE: If running in Docker, this data will not persist between container restarts")
			err = os.MkdirAll(OfflineAssetsPath, 0755)
			if err != nil {
				log.Error("OfflineMode", "err", err)
				return err
			}
		}

		files, readDirErr := os.ReadDir(OfflineAssetsPath)
		if readDirErr != nil {
			log.Error("OfflineMode: ReadDir", "err", readDirErr)
			return readDirErr
		}

		var nonDotFiles []string
		for _, file := range files {
			if !file.IsDir() && file.Name()[0] != '.' && file.Name()[0] != '_' {
				nonDotFiles = append(nonDotFiles, file.Name())
			}
		}

		if len(nonDotFiles) == 0 {
			return handleNoOfflineAssets(c, requestConfig, com, requestID, deviceID)
		}

		if requestConfig.OfflineMode.ExpirationHours > 0 {
			expired, expiredErr := checkOfflineAssetsExpiration(com.Context(), requestConfig.ImmichURL)
			if expiredErr != nil {
				return expiredErr
			}
			if expired {
				return handleNoOfflineAssets(c, requestConfig, com, requestID, deviceID)
			}
		}

		// check for duplicates if we have more assets then the history limit
		if len(nonDotFiles) > kiosk.HistoryLimit {
			utils.RemoveDuplicatesInPlace(&nonDotFiles, historyAsFilenames)
		}

		for range 3 {

			if len(nonDotFiles) == 0 {
				continue
			}

			picked := nonDotFiles[rand.IntN(len(nonDotFiles))]

			picked = filepath.Join(OfflineAssetsPath, picked)

			viewData, loadMsgpackErr := loadMsgpackZstd(picked)
			if loadMsgpackErr != nil {
				log.Error("OfflineMode: loadMsgpackZstd", "picked", picked, "err", loadMsgpackErr)
				continue
			}

			viewData.KioskVersion = KioskVersion
			viewData.RequestID = requestID
			viewData.DeviceID = deviceID
			utils.TrimHistory(&requestConfig.History, kiosk.HistoryLimit)
			viewData.History = requestConfig.History
			viewData.Theme = requestConfig.Theme
			viewData.Kiosk.DemoMode = requestConfig.Kiosk.DemoMode

			go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.NewOfflineAsset, viewData)

			return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))
		}

		return Render(c, http.StatusOK, partials.Error(partials.ErrorData{
			Title:   "No offline assets found",
			Message: "Check Kiosk logs for more information",
		}))

	}
}

// downloadOfflineAssets downloads and caches assets for offline mode viewing.
// It manages parallel downloads, enforces size limits, and handles expiration.
//
// Parameters:
//   - requestConfig: Configuration for the download request including parallel downloads,
//     number of assets, and max size limits
//   - requestCtx: Copy of the request context for async operations
//   - com: Common context and utilities
//   - requestID: Unique ID for this request
//   - deviceID: ID of the requesting device
//
// Returns an error if:
//   - Another download is already in progress
//   - Max size parsing fails
//   - File operations fail
//   - Asset downloads fail
func downloadOfflineAssets(requestConfig config.Config, requestCtx common.ContextCopy, com *common.Common, requestID, deviceID string) error {
	if !mu.TryLock() {
		log.Debug("DownloadOfflineAssets is already running")
		return nil
	}
	defer mu.Unlock()

	requestConfig.UseOfflineMode = true
	parallelDownloads := requestConfig.OfflineMode.ParallelDownloads
	numberOfAssets := requestConfig.OfflineMode.NumberOfAssets
	maxSize, maxSizeErr := utils.ParseSize(requestConfig.OfflineMode.MaxSize)
	if maxSizeErr != nil {
		return maxSizeErr
	}

	// Setup parent context with cancel
	ctx, cancel := context.WithCancel(com.Context())
	defer cancel()

	// Wrap with errgroup context
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(min(parallelDownloads, numberOfAssets))

	var offlineSize atomic.Int64
	var createdFiles sync.Map
	var once sync.Once

	fileExistsTolerance := calcFileExistsTolerance(numberOfAssets)
	var fileExistsCount atomic.Int64
	var errorCount atomic.Int64

	startTime := time.Now()

	// Write expiration timestamp
	expiration, expirationErr := os.Create(filepath.Join(OfflineAssetsPath, OfflineExpirationFilename))
	if expirationErr != nil {
		return expirationErr
	}
	defer expiration.Close()

	_, expirationErr = expiration.WriteString(
		startTime.Add(time.Hour * time.Duration(requestConfig.OfflineMode.ExpirationHours)).Format(time.RFC3339),
	)
	if expirationErr != nil {
		return expirationErr
	}

	for range numberOfAssets {
		eg.Go(func() error {

			for range 3 {
				if err := checkCanceled(egCtx); err != nil {
					return err
				}

				viewData, err := generateViewData(requestConfig, requestCtx, requestID, deviceID, false)
				if err != nil {
					log.Error("SaveOfflineAsset: generateViewData", "err", err)
					if errorCount.Add(1) > fileExistsTolerance*2 {
						once.Do(func() {
							log.Info("Too many errors — cancelling download",
								"error count", errorCount.Load(),
								"tolerance", fileExistsTolerance*2)
							cancel()
						})
						return nil
					}
					continue
				}

				viewData.UseOfflineMode = true
				viewData.History = []string{}

				var sb strings.Builder
				for _, asset := range viewData.Assets {
					sb.WriteString(asset.ImmichAsset.ID)
					sb.WriteString(asset.User)
				}

				filename := sb.String()
				filename = filepath.Join(OfflineAssetsPath, generateCacheFilename(filename))

				if _, exists := createdFiles.Load(filename); exists {
					if fileExistsCount.Add(1) > fileExistsTolerance {
						once.Do(func() {
							log.Info("Too many duplicate assets — cancelling download",
								"existingCount", fileExistsCount.Load(),
								"tolerance", fileExistsTolerance)
							cancel()
						})
						return nil
					}
					continue
				}
				if _, statErr := os.Stat(filename); statErr == nil {
					if fileExistsCount.Add(1) > fileExistsTolerance {
						once.Do(func() {
							log.Info("Too many duplicate assets — cancelling download",
								"existingCount", fileExistsCount.Load(),
								"tolerance", fileExistsTolerance)
							cancel()
						})
						return nil
					}
					continue
				}

				createdFiles.Store(filename, true)

				err = saveMsgpackZstd(egCtx, filename, viewData, &offlineSize, maxSize, &createdFiles)
				if errors.Is(err, ErrMaxStorageReached) {
					once.Do(func() {
						humanOfflineSize := humanize.Bytes(uint64(offlineSize.Load()))
						humanMaxSize := humanize.Bytes(uint64(maxSize))
						log.Info("Max offline storage size reached",
							"total assets saved", humanOfflineSize,
							"maxOfflineSize", humanMaxSize,
						)
						cancel()
					})
					return nil
				}
				if err != nil {
					log.Error("SaveOfflineAsset: saveMsgpackZstd", "err", err)
					continue
				}
				return nil
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("DownloadOfflineAssets finished with", "err", err)
		return err
	}

	duration := time.Since(startTime).Seconds()
	finishTime := fmt.Sprintf("%.2f seconds", duration)

	if duration > 60 {
		finishTime = fmt.Sprintf("%.2f minutes", duration/60)
	}

	size := 0
	createdFiles.Range(func(_, _ any) bool {
		size++
		return true
	})

	log.Info(fmt.Sprintf("%v offline assets downloaded", size), "in", finishTime)

	return nil
}

// saveMsgpackZstd saves view data to a file using msgpack encoding and zstd compression.
// It manages file size limits and updates offline storage size tracking.
// Returns an error if encoding, compression or file operations fail.
// Returns ErrMaxStorageReached if adding the file would exceed the configured max size.
func saveMsgpackZstd(ctx context.Context, filename string, data common.ViewData, offlineSize *atomic.Int64, maxSize int64, createdFiles *sync.Map) error {

	defer func() {
		createdFiles.Delete(filename)
	}()

	if cancelledErr := checkCanceled(ctx); cancelledErr != nil {
		return cancelledErr
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return err
	}

	if cancelledErr := checkCanceled(ctx); cancelledErr != nil {
		return cancelledErr
	}

	tmp, err := os.CreateTemp(path.Dir(filename), ".offline-*")
	if err != nil {
		return err
	}
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()

	compressionLevel := zstd.SpeedFastest
	if buf.Len() > 1024*1024 {
		compressionLevel = zstd.SpeedBestCompression
	}

	encoder, err := zstd.NewWriter(tmp, zstd.WithEncoderLevel(compressionLevel))
	if err != nil {
		return err
	}
	defer encoder.Close()

	if _, err = encoder.Write(buf.Bytes()); err != nil {
		return err
	}

	if err = encoder.Flush(); err != nil {
		return err
	}

	if cancelledErr := checkCanceled(ctx); cancelledErr != nil {
		return cancelledErr
	}

	fi, statErr := tmp.Stat()
	if statErr != nil {
		return statErr
	}

	newTotal := offlineSize.Add(fi.Size())
	if maxSize != 0 && newTotal > maxSize {
		offlineSize.Add(-fi.Size())
		return ErrMaxStorageReached
	}

	if err = os.Rename(tmp.Name(), filename); err != nil {
		offlineSize.Add(-fi.Size())
		return err
	}

	tmp = nil
	filename = ""

	return nil
}

// loadMsgpackZstd loads and decodes a msgpack+zstd compressed file into ViewData.
// Returns the decoded ViewData and any error encountered during the process.
func loadMsgpackZstd(filename string) (common.ViewData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return common.ViewData{}, err
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		return common.ViewData{}, err
	}
	defer decoder.Close()

	data, err := io.ReadAll(decoder)
	if err != nil {
		return common.ViewData{}, err
	}

	var viewData common.ViewData
	buf := bytes.NewBuffer(data)
	dec := msgpack.NewDecoder(buf)
	err = dec.Decode(&viewData)
	return viewData, err
}

// generateCacheFilename creates a SHA-256 hash from the concatenated UUIDs
// and returns it as a hex-encoded string to be used as a filename.
func generateCacheFilename(uuids ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(uuids, "")))
	return hex.EncodeToString(hash[:])
}

// handleNoOfflineAssets handles the case when no offline assets are available.
// It initiates an asynchronous download of assets and displays a status message to the user.
//
// Parameters:
//   - c: The Echo context
//   - requestConfig: Configuration for the request including offline mode settings
//   - com: Common context and utilities
//   - requestID: Unique ID for this request
//   - deviceID: ID of the requesting device
//
// The function:
//  1. Starts asynchronous download of offline assets
//  2. Formats user-friendly messages about storage limits and expiration
//  3. Renders a status page with download information
//
// Returns an error if the render operation fails
func handleNoOfflineAssets(c echo.Context, requestConfig config.Config, com *common.Common, requestID, deviceID string) error {
	requestCtx := common.CopyContext(c)
	go func(c common.ContextCopy) {
		downloadErr := downloadOfflineAssets(requestConfig, c, com, requestID, deviceID)
		if downloadErr != nil {
			log.Error("OfflineMode: DownloadOfflineAssets", "err", downloadErr)
		}
	}(requestCtx)

	maxSizeMessage := "None"
	if requestConfig.OfflineMode.MaxSize != "0" {
		maxSizeMessage = strings.ToUpper(requestConfig.OfflineMode.MaxSize)
	}

	expiryMessage := "None"
	hours := requestConfig.OfflineMode.ExpirationHours
	if hours > 0 {
		switch hours {
		case 1:
			expiryMessage = fmt.Sprintf("%d hour", hours)
		default:
			expiryMessage = fmt.Sprintf("%d hours", hours)
		}
	}

	t := i18n.T()

	message := fmt.Sprintf(`
		<ul>
			<li>%s: <strong>%v</strong> %s</li>
			<li>%s: <strong>%s</strong></li>
			<li>%s: <strong>%s</strong></li>
		</ul>
		`,
		t("limit"),
		requestConfig.OfflineMode.NumberOfAssets,
		t("assets"),
		t("storage_capacity"),
		maxSizeMessage,
		t("expiration"),
		expiryMessage,
	)

	return Render(c, http.StatusOK, partials.Message(partials.MessageData{
		Title:         t("downloading_assets"),
		Message:       message,
		IsDownloading: true,
	}))
}

// checkOfflineAssetsExpiration checks if cached offline assets have expired based on
// the expiration time stored in the _expiration file.
//
// It reads the expiration timestamp from the file, parses it, and compares it to the
// current time. If the assets have expired, it cleans up the offline assets directory.
//
// Returns:
//   - bool: true if assets have expired, false otherwise
//   - error: any error encountered during the process
func checkOfflineAssetsExpiration(ctx context.Context, immichURL string) (bool, error) {
	expirationContent, expirationErr := os.ReadFile(filepath.Join(OfflineAssetsPath, OfflineExpirationFilename))
	if expirationErr != nil {
		log.Warn("expiration missing", "err", expirationErr)
		return true, nil
	}

	expirationTime, timeparseErr := time.Parse(time.RFC3339, strings.TrimSpace(string(expirationContent)))
	if timeparseErr != nil {
		log.Error("OfflineMode", "err", timeparseErr)
		return false, timeparseErr
	}

	if time.Now().After(expirationTime) {
		if !immich.IsOnline(ctx, immichURL) {
			log.Warn("Offline assets have expired but Immich is offline")
			return false, nil
		}
		log.Info("Offline assets have expired")
		cleanErr := utils.CleanDirectory(OfflineAssetsPath)
		if cleanErr != nil {
			log.Error("Failed to clean offline assets directory", "err", cleanErr)
		}
		return true, nil
	}

	return false, nil
}

func IsDownloading(c echo.Context) error {
	if IsOfflineDownloadRunning() {
		return c.NoContent(http.StatusOK)
	}

	return Render(c, http.StatusOK, partials.DownloadingStatus(false))
}

func IsOfflineDownloadRunning() bool {
	locked := mu.TryLock()
	if locked {
		mu.Unlock()
	}
	return !locked
}

func checkCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func calcFileExistsTolerance(numAssets int) int64 {
	// Allow up to 10%, but cap at 20 for safety
	const tolerancePercentage = 0.10
	const maxToleranceCap = 20
	const minTolerance = 3

	percent := int(float64(numAssets) * tolerancePercentage)

	if percent > maxToleranceCap {
		return int64(maxToleranceCap)
	}

	return max(minTolerance, int64(percent))
}
