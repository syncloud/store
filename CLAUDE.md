# Project structure

- `api/`, `rest/`, `model/`, `storage/`, `cmd/` — Go backend (the snap-store implementation, served via Echo)
- `web/` — Vue 3 + Vite frontend (the public store app list and search UI)
- `test/` — Go integration tests (`store_test.go`, `snapd_test.go`) plus test snap fixtures
- `.drone.jsonnet` — Drone pipelines, one per arch (`amd64`, `arm64`, `arm`)

# Frontend (`web/`)

Vue 3 + Vite. **No Element Plus** — plain components and CSS variables for theming. Don't reintroduce a component library without a reason.

## Local preview

The app runs entirely in the browser; you don't need the Go backend to look at the design.

```sh
cd web
npm install
npm run dev:stub      # MirageJS stubs /v2/snaps/find and /v2/snaps/info
```

`npm run dev` (no `:stub`) expects the real Go store at `localhost:8080` and proxies `/v2` and `/api` to it.

## Theme

Dark/light is implemented via `data-theme` on `<html>` and CSS custom properties in `src/style/global.css`. Persisted in `localStorage` under `syncloud-store-theme`. Defaults to `prefers-color-scheme` if nothing is stored.

## Locator conventions

Tests select by `data-testid` only — never by role, text, or CSS selector.

When a new element needs to be targeted by a test, add `data-testid="..."` to the Vue template and use `page.getByTestId(...)` in the spec. And: navigate via clicks, not `page.goto('/some-inner-route')` — direct URL hits bypass the nav and hide routing bugs. (Currently this app is single-page; that rule kicks in once routes are added.)

Why:
- Role + accessible-name matching is partial-match by default and breaks when another element on the page contains the same text.
- Text matching couples tests to copy changes and translations.
- `data-testid` is stable, explicit, and survives DOM refactors.

## Playwright notes

```text
web/e2e/                  # tests
web/playwright.config.js  # config
```

Layout (mirrors `../redirect/www/`):
- shared `*.spec.js` flows run on both `desktop` and `mobile`
- `*.mobile.spec.js` holds mobile-only assertions
- screenshots are taken on failure in `web/e2e/fixtures.js`
- videos and traces are retained in `web/test-results`

Playwright boots its own server (`npm run preview:stub`) — it's a stub-backed production build, so tests don't depend on the Go store. To run against another URL, set `PLAYWRIGHT_BASE_URL`.

## Local limitations

Playwright does not run on this Termux/Android host. Syntax checks and repo edits can be done locally, but real Playwright execution must be validated in Drone's Linux environment.

# CI

Web UI:
```text
http://ci.syncloud.org:8080/syncloud/store
```

Drone API examples:
```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/store/builds?limit=5"
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/store/builds/{N}"
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/store/builds/{N}/logs/{stage}/{step}"
```

## Debugging CI failures

Stages (amd64, arm64, arm) run in parallel. A build can be `status: running` while individual stages have already failed — always drill into per-stage, per-step status instead of waiting on the top-level build status.

```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/store/builds/{N}" | python3 -c "
import json,sys
b=json.load(sys.stdin)
for stage in b.get('stages', []):
    print(stage.get('name'), '-', stage.get('status'))
    for step in stage.get('steps', []):
        if step.get('status') == 'failure':
            print('   ', step.get('number'), step.get('name'), '-', step.get('status'))
"
```

Get the log for a specific stage/step:

```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/store/builds/{N}/logs/{stage}/{step}" | python3 -c "
import json,sys
for line in json.load(sys.stdin):
    print(line.get('out', ''), end='')
" | tail -120
```

Drone's REST logs endpoint returns 404 while a step is still running — logs are only persisted to its DB on step completion. To watch a running step in real time, ssh to the CI host and `docker logs -f drone-<id>` instead.

## CI artifacts

Artifacts are uploaded by the `artifact` step and served by an nginx file browser at `http://ci.syncloud.org:8081`.

```sh
curl -s "http://ci.syncloud.org:8081/files/syncloud-store/{N}-{arch}/"
```

Playwright failures (once the web pipeline lands) show up under `playwright-report/` and `playwright-results/<slug>/` (`error-context.md`, `failure-full-page.png`, `trace.zip`, `video.webm`).
