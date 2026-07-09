# Prompt 21 — Desktop / wide-viewport treatment

**Set:** A (existing site, redesigned) — cross-cutting addendum, not a single screen.
**Status:** New, 2026-07-09. Site is live at torro.cat. Two mechanical, low-risk fixes already
shipped (topbar nav sizing, four secondary containers widened to 1200px) — this prompt covers
everything that needs an actual design decision, not just a wider number.

## Why this prompt exists

[Prompt 01](01-design-system-foundations.md)'s own commissioning instruction said, verbatim:
**"Design for mobile-first."** That was the right call at the time — there was no live site, no
users, and every other prompt in this set (02 through 20) inherited that same mobile-first scope.
Eight months later the site is live, real users are opening it on real desktop screens, and it
shows: `public/css/main.css` (6,990 lines) has dozens of `max-width` rules tuned for phone/tablet
widths and **almost no `min-width` breakpoint anywhere in the file** widening anything back out
for a 1280px+ viewport. This isn't a bug in the existing design system — it's a gap the original
scope never covered. This prompt is that missing pass.

A 4-agent screenshot audit (1920×1080 and 1366×800, ~16 routes) found the pattern is systemic:
every content container hard-caps somewhere between 480px and 900px and centers itself, leaving
anywhere from ~500px to ~700px of empty background on each side at 1920px width. Full findings
below, worst two screens first.

## Brand grounding (repeated here per this doc's own convention — self-contained, no other file
## needed to hand this to Claude Design)

Torrorèndum is a Catalan-language head-to-head voting competition where people compare pairs of
torrons (Catalan Christmas nougat) from the heritage brand Torrons Vicens (founded 1775, Agramunt,
PGI-protected craft). Tone is **"museum-grade craft heritage" crossed with "playful,
internet-native flavor drops"** — the same brand does both 1775-heritage marketing and viral
flavors like Dubai chocolate. Warm browns, caramel/gold, and cream — never generic
Christmas red-and-green, no snowflakes or holly. A single deep-burgundy "competition" accent is
reserved for results, winners, VS badges, and calls to action — never used as a generic UI color.
Every key screen needs to look intentional as a screenshot; it gets shared on Instagram/WhatsApp
constantly, and that expectation doesn't change just because the viewport got wider.

**Recurring visual motif to preserve, not reinvent:** cards and primary buttons use a "sticker"
treatment — `transform: rotate(-1deg)`, a hard offset drop-shadow instead of a soft blur
(`--shadow-sticker: 0 6px 0 var(--color-border)`, `--shadow-sticker-btn: 0 4px 0
var(--color-primary-dark)`), and `--radius-card: 20px` / `--radius-button: 18px` rounding. This
reads as "cut-out sticker on a table," not a generic SaaS card — any new desktop-only components
(a second column, a wider stat panel, etc.) should feel like they belong to this same physical
object language, not a different, flatter design system bolted onto the side.

### Full token system (use these values exactly — do not invent new colors/spacing/radii)

