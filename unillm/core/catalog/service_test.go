package catalog

import (
	"context"
	"testing"
)

type mockRepo struct {
	providers []Provider
	keys      map[int64][]string
}

func (m *mockRepo) ListActiveProviders(context.Context) ([]Provider, error) {
	return m.providers, nil
}
func (m *mockRepo) ListActiveKeys(_ context.Context, providerID int64) ([]string, error) {
	return m.keys[providerID], nil
}
func (m *mockRepo) FindModelByPublicName(context.Context, string) (*ModelConfig, error) {
	return nil, nil
}
func (m *mockRepo) FindProviderByID(context.Context, int64) (*Provider, error) {
	return nil, nil
}
func (m *mockRepo) ListActiveModels(context.Context) ([]ModelConfig, error) {
	return nil, nil
}

func TestNextKeyRoundRobin(t *testing.T) {
	svc := NewService(&mockRepo{
		providers: []Provider{{ID: 1, Name: "openai"}},
		keys:      map[int64][]string{1: {"k1", "k2", "k3"}},
	})
	if err := svc.Reload(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	got := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		key, err := svc.NextKey(context.Background(), "openai")
		if err != nil {
			t.Fatalf("NextKey failed: %v", err)
		}
		got = append(got, key)
	}

	want := []string{"k1", "k2", "k3", "k1", "k2", "k3"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("key[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNextKeyMissingProvider(t *testing.T) {
	svc := NewService(&mockRepo{})
	if err := svc.Reload(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	_, err := svc.NextKey(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing provider keys")
	}
}
