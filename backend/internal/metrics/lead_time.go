package metrics

import (
	"math"
	"sort"

	"github.com/akaitigo/agile-metrics-hub/internal/model"
)

// CalculateLeadTime はタスクイベントからリードタイム統計を計算する。
// startStatuses はリードタイム計測開始ステータス（例: "To Do"）。
// endStatuses はリードタイム計測終了ステータス（例: "Done"）。
func CalculateLeadTime(events []model.TaskEvent, startStatuses, endStatuses map[string]bool) model.LeadTimeStats {
	// タスクごとに開始時刻と終了時刻を追跡
	type taskTimes struct {
		started  int64 // unix millis
		finished int64
	}
	tasks := make(map[string]*taskTimes)

	for _, e := range events {
		tt, ok := tasks[e.TaskID]
		if !ok {
			tt = &taskTimes{}
			tasks[e.TaskID] = tt
		}
		if startStatuses[e.ToStatus] && tt.started == 0 {
			tt.started = e.Timestamp.UnixMilli()
		}
		if endStatuses[e.ToStatus] && tt.started > 0 {
			tt.finished = e.Timestamp.UnixMilli()
		}
	}

	// リードタイム（時間単位）を収集
	var durations []float64
	for _, tt := range tasks {
		if tt.started > 0 && tt.finished > tt.started {
			hours := float64(tt.finished-tt.started) / (1000 * 60 * 60)
			durations = append(durations, hours)
		}
	}

	if len(durations) == 0 {
		return model.LeadTimeStats{}
	}

	sort.Float64s(durations)

	return model.LeadTimeStats{
		Percentile50: percentile(durations, 0.50),
		Percentile85: percentile(durations, 0.85),
		Percentile95: percentile(durations, 0.95),
		Average:      average(durations),
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	if lower == upper {
		return sorted[lower]
	}
	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
