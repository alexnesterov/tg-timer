package bot

import (
	"context"
	"time"
)

// Timer represents an active timer
type Timer struct {
	ChatID     int64
	Duration   time.Duration
	StartTime  time.Time
	CancelFunc context.CancelFunc
}

// Command represents parsed command
type Command struct {
	Name string
	Args string
}

// TimerDuration represents parsed timer duration
type TimerDuration struct {
	Duration time.Duration
	Text     string // Human readable text (e.g., "10 минут")
}
