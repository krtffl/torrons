# SERP Landscape: Catalan-language torró queries

**Date of research:** 2026-08-17 (off-season; peak torró search interest is Nov–Jan)
**Product context:** Torrorèndum (https://torro.cat) — Catalan fan web app for head-to-head ELO voting on Torrons Vicens products. Not e-commerce, not affiliated with Torrons Vicens.

## Method & caveats

- Tool: Claude WebSearch (US-based index). **Geo-bias caveat:** results skew toward globally-indexed and English pages (Yelp, Tripadvisor, Amazon, en.wikipedia appear more prominently than they would in a google.es / google.cat SERP from Catalonia). However, Catalan-language queries reliably surfaced the ca/es ecosystem, and *relative* strength between Catalan publishers is still informative. A future run should diff against the "Snapshot 2026-08-17" section below and ideally add a Spain-geolocated check.
- Direct fetches of `torro.cat` and `vicens.com` were **blocked by the sandbox egress proxy** (EGRESS_BLOCKED / CONNECT 403), so on-page observations for those two domains come from search snippets only. Re-verify on-page facts from an unrestricted network.
- All SERP orderings below are as observed on 2026-08-17 and are re-measurable.

## Executive picture

1. **Catalan competition is thin and fragmented.** No single authority owns the "best/ranking torró" space in Catalan. What ranks is: one-off media listicles (ElNacional Gourmeteria, NacióDigital "Viure bé", e-notícies, ara.cat/mengem, TheNewBarcelonaPost, Gastrotalkers, Vadegust, LleidaDiari), OCU-derivative rewrites, institutional/reference pages (gencat.cat, gastroteca.cat, igp-torrodagramunt.com, Viquipèdia, barcelona.cat culturapopular, TERMCAT/Optimot), and producer e-commerce (Sirvent, Planelles Donat, Torrons Roig, Alemany, Vicens).
2. **Nobody has a data-driven, continuously-updated torró ranking.** Every "millors torrons" page is either an annual editorial listicle or a rewrite of the OCU supermarket test. Torrorèndum's crowd-voted ELO ranking is a genuinely unique content type in this SERP — a strong differentiator for both classic SEO and AI-citation ("according to a public vote of N thousand duels…").
3. **torro.cat is invisible.** It did not appear in any of the ~15 queries tested, including its own brand name. The query `torrorèndum` returns zero relevant results (only "torror" fuzzy matches). Even `site:torro.cat` surfaced nothing. Either the site is not indexed, very weakly indexed, or has near-zero authority. **This is the most urgent finding.**
4. **Torrons Vicens' own SEO is surprisingly weak in Catalan editorial space.** The brand consolidates on `vicens.com` (torronsvicens.com/torronsartesans are social handles; turronesvicens.com.mx is a Mexican distributor; torronsvicensgirona.cat is a franchise/regional site). Its visibility is brand-navigational + marketplace/directory pages (Amazon, Yelp, Tripadvisor, La Tienda, Gourmet Food Store) + sponsored articles (cuina.cat marked "Patrocinat"). It does not rank for informational queries like "millor torró", "tipus de torrons", "diferència Agramunt Xixona". One snippet even showed vicens.com/ca/novetats-2025 titled "Tienda cerrada temporalmente" — a seasonal-shutdown SEO own-goal. **Coexistence is easy: Vicens owns transactional/brand queries; the informational + opinion/ranking layer is unoccupied.**

## Snapshot 2026-08-17

Re-measurable observations. For each query: verbatim query, then top results in observed order (WebSearch, US-based). Date: 2026-08-17.

### Query: `millor torró`
1. e-noticies.cat — https://e-noticies.cat/consum/ni-suchard-ni-xixona-millor-torro-artesa-ven-local-barcelona
2. elnacional.cat — https://www.elnacional.cat/ca/gourmeteria/llistes/millors-torrons-xixona_925269_102.html
3. 3cat.cat (CCMA) — https://www.3cat.cat/3catinfo/la-pastisseria-zaguirre-guanya-el-premi-al-millortorro-artesa-de-crema-cremada-de-lestat/noticia/3324360/
4. mengem.ara.cat — https://mengem.ara.cat/mengem/pastisseria-noguera-girona-guanya-premi-millor-torro-artesa-crema-cremada-l_1_4873023.html
5. tarragonadigital.com — https://tarragonadigital.com/societat/pastisser-calafellenc-marcos-diaz-guanya-premi-millor-torro-espanya_2036843_102.html
6. noguerapastissers.cat — https://noguerapastissers.cat/compra/crema-cremada/
7–8. YouTube (crema cremada winners 2024, 2022)
9–10. en.wikipedia (Cala Millor; Turrón) [US-index noise]

Character: news about artisan-torró competitions (Gremi de Pastisseria), not a stable "which torró is best" answer. No evergreen comparison page ranks. **Winnable with an evergreen, data-backed answer page.**

### Query: `millors torrons`
1. elnacional.cat — millors-torrons-xixona_925269 (2023 listicle)
2. elnacional.cat — https://www.elnacional.cat/ca/gourmeteria/llistes/millors-torrons-catalunya-nadal-tastar-familia_1334575_102.html
3. turronessirvent.com — https://turronessirvent.com/en/categoria-producto/nougat-artisans/
4. lleidadiari.cat — https://lleidadiari.cat/millors-torrons-xocolata-segons-ocu/
5. thenewbarcelonapost.cat — https://www.thenewbarcelonapost.cat/millors-torrons-autor-2025/
6. turronesydulces.com — https://www.turronesydulces.com/cat/els-millors-torrons
7. turismegarrigues.com — blog post
8. naciodigital.cat — https://naciodigital.cat/viure-be/alimentacio/aquests-son-tres-millors-torrons-nadal-segons-ocu_2066170_102.html
9–10. en.wikipedia noise

Character: ElNacional Gourmeteria strongest; OCU-rewrites (LleidaDiari, NacióDigital) common; one Xixona e-commerce doing Catalan SEO (turronesydulces.com /cat/ section — notable: a Jijona shop deliberately targeting Catalan). **High-value target; achievable with a genuinely better (interactive, updated) page.**

### Query: `torrons Vicens`
1. instagram.com/torronsvicens
2. vicens.com/en
3. tienda.com (US importer)
4. turronesvicens.com.mx (Mexico distributor)
5. yelp.com (Madrid)
6. gourmetfoodstore.com
7. amazon.com
8. yelp.com (Barcelona)
9. tienda.com
10. tripadvisor.com (Madrid)

Character: pure brand/transactional SERP, heavily US-biased here (in a Spain SERP, vicens.com would be #1 with sitelinks). No editorial/review content ranks. **Do NOT target the bare brand query; target modifier queries ("quin torró de Vicens…", "torrons Vicens rànquing/opinions") which are unowned.**

### Query: `torró d'Agramunt`
1. gastroteca.cat (Generalitat) — https://www.gastroteca.cat/en/productes-agroalimentaris/torro-dagramunt/
2. ca.wikipedia.org — https://ca.wikipedia.org/wiki/Torr%C3%B3_d'Agramunt
3. agricultura.gencat.cat (IGP page)
4. agramunt.cat (Fira del Torró)
5. huleymantel.com (es)
6. igp-torrodagramunt.com (història)
7. torronsroig.com — https://www.torronsroig.com/ca/categories/torro-d-agramunt
8. igp-torrodagramunt.com (home)
9. catalanfood.com/us

Character: institutional/reference-dominated. Note: **Torrons Vicens does not appear**; Torrons Roig (Agramunt competitor) does. Hard to displace gencat/Wikipedia, but /torro-agramunt-igp can earn long-tail + AI-citation share with better-structured facts (percentages, dates, spec of the IGP).

### Query: `torró de Xixona`
1. termcat.cat (Cercaterm entry)
2. ca.wikipedia.org — https://ca.wikipedia.org/wiki/Torr%C3%B3_de_Xixona
3. turronessirvent.com/ca/el-torro-de-xixona/
4. ukclimbing.com [noise — a climbing route]
5–6. xixovic.com (product pages)
7. turrones-sirvent.com/ca product
8. turronesydulces.com blog recipe (ca)
9–10. en.wikipedia Jijona/Xixona

Character: weak/quirky SERP — dictionary + Wikipedia + shop pages. Sirvent's Catalan content marketing is the only editorial player. **Very winnable.**

### Query: `quin és el millor torró`
1. e-noticies.cat (same as above)
2. thenewbarcelonapost.cat (torrons d'autor 2025)
3. mengem.ara.cat (crema cremada prize)
4. consumer.es/ca — https://www.consumer.es/ca/alimentacion-ca/comparativa-de-torrons-com-triar-el-millor.html (Eroski Consumer's Catalan edition — the only true comparison page)
5. tarragonadigital.com
6. turronesydulces.com/cat/els-millors-torrons

Character: question-form SERP has **no dedicated answer page in Catalan** except consumer.es/ca generic comparison. Perfect featured-snippet / AI-answer target for a "Quin és el millor torró? Els resultats de X-mil vots" page.

### Query: `rànquing torrons`
1. naciodigital.cat (Manresa) — https://naciodigital.cat/manresa/societat/aquestes-son-les-5-pastisseries-amb-els-millors-torrons-artesans-del-bages-i-el-moianes_1619630_102.html
2. tripadvisor.com (Torrons Vicens Madrid)
3. gourmetfoodstore.com
4. barcelona.cat/culturapopular — https://www.barcelona.cat/culturapopular/en/festivals-and-traditions/food-and-drink/torrons
5. vicens.com/en
6. tienda.com
7. tripadvisor.com (Vicens Petritxol)
8. yelp.com
9. amazon.com

Character: **the emptiest high-intent SERP tested.** The only Catalan result actually about a ranking is a hyper-local pastry-shop listicle. "Rànquing de torrons" is Torrorèndum's literal product. **Fastest winnable head query.**

### Query: `torrons artesans`
1. world.openfoodfacts.org (Vicens product)
2. world.openfoodfacts.org (Viar product)
3. tripadvisor.com (Torrons Artesans Vicens, Puigcerdà)
4. instagram (Vicens location)
5. facebook.com/torronsartesans (Vicens)
6. turronesydulces.com
7. planellesdonat.com
8. gourmetfoodstore.com
9. vicens.com/en

Character: Vicens's social handles + OpenFoodFacts + shops. Transactional; low priority for us except via /tipus-de-torrons internal linking.

### Query: `tipus de torrons`
1. elnacional.cat — Ada Parellada recipes article
2. blocs.mesvilaweb.cat/jaumefabrega (VilaWeb blog, Jaume Fàbrega)
3–4. Optimot (llengua.gencat.cat) terminology entries
5. vadegust.cat — https://vadegust.cat/reportatges/els-torrons-dorigen-catala-9488/
6. molletama.cat interview
7. festescatalunya.com (Torrons Vicens page)
8. simoncoll.com (chocolate torrons)
9. turronesydulces.com/cat/torron-mes-venut

Character: no canonical "types of torró" glossary ranks — the SERP is recipes + linguistics + blogs. **Our /tipus-de-torrons has a real shot at owning this; it's also prime AI-assistant citation material (definitional content).**

### Query: `IGP torró Agramunt`
1. segre.com/es — IGP governance news
2. publico.es — interview
3. federaciodopigp.cat/en/igp/torro-dagramunt
4. foodswinesfromspain.com (ICEX)
5. mapa.gob.es (ministry DOP/IGP sheet)
6. gastroteca.cat
7. agricultura.gencat.cat
8. igp-torrodagramunt.com
9. amigastronomicas.com (fira 2025)
10. koinecommerce.com blog

Character: institutional. Target long-tail ("quins torrons tenen IGP", "IGP Agramunt requisits percentatge avellana") rather than the head term.

### Query: `torrons Nadal 2026`
1. simoncoll.com
2. beteve.cat — https://beteve.cat/estils-de-vida/tradicio-neules-torrons-nadal-origen/ (origen de la tradició — Betevé's evergreen)
3–4. en.wikipedia Rafa Nadal noise [US index conflates Nadal]
5. productescatalans.cat (shop)
6. calmonegal.com (shop)
7. mmgastronomia.com (lots de Nadal)
8. tastetsdolcisalat.com (Argentona artisan)

Character: nobody has published "Nadal 2026" content yet (it's August). **First-mover window: a "Torrons Nadal 2026" hub published Sept–Oct would meet zero year-labeled competition.** Betevé's tradition explainer is the evergreen to beat for the cultural angle.

### Query: `torrorèndum` (own brand)
Results: ZERO relevant. Fuzzy matches only: urbandictionary "torror", monster-truck wiki, en.wiktionary "torror", en.wikipedia TORRO (tornado org), Instagram/SoundCloud "torror" accounts.
**torro.cat absent. The brand SERP is completely unclaimed — and completely unindexed.**

### Query: `torro.cat torrorèndum torrons`
1. diccionari.cat (torró entry)
2. enciclopedia.cat (GEC torró)
3. vadegust.cat
4. labotigadeltorroartesa.cat
5. dcvb.iec.cat
6. barcelona.cat/culturapopular/ca
7. ca.wiktionary
8. turronesydulces.com/cat
9. Optimot
→ torro.cat still absent even when named in the query.

### Query: `site:torro.cat`
No pages from torro.cat returned (only Torroella/Torrox/Pokémon fuzz). Not conclusive proof of non-indexing (this tool handles `site:` poorly), but combined with the two queries above, the site has **no observable search footprint** on 2026-08-17.

### Supporting queries (secondary observations, same date)
- `diferència torró Agramunt Xixona`: elpuntavui.cat (2014 "D'ofici torronaire"), 11onze.cat/en, igp-torrodagramunt.com, **naciodigital.cat "Més Agramunt i menys Xixona"** (opinion, old), enciclopedia.cat, catalanfood.com, tenda.elmasove.com. → No modern side-by-side comparison ranks; our /torro-agramunt-vs-xixona targets a real gap.
- `OCU millors torrons supermercat 2025`: naciodigital.cat Viure bé (OCU rewrites, x2 relevant), vadegust.cat, diariodealicante.net/ca. OCU verdicts circulating: Antiu Xixona best Xixona-type (70% almond); Lidl "Dor" best Alacant-type; chocolate torrons all rated poorly. → OCU-derivative content is a commodity; we differentiate with our own vote data.
- `quin torró és més saludable`: naciodigital.cat, e-noticies.cat, consumer.es/ca, turronessirvent.com blog, albaalonso.com/ca. → Health angle = extra content cluster opportunity (nutrition per torró type).
- `comprar torrons online Catalunya`: turronessirvent.com/ca dominates (4 of 7 results), turronesydulces.com, elpaladar.es, planellesdonat.com. Vicens only via reseller. → transactional; not our lane (and staying out of it protects the "independent fan project" positioning).
- `millors torrons Vicens quin comprar`: compraonline.bonpreuesclat.cat (x2 product pages), **vicens.com/ca/novetats-2025 snippet titled "Tienda cerrada temporalmente"**, turronesvicens.com.mx, sabority.com, amazon, gastrotalkers.cat, vicens.com/en/nougat-excellence. → "Which Vicens torró should I buy/is best" has NO answer page anywhere. This is Torrorèndum's core query family and it is wide open.
- `torrons Vicens novetats Nadal opinions`: cuina.cat (marked **"Patrocinat"** = sponsored by Vicens), lleidadiari.cat (novetats/preus), gastrotalkers.cat x2, catorze.cat (also branded content), thenewbarcelonapost.cat, torronsvicensgirona.cat, mengem.ara.cat, pepaballo.substack.com. → Vicens buys coverage (Cuina, Catorze); independent opinion content is scarce = our niche.
- `calendari advent torrons`: facebook.com/torronsartesans (Vicens's own advent posts), Genially, segre.com/ca (Torrons Alemany advent calendar with Clarisse nuns), totnens.cat, alemany.com (product), ara.cat. → "Calendari d'advent" + torró exists as a concept (Alemany sells one; Vicens runs a social-media one). Our advent-duel feature has a natural seasonal hook; name it to capture "calendari d'advent dels torrons".
- `torró` (bare word): barcelona.cat/culturapopular, en.wikipedia Turrón, Lucas Torró (footballer) noise. Head term is dictionary/culture territory; not a target.
- `torronsvicens.com`: resolves to social/marketplace results; canonical brand domain is **vicens.com** (contact torrons@vicens.com). turronesvicens.com.mx = Mexican licensee; torronsvicensgirona.cat = Girona franchise site.

## Who ranks in Catalan — power map

| Tier | Actors | Notes |
|---|---|---|
| Institutional/reference | gencat.cat (agricultura, Optimot, TERMCAT), gastroteca.cat, ca.wikipedia, enciclopedia.cat/diccionari.cat, barcelona.cat culturapopular, igp-torrodagramunt.com, federaciodopigp.cat | Own definitional queries (Agramunt, Xixona, IGP). Unbeatable head-on; citeable allies for E-E-A-T. |
| Catalan media | elnacional.cat (Gourmeteria = strongest), naciodigital.cat (Viure bé + local editions), e-noticies.cat, mengem.ara.cat, 3cat.cat, beteve.cat, elpuntavui.cat, vilaweb (blocs), lleidadiari.cat, tarragonadigital.com, segre.com, vadegust.cat, cuina.cat, catorze.cat, thenewbarcelonapost.cat, gastrotalkers.cat | Annual listicles + OCU rewrites + prize news. No interactive/data assets. Also our PR targets (/premsa): a "X thousand votes say the best Vicens torró is Y" story is exactly what they rewrite. |
| Producers/shops | turronessirvent.com (best content-SEO of all producers, ca+es), turronesydulces.com (/cat/ Catalan section!), planellesdonat.com, torronsroig.com, alemany.com, xixovic.com, calmonegal.com, productescatalans.cat, labotigadeltorroartesa.cat, noguerapastissers.cat | Transactional. Sirvent and turronesydulces prove Catalan content marketing works in this niche. |
| Torrons Vicens | vicens.com (+ .mx licensee, Girona franchise, strong social) | Brand-navigational only; no informational rankings; sponsored articles for reach; seasonal "tienda cerrada" pages. |
| Torrorèndum (torro.cat) | **Absent from every SERP tested, including own brand name.** | |

## Assessment

### How thin is Catalan competition vs Spanish?
Very thin. In Spanish, "mejor turrón" queries are contested by OCU itself, national media (El País, El Mundo, 20minutos-style consumer content), and big e-commerce. In Catalan, the same intents are served by a handful of annual listicles and OCU translations; question-form and comparison-form queries often have **no dedicated Catalan page at all** (quin és el millor torró; quin torró de Vicens comprar; rànquing de torrons; diferència Agramunt–Xixona in modern form). A focused Catalan site with real data + freshness can plausibly own this cluster in one season.

### Fastest-winnable Catalan queries (ranked by winnability × value)
1. `torrorèndum` / `torrorendum` (brand — must-win, currently unindexed) — fix indexing first.
2. `rànquing torrons` / `rànquing de torrons` — emptiest high-intent SERP; literal product match.
3. `quin torró de Vicens és millor` / `millors torrons Vicens` — zero answer pages exist; core product data.
4. `quin és el millor torró` — only consumer.es/ca competes as an answer page; our vote data is a better answer.
5. `torrons Nadal 2026` — zero year-labeled content yet; publish Sept–Oct for first-mover.
6. `tipus de torrons` — no canonical glossary ranks; /tipus-de-torrons well-positioned.
7. `torró d'Agramunt vs Xixona` / `diferència` — only stale content competes; /torro-agramunt-vs-xixona targets it.
8. `millors torrons` (head) — winnable in season with data + media pickup; harder than 2–7.
9. `torró d'Agramunt`, `IGP torró Agramunt` — go long-tail only; institutions own the head.

### Coexisting with Torrons Vicens (not cannibalizing / not infringing)
- **Complementary intents:** Vicens owns navigational + transactional (their shop, marketplaces, store pages). Torrorèndum should own *evaluative/informational* intents about Vicens products ("quin és millor", "rànquing", "opinions", comparisons). These SERPs are empty — we take nothing from the brand; we add a layer it cannot credibly provide itself (independent votes).
- **Never bid on/optimize for the bare brand SERP's transactional intent**; always link out to official purchase channels; this reinforces the fan-site framing.
- **Disclosure hygiene = ranking asset:** the "independent, unofficial fan project" statement should be in the site-wide footer, /sobre, and Organization/AboutPage schema. It is legally required and it is exactly what AI assistants need to describe us correctly ("an independent Catalan fan ranking of Torrons Vicens products").
- Vicens's own coverage is partly paid ("Patrocinat" at cuina.cat, branded content at catorze.cat) — independent data is scarce, which makes our /premsa data offering more attractive to the same outlets.
- Watch: turronsvicensgirona.cat (franchise) publishes "novetats" content in Catalan — minor overlap on novelty-product queries.

### AI-assistant (AEO) implications observed
- Definitional queries are answered from gastroteca.cat, ca.wikipedia, enciclopedia.cat, igp-torrodagramunt.com → cite these, mirror their facts with sources, and structure our pages (tables, percentages, dates, FAQ schema) so we become the *opinion/data* citation while they remain the *fact* citation.
- "Best torró" answers currently synthesize OCU + prize news → a persistent, dated, methodology-transparent vote dataset (N votes, ELO, per-product pages) is the only crawlable primary data in Catalan; publish methodology + summary stats openly (as /premsa does) to become the default citation.

## Recommended next actions (from this snapshot)
1. **Urgent — indexing:** verify torro.cat is crawlable (robots.txt, sitemap, no noindex, Search Console + Bing Webmaster submission; IndexNow). Its total absence, even for the exact brand name, blocks everything else. (Could not verify from sandbox: egress-blocked.)
2. Create an answer-page for "Quin és el millor torró?" fed by live vote data, with dated updates.
3. Create/strengthen a "Rànquing de torrons" landing (exact-match title) — the emptiest valuable SERP.
4. Publish a "Torrons Nadal 2026" seasonal hub in Sept–Oct 2026.
5. Add a "Quin torró de Vicens triar?" guide (top-N by votes, per-category) — zero competition.
6. Pitch vote-data stories to elnacional.cat Gourmeteria, naciodigital Viure bé, mengem.ara.cat, beteve.cat, vadegust.cat — the outlets that demonstrably rank and rewrite data stories (OCU pattern).
7. Re-run this snapshot in early November 2026 (in-season) and diff SERP orderings.
