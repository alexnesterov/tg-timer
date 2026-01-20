package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tg-timer/internal/bot"
	"tg-timer/pkg/telegram"
)

func main() {
	// Get bot token from environment
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize components
	telegramClient := telegram.NewClient(token)
	timerManager := bot.NewTimerManager(telegramClient)
	commandHandler := bot.NewCommandHandler(timerManager, telegramClient)

	log.Println("Telegram timer bot started")

	// Start bot in goroutine
	go bot.Run(ctx, telegramClient, commandHandler)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received")

	// Cancel context to stop all operations
	cancel()

	// Stop all timers
	timerManager.StopAll()

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	log.Println("Bot stopped gracefully")
}
