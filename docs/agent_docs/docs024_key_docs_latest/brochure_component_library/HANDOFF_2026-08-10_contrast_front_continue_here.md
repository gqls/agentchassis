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

---

## ADDENDUM (evening) — the repair was ATTEMPTED and is BLOCKED on a platform gap. D1 is superseded by D1a

Owner said "go ahead with ai-agent-orchestration — use the easiest palette, dark or light".
**Dark chosen** (pin, `style_direction`, `colour_mood` and the site's own `avoid` list all
say dark). **The repair did not land.** Full evidence in `bugs_open/113`, last two sections.

**What happened:** queued `needs_design` `f7ceba19` → `webdesign-agent`. It reported
`complete` in 2 minutes and **changed nothing** — no palette row, collection unchanged,
`styles.css` last-modified hours earlier, `card_bg` still `#ffffff`.

**What it did prove:** its `result.color_scheme` is the pinned DARK palette byte-for-byte.
The long-standing `[UNMEASURED]` worry — that a re-render would pull in the stale light
`design_intent.color_scheme` — is now **settled twice over** (by reading
`extractPaletteSignal`, and by observing the run). Do not re-raise it.

**The gap, and it is the real finding:**
- `should_fork_theme` contributes a **library** theme; it explicitly does **not** touch
  `sites.style_collection_id` (`fork_theme_from_site_action.go:3-13`).
- That file names `site-design-planner` / `install_site_composition` as the only installer.
- `install_site_composition_action.go:148-158` **loud-fails on a site that already has a
  collection**, recommending *"clear sites.style_collection_id manually"*.
- Every `needs_composition` row ever written carries `reason: no_style_collection`.

**So: the platform can compose a site that has nothing, and cannot re-compose one that has
the wrong thing.** That is why this site's own `47ce091c` has sat `unresolved after 2
attempts` since 2026-08-06 — the detector is correct and no handler can satisfy it.

### D1a — the decision that replaces D1 (owner)

**Option 1 — do the operator action (2 statements, ~5 min).**
```sql
-- rollback value: 3196d966-24ef-4415-9dc8-1afbc02166ca
UPDATE sites SET style_collection_id = NULL WHERE domain='ai-agent-orchestration.com';
-- then INSERT a needs_composition item, status 'triaged', handler site-design-planner
```
*Risk, stated plainly:* between the clear and the re-resolve the site has no composition,
and anything that renders in that window hits the loader's **emergency fallback**
(`render_css_composition_loader.go:144-158`) and could deploy a `standard-brochure`
stylesheet over a live site. One in-flight item (`e97fb5c5`) could do exactly that. The
window is short and the rollback is one statement, but it is a real window.
**This needs explicit owner approval — the permission layer already declined it once.**

**Option 2 — fix the mechanism instead.** Give `install_site_composition` a supported
re-resolve path (an explicit `allow_reinstall` flag, unsafe default OFF, per the owner's
2026-08-02 RFC_010 ruling on opt-in fields). Costs a council round; closes the class, makes
`47ce091c` satisfiable, and removes the manual window for every future case. **Recommended
if this will ever happen again — and `47ce091c` proves it already has.**

**Option 3 — leave the site.** Its 38-of-58 white-card failures stay.

### Paired baseline is already recorded, so whoever acts can finish the measurement

`BEFORE` (2026-08-10, 3 pages): **58 failures** — 38 on `rgb(255,255,255)`, 4 on
`rgb(248,249,250)`, 14 on the dark grounds (primary-as-ink, NOT this fix), 1 over-image.
**Prediction recorded before the fact: ~42 fewer, ~15 remaining.** A drop to near zero
means something else changed and this fix should not be credited.

### Also found this session, unfiled

**`needs_design` / `needs_composition` items are stranded at `detected`.**
`claim_work_item_action.go:102` claims only `triaged`/`approved` and nothing promotes them;
I had to promote mine by hand. Three were stuck, one (`loancalculator.co.uk`) for ~33h.
Matches the standing "detection works; schedule and dispatch do not" pattern. Not filed.
