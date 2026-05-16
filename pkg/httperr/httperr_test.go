package httperr_test

import (
	"encoding/json"
	"testing"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorResponseShape(t *testing.T) {
	cases := []struct {
		name     string
		resp     *httperr.ErrorResponse
		wantCode string
	}{
		{"validation", httperr.Validation("bad input"), httperr.CodeValidation},
		{"unauthorized", httperr.Unauthorized("no auth"), httperr.CodeUnauth},
		{"forbidden", httperr.Forbidden("no access"), httperr.CodeForbidden},
		{"not_found", httperr.NotFound("missing"), httperr.CodeNotFound},
		{"conflict", httperr.Conflict("dup"), httperr.CodeConflict},
		{"rate_limited", httperr.RateLimited("slow down"), httperr.CodeRateLimited},
		{"ai_unavailable", httperr.AIUnavailable("ai down"), httperr.CodeAIUnavail},
		{"internal", httperr.Internal("oops"), httperr.CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, tc.resp.Success)
			assert.Equal(t, tc.wantCode, tc.resp.Error.Code)
			assert.NotEmpty(t, tc.resp.Error.Message)

			b, err := json.Marshal(tc.resp)
			require.NoError(t, err)
			var m map[string]any
			require.NoError(t, json.Unmarshal(b, &m))
			assert.False(t, m["success"].(bool))
			errObj := m["error"].(map[string]any)
			assert.Equal(t, tc.wantCode, errObj["code"])
		})
	}
}

func TestSuccessResponseShape(t *testing.T) {
	meta := &httperr.PaginationMeta{Page: 2, PerPage: 10, Total: 55, TotalPages: 6}
	resp := &httperr.SuccessResponse{Success: true, Data: fiber_map{"id": "abc"}, Meta: meta}

	b, err := json.Marshal(resp)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.True(t, m["success"].(bool))
	assert.NotNil(t, m["data"])
	metaObj := m["meta"].(map[string]any)
	assert.Equal(t, float64(2), metaObj["page"])
	assert.Equal(t, float64(10), metaObj["per_page"])
	assert.Equal(t, float64(55), metaObj["total"])
	assert.Equal(t, float64(6), metaObj["total_pages"])
}

type fiber_map map[string]any
