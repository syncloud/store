# Syncloud Store Web

Vue 3 + Vite single-page app that lists apps from the Syncloud store and lets you search them.

## Scripts (web/)

- `npm run dev` — dev server, expects a real backend at `localhost:8080` (proxied at `/v2`, `/api`)
- `npm run dev:stub` — dev server with MirageJS stubbing the backend so you can preview the design without running anything else
- `npm run build` — production build into `dist/` (no stub data)
- `npm run build:stub` — build into `dist/` with MirageJS bundled, for tests / standalone demo
- `npm run preview` / `preview:stub` — serve `dist/` on `127.0.0.1:4173`

## Playwright

E2e tests live in `web/e2e/` as a self-contained TypeScript package. Mirrors the layout of the `onlyoffice` app.

```sh
cd web/e2e
npm install
npm test                   # boots ../web's preview:stub via webServer, then runs all specs
```

Artifacts (screenshots, traces, html report) land under `web/e2e/artifact/playwright/<project>/`. Override the location via `PLAYWRIGHT_ARTIFACT_DIR`.

Tests select by `data-testid` only.
