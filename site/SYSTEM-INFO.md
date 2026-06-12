# SYSTEM-INFO — ASDT Site

Reference document for any AI agent or developer working on the ASDT marketing and documentation site (`site/`).

## What ASDT is

ASDT (AI Software Delivery Team) installs a team of AI specialists into Claude Code or OpenCode. The six specialists — PM, Architect, Developer, QA, Security, UX/UI — each own a distinct discipline and pass structured artifacts to the next step through Engram (persistent memory).

`/asdt` is a **pipeline advisor** — it analyzes a feature request and recommends which specialists to involve and in what order. It does NOT execute specialists automatically. The user confirms the routing plan and runs each specialist command. Each specialist is its own orchestrator: it runs through defined steps (explore → spec → design → implement) internally.

**Supported AI coding assistants:** Claude Code, OpenCode  
**Memory provider:** Engram MCP server (required; more providers planned)  
**Install:** `curl -fsSL …/install.sh | bash` (pre-built binary, no Go needed for end users)

## Tech stack

| Layer        | Technology                        |
|-------------|-----------------------------------|
| Framework    | Astro 6.4.6 (static output)       |
| CSS          | Tailwind CSS v4 (CSS-first)        |
| Package mgr  | Bun                               |
| Language     | TypeScript (strict)               |
| Deploy       | GitHub Pages via Actions          |
| Base URL     | `/asdt`                           |

## Project structure

```
site/
├── src/
│   ├── components/         # Zero client JS (except CodeBlockCopy + ThemeToggle)
│   ├── content/docs/en/    # Markdown docs — order field controls nav sort
│   ├── i18n/
│   │   ├── types.ts        # UIStrings interface — strict TS type contract
│   │   ├── locales/en.ts   # English strings
│   │   └── locales/es.ts   # Spanish strings
│   ├── layouts/            # BaseLayout, DocsLayout
│   ├── pages/              # File-based routing (en at root, es at /es/)
│   └── styles/global.css   # @theme CSS custom properties + Tailwind import
├── public/                 # Static assets (favicon etc.)
├── astro.config.mjs        # i18n config, base path, Vite plugins
├── TODO.md                 # Design backlog (this session)
└── SYSTEM-INFO.md          # This file
```

## Color tokens

| Token                | Dark (#0d1117 bg)  | Light (#ffffff bg) |
|----------------------|--------------------|---------------------|
| `--color-bg`         | `#0d1117`          | `#ffffff`           |
| `--color-surface`    | `#161b22`          | `#f6f8fa`           |
| `--color-text`       | `#e6edf3`          | `#1f2328`           |
| `--color-text-muted` | `#8b949e`          | `#636c76`           |
| `--color-accent`     | `#3fb950`          | `#1a7f37`           |
| `--color-accent-hover`| `#46c95e`         | `#218a3d`           |
| `--color-error`      | `#f85149`          | `#d1242f`           |
| `--color-border`     | `#30363d`          | `#d0d7de`           |

Theme is toggled by `data-theme="light"` on `<html>`. Persisted in localStorage as `asdt-theme`.

## Typography

- `--font-sans`: `'Inter Variable', ui-sans-serif, system-ui, sans-serif`
- `--font-mono`: `'JetBrains Mono Variable', ui-monospace, SFMono-Regular, monospace`

Tailwind v4 syntax: `font-mono` → `font-(--font-mono)` style is NOT needed; `font-mono` utility maps correctly via `@theme`.

## Key conventions

1. **Zero new JS** — no `<script>` tags beyond the theme-persistence IIFE in BaseLayout and the CodeBlockCopy island. Use CSS for all animation and interaction.
2. **Tailwind v4 CSS-first** — use `text-(--color-text)` syntax (arbitrary value with CSS var), never hardcode hex values in class names.
3. **All hrefs via `getBaseHref(path)`** — never use root-absolute paths; the base URL prefix is required for GitHub Pages.
4. **UIStrings atomic update rule** — `types.ts`, `en.ts`, and `es.ts` must always be updated together. Adding a key to only one file breaks the TS build.
5. **Content collection schema** — docs frontmatter requires `title: string`, `description: string`, `order: number`, `locale: 'en' | 'es'`.
6. **Docs nav order** — controlled by `order` frontmatter AND the `slugToNav` map in `DocsLayout.astro`. Both must be updated when adding a new doc page.

## Component inventory

| Component          | Type      | JS? | Purpose                                    |
|--------------------|-----------|-----|--------------------------------------------|
| BaseLayout         | Layout    | No  | HTML shell, theme script, skip link        |
| DocsLayout         | Layout    | No  | Sidebar nav + article area                 |
| Hero               | Section   | No  | Landing hero with install command + CTAs   |
| WhyAsdt            | Section   | No  | Value prop — 3-item benefit grid           |
| SpecialistsGrid    | Section   | No  | Orchestrator card + 6 specialist cards     |
| SpecialistCard     | Card      | No  | Individual specialist card                 |
| PipelineDiagram    | Section   | No  | Animated SVG pipeline (CSS @keyframes)     |
| Footer             | Footer    | No  | Links + tagline                            |
| NavBar             | Nav       | No  | Top navigation                             |
| ThemeToggle        | Button    | Yes | Dark/light theme toggle (existing island)  |
| LanguagePicker     | Nav       | No  | EN/ES switcher                             |
| CodeBlockCopy      | Island    | Yes | Copy-to-clipboard (only new JS island)     |
| FallbackNotice     | Alert     | No  | Untranslated page notice                   |
| DocsNavItem        | Nav item  | No  | Sidebar nav link                           |
| SkipLink           | A11y      | No  | Skip to main content                       |
| ReleaseBadge       | Badge     | No  | GitHub release tag badge                   |
