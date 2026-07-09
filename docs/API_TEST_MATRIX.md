# Torrorèndum — Black-Box HTTP API Test Matrix

Authoritative, code-derived test matrix for every HTTP endpoint. Every expected
status code below was read out of the handler + error-mapping source, not
guessed. Where the current behavior looks wrong, the **actual** behavior is
recorded as the expectation (so the suite locks in reality and catches
regressions) and flagged inline with `<!-- SUSPECT: ... -->`; a consolidated
list is at the end.

Route table source: `internal/http/server.go` (lines 151–282).
Error→status mapping: `internal/domain/error.go`, `internal/repository/error.go`.
Middleware: `internal/http/middleware.go`, `internal/http/server.go:46–142`.

---

## 0. Conventions, fixtures, and global behaviors

### 0.1 Middleware chain (applied to EVERY request, in order)

From `server.go:46–142`:

1. `middleware.Logger`, `Recoverer`, `RequestID`, `RealIP`, `URLFormat`,
   `render.SetContentType(JSON)`.
2. **Global rate limit** — `httprate.Limit(100, 1*time.Minute, KeyByIP)`
   (`server.go:54–62`). Over 100 req/min per IP → **429**, body
   `Rate limit exceeded. Please try again later.\n`, Content-Type
   `text/plain; charset=utf-8`. Sets `X-RateLimit-*` / `Retry-After` headers.
3. **Security headers** (`server.go:89–139`) — every response carries
   `X-Content-Type-Options: nosniff`, `X-XSS-Protection`, `Referrer-Policy`,
   `Permissions-Policy`, `X-Permitted-Cross-Domain-Policies: none`. Non-`/embed/`
   paths also get `X-Frame-Options: DENY` and the app CSP; `/embed/` paths get
   NO `X-Frame-Options` and a permissive CSP with `frame-ancestors *`.
4. **UserMiddleware** (`middleware.go:26–102`) — for any path NOT starting with
   `/embed/`: reads cookie `torrons_user_id`; if valid+known → uses it; if
   missing OR present-but-unknown → **mints a brand-new Users row and sets the
   cookie**. Always injects a valid, existing user id into context. Paths under
   `/embed/` skip this entirely (no cookie read, no user minted).
5. Route-scoped: **voteRateLimiter** `httprate.Limit(20, 1*time.Minute)` keyed by
   context user id (`server.go:76–86`), applied only to
   `POST /pairings/{id}/vote` and `POST /bracket/match/{matchId}/vote`. Over
   20/min per user → **429**, body `You're voting too quickly. Please slow down.\n`.
6. Route-scoped: **RequireAdminToken** (`middleware.go:125–147`) on
   `POST /bracket/{classId}/create` and `POST /bracket/{bracketId}/advance`.

Because UserMiddleware runs for all non-embed paths, a **cookie that is
absent** and a **cookie whose UUID is unknown** produce the SAME observable
outcome at the handler: a fresh valid user (VoteCount 0) is created and a new
`Set-Cookie: torrons_user_id=...` is returned. The only way to present the
handler with a KNOWN, high-vote user is to seed that Users row first (see 0.5).

### 0.2 Router defaults

No custom `NotFound`/`MethodNotAllowed` handlers are registered, so chi v5.3.0
defaults apply:

- Unmatched path → **404** (empty body, `text/plain`).
- Path matches but method not registered → **405** (chi sets an `Allow`
  header). All "wrong method" cases below expect **405**.
- `middleware.URLFormat` strips a trailing `.<ext>` from the routing path. This
  is why `/share/card.png`, `/wrapped/card.png`, `/reveal/card.png`,
  `/press-kit/card.png` are registered WITHOUT `.png` yet are hit WITH it. IDs
  in this app are UUIDs (no dots) so URLFormat never truncates a path param.

### 0.3 ID typing — critical for "bad id" cases

Every `"Id"` column is **`VARCHAR(36)`, NOT native `uuid`** (see
`migrations/0000*_create_*.up.sql`). Consequence: a malformed id, a non-UUID
id, and a SQL-metacharacter id are all just non-matching parameterized string
values — Postgres raises **no type error**. They therefore behave **identically
to a well-formed-but-nonexistent UUID**: the row lookup returns
`sql.ErrNoRows` → `repository.handleErrors` maps "no rows in result set" →
`errNotFound()` → `domain.NotFoundError`. There is no SQL-injection surface
(all queries are parameterized). Class ids are the literal strings `"1"`–`"5"`.

### 0.4 Error → HTTP status mapping (`domain/error.go`)

`render.Render(w, r, domain.ErrXxx(err))` writes JSON
`{"code":<ErrorCode>,"message":"<text after first colon>"}` with Content-Type
`application/json; charset=utf-8` and status:

