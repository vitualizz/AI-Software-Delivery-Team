---
title: Product Manager
description: Transforma peticiones de features sin estructura en backlog entries con historias de usuario, límites de alcance y priorización — el especialista a invocar antes de la arquitectura o el código cuando los requisitos necesitan formalización.
order: 20
locale: es
---

# Product Manager (`/asdt-pm`)

> Transforma peticiones de features sin estructura en backlog entries con historias de usuario, límites de alcance y priorización — el especialista a invocar antes de la arquitectura o el código cuando los requisitos necesitan formalización.

## Qué hace

El especialista PM convierte peticiones vagas en un artefacto de requisitos preciso que todos los demás especialistas pueden consumir sin ambigüedad. Extrae el problema central, identifica stakeholders, escribe historias de usuario con criterios de aceptación preliminares, define límites de alcance explícitos y consolida todo en un `backlog-entry` que fluye hacia los siguientes especialistas.

Dos propiedades hacen que el contrato del backlog-entry sea estricto: los límites de alcance son **obligatorios** — un backlog-entry sin ítems explícitos fuera del alcance se considera incompleto, porque la ambigüedad de alcance es la causa raíz del scope creep. Y los criterios de aceptación en el backlog-entry son condiciones en inglés simple de alto nivel — **no** son criterios de prueba finales. QA los formaliza en formato Given/When/Then.

El especialista PM nunca escribe decisiones de arquitectura, código de implementación ni specs de UX. Su único trabajo es hacer que los requisitos sean inequívocos para que ningún especialista posterior tenga que adivinar.

## Cuándo invocarlo

- La petición está formulada en lenguaje vago o centrado en el usuario ("agregar modo oscuro", "mejorar la búsqueda")
- Necesitás historias de usuario explícitas antes de que el Arquitecto o el Developer intervengan
- El alcance necesita quedar cerrado antes de que empiece el trabajo para evitar expansión en el medio del sprint
- Hay múltiples stakeholders con necesidades que hay que conciliar

## Posición en el pipeline

Funciona mejor como el **primer** especialista en un pipeline — su `backlog-entry` es la fuente principal de requisitos para Arquitecto, Developer y QA. Invocarlo después de que la arquitectura ya está decidida arriesga divergencia entre requisitos y diseño. Puede correr de forma standalone cuando solo necesitás requisitos formalizados sin continuar el pipeline.

## Qué produce

`pm/backlog-entry` — el artefacto canónico de requisitos. Contiene: nombre del feature, resumen ejecutivo, historias de usuario ordenadas con criterios de aceptación, bloque completo de alcance (dentro/fuera, puntos de integración, flags de riesgo) e ítems abiertos para especialistas posteriores.

Consumido por: **Arquitecto** (lee resumen ejecutivo + alcance), **Developer** (lee historias de usuario + orden de prioridad), **QA** (lee historias de usuario + criterios de aceptación como fuente principal de requisitos).

## Patrones comunes

```
/asdt-pm Agregar autenticación de usuario con email y contraseña
# → Requisitos ambiguos, necesita alcance antes de la arquitectura
```

```
/asdt-pm Rediseñar el sistema de notificaciones
# → Múltiples stakeholders con necesidades potencialmente en conflicto
```

```
/asdt-pm Agregar exportación CSV al panel de reportes
# → Simple en la superficie, pero los puntos de integración y el riesgo de alcance necesitan mapearse
```

## Límites — qué NO hace

- No escribe decisiones de arquitectura ni ADRs
- No escribe código de implementación ni diseños técnicos
- No escribe specs de UX, wireframes ni specs de componentes
- No produce criterios de aceptación finales y testeables (eso es trabajo de QA)
- Nunca produce un backlog-entry sin ítems explícitos fuera del alcance
