package metrics

import (
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

// CalculateBurndown はスナップショットデータからバーンダウンチャートデータを生成する。
// sprintStart/sprintEnd はスプリント期間。totalPoints は初期ポイント。
func CalculateBurndown(snapshots []SnapshotData, sprintStart, sprintEnd time.Time, totalPoints float64) []model.BurndownPoint {
	totalDays := sprintEnd.Sub(sprintStart).Hours() / 24
	if totalDays <= 0 {
		return nil
	}

	var points []model.BurndownPoint
	for _, s := range snapshots {
		if s.Date.Before(sprintStart) || s.Date.After(sprintEnd) {
			continue
		}
		daysSinceStart := s.Date.Sub(sprintStart).Hours() / 24
		idealRemaining := totalPoints * (1 - daysSinceStart/totalDays)
		if idealRemaining < 0 {
			idealRemaining = 0
		}

		points = append(points, model.BurndownPoint{
			Date:            s.Date,
			RemainingTasks:  s.TotalTasks - s.CompletedTasks,
			RemainingPoints: s.TotalPoints - s.CompletedPoints,
			IdealRemaining:  idealRemaining,
		})
	}
	return points
}

// SnapshotData はスナップショットから必要なデータを抽出したもの。
type SnapshotData struct {
	Date            time.Time
	TotalTasks      int
	CompletedTasks  int
	TotalPoints     float64
	CompletedPoints float64
	StatusCounts    map[string]int
}
