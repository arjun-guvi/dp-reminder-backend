package payments

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/mongo"
	"github.com/ares/dp-vc-webApp/models/payment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "payments"

type Service struct{ config *env.Config }

func New(config *env.Config) *Service { return &Service{config: config} }

func (s *Service) Create(ctx context.Context, req payment.CreateRequest) (*payment.Payment, error) {
	channels, err := normalizeChannels(req.NotificationChannels, req.RecipientEmail, req.TelegramChatID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item := &payment.Payment{
		ID: primitive.NewObjectID(), Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description),
		Amount: req.Amount, Currency: strings.ToUpper(strings.TrimSpace(req.Currency)), DueDate: req.DueDate.UTC(),
		RecipientName: strings.TrimSpace(req.RecipientName), RecipientEmail: strings.TrimSpace(req.RecipientEmail),
		TelegramChatID: strings.TrimSpace(req.TelegramChatID), NotificationChannels: channels,
		Status: payment.StatusPending, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := mongo.InsertOne(ctx, collectionName, item); err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, status string) ([]payment.Payment, error) {
	filter := bson.M{}
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
	if err := mongo.FindOne(ctx, collectionName, bson.M{"_id": objectID}, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid payment id")
	}
	modified, err := mongo.UpdateOne(ctx, collectionName, bson.M{"_id": objectID}, bson.M{"status": status, "updated_at": time.Now().UTC()})
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
	deleted, err := mongo.DeleteOne(ctx, collectionName, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if deleted == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func (s *Service) SendUpcomingNotifications(ctx context.Context, now time.Time) error {
	window := now.UTC().AddDate(0, 0, s.config.PaymentReminderDays)
	filter := bson.M{"status": payment.StatusPending, "notification_sent_at": bson.M{"$exists": false}, "due_date": bson.M{"$gte": now.UTC(), "$lte": window}}
	items := make([]payment.Payment, 0)
	if err := mongo.FindMany(ctx, collectionName, filter, &items); err != nil {
		return err
	}
	for i := range items {
		if err := s.notify(ctx, &items[i]); err != nil {
			fmt.Printf("[Worker] payment %s notification failed: %v\n", items[i].ID.Hex(), err)
			continue
		}
		sentAt := time.Now().UTC()
		_, _ = mongo.UpdateOne(ctx, collectionName, bson.M{"_id": items[i].ID, "notification_sent_at": bson.M{"$exists": false}}, bson.M{"notification_sent_at": sentAt, "updated_at": sentAt})
	}
	return nil
}

func (s *Service) notify(ctx context.Context, item *payment.Payment) error {
	message := fmt.Sprintf("Payment reminder: %s, %s %.2f is due on %s.", item.Title, item.Currency, item.Amount, item.DueDate.Format(time.RFC1123))
	for _, channel := range item.NotificationChannels {
		switch channel {
		case "email":
			if err := s.sendEmail(item.RecipientEmail, "Upcoming payment reminder", message); err != nil {
				return err
			}
		case "telegram":
			if err := s.sendTelegram(ctx, item.TelegramChatID, message); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) sendEmail(to, subject, body string) error {
	if s.config.SMTPHost == "" {
		return errors.New("SMTP_HOST is not configured")
	}
	address := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	message := "To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + body
	var auth smtp.Auth
	if s.config.SMTPUsername != "" {
		auth = smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	}
	return smtp.SendMail(address, auth, s.config.SMTPFrom, []string{to}, []byte(message))
}

func (s *Service) sendTelegram(ctx context.Context, chatID, message string) error {
	if s.config.TelegramBotToken == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is not configured")
	}
	endpoint := "https://api.telegram.org/bot" + s.config.TelegramBotToken + "/sendMessage"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(url.Values{"chat_id": {chatID}, "text": {message}}.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram returned status %s", response.Status)
	}
	return nil
}

func normalizeChannels(channels []string, email, chatID string) ([]string, error) {
	if len(channels) == 0 {
		if email != "" {
			channels = append(channels, "email")
		}
		if chatID != "" {
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
		if channel == "email" && email == "" {
			return nil, errors.New("recipient_email is required for email notifications")
		}
		if channel == "telegram" && chatID == "" {
			return nil, errors.New("telegram_chat_id is required for Telegram notifications")
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
