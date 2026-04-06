package clickup_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestAdapter_TestConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "1"})
	}))
	defer srv.Close()

	a := clickup.NewAdapterWithBaseURL("test-key", srv.URL)
	if err := a.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestAdapter_TestConnection_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	a := clickup.NewAdapterWithBaseURL("bad-key", srv.URL)
	err := a.TestConnection(context.Background())
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAdapter_FetchTasks_Success(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if page > 0 {
			_ = json.NewEncoder(w).Encode(map[string][]string{"tasks": {}})
			return
		}
		page++
		resp := map[string]interface{}{
			"tasks": []map[string]interface{}{
				{
					"id":           "t1",
					"name":         "Task 1",
					"status":       map[string]string{"status": "open"},
					"assignees":    []map[string]string{},
					"points":       nil,
					"tags":         []map[string]string{},
					"date_created": "1700000000000",
					"date_updated": "1700100000000",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := clickup.NewAdapterWithBaseURL("key", srv.URL)
	tasks, err := a.FetchTasks(context.Background(), "list1")
	if err != nil {
		t.Fatalf("FetchTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Task 1" {
		t.Errorf("expected 'Task 1', got %q", tasks[0].Title)
	}
	if tasks[0].Status != "open" {
		t.Errorf("expected 'open', got %q", tasks[0].Status)
	}
}

func TestAdapter_FetchTasks_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	a := clickup.NewAdapterWithBaseURL("key", srv.URL)
	_, err := a.FetchTasks(context.Background(), "list1")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestAdapter_FetchSprints_NotSupported(t *testing.T) {
	a := clickup.NewAdapterWithBaseURL("key", "http://unused")
	_, err := a.FetchSprints(context.Background(), "list1")
	if err != adapter.ErrSprintsNotSupported {
		t.Errorf("expected ErrSprintsNotSupported, got %v", err)
	}
}
