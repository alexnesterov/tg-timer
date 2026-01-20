package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

// Client represents Telegram Bot API client interface
type Client interface {
	GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error)
	SendMessage(ctx context.Context, chatID int64, text string) error
	SetWebhook(ctx context.Context, webhookURL string) error
	DeleteWebhook(ctx context.Context) error
}

// HTTPClient represents Telegram Bot API client
type HTTPClient struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewClient creates new Telegram client
func NewClient(token string) Client {
	return &HTTPClient{
		token:   token,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s/", token),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetUpdates retrieves updates from Telegram (long polling)
func (tc *HTTPClient) GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("timeout", fmt.Sprintf("%d", timeout))

	url := tc.baseURL + "getUpdates?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OK          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("API error: %s", result.Description)
	}

	return result.Result, nil
}

// SendMessage sends message to chat
func (tc *HTTPClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	return tc.sendMessageWithRetry(ctx, chatID, text, 3)
}

// sendMessageWithRetry sends message with exponential backoff retry
func (tc *HTTPClient) sendMessageWithRetry(ctx context.Context, chatID int64, text string, maxRetries int) error {
	req := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := tc.sendSingleMessage(ctx, req)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Failed to send message (attempt %d/%d): %v", attempt+1, maxRetries, err)
	}

	return fmt.Errorf("failed to send message after %d attempts: %w", maxRetries, lastErr)
}

// sendSingleMessage sends single message without retry
func (tc *HTTPClient) sendSingleMessage(ctx context.Context, req SendMessageRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := tc.baseURL + "sendMessage"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := tc.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("API error: %s", result.Description)
	}

	return nil
}
