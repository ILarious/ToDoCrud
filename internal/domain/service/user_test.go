package service

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerr "todo_crud/internal/domain/errors"
	"todo_crud/internal/domain/model"
)

type userRepoStub struct {
	createFn        func(ctx context.Context, user model.User) (int64, error)
	getByUsernameFn func(ctx context.Context, username string) (model.User, error)
}

func (s userRepoStub) Create(ctx context.Context, user model.User) (int64, error) {
	if s.createFn == nil {
		return 0, errors.New("createFn is nil")
	}
	return s.createFn(ctx, user)
}

func (s userRepoStub) GetByUsername(ctx context.Context, username string) (model.User, error) {
	if s.getByUsernameFn == nil {
		return model.User{}, domainerr.ErrUserNotFound
	}
	return s.getByUsernameFn(ctx, username)
}

func TestAuthServiceSignUp_Success(t *testing.T) {
	svc := NewAuthService(userRepoStub{
		createFn: func(_ context.Context, user model.User) (int64, error) {
			if user.Password == "" || user.Password == "password123" {
				t.Fatalf("expected hashed password, got %q", user.Password)
			}
			return 42, nil
		},
		getByUsernameFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{}, domainerr.ErrUserNotFound
		},
	}, "test-secret", time.Hour)

	token, err := svc.SignUp(context.Background(), "John", "john_doe", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	uid, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if uid != 42 {
		t.Fatalf("expected user id 42, got %d", uid)
	}
}

func TestAuthServiceSignUp_InvalidInput(t *testing.T) {
	svc := NewAuthService(userRepoStub{}, "test-secret", time.Hour)

	_, err := svc.SignUp(context.Background(), "A", "ab", "123")
	if !errors.Is(err, domainerr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAuthServiceSignUp_UsernameTaken(t *testing.T) {
	svc := NewAuthService(userRepoStub{
		getByUsernameFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{ID: 1, Username: "john"}, nil
		},
		createFn: func(_ context.Context, _ model.User) (int64, error) {
			return 0, nil
		},
	}, "test-secret", time.Hour)

	_, err := svc.SignUp(context.Background(), "John", "john", "password123")
	if !errors.Is(err, domainerr.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAuthServiceSignIn_InvalidCredentials(t *testing.T) {
	svc := NewAuthService(userRepoStub{
		getByUsernameFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{}, domainerr.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ model.User) (int64, error) {
			return 1, nil
		},
	}, "test-secret", time.Hour)

	_, err := svc.SignIn(context.Background(), "john", "password123")
	if !errors.Is(err, domainerr.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceParseToken_Expired(t *testing.T) {
	svc := NewAuthService(userRepoStub{
		createFn: func(_ context.Context, _ model.User) (int64, error) { return 7, nil },
		getByUsernameFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{}, domainerr.ErrUserNotFound
		},
	}, "test-secret", -1*time.Second)

	token, err := svc.SignUp(context.Background(), "John", "john", "password123")
	if err != nil {
		t.Fatalf("unexpected sign-up error: %v", err)
	}

	_, err = svc.ParseToken(token)
	if !errors.Is(err, domainerr.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
