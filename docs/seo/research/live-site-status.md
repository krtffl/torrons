# Live Site Status — torro.cat (Torrorèndum)

**Run date:** 2026-08-17 (~15:45–16:05 UTC)
**Repo state at run time:** branch `claude/seo-positioning-strategy-qkjloq`, HEAD `1fc5890` ("docs(seo): research snapshots …"); previous feature commit `d70707d` ("feat(seo): public ranking page, dynamic season year, product schema, full icon set").
**Purpose:** canonical "our side" baseline for future diffing.

---

## 0. CRITICAL METHOD LIMITATION (read first)

Direct network access to torro.cat (and to the general web) is **blocked in this sandbox**:

- `curl https://torro.cat/` via the mandated agent proxy → **`CONNECT tunnel failed, response 403`** (proxy log: `gateway answered 403 to CONNECT (policy denial or upstream failure)` for `torro.cat:443` and `www.torro.cat:443`, timestamps 2026-08-17T15:46:35–45Z). Per proxy README, policy denials must not be retried.
- `WebFetch https://torro.cat/` → `{"error_type":"EGRESS_BLOCKED","domain":"torro.cat","message":"Access to torro.cat is blocked by the network egress proxy."}`
- `WebFetch https://en.wikipedia.org/...` → also `EGRESS_BLOCKED` — the blockage is blanket, not torro.cat-specific.
- **WebSearch works** (Anthropic-side API), so Section 3 (indexation/visibility) IS a genuine live external measurement.

**Consequence:** Task items 1, 2 and 4 (live HTTP status/headers/weights, redirect hygiene, og-image 200 check, llms.txt live parse) could **not** be measured live today. In their place, Section 2 records the **expected as-deployed values derived from the repo at commit `1fc5890`** — exact strings a future run (from an unrestricted network) can diff against actual responses. Any mismatch then means either (a) deployed build ≠ this commit, or (b) infra (reverse proxy/DNS) alters behavior.

---

## 1. What could NOT be verified live (checklist for the next run)

For each of these, record on the next run: HTTP status, `Content-Encoding`, `Cache-Control`, `Strict-Transport-Security`, `X-Frame-Options`/CSP, `<title>`, meta description, canonical, JSON-LD `@type`s, robots meta, transfer size, `time_total`:

- `https://torro.cat/` · `/robots.txt` · `/sitemap.xml` · `/llms.txt` · `/sobre` · `/premsa` · `/tipus-de-torrons` · `/torro-agramunt-igp` · `/torro-agramunt-vs-xixona` · `/classes` · `/ranquing-de-torrons` · one `/torro/{id}` from the live sitemap
- `http://torro.cat/` (expect 301→https; **no in-app redirect exists — host/infra dependent, unverified**)
- `https://www.torro.cat/` (no in-app www handling; DNS/proxy dependent, **unverified** — proxy log shows we attempted `www.torro.cat:443`, blocked)
- Trailing-slash behavior (chi router: `/sobre/` is a distinct path; chi does NOT auto-redirect — likely **404 on trailing slash** unless infra normalizes. Verify!)
- `https://torro.cat/public/assets/og-image.jpg` (expect 200, `image/jpeg`, ~43,192 bytes, `Cache-Control: public, max-age=2592000`)

## 2. Expected as-deployed state (from repo, commit `1fc5890`) — the diff baseline

### 2.1 Global response behavior (internal/http/server.go, middleware.go)

- **Compression:** chi `middleware.Compress(5, ...)` for HTML/CSS/JS/JSON/XML/SVG → expect `Content-Encoding: gzip` on text responses.
- **Security headers on every response:** `X-Content-Type-Options: nosniff`; `X-XSS-Protection: 1; mode=block`; `Referrer-Policy: strict-origin-when-cross-origin`; `Permissions-Policy: geolocation=(), microphone=(), camera=()`; `X-Permitted-Cross-Domain-Policies: none`.
- **HSTS:** `Strict-Transport-Security: max-age=31536000; includeSubDomains` (no `preload`) — emitted **only** when `r.TLS != nil` or `X-Forwarded-Proto: https`. If the live check finds HSTS missing, the reverse proxy isn't forwarding `X-Forwarded-Proto`.
- **Framing/CSP:** non-embed routes: `X-Frame-Options: DENY` + `Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:`. `/embed/*` instead gets `frame-ancestors *` CSP and no XFO.
- **Cache-Control:** static template pages `public, max-age=3600`; `/public/images|icons|assets/*` → `public, max-age=2592000` (30 d); other `/public/*` → `public, max-age=3600`; `/premsa` & `/embed/leaderboard` → `public, max-age=300`; share/reveal/wrapped cards → `private, no-store`.
- **URLFormat quirk:** `middleware.URLFormat` strips extensions before routing; `/robots`, `/sitemap`, `/llms` are registered dotless so **both** `/robots.txt` and `/robots` (etc.) resolve. Historic bug: builds older than this fix 404 on `/robots.txt` — a key thing to verify live.

