package redisstore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrTokenNotFound = errors.New("token not found or expired")

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

func (s *TokenStore) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	key := refreshKey(token)
	return s.client.Set(ctx, key, userID, ttl).Err()
}

func (s *TokenStore) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	key := refreshKey(token)
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *TokenStore) RevokeRefreshToken(ctx context.Context, token string) error {
	key := refreshKey(token)
	return s.client.Del(ctx, key).Err()
}

func (s *TokenStore) StoreVerificationToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return s.client.Set(ctx, verifyKey(token), userID, ttl).Err()
}

func (s *TokenStore) GetUserIDByVerificationToken(ctx context.Context, token string) (string, error) {
	val, err := s.client.Get(ctx, verifyKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrTokenNotFound
	}
	return val, err
}

func (s *TokenStore) DeleteVerificationToken(ctx context.Context, token string) error {
	return s.client.Del(ctx, verifyKey(token)).Err()
}

func (s *TokenStore) StoreResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return s.client.Set(ctx, resetKey(token), userID, ttl).Err()
}

func (s *TokenStore) GetUserIDByResetToken(ctx context.Context, token string) (string, error) {
	val, err := s.client.Get(ctx, resetKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrTokenNotFound
	}
	return val, err
}

func (s *TokenStore) DeleteResetToken(ctx context.Context, token string) error {
	return s.client.Del(ctx, resetKey(token)).Err()
}

func refreshKey(token string) string {
	return fmt.Sprintf("refresh:%s", hash(token))
}

func verifyKey(token string) string {
	return fmt.Sprintf("verify:%s", token)
}

func resetKey(token string) string {
	return fmt.Sprintf("reset:%s", token)
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)
}
