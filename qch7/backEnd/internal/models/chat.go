package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageRole string

const (
	MsgUser      MessageRole = "user"
	MsgAssistant MessageRole = "assistant"
	MsgSystem    MessageRole = "system"
)

type Conversation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	Title     string             `bson:"title" json:"title"`
	Model     string             `bson:"model" json:"model"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

type ChatMessage struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversation_id" json:"conversationId"`
	UserID         primitive.ObjectID `bson:"user_id" json:"userId"`
	Role           MessageRole        `bson:"role" json:"role"`
	Content        string             `bson:"content" json:"content"`
	CreatedAt      time.Time          `bson:"created_at" json:"createdAt"`
}
