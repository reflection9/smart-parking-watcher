package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type ParkingCommandPublisher interface {
	Publish(ctx context.Context, command SpotCommand) error
	Close() error
}

type noopParkingCommandPublisher struct{}

type kafkaParkingCommandPublisher struct {
	writer *kafka.Writer
}

func NewNoopParkingCommandPublisher() ParkingCommandPublisher {
	return &noopParkingCommandPublisher{}
}

func NewKafkaParkingCommandPublisher(
	brokers []string,
	topic string,
) ParkingCommandPublisher {
	return &kafkaParkingCommandPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			BatchTimeout: 100 * time.Millisecond,
		},
	}
}

func (p *noopParkingCommandPublisher) Publish(
	context.Context,
	SpotCommand,
) error {
	return nil
}

func (p *noopParkingCommandPublisher) Close() error {
	return nil
}

func (p *kafkaParkingCommandPublisher) Publish(
	ctx context.Context,
	command SpotCommand,
) error {
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(strconv.FormatInt(command.ReservationID, 10)),
		Value: payload,
		Time:  command.OccurredAt,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(command.EventID)},
			{Key: "event_type", Value: []byte(command.EventType)},
			{Key: "source", Value: []byte(command.Source)},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish parking command: %w", err)
	}

	return nil
}

func (p *kafkaParkingCommandPublisher) Close() error {
	return p.writer.Close()
}
