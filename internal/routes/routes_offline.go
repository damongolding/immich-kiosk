package routes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/zstd"
	"github.com/labstack/echo/v4"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/errgroup"
)

const OfflineAssetsPath = "./offline-assets"

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
		requestConfig.ExperimentalAlbumVideo = false

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
			err = os.MkdirAll(OfflineAssetsPath, os.ModePerm)
			if err != nil {
				log.Error("OfflineMode", "err", err)
				return err
			}
		}

		files, readDirErr := os.ReadDir(OfflineAssetsPath)
		if readDirErr != nil {
			log.Error("OfflineMode", "err", readDirErr)
			return readDirErr
		}

		var nonDotFiles []string
		for _, file := range files {
			if !file.IsDir() && file.Name()[0] != '.' {
				nonDotFiles = append(nonDotFiles, file.Name())
			}
		}

		if len(nonDotFiles) == 0 {
			requestCtx := common.CopyContext(c)
			go func(c common.ContextCopy) {
				downloadErr := downloadOfflineAssets(requestConfig, c, com, requestID, deviceID)
				if downloadErr != nil {
					log.Error("OfflineMode: DownloadOfflineAssets", "err", downloadErr)
				}
			}(requestCtx)

			return Render(c, http.StatusOK, partials.Message(partials.MessageData{
				Title:         "Downloading Assets",
				Message:       fmt.Sprintf("Getting %v assets with a storage max capacity of %s", requestConfig.Kiosk.ExperimentalOfflineMode.NumberOfAssets, requestConfig.Kiosk.ExperimentalOfflineMode.MaxSize),
				IsDownloading: true,
			}))
		}

		for range 3 {

			if len(nonDotFiles) == 0 {
				continue
			}

			picked := nonDotFiles[rand.IntN(len(nonDotFiles))]

			// check if file has already been picked (in history)
			if slices.Contains(historyAsFilenames, picked) {
				continue
			}

			picked = path.Join(OfflineAssetsPath, picked)

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

			return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))

		}

		return Render(c, http.StatusOK, partials.Error(partials.ErrorData{
			Message: "No offline assets found",
		}))

	}
}

func downloadOfflineAssets(requestConfig config.Config, requestCtx common.ContextCopy, com *common.Common, requestID, deviceID string) error {

	if !mu.TryLock() {
		return errors.New("DownloadOfflineAssets is already running")
	}
	defer mu.Unlock()

	parallelDownloads := requestConfig.Kiosk.ExperimentalOfflineMode.ParallelDownloads
	numberOfAssets := requestConfig.Kiosk.ExperimentalOfflineMode.NumberOfAssets
	maxSize, maxSizeErr := utils.ParseSize(requestConfig.Kiosk.ExperimentalOfflineMode.MaxSize)
	if maxSizeErr != nil {
		return maxSizeErr
	}

	eg, _ := errgroup.WithContext(com.Context())
	eg.SetLimit(min(parallelDownloads, numberOfAssets))

	var offlineSize atomic.Int64
	var createdFiles sync.Map
	var maxReached atomic.Bool

	startTime := time.Now()

	for range numberOfAssets {
		eg.Go(func() error {

			for range 3 {

				sizeSoFar := offlineSize.Load()
				if maxSize != 0 && sizeSoFar >= maxSize {
					if !maxReached.Load() {
						maxReached.Store(true)
						humanOfflineSize := humanize.Bytes(uint64(offlineSize.Load()))
						humanMaxSize := humanize.Bytes(uint64(maxSize))
						log.Warn("SaveOfflineAsset: max offline storage size reached",
							"total assets saved", humanOfflineSize,
							"maxOfflineSize", humanMaxSize,
						)
					}
					return nil
				}

				viewData, err := generateViewData(requestConfig, requestCtx, requestID, deviceID, false)
				if err != nil {
					log.Error("SaveOfflineAsset: generateViewData", "err", err)
					continue
				}

				viewData.UseOfflineMode = true
				viewData.History = []string{}

				var filename string

				for _, asset := range viewData.Assets {
					filename += asset.ImmichAsset.ID + asset.User
				}

				filename = generateCacheFilename(filename)

				filename = path.Join(OfflineAssetsPath, filename)

				if _, exists := createdFiles.Load(filename); exists {
					continue
				}

				if _, err = os.Stat(filename); err == nil {
					continue
				}

				createdFiles.Store(filename, true)

				return saveMsgpackZstd(filename, viewData, &offlineSize, maxSize, &maxReached, &createdFiles)
			}

			return errors.New("DownloadOfflineAssets: max tries reached")
		})
	}

	err := eg.Wait()
	if err != nil {
		log.Error("DownloadOfflineAssets finished with", "err", err)
		return err
	}

	duration := time.Since(startTime).Seconds()

	size := 0
	createdFiles.Range(func(_, _ any) bool {
		size++
		return true
	})

	log.Info(fmt.Sprintf("%v offline assets downloaded", size), "in", fmt.Sprintf("%.2f seconds", duration))

	return nil
}

func saveMsgpackZstd(filename string, data common.ViewData, offlineSize *atomic.Int64, maxSize int64, maxReached *atomic.Bool, createdFiles *sync.Map) error {

	defer func() {
		createdFiles.Delete(filename)
	}()

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return err
	}

	sizeSoFar := offlineSize.Load()
	if maxSize != 0 && sizeSoFar+int64(buf.Len()) >= maxSize {
		if !maxReached.Load() {
			maxReached.Store(true)
			humanOfflineSize := humanize.Bytes(uint64(offlineSize.Load()))
			humanMaxSize := humanize.Bytes(uint64(maxSize))
			log.Warn("SaveOfflineAsset: max offline storage size reached", "total assets saved", humanOfflineSize, "maxOfflineSize", humanMaxSize)
		}
		return nil
	}

	tmp, err := os.CreateTemp(path.Dir(filename), ".offline-*")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		if tmp != nil {
			removeErr := os.Remove(tmp.Name())
			if removeErr != nil {
				log.Error("SaveOfflineAsset: failed to remove temporary file", "error", removeErr)
			}
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

	fi, statErr := tmp.Stat()
	if statErr != nil {
		return statErr
	}

	if newTotal := offlineSize.Add(fi.Size()); maxSize != 0 && newTotal > maxSize {
		offlineSize.Add(-fi.Size())
		return nil
	}

	if err = os.Rename(tmp.Name(), filename); err != nil {
		offlineSize.Add(-fi.Size())
		return err
	}

	// prevent deferred removal
	tmp = nil
	filename = ""

	return nil
}

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

func generateCacheFilename(uuids ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(uuids, "")))
	return hex.EncodeToString(hash[:])
}
