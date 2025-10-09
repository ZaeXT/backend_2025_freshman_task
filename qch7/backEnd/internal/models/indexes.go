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

// EnsureIndexes 为各集合创建所需索引；请在启动时调用。
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
    // users: 邮箱唯一索引
	userModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	}
	if _, err := db.Collection(ColUsers).Indexes().CreateOne(ctx, userModel); err != nil {
		return err
	}
    // conversations: user_id + updated_at 降序（用于列表）
	convModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "updated_at", Value: -1}},
		Options: options.Index().SetName("user_updated_at"),
	}
	if _, err := db.Collection(ColConversations).Indexes().CreateOne(ctx, convModel); err != nil {
		return err
	}
    // messages: conversation_id + created_at 升序
	msgModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "conversation_id", Value: 1}, {Key: "created_at", Value: 1}},
		Options: options.Index().SetName("conv_created_at"),
	}
	if _, err := db.Collection(ColMessages).Indexes().CreateOne(ctx, msgModel); err != nil {
		return err
	}
    // 可选：消息 TTL（例如 90 天）— 如不需要可忽略
	_ = time.Second // placeholder to avoid unused import if TTL is not used
	return nil
}
