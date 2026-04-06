package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

// Adapter はJira Cloud REST API v3のアダプター実装。
type Adapter struct {
	baseURL  string
	email    string
	apiToken string
	client   *http.Client
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
	return &Adapter{
		baseURL:  strings.TrimRight(baseURL, "/"),
		email:    email,
		apiToken: apiToken,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
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
	host := parsed.Hostname()
	// Atlassian Cloud ドメインのみ許可
	if !strings.HasSuffix(host, ".atlassian.net") {
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

	body, err := io.ReadAll(resp.Body)
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
		return nil, fmt.Errorf("jira API error: %d %s", resp.StatusCode, string(body))
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

	for {
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
	Issues []jiraIssue `json:"issues"`
	Total  int         `json:"total"`
}

type jiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		StoryPoints *float64 `json:"story_points"`
		Priority    *struct {
			Name string `json:"name"`
		} `json:"priority"`
		Labels         []string `json:"labels"`
		Created        string   `json:"created"`
		Updated        string   `json:"updated"`
		Resolutiondate *string  `json:"resolutiondate"`
	} `json:"fields"`
}

func (a *Adapter) FetchTasks(ctx context.Context, projectID string) ([]model.Task, error) {
	var allTasks []model.Task
	startAt := 0

	for {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/rest/agile/1.0/board/%s/issue?startAt=%d&maxResults=100&fields=summary,status,assignee,story_points,priority,labels,created,updated,resolutiondate",
				projectID, startAt))
		if err != nil {
			return nil, fmt.Errorf("fetch issues: %w", err)
		}

		var resp jiraSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse issues: %w", err)
		}

		for _, issue := range resp.Issues {
			task := convertJiraIssue(issue, projectID)
			allTasks = append(allTasks, task)
		}

		startAt += len(resp.Issues)
		if startAt >= resp.Total || len(resp.Issues) == 0 {
			break
		}
	}
	return allTasks, nil
}

func convertJiraIssue(issue jiraIssue, projectID string) model.Task {
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

	task.StoryPoints = issue.Fields.StoryPoints

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

	for {
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

	for {
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
