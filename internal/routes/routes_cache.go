package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/labstack/echo/v4"
)

func FlushCache(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		viewDataCacheMutex.Lock()
		defer viewDataCacheMutex.Unlock()

		log.Info("Cache before flush", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

		ViewDataCache.Flush()
		immich.FlushApiCache()

		log.Info("Cache after flush ", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

		c.Response().Header().Set("HX-Refresh", "true")
		go webhooks.Trigger(requestData, KioskVersion, webhooks.CacheFlush, common.ViewData{})
		return c.NoContent(http.StatusNoContent)
	}
}
