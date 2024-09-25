package routes

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// NewImage returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
//
// Parameters:
//   - baseConfig: A pointer to the base configuration.
//
// Returns:
//   - echo.HandlerFunc: A function that handles the image request.
func NewImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceID := c.Request().Header.Get("kiosk-device-id")
		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.String(http.StatusTemporaryRedirect, "")
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"deviceID", kioskDeviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			cacheKey := c.Request().URL.String() + kioskDeviceID
			if data, found := pageDataCache.Get(cacheKey); found {
				cachedPageData := data.([]views.PageData)
				if len(cachedPageData) >= 2 {
					log.Debug(
						requestId,
						"deviceID", kioskDeviceID,
						"cache hit for new image", true,
					)
					switch requestConfig.SplitView {
					case true:
						nextPageData := cachedPageData[:2]
						pageDataCache.Set(cacheKey, cachedPageData[2:], cache.DefaultExpiration)
						go imagePreFetch(2, requestConfig, c, kioskDeviceID)

						// Update history which will be outdated in cache
						trimHistory(&requestConfig.History, 10)
						nextPageData[0].History = requestConfig.History

						return Render(c, http.StatusOK, views.Image(nextPageData...))
					default:
						nextPageData := cachedPageData[0]
						pageDataCache.Set(cacheKey, cachedPageData[1:], cache.DefaultExpiration)
						go imagePreFetch(1, requestConfig, c, kioskDeviceID)

						// Update history which will be outdated in cache
						trimHistory(&requestConfig.History, 10)
						nextPageData.History = requestConfig.History

						return Render(c, http.StatusOK, views.Image(nextPageData))
					}
				} else {
					pageDataCache.Delete(cacheKey)
				}
			}
			log.Debug(
				requestId,
				"deviceID", kioskDeviceID,
				"cache miss for new image", false,
			)
		}

		imagesNeeded := 1
		if requestConfig.SplitView {
			imagesNeeded = 2
		}

		pageData := make([]views.PageData, imagesNeeded)

		for index := range imagesNeeded {
			pageDataSingle, err := processPageData(requestConfig, c, false)
			if err != nil {
				log.Error("processing image", "err", err)
				return Render(c, http.StatusOK, views.Error(views.ErrorData{
					Title:   "Error processing image",
					Message: err.Error(),
				}))
			}
			pageData[index] = pageDataSingle
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(2, requestConfig, c, kioskDeviceID)
		}

		return Render(c, http.StatusOK, views.Image(pageData...))

	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
//
// Parameters:
//   - baseConfig: A pointer to the base configuration.
//
// Returns:
//   - echo.HandlerFunc: A function that handles the raw image request.
func NewRawImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichImage := immich.NewImage(requestConfig)

		imgBytes, err := processImage(&immichImage, requestConfig, requestId, "", false)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}
}
