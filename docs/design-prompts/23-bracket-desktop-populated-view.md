# Prompt 23 — Bracket / knockout view: populated desktop treatment

**Set:** C (cross-cutting addendum, continuing [21](21-desktop-wide-viewport-treatment.md)) — a single
screen.
**Status:** New, 2026-07-09. Prompt 21 flagged this exact screen and deliberately left it
unaudited: *"the real bracket-in-progress view (once populated) hasn't been audited at desktop
width yet since no bracket is live yet this season."* No bracket has gone live in production since
then either. This prompt closes that gap by seeding a real bracket in an isolated local instance —
driven through the app's actual HTTP endpoints and vote-cascade logic, not hand-crafted fake
HTML — and auditing real screenshots of it at desktop width. [Prompt 10](10-bracket-knockout-view.md)
is the screen's original brief, called out in this doc set's own README as *"flagship,
highest-value redesign target"* among the newer screens.

## Why this prompt exists

The desktop-treatment work that already shipped from prompt 21 (`main.css:7016-7048`) is
**explicitly scoped to only two containers**: `#vote-page` and `#torro-detail-container`. Every
other page — including the bracket — was untouched. Three real screenshots (1280px and 1920px,
mid-tournament and finished-with-champion) confirm the bracket view shows the exact same
"mobile-first layout stretched wide" symptom prompt 21 catalogued and fixed for the vote and
product pages, not yet fixed here.

## Brand grounding (repeated here per this doc set's own convention — self-contained)

Torrorèndum is a Catalan-language head-to-head voting competition for the heritage brand Torrons
Vicens (founded 1775, Agramunt, PGI-protected craft). Tone: **"museum-grade craft heritage" crossed
with "playful, internet-native flavor drops."** Warm browns, caramel/gold, cream. A single
deep-burgundy "competition" accent is reserved for results, winners, VS badges, and calls to
action — **this screen is exactly the kind of moment that accent exists for:** it's the single
highest-stakes screen in the product (per prompt 10's own framing, "it should feel like the finals
of something real, not a wireframe with lines connecting boxes").

**Recurring visual motif:** cards/buttons use a "sticker" treatment — `transform: rotate(-1deg)`,
hard offset drop-shadow (`--shadow-sticker: 0 6px 0 var(--color-border)`), `--radius-card: 20px`.
The bracket already partly uses pieces of this, split across two elements, neither with the full
combination yet: `.bracket-match` (`main.css:4499-4510`) has the hard-offset `--shadow-sticker` but
no rotation; `.bracket-champion` (`main.css:4312-4327`) has the `-1deg` sticker rotation but its
shadow is the generic soft `--shadow-lg`, not `--shadow-sticker`. Any new desktop treatment should
extend this motif consistently, not replace it.

### Full token system (use these values exactly — do not invent new colors/spacing/radii)

```css
--color-primary: #E0923F;
--color-primary-light: #B96F26;   /* despite the name, this is the DARKER hover/pressed value */
--color-primary-dark: #B96F26;
--color-primary-tint: rgba(224, 146, 63, 0.08);
--color-brand-gold: #EFC26E;

--color-competition: #8A2638;         /* USE THIS on this screen — winner path, champion moment */
--color-competition-dark: #6B1D2A;
--color-competition-tint: rgba(138, 38, 56, 0.08);
--color-competition-contrast: #FFFAF1;

--color-white: #FFFFFF;
--color-card: #FFFAF1;
--color-background: #FBF2E3;
--color-surface: #F0DFC2;
--color-desktop-background: #E7E0D0;  /* wide-viewport canvas token, already revived for vote/product pages — reuse it here for consistency, don't reinvent */

--color-text: #3A2A1C;                /* also the .bracket-champion banner's current background color */
--color-text-light: #8A7458;
--color-text-light-dark: #6B5A44;

--color-border: #E8D6B4;              /* current connector-line color — see findings below */
--color-border-dashed: #D8C097;

--color-shadow: rgba(58, 42, 28, 0.12);
--color-shadow-heavy: rgba(58, 42, 28, 0.28);
--color-shadow-soft: rgba(58, 42, 28, 0.16);
--color-shadow-faint: rgba(58, 42, 28, 0.05);

--color-gold: #FFD700; --color-silver: #C0C0C0; --color-bronze: #CD7F32;

--spacing-xs: 5px; --spacing-sm: 10px; --spacing-md: 15px; --spacing-lg: 25px; --spacing-xl: 40px;

--font-family: 'Newsreader', Georgia, serif;
--font-family-display: 'Bricolage Grotesque', sans-serif;
--font-size-base: 16px; --font-size-sm: 0.875rem; --font-size-lg: 1.25rem;
--font-size-xl: 1.5rem; --font-size-2xl: 2rem;

--radius-card: 20px; --radius-button: 18px; --radius-pill: 999px;
--shadow-sticker: 0 6px 0 var(--color-border);
--shadow-sticker-btn: 0 4px 0 var(--color-primary-dark);
--shadow-sticker-badge: 2px 2px 0 rgba(58, 42, 28, .18);
```

## Exact current state, confirmed via real screenshots of a real seeded bracket

Method: an isolated local instance was seeded with a genuine 8-team single-elimination bracket for
class "Clàssics," advanced round-by-round through the app's real vote endpoints (not hand-set
database flags), and screenshotted at each stage. Findings:

