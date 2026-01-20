package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tg-timer/pkg/telegram"
)

// CommandHandler handles bot commands
type CommandHandler struct {
	timerManager *TimerManager
	telegram     telegram.Client
}

// NewCommandHandler creates new command handler
func NewCommandHandler(timerManager *TimerManager, telegram telegram.Client) *CommandHandler {
	return &CommandHandler{
		timerManager: timerManager,
		telegram:     telegram,
	}
}

// HandleUpdate processes incoming update
func (ch *CommandHandler) HandleUpdate(ctx context.Context, update telegram.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	command := parseCommand(update.Message.Text)
	if command == nil {
		return
	}

	chatID := update.Message.Chat.ID
	log.Printf("Received command '%s' from chat %d", command.Name, chatID)

	switch command.Name {
	case "timer":
		ch.handleTimerCommand(ctx, chatID, command.Args)
	case "cancel":
		ch.handleCancelCommand(ctx, chatID)
	default:
		ch.sendUnknownCommandMessage(ctx, chatID)
	}
}

// parseCommand parses message text into command
func parseCommand(text string) *Command {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return nil
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}

	commandName := strings.TrimPrefix(parts[0], "/")
	commandName = strings.ToLower(commandName)

	var args string
	if len(parts) > 1 {
		args = strings.Join(parts[1:], " ")
	}

	return &Command{
		Name: commandName,
		Args: args,
	}
}

// handleTimerCommand processes /timer command
func (ch *CommandHandler) handleTimerCommand(ctx context.Context, chatID int64, args string) {
	if args == "" {
		ch.sendMessage(ctx, chatID, "Использование: /timer Xs или /timer Xm\nПример: /timer 30s или /timer 10m")
		return
	}

	duration, durationText, err := parseTimerDuration(args)
	if err != nil {
		ch.sendMessage(ctx, chatID, "Неверный формат времени. Используйте: /timer 30s или /timer 10m")
		return
	}

	if duration > 24*time.Hour {
		ch.sendMessage(ctx, chatID, "Максимальное время таймера - 24 часа")
		return
	}

	err = ch.timerManager.SetTimer(ctx, chatID, duration, durationText)
	if err != nil {
		ch.sendMessage(ctx, chatID, "Ошибка при установке таймера. Попробуйте еще раз.")
		log.Printf("Failed to set timer for chat %d: %v", chatID, err)
		return
	}

	message := fmt.Sprintf("Таймер на %s установлен.", durationText)
	ch.sendMessage(ctx, chatID, message)
}

// handleCancelCommand processes /cancel command
func (ch *CommandHandler) handleCancelCommand(ctx context.Context, chatID int64) {
	if ch.timerManager.CancelTimer(chatID) {
		ch.sendMessage(ctx, chatID, "Таймер отменён.")
	} else {
		ch.sendMessage(ctx, chatID, "Активный таймер не найден.")
	}
}

// sendUnknownCommandMessage sends message for unknown command
func (ch *CommandHandler) sendUnknownCommandMessage(ctx context.Context, chatID int64) {
	ch.sendMessage(ctx, chatID, "Неизвестная команда. Доступные команды:\n/timer Xs или /timer Xm - установить таймер\n/cancel - отменить таймер")
}

// sendMessage sends message with error logging
func (ch *CommandHandler) sendMessage(ctx context.Context, chatID int64, text string) {
	err := ch.telegram.SendMessage(ctx, chatID, text)
	if err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
	}
}

// parseTimerDuration parses timer duration string (e.g., "30s", "10m")
func parseTimerDuration(input string) (time.Duration, string, error) {
	// Regular expression to match number + unit
	re := regexp.MustCompile(`^(\d+)([sm])$`)
	matches := re.FindStringSubmatch(strings.ToLower(strings.TrimSpace(input)))

	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid format")
	}

	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("invalid number: %w", err)
	}

	unit := matches[2]

	var duration time.Duration
	var durationText string

	switch unit {
	case "s":
		if number > 86400 { // Max 24 hours in seconds
			return 0, "", fmt.Errorf("too many seconds")
		}
		duration = time.Duration(number) * time.Second
		if number == 1 {
			durationText = "1 секунду"
		} else if number < 5 {
			durationText = fmt.Sprintf("%d секунды", number)
		} else {
			durationText = fmt.Sprintf("%d секунд", number)
		}
	case "m":
		if number > 1440 { // Max 24 hours in minutes
			return 0, "", fmt.Errorf("too many minutes")
		}
		duration = time.Duration(number) * time.Minute
		if number == 1 {
			durationText = "1 минуту"
		} else if number < 5 {
			durationText = fmt.Sprintf("%d минуты", number)
		} else {
			durationText = fmt.Sprintf("%d минут", number)
		}
	default:
		return 0, "", fmt.Errorf("unsupported unit: %s", unit)
	}

	return duration, durationText, nil
}
