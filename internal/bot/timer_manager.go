package bot

import (
	"context"
	"log"
	"sync"
	"time"

	"tg-timer/pkg/telegram"
)

// TimerManager manages active timers with thread safety
type TimerManager struct {
	timers   map[int64]*Timer // chatID -> Timer
	mu       sync.RWMutex
	telegram telegram.Client
}

// NewTimerManager creates new timer manager
func NewTimerManager(telegram telegram.Client) *TimerManager {
	return &TimerManager{
		timers:   make(map[int64]*Timer),
		telegram: telegram,
	}
}

// SetTimer creates new timer for chat, canceling existing one if present
func (tm *TimerManager) SetTimer(ctx context.Context, chatID int64, duration time.Duration, durationText string) error {
	// Cancel existing timer if any
	tm.CancelTimer(chatID)

	// Create context for this timer
	timerCtx, cancel := context.WithCancel(ctx)

	// Create timer object
	timer := &Timer{
		ChatID:     chatID,
		Duration:   duration,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	// Store timer
	tm.mu.Lock()
	tm.timers[chatID] = timer
	tm.mu.Unlock()

	// Start timer in goroutine
	go tm.runTimer(timerCtx, chatID, duration, durationText)

	log.Printf("Timer set for chat %d: %s", chatID, durationText)
	return nil
}

// CancelTimer cancels active timer for chat
func (tm *TimerManager) CancelTimer(chatID int64) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if timer, exists := tm.timers[chatID]; exists {
		timer.CancelFunc()
		delete(tm.timers, chatID)
		log.Printf("Timer cancelled for chat %d", chatID)
		return true
	}

	return false
}

// HasActiveTimer checks if chat has active timer
func (tm *TimerManager) HasActiveTimer(chatID int64) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.timers[chatID]
	return exists
}

// StopAll stops all active timers
func (tm *TimerManager) StopAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for chatID, timer := range tm.timers {
		timer.CancelFunc()
		log.Printf("Timer stopped for chat %d", chatID)
	}

	// Clear all timers
	tm.timers = make(map[int64]*Timer)
}

// runTimer runs timer and sends notification when done
func (tm *TimerManager) runTimer(ctx context.Context, chatID int64, duration time.Duration, durationText string) {
	select {
	case <-ctx.Done():
		// Timer was cancelled
		return
	case <-time.After(duration):
		// Timer completed
		tm.mu.Lock()
		delete(tm.timers, chatID)
		tm.mu.Unlock()

		// Send notification
		err := tm.telegram.SendMessage(ctx, chatID, "Время вышло!")
		if err != nil {
			log.Printf("Failed to send timer completion message to chat %d: %v", chatID, err)
		} else {
			log.Printf("Timer completed for chat %d", chatID)
		}
	}
}

// GetActiveTimerInfo returns info about active timer for chat
func (tm *TimerManager) GetActiveTimerInfo(chatID int64) (time.Duration, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if timer, exists := tm.timers[chatID]; exists {
		elapsed := time.Since(timer.StartTime)
		remaining := timer.Duration - elapsed
		if remaining < 0 {
			remaining = 0
		}
		return remaining, true
	}

	return 0, false
}
