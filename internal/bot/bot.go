package bot

import (
	"context"
	"log"
	"time"

	"tg-timer/pkg/telegram"
)

// Run runs the main bot loop
func Run(ctx context.Context, telegramClient telegram.Client, commandHandler *CommandHandler) {
	var lastUpdateID int

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot loop stopped due to context cancellation")
			return
		default:
			// Get updates with timeout
			updates, err := telegramClient.GetUpdates(ctx, lastUpdateID+1, 30)
			if err != nil {
				if ctx.Err() != nil {
					// Context cancelled, exit
					return
				}
				log.Printf("Failed to get updates: %v", err)
				// Wait before retry
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Process updates
			for _, update := range updates {
				if update.UpdateID > lastUpdateID {
					lastUpdateID = update.UpdateID
				}

				// Handle update in goroutine to avoid blocking
				go func(u telegram.Update) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Panic in update handler: %v", r)
						}
					}()
					commandHandler.HandleUpdate(ctx, u)
				}(update)
			}
		}
	}
}
