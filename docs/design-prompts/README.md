# Design prompts

21 copy-paste prompts for redesigning Torrorèndum: 19 extracted from the [Growth & Redesign
Dossier](https://claude.ai/code/artifact/06054553-5090-4ad7-8df6-7c5d94ffa8d7) (§04), grounded
against what's actually shipped in this repo as of 2026-07-05, plus two prompts added after the
fact (§ below). Each screen prompt is self-contained — brand grounding (palette, type
pairing, tone) is repeated inside every one — so any single file here can be pasted into Claude
Design or handed to a designer on its own, with no other context required.

**Always commission [Prompt 01](01-design-system-foundations.md) first** — every other prompt
assumes its palette, type scale, and component styles are already locked — **and commission
[Prompt 20](20-design-system-audit.md) last**, once real mockups/exports exist for the other 19,
to catch any drift or inconsistency before calling the redesign done.

## Set A — the existing site, redesigned

| # | Screen | Status |
|---|---|---|
| [01](01-design-system-foundations.md) | Design system foundations | Foundational — do this first |
| [02](02-homepage-landing.md) | Homepage / landing | Live — `GET /` |
| [03](03-category-selection.md) | Category selection screen | Live — `GET /classes` ⚠️ naming decision needed, see file |
| [04](04-the-duel-vote-screen.md) | The duel (vote screen) | Live — `GET /classes/{id}/vote` |
| [05](05-leaderboard.md) | Leaderboard (ranked list) | Live — `GET /leaderboard` |
| [06](06-stats-achievements-dashboard.md) | Stats & achievements dashboard | Live — `GET /stats` |
| [07](07-voting-history.md) | Voting history | Live — `GET /history` |
| [08](08-product-detail-page.md) | Product detail page | Live — `GET /torro/{id}` |
| [09](09-navigation-countdown-onboarding.md) | Navigation, countdown & onboarding | Nav + countdown live; onboarding modal is new |

## Set B — new screens for features built since the dossier

| # | Screen | Status |
|---|---|---|
| [10](10-bracket-knockout-view.md) | Bracket / knockout view | Live — `GET /bracket/{classId}` — flagship, highest-value redesign target |
| [11](11-advent-calendar.md) | Advent calendar | Live — `GET /advent` |
| [12](12-friend-leaderboard-challenge-link.md) | Friend leaderboard & challenge link | Live — `GET /friends` |
| [13](13-streak-indicator.md) | Streak indicator | Live — small component on `/stats` and the vote flow |
| [14](14-torro-personality-reveal.md) | "Torró personality" reveal | **Not built** — needs real copywriting first |
| [15](15-shareable-result-card.md) | Shareable result card | Live (plain v1) — `GET /share/card.png` — reused by 16 & 18, prioritize this one |
| [16](16-wrapped-recap.md) | Wrapped recap | Live (plumbing only) — `GET /wrapped` |
| [17](17-embeddable-leaderboard-widget.md) | Embeddable leaderboard widget | Live — `GET /embed/leaderboard` |
| [18](18-press-ready-numbers-page.md) | Press-ready "the numbers" page | Live — `GET /premsa` |
| [19](19-insights-dashboard-vicens.md) | Insights dashboard (Vicens-facing) | **Not built** — event-triggered, no auth system exists yet |

## Set C — final audit and cross-cutting follow-ups

| # | Screen | Status |
|---|---|---|
| [20](20-design-system-audit.md) | Design system consistency & UX audit | Not built — a QA/critique pass over 01–19's output, commission last |
| [21](21-desktop-wide-viewport-treatment.md) | Desktop / wide-viewport treatment | Cross-cutting addendum — 01 was explicitly scoped mobile-first, site is now live and desktop was never designed. Priority: [04](04-the-duel-vote-screen.md) (vote screen, critical) then [08](08-product-detail-page.md) (product detail, major). Shipped (vote + product pages only). |
| [22](22-comparativa-two-column-desktop.md) | Comparativa (Agramunt vs Xixona) two-column desktop layout | Addendum to 21 — resolves 21's own open question about this page. Not designed/built yet. |
| [23](23-bracket-desktop-populated-view.md) | Bracket / knockout view: populated desktop treatment | Addendum to [10](10-bracket-knockout-view.md) — audits the never-before-screenshotted populated bracket at desktop width (real seeded data, not guesses). Not designed/built yet. |

## Notes for whoever commissions these

- "Live" means there's a real screen at that route today — screenshot it alongside the
  prompt for extra grounding, especially for Set B, since those screens were deliberately
  shipped with plain, unstyled placeholder CSS/PNG rendering ("functional now, restyle
  later") specifically so this redesign pass has something concrete to replace.
- Prompt 15 (shareable result card) is the single highest-leverage prompt in Set B: both
  Wrapped (16) and the press kit (18) reuse its rendering engine, so whatever visual
  language it gets will carry through to both automatically.
- Prompts 14 and 19 aren't reskins of something live — they're genuinely new screens with
  no existing implementation to reference.
- For full brand research (Vicens' real palette/photography/positioning), the SEO plan,
  and the Vicens partnership pitch, see the full dossier linked above — the prompts here
  only carry the minimum grounding needed to stand alone.
- Prompt 20 is not part of the original dossier — it's a final consistency/QA pass added after
  the fact, meant to catch style drift or incoherence across the finished set of screens before
  the redesign is considered done. It needs real mockups/exports of 01–19 in hand to be useful,
  not just the written prompts, and its output is a findings-and-fixes list, not new visuals.
