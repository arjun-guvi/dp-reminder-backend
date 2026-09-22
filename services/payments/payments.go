package payments

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/logger"
	"github.com/ares/dp-vc-webApp/configs/mongo"
	"github.com/ares/dp-vc-webApp/models/payment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "payments"

var log = logger.Payment()

type Service struct{ config *env.Config }

func New(config *env.Config) *Service { return &Service{config: config} }

func (s *Service) Create(ctx context.Context, req payment.CreateRequest) (*payment.Payment, error) {
	emails := append([]string{}, req.RecipientEmails...)
	if req.RecipientEmail != "" {
		emails = append(emails, req.RecipientEmail)
	}
	chatIDs := append([]string{}, req.TelegramChatIDs...)
	if req.TelegramChatID != "" {
		chatIDs = append(chatIDs, req.TelegramChatID)
	}
	emails, chatIDs, err := normalizeRecipients(emails, chatIDs)
	if err != nil {
		return nil, err
	}
	channels, err := normalizeChannels(req.NotificationChannels, len(emails), len(chatIDs))
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item := &payment.Payment{
		ID: primitive.NewObjectID(), Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description),
		Amount: req.Amount, Currency: strings.ToUpper(strings.TrimSpace(req.Currency)), DueDate: req.DueDate.UTC(),
		RecipientName: strings.TrimSpace(req.RecipientName), NotificationChannels: channels,
		RecipientEmails: emails, TelegramChatIDs: chatIDs,
		Status: payment.StatusPending, CreatedAt: now, UpdatedAt: now,
		PaymentType: req.PaymentType,
		DownPayment: req.DownPayment,
		EmiMonths:   req.EmiMonths,
	}
	if len(emails) > 0 {
		item.RecipientEmail = emails[0]
	}
	if len(chatIDs) > 0 {
		item.TelegramChatID = chatIDs[0]
	}
	item.Deleted = false
	if _, err := mongo.InsertOne(ctx, collectionName, item); err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, status string) ([]payment.Payment, error) {
	filter := bson.M{
		"deleted": false,
	}
	if status != "" {
		filter["status"] = status
	}
	items := make([]payment.Payment, 0)
	err := mongo.FindMany(ctx, collectionName, filter, &items, options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}}))
	return items, err
}

func (s *Service) Get(ctx context.Context, id string) (*payment.Payment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid payment id")
	}
	var item payment.Payment
	if err := mongo.FindOne(ctx, collectionName, bson.M{"_id": objectID, "deleted": false}, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid payment id")
	}
	modified, err := mongo.UpdateOne(ctx, collectionName, bson.M{"_id": objectID, "deleted": false}, bson.M{"status": status, "updated_at": time.Now().UTC()})
	if err != nil {
		return err
	}
	if modified == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid payment id")
	}
	modified, err := mongo.UpdateOne(ctx, collectionName, bson.M{"_id": objectID}, bson.M{"deleted": true, "updated_at": time.Now().UTC()})
	if err != nil {
		return err
	}
	if modified == 0 {
		return errors.New("payment not found")
	}
	return nil
}

