# Prompt 20 — Design system consistency & UX audit

**Set:** C (final audit — run only after Prompts 01–19 have actually been designed/applied)
**Status:** Not built — this is a review/QA pass over the other 19 prompts' output, not a new screen.

## Prompt (copy-paste this into Claude Design, alongside screenshots/exports of every redesigned screen)

```
Torrorèndum's redesign has now been produced screen-by-screen against a locked design system
(warm browns, caramel/gold, and cream, a single reserved burgundy "competition" accent, a heavy
display sans paired with a warm readable serif, "museum-grade craft heritage" crossed with
"playful, internet-native flavor drops"). Do a thorough, in-depth audit of the full set — every
redesigned screen, from the homepage and duel/vote screen through the bracket, advent calendar,
friend leaderboards, streak indicator, shareable result card, Wrapped recap, embeddable widget,
and press page — and report back with specifics, not a general sign-off. Check: adherence to the
locked palette/type scale/button-and-card styles/iconography, with zero one-off colors, fonts, or
spacing values that snuck into any single screen; visual and tonal coherence screen-to-screen,
especially for elements that repeat across multiple screens (the countdown widget, streak
indicator, share-card engine reused by Wrapped and the press kit) so they don't visually diverge
depending on which screen they appear on; complete, deliberately-designed coverage of every
non-happy-path state per screen — empty states, loading states, error states, and honest
zero-data states (several screens correctly show "not enough data yet" instead of fabricating
numbers, and the redesign must preserve and visually support that truthfully, not paper over it
with fake-looking placeholder data); accessibility — WCAG AA contrast at minimum, legible type
sizes at real mobile viewport widths, obvious focus and tap states, and colorblind-safe use of
the burgundy accent as the only "this matters" signal; and mobile-first responsive behavior
holding up at real breakpoints, since these screens are screenshotted and shared on Instagram and
WhatsApp constantly and any single broken layout undermines that. Deliver a structured findings
list — one entry per screen or shared component with a deviation, its severity, and a specific,
actionable fix — flagging anything that isn't perfectly aligned with the locked design system or
that would read as broken, inconsistent, or unfinished to a real user.
```

## Notes

- Commission this last, and only once real mockups or exported screens exist for Prompts 01–19 —
  an audit needs finished artifacts to compare against each other and against Prompt 01's locked
  foundations, not just the written briefs that generated them.
- This is a QA/critique pass, not a generative one: the expected deliverable is a findings-and-fixes
  report, not new visuals. Feed anything it flags back into the relevant screen's prompt for a
  revision pass before considering the redesign done.
- Pay particular attention to whatever this flags on [Prompt 15](15-shareable-result-card.md) —
  since [Wrapped](16-wrapped-recap.md) and the [press kit](18-press-ready-numbers-page.md) reuse
  its rendering engine directly, any drift found there propagates to both automatically, so it's
  worth fixing before the other two are considered final.
- Prompts 14 and 19 have no live implementation to screenshot (see the main README) — this audit
  can only meaningfully cover them once/if they're actually built and designed.
