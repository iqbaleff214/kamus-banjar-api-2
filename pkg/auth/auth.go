package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour

	claimsKey = "claims"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var secret []byte

// Init must be called once at startup with the JWT secret.
func Init(jwtSecret string) {
	secret = []byte(jwtSecret)
}

// GenerateAccessToken creates a signed HS256 JWT with 15-minute TTL.
func GenerateAccessToken(userID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// ParseToken validates a JWT and returns the embedded claims.
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GetClaims retrieves injected claims from the Fiber context.
func GetClaims(c *fiber.Ctx) *Claims {
	v := c.Locals(claimsKey)
	if v == nil {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}

// RequireAuth validates the Bearer token and injects Claims into context.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return httperr.Send(c, fiber.StatusUnauthorized, httperr.Unauthorized("missing or invalid authorization header"))
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := ParseToken(tokenStr)
		if err != nil {
			return httperr.Send(c, fiber.StatusUnauthorized, httperr.Unauthorized("invalid or expired token"))
		}
		c.Locals(claimsKey, claims)
		return c.Next()
	}
}

// RequireRole returns a middleware that permits only the specified roles.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return httperr.Send(c, fiber.StatusUnauthorized, httperr.Unauthorized("authentication required"))
		}
		if _, ok := allowed[claims.Role]; !ok {
			return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden("insufficient permissions"))
		}
		return c.Next()
	}
}
