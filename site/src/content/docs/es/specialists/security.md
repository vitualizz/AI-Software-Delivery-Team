---
title: Seguridad
description: Busca los huecos que encontraría un atacante primero — modelos de amenazas, revisiones OWASP y checklists de hardening — el especialista a invocar siempre que auth, manejo de datos o integraciones externas estén sobre la mesa, en cualquier punto del pipeline.
order: 24
locale: es
---

# Seguridad (`/asdt-security`)

> Busca los huecos que encontraría un atacante primero — modelos de amenazas, revisiones OWASP y checklists de hardening — el especialista a invocar siempre que auth, manejo de datos o integraciones externas estén sobre la mesa, en cualquier punto del pipeline.

## Qué hace

El especialista de Seguridad realiza modelado de amenazas y análisis de seguridad usando STRIDE y el OWASP Top 10. Mapea la superficie de ataque, identifica amenazas sistemáticamente y produce un checklist de hardening priorizado donde cada hallazgo tiene una mitigación concreta y accionable — no "monitorearlo" ni "agregar logging."

El invariante crítico: **Seguridad no tiene predecesor requerido.** Puede correr en cualquier etapa — en un proyecto nuevo sin artefactos previos, a mitad del desarrollo o después del lanzamiento. Si existen artefactos previos (decisiones de arquitectura, implementación), los lee. Si no, trabaja desde el contexto de plataforma y la petición sola, notando las brechas en `open_items` y continuando.

La profundidad está controlada por `risk_surface`, no por complejidad. Este es el único especialista donde la pregunta no es "¿qué tan compleja es la feature?" sino "¿qué tan grande es la superficie de ataque?"

## Cuándo invocarlo

- Autenticación, gestión de sesiones o autorización están involucrados
- La feature maneja o almacena información de identificación personal
- Hay integraciones externas, webhooks o URLs controladas por el usuario
- Se están exponiendo nuevos endpoints de API públicamente
- En cualquier momento antes de salir a producción cuando la seguridad no ha sido revisada

## Posición en el pipeline

**Sin predecesor requerido** — invocalo en cualquier punto. Para máximo impacto, correlo después de que el Arquitecto produce `system-design` (Seguridad puede analizar la superficie de API y los límites de servicios). Para un modelo de amenazas temprano en el diseño, correlo antes de que la arquitectura esté finalizada para exponer riesgos a nivel de diseño antes de que queden incorporados.

Sus outputs (`security-findings` + `hardening-checklist`) son consumidos por Developer y Arquitecto para abordar las mitigaciones.

## Qué produce

Dos artefactos finales:

- **`security-findings`** — todos los hallazgos con ratings de severidad (Crítico/Alto/Medio/Bajo siguiendo CVSS-lite), referencias CWE y recomendaciones concretas
- **`hardening-checklist`** — ítems accionables agrupados por esfuerzo, con separación de debe-corregirse-antes-del-lanzamiento vs. puede-diferirse

Consumido por: **Developer** (para implementar mitigaciones), **Arquitecto** (para ajustar decisiones de diseño que introdujeron riesgos estructurales).

## Patrones comunes

```
/asdt-security Auditar la integración OAuth
# → Flujo de auth externo con manejo de tokens — superficie de riesgo alta
```

```
/asdt-security Modelar amenazas para el nuevo handler de webhooks de pago
# → Input controlado por el usuario que impacta lógica financiera
```

```
/asdt-security Revisión de seguridad rápida antes del lanzamiento de v2
# → No se necesitan artefactos previos — corre solo desde el contexto de plataforma
```

## Límites — qué NO hace

- No escribe código de implementación
- No produce decisiones de arquitectura ni specs de UX
- No produce planes de prueba (aunque sus hallazgos informan qué debería cubrir QA)
- Cada hallazgo debe tener una mitigación concreta — "agregar monitoring" no es una mitigación
- La severidad siempre sigue CVSS-lite: Crítico / Alto / Medio / Bajo — sin otra escala
