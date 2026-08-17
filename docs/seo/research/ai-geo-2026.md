# AI / Answer-Engine Optimization (GEO/AEO) — State of Play 2026 and Playbook for Torrorèndum

**Research date:** 2026-08-17 · **Researcher:** Claude (subagent, SEO research workflow)
**Scope:** How a site becomes the source that AI assistants (ChatGPT, Perplexity, Gemini, Claude) cite and recommend, and what Torrorèndum (torro.cat) should do about it before Nadal 2026.

**Method caveats:**
- WebSearch used here is **US-based**; Catalan/Spanish queries were run and do surface the es/ca ecosystem, but rankings observed may differ from what a user in Catalonia sees (geo/language personalization). Treat the Snapshot section as a *relative* baseline, re-measured with the same tool.
- The research sandbox's egress proxy **blocked direct fetches to torro.cat, platform.openai.com, docs.perplexity.ai, support.anthropic.com, llmstxt.org and ppc.land** — live-site checks (robots.txt, llms.txt, sitemap status of torro.cat) could not be performed from this environment and are flagged below as "verify locally". Claims about those primary docs are sourced via secondary reporting.

---

## 1. The AI crawler landscape (mid-2026)

There are now **three distinct crawler roles**, and the same vendor typically runs one bot per role. Getting this distinction right is the foundation of robots.txt policy:

| Role | What it does | Bots |
|---|---|---|
| **Training** | Collects content for model pre-training. Blocking it removes you from *future model weights* (long-term "the model just knows about you") but not from live search answers. | GPTBot, ClaudeBot, Google-Extended, Applebot-Extended, Meta-ExternalAgent, Bytespider, Amazonbot (partly) |
| **Search indexing** | Builds the vendor's answer-engine index. Blocking it removes you from AI search citations. | OAI-SearchBot, Claude-SearchBot, PerplexityBot |
| **Live user fetch** | Fetches a page in real time when a user's prompt needs it (the fetch that produces most inline citations + referral clicks). | ChatGPT-User, Claude-User, Perplexity-User |

