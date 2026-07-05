# Prompt 03 — Category selection screen

**Set:** A (existing site, redesigned)
**Status:** Live at `GET /classes` (`public/templates/classes.html`).

**Naming decision needed before finalizing this screen.** The dossier flagged this category's name as unconfirmed; a follow-up Catalog Audit resolved it against the live vicens.com catalog. The app's DB currently labels this category exactly `"Albert Adrià"` (see `migrations/000005_insert_classes.up.sql`), with description `"Essència Adrià"` — a name the audit found does **not** exist on Vicens' live site. The two real names in use there are **"Adrià Natura"** (Albert Adrià's legacy solo line, running since 2013) and **"Sinergia. Los chefs en casa."** (the newer multi-chef umbrella — Jordi Roca, Quique Dacosta, Ángel León, and Adrià — his recent work now sits inside). The prompt below uses "Adrià Natura" as a working placeholder; confirm with the product owner whether to rename this category, split it into two, or leave it as-is before this screen's copy is locked.

## Prompt (copy-paste this into Claude Design)

```
Torrorèndum: Catalan voting competition for Torrons Vicens, warm brown/gold/cream palette, burgundy accent. Design the category-choice screen shown right after landing: five cards — Clàssics, Novetats, Xocolata, the Adrià/chef-collaboration line (working name: "Adrià Natura" — see note above), and Global (cross-category). Each card needs a one-line description, a distinct but harmonious icon or texture per category (not identical templated cards), and a sense of "which arena do you want to enter" rather than a plain settings menu. Make the Global option feel like the premium/hardest mode — it's the cross-category showdown. Mobile-first, five options need to work as a single-column stack without feeling like a long form.
```
