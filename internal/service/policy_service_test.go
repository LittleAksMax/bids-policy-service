package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LittleAksMax/policy-service/internal/repository"
)

type mockRepo struct {
	policy *repository.Policy
	called bool
}

func (m *mockRepo) GetPolicy(ctx context.Context, id string) (*repository.Policy, error) {
	m.called = true
	if m.policy != nil && m.policy.ID == id {
		return m.policy, nil
	}
	return nil, errors.New("not found")
}
func (m *mockRepo) CreatePolicy(ctx context.Context, p *repository.Policy) error   { return nil }
func (m *mockRepo) ListPolicies(ctx context.Context) ([]*repository.Policy, error) { return nil, nil }
func (m *mockRepo) UpdatePolicy(ctx context.Context, p *repository.Policy) error   { return nil }

type mockCache struct {
	store   map[string]string
	expires map[string]time.Time
}

func (m *mockCache) Save(ctx context.Context, token string, userID string, expiresAt time.Time) error {
	m.store[token] = userID
	m.expires[token] = expiresAt
	return nil
}
func (m *mockCache) Get(ctx context.Context, token string) (string, time.Time, error) {
	v, ok := m.store[token]
	e, ok2 := m.expires[token]
	if ok && ok2 {
		return v, e, nil
	}
	return "", time.Time{}, errors.New("not found")
}
func (m *mockCache) Delete(ctx context.Context, token string) error {
	delete(m.store, token)
	delete(m.expires, token)
	return nil
}
func (m *mockCache) HealthCheck(ctx context.Context) error { return nil }

func TestGetPolicy_CacheHit(t *testing.T) {
	p := &repository.Policy{ID: "123", UserID: "u1", Name: "n", Rules: []string{"r"}}
	cacheData, _ := json.Marshal(p)
	mc := &mockCache{store: map[string]string{"123": string(cacheData)}, expires: map[string]time.Time{"123": time.Now().Add(1 * time.Hour)}}
	repo := &mockRepo{policy: p}
	svc := NewPolicyService(repo, mc)
	got, err := svc.GetPolicy(context.Background(), "123")
	if err != nil || got.ID != "123" {
		t.Fatalf("expected cache hit, got %v, err %v", got, err)
	}
	if repo.called {
		t.Fatalf("expected repo not called on cache hit")
	}
}

func TestGetPolicy_CacheMiss(t *testing.T) {
	p := &repository.Policy{ID: "123", UserID: "u1", Name: "n", Rules: []string{"r"}}
	mc := &mockCache{store: map[string]string{}, expires: map[string]time.Time{}}
	repo := &mockRepo{policy: p}
	svc := NewPolicyService(repo, mc)
	got, err := svc.GetPolicy(context.Background(), "123")
	if err != nil || got.ID != "123" {
		t.Fatalf("expected repo hit, got %v, err %v", got, err)
	}
	if !repo.called {
		t.Fatalf("expected repo called on cache miss")
	}
	// Should now be cached
	cached, _, err := mc.Get(context.Background(), "123")
	if err != nil || cached == "" {
		t.Fatalf("expected cache to be set after miss")
	}
}
