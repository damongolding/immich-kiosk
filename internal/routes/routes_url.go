package routes

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/labstack/echo/v4"
)

func BuildUrl() echo.HandlerFunc {
	return func(c echo.Context) error {
		kioskHost := c.Request().Host
		kioskUrl, err := url.Parse(kioskHost)
		if err != nil {
			return err
		}

		// HACK: remove empty form values so optional fields parse as nil
		// and we can ignore them from the URL parameters to default to the servers configured defaults
		c.Request().ParseForm()
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
			return err
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

		kioskUrl.RawQuery = q.Encode()

		// TODO(jj): htmx component to render
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
