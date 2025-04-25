package routes

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/labstack/echo/v4"
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

		replacer := strings.NewReplacer(
			kiosk.HistoryIndicator, "",
			":", "",
			",", "",
		)
		for i, h := range requestConfig.History {
			requestConfig.History[i] = replacer.Replace(h)
		}

		if _, err = os.Stat(OfflineAssetsPath); os.IsNotExist(err) {
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

		if len(files) == 0 {
			downloadErr := DownloadOfflineAssets(requestConfig, c, com, requestID, deviceID)
			if downloadErr != nil {
				log.Error("OfflineMode: DownloadOfflineAssets", "err", downloadErr)
				return downloadErr
			}
		}

		var nonDotFiles []string
		for _, file := range files {
			if !file.IsDir() && file.Name()[0] != '.' {
				nonDotFiles = append(nonDotFiles, file.Name())
			}
		}

		if len(nonDotFiles) == 0 {
			return c.String(http.StatusNotFound, "No offline assets found")
		}

		for range 3 {

			picked := nonDotFiles[rand.IntN(len(nonDotFiles))]

			if slices.Contains(requestConfig.History, strings.Replace(picked, ".html", "", 1)) {
				log.Info("Same file!", "file", picked, "history", requestConfig.History)
				continue
			}

			log.Info("Picked", "file", picked, "history", requestConfig.History)

			picked = path.Join(OfflineAssetsPath, picked)

			return c.File(picked)
		}

		return c.String(http.StatusNotFound, "No offline assets found")
	}
}

func DownloadOfflineAssets(baseConfig config.Config, echoCtx echo.Context, com *common.Common, requestID, deviceID string) error {

	if !mu.TryLock() {
		return errors.New("DownloadOfflineAssets is already running")
	}
	defer mu.Unlock()

	parallelDownloads := baseConfig.Kiosk.ExperimentalOfflineMode.ParallelDownloads
	numberOfAssets := baseConfig.Kiosk.ExperimentalOfflineMode.NumberOfAssets
	maxSize := baseConfig.Kiosk.ExperimentalOfflineMode.MaxSize

	var eg errgroup.Group
	eg.SetLimit(parallelDownloads)

	var offlineSize atomic.Int64

	requestCtx := common.CopyContext(echoCtx)

	for i := range numberOfAssets {
		eg.Go(func() error {

			defer log.Info("Done", "#", i)

			for range 3 {
				viewData, err := generateViewData(baseConfig, requestCtx, requestID, deviceID, false)
				if err != nil {
					return err
				}

				var filename string

				for _, asset := range viewData.Assets {
					filename += asset.ImmichAsset.ID + asset.User
				}

				filename = path.Join(OfflineAssetsPath, filename+".html")

				if _, err = os.Stat(filename); os.IsExist(err) {
					continue
				}

				return SaveOfflineAsset(com.Context(), filename, imageComponent.Image(viewData, com.Secret()), maxSize, &offlineSize)
			}

			return errors.New("DownloadOfflineAssets: max tries reached")

		})
	}

	err := eg.Wait()
	if err != nil {
		log.Error("DownloadOfflineAssets", "err", err)
		return err
	}

	return nil
}
