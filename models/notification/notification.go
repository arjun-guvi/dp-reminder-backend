package notification

import "time"

// Type represents notification type
type Type string

const (
	TypeEmail    Type = "email"
	TypeTelegram Type = "telegram"
	TypeBoth     Type = "both"
)

// Priority represents notification priority
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Notification represents a notification to be sent
type Notification struct {
	Type        Type              `json:"type" bson:"type"`
	EmailRecipients []string      `json:"email_recipients" bson:"email_recipients"`
	TelegramChats   []string      `json:"telegram_chats" bson:"telegram_chats"`
	Subject     string            `json:"subject" bson:"subject"`
	Message     string            `json:"message" bson:"message"`
	HTMLMessage string            `json:"html_message" bson:"html_message"`
	Priority    Priority          `json:"priority" bson:"priority"`
	Metadata    map[string]string `json:"metadata" bson:"metadata"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
}

// Request is the API request for sending a notification
type Request struct {
	Type            Type              `json:"type" binding:"required"`
	EmailRecipients []string          `json:"email_recipients"`
	TelegramChats   []string          `json:"telegram_chats"`
	Subject         string            `json:"subject"`
	Message         string            `json:"message" binding:"required"`
	HTMLMessage     string            `json:"html_message"`
	Priority        Priority          `json:"priority"`
	Metadata        map[string]string `json:"metadata"`
}

// Response is the API response for sending a notification
type Response struct {
	Success bool     `json:"success"`
	Results  []Result `json:"results"`
}

// Result represents the result of a single notification send
type Result struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Config holds notification destinations
type Config struct {
	EmailRecipients []string
	TelegramChats   []string
}
