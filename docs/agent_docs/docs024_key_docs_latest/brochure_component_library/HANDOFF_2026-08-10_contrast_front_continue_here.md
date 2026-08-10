# HANDOFF 2026-08-10 — contrast front. 113 is done in substance; the remaining work is a decision, not a fix

**Continues:** `HANDOFF_2026-08-09b_contrast_front_continue_here.md`.
**Bug files touched:** `bugs_open/113` (status changed), `bugs_open/122` (pointer only).
**Commits this session:** `4bd0fb519` (113 correction + LANDMINES), `cfb05757a` (122 pointer),
`4b28bc1cf` (113 roll + census). A matching `WRONG_CALLS.md` entry is at HEAD inside
`bc6b03ec4` — another session swept it as a same-file passenger; nothing was lost.

---

## Where this front stands in one paragraph

**113's own mechanism — a dark site's palette omitting a specialised slot, so the layout's
light literal ships — is repaired on every dark site on the fleet, and that is measured at
the served stylesheet, not inferred.** Four dark sites carry the derivation's signature
(`card_bg` byte-equal to `surface`, on a palette that defines no `card_bg`), and three of
them were never touched by hand. The old fleet figures ("11 other palettes", "12 guaranteed
white-card") were counted over all 31 `palettes` rows, most of which no live site reaches;
they are retired. **One site still serves a white card on a dark page, and it is a
different defect that this fix cannot reach by design.** That is the decision below.

## What changed in my understanding this session (the important part)

Yesterday's audit concluded `ai-agent-orchestration.com` was 113's last live instance. **It
is not, and the reasoning that got there is worth keeping**, because it is a trap anyone
auditing colours will hit:

- The site renders from a **shared seed palette** (`professional-dark`) that **defines
  `card_bg: "#ffffff"` explicitly**. It is a specialised slot, so the theme wins it; the
  site's dark spec can only reach the 8 `corePaletteKeys`; and
  `fillDarkSchemeSpecialisedSlots` correctly skips any slot the palette defines
  (`palette_specialised_slots.go:144`).
- **The served value is over-determined.** The layout literal is `#ffffff` and the palette
  value is `#ffffff`. No reading of the stylesheet can separate them — only
  `colours ? 'card_bg'` on the input can.
- The miss came from asking `palettes` by `source_domain`, getting 0 rows, and reading that
  as "no palette". **`source_domain` is stamped only on a per-site fork.** Logged as a
  LANDMINE (synced to `doc_notes`) and in `WRONG_CALLS.md`.

## The queries you will want (all verified today)

```sql
-- the ONLY correct way to ask which palette a site renders from
SELECT s.domain, t.name AS theme, p.name AS palette, p.origin, l.name AS layout,
       (p.colours ? 'card_bg') AS supplies_card_bg, p.colours->>'card_bg' AS card_bg,
       p.colours->>'background' AS palette_bg, p.colours->>'surface' AS palette_surface
FROM sites s
LEFT JOIN style_collections sc ON sc.id = s.style_collection_id
LEFT JOIN css_themes  t ON t.id = sc.css_theme_id
LEFT JOIN palettes    p ON p.id = t.palette_id AND p.is_active
LEFT JOIN layouts     l ON l.id = t.layout_id  AND l.is_active
WHERE s.domain = '<domain>';

-- who else rides that collection — an edit is never one site's
SELECT domain FROM sites WHERE style_collection_id = '<id>' ORDER BY domain;

-- the live steer for the core slots (NOT design_intent.color_scheme, which is unread)
SELECT jsonb_pretty(data->'palette'->'reference_values')
FROM site_specs WHERE site_id='<id>' AND aspect='design_intent' AND is_current;
```

**Derivation signature, on the served artefact** — this is what proves the fix live:

```bash
curl -s https://<domain>/assets/css/styles.css | grep -oE -- "--color-(card-bg|surface):[^;]*;"
# card_bg == surface, on a palette with no card_bg, can only be fillDarkSchemeSpecialisedSlots
```

## Decisions waiting on the owner

**D1 — how to repair `ai-agent-orchestration.com`.** It serves `#ffffff` cards on a
`#080B10` page; 44 of its 124 measured failures are ink on that white card. Three options,
and they are not equal:
- *(a) fork the palette for this site* with dark specialised slots. Contained, reversible,
  matches how every other adopted site works (`origin='adopted'`). Costs one more palette
  row to maintain.
- *(b) move the site to a genuinely dark collection.* Cheapest if a suitable one exists —
  none currently does, so this is really "build one", i.e. (a) with extra steps.
- *(c) leave it.* Defensible only if the site is being retired.
- **Do NOT edit the shared seed row.** `finetuning.uk` and `gaswholesalers.com` ride it and
  are light sites where `#ffffff` is correct.
- **Blocked on a prior question:** the site's `design_intent.color_scheme` is a *light*
  scheme while its `design_intent.palette.reference_values` is *dark* and pinned. Those two
  disagreeing is unexplained and `[UNMEASURED]`. Resolve it before re-rendering, or the
  re-render may pull the light scheme in.

**D2 — does the specialised-slot authority defect get its own bug file?** My
recommendation: **yes.** It is a different mechanism from 113 (a slot that *is* supplied,
with a value valid for one scheme and wrong for the other), it has a different fix, and
keeping 113 open for it makes 113's status unreadable. 113 is currently held open solely
because of it.

**D3 — the structural option, if you want the class closed rather than the instance.**
Nothing today refuses a *light* theme palette on a *dark* site spec. A guard at the merge —
"if the merged background is dark and a theme-owned slot is light, warn or refuse" — is the
same shape as the scheme guard 022 already installed, and would have caught this site
without anyone looking at it. Costs a council round; it is a platform-code change.

**D4 — the deferred fleet sweep, still open from 07-27.** Three
`tool-llm-cost-calculator` instances (`finetuning.uk`, `leopardessconsulting.co.uk`,
`ai-agent-orchestration.com`) hard-code `color: #fff` over `var(--color-primary)`.
`var(--color-primary-text, #fff)` is correct on all three. Currently harmless — it only
bites when their primary lightens.

**D5 — where the real remaining volume is, and it is not 113.** `features_open/026`
families 2/3 (primary used as ink) covers 5 sites and ~51 live components, and
`bugs_open/122`'s `.news-list-tag` was 181 of 442 failures in the three-site audit.
**Both dwarf what is left of 113.** If contrast work continues, it should continue there.

## What I deliberately did NOT do

- **Did not repaint any site.** 113's standing instruction, and every affected site is
  another lane's.
- **Did not run `render_audit.py --sitemap` for after-numbers.** Item 4 of 113's
  verification list is explicitly left open with its reason: the four repaired sites'
  totals are now dominated by the `.news-list-tag` and primary-as-ink families, so a fresh
  run measures 122 and 026 far more than it measures 113. Whoever runs it should attribute
  **per selector**, not per site total.
- **Did not run `090`.** Substitution declared in 113: served artefacts fetched first-hand,
  the four-table chain queried, three code paths read at line level, and two controls that
  could have come out otherwise (dartsonline; the two light siblings).
- **Did not touch the four domains with no composition linked** (`cookly.uk`,
  `loanandmortgagecalculator.co.uk`, `loancalculator.co.uk`, `loancash.co.uk`). All light,
  so 113 almost certainly does not apply — but *not measured is not fine*, and "why does a
  live domain have no style collection" is a real question nobody owns.

## Cold-start for a fresh session

1. Read `bugs_open/113`, **bottom two sections first** (the 08-09 third-pass correction and
   the 08-10 census). The head of the file is now strike-through-corrected; do not quote its
   original figures.
2. Read the LANDMINE `palettes.source_domain is stamped only on a per-site FORK…` before
   running any colour query.
3. If the owner has ruled on **D1**, the work is a palette fork + a re-render, and the
   before/after audit must be run **both** sides (113's own transferable lesson).
4. If not, the highest-value work on this front is **D5**, not 113.