```css
/* Action / brand accent (caramel) — generic buttons, links, active states */
--color-primary: #E0923F;
--color-primary-light: #B96F26;   /* despite the name, this is the DARKER hover/pressed value */
--color-primary-dark: #B96F26;
--color-primary-tint: rgba(224, 146, 63, 0.08);
--color-brand-gold: #EFC26E;      /* lighter decorative accent */

/* Competition-only accent (burgundy) — RESERVED for VS badges, "GUANYADOR" banners,
   the share-result CTA, and live/active-round pills. Never a generic action color. */
--color-competition: #8A2638;
--color-competition-dark: #6B1D2A;
--color-competition-tint: rgba(138, 38, 56, 0.08);
--color-competition-contrast: #FFFAF1;

--color-white: #FFFFFF;
--color-card: #FFFAF1;                /* card surface, near-white */
--color-background: #FBF2E3;          /* page/content background, cream */
--color-surface: #F0DFC2;             /* secondary surface / hover background */
--color-desktop-background: #E7E0D0;  /* DEFINED BUT UNUSED — see call-out below */

--color-text: #3A2A1C;                /* primary text, cocoa */
--color-text-light: #8A7458;
--color-text-light-dark: #6B5A44;     /* AA-contrast secondary text */

--color-border: #E8D6B4;
--color-border-dashed: #D8C097;       /* dashed border — reserved sponsorship space */

--color-shadow: rgba(58, 42, 28, 0.12);
--color-shadow-heavy: rgba(58, 42, 28, 0.28);
--color-shadow-soft: rgba(58, 42, 28, 0.16);
--color-shadow-faint: rgba(58, 42, 28, 0.05);

/* Ranking/medal accents — podium gold/silver/bronze, unrelated to the brand palette */
--color-gold: #FFD700; --color-silver: #C0C0C0; --color-bronze: #CD7F32;

/* Spacing scale */
--spacing-xs: 5px; --spacing-sm: 10px; --spacing-md: 15px; --spacing-lg: 25px; --spacing-xl: 40px;

/* Typography */
--font-family: 'Newsreader', Georgia, serif;                 /* body copy, italic taglines */
--font-family-display: 'Bricolage Grotesque', sans-serif;    /* headings, UI labels, buttons, nav */
--font-size-base: 16px; --font-size-sm: 0.875rem; --font-size-lg: 1.25rem;
--font-size-xl: 1.5rem; --font-size-2xl: 2rem;

/* Radius / shadow (the "sticker" system) */
--radius-card: 20px; --radius-button: 18px; --radius-pill: 999px;
--shadow-sticker: 0 6px 0 var(--color-border);
--shadow-sticker-btn: 0 4px 0 var(--color-primary-dark);
--shadow-sticker-badge: 2px 2px 0 rgba(58, 42, 28, .18);
```

**Call-out — an unused token already exists for this exact problem:** `--color-desktop-background:
#E7E0D0` ("wide-viewport page canvas") is defined in `main.css` but never referenced anywhere.
Grepped to confirm. It looks like a distinct, slightly-darker canvas color for wide viewports
(behind a centered content column, `--color-background` cream for the column itself) was planned
during the original foundations pass and never finished. **Explicit decision needed: revive this
token as the desktop `<body>` background so wide viewports get a deliberate two-tone canvas
instead of empty cream, or replace it with something else — don't leave it defined-and-unused.**

## Priority 1 — The duel / vote screen (critical, fix this one first)

Live at `GET /classes/{id}/vote` (`public/templates/vote.html` + `pairing.html`). Original brief:
[Prompt 04](04-the-duel-vote-screen.md) — *"two torrons shown side by side, tap one to vote... a
large typographic VS between the cards, satisfying tap/hover states, a progress indicator toward
unlocking results, a quiet sense of momentum (streak count or vote count nearby)."* That brief is
still exactly right for what the screen IS — it just never got a desktop composition.

**Exact current state:** `.torron-comparison.vote-duel` is `max-width: 480px; margin: 15px auto`
(`main.css:2067-2072`), with only one responsive override at `@media (max-width: 768px)`
(`main.css:2180`) — there is no rule above 480px at all. Confirmed via a real screenshot at
1920×1080: **the two product cards + VS badge sit in a small ~480px box near the top-center of the
screen, with roughly 400-450px of completely empty tan background below the footer, and ~700px of
empty background on each side of the cards horizontally.** This is the core voting interaction —
the single most important conversion screen on the site — and it currently looks unfinished on any
real desktop monitor.

**Existing DOM/content inventory to design around** (don't invent new fields, these are real and
already wired to live data):
- `#progress` header: round label + a streak-count pill (flame icon, day count, an "at risk" red
  state) + a progress bar toward unlocking results.
- Two `.torron-card.vote-card` cards, each: product photo (`.torron-image`), a "NOU" (new) ribbon
  badge for 2025 additions, product name, up to 4 dietary icon badges (vegan/gluten-free/
  lactose-free/organic, emoji-based), and a small "more info" link out to the product detail page.
- A `.vote-vs-badge` circular "VS" mark between the two cards.

