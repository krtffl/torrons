# SEO & AI-Positioning Strategy — Torrorèndum (torro.cat)

**Goal:** #1 (or the cited AI-assistant source) for every search a user might make
about which torró is best, torró rankings, Torrons Vicens product evaluation, and the
supporting informational cluster — in Catalan first, Spanish second, English long-tail
third.

**Living document.** Created 2026-08-17 from the baseline research
(`BASELINE_2026-08-17.md`, `research/*.md`). Update checkboxes and re-prioritize on
every check run (`CHECK_PROTOCOL.md`).

---

## 1. Positioning

**The one thing nobody else has:** a persistent, independent, transparent,
continuously-updated dataset of real people's torró preferences (head-to-head duels,
ELO, 107 products). Every competitor is static, annual, paywalled, self-interested,
or ephemeral. So:

> **Torrorèndum is the people's torró ranking** — "els experts caten; el poble vota."
> OCU is the lab; we are the crowd. Vicens is the maker; we are the independent fans.
> Free, open (CC BY 4.0 stats), methodology in public, updated live.

Guardrails that are also ranking assets:
- The **independence disclosure** ("projecte de fans independent, no oficial") stays
  on every page, in Organization/WebSite schema, in llms.txt and in press copy. It is
  the legal shield *and* the E-E-A-T differentiator vs affiliate rankings — and it
  must be machine-readable so AI answers describe us correctly.
- **Never** target the bare brand SERP's transactional intent; always link official
  purchase channels; target evaluative modifiers ("quin és millor", "rànquing",
  "opinions") which are empty.
- **No aggregateRating/star schema from ELO** (policy + misrepresentation risk). If a
  real 1–5 rating feature ever ships, revisit.
- **Product-photo copyright is an open legal question** (photos are presumably Vicens
  imagery) — resolve before the PR push makes the site conspicuous
  (`gaps-and-contradictions.md` G8).

## 2. The five pillars

### Pillar 1 — Exist in the indexes (P0, mostly operator actions)

The site is invisible everywhere. Nothing else pays off until this clears.

- [ ] **Verify the deployed build** serves this branch (live sitemap contains
      `/ranquing-de-torrons`; `/robots.txt` 200). If not → deploy.
- [ ] **Google Search Console**: DNS-verify torro.cat, submit sitemap, request
      indexing of the ~12 key pages, watch Page-indexing report until >100 URLs
      indexed.
- [ ] **Bing Webmaster Tools** (feeds ChatGPT/Copilot answers): verify, submit
      sitemap. **IndexNow**: generate key, serve `{key}.txt`, ping on content
      updates (small Go addition — roadmap item C-7).
- [ ] **Brave Search** (feeds Claude): check `site:torro.cat`, submit if absent.
- [ ] **Infra hygiene** (live checks were sandbox-blocked): http→https 301,
      www→apex 301, `X-Forwarded-Proto` reaching the app (HSTS depends on it), no
      CDN/WAF "block AI bots" toggle undoing our robots policy.
- [ ] **Analytics**: install a privacy-respecting analytics with an AI-referrer
      channel (chatgpt.com, perplexity.ai, claude.ai, gemini.google.com, copilot…);
      confirm server-log access for AI-crawler UA counts. Without this, the entire
      measurement plan has no substrate.
- [ ] **Claim social handles** (@torrorendum on IG/TikTok/X/YouTube minimum) and add
      them as `sameAs` in Organization schema — entity corroboration + platform
      search presence.

### Pillar 2 — Technical excellence on-site (code; largely DONE this run)

Done 2026-08-17: public ranking page (`/ranquing-de-torrons`), dynamic season year,
Product+ItemList+dated-Article+licensed-Dataset schema, 19→3 MB images, self-hosted
htmx, explicit AI-crawler robots.txt, error/404 pages, trailing-slash 301s, full
icon/manifest set, per-product og:images, X-Robots-Tag on personal PNGs.

Remaining backlog (ordered):
- [ ] **C-1** Organization JSON-LD sitewide (name, logo, description with the
      independence disclosure verbatim, `sameAs` once handles exist) — currently only
      WebSite+WebApplication on the homepage carry entity data.
