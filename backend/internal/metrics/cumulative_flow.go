package metrics

import "github.com/akaitigo/agile-metrics-hub/internal/model"

// CalculateCumulativeFlow はスナップショットデータから累積フローダイアグラムデータを生成する。
func CalculateCumulativeFlow(snapshots []SnapshotData) []model.CumulativeFlowPoint {
	points := make([]model.CumulativeFlowPoint, 0, len(snapshots))
	for _, s := range snapshots {
		// マップをコピーして呼び出し元との共有を防ぐ
		statuses := make(map[string]int, len(s.StatusCounts))
		for k, v := range s.StatusCounts {
			statuses[k] = v
		}
		points = append(points, model.CumulativeFlowPoint{
			Date:     s.Date,
			Statuses: statuses,
		})
	}
	return points
}
