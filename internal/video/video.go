package video

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/charmbracelet/log"
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
}

type VideoManager struct {
	mu sync.RWMutex

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

func (v *VideoManager) AddVideo(id, fileName string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Videos = append(v.Videos, Video{
		ID:           id,
		FileName:     fileName,
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

}
