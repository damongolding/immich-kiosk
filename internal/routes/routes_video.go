package routes

import (
	"net/http"
	"net/url"
	"path"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/labstack/echo/v4"
)

// func GetVideo(baseConfig *config.Config) echo.HandlerFunc {
// 	return func(c echo.Context) error {

// 		requestConfig := *baseConfig

// 		immichImage := immich.NewImage(requestConfig)

// 		immichImage.ID = "429d336e-48f3-4d2b-8b63-7b0a04758b48"

// 		b := immichImage.VideoPlayback()

//			return c.Blob(http.StatusOK, "video/quicktime", b)
//		}
//	}

func GetVideo(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		u, err := url.Parse(baseConfig.ImmichUrl)
		if err != nil {
			log.Error(err)
			return err
		}

		apiUrl := url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
			Path:   path.Join("api", "assets", "429d336e-48f3-4d2b-8b63-7b0a04758b48", "video", "playback"),
		}

		// Create HTTP client
		client := &http.Client{}

		// Create request to your API
		req, err := http.NewRequest("GET", apiUrl.String(), nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error creating request")
		}

		req.Header.Set("Accept", "application/octet-stream")
		req.Header.Set("x-api-key", baseConfig.ImmichApiKey)

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error fetching video")
		}
		defer resp.Body.Close()

		// Set response headers
		c.Response().Header().Set("Content-Type", "video/quicktime") // Adjust content type as needed
		c.Response().Header().Set("Content-Length", resp.Header.Get("Content-Length"))

		// Stream the video
		return c.Stream(http.StatusOK, "video/quicktime", resp.Body)
	}
}

func Video(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		return c.HTML(http.StatusOK, `
			<!doctype html>
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
			<video id="myVideo" autoplay playsinline muted src="/getvideo"></video>
			<script>
			const video = document.getElementById('myVideo');

video.addEventListener('ended', function() {
    console.log('Video has ended');
    // Your code here
});
			</script>
		</body>
		</html>

	`)
	}
}
