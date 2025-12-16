// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"context"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v4"
)

type Common struct {
	ctx    context.Context
	cancel context.CancelFunc
	secret string
}

func New() *Common {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Common{
		ctx:    ctx,
		cancel: cancel,
	}

	if err := c.initializeSecret(); err != nil {
		log.Fatal("failed to initialize shared secret", "error", err)
	}

	c.handleGracefulShutdown()
	return c
}

// initializeSecret generates and sets a secret token that is shared between application components
// this shared secret is used for secure communication and authentication between services
func (c *Common) initializeSecret() error {

	secret, err := utils.GenerateSharedSecret()
	if err != nil {
		return fmt.Errorf("failed to generate shared secret: %w", err)
	}
	c.secret = secret

	return nil
}

func (c *Common) handleGracefulShutdown() {
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-sigChan:
			c.cancel()
		case <-c.ctx.Done():
		}
		signal.Stop(sigChan)
	}()
}

func (c *Common) Context() context.Context {
	return c.ctx
}

func (c *Common) Secret() string {
	return c.secret
}

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
	RequestConfig config.Config // Configuration for the current request
}

// ViewImageData contains the image data and metadata for displaying an image in the view
type ViewImageData struct {
	ImageData          string       // ImageData contains the image as base64 data
	ImageBlurData      string       // ImageBlurData contains the blurred image as base64 data
	ImageDate          string       // ImageDate contains the date of the image
	User               string       // User the user api key used
	ImmichAsset        immich.Asset // ImmichAsset contains immich asset data
	ImageDominantColor color.RGBA   // ImageDominantColor contains the dominant color of the image
}

// ViewData contains all the data needed to render a view in the application
type ViewData struct {
	KioskVersion  string          // KioskVersion contains the current build version of Kiosk
	RequestID     string          // RequestID contains the unique identifier for the request
	DeviceID      string          // DeviceID contains the unique identifier for the device
	Assets        []ViewImageData // Assets contains the collection of assets to display in view
	Queries       url.Values      // Queries contains the URL query parameters
	CustomCSS     []byte          // CustomCSS contains custom CSS styling as bytes
	config.Config                 // Config contains the instance configuration
}

type ViewImageDataOptions struct {
	RelativeAssetBucket   kiosk.Source
	RelativeAssetBucketID string
	ImageOrientation      immich.ImageOrientation
	RelativeAssetWanted   bool
}

// ContextCopy stores a copy of key HTTP context information including URL and headers
type ContextCopy struct {
	RequestHeader  http.Header // Headers from the incoming request
	ResponseHeader http.Header // Headers for the outgoing response
	URL            url.URL     // The request URL
}

// CopyContext creates a copy of essential context data from an echo.Context
// This allows preserving context information without maintaining a reference to the original context
// Returns a ContextCopy containing the URL and header information
func CopyContext(c echo.Context) ContextCopy {

	ctxCopy := ContextCopy{
		URL:            *c.Request().URL,
		RequestHeader:  c.Request().Header.Clone(),
		ResponseHeader: c.Response().Header().Clone(),
	}

	return ctxCopy
}

type URLViewData struct {
	People []immich.Person
	Albums []immich.Album
	Tags   []immich.Tag
}

