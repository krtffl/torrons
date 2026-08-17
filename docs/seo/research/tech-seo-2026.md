# Technical & On-Page SEO Standard 2026 — Definitive Checklist for Torrorèndum (torro.cat)

**Research date:** 2026-08-17 · **Researcher:** Claude (subagent, best-practices track)
**Scope:** the current (mid-2026) technical/on-page SEO standard from authoritative sources, mapped to a Catalan-language, independent, non-commercial fan voting/ranking site for Torrons Vicens products.

**Method caveats:**
- `WebSearch` used here is **US-geo-based**; result orderings may differ from google.es/google.cat. Spanish/Catalan-language queries were still run where relevant; absolute policy/documentation facts are geo-independent.
- Direct fetches of `developers.google.com` and `web.dev` were **blocked by the sandbox egress proxy** in this environment, so Google-doc claims are verified via search-result snippets of the official docs plus reputable secondary sources (Search Engine Land, Search Engine Journal, seroundtable, etc.). Official URLs are cited as the canonical reference; a future run with direct access should re-verify against them.

---

## 0. Executive framing

Two audiences now matter equally: **classic Google/Bing rankings** and **AI answer engines** (ChatGPT/OAI-SearchBot, Perplexity, Gemini/AI Overviews & AI Mode, Claude). The 2026 standard is: fast (CWV incl. INP), crawlable server-rendered HTML, ruthless structured-data hygiene (several types were killed in 2023–2026), demonstrable E-E-A-T with transparent methodology, correct ca/es hreflang if bilingual, and explicit AI-crawler policy in robots.txt. The single biggest 2026 policy change: **FAQ rich results are fully dead as of May 7, 2026** — FAQPage markup no longer produces any rich result for anyone.

---

## 1. Core Web Vitals & page experience

**Standard (unchanged thresholds, INP replaced FID in March 2024):**

| Metric | Good | Poor | Measured at |
|---|---|---|---|
| LCP (Largest Contentful Paint) | ≤ 2.5 s | > 4.0 s | 75th percentile, CrUX 28-day field data |
| INP (Interaction to Next Paint) | ≤ 200 ms | > 500 ms | 75th percentile |
| CLS (Cumulative Layout Shift) | ≤ 0.1 | > 0.25 | 75th percentile |

- A page "passes" a metric only when ≥75% of real page views are in the Good bucket (CrUX, 28-day rolling window). Sources: https://web.dev/articles/inp , https://web.dev/articles/vitals , corroborated by https://www.corewebvitals.io/core-web-vitals and https://meteoraweb.com/en/analisi-dei-dati-e-metriche/core-web-vitals-2026-lcp-inp-cls-thresholds-and-seo-impact (observed 2026-08-17).
- CWV is a real but **modest tie-breaker** ranking signal; it matters more for Discover and for AI search crawlers (real-time search bots time out on slow pages — see §15).
- **INP is the metric most at risk for Torrorèndum**: the duel-voting interaction (click → ELO update → next duel render) is exactly what INP measures. Keep main-thread work per vote < 200 ms; optimistic UI + deferred network write is the standard pattern.

**Pass/fail criteria:**
- [ ] PageSpeed Insights field data (or CrUX API) shows all three metrics Good on mobile for the homepage, /torro/{id} template, and ranking pages.
- [ ] Lab check: Lighthouse mobile performance ≥ 90; TBT < 200 ms.
- [ ] Voting click-to-paint (measured with the web-vitals JS library `onINP`) ≤ 200 ms at p75.
- [ ] No layout shift when duel images/results load (explicit width/height or aspect-ratio on all images; reserve space for result bars).

## 2. Mobile-first indexing

Mobile-first indexing has been complete since Oct 2023 — Googlebot Smartphone is effectively the only indexer. Source: https://developers.google.com/search/blog/2023/10/mobile-first-is-here (canonical); Google Search Central documentation.

**Pass/fail:**
- [ ] Identical content, structured data, and meta robots on mobile and desktop (responsive design = automatic pass).
- [ ] Viewport meta present; tap targets ≥ 48px; no horizontal scroll.
- [ ] `curl -A "Googlebot-Smartphone UA"` returns full content server-side (or verified working rendering).

## 3. HTTPS & security headers

- HTTPS is a confirmed (lightweight) ranking signal since 2014 and part of the page-experience assessment. Source: https://developers.google.com/search/blog/2014/08/https-as-ranking-signal ; corroborated https://websentry.dev/blog/does-https-affect-seo-rankings-what-google-actually-rewards/ .
- **Security headers (HSTS, CSP, X-Content-Type-Options) are NOT direct ranking factors** — John Mueller confirmed this explicitly. Source: https://www.searchenginejournal.com/security-headers-and-ranking-influence/488781/ . They remain best practice for trust/E-E-A-T and eliminate the HTTP→HTTPS redirect round-trip (HSTS helps LCP on repeat visits).

