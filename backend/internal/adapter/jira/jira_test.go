package jira_test

import (
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/adapter/jira"
)

func TestNewAdapter_MissingConfig(t *testing.T) {
	_, err := jira.NewAdapter("token", nil)
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNewAdapter_EmptyToken(t *testing.T) {
	_, err := jira.NewAdapter("", map[string]string{"base_url": "https://test.atlassian.net", "email": "test@example.com"})
	if err != adapter.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestNewAdapter_Valid(t *testing.T) {
	a, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://test.atlassian.net",
		"email":    "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "jira" {
		t.Errorf("expected name 'jira', got %q", a.Name())
	}
}

func TestNewAdapter_SSRF_HttpBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "http://test.atlassian.net",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for http:// URL")
	}
}

func TestNewAdapter_SSRF_NonAtlassianBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://169.254.169.254",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for non-atlassian domain")
	}
}

func TestNewAdapter_InvalidStoryPointsField(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url":           "https://test.atlassian.net",
		"email":              "test@example.com",
		"story_points_field": "field;DROP TABLE",
	})
	if err == nil {
		t.Fatal("expected error for invalid field name")
	}
}

func TestNewAdapter_ValidStoryPointsField(t *testing.T) {
	a, err := jira.NewAdapter("token", map[string]string{
		"base_url":           "https://test.atlassian.net",
		"email":              "test@example.com",
		"story_points_field": "customfield_10020",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "jira" {
		t.Errorf("expected name 'jira', got %q", a.Name())
	}
}

func TestNewAdapter_SSRF_LocalhostBlocked(t *testing.T) {
	_, err := jira.NewAdapter("token", map[string]string{
		"base_url": "https://localhost",
		"email":    "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for localhost")
	}
}
