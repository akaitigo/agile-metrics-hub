package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

var validFieldName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

const maxPaginationIterations = 500

// Adapter はJira Cloud REST API v3のアダプター実装。
type Adapter struct {
	baseURL          string
	email            string
	apiToken         string
	storyPointsField string
	client           *http.Client
}

// NewAdapter はJiraアダプターを生成する。
// config に "base_url" と "email" が必要。
func NewAdapter(apiToken string, config map[string]string) (adapter.PMToolAdapter, error) {
	baseURL := config["base_url"]
	email := config["email"]
	if apiToken == "" || baseURL == "" || email == "" {
		return nil, adapter.ErrUnauthorized
	}
	if err := validateBaseURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base_url: %w", err)
	}
	spField := config["story_points_field"]
	if spField == "" {
		spField = "customfield_10016"
	}
	if !validFieldName.MatchString(spField) {
		return nil, fmt.Errorf("invalid story_points_field: %q", spField)
	}
	return &Adapter{
		baseURL:          strings.TrimRight(baseURL, "/"),
		email:            email,
		apiToken:         apiToken,
		storyPointsField: spField,
		client:           &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// NewAdapterWithBaseURL はテスト用にbaseURLを差し替えたアダプターを生成する。
func NewAdapterWithBaseURL(baseURL, email, apiToken string) *Adapter {
	return &Adapter{
		baseURL:          baseURL,
		email:            email,
		apiToken:         apiToken,
		storyPointsField: "customfield_10016",
		client:           &http.Client{Timeout: 5 * time.Second},
	}
}

// validateBaseURL はSSRF防止のためbase_urlを検証する。
func validateBaseURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("only https is allowed, got %q", parsed.Scheme)
	}
	if parsed.User != nil {
		return fmt.Errorf("userinfo in URL is not allowed")
	}
	host := parsed.Hostname()
	// Atlassian Cloud ドメインのみ許可（厳密なドメイン照合）
	if host != "atlassian.net" && !strings.HasSuffix(host, ".atlassian.net") {
		return fmt.Errorf("only *.atlassian.net domains are allowed, got %q", host)
	}
	return nil
}

func (a *Adapter) Name() string { return "jira" }

func (a *Adapter) doRequest(ctx context.Context, method, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(a.email, a.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	const maxResponseSize = 10 << 20 // 10MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, adapter.ErrUnauthorized
	case http.StatusTooManyRequests:
		return nil, adapter.ErrRateLimited
	case http.StatusNotFound:
		return nil, adapter.ErrNotFound
	default:
		return nil, fmt.Errorf("jira API error: status %d", resp.StatusCode)
	}
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	_, err := a.doRequest(ctx, http.MethodGet, "/rest/api/3/myself")
	return err
}

type jiraBoardsResponse struct {
	Values []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"values"`
	IsLast bool `json:"isLast"`
}

func (a *Adapter) FetchProjects(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	startAt := 0

	for range maxPaginationIterations {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/agile/1.0/board?startAt=%d&maxResults=50", startAt))
		if err != nil {
			return nil, fmt.Errorf("fetch boards: %w", err)
		}

		var resp jiraBoardsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse boards: %w", err)
		}

		for _, board := range resp.Values {
			projects = append(projects, model.Project{
				ExternalID: fmt.Sprintf("%d", board.ID),
				Source:     "jira",
				Name:       fmt.Sprintf("%s (%s)", board.Name, board.Type),
			})
		}

		if resp.IsLast || len(resp.Values) == 0 {
			break
		}
		startAt += len(resp.Values)
	}
	return projects, nil
}

type jiraSearchResponse struct {
	Issues []jiraIssueRaw `json:"issues"`
	Total  int            `json:"total"`
}

// jiraIssueRaw は2段階デシリアライズ用。fields を RawMessage で受ける。
type jiraIssueRaw struct {
	ID     string          `json:"id"`
	Key    string          `json:"key"`
	Fields json.RawMessage `json:"fields"`
}

type jiraIssueFields struct {
	Summary string `json:"summary"`
	Status  struct {
		Name string `json:"name"`
	} `json:"status"`
	Assignee *struct {
		DisplayName string `json:"displayName"`
	} `json:"assignee"`
	Priority *struct {
		Name string `json:"name"`
	} `json:"priority"`
	Labels         []string `json:"labels"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
	Resolutiondate *string  `json:"resolutiondate"`
}

type jiraIssue struct {
	ID        string
	Key       string
	Fields    jiraIssueFields
	RawFields json.RawMessage
}