**What's needed:** an actual desktop composition, not a wider max-width number. Open questions for
Claude Design to resolve: do the cards grow larger (bigger photos, more breathing room) while
staying side-by-side, or does the extra width go to surrounding context (e.g. a persistent
mini-leaderboard or streak detail beside the duel)? Does `#main-content` get a min-height/
vertical-centering rule so the duel sits in the visual center of the viewport instead of pinned to
the top? Whatever the answer, it needs to keep working as an `htmx` partial swap (see Technical
constraints below) and read as a "real round," per the original brief's own framing, at any width
from 768px to 2560px+.

## Priority 2 — Product detail page (major, second priority)

Live at `GET /torro/{id}` (`public/templates/torro.html`). Original brief:
[Prompt 08](08-product-detail-page.md) — *"a small, well-made product spec sheet — informative and
calm — in deliberate contrast to the high-energy duel screen, while staying in the same visual
system."*

**Exact current state:** `.torro-detail-page` is `max-width: 560px; margin: 0 auto`
(`main.css:4715-4721`), and the hero photo `.torro-photo` is hard-fixed at `max-width: 220px;
height: 260px` (`main.css:4774-4783`) with **no responsive rule at any width, mobile or desktop** —
it's the same thumbnail-sized image on a phone and a 27" monitor. Confirmed via screenshot: this is
the single worst-scoring page in the whole audit, since it's also the page with the most content
(photo, title, intensity meter, allergen chips, rank/rating stat card, "Compra'l a Vicens.com" CTA)
crammed into the narrowest fixed box of any screen reviewed — roughly 680px of empty background on
each side at 1920px.

**Existing DOM/content inventory:** photo with a rank-position chip overlaid on one corner, product
title, a 1-5 intensity meter (dot/segment bar + a text descriptor like "cruixent"), allergen chips,
a rank card (current position + ELO rating), and the outbound "Compra'l a Vicens.com" CTA.

**What's needed:** the brief's own "spec sheet" framing already implies the right desktop answer —
a two-column layout (photo left, everything else right) that a wide viewport can naturally support
and a narrow one can't. Needs an actual larger photo treatment (the current 220×260px asset/crop
may need re-export at a larger size — check before assuming the source image supports it), and a
decision on whether the rank/rating stat card sits beside or below the info column at desktop
width.

## Secondary screens (lower priority — note if Claude Design has strong opinions, otherwise fine as narrow single columns)

- **[Prompt 10 — bracket/knockout view](10-bracket-knockout-view.md)**, `GET /bracket/{classId}`:
  called out in this doc's own README as *"flagship, highest-value redesign target"* among the
  newer screens, worth extra attention if there's room in scope. The empty-state card shown before
  a bracket starts (`.history-empty`, `max-width: 500px`) is centered in a lot of empty canvas —
  probably fine as a small centered message for an *empty* state, but the real bracket-in-progress
  view (once populated) hasn't been audited at desktop width yet since no bracket is live yet this
  season.
- **Prose/content pages** (`/sobre`, `/torro-agramunt-igp`, `/torro-agramunt-vs-xixona`,
  `/tipus-de-torrons`) — shared `#content-page-container`, `max-width: 720px; margin: 0 auto`
  (`main.css:6852`). Reads fine narrow since it's long-form prose meant to be read, not scanned.
  Open question: would the comparison page (Agramunt vs Xixona, inherently two-sided) or the
  glossary (a term list) specifically benefit from a 2-column desktop layout? Lower priority either
  way.
- **Category selection** (`GET /classes`, `.arena-list`/`.arena-intro`, both `max-width: 640px`,
  `main.css:1141-1170`) — a vertical list of 5-6 category cards. Fine narrow; a desktop grid is
  possible but not clearly better for a linear "pick one" decision — Claude Design's call.
- **Already fixed, for consistency reference only:** topbar nav (was `flex: 1 1 0` stretching every
  link edge-to-edge at ≥640px, now `flex: 0 0 auto` + 140px cap, centered) and
  `#leaderboard-container` / `#stats-container` / `#history-container` / `#press-container` (all
  were `max-width: 900px`, now `1200px`). These didn't get a bespoke redesign, just a wider number
  — flag if Claude Design thinks that's insufficient for any of them.

## Technical constraints (binding — whatever comes back needs to fit these)

