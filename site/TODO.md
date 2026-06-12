# TODO — site-v2 (for Claude Design)

Design and polish tasks queued for the next design pass.

## Visual design

- [ ] WhyAsdt section: evaluate whether the three cards need illustration or icon polish beyond the current Heroicons outline paths
- [ ] Orchestrator card: add a visual "super-command" treatment — gradient border, or a label strip "Orchestrate everything" above the card
- [ ] PipelineDiagram: review animation timing (draw at 0.5s/node, glow loop); add an accent glow on the active node
- [ ] Hero secondary CTA: verify ghost button contrast in light theme meets WCAG AA
- [ ] Footer: consider adding a version badge or "Built with ASDT" attribution line

## Content

- [ ] Spanish translation for user-flows.md (currently FallbackNotice fallback)
- [ ] OG image (1200×630, dark theme, ASDT brand) for social sharing meta tags
- [ ] Add `<meta name="description">` to all doc pages (frontmatter description is present; BaseLayout needs to wire it to the meta tag if not already done)

## QA

- [ ] Run axe-core / Lighthouse against the built site before next deploy
- [ ] Smoke test docs sidebar navigation at 375px viewport (iPhone SE)
- [ ] Verify PipelineDiagram animation in Safari (SVG CSS animations have historically had issues)
- [ ] Confirm prefers-reduced-motion completely disables the pipeline animation (not just slows it)
