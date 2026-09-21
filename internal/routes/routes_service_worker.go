package routes

import (
	"bytes"
	"embed"
	"html/template"

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

		swTmpl := template.Must(template.New("sw").Parse(string(sw)))

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

		html := buf.String()

		w := c.Response()

		err = swTmpl.Execute(w, map[string]string{
			"Version":      baseConfig.Kiosk.Version,
			"FallbackHtml": html,
		})

		return err
	}
}
