package video

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

var (
	tmpDirectory       string
	videoDirectory     string
	customTempVideoDir = filepath.Join(os.TempDir(), "immich-kiosk", "videos")

	requestConfig config.Config
)

type Video struct {
	ID           string
	LastAccessed time.Time
	FileName     string
	FilePath     string
	ImmichAsset  immich.ImmichAsset
}

type VideoManager struct {
	mu sync.RWMutex

	DownloadQueue []string

	Videos []Video
	MaxAge time.Duration
}

func New(ctx context.Context, base config.Config) (*VideoManager, error) {
	if err := VideoInit(); err != nil {
		return nil, err
	}

	requestConfig = base

	v := &VideoManager{}
	go v.VideoCleanup(ctx)

	return v, nil
}

func VideoInit() error {

	// Create custom temp directory if it doesn't exist
	err := os.MkdirAll(customTempVideoDir, 0755)
	if err != nil {
		log.Error("Error creating custom temp directory", "err", err)
		return err
	}

	log.Info("created video tmp dir at", "path", customTempVideoDir)

	return nil
}

func (v *VideoManager) VideoCleanup(ctx context.Context) {
	// Run cleanup every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	// Add cleanup on function exit
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.cleanup()
		}
	}
}

func Delete() {
	log.Debug("Remove custom temp video directory")
	err := os.RemoveAll(customTempVideoDir)
	if err != nil {
		log.Error("Error removing custom temp directory", "err", err)
		fmt.Print(err)
	}
}

func (v *VideoManager) RemoveVideo(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i, video := range v.Videos {
		if video.ID == id {
			filePath := filepath.Join(customTempVideoDir, video.FileName)
			if _, err := os.Stat(filePath); err == nil {
				err := os.Remove(filePath)
				if err != nil {
					log.Error("deleting video", "video", filePath, "err", err)
					return
				}

				v.Videos = append(v.Videos[:i], v.Videos[i+1:]...)

				log.Debug("deleted video", "video", filePath, "err", err)

			} else {
				return
			}

		}
	}
}

func (v *VideoManager) cleanup() {

	now := time.Now()

	for i := len(v.Videos) - 1; i >= 0; i-- {
		if now.Sub(v.Videos[i].LastAccessed) > v.MaxAge {
			v.RemoveVideo(v.Videos[i].ID)
		}
	}
}

func (v *VideoManager) IsDownloaded(id string) bool {

	if _, err := v.GetVideo(id); err == nil {
		return true
	}

	return false
}

func (v *VideoManager) IsDownloading(id string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return slices.Contains(v.DownloadQueue, id)
}

func (v *VideoManager) GetVideo(id string) (Video, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, video := range v.Videos {
		if video.ID == id {
			video.LastAccessed = time.Now()
			return video, nil
		}
	}

	return Video{}, fmt.Errorf("video not found")
}

func (v *VideoManager) AddVideoToViewCache(id, fileName, filePath string, requestConfig *config.Config, deviceID, requestUrl string, immichAsset immich.ImmichAsset) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Videos = append(v.Videos, Video{
		ID:           id,
		FileName:     fileName,
		FilePath:     filePath,
		LastAccessed: time.Now(),
		ImmichAsset:  immichAsset,
	})

	viewDataToAdd := common.ViewData{
		DeviceID: deviceID,
		Config:   *requestConfig,
		Images: []common.ViewImageData{
			common.ViewImageData{
				ImmichAsset: immichAsset,
			},
		},
	}

	log.Info("Adding video to cache")

	ViewDataToCache(viewDataToAdd, requestConfig, deviceID, nil, requestUrl)
}

func (v *VideoManager) updateLastAccessed(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i := range v.Videos {
		if v.Videos[i].ID == id {
			v.Videos[i].LastAccessed = time.Now()
			break
		}
	}
}

func (v *VideoManager) removeFromQueue(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i, videoID := range v.DownloadQueue {
		if videoID == id {
			v.DownloadQueue = append(v.DownloadQueue[:i], v.DownloadQueue[i+1:]...)
			break
		}
	}
}

func (v *VideoManager) DownloadVideo(immichAsset immich.ImmichAsset, requestConfig config.Config, deviceID string, requestUrl string) {

	log.Info("Downloading video")

	videoID := immichAsset.ID

	v.DownloadQueue = append(v.DownloadQueue, videoID)
	defer v.removeFromQueue(videoID)

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Error(err)
		return
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", videoID, "video", "playback"),
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request to your API
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error("Error fetching video: NewRequest", "err", err)
		return
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("x-api-key", requestConfig.ImmichApiKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error fetching video: Do", "err", err)
		return
	}
	defer resp.Body.Close()

	ext := filepath.Ext(immichAsset.OriginalFileName)

	// Get the video filename
	filename := videoID + ext
	filePath := filepath.Join(customTempVideoDir, filename)

	// Create a file to save the video
	out, err := os.Create(filePath)
	if err != nil {
		log.Error("Error creating video file", "err", err)
		return
	}
	defer out.Close()

	// Copy the video data to the file
	_, err = out.ReadFrom(resp.Body)
	if err != nil {
		log.Error("Error writing video file", "err", err)
		return
	}

	v.AddVideoToViewCache(videoID, filename, filePath, &requestConfig, deviceID, requestUrl, immichAsset)

	log.Debug("downloaded video", "path", filePath)
}

func ViewDataToCache(viewDataToAdd common.ViewData, requestConfig *config.Config, deviceID string, requestData any, url string) {
	utils.TrimHistory(&requestConfig.History, 10)

	cachedViewData := []common.ViewData{}

	viewCacheKey := cache.ViewCacheKey(url, deviceID)

	if data, found := cache.Get(viewCacheKey); found {
		cachedViewData = data.([]common.ViewData)
	}

	cachedViewData = append(cachedViewData, viewDataToAdd)

	cache.Set(viewCacheKey, cachedViewData)

	// go webhooks.Trigger(requestData, KioskVersion, webhooks.PrefetchAsset, viewDataToAdd)
}
