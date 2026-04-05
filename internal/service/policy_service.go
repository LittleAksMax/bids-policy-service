package service

import (
	"context"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PolicyService provides business logic for policies.
type PolicyService struct {
	repo repository.PolicyRepository
}

// PolicyServiceInterface defines the contract for policy service logic.
type PolicyServiceInterface interface {
	GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*repository.Policy, error)
	CreatePolicy(ctx context.Context, userID uuid.UUID, marketplace, name, script string) (*repository.Policy, error)
	ListPolicies(ctx context.Context, userID uuid.UUID) ([]*repository.Policy, error)
	ListPoliciesByMarketplace(ctx context.Context, userID uuid.UUID, marketplace string) ([]*repository.Policy, error)
	UpdatePolicy(ctx context.Context, userID uuid.UUID, id, name, script string) (*repository.Policy, error)
	DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (bool, error)
}

// NewPolicyService creates a new PolicyService.
func NewPolicyService(repo repository.PolicyRepository) *PolicyService {
	return &PolicyService{repo: repo}
}

// GetPolicy retrieves a policy by its ID.
func (s *PolicyService) GetPolicy(ctx context.Context, userID uuid.UUID, id string) (*repository.Policy, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	return s.repo.GetPolicy(ctx, userID, objID)
}

// CreatePolicy creates a new policy.
func (s *PolicyService) CreatePolicy(ctx context.Context, userID uuid.UUID, marketplace, name, script string) (*repository.Policy, error) {
	p := &repository.Policy{
		UserID:      userID.String(),
		Marketplace: marketplace,
		Name:        name,
		Script:      script,
	}
	err := s.repo.CreatePolicy(ctx, p)
	return p, err
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
func (s *PolicyService) UpdatePolicy(ctx context.Context, userID uuid.UUID, id, name, script string) (*repository.Policy, error) {
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
	existing.Script = script
	return existing, s.repo.UpdatePolicy(ctx, userID, existing)
}

// DeletePolicy deletes a policy.
func (s *PolicyService) DeletePolicy(ctx context.Context, userID uuid.UUID, id string) (bool, error) {
	policy, err := s.repo.DeletePolicy(ctx, userID, id)
	return policy != nil, err
}
