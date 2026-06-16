package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"wallet-api/internal/model"
	"wallet-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

type AuthResult struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func NewAuthService(users repository.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), jwtTTL: jwtTTL}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := validateCredentials(email, password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}
	token, err := s.makeJWT(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	token, err := s.makeJWT(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: *user}, nil
}

func (s *AuthService) makeJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(s.jwtTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	result, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return result, nil
}

func validateCredentials(email, password string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidInput
	}
	if len(password) < 8 {
		return ErrInvalidInput
	}
	return nil
}