**Pass/fail:**
- [ ] All URLs 301 → https; zero mixed content.
- [ ] `Strict-Transport-Security: max-age=31536000; includeSubDomains` present.
- [ ] `X-Content-Type-Options: nosniff`, sensible `Content-Security-Policy`, `Referrer-Policy` present (hygiene, not ranking).
- [ ] Valid cert, TLS ≥ 1.2.

## 4. Crawling & indexing

### 4.1 robots.txt
- Keep robots.txt for **crawl control only**, never for de-indexing (a blocked URL can still be indexed URL-only). Canonical doc: https://developers.google.com/search/docs/crawling-indexing/robots/intro .
- **Never combine robots.txt disallow with noindex on the same URL** — Google can't see the noindex if blocked. Source: Google "Crawling December: faceted navigation" https://developers.google.com/search/blog/2024/12/crawling-december-faceted-nav (corroborated by https://ppc.land/managing-faceted-navigation-urls-new-google-documentation-2/ ).

### 4.2 Sitemaps & lastmod
- The sitemap **ping endpoint is dead** (deprecated June 2023, 404 since); submit via Search Console + `Sitemap:` line in robots.txt. Source: https://developers.google.com/search/blog/2023/06/sitemaps-lastmod-ping .
- **`lastmod` is actively used by Google for crawl scheduling — but only if it's honest.** Update it only when primary content, structured data, or links materially change; not for footer/sidebar churn. Same source. `changefreq` and `priority` are ignored by Google.
- For Torrorèndum: ELO ranks changing daily arguably changes primary content of ranking pages — set lastmod on true content change of the page (rank order changed), not on every vote.

**Pass/fail:**
- [ ] `/sitemap.xml` valid, ≤ 50k URLs, referenced in robots.txt, submitted in GSC + Bing WMT.
- [ ] Every `<lastmod>` matches a real, material content change (spot-check 5 URLs vs. actual diff).
- [ ] No pinging code; no changefreq/priority reliance.

### 4.3 Canonicals
- One self-referencing `rel=canonical` per indexable URL; absolute URLs; consistent with sitemap and internal links (canonical doc: https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls ).
- Canonicals are hints, not directives — conflicting signals (canonical → A, internal links → B) get overridden.

**Pass/fail:**
- [ ] Each page has exactly one canonical; it equals the sitemap URL and the internally-linked URL (protocol, host, trailing-slash consistent).
- [ ] Parameterized/share-state URLs (e.g. `?duel=`, UTM) canonicalize to the clean URL.

### 4.4 Faceted/parameterized URLs (duel permalinks, filters, sort orders)
Google's Dec 2024 guidance (the current reference — URL Parameters tool is long gone): https://developers.google.com/search/blog/2024/12/crawling-december-faceted-nav
- Zero-value parameters (sort, view, session, individual duel states) → **disallow in robots.txt** (crawl-budget lever).
- Facets with search value → keep crawlable, consistent parameter order, `&` separator, canonical to the main page if content is duplicative, and return 404 for empty result combinations.
- Alternatively `noindex,follow` for thin combination pages — but then do NOT robots-block them.

**Pass/fail:**
- [ ] Inventory of every URL parameter torro.cat can emit, each assigned exactly one strategy (robots-block / canonical / noindex / indexable).
- [ ] GSC Crawl Stats shows < 10% of crawl on parameterized URLs.

### 4.5 Pagination
- `rel=prev/next` has been ignored by Google since 2019. Standard: each page self-canonical, unique title ("… — pàgina 2"), linked with plain `<a href>`; don't canonicalize page 2+ to page 1. Source: https://developers.google.com/search/docs/specialty/ecommerce/pagination-and-incremental-page-loading .
- For a ~50–100-product ranking, prefer **one single unpaginated ranking page** — best for users and for AI-engine quotability.

