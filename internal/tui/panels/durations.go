package panels

import (
	"fmt"
	"time"

	"github.com/vitualizz/asdt/internal/pipeline"
)

// StepDuration is a pure, presentation-only value pairing two consecutive
// StepRecord entries with the elapsed time between them.
type StepDuration struct {
	From    pipeline.StepRecord
	To      pipeline.StepRecord
	Elapsed time.Duration
}

// ComputeStepDurations walks consecutive pairs in steps (already in
// chronological append-order per the append-only steps_completed log
// convention) and returns the pairwise elapsed durations. Returns an empty
// slice for 0 or 1 steps — there are no pairs to compute.
func ComputeStepDurations(steps []pipeline.StepRecord) []StepDuration {
	if len(steps) < 2 {
		return []StepDuration{}
	}

	durations := make([]StepDuration, 0, len(steps)-1)
	for i := 1; i < len(steps); i++ {
		from := steps[i-1]
		to := steps[i]
		durations = append(durations, StepDuration{
			From:    from,
			To:      to,
			Elapsed: to.Timestamp.Sub(from.Timestamp),
		})
	}
	return durations
}

// FormatDuration renders d as a compact human string: "45s", "2m15s", or
// "1h05m" — minutes/seconds are zero-padded to two digits when a larger unit
// is present, matching common terminal-UI duration idioms.
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
