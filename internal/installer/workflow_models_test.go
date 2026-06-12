package installer_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

const modelWorkflowYAML = `specialist: pm
steps:
  # recall runs inline — context only
  - name: knowledge-recall
    skill: ../asdt-shared/skills/knowledge-recall.md
    execution: inline

  - name: feature-intake
    skill: steps/feature-intake.md
    execution: subagent
    agent: analyst
    inputs: []
    output_topic_key: "{project}/{change}/pm/feature-intake"

  - name: backlog-entry
    skill: steps/backlog-entry.md
    execution: subagent
    agent: analyst
    model: sonnet
    inputs:
      - "{project}/{change}/pm/feature-intake"
    output_topic_key: "{project}/{change}/pm/backlog-entry"
`

func TestWorkflowModelStepsListsOnlySubagentSteps(t *testing.T) {
	fsys := fstest.MapFS{
		"asdt-pm/workflow.yaml": {Data: []byte(modelWorkflowYAML)},
		"asdt-shared/skills/x":  {Data: []byte("not a workflow")},
	}

	steps, err := installer.WorkflowModelSteps(fsys)
	if err != nil {
		t.Fatalf("WorkflowModelSteps: %v", err)
	}

	if len(steps) != 2 {
		t.Fatalf("got %d steps, want 2 (inline excluded)", len(steps))
	}
	if steps[0].Key() != "pm/feature-intake" || steps[0].Model != "" {
		t.Errorf("steps[0] = %+v, want pm/feature-intake with empty model", steps[0])
	}
	if steps[1].Key() != "pm/backlog-entry" || steps[1].Model != "sonnet" {
		t.Errorf("steps[1] = %+v, want pm/backlog-entry with model sonnet", steps[1])
	}
}

func TestInjectModelsUpdatesAndInserts(t *testing.T) {
	models := map[string]string{
		"pm/feature-intake": "openai/gpt-4o-mini", // insert: step has no model field
		"pm/backlog-entry":  "opus",               // update: step has model: sonnet
	}

	out, err := installer.InjectModels([]byte(modelWorkflowYAML), models)
	if err != nil {
		t.Fatalf("InjectModels: %v", err)
	}
	got := string(out)

	if !strings.Contains(got, "model: openai/gpt-4o-mini") {
		t.Errorf("inserted model missing:\n%s", got)
	}
	if !strings.Contains(got, "model: opus") {
		t.Errorf("updated model missing:\n%s", got)
	}
	if strings.Contains(got, "model: sonnet") {
		t.Errorf("old model value still present:\n%s", got)
	}
	if !strings.Contains(got, "# recall runs inline — context only") {
		t.Errorf("comment lost in round-trip:\n%s", got)
	}
}

func TestInjectModelsLeavesUnselectedStepsUntouched(t *testing.T) {
	out, err := installer.InjectModels([]byte(modelWorkflowYAML), map[string]string{
		"pm/backlog-entry": "opus",
	})
	if err != nil {
		t.Fatalf("InjectModels: %v", err)
	}
	got := string(out)

	// feature-intake had no model and no selection — still none.
	intakeBlock := got[strings.Index(got, "feature-intake"):strings.Index(got, "backlog-entry")]
	if strings.Contains(intakeBlock, "model:") {
		t.Errorf("feature-intake gained a model field without a selection:\n%s", intakeBlock)
	}
}

func TestInjectModelsEmptyMapIsNoop(t *testing.T) {
	out, err := installer.InjectModels([]byte(modelWorkflowYAML), nil)
	if err != nil {
		t.Fatalf("InjectModels: %v", err)
	}
	if string(out) != modelWorkflowYAML {
		t.Error("nil models map must return content byte-identical")
	}
}

func TestInjectModelsSameValueIsByteIdenticalNoop(t *testing.T) {
	// Selections equal to the source defaults (the prefilled skip path) must
	// not re-encode the YAML — output stays byte-identical to the source.
	out, err := installer.InjectModels([]byte(modelWorkflowYAML), map[string]string{
		"pm/backlog-entry": "sonnet", // same as source default
	})
	if err != nil {
		t.Fatalf("InjectModels: %v", err)
	}
	if string(out) != modelWorkflowYAML {
		t.Error("same-value selection must leave content byte-identical")
	}
}

func TestInjectModelsIgnoresUnknownSteps(t *testing.T) {
	out, err := installer.InjectModels([]byte(modelWorkflowYAML), map[string]string{
		"pm/no-such-step":   "opus",
		"other/some-step":   "haiku",
		"pm/backlog-entry ": "opus", // trailing space — must not match
	})
	if err != nil {
		t.Fatalf("InjectModels: %v", err)
	}
	if string(out) != modelWorkflowYAML {
		t.Error("selections matching no step must leave content unchanged")
	}
}

func TestRemoveModelsStripsEveryModelField(t *testing.T) {
	out, err := installer.RemoveModels([]byte(modelWorkflowYAML))
	if err != nil {
		t.Fatalf("RemoveModels: %v", err)
	}
	got := string(out)

	if strings.Contains(got, "model:") {
		t.Errorf("RemoveModels left a model field behind:\n%s", got)
	}
	// Comments and step order must survive the round-trip.
	if !strings.Contains(got, "# recall runs inline — context only") {
		t.Errorf("comment lost in round-trip:\n%s", got)
	}
	if strings.Index(got, "feature-intake") >= strings.Index(got, "backlog-entry") {
		t.Errorf("step order changed:\n%s", got)
	}
}

func TestRemoveModelsNoModelFieldIsByteIdenticalNoop(t *testing.T) {
	// A workflow with no model: field anywhere must be returned byte-identical
	// — no re-encode, no comment/whitespace drift.
	const noModelYAML = `specialist: pm
steps:
  # only inline + a subagent without a model
  - name: knowledge-recall
    execution: inline

  - name: feature-intake
    execution: subagent
    agent: analyst
`
	out, err := installer.RemoveModels([]byte(noModelYAML))
	if err != nil {
		t.Fatalf("RemoveModels: %v", err)
	}
	if string(out) != noModelYAML {
		t.Errorf("RemoveModels on model-free content must be byte-identical, got:\n%s", string(out))
	}
}
