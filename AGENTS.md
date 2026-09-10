# Project Instructions

## Product Context

Read [product.md](product.md) before planning or implementing product features.
It defines the product's purpose and boundaries. Keep implementation aligned
with it, and do not treat provisional terminology or unspecified capabilities
as approved requirements.

## Delivery Priority

This project is in a rapid implementation phase. Prioritize delivering the
smallest working feature so the user can validate it manually. Keep changes
focused, straightforward, and easy to extend. Do not overengineer or expand the
scope with speculative abstractions or unrelated cleanup.

## Validation Policy

- Do not run full test suites unless the user explicitly asks.
- Do not perform adversarial reviews or launch review agents unless the user
  explicitly asks. These activities come later, after human validation.
- Do not write integration tests or complex unit tests during normal feature
  implementation. Add them only when explicitly requested.
- Avoid repeated test runs and broad verification workflows. Human validation is
  the primary acceptance gate for feature behavior at this stage.
- Still meet baseline engineering quality: valid syntax, valid imports, sound
  types, intact files, and a minimally working implementation.
- Use the smallest relevant verification step, such as a targeted type check,
  syntax check, or build when needed to catch compilation or import problems.
  Do not run several overlapping checks by default.
- Fix obvious errors introduced by the change. Do not knowingly leave broken
  code, expose secrets, or compromise essential security or data integrity for
  speed.
- State briefly what was implemented, which lightweight check was performed,
  and what remains for the user to validate. Never claim tests passed or behavior
  was verified when those checks were not run.
- Human validation alone does not authorize full suites or adversarial reviews;
  wait for an explicit request before performing them.

## Project Layout

- `clients/desktop`: React, TypeScript, and Vite frontend.
- `engine`: Standalone Go engine for persistence and headless harness execution.
- `clients/desktop/src/design-system`: Shared UI primitives and visual tokens.
- `clients/desktop/src/features`: Feature-specific components and behavior.
- `code.html`: Original visual reference. Preserve it unless asked to change it.

Reuse existing design-system components and tokens. Keep application behavior
outside generic UI primitives. Preserve the established visual language unless
the user requests a redesign.

## UI Copy

- Keep all application-controlled UI text in English, including picker buttons.
- Assume users understand the tool. Avoid obvious explanations, tutorial-like
  helper text, and redundant labels such as "Mock", "Demo data", or "UI preview".
- Show only information that helps users decide, act, or recover from an error.
  Keep implementation limitations in documentation rather than repeating them
  throughout the interface.
- Use custom, English-labeled buttons for native file inputs instead of exposing
  browser-generated text such as "No file chosen". Native system dialogs may
  still follow the operating system's language.

## Collaboration

Implement directly when the request is clear. Ask only when a missing decision
materially blocks the work. Keep progress updates and final explanations concise.
Preserve unrelated user changes. Do not commit or push unless explicitly asked.
