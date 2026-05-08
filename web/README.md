# Syncloud Store Web

Vue 3 + Vite single-page app that lists apps from the Syncloud store and lets you search them.

## Scripts

- `npm run dev` — dev server, expects a real backend at `localhost:8080` (proxied at `/v2`, `/api`)
- `npm run dev:stub` — dev server with MirageJS stubbing the backend so you can preview the design without running anything else
- `npm run build` — production build into `dist/`
- `npm run preview:stub` — production build with the stub baked in, served on `:4173` (used by Playwright)
- `npm run test:e2e` — Playwright tests (boots `preview:stub` automatically)

## Locator convention

Tests select by `data-testid` only. When you add a new element a test needs to reach, add `data-testid="..."` to the template and use `page.getByTestId(...)` in the spec.
