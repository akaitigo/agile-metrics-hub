package clickup_test

import (
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/adapter/clickup"
)

func TestNewAdapter_EmptyKey(t *testing.T) {
	_, err := clickup.NewAdapter("", nil)
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNewAdapter_ValidKey(t *testing.T) {
	a, err := clickup.NewAdapter("pk_test_key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "clickup" {
		t.Errorf("expected name 'clickup', got %q", a.Name())
	}
}