**Dead space is the headline problem — worse than either screen prompt 21 already fixed.**
`.bracket-tree` (`main.css:4407-4413`) is `display: flex` with each `.bracket-round`
(`main.css:4415-4421`) fixed at `flex: 0 0 250px`, `--bracket-gap: 40px`, and **`#bracket-container`
itself has no `max-width` or centering rule anywhere** — confirmed by grep, the only
`#bracket-container`-scoped rules are component-level (image sizing etc.), not layout. Measured
from real screenshots:
- **Mid-tournament (round 1 decided, round 2 open), 2 visible round columns:** tree width ≈ 540px.
  At 1280px viewport, ~740px (58%) is empty background to the right. At 1920px, ~1380px (**72%**)
  is empty.
- **Finished bracket (3 rounds, champion decided), 1920px:** tree width ≈ 830px. Still ~1090px
  (**57%**) empty background to the right of the tree.

The tree is left-anchored and does not grow, recenter, or gain a second visual dimension as the
viewport widens — it just sits in a much bigger empty canvas. This is the identical pre-fix
symptom prompt 21 documented for the vote screen (small box, empty space below) and product page
(narrow column, empty space beside) before their redesigns.

**Page header reads fine, just small.** The masthead — "Bracket - {ClassName}" title, "Ronda X de
Y" status line, the `TORNEIG OBERT`/`TORNEIG TANCAT` status pill (`.bracket-status-pill`,
`main.css:4285-4306`), and the "Vota a la ronda actual" CTA — is centered in a narrow (~350-400px)
column. It doesn't look broken (the topbar-nav fix from prompt 21 already benefits every page
generically), it's just proportionally tiny relative to a 1920px canvas, since nothing here has an
intentional desktop composition yet.

**Connector lines are functionally correct but visually too quiet for the highest-stakes screen in
the product.** The lines joining a round's two source matches to the next round's slot
(`main.css:4469-4497`) are exactly `2px` solid `var(--color-border)` — `#E8D6B4`, a light warm tan
close in luminance to the cream page background. In the real screenshots this reads as a thin,
low-contrast hairline: technically traceable, but it does not project "printed tournament bracket"
clarity, and prompt 10's original brief specifically wanted the eventual winner's path picked out
in burgundy "so it reads as inevitable in hindsight" — **that burgundy winner-path treatment does
not exist yet in any form**, decided matches are only distinguished by a small green "✓ Guanyador"
text badge (`.winner-badge`) and 50% opacity on the loser (`.bracket-competitor.loser`,
`main.css:4529-4532`), not by the connector lines themselves.

**Champion banner is full-bleed but its content is small inside it — same asset-sizing gap as the
product photo prompt 21 already fixed.** `.bracket-champion` (`main.css:4312-4327`) is the one
element on this page that already spans the full container width edge-to-edge, with a dark cocoa
background (`var(--color-text)`, `#3A2A1C`) and gold border — a genuinely strong full-width
composition, unlike the rest of the page. But `.bracket-champion-image` is **hardcoded at
`120×120px`** (`main.css:4328-4336`) with no responsive override at any width. In the real
screenshot the banner reads as roughly 400px tall and the full 1920px wide, and that 120×120px
photo sits small and centered with a large amount of empty dark space around it, above, and below —
this is the exact "mobile-sized asset inside a much bigger frame" problem the product-detail photo
had (`.torro-photo`, `main.css:4797-4805`, `max-width: 220px; height: 260px`, no responsive rule at
its original size) before prompt 21 flagged and fixed it there.

## Existing DOM/content inventory to design around (don't invent new fields — these are real,
## already-wired template blocks in `public/templates/bracket.html`)

- Masthead (`.stats-header.bracket-page-header`): eyebrow with a small "VS" badge + "torrorèndum ·
  fase eliminatòria" label, `<h1>` "Bracket - {ClassName}", a status line ("Ronda X de Y" or
  "Bracket finalitzat!"), and a status pill (open/closed).
- A CTA button ("Vota a la ronda actual") shown only while the bracket has no decided champion yet.
- `.bracket-champion` banner (photo, trophy emoji, "CAMPIÓ DEL BRACKET" label, champion name) —
  shown once a champion exists, replaces the CTA.
