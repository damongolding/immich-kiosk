package routes

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/templates/views"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/labstack/echo/v4"
)

func About(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

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
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     deviceID,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.About(viewData))
	}
}

func BuildUrl() echo.HandlerFunc {
	return func(c echo.Context) error {
		kioskHost := c.Request().Host
		kioskUrl, err := url.Parse(kioskHost)
		if err != nil {
			return err
		}

		if err := c.Request().ParseForm(); err != nil {
			return err
		}

		people := c.Request().Form["people"]
		q := url.Values{}
		for _, person := range people {
			q.Add("person", person)
		}

		albums := c.Request().Form["album"]
		for _, album := range albums {
			q.Add("album", album)
		}

		kioskUrl.RawQuery = q.Encode()

		return c.String(200, fmt.Sprintf(`<div id="url-result"><a href="%s">%s</a></div>`, kioskUrl.String(), kioskUrl.String()))
	}
}

func Url(baseConfig *config.Config, im immich_open_api.ClientWithResponsesInterface) echo.HandlerFunc {
	return func(c echo.Context) error {

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
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		size := float32(100)
		withHidden := true
		page := float32(1)
		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     deviceID,
			Config:       requestConfig,
		}

		ppl, err := im.GetAllPeopleWithResponse(c.Request().Context(), &immich_open_api.GetAllPeopleParams{Page: &page, Size: &size, WithHidden: &withHidden})
		if err != nil {
			return err
		}
		if ppl.StatusCode() != 200 {
			return errors.New("bad request: " + ppl.Status() + string(ppl.Body))
		}

		peopleWithNames := make([]immich_open_api.PersonResponseDto, 0)
		for _, p := range ppl.JSON200.People {
			if p.Name != "" {
				peopleWithNames = append(peopleWithNames, p)
			}
		}

		alb, err := im.GetAllAlbumsWithResponse(c.Request().Context(), &immich_open_api.GetAllAlbumsParams{})
		if err != nil {
			return err
		}
		if alb.StatusCode() != 200 {
			return errors.New("bad request: " + alb.Status() + string(alb.Body))
		}

		urlData := common.UrlViewData{
			People: peopleWithNames,
			Albums: *alb.JSON200,
		}

		return Render(c, http.StatusOK, views.Url(viewData, urlData))
	}
}