Per-vendor detail (sources: [OpenAI crawler docs summary — sorank.com](https://www.sorank.com/glossary-geo-seo/openai-crawlers), [menra.ai ChatGPT crawler guide](https://www.menra.ai/guides/chatgpt-crawler-guide), [ppc.land on OpenAI doc revisions](https://ppc.land/openai-revises-chatgpt-crawler-documentation-with-significant-policy-changes/), [Search Engine Land on Anthropic's bots](https://searchengineland.com/anthropic-claude-bots-470171), [Search Engine Journal on Claude bots](https://www.searchenginejournal.com/anthropics-claude-bots-make-robots-txt-decisions-more-granular/568253/), [nohacks.co 2026 AI user-agent landscape](https://nohacks.co/blog/ai-user-agents-landscape-2026)):

- **OpenAI** — `GPTBot/1.1` (training), `OAI-SearchBot/1.0` (ChatGPT Search index; explicitly *not* used for training anymore), `ChatGPT-User/1.0` (live fetch; OpenAI has *removed* robots.txt-compliance language for it in a 2025 doc revision — user-triggered fetches may not honor robots.txt). Verify real OpenAI bots by IP: `openai.com/gptbot.json`, `openai.com/searchbot.json`, `openai.com/chatgpt-user.json`. GPTBot's share of AI crawl traffic grew 4.7% → 11.7% (Jul 2024 → Jul 2025).
- **Anthropic** — `ClaudeBot` (training), `Claude-SearchBot` (search-quality indexing), `Claude-User` (live fetch). Anthropic states **all three honor robots.txt**, including Crawl-delay; documentation formalized Feb 2026 ([seroundtable](https://www.seroundtable.com/anthropic-updates-its-crawler-docs-40978.html)).
- **Perplexity** — `PerplexityBot` (index) and `Perplexity-User` (live fetch). In Aug 2025 **Cloudflare accused Perplexity of stealth-crawling** sites that blocked it (rotating to a generic Chrome UA), de-listed it as a verified bot ([Search Engine Journal](https://www.searchenginejournal.com/cloudflare-delists-and-blocks-perplexity-from-crawling-websites/552899/), [webpronews](https://www.webpronews.com/cloudflare-accuses-perplexity-of-evading-robots-txt-with-stealth-scrapers/)). Practical takeaway for a site that *wants* visibility: irrelevant — just allow it.
- **Google** — `Google-Extended` is a robots.txt *token only* (no separate crawler): it controls use of your content for **Gemini training and grounding**. Regular Googlebot still feeds Search, AI Overviews and AI Mode. Blocking Google-Extended while wanting Gemini visibility is self-defeating ([devhut.net 2026 robots.txt guide](https://www.devhut.net/robots-txt-and-blocking-ai-bots-what-website-owners-need-to-know-in-2026/)).
- **Apple** — `Applebot` feeds Siri/Spotlight/Safari; `Applebot-Extended` is the opt-out token for Apple Intelligence training. Allow both.
- **Meta** — `Meta-ExternalAgent` (AI training/crawling), `Meta-ExternalFetcher` (user fetch). Allow.
- **ByteDance** — `Bytespider`: aggressive, widely reported to **ignore robots.txt**; among the most-blocked AI bots. Allowing it costs nothing for visibility in TikTok/Doubao AI features; blocking it is largely symbolic anyway.
- **Amazon** — `Amazonbot`: feeds Alexa+ and Amazon AI services; respects robots.txt. Allow ([originality.ai bot list](https://originality.ai/ai-bot-blocking), [robotstxt.com/ai](https://robotstxt.com/ai)).

### robots.txt best practice for a site that WANTS AI visibility

The 2026 consensus ([genrank.io robots.txt-for-AI guide](https://genrank.io/blog/configure-robots-txt-for-ai), [nohacks.co](https://nohacks.co/blog/ai-user-agents-landscape-2026)): **default-allow everything, explicitly allow the search/user-fetch bots** (explicit `Allow` records are documentation + insurance against inherited `Disallow` rules), keep `Disallow` only for genuinely private paths (admin, API endpoints you don't want hammered), and reference the sitemap. For Torrorèndum there is no reason to block even the training bots — being *in the weights* is the long game for "quin és el millor torró?" answered without browsing. Recommended file:

```
User-agent: *
Allow: /

# AI search & assistant bots — explicitly welcome
User-agent: OAI-SearchBot
Allow: /
User-agent: ChatGPT-User
Allow: /
User-agent: GPTBot
Allow: /
User-agent: ClaudeBot
Allow: /
User-agent: Claude-SearchBot
Allow: /
User-agent: Claude-User
Allow: /
User-agent: PerplexityBot
Allow: /
User-agent: Perplexity-User
Allow: /
User-agent: Google-Extended
Allow: /
User-agent: Applebot
Allow: /
User-agent: Applebot-Extended
Allow: /
User-agent: Meta-ExternalAgent
Allow: /
User-agent: Amazonbot
Allow: /

Sitemap: https://torro.cat/sitemap.xml
```

Also check any **CDN/WAF layer** (Cloudflare "Block AI bots" toggle, managed robots.txt): Cloudflare's one-click AI-bot block and its managed rules are the most common way sites *accidentally* block GPTBot/ClaudeBot while their robots.txt says allow ([Cloudflare managed robots.txt docs](https://developers.cloudflare.com/bots/additional-configurations/managed-robots-txt/)).

---

## 2. llms.txt in 2026: spec status and real adoption

- **Spec** (llmstxt.org, proposed by Jeremy Howard / Answer.AI, Sept 2024): a Markdown file at `/llms.txt` — H1 site name, blockquote summary, H2 sections of curated links with one-line descriptions; optional `/llms-full.txt` with full flattened content. (Primary spec site blocked from this sandbox; description per secondary sources below.)
- **Adoption is real but consumption is not**: adoption rose **8.8×**, yet **97% of llms.txt files receive zero AI-bot requests**; most fetches of the file come from SEO audit tools, not from answer engines ([ppc.land study](https://ppc.land/llms-txt-adoption-rises-8-8x-but-97-of-files-get-zero-ai-requests/)).
- **No major AI vendor has committed to it in production** (OpenAI, Google, Anthropic, Meta, Mistral — none, as of Q1 2026). Google's John Mueller: "no AI system currently uses llms.txt" (Jun 2025, still the operative Google position) ([getpassionfruit 2026 guide](https://www.getpassionfruit.com/blog/should-i-create-an-llms.txt-file-google-s-2026-guidance-explained), [Presenc "State of llms.txt 2026"](https://presenc.ai/research/state-of-llms-txt-2026), [ariashaw evidence review](https://ariashaw.com/does-llms-txt-actually-work)).
- Nuance: Anthropic recommends llms.txt in agent-authoring guidance and OpenAI uses it in the Agents SDK / Agentic Commerce Protocol context — i.e., it is emerging as an **agent-readiness convention**, not a search-citation lever.
- **Verdict for Torrorèndum:** create it (30 minutes, zero risk, positions the site for the agentic-browsing wave), but expect **no measurable citation impact** in 2026. Do not prioritize it over Bing indexing or content work.

---

## 3. How each answer engine sources and cites the web — and what that implies

| Engine | Retrieval index | Live fetcher | Implication |
|---|---|---|---|
| **ChatGPT Search** | **Bing's index** is the backbone (~87% of ChatGPT citations align with Bing); OpenAI is building supplementary proprietary crawl (OAI-SearchBot) | ChatGPT-User | **Bing Webmaster Tools + IndexNow are mandatory infrastructure.** A page absent from Bing is invisible to ChatGPT and Copilot regardless of Google rank |
| **Perplexity** | Own index (~200B URLs via PerplexityBot) + Bing + live fetches; RAG pipeline assigns citations while composing | Perplexity-User | Freshness is Perplexity's most distinctive citation factor (cited pages avg ~1,166 days old vs older elsewhere; recency wins ties). Keep pages dated and updated |
| **Gemini / AI Overviews / AI Mode** | **Google's own index + Knowledge Graph**; AI Mode does "query fan-out" into sub-queries; only ~14% URL overlap between AI Mode and AI Overviews citations (SLIDEFACTORY, Jun 2026) | Google infrastructure | Classic Google SEO *is* Gemini GEO: rank for the underlying query and have a passage that directly supports a claim. Cover sub-question variants (fan-out) |
| **Claude** | **Brave Search index** (~40B pages; never officially confirmed by Anthropic but shown by subprocessor list + 86.7% citation overlap with Brave top results) | Claude-User | Check/ensure presence in Brave Search (search.brave.com `site:torro.cat`); Brave has its own crawler and a URL submission path — Bing-only playbooks miss Claude |

Sources: [aiplusautomation — ChatGPT search architecture](https://aiplusautomation.com/blog/chatgpt-bing-or-google), [stackmatix — Bing Webmaster Tools for ChatGPT](https://www.stackmatix.com/blog/bing-webmaster-tools-chatgpt), [docdigitalsem — Bing indexing for AI](https://docdigitalsem.com/bing-indexing-for-ai-search/), [eseospace — how Perplexity indexing works](https://eseospace.com/blog/how-perplexity-indexing-works-2026/), [zeroclicklabs — Perplexity citation guide](https://zeroclicklabs.ai/perplexity-rank-citation-visibility-guide-2026/), [wislr — Gemini vs AI Overviews vs AI Mode](https://www.wislr.com/articles/gemini-vs-ai-overviews-vs-ai-mode), [stridec — how to get cited in Gemini](https://stridec.com/blog/how-to-get-cited-in-gemini/), [tryprofound — Claude web search explained](https://www.tryprofound.com/blog/what-is-claude-web-search-explained), [rivalhound — Claude runs on Brave](https://www.rivalhound.com/blog/claude-brave-search-visibility/), [isagentready — how Claude selects sources](https://isagentready.com/en/blog/how-claude-selects-sources-to-cite).

**Operational consequences:**
1. **Bing Webmaster Tools**: verify torro.cat, submit sitemap, watch the Index Explorer. This is the single highest-leverage infrastructure task for ChatGPT visibility ([subscribepr — indexed on Bing 2026](https://subscribepr.com/blog/how-to-get-indexed-on-bing/)).
2. **IndexNow**: push URL changes (new products, updated rankings) — Bing typically processes within hours-to-a-day; free, one API key file + a POST per change ([jetfuel.agency BWT/IndexNow guide](https://jetfuel.agency/how-to-set-up-bing-webmaster-tools-for-your-site-step-by-step-guide/)).
3. **Brave**: confirm indexing; Claude's citations track Brave's top-10 at ~87% ([rivalhound](https://www.rivalhound.com/blog/claude-brave-search-visibility/), [convertos — Brave submit URL](https://convertos.ai/geo/claude-brave-search-submit-url)).
4. **JS-dependence kills AI retrieval**: most AI fetchers don't execute JavaScript. Torrorèndum is Go server-rendered — an advantage; keep vote counts and rankings in the raw HTML, never client-side-only.

---

## 4. Content patterns that win citations

**The Princeton GEO paper** (Aggarwal et al., KDD 2024 — still the canonical experiment; [Princeton record](https://collaborate.princeton.edu/en/publications/geo-generative-engine-optimization/), [methodology deep-dive](https://blckalpaca.at/en/knowledge-base/seo-geo/geo-generative-engine-optimization/the-princeton-geo-study-methodology-results-and-critique)): tested ~10K queries; targeted rewrites boosted generative-engine visibility **22–41%**. Winning tactics, in order: **adding statistics, adding quotations, citing sources** (30–40% improvement in position-adjusted word count; up to **+115%** for sites ranked #5 in the underlying SERP — GEO helps underdogs most). Keyword stuffing *reduced* visibility.

Corroborating 2025–26 industry data:
- Quantitative claims get **~40% higher citation rates** than qualitative ones; statistics-focused pages ~40% higher than regular posts ([thedigitalbloom AI visibility report](https://thedigitalbloom.com/learn/2025-ai-citation-llm-visibility-report/)).
- **Original research / first-party data is the most reliable citation magnet**: benchmarks, surveys, proprietary datasets give LLMs a primary source to point at ([Ahrefs — how to earn LLM citations](https://ahrefs.com/blog/llm-citations/), [averi.ai — becoming a data source for LLMs](https://www.averi.ai/blog/building-citation-worthy-content-making-your-brand-a-data-source-for-llms)).
- Ahrefs' 1.4M-prompt study: pages matching ChatGPT's *narrower reformulated queries* get cited; **clear descriptive URLs** get cited more ([Ahrefs — why ChatGPT cites pages](https://ahrefs.com/blog/why-chatgpt-cites-pages/)).
- Format winners: **listicles, comparison pages, definition/glossary content, Q&A blocks, tables** — anything a model can lift as a self-contained passage ([BuzzStream cross-study review](https://www.buzzstream.com/blog/top-sites-chatgpt/), [Onely — LLM-friendly content](https://www.onely.com/blog/llm-friendly-content/)).
- **Brand search volume correlates with LLM citations (r≈0.334) more than backlinks** — brand/entity building beats link building for GEO ([Ahrefs via position.digital stats roundup](https://www.position.digital/blog/ai-seo-statistics/)).
- Structural rules of thumb: answer the page's question in the **first 40–80 words** under the H1; one claim per sentence; put numbers in text (not only in charts/images); date every data claim.

---

## 5. Schema.org's role for LLMs

Structured data does **not independently trigger citations**, but it strongly correlates with them and makes extraction reliable: ~71% of ChatGPT-cited pages and ~65% of Google AI Mode-cited pages carry structured data; pages with complete relevant schema see 20–40% more AI-answer appearances — *conditional on already ranking* ([stackmatix structured-data guide](https://www.stackmatix.com/blog/structured-data-ai-search), [derivatex — schema types that matter](https://derivatex.agency/blog/schema-markup-llm-seo/), [llmreach 2026 guide](https://www.llmreach.ai/blog/implement-structured-data-for-ai-2025-guide)). Highest-value types for Torrorèndum: **Dataset** (the voting data itself — rare and differentiating), **ItemList** (rankings), **FAQPage**, **Article** with `datePublished`/`dateModified`, **WebSite + Organization** (entity disambiguation, including the "independent fan project" description), **Product-adjacent** info on /torro/{id} pages *without* Offer markup (not e-commerce — don't fake it).

---

## 6. Wikipedia / Wikidata as AI-knowledge anchor

- Wikipedia is the single most influential source in LLM training (~22% by influence weight; 3–4.5% of tokens) and feeds the Knowledge Graph behind AI Overviews; ChatGPT cites Wikipedia in ~7.8% of all citations ([statuslabs — Wikipedia as truth anchor](https://statuslabs.com/blog/how-ai-models-use-wikipedia-as-a-truth-anchor/), [allmo — Wikipedia's impact on ChatGPT](https://allmo.ai/articles/what-we-know-about-the-impact-of-wikipedia-on-chatgpt-search-results)).
- A **Wikidata Q-ID** is the low-bar entity anchor: acceptable without a Wikipedia article if the item is "a clearly identifiable entity describable with serious, publicly available references" or fills a structural need ([Wikidata:Notability](https://www.wikidata.org/wiki/Wikidata:Notability)).
- **Critical constraint: COI.** Wikidata policy forbids creating an item about *your own* project; Viquipèdia likewise. **The legitimate path**: (1) earn independent press coverage (Catalan media love a Nadal data story), (2) let a third-party editor create the item / add torro.cat as a reference in the Viquipèdia "Torró" article, or transparently request it via talk pages declaring the COI. Existing anchors to build on: torró, Torró d'Agramunt (IGP), Torrons Vicens all have Viquipèdia articles and Wikidata items — Torrorèndum's data can become a *cited reference* inside those long before it merits its own article.

---

## 7. Measuring AI visibility in 2026

- **Referrer tracking**: custom GA4 channel group (placed *above* Referral) with source regex like `chatgpt\.com|chat\.openai\.com|perplexity\.ai|claude\.ai|gemini\.google\.com|copilot\.microsoft\.com|grok\.com|meta\.ai|you\.com|deepseek\.com|edgeservices` — or the equivalent in a privacy-friendly analytics stack ([abmatic GA4 setup](https://abmatic.ai/blog/track-ai-referral-traffic-ga4), [authoritytech guide](https://authoritytech.io/blog/ai-traffic-attribution-how-to-track-chatgpt-perplexity-gemini)). Caveat: **35–70% of AI-referred sessions arrive with no referrer** and land in Direct — referrer counts are a floor, not a total ([vyncedigital](https://vyncedigital.com/blog/measure-ai-search-traffic-in-ga4-track-chatgpt-gemini-perplexity)).
- **Server-log bot monitoring**: count monthly hits by UA for GPTBot, OAI-SearchBot, ChatGPT-User, ClaudeBot, Claude-SearchBot, Claude-User, PerplexityBot, Bingbot. ChatGPT-User/Claude-User hits ≈ your pages being read into live answers — the leading indicator of citations.
- **Prompt-panel testing**: fixed set of Catalan/Spanish prompts run monthly in ChatGPT/Perplexity/Gemini/Claude, recording whether torro.cat is cited (manual, free, adequate at this scale). SaaS trackers (Profound — enterprise; Otterly, Peec, ZipTie, Ahrefs Brand Radar — SMB tiers) exist but are overkill/cost-inefficient for a fan project ([backlinko tool review](https://backlinko.com/llm-tracking-tools), [otterly comparison](https://otterly.ai/blog/best-ai-search-monitoring-and-llm-monitoring-solutions/)).

## 8. Risks

- **Zero-click erosion**: with an AI Overview present, organic CTR drops ~47% (Pew, Jul 2025: 8% vs 15% click; only **1% ever click a citation**); Ahrefs (Feb 2026) measures **-58% CTR** on AIO queries ([wholewhale on Pew](https://wholewhale.com/tips/googles-gaslighting-pew-research-confirms-what-seos-already-know-ai-overviews/), [eduearnhub stats report](https://eduearnhub.com/ai-overviews-zero-click-search-statistics/)). Mitigation: make the *brand* the destination (interactive voting can't be summarized away — a duel you play is not a snippet).
- **Retrieved-but-not-cited**: models consume content to form answers without crediting it (Reddit citation rate ~1.9% despite heavy retrieval — [Search Engine Journal](https://www.searchenginejournal.com/chatgpt-often-retrieves-but-rarely-cites-reddit-pages-data-shows/572243/)). Mitigation: brand the data itself ("segons el Torrorèndum, amb N votes…") so paraphrase still carries the name; unique named statistics are harder to strip than generic prose.
- **Misattribution as official Vicens property**: LLMs compress "fan site about Torrons Vicens products" into "Torrons Vicens site". The legal disclosure must be machine-readable (Organization schema description, About page first paragraph, llms.txt summary) so retrieval-time answers repeat "independent, unofficial".
- **Rogue crawlers** (Bytespider, stealth Perplexity): for a visibility-seeking site this is a non-issue; just don't install anti-bot layers that also catch the good bots.

---

## 9. Recommendations for Torrorèndum (each: action → source → checkable criterion)

**Infrastructure**
1. **Deploy the explicit-allow robots.txt** (§1) with sitemap line; audit any CDN/WAF for AI-bot blocking. *Source:* [genrank.io](https://genrank.io/blog/configure-robots-txt-for-ai), [developers.cloudflare.com](https://developers.cloudflare.com/bots/additional-configurations/managed-robots-txt/). *Criterion:* `curl -A "GPTBot" https://torro.cat/ -o /dev/null -w "%{http_code}"` returns 200 for each of the 8 major AI UAs; robots.txt contains no Disallow affecting content pages. *(Could not verify current state from sandbox — egress blocked.)*
2. **Verify torro.cat in Bing Webmaster Tools; submit sitemap; wire IndexNow** pings into publish/update flow (Go: one HTTP POST to `api.indexnow.org` on content change). *Source:* [stackmatix](https://www.stackmatix.com/blog/bing-webmaster-tools-chatgpt), [subscribepr](https://subscribepr.com/blog/how-to-get-indexed-on-bing/). *Criterion:* BWT shows all sitemap URLs indexed; `https://torro.cat/{indexnow-key}.txt` returns 200; log shows Bingbot hits within 48h of a push.
3. **Confirm Brave Search indexing** (Claude's index). *Source:* [rivalhound](https://www.rivalhound.com/blog/claude-brave-search-visibility/). *Criterion:* `site:torro.cat` on search.brave.com returns ≥10 pages including /torro-agramunt-igp.
4. **Add /llms.txt** (site summary incl. unofficial-fan-project disclosure + curated links to SEO pages and the data page). Low expected impact; near-zero cost. *Source:* [ppc.land adoption study](https://ppc.land/llms-txt-adoption-rises-8-8x-but-97-of-files-get-zero-ai-requests/). *Criterion:* `https://torro.cat/llms.txt` returns 200, valid spec format.
5. **Guarantee no-JS readability** of rankings, vote counts and product data. *Source:* [tryprofound](https://www.tryprofound.com/blog/what-is-claude-web-search-explained). *Criterion:* `curl` of / and /torro/{id} shows ELO ranking numbers and vote totals in raw HTML.

**Content — the original-data moat**
6. **Build a flagship data page** (e.g. `/dades` or expand `/premsa`): "Observatori del Torró" with named, dated, quotable statistics — "Amb X.XXX vots, el torró més ben valorat de Torrons Vicens és Y (ELO Z)", head-to-head win rates, Agramunt-vs-Xixona vote splits, seasonal trends. One claim per sentence; numbers in text; update dated. This is the exact asset class (first-party survey data) that OCU rides to dominate "mejor turrón" queries. *Source:* Princeton GEO ([+22–41% visibility from statistics](https://collaborate.princeton.edu/en/publications/geo-generative-engine-optimization/)), [Ahrefs LLM citations](https://ahrefs.com/blog/llm-citations/). *Criterion:* page live with ≥10 dated statistics, each ≤25 words, each containing an absolute number.
7. **Open every SEO page with a 40–80-word direct answer** under the H1, and add FAQ (Q&A) blocks with self-contained answers. *Source:* [Onely](https://www.onely.com/blog/llm-friendly-content/), [Frase GEO playbook](https://www.frase.io/blog/how-to-get-cited-by-ai-search-engines-the-complete-geo-playbook). *Criterion:* first paragraph of each /torro-*, /tipus-de-torrons, /sobre answers its title query in <80 words with no preamble.
8. **Cover query fan-out variants** for Gemini AI Mode: sub-pages/sections answering "quin torró té menys sucre", "diferència torró dur i tou", "quin torró comprar per Nadal 2026", each with its own extractable passage. *Source:* [wislr](https://www.wislr.com/articles/gemini-vs-ai-overviews-vs-ai-mode). *Criterion:* ≥8 distinct sub-questions each answered in a dedicated, linkable H2/H3 block.
9. **Visible freshness**: "Actualitzat: {date} — {N} vots" on rankings and data pages (Perplexity's #1 differentiator; live vote counts make this genuinely dynamic). *Source:* [zeroclicklabs](https://zeroclicklabs.ai/perplexity-rank-citation-visibility-guide-2026/). *Criterion:* visible date + `dateModified` in schema on / and /dades, auto-updating.

**Machine-readable trust**
10. **Schema rollout**: `Dataset` (voting data, with license + temporalCoverage), `ItemList` (rankings), `FAQPage`, `Article` (dated), `WebSite`+`Organization` whose `description` states "web independent i no oficial, sense relació amb Torrons Vicens". *Source:* [derivatex](https://derivatex.agency/blog/schema-markup-llm-seo/), [stackmatix](https://www.stackmatix.com/blog/structured-data-ai-search). *Criterion:* all types pass validator.schema.org; Organization.description contains the unofficial disclosure verbatim.
11. **Brand the statistic, not just the page**: publish numbers phrased as "segons el Torrorèndum (torro.cat)…" in the press kit so paraphrase carries attribution. *Source:* [SEJ Reddit-paradox data](https://www.searchenginejournal.com/chatgpt-often-retrieves-but-rarely-cites-reddit-pages-data-shows/572243/). *Criterion:* /premsa offers ≥5 copy-paste "citable claims" embedding the brand name.

**Entity & authority**
12. **Nadal 2026 data-story PR** (Oct–Nov): pitch Catalan media (Vilaweb, Ara, NacióDigital, 3Cat, betevé, The New Barcelona Post, Vadegust — the outlets currently winning these SERPs, see Snapshot) a story from the voting data. Press coverage is simultaneously the citation source LLMs trust, the brand-search driver (r≈0.334 with LLM citations), and the notability basis for Wikidata. *Source:* [position.digital](https://www.position.digital/blog/ai-seo-statistics/), [Fractl](https://www.frac.tl/ai-citation-research-digital-pr-strategy/). *Criterion:* ≥2 independent Catalan media articles naming "Torrorèndum" published before 2026-12-01.
13. **Wikidata/Viquipèdia the legitimate way**: after press coverage exists, seek a third-party-created Wikidata item (or COI-declared request); propose torro.cat data as a *reference* in Viquipèdia's Torró / Torró d'Agramunt articles via talk page, never self-inserted. *Source:* [Wikidata:Notability](https://www.wikidata.org/wiki/Wikidata:Notability). *Criterion:* Torrorèndum Q-ID exists with ≥2 independent references; zero COI-policy violations (no self-created items).

**Measurement**
14. **Stand up the AI-visibility dashboard**: (a) analytics channel with the AI-referrer regex (§7); (b) monthly server-log counts per AI UA; (c) monthly prompt panel — fixed prompts incl. "Quin és el millor torró de Torrons Vicens?", "Quin torró de Vicens comprar per Nadal?", "millor torró d'Agramunt" across ChatGPT/Perplexity/Gemini/Claude, logging cited domains. *Source:* [abmatic](https://abmatic.ai/blog/track-ai-referral-traffic-ga4), [backlinko](https://backlinko.com/llm-tracking-tools). *Criterion:* a tracked log/dashboard exists with ≥1 monthly entry per stream; baseline row dated ≤2026-09-30.

**Priority order for the ~4 months before peak season:** 1–2–5 (week 1) → 6–7–10 (weeks 2–4) → 3–9–14 (September) → 12 (October–November) → 4–8–11–13 opportunistic.

---

## Snapshot 2026-08-17

Re-measurable observations. Tool: Claude WebSearch (**US-based — geo-bias caveat**: rankings for users in Catalonia/Spain may differ; re-measure with the same tool for comparability). Date: **2026-08-17**.

### Query snapshots (verbatim query → observed results in order)

**Q1: `quin és el millor torró de Torrons Vicens`**
1. thenewbarcelonapost.cat — /millors-torrons-autor-2025/
2. vadegust.cat — /actualitat/torrons-vicens-presenta-sinergia-…
3. botiguesdecatalunya.cat — /cat/soci/turrones-vicens-agramunt
4. festescatalunya.com — /torrons-vicens/
5. gastrotalkers.cat — /noticia/658/10-torrons-molt-originals-…
6. lacasadeltorro.com — /1288.html
7. vicens.com — /en

→ **torro.cat: ABSENT.** No source answers the question with data; the AI-synthesized answer hedged ("depends on taste") and highlighted Sinèrgia/Dubai/Plàncton novelty products. An "N votes say X" source would own this query.

**Q2: `millor torró rànquing votacions`**
1. thenewbarcelonapost.cat — /millors-torrons-autor-2025/
2. enderrock.cat (irrelevant — music awards)
3. xcatalunya.cat — /societat/torro-catala-que-millor-…
4. e-noticies.cat — /consum/ni-suchard-ni-xixona-…
5. turronesydulces.com — /cat/els-millors-torrons
6. mengem.ara.cat (irrelevant — restaurants)
7. tarragonadigital.com — /societat/pastisser-calafellenc-…

→ **torro.cat: ABSENT.** SERP is professional-jury award news (Gremi de Pastissers, Campionat Maestro Turronero); zero consumer-vote data sources. Direct content gap.

**Q3: `Torrorèndum torro.cat`** (brand query)
→ Results entirely irrelevant (en.wikipedia.org municipality pages: Torroella de Fluvià, Torroja del Priorat…; tiktok.com/tag/torrocat; pokemon.gameinfo.io). **The brand has zero presence in this (US) search index — the single most urgent finding.** Re-check after press/PR work; success = torro.cat is result #1 for its own name.

**Q4: `cuál es el mejor turrón según los consumidores encuesta ranking`** (Spanish)
1. cronista.com — /espana/actualidad-es/estos-son-los-mejores-turrones-del-supermercado-segun-la-ocu-…
2. ocu.org — /alimentacion/dulces/como-elegir-turron
3. elconfidencialdigital.com — /articulo/consumo/son-mejores-turrones-supermercado-ocu/…
4. infobae.com — /espana/2024/12/16/este-es-el-mejor-turron-…segun-la-ocu/
5. ebay.de (irrelevant)

→ **OCU (consumer-testing org) owns the "según consumidores" framing** — 4 of 5 results are OCU or OCU-derivative press. Model to emulate at Vicens/Catalonia scale: publish the data, let media derivatives multiply.

### Infrastructure facts to re-verify on next run
- OpenAI bot IP ranges published at: openai.com/gptbot.json, openai.com/searchbot.json, openai.com/chatgpt-user.json (per [sorank.com](https://www.sorank.com/glossary-geo-seo/openai-crawlers), 2026-08-17).
- ChatGPT Search ≈ Bing index (~87% citation alignment); Claude ≈ Brave (86.7% top-result overlap); Gemini = Google index + KG; Perplexity = own ~200B-URL index + Bing.
- llms.txt: 97% of files get zero AI requests; no major vendor production commitment (as of Q1 2026).
- Pew Jul 2025: AIO present → 8% organic CTR (vs 15%); 1% citation CTR. Ahrefs Feb 2026: -58% CTR on AIO queries.
- GPTBot share of AI crawl traffic: 11.7% (Jul 2025).
- torro.cat live checks (robots.txt / llms.txt / sitemap HTTP status): **BLOCKED from research sandbox (egress proxy)** — 2026-08-17 status unknown; verify from an unrestricted environment and record here.
