package repository

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PolicyRepository interface {
	GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*Policy, error)
	CreatePolicy(ctx context.Context, p *Policy) error
	ListPolicies(ctx context.Context, userID uuid.UUID) ([]*Policy, error)
	ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, p *Policy) error
	DeletePolicy(ctx context.Context, userID uuid.UUID, id string) error
}

type MongoPolicyRepository struct {
	coll *mongo.Collection
}

func NewMongoPolicyRepository(db *mongo.Database) *MongoPolicyRepository {
	return &MongoPolicyRepository{coll: db.Collection("policies")}
}

func (r *MongoPolicyRepository) GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*Policy, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid policy id")
	}
	var policy Policy
	filter := bson.M{"_id": objID, "user_id": userID.String()}
	err = r.coll.FindOne(ctx, filter).Decode(&policy)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *MongoPolicyRepository) CreatePolicy(ctx context.Context, p *Policy) error {
	res, err := r.coll.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		p.ID = oid
	}
	return nil
}

func (r *MongoPolicyRepository) ListPolicies(ctx context.Context, userID uuid.UUID) ([]*Policy, error) {
	cur, err := r.coll.Find(ctx, bson.M{"user_id": userID.String()})
	if err != nil {
		return nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			log.Printf("failed to close cursor: %v\n", err)
		}
	}(cur, ctx)
	policies := make([]*Policy, 0)
	if err := cur.All(ctx, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// ListPoliciesByMarketplace retrieves all policies for a specific marketplace.
func (r *MongoPolicyRepository) ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*Policy, error) {
	cur, err := r.coll.Find(ctx, bson.M{"user_id": userID.String(), "marketplace": marketplace})
	if err != nil {
		return nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			log.Printf("failed to close cursor: %v\n", err)
		}
	}(cur, ctx)
	policies := make([]*Policy, 0)
	if err := cur.All(ctx, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// DeletePolicy deletes a policy by its ID.
func (r *MongoPolicyRepository) DeletePolicy(ctx context.Context, userID uuid.UUID, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid policy id")
	}
	filter := bson.M{"_id": objID, "user_id": userID.String()}
	res, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoPolicyRepository) UpdatePolicy(ctx context.Context, userID uuid.UUID, p *Policy) error {
	filter := bson.M{"_id": p.ID, "user_id": userID.String()}
	update := bson.M{"$set": bson.M{
		"name":  p.Name,
		"rules": p.Rules,
	}}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
