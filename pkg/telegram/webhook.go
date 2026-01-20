package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SetWebhook sets webhook for Telegram bot
func (tc *HTTPClient) SetWebhook(ctx context.Context, webhookURL string) error {
	params := url.Values{}
	params.Add("url", webhookURL)

	url := tc.baseURL + "setWebhook?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
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

// DeleteWebhook deletes webhook
func (tc *HTTPClient) DeleteWebhook(ctx context.Context) error {
	url := tc.baseURL + "deleteWebhook"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
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
