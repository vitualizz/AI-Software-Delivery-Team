// Package i18n provides a simple locale-based string catalog for the TUI.
// Add new feature areas as nested structs inside Catalog when they appear.
package i18n

// Catalog holds all translatable strings, organized by feature area so it
// scales naturally as new TUI surfaces (dashboard, settings) are added.
type Catalog struct {
	Installer InstallerStrings
	Dashboard DashboardStrings
	Personas  PersonaStrings
}

// PersonaStrings holds the localized one-line descriptions for the built-in
// persona presets shown in StateAgentSetup and StateReview. Persona names are
// proper nouns and never translated — only their descriptions are.
type PersonaStrings struct {
	Sky         string
	Toffy       string
	Atreus      string
	Babi        string
	LeePalacios string
}

// PersonaDescription returns the localized description for the given persona
// preset ID. An empty field falls back to the English catalog's same field;
// an unknown ID returns "".
func (c Catalog) PersonaDescription(presetID string) string {
	pick := func(p PersonaStrings) string {
		switch presetID {
		case "sky":
			return p.Sky
		case "toffy":
			return p.Toffy
		case "atreus":
			return p.Atreus
		case "babi":
			return p.Babi
		case "lee-palacios":
			return p.LeePalacios
		}
		return ""
	}
	if s := pick(c.Personas); s != "" {
		return s
	}
	return pick(English.Personas)
}

// InstallerStrings holds every user-visible string in the wizard TUI.
type InstallerStrings struct {
	// Frame titles
	TitleMainMenu         string
	TitlePreflightCheck   string
	TitleSelectAssistants string
	TitleSelectProvider   string
	TitleAgentSetup       string
	TitleEmojiPref        string
	TitleAgentWriteMode   string
	TitleReview           string
	TitleInstalling       string
	TitleDone             string
	TitleDashboard        string
	TitleLanguageSelect   string

	// Step indicator words ("step 1 of 5" / "paso 1 de 5")
	StepWord   string
	StepOfWord string

	// Main menu items
	MenuInstall   string
	MenuDashboard string
	MenuQuit      string

	// Keyboard hint descriptions (what a key does)
	HintNavigate       string
	HintSelect         string
	HintQuit           string
	HintToggle         string
	HintAllNone        string
	HintBack           string
	HintBackToMenu     string
	HintCycleMode      string
	HintContinue       string
	HintChecking       string
	HintEnvironment    string
	HintEngramRequired string

	// Hint group labels
	HintGroupNav      string
	HintGroupActions  string
	HintGroupStatus   string
	HintGroupRequired string

	// Inline confirm buttons (full label including brackets)
	BtnContinue string
	BtnSkip     string
	BtnInstall  string

	// Body / subtitle text
	BodyAgentSetupSubtitle     string
	BodyEmojiPrefSubtitle      string
	BodyAgentWriteMode         string
	BodyInstalling             string
	BodyLanguageSelectSubtitle string

	// Emoji preference options (radio rows + Review value)
	OptionEmojiYes string
	OptionEmojiNo  string

	// AgentSetup state messages
	WarnExistingConfig     string
	WarnExistingConfigNote string
	InfoWriteGlobal        string

	// Review screen labels
	LabelAssistants string
	LabelProvider   string
	LabelPersona    string
	LabelEmojis     string
	LabelConfig     string
	LabelSkipped    string

	// Agent write mode labels (Keep / Overwrite / Append)
	LabelModeKeep      string
	LabelModeOverwrite string
	LabelModeAppend    string

	// Done screen
	LabelAgentConfig      string
	MsgAgentConfigWritten string // printf format: first %s = assistant name
	MsgAgentConfigSkipped string // printf format: first %s = assistant name
	MsgOpenCodeNote       string
	MsgGetStarted         string

	// Preflight section group titles (display only; row labels are lookup keys)
	SectionYourEnvironment string
	SectionMemoryProvider  string
	SectionAIEnhancements  string

	// Preflight Engram recovery block (shown when engram is not found)
	PrefEngramRequired string
	PrefEngramInstall  string
	PrefEngramRestart  string
}

// DashboardStrings holds user-visible strings for the dashboard TUI.
type DashboardStrings struct {
	LabelInstalled string // "Installed" — prefix for the install date
	LabelPersona   string // "Persona" — prefix for the configured persona name
}
