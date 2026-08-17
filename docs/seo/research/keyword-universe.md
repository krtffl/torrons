# Keyword / Query Universe — Torrorèndum (torro.cat)

**Date of research:** 2026-08-17 (off-season; peak torró/turrón search interest is Nov–Jan; next season = Nadal/Navidad 2026)
**Scope:** every search — classic Google/Bing AND conversational AI-assistant queries — that Torrorèndum should want to rank for or be cited for, in Catalan (primary), Spanish (secondary) and a small English long-tail.
**Companion docs:** `serp-landscape-ca.md` and `serp-landscape-es.md` (same date, same repo) hold deep SERP breakdowns for ~22 head queries; this doc adds 16 newly-measured queries (see "Snapshot 2026-08-17" below) and organizes the full universe by intent cluster.

---

## Method & caveats

- **Tools:** Claude WebSearch + WebFetch. WebSearch is **US-based**; orderings shown are approximate composition for a Spain/Catalonia SERP (expect more local media, Shopping units and local packs in google.es). Catalan/Spanish queries still reliably surface the ca/es ecosystem.
- **Autocomplete APIs were unreachable** from this sandbox: `suggestqueries.google.com`, `duckduckgo.com/ac`, and Bing suggest endpoints all returned proxy CONNECT 403. Autocomplete-style expansions below are therefore derived from (a) query patterns observed *inside* SERP titles ("People also search/ask"-style phrasing mined from ranking headlines), (b) morphological expansion of measured heads, and (c) the product catalog. **A future run with real gl=es/hl=ca autocomplete data should re-verify the long-tail lists.**
- **Volume/seasonality is qualitative** (no keyword-tool data available): scale = Very High / High / Med / Low / Micro, with seasonality noted. Nearly everything in this universe is extremely seasonal (≈10–20x Nov–Dec vs summer), except evergreen informational queries which flatten somewhat.
- **Difficulty:** Low / Med / High / Very High, based on who currently ranks (see snapshots + companion docs).
- **Priority:** P0 = core identity, must win; P1 = high-value and winnable this season; P2 = supporting long-tail, capture with existing/planned pages; P3 = monitor only / not our fight.
- Product catalog ground truth: 107 products in 4 classes (Clàssics, Novetats, Xocolata, Adrià Natura) from `migrations/000005_insert_classes.up.sql`, `000006_insert_torrons.up.sql`, `000017_rename_adria_natura.up.sql`.
- **Legal guardrail for all brand-adjacent targeting:** every page entering "Torrons Vicens" SERPs must carry the prominent independent/fan/non-official disclosure. Never bid for the bare brand query's #1; target modifier queries.

---

## Cluster 1 — Navigational / Brand

Users looking for us, or for Torrons Vicens by name.

| Query | Lang | Volume / seasonality | Difficulty | Page | Priority |
|---|---|---|---|---|---|
| torrorendum / torrorèndum | ca | Micro today; grows with brand | **Should be trivial — currently FAILING: zero results found for `"torrorendum" OR "torrorèndum"` on 2026-08-17** | Home | **P0 (urgent)** |
| torro.cat / torro cat / "torro punt cat" | ca | Micro | Trivial once indexed; today `site:torro.cat` surfaces nothing (see serp-landscape-ca.md) | Home | **P0 (urgent)** |
| torrorendum ranking / resultats torrorèndum | ca | Micro | Low | /classificacio (ranking page) | P1 |
| torrons vicens | ca/es | High, Nov–Dec spike | Very High (brand owns it + Amazon/Tripadvisor/Yelp) | n/a — do not target head | P3 |
| torrons vicens rànquing / els torrons vicens més votats | ca | Low | Low (unoccupied) | Ranking landing "Els torrons Vicens més votats" | **P0** |
| turrones vicens opiniones / análisis / cuál comprar | es | Med, seasonal | Low — current occupants are AI-spam pages (lorcaalacarta.es, cumul.es per serp-landscape-es.md) | ES ranking landing | **P1** |
| torrons vicens agramunt (botiga/fàbrica) | ca | Med, local | High (Maps/brand) | P3 — mention only | P3 |
| vicens albert adrià / adrià natura torró | ca/es | Low-Med, Dec spike | Med (vicens.com, Amazon, press: hellochefs.es, financialfood.es) | Class page for Adrià Natura + per-product pages | P1 |
| torró vicens [product name] (e.g. "torró vicens crema cremada") | ca | Micro each, long fat tail | Low | /torro/{id} pages | P1 |

