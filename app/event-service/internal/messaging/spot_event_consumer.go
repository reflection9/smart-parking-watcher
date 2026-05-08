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

type SpotEventHandler interface {
	HandleSpotEvent(ctx context.Context, event SpotStatusEvent) error
}

type SpotEventConsumer interface {
	Start(ctx context.Context, handler SpotEventHandler) error
	Close() error
}

type noopSpotEventConsumer struct{}

type kafkaSpotEventConsumer struct {
	reader *kafka.Reader
}

func NewNoopSpotEventConsumer() SpotEventConsumer {
	return &noopSpotEventConsumer{}
}

func NewKafkaSpotEventConsumer(
	brokers []string,
	topic string,
	groupID string,
) SpotEventConsumer {
	return &kafkaSpotEventConsumer{
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

func (c *noopSpotEventConsumer) Start(context.Context, SpotEventHandler) error {
	return nil
}

func (c *noopSpotEventConsumer) Close() error {
	return nil
}

func (c *kafkaSpotEventConsumer) Start(ctx context.Context, handler SpotEventHandler) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			log.Printf("failed to fetch spot kafka message: %v", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		var event SpotStatusEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to decode spot event: %v", err)
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				return fmt.Errorf("failed to commit malformed spot kafka message: %w", commitErr)
			}
			continue
		}

		if err := handler.HandleSpotEvent(ctx, event); err != nil {
			log.Printf("failed to handle spot event %s: %v", event.EventID, err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("failed to commit spot kafka message: %w", err)
		}
	}
}

func (c *kafkaSpotEventConsumer) Close() error {
	return c.reader.Close()
}
