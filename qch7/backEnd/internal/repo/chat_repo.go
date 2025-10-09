package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"backEnd/internal/db"
	"backEnd/internal/models"
)

type ChatRepository struct {
	convCol *mongo.Collection
	msgCol  *mongo.Collection
}

func NewChatRepository() *ChatRepository {
	database := db.DB()
	return &ChatRepository{
		convCol: database.Collection(models.ColConversations),
		msgCol:  database.Collection(models.ColMessages),
	}
}

func (r *ChatRepository) UpsertConversation(ctx context.Context, userID primitive.ObjectID, title, model string) (*models.Conversation, error) {
	conv := &models.Conversation{
		UserID:    userID,
		Title:     title,
		Model:     model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	res, err := r.convCol.InsertOne(ctx, conv)
	if err != nil {
		return nil, err
	}
	conv.ID = res.InsertedID.(primitive.ObjectID)
	return conv, nil
}

func (r *ChatRepository) InsertMessage(ctx context.Context, m *models.ChatMessage) error {
	m.CreatedAt = time.Now()
	_, err := r.msgCol.InsertOne(ctx, m)
	return err
}

func (r *ChatRepository) ListMessages(ctx context.Context, convID primitive.ObjectID, limit int64) ([]models.ChatMessage, error) {
	cur, err := r.msgCol.Find(ctx, bson.M{"conversation_id": convID}, nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []models.ChatMessage
	for cur.Next(ctx) {
		var m models.ChatMessage
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, cur.Err()
}

func (r *ChatRepository) FindConversationByIDAndUser(ctx context.Context, convID primitive.ObjectID, userID primitive.ObjectID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.convCol.FindOne(ctx, bson.M{"_id": convID, "user_id": userID}).Decode(&conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}
