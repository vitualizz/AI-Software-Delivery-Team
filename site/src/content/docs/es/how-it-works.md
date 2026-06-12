---
title: Cómo funciona
description: Flujo de punta a punta desde la petición de una feature hasta el código entregado — cómo ASDT orquesta especialistas a través de artefactos estructurados.
order: 5
locale: es
---

# Cómo funciona

## El modelo de ejecución

ASDT te da estructura sin quitarte el control. Cuando ejecutás un especialista, orquesta una secuencia de pasos enfocados — cada uno produce un único artefacto y solo lee lo que produjo el paso anterior.

Vos invocás a los especialistas. ASDT nunca los ejecuta automáticamente. Eso es intencional: cada paso donde un humano confirma un plan es un paso donde los supuestos incorrectos se detectan antes de que se acumulen.

El asesor `/asdt` analiza tu petición y sugiere qué especialistas involucrar y en qué orden. Vos confirmás el plan y ejecutás cada comando.

## El conocimiento fluye hacia adelante automáticamente

Cada paso de cada especialista produce un artefacto — un documento estructurado guardado en la base de conocimiento con una clave estable. El siguiente especialista lo recupera por clave. Sin pasar contexto manualmente. Sin copiar y pegar entre comandos.

Esto significa:

- **Los especialistas están desacoplados.** El Developer lee el registro de decisión del Arquitecto como un documento — no como una variable compartida ni un import de archivo.
- **Los artefactos sobreviven a las sesiones.** Ejecutás PM el lunes, continuás con el Arquitecto el jueves. La base de conocimiento retiene el contexto.
- **Los inputs faltantes se degradan sin errores.** Si falta un artefacto, el siguiente especialista lo anota en `open_items` y continúa con lo que tiene disponible.

## Los especialistas se adaptan a la complejidad

Cada especialista ejecuta la profundidad de pasos adecuada para la complejidad de la petición. Un bugfix rápido corre menos pasos que un nuevo sistema de autenticación. La tabla a continuación muestra la secuencia completa a complejidad moderada:

| Especialista | Pasos |
|---|---|
| PM | `feature-intake` → `user-stories` → `scope-analysis` → `backlog-entry` |
| Arquitecto | `load-constraints` → `evaluate-approaches` → `decision-record` → `system-design` → `risk-analysis` → `technical-handoff` |
| Developer | `explore` → `spec` → `design` → `tasks` → `implement` |
| QA | `load-requirements` → `ac-validation` → `edge-case-analysis` → `test-strategy` → `test-case-generation` → `quality-report` |
| Seguridad | `threat-modeling` → `attack-surface` → `owasp-analysis` → `hardening-checklist` |
| UX/UI | `feature-brief` → `information-architecture` → `user-flows` → `component-mapping` → `responsive-strategy` → `ux-handoff` |

Los pasos se ejecutan como sub-agentes aislados — no comparten contexto entre sí, lo que evita que el razonamiento temprano contamine los pasos posteriores. Cada paso lee solo sus inputs declarados, escribe un artefacto y pasa el control.

## El humano siempre está en el circuito

ASDT aplica una compuerta suave en dos momentos:

1. **Después de `/asdt`** — el asesor de pipeline presenta un plan de routing y espera confirmación antes de darte los comandos a ejecutar.
2. **Entre especialistas** — vos decidís cuándo correr el siguiente. Nada se automatiza.

Esto no es una limitación. Es el diseño. Las decisiones de arquitectura generadas por IA se benefician de la revisión humana antes de que un developer actúe sobre ellas. Los planes de QA se benefician de la revisión humana antes de que definan qué significa "listo". ASDT te da la estructura; vos aplicás el criterio.
