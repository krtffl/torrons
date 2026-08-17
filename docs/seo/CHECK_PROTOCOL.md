# SEO / positioning — recurring check protocol

Purpose: make every future SEO review run **comparable** with the previous ones, so
trends (ours and competitors') are visible instead of anecdotal. The first baseline
is `BASELINE_2026-08-17.md`. Each re-run produces a new dated snapshot next to it
(`BASELINE_YYYY-MM-DD.md`) using the same structure, plus (optionally) refreshed
research files under `research/`.

## Cadence

- **Off-season (Feb–Sep):** quarterly is enough — the SERPs barely move.
- **Pre-season (Oct):** one full run *before* publishing seasonal content, to catch
  the year's first "mejores turrones 2026 / torrons Nadal"-labeled competitors.
- **In-season (Nov–early Jan):** monthly, plus an extra pass the week after OCU
  publishes its annual test (mid/late December) — that event reshuffles everything.

## What to measure every run (same queries, same order)

### 1. Our side (torro.cat)

| Check | How | 2026-08-17 baseline |
|---|---|---|
| Brand query indexed | WebSearch `torrorèndum`, `torrorendum`, `"torro.cat"` | **Absent — not indexed at all** |
| `site:torro.cat` footprint | search engine site: query (tool support varies; note method) | No pages observed |
| robots.txt / sitemap.xml / llms.txt reachable | `curl -s -o /dev/null -w "%{http_code}" https://torro.cat/{robots.txt,sitemap.xml,llms.txt}` | Unverifiable from CI sandbox (egress-blocked); routes exist in code |
| Compression / cache / redirect hygiene | curl headers on `/`, http→https, www→apex | Unverifiable from sandbox this run |
| Rankings for target queries | run the query set below, note our position or absence | Absent everywhere |
| AI-assistant answers | ask ChatGPT/Perplexity/Claude (manually or via API) "quin és el millor torró?", "millors torrons Vicens?" and note whether torro.cat is cited | Not cited (OCU verdicts + Wikipedia dominate) |

> **Sandbox note:** the Claude Code environment used on 2026-08-17 blocks egress to
> torro.cat (proxy 403), so live-site checks must either run from an environment
> whose network policy allows torro.cat, or be done manually by the operator.

### 2. The query set (run verbatim, record top ~8 results in order)

Catalan (primary): `millor torró` · `millors torrons` · `rànquing torrons` ·
`quin és el millor torró` · `torrons Vicens` · `millors torrons Vicens` ·
`torró d'Agramunt` · `torró de Xixona` · `tipus de torrons` · `IGP torró Agramunt` ·
`torrons Nadal <year>` · `torrorèndum`

Spanish (secondary): `mejor turrón` · `mejor turrón <year>` · `ranking turrones` ·
`comparativa turrones` · `mejor turrón de chocolate` · `turrones Vicens` ·
`turrón de Agramunt` · `tipos de turrón` · `cata de turrones` ·
`turrón de Jijona o Alicante diferencia`

Record: query → ordered result URLs → date → tool used (note the US geo-bias of
WebSearch; a Spain-geolocated check is strictly better when available).

### 3. Competitor watch-list (check for changes)

- **OCU** (ocu.org/alimentacion/dulces/…): new annual test? sample size? new verdicts?
  (Baseline verdicts and URLs: see `research/competitors.md` §1 and
  `research/serp-landscape-es.md` "Key OCU verdicts".)
- **Directo al Paladar**: new versions of the annual catas ("…-3" URL suffix bumping).
- **ElNacional.cat Gourmeteria / NacióDigital Viure bé / e-notícies / vadegust /
  thenewbarcelonapost.cat**: new "millors torrons" listicles; do they cite anyone's data?
- **turronesydulces.com/cat/**: still ranking for `millors torrons`? (thin affiliate —
  if it still outranks us in-season, our page needs work).
- **vicens.com**: new content sections, whether the seasonal shop closure ("Tienda
  cerrada temporalmente" snippet) recurs, campeonatoturron.vicens.com activity.
- **New entrants:** re-run `votación torneo turrones bracket` / `encuesta mejor turrón`
  style queries — the interactive lane was **empty** at baseline; if anyone launches a
  voting/ranking product, treat it as a priority alert.
- **AI answers:** what do assistants answer for "mejor turrón según votos" — still OCU?

### 4. Diffing and reporting

1. Copy the previous `BASELINE_*.md`, update every table/observation, keep the same
   section order.
2. Add a "## Changes since last run" section at the top: moved queries, new/lost
   competitor assets, our position changes, completed roadmap items.
3. Update `STRATEGY.md` checkboxes; re-prioritize if a gap closed or a new one opened.
4. Commit both files; the git history of `docs/seo/` **is** the trend record.

## Operator actions that code can't do (verify status each run)

These block or unblock everything else and require the site owner (status at
baseline: **all pending/unknown — the site was not even indexed**):

- [ ] Google Search Console property for torro.cat verified; sitemap submitted;
      indexing requested for key pages.
- [ ] Bing Webmaster Tools property verified (feeds Bing + ChatGPT/Copilot answers);
      sitemap submitted; IndexNow key configured.
- [ ] DNS/redirect hygiene confirmed live (http→https, www→apex 301, no proxy quirks).
- [ ] Analytics that can attribute AI/LLM referrers (chatgpt.com, perplexity.ai
      referrals) — privacy-respecting is fine, but measure *something*.
- [ ] Social profiles claimed (at minimum Instagram/TikTok handle parity for the brand
      name) and linked from the site (sameAs).
