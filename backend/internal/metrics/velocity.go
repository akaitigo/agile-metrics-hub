package metrics

import "github.com/akaitigo/agile-metrics-hub/internal/model"

// SprintData はベロシティ計算に必要なスプリントデータ。
type SprintData struct {
	Name            string
	CommittedPoints float64
	CompletedPoints float64
}

// CalculateVelocity はスプリントデータからベロシティチャートデータを生成する。
func CalculateVelocity(sprints []SprintData) []model.VelocityPoint {
	points := make([]model.VelocityPoint, 0, len(sprints))
	for _, s := range sprints {
		points = append(points, model.VelocityPoint{
			SprintName:      s.Name,
			CommittedPoints: s.CommittedPoints,
			CompletedPoints: s.CompletedPoints,
		})
	}
	return points
}

// AverageVelocity は直近Nスプリントの平均ベロシティを計算する。
func AverageVelocity(sprints []SprintData, n int) float64 {
	if len(sprints) == 0 {
		return 0
	}
	if n > len(sprints) {
		n = len(sprints)
	}
	recent := sprints[len(sprints)-n:]
	var sum float64
	for _, s := range recent {
		sum += s.CompletedPoints
	}
	return sum / float64(n)
}