// SendUpcomingNotifications finds payments due within the reminder window and sends notifications
func (s *Service) SendUpcomingNotifications(ctx context.Context, now time.Time) error {
	// Calculate the window: payments due within PaymentReminderDays from now
	window := now.UTC().AddDate(0, 0, s.config.PaymentReminderDays)

	log.Info("Scanning for upcoming payments", map[string]interface{}{
		"window_start":  now.UTC().Format(time.RFC3339),
		"window_end":    window.Format(time.RFC3339),
		"reminder_days": s.config.PaymentReminderDays,
	})

	// Find pending payments where:
	// - status is pending
	// - notification hasn't been sent yet
	// - due_date is within the window
	// - not deleted
	filter := bson.M{
		"status":               payment.StatusPending,
		"notification_sent_at": bson.M{"$exists": false},
		"due_date":             bson.M{"$lte": window},
		"deleted":              false,
	}

	items := make([]payment.Payment, 0)
	if err := mongo.FindMany(ctx, collectionName, filter, &items); err != nil {
		log.Error("Failed to query payments", err, nil)
		return fmt.Errorf("query payments: %w", err)
	}

	log.Info("Found payments to notify", map[string]interface{}{
		"count": len(items),
	})

	if len(items) == 0 {
		return nil
	}

	// Track results
	successCount := 0
	failureCount := 0

	for i := range items {
		paymentID := items[i].ID.Hex()
		log.Info("Processing payment notification", map[string]interface{}{
			"payment_id": paymentID,
			"title":      items[i].Title,
			"amount":     items[i].Amount,
			"currency":   items[i].Currency,
			"due_date":   items[i].DueDate.Format(time.RFC3339),
			"recipient":  items[i].RecipientName,
			"channels":   items[i].NotificationChannels,
		})

		if err := s.notify(ctx, &items[i]); err != nil {
			log.Error("Failed to send notification for payment", err, map[string]interface{}{
				"payment_id": paymentID,
			})
			failureCount++
			continue
		}

		// Mark notification as sent
		sentAt := time.Now().UTC()
		updatedCount, err := mongo.UpdateOne(
			ctx,
			collectionName,
			bson.M{
				"_id":                  items[i].ID,
				"notification_sent_at": bson.M{"$exists": false},
			},
			bson.M{
				"notification_sent_at": sentAt,
				"updated_at":           sentAt,
			},
		)
		if err != nil {
			log.Error("Failed to mark notification as sent", err, map[string]interface{}{
				"payment_id": paymentID,
			})
		} else if updatedCount > 0 {
			successCount++
			log.Info("Notification sent and marked", map[string]interface{}{
				"payment_id": paymentID,
				"sent_at":    sentAt.Format(time.RFC3339),
			})
		}

		// Prevent rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Info("Notification batch completed", map[string]interface{}{
		"total_processed": len(items),
		"success":         successCount,
		"failed":          failureCount,
	})

	return nil
}

func (s *Service) notify(ctx context.Context, item *payment.Payment) error {
	message := fmt.Sprintf("Payment reminder: %s, %s %.2f is due on %s.", item.Title, item.Currency, item.Amount, item.DueDate.Format(time.RFC1123))

	emailSubject := "Upcoming Payment Reminder: " + item.Title

	emails := item.RecipientEmails
	if len(emails) == 0 && item.RecipientEmail != "" {
		emails = []string{item.RecipientEmail}
	}
	chatIDs := item.TelegramChatIDs
	if len(chatIDs) == 0 && item.TelegramChatID != "" {
		chatIDs = []string{item.TelegramChatID}
	}

	var errors []error
	for _, channel := range item.NotificationChannels {
		switch channel {
		case "email":
			if len(emails) == 0 {
				log.Warn("No email recipients configured for payment", map[string]interface{}{
					"payment_id": item.ID.Hex(),
				})
				continue
			}

			for _, emailAddr := range emails {
				err := s.sendEmail(emailAddr, emailSubject, message)
				if err != nil {
					errors = append(errors, fmt.Errorf("email to %s: %w", emailAddr, err))
					log.Error("Failed to send email notification", err, map[string]interface{}{
						"payment_id": item.ID.Hex(),
						"recipient":  emailAddr,
					})
				} else {
					log.Info("Email notification sent", map[string]interface{}{
						"payment_id": item.ID.Hex(),
						"recipient":  emailAddr,
					})
				}
			}

		case "telegram":
			if s.config.TelegramBotToken == "" {
				log.Warn("Telegram bot token not configured", map[string]interface{}{
					"payment_id": item.ID.Hex(),
				})
				continue
			}

			if len(chatIDs) == 0 {
				log.Warn("No Telegram chat IDs configured for payment", map[string]interface{}{
					"payment_id": item.ID.Hex(),
				})
				continue
			}

			for _, chatID := range chatIDs {
				err := s.sendTelegram(ctx, chatID, message)
				if err != nil {
					errors = append(errors, fmt.Errorf("telegram to %s: %w", chatID, err))
					log.Error("Failed to send Telegram notification", err, map[string]interface{}{
						"payment_id": item.ID.Hex(),
						"chat_id":    chatID,
					})
				} else {
					log.Info("Telegram notification sent", map[string]interface{}{
						"payment_id": item.ID.Hex(),
						"chat_id":    chatID,
					})
				}
			}
		}
	}

	// Return first error if all channels failed
	if len(errors) > 0 && len(errors) == len(item.NotificationChannels) {
		return errors[0]
	}

	return nil
}

