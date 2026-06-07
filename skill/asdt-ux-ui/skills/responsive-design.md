# Responsive Design Guidelines

## Purpose

Mobile-first approach and breakpoint strategy. Applied during Step 6 of the UX/UI workflow.

## Mobile-First Principle

Design the smallest viewport first, then add complexity as viewport grows. Every component must be functional and readable at 320px before considering larger breakpoints.

## Standard Breakpoints

| Name | Min width | Target device |
|------|-----------|---------------|
| `xs` | 320px | Small phones |
| `sm` | 640px | Large phones / small tablets |
| `md` | 768px | Tablets portrait |
| `lg` | 1024px | Tablets landscape / small laptops |
| `xl` | 1280px | Desktop |
| `2xl` | 1440px | Wide desktop |

When documenting responsive behavior, always state which breakpoints trigger layout changes — don't just say "mobile" and "desktop".

## Fluid vs Fixed Layouts

- **Fluid**: content stretches proportionally with the viewport. Use for content-heavy pages (articles, dashboards, forms).
- **Fixed**: content has a max-width container centered in the viewport. Use when line length or readability matters (max ~80 characters per line for body text).
- **Hybrid**: fluid within a max-width cap. Preferred default for most application UIs.

Max-width caps by content type:
- Narrow forms / wizards: `480px`
- Standard content panels: `768px`
- Wide dashboards / data tables: `1280px`
- Full-bleed layouts: none (fluid)

## Touch Target Sizing

- Minimum tappable area: **44×44px** (Apple HIG / WCAG 2.5.5).
- Minimum spacing between adjacent targets: **8px**.
- Inline text links in dense content are exempt only when an explicit icon/button alternative is nearby.
- Flag every component that has clickable elements smaller than 44px in the `open_items[]` of `component-spec.yaml`.

## Responsive Typography

- Base font size: `16px` (1rem). Never go below `14px` for body text.
- Use relative units (`rem`, `em`, `%`) — never `px` for font sizes in components.
- Heading scale: reduce by 1 step at `sm` and below (e.g., H1 becomes H2 size on mobile).
- Line height for body: `1.5`. For headings: `1.2`.
- Max line length for readable body text: 45–75 characters.

## Image Handling

- Always specify `width` and `height` attributes (or CSS equivalents) to prevent layout shift.
- Use `srcset` / `sizes` for raster images above 200px wide.
- Prefer SVG for icons and illustrations.
- Lazy-load images below the fold.
- Document in `component-spec.yaml` if a component contains images that need responsive handling.

## Component Behavior at Breakpoints

For each component in `component-spec.yaml`, document:
1. Default layout (mobile-first)
2. Which breakpoint triggers layout change
3. What changes at that breakpoint (stacked → side-by-side, hidden → visible, etc.)

Example notation:
```
xs–sm: single-column stacked layout
md+: two-column grid, label left / input right
lg+: three-column grid with sidebar
```
