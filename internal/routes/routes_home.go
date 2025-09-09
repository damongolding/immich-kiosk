package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/weather"
)

// Home returns an HTTP handler for the home endpoint, initializing request data, applying custom CSS if available, and rendering the home view with device identification and configuration context.
func Home(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {

		c.SetCookie(&http.Cookie{
			Name:   redirectCountHeader,
			MaxAge: -1,
		})

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		var customCSS []byte

		customCSS, err = loadCustomCSS()
		if err != nil {
			log.Error("loading custom css", "err", err)
		}

		queryParams := c.QueryParams()
		if !queryParams.Has("weather") && requestConfig.HasWeatherDefault {
			queryParams.Set("weather", weather.DefaultLocation())
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     generateDeviceID(c),
			Queries:      queryParams,
			CustomCSS:    customCSS,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.Home(viewData, com.Secret()))
	}
}

// generateDeviceID generates a stable device identifier based on request-specific information.
//
// It uses the "kiosk-device-id" header if present, otherwise falls back to a combination of
// the client's IP address and User-Agent. The identifier also incorporates normalized query
// parameters. The resulting string is hashed using SHA-256 and returned as a hex string.
func generateDeviceID(c echo.Context) string {
	queryParams := c.QueryParams()
	var parts []string
	for key, values := range queryParams {
		joined := strings.Join(values, ",")
		parts = append(parts, key+"="+joined)
	}
	sort.Strings(parts)
	normalizedQuery := strings.Join(parts, "&")

	deviceTag := c.Request().Header.Get("kiosk-device-id")
	if deviceTag == "" {
		ip := c.RealIP()
		if ip == "" {
			ip = utils.GenerateUUID()
		}
		userAgent := c.Request().UserAgent()
		deviceTag = ip + "|" + userAgent
	}

	idSource := deviceTag + "|" + normalizedQuery

	hash := sha256.Sum256([]byte(idSource))
	deviceID := hex.EncodeToString(hash[:])

	return deviceID
}

func loadCustomCSS() ([]byte, error) {
	if !utils.FileExists("./custom.css") {
		return nil, nil
	}
	return os.ReadFile("./custom.css")
}
