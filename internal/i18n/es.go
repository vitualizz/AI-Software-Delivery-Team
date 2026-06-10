package i18n

// Spanish is the Rioplatense Spanish catalog (voseo for action descriptions).
var Spanish = Catalog{
	Installer: InstallerStrings{
		TitleMainMenu:         "asdt-tui",
		TitlePreflightCheck:   "Verificación Previa",
		TitleSelectAssistants: "Seleccionar Asistentes",
		TitleSelectProvider:   "Seleccionar Proveedor de Memoria",
		TitleAgentSetup:       "Personalidad del Agente",
		TitleAgentWriteMode:   "Configs Existentes Detectadas",
		TitleReview:           "Revisar y Confirmar",
		TitleInstalling:       "Instalando...",
		TitleDone:             "Instalación Completa",
		TitleDashboard:        "Dashboard",

		StepWord:   "paso",
		StepOfWord: "de",

		MenuInstall:   "Instalar / Actualizar Skills",
		MenuDashboard: "Dashboard",
		MenuQuit:      "Salir",

		HintNavigate:       "navegar",
		HintSelect:         "seleccionar",
		HintQuit:           "salir",
		HintToggle:         "alternar",
		HintAllNone:        "todos/ninguno",
		HintBack:           "volver",
		HintBackToMenu:     "volver al menú",
		HintCycleMode:      "cambiar modo",
		HintContinue:       "continuar",
		HintChecking:       "verificando",
		HintEnvironment:    "entorno",
		HintEngramRequired: "requerido — ver arriba",

		HintGroupNav:      "Nav",
		HintGroupActions:  "Acciones",
		HintGroupStatus:   "Estado",
		HintGroupRequired: "Requerido",

		BtnContinue: "[ Continuar → ]",
		BtnSkip:     "[ Omitir → ]",
		BtnInstall:  "[ Instalar → ]",

		BodyAgentSetupSubtitle: "Configurá cómo se comportan tus asistentes de IA en todas las herramientas",
		BodyAgentWriteMode:     "Elegí cómo manejar cada config existente:",
		BodyInstalling:         "Instalando asistentes y skills...",

		WarnExistingConfig:     "⚠  Se detectó una config global de agente existente — será sobreescrita.",
		WarnExistingConfigNote: "   Continuá bajo tu responsabilidad.",
		InfoWriteGlobal:        "Esto va a escribir en tu config global de asistentes de IA.",

		LabelAssistants: "Asistentes:",
		LabelProvider:   "Proveedor: ",
		LabelPersona:    "Persona:   ",
		LabelConfig:     "Config:",
		LabelSkipped:    "omitido",

		LabelModeKeep:      "Mantener",
		LabelModeOverwrite: "Sobreescribir",
		LabelModeAppend:    "Agregar",

		LabelAgentConfig:      "Config del Agente:",
		MsgAgentConfigWritten: "config de %s escrita",
		MsgAgentConfigSkipped: "– %s: omitido (se mantuvo la config existente)",
		MsgOpenCodeNote:       "Nota: OpenCode usa esto como override global para todos los proyectos.",
		MsgGetStarted:         "Abrí un proyecto y ejecutá /asdt para arrancar.",

		SectionYourEnvironment: "Tu Entorno",
		SectionMemoryProvider:  "Proveedor de Memoria",
		SectionAIEnhancements:  "Mejoras de IA",

		PrefEngramRequired: "⚠  Engram es requerido para la persistencia de memoria entre sesiones de IA.",
		PrefEngramInstall:  "Instalá: https://github.com/Gentleman-Programming/engram",
		PrefEngramRestart:  "Después de instalarlo, reiniciá asdt-tui.",
	},
}