### 2.2 robots.txt (exact expected body)

```
User-agent: *
Allow: /
Sitemap: https://torro.cat/sitemap.xml
```

No AI-crawler blocks (deliberate: GPTBot/ClaudeBot/PerplexityBot all allowed via `*`).

### 2.3 sitemap.xml (expected structure)

Static entries: `/` (1.0), `/ranquing-de-torrons` (0.9), `/classes` (0.8), `/premsa` (0.5), `/advent` (0.5), `/sobre` (0.6), `/torro-agramunt-igp` (0.6), `/torro-agramunt-vs-xixona` (0.6), `/tipus-de-torrons` (0.6); plus every `/torro/{id}` (0.7) from DB and `/bracket/{classId}` (0.6) for classes with a bracket. **Note:** `/ranquing-de-torrons` only exists since commit `d70707d` — if the live sitemap lacks it, the deploy predates that commit.

### 2.4 llms.txt (expected)

Markdown map beginning `# Torrorèndum`, with the independence disclaimer blockquote ("independent fan project, not an official Torrons Vicens property"), Key pages list (Inici, Rànquing de torrons, Categories, Classificació, Premsa i dades, Advent, Sobre, IGP, Agramunt vs Xixona, Tipus de torrons) and "Notes for tools" section. Served `text/plain; charset=utf-8`. Parses cleanly as llms.txt (H1 + blockquote + H2 sections with link lists).

### 2.5 Per-page expected metadata (from public/templates/*.html)

All pages: `robots` meta `index, follow, max-image-preview:large` (public pages), `og:image` = `https://torro.cat/public/assets/og-image.jpg` (file present in repo, 43,192 bytes). `{{ seasonYear }}` renders dynamically (2026 for next season).

| Path | template | `<title>` (expected) | canonical | JSON-LD @types |
|---|---|---|---|---|
| `/` | index.html | `Torrorèndum {year} - Vota el millor torró` | `https://torro.cat` | WebSite, WebApplication, Brand |
| `/sobre` | about.html | `Torrorèndum · Sobre el projecte i preguntes freqüents` | `https://torro.cat/sobre` | FAQPage, Question, Answer, BreadcrumbList |
| `/premsa` | press.html | `Torrorèndum · Premsa i dades` | `https://torro.cat/premsa` | Dataset, Organization, BreadcrumbList |
| `/tipus-de-torrons` | glossari.html | `Torrorèndum · Glossari de torrons: tipus i varietats` | `https://torro.cat/tipus-de-torrons` | DefinedTermSet, DefinedTerm, BreadcrumbList |
| `/torro-agramunt-igp` | igp.html | `Torrorèndum · Què és la IGP Torró d'Agramunt?` | `https://torro.cat/torro-agramunt-igp` | Article, Organization, BreadcrumbList |
| `/torro-agramunt-vs-xixona` | comparativa.html | `Torrorèndum · Torró d'Agramunt vs Torró de Xixona` | `https://torro.cat/torro-agramunt-vs-xixona` | Article, Organization, BreadcrumbList |
| `/classes` | classes.html | `Categories - Torrorèndum {year}` | `https://torro.cat/classes` | BreadcrumbList |
| `/torro/{id}` | torro.html | `{Name} - Torrorèndum {year}` | `https://torro.cat/torro/{id}` | Product, Brand, BreadcrumbList |
| `/ranquing-de-torrons` | ranquing.html | `Rànquing de torrons {year} — Els millors torrons segons els vots` | `https://torro.cat/ranquing-de-torrons` | ItemList, BreadcrumbList |

Meta descriptions (expected, verbatim in templates): homepage "Participa al Torrorèndum {year}! Vota i escull els teus torrons favorits…"; sobre "Què és Torrorèndum, com funciona la votació i el rànquing ELO, i per què no és un projecte oficial de Torrons Vicens."; glossari "Glossari de torrons: què és el de Xixona, d'Agramunt, el dur, el guirlache o els panellets? …"; igp "Què significa la IGP Torró d'Agramunt: zona de producció, ingredients, percentatges mínims i la diferència legal amb la DOP."; comparativa "Comparativa neutra i documentada entre el torró d'Agramunt i el torró de Xixona: …".

