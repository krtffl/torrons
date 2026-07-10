# Prompt 22 — Comparativa: Agramunt vs Xixona two-column desktop layout

**Set:** C (cross-cutting addendum, continuing [21](21-desktop-wide-viewport-treatment.md)) — a single
screen, not a new feature.
**Status:** New, 2026-07-09. Site is live at torro.cat. Prompt 21 flagged this page as an open
question and deliberately left it unresolved: *"would the comparison page (Agramunt vs Xixona,
inherently two-sided) ... specifically benefit from a 2-column desktop layout? Lower priority
either way."* This prompt resolves that question into a full commissioning brief.

## Why this prompt exists

`/torro-agramunt-vs-xixona` (`public/templates/comparativa.html`) is currently a single column of
long-form prose — five `<h2>` sections describing Agramunt and Xixona side by side in prose. The
exact shape varies per section (mostly paired Agramunt-paragraph-then-Xixona-paragraph, but
"Origen geogràfic" is one merged paragraph covering both, and "Ingredients i percentatges mínims"
is an intro paragraph followed by a 3-item list) — see the exact content inventory below, don't
assume a uniform two-paragraph structure to redesign from. The content is *already* structured as a
comparison in substance, it just isn't laid out as one. At desktop width (≥1280px) that linear back-and-forth
reads slower than it needs to — a reader scanning "how is the texture different" has to hold one
side in their head while reading the other. A genuine two-column layout (Agramunt | Xixona) lets
both sides be scanned side by side per topic, which is a better fit for what this content actually
is: a documented, neutral, point-by-point comparison, not an essay.

This page is currently shared with three other prose pages (`/sobre`, `/torro-agramunt-igp`,
`/tipus-de-torrons`) via one CSS container, `#content-page-container`. **This prompt only covers
the comparativa page** — the other three stay single-column prose, which prompt 21 already judged
correct for them.

## Brand grounding (repeated here per this doc set's own convention — self-contained, no other file
## needed to hand this to Claude Design)

Torrorèndum is a Catalan-language head-to-head voting competition where people compare pairs of
torrons (Catalan Christmas nougat) from the heritage brand Torrons Vicens (founded 1775, Agramunt,
PGI-protected craft). Tone is **"museum-grade craft heritage" crossed with "playful,
internet-native flavor drops"** — the same brand does both 1775-heritage marketing and viral
flavors like Dubai chocolate. Warm browns, caramel/gold, and cream — never generic Christmas
red-and-green. A single deep-burgundy "competition" accent is reserved for results, winners, VS
badges, and calls to action — **never** used as a generic UI color, which matters here since this
page is informational, not a competition moment.

**Recurring visual motif, cards/buttons:** a "sticker" treatment — `transform: rotate(-1deg)`, a
hard offset drop-shadow instead of a soft blur (`--shadow-sticker: 0 6px 0 var(--color-border)`),
`--radius-card: 20px` rounding. This reads as "cut-out sticker on a table."

### Full token system (use these values exactly — do not invent new colors/spacing/radii)

```css
--color-primary: #E0923F;
--color-primary-light: #B96F26;   /* despite the name, this is the DARKER hover/pressed value */
--color-primary-dark: #B96F26;
--color-primary-tint: rgba(224, 146, 63, 0.08);
--color-brand-gold: #EFC26E;

--color-competition: #8A2638;        /* RESERVED — do not use for this page's UI chrome */
--color-competition-dark: #6B1D2A;
--color-competition-tint: rgba(138, 38, 56, 0.08);
--color-competition-contrast: #FFFAF1;

--color-white: #FFFFFF;
--color-card: #FFFAF1;
--color-background: #FBF2E3;
--color-surface: #F0DFC2;
--color-desktop-background: #E7E0D0;  /* wide-viewport canvas, revived by prompt 21 — reuse it here too, don't reinvent */

--color-text: #3A2A1C;
--color-text-light: #8A7458;
--color-text-light-dark: #6B5A44;

--color-border: #E8D6B4;
--color-border-dashed: #D8C097;

--color-shadow: rgba(58, 42, 28, 0.12);
--color-shadow-heavy: rgba(58, 42, 28, 0.28);
--color-shadow-soft: rgba(58, 42, 28, 0.16);
--color-shadow-faint: rgba(58, 42, 28, 0.05);

--spacing-xs: 5px; --spacing-sm: 10px; --spacing-md: 15px; --spacing-lg: 25px; --spacing-xl: 40px;

--font-family: 'Newsreader', Georgia, serif;                 /* body copy */
--font-family-display: 'Bricolage Grotesque', sans-serif;    /* headings, UI labels */
--font-size-base: 16px; --font-size-sm: 0.875rem; --font-size-lg: 1.25rem;
--font-size-xl: 1.5rem; --font-size-2xl: 2rem;

--radius-card: 20px; --radius-button: 18px; --radius-pill: 999px;
--shadow-sticker: 0 6px 0 var(--color-border);
--shadow-sticker-badge: 2px 2px 0 rgba(58, 42, 28, .18);
```

