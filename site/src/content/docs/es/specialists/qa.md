---
title: QA Engineer
description: Construye la red de seguridad antes de que el código salga a producción — planes de prueba, validación de criterios de aceptación, análisis de casos borde e informes de calidad — el especialista a invocar cuando "funciona en mi máquina" no es suficiente.
order: 23
locale: es
---

# QA Engineer (`/asdt-qa`)

> Construye la red de seguridad antes de que el código salga a producción — planes de prueba, validación de criterios de aceptación, análisis de casos borde e informes de calidad — el especialista a invocar cuando "funciona en mi máquina" no es suficiente.

## Qué hace

El especialista QA valida los criterios de aceptación, descubre casos borde de forma sistemática, define la estrategia de testing en la pirámide (unitario / integración / e2e) y produce un informe de calidad con un veredicto de ship-readiness. Parte de cualquier artefacto previo que exista — implementación del Developer, decisiones de arquitectura o requisitos crudos — y los normaliza en una lista de ACs testeables antes de escribir un solo caso de prueba.

`ac-validation` siempre corre sin importar la complejidad — las brechas en los ACs deben exponerse, no ignorarse en silencio. Un AC malo produce un test malo; el especialista QA corrige el AC primero, después genera casos de prueba contra la versión corregida.

El especialista QA no es elegible para complejidad trivial. En trivial vuelve a `simple`, porque no existe un conjunto de pasos completo por dependencias por debajo de ese nivel.

## Cuándo invocarlo

- El código está listo para revisión y necesitás un quality gate antes de que salga a producción
- Los criterios de aceptación existen pero no han sido validados formalmente (atomicidad, mensurabilidad, independencia)
- Querés cobertura sistemática de casos borde, no solo tests del happy path
- Necesitás un plan de pruebas estructurado que un Developer pueda implementar sin adivinar

## Posición en el pipeline

Típicamente corre **después del Developer** (lee `dev-implementation`) y es el sign-off final antes de que el código se mergee. Puede correr antes — contra el `backlog-entry` del PM o el `architectural-decision` del Arquitecto — para detectar problemas de calidad en los ACs antes de que empiece la implementación. Ese pase temprano ahorra mucho más tiempo que encontrar las brechas después de que el código está escrito.

## Qué produce

`test-plan` — el artefacto de calidad final y sign-off. Contiene: resumen de tests (conteos unitario/integración/e2e), porcentaje de cobertura de ACs, brechas en ACs sin cobertura, el veredicto de calidad con rationale y la lista completa de casos de prueba.

Consumido por: **Developer** (para implementar la suite de tests), usado como artefacto de sign-off antes del merge.

## Patrones comunes

```
/asdt-qa Revisar el flujo de checkout en busca de casos borde
# → El happy path está testeado pero las condiciones límite y las rutas de error necesitan cobertura
```

```
/asdt-qa Validar criterios de aceptación antes de que empiece la implementación
# → Correr QA sobre el backlog-entry del PM para detectar problemas de calidad en los ACs temprano
```

```
/asdt-qa Construir un plan de pruebas para el módulo de autenticación
# → Estrategia completa de pirámide de testing para código sensible a seguridad
```

## Límites — qué NO hace

- No escribe código de implementación
- No escribe decisiones de arquitectura ni specs de UX
- `ac-validation` no puede omitirse — las brechas en los ACs siempre deben exponerse
- `test-strategy` es un input requerido para la generación de casos de prueba en moderate+ — no se puede omitir
- Los casos de prueba son especificaciones (Given/When/Then) — no código ejecutable
