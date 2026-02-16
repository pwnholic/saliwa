// Package sync provides clock synchronization and nonce generation.
// Critical for exchange API signature validation.
package sync

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lilwiggy/ex-act/pkg/errors"
)

// ClockSync maintains synchronized time with exchange server.
// Binance requires clock offset < 500ms for signed requests.
type ClockSync struct {
	exchange string

	// State
	offset   atomic.Int64 // Server time - Local time (in milliseconds)
	lastSync atomic.Int64 // Last sync timestamp (Unix milliseconds)

	// Configuration
	maxOffset    time.Duration // Maximum allowed offset
	syncInterval time.Duration // How often to sync
	timeProvider TimeProvider  // Function to get server time

	// Control
	mutex   sync.Mutex
	stopCh  chan struct{}
	running atomic.Bool
}

// TimeProvider returns the server time in milliseconds.
type TimeProvider func(ctx context.Context) (int64, error)

// ClockConfig contains clock synchronization configuration.
type ClockConfig struct {
	MaxOffset    time.Duration // Maximum allowed offset (default: 500ms)
	SyncInterval time.Duration // Sync interval (default: 5m)
	TimeProvider TimeProvider  // Function to get server time
}

// DefaultClockConfig returns default clock configuration.
func DefaultClockConfig() ClockConfig {
	return ClockConfig{
		MaxOffset:    500 * time.Millisecond,
		SyncInterval: 5 * time.Minute,
	}
}

// NewClockSync creates a new clock synchronizer.
func NewClockSync(exchange string, cfg ClockConfig) *ClockSync {
	if cfg.MaxOffset == 0 {
		cfg.MaxOffset = DefaultClockConfig().MaxOffset
	}
	if cfg.SyncInterval == 0 {
		cfg.SyncInterval = DefaultClockConfig().SyncInterval
	}

	return &ClockSync{
		exchange:     exchange,
		maxOffset:    cfg.MaxOffset,
		syncInterval: cfg.SyncInterval,
		timeProvider: cfg.TimeProvider,
		stopCh:       make(chan struct{}),
	}
}

// Start begins periodic clock synchronization.
func (cs *ClockSync) Start() error {
	if cs.running.Swap(true) {
		return nil // Already running
	}

	// Initial sync
	if err := cs.Sync(); err != nil {
		cs.running.Store(false)
		return err
	}

	// Start periodic sync
	go cs.syncLoop()

	log.Info().
		Str("exchange", cs.exchange).
		Dur("interval", cs.syncInterval).
		Msg("clock sync started")

	return nil
}

// Stop stops the clock synchronizer.
func (cs *ClockSync) Stop() {
	if !cs.running.Swap(false) {
		return
	}

	close(cs.stopCh)
	log.Info().Str("exchange", cs.exchange).Msg("clock sync stopped")
}

// Sync performs a single clock synchronization.
func (cs *ClockSync) Sync() error {
	if cs.timeProvider == nil {
		return errors.NewValidationError("timeProvider", nil, "must not be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Measure round-trip time
	localStart := time.Now().UnixMilli()
	serverTime, err := cs.timeProvider(ctx)
	if err != nil {
		return errors.NewConnectionError(cs.exchange, "clock", "sync failed: "+err.Error(), true)
	}
	localEnd := time.Now().UnixMilli()

	// Calculate offset (accounting for network latency)
	// Assume server time is at midpoint of round-trip
	localMid := (localStart + localEnd) / 2
	offset := serverTime - localMid

	cs.offset.Store(offset)
	cs.lastSync.Store(time.Now().UnixMilli())

	log.Debug().
		Str("exchange", cs.exchange).
		Int64("offset_ms", offset).
		Msg("clock synchronized")

	// Check if offset is acceptable
	if abs(offset) > cs.maxOffset.Milliseconds() {
		return errors.NewClockSyncError(
			cs.exchange,
			time.UnixMilli(localMid),
			time.UnixMilli(serverTime),
			time.Duration(abs(offset))*time.Millisecond,
		)
	}

	return nil
}

// syncLoop runs periodic synchronization.
func (cs *ClockSync) syncLoop() {
	ticker := time.NewTicker(cs.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopCh:
			return
		case <-ticker.C:
			if err := cs.Sync(); err != nil {
				log.Error().Err(err).Str("exchange", cs.exchange).Msg("clock sync failed")
			}
		}
	}
}

// Now returns the current synchronized time.
func (cs *ClockSync) Now() time.Time {
	offset := cs.offset.Load()
	return time.UnixMilli(time.Now().UnixMilli() + offset)
}

// UnixMilli returns the current synchronized time in milliseconds.
func (cs *ClockSync) UnixMilli() int64 {
	offset := cs.offset.Load()
	return time.Now().UnixMilli() + offset
}

// Offset returns the current clock offset.
func (cs *ClockSync) Offset() time.Duration {
	return time.Duration(cs.offset.Load()) * time.Millisecond
}

// IsSynchronized returns true if clock has been synchronized.
func (cs *ClockSync) IsSynchronized() bool {
	return cs.lastSync.Load() > 0
}

// LastSync returns the time of the last synchronization.
func (cs *ClockSync) LastSync() time.Time {
	return time.UnixMilli(cs.lastSync.Load())
}

// ValidateOffset checks if the current offset is within acceptable range.
func (cs *ClockSync) ValidateOffset() error {
	offset := abs(cs.offset.Load())
	maxMs := cs.maxOffset.Milliseconds()

	if offset > maxMs {
		return errors.NewClockSyncError(
			cs.exchange,
			time.Now(),
			cs.Now(),
			time.Duration(offset)*time.Millisecond,
		)
	}

	return nil
}

// SetTimeProvider sets the time provider function.
func (cs *ClockSync) SetTimeProvider(provider TimeProvider) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.timeProvider = provider
}

// abs returns the absolute value of an int64.
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
