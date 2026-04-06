package clickup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

const baseURL = "https://api.clickup.com/api/v2"

// Adapter はClickUp API v2のアダプター実装。
type Adapter struct {
	apiKey string
	client *http.Client
}

// NewAdapter はClickUpアダプターを生成する。
func NewAdapter(apiKey string, _ map[string]string) (adapter.PMToolAdapter, error) {
	if apiKey == "" {
		return nil, adapter.ErrUnauthorized
	}
	return &Adapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (a *Adapter) Name() string { return "clickup" }

func (a *Adapter) doRequest(ctx context.Context, method, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", a.apiKey)
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
	case http.StatusUnauthorized:
		return nil, adapter.ErrUnauthorized
	case http.StatusTooManyRequests:
		return nil, adapter.ErrRateLimited
	case http.StatusNotFound:
		return nil, adapter.ErrNotFound
	default:
		return nil, fmt.Errorf("clickup API error: status %d", resp.StatusCode)
	}
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	_, err := a.doRequest(ctx, http.MethodGet, "/user")
	return err
}

// clickupTeam はClickUp API のチームレスポンス。
type clickupTeamsResponse struct {
	Teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"teams"`
}

type clickupSpacesResponse struct {
	Spaces []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"spaces"`
}

type clickupListsResponse struct {
	Lists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"lists"`
}

func (a *Adapter) FetchProjects(ctx context.Context) ([]model.Project, error) {
	body, err := a.doRequest(ctx, http.MethodGet, "/team")
	if err != nil {
		return nil, fmt.Errorf("fetch teams: %w", err)
	}
	var teams clickupTeamsResponse
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, fmt.Errorf("parse teams: %w", err)
	}

	var projects []model.Project
	for _, team := range teams.Teams {
		body, err := a.doRequest(ctx, http.MethodGet, fmt.Sprintf("/team/%s/space?archived=false", team.ID))
		if err != nil {
			return nil, fmt.Errorf("fetch spaces for team %s: %w", team.ID, err)
		}
		var spaces clickupSpacesResponse
		if err := json.Unmarshal(body, &spaces); err != nil {
			return nil, fmt.Errorf("parse spaces: %w", err)
		}

		for _, space := range spaces.Spaces {
			body, err := a.doRequest(ctx, http.MethodGet, fmt.Sprintf("/space/%s/list?archived=false", space.ID))
			if err != nil {
				continue // skip inaccessible spaces
			}
			var lists clickupListsResponse
			if err := json.Unmarshal(body, &lists); err != nil {
				continue
			}
			for _, list := range lists.Lists {
				projects = append(projects, model.Project{
					ExternalID: list.ID,
					Source:     "clickup",
					Name:       fmt.Sprintf("%s / %s / %s", team.Name, space.Name, list.Name),
				})
			}
		}
	}
	return projects, nil
}

type clickupTasksResponse struct {
	Tasks []clickupTask `json:"tasks"`
}

type clickupTask struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Status string `json:"status"`
	} `json:"status"`
	Assignees []struct {
		Username string `json:"username"`
	} `json:"assignees"`
	Points   json.RawMessage `json:"points"`
	Priority *struct {
		Priority string `json:"priority"`
	} `json:"priority"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	DateCreated string  `json:"date_created"`
	DateUpdated string  `json:"date_updated"`
	DateClosed  *string `json:"date_closed"`
}

func (a *Adapter) FetchTasks(ctx context.Context, projectID string) ([]model.Task, error) {
	var allTasks []model.Task
	const maxPages = 500
	page := 0

	for page < maxPages {
		body, err := a.doRequest(ctx, http.MethodGet,
			fmt.Sprintf("/list/%s/task?page=%d&include_closed=true&subtasks=true", projectID, page))
		if err != nil {
			return nil, fmt.Errorf("fetch tasks page %d: %w", page, err)
		}

		var resp clickupTasksResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse tasks: %w", err)
		}

		if len(resp.Tasks) == 0 {
			break
		}

		for _, ct := range resp.Tasks {
			task := convertClickUpTask(ct, projectID)
			allTasks = append(allTasks, task)
		}
		page++
	}
	return allTasks, nil
}

func convertClickUpTask(ct clickupTask, projectID string) model.Task {
	task := model.Task{
		ExternalID: ct.ID,
		Source:     "clickup",
		ProjectID:  projectID,
		Title:      ct.Name,
		Status:     ct.Status.Status,
	}

	if len(ct.Assignees) > 0 {
		task.Assignee = ct.Assignees[0].Username
	}

	if ct.Points != nil {
		var points float64
		if json.Unmarshal(ct.Points, &points) == nil && points > 0 {
			task.StoryPoints = &points
		}
	}

	if ct.Priority != nil {
		task.Priority = ct.Priority.Priority
	}

	for _, tag := range ct.Tags {
		task.Labels = append(task.Labels, tag.Name)
	}

	if ts, err := strconv.ParseInt(ct.DateCreated, 10, 64); err == nil {
		task.CreatedAt = time.UnixMilli(ts)
	}
	if ts, err := strconv.ParseInt(ct.DateUpdated, 10, 64); err == nil {
		task.UpdatedAt = time.UnixMilli(ts)
	}
	if ct.DateClosed != nil {
		if ts, err := strconv.ParseInt(*ct.DateClosed, 10, 64); err == nil {
			closed := time.UnixMilli(ts)
			task.CompletedAt = &closed
		}
	}

	return task
}

func (a *Adapter) FetchSprints(_ context.Context, _ string) ([]model.Sprint, error) {
	return nil, adapter.ErrSprintsNotSupported
}

// timeInStatusResponse はClickUp Time in Status APIのレスポンス。
type timeInStatusResponse struct {
	CurrentStatus struct {
		Status    string `json:"status"`
		TotalTime struct {
			ByMinute int64  `json:"by_minute"`
			Since    string `json:"since"`
		} `json:"total_time"`
	} `json:"current_status"`
	StatusHistory []struct {
		Status    string `json:"status"`
		TotalTime struct {
			ByMinute int64  `json:"by_minute"`
			Since    string `json:"since"`
		} `json:"total_time"`
		OrderIndex int `json:"orderindex"`
	} `json:"status_history"`
}

func (a *Adapter) FetchTaskHistory(ctx context.Context, taskID string) ([]model.TaskEvent, error) {
	body, err := a.doRequest(ctx, http.MethodGet, fmt.Sprintf("/task/%s/time_in_status", taskID))
	if err != nil {
		return nil, fmt.Errorf("fetch time in status: %w", err)
	}

	var resp timeInStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse time in status: %w", err)
	}

	var events []model.TaskEvent
	for i, hist := range resp.StatusHistory {
		since, err := strconv.ParseInt(hist.TotalTime.Since, 10, 64)
		if err != nil {
			continue
		}
		event := model.TaskEvent{
			TaskID:    taskID,
			ToStatus:  hist.Status,
			Timestamp: time.UnixMilli(since),
			Source:    "time_in_status",
		}
		if i > 0 {
			event.FromStatus = resp.StatusHistory[i-1].Status
		}
		events = append(events, event)
	}

	return events, nil
}
