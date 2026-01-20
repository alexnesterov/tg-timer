package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tg-timer/internal/bot"
	"tg-timer/pkg/telegram"
)

func main() {
	// Get configuration from environment
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("WEBHOOK_URL environment variable is required")
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

	// Setup webhook
	err := telegramClient.SetWebhook(ctx, webhookURL)
	if err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}
	log.Printf("Webhook set to: %s", webhookURL)

	// Setup HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: setupWebhookHandler(commandHandler),
	}

	log.Printf("Telegram timer bot started with webhook on port %s", port)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received")

	// Cancel context to stop all operations
	cancel()

	// Stop all timers
	timerManager.StopAll()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Delete webhook
	if err := telegramClient.DeleteWebhook(ctx); err != nil {
		log.Printf("Failed to delete webhook: %v", err)
	}

	log.Println("Bot stopped gracefully")
}

func setupWebhookHandler(commandHandler *bot.CommandHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update telegram.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("Failed to decode update: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Handle update in goroutine to avoid blocking
		go func(u telegram.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in update handler: %v", r)
				}
			}()
			commandHandler.HandleUpdate(context.Background(), u)
		}(update)

		w.WriteHeader(http.StatusOK)
	})
}
