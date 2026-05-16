package mailer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/mailer"
)

func TestSendVerificationEmail_MockCapture(t *testing.T) {
	m := &mailer.MockMailer{}
	err := m.SendVerificationEmail("user@example.com", "abc123token")
	require.NoError(t, err)
	require.Len(t, m.VerificationCalls, 1)
	assert.Equal(t, "user@example.com", m.VerificationCalls[0].To)
	assert.Equal(t, "abc123token", m.VerificationCalls[0].Token)
}

func TestSendPasswordResetEmail_MockCapture(t *testing.T) {
	m := &mailer.MockMailer{}
	err := m.SendPasswordResetEmail("user@example.com", "reset456token")
	require.NoError(t, err)
	require.Len(t, m.ResetCalls, 1)
	assert.Equal(t, "reset456token", m.ResetCalls[0].Token)
}
