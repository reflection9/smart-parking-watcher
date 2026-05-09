package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Event struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	EventID       string        `bson:"event_id"`
	Source        string        `bson:"source"`
	ReservationID *int64        `bson:"reservation_id,omitempty"`
	UserID        *int64        `bson:"user_id,omitempty"`
	ZoneID        int64         `bson:"zone_id"`
	SpotID        int64         `bson:"spot_id"`
	EventType     string        `bson:"event_type"`
	Status        string        `bson:"status,omitempty"`
	OldStatus     string        `bson:"old_status,omitempty"`
	NewStatus     string        `bson:"new_status,omitempty"`
	ExpiresAt     *time.Time    `bson:"expires_at,omitempty"`
	ConfirmedAt   *time.Time    `bson:"confirmed_at,omitempty"`
	OccurredAt    time.Time     `bson:"occurred_at"`
	CreatedAt     time.Time     `bson:"created_at"`
}
