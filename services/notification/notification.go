package notification

import (
	"context"
	"fmt"
	"sync"

	"github.com/ares/dp-vc-webApp/configs/email"
	"github.com/ares/dp-vc-webApp/configs/telegram"
	"github.com/ares/dp-vc-webApp/models/notification"
)

// Service handles sending notifications via multiple channels
type Service struct {
	emailClient    *email.Client
	telegramClient *telegram.Client
	emailConfig    notification.Config
	telegramConfig notification.Config
}

// New creates a new notification service
func New(
	emailClient *email.Client,
	telegramClient *telegram.Client,
	emailConfig, telegramConfig notification.Config,
) *Service {
	return &Service{
		emailClient:    emailClient,
		telegramClient: telegramClient,
		emailConfig:    emailConfig,
		telegramConfig: telegramConfig,
	}
}

// Send sends a notification through configured channels
func (s *Service) Send(ctx context.Context, notif notification.Request) (*notification.Response, error) {
	var results []notification.Result
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Use default recipients if not specified
	emailRecipients := notif.EmailRecipients
	if len(emailRecipients) == 0 {
		emailRecipients = s.emailConfig.EmailRecipients
	}

	telegramChats := notif.TelegramChats
	if len(telegramChats) == 0 {
		telegramChats = s.telegramConfig.TelegramChats
	}

	// Prepare message
	message := notif.Message
	if notif.HTMLMessage != "" {
		message = notif.HTMLMessage
	}

	// Send email notifications
	if (notif.Type == notification.TypeEmail || notif.Type == notification.TypeBoth) && len(emailRecipients) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if s.emailClient == nil {
				mu.Lock()
				results = append(results, notification.Result{
					Type:    "email",
					Target:  "N/A",
					Success: false,
					Error:   "email client not configured",
				})
				mu.Unlock()
				return
			}

			for _, recipient := range emailRecipients {
				err := s.emailClient.Send(email.Message{
					To:      []string{recipient},
					Subject: notif.Subject,
					Body:    notif.HTMLMessage,
					IsHTML:  notif.HTMLMessage != "",
				})

				mu.Lock()
				results = append(results, notification.Result{
					Type:    "email",
					Target:  recipient,
					Success: err == nil,
					Error:   formatError(err),
				})
				mu.Unlock()
			}
		}()
	}

	// Send Telegram notifications
	if (notif.Type == notification.TypeTelegram || notif.Type == notification.TypeBoth) && len(telegramChats) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if s.telegramClient == nil {
				mu.Lock()
				results = append(results, notification.Result{
					Type:    "telegram",
					Target:  "N/A",
					Success: false,
					Error:   "telegram client not configured",
				})
				mu.Unlock()
				return
			}

			for _, chatID := range telegramChats {
				err := s.telegramClient.SendHTMLMessage(chatID, notif.Subject, message)

				mu.Lock()
				results = append(results, notification.Result{
					Type:    "telegram",
					Target:  chatID,
					Success: err == nil,
					Error:   formatError(err),
				})
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Aggregate errors
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	return &notification.Response{
		Success: successCount > 0,
		Results: results,
	}, nil
}

// SendEmail sends only email notification
func (s *Service) SendEmail(ctx context.Context, recipients []string, subject, body string) error {
	if s.emailClient == nil {
		return fmt.Errorf("email client not configured")
	}

	return s.emailClient.SendHTML(recipients, subject, body)
}

// SendTelegram sends only Telegram notification
func (s *Service) SendTelegram(ctx context.Context, chatIDs []string, subject, message string) error {
	if s.telegramClient == nil {
		return fmt.Errorf("telegram client not configured")
	}

	for _, chatID := range chatIDs {
		if err := s.telegramClient.SendHTMLMessage(chatID, subject, message); err != nil {
			return err
		}
	}
	return nil
}

// SendAlert sends a critical alert through all channels
func (s *Service) SendAlert(ctx context.Context, subject, message string) error {
	_, err := s.Send(ctx, notification.Request{
		Type:        notification.TypeBoth,
		Subject:     "🔴 ALERT: " + subject,
		Message:     message,
		HTMLMessage: message,
		Priority:    notification.PriorityCritical,
	})
	return err
}

// SendPaymentReminder sends a payment reminder notification
func (s *Service) SendPaymentReminder(ctx context.Context, recipientEmail, recipientName, paymentDetails string) error {
	subject := "Payment Reminder"
	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #4F46E5;">Payment Reminder</h2>
			<p>Dear %s,</p>
			<p>This is a friendly reminder about your upcoming payment.</p>
			<div style="background: #f5f5f5; padding: 15px; border-radius: 8px; margin: 20px 0;">
				%s
			</div>
			<p>Please ensure timely payment to avoid any service interruption.</p>
			<p style="color: #666; font-size: 14px;">Best regards,<br>DP VC Team</p>
		</div>
	`, recipientName, paymentDetails)

	return s.SendEmail(ctx, []string{recipientEmail}, subject, htmlBody)
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// IsConfigured returns whether the notification service is properly configured
func (s *Service) IsConfigured() (emailConfigured, telegramConfigured bool) {
	emailConfigured = s.emailClient != nil
	telegramConfigured = s.telegramClient != nil
	return
}

// GetStatus returns the current status of notification channels
func (s *Service) GetStatus() map[string]interface{} {
	emailCfg, tgCfg := s.IsConfigured()
	return map[string]interface{}{
		"email": map[string]interface{}{
			"configured": emailCfg,
			"recipients": len(s.emailConfig.EmailRecipients),
		},
		"telegram": map[string]interface{}{
			"configured": tgCfg,
			"chats":      len(s.telegramConfig.TelegramChats),
		},
	}
}
