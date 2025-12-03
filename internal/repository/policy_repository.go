package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Policy struct {
	ID     string   `bson:"_id,omitempty"`
	UserID string   `bson:"user_id"`
	Name   string   `bson:"name"`
	Rules  []string `bson:"rules"`
}

type PolicyRepository interface {
	GetPolicy(ctx context.Context, id string) (*Policy, error)
	CreatePolicy(ctx context.Context, p *Policy) error
	ListPolicies(ctx context.Context) ([]*Policy, error)
	UpdatePolicy(ctx context.Context, p *Policy) error
}

type MongoPolicyRepository struct {
	coll *mongo.Collection
}

func NewMongoPolicyRepository(db *mongo.Database) *MongoPolicyRepository {
	return &MongoPolicyRepository{coll: db.Collection("policies")}
}

func (r *MongoPolicyRepository) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	// TODO: implement
	return nil, nil
}

func (r *MongoPolicyRepository) CreatePolicy(ctx context.Context, p *Policy) error {
	// TODO: implement
	return nil
}

func (r *MongoPolicyRepository) ListPolicies(ctx context.Context) ([]*Policy, error) {
	// TODO: implement
	return []*Policy{}, nil
}

func (r *MongoPolicyRepository) UpdatePolicy(ctx context.Context, p *Policy) error {
	// TODO: implement
	return nil
}
