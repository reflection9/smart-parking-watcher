package expiration

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisReservationTTLTracker struct {
	client    *redis.Client
	db        int
	keyPrefix string
	pubsub    *redis.PubSub
}

func NewRedisReservationTTLTracker(
	addr string,
	password string,
	db int,
	keyPrefix string,
) ReservationTTLTracker {
	return &redisReservationTTLTracker{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		db:        db,
		keyPrefix: keyPrefix,
	}
}

func (t *redisReservationTTLTracker) TrackReservation(
	ctx context.Context,
	reservationID uint,
	expiresAt time.Time,
) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	return t.client.Set(ctx, t.key(reservationID), reservationID, ttl).Err()
}

func (t *redisReservationTTLTracker) RemoveReservation(ctx context.Context, reservationID uint) error {
	return t.client.Del(ctx, t.key(reservationID)).Err()
}

func (t *redisReservationTTLTracker) Start(
	ctx context.Context,
	onExpired func(context.Context, uint) error,
) error {
	if onExpired == nil {
		return nil
	}

	if err := t.client.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err(); err != nil {
		log.Printf("failed to enable Redis keyspace notifications: %v", err)
	}

	channelName := fmt.Sprintf("__keyevent@%d__:expired", t.db)
	pubsub := t.client.Subscribe(ctx, channelName)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return err
	}

	t.pubsub = pubsub
	channel := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-channel:
			if !ok {
				return nil
			}

			reservationID, matched := t.parseReservationID(message.Payload)
			if !matched {
				continue
			}

			if err := onExpired(context.Background(), reservationID); err != nil {
				log.Printf("failed to handle Redis reservation expiration for %d: %v", reservationID, err)
			}
		}
	}
}

func (t *redisReservationTTLTracker) Close() error {
	if t.pubsub != nil {
		if err := t.pubsub.Close(); err != nil {
			return err
		}
	}

	if t.client != nil {
		return t.client.Close()
	}

	return nil
}

func (t *redisReservationTTLTracker) key(reservationID uint) string {
	return fmt.Sprintf("%s%d", t.keyPrefix, reservationID)
}

func (t *redisReservationTTLTracker) parseReservationID(key string) (uint, bool) {
	if !strings.HasPrefix(key, t.keyPrefix) {
		return 0, false
	}

	rawID := strings.TrimPrefix(key, t.keyPrefix)
	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, false
	}

	return uint(parsedID), true
}
