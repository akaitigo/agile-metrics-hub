package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/metrics"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

// MetricsHandler はメトリクスAPIのHTTPハンドラー。
// MVP段階ではインメモリのデモデータを返す。DB統合は後続Issueで実装。
type MetricsHandler struct{}

// Burndown はバーンダウンチャートデータを返す。
func (h *MetricsHandler) Burndown(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		JSONError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	// デモデータ生成（DB統合後に置き換え）
	now := time.Now()
	start := now.AddDate(0, 0, -14)
	end := now

	var snapshots []metrics.SnapshotData
	totalTasks := 20
	totalPoints := 50.0
	for i := 0; i <= 14; i++ {
		day := start.AddDate(0, 0, i)
		completed := int(float64(i) / 14.0 * float64(totalTasks))
		completedPts := float64(i) / 14.0 * totalPoints
		snapshots = append(snapshots, metrics.SnapshotData{
			Date:            day,
			TotalTasks:      totalTasks,
			CompletedTasks:  completed,
			TotalPoints:     totalPoints,
			CompletedPoints: completedPts,
		})
	}

	result := metrics.CalculateBurndown(snapshots, start, end, totalPoints)
	JSONResponse(w, http.StatusOK, result)
}

// Velocity はベロシティチャートデータを返す。
func (h *MetricsHandler) Velocity(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count := 5
	if countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 && c <= 50 {
			count = c
		}
	}

	// デモデータ
	sprints := make([]metrics.SprintData, 0, count)
	for i := 1; i <= count; i++ {
		sprints = append(sprints, metrics.SprintData{
			Name:            "Sprint " + strconv.Itoa(i),
			CommittedPoints: float64(20 + i),
			CompletedPoints: float64(18 + i),
		})
	}

	points := metrics.CalculateVelocity(sprints)
	avg := metrics.AverageVelocity(sprints, count)
	JSONResponse(w, http.StatusOK, map[string]any{
		"velocity":         points,
		"average_velocity": avg,
	})
}

// CumulativeFlow は累積フローダイアグラムデータを返す。
func (h *MetricsHandler) CumulativeFlow(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// デモデータ
	now := time.Now()
	var snapshots []metrics.SnapshotData
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -days+i)
		todo := 10 - i/5
		if todo < 0 {
			todo = 0
		}
		snapshots = append(snapshots, metrics.SnapshotData{
			Date: day,
			StatusCounts: map[string]int{
				"Todo":        todo,
				"In Progress": 5 + i/10,
				"Done":        i / 3,
			},
		})
	}

	result := metrics.CalculateCumulativeFlow(snapshots)
	JSONResponse(w, http.StatusOK, result)
}

// LeadTime はリードタイム統計を返す。
func (h *MetricsHandler) LeadTime(w http.ResponseWriter, _ *http.Request) {
	// デモデータ
	now := time.Now()
	events := []model.TaskEvent{
		{TaskID: "t1", ToStatus: "In Progress", Timestamp: now.Add(-72 * time.Hour)},
		{TaskID: "t1", ToStatus: "Done", Timestamp: now.Add(-48 * time.Hour)},
		{TaskID: "t2", ToStatus: "In Progress", Timestamp: now.Add(-96 * time.Hour)},
		{TaskID: "t2", ToStatus: "Done", Timestamp: now.Add(-24 * time.Hour)},
		{TaskID: "t3", ToStatus: "In Progress", Timestamp: now.Add(-48 * time.Hour)},
		{TaskID: "t3", ToStatus: "Done", Timestamp: now.Add(-12 * time.Hour)},
	}

	startStatuses := map[string]bool{"In Progress": true}
	endStatuses := map[string]bool{"Done": true}

	stats := metrics.CalculateLeadTime(events, startStatuses, endStatuses)
	JSONResponse(w, http.StatusOK, stats)
}
