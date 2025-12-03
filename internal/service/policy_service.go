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

func NewPolicyService(repo repository.PolicyRepository, cache cache.RequestCache) *PolicyService {
	return &PolicyService{repo: repo, cache: cache}
}

func (s *PolicyService) GetPolicy(ctx context.Context, id string) (*repository.Policy, error) {
	// Try cache first
	cached, expiresAt, err := s.cache.Get(ctx, id)
	if err == nil && cached != "" && expiresAt.After(time.Now()) {
		var policy repository.Policy
		if err := json.Unmarshal([]byte(cached), &policy); err == nil {
			return &policy, nil
		}
	}
	// Not found or expired, get from repo
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return nil, err
	}
	// Save to cache for future requests (cache for 5 minutes)
	data, _ := json.Marshal(policy)
	expires := time.Now().Add(5 * time.Minute)
	_ = s.cache.Save(ctx, id, string(data), expires)
	return policy, nil
}

func (s *PolicyService) CreatePolicy(ctx context.Context, p *repository.Policy) error {
	return s.repo.CreatePolicy(ctx, p)
}

func (s *PolicyService) ListPolicies(ctx context.Context) ([]*repository.Policy, error) {
	return s.repo.ListPolicies(ctx)
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, p *repository.Policy) error {
	err := s.repo.UpdatePolicy(ctx, p)
	if err == nil {
		_ = s.cache.Delete(ctx, p.ID) // Invalidate cache for updated policy
	}
	return err
}
