package installer

// StepTier classifies a workflow step by the capability/cost weight of its
// source-default model. Every step's shipped default maps to exactly one tier,
// and a preset assigns one model per tier — so a preset is fully described by
// three model choices regardless of how many steps a specialist ships.
type StepTier int

const (
	// TierLight covers low-cost, high-throughput steps (source default: haiku).
	TierLight StepTier = iota
	// TierAnalysis covers balanced reasoning steps (source default: sonnet).
	TierAnalysis
	// TierDecision covers the heaviest judgement steps (source default: opus).
	TierDecision
)

// Source-default model IDs used to classify a step into a tier. These match the
// `model:` values shipped in the embedded workflow.yaml files.
const (
	sourceModelLight    = "haiku"
	sourceModelAnalysis = "sonnet"
	sourceModelDecision = "opus"
)

// Classify maps a step's source-default model to its tier. Anything that is not
// the balanced or decision default falls back to TierLight — the safe, cheapest
// bucket — so an unexpected or empty value never gets promoted to a costly tier.
func Classify(sourceModel string) StepTier {
	switch sourceModel {
	case sourceModelDecision:
		return TierDecision
	case sourceModelAnalysis:
		return TierAnalysis
	default:
		return TierLight
	}
}

// Preset choice indices used by the model gate. 0 (Chameleon) has no tier
// mapping — it strips every model field so each step inherits the model the
// assistant already has defined.
const (
	PresetChameleon  = 0
	PresetSprinter   = 1
	PresetCraftsman  = 2
	PresetStrategist = 3
	PresetMastermind = 4
)

// modelFable is the top-capability decision model assigned by the Mastermind
// preset. It is a bare model ID, deliberately not registered in ai_providers.go.
const modelFable = "fable"

// presetModels maps a preset choice to the model assigned at each tier. Choice 0
// (Chameleon) is absent on purpose: it removes every model field rather than
// assigning one. Craftsman (2) reproduces the source defaults exactly, so it
// classifies as zero customizations and installs workflow.yaml files verbatim.
var presetModels = map[int]map[StepTier]string{
	PresetSprinter: {
		TierLight:    sourceModelLight,
		TierAnalysis: sourceModelLight,
		TierDecision: sourceModelAnalysis,
	},
	PresetCraftsman: {
		TierLight:    sourceModelLight,
		TierAnalysis: sourceModelAnalysis,
		TierDecision: sourceModelDecision,
	},
	PresetStrategist: {
		TierLight:    sourceModelAnalysis,
		TierAnalysis: sourceModelAnalysis,
		TierDecision: sourceModelDecision,
	},
	PresetMastermind: {
		TierLight:    sourceModelAnalysis,
		TierAnalysis: sourceModelDecision,
		TierDecision: modelFable,
	},
}

// PresetModels returns the tier→model assignment for a preset choice, or nil
// when the choice has no mapping (Chameleon, or any out-of-range value).
func PresetModels(choice int) map[StepTier]string {
	return presetModels[choice]
}

// presetNames holds the canonical English preset names, used for tests and
// logging. User-facing labels come from the i18n catalog instead.
var presetNames = map[int]string{
	PresetChameleon:  "Chameleon",
	PresetSprinter:   "Sprinter",
	PresetCraftsman:  "Craftsman",
	PresetStrategist: "Strategist",
	PresetMastermind: "Mastermind",
}

// PresetName returns the canonical English name for a preset choice, or "" for
// an unknown choice.
func PresetName(choice int) string {
	return presetNames[choice]
}