| Constructor | HTTP status |
|---|---|
| `ErrBadRequest` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrNotFound` | 404 |
| `ErrInternal` | 500 |

The embedded `code` is looked up from the error MESSAGE prefix, independent of
the HTTP status. So a NotFound-classified repo error passed into `ErrInternal`
yields **HTTP 500 with body `{"code":2506,"message":"Record not found"}`** — the
status and code disagree. This underlies SUSPECT-1/2. Error codes:
`ValidationError=2400`, `Unauthorized=2401`, `NotFound=2506`, `Unknown=2507`,
`DuplicateKey=2503`, `ForeignKey=2504`.

Handlers that use `http.Error(...)` instead return `text/plain; charset=utf-8`
with a plain string body.

### 0.5 Test fixtures (referenced by every case)

Seed these directly via SQL before the suite (mirrors `integration_test.go`'s
`insertTestClass` / `insertTestTorro` helpers; classes/torrons/users have no
repo `Create` used by these paths). Classes `"1"`–`"5"` exist from
`migrations/000005` + `000012`.

| Token | Meaning |
|---|---|
| `USER_0` | `torrons_user_id` cookie for a seeded Users row, `VoteCount=0`, `ClassVotes='{}'` |
| `USER_50` | seeded Users row, `VoteCount=50`, `ClassVotes='{"1":50,"2":50,"3":50,"4":50}'` (above every threshold) |
| `USER_49` | seeded Users row, `VoteCount=49` (one below the 50 global gate) |
| `USER_UNKNOWN` | a well-formed UUID with NO Users row → middleware mints a fresh user |
| `NX_UUID` | `00000000-0000-0000-0000-000000000000` (well-formed, never present) |
| `BAD_ID` | `not-a-uuid` |
| `SQL_ID` | url-encoded `%27%20OR%20%271%27%3D%271` (`' OR '1'='1`) |
| `PAIRING_1` | a real Pairings row in class `"1"` with torrons `TORRO_A`,`TORRO_B` |
| `TORRO_A` / `TORRO_B` | torrons in `PAIRING_1` |
| `TORRO_X` | any real torró id (for `/torro/{id}`) |
| `BRACKET_IP` | a real Brackets row, `Status=in_progress`, class `"1"` |
| `MATCH_PENDING` | a pending BracketMatches row in `BRACKET_IP` current round, torrons `M1`,`M2` |
| `CIRCLE_MEMBER` | FriendCircles row `USER_50` is a member of |
| `CIRCLE_NONMEMBER` | FriendCircles row `USER_50` is NOT a member of |
| `INVITE_OK` | valid invite code for some circle |
| `ADMIN_TOKEN` | value the server was started with via `ADMIN_TOKEN` env (empty by default) |

**Creating a user with N votes:** insert a Users row with `"VoteCount"=N` and,
for per-class gates, a `"ClassVotes"` JSONB like `{"1":N}` (read by
`GetVoteCountForClass`, `postgres_user.go:130`; total `VoteCount` read by
`Get`). Then send `Cookie: torrons_user_id=<that id>`. Alternatively cast N real
votes via `POST /pairings/{id}/vote`, but direct SQL is the reproducible path.

### 0.6 Vote-threshold constants (asked for explicitly)

`getMinVotesForClass(classId)` — `user_api.go:170–185`:
`"1"→30`, `"2"→25`, `"3"→30`, `"4"→40`, `"5"→50`, **default→25**.

- **`/wrapped`** gate = `getMinVotesForClass("5")` = **50** (`wrapped_handler.go:126`).
- **`/reveal`** gate = `getMinVotesForClass("5")` = **50** (`reveal_handler.go:172`).
- **`/api/user/leaderboard/global`** gate = hardcoded **50** (`user_api.go:160–161`).
- **`GET /leaderboard?view=personal&category=global`** gate =
  `getMinVotesForClass("global")` → default → **25** (`leaderboard.go:166`).
  ← inconsistent with the 50 used everywhere else (**SUSPECT-3**).
- **`GET /api/leaderboard/global`** (community JSON) — **no gate**, always top 100.
- `/stats` uses `getMinVotesForClass(class.Id)` per class with `"5"`→50.

### 0.7 Which endpoints need seed state

- **Active `Campaigns` row required (else empty/boundary state, not error):**
  none hard-require it. `POST /bracket/{classId}/create` DOES require an active
  campaign (400 without one). `/advent` shows a "no campaign" state without one.
  `/api/campaign/countdown|info|countdown/widget` return their not-active
  bodies without one. `POST /pairings/{id}/vote` soft-attaches the campaign
  (succeeds either way).
- **A bracket must exist for:** `POST /bracket/match/{matchId}/vote`,
  `POST /bracket/{bracketId}/advance`, and meaningful (non-empty) output from
  `GET /bracket/{classId}`, `GET /bracket/{classId}/vote`.
- **`/reveal`, `/wrapped`** need a user with ≥50 votes to show real data (else
  the "not unlocked" state, still 200).

---

## WEB UI ROUTES

## R1 — `GET /healthcheck` → `handleHealthcheck` (`healthcheck.go:19`)

No auth. No params.

| id | request | expect |
|---|---|---|
| R1-01 | `GET /healthcheck` | **200**, Content-Type `application/json; charset=utf-8`, body contains `"answer":42` |
| R1-02 | `POST /healthcheck` | **405** |

## R2 — `GET /robots.txt` → `robotsTxt` (`seo_handler.go:25`)

| id | request | expect |
|---|---|---|
| R2-01 | `GET /robots.txt` | **200**, Content-Type `text/plain; charset=utf-8`, body contains `Sitemap: https://torro.cat/sitemap.xml` and `User-agent: *` |
| R2-02 | `POST /robots.txt` | **405** |

## R3 — `GET /sitemap.xml` → `sitemapXML` (`seo_handler.go:70`)

Reads `torroRepo.List`, `classRepo.List`, `bracketRepo.GetLatestByClass`.

| id | request | expect |
|---|---|---|
| R3-01 | `GET /sitemap.xml` | **200**, Content-Type `application/xml; charset=utf-8`, body starts `<?xml`, contains `<loc>https://torro.cat/</loc>` and one `<loc>https://torro.cat/torro/...` per torró |
| R3-02 | `POST /sitemap.xml` | **405** |

## R4 — `GET /llms.txt` → `llmsTxt` (`seo_handler.go:36`)

| id | request | expect |
|---|---|---|
| R4-01 | `GET /llms.txt` | **200**, `text/plain; charset=utf-8`, body contains `# Torrorèndum` and `independent fan project` |
| R4-02 | `PUT /llms.txt` | **405** |

## R5 — `GET /` → `index` (`handler.go:107`)

Static template, no DB. Reads `HX-Request` header only.

| id | request | expect |
|---|---|---|
| R5-01 | `GET /` (no cookie) | **200**, `text/html; charset=utf-8`, body contains `<!DOCTYPE html>`; response sets `Set-Cookie: torrons_user_id=...` (middleware minted a user) |
| R5-02 | `GET /` with `HX-Request: true` | **200**, `text/html`; renders the `index.html` fragment (still 200; assert body present). |
| R5-03 | `GET /` with valid `Cookie: torrons_user_id=USER_50` | **200**; a fresh `Set-Cookie` refreshing expiry is still returned |
| R5-04 | `POST /` | **405** |

## R6 — `GET /classes` → `classes` (`handler.go:128`)

Reads `classRepo.List`. `HX-Request` toggles fragment vs full page.

| id | request | expect |
|---|---|---|
| R6-01 | `GET /classes` | **200**, `text/html`, body lists categories |
| R6-02 | `GET /classes` `HX-Request: true` | **200**, `text/html` fragment |
| R6-03 | `POST /classes` | **405** |

## R7 — `GET /classes/{id}/vote` → `vote` (`handler.go:157`)

Path `{id}` = classId. Reads streak from context user (decorative). Calls
`pairingRepo.GetRandom(classId)` then `torroRepo.Get` twice.

