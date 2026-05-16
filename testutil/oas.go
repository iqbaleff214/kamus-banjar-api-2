package testutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertSuccessEnvelope validates {"success":true,"data":...} and returns the
// parsed envelope so callers can assert on the data field.
func AssertSuccessEnvelope(t testing.TB, body []byte) map[string]any {
	t.Helper()
	var env map[string]any
	require.NoError(t, json.Unmarshal(body, &env), "response body must be valid JSON")
	assert.Equal(t, true, env["success"], "success field must be true")
	assert.Contains(t, env, "data", "success response must contain a data key")
	return env
}

// AssertErrorEnvelope validates {"success":false,"error":{"code":...,"message":...}}
// and returns the parsed envelope.
func AssertErrorEnvelope(t testing.TB, body []byte) map[string]any {
	t.Helper()
	var env map[string]any
	require.NoError(t, json.Unmarshal(body, &env), "response body must be valid JSON")
	assert.Equal(t, false, env["success"], "success field must be false for error responses")
	errObj, ok := env["error"].(map[string]any)
	require.True(t, ok, "error response must contain an error object")
	assert.NotEmpty(t, errObj["code"], "error.code must be present and non-empty")
	assert.NotEmpty(t, errObj["message"], "error.message must be present and non-empty")
	return env
}
