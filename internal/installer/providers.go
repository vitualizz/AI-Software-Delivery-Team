package installer

// ProviderID identifies a known memory provider.
type ProviderID = string

const ProviderEngram ProviderID = "engram"

// ProviderDescriptor describes a known memory provider.
// CustomizeSkill transforms skill file content before writing; it must be non-nil.
type ProviderDescriptor struct {
	ID             ProviderID
	Name           string
	Description    string
	CustomizeSkill func(content string) string
}

// Providers lists all known memory providers.
var Providers = []ProviderDescriptor{
	{
		ID:             ProviderEngram,
		Name:           "Engram",
		Description:    "Persistent cross-session memory via the Engram MCP server.",
		CustomizeSkill: func(content string) string { return content },
	},
}