type URLBuilderRequest struct {
	// Duration
	Duration       *uint64 `form:"duration" url:"duration,omitempty"`
	OptimizeImages *bool   `form:"optimize_images" url:"optimize_images,omitempty"`
	BurnInInterval *uint64 `form:"burn_in_interval" url:"burn_in_interval,omitempty"`
	BurnInDuration *uint64 `form:"burn_in_duration" url:"burn_in_duration,omitempty"`
	BurnInOpacity  *uint64 `form:"burn_in_opacity" url:"burn_in_opacity,omitempty"`

	// Buckets
	People           []string `form:"people" url:"person,omitempty"`
	RequireAllPeople *bool    `form:"require_all_people" url:"require_all_people,omitempty"`
	ExcludedPeople   []string `form:"excluded_people" url:"excluded_person,omitempty"`
	Albums           []string `form:"album" url:"album,omitempty"`
	AlbumOrder       *string  `form:"album_order" url:"album_order,omitempty"`
	ExcludedAlbums   []string `form:"excluded_albums" url:"excluded_album,omitempty"`
	Tags             []string `form:"tag" url:"tag,omitempty"`
	ExcludedTags     []string `form:"excluded_tags" url:"excluded_tag,omitempty"`
	ShowMemories     *bool    `form:"memories" url:"memories,omitempty"`
	PastMemoryDays   *uint64  `form:"past_memory_days" url:"past_memory_days,omitempty"`

	ShowArchived *bool `form:"show_archived" url:"show_archived,omitempty"`

	// Video
	ShowVideos         *bool   `form:"show_videos" url:"show_videos,omitempty"`
	LivePhotos         *bool   `form:"live_photos" url:"live_photos,omitempty"`
	LivePhotoLoopDelay *uint64 `form:"live_photo_loop_delay" url:"live_photo_loop_delay,omitempty"`

	// Clock
	ShowTime    *bool   `form:"show_time" url:"show_time,omitempty"`
	TimeFormat  *string `form:"time_format" url:"time_format,omitempty"`
	ShowDate    *bool   `form:"show_date" url:"show_date,omitempty"`
	DateFormat  *string `form:"date_format" url:"date_format,omitempty"`
	ClockSource *string `form:"clock_source" url:"clock_source,omitempty"`

	// UI
	ShowClearCacheButton *bool   `form:"show_clear_cache_button" url:"show_clear_cache_button,omitempty"`
	ShowProgressBar      *bool   `form:"show_progress_bar" url:"show_progress_bar,omitempty"`
	ProgressBarPosition  *string `form:"progress_bar_position" url:"progress_bar_position,omitempty"`
	HideCursor           *bool   `form:"hide_cursor" url:"hide_cursor,omitempty"`
	FontSize             *uint64 `form:"font_size" url:"font_size,omitempty"`
	Theme                *string `form:"theme" url:"theme,omitempty"`
	Layout               *string `form:"layout" url:"layout,omitempty"`

	// Transition
	Transition *string `form:"transition" url:"transition,omitempty"`

	// Image
	ImageFit          *string `form:"image_fit" url:"image_fit,omitempty"`
	ImageEffect       *string `form:"image_effect" url:"image_effect,omitempty"`
	ImageEffectAmount *uint64 `form:"image_effect_amount" url:"image_effect_amount,omitempty"`
	UseOriginalImage  *bool   `form:"use_original_image" url:"use_original_image,omitempty"`

	// Metadata
	ShowOwner            *bool   `form:"show_owner" url:"show_owner,omitempty"`
	ShowAlbumName        *bool   `form:"show_album_name" url:"show_album_name,omitempty"`
	ShowPersonName       *bool   `form:"show_person_name" url:"show_person_name,omitempty"`
	ShowPersonAge        *bool   `form:"show_person_age" url:"show_person_age,omitempty"`
	ShowImageTime        *bool   `form:"show_image_time" url:"show_image_time,omitempty"`
	ImageTimeFormat      *string `form:"image_time_format" url:"image_time_format,omitempty"`
	ShowImageDate        *bool   `form:"show_image_date" url:"show_image_date,omitempty"`
	ImageDateFormat      *string `form:"image_date_format" url:"image_date_format,omitempty"`
	ShowImageDescription *bool   `form:"show_image_description" url:"show_image_description,omitempty"`
	ShowImageCamera      *bool   `form:"show_image_camera" url:"show_image_camera,omitempty"`
	ShowImageEXIF        *bool   `form:"show_image_exif" url:"show_image_exif,omitempty"`
	ShowImageLocation    *bool   `form:"show_image_location" url:"show_image_location,omitempty"`
	ShowImageQR          *bool   `form:"show_image_qr" url:"show_image_qr,omitempty"`
	ShowImageID          *bool   `form:"show_image_id" url:"show_image_id,omitempty"`

	// Show more overlay
	ShowMoreInfo          *bool    `form:"show_more_info" url:"show_more_info,omitempty"`
	ShowMoreInfoImageLink *bool    `form:"show_more_info_image_link" url:"show_more_info_image_link,omitempty"`
	ShowMoreInfoQRCode    *bool    `form:"show_more_info_qr_code" url:"show_more_info_qr_code,omitempty"`
	LikeButtonAction      []string `form:"like_button_action" url:"like_button_action,omitempty"`
	HideButtonAction      []string `form:"hide_button_action" url:"hide_button_action,omitempty"`
}
