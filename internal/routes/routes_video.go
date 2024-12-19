package routes

import (
	"net/http"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/labstack/echo/v4"
)

func GetVideo(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestConfig := *baseConfig

		immichImage := immich.NewImage(requestConfig)

		immichImage.ID = "429d336e-48f3-4d2b-8b63-7b0a04758b48"

		b := immichImage.VideoPlayback()

		return c.Blob(http.StatusOK, "video/quicktime", b)
	}
}

func Video(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		return c.HTML(http.StatusOK, `
		<html>
		<head>
		<style>
		video{
			max-width: 100%;
			max-height: 100%;
		}
		</style>
		</head>
		<body>
			<video id="myVideo" autoplay playsinline  src="/getvideo" style="touch-action: none; />
		</body>
		</html>

	`)
	}
}
