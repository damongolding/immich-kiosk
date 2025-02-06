package video

import (
	"context"
	"fmt"
	"os"
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
	customTempVideoDir = filepath.Join(os.TempDir(), "immich-kiosk", "videos")
)

// Video represents a downloaded video file and its metadata
type Video struct {
	ID           string
	LastAccessed time.Time
	FileName     string
	FilePath     string
	ImmichAsset  immich.ImmichAsset
}

// VideoManager handles downloading and managing video files
type VideoManager struct {
	mu sync.RWMutex

	DownloadQueue []string

	Videos []Video
	MaxAge time.Duration
}

// New creates a new VideoManager instance
func New(ctx context.Context, base config.Config) (*VideoManager, error) {
	if err := VideoInit(); err != nil {
		return nil, err
	}

	v := &VideoManager{}
	go v.VideoCleanup(ctx)

	return v, nil
}

// VideoInit initializes the video temp directory
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

// VideoCleanup runs periodic cleanup of old video files
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

// Delete removes the video temp directory and all its contents
func Delete() {
	log.Debug("Remove custom temp video directory")
	err := os.RemoveAll(customTempVideoDir)
	if err != nil {
		log.Error("Error removing custom temp directory", "err", err)
	}
}

// RemoveVideo deletes a video file and removes it from the manager
func (v *VideoManager) RemoveVideo(id string) {

	for i, video := range v.Videos {
		if video.ID == id {
			filePath := filepath.Join(customTempVideoDir, video.FileName)
			if _, err := os.Stat(filePath); err == nil {
				err := os.Remove(filePath)
				if err != nil {
					log.Error("deleting video", "video", filePath, "err", err)
					continue
				}

				v.Videos = append(v.Videos[:i], v.Videos[i+1:]...)

				log.Debug("deleted video", "video", filePath, "err", err)

			} else {
				log.Debug("video file not found", "video", filePath)
				v.Videos = append(v.Videos[:i], v.Videos[i+1:]...)
			}

		}
	}
}

// cleanup removes videos that have exceeded their maximum age
func (v *VideoManager) cleanup() {
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
func (v *VideoManager) IsDownloaded(id string) bool {

	if _, err := v.GetVideo(id); err == nil {
		return true
	}

	return false
}

// IsDownloading checks if a video is currently being downloaded
func (v *VideoManager) IsDownloading(id string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return slices.Contains(v.DownloadQueue, id)
}

// GetVideo retrieves a video by ID
func (v *VideoManager) GetVideo(id string) (Video, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, video := range v.Videos {
		if video.ID == id {
			video.LastAccessed = time.Now()
			return video, nil
		}
	}

	return Video{}, fmt.Errorf("video not found")
}

// AddVideoToViewCache adds a downloaded video to the cache
func (v *VideoManager) AddVideoToViewCache(id, fileName, filePath string, requestConfig *config.Config, deviceID, requestUrl string, immichAsset immich.ImmichAsset, imageData, imageBlurData string) {
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
		Assets: []common.ViewImageData{
			{
				ImmichAsset:   immichAsset,
				ImageData:     imageData,
				ImageBlurData: imageBlurData,
			},
		},
	}

	cache.AssetToCacheWithPosition(viewDataToAdd, requestConfig, deviceID, requestUrl, cache.PREPEND)
}

// removeFromQueue removes a video ID from the download queue
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

// addToQueue adds a video ID to the download queue
func (v *VideoManager) addToQueue(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.DownloadQueue = append(v.DownloadQueue, id)
}

// DownloadVideo downloads a video file and adds it to the cache
func (v *VideoManager) DownloadVideo(immichAsset immich.ImmichAsset, requestConfig config.Config, deviceID string, requestUrl string) {

	videoID := immichAsset.ID

	v.addToQueue(videoID)
	defer v.removeFromQueue(videoID)

	// Get the video data
	videoBytes, _, err := immichAsset.Video()
	if err != nil {
		log.Error("getting video", "err", err)
		return
	}

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

	// Write the video data to the file
	_, err = out.Write(videoBytes)
	if err != nil {
		log.Error("Error writing video file", "err", err)
		return
	}

	imgBytes, err := immichAsset.ImagePreview()
	if err != nil {
		log.Error("getting image preview", "err", err)
	}

	img, err := utils.BytesToImage(imgBytes)
	if err != nil {
		log.Error("image BytesToImage", "err", err)
	}

	img = utils.ApplyExifOrientation(img, immichAsset.IsLandscape, immichAsset.ExifInfo.Orientation)

	if requestConfig.OptimizeImages {
		img, err = utils.OptimizeImage(img, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
		if err != nil {
			log.Error("OptimizeImages", "err", err)
		}
	}

	imgBlur, err := utils.BlurImage(img, false, 0, 0)
	if err != nil {
		log.Error("getting image preview", "err", err)
	}

	imageData, err := utils.ImageToBase64(img)
	if err != nil {
		log.Error("converting image to base64", "err", err)
	}

	imageBlurData, err := utils.ImageToBase64(imgBlur)
	if err != nil {
		log.Error("converting image to base64", "err", err)
	}

	log.Debug("downloaded video", "path", filePath)

	v.AddVideoToViewCache(videoID, filename, filePath, &requestConfig, deviceID, requestUrl, immichAsset, imageData, imageBlurData)
}
