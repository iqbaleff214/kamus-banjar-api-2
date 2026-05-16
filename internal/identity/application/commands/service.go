package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/mailer"
)

const (
	verifyTokenTTL = 24 * time.Hour
	resetTokenTTL  = 1 * time.Hour
	refreshTTL     = 7 * 24 * time.Hour
)

// TokenStore defines the Redis token operations needed by Identity commands.
type TokenStore interface {
	StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (string, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	StoreVerificationToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetUserIDByVerificationToken(ctx context.Context, token string) (string, error)
	DeleteVerificationToken(ctx context.Context, token string) error
	StoreResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetUserIDByResetToken(ctx context.Context, token string) (string, error)
	DeleteResetToken(ctx context.Context, token string) error
}

// Service is the application service for Identity commands.
type Service struct {
	users  domain.UserRepository
	tokens TokenStore
	mail   mailer.Mailer
}

func NewService(users domain.UserRepository, tokens TokenStore, mail mailer.Mailer) *Service {
	return &Service{users: users, tokens: tokens, mail: mail}
}

// RegisterUser creates a new user and sends a verification email.
func (s *Service) RegisterUser(ctx context.Context, name, email, password, passwordConfirmation string) (*domain.User, error) {
	if password != passwordConfirmation {
		return nil, ErrPasswordMismatch
	}
	user, err := domain.NewUser(name, email, password)
	if err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	token, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	if err := s.tokens.StoreVerificationToken(ctx, token, user.ID.String(), verifyTokenTTL); err != nil {
		return nil, err
	}
	_ = s.mail.SendVerificationEmail(user.Email, token)
	return user, nil
}

// VerifyEmail verifies the user's email via token.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.tokens.GetUserIDByVerificationToken(ctx, token)
	if err != nil {
		return domain.ErrUnverifiedEmail
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrUnverifiedEmail
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsEmailVerified() {
		_ = s.tokens.DeleteVerificationToken(ctx, token)
		return nil
	}
	user.MarkEmailVerified()
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}
	return s.tokens.DeleteVerificationToken(ctx, token)
}

// LoginUser authenticates a user and returns an access+refresh token pair.
func (s *Service) LoginUser(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", domain.ErrWrongPassword
	}
	if !user.IsActive {
		return "", "", domain.ErrUserBanned
	}
	if !user.VerifyPassword(password) {
		return "", "", domain.ErrWrongPassword
	}

	accessToken, err = auth.GenerateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return "", "", err
	}
	refreshToken, err = randomHex(32)
	if err != nil {
		return "", "", err
	}
	if err := s.tokens.StoreRefreshToken(ctx, refreshToken, user.ID.String(), refreshTTL); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// RefreshToken issues a new token pair, revoking the old refresh token.
func (s *Service) RefreshToken(ctx context.Context, oldRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	userID, err := s.tokens.GetUserIDByRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return "", "", domain.ErrWrongPassword
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", "", domain.ErrWrongPassword
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err = auth.GenerateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err = randomHex(32)
	if err != nil {
		return "", "", err
	}
	if err := s.tokens.StoreRefreshToken(ctx, newRefreshToken, user.ID.String(), refreshTTL); err != nil {
		return "", "", err
	}
	_ = s.tokens.RevokeRefreshToken(ctx, oldRefreshToken)
	return newAccessToken, newRefreshToken, nil
}

// LogoutUser revokes the given refresh token.
func (s *Service) LogoutUser(ctx context.Context, refreshToken string) error {
	return s.tokens.RevokeRefreshToken(ctx, refreshToken)
}

// ForgotPassword sends a password-reset email if the email exists.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil // silent — no user enumeration
	}
	token, err := randomHex(32)
	if err != nil {
		return err
	}
	if err := s.tokens.StoreResetToken(ctx, token, user.ID.String(), resetTokenTTL); err != nil {
		return err
	}
	_ = s.mail.SendPasswordResetEmail(user.Email, token)
	return nil
}

// ResetPassword resets the user's password using a valid reset token.
func (s *Service) ResetPassword(ctx context.Context, token, password, passwordConfirmation string) error {
	if password != passwordConfirmation {
		return ErrPasswordMismatch
	}
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}
	userID, err := s.tokens.GetUserIDByResetToken(ctx, token)
	if err != nil {
		return domain.ErrUnverifiedEmail
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrUnverifiedEmail
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.SetPasswordHash(string(hash))
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}
	return s.tokens.DeleteResetToken(ctx, token)
}

// UpdateProfile updates the user's display name.
func (s *Service) UpdateProfile(ctx context.Context, userID, name string) (*domain.User, error) {
	if name == "" {
		return nil, domain.ErrEmptyName
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword verifies the current password before updating it.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !user.VerifyPassword(currentPassword) {
		return domain.ErrWrongPassword
	}
	if len(newPassword) < 8 {
		return domain.ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.SetPasswordHash(string(hash))
	return s.users.Update(ctx, user)
}

var ErrPasswordMismatch = sentinelError("passwords do not match")

type sentinelError string

func (e sentinelError) Error() string { return string(e) }

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
