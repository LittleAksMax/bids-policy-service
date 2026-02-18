package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PolicyService provides business logic for policies, with cache support.
type PolicyService struct {
	repo  repository.PolicyRepository
	cache cache.RequestCache
}

// PolicyServiceInterface defines the contract for policy service logic.
type PolicyServiceInterface interface {
	GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*repository.Policy, error)
	CreatePolicy(ctx context.Context, p *repository.Policy) error
	ListPolicies(ctx context.Context, userID uuid.UUID) ([]*repository.Policy, error)
	ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*repository.Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, id, name string, rules repository.RuleNode) (*repository.Policy, error)
	DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (bool, error)
}

// NewPolicyService creates a new PolicyService
func NewPolicyService(repo repository.PolicyRepository, cache cache.RequestCache) *PolicyService {
	return &PolicyService{repo: repo, cache: cache}
}

// GetPolicy retrieves a policy by its ID, first checking the cache.
func (s *PolicyService) GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*repository.Policy, error) {
	cacheKey := userID.String() + ":policy:" + id
	cached, expiresAt, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" && expiresAt.After(time.Now()) {
		var policy repository.Policy
		if err := json.Unmarshal([]byte(cached), &policy); err == nil {
			return &policy, nil
		}
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	policy, err := s.repo.GetPolicy(ctx, userID, objID)
	if err != nil || policy == nil {
		return policy, err
	}
	b, _ := json.Marshal(policy)
	_ = s.cache.Set(ctx, cacheKey, string(b), 5*time.Minute)
	return policy, nil
}

// CreatePolicy creates a new policy.
func (s *PolicyService) CreatePolicy(ctx context.Context, p *repository.Policy) error {
	return s.repo.CreatePolicy(ctx, p)
}

// ListPolicies retrieves all policies.
func (s *PolicyService) ListPolicies(ctx context.Context, userID uuid.UUID) ([]*repository.Policy, error) {
	return s.repo.ListPolicies(ctx, userID)
}

// ListPoliciesByMarketplace retrieves all policies for a specific marketplace.
func (s *PolicyService) ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*repository.Policy, error) {
	return s.repo.ListPoliciesByMarketplace(ctx, userID, marketplace)
}

// UpdatePolicy updates an existing policy.
func (s *PolicyService) UpdatePolicy(ctx context.Context, userID uuid.UUID, id, name string, rules repository.RuleNode) (*repository.Policy, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}

	// Get existing, return if error or not found
	existing, err := s.repo.GetPolicy(ctx, userID, objID)
	if err != nil || existing == nil {
		return nil, err
	}

	existing.Name = name
	existing.Rules = rules
	err = s.repo.UpdatePolicy(ctx, userID, existing)
	if err == nil {
		_ = s.cache.Delete(ctx, userID.String()+":policy:"+existing.ID.Hex()) // Invalidate cache for updated policy
	}
	// It's fine to put true, since if err != nil, the update failed, and we handle the error first
	return existing, err
}

// DeletePolicy deletes a policy.
func (s *PolicyService) DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (bool, error) {
	policy, err := s.repo.DeletePolicy(ctx, userID, id)
	if err == nil {
		_ = s.cache.Delete(ctx, userID.String()+":policy:"+id) // Invalidate cache for deleted policy
	}
	return policy != nil, err
}
