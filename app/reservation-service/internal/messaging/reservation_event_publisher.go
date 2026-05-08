package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type ReservationEventPublisher interface {
	Publish(ctx context.Context, event ReservationLifecycleEvent) error
	Close() error
}

type noopReservationEventPublisher struct{}

type kafkaReservationEventPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewNoopReservationEventPublisher() ReservationEventPublisher {
	return &noopReservationEventPublisher{}
}

func NewKafkaReservationEventPublisher(
	brokers []string,
	topic string,
) ReservationEventPublisher {
	return &kafkaReservationEventPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			BatchTimeout: 100 * time.Millisecond,
		},
		topic: topic,
	}
}

func (p *noopReservationEventPublisher) Publish(context.Context, ReservationLifecycleEvent) error {
	return nil
}

func (p *noopReservationEventPublisher) Close() error {
	return nil
}

func (p *kafkaReservationEventPublisher) Publish(
	ctx context.Context,
	event ReservationLifecycleEvent,
) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(strconv.FormatInt(event.ReservationID, 10)),
		Value: payload,
		Time:  event.OccurredAt,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "source", Value: []byte(event.Source)},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish reservation event: %w", err)
	}

	return nil
}

func (p *kafkaReservationEventPublisher) Close() error {
	return p.writer.Close()
}
