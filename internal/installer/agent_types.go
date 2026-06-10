package installer

// PersonaPreset is a selectable agent persona shipped with the installer.
type PersonaPreset struct {
	ID          string // matches persona filename stem: sky|toffy|atreus|babi|lee-palacios
	Name        string // "Sky"
	Description string // one-line summary shown in StateAgentSetup
	File        string // embedded path: "asdt-init/personas/sky.md"
}

// PersonaPresets lists the built-in agent persona presets in display order.
var PersonaPresets = []PersonaPreset{
	{
		ID:          "sky",
		Name:        "Sky",
		Description: "Sharp and thorough. Roasts bad practices with wit, then explains everything.",
		File:        "asdt-init/personas/sky.md",
	},
	{
		ID:          "toffy",
		Name:        "Toffy",
		Description: "Warm and enthusiastic. Simple answers, accidental sarcasm.",
		File:        "asdt-init/personas/toffy.md",
	},
	{
		ID:          "atreus",
		Name:        "Atreus",
		Description: "Bold and reckless. Ships the unconventional solution first — and it works.",
		File:        "asdt-init/personas/atreus.md",
	},
	{
		ID:          "babi",
		Name:        "Babi",
		Description: "Your biggest fan. Warm, fast, and celebrates everything loudly.",
		File:        "asdt-init/personas/babi.md",
	},
	{
		ID:          "lee-palacios",
		Name:        "Lee Palacios",
		Description: "Cat lover, coder, otaku. Your pair-programming companion for the whole journey.",
		File:        "asdt-init/personas/lee-palacios.md",
	},
}

// AgentWriteMode controls how an existing global agent config is handled.
type AgentWriteMode int

const (
	// AgentModeOverwrite replaces any existing config entirely.
	AgentModeOverwrite AgentWriteMode = iota
	// AgentModeAppend adds the new config at the end of the existing file.
	AgentModeAppend
	// AgentModeSkip leaves the existing config untouched.
	AgentModeSkip
)

// AgentConfigResult is the per-assistant outcome of writing agent config.
type AgentConfigResult struct {
	AssistantID AssistantID
	Written     []string // files written/updated (AGENTS.md, CLAUDE.md)
	Skipped     bool     // true when a pre-existing file was kept (overwrite declined)
	Err         error
}

// AgentConfigAdapterDescriptor declares how one assistant persists agent config.
type AgentConfigAdapterDescriptor struct {
	AssistantID       AssistantID
	ConfigPath        func() string // display path shown in the TUI (e.g. "~/.claude/CLAUDE.md")
	AgentConfigExists func() bool
	Write             func(rendered string, mode AgentWriteMode) (AgentConfigResult, error)
}
