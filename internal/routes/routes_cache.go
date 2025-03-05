package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/labstack/echo/v4"
)

func FlushCache(baseConfig *config.Config, com common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		log.Info("Cache before flush", "cache_items", cache.ItemCount())

		cache.Flush()

		log.Info("Cache after flush ", "cache_items", cache.ItemCount())

		c.Response().Header().Set("HX-Refresh", "true")
		go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.CacheFlush, common.ViewData{})
		return c.NoContent(http.StatusNoContent)
	}
}
