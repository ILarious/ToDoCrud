package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type userRepository interface {
	Create(ctx context.Context, user model.User) (int64, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
}

type authService struct {
	users      userRepository
	tokenKey   []byte
	tokenTTL   time.Duration
	hashPepper string
}

func NewAuthService(users userRepository, tokenKey string, tokenTTL time.Duration) *authService {
	key := strings.TrimSpace(tokenKey)
	if key == "" {
		key = "dev-insecure-signing-key"
	}

	return &authService{
		users:      users,
		tokenKey:   []byte(key),
		tokenTTL:   tokenTTL,
		hashPepper: key,
	}
}

func (s *authService) SignUp(ctx context.Context, name, username, password string) (string, error) {
	name = strings.TrimSpace(name)
	username = strings.TrimSpace(username)
	if err := validateSignUpInput(name, username, password); err != nil {
		return "", err
	}

	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return "", domainerr.ErrUsernameTaken
	} else if !errors.Is(err, domainerr.ErrUserNotFound) {
		return "", err
	}

	hash, err := s.hashPassword(password)
	if err != nil {
		return "", err
	}

	id, err := s.users.Create(ctx, model.User{
		Name:     name,
		Username: username,
		Password: hash,
	})
	if err != nil {
		if errors.Is(err, domainerr.ErrUserExists) {
			return "", domainerr.ErrUsernameTaken
		}
		return "", err
	}

	return s.generateToken(id)
}

func (s *authService) SignIn(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", domainerr.ErrInvalidInput
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domainerr.ErrUserNotFound) {
			return "", domainerr.ErrInvalidCredentials
		}
		return "", err
	}

	if !s.comparePassword(user.Password, password) {
		return "", domainerr.ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

func (s *authService) ParseToken(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, domainerr.ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, domainerr.ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, domainerr.ErrInvalidToken
	}

	signed := sign(payload, s.tokenKey)
	if !hmac.Equal(sig, signed) {
		return 0, domainerr.ErrInvalidToken
	}

	body := strings.Split(string(payload), ":")
	if len(body) != 2 {
		return 0, domainerr.ErrInvalidToken
	}

	uid, err := strconv.ParseInt(body[0], 10, 64)
	if err != nil || uid <= 0 {
		return 0, domainerr.ErrInvalidToken
	}
	exp, err := strconv.ParseInt(body[1], 10, 64)
	if err != nil {
		return 0, domainerr.ErrInvalidToken
	}
	if time.Now().Unix() > exp {
		return 0, domainerr.ErrInvalidToken
	}

	return uid, nil
}

func validateSignUpInput(name, username, password string) error {
	if len(name) < 2 || len(username) < 3 || len(password) < 8 {
		return domainerr.ErrInvalidInput
	}
	return nil
}

func (s *authService) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	h := sha256.Sum256([]byte(hex.EncodeToString(salt) + ":" + password + ":" + s.hashPepper))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(h[:]), nil
}

func (s *authService) comparePassword(stored, plain string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}

	h := sha256.Sum256([]byte(parts[0] + ":" + plain + ":" + s.hashPepper))
	expected := hex.EncodeToString(h[:])
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

func (s *authService) generateToken(userID int64) (string, error) {
	exp := time.Now().Add(s.tokenTTL).Unix()
	payload := []byte(fmt.Sprintf("%d:%d", userID, exp))
	sig := sign(payload, s.tokenKey)

	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func sign(payload, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return mac.Sum(nil)
}