- `.bracket-tree-hint`: a small label row above the tree with a "llisca en horitzontal →" hint,
  since the tree is horizontally scrollable on narrow viewports — **this scroll affordance should
  probably disappear entirely once the tree comfortably fits the viewport at desktop width**,
  Claude Design's call on the exact breakpoint.
- The tree itself: N round columns (`RONDA 1`, `RONDA 2`, ..., final column labeled `GRAN FINAL`
  and gets `--color-competition` on its title via `.bracket-round.is-final .section-title`), each
  containing a vertical stack of `.bracket-match` cards. Each match card shows two competitors —
  seed number, product photo, name, and (once decided) a winner badge on one and 50% opacity on
  the other — separated by a small "VS" label, or a "Passa directe a la següent ronda" bye slot
  when a competitor has no opponent.
- A current-round badge (`.bracket-current-badge`, burgundy tint pill) marks whichever round column
  is still open for voting.

## Technical constraints (binding — inherited from prompt 21, unchanged)

- **No build step, no CSS preprocessor.** One hand-written file, `public/css/main.css`, native CSS
  custom properties only.
- **Server-rendered Go (`html/template`) + htmx, not a client-side framework.** No React/Vue/
  client-side state; any interaction stays achievable with CSS transitions + htmx swaps.
- **Mobile treatment must not regress.** The existing narrow-viewport horizontal-scroll tree
  behavior (`.bracket-tree-scroll`, `@media (max-width: 480px)` gap/column-width reduction at
  `main.css:4610-4619`) stays exactly as-is; this prompt only adds `@media (min-width: ...)` rules
  on top, matching the `1280px` breakpoint prompt 21 established for the vote/product pages.
- **Bracket size and round count are dynamic, not fixed.** A bracket can be size 4, 8, 16, etc.
  (`domain.DefaultBracketSize = 8`, but classes can differ), meaning anywhere from 2 to 4+ round
  columns can be visible at once. The desktop composition needs to hold up whether the tree
  currently has enough columns to fill the canvas on its own or not — don't design only for the
  8-team/3-round case shown in this audit's screenshots.
- **This is the one screen where the reserved burgundy accent should be used expressively**, per
  prompt 10's original brief (winner's path in burgundy) — that's different from every other
  screen in this doc set, where burgundy is explicitly restricted.

## Prompt (copy-paste this into Claude Design)

```
Torrorèndum: Catalan voting competition for the heritage brand Torrons Vicens (founded 1775,
Agramunt, PGI-protected craft), warm brown/caramel/cream palette, a single reserved burgundy
accent normally kept only for competition/results moments — THIS screen is the one place that
accent should be used expressively, not sparingly. Heavy display sans (Bricolage Grotesque) for
headings/UI, warm serif (Newsreader) for body text. Cards use a recurring "sticker" treatment —
slight rotation, hard offset drop-shadow, generously rounded corners.

This is the flagship screen: a single-elimination knockout tournament bracket (torrons compete
round by round — quarts de final, semifinals, gran final — toward one champion). It already has
full mobile styling and works correctly, but has never been designed for desktop. Real screenshots
of a live populated bracket at 1920px show: the round-by-round tree (each round a fixed 250px-wide
column of match cards, connected by thin 2px lines) only fills 30-45% of the viewport width and
does not grow, recenter, or gain any additional composition as the screen widens — it just sits
left-anchored in a large empty cream canvas, identical to how the vote and product-detail screens
looked before their own desktop redesigns (see reference: those two now use a wider centered
composition against a two-tone canvas). The connector lines between rounds are functionally
correct but too low-contrast to read as a real tournament bracket at a glance. The champion
banner, once decided, is already a good full-width dark "hero" band with a trophy icon and the
winner's name — but the champion's product photo inside it is fixed at a mobile-sized 120x120px
with a lot of unused dark space around it.

Design a desktop (1280px-2560px) treatment that: (1) gives the tree itself more visual weight and
intentional use of the wider canvas — bigger match cards, more generous spacing, or a background
treatment that makes the empty space feel deliberate rather than unfinished; (2) makes the
connector lines/winner's path read clearly as a tournament bracket, using the reserved burgundy
accent to trace the path toward the eventual champion so it "reads as inevitable in hindsight";
(3) gives the champion banner a properly-sized champion photo for its new width, not a stretched
mobile asset. The composition needs to keep working whether 2 rounds are visible or 4+, since
bracket size varies by category.

Deliver: an annotated desktop mockup (1280px and 1920px) of both a mid-tournament state (some
rounds decided, one round still open) and the completed/champion state, expressed as
CSS-translatable values (colors, spacing, radii) using the existing token names given, not new
arbitrary values.
```

## Deliverable format needed back

Same as prompts 21/22: a static mockup or a written layout description precise enough to translate
directly into CSS — no `DesignSync`/component-library project needed for a single hand-written
stylesheet.
