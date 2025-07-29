package asset

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
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/disintegration/imaging"
)

var (
	customTempDir      = filepath.Join(os.TempDir(), "immich-kiosk")
	customTempVideoDir = filepath.Join(customTempDir, "videos")
	customTempImageDir = filepath.Join(customTempDir, "images")
)

// Asset represents a downloaded video file and its metadata
type Asset struct {
	ID           string
	LastAccessed time.Time
	FileName     string
	FilePath     string
	ContentType  string
	ImmichAsset  immich.Asset
}

func (a Asset) Get() ([]byte, error) {
	if a.FilePath == "" {
		return nil, errors.New("asset file path is empty")
	}

	data, err := os.ReadFile(a.FilePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a Asset) GetBlurred() ([]byte, error) {
	if a.FilePath == "" {
		return nil, errors.New("asset file path is empty")
	}

	fileExt := filepath.Ext(a.FilePath)
	filePath := strings.Replace(a.FilePath, fileExt, "_blurred.jpg", 1)

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	data, err := os.ReadFile(a.FilePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a Asset) GetBlurredPath() string {
	if a.FilePath == "" {
		return ""
	}

	fileExt := filepath.Ext(a.FilePath)
	filePath := strings.Replace(a.FilePath, fileExt, "_blurred.jpg", 1)

	return filePath
}

// Manager handles downloading and managing video files
type Manager struct {
	mu sync.RWMutex

	DownloadQueue []string

	Videos []Asset
	Images []Asset
	MaxAge time.Duration
}

// New creates a new VideoManager instance
func New(ctx context.Context) (*Manager, error) {
	if err := initialise(); err != nil {
		return nil, err
	}

	m := &Manager{}
	go m.AssetCleanup(ctx)

	return m, nil
}

// initialise initializes the video temp directory
func initialise() error {

	// Create custom temp directory if it doesn't exist
	err := os.MkdirAll(customTempVideoDir, 0755)
	if err != nil {
		log.Error("Error creating custom temp directory", "err", err)
		return err
	}

	err = os.MkdirAll(customTempImageDir, 0755)
	if err != nil {
		log.Error("Error creating custom temp directory", "err", err)
		return err
	}

	log.Info("created image tmp dir at", "path", customTempImageDir)
	log.Info("created video tmp dir at", "path", customTempVideoDir)

	return nil
}

// AssetCleanup runs periodic cleanup of old asset files
func (m *Manager) AssetCleanup(ctx context.Context) {
	// Run cleanup every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	// Add cleanup on function exit
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// Delete removes the temp directory and all its contents
func Delete() {
	log.Debug("Remove temp directory")
	err := os.RemoveAll(customTempDir)
	if err != nil {
		log.Error("Error removing temp directory", "err", err)
	}

}

// RemoveVideo deletes a video file and removes it from the manager
func (m *Manager) RemoveVideo(id string) {

	for i, video := range m.Videos {
		if video.ID == id {
			filePath := filepath.Join(customTempVideoDir, video.FileName)
			if _, err := os.Stat(filePath); err == nil {
				fileRemoveErr := os.Remove(filePath)
				if fileRemoveErr != nil {
					log.Error("deleting video", "video", filePath, "err", fileRemoveErr)
					continue
				}

				m.Videos = slices.Delete(m.Videos, i, i+1)

				log.Debug("deleted video", "video", filePath, "err", err)

			} else {
				log.Debug("video file not found", "video", filePath)
				m.Videos = slices.Delete(m.Videos, i, i+1)
			}

		}
	}
}

// cleanup removes videos that have exceeded their maximum age
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for i := len(m.Videos) - 1; i >= 0; i-- {
		if now.Sub(m.Videos[i].LastAccessed) > m.MaxAge {
			m.RemoveVideo(m.Videos[i].ID)
		}
	}

	for i := len(m.Images) - 1; i >= 0; i-- {
		if now.Sub(m.Images[i].LastAccessed) > m.MaxAge {
			m.RemoveImage(m.Images[i].ID)
		}
	}
}

// IsDownloaded checks if a video has already been downloaded
func (m *Manager) IsDownloaded(id string) bool {

	if _, err := m.GetVideo(id); err == nil {
		return true
	}

	return false
}

// IsDownloading checks if a video is currently being downloaded
func (m *Manager) IsDownloading(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return slices.Contains(m.DownloadQueue, id)
}

// GetVideo retrieves a video by ID
func (m *Manager) GetVideo(id string) (Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, video := range m.Videos {
		if video.ID == id {
			video.LastAccessed = time.Now()
			return video, nil
		}
	}

	return Asset{}, errors.New("video not found")
}

// AddVideoToViewCache adds a downloaded video to the cache
func (m *Manager) AddVideoToViewCache(id, fileName, filePath, contentType string, requestConfig *config.Config, deviceID, requestURL string, immichAsset immich.Asset, imageData, imageBlurData string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Videos = append(m.Videos, Asset{
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
func (m *Manager) removeFromQueue(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, videoID := range m.DownloadQueue {
		if videoID == id {
			m.DownloadQueue = slices.Delete(m.DownloadQueue, i, i+1)
			break
		}
	}
}

// addToQueue adds a video ID to the download queue
func (m *Manager) addToQueue(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DownloadQueue = append(m.DownloadQueue, id)
}

// DownloadVideo downloads a video file and adds it to the cache
func (m *Manager) DownloadVideo(immichAsset immich.Asset, requestConfig config.Config, deviceID string, requestURL string) {

	videoID := immichAsset.ID

	m.addToQueue(videoID)
	defer m.removeFromQueue(videoID)

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
		log.Debug("downloaded video", "path", filePath)
		m.AddVideoToViewCache(videoID, filename, filePath, contentType, &requestConfig, deviceID, requestURL, immichAsset, imageData, imageBlurData)
	}()

	imgBytes, _, imgBytesErr := immichAsset.ImagePreview()
	if imgBytesErr != nil {
		log.Debug("getting image preview for video", "id", videoID, "err", imgBytesErr)
		return
	}

	img, imgErr := utils.BytesToImage(imgBytes)
	if imgErr != nil {
		log.Error("image BytesToImage", "err", imgErr)
	}

	img = utils.ApplyExifOrientation(img, immichAsset.ExifInfo.Orientation)

	if requestConfig.OptimizeImages {
		img, imgErr = utils.OptimizeImage(img, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
		if imgErr != nil {
			log.Error("OptimizeImages", "err", imgErr)
		}
	}

	imgBlur, imgBlurErr := utils.BlurImage(img, requestConfig.BackgroundBlurAmount, false, 0, 0)
	if imgBlurErr != nil {
		log.Error("getting image preview", "err", imgBlurErr)
	}

	imageData, imageDataErr := utils.ImageToBase64(img)
	if imageDataErr != nil {
		log.Error("converting image to base64", "err", imageDataErr)
	}

	imageBlurData, err := utils.ImageToBase64(imgBlur)
	if err != nil {
		log.Error("converting image to base64", "err", err)
	}
}

// GetImage retrieves an image by ID
func (m *Manager) GetImage(id string) (Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, image := range m.Images {
		if image.ID == id {
			m.Images[i].LastAccessed = time.Now()
			return image, nil
		}
	}

	return Asset{}, errors.New("image not found")
}

func (m *Manager) DownloadImage(immichAsset immich.Asset, requestConfig config.Config, deviceID string) {

	// Get the image data
	imageData, contentType, err := immichAsset.ImagePreview()
	if err != nil {
		log.Error("getting image preview", "err", err)
		return
	}

	ext := filepath.Ext(immichAsset.OriginalFileName)
	if strings.HasPrefix(contentType, "image/") {
		mediaType := strings.Split(contentType, ";")[0]
		parts := strings.Split(mediaType, "/")
		if len(parts) == 2 && parts[1] != "" {
			ext = "." + parts[1]
		}
	}

	fileName := immichAsset.ID + ext
	filePath := filepath.Join(customTempImageDir, fileName)
	blurredFilePath := filepath.Join(customTempImageDir, immichAsset.ID+"_blurred.jpg")

	img, imgErr := utils.BytesToImage(imageData)
	if imgErr != nil {
		log.Error("converting image bytes to image", "err", imgErr)
		return
	}

	blurredImage, blurredImageErr := utils.BlurImage(img, requestConfig.BackgroundBlurAmount, requestConfig.OptimizeImages, requestConfig.ClientData.Width, requestConfig.ClientData.Height)
	if blurredImageErr != nil {
		log.Error("blurring image", "err", blurredImageErr)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a file to save the image
	imageFile, imageFileErr := os.Create(filePath)
	if imageFileErr != nil {
		log.Error("Error creating image file", "err", imageFileErr)
		return
	}
	defer imageFile.Close()

	// Write the video data to the file
	_, imageFileErr = imageFile.Write(imageData)
	if imageFileErr != nil {
		log.Error("Error writing image file", "err", imageFileErr)
		return
	}

	// Create a file to save the blurred image
	blurredImageFile, blurredImageFileErr := os.Create(blurredFilePath)
	if blurredImageFileErr != nil {
		log.Error("Error creating image file", "err", blurredImageFileErr)
		return
	}
	defer blurredImageFile.Close()

	blurredImageFileErr = imaging.Encode(blurredImageFile, blurredImage, imaging.JPEG, imaging.JPEGQuality(50))
	if blurredImageFileErr != nil {
		log.Error("Error writing image file", "err", blurredImageFileErr)
		return
	}

	m.Images = append(m.Images, Asset{
		ID:           immichAsset.ID,
		LastAccessed: time.Now(),
		FileName:     fileName,
		FilePath:     filePath,
		ContentType:  contentType,
		ImmichAsset:  immichAsset,
	})
}

func (m *Manager) RemoveImage(id string) {

	for i, image := range m.Images {
		if image.ID == id {
			filePath := filepath.Join(customTempImageDir, image.FileName)
			if _, err := os.Stat(filePath); err == nil {
				fileRemoveErr := os.Remove(filePath)
				if fileRemoveErr != nil {
					log.Error("deleting image", "image", filePath, "err", fileRemoveErr)
					continue
				}

				m.Images = slices.Delete(m.Images, i, i+1)

				log.Debug("deleted image", "image", filePath, "err", err)

			} else {
				log.Debug("image file not found", "image", filePath)
				m.Images = slices.Delete(m.Images, i, i+1)
			}

		}
	}
}
