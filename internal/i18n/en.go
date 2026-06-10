package i18n

// English is the default catalog. All other catalogs must match this structure.
var English = Catalog{
	Installer: InstallerStrings{
		TitleMainMenu:         "asdt-tui",
		TitlePreflightCheck:   "Pre-flight Check",
		TitleSelectAssistants: "Select Assistants",
		TitleSelectProvider:   "Select Memory Provider",
		TitleAgentSetup:       "Agent Persona",
		TitleAgentWriteMode:   "Existing Configs Detected",
		TitleReview:           "Review & Confirm",
		TitleInstalling:       "Installing...",
		TitleDone:             "Installation Complete",
		TitleDashboard:        "Dashboard",

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

		BodyAgentSetupSubtitle: "Configure how AI assistants behave across all tools",
		BodyAgentWriteMode:     "Choose how to handle each existing config:",
		BodyInstalling:         "Installing assistants and skills...",

		WarnExistingConfig:     "⚠  Existing global agent config detected — will be overwritten.",
		WarnExistingConfigNote: "   Proceed at your own risk.",
		InfoWriteGlobal:        "This will write to your global AI assistant config.",

		LabelAssistants: "Assistants:",
		LabelProvider:   "Provider:  ",
		LabelPersona:    "Persona:   ",
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
}
