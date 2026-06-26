package auth

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
)

var (
	ErrInvalidPassword        = errors.New("password must contain 12 to 128 characters")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInitialPasswordMissing = errors.New("BOPHOTOS_INITIAL_PASSWORD must be set before first startup")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) IsInitialized(ctx context.Context) (bool, error) {
	return s.repository.IsInitialized(ctx)
}

func (s *Service) Setup(ctx context.Context, password string) error {
	if !validPassword(password) {
		return ErrInvalidPassword
	}
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("hash administrator password: %w", err)
	}
	return s.repository.Create(ctx, hash)
}

func (s *Service) EnsureInitialized(ctx context.Context, initialPassword string) error {
	initialized, err := s.IsInitialized(ctx)
	if err != nil {
		return err
	}
	if initialized {
		return nil
	}
	if initialPassword == "" {
		return ErrInitialPasswordMissing
	}
	if err := s.Setup(ctx, initialPassword); errors.Is(err, ErrAlreadyInitialized) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Service) VerifyPassword(ctx context.Context, password string) error {
	hash, err := s.repository.PasswordHash(ctx)
	if errors.Is(err, ErrNotInitialized) {
		return ErrInvalidCredentials
	}
	if err != nil {
		return err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return fmt.Errorf("verify administrator password: %w", err)
	}
	if !match {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, currentPassword, newPassword string) error {
	if err := s.VerifyPassword(ctx, currentPassword); err != nil {
		return err
	}
	if !validPassword(newPassword) {
		return ErrInvalidPassword
	}
	hash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("hash administrator password: %w", err)
	}
	return s.repository.UpdatePassword(ctx, hash)
}

func (s *Service) RevokeSessions(ctx context.Context) error {
	return s.repository.DeleteAllSessions(ctx)
}

func validPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 12 && length <= 128
}