Approx page weight (template bytes, pre-render, pre-gzip): index 23.5 KB; comparativa 16.2 KB; glossari 14.3 KB; press 13.9 KB; torro 12.4 KB; about 11.6 KB; igp 9.9 KB; ranquing 8.9 KB; classes 6.7 KB.

### 2.6 Known repo↔deploy discrepancy risks

1. **Deployed commit unknown.** `/ranquing-de-torrons`, dynamic season year, Product JSON-LD, and the dotless robots/sitemap/llms registration all landed in `d70707d`. If the production deploy predates it: `/ranquing-de-torrons` 404s, `/robots.txt`·`/sitemap.xml`·`/llms.txt` 404 (the pre-fix dead-code bug), and titles may show a stale hardcoded year. **First thing to check next run.**
2. No in-app apex/www or http→https redirects — entirely infra-dependent; behavior unrecorded.
3. chi does not redirect trailing slashes; `/sobre/` likely 404s unless the proxy normalizes.

---

## Snapshot 2026-08-17

Re-measurable observations. **Geo-bias caveat:** WebSearch is US-based; Catalan-market SERPs (google.es/google.cat, Catalan locale) may differ — treat presence/absence as directional, re-measure from the same tool for comparability.

### A. Network reachability (from this sandbox)

| Check | Result 2026-08-17 |
|---|---|
| `curl https://torro.cat/` via agent proxy | CONNECT 403 (proxy policy denial), 15:46:35Z |
| `curl https://www.torro.cat/` | CONNECT 403, 15:46:36Z |
| WebFetch torro.cat | EGRESS_BLOCKED |
| WebFetch en.wikipedia.org (control) | EGRESS_BLOCKED (blockage is blanket) |
| WebSearch | working |

### B. Brand-query visibility — **torro.cat is ABSENT from all results**

**Q1. `torrorendum torro.cat`** — top results in order: 1. diccionari.cat/GDLC/torro 2. bulbapedia.bulbagarden.net/wiki/Torracat_(Pokémon) 3. en.wikipedia.org/wiki/Torroella_de_Fluvià 4. en.wikipedia.org/wiki/Torroja_del_Priorat 5. tiktok.com/tag/torrocat 6. veekun.com/dex/pokemon/torracat 7. en.wikipedia.org/wiki/Torroella_de_Montgrí 8. en.wikipedia.org/wiki/Els_Torms 9. en.wikipedia.org/wiki/Arturo_Torró. → **torro.cat not present.**

**Q2. `"torrorèndum"` (quoted, accented)** — top results: 1. urbandictionary.com/define.php?term=torror 2. monstertruck.fandom.com/wiki/Torror 3. en.wiktionary.org/wiki/torror 4. en.wikipedia.org/wiki/TORRO 5. en.wikipedia.org/wiki/Tor 6. instagram.com/torror 7. soundcloud.com/torror 8. en.wikipedia.org/wiki/Terror_Universal 9. facebook.com/torror17 10. en.wikipedia.org/wiki/Holy_Terror. → **Zero relevant results; brand term unindexed.**

**Q3. `site:torro.cat`** — results: 1. en.wikipedia.org/wiki/Torroja_del_Priorat 2. en.wikipedia.org/wiki/Torrox 3. en.wikipedia.org/wiki/Torroella_de_Fluvià 4. en.wikipedia.org/wiki/Torroella_de_Montgrí 5. en.wikipedia.org/wiki/Torroella_de_Baix 6. sg.portal-pokemon.com/play/pokedex/0726 7. pokemongo.fandom.com/wiki/Torracat 8. amazon.com/toro-cat/s?k=toro+cat. → **No pages from torro.cat returned at all** (operator possibly unsupported by this search backend, but consistent with non-indexation).

**Q4. `torrorendum votar torró`** — results: 1. www2.tortosa.cat (inscripció votar) 2. en.wikipedia.org/wiki/Turrón 3. en.wikipedia.org/wiki/Torrox 4. en.wikipedia.org/wiki/Arturo_Torró 5. en.wikipedia.org/wiki/Torbe 6. en.wikipedia.org/wiki/Torroja_del_Priorat 7. en.wikipedia.org/wiki/Tor,_Pallars 8. en.wikipedia.org/wiki/Juan_Jose_Aizcorbe_Torra 9. andaluciainformacion.es (elecciones). → **torro.cat absent.**

