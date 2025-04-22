package routes

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

func demoAssets(requestConfig config.Config, requestCtx common.ContextCopy, com *common.Common) {
	log.Info("getting demo assets")

	errGroup := errgroup.Group{}
	errGroup.SetLimit(10)

	for i := range 20 {
		errGroup.Go(func() error {

			log.Info("getting demo asset", "id", i)

			for range 10 {
				viewData, err := generateViewData(requestConfig, requestCtx, "demo", "demo", false)
				if err != nil {
					log.Error("generateViewData", "err", err)
					return err
				}

				filename := fmt.Sprintf("./demo-assets/%s.html", viewData.Assets[0].ImmichAsset.ID)

				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					continue
				}

				file, fileErr := os.Create(filename)
				if fileErr != nil {
					log.Error("os.Create", "err", fileErr)
					return fileErr
				}
				defer file.Close()

				imageComponent.Image(viewData, com.Secret()).Render(com.Context(), file)

				return nil
			}

			return nil

		})
	}

	errs := errGroup.Wait()
	if errs != nil {
		log.Error("failed to retrieve demo assets", "errs", errs)
	}

	log.Info("retrieved demo assets")
}

func DemoAsset(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {

		if !baseConfig.Kiosk.DemoMode {
			return c.NoContent(http.StatusOK)
		}

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
		deviceID := requestData.DeviceID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"deviceID", deviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		// use history
		if len(requestConfig.History) > 1 && !strings.HasPrefix(requestConfig.History[len(requestConfig.History)-1], "*") {
			return NextHistoryAsset(baseConfig, com, c)
		}

		mkdirErr := os.Mkdir("./demo-assets", 0755)
		if mkdirErr != nil {
			log.Error("os.Mkdir", "err", mkdirErr)
		}

		files := []os.DirEntry{}
		entries, fileErr := os.ReadDir("./demo-assets")
		if fileErr != nil {
			return RenderError(c, fileErr, "reading demo directory")
		}
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				files = append(files, entry)
			}
		}

		requestCtx := common.CopyContext(c)

		if len(files) == 0 {
			demoAssets(requestConfig, requestCtx, com)
		}

		if len(files) > 0 {
			randomIndex := rand.IntN(len(files))
			file, openErr := os.ReadFile(fmt.Sprintf("./demo-assets/%s", files[randomIndex].Name()))
			if openErr != nil {
				return RenderError(c, openErr, "reading demo file")
			}

			log.Info("Demo mode selected")

			return c.HTMLBlob(http.StatusOK, file)
		}

		return echo.NewHTTPError(500, "no demo assets found")
	}
}
