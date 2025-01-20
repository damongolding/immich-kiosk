package video

import (
	"slices"
	"sync"
	"time"
)

var (
	tmpDirectory   string
	videoDirectory string
)

type Video struct {
	ID           string
	LastAccessed time.Time
}

type VideoManager struct {
	mu sync.RWMutex

	DownloadQueue []string

	Videos []Video
}

func init() {

}

func New() *VideoManager {
	return &VideoManager{}
}

func (v *VideoManager) IsDownloading(id string) bool {
	v.mu.RLock()
	defer v.mu.Unlock()

	return slices.Contains(v.DownloadQueue, id)
}
