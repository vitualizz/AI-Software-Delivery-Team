# Responsive Strategy — UX/UI Specialist

## Purpose
Define how each component and layout adapts across device sizes.
Mobile-first: start with the smallest viewport, expand up.

## Inputs
- `ux-ui/components`: component inventory (reused, extended, new)

Extract: new_components and extended_components (these need explicit responsive specs).
Reused components already have responsive behavior — note their breakpoint behavior only if the
feature requires overriding it.

## Context budget
ux-ui/components (new + extended only): max 1,000 tokens.

## Processing
For each new and extended component:
1. MOBILE (320-767px): describe the layout, what is visible, what collapses or stacks.
2. TABLET (768-1023px): describe changes from mobile.
3. DESKTOP (1024px+): describe the full desktop layout.
4. TOUCH TARGETS: confirm all interactive elements are ≥ 44×44px on mobile.
5. CONTENT PRIORITY: if content must be hidden on small screens, which content and why?

Apply mobile-first thinking: if something works on mobile, it works everywhere.
Never hide critical actions on mobile — collapse or reorder instead.

## Output
Produces: `ux-ui/responsive`

Persist via mem_save under the output_topic_key in workflow.yaml; return envelope.

Schema:
```yaml
payload:
  breakpoint_strategy:
    mobile: ""    # general layout approach
    tablet: ""
    desktop: ""
  component_behavior:
    - component: ""
      mobile: ""
      tablet: ""
      desktop: ""
      touch_target_compliant: true
  hidden_on_mobile: []    # with justification
  open_items: []
```
