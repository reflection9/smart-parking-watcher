package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"subscription-service/internal/dto"

	"github.com/redis/go-redis/v9"
)

type SubscriptionCache interface {
	GetByZoneID(ctx context.Context, zoneID int64) ([]dto.SubscriptionResponse, bool, error)
	SetByZoneID(ctx context.Context, zoneID int64, subscriptions []dto.SubscriptionResponse) error
	InvalidateZoneID(ctx context.Context, zoneID int64) error
	GetByUserID(ctx context.Context, userID int64) ([]dto.SubscriptionResponse, bool, error)
	SetByUserID(ctx context.Context, userID int64, subscriptions []dto.SubscriptionResponse) error
	InvalidateUserID(ctx context.Context, userID int64) error
	Close() error
}

type noopSubscriptionCache struct{}

func NewNoopSubscriptionCache() SubscriptionCache {
	return &noopSubscriptionCache{}
}

func (c *noopSubscriptionCache) GetByZoneID(context.Context, int64) ([]dto.SubscriptionResponse, bool, error) {
	return nil, false, nil
}

func (c *noopSubscriptionCache) SetByZoneID(context.Context, int64, []dto.SubscriptionResponse) error {
	return nil
}

func (c *noopSubscriptionCache) InvalidateZoneID(context.Context, int64) error {
	return nil
}

func (c *noopSubscriptionCache) GetByUserID(context.Context, int64) ([]dto.SubscriptionResponse, bool, error) {
	return nil, false, nil
}

func (c *noopSubscriptionCache) SetByUserID(context.Context, int64, []dto.SubscriptionResponse) error {
	return nil
}

func (c *noopSubscriptionCache) InvalidateUserID(context.Context, int64) error {
	return nil
}

func (c *noopSubscriptionCache) Close() error {
	return nil
}

type redisSubscriptionCache struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedisSubscriptionCache(
	addr string,
	password string,
	db int,
	keyPrefix string,
	ttl time.Duration,
) SubscriptionCache {
	return &redisSubscriptionCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (c *redisSubscriptionCache) GetByZoneID(
	ctx context.Context,
	zoneID int64,
) ([]dto.SubscriptionResponse, bool, error) {
	return c.get(ctx, c.zoneKey(zoneID))
}

func (c *redisSubscriptionCache) SetByZoneID(
	ctx context.Context,
	zoneID int64,
	subscriptions []dto.SubscriptionResponse,
) error {
	return c.set(ctx, c.zoneKey(zoneID), subscriptions)
}

func (c *redisSubscriptionCache) InvalidateZoneID(ctx context.Context, zoneID int64) error {
	return c.client.Del(ctx, c.zoneKey(zoneID)).Err()
}

func (c *redisSubscriptionCache) GetByUserID(
	ctx context.Context,
	userID int64,
) ([]dto.SubscriptionResponse, bool, error) {
	return c.get(ctx, c.userKey(userID))
}

func (c *redisSubscriptionCache) SetByUserID(
	ctx context.Context,
	userID int64,
	subscriptions []dto.SubscriptionResponse,
) error {
	return c.set(ctx, c.userKey(userID), subscriptions)
}

func (c *redisSubscriptionCache) InvalidateUserID(ctx context.Context, userID int64) error {
	return c.client.Del(ctx, c.userKey(userID)).Err()
}

func (c *redisSubscriptionCache) Close() error {
	return c.client.Close()
}

func (c *redisSubscriptionCache) get(
	ctx context.Context,
	key string,
) ([]dto.SubscriptionResponse, bool, error) {
	payload, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}

		return nil, false, err
	}

	var subscriptions []dto.SubscriptionResponse
	if err := json.Unmarshal([]byte(payload), &subscriptions); err != nil {
		return nil, false, err
	}

	return subscriptions, true, nil
}

func (c *redisSubscriptionCache) set(
	ctx context.Context,
	key string,
	subscriptions []dto.SubscriptionResponse,
) error {
	payload, err := json.Marshal(subscriptions)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, payload, c.ttl).Err()
}

func (c *redisSubscriptionCache) zoneKey(zoneID int64) string {
	return fmt.Sprintf("%szones:%d", c.keyPrefix, zoneID)
}

func (c *redisSubscriptionCache) userKey(userID int64) string {
	return fmt.Sprintf("%susers:%d", c.keyPrefix, userID)
}
