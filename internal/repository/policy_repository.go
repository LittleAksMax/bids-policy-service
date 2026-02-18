package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PolicyRepository interface {
	GetPolicy(ctx context.Context, userID uuid.UUID, id primitive.ObjectID) (*Policy, error)
	CreatePolicy(ctx context.Context, p *Policy) error
	ListPoliciesWithMarketplace(ctx context.Context, userID uuid.UUID, marketplace *string) ([]*Policy, error)
	ListPolicies(ctx context.Context, userID uuid.UUID) ([]*Policy, error)
	ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, p *Policy) error
	DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (*Policy, error)
}

type MongoPolicyRepository struct {
	coll *mongo.Collection
}

func NewMongoPolicyRepository(db *mongo.Database) *MongoPolicyRepository {
	return &MongoPolicyRepository{coll: db.Collection("policies")}
}

func (r *MongoPolicyRepository) GetPolicy(ctx context.Context, userID uuid.UUID, id primitive.ObjectID) (*Policy, error) {
	filter := bson.M{"_id": id, "user_id": userID.String()}

	var doc policyDoc
	err := r.coll.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rules RuleNode
	if len(doc.Rules) != 0 {
		rules, err = UnmarshalRuleNodeBSON(doc.Rules)
		if err != nil {
			return nil, fmt.Errorf("policy %s: decode rules: %w", doc.ID.Hex(), err)
		}
	}

	return &Policy{
		ID:          doc.ID,
		UserID:      doc.UserID,
		Marketplace: doc.Marketplace,
		Name:        doc.Name,
		Type:        doc.Type,
		Rules:       rules,
	}, nil
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

type policyDoc struct {
	ID          primitive.ObjectID `bson:"_id"`
	UserID      string             `bson:"user_id"`
	Marketplace string             `bson:"marketplace"`
	Name        string             `bson:"name"`
	Type        string             `bson:"type"`

	// Capture the recursive tree without decoding it yet
	Rules bson.Raw `bson:"rules"`
}

// ListPoliciesWithMarketplace lists policies for a user, optionally filtered by marketplace.
func (r *MongoPolicyRepository) ListPoliciesWithMarketplace(ctx context.Context, userID uuid.UUID, marketplace *string) ([]*Policy, error) {
	filter := bson.M{"user_id": userID.String()}
	if marketplace != nil {
		if m := strings.TrimSpace(*marketplace); m != "" {
			filter["marketplace"] = m
		}
	}
	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		if err := cur.Close(ctx); err != nil {
			log.Printf("failed to close cursor: %v\n", err)
		}
	}(cur, ctx)

	// Decode straight into policyDoc (non-recursive fields + raw rules)
	var docs []policyDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}

	policies := make([]*Policy, 0, len(docs))
	for _, d := range docs {
		var root RuleNode
		if len(d.Rules) != 0 {
			var decErr error
			root, decErr = UnmarshalRuleNodeBSON(d.Rules)
			if decErr != nil {
				return nil, fmt.Errorf("policy %s: decode rules: %w", d.ID.Hex(), decErr)
			}
		}

		policies = append(policies, &Policy{
			ID:          d.ID,
			UserID:      d.UserID,
			Marketplace: d.Marketplace,
			Name:        d.Name,
			Type:        d.Type,
			Rules:       root,
		})
	}

	return policies, nil
}

func (r *MongoPolicyRepository) ListPolicies(ctx context.Context, userID uuid.UUID) ([]*Policy, error) {
	return r.ListPoliciesWithMarketplace(ctx, userID, nil)
}

// ListPoliciesByMarketplace retrieves all policies for a specific marketplace.
func (r *MongoPolicyRepository) ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*Policy, error) {
	return r.ListPoliciesWithMarketplace(ctx, userID, &marketplace)
}

// DeletePolicy deletes a policy by its ID.
func (r *MongoPolicyRepository) DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (*Policy, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	filter := bson.M{"_id": objID, "user_id": userID.String()}

	// Use FindOneAndDelete to retrieve the document and validate/parse rules similarly
	var doc policyDoc
	res := r.coll.FindOneAndDelete(ctx, filter)
	if err := res.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	rules, decErr := UnmarshalRuleNodeBSON(doc.Rules)
	if decErr != nil {
		return nil, fmt.Errorf("policy %s: decode rules after delete: %w", doc.ID.Hex(), decErr)
	}
	return &Policy{
		ID:          doc.ID,
		UserID:      doc.UserID,
		Marketplace: doc.Marketplace,
		Name:        doc.Name,
		Type:        doc.Type,
		Rules:       rules,
	}, nil
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
