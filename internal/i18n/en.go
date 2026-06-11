package i18n

// English is the default catalog. All other catalogs must match this structure.
var English = Catalog{
	Installer: InstallerStrings{
		TitleMainMenu:         "asdt-tui",
		TitlePreflightCheck:   "Pre-flight Check",
		TitleSelectAssistants: "Select Assistants",
		TitleSelectProvider:   "Select Memory Provider",
		TitleAgentSetup:       "Agent Persona",
		TitleEmojiPref:        "Emoji Preference",
		TitleAgentWriteMode:   "Existing Configs Detected",
		TitleReview:           "Review & Confirm",
		TitleInstalling:       "Installing...",
		TitleDone:             "Installation Complete",
		TitleDashboard:        "Dashboard",
		TitleLanguageSelect:   "Language",

		StepWord:   "step",
		StepOfWord: "of",

		MenuInstall:   "Install / Update Skills",
		MenuDashboard: "Dashboard",
		MenuQuit:      "Quit",

		HintNavigate:       "navigate",
		HintSelect:         "select",
		HintQuit:           "quit",
		HintToggle:         "toggle",
		HintAllNone:        "all/none",
		HintBack:           "back",
		HintBackToMenu:     "back to menu",
		HintCycleMode:      "cycle mode",
		HintContinue:       "continue",
		HintChecking:       "checking",
		HintEnvironment:    "environment",
		HintEngramRequired: "required — see above",

		HintGroupNav:      "Nav",
		HintGroupActions:  "Actions",
		HintGroupStatus:   "Status",
		HintGroupRequired: "Required",

		BtnContinue: "[ Continue → ]",
		BtnSkip:     "[ Skip → ]",
		BtnInstall:  "[ Install → ]",

		BodyAgentSetupSubtitle:     "Configure how AI assistants behave across all tools",
		BodyEmojiPrefSubtitle:      "Should your assistants use emojis in their responses?",
		BodyAgentWriteMode:         "Choose how to handle each existing config:",
		BodyInstalling:             "Installing assistants and skills...",
		BodyLanguageSelectSubtitle: "Which language should the installer use?",

		OptionEmojiYes: "Yes — use emojis",
		OptionEmojiNo:  "No — plain text only",

		WarnExistingConfig:     "⚠  Existing global agent config detected — will be overwritten.",
		WarnExistingConfigNote: "   Proceed at your own risk.",
		InfoWriteGlobal:        "This will write to your global AI assistant config.",

		LabelAssistants: "Assistants:",
		LabelProvider:   "Provider:  ",
		LabelPersona:    "Persona:   ",
		LabelEmojis:     "Emojis:    ",
		LabelConfig:     "Config:",
		LabelSkipped:    "skipped",

		LabelModeKeep:      "Keep",
		LabelModeOverwrite: "Overwrite",
		LabelModeAppend:    "Append",

		LabelAgentConfig:      "Agent Config:",
		MsgAgentConfigWritten: "%s agent config written",
		MsgAgentConfigSkipped: "– %s: skipped (existing config kept)",
		MsgOpenCodeNote:       "Note: OpenCode reads this as a global override for all projects.",
		MsgGetStarted:         "Open a project and run /asdt to get started.",
		MsgStaleRemoved:       "%d stale file(s) cleaned up",

		SectionYourEnvironment: "Your Environment",
		SectionMemoryProvider:  "Memory Provider",
		SectionAIEnhancements:  "AI Enhancements",

		PrefEngramRequired: "⚠  Engram is required for memory persistence across AI sessions.",
		PrefEngramInstall:  "Install: https://github.com/Gentleman-Programming/engram",
		PrefEngramRestart:  "After installing, restart asdt-tui.",
	},
	Dashboard: DashboardStrings{
		LabelInstalled: "Installed",
		LabelPersona:   "Persona",
	},
	Personas: PersonaStrings{
		Sky:         "Sharp and thorough. Roasts bad practices with wit, then explains everything.",
		Toffy:       "Warm and enthusiastic. Simple answers, accidental sarcasm.",
		Atreus:      "Bold and reckless. Ships the unconventional solution first — and it works.",
		Babi:        "Your biggest fan. Warm, fast, and celebrates everything loudly.",
		LeePalacios: "Cat lover, coder, otaku. Your pair-programming companion for the whole journey.",
	},
}
