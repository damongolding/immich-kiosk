package routes

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/log/v2"
)

type MaintenanceState struct {
	mu       sync.RWMutex
	active   bool
	message  string
	clients  map[chan string]struct{}
	clientMu sync.Mutex
}

func NewMaintenance(ctx context.Context, path string, pollInterval time.Duration) *MaintenanceState {
	m := &MaintenanceState{clients: make(map[chan string]struct{})}
	go m.watch(ctx, path, pollInterval)
	return m
}

func (m *MaintenanceState) watch(ctx context.Context, path string, interval time.Duration) {
	check := func() {
		data, err := os.ReadFile(path)
		active := err == nil
		msg := strings.TrimSpace(string(data))

		m.mu.Lock()
		changed := m.active != active || m.message != msg
		m.active, m.message = active, msg
		m.mu.Unlock()

		if changed {
			m.broadcast(active, msg)
		}
	}

	check() // initial state on startup
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			check()
		}
	}
}

func (m *MaintenanceState) broadcast(active bool, msg string) {
	log.Info("broadcasting", "active", active, "message", msg)
	payload := "inactive"
	if active {
		payload = "active:" + msg
	}
	m.clientMu.Lock()
	defer m.clientMu.Unlock()
	for ch := range m.clients {
		select {
		case ch <- payload:
		default: // slow client, drop rather than block
		}
	}
}

func (m *MaintenanceState) IsActive() (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active, m.message
}