- [ ] **C-2** Quotable first-paragraph answers (40–80 words, number + date) on every
      money page — partially done on /ranquing-de-torrons; extend to /sobre, IGP,
      comparativa, glossari (GEO rec 7).
- [ ] **C-3** Honest `<lastmod>` in the sitemap for the ranking page (from latest
      Result timestamp) and content pages (from a maintained constant) — Google uses
      honest lastmod for crawl scheduling; fake dates are worse than none.
- [ ] **C-4** Self-host the two Google Fonts families (removes the last
      render-blocking third-party + GDPR-gray dependency).
- [ ] **C-5** Visible breadcrumb UI matching the BreadcrumbList markup.
- [ ] **C-6** `web-vitals` INP/LCP field measurement on the vote flow (INP is the
      at-risk metric: click → ELO write → next-duel render must stay <200 ms).
- [ ] **C-7** IndexNow ping hook (daily or on ranking-order change).
- [ ] **C-8** htmx 1.9.12/2.x upgrade evaluation (currently pinned 1.9.9).

### Pillar 3 — Content architecture (the page plan, from `research/keyword-universe.md`)

Shipped: ranking page (CA), IGP explainer, Agramunt-vs-Xixona, glossary, FAQ/about,
press/data page, 107 product pages, advent page (retitled).

Build order (each page: exact-match title ≤60 chars = H1, quotable lead with numbers,
FAQ block, interlinked hub-and-spoke):

- [ ] **P0-a** "Quin és el millor torró?" evergreen answer page fed by live vote data
      (or make /ranquing-de-torrons carry the question explicitly in an H2+lead).
- [ ] **P0-b** "Els torrons Vicens més votats" guide (zero competition; disclosure
      prominent).
- [ ] **P0-c** Spanish twin of the IGP explainer: "Turrón de Agramunt: qué es, IGP,
      diferencias con Jijona" — the highest-affinity ES gap (no editorial reference
      exists; TasteAtlas ranks #1). First page of the `/es/` subtree.
- [ ] **P0-d** ES ranking landing "Ranking de turrones (por votos)".
- [ ] **P1-a** Seasonal hub "Torrons Nadal 2026: novetats i resultats" — **publish
      by early October**; freeze a "Classificació Nadal 2026" edition in January
      (year-labeled URLs are how every incumbent wins December queries).
- [ ] **P1-b** Per-category best pages: xocolata, crema cremada (the most-Catalan
      variety — unowned), festuc/pistacho (rising), praliné, sense sucre.
- [ ] **P1-c** FAQ enrichment on /sobre or glossary: calories per type, caducitat/
      conservació (15–18 months — clean featured-snippet shape), quin torró és típic
      de Catalunya, diabetics/gluten (cite sources, no medical advice).
- [ ] **P1-d** Adrià Natura line hub ("Albert Adrià × Vicens: la línia Natura,
      rànquing per vots") — press gravity, no independent review exists. Use the
      current line name "Adrià Natura"/"Natura", not the obsolete "Essència".
- [ ] **P1-e** Fira del Torró d'Agramunt 2026 section on the IGP page (dates
      published ~Sept; reliable pre-season traffic).
