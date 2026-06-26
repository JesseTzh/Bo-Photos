package auth

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestSetupCreatesOnlyOneAdministrator(t *testing.T) {
	service := newTestService(t)

	if err := service.Setup(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("first Setup() error = %v", err)
	}
	if err := service.Setup(context.Background(), "another strong password"); err != ErrAlreadyInitialized {
		t.Fatalf("second Setup() error = %v, want ErrAlreadyInitialized", err)
	}
}

func TestSetupRejectsPasswordsOutsideLengthRange(t *testing.T) {
	service := newTestService(t)

	for _, password := range []string{"too short", strings.Repeat("长", 129)} {
		if err := service.Setup(context.Background(), password); err != ErrInvalidPassword {
			t.Errorf("Setup(%d runes) error = %v, want ErrInvalidPassword", len([]rune(password)), err)
		}
	}
}

func TestEnsureInitializedRequiresInitialPasswordForEmptyDatabase(t *testing.T) {
	service := newTestService(t)

	if err := service.EnsureInitialized(context.Background(), ""); err != ErrInitialPasswordMissing {
		t.Fatalf("EnsureInitialized() error = %v, want ErrInitialPasswordMissing", err)
	}
}

func TestEnsureInitializedCreatesAdministratorFromInitialPassword(t *testing.T) {
	service := newTestService(t)

	if err := service.EnsureInitialized(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("EnsureInitialized() error = %v", err)
	}
	if err := service.VerifyPassword(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
}

func TestEnsureInitializedIgnoresInitialPasswordWhenAlreadyInitialized(t *testing.T) {
	service := newTestService(t)
	if err := service.Setup(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := service.EnsureInitialized(context.Background(), "short"); err != nil {
		t.Fatalf("EnsureInitialized() error = %v", err)
	}
	if err := service.VerifyPassword(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	service := newTestService(t)
	if err := service.Setup(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := service.VerifyPassword(context.Background(), "wrong password"); err != ErrInvalidCredentials {
		t.Fatalf("VerifyPassword() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestChangePasswordReplacesCurrentPassword(t *testing.T) {
	service := newTestService(t)
	if err := service.Setup(context.Background(), "correct horse battery staple"); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if err := service.ChangePassword(
		context.Background(),
		"correct horse battery staple",
		"new correct horse battery staple",
	); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if err := service.VerifyPassword(context.Background(), "correct horse battery staple"); err != ErrInvalidCredentials {
		t.Errorf("old password error = %v, want ErrInvalidCredentials", err)
	}
	if err := service.VerifyPassword(context.Background(), "new correct horse battery staple"); err != nil {
		t.Errorf("new password error = %v", err)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("repository.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := repository.Migrate(db); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}

	return NewService(NewRepository(db))
}
