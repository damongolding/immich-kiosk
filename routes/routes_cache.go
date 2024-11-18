package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/webhooks"
	"github.com/labstack/echo/v4"
)

func FlushCache(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		kioskDeviceID := c.Request().Header.Get("kiosk-device-id")

		viewDataCacheMutex.Lock()
		defer viewDataCacheMutex.Unlock()

		log.Info("Cache before flush", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

		ViewDataCache.Flush()
		immich.FluchApiCache()

		log.Info("Cache after flush ", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

		c.Response().Header().Set("HX-Refresh", "true")
		go webhooks.Trigger(*baseConfig, KioskVersion, webhooks.CacheFlush, views.ViewData{DeviceID: kioskDeviceID})
		return c.NoContent(http.StatusNoContent)
	}
}
