package routes

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"text/template"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v5"
)

func ServiceWorker(baseConfig *config.Config, public embed.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/javascript")
		c.Response().Header().Set("Service-Worker-Allowed", "/")
		c.Response().Header().Set("Cache-Control", "no-cache")

		var sw, css []byte
		var err error

		sw, err = public.ReadFile("frontend/public/assets/js/sw.js.tmpl")
		if err != nil {
			return err
		}

		css, err = public.ReadFile("frontend/public/assets/css/kiosk.css")
		if err != nil {
			return err
		}

		var customCSS []byte
		customCSS, err = utils.LoadCustomCSS()
		if err != nil {
			log.Error("ServiceWorker: loading custom css", "err", err)
		}

		var buf bytes.Buffer

		err = views.Recovering(baseConfig.SystemLang, baseConfig.Kiosk.Version, css, customCSS, baseConfig.CustomCSS).Render(c.Request().Context(), &buf)
		if err != nil {
			return err
		}

		return renderServiceWorker(c.Response(), sw, baseConfig.Kiosk.Version, buf.String())
	}
}

func renderServiceWorker(w io.Writer, sw []byte, version, fallbackHTML string) error {
	cacheName, err := json.Marshal("immich-kiosk-" + version)
	if err != nil {
		return err
	}

	fallbackJSON, err := json.Marshal(fallbackHTML)
	if err != nil {
		return err
	}

	swTmpl, err := template.New("sw").Parse(string(sw))
	if err != nil {
		return err
	}

	return swTmpl.Execute(w, map[string]string{
		"CacheName":    string(cacheName),
		"FallbackHtml": string(fallbackJSON),
	})
}
