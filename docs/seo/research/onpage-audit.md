# On-Repo SEO Audit — Torrorèndum (torro.cat)

**Date:** 2026-08-17 · **Scope:** static analysis of the repo at `/home/user/torrons` (all 21 templates in `public/templates/`, HTTP layer in `internal/http/`, assets in `public/`). No live-site fetches were made in this audit (that is the live-site audit's job); every observation below cites `file:line` in the repo at the commit present on 2026-08-17.

**Overall verdict:** the on-page layer is in *very good* shape — far better than typical for a fan project. Every indexable page has a unique title, meta description, canonical, correct robots meta, full OG/Twitter sets, JSON-LD with breadcrumbs, one `h1`, `lang="ca"`, and 100 % image alt coverage. The historical P0 (sitemap/robots/llms 404 via `middleware.URLFormat`) **is fixed**. The remaining work is: image weight (19 MB of ~1 MB PNGs — the single biggest ranking-relevant defect), a dead `error.html` template reference, third-party render-blocking dependencies (unpkg htmx + Google Fonts), and a tail of metadata polish items.

---

## 1. Verification of the known URLFormat bug (AUDIT_2026-07-09 finding #4)

**Status: FIXED — verified in code and tests.**

- `internal/http/server.go:64` still registers `middleware.URLFormat` globally, but `server.go:214-216` now registers the SEO routes **dotless** (`/robots`, `/sitemap`, `/llms`) with an explanatory comment (`server.go:207-213`), mirroring the `/share/card` trick (`server.go:247`). URLFormat strips `.txt`/`.xml` before chi matches, so both `/robots.txt` and `/robots` resolve to the handler.
- Regression coverage exists: `internal/http/seo_handler_test.go:64-132` requests `/robots.txt` and `/sitemap.xml` and asserts content; `test/e2e/test_seo.sh` exists as an end-to-end guard.
- **Side effect worth knowing (P3):** URLFormat makes *every* route resolve under any extension — `/sobre.html`, `/sobre.json`, `/sitemap.xml`/`/sitemap` all return 200 with identical content. Duplicate-URL risk is mitigated by the absolute `rel=canonical` on every page, but the dotless `/sitemap`, `/robots`, `/llms` variants are live duplicate endpoints. Harmless; do not "fix" by removing URLFormat without re-checking `/share/card.png`, `/wrapped/card.png`, `/press-kit/card.png`, `/reveal/card.png`.

## 2. Per-page matrix (all templates, `public/templates/`)

Legend: ✔ present/correct · ✘ missing · n/a not applicable (fragment or noindex page).

| Template (route) | Title unique | Meta desc | Canonical | Robots | OG/Tw | JSON-LD | h1 | img alt | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `index.html` (/) | ✔ "Torrorèndum {year} - Vota el millor torró" | ✔ good | ✔ `https://torro.cat` (no trailing slash; og:url has slash — normalize, P3) | index | 9/5 ✔ | WebSite + WebApplication + Brand `mentions` (disclosure-safe, index.html:17-46) | 1 | 3/3 | Anchor page; JSON-LD deliberately avoids claiming Vicens as publisher — keep. |
| `ranquing.html` (/ranquing-de-torrons) | ✔ 63-char keyword title | ✔ dynamic w/ vote count (ranquing.html:7) | ✔ | index | 9/5 ✔ | ItemList mirroring visible standings + Breadcrumb (ranquing.html:13-42) | 1 | 1/1 | Flagship SEO page. No aggregateRating anywhere — policy-safe. ItemList comma logic `{{ if gt .Rank 1 }},{{ end }}` assumes ranks start at 1 (holds today). |
| `torro.html` (/torro/{id}) | ✔ product name | ✔ ~200 chars (slightly long) | ✔ per-id | index | 9/5 ✔ | Product (deliberately **no** aggregateRating — documented at torro.html:11-15) + Breadcrumb | 1 | 2/2 | `weight` emitted as plain string not QuantitativeValue (torro.html:24) — validator warning only. og:image is generic site image, not the product photo (torro.html:50) — missed CTR win. og:type "website" not "product". |
| `about.html` (/sobre) | ✔ | ✔ | ✔ | index | 9/5 ✔ | FAQPage (8 Q/A) + Breadcrumb | 1 | n/a | FAQPage rich results are restricted by Google to gov/health since 2023 — keep for AEO/LLM extraction, expect no SERP rich result. Answers match visible content. |
| `igp.html` (/torro-agramunt-igp) | ✔ | ✔ | ✔ | index | 9/5, og:type=article ✔ | Article + Breadcrumb (igp.html:22-32) | 1 | n/a | Article lacks `datePublished`/`dateModified` and `image` — P3 enrichment. |
| `comparativa.html` (/torro-agramunt-vs-xixona) | ✔ | ✔ | ✔ | index | 9/5 ✔ | Article + Breadcrumb | 1 | n/a | Same Article date/image gap. |
| `glossari.html` (/tipus-de-torrons) | ✔ | ✔ | ✔ | index | 9/5 ✔ | DefinedTermSet + 12 DefinedTerm + Breadcrumb | 1 (h2×4, h3×11, ordered) | n/a | Excellent AEO structure. |
| `press.html` (/premsa) | ✔ | ✔ | ✔ | index | 9/5 ✔ | Dataset + Breadcrumb (press.html:24-34) | 1 | 5/5 | Dataset lacks `license` + `distribution` — needed for Google Dataset Search eligibility (P3). |
| `classes.html` (/classes) | ✔ | thin: "Escull la teva categoria de torrons {year}" | ✔ | index | 9/5 ✔ | Breadcrumb | 1 | n/a | Desc thin but page is navigational — acceptable; could be enriched. |
| `advent.html` (/advent) | ✔ | **✘ desc = title verbatim** ("El duel de l'Advent - Torrorèndum {year}", advent.html:7) | ✔ | index | 9/5 ✔ | Breadcrumb only | 1 | 1/1 | P2: write a real description (what an advent duel is, one per day, Dec 1–24). Indexed year-round though content is seasonal/empty off-season — consider explanatory off-season copy (template already has CampaignActive states). |
| `bracket.html` (/bracket/{classId}) | **✘ NOT unique per class** — "Bracket - Torrorèndum {year}" for every class (bracket.html:52) though `.ClassName` is available (used in breadcrumb, bracket.html:18) | thin + duplicated per class (bracket.html:7) | ✔ per-class | index | full ✔ | Breadcrumb | 1 | 5/5 | P2. Second embedded doc in same file = /bracket/{id}/vote view: correctly `noindex, follow` with comment (bracket.html:204), no canonical — correct. |
| `leaderboard.html` (/leaderboard) | ✔ | ✔ | none | **noindex, follow** — correct (personalized view; public twin is /ranquing-de-torrons per server.go:231-233) | 9/5 | — | 1 | 2/2 | Correct choice. |
| `vote.html` (/classes/{id}/vote) | ✔ | ✔ | none | noindex ✔ (randomized per request — correct) | 8/4 (no og:url/twitter:url — fine for noindex) | — | **0 h1** | 1/1 | h1 absence is cosmetic on a noindex page (P3 accessibility nicety). |
| `stats.html`, `history.html`, `wrapped.html`, `reveal.html`, `friends.html` | ✔ each | ✔ | none | noindex, follow ✔ (personal pages — correct) | 9/5 | — | 1 (friends.html has two h1s at :80 and :164 but in mutually exclusive template defines — only one renders per view) | ✔ | Correct pattern throughout. |
| `embed_leaderboard.html` (/embed/leaderboard) | ✔ | ✔ | none | **noindex, follow** with rationale comment (embed_leaderboard.html:8-11) — correct: iframe-only widget, avoids thin-duplicate indexing while letting the host-page backlink pass signals | 0/0 (n/a) | — | 0 (n/a) | 1/1 | Framing enabled via per-path CSP `frame-ancestors *` (server.go:147-152) while all other routes keep X-Frame-Options DENY — correct split. |
| `countdown.html`, `pairing.html` | n/a — head-less fragments only ever returned to htmx requests (POST result / widget endpoint) | | | | | | | | `/api/campaign/countdown/widget` returns HTML with no robots meta and no X-Robots-Tag — practically un-indexed (never linked, under /api/), but `Disallow: /api/` in robots.txt would be a cheap belt-and-braces (P3). |

**Cross-cutting positives:** `lang="ca"` on every full-page template; `{{ if not .HX }}` guard present on all 18 full-page templates — `isHX` (handler.go:584-589) keys solely off the `HX-Request` header, which crawlers never send, so **crawlers always receive the complete document** (head, JSON-LD, header/topbar/footer). `Vary: HX-Request` is set on cached pages (middleware.go:283-286) so shared caches can't serve a fragment to a crawler. Every `<img>` in every template has `alt` (spot-verified counts match template-wide).

## 3. Sitemap vs. routes diff (seo_handler.go:71-125 vs server.go:200-307)

Sitemap emits: `/`, `/ranquing-de-torrons`, `/classes`, `/premsa`, `/advent`, `/sobre`, `/torro-agramunt-igp`, `/torro-agramunt-vs-xixona`, `/tipus-de-torrons`, every `/torro/{id}` (DB loop), and `/bracket/{classId}` only for classes with an existing bracket (seo_handler.go:113-119 — correctly avoids sitemap-listing URLs that would render an empty state).

- **Sitemap entries that could 404: none found.** All static entries map to registered routes; `/torro/{id}` entries come from the same repo `List()` the handler serves from; bracket entries are existence-checked.
- **Indexable routes missing from sitemap: none.** Every template with `index` robots meta is either listed or covered by a loop. Noindex pages are deliberately excluded (comment at seo_handler.go:66-70) — consistent signals.
- Gaps (P3): no `<lastmod>` (would help recrawl scheduling on the ranking page), no `Cache-Control` on the sitemap/robots/llms responses themselves.

## 4. llms.txt accuracy (seo_handler.go:36-64)

Accurate overall; disclosure framing consistent with JSON-LD and footer. Two nits: (a) it points LLMs at `/leaderboard` ("Classificació — live rankings (per-visitor view)") which is a noindex personalized page — the ranking claim should route exclusively to `/ranquing-de-torrons` (it already flags it as "best page to cite", so this is minor); (b) `/torro/{id}` product pages and `/bracket/*` aren't mentioned even as a pattern. P3.

## 5. Server layer

| Area | Status | Evidence |
|---|---|---|
| Compression | ✔ gzip level 5 for HTML/CSS/JS/JSON/XML/SVG | server.go:71-75 |
| Cache headers | ✔ tiered: 30 d images/icons/assets, 1 h CSS/JS (server.go:187-197); 1 h + `Vary: HX-Request` on static pages and /ranquing-de-torrons (handler.go:156, content_handler.go:41, ranking_handler.go:82); 5 min on /premsa + /embed/leaderboard (press_handler.go:417, embed_handler.go:123); `private, no-store` on personal PNG cards (sharecard_handler.go:79, wrapped_handler.go:119, reveal_handler.go:166) | correct choices |
| ETag | ✘ none anywhere (http.FileServer over embedded FS emits no ETag/Last-Modified since embedded files have zero modtime) | P3 — Cache-Control mostly covers it; fingerprinted asset names noted as follow-up in server.go:184-185 |
| Security headers | ✔ HSTS (proxy-aware), nosniff, Referrer-Policy, Permissions-Policy, CSP; X-Frame-Options DENY except /embed/* | server.go:113-168 |
| X-Robots-Tag | none set anywhere (grep across internal/) — all robots control is via meta tags, which is fine for HTML; the four PNG card endpoints have no robots control (they're no-store and unlinked-for-crawlers; P3: add `X-Robots-Tag: noindex` to PNG responses) | |
| Trailing slash / case | ✘ no RedirectSlashes/CleanPath middleware: `/sobre/` and `/SOBRE` → 404 (chi exact-match). Not a duplicate-content risk (404, not 200), only a lost-redirect nicety. Host canonicalization (www→apex, http→https) not in repo — lives at the proxy; must be verified in the live audit | server.go:60-65 |
| 404 behavior | chi default: correct `404` status, plain-text "404 page not found" — no branded HTML 404 with recovery links (P3 UX) | |
| 5xx fallback | **BUG:** every handler's error path does `ExecuteTemplate(w, "error.html", …)` but **no template named `error.html` exists** in `public/templates/` and no `{{define "error…"}}` anywhere (verified by grep). The call always fails, logs a second error, and falls through to `http.Error` after `WriteHeader(500)` was already sent → "superfluous WriteHeader" logs and a plain-text 500. 10+ call sites: handler.go:163, 192, 298, 574; content_handler.go:49; and equivalents in other handlers | P2 — dead code masking the intended branded error page |
| Rate limiting vs crawlers | 100 req/min/IP global (server.go:78-86). Googlebot typically stays under this, but an image-heavy crawl (19 MB of product images route through `/public/*` which **is** rate-limited) could hit 429s during seasonal recrawl spikes. Consider exempting `/public/*` or raising the limit for verified bots | P3 |

## 6. Performance (ranking-relevant)

1. **Images are the #1 defect (P1).** `public/images/` totals **19 MB**; top offenders ~1 MB each (`xoco_festuc.png` 1012 K, `tou.jpg` 988 K, `iogurt.png` 976 K, `brutal.png` 968 K, `xoco_cruixent.png` 952 K). These render at ≤400 px on the vote screen and 64 px thumbnails on the ranking (ranquing.html:113 does set `width/height` + `lazy`; vote.html:129 sets neither). `optimize-images.sh` + `IMAGE_OPTIMIZATION.md` exist in-repo but **have not been run** ("optional for launch" note, IMAGE_OPTIMIZATION.md:21). This is the main LCP/CWV drag on the two most important page types (vote + torró detail). Fix: run the script, prefer WebP/AVIF `<picture>` or at least resized PNGs, add explicit `width`/`height` on `torro-photo` and `torron-image`.
2. **CSS:** single `main.css` = 194 KB raw (public/css/, gzip ~compresses well, 1 h TTL). No render-blocking third-party CSS other than fonts. P3: consider critical-CSS or pruning; not urgent.
3. **Fonts:** render-blocking Google Fonts stylesheet on every page (e.g. index.html:76-78) with `preconnect` + `display=swap` — acceptable but a cross-origin SPOF and a GDPR-gray dependency; self-hosting the two families would remove a blocking request chain (P2/P3).
4. **htmx from unpkg (P2):** `htmx.org@1.9.9` + `json-enc` from `unpkg.com` with SRI (index.html:79, torro.html:72-73). unpkg is a known single point of failure (repeated outages); it's render-blocking-ish in `<head>` (no `defer`), CSP must whitelist unpkg (server.go:160), and htmx 1.9.9 is old (1.9.12 fixed bugs; 2.x is current). Self-host the two files under `/public/js/` (they're ~14 KB gzipped), drop `https://unpkg.com` from CSP.

## 7. Structured-data policy review

- **No self-serving AggregateRating anywhere** — explicitly and correctly avoided with an in-code rationale (torro.html:11-15). Keep this stance; adding `aggregateRating` from own-site ELO votes would risk a structured-data manual action.
- FAQPage on /sobre: valid but rich-result-ineligible for this site class (Google Aug-2023 restriction); harmless, useful for AEO.
- ItemList on the public ranking mirrors visible content (policy-safe); `itemListOrder` says "Ascending" while the list is best-first — semantically debatable, no practical risk.
- Product on /torro/{id}: no `offers` → ineligible for merchant listing rich results, which is **correct** for a non-commerce fan site (adding offers pointing at vicens.com would be misrepresentative). `weight` as a bare string is a validator warning (torro.html:24).
- Article on igp/comparativa: add `datePublished`/`dateModified` (+`image`) for freshness signals (P3).
- Dataset on /premsa: add `license` and `distribution` to become Dataset-Search-eligible (P3).
- Go `html/template` auto-escapes inside `<script type="application/ld+json">` using JS/JSON-safe escaping, so dynamic values (`.Torro.Name`, descriptions) can't break the JSON — no injection/validity risk found. Conditional comma patterns (torro.html:23-25) are safe because the final `brand` key is unconditional.

## 8. Internal linking & discovery

- Global footer (index.html:450-469, included by all 18 full-page templates) links every SEO page: /, /classes, /leaderboard, /ranquing-de-torrons, /bracket/5, /premsa, /sobre, all three content pages. Global topbar adds /advent, /stats, etc. **No orphan indexable pages.** `/torro/{id}` pages are reachable from ranquing/leaderboard/press/friends lists and cross-link related products (torro.html:235-252) — good crawl mesh.
- BreadcrumbList JSON-LD on all indexable pages; no *visible* breadcrumb UI (P3 — visible breadcrumbs are what appear in SERPs alongside the markup).
- Footer disclosure ("projecte de fans independent, sense cap relació oficial amb Torrons Vicens", index.html:464-467) renders site-wide on full-page loads — the legal disclosure reaches crawlers on every page. Do not weaken.
- Links use `href` + hx-boost, so plain-HTML crawlers follow normal anchors. ✔

## 9. hreflang / i18n readiness

No `hreflang` anywhere (grep: zero hits) — correct today (single ca version; `og:locale ca_ES`, `lang="ca"`, JSON-LD `inLanguage: ca` are consistent). **When a Spanish version ships:** add per-page `<link rel="alternate" hreflang="ca" …>` + `hreflang="es"` + `x-default` in each template head (between canonical and JSON-LD), mirror in sitemap `<xhtml:link>` entries (seo_handler.go:100 loop), and decide URL scheme (`/es/…` subtree is simplest with chi mounting). Also translate `llms.txt` sections.

## 10. Semantic HTML / landmarks

- ✔ `<footer>` element, `<nav>` with aria-labels (footer nav, topbar `role="navigation"`), `<time datetime>` on ranking update stamp (ranquing.html:104), aria-labels on icons/decorative spans, `role="img"` with labels on emoji badges (vote.html:138-141).
- ✘ Main content is `<div id="main-content">` not `<main>` (all page shells, e.g. torro.html:82); header is `<div id="header">` not `<header>`. P3 accessibility/semantics polish — htmx targets the div by id, so switching tag is a one-line change per template.

## 11. Favicons / manifest / OG image

Complete set in `public/icons/`: `favicon.ico`, 16/32 PNGs, `apple-touch-icon.png`, `android-chrome-192/512`, `site.webmanifest` — all referenced in every full-page head (index.html:70-74). `public/assets/og-image.jpg` exists (referenced 1200×630 everywhere). ✔ No gaps.

## 12. Prioritized findings

| P | Finding | Where | Fix |
|---|---|---|---|
| **P0** | *(none — the historical P0 sitemap/robots 404 is verified fixed)* | server.go:214-216; seo_handler_test.go | keep e2e guard |
| **P1** | 19 MB unoptimized product images, ~1 MB PNGs on vote/detail (LCP/CWV drag on the highest-value pages); `optimize-images.sh` never run; vote.html img has no width/height | public/images/*, vote.html:129, torro.html:109 | run script → resized WebP/AVIF, add dimensions |
| **P2** | `error.html` template referenced by every handler error path but does not exist → intended branded error page unreachable, double-WriteHeader logs | handler.go:163 et al., content_handler.go:49 | add `public/templates/error.html` (or a `{{define "error.html"}}`) |
| **P2** | htmx 1.9.9 + json-enc loaded from unpkg.com — SPOF, extra origin, stale version, CSP widening | all heads e.g. index.html:79, torro.html:72-73; CSP server.go:160 | self-host under /public/js/, add `defer`, tighten CSP |
| **P2** | /bracket/{classId} titles+descriptions identical across classes despite `.ClassName` being available | bracket.html:7,52 | interpolate ClassName |
| **P2** | /advent meta description duplicates the title verbatim | advent.html:7 | write real description |
| **P3** | Render-blocking Google Fonts (SPOF/consent-gray); consider self-host | index.html:76-78 etc. | self-host woff2 |
| **P3** | Index canonical `https://torro.cat` vs og:url `https://torro.cat/` inconsistency | index.html:15,51 | normalize to `/` |
| **P3** | torro og:image = generic site card instead of product photo; og:type "website" | torro.html:46-50 | per-product og:image |
| **P3** | Twitter tags use `property=` instead of `name=` (X parses OG fallback, so low impact) | all templates | switch to name= |
| **P3** | No `<lastmod>` in sitemap; no Cache-Control on sitemap/robots/llms responses | seo_handler.go:100-121 | add lastmod for ranking/content pages |
| **P3** | llms.txt cites noindexed /leaderboard; omits /torro/{id} pattern | seo_handler.go:49 | point rankings only at /ranquing-de-torrons |
| **P3** | No trailing-slash/case redirect (404 instead of 301); no branded 404 page; host canonicalization delegated to proxy (verify live) | server.go router | chi middleware.RedirectSlashes + custom NotFound |
| **P3** | Article schema missing dates/image; Dataset missing license/distribution; Product `weight` bare string | igp.html:22, comparativa, press.html:24, torro.html:24 | enrich |
| **P3** | `<div id="main-content">`→`<main>`, `<div id="header">`→`<header>`; vote page lacks h1; no visible breadcrumb UI | page shells | semantic swap |
| **P3** | PNG card endpoints lack `X-Robots-Tag: noindex`; `/api/` not disallowed in robots.txt; global 100 req/min/IP limit also covers /public/* assets during crawl spikes | sharecard/wrapped/reveal handlers; seo_handler.go:25-28; server.go:78 | cheap hardening |

## 13. Answers to specific checklist questions

- **`{{ if not .HX }}` / crawler completeness:** confirmed safe. `isHX` = `HX-Request: true` header only (handler.go:584-589); crawlers get the full document; cached variants are separated by `Vary: HX-Request` (middleware.go:285). No cloaking risk: both variants contain the same content body.
- **X-Robots-Tag:** none emitted anywhere; all robots directives are per-page meta tags, and every choice (index vs noindex) was reviewed above and found appropriate for its page type.
- **/embed/leaderboard indexability:** `noindex, follow` meta (embed_leaderboard.html:11) + permissive `frame-ancestors *` CSP only for /embed/* (server.go:147-152) + 5-min cache (embed_handler.go:123). Correct configuration for a backlink-earning widget.
- **Sitemap vs routes:** zero mismatches in either direction (section 3).
- **Compression/ETag/cache:** gzip yes; ETag no; Cache-Control tiered and sensible (section 5).

## Snapshot 2026-08-17 (re-measurable baseline)

Re-run these exact greps/commands against the repo to diff a future state:

- `middleware.URLFormat` registered: server.go:64; dotless SEO routes: server.go:214-216 (`/robots`, `/sitemap`, `/llms`). Expected: unchanged or URLFormat removed with card routes re-verified.
- `grep -rn 'define "error' public/templates/` → **0 hits** (the P2 bug). Expect ≥1 after fix.
- `du -sh public/images` → **19 M**; top-5: xoco_festuc.png 1012K, tou.jpg 988K, iogurt.png 976K, brutal.png 968K, xoco_cruixent.png 952K. Expect ≤5 M after optimization.
- `stat -c %s public/css/main.css` → **194 777 bytes**.
- `grep -c 'unpkg.com' public/templates/*.html` → nonzero in 18 full-page templates (htmx 1.9.9 + json-enc). Expect 0 after self-hosting.
- `grep -c hreflang public/templates/*.html` → 0 everywhere (correct until /es ships).
- Robots meta census (template → directive): index/ranquing/torro/about/igp/comparativa/glossari/press/classes/advent/bracket-overview = `index, follow, max-image-preview:large`; leaderboard/vote/stats/history/wrapped/reveal/friends/embed_leaderboard/bracket-vote = `noindex, follow`.
- Sitemap static set (seo_handler.go:85-98): `/ 1.0`, `/ranquing-de-torrons 0.9`, `/classes 0.8`, `/premsa 0.5`, `/advent 0.5`, `/sobre 0.6`, `/torro-agramunt-igp 0.6`, `/torro-agramunt-vs-xixona 0.6`, `/tipus-de-torrons 0.6`, + `/torro/{id} 0.7` loop + conditional `/bracket/{classId} 0.6`.
- JSON-LD type census: WebSite, WebApplication, Brand (index); ItemList+Breadcrumb (ranquing); Product+Brand+Breadcrumb (torro); FAQPage 8×Q/A+Breadcrumb (about); Article+Breadcrumb (igp, comparativa); DefinedTermSet+12 DefinedTerm+Breadcrumb (glossari); Dataset+Breadcrumb (press); Breadcrumb only (classes, advent, bracket). **AggregateRating: 0 occurrences repo-wide.**
- Cache-header call sites: handler.go:156, content_handler.go:41, ranking_handler.go:82 (1 h public + Vary); press_handler.go:417, embed_handler.go:123 (5 min); sharecard_handler.go:79, wrapped_handler.go:119, reveal_handler.go:166 (private no-store); server.go:192/194 (assets 30 d / 1 h).
- Titles observed (for uniqueness diffing): see per-page matrix, section 2.

*Note for companion reports: WebSearch-based SERP work in this workflow runs from US infrastructure — geo-biased vs. real es/ca SERPs; this on-repo audit is unaffected.*