func (a *Adapter) FetchTasks(ctx context.Context, projectID string) ([]model.Task, error) {
	var allTasks []model.Task
	startAt := 0

	for range maxPaginationIterations {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/agile/1.0/board/%s/issue?startAt=%d&maxResults=100&fields=summary,status,assignee,%s,priority,labels,created,updated,resolutiondate",
				projectID, startAt, a.storyPointsField))
		if err != nil {
			return nil, fmt.Errorf("fetch issues: %w", err)
		}

		var resp jiraSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse issues: %w", err)
		}

		for _, raw := range resp.Issues {
			var fields jiraIssueFields
			if err := json.Unmarshal(raw.Fields, &fields); err != nil {
				continue
			}
			issue := jiraIssue{ID: raw.ID, Key: raw.Key, Fields: fields, RawFields: raw.Fields}
			task := convertJiraIssue(issue, projectID, a.storyPointsField)
			allTasks = append(allTasks, task)
		}

		startAt += len(resp.Issues)
		if startAt >= resp.Total || len(resp.Issues) == 0 {
			break
		}
	}
	return allTasks, nil
}

func convertJiraIssue(issue jiraIssue, projectID, storyPointsField string) model.Task {
	task := model.Task{
		ExternalID: issue.Key,
		Source:     "jira",
		ProjectID:  projectID,
		Title:      issue.Fields.Summary,
		Status:     issue.Fields.Status.Name,
		Labels:     issue.Fields.Labels,
	}

	if issue.Fields.Assignee != nil {
		task.Assignee = issue.Fields.Assignee.DisplayName
	}

	// ストーリーポイントをカスタムフィールドから動的に取得
	if issue.RawFields != nil {
		var fields map[string]json.RawMessage
		if json.Unmarshal(issue.RawFields, &fields) == nil {
			if raw, ok := fields[storyPointsField]; ok {
				var pts float64
				if json.Unmarshal(raw, &pts) == nil && pts > 0 {
					task.StoryPoints = &pts
				}
			}
		}
	}

	if issue.Fields.Priority != nil {
		task.Priority = issue.Fields.Priority.Name
	}

	if t, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Created); err == nil {
		task.CreatedAt = t
	}
	if t, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Updated); err == nil {
		task.UpdatedAt = t
	}
	if issue.Fields.Resolutiondate != nil {
		if t, err := time.Parse("2006-01-02T15:04:05.000-0700", *issue.Fields.Resolutiondate); err == nil {
			task.CompletedAt = &t
		}
	}

	return task
}

type jiraSprintsResponse struct {
	Values []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		State        string `json:"state"`
		StartDate    string `json:"startDate"`
		EndDate      string `json:"endDate"`
		CompleteDate string `json:"completeDate"`
	} `json:"values"`
	IsLast bool `json:"isLast"`
}

func (a *Adapter) FetchSprints(ctx context.Context, projectID string) ([]model.Sprint, error) {
	var sprints []model.Sprint
	startAt := 0

	for range maxPaginationIterations {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/agile/1.0/board/%s/sprint?startAt=%d&maxResults=50", projectID, startAt))
		if err != nil {
			return nil, fmt.Errorf("fetch sprints: %w", err)
		}

		var resp jiraSprintsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse sprints: %w", err)
		}

		for _, s := range resp.Values {
			sprint := model.Sprint{
				ExternalID: fmt.Sprintf("%d", s.ID),
				Source:     "jira",
				ProjectID:  projectID,
				Name:       s.Name,
				Status:     mapJiraSprintState(s.State),
			}
			if t, err := time.Parse("2006-01-02T15:04:05.000Z", s.StartDate); err == nil {
				sprint.StartDate = t
			}
			if t, err := time.Parse("2006-01-02T15:04:05.000Z", s.EndDate); err == nil {
				sprint.EndDate = t
			}
			sprints = append(sprints, sprint)
		}

		if resp.IsLast || len(resp.Values) == 0 {
			break
		}
		startAt += len(resp.Values)
	}
	return sprints, nil
}

func mapJiraSprintState(state string) string {
	switch state {
	case "active":
		return "active"
	case "closed":
		return "closed"
	default:
		return "planned"
	}
}

type jiraChangelogResponse struct {
	Values []struct {
		Created string `json:"created"`
		Items   []struct {
			Field      string `json:"field"`
			FromString string `json:"fromString"`
			ToString   string `json:"toString"`
		} `json:"items"`
	} `json:"values"`
	IsLast bool `json:"isLast"`
}

func (a *Adapter) FetchTaskHistory(ctx context.Context, taskID string) ([]model.TaskEvent, error) {
	var events []model.TaskEvent
	startAt := 0

	for range maxPaginationIterations {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/api/3/issue/%s/changelog?startAt=%d&maxResults=100", taskID, startAt))
		if err != nil {
			return nil, fmt.Errorf("fetch changelog: %w", err)
		}

		var resp jiraChangelogResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse changelog: %w", err)
		}

		for _, entry := range resp.Values {
			for _, item := range entry.Items {
				if item.Field != "status" {
					continue
				}
				ts, err := time.Parse("2006-01-02T15:04:05.000-0700", entry.Created)
				if err != nil {
					continue
				}
				events = append(events, model.TaskEvent{
					TaskID:     taskID,
					FromStatus: item.FromString,
					ToStatus:   item.ToString,
					Timestamp:  ts,
					Source:     "api",
				})
			}
		}

		if resp.IsLast || len(resp.Values) == 0 {
			break
		}
		startAt += len(resp.Values)
	}
	return events, nil
}
