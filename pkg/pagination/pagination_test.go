package pagination_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testParse builds a minimal Fiber app that calls pagination.Parse and
// returns the result as JSON so we can inspect it outside the handler.
func testParse(t *testing.T, query string) (pagination.Params, *map[string]any, int) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		p, errResp := pagination.Parse(c)
		if errResp != nil {
			return c.Status(fiber.StatusBadRequest).JSON(errResp)
		}
		return c.JSON(fiber.Map{
			"page":     p.Page,
			"per_page": p.PerPage,
			"offset":   p.Offset(),
		})
	})
	req := httptest.NewRequest("GET", "/?"+query, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return pagination.Params{}, &body, resp.StatusCode
}

func TestPaginationDefaults(t *testing.T) {
	_, body, status := testParse(t, "")
	assert.Equal(t, 200, status)
	assert.Equal(t, float64(1), (*body)["page"])
	assert.Equal(t, float64(20), (*body)["per_page"])
}

func TestPaginationMaxPerPage(t *testing.T) {
	_, body, status := testParse(t, "per_page=200")
	assert.Equal(t, 400, status)
	assert.Equal(t, "VALIDATION_ERROR", (*body)["error"].(map[string]any)["code"])
}

func TestPaginationOffset(t *testing.T) {
	_, body, status := testParse(t, "page=3&per_page=20")
	assert.Equal(t, 200, status)
	assert.Equal(t, float64(40), (*body)["offset"])
}

func TestPaginationNegativePage(t *testing.T) {
	_, _, status := testParse(t, "page=-1")
	assert.Equal(t, 400, status)
}

func TestPaginationNonInteger(t *testing.T) {
	_, _, status := testParse(t, "page=abc")
	assert.Equal(t, 400, status)
}

func TestTotalPages(t *testing.T) {
	assert.Equal(t, 3, pagination.TotalPages(55, 20))
	assert.Equal(t, 3, pagination.TotalPages(60, 20))
	assert.Equal(t, 0, pagination.TotalPages(10, 0))
}