### 4.6 Rendering
- Googlebot renders JS, but AI-search crawlers (OAI-SearchBot, PerplexityBot, ClaudeBot) largely **do not execute JavaScript**. SSR/SSG of all content pages and rankings is mandatory for the AI-citation goal. (Corroborated across https://www.frase.io/blog/how-to-get-cited-by-ai-search-engines-the-complete-geo-playbook and vendor docs, 2026-08-17.)

**Pass/fail:**
- [ ] `curl` (no JS) of every indexable template returns full ranking/content HTML including product names and rank numbers.

## 5. Structured data — what's alive, what's dead, what fits a voting site (state: 2026-08-17)

Canonical gallery: https://developers.google.com/search/docs/appearance/structured-data/search-gallery · policies: https://developers.google.com/search/docs/appearance/structured-data/sd-policies

### 5.1 Deprecation timeline (must-know)
| Feature | Status | Date | Source |
|---|---|---|---|
| HowTo rich results | **Dead** (desktop+mobile) | Sept 2023 | https://developers.google.com/search/blog/2023/08/howto-faq-changes |
| FAQPage rich results | Restricted to well-known **government/health** sites only | Aug 2023 | same |
| FAQPage rich results | **Fully dead for everyone** — no SERP display since May 7 2026; RRT/report support removed June 2026; Search Console API support ends Aug 2026 | May–Aug 2026 | https://www.searchenginejournal.com/google-drops-faq-rich-results-from-search/574429/ (doc-note change, no blog post) |
| Sitelinks search box (`WebSite`→`potentialAction: SearchAction`) | **Dead** | Nov 21 2024 | https://www.searchenginejournal.com/google-removes-sitelinks-search-box-documentation/533973/ |
| Book Actions, Course Info, Claim Review, Estimated Salary, Learning Video, Special Announcement, Vehicle Listing | **Dead** (7 types) | June 2025 (reports removed Sept 8 2025) | https://developers.google.com/search/blog/2025/06/simplifying-search-results ; https://searchengineland.com/google-to-drop-support-for-several-rich-result-types-to-simplify-the-search-results-page-456969 |

Leftover dead markup causes no penalty, but ship no new dead markup. FAQPage JSON-LD may still aid AI-engine comprehension — keep the visible Q&A content, drop expectations of rich results.

### 5.2 Types that fit Torrorèndum, with policy notes
- **Organization** (site publisher, on every page): name "Torrorèndum", `url`, `logo`, `sameAs`, and crucially a `description` stating independence/fan status. **Must NOT impersonate Torrons Vicens** — sd-policies prohibit misleading markup; the legal disclosure (unofficial fan project) must be consistent in schema. Doc: https://developers.google.com/search/docs/appearance/structured-data/organization
- **WebSite** (site name in SERP): keep `WebSite` with `name` + `alternateName`; site-name feature is alive (only SearchAction is dead). Doc: https://developers.google.com/search/docs/appearance/site-names
- **BreadcrumbList**: fully supported, low effort, do it on /torro/{id} and content pages. Doc: https://developers.google.com/search/docs/appearance/structured-data/breadcrumb
- **ItemList** on ranking pages: valid schema.org, but Google carousels from ItemList are limited to specific host types (Recipe, Course, Movie, Restaurant, and the newer travel/product-variant carousels beta) — a torró ELO ranking gets **no rich result** from ItemList. Still worth shipping: it is machine-readable ranking data that AI engines parse. Doc: https://developers.google.com/search/docs/appearance/structured-data/carousel
- **Product + AggregateRating — HANDLE WITH CARE.** Policies (review snippet doc: https://developers.google.com/search/docs/appearance/structured-data/review-snippet ):
  - *Self-serving ban*: ratings of an entity placed on that entity's own site are ineligible. Torrorèndum reviews **third-party** products (Torrons Vicens's), so it is **not self-serving** — an independent aggregator is actually the intended use case.
  - *But*: "ratings must be sourced directly from users" and must refer to a genuine rating scheme. **ELO scores from pairwise duels are not user ratings on a bounded scale.** Mapping ELO→1–5 stars synthetically risks violating the "misleading/fabricated ratings" policy and, worse, could look like fake official product ratings — a legal/brand risk for an unofficial fan site. **Recommendation: do NOT emit `aggregateRating` from ELO.** If a separate explicit 1–5 user-rating feature is ever added, aggregateRating becomes legitimate.
  - Safe pattern for /torro/{id}: `Product` (or `Product`-free `WebPage`+`about`) with name, image, brand ("Torrons Vicens", factual nominative use), description including duel stats as text — **no `offers`** (not selling), no `review`/`aggregateRating`.
- **Dataset** on /premsa: the ELO/vote dataset is a genuine dataset; Dataset markup is supported (Dataset Search) and reinforces "cited source" positioning for journalists and AI engines. Doc: https://developers.google.com/search/docs/appearance/structured-data/dataset
- **Article/NewsArticle** on explainer pages (IGP, Agramunt vs Xixona) with `author` (Person), `datePublished`, `dateModified`. Doc: https://developers.google.com/search/docs/appearance/structured-data/article
- **VideoObject**: only worth it if a video is the **main content of a dedicated watch page** — since 2023–24 Google only indexes/thumbnails videos on watch pages; embedded supplementary videos are ignored. Sources: https://searchengineland.com/google-video-must-be-main-content-to-appear-as-thumbnail-395597 , https://searchengineland.com/google-expands-video-requirements-for-video-mode-where-video-must-be-main-content-of-the-page-435382 . Verdict for torro.cat: **skip** unless a video hub is planned.
- **Speakable**: still **beta**, US-English + Google Home only — useless for a Catalan site. Doc: https://developers.google.com/search/docs/appearance/structured-data/speakable . Verdict: skip.
- **FAQPage** on /sobre: no rich result anymore (see 5.1). Keep visible FAQs (great for AI answers); markup optional/harmless.

**Pass/fail:**
- [ ] Rich Results Test: 0 errors on every template; only living types emitted.
- [ ] No `aggregateRating`/star markup derived from ELO anywhere.
- [ ] Organization schema states independent/fan status consistently with the visible disclaimer.
- [ ] JSON-LD (Google's preferred format) server-rendered in initial HTML.

## 6. E-E-A-T for a fan project

E-E-A-T (Experience added Dec 2022) is a quality-rater rubric, not a direct ranking factor, but proxies for it feed ranking systems and AI-engine source selection; **Trust is the dominant member**. Sources: Google Search Quality Rater Guidelines (https://static.googleusercontent.com/media/guidelines.raterhub.com/en//searchqualityevaluatorguidelines.pdf ), https://developers.google.com/search/docs/fundamentals/creating-helpful-content , https://www.searchenginejournal.com/google-e-e-a-t-how-to-demonstrate-first-hand-experience/474446/ .

For an independent fan site the winning levers are:
- **Experience**: first-hand tasting/product familiarity, photos of actual products, vote-count receipts ("N votes cast by M visitors").
- **Methodology transparency**: a public "Com funciona el rànquing" page — ELO formula, K-factor, sample sizes, anti-abuse measures, data update cadence. This is the strongest differentiator vs. generic listicles and exactly what AI engines quote.
- **Author/about**: named human(s) on /sobre with bio and contact; `Person` schema; contact email; the independence disclaimer (which is itself a trust signal — do not bury it).
- **Citations out**: link IGP/DO official sources (e.g. IGP Torró d'Agramunt council, DOGC), Torrons Vicens official site clearly labeled as the official brand site — outbound citation is an E-E-A-T and GEO best practice.
- **Citations in**: /premsa with citable stats + Dataset markup is the earned-media engine.

**Pass/fail:**
- [ ] Public methodology page exists, linked from footer and every ranking page.
- [ ] Named author/maintainer with contact info; visible "última actualització" and vote counts.
- [ ] Independence disclaimer visible sitewide + in Organization schema + in meta description of home.
- [ ] ≥ 3 authoritative outbound citations per explainer page.

## 7. Freshness & dates

- Google shows dates from visible on-page dates + `datePublished`/`dateModified` (must agree). Doc: https://developers.google.com/search/docs/appearance/publication-dates . Do not fake-freshen: changing dates without content changes is against guidance and erodes trust.
- Perplexity and ChatGPT search **strongly prefer recently-updated content** — visible "Actualitzat: {date}" + honest `dateModified` + honest sitemap lastmod is a triple win. (GEO sources §15.)
- Seasonal reality: refresh all money pages Oct–Nov 2026 ahead of the Nadal peak; add year-scoped freshness ("Rànquing de torrons 2026") where honest.

**Pass/fail:**
- [ ] Visible updated-date on rankings & explainers matches `dateModified` matches sitemap lastmod.
- [ ] All evergreen pages get a real content review + update before Nov 1, 2026.

## 8. Internal linking architecture

Standard (Google SEO starter guide + link best practices, https://developers.google.com/search/docs/crawling-indexing/links-crawlable ):
- Every indexable page ≤ 3 clicks from home; crawlable `<a href>` (not onclick/button navigation — important for the SPA-style duel UI).
- Descriptive anchor text ("torró de praliné de Vicens", not "veure més").
- Hub-and-spoke: ranking pages ↔ /torro/{id} ↔ explainers (/torro-agramunt-igp etc.) ↔ glossary terms deep-linked from product descriptions.
- No orphan pages; breadcrumbs on all sub-pages.

**Pass/fail:**
- [ ] Crawl (Screaming Frog / custom) shows 0 orphans, max depth 3, all nav links are real `<a href>`.
- [ ] Each /torro/{id} links: its ranking position page, ≥2 related products, ≥1 glossary/explainer.

## 9. Image SEO

Doc: https://developers.google.com/search/docs/appearance/google-images
- Descriptive `alt` in Catalan (product name + type, e.g. "Torró d'Agramunt IGP de Torrons Vicens"); descriptive filenames.
- `srcset`+`sizes` responsive images; **WebP/AVIF both fully supported by Google Images**; serve AVIF with WebP/JPEG fallback via `<picture>`.
- Native `loading="lazy"` below the fold — but **never lazy-load the LCP image**; give it `fetchpriority="high"`.
- Explicit dimensions everywhere (CLS).
- `max-image-preview:large` robots meta (also a Discover prerequisite, §13).
- Image sitemap or images in the main sitemap for product shots.

**Pass/fail:**
- [ ] 100% images have alt; product images ≥ 1200 px wide available.
- [ ] LCP image not lazy; `fetchpriority=high`; modern format with fallback.
- [ ] `<meta name="robots" content="max-image-preview:large">` sitewide.

## 10. i18n strategy for ca/es

Docs: https://developers.google.com/search/docs/specialty/international/localized-versions , https://developers.google.com/search/blog/2013/04/x-default-hreflang-for-international-pages
- **URL strategy: subdirectories, not parameters.** Catalan stays at root (`torro.cat/...`) as the primary/original; Spanish at `torro.cat/es/...`. Parameters (`?lang=es`) are explicitly discouraged for localized content; separate ccTLD unnecessary.
- **hreflang pairs, bidirectional, per page**: `ca`, `es`, and `x-default` → the Catalan version (it's the flagship + language-selection default). Use `hreflang="ca"`/`"es"` (ISO 639-1; region codes like `es-ES` only if a Latin-American variant ever exists). One missing return tag invalidates the cluster (75% of implementations have errors — https://www.linkgraph.com/blog/hreflang-implementation-guide/ ).
- Each language version: self-canonical (never canonical cross-language), fully translated (title/meta/schema/alt too), interlinked via visible language switcher with crawlable links.
- hreflang can live in `<head>` or sitemap — pick ONE mechanism to avoid conflicts.
- Note: hreflang is about serving the right version, not a ranking boost; a Spanish version's real win is capturing "turrón" (much higher volume than "torró") queries and AI answers in Spanish.

**Pass/fail:**
- [ ] Every ca page emits `ca` + `es` + `x-default` hreflang; every es page emits the mirror set (validated with hreflang checker, 0 return-tag errors).
- [ ] `x-default` = Catalan URL; no cross-language canonicals; no `?lang=` URLs indexable.

## 11. Titles & meta 2026

Docs: https://developers.google.com/search/docs/appearance/title-link , https://developers.google.com/search/docs/appearance/snippet
- Google rewrites titles ~76% and descriptions ~62–70% of the time (Q1 2025 studies: https://destination-digital.co.uk/news-blogs-case-studies/title-meta-description-length-google-serps-2025/ ). Rewrite triggers: boilerplate, keyword stuffing, title≠H1≠content mismatch. Defense: **one accurate, unique title per page that matches the H1 and the dominant query**, key info in the first ~50 characters.
- Practical limits: title ≈ ≤ 580–600 px (~55–60 chars); description ≈ 150–158 chars. These are display limits, not ranking limits.
- Pattern for torro.cat: `{Primary term} — {value} | Torrorèndum` e.g. "Rànquing de torrons Vicens 2026 — votat en X duels | Torrorèndum". Brand last; year only where honestly maintained.
- Meta description: front-load the answer + the independence/fan angle (unique selling point + legal disclosure in one move).

**Pass/fail:**
- [ ] 0 duplicate titles/descriptions (crawl check); every title ≤ 60 chars matches its H1 semantically.
- [ ] GSC check (post-launch): sampled SERP titles = authored titles for ≥ 70% of top pages.

## 12. Favicon, branding, sitelinks

- Favicon doc (updated Oct 2024): square, **min 8×8 px but ≥ 48×48 recommended** (the old 48px-multiple rule was relaxed), stable URL, representative of brand. Source: https://developers.google.com/search/docs/appearance/favicon-in-search , https://www.searchenginejournal.com/google-now-recommends-higher-resolution-favicons/530793/ .
- Site name in SERP: controlled by `WebSite` schema `name` (§5.2) + consistent branding; works for domain-level.
- Sitelinks: not directly controllable; earned via clear hierarchy, crawlable nav, descriptive anchors, breadcrumbs. Doc: https://developers.google.com/search/docs/appearance/sitelinks .

**Pass/fail:**
- [ ] 96×96 (or larger) square favicon at a stable URL, referenced with `<link rel="icon">`; renders in google SERP within weeks of indexing.
- [ ] Brand SERP for "torrorèndum" shows correct site name + sitelinks to key sections (re-check quarterly).

## 13. Google Discover

Doc: https://developers.google.com/search/docs/appearance/google-discover
- Eligibility is automatic: indexed + policy-compliant; **no special feed or tag**. Practical requirements: images ≥ **1200 px** wide + `max-image-preview:large`; compelling non-clickbait titles; E-E-A-T; freshness. (Corroborated: https://www.seoworks.co.uk/google-discover-guide/ .)
- Discover is entity/interest driven and seasonal — torró content realistically surfaces Nov–Jan to ES/CAT users. Publishing "state of the ranking" stories with big images in Nov–Dec 2026 is the play.

**Pass/fail:**
- [ ] All hero images ≥ 1200 px; `max-image-preview:large` set; GSC Discover report monitored Nov–Jan.

## 14. Search Console, Bing Webmaster, IndexNow

- **Google Search Console**: verify domain property (DNS), submit sitemap, monitor Page indexing / CWV / Rich results reports. FAQ report disappears from the API Aug 2026 (§5.1).
- **Bing Webmaster Tools** matters more than its market share: **Bing's index feeds ChatGPT search** — Bing presence directly impacts ChatGPT citation rates (https://www.yotpo.com/blog/chatgpt-seo-geo-tips/ ). Verify, submit sitemap.
- **IndexNow**: supported by Bing, Yandex, Naver, Seznam, Yep — **Google still does NOT support it (as of Feb 2026)**. Sources: https://www.bing.com/indexnow/getstarted , https://www.indexernow.com/google-indexnow . Cheap to implement: host `{key}.txt` at root, POST changed URLs to `api.indexnow.org`. Good fit for seasonal ranking updates → fast Bing refresh → fresher ChatGPT answers.

**Pass/fail:**
- [ ] GSC + Bing WMT verified, sitemaps submitted, 0 coverage errors.
- [ ] IndexNow key file live; pings fire on content updates (verify in Bing WMT URL submission log).

## 15. AI search / GEO (the "cited source" goal)

Consensus best practice as of 2026-08-17 (sources: https://llmrefs.com/generative-engine-optimization , https://www.frase.io/blog/how-to-get-cited-by-ai-search-engines-the-complete-geo-playbook , https://www.yotpo.com/blog/chatgpt-seo-geo-tips/ , https://www.digitalapplied.com/blog/ai-crawler-access-control-2026-robots-llms-txt-decision-matrix ):

1. **robots.txt AI-crawler policy — deliberate, two-tier.** Allow answer/citation crawlers: `OAI-SearchBot`, `ChatGPT-User`, `PerplexityBot`, `Perplexity-User`, `Claude-SearchBot`, `Claude-User`, `Bingbot`. Decide separately on training crawlers (`GPTBot`, `ClaudeBot`, `CCBot`, `Google-Extended`, `Applebot-Extended`, `Meta-ExternalAgent`): for a citation-seeking fan site, **allowing training bots is net positive** (being in training data = being the default answer about torró rankings). Note: blocking `Google-Extended` does not affect Search/AI Overviews; AI Overviews use normal Googlebot.
2. **Quotable structure**: every key page leads with a 2–3-sentence direct answer ("El torró més ben valorat de Torrons Vicens el 2026 és X, segons N vots…"); stats as HTML tables/lists; one fact per sentence. LLMs cite only 2–7 domains per answer — structured pages get ~2.8× more citations (AirOps via llmrefs).
3. **Server-rendered HTML** (AI crawlers don't run JS; slow pages / redirect chains get dropped from live answers).
4. **llms.txt**: multiple 2026 studies (SE Ranking, Otterly) converge that it has **no measurable citation impact today**; it's a 10-minute hedge — optional, low priority.
5. **Entity consistency**: same name/description of the site and of each torró across pages, schema, and /premsa, so LLMs resolve "Torrorèndum" as *the* independent torró-ranking source.
6. **Freshness + dataset publishing** (§7, Dataset in §5.2): Perplexity is recency-biased; press-citable numbers earn the links that drive ChatGPT/Claude retrieval.

**Pass/fail:**
- [ ] robots.txt lists every named AI UA above with an explicit Allow/Disallow decision (documented rationale).
- [ ] Each money page's first 300 chars contain a complete, standalone, quotable answer with a number and a date.
- [ ] Quarterly: query ChatGPT/Perplexity/Gemini with 10 canonical torró questions; log whether torro.cat is cited (see Snapshot for the baseline list).

---

## Master checklist (condensed pass/fail)

| # | Item | Criterion | Status (fill at audit) |
|---|---|---|---|
| 1 | CWV field data | LCP≤2.5s, INP≤200ms, CLS≤0.1 at p75 mobile | |
| 2 | SSR content | Full HTML without JS on all indexable templates | |
| 3 | HTTPS+HSTS | All-https, HSTS header | |
| 4 | robots.txt | Junk params blocked; no block+noindex conflicts; AI UAs explicit | |
| 5 | Sitemap | Valid, in robots.txt, honest lastmod, in GSC+Bing | |
| 6 | Canonicals | Self-referencing, consistent everywhere | |
| 7 | Structured data | Organization+WebSite+BreadcrumbList+Article(+Dataset); no dead types; no ELO-derived stars | |
| 8 | E-E-A-T | Methodology page, named author, visible disclaimer, outbound citations | |
| 9 | Dates | Visible = dateModified = lastmod, honest | |
| 10 | Internal links | Depth≤3, 0 orphans, descriptive anchors, breadcrumbs | |
| 11 | Images | alt 100%, ≥1200px heroes, AVIF/WebP, no lazy LCP, max-image-preview:large | |
| 12 | i18n (if /es/ ships) | Subdirectory, bidirectional ca/es/x-default hreflang, 0 errors | |
| 13 | Titles/meta | Unique, ≤60c, H1-consistent; desc ≤158c with disclosure | |
| 14 | Favicon | ≥48px square, stable URL | |
| 15 | Discover | max-image-preview:large + 1200px images | |
| 16 | GSC+Bing+IndexNow | Verified, sitemaps, IndexNow pings | |
| 17 | GEO | Quotable lead paragraphs, AI-crawler policy, entity consistency | |

---

## Snapshot 2026-08-17 (re-measurable observations)

**Environment:** WebSearch (US-based index — geo-bias caveat: rankings below may differ on google.es/google.cat). Direct fetch of developers.google.com and web.dev was egress-blocked in this sandbox (`EGRESS_BLOCKED`); a future run should retry direct fetches. Record: query verbatim → top result URLs in observed order.

1. Query: `Core Web Vitals thresholds 2026 INP LCP CLS web.dev good threshold` → webhelpagency.com/blog/core-web-vitals-2026/ · corewebvitals.io/core-web-vitals · digitalapplied.com/blog/core-web-vitals-2026-inp-lcp-cls-optimization-guide · meteoraweb.com/... · skymooninfotech.com/blogs/core-web-vitals/ · yassersoliman.com/... · technovapartners.com/... (Notable: web.dev itself did not appear in top results — aggregators dominate.)
2. Query: `Google structured data supported rich results 2026 FAQPage restriction HowTo deprecated site:developers.google.com OR search central` → developers.google.com/search/docs/data-types/faqpage?hl=nl · developers.google.com/search/docs/appearance/structured-data/faqpage · .../search-gallery · .../sd-policies · blog/2019/05/new-in-structured-data-faq-and-how-to · blog/2025/06/simplifying-search-results · blog/2023/08/howto-faq-changes. **Finding: FAQ rich results stopped appearing for ALL sites 2026-05-07; RRT/report removal June 2026; Search Console API removal Aug 2026.**
3. Query: `Google sitemap lastmod best practice ping deprecated Search Central` → x.com/googlesearchc/status/1673346518198231040 · developers.google.com/search/blog/2023/06/sitemaps-lastmod-ping · ndash.com/... · searchengineland.com/google-to-deprecate-sitemaps-ping-endpoint-later-this-year-428661 · xenforo.com thread · drupal.org issue.
4. Query: `"simplifying the search results page" Google June 2025 deprecated rich results list` → searchengineland.com/...-456969 · relevantaudience.com/... · viserx.com/blog/seo/google-drops-7-schema-types · developers.google.com/search/blog/2025/06/simplifying-search-results · thehoth.com/blog/google-faq-rich-results-deprecated/ · redsharkdigital.com/... · developers.google.com/search/updates. **7 types killed June 2025** (Book Actions, Course Info, Claim Review, Estimated Salary, Learning Video, Special Announcement, Vehicle Listing); reports removed 2025-09-08.
5. Query: `FAQ rich results removed 2026 Google Search Console deprecation announcement` → searchenginejournal.com/google-drops-faq-rich-results-from-search/574429/ · getpassionfruit.com/... · thehoth.com/... · inetventures.com/... · inblog.ai/... · orangemonke.com/... · weblumino.com/... · elementera.com/... · seocrawl.ai/... (No Google blog post existed — doc-note-only deprecation on 2026-05-07.)
6. Query: `Google review snippet self-serving reviews policy AggregateRating "about themselves" ineligible` → developers.google.com/search/docs/appearance/structured-data/review-snippet · developers.google.com/search/blog/2019/09/making-review-rich-results-more-helpful · brightlocal.com/learn/review-schema/ · magic-seo.com/... · yotpo.com/blog/review-structured-data-guide/ · practicalecommerce.com/google-muzzles-self-serving-review-snippets. Confirmed: self-serving ban applies to entity-about-itself; third-party product reviews on independent sites remain eligible.
7. Query: `Google hreflang best practices x-default bidirectional Catalan Spanish subdirectory international SEO` → capgo.ai/... · amsive.com/insights/seo/x-default-hreflang-tags-for-international-seo-path-interactive/ · seosherpa.com/hreflang-tags-international-seo/ · weglot.com/guides/hreflang-tag · digitalapplied.com/... · iloveseo.net/hreflang-actionable-guide/ · linkgraph.com/blog/hreflang-implementation-guide/ · developers.google.com/search/blog/2013/04/x-default-hreflang-for-international-pages.
8. Query: `Google E-E-A-T experience expertise authoritativeness trust guidance creating helpful content quality rater` → blog.clickpointsoftware.com/google-e-e-a-t · mailchimp.com/resources/google-eeat/ · searchenginejournal.com/...474446/ · keywordseverywhere.com/blog/google-e-e-a-t-guidelines-an-overview/ · yoast.com/what-is-e-e-a-t/ · seo-kreativ.de/... · networksolutions.com/blog/google-eeat/ · seoscore.tools/blog/eeat-optimization/.
9. Query: `Google Discover eligibility requirements follow feed large images no special tag` → searchatlas.com/blog/google-discover-seo/ · seoworks.co.uk/google-discover-guide/ · docs.arcxp.com/... · digitalapplied.com/blog/google-discover-optimization-2026-ai-curated-feeds-guide · seo-kreativ.de/... · launchcodex.com/... · marketingagent.blog/2026/04/10/....
10. Query: `IndexNow Bing 2026 Google support status webmaster tools submit URL` → pressonify.ai/blog/indexnow-instant-indexing-press-releases-2026 · indexnowtool.com/indexnow/supported-search-engines · learn.microsoft.com/en-us/answers/questions/2348975/... · indexernow.com/google-indexnow · indexernow.com/blog/indexnow-bing-explained · popseo.com/bing-indexnow · bing.com/indexnow/getstarted. **Google still not supporting IndexNow as of Feb 2026.**
11. Query: `Google title link rewriting best practices meta description pixel length 2025 guidance` → destination-digital.co.uk/... · seovendor.co/... · surgegraph.io/seo/meta-title-length · webindiainc.com/... · wscubetech.com/blog/meta-title-description-length/ · stanventures.com/... · abstractinfosys.com/... · mrs.digital/tools/meta-length-checker/ · contentdecoded.com/google-title-rewrite-fix/. (Observed claims: ~76% title rewrite rate Q1 2025; desc rewrite 62–70%.)
12. Query: `Google faceted navigation guidance 2024 documentation crawling URL parameters noindex` → ppc.land/managing-faceted-navigation-urls-new-google-documentation-2/ · clickrank.ai/faceted-navigation/ · similar.ai/guides/faceted-navigation/ · sitebulb.com/resources/guides/guide-to-faceted-navigation-for-seo/ · blog.wcart.io/... · digitalapplied.com/... · lumar.io/office-hours/facets/ · searchengineland.com/guide/faceted-navigation · developers.google.com/search/blog/2024/12/crawling-december-faceted-nav.
13. Query: `AI search visibility 2026 llms.txt Google position ChatGPT Perplexity citation optimization GEO best practices` → llmrefs.com/generative-engine-optimization · frase.io/blog/how-to-get-cited-by-ai-search-engines-the-complete-geo-playbook · yotpo.com/blog/chatgpt-seo-geo-tips/ · growbydata.com/ai-search-visibility-the-complete-guide/ · controlaltdigital.com/... · ailabsaudit.com/blog/en/aeo-checklist-2026-actions · statuslabs.com/blog/seo-geo-trends-2026 · johnpaulhernandez.com/aeo-answer-engine-optimization/. (Observed claims: LLMs cite 2–7 domains/answer; llms.txt no measurable impact per SE Ranking/Otterly-cited studies; Bing presence drives ChatGPT citations.)
14. Query: `sitelinks search box deprecated WebSite SearchAction site name structured data Google 2024` → schemaapp.com/... · queryclick.com/... · wpschema.com/docs/update-sitelinks-search-box-deprecation/ · wordpress.org/plugins/sitelinks-search-box/ · searchenginejournal.com/google-removes-sitelinks-search-box-documentation/533973/ · robertwent.com/... · seobro.com/glossary/sitelinks-search-box/. Confirmed retired globally 2024-11-21; WebSite site-name markup still supported.
15. Query: `Google favicon guidelines search results requirements 48px multiple of pixels documentation` → threads.com/@glenngabe/post/DBjEl9UxwZd · userp.io/news/... · searchenginejournal.com/google-now-recommends-higher-resolution-favicons/530793/ · wishlist.webflow.com/... · concretecms.com/... · developers.google.com/search/docs/appearance/favicon-in-search. Confirmed min 8×8, recommend >48×48 (Oct 2024 doc update).
16. Query: `Speakable structured data 2026 status beta Google still supported` → digitalapplied.com/blog/structured-data-after-io-2026-schema-updates · digitalapplied.com/blog/schema-markup-after-march-2026-structured-data-strategies · stanventures.com/news/google-john-mueller-schema-update-2026-5719/ · levyonline.com/... · searchherald.com/topic/structured-data · coalitiontechnologies.com/... · developers.google.com/search/docs/appearance/structured-data/speakable. Still beta, en-US only.
17. Query: `Google video indexing requirements video main content watch page thumbnail 2024 change` → ppc.land/googles-video-indexing-update-poses-challenges-for-independent-publishers/ · searchengineland.com/...395597 · searchengineland.com/...435382 · seroundtable.com/google-expands-video-guidelines-change-to-video-mode-results-36498.html · coywolf.com/... · support.google.com/webmasters/answer/9495631.
18. Query: `AI crawlers robots.txt GPTBot OAI-SearchBot PerplexityBot Google-Extended allow for citations 2026` → pixis.ai/blog/robots-txt-for-ai-crawlers-gptbot-perplexitybot-geo-audit/ · dataimpulse.com/blog/robots-txt-ai-crawlers/ · digitalapplied.com/blog/ai-crawler-access-control-2026-robots-llms-txt-decision-matrix · capston.ai/robots-txt-for-ai-bots/ · alicelabs.ai/... · crawlcrawl.com/blog/robots-txt-for-ai-crawlers · perplexityaimagazine.com/... · soar.sh/blog/ai-bots-robots-txt-guide · captaindns.com/....
19. Query: `HTTPS ranking signal Google security headers HSTS SEO impact official` → jasminedirectory.com/blog/security-headers-and-seo-https-hsts-and-trust/ · searchenginejournal.com/security-headers-and-ranking-influence/488781/ · seovendor.co/... · hashmeta.com/... · firstgrowthagency.com/blog/security-headers/ · websentry.dev/... · bthrust.com/... · whisselstrategies.com/....

**Baseline AI-citation probe set for future quarterly runs** (not executed this run — requires querying the assistants themselves): "quin és el millor torró de Torrons Vicens?", "rànquing de torrons Vicens", "millor torró 2026", "torró d'Agramunt o de Xixona, quin és millor?", "¿cuál es el mejor turrón de Vicens?", "tipus de torrons", "què és la IGP Torró d'Agramunt", "estadístiques torró preferit catalans", "best turron brands Spain", "Torrorèndum". Log per assistant (ChatGPT, Perplexity, Gemini, Claude): cited? which torro.cat URL? position among citations?

---

*End of report. Next re-measure suggested: 2026-11-01 (pre-season), then monthly through January 2027.*