func (s *Service) sendEmail(to, subject, body string) error {
	host := s.config.SMTPHost
	port := s.config.SMTPPort
	username := s.config.SMTPUsername
	password := s.config.SMTPPassword
	from := s.config.SMTPFrom

	// Use Gmail if configured
	if s.config.GmailUsername != "" && s.config.GmailAppPassword != "" {
		host = "smtp.gmail.com"
		port = 587
		username = s.config.GmailUsername
		password = s.config.GmailAppPassword
		from = s.config.GmailFrom
	}

	if host == "" || username == "" || password == "" || from == "" {
		return errors.New("email configuration incomplete - check SMTP or Gmail settings")
	}

	address := fmt.Sprintf("%s:%d", host, port)
	message := "To: " + to + "\r\nSubject: " + subject + "\r\nMIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n" + body

	auth := smtp.PlainAuth("", username, password, host)
	err := smtp.SendMail(address, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *Service) sendTelegram(ctx context.Context, chatID, message string) error {
	if s.config.TelegramBotToken == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is not configured")
	}

	endpoint := "https://api.telegram.org/bot" + s.config.TelegramBotToken + "/sendMessage"

	formData := url.Values{}
	formData.Set("chat_id", chatID)
	formData.Set("text", message)
	formData.Set("parse_mode", "HTML")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram API returned status %d", response.StatusCode)
	}

	return nil
}

func normalizeChannels(channels []string, emailCount, chatIDCount int) ([]string, error) {
	if len(channels) == 0 {
		if emailCount > 0 {
			channels = append(channels, "email")
		}
		if chatIDCount > 0 {
			channels = append(channels, "telegram")
		}
	}
	seen := make(map[string]bool)
	result := make([]string, 0, len(channels))
	for _, channel := range channels {
		channel = strings.ToLower(strings.TrimSpace(channel))
		if channel != "email" && channel != "telegram" {
			return nil, fmt.Errorf("unsupported notification channel: %s", channel)
		}
		if channel == "email" && emailCount == 0 {
			return nil, errors.New("at least one recipient email is required for email notifications")
		}
		if channel == "telegram" && chatIDCount == 0 {
			return nil, errors.New("at least one Telegram chat ID is required for Telegram notifications")
		}
		if !seen[channel] {
			result = append(result, channel)
			seen[channel] = true
		}
	}
	if len(result) == 0 {
		return nil, errors.New("at least one notification recipient is required")
	}
	return result, nil
}

func normalizeRecipients(emails, chatIDs []string) ([]string, []string, error) {
	uniqueEmails := make([]string, 0, len(emails))
	seenEmails := make(map[string]bool)
	for _, rawEmail := range emails {
		address, err := mail.ParseAddress(strings.TrimSpace(rawEmail))
		if err != nil || address.Address != strings.TrimSpace(rawEmail) {
			return nil, nil, fmt.Errorf("invalid recipient email: %s", rawEmail)
		}
		email := strings.ToLower(address.Address)
		if !seenEmails[email] {
			uniqueEmails = append(uniqueEmails, email)
			seenEmails[email] = true
		}
	}
	uniqueChatIDs := make([]string, 0, len(chatIDs))
	seenChatIDs := make(map[string]bool)
	for _, rawChatID := range chatIDs {
		chatID := strings.TrimSpace(rawChatID)
		if chatID == "" {
			return nil, nil, errors.New("Telegram chat IDs cannot be empty")
		}
		if !seenChatIDs[chatID] {
			uniqueChatIDs = append(uniqueChatIDs, chatID)
			seenChatIDs[chatID] = true
		}
	}
	return uniqueEmails, uniqueChatIDs, nil
}
