package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles Telegram API communication
type Client struct {
	botToken string
	httpClient *http.Client
}

// New creates a new Telegram client
func New(botToken string) *Client {
	return &Client{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessage sends a text message to a Telegram chat
func (c *Client) SendMessage(chatID string, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	
	if !result["ok"].(bool) {
		return fmt.Errorf("telegram API error: %v", result["description"])
	}
	
	return nil
}

// SendHTMLMessage sends an HTML-formatted message
func (c *Client) SendHTMLMessage(chatID, subject, body string) error {
	message := fmt.Sprintf("<b>%s</b>\n\n%s", subject, body)
	return c.SendMessage(chatID, message)
}

// GetBotInfo returns information about the bot
func (c *Client) GetBotInfo() (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", c.botToken)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get bot info: %w", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return result, nil
}

// GetUpdates fetches bot updates (useful to get chat IDs)
func (c *Client) GetUpdates(offset int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d", c.botToken, offset)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	if updates, ok := result["result"].([]interface{}); ok {
		var updatesList []map[string]interface{}
		for _, u := range updates {
			if update, ok := u.(map[string]interface{}); ok {
				updatesList = append(updatesList, update)
			}
		}
		return updatesList, nil
	}
	
	return nil, nil
}
