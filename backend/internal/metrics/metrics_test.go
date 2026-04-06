package metrics_test

import (
	"math"
	"testing"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/metrics"
	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

func TestCalculateBurndown(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC) // 10日間

	snapshots := []metrics.SnapshotData{
		{Date: start, TotalTasks: 10, CompletedTasks: 0, TotalPoints: 50, CompletedPoints: 0},
		{Date: start.AddDate(0, 0, 5), TotalTasks: 10, CompletedTasks: 6, TotalPoints: 50, CompletedPoints: 30},
		{Date: end, TotalTasks: 10, CompletedTasks: 10, TotalPoints: 50, CompletedPoints: 50},
	}

	result := metrics.CalculateBurndown(snapshots, start, end, 50)
	if len(result) != 3 {
		t.Fatalf("expected 3 points, got %d", len(result))
	}

	// Day 0: ideal = 50, remaining = 50
	if result[0].IdealRemaining != 50 {
		t.Errorf("day 0 ideal: expected 50, got %f", result[0].IdealRemaining)
	}
	if result[0].RemainingPoints != 50 {
		t.Errorf("day 0 remaining: expected 50, got %f", result[0].RemainingPoints)
	}

	// Day 5: ideal = 25
	if math.Abs(result[1].IdealRemaining-25) > 0.01 {
		t.Errorf("day 5 ideal: expected 25, got %f", result[1].IdealRemaining)
	}

	// Day 10: ideal = 0, remaining = 0
	if result[2].RemainingPoints != 0 {
		t.Errorf("day 10 remaining: expected 0, got %f", result[2].RemainingPoints)
	}
}

func TestCalculateVelocity(t *testing.T) {
	sprints := []metrics.SprintData{
		{Name: "Sprint 1", CommittedPoints: 20, CompletedPoints: 18},
		{Name: "Sprint 2", CommittedPoints: 22, CompletedPoints: 22},
		{Name: "Sprint 3", CommittedPoints: 25, CompletedPoints: 20},
	}

	result := metrics.CalculateVelocity(sprints)
	if len(result) != 3 {
		t.Fatalf("expected 3 points, got %d", len(result))
	}
	if result[0].SprintName != "Sprint 1" {
		t.Errorf("expected 'Sprint 1', got %q", result[0].SprintName)
	}
}

func TestAverageVelocity(t *testing.T) {
	sprints := []metrics.SprintData{
		{CompletedPoints: 10},
		{CompletedPoints: 20},
		{CompletedPoints: 30},
	}

	avg := metrics.AverageVelocity(sprints, 2)
	if avg != 25 { // (20+30)/2
		t.Errorf("expected 25, got %f", avg)
	}

	avgAll := metrics.AverageVelocity(sprints, 10) // n > len
	if avgAll != 20 {                              // (10+20+30)/3
		t.Errorf("expected 20, got %f", avgAll)
	}
}

func TestCalculateLeadTime(t *testing.T) {
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	events := []model.TaskEvent{
		{TaskID: "t1", ToStatus: "In Progress", Timestamp: base},
		{TaskID: "t1", FromStatus: "In Progress", ToStatus: "Done", Timestamp: base.Add(24 * time.Hour)},
		{TaskID: "t2", ToStatus: "In Progress", Timestamp: base},
		{TaskID: "t2", FromStatus: "In Progress", ToStatus: "Done", Timestamp: base.Add(48 * time.Hour)},
	}

	startStatuses := map[string]bool{"In Progress": true}
	endStatuses := map[string]bool{"Done": true}

	stats := metrics.CalculateLeadTime(events, startStatuses, endStatuses)
	if stats.Average != 36 { // (24+48)/2
		t.Errorf("expected average 36h, got %f", stats.Average)
	}
	if stats.Percentile50 != 36 { // median of [24, 48]
		t.Errorf("expected p50 36h, got %f", stats.Percentile50)
	}
}

func TestCalculateLeadTime_NoEvents(t *testing.T) {
	stats := metrics.CalculateLeadTime(nil, nil, nil)
	if stats.Average != 0 {
		t.Errorf("expected 0 for empty events, got %f", stats.Average)
	}
}

func TestCalculateCumulativeFlow(t *testing.T) {
	snapshots := []metrics.SnapshotData{
		{
			Date:         time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			StatusCounts: map[string]int{"Todo": 5, "In Progress": 3, "Done": 2},
		},
	}
	result := metrics.CalculateCumulativeFlow(snapshots)
	if len(result) != 1 {
		t.Fatalf("expected 1 point, got %d", len(result))
	}
	if result[0].Statuses["Todo"] != 5 {
		t.Errorf("expected Todo=5, got %d", result[0].Statuses["Todo"])
	}
}
