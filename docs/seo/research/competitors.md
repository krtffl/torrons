# Competitor Analysis — the "which turrón/torró is best" attention space

**Project:** Torrorèndum (https://torro.cat) — Catalan-language fan web app: head-to-head duels between Torrons Vicens products, ELO rankings, brackets, advent-calendar duel, shareable stats. Independent, non-commercial, **not** official Torrons Vicens property.
**Date of research:** 2026-08-17 (off-season; next peak: Nadal/Navidad 2026, Nov–Jan).
**Method:** WebSearch + WebFetch. **Caveats:** (1) WebSearch is **US-based** — Spanish/Catalan queries still surface the es/ca ecosystem, but rankings may differ from google.es / google.cat results seen by real users; treat orderings as approximate. (2) The research sandbox's egress proxy **blocked direct fetches** of ocu.org, vicens.com, turronesydulces.com and es.wikipedia.org, so their on-page details come from SERP snippets, secondary coverage and prior knowledge — flagged inline where relevant.

---

## 1. Consumer-testing authorities: OCU

**Who:** Organización de Consumidores y Usuarios — Spain's dominant consumer-testing body. Its turrón vertical lives at `ocu.org/alimentacion/dulces/`:
- Methodology page: https://www.ocu.org/alimentacion/dulces/asi-comparamos-turrones
- Interactive comparator: https://www.ocu.org/alimentacion/dulces/comparador-turron
- Buying guide: https://www.ocu.org/alimentacion/dulces/como-elegir-turron
- Chocolate-turrón report: https://www.ocu.org/alimentacion/dulces/informe/turron-chocolate
- Press note (media seed): https://www.ocu.org/organizacion/prensa/notas-de-prensa/2020/estudio-comparativo-turrones

**Methodology & scale:** ~154 turrones analyzed (duro, blando, chocolate crujiente, sin azúcar, IGP Jijona): labeling analysis, nutritional composition, price comparison, and **blind expert tasting**. Headline verdicts: best Jijona = **1880** (70% almond, Calidad Suprema); best value = **Flor de Navidad (Aldi)** / **Dor (Lidl)**; most chocolate turrones rated poor quality. (Sources: ocu.org pages above via SERP; echo in https://www.directoalpaladar.com/actualidad-1/ocu-no-deja-titere-cabeza-su-analisis-turrones-almendra-mitad-duros-blandos-suspende-cata, https://www.elespanol.com/sociedad/consumo/20241220/mejor-turron-supermercado-navidades-ocu-cuesta-euros/910159006_0.html, https://theobjective.com/economia/consumo/2024-12-20/ocu-mejor-turron-supermercados/.)

**Paywall model:** freemium — headline "3 best" conclusions are free + pushed via press notes; the full comparator (per-product scores) is members-only. Subscription is tiered (~2 €/mo intro escalating to ~23,90 €/mo from year 3; no permanence) — https://www.ocu.org/info/precios-suscripcion, third-party review https://preahorro.com/como-ahorrar/ocu-analisis-y-opinion-merece-la-pena/. Recurring criticism: cost and commercial-agreement conflicts (https://puntua.net/empresa/opiniones-sobre-organizacion-de-consumidores-unidos-ocu-ocu-org/).

**Media echo (their real moat):** every Nov–Dec, dozens of outlets in ES **and CA** rewrite the OCU press note: elconfidencialdigital.com, canarias7.es, theobjective.com, losreplicantes.com, elespanol.com, and Catalan media e-noticies.cat, catalunyapress.cat, naciodigital.cat, vadegust.cat. OCU thus owns "según la OCU" as the trust anchor for both classic SERPs and AI-assistant answers.

**vs Torrorèndum:** They have: massive domain authority, lab + expert methodology credibility, yearly refresh, national press pipeline, an interactive comparator. They lack: Catalan-language depth (echo is via third parties), any community/vote data, per-product pages for artisan/Vicens catalog (they test supermarket tablets), free full results, embeddability, fun. We can never beat their "lab test" authority — we can own the complementary axis: **"what people actually prefer" (revealed preference, N votes, live), in Catalan, free and fully open.**

## 2. Food media annual rankings / catas

- **Directo al Paladar (Webedia)** — the strongest editorial player. Annual blind-tasting franchises by their "mayor experto en dulces navideños": best turrones duros de supermercado (https://www.directoalpaladar.com/actualidad-1/cuales-mejores-turrones-duros-supermercado-mayor-experto-dulces-navidenos-dap), best blandos marca blanca (https://www.directoalpaladar.com/actualidad-1/cuales-mejores-turrones-blandos-marca-blanca-nuestro-mayor-experto-dulces-navidenos), plus the evergreen traffic magnet "quién fabrica los turrones de cada supermercado" (https://www.directoalpaladar.com/consumidores/mercadona-lidl-carrefour-quien-fabrica-turrones-cada-supermercado-esta-navidad-cuales-baratos-3). Structure: one expert, blind cata, ranked list with prices, refreshed each November (URL slug versioned "-3"). Findings echo OCU too. Verdicts observed: best duro = DIA (made by Almendra y Miel, IGP Jijona); best blando = Lidl (made by Delaviuda).
- **El Español** — mixes OCU-rewrite pieces (https://www.elespanol.com/ciencia/nutricion/20211201/este-mejor-turron-supermercado-segun-ocu/631437456_0.html) with own expert catas (maestro repostero Antonio Palomo ranking Alcampo > Carrefour > Mercadona: https://www.elespanol.com/reportajes/20211213/mejores-turrones-antonio-palomo-lidl-carrefour-mercadona/633687578_0.html) and covers the Vicens creative championship (https://www.elespanol.com/cocinillas/actualidad-gastronomica/20241013/mejor-turron-creativo-espana-hace-tarragona-sabor-texturas-madagascar/892910881_0.html).
- **The New Barcelona Post** — bilingual ES/CA annual "millors torrons d'autor" listicle (https://www.thenewbarcelonapost.com/mejores-turrones-2025/, https://www.thenewbarcelonapost.cat/millors-torrons-autor-2025/) — artisan/pastry-chef angle, heavily features Vicens & L'Atelier.
- **ElNacional.cat (Gourmeteria)** — Catalan listicle "Els millors torrons de Catalunya" (https://www.elnacional.cat/ca/gourmeteria/llistes/millors-torrons-catalunya-nadal-tastar-familia_1334575_102.html) — the closest thing to a Catalan-language editorial ranking.
- **Others in the annual churn:** okdiario.com gastronomía, hola.com cocina, cadena100.es, codigounico.com ("8 mejores turrones artesanales del mundo": https://www.codigounico.com/placeres/gastronomia/mejores-turrones-del-mundo-artesanos-espana.html), gastronomistas.com, HuffPost's YouTube supermarket cata (https://www.youtube.com/watch?v=5ESZ17rGV2U), e-commerce content plays like dulcealmacen.com "guía completa 2025" (https://dulcealmacen.com/mejores-turrones-artesanos-navidad-2025/). Notably, La Vanguardia Comer, ABC and bonviveur did **not** surface in the US-index top results for the tested queries — they participate but are not the visible leaders from this vantage point.

**Pattern:** annual, dated listicles (year in title), one expert or pure editorial, supermarket price hooks, published late Oct–Dec, refreshed by re-versioning URLs. **They have:** news-domain authority, Discover/News distribution, yearly freshness signals. **They lack:** any persistent dataset, interactivity, Catalan depth (except ElNacional/TNBP), per-product pages, methodology beyond one palate. **We have:** continuously updating ELO from thousands of votes (a genuinely different data story they could cite/embed).

## 3. Supermarket own-brand ranking content

The highest-volume seasonal queries are supermarket-anchored ("mejor turrón Mercadona/Lidl/Aldi"). Who captures them: OCU-echo pieces (e-noticies.cat: https://e-noticies.cat/es/sociedad/ocu-tiene-claro-mejor-y-peor-turron; canarias7.es: https://www.canarias7.es/gastronomia-c7/mejores-turrones-mercadona-20211202101934-nt.html), DAP's "quién fabrica" piece, influencer catas on TikTok (e.g. https://www.tiktok.com/@silvia_fatfood/video/7422248612819766560, TikTok discover page "Ranking Mejores Turrones 2025") and okdiario influencer-cata write-ups (https://okdiario.com/economia/dos-influencers-catan-turrones-almendras-aldi-lidl-veredicto-inapelable-13815429). Key evergreen fact set: Mercadona ← Antiu Xixona; Lidl "Dor" & Aldi "Flor de Navidad" ← Delaviuda / José Garrigós. The supermarkets themselves publish no ranking content — it's all third-party. **Relevance to us:** adjacent but off-catalog (Torrorèndum is Vicens-only); useful mainly as internal-linking context (Agramunt vs supermarket Xixona comparisons) rather than a battlefield to fight on.

## 4. Brand publishers

- **Torrons Vicens (vicens.com)** — e-commerce with an active ES/CA/EN blog (recipes, brand news, RAC1 solidarity turrón, awards) and a separate event property **campeonatoturron.vicens.com** (Campeonato Nacional "Turrón Creativo" Maestro Turronero Àngel Velasco), which generates national press every autumn (gastronomistas.com, okdiario, El Español, hola.com all covered 2024/2025 editions; 2025 winner: Brandy Smoke, L'Atelier Barcelona). Has a Spanish Wikipedia article (https://es.wikipedia.org/wiki/Torrons_Vicens), strong social presence, chef collabs (Albert Adrià, José Andrés). Blog is brand/news-oriented, **not** optimized for informational queries ("tipos de turrón", "mejor turrón") and offers no ranking/comparison content — they can't rank their own products against each other without commercial awkwardness. (Direct fetch blocked; based on SERP + https://www.last.app/recursos/blog/torrons-vicens-tradicion-innovacion-exito.)
- **Xixona brands:** turron1880.com (slogan-led brand site, "el turrón más caro del mundo" — see analysis at https://www.directoalpaladar.com/ingredientes-y-alimentos/que-turron-1880-se-anuncia-como-caro-mundo-hay-otros-que-cuestan); El Lobo (same Sirvent family/Sanchis Mira group); Delaviuda (white-label giant behind Lidl/Aldi); Pablo Garrigós Ibáñez (premium/innovation positioning). Their sites are catalog-first; none does serious informational SEO.
- **Sector blog: Made in Jijona (madeinjijona.com)** — "el blog del turrón": deep history/heritage/brand-story content (origin of Jijona blando, nougat vs turrón, TV ad history, #turronemoji campaign, brand founding dates). Strongest informational-content competitor on the Xixona side; no rankings, no interactivity, Spanish-only, Jijona-centric (weak on Agramunt/Catalan angle). Direct rival to our /torro-agramunt-igp and /tipus-de-torrons style content — and a prime **link-exchange/citation target**.
- **Affiliate/e-commerce content:** turronesydulces.com — "Las mejores marcas de turrón" (https://www.turronesydulces.com/marcas) and a **Catalan version** "Els millors torrons segons valoracions i marques" (https://www.turronesydulces.com/cat/els-millors-torrons) ranking by customer star-ratings on their shop. It ranked #2 for the Catalan query "millors torrons" in our test — proof that thin, commercially-biased Catalan content currently wins for lack of alternatives. Also dulcealmacen.com guide pages. **They have:** transactional intent capture. **They lack:** independence (they sell what they rank), sample size, methodology.

## 5. Interactive / community rankings & voting

**Finding: this category is effectively empty.** Searches for turrón brackets/tournaments/polls ("mundial de turrones", bracket eliminatorias, reddit encuestas, TV3/RAC1 "torró preferit" polls) returned **no dedicated interactive competitor** — only generic bracket tools (challonge.com), World-Cup bracket apps, and unrelated content. No Reddit ranking thread surfaced in the US index; TikTok "ranking" content exists but is one-off video, not a persistent voting product. Radio/TV do ephemeral polls at best. **Torrorèndum is, as of 2026-08-17, the only persistent head-to-head voting/ELO product in the turrón space that this research could find.** This is the core defensible position — but it also means zero existing search demand for "torró duel/votació"-type queries: demand must be captured via the informational queries (categories 1–2, 6) and via press/social distribution, then converted into the interactive loop.

## 6. Wikipedia / knowledge-panel sources

- **es.wikipedia.org/wiki/Turrón_de_Agramunt** — ranks #1 for Agramunt-history queries; covers 1741 Siscar letters, IGP since 2001/2002, ingredients. Interlanguage: Catalan (Viquipèdia "Torró d'Agramunt") and the English "Turrón" article (https://en.wikipedia.org/wiki/Turr%C3%B3n) feed AI-assistant answers and Google knowledge panels.
- **es.wikipedia.org/wiki/Torrons_Vicens** — the brand has its own article (surface for "Torrons Vicens" knowledge panel).
- Secondary knowledge-ish sites filling the SERP: espanafascinante.com, arecetas.com, koinecommerce.com/blog-cerespain, balansiya.com, restaurantatipic.com (an Agramunt restaurant blog!), and Catalan linguistic sources (beteve.cat on the word "torró", ca.wiktionary.org, dsff.uab.cat).
- **Implication for AI-citation goal:** LLMs currently ground turrón answers in Wikipedia + OCU-echo + DAP. Getting torro.cat cited requires (a) being referenced FROM these pages (Viquipèdia external links, /premsa data being citable), (b) unique data no one else has (vote counts, ELO time series) that journalists and Wikipedia editors can cite with a stable URL.

---

## Competitive-gap matrix

| Capability | OCU | Food media (DAP/ElEspañol/ElNacional) | Supermarket-echo/TikTok | Brand publishers (Vicens/Xixona) | turronesydulces (affiliate) | Wikipedia | **Torrorèndum** |
|---|---|---|---|---|---|---|---|
| Domain authority / backlinks | ●●● | ●●● | ●● | ●● | ● | ●●● | ○ |
| Methodology credibility | ●●● (lab+cata) | ●● (expert cata) | ○ | ○ (self-interested) | ○ | ●● (sourced) | ●● (transparent ELO, large N) |
| Independence / no paywall | ○ (paywall) / ●● | ●● (ads) | ● | ○ | ○ | ●●● | ●●● |
| Annual freshness signal | ●●● | ●●● | ●●● | ●● | ● | ● | ●●● (live, continuous) |
| Catalan-language depth | ○ | ● (ElNacional/TNBP only) | ● | ●● | ● (thin) | ●● (Viquipèdia) | ●●● |
| Per-product depth (Vicens catalog) | ○ | ○ | ○ | ●● (catalog, no comparisons) | ○ | ○ | ●●● |
| Interactivity / voting / community data | ● (members comparator) | ○ | ○ (ephemeral) | ○ | ● (star ratings) | ○ | ●●● (unique) |
| Embeds / shareable widgets | ○ | ○ | ○ | ○ | ○ | ○ | ●●● (potential) |
| Press pipeline / media echo | ●●● | ●●● (they ARE the echo) | ●● | ●● (championship PR) | ○ | n/a | ○ (must build via /premsa) |
| Multilingual reach (ES/EN) | ●● | ●● | ● | ●● | ● | ●●● | ○ (CA only today) |
| AI-assistant citation likelihood today | ●●● | ●● | ○ | ● | ○ | ●●● | ○ |

## The 5 biggest exploitable gaps

1. **The interactive/community lane is empty.** Nobody — not OCU, not any media, not any brand — has a persistent public voting/ELO ranking for turrón. Own it loudly: publish the methodology page ("com funciona l'ELO del Torrorèndum") as the credibility counterpart to OCU's "así comparamos", and pitch the dataset itself ("X,000 votes by Catalans rank Vicens' torrons") to the very outlets that annually rewrite OCU (elnacional.cat, naciodigital, vadegust, catalunyapress, e-noticies) — they demonstrably need fresh turrón angles every December.
2. **Catalan informational SERP is weak.** "millors torrons" is currently won by one ElNacional listicle and a thin affiliate page (turronesydulces.com/cat) that ranks despite selling what it rates. High-quality, independent Catalan pages (rankings + /tipus-de-torrons + Agramunt content) can plausibly take these SERPs and become the default Catalan citation for AI assistants, which today have almost nothing native in ca.
3. **OCU's paywall + supermarket-only scope.** OCU's full results are members-only and cover supermarket tablets, not artisan/Vicens products. Position Torrorèndum as the free, open, always-on complement: "OCU tests the lab; el poble vota el gust" — and publish open data (CSV/JSON on /premsa) that journalists can use without a subscription, something literally no competitor offers.
4. **Annual-freshness arms race can be jumped, not joined.** All media rely on a once-a-year cata with a dated URL. A live ranking updated daily — plus a yearly frozen "Classificació Nadal 2026" edition page (mirroring their `mejores-turrones-2026` pattern so it can rank for those queries and be diffed year over year) — gives both the evergreen and the freshness signal. Ship the season page **before November**, when the churn begins.
5. **Nobody offers embeds or per-product comparison depth.** Media write about products with no product pages; Vicens has product pages but can't rank/compare them. Torrorèndum's /torro/{id} pages + embeddable duel/ranking widgets (for blogs, and even for Vicens' own social/PR — the creative-championship coverage shows appetite) are an unclaimed backlink engine. Every embed is a link; every December listicle writer is a potential embedder. (Guardrail: keep the "independent fan project, not official Torrons Vicens" disclosure on every embed and page — it is also the trust signal that distinguishes us from turronesydulces-style self-interested rankings.)

---

## Snapshot 2026-08-17

Re-measurable observations. All queries run on 2026-08-17 via US-based WebSearch (geo-bias caveat: orderings approximate vs google.es/ca). Domains/URLs listed in the order observed.

**Q1: `OCU comparativa mejores turrones análisis`**
1. ocu.org/alimentacion/dulces/asi-comparamos-turrones 2. ocu.org/alimentacion/dulces/comparador-turron 3. ocu.org/alimentacion/dulces/como-elegir-turron 4. ocu.org/organizacion/prensa/notas-de-prensa/2020/estudio-comparativo-turrones 5. ocu.org/alimentacion/dulces/informe/turron-chocolate 6. wanderlog.com (noise)

**Q2: `mejores turrones 2025 cata ranking`**
1. tiktok.com/discover/ranking-mejores-turrones-2025 2. okdiario.com/gastronomia/mejor-turron-2025-…-15910250 3. thenewbarcelonapost.com/mejores-turrones-2025/ 4. hola.com/cocina/noticias/20251126869522/… 5. gastronomistas.com/latelier-barcelona-…-mejor-turron-creativo-2025/ 6. cadena100.es/…/turrones-raros-2025-… 7. turronesydulces.com/marcas

**Q3: `Directo al Paladar mejores turrones supermercado cata`**
1. directoalpaladar.com/actualidad-1/cuales-mejores-turrones-duros-supermercado-… 2. elconfidencialdigital.com/…/son-mejores-turrones-supermercado-ocu/… 3. directoalpaladar.com/actualidad-1/cuales-mejores-turrones-blandos-marca-blanca-… 4. directoalpaladar.com/viajes/donde-comprar-turron-artesano-madrid-… 5. directoalpaladar.com/consumidores/mercadona-lidl-carrefour-quien-fabrica-turrones-…-3 6. directoalpaladar.com/actualidad-1/ocu-no-deja-titere-cabeza-… 7. directoalpaladar.com/actualidad-1/mejor-turron-creativo-…

**Q4: `mejores turrones Mercadona Lidl Aldi comparativa`**
1. tiktok.com/@silvia_fatfood/video/7422248612819766560 2. e-noticies.cat/es/sociedad/ocu-tiene-claro-mejor-y-peor-turron 3. canarias7.es/gastronomia-c7/mejores-turrones-mercadona-20211202101934-nt.html 4. okdiario.com/economia/dos-influencers-catan-turrones-…-13815429 5. directoalpaladar.com/consumidores/…-quien-fabrica-turrones-…-3 6. elespanol.com/sociedad/consumo/mercadona-lidl-carrefour-mejor-turron-…/813918670_0.html 7. youtube.com/watch?v=5ESZ17rGV2U (HuffPost) 8. en.wikipedia.org/wiki/Mercadona

**Q5: `torronsvicens.com blog turró Vicens noticias contenidos`**
1. vicens.com/en 2. x.com/torronsvicens 3. campeonatoturron.vicens.com/bases-del-campeonato/ 4. vicens.com/en/torrons-vicens/quality-tradition-innovation/ 5. youtube.com channel 6. vicens.com/en/blog/…national-marketing-awards 7. last.app/recursos/blog/torrons-vicens-… 8. instagram.com/torronsvicens 9. vicens.com/es/blog/bloggers/2?page=5

**Q6: `"millors torrons" rànquing cata Nadal català`**
1. elnacional.cat/ca/gourmeteria/llistes/millors-torrons-catalunya-nadal-tastar-familia_1334575_102.html 2. turronesydulces.com/cat/els-millors-torrons 3. catalunyapress.cat/…/4108324-aquests-els-millors-torrons-…-locu 4. thenewbarcelonapost.cat/millors-torrons-autor-2025/ 5. vadegust.cat/aldia/tres-millors-torrons-…-ocu-66403/ 6. naciodigital.cat/viure-be/alimentacio/…_2066170_102.html 7. valenciaextra.com/valencia/els-millors-torrons-de-valencia_257216_102.html

**Q7: `reddit mejor turrón favorito encuesta ranking`** — no Reddit results surfaced. 1. turronesydulces.com/marcas 2. en.wikipedia.org/wiki/Turrón 3–7. ebay/walmart/statista noise.

**Q8: `El Español mejores turrones supermercado ranking cata expertos`**
1. directoalpaladar.com (duros) 2. elespanol.com/sociedad/consumo/20241220/…-ocu-…/910159006_0.html 3. elespanol.com/reportajes/20211213/mejores-turrones-antonio-palomo-…/633687578_0.html 4. elespanol.com/ciencia/nutricion/20211201/…-ocu/631437456_0.html 5. losreplicantes.com/articulos/…-ocu/ 6. directoalpaladar.com (blandos) 7. theobjective.com/economia/consumo/2024-12-20/ocu-mejor-turron-supermercados/

**Q9: `Wikipedia torró Agramunt turrón historia`**
1. es.wikipedia.org/wiki/Turrón_de_Agramunt 2. es.wikipedia.org/wiki/Torrons_Vicens 3. espanafascinante.com/gastronomia/turron_de_agramunt/ 4. arecetas.com/articulos/d-o-turron-de-agramunt-… 5. koinecommerce.com/blog-cerespain/turron-de-agramunt/ 6. restaurantatipic.com/es/b/…/el-turron-de-agramunt-historia-18-3 7. balansiya.com/herencia-turrones/

**Q10: `bonviveur La Vanguardia mejores turrones artesanos navidad prueba`** — neither bonviveur nor La Vanguardia surfaced. 1. elespanol.com/cocinillas/…/mejor-turron-creativo-espana-…/892910881_0.html 2. codigounico.com/placeres/gastronomia/mejores-turrones-del-mundo-artesanos-espana.html 3. thenewbarcelonapost.com/mejores-turrones-2025/ 4–5. dulcealmacen.com guide pages.

**Q11: `votación torneo turrones bracket eliminatorias twitter "mundial de turrones"`** — zero relevant results (only World-Cup bracket apps, challonge.com, wrestling). **No interactive turrón-voting competitor exists in the index.**

**Q12: `"torro.cat" OR "Torrorèndum" torró duel votar`** — **torro.cat did not appear at all** (only Viquipèdia "Torró", wiktionary, beteve.cat, TikTok tag #torrocat, noise). As seen from the US index on 2026-08-17, the site has no visibility for its own brand terms.

**Q13: `enquesta "quin és el teu torró preferit" TV3 RAC1 catalans`** — no relevant poll found (radioteca.cat noise, racocatala.cat old forum, Wikipedia TV articles).

**Other observations to re-check next run:**
- OCU turrón sample size claim: 154 products (via SERP snippet of ocu.org, 2026-08-17).
- OCU membership pricing: 2 €/mo intro → 7,45 → 13,95 → 23,90 €/mo (ocu.org/info/precios-suscripcion via SERP, 2026-08-17).
- DAP "quién fabrica" URL at version `-3`.
- 2025 award winners circulating: Brandy Smoke / L'Atelier Barcelona (creative), Pastisseria Turull (yema).
- Egress-blocked (unverifiable directly this run): ocu.org, vicens.com, turronesydulces.com, es.wikipedia.org.
