package summary

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/stephensun/whoop-cli/whoop"
)

var TargetSports = map[string]string{"muay-thai": "Muay Thai", "weightlifting": "lift"}

func Adherence(workouts []whoop.Workout) string {
	counts := map[string]int{}
	for _, workout := range workouts {
		if label, ok := TargetSports[strings.ToLower(workout.SportName)]; ok {
			counts[label]++
		}
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	labels := make([]string, 0, len(counts))
	for label := range counts {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	detail := "none"
	if len(labels) > 0 {
		parts := make([]string, 0, len(labels))
		for _, label := range labels {
			parts = append(parts, fmt.Sprintf("%d %s", counts[label], label))
		}
		detail = strings.Join(parts, ", ")
	}
	return fmt.Sprintf("adherence | %d/5 target sessions in last 7 days (%s); target: 2 lift + 2 Muay Thai + 1 flex", total, detail)
}

func Brief(recoveries []whoop.Recovery, sleeps []whoop.Sleep, workouts []whoop.Workout) []string {
	lines := make([]string, 0, 3)
	if len(recoveries) > 0 && recoveries[0].Score != nil {
		s := recoveries[0].Score
		lines = append(lines, fmt.Sprintf("recovery | %s | hrv %s ms | rhr %s bpm", number(s.RecoveryScore, 0), number(s.HRVRmssdMilli, 1), number(s.RestingHeartRate, 0)))
	}
	if len(sleeps) > 0 && sleeps[0].Score != nil && sleeps[0].Score.StageSummary != nil {
		s := sleeps[0].Score
		stage := s.StageSummary
		total := stage.TotalLightSleepTimeMilli + stage.TotalRemSleepTimeMilli + stage.TotalSlowWaveSleepTimeMilli
		lines = append(lines, fmt.Sprintf("sleep | %s%% performance | %s h asleep", number(s.SleepPerformancePercentage, 0), number(float64(total)/3600000, 1)))
	}
	lines = append(lines, Adherence(workouts))
	return lines
}

func Weekly(recoveries []whoop.Recovery, sleeps []whoop.Sleep, workouts []whoop.Workout) []string {
	lines := []string{}
	var scores, hrv, rhr []float64
	for _, recovery := range recoveries {
		if recovery.Score == nil {
			continue
		}
		if recovery.Score.RecoveryScore != 0 {
			scores = append(scores, recovery.Score.RecoveryScore)
		}
		if recovery.Score.HRVRmssdMilli != 0 {
			hrv = append(hrv, recovery.Score.HRVRmssdMilli)
		}
		if recovery.Score.RestingHeartRate != 0 {
			rhr = append(rhr, recovery.Score.RestingHeartRate)
		}
	}
	if len(scores) > 0 {
		lines = append(lines, fmt.Sprintf("recovery | avg %s%% | min %s%% | %d days", number(avg(scores), 0), number(min(scores), 0), len(scores)))
	}
	if len(hrv) > 0 && len(rhr) > 0 {
		lines = append(lines, fmt.Sprintf("hrv | avg %sms | rhr avg %sbpm", number(avg(hrv), 1), number(avg(rhr), 0)))
	}
	var hours []float64
	for _, sleep := range sleeps {
		if sleep.Nap || sleep.Score == nil || sleep.Score.StageSummary == nil {
			continue
		}
		stage := sleep.Score.StageSummary
		total := stage.TotalLightSleepTimeMilli + stage.TotalRemSleepTimeMilli + stage.TotalSlowWaveSleepTimeMilli
		if total > 0 {
			hours = append(hours, float64(total)/3600000)
		}
	}
	if len(hours) > 0 {
		lines = append(lines, fmt.Sprintf("sleep | avg %sh | min %sh | %d nights", number(avg(hours), 1), number(min(hours), 1), len(hours)))
	}
	lines = append(lines, Adherence(workouts))
	return lines
}

func number(value float64, digits int) string {
	if digits == 0 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}
func avg(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}
func min(values []float64) float64 {
	result := math.Inf(1)
	for _, value := range values {
		if value < result {
			result = value
		}
	}
	return result
}
