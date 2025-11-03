package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/labstack/echo/v4"
)

func BuildUrl() echo.HandlerFunc {
	return func(c echo.Context) error {
		kioskHost := c.Request().Header.Get("X-Forwarded-Host")
		if kioskHost == "" {
			kioskHost = c.Request().Host
		}
		scheme := c.Scheme()
		if xf := c.Request().Header.Get("X-Forwarded-Proto"); xf != "" {
			s := strings.ToLower(xf)
			if s == "http" || s == "https" {
				scheme = s
			}
		}
		prefix := c.Request().Header.Get("X-Forwarded-Prefix")
		if prefix != "" && !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		kioskUrl, err := url.Parse(fmt.Sprintf("%s://%s", scheme, kioskHost))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "parsing url from request")
		}

		// HACK: remove empty form values so optional fields parse as nil
		// and we can ignore them from the URL parameters to default to the servers configured defaults
		if err := c.Request().ParseForm(); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid form")
		}
		newValues := make(url.Values)
		for k, v := range c.Request().Form {
			if len(v) == 1 && (v[0] == "") {
				continue
			}
			newValues[k] = v
		}
		c.Request().Form = newValues

		var req common.UrlBuilderRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "parsing form")
		}

		q := url.Values{}
		for _, person := range req.People {
			q.Add("person", person)
		}

		for _, album := range req.Albums {
			q.Add("album", album)
		}

		if sd := req.ShowDate; sd != nil {
			q.Add("show_date", strconv.FormatBool(*sd))
		}

		if st := req.ShowTime; st != nil {
			q.Add("show_time", strconv.FormatBool(*st))
		}

		if rap := req.RequireAllPeople; rap != nil {
			q.Add("require_all_people", strconv.FormatBool(*rap))
		}

		if spb := req.ShowProgressBar; spb != nil {
			q.Add("show_progress_bar", strconv.FormatBool(*spb))
		}

		if pbp := req.ProgressBarPosition; pbp != nil {
			q.Add("progress_bar_position", *pbp)
		}

		if tr := req.Transition; tr != nil {
			q.Add("transition", *tr)
		}

		if lyt := req.Layout; lyt != nil {
			q.Add("layout", *lyt)
		}

		if dur := req.Duration; dur != nil {
			if *dur > 0 {
				q.Add("duration", strconv.FormatUint(*dur, 10))
			}
		}

		kioskUrl.RawQuery = q.Encode()

		return Render(c, http.StatusOK, partials.UrlResult(kioskUrl.String()))
	}
}

func UrlBuilderPage(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		im := immich.New(com.Context(), requestConfig)
		ppl, err := im.AllNamedPeople(requestID, deviceID)
		if err != nil {
			return err
		}

		albs, err := im.AllAlbums(requestID, deviceID)
		if err != nil {
			return err
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     deviceID,
			Config:       requestConfig,
		}

		urlData := common.UrlViewData{
			People: ppl,
			Albums: albs,
		}

		return Render(c, http.StatusOK, views.Url(viewData, urlData))
	}
}
