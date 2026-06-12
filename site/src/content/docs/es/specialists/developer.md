---
title: Developer
description: Convierte specs y diseños en código funcional — planes de implementación, código de producción y suites de tests — el especialista a invocar una vez que la forma de la solución está definida y es momento de construirla.
order: 22
locale: es
---

# Developer (`/asdt-developer`)

> Convierte specs y diseños en código funcional — planes de implementación, código de producción y suites de tests — el especialista a invocar una vez que la forma de la solución está definida y es momento de construirla.

## Qué hace

El especialista Developer transforma requisitos existentes, specs de UX y decisiones de arquitectura en una implementación concreta. Lee el codebase primero (siempre), define qué se construirá y qué no, elige el enfoque técnico, divide el trabajo en tareas atómicas y produce un plan de implementación o escribe archivos directamente en el repositorio.

Dos modos de operación controlan las escrituras reales de archivos. En **modo plan-only** (por defecto), produce código como snippets en la base de conocimiento — sin cambios en el repositorio. En **modo escritura**, los targets de archivos declarados se resuelven y validan antes de cualquier escritura — si un path necesario está fuera de los targets declarados, se detiene y reporta el problema en lugar de escribir de forma no declarada.

`explore` y `spec` son irrenunciables — siempre corren sin importar la complejidad. El paso `test` es condicional: solo corre cuando `strict_tdd: true` está configurado en `.asdt/config.yaml`.

## Cuándo invocarlo

- La forma de la solución ya está definida (requisitos, arquitectura o UX están definidos)
- Necesitás un plan de implementación concreto con tareas ordenadas y targets a nivel de archivo
- Querés código de producción escrito directamente en el codebase (modo escritura, con targets declarados)
- Estás retomando desde un artefacto anterior del Arquitecto o PM almacenado en la base de conocimiento

## Posición en el pipeline

Típicamente corre **después del Arquitecto** (lee `architectural-decision` + `system-design`) y produce el `dev-implementation` final consumido por QA. Puede correr standalone con solo una descripción de la petición — explorará y especificará el problema él mismo sin artefactos previos. En complejidad `simple`, bypassa el Arquitecto por completo.

## Qué produce

`developer/dev-implementation` — el artefacto de implementación consolidado. En modo plan-only contiene snippets de código ordenados. En modo escritura contiene el manifest de archivos escritos y el rationale de cada uno.

Consumido por: **QA** (lee la implementación para validar contra criterios de aceptación y producir casos de prueba).

## Patrones comunes

```
/asdt-developer Implementar el componente de perfil de usuario
# → Forma definida por trabajo previo de UX/UI y Arquitecto, ahora es momento de construir
```

```
/asdt-developer Agregar exportación CSV al panel de reportes
# → Petición standalone — explorará, especificará e implementará sin artefactos previos
```

```
/asdt-developer Implementar basándose en el ADR del Arquitecto
# → Lee architectural-decision de la base de conocimiento automáticamente
```

## Límites — qué NO hace

- No produce decisiones de arquitectura ni ADRs
- No escribe specs de UX, wireframes ni specs de componentes
- No produce planes de prueba ni informes de calidad (el paso test solo genera código de tests, no planes)
- Nunca escribe archivos fuera de los targets declarados en modo escritura — se detiene y reporta
- `explore` y `spec` no pueden omitirse sin importar la complejidad
