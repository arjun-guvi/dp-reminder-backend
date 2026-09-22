package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/mongo"
	"github.com/ares/dp-vc-webApp/configs/types"
	"github.com/ares/dp-vc-webApp/models/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const userCollection = "users"

type Service struct{ config *env.Config }

type tokenClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Expires int64  `json:"exp"`
}

func New(config *env.Config) *Service { return &Service{config: config} }

func (s *Service) Signup(ctx context.Context, request user.SignupRequest) (*user.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	var existing user.User
	if err := mongo.FindOne(ctx, userCollection, bson.M{"email": email}, &existing); err == nil {
		return nil, errors.New("email is already registered")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	now := time.Now().UTC()
	newUser := user.User{ID: primitive.NewObjectID(), Name: strings.TrimSpace(request.Name), Email: email, PasswordHash: string(passwordHash), CreatedAt: now, UpdatedAt: now}
	if _, err := mongo.InsertOne(ctx, userCollection, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	token, err := s.issueToken(newUser)
	if err != nil {
		return nil, err
	}
	newUser.PasswordHash = ""
	return &user.AuthResponse{Token: token, User: newUser}, nil
}

func (s *Service) Login(ctx context.Context, request user.LoginRequest) (*user.AuthResponse, error) {
	var existing user.User
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if err := mongo.FindOne(ctx, userCollection, bson.M{"email": email}, &existing); err != nil {
		return nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(request.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}
	token, err := s.issueToken(existing)
	if err != nil {
		return nil, err
	}
	existing.PasswordHash = ""
	return &user.AuthResponse{Token: token, User: existing}, nil
}

func (s *Service) ValidateToken(token string) (*types.User, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, []byte(s.config.AuthSecret))
	_, _ = mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, errors.New("invalid token")
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Subject == "" || claims.Expires < time.Now().Unix() {
		return nil, errors.New("invalid or expired token")
	}
	return &types.User{ID: claims.Subject, Email: claims.Email, Permissions: []string{"payments.view", "payments.manage"}, Metadata: map[string]interface{}{}}, nil
}

func (s *Service) issueToken(account user.User) (string, error) {
	claims := tokenClaims{Subject: account.ID.Hex(), Email: account.Email, Expires: time.Now().Add(time.Duration(s.config.AuthTokenHours) * time.Hour).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(s.config.AuthSecret))
	_, _ = mac.Write(payload)
	return encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
