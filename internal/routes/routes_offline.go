package routes

import (
	"errors"
	"math/rand/v2"
	"os"
	"path"
	"sync/atomic"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
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

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestID := requestData.RequestID
		deviceID := requestData.DeviceID

		if _, err := os.Stat(OfflineAssetsPath); os.IsNotExist(err) {
			err := os.MkdirAll(OfflineAssetsPath, os.ModePerm)
			if err != nil {
				log.Error("OfflineMode", "err", err)
				return err
			}
		}

		files, ReadDirErr := os.ReadDir(OfflineAssetsPath)
		if ReadDirErr != nil {
			log.Error("OfflineMode", "err", ReadDirErr)
			return ReadDirErr
		}

		if len(files) == 0 {
			DownloadOfflineAssets(baseConfig, c, com, requestID, deviceID)
		}

		var nonDotFiles []string
		for _, file := range files {
			if !file.IsDir() && file.Name()[0] != '.' {
				nonDotFiles = append(nonDotFiles, file.Name())
			}
		}

		picked := nonDotFiles[rand.IntN(len(nonDotFiles))]

		picked = path.Join(OfflineAssetsPath, picked)

		return c.File(picked)
	}
}

func DownloadOfflineAssets(baseConfig *config.Config, echoCtx echo.Context, com *common.Common, requestID, deviceID string) error {

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
				viewData, err := generateViewData(*baseConfig, requestCtx, requestID, deviceID, false)
				if err != nil {
					return err
				}

				filename := path.Join(OfflineAssetsPath, viewData.Assets[0].ImmichAsset.ID+".html")

				if _, err := os.Stat(filename); os.IsExist(err) {
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
