package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

var (
	customTempVideoDir = filepath.Join(os.TempDir(), "immich-kiosk", "videos")
)

// Video represents a downloaded video file and its metadata
type Video struct {
	ID           string
	LastAccessed time.Time
	FileName     string
	FilePath     string
	ContentType  string
	ImmichAsset  immich.Asset
}

// Manager handles downloading and managing video files
type Manager struct {
	DownloadQueue []string

	Videos []Video
	MaxAge time.Duration
	mu     sync.RWMutex
}

// New creates a new VideoManager instance
func New(ctx context.Context) (*Manager, error) {
	if err := initialise(); err != nil {
		return nil, err
	}

	v := &Manager{}
	go v.VideoCleanup(ctx)

	return v, nil
}

// initialise initializes the video temp directory
func initialise() error {

	// Create custom temp directory if it doesn't exist
	err := os.MkdirAll(customTempVideoDir, 0755)
	if err != nil {
		log.Error("Error creating custom temp directory", "err", err)
		return err
	}

	log.Info("Created video tmp dir", "path", customTempVideoDir)

	return nil
}

// VideoCleanup runs periodic cleanup of old video files
func (v *Manager) VideoCleanup(ctx context.Context) {
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

// Delete removes the video temp directory and all its contents
func Delete() {
	log.Debug("Remove custom temp video directory")
	err := os.RemoveAll(customTempVideoDir)
	if err != nil {
		log.Error("Error removing custom temp directory", "err", err)
	}
}

// RemoveVideo deletes a video file and removes it from the manager
func (v *Manager) RemoveVideo(id string) {

	for i, video := range v.Videos {
		if video.ID == id {
			filePath := filepath.Join(customTempVideoDir, video.FileName)
			if _, err := os.Stat(filePath); err == nil {
				fileRemoveErr := os.Remove(filePath)
				if fileRemoveErr != nil {
					log.Error("deleting video", "video", filePath, "err", fileRemoveErr)
					continue
				}

				v.Videos = slices.Delete(v.Videos, i, i+1)

				log.Debug("deleted video", "video", filePath, "err", err)

			} else {
				log.Debug("video file not found", "video", filePath)
				v.Videos = slices.Delete(v.Videos, i, i+1)
			}

		}
	}
}

// cleanup removes videos that have exceeded their maximum age
func (v *Manager) cleanup() {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()

	for i := len(v.Videos) - 1; i >= 0; i-- {
		if now.Sub(v.Videos[i].LastAccessed) > v.MaxAge {
			v.RemoveVideo(v.Videos[i].ID)
		}
	}
}

// IsDownloaded checks if a video has already been downloaded
func (v *Manager) IsDownloaded(id string) bool {

	if _, err := v.GetVideo(id); err == nil {
		return true
	}

	return false
}

// IsDownloading checks if a video is currently being downloaded
func (v *Manager) IsDownloading(id string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return slices.Contains(v.DownloadQueue, id)
}

// GetVideo retrieves a video by ID
func (v *Manager) GetVideo(id string) (Video, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, video := range v.Videos {
		if video.ID == id {
			video.LastAccessed = time.Now()
			return video, nil
		}
	}

	return Video{}, errors.New("video not found")
}

// AddVideoToViewCache adds a downloaded video to the cache
func (v *Manager) AddVideoToViewCache(id, fileName, filePath, contentType string, requestConfig *config.Config, deviceID, requestURL string, immichAsset immich.Asset, imageData, imageBlurData string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Videos = append(v.Videos, Video{
		ID:           id,
		LastAccessed: time.Now(),
		FileName:     fileName,
		FilePath:     filePath,
		ContentType:  contentType,
		ImmichAsset:  immichAsset,
	})

	viewDataToAdd := common.ViewData{
		DeviceID: deviceID,
		Config:   *requestConfig,
		Assets: []common.ViewImageData{
			{
				ImmichAsset:   immichAsset,
				ImageData:     imageData,
				ImageBlurData: imageBlurData,
			},
		},
	}

	if requestURL != "" {
		cache.AssetToCacheWithPosition(viewDataToAdd, requestConfig, deviceID, requestURL, cache.PREPEND)
	}
}

// removeFromQueue removes a video ID from the download queue
func (v *Manager) removeFromQueue(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i, videoID := range v.DownloadQueue {
		if videoID == id {
			v.DownloadQueue = slices.Delete(v.DownloadQueue, i, i+1)
			break
		}
	}
}

// addToQueue adds a video ID to the download queue
func (v *Manager) addToQueue(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.DownloadQueue = append(v.DownloadQueue, id)
}

// DownloadVideo downloads a video file and adds it to the cache
func (v *Manager) DownloadVideo(immichAsset immich.Asset, requestConfig config.Config, deviceID string, requestURL string) {

	videoID := immichAsset.ID

	v.addToQueue(videoID)
	defer v.removeFromQueue(videoID)

	// Get the video data
	videoBytes, contentType, videoBytesErr := immichAsset.Video()
	if videoBytesErr != nil {
		log.Error("getting video", "err", videoBytesErr)
		return
	}

	ext := filepath.Ext(immichAsset.OriginalFileName)
	if strings.HasPrefix(contentType, "video/") {
		mediaType := strings.Split(contentType, ";")[0]
		parts := strings.Split(mediaType, "/")
		if len(parts) == 2 && parts[1] != "" {
			ext = "." + parts[1]
		}
	}

	filename := videoID + ext
	filePath := filepath.Join(customTempVideoDir, filename)

	// Create a file to save the video
	videoFile, videoFileErr := os.Create(filePath)
	if videoFileErr != nil {
		log.Error("Error creating video file", "err", videoFileErr)
		return
	}
	defer videoFile.Close()

	// Write the video data to the file
	_, videoFileErr = videoFile.Write(videoBytes)
	if videoFileErr != nil {
		log.Error("Error writing video file", "err", videoFileErr)
		return
	}

	var imageData, imageBlurData string

	defer func() {
		log.Debug(kiosk.DebugID+" Downloaded video", "path", filePath)
		v.AddVideoToViewCache(videoID, filename, filePath, contentType, &requestConfig, deviceID, requestURL, immichAsset, imageData, imageBlurData)
	}()

	imgBytes, _, imgBytesErr := immichAsset.ImagePreview()
	if imgBytesErr != nil {
		log.Debug("Getting image preview for video", "id", videoID, "err", imgBytesErr)
		return
	}

	img, imgErr := utils.BytesToImage(imgBytes, false)
	if imgErr != nil {
		log.Error("Image BytesToImage", "err", imgErr)
	}

	if requestConfig.OptimizeImages {
		img, imgErr = utils.OptimizeImage(img, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
		if imgErr != nil {
			log.Error("OptimizeImages", "err", imgErr)
		}
	}

	imgBlur, imgBlurErr := utils.BlurImage(img, requestConfig.BackgroundBlurAmount, false, 0, 0)
	if imgBlurErr != nil {
		log.Error("Getting image preview", "err", imgBlurErr)
	}

	imageData, imageDataErr := utils.ImageToBase64(img)
	if imageDataErr != nil {
		log.Error("Converting image to base64", "err", imageDataErr)
	}

	imageBlurData, err := utils.ImageToBase64(imgBlur)
	if err != nil {
		log.Error("Converting image to base64", "err", err)
	}
}
