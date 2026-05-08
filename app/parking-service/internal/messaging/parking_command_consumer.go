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

type ParkingCommandHandler interface {
	HandleSpotCommand(ctx context.Context, command SpotCommand) error
}

type ParkingCommandConsumer interface {
	Start(ctx context.Context, handler ParkingCommandHandler) error
	Close() error
}

type noopParkingCommandConsumer struct{}

type kafkaParkingCommandConsumer struct {
	reader *kafka.Reader
}

func NewNoopParkingCommandConsumer() ParkingCommandConsumer {
	return &noopParkingCommandConsumer{}
}

func NewKafkaParkingCommandConsumer(
	brokers []string,
	topic string,
	groupID string,
) ParkingCommandConsumer {
	return &kafkaParkingCommandConsumer{
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

func (c *noopParkingCommandConsumer) Start(context.Context, ParkingCommandHandler) error {
	return nil
}

func (c *noopParkingCommandConsumer) Close() error {
	return nil
}

func (c *kafkaParkingCommandConsumer) Start(
	ctx context.Context,
	handler ParkingCommandHandler,
) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			log.Printf("failed to fetch parking command kafka message: %v", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		var command SpotCommand
		if err := json.Unmarshal(message.Value, &command); err != nil {
			log.Printf("failed to decode parking command: %v", err)
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				return fmt.Errorf("failed to commit malformed parking command: %w", commitErr)
			}
			continue
		}

		if err := handler.HandleSpotCommand(ctx, command); err != nil {
			log.Printf("failed to handle parking command %s: %v", command.EventID, err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * time.Second):
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("failed to commit parking command: %w", err)
		}
	}
}

func (c *kafkaParkingCommandConsumer) Close() error {
	return c.reader.Close()
}
