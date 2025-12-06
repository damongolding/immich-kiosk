package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/labstack/echo/v4"
)

// Albums returns an Echo handler that fetches and returns all albums from Immich.
func Albums(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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
		requestConfig := requestData.RequestConfig

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
		)

		// Check if API albums are password protected
		if requestConfig.Kiosk.APIAlbumsPassword != "" {
			providedPassword := c.Request().Header.Get("X-Kiosk-Password")
			if providedPassword == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Password required", "code": "password_required"})
			}
			if providedPassword != requestConfig.Kiosk.APIAlbumsPassword {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid password", "code": "invalid_password"})
			}
		}

		// Initialize response map
		response := make(map[string]immich.Albums)

		// Fetch main library albums
		asset := immich.New(com.Context(), requestConfig)
		mainAlbums, err := asset.AllAlbums(requestID, deviceID)
		if err != nil {
			log.Error("Failed to fetch albums for main library", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch albums"})
		}
		response["Main Library"] = mainAlbums

		// Fetch albums for other users
		for user, apiKey := range requestConfig.ImmichUsersAPIKeys {
			// Create a copy of config with user's API key
			userConfig := requestConfig
			userConfig.ImmichAPIKey = apiKey
			userConfig.SelectedUser = user

			userAsset := immich.New(com.Context(), userConfig)
			userAlbums, err := userAsset.AllAlbums(requestID, deviceID)
			if err != nil {
				log.Error("Failed to fetch albums for user", "user", user, "error", err)
				// Continue fetching for other users even if one fails
				continue
			}
			response[user] = userAlbums
		}

		return c.JSON(http.StatusOK, response)
	}
}
