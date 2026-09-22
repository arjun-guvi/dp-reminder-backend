package payment

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusCancelled = "cancelled"
)

// Payment is a scheduled payment and its notification recipients.
type Payment struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title                string             `json:"title" bson:"title"`
	Description          string             `json:"description,omitempty" bson:"description,omitempty"`
	Amount               float64            `json:"amount" bson:"amount"`
	Currency             string             `json:"currency" bson:"currency"`
	PaymentType          string             `json:"payment_type" bson:"payment_type"`
	DownPayment          float64            `json:"down_payment" bson:"down_payment"`
	DueDate              time.Time          `json:"due_date" bson:"due_date"`
	RecipientName        string             `json:"recipient_name" bson:"recipient_name"`
	RecipientEmail       string             `json:"recipient_email,omitempty" bson:"recipient_email,omitempty"`
	TelegramChatID       string             `json:"telegram_chat_id,omitempty" bson:"telegram_chat_id,omitempty"`
	RecipientEmails      []string           `json:"recipient_emails,omitempty" bson:"recipient_emails,omitempty"`
	TelegramChatIDs      []string           `json:"telegram_chat_ids,omitempty" bson:"telegram_chat_ids,omitempty"`
	NotificationChannels []string           `json:"notification_channels" bson:"notification_channels"`
	Status               string             `json:"status" bson:"status"`
	NotificationSentAt   *time.Time         `json:"notification_sent_at,omitempty" bson:"notification_sent_at,omitempty"`
	CreatedAt            time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" bson:"updated_at"`
	Deleted              bool
	EmiMonths            float64 `json:"emi_months"`
}

type CreateRequest struct {
	Title                string    `json:"title" binding:"required"`
	Description          string    `json:"description"`
	Amount               float64   `json:"amount" binding:"required,gt=0"`
	Currency             string    `json:"currency" binding:"required,len=3"`
	PaymentType          string    `json:"payment_type" binding:"required,oneof=one_time emi"`
	DownPayment          float64   `json:"down_payment" binding:"gte=0"`
	DueDate              time.Time `json:"due_date" binding:"required"`
	RecipientName        string    `json:"recipient_name" binding:"required"`
	RecipientEmail       string    `json:"recipient_email" binding:"omitempty,email"`
	TelegramChatID       string    `json:"telegram_chat_id"`
	RecipientEmails      []string  `json:"recipient_emails"`
	TelegramChatIDs      []string  `json:"telegram_chat_ids"`
	NotificationChannels []string  `json:"notification_channels"`
	EmiMonths            float64   `json:"emi_months"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending paid cancelled"`
}
