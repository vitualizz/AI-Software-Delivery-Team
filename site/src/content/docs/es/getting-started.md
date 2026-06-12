---
title: Comenzando
description: Instalá y ejecutá tu primer pipeline ASDT en minutos.
order: 1
locale: es
---

# Comenzando

## Requisitos

Antes de usar ASDT necesitás:

- **Claude Code** u **OpenCode** — instalado y autenticado
- **Un memory provider** — requerido para persistencia entre sesiones (por defecto: [Engram](https://github.com/Gentleman-Programming/engram))
- Una terminal (bash o zsh)

> **¿Compilando desde el código fuente?** Se requiere Go 1.22+. El instalador de una línea descarga un binario precompilado — no necesitás compilador.

## Instalación

```bash
curl -fsSL https://raw.githubusercontent.com/vitualizz/asdt/main/install.sh | bash
```

Descarga el último binario precompilado para tu plataforma e instala en `~/.local/bin/`.

## Inicialización

Inicializá ASDT en tu proyecto:

```bash
asdt init
```

Crea `.asdt/config.yaml` con valores por defecto razonables.

## Tu primer pipeline

```
/asdt Agregar autenticación de usuario con email y contraseña
```

ASDT analiza la petición y recomienda una secuencia de especialistas — por ejemplo: `/asdt-pm` → `/asdt-architect` → `/asdt-developer`. Confirmá el plan y ejecutá cada comando. Cada especialista guarda su output en la base de conocimiento para que el siguiente retome donde dejó el anterior.

## Ejecutar especialistas individuales

```
/asdt-pm Agregar modo oscuro a la página de configuración
/asdt-architect Diseñar la estrategia de caché
/asdt-developer Implementar el componente de perfil de usuario
```
