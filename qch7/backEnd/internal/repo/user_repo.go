package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backEnd/internal/db"
	"backEnd/internal/models"
)

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository() *UserRepository {
	return &UserRepository{col: db.DB().Collection(models.ColUsers)}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	_, err := r.col.InsertOne(ctx, u)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var out models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	opts := options.Count().SetLimit(1)
	n, err := r.col.CountDocuments(ctx, bson.M{"email": email}, opts)
	return n > 0, err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID primitive.ObjectID, role models.UserRole) error {
	update := bson.M{
		"$set": bson.M{
			"role":       role,
			"updated_at": time.Now(),
		},
	}
	_, err := r.col.UpdateByID(ctx, userID, update)
	return err
}