**Open decision — which existing card pattern to build from:** this codebase has two competing
card visual languages already in production, and they use different tokens. Note the base
`.torron-card` rule is itself the calmer pattern by default — the sticker treatment only exists in
scoped per-screen overrides, not as `.torron-card`'s own baseline:
1. **The sticker treatment** — `--radius-card`, `--shadow-sticker`, `rotate(-1deg)`. Not on any
   base class; it's applied via scoped overrides on specific screens:
   `#bracket-vote-container .torron-card` (`main.css:4624-4636`, bracket vote screen) and
   `.advent-torron-card` (`main.css:5933-5940`, advent calendar) get the full combination.
   `.torron-card.vote-card` (`main.css:2075-2087`, the main vote screen) gets `--radius-card` only
   — no rotation, and a soft `--shadow-sm` instead of `--shadow-sticker`.
2. **The calmer, non-rotated pattern** — this is actually the *default*: both the base
   `.torron-card` rule (`main.css:1587-1601`) and `.stats-card` (`main.css:2873-2880`) share it:
   `background-color: var(--color-white)`, `border-radius: var(--border-radius-lg)` (**a different
   radius token than `--radius-card`**), `box-shadow: var(--shadow-md)` (a soft blur shadow, not
   the hard-offset sticker shadow), no rotation. `.stats-card` is used on the stats/achievements
   dashboard; the base `.torron-card` styling is what any *unscoped* use of that class would fall
   back to.

This page is documented legal/technical fact-checking (IGP status, percentages, dates), closer in
tone to the calm "spec sheet" framing prompt 21 used for the product detail page than to a
competition moment — **explicit decision needed: build the two comparison cards from the
`.stats-card` calmer pattern, from the `.torron-card` sticker pattern, or a third treatment — don't
silently default to one.**

## Exact current state

Route: `GET /torro-agramunt-vs-xixona`, template `public/templates/comparativa.html`. Shares
`#content-page-container` with three other prose pages: `max-width: 720px; margin: 0 auto`
(`main.css:6875-6879`). `.content-masthead` (`main.css:6881`) holds an eyebrow/title/subtitle
block. Body copy rules: `.content-body h2` (`main.css:6912`), `.content-body p`
(`main.css:6926`), `.content-body ul`/`ol` (`main.css:6930`). A `.content-cross-links` nav
(`main.css:6983`) with three `.content-cross-link` pill buttons (`main.css:6992`) sits at the
bottom, linking to the IGP explainer, the glossary, and the vote flow. **None of this has a
`min-width` breakpoint** — it renders identically narrow at any viewport width today.

## Existing DOM/content inventory to design around (don't invent new fields or research new facts —
## this is the real, fact-checked copy already live on the page)

**Stays full-width, both above and below the two-column section** (these are framing content, not
themselves comparable side-by-side):
- Masthead: eyebrow "Comparativa", title "Torró d'Agramunt vs Torró de Xixona", subtitle.
- Opening section, headed "Dues tradicions torroneres, dues fitxes tècniques diferents" — its
  paragraph establishes neutrality (no normativa ranks one over the other).
- Closing section, headed "Dues maneres legítimes d'entendre el torró" — its paragraph is the
  neutral conclusion.
- The three cross-links at the bottom.

**Becomes the two-column comparison** — five topic sections, each split into an Agramunt side and
a Xixona side. **Important asymmetry Claude Design needs to know about: Xixona is not one thing —
it covers *two* separate IGPs (Jijona, soft texture; Turrón de Alicante, hard texture) under one
regulatory council. The Xixona column is inherently longer/denser than the Agramunt column in every
section below.** Don't design assuming mirrored, equal-length columns — the layout needs to hold up
with a visibly heavier right side.

1. **Origen geogràfic**
   - Agramunt: municipi d'Agramunt, comarca de l'Urgell, Lleida, Catalunya. Elaboració i envasat
     han de tenir lloc dins del terme municipal.
   - Xixona: municipi de Xixona (Jijona), comarca de l'Alacantí, Alacant, Comunitat Valenciana.
     Mateixa regla d'elaboració/envasat dins del terme municipal.
2. **Estatus legal: totes dues són IGP, no DOP**
   - Agramunt: una sola IGP, "Torró d'Agramunt," inscrita el 2002 (Reglament CE 1241/2002).
   - Xixona: **dues** IGP diferents sota el mateix consell regulador, seu a Xixona — "Jijona"
     (tova) i "Turrón de Alicante" (dura) — totes dues reconegudes el 1996.
3. **Textura i mètode d'elaboració**
   - Agramunt: "irregular, tosca, amb porositats i de pasta dura, però es trenca sense esforç,"
     color lleugerament marró/daurat. Mel + sucre/xarop de glucosa coent en un perol, clara d'ou,
     ametlles o avellanes senceres torrades i pelades.
   - Xixona — Jijona: textura tova, ametlla mòlta barrejada amb xarop caramel·litzat ja endurit,
     pasta marró blanquinós amb grànuls d'ametlla irregulars. Xixona — Turrón de Alicante: textura
     dura, ametlles senceres torrades "a modo de jaspe blanco," pot anar recoberta d'oblia.
