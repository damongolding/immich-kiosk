package routes

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/log/v2"
)

type MaintenanceState struct {
	mu       sync.RWMutex
	active   bool
	message  string
	clients  map[chan string]struct{}
	clientMu sync.Mutex
	kioskCtx context.Context
	cancel   context.CancelFunc
	stopped  atomic.Bool
}

func NewMaintenance(ctx context.Context, path string, pollInterval time.Duration) *MaintenanceState {
	mCtx, cancel := context.WithCancel(ctx)
	m := &MaintenanceState{clients: make(map[chan string]struct{}), kioskCtx: ctx, cancel: cancel}
	go m.watch(mCtx, path, pollInterval)
	return m
}

func (m *MaintenanceState) watch(ctx context.Context, path string, interval time.Duration) {
	log.Info("Watching for maintenance file", "at", path)
	check := func() {
		data, err := os.ReadFile(path)
		active := !errors.Is(err, os.ErrNotExist)
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
			log.Info("Stopped watching for maintenance file", "at", path)
			return
		default:
			check()
		}
	}
}

func (m *MaintenanceState) broadcast(active bool, msg string) {
	log.Debug("broadcasting", "active", active, "message", msg)
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

func (m *MaintenanceState) Stop() {
	if m.cancel == nil {
		return
	}
	m.cancel()
	m.stopped.Store(true)
}

func (m *MaintenanceState) IsStopped() bool { return m.stopped.Load() }

type MaintenanceHolder struct {
	p atomic.Pointer[MaintenanceState]
}

func (h *MaintenanceHolder) Get() *MaintenanceState  { return h.p.Load() }
func (h *MaintenanceHolder) Set(m *MaintenanceState) { h.p.Store(m) }
