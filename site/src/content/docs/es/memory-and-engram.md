---
title: Base de conocimiento y memoria
description: Cómo ASDT persiste artefactos entre especialistas y sesiones — y cómo la base de conocimiento mantiene al equipo sincronizado.
order: 7
locale: es
---

# Base de conocimiento y memoria

## Por qué importa la memoria

Sin memoria persistente, cada especialista empieza de cero. El Developer no puede leer el registro de decisión del Arquitecto. El especialista de QA no sabe qué construyó el Developer. Tendrías que copiar el contexto manualmente entre cada paso — lo que cancela el propósito de tener un equipo.

ASDT resuelve esto con una base de conocimiento. Cada artefacto que produce un especialista se guarda con una clave estable. El siguiente especialista lo recupera automáticamente por clave. El contexto fluye hacia adelante sin intervención manual, incluso entre sesiones separadas por días.

## Cómo fluye el conocimiento

Cada paso de cada especialista produce un artefacto. Ese artefacto se guarda en la base de conocimiento con dos identificadores:

- Un **título legible por humanos** — ej. `"add-auth/developer/dev-implementation"`
- Una **clave estable** para recuperación automática — usada por los especialistas siguientes para obtener exactamente el artefacto que necesitan

Cuando corre el siguiente especialista, consulta la base de conocimiento por clave. Si el artefacto existe, procede normalmente. Si falta, anota la brecha en `open_items` y continúa con lo que tiene disponible — sin errores duros, sin pipelines bloqueados.

Esto significa que el orden de ejecución es flexible. Podés empezar con cualquier especialista. Podés pausar entre pasos. La base de conocimiento retiene el estado.

## Memory providers

ASDT requiere un **memory provider** — un almacén persistente que sobrevive a los cierres de sesión y hace los artefactos disponibles entre distintas invocaciones de especialistas.

Un memory provider le da a ASDT:
- **Persistencia entre sesiones** — el output del Arquitecto del lunes está disponible para el Developer del jueves
- **Scope por proyecto** — los artefactos de proyectos distintos no se mezclan
- **Recuperación por clave** — búsqueda determinística sin matching difuso

Hoy, **Engram** es el memory provider soportado. Más providers están planificados.

### Engram — implementación actual

[Engram](https://github.com/Gentleman-Programming/engram) es un servidor MCP (Model Context Protocol) que provee memoria persistente y con scope de proyecto para asistentes de IA.

Engram tiene que estar corriendo antes de invocar cualquier especialista (o cualquier memory provider que tengas configurado). Los artefactos guardados en Engram sobreviven al cierre de una sesión de Claude Code — esa es la propiedad de la que ASDT depende para la continuidad del pipeline entre sesiones.

#### Instalación

Seguí la [guía de setup de Engram](https://github.com/Gentleman-Programming/engram) para instalar e iniciar el servidor MCP. Luego agregalo a la configuración MCP de tu asistente de IA.

#### Verificar que está corriendo

```
/asdt-init
```

El especialista de init verifica la conectividad con el memory provider como parte del setup e informa si el servidor MCP es inalcanzable.

## Cómo se almacenan los artefactos

Cada artefacto se guarda con:

- **`title`** — legible, ej. `"add-auth/developer/dev-implementation"`
- **`topic_key`** — clave para recuperación automática, ej. `"add-auth/developer/dev-implementation"`
- **`type`** — `architecture`, `decision`, `bugfix`, etc.
- **`project`** — el nombre del proyecto, para delimitar resultados de búsqueda

Cuando un especialista necesita un artefacto previo, llama a `mem_search` con el topic_key y luego a `mem_get_observation` para obtener el contenido completo. Es una búsqueda de un paso — sin matching difuso, sin escaneo de contexto.

## Continuidad entre sesiones

Retomar un pipeline después de cerrar una sesión es igual que continuarlo en medio de una:

```
/asdt-developer Implementar basándose en el ADR del Arquitecto
```

El Developer busca el artefacto del Arquitecto en la base de conocimiento por clave. Si lo encuentra, procede normalmente. Si no, anota el input faltante y continúa con el contexto disponible.

Los límites de sesión no importan. El output del Arquitecto del lunes es tan legible para el Developer del jueves como si hubieran corrido uno tras el otro.

## Scope por proyecto

Cada llamada a `mem_save` incluye un campo `project`. ASDT deriva el nombre del proyecto de `.asdt/config.yaml` o del nombre del directorio. Los artefactos son buscables dentro del scope de un proyecto — correr ASDT en dos proyectos distintos no mezcla su memoria.

## Qué no es Engram

Engram no es un sistema de archivos. Los artefactos son documentos, no archivos de código. Cuando el Developer produce snippets de código como parte de su artefacto de implementación, esos snippets viven dentro del documento de Engram — son especificaciones de qué escribir, no archivos en disco. El humano (o el asistente de IA en un paso posterior) los aplica al codebase real.

Esta separación es intencional. Los archivos de código tienen historial de git. Los artefactos tienen la base de conocimiento. Cada uno vive donde corresponde.
