package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
)

func TestNewUser_ValidInput(t *testing.T) {
	u, err := domain.NewUser("Alice", "alice@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, "alice@example.com", u.Email)
	assert.NotEmpty(t, u.PasswordHash)
	assert.NotEqual(t, "password123", u.PasswordHash, "password must be hashed")
	assert.Equal(t, domain.RoleUser, u.Role)
	assert.True(t, u.IsActive)
	assert.Nil(t, u.EmailVerifiedAt)
	assert.NotEqual(t, "", u.ID.String())
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := domain.NewUser("Alice", "not-an-email", "password123")
	assert.ErrorIs(t, err, domain.ErrInvalidEmail)
}

func TestNewUser_ShortPassword(t *testing.T) {
	_, err := domain.NewUser("Alice", "alice@example.com", "short")
	assert.ErrorIs(t, err, domain.ErrPasswordTooShort)
}

func TestNewUser_EmptyName(t *testing.T) {
	_, err := domain.NewUser("", "alice@example.com", "password123")
	assert.ErrorIs(t, err, domain.ErrEmptyName)
}

func TestVerifyPassword(t *testing.T) {
	u, err := domain.NewUser("Alice", "alice@example.com", "correctpassword")
	require.NoError(t, err)

	assert.True(t, u.VerifyPassword("correctpassword"))
	assert.False(t, u.VerifyPassword("wrongpassword"))
}

func TestPromote_InvalidRole(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "password123")
	err := u.Promote("superuser")
	assert.ErrorIs(t, err, domain.ErrInvalidRole)
}

func TestPromote_ValidRole(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "password123")
	err := u.Promote(domain.RoleAdmin)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleAdmin, u.Role)
}

func TestBanUnban(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "password123")
	assert.True(t, u.IsActive)

	u.Ban()
	assert.False(t, u.IsActive)

	u.Unban()
	assert.True(t, u.IsActive)
}

func TestMarkEmailVerified(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "password123")
	assert.False(t, u.IsEmailVerified())

	u.MarkEmailVerified()
	assert.True(t, u.IsEmailVerified())
	assert.NotNil(t, u.EmailVerifiedAt)
}
