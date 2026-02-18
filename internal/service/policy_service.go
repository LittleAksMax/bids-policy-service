package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/google/uuid"
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
	UpdatePolicy(ctx context.Context, userID uuid.UUID, p *repository.Policy) error
	DeletePolicy(ctx context.Context, userID uuid.UUID, id string) error
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
	policy, err := s.repo.GetPolicy(ctx, userID, id)
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
func (s *PolicyService) UpdatePolicy(ctx context.Context, userID uuid.UUID, p *repository.Policy) error {
	err := s.repo.UpdatePolicy(ctx, userID, p)
	if err == nil {
		_ = s.cache.Delete(ctx, userID.String()+":policy:"+p.ID.Hex()) // Invalidate cache for updated policy
	}
	return err
}

// DeletePolicy deletes a policy.
func (s *PolicyService) DeletePolicy(ctx context.Context, userID uuid.UUID, id string) error {
	err := s.repo.DeletePolicy(ctx, userID, id)
	if err == nil {
		_ = s.cache.Delete(ctx, userID.String()+":policy:"+id) // Invalidate cache for deleted policy
	}
	return err
}
