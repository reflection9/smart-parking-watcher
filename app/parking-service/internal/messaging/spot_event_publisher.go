package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type SpotEventPublisher interface {
	Publish(ctx context.Context, event SpotStatusEvent) error
	Close() error
}

type noopSpotEventPublisher struct{}

type kafkaSpotEventPublisher struct {
	writer *kafka.Writer
}

func NewNoopSpotEventPublisher() SpotEventPublisher {
	return &noopSpotEventPublisher{}
}

func NewKafkaSpotEventPublisher(brokers []string, topic string) SpotEventPublisher {
	return &kafkaSpotEventPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			BatchTimeout: 100 * time.Millisecond,
		},
	}
}

func (p *noopSpotEventPublisher) Publish(context.Context, SpotStatusEvent) error {
	return nil
}

func (p *noopSpotEventPublisher) Close() error {
	return nil
}

func (p *kafkaSpotEventPublisher) Publish(ctx context.Context, event SpotStatusEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(strconv.FormatInt(event.SpotID, 10)),
		Value: payload,
		Time:  event.OccurredAt,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "source", Value: []byte(event.Source)},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish spot event: %w", err)
	}

	return nil
}

func (p *kafkaSpotEventPublisher) Close() error {
	return p.writer.Close()
}
