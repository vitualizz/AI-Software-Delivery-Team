package i18n

// Spanish is the Rioplatense Spanish catalog (voseo for action descriptions).
var Spanish = Catalog{
	Installer: InstallerStrings{
		TitleMainMenu:         "asdt-tui",
		TitlePreflightCheck:   "Verificación Previa",
		TitleSelectAssistants: "Seleccionar Asistentes",
		TitleSelectProvider:   "Seleccionar Proveedor de Memoria",
		TitleAgentSetup:       "Personalidad del Agente",
		TitleEmojiPref:        "Preferencia de Emojis",
		TitleAgentWriteMode:   "Configs Existentes Detectadas",
		TitleReview:           "Revisar y Confirmar",
		TitleInstalling:       "Instalando...",
		TitleDone:             "Instalación Completa",
		TitleDashboard:        "Dashboard",
		TitleLanguageSelect:   "Idioma",

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

		BodyAgentSetupSubtitle:     "Configurá cómo se comportan tus asistentes de IA en todas las herramientas",
		BodyEmojiPrefSubtitle:      "¿Querés que tus asistentes usen emojis en sus respuestas?",
		BodyAgentWriteMode:         "Elegí cómo manejar cada config existente:",
		BodyInstalling:             "Instalando asistentes y skills...",
		BodyLanguageSelectSubtitle: "¿En qué idioma querés usar el instalador?",

		OptionEmojiYes: "Sí — usar emojis",
		OptionEmojiNo:  "No — solo texto plano",

		WarnExistingConfig:     "⚠  Se detectó una config global de agente existente — será sobreescrita.",
		WarnExistingConfigNote: "   Continuá bajo tu responsabilidad.",
		InfoWriteGlobal:        "Esto va a escribir en tu config global de asistentes de IA.",

		LabelAssistants: "Asistentes:",
		LabelProvider:   "Proveedor: ",
		LabelPersona:    "Persona:   ",
		LabelEmojis:     "Emojis:    ",
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
		MsgStaleRemoved:       "se limpiaron %d archivo(s) obsoleto(s)",

		SectionYourEnvironment: "Tu Entorno",
		SectionMemoryProvider:  "Proveedor de Memoria",
		SectionAIEnhancements:  "Mejoras de IA",

		PrefEngramRequired: "⚠  Engram es requerido para la persistencia de memoria entre sesiones de IA.",
		PrefEngramInstall:  "Instalá: https://github.com/Gentleman-Programming/engram",
		PrefEngramRestart:  "Después de instalarlo, reiniciá asdt-tui.",
	},
	Dashboard: DashboardStrings{
		LabelInstalled: "Instalado",
		LabelPersona:   "Persona",
	},
	Personas: PersonaStrings{
		Sky:         "Aguda y minuciosa. Destroza las malas prácticas con ingenio, y después te explica todo.",
		Toffy:       "Cálida y entusiasta. Respuestas simples, sarcasmo accidental.",
		Atreus:      "Audaz y temerario. Shipea primero la solución poco convencional — y funciona.",
		Babi:        "Tu fan número uno. Cálida, rápida, y celebra todo a los gritos.",
		LeePalacios: "Amante de los gatos, coder, otaku. Tu compañero de pair programming para todo el viaje.",
	},
}
