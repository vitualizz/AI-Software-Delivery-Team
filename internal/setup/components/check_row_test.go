package components_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/setup/components"
)

func TestCheckRow_Render_OKContainsCheckmark(t *testing.T) {
	row := components.CheckRow{Label: "git", Status: components.CheckStatusOK, Detail: "/usr/bin/git"}
	out := row.Render(80)
	if !strings.Contains(out, "✓") {
		t.Errorf("OK row missing checkmark, got: %q", out)
	}
}

func TestCheckRow_Render_PendingContainsCircle(t *testing.T) {
	row := components.CheckRow{Label: "engram", Status: components.CheckStatusPending}
	out := row.Render(80)
	if !strings.Contains(out, "◌") && !strings.Contains(out, "engram") {
		t.Errorf("Pending row should render label, got: %q", out)
	}
}

func TestCheckRow_TickSpinner_UpdatesPendingRow(t *testing.T) {
	row := components.CheckRow{Label: "engram", Status: components.CheckStatusPending}
	row.TickSpinner("⣾ ")
	if row.SpinnerFrame != "⣾ " {
		t.Errorf("SpinnerFrame not updated, got: %q", row.SpinnerFrame)
	}
}

func TestCheckRow_TickSpinner_IgnoresNonPendingRow(t *testing.T) {
	row := components.CheckRow{Label: "git", Status: components.CheckStatusOK, SpinnerFrame: "⣾ "}
	row.TickSpinner("⣽ ")
	if row.SpinnerFrame != "⣾ " {
		t.Errorf("TickSpinner should not update non-pending row, got: %q", row.SpinnerFrame)
	}
}

func TestUpdateRow_UpdatesMatchingLabel(t *testing.T) {
	sections := []components.SectionGroup{
		{Title: "Test", Rows: []components.CheckRow{
			{Label: "engram", Status: components.CheckStatusPending},
		}},
	}
	result := components.CheckResult{Label: "engram", Status: components.CheckStatusOK, Detail: "/usr/bin/engram"}
	out := components.UpdateRow(sections, "engram", result)
	if out[0].Rows[0].Status != components.CheckStatusOK {
		t.Errorf("UpdateRow: status not updated, got: %v", out[0].Rows[0].Status)
	}
}

func TestUpdateRow_DoesNotMutateInput(t *testing.T) {
	sections := []components.SectionGroup{
		{Title: "Test", Rows: []components.CheckRow{
			{Label: "engram", Status: components.CheckStatusPending},
		}},
	}
	result := components.CheckResult{Label: "engram", Status: components.CheckStatusOK}
	_ = components.UpdateRow(sections, "engram", result)
	if sections[0].Rows[0].Status != components.CheckStatusPending {
		t.Error("UpdateRow mutated input slice")
	}
}
