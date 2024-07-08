package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
)

func home(c echo.Context) error {
	fmt.Println()

	// create a copy of the global config to use with this instance
	instanceConfig := config

	queries := c.Request().URL.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	return c.Render(http.StatusOK, "index.html", instanceConfig)

}

func newImage(c echo.Context) error {
	fmt.Println()

	instanceConfig := config

	referer, err := url.Parse(c.Request().Referer())
	if err != nil {
		log.Fatal(err)
	}

	queries := referer.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	immichImage := NewImage()

	if instanceConfig.People != "" {
		randomPersonImageErr := immichImage.GetRandomImageOfPerson(instanceConfig.People)
		if randomPersonImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomPersonImageErr.Error()})
		}
	} else {
		randomImageErr := immichImage.GetRandomImage()
		if randomImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomImageErr.Error()})
		}
	}

	// imageGet := time.Now()
	imgBytes, err := immichImage.GetImagePreview()
	if err != nil {
		return err
	}

	// fmt.Println(immichImage.OriginalFileName, ": Got image in", time.Since(imageGet).Seconds())

	// imageConvert := time.Now()
	img, err := ImageToBase64(imgBytes)
	if err != nil {
		return err
	}
	// fmt.Println(immichImage.OriginalFileName, ": Converted image in", time.Since(imageConvert).Seconds())

	date := fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format("02-01-2006"), immichImage.LocalDateTime.Format(time.Kitchen))

	data := PageData{
		ImageUrl:   img,
		Date:       date,
		FillScreen: instanceConfig.FillScreen,
	}

	return c.Render(http.StatusOK, "image.html", data)
}
