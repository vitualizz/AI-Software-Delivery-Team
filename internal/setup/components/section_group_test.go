package components_test

import (
	"strings"
	"testing"

	"github.com/vitualizz/asdt/internal/setup/components"
)

func TestSectionGroup_Render_ContainsTitle(t *testing.T) {
	sg := components.SectionGroup{
		Title: "Memory Provider",
		Rows:  []components.CheckRow{{Label: "Engram", Status: components.CheckStatusPending}},
	}
	out := sg.Render(80)
	if !strings.Contains(out, "Memory Provider") {
		t.Errorf("SectionGroup render missing title, got: %q", out)
	}
}

func TestSectionGroup_Render_ContainsRowLabel(t *testing.T) {
	sg := components.SectionGroup{
		Title: "Memory Provider",
		Rows:  []components.CheckRow{{Label: "Engram", Status: components.CheckStatusOK}},
	}
	out := sg.Render(80)
	if !strings.Contains(out, "Engram") {
		t.Errorf("SectionGroup render missing row label, got: %q", out)
	}
}