| id | request | expect |
|---|---|---|
| R7-01 | `GET /classes/1/vote` (class 1 has pairings) | **200**, `text/html`, renders `vote.html` with two torró options |
| R7-02 | `GET /classes/1/vote` `HX-Request: true` | **200**, `text/html` fragment (no full-page chrome) |
| R7-03 | `GET /classes/1/vote` with `Cookie: torrons_user_id=USER_50` | **200**; streak pill may render but never blocks |
| R7-04 | `GET /classes/99/vote` (well-formed but nonexistent class, no pairings) | **500**, `application/json`, body `{"code":2506,"message":"Record not found"}` <!-- SUSPECT-1: nonexistent/empty class → GetRandom returns NotFound but handler wraps it in ErrInternal (handler.go:162-167); should be 404 --> |
| R7-05 | `GET /classes/BAD_ID/vote` | **500** (same as R7-04; varchar id, no pairings → ErrNoRows) <!-- SUSPECT-1 --> |
| R7-06 | `GET /classes/SQL_ID/vote` | **500** (parameterized, no injection; behaves as nonexistent) <!-- SUSPECT-1 --> |
| R7-07 | missing path segment: `GET /classes//vote` | **404** (chi doesn't match empty path param) |
| R7-08 | `POST /classes/1/vote` | **405** |

## R8 — `POST /pairings/{id}/vote` → `result` (`handler.go:229`), voteRateLimiter

Path `{id}` = pairingId. Query `id` = winner torró id (REQUIRED). Query
`advent=true` optional. Context user id from middleware. On success renders next
`pairing.html` fragment (or advent "come back tomorrow" if `advent=true`).

| id | request | expect |
|---|---|---|
| R8-01 | `POST /pairings/PAIRING_1/vote?id=TORRO_A` cookie `USER_50` | **200**, `text/html`, renders next `pairing.html`; a Results row is written, ELO updated |
| R8-02 | `POST /pairings/PAIRING_1/vote?id=TORRO_A` `HX-Request: true` | **200**, `text/html` fragment |
| R8-03 | `POST /pairings/PAIRING_1/vote` (missing `id` query) | **400**, `application/json`, `{"code":2400,"message":"Winner ID does not match pairing torros"}` (empty winner ≠ either torró) |
| R8-04 | `POST /pairings/PAIRING_1/vote?id=NX_UUID` (winner not in pairing) | **400**, `{"code":2400,"message":"Winner ID does not match pairing torros"}` |
| R8-05 | `POST /pairings/NX_UUID/vote?id=TORRO_A` (nonexistent pairing) | **500**, `application/json`, `{"code":2506,"message":"Record not found"}` <!-- SUSPECT-2: pairingRepo.Get NotFound wrapped in ErrInternal (handler.go:250-255); should be 404 --> |
| R8-06 | `POST /pairings/BAD_ID/vote?id=TORRO_A` | **500** (varchar id → ErrNoRows) <!-- SUSPECT-2 --> |
| R8-07 | `POST /pairings/SQL_ID/vote?id=TORRO_A` | **500** (parameterized; behaves as nonexistent) <!-- SUSPECT-2 --> |
| R8-08 | `POST /pairings/PAIRING_1/vote?id=TORRO_A&advent=true` cookie `USER_50` | **200**, `text/html` advent template ("come back tomorrow"); an AdventVotes row is written (once/day) |
| R8-09 | second `advent=true` vote same user same day | **500** `{"code":...,"message":...}` — the duplicate AdventVotes unique-constraint violation flows through `ErrInternal` (handler.go:415) <!-- SUSPECT-2b: a duplicate-advent-today is a 4xx condition (DuplicateKey) but surfaces as 500 --> |
| R8-10 | `GET /pairings/PAIRING_1/vote?id=TORRO_A` | **405** |
| R8-11 | rate-limit trip: 21st `POST /pairings/PAIRING_1/vote?id=TORRO_A` within 60s as the SAME context user | **429**, `text/plain`, body `You're voting too quickly. Please slow down.\n` |

## R9 — `GET /torro/{id}` → `torroDetail` (`torro_handler.go:76`)

Path `{id}` = torró id. Uses `http.Error` (plain text), not domain errors.

| id | request | expect |
|---|---|---|
| R9-01 | `GET /torro/TORRO_X` | **200**, `text/html`, product detail with name/rating |
| R9-02 | `GET /torro/TORRO_X` `HX-Request: true` | **200**, `text/html` fragment |
| R9-03 | `GET /torro/NX_UUID` | **404**, `text/plain; charset=utf-8`, body `Torró no trobat` (torro_handler.go:83-86 — correctly maps NotFound to 404) |
| R9-04 | `GET /torro/BAD_ID` | **404**, body `Torró no trobat` (varchar → ErrNoRows → NotFound) |
| R9-05 | `GET /torro/SQL_ID` | **404**, body `Torró no trobat` |
| R9-06 | `POST /torro/TORRO_X` | **405** |

## R10 — `GET /leaderboard` → `leaderboard` (`leaderboard_handler.go:62`)

Query `view` (`personal`|`global`, default `personal`), `category`
(`global`|classId, default `global`), dietary filters
`vegan|gluten_free|lactose_free|organic` (`=true`). Requires a context user
(always present via middleware). Uses `http.Error` for failures.

| id | request | expect |
|---|---|---|
| R10-01 | `GET /leaderboard` (defaults: personal/global) cookie `USER_50` | **200**, `text/html`; ≥50 votes ⇒ shows personalized rows |
| R10-02 | `GET /leaderboard` cookie `USER_49` | **200**, `text/html`; body contains the "not enough votes" copy `No tens prou vots` (49 < 25? no — 49 ≥ 25 so it UNLOCKS) → actually shows rows <!-- SUSPECT-3: personal global gate is 25 (getMinVotesForClass("global")→default), not 50; USER_49 is unlocked here but locked on /wrapped --> |
| R10-03 | `GET /leaderboard` cookie `USER_0` | **200**; 0 < 25 ⇒ body contains `No tens prou vots` |
| R10-04 | `GET /leaderboard?view=global` | **200**; community view, no vote gate |
| R10-05 | `GET /leaderboard?view=global&category=1` | **200**; class-1 community ranking |
| R10-06 | `GET /leaderboard?view=personal&category=1` cookie `USER_50` | **200**; class-1 gate=30, 50≥30 ⇒ unlocked |
| R10-07 | `GET /leaderboard?category=99` (nonexistent class) | **200**; class name resolves to `Desconegut`, empty entries, no 404 <!-- SUSPECT-6: class id never validated here --> |
| R10-08 | `GET /leaderboard?view=bogus` | **200**; any non-`personal` value falls into the `global` branch (leaderboard.go:99/107) |
| R10-09 | `GET /leaderboard?view=global&vegan=true&gluten_free=true` | **200**; dietary filter applied |
| R10-10 | `GET /leaderboard` `HX-Request: true` | **200**, `text/html` fragment |
| R10-11 | `POST /leaderboard` | **405** |

Note: the `HTTP 401 "User not found"` branch (leaderboard_handler.go:77-81) is
unreachable through the real router because UserMiddleware always injects a user.

## R11 — `GET /stats` → `stats` (`stats_handler.go:79`)

Requires context user (always present). Reads user + per-class vote counts.

| id | request | expect |
|---|---|---|
| R11-01 | `GET /stats` cookie `USER_50` | **200**, `text/html`; shows total votes, category progress, achievements |
| R11-02 | `GET /stats` cookie `USER_0` | **200**; all categories locked, rank `Principiant` |
| R11-03 | `GET /stats` no cookie | **200**; middleware mints a 0-vote user, renders locked state |
| R11-04 | `GET /stats` `HX-Request: true` | **200**, `text/html` fragment |
| R11-05 | `POST /stats` | **405** |

## R12 — `GET /history` → `history` (`history_handler.go:45`)

Query `category` (`all`|classId, default `all`), `offset` (int, default 0).
Requires context user. Raw SQL over `Results`.

| id | request | expect |
|---|---|---|
| R12-01 | `GET /history` cookie `USER_50` | **200**, `text/html`, vote history list |
| R12-02 | `GET /history?category=1` cookie `USER_50` | **200**; filtered to class 1 |
| R12-03 | `GET /history?offset=20` | **200**; paginated page 2 |
| R12-04 | `GET /history?offset=abc` (non-int) | **200**; `strconv.Atoi` error ignored → offset stays 0 (history_handler.go:63-65) |
| R12-05 | `GET /history?offset=-5` (negative) | **200**; passed straight into SQL `OFFSET -5` → Postgres rejects negative OFFSET → query error → **500**, `text/plain`, `Internal Server Error` <!-- SUSPECT-7: negative offset reaches SQL OFFSET and 500s; not clamped --> |
| R12-06 | `GET /history?category=99` | **200**; empty list (no rows match class 99) |
| R12-07 | `GET /history` `HX-Request: true` | **200**, `text/html` fragment |
| R12-08 | `POST /history` | **405** |

## R13 — `GET /share/card` (hit as `/share/card.png`) → `shareCard` (`sharecard_handler.go:24`)

Per-user PNG. Requires context user.

| id | request | expect |
|---|---|---|
| R13-01 | `GET /share/card.png` cookie `USER_50` | **200**, Content-Type `image/png`, `Cache-Control: private, no-store`, body is PNG bytes (`\x89PNG`) |
| R13-02 | `GET /share/card.png` cookie `USER_0` | **200**, `image/png` — fallback "vota per generar" card (no votes) |
| R13-03 | `GET /share/card` (no `.png`) | **200**, `image/png` (URLFormat makes both resolve to the same route) |
| R13-04 | `POST /share/card.png` | **405** |

## R14 — `GET /bracket/{classId}` → `bracketOverview` (`bracket_handler.go:89`)

Path `{classId}`. No bracket for the class = empty state (200), not an error.

| id | request | expect |
|---|---|---|
| R14-01 | `GET /bracket/1` (BRACKET_IP exists in class 1) | **200**, `text/html`; rounds + matches rendered |
| R14-02 | `GET /bracket/2` (no bracket) | **200**, `text/html`; empty/`BracketExists=false` state |
| R14-03 | `GET /bracket/99` (nonexistent class) | **200**; class name falls back to the raw id, empty state |
| R14-04 | `GET /bracket/BAD_ID` | **200**; GetLatestByClass errors → treated as empty state (bracket_handler.go:111-116) |
| R14-05 | `GET /bracket/1` `HX-Request: true` | **200**, `text/html` fragment |
| R14-06 | `POST /bracket/1` | **405** |

## R15 — `GET /bracket/{classId}/vote` → `bracketVote` (`bracket_handler.go:197`)

Path `{classId}`. Requires context user (always present). Serves a random open
match, or completed/no-more-matches state.

| id | request | expect |
|---|---|---|
| R15-01 | `GET /bracket/1/vote` cookie `USER_50` (open matches) | **200**, `text/html`; a single match card |
| R15-02 | `GET /bracket/2/vote` (no bracket) | **200**; `BracketExists=false` empty state |
| R15-03 | `GET /bracket/1/vote` when all current-round matches already voted by this user | **200**; `NoMoreMatches` state |
| R15-04 | `GET /bracket/1/vote` when bracket completed | **200**; champion/completed state |
| R15-05 | `GET /bracket/1/vote` `HX-Request: true` | **200**, `text/html` fragment |
| R15-06 | `POST /bracket/1/vote` | **405** |

## R16 — `POST /bracket/match/{matchId}/vote` → `bracketMatchVote` (`bracket_handler.go:298`), voteRateLimiter

Path `{matchId}`. Query `winner` = torró id (REQUIRED). Context user id.
Errors routed via `renderBracketError` (NotFound→404, Validation/Dup/FK→400,
else→500).

| id | request | expect |
|---|---|---|
| R16-01 | `POST /bracket/match/MATCH_PENDING/vote?winner=M1` cookie `USER_50` | **200**, `text/html`; vote recorded, next vote card served |
| R16-02 | `POST /bracket/match/MATCH_PENDING/vote` (missing `winner`) | **400**, `application/json`, `{"code":2400,"message":"missing winner query parameter"}` |
| R16-03 | `POST /bracket/match/MATCH_PENDING/vote?winner=NX_UUID` (winner not in match) | **400**, `{"code":2400,"message":"winner does not match either competitor in this match"}` |
| R16-04 | `POST /bracket/match/NX_UUID/vote?winner=M1` (nonexistent match) | **404**, `application/json`, `{"code":2506,"message":"Record not found"}` (renderBracketError maps NotFound→404) |
| R16-05 | `POST /bracket/match/BAD_ID/vote?winner=M1` | **404** (varchar → ErrNoRows → NotFound → 404) |
| R16-06 | `POST /bracket/match/SQL_ID/vote?winner=M1` | **404** |
| R16-07 | duplicate vote: same user votes MATCH_PENDING twice | **400**, `{"code":2400,"message":"you have already voted on this match"}` (unique-constraint caught, bracket_handler.go:358-361) |
| R16-08 | vote on an already-decided match | **400**, `{"code":2400,"message":"this match is no longer open for voting"}` |
| R16-09 | `GET /bracket/match/MATCH_PENDING/vote?winner=M1` | **405** |
| R16-10 | 21st vote in 60s as same user | **429**, `text/plain`, `You're voting too quickly. Please slow down.\n` |

## R17 — `POST /bracket/{classId}/create` → `bracketCreate` (`bracket_handler.go:578`), RequireAdminToken

Path `{classId}`. Query `size` (int, default `DefaultBracketSize`, must be a
power of two). Requires `Authorization: Bearer <ADMIN_TOKEN>`.

| id | request | expect |
|---|---|---|
| R17-01 | `POST /bracket/1/create?size=4` header `Authorization: Bearer ADMIN_TOKEN`, active campaign, class 1 ≥2 torrons | **201**, `application/json`, body is the bracket JSON (`"ClassId":"1"`,`"Size":4`,`"Status":"in_progress"`) |
| R17-02 | same but NO `Authorization` header | **401**, `application/json`, `{"code":2401,"message":"invalid or missing admin token"}`, header `WWW-Authenticate: Bearer` |
| R17-03 | same but `Authorization: Bearer wrong-token` | **401**, `{"code":2401,...}`, `WWW-Authenticate: Bearer` |
| R17-04 | correct token but server started with EMPTY `ADMIN_TOKEN` | **401** (fails closed — middleware.go:131-135) |
| R17-05 | correct token, `size=5` (not power of two) | **400**, `{"code":2400,"message":"bracket size must be a power of two (got 5)"}` |
| R17-06 | correct token, `size=abc` (non-int) | **400**, `{"code":2400,"message":"size must be an integer"}` |
| R17-07 | correct token, `size=0` | **400**, `bracket size must be a power of two (got 0)` (isPowerOfTwo(0)=false) |
| R17-08 | correct token, `size=-4` | **400**, `bracket size must be a power of two (got -4)` |
| R17-09 | correct token, `size=1` | **400**, `class ... needs at least 2 active torrons` — 1 passes isPowerOfTwo but TopNByClass LIMIT 1 yields <2 <!-- SUSPECT-8: size=1 accepted by power-of-two check, rejected downstream with a misleading message; DB CHECK is Size>=2 --> |
| R17-10 | correct token, NO active campaign | **400**, `{"code":2400,"message":"no active campaign to attach a bracket to"}` |
| R17-11 | correct token, bracket already exists for (active campaign, class 1) | **400**, `a bracket already exists for class 1 in the active campaign` |
| R17-12 | correct token, class `99` (or a class with <2 active torrons) | **400**, `class 99 needs at least 2 active torrons to start a bracket` |
| R17-13 | `GET /bracket/1/create` header valid | **405** (method checked after — see note) |

Note: with a real router the middleware runs before method dispatch only if the
path+method matches; `GET /bracket/1/create` matches no GET route → chi returns
**405/404** at the router. Assert 405 (a POST route exists at that path).

## R18 — `POST /bracket/{bracketId}/advance` → `bracketAdvance` (`bracket_handler.go:737`), RequireAdminToken

Path `{bracketId}`. Requires admin bearer token.

| id | request | expect |
|---|---|---|
| R18-01 | `POST /bracket/BRACKET_IP/advance` header `Authorization: Bearer ADMIN_TOKEN` | **200**, `application/json`, updated bracket JSON (advanced round or completed) |
| R18-02 | no `Authorization` | **401**, `{"code":2401,"message":"invalid or missing admin token"}`, `WWW-Authenticate: Bearer` |
| R18-03 | `Authorization: Bearer wrong-token` | **401** |
| R18-04 | valid token, `POST /bracket/NX_UUID/advance` (nonexistent) | **404**, `{"code":2506,"message":"Record not found"}` (renderBracketError) |
| R18-05 | valid token, `POST /bracket/BAD_ID/advance` | **404** |
| R18-06 | valid token, bracket already completed | **400**, `{"code":2400,"message":"bracket ... is already completed"}` |
| R18-07 | `GET /bracket/BRACKET_IP/advance` valid token | **405** |

Collision note: `/bracket/{classId}/create` and `/bracket/{bracketId}/advance`
are distinct literal suffixes; `/bracket/{classId}` (R14) is GET-only so no path
ambiguity with the POST routes.

## R19 — `GET /advent` → `advent` (`advent_handler.go:31`)

Context user optional-ish (used for "already voted"). Needs active campaign for
the real duel.

| id | request | expect |
|---|---|---|
| R19-01 | `GET /advent` with active campaign, user hasn't voted today | **200**, `text/html`; today's featured pairing |
| R19-02 | `GET /advent` with NO active campaign | **200**, `text/html`; `CampaignActive=false` boundary state |
| R19-03 | `GET /advent` active campaign, user already voted today (cookie `USER_50` with an AdventVotes row for today) | **200**; `AlreadyVoted=true` "come back tomorrow" state |
| R19-04 | `GET /advent` active campaign but featured class has no pairings | **500**, `text/plain`, `Internal Server Error` (GetDeterministic → ErrNoRows → advent_handler.go:74-78) |
| R19-05 | `GET /advent` `HX-Request: true` | **200**, `text/html` fragment |
| R19-06 | `POST /advent` | **405** |

## R20 — `GET /friends` → `friendsIndex` (`friends_handler.go:38`)

Context user (always present). Uses `http.Error`.

| id | request | expect |
|---|---|---|
| R20-01 | `GET /friends` cookie `USER_50` | **200**, `text/html`; lists the user's circles + "create" action |
| R20-02 | `GET /friends` no cookie | **200**; middleware mints user, empty circle list |
| R20-03 | `GET /friends` `HX-Request: true` | **200**, `text/html` fragment |
| R20-04 | `POST /friends` | **405** |

## R21 — `POST /friends/create` → `friendsCreate` (`friends_handler.go:64`)

Context user. Creates a circle, returns the invite link view.

| id | request | expect |
|---|---|---|
| R21-01 | `POST /friends/create` cookie `USER_50` | **200**, `text/html`; body contains an invite URL `.../friends/join/<code>` |
| R21-02 | `POST /friends/create` no cookie | **200**; middleware mints a user first, then creates the circle owned by it |
| R21-03 | `POST /friends/create` `HX-Request: true` | **200**, `text/html` fragment |
| R21-04 | `GET /friends/create` | **405** |

## R22 — `GET /friends/join/{inviteCode}` → `friendsJoin` (`friends_handler.go:96`)

Path `{inviteCode}`. Adds user to circle, then 302 redirects to the circle.

| id | request | expect |
|---|---|---|
| R22-01 | `GET /friends/join/INVITE_OK` cookie `USER_50` | **302**, header `Location: /friends/<circleId>` |
| R22-02 | `GET /friends/join/INVITE_OK` no cookie | **302** (middleware mints a user, adds it, redirects) |
| R22-03 | `GET /friends/join/nonexistent-code` | **200**, `text/html`; `invalid-invite` view (NOT an error — friends_handler.go:108-116) |
| R22-04 | `GET /friends/join/SQL_ID` | **200**; `invalid-invite` view |
| R22-05 | `GET /friends/join/` (empty code) | **404** (chi won't match empty param) |
| R22-06 | `POST /friends/join/INVITE_OK` | **405** |

## R23 — `GET /friends/{circleId}` → `friendsLeaderboard` (`friends_handler.go:130`)

Path `{circleId}`. Query `category` (`global`|classId, default `global`).
Context user. Non-members get a "not-member" view; unknown circle → 404.

| id | request | expect |
|---|---|---|
| R23-01 | `GET /friends/CIRCLE_MEMBER` cookie `USER_50` (member) | **200**, `text/html`; circle leaderboard |
| R23-02 | `GET /friends/CIRCLE_MEMBER?category=1` member | **200**; class-1 circle leaderboard |
| R23-03 | `GET /friends/CIRCLE_NONMEMBER` cookie `USER_50` (not a member) | **200**, `text/html`; `not-member` view (friends_handler.go:159-166) |
| R23-04 | `GET /friends/NX_UUID` (unknown circle) | **404**, `text/plain; charset=utf-8`, body `Not Found` (friends_handler.go:142-147) |
| R23-05 | `GET /friends/BAD_ID` | **404**, body `Not Found` |
| R23-06 | `GET /friends/SQL_ID` | **404**, body `Not Found` |
| R23-07 | `GET /friends/CIRCLE_MEMBER` `HX-Request: true` | **200**, `text/html` fragment |
| R23-08 | `POST /friends/CIRCLE_MEMBER` | **405** |

## R24 — `GET /embed/leaderboard` → `embedLeaderboard` (`embed_handler.go:30`)

**Cookie-less path** — UserMiddleware SKIPS it (no `Set-Cookie`, no user
minted). Query `classId` (default `"5"`→all-classes), `limit` (int, default 10,
capped 25). Public cache.

| id | request | expect |
|---|---|---|
| R24-01 | `GET /embed/leaderboard` (no cookie) | **200**, `text/html`; NO `Set-Cookie` header present; `Cache-Control: public, max-age=300`; NO `X-Frame-Options`; CSP contains `frame-ancestors *` |
| R24-02 | `GET /embed/leaderboard?classId=1` | **200**; class-1 ranking |
| R24-03 | `GET /embed/leaderboard?classId=5` | **200**; global (all-classes) ranking (classId 5 → empty listClassId, embed_handler.go:57-60) |
| R24-04 | `GET /embed/leaderboard?limit=3` | **200**; at most 3 entries |
| R24-05 | `GET /embed/leaderboard?limit=1000` | **200**; capped to 25 entries |
| R24-06 | `GET /embed/leaderboard?limit=0` | **200**; `parsed>0` false → default 10 |
| R24-07 | `GET /embed/leaderboard?limit=-5` | **200**; default 10 |
| R24-08 | `GET /embed/leaderboard?limit=abc` | **200**; parse fails → default 10 |
| R24-09 | `GET /embed/leaderboard?classId=99` | **200**; empty list (class never validated) <!-- SUSPECT-6 --> |
| R24-10 | `GET /embed/leaderboard` with `Cookie: torrons_user_id=USER_50` | **200**; cookie ignored, NO new Set-Cookie |
| R24-11 | `POST /embed/leaderboard` | **405** |

## R25 — `GET /premsa` → `press` (`press_handler.go:73`)

Aggregate public stats + embed snippet generator. `http.Error` on failure.

| id | request | expect |
|---|---|---|
| R25-01 | `GET /premsa` | **200**, `text/html`; most-voted / riser / closest-duel / champion (each shown or its empty state) + category picker |
| R25-02 | `GET /premsa` `HX-Request: true` | **200**, `text/html` fragment |
| R25-03 | `POST /premsa` | **405** |

## R26 — `GET /wrapped` (hit as `/wrapped`) → `wrapped` (`wrapped_handler.go:48`)

Context user. **Gate = 50 total votes** (`getMinVotesForClass("5")`).

| id | request | expect |
|---|---|---|
| R26-01 | `GET /wrapped` cookie `USER_50` | **200**, `text/html`; `HasEnoughVotes=true`, real recap |
| R26-02 | `GET /wrapped` cookie `USER_49` | **200**; "not unlocked yet" state, `VotesRemaining=1` |
| R26-03 | `GET /wrapped` cookie `USER_0` | **200**; locked, `VotesRemaining=50` |
| R26-04 | `GET /wrapped` no cookie | **200**; middleware mints a 0-vote user → locked state |
| R26-05 | `GET /wrapped` `HX-Request: true` | **200**, `text/html` fragment |
| R26-06 | `POST /wrapped` | **405** |

## R27 — `GET /wrapped/card` (hit as `/wrapped/card.png`) → `wrappedCard` (`wrapped_handler.go:82`)

Context user. Same 50-vote gate (the PNG renders a locked card below it).

| id | request | expect |
|---|---|---|
| R27-01 | `GET /wrapped/card.png` cookie `USER_50` | **200**, `image/png`, `Cache-Control: private, no-store`, PNG bytes |
| R27-02 | `GET /wrapped/card.png` cookie `USER_49` | **200**, `image/png` — locked "keep voting" card (still 200) |
| R27-03 | `GET /wrapped/card` (no `.png`) | **200**, `image/png` |
| R27-04 | `POST /wrapped/card.png` | **405** |

## R28 — `GET /press-kit/card` (hit as `/press-kit/card.png`) → `pressKitCard` (`press_handler.go:215`)

Global aggregate PNG (champion). No per-user data. Public cache.

| id | request | expect |
|---|---|---|
| R28-01 | `GET /press-kit/card.png` (bracket completed, champion exists) | **200**, `image/png`, `Cache-Control: public, max-age=300`, PNG bytes |
| R28-02 | `GET /press-kit/card.png` (no champion yet) | **200**, `image/png` — empty-state card |
| R28-03 | `GET /press-kit/card` (no `.png`) | **200**, `image/png` |
| R28-04 | `POST /press-kit/card.png` | **405** |

## R29 — `GET /reveal` (hit as `/reveal`) → `reveal` (`reveal_handler.go:95`)

Context user. **Gate = 50 total votes** (`getMinVotesForClass("5")`).

| id | request | expect |
|---|---|---|
| R29-01 | `GET /reveal` cookie `USER_50` | **200**, `text/html`; persona badge + tagline |
| R29-02 | `GET /reveal` cookie `USER_49` | **200**; "not unlocked yet", `VotesRemaining=1` |
| R29-03 | `GET /reveal` cookie `USER_0` | **200**; locked, `VotesRemaining=50` |
| R29-04 | `GET /reveal` `HX-Request: true` | **200**, `text/html` fragment |
| R29-05 | `POST /reveal` | **405** |

## R30 — `GET /reveal/card` (hit as `/reveal/card.png`) → `revealCard` (`reveal_handler.go:129`)

Context user. Same 50-vote gate.

| id | request | expect |
|---|---|---|
| R30-01 | `GET /reveal/card.png` cookie `USER_50` | **200**, `image/png`, `Cache-Control: private, no-store`, PNG bytes |
| R30-02 | `GET /reveal/card.png` cookie `USER_49` | **200**, `image/png` — locked card |
| R30-03 | `GET /reveal/card` (no `.png`) | **200**, `image/png` |
| R30-04 | `POST /reveal/card.png` | **405** |

## R31–R34 — Static SEO pages (`content_handler.go`)

All zero-DB, identical pattern: `GET` → 200 `text/html`; `HX-Request: true` →
200 fragment; wrong method → 405.

| id | request | expect |
|---|---|---|
| R31-01 | `GET /sobre` | **200**, `text/html`, About/FAQ page |
| R31-02 | `POST /sobre` | **405** |
| R32-01 | `GET /torro-agramunt-igp` | **200**, `text/html`, IGP explainer |
| R32-02 | `POST /torro-agramunt-igp` | **405** |
| R33-01 | `GET /torro-agramunt-vs-xixona` | **200**, `text/html`, comparison page |
| R33-02 | `POST /torro-agramunt-vs-xixona` | **405** |
| R34-01 | `GET /tipus-de-torrons` | **200**, `text/html`, glossary |
| R34-02 | `GET /tipus-de-torrons` `HX-Request: true` | **200**, fragment |
| R34-03 | `PUT /tipus-de-torrons` | **405** |

---

## USER API (`/api/user`)

## R35 — `GET /api/user/stats` → `handleUserStats` (`user_api.go:26`)

Context user. JSON.

| id | request | expect |
|---|---|---|
| R35-01 | `GET /api/user/stats` cookie `USER_50` | **200**, `application/json; charset=utf-8`, body has `"user_id"`, `"total_votes":50`, `"class_votes"`, `"snapshot_count"`, `"current_streak"` |
| R35-02 | `GET /api/user/stats` no cookie | **200**; middleware mints user → `"total_votes":0`, `"class_votes":{}` |
| R35-03 | `GET /api/user/stats` cookie `USER_UNKNOWN` | **200**; middleware mints a fresh user (unknown UUID) → `total_votes:0`; a NEW `Set-Cookie` is returned with a different id |
| R35-04 | `POST /api/user/stats` | **405** |

The `401 {"error":"No user session found"}` branch (user_api.go:27-32) is
unreachable via the router (middleware guarantees a user).

## R36 — `GET /api/user/leaderboard/class/{classId}` → `handleUserLeaderboard` (`user_api.go:79`)

Path `{classId}`. Optional dietary query filters. JSON.

| id | request | expect |
|---|---|---|
| R36-01 | `GET /api/user/leaderboard/class/1` cookie `USER_50` | **200**, `application/json`, body has `"class_id":"1"`, `"vote_count":50`, `"entries"`, `"min_votes_required":30`, `"min_votes_met":true` |
| R36-02 | `GET /api/user/leaderboard/class/1` cookie `USER_0` | **200**; `"vote_count":0`, `"min_votes_met":false`, `"min_votes_required":30` |
| R36-03 | `GET /api/user/leaderboard/class/99` (nonexistent) | **200**; `"min_votes_required":25` (default), `"entries":[]`, no 404 <!-- SUSPECT-6: class never validated here, unlike /api/leaderboard/class/{classId} which 404s --> |
| R36-04 | `GET /api/user/leaderboard/class/BAD_ID` | **200**; empty entries, default min_votes 25 |
| R36-05 | `GET /api/user/leaderboard/class/1?vegan=true` | **200**; filtered entries |
| R36-06 | missing segment `GET /api/user/leaderboard/class/` | **404** (chi; empty param not matched → falls to `/api/user/leaderboard/global`? No — distinct literal; 404) |
| R36-07 | `POST /api/user/leaderboard/class/1` | **405** |

## R37 — `GET /api/user/leaderboard/global` → `handleUserGlobalLeaderboard` (`user_api.go:126`)

Context user. **Gate = hardcoded 50.** JSON.

| id | request | expect |
|---|---|---|
| R37-01 | `GET /api/user/leaderboard/global` cookie `USER_50` | **200**, `application/json`, `"total_votes":50`, `"min_votes_required":50`, `"min_votes_met":true`, `"entries"` |
| R37-02 | `GET /api/user/leaderboard/global` cookie `USER_49` | **200**; `"min_votes_met":false`, `"min_votes_required":50` |
| R37-03 | `GET /api/user/leaderboard/global?organic=true` | **200**; filtered |
| R37-04 | `POST /api/user/leaderboard/global` | **405** |

---

## CAMPAIGN API (`/api/campaign`)

## R38 — `GET /api/campaign/countdown` → `handleCountdown` (`campaign_api.go:38`)

JSON. No active campaign is a **200** with an inactive body, not an error.

| id | request | expect |
|---|---|---|
| R38-01 | `GET /api/campaign/countdown` (active campaign) | **200**, `application/json`, `"is_active":true`, `"campaign_id"`, `"time_remaining_seconds"`, `"days_remaining"` |
| R38-02 | `GET /api/campaign/countdown` (no active campaign) | **200**, `application/json`, `{"is_active":false,"has_ended":true}` (other fields zero/omitted) |
| R38-03 | `POST /api/campaign/countdown` | **405** |

## R39 — `GET /api/campaign/countdown/widget` → `handleCountdownWidget` (`campaign_api.go:93`)

Returns HTML fragment. Never errors hard (falls back to inline error div).

| id | request | expect |
|---|---|---|
| R39-01 | `GET /api/campaign/countdown/widget` (active) | **200**, `text/html`; countdown widget markup |
| R39-02 | `GET /api/campaign/countdown/widget` (no active) | **200**, `text/html`; inactive widget state |
| R39-03 | `POST /api/campaign/countdown/widget` | **405** |

## R40 — `GET /api/campaign/info` → `handleCampaignInfo` (`campaign_api.go:298`)

JSON. No active campaign → **404** (unlike countdown).

| id | request | expect |
|---|---|---|
| R40-01 | `GET /api/campaign/info` (active) | **200**, `application/json`, `"id"`, `"name"`, `"start_date"`, `"end_date"`, `"status"`, `"year"` |
| R40-02 | `GET /api/campaign/info` (no active) | **404**, `application/json`, `{"error":"No active campaign"}` |
| R40-03 | `POST /api/campaign/info` | **405** |

---

## LEADERBOARD API (`/api/leaderboard`)

## R41 — `GET /api/leaderboard/global` → `handleGlobalLeaderboard` (`campaign_api.go:160`)

Public, no gate. Raw SQL, top 100 non-discontinued by rating. JSON.

| id | request | expect |
|---|---|---|
| R41-01 | `GET /api/leaderboard/global` | **200**, `application/json`, `"entries"` (array of `{rank,torron_id,torron_name,image,rating,class_name}`), `"total_entries"`, `"timestamp"` |
| R41-02 | `GET /api/leaderboard/global` no cookie | **200**; identical (no user dependency) |
| R41-03 | `POST /api/leaderboard/global` | **405** |

## R42 — `GET /api/leaderboard/class/{classId}` → `handleClassLeaderboard` (`campaign_api.go:213`)

Path `{classId}`. Validates class existence → 404 if unknown. JSON.

| id | request | expect |
|---|---|---|
| R42-01 | `GET /api/leaderboard/class/1` | **200**, `application/json`, `"class_id":"1"`, `"class_name"`, `"entries"`, `"total_entries"`, `"timestamp"` |
| R42-02 | `GET /api/leaderboard/class/99` (nonexistent) | **404**, `application/json`, `{"error":"Class not found"}` (campaign_api.go:239-243) |
| R42-03 | `GET /api/leaderboard/class/BAD_ID` | **404**, `{"error":"Class not found"}` |
| R42-04 | `GET /api/leaderboard/class/SQL_ID` | **404**, `{"error":"Class not found"}` |
| R42-05 | missing segment `GET /api/leaderboard/class/` | **404** |
| R42-06 | `POST /api/leaderboard/class/1` | **405** |

---

## R43 — `GET /public/*` → static `http.FileServer` (`server.go:151`)

Serves the embedded `public/` FS. No user dependency (still passes through
UserMiddleware since it's not `/embed/` — a cookie is minted).

| id | request | expect |
|---|---|---|
| R43-01 | `GET /public/<a real asset>` (e.g. an existing css/js/img) | **200**, correct sniffed/served Content-Type, asset bytes |
| R43-02 | `GET /public/does-not-exist.xyz` | **404** |
| R43-03 | `GET /public/../server.go` (path traversal attempt) | **404** (http.FileServer cleans the path; escape blocked) |
| R43-04 | `POST /public/<asset>` | **405** |

---

## GLOBAL / CROSS-CUTTING CASES

| id | request | expect |
|---|---|---|
| G-01 | `GET /totally-unknown-path` | **404** (chi default NotFound) |
| G-02 | Global rate limit: 101st request in 60s from the SAME IP (set via `X-Forwarded-For`/RemoteAddr; RealIP canonicalizes) across any routes | **429**, `text/plain`, `Rate limit exceeded. Please try again later.\n` |
| G-03 | Any non-`/embed/` GET with no cookie | response includes `Set-Cookie: torrons_user_id=<uuid>; Path=/; Max-Age=7776000; HttpOnly; SameSite=Lax` |
| G-04 | Any `/embed/` GET | response has NO `Set-Cookie`; CSP contains `frame-ancestors *`; NO `X-Frame-Options` |
| G-05 | Any non-`/embed/` GET | headers include `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, a `Content-Security-Policy` |
| G-06 | cookie present-but-unknown UUID on any user-backed route | handler sees a freshly-minted 0-vote user AND a new `Set-Cookie` replaces the unknown id (middleware.go:60-96) |

---

## SUSPECT LIST (behavior that looks wrong; encoded as-is above)

- **SUSPECT-1** — `GET /classes/{id}/vote` (`handler.go:162-167`): a nonexistent
  or empty class (no pairings) makes `pairingRepo.GetRandom` return a
  NotFound-classified error, but the handler wraps it in `domain.ErrInternal`
  → **HTTP 500** with body `{"code":2506,"message":"Record not found"}`. Correct
  would be **404**. Affects R7-04/05/06. (Also fires for a real class that
  simply has zero pairings.)

- **SUSPECT-2** — `POST /pairings/{id}/vote` (`handler.go:250-255`): a
  nonexistent/malformed pairing id → `pairingRepo.Get` NotFound wrapped in
  `ErrInternal` → **HTTP 500** (`{"code":2506,...}`) instead of **404**. Affects
  R8-05/06/07.
  - **SUSPECT-2b** — same handler (`handler.go:409-417`): a second advent vote
    the same day violates the `AdventVotes` unique index; that DuplicateKey
    error is wrapped in `ErrInternal` → **500** instead of a 4xx. Affects R8-09.

- **SUSPECT-3** — Inconsistent global-leaderboard vote gate. The personal-global
  view of `GET /leaderboard` calls `getMinVotesForClass("global")`, which hits
  the `default` case → **25** (`leaderboard.go:166` + `user_api.go:182-184`),
  while `/wrapped`, `/reveal`, and `/api/user/leaderboard/global` all gate the
  "global" recap at **50**. So a user with 25–49 votes is "unlocked" on
  `/leaderboard` but "locked" everywhere else. Affects R10-02.

- **SUSPECT-6** — Class ids are not validated on several read paths, so an
  unknown class returns **200 with empty data** instead of 404:
  `GET /leaderboard?category=99` (R10-07), `GET /embed/leaderboard?classId=99`
  (R24-09), `GET /api/user/leaderboard/class/99` (R36-03). Contrast
  `GET /api/leaderboard/class/99`, which DOES 404 (R42-02) — inconsistent
  contract between the two class-scoped leaderboard APIs.

- **SUSPECT-7** — `GET /history?offset=-5` (`history_handler.go:61-99`): a
  negative `offset` is passed straight into SQL `OFFSET $n`; Postgres rejects a
  negative OFFSET, so the query errors → **HTTP 500** `Internal Server Error`.
  Not clamped to 0. Affects R12-05.

- **SUSPECT-8** — `POST /bracket/{classId}/create?size=1` (`bracket_handler.go:611`):
  `isPowerOfTwo(1)` is `true`, so size=1 passes the power-of-two guard and is
  only rejected later by the "needs at least 2 active torrons" check (because
  `TopNByClass(..., 1)` returns <2). The user-facing message is misleading, and
  the real invariant (`Brackets.Size >= 2`, DB CHECK `chk_bracket_size_positive`)
  is never surfaced. Affects R17-09.
