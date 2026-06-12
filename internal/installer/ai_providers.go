package installer

import "os"

// AIModel describes one selectable model offered by an AI provider.
//
// ID is the exact value written into a workflow.yaml step's `model:` field.
// For Anthropic models it is the Claude Code Agent-tool enum (haiku, sonnet,
// opus) so installed files work natively on Claude Code; for every other
// provider it is a "provider/model" string that OpenCode consumes directly
// and Claude Code maps to ClaudeCodeEnum at invocation time.
type AIModel struct {
	ID             string
	Display        string
	ClaudeCodeEnum string
	OpenCodeID     string
}

// AIProvider describes a known AI provider and the models it offers.
// EnvKey is the environment variable whose presence marks the provider as
// available — detection is passive: no network calls, no key validation.
type AIProvider struct {
	ID     string
	Name   string
	EnvKey string
	Models []AIModel
}

// AIProviderAnthropic is the provider always offered: Claude Code runs
// Anthropic models natively, so its catalog never depends on an env var.
const AIProviderAnthropic = "anthropic"

// AIProviders lists all known AI providers with their cataloged models.
// The catalog ships compiled into the binary and is updated with releases.
var AIProviders = []AIProvider{
	{
		ID:     AIProviderAnthropic,
		Name:   "Anthropic",
		EnvKey: "ANTHROPIC_API_KEY",
		Models: []AIModel{
			{ID: "haiku", Display: "Claude Haiku — fast, low cost", ClaudeCodeEnum: "haiku", OpenCodeID: "anthropic/claude-haiku-4-5"},
			{ID: "sonnet", Display: "Claude Sonnet — balanced", ClaudeCodeEnum: "sonnet", OpenCodeID: "anthropic/claude-sonnet-4-6"},
			{ID: "opus", Display: "Claude Opus — most capable", ClaudeCodeEnum: "opus", OpenCodeID: "anthropic/claude-opus-4-8"},
		},
	},
	{
		ID:     "openai",
		Name:   "OpenAI",
		EnvKey: "OPENAI_API_KEY",
		Models: []AIModel{
			{ID: "openai/gpt-4o-mini", Display: "GPT-4o Mini — fast, low cost", ClaudeCodeEnum: "haiku", OpenCodeID: "openai/gpt-4o-mini"},
			{ID: "openai/gpt-4o", Display: "GPT-4o — balanced", ClaudeCodeEnum: "sonnet", OpenCodeID: "openai/gpt-4o"},
			{ID: "openai/o3", Display: "o3 — deep reasoning", ClaudeCodeEnum: "opus", OpenCodeID: "openai/o3"},
		},
	},
	{
		ID:     "google",
		Name:   "Google Gemini",
		EnvKey: "GOOGLE_API_KEY",
		Models: []AIModel{
			{ID: "google/gemini-2.0-flash", Display: "Gemini 2.0 Flash — fast, low cost", ClaudeCodeEnum: "haiku", OpenCodeID: "google/gemini-2.0-flash"},
			{ID: "google/gemini-1.5-pro", Display: "Gemini 1.5 Pro — long context", ClaudeCodeEnum: "sonnet", OpenCodeID: "google/gemini-1.5-pro"},
			{ID: "google/gemini-2.0-pro", Display: "Gemini 2.0 Pro — most capable", ClaudeCodeEnum: "opus", OpenCodeID: "google/gemini-2.0-pro"},
		},
	},
}

// DetectAIProviders returns the providers available in this environment:
// Anthropic always, every other provider only when its EnvKey is set.
func DetectAIProviders() []AIProvider {
	var detected []AIProvider
	for _, p := range AIProviders {
		if p.ID == AIProviderAnthropic || os.Getenv(p.EnvKey) != "" {
			detected = append(detected, p)
		}
	}
	return detected
}

// FlattenModels returns the models of the given providers as one list,
// preserving provider order — the cycle order shown in the TUI picker.
func FlattenModels(providers []AIProvider) []AIModel {
	var models []AIModel
	for _, p := range providers {
		models = append(models, p.Models...)
	}
	return models
}

// FindModel returns the cataloged model with the given ID and true, or a
// zero AIModel and false when the ID is not in the catalog.
func FindModel(id string) (AIModel, bool) {
	for _, p := range AIProviders {
		for _, m := range p.Models {
			if m.ID == id {
				return m, true
			}
		}
	}
	return AIModel{}, false
}
