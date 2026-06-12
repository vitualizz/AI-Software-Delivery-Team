package installer_test

import (
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/installer"
)

func TestClassifyMapsSourceDefaultsToTiers(t *testing.T) {
	cases := []struct {
		source string
		want   installer.StepTier
	}{
		{"haiku", installer.TierLight},
		{"sonnet", installer.TierAnalysis},
		{"opus", installer.TierDecision},
		{"", installer.TierLight},        // empty default → cheapest bucket
		{"unknown", installer.TierLight}, // unexpected value → cheapest bucket
	}
	for _, c := range cases {
		if got := installer.Classify(c.source); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.source, got, c.want)
		}
	}
}

func TestPresetModelsChameleonHasNoMapping(t *testing.T) {
	if installer.PresetModels(installer.PresetChameleon) != nil {
		t.Error("Chameleon must have no tier mapping (nil)")
	}
}

func TestPresetModelsSprinterAssignment(t *testing.T) {
	tiers := installer.PresetModels(installer.PresetSprinter)
	if tiers == nil {
		t.Fatal("Sprinter must have a tier mapping")
	}
	if tiers[installer.TierLight] != "haiku" || tiers[installer.TierAnalysis] != "haiku" {
		t.Errorf("Sprinter light/analysis = %q/%q, want haiku/haiku", tiers[installer.TierLight], tiers[installer.TierAnalysis])
	}
	if tiers[installer.TierDecision] != "sonnet" {
		t.Errorf("Sprinter decision = %q, want sonnet", tiers[installer.TierDecision])
	}
}

func TestPresetModelsCraftsmanReproducesSourceDefaults(t *testing.T) {
	tiers := installer.PresetModels(installer.PresetCraftsman)
	if tiers[installer.TierLight] != "haiku" ||
		tiers[installer.TierAnalysis] != "sonnet" ||
		tiers[installer.TierDecision] != "opus" {
		t.Errorf("Craftsman must reproduce source defaults (haiku/sonnet/opus), got %v", tiers)
	}
}

func TestPresetModelsMastermindUsesFableForDecisions(t *testing.T) {
	tiers := installer.PresetModels(installer.PresetMastermind)
	if tiers[installer.TierDecision] != "fable" {
		t.Errorf("Mastermind decision = %q, want fable", tiers[installer.TierDecision])
	}
}

// TestPresetDecisionTierNeverLight is the NFR invariant: no preset may route a
// decision-tier step to the cheapest model — decisions always get at least the
// balanced tier.
func TestPresetDecisionTierNeverLight(t *testing.T) {
	for _, choice := range []int{
		installer.PresetSprinter,
		installer.PresetCraftsman,
		installer.PresetStrategist,
		installer.PresetMastermind,
	} {
		tiers := installer.PresetModels(choice)
		if tiers[installer.TierDecision] == "haiku" {
			t.Errorf("preset %d routes decisions to haiku — violates NFR invariant", choice)
		}
	}
}

func TestPresetNameReturnsCanonicalNames(t *testing.T) {
	cases := map[int]string{
		installer.PresetChameleon:  "Chameleon",
		installer.PresetSprinter:   "Sprinter",
		installer.PresetCraftsman:  "Craftsman",
		installer.PresetStrategist: "Strategist",
		installer.PresetMastermind: "Mastermind",
	}
	for choice, want := range cases {
		if got := installer.PresetName(choice); got != want {
			t.Errorf("PresetName(%d) = %q, want %q", choice, got, want)
		}
	}
	if installer.PresetName(99) != "" {
		t.Error("PresetName for an unknown choice must be empty")
	}
}