- **No build step, no CSS preprocessor.** Everything lives in one hand-written file,
  `public/css/main.css` (currently 6,990 lines), using native CSS custom properties (the token
  system above). Deliverables should be translatable directly into plain CSS added to this file —
  Sass/Less/CSS-in-JS/Tailwind are not usable here.
- **Server-rendered Go (`html/template`) + htmx, not a client-side framework.** Pages are rendered
  server-side; `htmx` handles partial-page swaps (e.g. voting submits via `hx-post` and the response
  replaces just the vote container). No React/Vue/client-side state — any interaction design needs
  to be achievable with CSS transitions and htmx swaps, matching how Prompt 04 already scoped this
  ("no heavy client-side animation framework").
- **Mobile treatment is done and must not regress.** Every existing `max-width`/`@media
  (max-width: ...)` rule stays exactly as-is; this prompt only adds `@media (min-width: ...)` rules
  on top. Don't touch anything below ~640-768px without a specific reason tied to one of the issues
  above.
- **Existing component primitives to reuse, not reinvent:** `.btn` (primary sticker button —
  rotate(-1deg), `--shadow-sticker-btn`, `--radius-button`), `.torron-card`/`.torro-photo`-style
  card treatment (`--radius-card`, `--shadow-sticker`), `--radius-pill` for badges/chips. New
  desktop-only elements (e.g. a second content column, a wider stat panel) should be built from
  these same primitives.

## Prompt (copy-paste this into Claude Design)

```
Torrorèndum: Catalan voting competition for the heritage brand Torrons Vicens (founded 1775,
Agramunt, PGI-protected craft), warm brown/caramel/cream palette, a single reserved burgundy
accent for competition/results moments only, heavy display sans (Bricolage Grotesque) for
headings/UI paired with a warm serif (Newsreader) for body text. Tone: "museum-grade craft
heritage" crossed with "playful, internet-native flavor drops." Cards and primary buttons use a
recurring "sticker" treatment — slight rotation, hard offset drop-shadow instead of a soft blur,
generously rounded corners — like a cut-out sticker on a table, not a flat SaaS card.

The site was designed and built mobile-first and is now live with real desktop traffic that was
never designed for — every content container is capped between 480px and 900px and centered,
leaving 500-700px of empty background on each side of a 1920px screen. Design a wide-viewport
(1280px-2560px) treatment for two screens specifically, in priority order:

1. THE VOTE SCREEN: two product cards shown side-by-side with a large "VS" mark between them, tap
   one to vote. Currently locked at 480px wide near the top of the screen, mostly empty space
   below. Needs an intentional desktop composition — bigger cards with more room, or extra width
   used for supporting context (streak/progress detail) — while keeping the "quiet sense of
   momentum" (streak pill, progress bar) from the original mobile design and staying achievable
   with CSS transitions + server-side partial-page swaps, no client-side animation framework.

2. THE PRODUCT DETAIL PAGE: currently a single narrow column (560px) with a small 220x260px
   product photo, title, a 1-5 intensity meter, allergen icons, a rank/rating stat card, and a
   "buy at Vicens.com" link — reads as a cramped spec sheet. Redesign as a two-column layout at
   desktop width: larger product photography on one side, the informational spec-sheet content on
   the other, staying "informative and calm" in contrast to the high-energy vote screen.

Also decide what the wide-viewport page background should be — there's an already-designed but
unused canvas color (#E7E0D0, a slightly darker warm gray-beige than the E7E0D0 cream) intended for
exactly this: a two-tone canvas where the centered content sits on cream against a subtly darker
desktop backdrop, rather than just floating in empty space.

Deliver: annotated desktop mockups (1280px and 1920px) for both screens above, plus the resolved
page-background decision, all expressed as CSS-translatable values (colors, spacing, radii) using
the existing token names given, not new arbitrary values.
```

## Deliverable format needed back

Static mockups (image or a described layout) are fine — this does **not** need to go through
`DesignSync`/a claude.ai design-system project, since it's one hand-written stylesheet, not a
component library. What's needed to actually implement it: for each of the two priority screens,
either (a) an image export at 1280px and 1920px widths, or (b) a written layout description
precise enough to translate directly into CSS (exact proportions, what moves where, what's new vs.
resized) — plus an explicit answer on the `--color-desktop-background` question above.
