package notification

import (
	"fmt"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/response"
	"github.com/ares/dp-vc-webApp/models/payment"
	paymentservice "github.com/ares/dp-vc-webApp/services/payments"
	"github.com/gin-gonic/gin"
)

// Controller handles notification-related requests
type Controller struct {
	paymentService *paymentservice.Service
}

// New creates a new notification controller
func New() *Controller {
	return &Controller{
		paymentService: paymentservice.New(env.GetConfig()),
	}
}

// TestWorkerRequest is the request body for testing worker
type TestWorkerRequest struct {
	// Force forces the notification even if already sent
	Force bool `json:"force"`
	// DryRun only simulates without sending
	DryRun bool `json:"dry_run"`
}

// TestWorkerResponse is the response for worker test
type TestWorkerResponse struct {
	Success        bool                  `json:"success"`
	Message        string                `json:"message"`
	PaymentsFound  int                   `json:"payments_found"`
	Notifications  SentNotifications     `json:"notifications"`
	Duration        int64                 `json:"duration_ms"`
	Details        []PaymentNotification `json:"details,omitempty"`
	ReminderDays    int                  `json:"reminder_days"`
	WindowEnd      string                `json:"window_end"`
}

// SentNotifications counts sent notifications
type SentNotifications struct {
	Email    int `json:"email"`
	Telegram int `json:"telegram"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// PaymentNotification represents a single notification result
type PaymentNotification struct {
	PaymentID     string   `json:"payment_id"`
	Title         string   `json:"title"`
	Amount        string   `json:"amount"`
	DueDate       string   `json:"due_date"`
	Recipient     string   `json:"recipient_name"`
	Channels      []string `json:"channels"`
	EmailSent     bool     `json:"email_sent"`
	TelegramSent  bool     `json:"telegram_sent"`
	AlreadySent   bool     `json:"already_sent"`
	Error         string   `json:"error,omitempty"`
}

// TestNotifications manually triggers the notification check (for testing)
// POST /notifications/test
func (ctrl *Controller) TestNotifications(c *gin.Context) {
	startTime := time.Now()

	var req TestWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to dry-run if no body provided
		req.Force = false
		req.DryRun = true
	}

	config := env.GetConfig()
	reminderDays := config.PaymentReminderDays
	window := time.Now().UTC().AddDate(0, 0, reminderDays)

	// Get pending payments
	payments, err := ctrl.paymentService.List(c.Request.Context(), payment.StatusPending)
	if err != nil {
		response.Error(c, 500, "Failed to fetch payments: "+err.Error())
		return
	}

	result := TestWorkerResponse{
		Success:       true,
		ReminderDays:  reminderDays,
		WindowEnd:     window.Format(time.RFC3339),
		Notifications: SentNotifications{},
		Details:       []PaymentNotification{},
	}

	// Filter payments that need notification
	for _, p := range payments {
		detail := PaymentNotification{
			PaymentID: p.ID.Hex(),
			Title:     p.Title,
			Amount:    fmt.Sprintf("%s %.2f", p.Currency, p.Amount),
			DueDate:   p.DueDate.Format(time.RFC3339),
			Recipient: p.RecipientName,
			Channels:  p.NotificationChannels,
		}

		// Check if payment is within notification window
		isWithinWindow := p.DueDate.Before(window) || p.DueDate.Equal(window)
		alreadyNotified := p.NotificationSentAt != nil

		if !isWithinWindow {
			// Skip - not due yet
			result.Notifications.Skipped++
			detail.AlreadySent = false
		} else if alreadyNotified && !req.Force {
			// Skip - already notified (unless forced)
			result.Notifications.Skipped++
			detail.AlreadySent = true
			detail.Error = "Already notified"
		} else {
			// Would send notification
			if req.DryRun {
				// In dry-run mode, just count
				for _, channel := range p.NotificationChannels {
					switch channel {
					case "email":
						result.Notifications.Email++
						detail.EmailSent = true
					case "telegram":
						result.Notifications.Telegram++
						detail.TelegramSent = true
					}
				}
			} else {
				// Actually send notification
				// Note: This would call the actual send logic
				// For safety, we keep it as dry-run by default
				result.Notifications.Failed++
				detail.Error = "Use POST /notifications/trigger to actually send"
			}
		}

		result.Details = append(result.Details, detail)
	}

	result.PaymentsFound = len(payments)
	result.Duration = time.Since(startTime).Milliseconds()

	if req.DryRun {
		result.Message = "Dry-run completed. Use 'dry_run: false' to see actual results."
	} else {
		result.Message = fmt.Sprintf("Found %d payments, %d need notification", 
			result.PaymentsFound, result.Notifications.Email+result.Notifications.Telegram)
	}

	response.Success(c, result)
}

// TriggerNotifications triggers the actual notification send (admin only)
// POST /notifications/trigger
func (ctrl *Controller) TriggerNotifications(c *gin.Context) {
	startTime := time.Now()

	err := ctrl.paymentService.SendUpcomingNotifications(c.Request.Context(), time.Now().UTC())
	if err != nil {
		response.Error(c, 500, "Failed to send notifications: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success":     true,
		"message":     "Notifications triggered successfully",
		"duration_ms": time.Since(startTime).Milliseconds(),
	})
}

// GetWorkerStatus returns the current worker status and configuration
// GET /notifications/status
func (ctrl *Controller) GetWorkerStatus(c *gin.Context) {
	config := env.GetConfig()

	emailConfigured := config.SMTPHost != "" || (config.GmailUsername != "" && config.GmailAppPassword != "")
	telegramConfigured := config.TelegramBotToken != ""

	response.Success(c, gin.H{
		"worker": gin.H{
			"notification_interval_minutes": config.NotificationIntervalMinutes,
			"payment_reminder_days":         config.PaymentReminderDays,
			"email_configured":              emailConfigured,
			"telegram_configured":           telegramConfigured,
			"next_run_in_minutes":           config.NotificationIntervalMinutes,
		},
		"config": gin.H{
			"smtp_host":       maskSecret(config.SMTPHost),
			"smtp_port":       config.SMTPPort,
			"smtp_username":   maskSecret(config.SMTPUsername),
			"smtp_from":       maskSecret(config.SMTPFrom),
			"gmail_username":  maskSecret(config.GmailUsername),
			"gmail_from":      maskSecret(config.GmailFrom),
			"telegram_token":  maskSecret(config.TelegramBotToken),
		},
		"status": gin.H{
			"healthy":         true,
			"last_check":      time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// ClearLock clears the notification lock (for testing)
// DELETE /notifications/lock
func (ctrl *Controller) ClearLock(c *gin.Context) {
	config := env.GetConfig()
	// This would need Redis client access
	// For now, return instructions
	response.Success(c, gin.H{
		"success": true,
		"message": "Use redis-cli to clear lock manually",
		"command": fmt.Sprintf("redis-cli DEL %s:lock:payment-notifications", config.RedisPrefix),
	})
}

// maskSecret masks sensitive information
func maskSecret(s string) string {
	if len(s) == 0 {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
