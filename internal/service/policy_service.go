package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/cache"
	"github.com/LittleAksMax/bids-policy-service/internal/repository"
)

// PolicyService provides business logic for policies, with cache support.
type PolicyService struct {
	repo  repository.PolicyRepository
	cache cache.RequestCache
}

// PolicyServiceInterface defines the contract for policy service logic.
type PolicyServiceInterface interface {
	GetPolicy(ctx context.Context, id string) (*repository.Policy, error)
	CreatePolicy(ctx context.Context, p *repository.Policy) error
	ListPolicies(ctx context.Context) ([]*repository.Policy, error)
	ListPoliciesByMarketplace(ctx context.Context, marketplace string) ([]*repository.Policy, error)
	UpdatePolicy(ctx context.Context, p *repository.Policy) error
	DeletePolicy(ctx context.Context, id string) error
}

// NewPolicyService creates a new PolicyService
func NewPolicyService(repo repository.PolicyRepository, cache cache.RequestCache) *PolicyService {
	return &PolicyService{repo: repo, cache: cache}
}

// GetPolicy retrieves a policy by its ID, first checking the cache.
func (s *PolicyService) GetPolicy(ctx context.Context, id string) (*repository.Policy, error) {
	cached, expiresAt, err := s.cache.Get(ctx, id)
	if err == nil && cached != "" && expiresAt.After(time.Now()) {
		var policy repository.Policy
		if err := json.Unmarshal([]byte(cached), &policy); err == nil {
			return &policy, nil
		}
	}
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(policy)
	expires := time.Now().Add(5 * time.Minute)
	_ = s.cache.Save(ctx, id, string(data), expires)
	return policy, nil
}

// CreatePolicy creates a new policy.
func (s *PolicyService) CreatePolicy(ctx context.Context, p *repository.Policy) error {
	return s.repo.CreatePolicy(ctx, p)
}

// ListPolicies retrieves all policies.
func (s *PolicyService) ListPolicies(ctx context.Context) ([]*repository.Policy, error) {
	return s.repo.ListPolicies(ctx)
}

// ListPoliciesByMarketplace retrieves all policies for a specific marketplace.
func (s *PolicyService) ListPoliciesByMarketplace(ctx context.Context, marketplace string) ([]*repository.Policy, error) {
	return s.repo.ListPoliciesByMarketplace(ctx, marketplace)
}

// UpdatePolicy updates an existing policy.
func (s *PolicyService) UpdatePolicy(ctx context.Context, p *repository.Policy) error {
	err := s.repo.UpdatePolicy(ctx, p)
	if err == nil {
		_ = s.cache.Delete(ctx, p.ID) // Invalidate cache for updated policy
	}
	return err
}

// DeletePolicy deletes a policy.
func (s *PolicyService) DeletePolicy(ctx context.Context, id string) error {
	err := s.repo.DeletePolicy(ctx, id)
	if err == nil {
		_ = s.cache.Delete(ctx, id) // Invalidate cache for deleted policy
	}
	return err
}
