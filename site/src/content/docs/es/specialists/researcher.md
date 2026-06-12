---
title: Researcher
description: Explora problemas y oportunidades difusas mediante ideación divergente y escaneo de factibilidad, convergiendo en una única dirección recomendada — el especialista a invocar antes de que existan los requisitos, cuando todavía no sabés qué construir.
order: 26
locale: es
---

# Researcher (`/asdt-researcher`)

> Explora problemas y oportunidades difusas mediante ideación divergente y escaneo de factibilidad, convergiendo en una única dirección recomendada — el especialista a invocar antes de que existan los requisitos, cuando todavía no sabés qué construir.

## Qué hace

El especialista Researcher diverge antes de que PM converja. Toma un problema u oportunidad difusa y ejecuta una secuencia de descubrimiento estructurada: encuadra el problema y genera direcciones candidatas deliberadamente divergentes, evalúa cada una con un veredicto de factibilidad basado en evidencia, y luego converge en una única dirección recomendada empaquetada como un `discovery-brief`.

Dos propiedades mantienen el contrato honesto: la ideación es **generativa, nunca selectiva** — el paso de ideación produce candidatas sin rankearlas, así las direcciones prometedoras pero inusuales sobreviven lo suficiente para ser evaluadas. Y el brief recomienda exactamente **una** dirección con justificación explícita; las candidatas que no quedaron se registran como ítems descartados que alimentan la lista de fuera de alcance de PM.

El Researcher es solo analista — nunca escribe en el filesystem. Su único trabajo es convertir "no sabemos qué construir" en una recomendación con base de factibilidad que PM pueda tratar como una petición bien formada.

## Cuándo invocarlo

- El problema o la oportunidad es difusa ("estamos perdiendo usuarios en algún lado", "los costos parecen demasiado altos")
- La dirección no está clara — existen múltiples soluciones plausibles y nadie las comparó
- Necesitás una recomendación con base de factibilidad **antes** de escribir requisitos
- Estás sopesando trade-offs de construir-vs-comprar o enfoque-vs-enfoque en la etapa de idea

## Posición en el pipeline

El único especialista **pre-PM** del pipeline — corre antes de que existan los requisitos. El resumen y la dirección recomendada de su `discovery-brief` se renderizan como prosa y se entregan al feature-intake de `/asdt-pm` como la petición cruda, así PM arranca desde una dirección explorada y verificada en factibilidad en lugar de una suposición. Puede correr de forma standalone cuando solo necesitás exploración estructurada sin continuar hacia requisitos.

## Qué produce

**researcher/ideation** — el problema encuadrado más las direcciones candidatas divergentes, sin rankear por diseño. **researcher/feasibility** — un veredicto verde/amarillo/rojo por candidata con evidencia de soporte y estimaciones de esfuerzo. **researcher/discovery-brief** — la recomendación convergida: una dirección, justificación, notas de factibilidad y candidatas descartadas.

Consumido por: **PM** (lee el discovery-brief como su petición cruda; las `wont_candidates` alimentan la lista de fuera de alcance del backlog-entry). En el nivel trivial solo se produce el artefacto de ideación.

## Patrones comunes

```
/asdt-researcher Estamos perdiendo usuarios durante el onboarding pero no sabemos por qué
# → Problema difuso, necesita encuadre y direcciones candidatas antes de los requisitos
```

```
/asdt-researcher ¿App móvil nativa o PWA para soporte offline?
# → Direcciones en competencia, necesita veredictos de factibilidad antes de comprometerse
```

```
/asdt-researcher Explorar formas de reducir nuestros costos de infraestructura
# → Oportunidad abierta, necesita ideación divergente antes de que alguien elija un camino
```

## Límites — qué NO hace

- No escribe requisitos ni historias de usuario (eso es trabajo de PM)
- No escribe decisiones de arquitectura ni ADRs
- No escribe código de implementación ni tests
- Nunca actúa como builder — solo analista, nunca escribe en el filesystem
- Nunca reemplaza a PM — su brief alimenta el intake de PM, no lo saltea
- La ideación nunca rankea candidatas — la convergencia ocurre solo en el brief, que recomienda exactamente UNA dirección
