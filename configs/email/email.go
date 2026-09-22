package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// Config holds email configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Client handles email sending
type Client struct {
	config Config
}

// New creates a new email client
func New(config Config) *Client {
	return &Client{config: config}
}

// Message represents an email message
type Message struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// Send sends an email
func (c *Client) Send(msg Message) error {
	// Build email headers
	headers := make(map[string]string)
	headers["From"] = c.config.From
	
	if len(msg.To) > 0 {
		headers["To"] = strings.Join(msg.To, ", ")
	}
	
	headers["Subject"] = msg.Subject
	
	if msg.IsHTML {
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"
	}
	
	// Build message body
	var emailBody string
	for k, v := range headers {
		emailBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	emailBody += "\r\n" + msg.Body
	
	// Setup authentication
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	
	// Send email using TLS
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer client.Close()
	
	// Start TLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: c.config.Host}
		if err := client.StartTLS(config); err != nil {
			return fmt.Errorf("start TLS: %w", err)
		}
	}
	
	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	
	// Set sender
	if err := client.Mail(c.config.From); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}
	
	// Add recipients
	for _, to := range msg.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("add recipient %s: %w", to, err)
		}
	}
	
	// Get data writer
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("get data writer: %w", err)
	}
	
	// Write email content
	_, err = fmt.Fprint(w, emailBody)
	if err != nil {
		return fmt.Errorf("write email content: %w", err)
	}
	
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data writer: %w", err)
	}
	
	return nil
}

// SendSimple sends a simple text email
func (c *Client) SendSimple(to []string, subject, body string) error {
	return c.Send(Message{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

// SendHTML sends an HTML email
func (c *Client) SendHTML(to []string, subject, body string) error {
	return c.Send(Message{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}
