package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type ReservationEventHandler interface {
	HandleReservationEvent(ctx context.Context, event ReservationLifecycleEvent) error
}

type ReservationEventConsumer interface {
	Start(ctx context.Context, handler ReservationEventHandler) error
	Close() error
}

type noopReservationEventConsumer struct{}

type kafkaReservationEventConsumer struct {
	reader *kafka.Reader
}

func NewNoopReservationEventConsumer() ReservationEventConsumer {
	return &noopReservationEventConsumer{}
}

func NewKafkaReservationEventConsumer(
	brokers []string,
	topic string,
	groupID string,
) ReservationEventConsumer {
	return &kafkaReservationEventConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:                brokers,
			Topic:                  topic,
			GroupID:                groupID,
			MinBytes:               1,
			MaxBytes:               10e6,
			CommitInterval:         0,
			WatchPartitionChanges:  true,
			PartitionWatchInterval: time.Second,
		}),
	}
}

func (c *noopReservationEventConsumer) Start(context.Context, ReservationEventHandler) error {
	return nil
}

func (c *noopReservationEventConsumer) Close() error {
	return nil
}

func (c *kafkaReservationEventConsumer) Start(
	ctx context.Context,
	handler ReservationEventHandler,
) error {
	for {
		fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		message, err := c.reader.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}

			log.Printf("failed to fetch kafka message: %v", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		var event ReservationLifecycleEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to decode reservation event: %v", err)
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				return fmt.Errorf("failed to commit malformed kafka message: %w", commitErr)
			}
			continue
		}

		if err := handler.HandleReservationEvent(ctx, event); err != nil {
			log.Printf("failed to handle reservation event %s: %v", event.EventID, err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("failed to commit kafka message: %w", err)
		}
	}
}

func (c *kafkaReservationEventConsumer) Close() error {
	return c.reader.Close()
}
