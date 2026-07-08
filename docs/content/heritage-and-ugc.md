# Heritage content + UGC hashtag — draft copy

Content-only deliverable, zero build required. Two pieces: (1) heritage/brand-story copy that
could sit on a future "About Torrorèndum" or footer section, (2) a UGC hashtag concept for
fans to share their own results/cards.

**Legal framing note — read before publishing anything here.** The pending Spanish IP/trademark
consult (see `reference_vicens_contacts.md`) flagged real risk in *implying* an official
Vicens/Torrorèndum partnership. Every piece below is written to read as independent fan
appreciation — factual/historical statements about Vicens, not marketing copy attributed to
them. Don't publish before the lawyer consult if that consult is still pending; this is drafted
content, not cleared content. Also worth noting separately: `index.html` currently has
`<meta name="author" content="Torrons Vicens">` — that line asserts the opposite of this framing
and is a quick one-line fix whenever it's convenient to touch it.

---

## 1. Heritage copy (Catalan, primary language of the app)

Grounded only in cross-verified facts already in project memory (founding year/location, PGI
status, company scale) — nothing invented. Three lengths, pick per placement.

**Micro (footer / about tooltip, ~30 words):**

> Els torrons Vicens es fan a Agramunt des de 1775. El Torró d'Agramunt té Indicació Geogràfica
> Protegida (IGP) — no és un torró qualsevol, és patrimoni. El Torrorèndum és un projecte de fans,
> no oficial de Vicens.

**Short (about section, ~80 words):**

> Agramunt fa torrons des de fa segles, i la casa Vicens n'és una de les cases més conegudes,
> amb arrels que es remunten a 1775. El Torró d'Agramunt està protegit per una Indicació
> Geogràfica Protegida (IGP), que garanteix que es fa amb mètodes i ingredients tradicionals de
> la zona: ametlla, mel, ou. Avui Vicens té botigues arreu de Catalunya, Espanya i més enllà —
> però la recepta de fons és la mateixa que fa generacions. El Torrorèndum neix de l'afició, no
> de Vicens: és una manera de celebrar aquesta tradició votant, comparant i discutint quin torró
> guanya cada any.

**Long (standalone heritage page, ~200 words):**

> Agramunt, a la Ribera del Sió (Lleida), és sinònim de torró des de fa generacions. No és
> casualitat: el Torró d'Agramunt és una de les poques varietats de torró amb Indicació
> Geogràfica Protegida (IGP) a Espanya, un segell que no es dona per màrqueting sinó perquè el
> mètode de producció —ametlla, mel, ou, i la manera de coure'ls— està lligat a la zona i a un
> ofici transmès de generació en generació.
>
> Vicens és una de les cases torroneres més conegudes d'Agramunt, amb un origen que es remunta a
> 1775. Del que era un obrador familiar ha crescut fins a tenir botigues pròpies arreu de
> Catalunya, la resta d'Espanya, i presència a Mèxic i França — però la base de l'ofici segueix
> sent la mateixa: torró fet bé, amb ingredients reals.
>
> El Torrorèndum agafa aquesta tradició i hi afegeix un joc: cada any, milers de vots decideixen
> quin torró guanya la temporada. No és un projecte de Vicens ni hi té cap relació oficial — és
> un projecte de fans que vol celebrar l'ofici torroner d'Agramunt i donar una excusa més per
> parlar-ne a taula.

**Sourcing for the above (don't restate these as facts beyond what's cited):** founding year
1775 and Agramunt location — company's own public-facing history (referenced consistently
across trade press, not independently re-verified this session); "Torró d'Agramunt" IGP status —
public, government-registered geographical indication, not Vicens-specific; store count/revenue
scale — see `reference_vicens_contacts.md` (~€112M 2025 revenue, 67 company-owned stores,
cross-verified 2026-07-06). Nothing about ingredients/process beyond the IGP's public
description was invented or assumed.

---

## 2. UGC hashtag concept

**Hashtag: `#ElMeuTorró`** ("My torró") over a Vicens-branded option (e.g. something built on
"Torrorèndum" or "Vicens" directly) — deliberately generic/fan-voiced rather than looking like an
official campaign tag, for the same legal-framing reason as the copy above. Secondary/backup:
`#TorronèndumFans` if a more on-brand tag is wanted later (still says "fans", not neutral).

**Mechanic — ties into features that already exist, no new build:**
1. After voting reaches the existing 50-vote threshold, a user can generate their **Wrapped
   recap card** (`/wrapped/card.png`, already shipped) or **persona reveal card**
   (`/reveal/card`, already shipped) — both are already designed to be downloaded and shared.
2. Prompt (in-app copy, not yet wired anywhere): *"Comparteix el teu resultat amb
   `#ElMeuTorró` i etiqueta els teus amics perquè votin també."*
3. No auto-posting, no API integration with Instagram/X — the existing pattern (user downloads
   PNG, posts manually) stays exactly as-is. This keeps the "zero build" property: the hashtag is
   just copy to add near the existing download button, not new plumbing.

**Sample post captions (for the user's own seeding posts, Catalan):**
- "El meu torró preferit d'aquesta temporada, oficial 😏 #ElMeuTorró"
- "Qui guanya al teu rànquing? Vota al Torrorèndum i descobreix-ho #ElMeuTorró"
- "Cada Nadal el mateix debat a taula: quin torró és el millor. Ara es pot votar. #ElMeuTorró"

**What this explicitly does NOT do:** claim Vicens sponsorship, use Vicens' logo/branding in the
hashtag itself, or imply the hashtag is run by Vicens. If Vicens outreach (the separate,
not-yet-sent artifact) ever succeeds and becomes a real partnership, this hashtag can be
revisited/rebranded then — don't pre-empt that by making it look official now.

**Open question for the user, not decided here:** whether to seed this hashtag before or after
the lawyer consult. Given the same legal-risk finding that's already pausing Vicens outreach, the
conservative default is to hold public UGC push until that consult happens too — flagging, not
deciding.
