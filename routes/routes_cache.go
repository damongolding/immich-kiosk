package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/labstack/echo/v4"
)

func FlushCache(c echo.Context) error {

	viewDataCacheMutex.Lock()
	defer viewDataCacheMutex.Unlock()

	log.Info("Cache before flush", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

	ViewDataCache.Flush()
	immich.FluchApiCache()

	log.Info("Cache after flush ", "viewDataCache_items", ViewDataCache.ItemCount(), "apiCache_items", immich.ApiCacheCount())

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusNoContent)

}
