package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ColUsers         = "users"
	ColConversations = "conversations"
	ColMessages      = "messages"
)

// EnsureIndexes creates indexes for collections. Call at startup.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// users: unique email
	userModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	}
	if _, err := db.Collection(ColUsers).Indexes().CreateOne(ctx, userModel); err != nil {
		return err
	}
	// conversations: user_id + updated_at desc index (for listing)
	convModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "updated_at", Value: -1}},
		Options: options.Index().SetName("user_updated_at"),
	}
	if _, err := db.Collection(ColConversations).Indexes().CreateOne(ctx, convModel); err != nil {
		return err
	}
	// messages: conversation_id + created_at asc
	msgModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "conversation_id", Value: 1}, {Key: "created_at", Value: 1}},
		Options: options.Index().SetName("conv_created_at"),
	}
	if _, err := db.Collection(ColMessages).Indexes().CreateOne(ctx, msgModel); err != nil {
		return err
	}
	// optional TTL for messages (e.g., 90 days) — comment out if not needed
	_ = time.Second // placeholder to avoid unused import if TTL is not used
	return nil
}
