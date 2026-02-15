package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LittleAksMax/bids-policy-service/internal/repository"
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

func (m *mockCache) Save(ctx context.Context, key string, value string, expiresAt time.Time) error {
	m.store[key] = value
	m.expires[key] = expiresAt
	return nil
}
func (m *mockCache) Get(ctx context.Context, key string) (string, time.Time, error) {
	println("mockCache.Get called with key=", key)
	for k, v := range m.store {
		println("store[", k, "] = ", v)
	}
	v, ok := m.store[key]
	e, ok2 := m.expires[key]
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
	t.Parallel()
	key := "cachehit-123"
	p := &repository.Policy{ID: key, UserID: "u1", Name: "n", Rules: &repository.Terminal{Type: "terminal", Op: repository.RuleOp{Type: "add", Amount: repository.RuleAmount{Neg: false, Amount: 1, Perc: false}}}}
	cacheData, _ := json.Marshal(p)
	expiresAt := time.Now().Add(1 * time.Hour)
	mc := &mockCache{store: make(map[string]string), expires: make(map[string]time.Time)}
	mc.store[key] = string(cacheData)
	mc.expires[key] = expiresAt
	repo := &mockRepo{policy: p}
	svc := NewPolicyService(repo, mc)
	got, err := svc.GetPolicy(context.Background(), key)
	if err != nil || got.ID != key {
		t.Fatalf("expected cache hit, got %v, err %v", got, err)
	}
	if repo.called {
		t.Fatalf("expected repo not called on cache hit; cacheData=%s expiresAt=%v now=%v repo.called=%v", cacheData, expiresAt, time.Now(), repo.called)
	}
}

func TestGetPolicy_CacheMiss(t *testing.T) {
	t.Parallel()
	key := "cachemiss-123"
	p := &repository.Policy{ID: key, UserID: "u1", Name: "n", Rules: &repository.Terminal{Type: "terminal", Op: repository.RuleOp{Type: "add", Amount: repository.RuleAmount{Neg: false, Amount: 1, Perc: false}}}}
	mc := &mockCache{store: make(map[string]string), expires: make(map[string]time.Time)}
	repo := &mockRepo{policy: p}
	svc := NewPolicyService(repo, mc)
	got, err := svc.GetPolicy(context.Background(), key)
	if err != nil || got.ID != key {
		t.Fatalf("expected repo hit, got %v, err %v", got, err)
	}
	if !repo.called {
		t.Fatalf("expected repo called on cache miss")
	}
	// Should now be cached
	cached, _, err := mc.Get(context.Background(), key)
	if err != nil || cached == "" {
		t.Fatalf("expected cache to be set after miss")
	}
}

func TestUpdatePolicy_DeletesCache(t *testing.T) {
	t.Parallel()
	key := "update-123"
	p := &repository.Policy{ID: key, UserID: "u1", Name: "n", Rules: &repository.Terminal{Type: "terminal", Op: repository.RuleOp{Type: "add", Amount: repository.RuleAmount{Neg: false, Amount: 1, Perc: false}}}}
	mc := &mockCache{store: make(map[string]string), expires: make(map[string]time.Time)}
	mc.store[key] = "cached"
	mc.expires[key] = time.Now().Add(1 * time.Hour)
	repo := &mockRepo{policy: p}
	svc := NewPolicyService(repo, mc)
	err := svc.UpdatePolicy(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := mc.store[key]; ok {
		t.Fatalf("expected cache to be deleted after update")
	}
}