- [ ] **P2** English one-pager (what is torró, turron vs torrone, "best turron in
      Barcelona — what locals voted"); cata-at-home guide funneling into the app;
      history explainer. When `/es/` ships: full bidirectional ca/es/x-default
      hreflang per `research/tech-seo-2026.md` §10.

### Pillar 4 — Authority: links, press, entities (operator + content)

The site has ~zero links; brand-search volume correlates with AI citations more than
backlinks — so PR is the lever, not link-begging.

- [ ] **Backlink/PR research track** (the one missing research report — G1): named
      journalists/sections at elnacional.cat Gourmeteria, NacióDigital Viure bé,
      mengem.ara.cat, betevé, vadegust, catalunyapress, TNBP; directory/resource
      targets (.cat ecosystem, gastroteca).
- [ ] **December data-story pitch** ("X mil vots: aquest és el millor torró segons
      els catalans") timed to the OCU-echo week (~Dec 17–19) — the outlets that
      rewrite OCU need a fresh angle every year; we are the only alternative dataset.
      Criterion: ≥2 independent Catalan articles naming Torrorèndum before Dec 1
      (stretch: before reveal Jan 6).
- [ ] **Embed program**: the `/embed/leaderboard` widget offered to bloggers/media
      (every embed = a followed link from a ranking article). Add an "embed this"
      snippet box on /premsa (partially exists — verify).
- [ ] **Wikipedia/Wikidata the legitimate way** (COI-safe): after press coverage
      exists, torro.cat data as a *reference* in Viquipèdia "Torró"/"Torró
      d'Agramunt" via talk pages; third-party-created Wikidata item. Never
      self-insert.
- [ ] **Reddit/forums/platform search** (G2/G3): decide presence on r/catalunya,
      TikTok/IG (where "ranking turrones" demand demonstrably lives) — accounts
      posting seasonal results clips.

### Pillar 5 — AI/GEO specifics (mostly satisfied by pillars 1–3; deltas:)

- [ ] **Branded statistics**: /premsa offers ≥5 copy-paste citable claims embedding
      the name ("Segons el Torrorèndum (torro.cat), amb N vots…") — paraphrase then
      carries attribution even without a link.
- [ ] **Query fan-out coverage**: ≥8 sub-question H2/H3 blocks across the cluster
      (quin torró té menys sucre, dur o tou, per regalar…), each self-contained.
- [ ] **Monthly prompt panel** (from `research/tech-seo-2026.md` end): 10 fixed
      questions across ChatGPT/Perplexity/Gemini/Claude, log citations. First
      baseline row: due ≤2026-09-30.
- llms.txt: shipped; zero measured impact industry-wide — keep, don't invest more.

## 3. Season calendar (Aug 2026 → Jan 2027)

| When | What |
|---|---|
| **Now–Aug 31** | Pillar-1 operator actions (deploy check, GSC/Bing/Brave, analytics, handles). Legal check on product photos (G8). Verify OCU verdict details (C1/C2) before quoting anywhere. |
| **September** | P0 content pages (a–d); C-1..C-3 technical items; Fira d'Agramunt 2026 content when dates publish; first monthly prompt-panel run; Spain-geolocated SERP spot-check (G7). |
| **Early October** | "Torrons Nadal 2026" hub + novetats coverage (Vicens announces new products ~Sept/Oct); per-category best pages; PR list ready. |
| **November** | Season opens: freshness updates on all money pages; press outreach round 1 (novetats angle); advent page pre-launch; **full check run** (CHECK_PROTOCOL) — first in-season snapshot. |
| **Dec 1–24** | Advent duels live (daily social clips); mid-December data-story pitch timed against the OCU echo week; monitor Discover. |
| **Jan 6** | Reveal/Gran Final press push ("el guanyador del Torrorèndum 2026"); freeze "Classificació Nadal 2026" edition page. |
| **Late Jan** | Post-season check run; year-over-year diff; update this doc. |

## 4. What we will NOT do

- Chase `mejor turrón del supermercado` / bare `torrons Vicens` / transactional
  queries (wrong catalog, wrong intent, unbeatable incumbents).
- Emit review/star schema from ELO, fake freshness dates, or buy links.
- Publish OCU-derivative rewrites (commodity content that undermines the
  original-data positioning).
- Paid search (non-commercial fan project; out of scope — recorded per G-review).

## 5. Success metrics

| Horizon | Metric |
|---|---|
| Sept 2026 | Indexed in Google/Bing/Brave; #1 for `torrorèndum`; GSC/BWT live; analytics live |
| Nov 2026 | Top-10 for `rànquing torrons`, `tipus de torrons`, `millors torrons Vicens`; first AI-assistant citation observed |
| Jan 2027 | Top-3 for the Catalan P0 cluster; ≥2 press mentions with links; torro.cat cited by ≥1 assistant for "quin és el millor torró"; CWV all-green field data |
| Ongoing | Every check run diffs green vs `BASELINE_2026-08-17.md`; zero policy/legal incidents |