4. **Ingredients i percentatges mínims**
   - Agramunt: ametlla o avellana ≥46% (Extra) / ≥60% (Suprema); mel ≥10%; clara d'ou ≥1%.
   - Xixona — Jijona: ametlla ≥50% (Extra) / ≥64% (Suprema); mel ≥10%. Xixona — Turrón de
     Alicante: ametlla ≥46% (Extra) / ≥60% (Suprema); mel ≥10%.
5. **Un apunt històric**
   - Agramunt: documents des de principis del s. XIX ("Cal Torró"), ofici documentat el 1741,
     Denominació de Qualitat creada per la Generalitat el 1984, reconeixement europeu (IGP) el
     2002.
   - Xixona: Consell Regulador constituït el 1939, Denominació Específica regulada el 1991,
     reconeixement europeu (IGP) el 1996.

## Technical constraint specific to this page — no Xixona product photography exists

The site's entire product catalog (`Torrons` table) is Vicens products, and Vicens is an
Agramunt-heritage brand — **there are no Xixona-brand torró product photos anywhere in this
codebase's asset library.** Whatever the Agramunt side gets visually (a product photo, an IGP seal,
a map pin), the Xixona side cannot use a real photographed torró to match it, since none exists in
`/public/images`. The current page has zero imagery at all (pure typography), which sidesteps this
entirely — **explicit decision needed: stay photo-free and lean on typography/iconography for both
sides equally (e.g. an IGP seal-style badge, a small map/region indicator), or introduce imagery
only if something exists for *both* sides symmetrically (stock IGP certification marks, region
outline icons) — don't design one side with real photography the other side can't match.**

## Technical constraints (binding — inherited from prompt 21, unchanged)

- **No build step, no CSS preprocessor.** One hand-written file, `public/css/main.css` (currently
  ~7,290 lines), native CSS custom properties only.
- **Server-rendered Go (`html/template`) + htmx, not a client-side framework.** No React/Vue/
  client-side state.
- **Mobile treatment is done and must not regress.** This page's current single-column narrow
  layout stays exactly as-is below the chosen breakpoint; this prompt only adds `@media
  (min-width: ...)` rules for the two-column treatment above it. Match prompt 21's precedent: the
  vote screen and product detail page both got a `1280px` breakpoint for their desktop
  compositions — reuse that same breakpoint here unless there's a specific reason not to, for
  consistency across the site's desktop treatments.
- **This page's content is legally/factually sourced** (official IGP plecs de condicions) — the
  five topic sections and their exact copy above are not placeholder text, don't paraphrase or
  invent additional comparison points beyond what's listed.

## Prompt (copy-paste this into Claude Design)

```
Torrorèndum: Catalan voting competition for the heritage brand Torrons Vicens (founded 1775,
Agramunt, PGI-protected craft), warm brown/caramel/cream palette, a single reserved burgundy
accent for competition/results moments only (not used on this page — this is neutral, documented
content, not a competition moment). Heavy display sans (Bricolage Grotesque) for headings/UI
paired with a warm serif (Newsreader) for body text.

Redesign the desktop treatment (1280px-2560px) of one page: a neutral, documented comparison
between two Catalan/Valencian torró traditions, "Torró d'Agramunt" and "Torró de Xixona." The page
currently reads as a single centered column of prose (five topic sections, each describing
Agramunt then Xixona back to back) capped at 720px — fine on mobile, slow to scan on a wide
screen. Redesign the five comparable topic sections (origin, legal/IGP status, texture, ingredient
percentages, history) as a two-column layout — Agramunt on one side, Xixona on the other, per
topic — while keeping the opening/closing framing paragraphs and the page masthead full-width
above and below the two-column section.

Important content constraint: the two sides are NOT symmetric. Xixona covers two separate IGPs
(a soft-textured "Jijona" and a hard-textured "Turrón de Alicante") under one regulatory body, so
its column is consistently longer and denser than the Agramunt column in every section — design
for a visibly unbalanced two-column layout, not mirrored equal-height cards. There is also no
Xixona product photography available in this project's asset library (the whole product catalog
is Agramunt-heritage Vicens products) — stay typography/iconography-led for both sides rather than
assuming photography, unless proposing something symmetric both sides can use (e.g. IGP seal-style
badges, region map icons).

This project has two existing card visual languages to choose from, not both: (1) a "sticker" card
— slight rotation, hard offset drop-shadow, `--radius-card` — used for playful/competition
moments; (2) a calmer, non-rotated white card with a soft shadow — used on the stats dashboard.
Recommend which fits a "documented legal facts, museum-grade heritage" tone better for this page,
don't default silently.

Deliver: an annotated desktop mockup (1280px) for the two-column comparison section, expressed as
CSS-translatable values (colors, spacing, radii) using the existing token names given, not new
arbitrary values.
```

## Deliverable format needed back

Same as prompt 21: a static mockup or a written layout description precise enough to translate
directly into CSS — no `DesignSync`/component-library project needed for a single hand-written
stylesheet.
