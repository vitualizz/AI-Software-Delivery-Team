# ADR-012: In-Repo Astro Site under site/ Deployed via GitHub Pages Actions Mode

Date: 2026-06-11
Status: Accepted

## Context

The ASDT project needed a public web presence beyond the GitHub README. The user
wanted an Astro application in this repo with a GitHub Actions pipeline. Three
decisions had to be made together: where the site lives, which platform deploys it,
and the workflow shape.

`docs/` was proposed by the user but rejected — it is the live home of ADR-001..011
plus contributing.md and development.md, making relocation a hard constraint.

## Decision

1. The Astro site lives in `site/` — a fully self-contained directory with its own
   `package.json`, lockfile, and `tsconfig.json`. The Go toolchain ignores it.

2. Deployment uses GitHub Pages in Actions-artifact mode: `configure-pages` +
   `upload-pages-artifact` + `deploy-pages`. Permissions `pages:write` and
   `id-token:write` (OIDC — zero stored secrets) are scoped to the deploy job only.
   Published URL: `https://vitualizz.github.io/asdt/`.

3. A single `.github/workflows/site.yml` handles both PR build checks (no deploy)
   and main-branch deploys (with deploy job gated on push). Path filters
   (`site/**`, `.github/workflows/site.yml`) prevent site changes from triggering
   the Go `ci.yml`.

4. `astro.config.mjs` sets `base: '/asdt/'`. All internal
   hrefs use `import.meta.env.BASE_URL` via the `getBaseHref()` utility.

5. If SSR, API functions, or Vercel-specific features enter scope in a future
   iteration, the migration path is: remove `base`, add `@astrojs/vercel` adapter,
   swap the deploy job. Site content and components travel intact.

## Consequences

- Entire CI/CD pipeline lives in this repo's GitHub Actions — zero third-party
  accounts or stored secrets.
- `base: '/asdt/'` must be set in every internal href; removing
  it requires a sweep if the repo is ever renamed or a custom domain added.
- No PR preview URLs — pull requests get a build check only, not a browsable deploy.
- The Node/Bun toolchain now coexists with Go in the same repo; contributors working
  on the site need Bun installed locally.
