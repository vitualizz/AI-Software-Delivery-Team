package panels_test

import (
	"testing"
	"time"

	"github.com/vitualizz/asdt/internal/pipeline"
	"github.com/vitualizz/asdt/internal/tui/panels"
)

// TestComputeStepDurationsEmptyForZeroSteps verifies that ComputeStepDurations
// returns an empty slice when there are no steps to pair.
func TestComputeStepDurationsEmptyForZeroSteps(t *testing.T) {
	got := panels.ComputeStepDurations(nil)
	if len(got) != 0 {
		t.Errorf("expected empty slice for 0 steps, got %d entries", len(got))
	}
}

// TestComputeStepDurationsEmptyForOneStep verifies that ComputeStepDurations
// returns an empty slice when there is only one step (no pairs to compute).
func TestComputeStepDurationsEmptyForOneStep(t *testing.T) {
	steps := []pipeline.StepRecord{{ID: "explore", Timestamp: time.Now()}}
	got := panels.ComputeStepDurations(steps)
	if len(got) != 0 {
		t.Errorf("expected empty slice for 1 step (no pairs), got %d entries", len(got))
	}
}

// TestComputeStepDurationsPairwiseDeltas verifies that ComputeStepDurations
// walks consecutive pairs and computes the elapsed time between them.
func TestComputeStepDurationsPairwiseDeltas(t *testing.T) {
	base := time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC)
	steps := []pipeline.StepRecord{
		{ID: "explore", Timestamp: base},
		{ID: "spec", Timestamp: base.Add(2 * time.Minute)},
		{ID: "design", Timestamp: base.Add(5 * time.Minute)},
	}

	got := panels.ComputeStepDurations(steps)
	if len(got) != 2 {
		t.Fatalf("expected 2 pairwise durations for 3 steps, got %d", len(got))
	}

	if got[0].From.ID != "explore" || got[0].To.ID != "spec" {
		t.Errorf("pair 0: expected explore->spec, got %s->%s", got[0].From.ID, got[0].To.ID)
	}
	if got[0].Elapsed != 2*time.Minute {
		t.Errorf("pair 0: expected 2m elapsed, got %v", got[0].Elapsed)
	}

	if got[1].From.ID != "spec" || got[1].To.ID != "design" {
		t.Errorf("pair 1: expected spec->design, got %s->%s", got[1].From.ID, got[1].To.ID)
	}
	if got[1].Elapsed != 3*time.Minute {
		t.Errorf("pair 1: expected 3m elapsed, got %v", got[1].Elapsed)
	}
}

// TestFormatDurationVariants verifies FormatDuration produces compact strings
// for seconds-only, minutes+seconds, and hours+minutes magnitudes.
func TestFormatDurationVariants(t *testing.T) {
	cases := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes and seconds", 2*time.Minute + 15*time.Second, "2m15s"},
		{"hours and minutes", 1*time.Hour + 5*time.Minute, "1h05m"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := panels.FormatDuration(tc.in)
			if got != tc.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