**Autocomplete-style expansions to cover on brand pages:** "torrons vicens preus", "torrons vicens novetats 2026", "torrons vicens nadal", "turrones vicens amazon", "torrons vicens madrid/barcelona botiga" (P3, transactional/local — leave to brand).

---

## Cluster 2 — Best-of / Ranking (the core battleground)

Torrorèndum's native format. CA side is thin and winnable; ES side is OCU-dominated at the head but open at "ranking".

| Query | Lang | Volume / seasonality | Difficulty | Page | Priority |
|---|---|---|---|---|---|
| millor torró / quin és el millor torró | ca | High, Nov–Dec | Med — no evergreen answer page exists; SERP is prize-news (e-noticies, elnacional, 3cat) per serp-landscape-ca.md | **"El millor torró segons X mil vots" evergreen answer page** | **P0** |
| millors torrons (2026 / de Nadal / de Catalunya) | ca | High, Nov–Dec | Med (elnacional listicles, OCU rewrites) | Same + year-stamped seasonal edition | **P0** |
| rànquing torrons / rànquing de torrons 2026 | ca | Med | **Low — no owner found** | Ranking page (ELO leaderboard) with this exact phrase in title | **P0** |
| ranking turrones / ranking de turrones 2026 | es | Med-High | **Low — weakest head SERP found in ES (TierMaker UGC #1)** | ES ranking landing | **P0** |
| mejor turrón / mejor turrón 2026 | es | Very High, Dec | Very High (OCU ecosystem) — enter via freshness + original data, aim top-10 + AI citation, not #1 | ES seasonal results page | P1 |
| millor torró de xocolata / mejor turrón de chocolate | ca/es | Med-High, Dec | Med (OCU-rewrites only; LleidaDiari, Vitónica) | Per-category best page (Xocolata class = 12 products) | **P1** |
| millor torró de crema cremada / mejor turrón de yema | ca/es | Med, Dec | Med (annual prize news; beteve.cat) | Per-category best page — note Vicens crema cremada line is deep (6+ products) | **P1** |
| millor torró de festuc / mejor turrón de pistacho | ca/es | Med and rising (viral ingredient) | Low-Med (recipes + e-comm rank, no rankings) | Per-category best page | P1 |
| millor torró de praliné / massapà / coco / trufa | ca | Low each | Low | Per-category bests from ELO data | P2 |
| mejor turrón sin azúcar / millor torró sense sucre | ca/es | Low-Med, Dec | Low (e-comm only: Bonpreu, Torrons Roig, Pons) | Category page IF catalog tagged; else FAQ answer | P2 |
| millors torrons d'autor / mejores turrones de autor 2026 | ca/es | Med, Nov–Dec | Med (thenewbarcelonapost owns it) | Adrià Natura + Novetats class pages; pitch as data source | P1 |
| mejor turrón artesano | es | Med, Dec | High (award news + e-comm) | P2 — angle: "el premio popular del torró" | P2 |
| mejor turrón del supermercado / OCU turrones | es | Very High, Dec | Not winnable (OCU semantics) | Context/press page only | P3 |
| el torró més votat / el turrón más votado | ca/es | Micro today — **we can create this query** | Trivial | Ranking page | P0 |
| tier list turrones / tier list torrons | es/ca | Low, growing (TierMaker/TikTok demand proven) | Low | Ranking page framed shareably | P2 |

---

## Cluster 3 — Informational / Educational

Where AI-citation share is won. Existing pages already cover several heads.

| Query | Lang | Volume / seasonality | Difficulty | Page | Priority |
|---|---|---|---|---|---|
| què és el torró d'Agramunt / torró d'Agramunt IGP | ca | Med, Oct–Dec (fira spike in Oct) | High head (gastroteca.cat, Viquipèdia, gencat) but long-tail + AI-citation winnable | **/torro-agramunt-igp (exists)** | **P0** |
| turrón de Agramunt (IGP) | es | Med | **Low — no ES editorial reference exists; TasteAtlas ranks #1** (serp-landscape-es.md) | **Spanish twin of /torro-agramunt-igp — highest-affinity gap found** | **P0** |
| diferencia turrón Jijona y Alicante / torró Xixona vs Alacant | es/ca | High, Nov–Dec | Med (bonviveur wins ES; CA has only consumer.es/ca) | Extend /torro-agramunt-vs-xixona; add ES version | **P1** |
| tipus de torrons / tipos de turrón | ca/es | Med-High, Nov–Dec | Med (brand blogs only) | **/tipus-de-torrons (exists)** + ES version enriched with "el més votat de cada tipus" | **P1** |
| cuántas calorías tiene el turrón / quantes calories té el torró | es/ca | High, Dec–Jan (post-Christmas guilt spike) | Med (eleconomista, fatsecret, diariofemenino — no CA page found) | FAQ/glossary section w/ per-type kcal table + schema | P1 |
| historia del turrón / origen / història del torró | es/ca | Med, evergreen+Dec | Med (brand blogs: turronesydulces, manueliborra, turronesmanuelpico) | History section on glossary or dedicated page; CA angle (beteve.cat origin piece exists) | P2 |
| com es fa el torró / cómo se hace el turrón (de Jijona / a la pedra) | ca/es | Med, Dec (recipe intent mixes in) | Med-High (recipe sites: marialunarillos, DAP, bonviveur) | Explainer on process (NOT a recipe site — link out); target "com es fa" informational half | P2 |
| torró IGP / IGP torró d'Agramunt / IGP Jijona | ca/es | Low, evergreen | Low | /torro-agramunt-igp | P1 |
| cuánto dura el turrón / caduca el turrón / com conservar el torró | es/ca | Med, Dec–Jan | Low-Med (elespanol, consumer.es, brand blogs) | FAQ entry (great featured-snippet shape: 15–18 months, per-type table) | P2 |
| torró de crema cremada vs crema catalana / què és el torró de crema | ca | Med (most-Catalan variety; terminology quirk "gema/crema" per Optimot) | Low (recipes + beteve news only) | Glossary entry + category page; cite Optimot/ésAdir naming | **P1** |
| quin torró és típic de Catalunya | ca | Low-Med, Dec | Low (beteve, catalunyacomestible — no structured answer) | Glossary/FAQ | P2 |
| per què es mengen torrons per Nadal / tradició neules i torrons | ca | Low-Med, Dec | Low (beteve owns it) | FAQ / blog | P2 |
| turrón embarazo / turrón para diabéticos / turrón sin gluten | es | Low each, Dec | Low | FAQ entries (health-adjacent: cite sources, no medical advice) | P2 |
| spanish nougat / what is turron | en | Med (US/UK expat + tourist), Dec | High (TasteAtlas, Wikipedia, chefspencil) | English one-pager /en (glossary summary) | P2 |
| turron vs torrone (vs nougat) | en | Low-Med | Low-Med (yummybazaar FAQ, amigofoods) | Same English one-pager section | P2 |

---

## Cluster 4 — Comparison

| Query | Lang | Volume / seasonality | Difficulty | Page | Priority |
|---|---|---|---|---|---|
| torró d'Agramunt vs Xixona / Agramunt o Jijona | ca/es | Low-Med | Low | **/torro-agramunt-vs-xixona (exists)** + ES twin "¿Turrón de Agramunt o de Jijona?" | **P0** |
| turrón 1880 o Delaviuda (cuál es mejor) | es | Med, Dec | Med (okdiario expert picks, merca2, DAP) — adjacent, not our catalog | P3 context mention in comparisons hub | P3 |
| turrón duro o blando (cuál es mejor / diferencias) | es | Med, Dec | Med | Glossary/FAQ + duel framing ("vota tu") | P2 |
| torró X vs torró Y (any pair of the 107 products) | ca | Micro each; **we generate these SERPs** — head-to-head is literally the product | Trivial | Duel/pairing pages or per-product "rivals" section on /torro/{id} | **P1** |
| comparativa torrons / comparador de torrons | ca | Low | Low | Position the app itself as "el comparador de torrons" | P1 |
| comparativa turrones (Vicens) | es | Med | High head (OCU vocabulary) / Low with "Vicens" modifier | ES ranking landing | P2 |
| turrón vicens o torrons roig / vicens o alemany (Agramunt rivals) | ca | Micro | Low | P3 — risky framing vs brands; only editorial neutrality | P3 |
| turrón Dubái vs turrón clásico / mejor turrón estilo Dubái | es | Med, viral (2025's trend; OCU tested it Dec-2025) | Med — freshness game | Trend page/news hook if a Vicens Dubái-style product exists (catalog: "Dubai hazelnut" novelty confirmed in press) | P2 |

---

## Cluster 5 — Seasonal / Event

| Query | Lang | Volume / seasonality | Difficulty | Page | Priority |
|---|---|---|---|---|---|
| torrons nadal 2026 / novetats torrons 2026 | ca | High, Oct–Dec only | Med (cuina.cat sponsored, thenewbarcelonapost, beteve) | Annual "Novetats i resultats Nadal 2026" page, published Oct | **P1** |
| turrones navidad 2026 / novedades turrón 2026 | es | High, Oct–Dec | Med-High | ES twin | P1 |
| calendari d'advent (de) torró / calendario de adviento turrón | ca/es | Low-Med, Nov | **Low — SERP is chocolate calendars (Lacasa, ECI, Planète Chocolat); nobody owns "advent + torró" as an experience** | **Advent-duel feature page — unique product, name the page exactly this** | **P0** |
| calendari advent nadal per adults / calendario adviento gastronómico | ca/es | Med, Nov | High (retail) | P3; mention on advent page | P3 |
| fira del torró agramunt 2026 (dates/programa) | ca | Med, Sept–Oct spike (2025 ed.: Oct 11–12, agramunt.cat) | Med (agramunt.cat, socpetit, escapadaambnens) | Event section on /torro-agramunt-igp (dates + link) — good pre-season traffic | P1 |
| regals de nadal gourmet / qué turrón regalar / cestas navidad turrón | ca/es | High, Dec | High (e-comm) | P3 head; capture "quin torró regalar" question form via ranking page section | P2 |
| loteria del torró? / torró per sant joan? (off-season curiosities) | ca | Micro | Low | FAQ ("es pot comprar torró tot l'any?") | P3 |
| resultats torrorèndum 2026 / guanyador torrorèndum | ca | Micro→Low (created by us; press hook) | Trivial | /premsa + seasonal results page | P0 |
| black friday torró / torró rebaixes | ca | Micro | Low | Skip — transactional | P3 |

---

## Cluster 6 — Question / AI-assistant conversational queries

These rarely show meaningful classic-SERP volume but are exactly what users ask ChatGPT/Perplexity/Gemini/Claude. Goal: be the **cited source**. Current observed assistant behavior (see serp-landscape-es.md §AI-angle): they recite OCU verdicts. Counter-strategy: unique, quantified, dated claims + stable /premsa URL + Dataset/ItemList schema + free data (OCU's comparator is paywalled — assistants can't read it).

Representative conversational queries to satisfy with structured, quotable answers:

- CA: "quin torró recomanes per regalar a algú que no li agrada la xocolata?" · "quin és el millor torró de Torrons Vicens segons la gent?" · "quin torró d'Albert Adrià val més la pena?" · "quina diferència hi ha entre el torró d'Agramunt i el de Xixona?" · "quants tipus de torró existeixen?" · "és el torró de crema cremada típic català?" · "quin torró té menys calories/sucre?" · "on puc votar quin és el millor torró?"
- ES: "¿qué turrón me recomiendas comprar esta Navidad?" · "¿cuál es el turrón más votado de España?" · "¿qué turrón de Vicens está mejor valorado?" · "¿el turrón de Agramunt tiene denominación de origen?" · "¿qué turrón regalo a un diabético?"
- EN: "what's the best turron to buy in Barcelona?" (current SERP: Tripadvisor + ForeverBarcelona + bcn.travel — a "visitor's guide to torró + what locals voted best" EN page could enter) · "what is the difference between turron and torrone?" · "best turron brands" (amigofoods listicle owns it) · "what souvenir food to bring back from Barcelona?"

**Page implications:** every cluster-2/3 page needs an explicit Q&A block (FAQPage schema), first-paragraph direct answers with numbers ("Amb N mil vots, el torró més ben valorat del 2026 és X, amb un ELO de Y"), dates on every claim, and a machine-readable stats endpoint linked from /premsa. Priority **P0 as a cross-cutting requirement**, not a page.

---

## Cluster 7 — Product-level long-tail (107 products, 4 classes)

Each /torro/{id} page should target `torró (de) {name}` + `torró {name} vicens` + `{name} vicens opinions/opiniones`. Micro-volume each, but ~200–400 indexable long-tail queries in aggregate, near-zero competition, and they feed AI assistants per-product facts nobody else publishes (ELO, win-rate, rival duels).

By class (from migrations seed):

1. **Adrià Natura (34 products)** — highest search-worthy names: Aire ("el torró més lleuger del món" — press hook confirmed: financialfood.es), Festucs, Barcelona, Tiramisú, Cervesa, Gintonic, Mojito, Tòfona blanca, Te matcha, Baklava, Cherry Times, Xocolata i xurros… Queries: "torró aire albert adrià", "turrón de cerveza vicens", "torró gintonic". **Note the 000017 rename: the public line name is "Adrià Natura" (was "Essència") — use "Adrià Natura" + "Albert Adrià" in titles; do NOT optimize for the obsolete "Essència Adrià".** Priority P1 (the line has real press gravity: hellochefs, financialfood, sweetpress, Amazon listings).
2. **Novetats (43)** — collab names carry their own brand queries: "torró chupa chups", "torró pernil enrique tomàs", "torró vermut miró", "torró ratafia", "torró baileys", "torró dulce de leche", "torró donuts", "torró marc ribas brutal", "torró tea shop". Micro but ultra-specific = easy #1s and AI-quotable novelty facts. P1 for named collabs, P2 rest.
3. **Clàssics (18)** — map to generic variety queries (crema cremada, massapà, nata i nous, dur/tou ametlla, coco): these pages should link up to the per-category best pages (cluster 2) which carry the volume. P2.
4. **Xocolata (12)** — same: feed "millor torró de xocolata" category page; individual pages P2.

---

## Page gap summary (what to create, in priority order)

| # | Page | Clusters served | Priority |
|---|---|---|---|
| 0 | Fix indexing of torro.cat (Search Console, sitemap, canonical, robots) — brand queries return nothing today | 1 | **P0 blocker** |
| 1 | Evergreen CA answer page "Quin és el millor torró? Rànquing per vots" (title carries: millor torró, rànquing torrons, més votat) | 2, 6 | P0 |
| 2 | ES landing "Ranking de turrones (por votos) — los turrones Vicens más votados" | 2, 1, 6 | P0 |
| 3 | ES twin of Agramunt IGP explainer ("Turrón de Agramunt: qué es, IGP, diferencias con Jijona") | 3, 4 | P0 |
| 4 | Advent-calendar landing "Calendari d'advent del torró" (+ ES) | 5 | P0 |
| 5 | Per-category best pages (xocolata, crema cremada, festuc, praliné, massapà, sense sucre) CA+ES | 2 | P1 |
| 6 | Seasonal "Torrons Nadal 2026: novetats i resultats" (annual, published Oct, year in title) | 5, 2 | P1 |
| 7 | FAQ enrichment: calories, caducitat/conservació, typical-of-Catalonia, diabetics, per-Nadal tradition (FAQPage schema) | 3, 6 | P1 |
| 8 | Adrià Natura line hub + upgraded per-product pages (rivals, ELO, quotable facts) | 7, 1 | P1 |
| 9 | English one-pager: what is torró/turron, turron vs torrone, best turron in Barcelona (locals' vote) | 3, 6 | P2 |
| 10 | History/how-it's-made explainer sections | 3 | P2 |

---

## Snapshot 2026-08-17

Re-measurable observations for the 16 queries newly analyzed in this run. For each: verbatim query, then top results in observed order (Claude WebSearch, US-based — expect ordering drift vs google.es). Date: **2026-08-17**. (For millor torró, millors torrons, torrons Vicens, torró d'Agramunt, torró de Xixona, quin és el millor torró, rànquing torrons, and the ES heads mejor turrón / ranking turrones / comparativa / tipos / Jijona-Alicante / cata / OCU / Mercadona / yema / Dubái / turrón de Agramunt — see the same-dated snapshots in `serp-landscape-ca.md` and `serp-landscape-es.md`.)

### Query: `cuántas calorías tiene el turrón`
1. eleconomista.es — https://www.eleconomista.es/salud-bienestar/nutricion/noticias/12523049/11/23/cuantas-calorias-tiene-el-turron-duro-blando-o-de-chocolate.html
2. elcomercio.pe (Doña Pepa — LatAm noise)
3. tiktok.com/discover/cuantas-calorías-tiene-el-turrón
4. diariofemenino.com — https://www.diariofemenino.com/articulos/dieta/cuanto-engorda-el-turron/
5. turronesydulces.com/blog — información nutricional
6. fatsecret.es — https://www.fatsecret.es/calorías-nutrición/search?q=Turrón
7–8. fitia.app (LatAm)
Facts circulating: ~490 kcal/100 g traditional; blando ≈134 kcal/porción > duro ≈125; chocolate-almendra 573/100 g. No Catalan-language result at all → CA gap open.

### Query: `historia del turrón origen`
1. utp.edu.pe (Peru noise)
2. turronesydulces.com — https://www.turronesydulces.com/historiaturron.htm
3. manueliborra.com — https://www.manueliborra.com/blog/historia-y-origen-de-los-turrones/
4. magic-edu.es
5. turronesmanuelpico.com — https://turronesmanuelpico.com/historia-del-turron-sus-origenes-y-tradicion-familiar/
6. turroneslacolmena.com/en — https://turroneslacolmena.com/en/origin-history-nougat/
7. chocolatescanayas.com
Character: 100% brand-blog content; facts repeated: Arab origin ("turun"), Sexona/Jijona 16th c., Greek athlete legend. No consumer-media or academic page ranks → medium-low bar; a sourced, dated history section can win AI citations.

### Query: `cómo se hace el turrón de Jijona`
1. marialunarillos.com — https://www.marialunarillos.com/blog/receta-turron-jijona-casero.html
2. directoalpaladar.com — https://www.directoalpaladar.com/postres/receta-turron-jijona-popular-dulce-navideno-alicante
3. turronesydulces.com/blog (receta blando casero)
4. bonviveur.com — https://bonviveur.com/es/recetas/turron-de-jijona-o-blando
5. lolitalapastelera.com
6. turronessirvent.com — https://turronessirvent.com/turron-de-jijona-origen-y-receta/
7. cosascaseras.com
8. jijona.com/el-turron/
9. turroneslacolmena.com/en/how-to-make-turron-recipe/
Character: recipe intent dominates; informational "industrial process" half (boixet, molinos de granito, 24h reposo) underserved → target only that half.

### Query: `turrón sin azúcar`
1. amigofoods.com (US shop)
2. dulcesdiabeticos.com — https://dulcesdiabeticos.com/turron-jijona-blando-sin-azucar/
3. hola.com (receta sin azúcar, dated 2025-12-14)
4. dulcesdiabeticos.com/tag/turron/
5. colomagarcia.com — https://colomagarcia.com/informacion/turron-sin-azucar-beneficios/
6. elalmendro.com/en (product)
7. diegoverdu.com (product)
8. turronesydulces.com (category)
Character: e-comm + niche diabetic blog; sweeteners facts (maltitol/stevia). No ranking/comparison of sugar-free turrones exists.

### Query: `turrón 1880 o delaviuda cuál es mejor`
1. okdiario.com — https://okdiario.com/economia/nunca-imaginarias-cual-mejor-turron-espana-experto-elige-alcampo-dia-delaviuda-13729299
2. merca2.es — https://www.merca2.es/2020/12/19/delaviuda-1880-el-almendro-turron-528169/
3. directoalpaladar.com — por qué 1880 se anuncia como "el más caro del mundo"
4. gastronomicspain.com/blog/en (1880 brand piece)
5. directoalpaladar.com (quién fabrica los turrones de cada supermercado)
6. ocu.org comparador — https://www.ocu.org/alimentacion/dulces/comparador-turron/delaviuda-turron-duro/72505_110347
7. turron1880.com
8–10. ebay.de / wanderlog / justia (noise)
Character: brand-vs-brand comparison queries resolve to expert picks + price angles (1880 ≈47.5 €/kg vs Delaviuda ≈22.8 €/kg). Not our catalog → P3 confirmed.

### Query: `Albert Adrià turrón Vicens Essència`
1. hellochefs.es — https://www.hellochefs.es/pasteleria/v/albert-adria-crea-los-turrones-natura-junto-a-torrons-vicens
2. financialfood.es — https://financialfood.es/albert-adria-y-torrons-vicens-se-unen-para-crear-el-turron-mas-ligero-del-mundo/ ("Turrón Aire, el más ligero del mundo")
3. sweetpress.com — https://www.sweetpress.com/actualidad/torrons-vicens-la-revolucion-artesana-del-turron-AISP12323
4. amazon.com — Natura Collection product (B0D576BWSF)
5–6. vicens.com/en (Natura products: nougat cream tubes, cerveza+tiramisú case)
7. albertadria.com/en/id/turrones-albert-adria/
8. vicens.com/en (Barcelona nougat 150g)
9. vicens.com/en/albert-adria-6854
10. en.wikipedia.org (Albert Adrià)
Character: press + brand + Amazon; line name in the wild is "Natura"/"Adrià Natura" (confirms migration 000017). Collaboration since 2013. No independent review/ranking of the line exists → open for "quin torró d'Adrià és el millor?".

### Query: `spanish nougat turron`
1. tasteatlas.com/turron
2. en.wikipedia.org/wiki/Turrón
3. chefspencil.com — https://www.chefspencil.com/turron-spanish-almond-nougat/
4. carolinescooking.com (recipe)
5. lespanola.com blog
6. 196flavors.com/turron/
7. sundaybaker.co (Alicante recipe)
8. turronesydulces.com/en — nougat vs turrón
Character: EN head owned by TasteAtlas/Wikipedia/recipe blogs. P2: only worth a summary page.

### Query: `turron vs torrone difference`
1. yummybazaar.com — https://yummybazaar.com/blogs/blog/frequently-asked-questions-about-turron-and-torrone (FAQ format wins)
2. thegreencreator.com (recipe, noise)
3. blog.amigofoods.com — turron/nougat/halva difference
4. blog.amigofoods.com — Spanish turron ultimate guide
5. researchgate.net — instrumental texture study (Turrón vs Torrone vs French nougat)
6. natachasanzcaballero.com
7. en.wikipedia.org/wiki/Turrón
8. mrdach.com
Facts circulating: turrón nut ratio 60–64% > torrone; torrone chewier, more flavorings; Spanish regulation stricter. Low bar, FAQ format wins.

### Query: `best turron brands Spain`
1. blog.amigofoods.com — https://blog.amigofoods.com/index.php/spanish-foods/best-spanish-turron-brands/ (7 best brands: 1880, El Almendro, Sanchis Mira…)
2. pinterest.com (same piece)
3. blog.amigofoods.com (ultimate guide)
4. foreverbarcelona.com — https://www.foreverbarcelona.com/where-to-buy-turron-in-barcelona-shops/
5. legourmetcentral.com
6–9. US shops (despanabrandfoods, ibericotaste, turronesydulces/en, laespanolameats)
Character: US-importer content. Vicens barely present → EN page listing "brands + what Catalans actually vote" has an angle.

### Query: `"torrorendum" OR "torrorèndum" torro.cat`
**Zero relevant results.** SERP fell back to dictionary entries for "torró" (diccionari.cat, enciclopedia.cat, ca.wiktionary), Wikipedia villages (Torroella…), barcelona.cat culturapopular, beteve.cat, tiktok #torrocat tag. **Brand invisible in index — confirms the P0 indexing blocker also seen in serp-landscape-ca.md.**

### Query: `calendario de adviento turrón chocolate`
1. tienda.lacasa.es — https://www.tienda.lacasa.es/turrones-lacasitos/1126-turron-crujiente-calendario-de-adviento.html (a turrón bar cut into 24 numbered pieces!)
2. elcorteingles.es/aptc (mejores calendarios de Adviento)
3. hipercor.es/aptc (same)
4. directoalpaladar.com/seleccion — mejores calendarios adviento 2024
5. larecetadelafelicidad.com (recipe, noise)
6. planetechocolat.com/es
7. calendariosdeadviento.com/chocolate/
8. bomboneriapons.com/en (chocolate advent calendar)
Character: all retail chocolate calendars; **no interactive/experience result**. "Calendari d'advent del torró" (CA) is effectively virgin territory.

### Query: `torró de crema catalana`
1. greatbritishchefs.com — Torró de Gema Cremada recipe (EN!)
2. receptescartesianes.cat/recipes/405
3. beteve.cat — https://beteve.cat/estils-de-vida/millor-torro-crema-cremada-gotic-colmena-barcelona/
4. Optimot (gencat) — terminology card "torró de gema o torró de crema?"
5. beteve.cat — "Així s'elabora el millor torró de crema cremada d'Espanya"
6. receptes.cat/recepta3403
7. saltdelcolom.com (product)
8. esadir.cat/entrades/fitxa/node/torrodecrema (naming)
9. receptescartesianes.cat/recipes/222
10. en.wikipedia (Crema catalana)
Character: recipes + terminology + prize news; no explainer/ranking. The most-Catalan variety is unowned → strong P1.

### Query: `torró sense sucre`
1. compraonline.bonpreuesclat.cat — **VICENS Assortit torró sense sucre** (retail)
2. eliasseleccio.cat (Pablo Garrigós product)
3. endocrino.cat (Dexeus hospital blog — sense sucres afegits)
4. torronsroig.com/ca/categories/sense-sucre
5. turrones-sirvent.com/ca (product)
6. planellesdonat.com (product)
7. botigatorrons.com (M.Mira product)
8–9. bomboneriapons.com (products/collection)
Character: pure e-comm + one clinical blog; no comparison/ranking content in CA. Vicens has a sugar-free assortment (Bonpreu listing) → category page viable.

### Query: `torrons nadal 2025 novetats`
1. totnens.cat — Fira dels Torrons de Llinars del Vallès 2025
2. escapadaambnens.com — Nadal a Girona 2025
3. cuina.cat — **"Patrocinat: Torrons Vicens marca tendència aquest Nadal"** — https://www.cuina.cat/actualitat/noticies/torrons-vicens-marca-tendencia-aquest-nadal_404154_102.html (Vicens 2025 novelties: Nocilla collab, Ángel León vi/ametlla fregida, Dacosta panettone/orxata, Jordi Roca Cromatisme Taronja, Adrià "Aire de Festuc" & "Aire de Cacau i Avellana", Dubai avellana)
4. thenewbarcelonapost.cat — millors torrons d'autor 2025
5. totnens.cat — calendari advent Sabadell
6. tarragonadigital.com — fires de Nadal
7. metropoliabierta/elespanol ca — llums de Nadal
8. beteve.cat — La Campana torró pa amb oli i xocolata
Character: "novetats" SERP = family-agenda sites + sponsored brand content + author-torró listicle. An independent annual novelty round-up with vote data would be differentiated. Also: Fira de Llinars = Dec 6–8, 2025 (second CA torró fair worth covering).

### Query: `best turron to buy in Barcelona`
1. tripadvisor.com — Planelles Donat review
2. cataloniaholidaylettings.co.uk blog
3. bcn.travel — https://www.bcn.travel/turron-barcelona/ ("Turron Barcelona: The Ultimate Guide")
4. tripadvisor.com — **Torrons Vicens (Madrid page) titled "Best Turrones in Barcelona!"**
5. foreverbarcelona.com — where to buy turron in Barcelona
6. barcelona-top-travel-tips.com — handmade turrón
Character: tourist content; shops named: Planelles Donat, Casa Colomina, Sirvent, La Campana. Vicens appears only via Tripadvisor. EN "what locals voted best" page has a clear wedge.

### Query: `turrón de pistacho`
1. okdiario.com/recetas — https://okdiario.com/recetas/turron-pistacho-casero-15905251
2. uncomo.com (receta)
3. juanideanasevilla.com (Thermomix)
4. amazon.com (US product)
5. comedera.com
6. conlaszarpasenlamasa.com
7. turroneslacolmena.com/turron-pistacho/
8. turronesydulces.com/en (product)
Character: recipe-dominated, rising trend ingredient (Dubái effect); no "mejor turrón de pistacho" ranking exists → per-category best page (Vicens: Festuc, Festucs Adrià, xocolata+festuc x3, iogurt+festuc…) is well-armed.

### Query: `fira del torró d'Agramunt 2025 dates`
1. escapadaambnens.com — Fira del Torró i la Xocolata a la Pedra 2025 (Oct 11–12, 36th ed.)
2. agramunt.cat — https://www.agramunt.cat/actualitat/agenda/xxxvi-fira-del-torro-i-la-xocolata-a-la-pedra
3. socpetit.cat
4. festacatalunya.cat
5. agramunt.cat (permanent fira page)
6. territoris.cat (35a edició, 2024)
7. agramuntesports.cat (Cursa del Torró)
8. femturisme.cat/en
Character: agenda sites + ajuntament. Fira 2026 dates will be the pre-season traffic magnet (publish Sept).

### Query: `cuánto dura el turrón caduca`
1. elespanol.com/cocinillas — https://www.elespanol.com/cocinillas/actualidad-gastronomica/20241210/caduca-turron-puedo-comerlo-ano/903659632_0.html
2. consumer.es — https://www.consumer.es/seguridad-alimentaria/los-turrones-que-duran-mas.html
3. cestas-marti.com
4. turronpico.com/cuando-caduca-el-turron/
5. galeraregalos.com/blog
6–8. ebay/diariodecuyo noise
Facts circulating: 15–18 months typical; duro 12–18m, blando 6–12m, chocolate 3–9m; "consumo preferente", not caducidad. Clean featured-snippet target.

### Query: `quin torró és típic de Catalunya`
1. beteve.cat — https://beteve.cat/estils-de-vida/tradicio-neules-torrons-nadal-origen/
2. catalunyacomestible.com/els-torrons/
3. enciclopedia.cat (torró entry)
4. blogspot noise
5. naciodigital.cat — gastronomia tradicional Nadal en català
6. ca.wiktionary
7. blocs.mesvilaweb.cat (Jaume Fàbrega)
8. turronesydulces.com/blog/ca — turró/terró/torro naming
Answer circulating: Agramunt (avellana + pa d'àngel) is THE Catalan one; crema cremada most popular locally. No structured answer page → FAQ target.

### Query: `qué turrón regalar navidad gourmet`
1. saborgourmet.com
2. amazon.com (US noise)
3. colomagarcia.com — https://colomagarcia.com/coloma-garcia/turrones-para-regalar-dulces-y-exquisitos/
4. saborazogourmet.com (cestas)
5. manueliborra.com — mejores turrones para regalar
6. turronessaxum.com/regalar-turron
Character: brand-blog gift guides; no independent "which to gift, per recipient type" content → a data-backed gift-picker section ("el més votat per a…") is a P2 wedge.

### Infrastructure notes (this run)
- Autocomplete endpoints (Google suggest, DDG ac, Bing osjson) → proxy CONNECT 403 on 2026-08-17; re-try from unrestricted network for true suggest mining.
- torro.cat and vicens.com direct fetches remain blocked by sandbox egress (per companion docs) — on-page audits need another environment.
