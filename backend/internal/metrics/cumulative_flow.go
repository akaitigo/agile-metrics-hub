package metrics

import "github.com/akaitigo/agile-metrics-hub/internal/model"

// CalculateCumulativeFlow はスナップショットデータから累積フローダイアグラムデータを生成する。
func CalculateCumulativeFlow(snapshots []SnapshotData) []model.CumulativeFlowPoint {
	points := make([]model.CumulativeFlowPoint, 0, len(snapshots))
	for _, s := range snapshots {
		points = append(points, model.CumulativeFlowPoint{
			Date:     s.Date,
			Statuses: s.StatusCounts,
		})
	}
	return points
}
