package installer_test

import (
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

func TestDetectAIProvidersAlwaysIncludesAnthropic(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")

	detected := installer.DetectAIProviders()

	if len(detected) != 1 {
		t.Fatalf("DetectAIProviders with no env keys: got %d providers, want 1 (anthropic only)", len(detected))
	}
	if detected[0].ID != installer.AIProviderAnthropic {
		t.Errorf("DetectAIProviders[0].ID = %q, want %q", detected[0].ID, installer.AIProviderAnthropic)
	}
}

func TestDetectAIProvidersIncludesProviderWhenEnvKeySet(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("GOOGLE_API_KEY", "")

	detected := installer.DetectAIProviders()

	if len(detected) != 2 {
		t.Fatalf("got %d providers, want 2 (anthropic + openai)", len(detected))
	}
	if detected[1].ID != "openai" {
		t.Errorf("detected[1].ID = %q, want openai", detected[1].ID)
	}
}

func TestFlattenModelsPreservesProviderOrder(t *testing.T) {
	providers := []installer.AIProvider{
		{ID: "a", Models: []installer.AIModel{{ID: "a1"}, {ID: "a2"}}},
		{ID: "b", Models: []installer.AIModel{{ID: "b1"}}},
	}

	models := installer.FlattenModels(providers)

	want := []string{"a1", "a2", "b1"}
	if len(models) != len(want) {
		t.Fatalf("got %d models, want %d", len(models), len(want))
	}
	for i, id := range want {
		if models[i].ID != id {
			t.Errorf("models[%d].ID = %q, want %q", i, models[i].ID, id)
		}
	}
}

func TestFindModel(t *testing.T) {
	m, ok := installer.FindModel("sonnet")
	if !ok {
		t.Fatal("FindModel(sonnet): not found, want found")
	}
	if m.ClaudeCodeEnum != "sonnet" {
		t.Errorf("FindModel(sonnet).ClaudeCodeEnum = %q, want sonnet", m.ClaudeCodeEnum)
	}

	if _, ok := installer.FindModel("no-such-model"); ok {
		t.Error("FindModel(no-such-model): found, want not found")
	}
}

func TestAnthropicModelIDsAreClaudeCodeEnums(t *testing.T) {
	// Anthropic model IDs must equal their ClaudeCodeEnum so installed
	// workflow.yaml files work natively on Claude Code without mapping.
	for _, p := range installer.AIProviders {
		if p.ID != installer.AIProviderAnthropic {
			continue
		}
		for _, m := range p.Models {
			if m.ID != m.ClaudeCodeEnum {
				t.Errorf("anthropic model %q: ID != ClaudeCodeEnum (%q)", m.ID, m.ClaudeCodeEnum)
			}
		}
	}
}
