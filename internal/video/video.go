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
	"github.com/damongolding/immich-kiosk/internal/config"
)

var (
	tmpDirectory       string
	videoDirectory     string
	customTempVideoDir = filepath.Join(os.TempDir(), "immich-kiosk", "videos")
)

type Video struct {
	ID           string
	LastAccessed time.Time
	FileName     string
	FilePath     string
}

type VideoManager struct {
	mu sync.RWMutex

	baseConfig config.Config

	DownloadQueue []string

	Videos []Video
	MaxAge time.Duration
}

func New(ctx context.Context) (*VideoManager, error) {
	if err := VideoInit(); err != nil {
		return nil, err
	}

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

	for {
		select {
		case <-ctx.Done():
			log.Debug("Remove custom temp video directory")
			err := os.Remove(customTempVideoDir)
			if err != nil {
				log.Error("Error removing custom temp directory", "err", err)
			}
			return
		case <-ticker.C:
			v.cleanup()
		}
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

func (v *VideoManager) IsDownloading(id string) bool {
	v.mu.RLock()
	defer v.mu.Unlock()

	return slices.Contains(v.DownloadQueue, id)
}

func (v *VideoManager) GetVideo(id string) (Video, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, video := range v.Videos {
		if video.ID == id {
			return video, nil
		}
	}

	return Video{}, fmt.Errorf("video not found")
}

func (v *VideoManager) AddVideo(id, fileName, filePath string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Videos = append(v.Videos, Video{
		ID:           id,
		FileName:     fileName,
		FilePath:     filePath,
		LastAccessed: time.Now(),
	})
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

func (v *VideoManager) downloadVideo(id string) {

	u, err := url.Parse(v.baseConfig.ImmichUrl)
	if err != nil {
		log.Error(err)
		return
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", id, "video", "playback"),
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request to your API
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error("Error fetching video")
		return
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("x-api-key", v.baseConfig.ImmichApiKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error fetching video")
		return
	}
	defer resp.Body.Close()

	// Get the video filename
	filename := id + ".mp4"
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

	v.AddVideo(id, filename, filePath)
	log.Debug("downloaded video", "path", filePath)
}
