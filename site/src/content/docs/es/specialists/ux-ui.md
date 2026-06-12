---
title: Diseño UX/UI
description: Da forma a cómo las personas experimentan el producto — flujos de usuario, arquitectura de información, specs de componentes, estrategia responsive y de accesibilidad — el especialista a invocar antes de que se construya una sola pantalla.
order: 25
locale: es
---

# Diseño UX/UI (`/asdt-ux-ui`)

> Da forma a cómo las personas experimentan el producto — flujos de usuario, arquitectura de información, specs de componentes, estrategia responsive y de accesibilidad — el especialista a invocar antes de que se construya una sola pantalla.

## Qué hace

El especialista UX/UI transforma un brief de feature en una especificación UX estructurada que el Developer puede implementar sin ambigüedad. Identifica primero al actor principal y el problema central, luego organiza el contenido en una arquitectura de información, mapea las secuencias de interacción completas (happy path + rutas de error + casos borde), cataloga qué componentes reutilizar, extender o crear desde cero y — en el nivel complex — define cómo se comporta cada componente en distintos breakpoints.

`ux-handoff` siempre corre sin importar la complejidad. Ese es el paso de consolidación que produce los dos artefactos finales consumidos por Developer y Arquitecto. Los niveles `simple` y `moderate` son intencionalmente idénticos — `responsive-strategy` es el único paso que está restringido a `complex`, porque los features simples raramente necesitan specs explícitas de múltiples breakpoints.

Una dependencia fuerte: `information-architecture` debe correr antes que `user-flows`. No se pueden mapear secuencias de interacción antes de conocer la jerarquía de contenido y la ruta de navegación.

## Cuándo invocarlo

- Una pantalla, diálogo o feature-level UI nueva necesita ser diseñada
- Los flujos de usuario necesitan mapearse antes de que empiece la arquitectura o la implementación
- La estrategia de reutilización de componentes necesita decidirse (extender existentes vs. crear nuevos)
- Los requisitos de accesibilidad necesitan especificarse explícitamente
- Querés que el Developer reciba una spec en lugar de inferir la UX de los requisitos

## Posición en el pipeline

Funciona mejor **antes del Developer** — el `ux-brief` y el `component-spec` son inputs que el Developer lee para implementar la UI correctamente. Puede correr en paralelo con el Arquitecto, ya que el diseño UX y las decisiones de arquitectura son en gran medida independientes. Correrlo después de que el Developer ya construyó una pantalla significa que la spec llega demasiado tarde para guiar la implementación.

## Qué produce

Dos artefactos finales:

- **`ux-brief`** — resumen del feature, actor principal, criterios de éxito, flujos de usuario (happy path + puntos de decisión), arquitectura de información
- **`component-spec`** — inventario completo de componentes: reutilizados (con caso de uso), extendidos (con cambios necesarios), nuevos (con razón, props, eventos, comportamiento responsive)

Consumido por: **Developer** (lee ambos para implementar la UI), **Arquitecto** (lee `ux-brief` para entender los flujos de usuario al diseñar contratos de API).

## Patrones comunes

```
/asdt-ux-ui Diseñar el flujo de onboarding para nuevos usuarios
# → UI multi-paso nueva — necesita IA completa + flujos antes de cualquier trabajo de componentes
```

```
/asdt-ux-ui Mapear la pantalla de preferencias de notificaciones
# → Patrón de UI existente a extender — component-mapping identificará oportunidades de reutilización
```

```
/asdt-ux-ui Especificar el layout móvil del dashboard
# → Nivel complex con responsive-strategy para comportamiento explícito de breakpoints
```

## Límites — qué NO hace

- No escribe código de implementación — solo especificaciones y estructura
- No produce decisiones de arquitectura ni planes de prueba
- Nunca propone componentes inconsistentes con el design system existente
- La UI generada debe sentirse como parte de la aplicación existente
- `information-architecture` no puede omitirse antes de `user-flows`
- `ux-handoff` siempre corre — la consolidación no es opcional
