package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Event struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	ZoneID    int64         `bson:"zone_id"`
	SpotID    int64         `bson:"spot_id"`
	EventType string        `bson:"event_type"`
	OldStatus string        `bson:"old_status,omitempty"`
	NewStatus string        `bson:"new_status"`
	CreatedAt time.Time     `bson:"created_at"`
}
