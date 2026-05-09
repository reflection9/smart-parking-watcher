package repository

import (
	"context"
	"history-service/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoEventRepository struct {
	collection *mongo.Collection
}

func NewMongoEventRepository(collection *mongo.Collection) *MongoEventRepository {
	return &MongoEventRepository{collection: collection}
}

func (r *MongoEventRepository) Create(ctx context.Context, event *model.Event) (bool, error) {
	result, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}

		return false, err
	}

	if objectID, ok := result.InsertedID.(bson.ObjectID); ok {
		event.ID = objectID
	}

	return true, nil
}

func (r *MongoEventRepository) ListByZoneID(ctx context.Context, zoneID int64) ([]model.Event, error) {
	return r.findMany(ctx, bson.M{"zone_id": zoneID})
}

func (r *MongoEventRepository) ListBySpotID(ctx context.Context, spotID int64) ([]model.Event, error) {
	return r.findMany(ctx, bson.M{"spot_id": spotID})
}

func (r *MongoEventRepository) ListByReservationID(ctx context.Context, reservationID int64) ([]model.Event, error) {
	return r.findMany(ctx, bson.M{"reservation_id": reservationID})
}

func (r *MongoEventRepository) ListOlderThan(
	ctx context.Context,
	cutoff time.Time,
	limit int64,
) ([]model.Event, error) {
	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: 1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"occurred_at": bson.M{"$lte": cutoff}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	events := make([]model.Event, 0)
	for cursor.Next(ctx) {
		var event model.Event
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *MongoEventRepository) DeleteByIDs(ctx context.Context, ids []bson.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	return err
}

func (r *MongoEventRepository) findMany(ctx context.Context, filter bson.M) ([]model.Event, error) {
	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	events := make([]model.Event, 0)
	for cursor.Next(ctx) {
		var event model.Event
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
