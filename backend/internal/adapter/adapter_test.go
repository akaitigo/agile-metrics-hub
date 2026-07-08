package adapter_test

import (
	"context"
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

type mockAdapter struct {
	name string
}

func (m *mockAdapter) Name() string { return m.name }
func (m *mockAdapter) TestConnection(_ context.Context) error {
	return nil
}

func (m *mockAdapter) FetchProjects(_ context.Context) ([]model.Project, error) {
	return nil, nil
}

func (m *mockAdapter) FetchTasks(_ context.Context, _ string) ([]model.Task, error) {
	return nil, nil
}

func (m *mockAdapter) FetchSprints(_ context.Context, _ string) ([]model.Sprint, error) {
	return nil, nil
}

func (m *mockAdapter) FetchTaskHistory(_ context.Context, _ string) ([]model.TaskEvent, error) {
	return nil, adapter.ErrHistoryNotSupported
}

func TestRegistry_RegisterAndCreate(t *testing.T) {
	reg := adapter.NewRegistry()

	reg.Register("mock", func(apiKey string, config map[string]string) (adapter.PMToolAdapter, error) {
		return &mockAdapter{name: "mock"}, nil
	})

	a, err := reg.Create("mock", "test-key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "mock" {
		t.Errorf("expected name 'mock', got %q", a.Name())
	}
}

func TestRegistry_CreateUnsupported(t *testing.T) {
	reg := adapter.NewRegistry()

	_, err := reg.Create("nonexistent", "key", nil)
	if err == nil {
		t.Fatal("expected error for unsupported source")
	}
}

func TestRegistry_Sources(t *testing.T) {
	reg := adapter.NewRegistry()
	reg.Register("clickup", func(_ string, _ map[string]string) (adapter.PMToolAdapter, error) { return nil, nil })
	reg.Register("jira", func(_ string, _ map[string]string) (adapter.PMToolAdapter, error) { return nil, nil })

	sources := reg.Sources()
	if len(sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(sources))
	}
}
