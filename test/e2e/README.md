# End-to-End Black-Box Test Suite

A self-contained, reproducible regression suite that stands up a throwaway
Postgres, boots the **real** server binary against it, and exercises every
endpoint and every use case (happy path, validation, errors, security,
concurrency, performance) over plain HTTP — the same way a browser or crawler
hits production. It complements the Go unit/integration tests by catching
things that only appear in the full middleware + routing + DB stack (the
`/sitemap.xml` extension-stripping 404, the rate-limit bypass, the ELO race).

Built from the 2026-07-09 audit (`docs/AUDIT_2026-07-09.md`) and the endpoint
matrix (`docs/API_TEST_MATRIX.md`).

## Run it

```bash
test/e2e/run.sh                 # full run: spins up podman Postgres + server, tears down after
KEEP=1 test/e2e/run.sh          # leave server + DB up afterward (poke at http://localhost:8477)
ONLY=security test/e2e/run.sh   # run a single module (security|errors|smoke|seo|concurrency|perf)
TORRO_E2E_REUSE_DB_URL=postgres://user:pass@host:5432/db test/e2e/run.sh   # use an existing DB
```

Requirements: `podman` (or set `TORRO_E2E_REUSE_DB_URL`), `curl`, `python3`, `go`.
No host Postgres client needed — SQL runs via `podman exec`. Uses ports 55440
(DB) and 8477 (app) by default; override with `PG_PORT` / `APP_PORT`.

Exit code is non-zero iff a **hard** assertion fails. Known-but-unfixed defects
are `xfail` guards (see below), which keep the run green while the bug exists.

## How the known-bug guards work

The suite encodes **current reality first** so it detects regressions in both
directions:

- `assert_status` / `assert_body_contains` / `assert_header` — hard assertions
  for behavior that is correct today. A failure fails the run.
- `xfail_status ID desc BUGGY CORRECT ...` — a **confirmed, unfixed defect**. It
  asserts the *buggy* status, so the run stays green, but:
  - prints a yellow `xfail` line naming the bug, and
  - if the endpoint ever returns the `CORRECT` status instead, prints a cyan
    **`XPASS`** ("bug fixed? tighten this guard") — your cue to flip it to a hard
    `assert_status`. A brand-new third status is a hard failure.

So when you fix, say, the `/classes/{id}/vote` 500→404 bug, the suite will
`XPASS` on `S1` and remind you to convert the guard. Nothing silently drifts.

## What's covered

| Module | Focus | Notable guards |
|--------|-------|----------------|
| `test_smoke.sh` | every route's happy path + content-type/cookie sanity | all 43 routes answer; `/embed/` sets no cookie |
| `test_errors.sh` | validation, status-code correctness | **S1/S2** bad-id→500, **S7** negative offset→500, **S6** unknown-class contract |
| `test_security.sh` | rate limiting, admin auth, injection safety | **SEC07** XFF rate-limit bypass, admin fail-closed, injection inert |
| `test_seo.sh` | SEO surface | **SEO01–03** sitemap/robots/llms 404 (URLFormat strip); assets unaffected |
| `test_concurrency.sh` | the ELO lost-update race | **RACE1** fires N concurrent votes, compares vs. serialized-correct ELO |
| `test_perf.sh` | per-request latency budgets | `/premsa` hotspot + a small concurrent burst (non-200 = pool/timeout exhaustion) |

## Notes

- **Perf module** uses generous, order-of-magnitude budgets and runs against the
  suite's freshly-seeded (small) DB, so it catches gross regressions but not the
  under-load collapse. To reproduce the real `/premsa` degradation, seed volume
  (see the methodology in `docs/AUDIT_2026-07-09.md`) and re-run with
  `BUDGET_PREMSA_MS` tightened.
- **Concurrency module** mutates `Torrons.Rating` / `Results` for the fixture
  pairing by design — the suite runs against a throwaway DB, so this is safe and
  repeatable.
- Fixtures (`e2e-user-unlocked` with 60 votes, `e2e-user-fresh` with 0, a known
  pairing + its two torró ids) are seeded deterministically by `run.sh` and
  referenced by stable handle across modules.
- Teardown kills the **real** bound server PID (looked up via `ss`, not `$!`) and
  removes the container. `KEEP=1` skips teardown for debugging.