### C. Target-keyword SERPs (who owns them today) — torro.cat absent from every one

**Q5. `torró d'Agramunt vs torró de Xixona diferències`** — 1. revista.consumer.es/ca/.../torrons-triar-be-no-nomes-depen-del-paladar.html 2. elpuntavui.cat/economia/article/18-economia/348972-dofici-torronaire.html 3. 11onze.cat/en/magazine/torro-sweet-long-tradition/ 4. igp-torrodagramunt.com/ca/.../historia-del-torro-dagramunt/41923.html 5. naciodigital.cat/noticia/38188/agramunt/menys/xixona 6. enciclopedia.cat/gran-enciclopedia-catalana/torro-1 7. ca.wikipedia.org/wiki/Torrons.

**Q6. `tipus de torrons glossari varietats`** — 1. barcelona.cat/culturapopular/ca/festes-i-tradicions/gastronomia/torrons 2. turronesydulces.com/cat/tipus-ametlles 3. festescatalunya.com/torrons-vicens/ 4. blocs.mesvilaweb.cat/jaumefabrega/torrons-torrons/ 5. vadegust.cat/reportatges/els-torrons-dorigen-catala-9488/ 6. molletama.cat (entrevista torronaire) 7–8. aplicacions.llengua.gencat.cat (Optimot, "Denominacions de torrons").

**Q7. `quin és el millor torró de Torrons Vicens rànquing`** — 1. thenewbarcelonapost.cat/millors-torrons-autor-2025/ 2. segre.com/ca/economia/250116/torrons-vicens-segona-gran-productora… 3. vicens.com/ca/blog/torrons-vicens-sorpren-aquest-any-amb-vuit-noves-varietats 4. gastrotalkers.cat/noticia/818/torrons-originals-2023… 5. escalabarcelona.com/2023/09/06/torrons-vicens… 6. tripadvisor.com (Torrons Vicens Petritxol) 7. lacasadeltorro.com/1288.html. → Search engine itself notes **no ranking of Vicens products exists** — exactly the gap /ranquing-de-torrons targets.

**Q8. `"IGP" "torró d'Agramunt" què és`** — 1. gastroteca.cat/en/productes-agroalimentaris/torro-dagramunt/ 2. ca.wikipedia.org/wiki/Torró_d'Agramunt 3. agramunt.cat/el-municipi/fires-i-festes/fira-del-torro 4. mapa.gob.es (pliego de condiciones PDF) 5. federaciodopigp.cat/en/igp/torro-dagramunt 6. agricultura.gencat.cat (fitxa IGP) 7. radiotarrega.cat 8. igp-torrodagramunt.com 9. gastronomia.aralleida.com 10. elnacional.cat/ca/gourmeteria/.../934107_102.html.

### D. Verdicts (2026-08-17)

1. **torro.cat is not visible in any search result observed today** — neither brand queries nor target keyword queries. Consistent with a very young/unindexed site (and/or the deployed build's historical robots/sitemap 404 bug having delayed crawling).
2. **Brand SERP for "torrorèndum" is completely empty** — an easy first win once indexed; also currently zero risk of brand confusion, but note Q1 surfaces `diccionari.cat`'s "torró" entry, i.e. the .cat "torro" namespace is dictionary-dominated.
3. **Keyword incumbents** are institutional (gencat, enciclopedia.cat, Wikipedia, IGP consell regulador, barcelona.cat) and media (consumer.es, naciodigital, elpuntavui) — no interactive/ranking competitor exists in the observed SERPs.

### E. Re-measure protocol for the next run

1. Run from a network that can reach torro.cat; complete Section 1's checklist with `curl -sS -D- -o page.html -w '%{http_code} %{time_total} %{size_download}'` per URL; diff against Section 2 expected values.
2. Re-run Q1–Q8 verbatim on the same search tool; diff ordered result lists against Section B/C.
3. Confirm sitemap contains `/ranquing-de-torrons` (deploy freshness signal) and that `/robots.txt` returns 200 (not the pre-`d70707d` 404).
4. Check Google Search Console / `site:torro.cat` on google.es directly for indexation counts.

---

*Sources for SERP claims: the WebSearch result lists reproduced verbatim above (queries Q1–Q8, run 2026-08-17). Repo claims: `/home/user/torrons/internal/http/server.go`, `internal/http/seo_handler.go`, `internal/http/middleware.go`, `public/templates/*.html`, `public/assets/og-image.jpg` at commit `1fc5890`.*
