package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type User struct {
	ID              uuid.UUID
	Name            string
	Email           string
	PasswordHash    string
	Role            Role
	IsActive        bool
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUser creates and validates a new User, hashing the password.
func NewUser(name, email, password string) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if !emailRe.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         RoleUser,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// VerifyPassword returns true if plain matches the stored hash.
func (u *User) VerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain))
	return err == nil
}

// Promote changes the user's role. Only RoleAdmin is accepted.
func (u *User) Promote(role Role) error {
	if role != RoleAdmin && role != RoleUser {
		return ErrInvalidRole
	}
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// Ban deactivates the user account.
func (u *User) Ban() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

// Unban reactivates the user account.
func (u *User) Unban() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

// SetPasswordHash directly sets a pre-hashed password (used by reset flow).
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now().UTC()
}

// MarkEmailVerified sets the verification timestamp.
func (u *User) MarkEmailVerified() {
	now := time.Now().UTC()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// IsEmailVerified returns true if the email has been verified.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
