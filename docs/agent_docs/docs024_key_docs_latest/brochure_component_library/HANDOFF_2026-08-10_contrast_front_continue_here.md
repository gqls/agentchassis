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

---

## ADDENDUM 2 (late evening) — OPTION 2 CHOSEN AND BUILT. D1a is closed; the site repair is now unblocked but NOT done

Owner chose **option 2 — fix the mechanism, not the instance**. Committed `5c7b115c5`.

**What shipped (inert until the next roll):**
- `install_site_composition` takes **`allow_reinstall`** (step-config literal, **default false**,
  read via `GetBoolFieldLoud` so a malformed declaration falls back to the SAFE branch). The
  swap happens **inside the action's existing transaction**, so the "site briefly uncomposed"
  window that made the manual route dangerous never opens.
- The link UPDATE's race guard moved `IS NULL` → `IS NOT DISTINCT FROM $3::uuid` rather than
  being dropped — it still refuses to clobber a concurrent install, in both modes.
- `previous_collection_id` is returned: the rollback value, and the **only** record of it.
- Discovery's composition pair (`check_integrity.go`) now emits **`triaged`**, matching
  `emit_design_items` which emits the same two `item_key`s.
- Registered as **DES-082 / DES-083** in the same commit. Council: **`Council-Submitted:
  b8e341b9-4709-49ad-8b7b-f4c8894ba551`** — **verdict NOT yet read. Owed.**

**Three tests, proven by mutation run before commit** (results as observed, not expected):
`if !allowReinstall` → `if false` fails A and C; hardcoding the refusal fails B.

### The scheduler question, answered — and the answer is "do NOT enable a sweep"

The promoter exists (`TriageDetectedItemsAction`) and is **undriven**: its three callers
(`improvement-loop`, `design-audit-agent`, `site-review-agent`) have **no enabled scheduled
task between them** — `improvement-sweep` is `enabled=false`.

**But enabling one is the wrong fix.** That action promotes **every** `detected` row for a
site with **no type filter**, and there are **448** `detected` build-pipeline rows fleet-wide
(193 `page_rerender`, 79 `contrast_failure`, 23 `audit_tool`, …). Turning it on dispatches all
of them at once. The real defect was narrower and is now fixed at source: **two producers of
the same `item_key` disagreed about status.** The general promoter remains undriven — left
open deliberately, and it is a genuine open question for whoever wants the other 448.

### What is still owed on this front

1. **Read the council verdict** (`b8e341b9`) and act on REVISE/REJECTED — the code is already
   on the shared branch, so this is not optional.
2. **The site is still unrepaired.** Nothing sets `allow_reinstall` yet. After the roll, the
   repair is: queue `needs_composition` for `ai-agent-orchestration.com` with the
   `site-design-planner` step carrying `allow_reinstall: true`. **Rollback value:
   `3196d966-24ef-4415-9dc8-1afbc02166ca`.**
3. **Pod-grep after the roll** — `allow_reinstall` / `previous_collection_id`, with a negative
   control. Nothing has been proven live; DES-082/083 say BUILT, not deployed.
4. **The paired AFTER measurement**: BEFORE is 58 failures on 3 pages (38 white-card, 4 second
   light literal, 14 primary-as-ink, 1 over-image). **Prediction, pre-registered: ~42 fewer,
   ~15 left.** A drop to near zero means something else changed too.

---

## ADDENDUM 3 (late) — LIVE, but REVISE. Start a fresh session HERE

**Status in one line:** `allow_reinstall` is **in the running binary and proven so**, the
council said **REVISE**, both objections hold, and **the site is still unrepaired** — now
blocked on a small code revision rather than on an owner decision.

### Proven live (do not re-do this)

Chassis `696d88b4c7`, both replicas: `allow_reinstall` 4, `previous_collection_id` 1,
`re-resolve not requested` 1, and **`re-resolve not supported` = 0** — the string the change
*deleted*. That is a removal-based negative control, so it proves *this* build, not any build.

### THE ONE THING TO FIX FIRST — the flag cannot be used on one site

`allow_reinstall` is read from `StepConfig.Config`. `site-design-planner`'s install step
config holds only path references, so the only way to set it is to edit the **agent
definition** — which turns re-install on for **every** composition install fleet-wide, the
exact unsafe-default-ON state the flag exists to prevent.

**The revision (small, and it is the whole remaining blocker):** make the flag per-request —
read it from the work item spec / `input_data` **as well as** step config, keeping default
false and the `GetBoolFieldLoud` loud-fallback. Then a single `needs_composition` item can
carry `spec.allow_reinstall = true` and nobody else's behaviour changes.

Then resubmit with `RESUBMIT_CORR=b8e341b9-4709-49ad-8b7b-f4c8894ba551` so the trail
accumulates, and answer the second objection in the resubmission (see below).

### What I got wrong, so you do not repeat it in the resubmission

`47ce091c` is **`needs_design_review`**, created **2026-04-24** — NOT the
`needs_composition`/`needs_design` pair, NOT 2026-08-06 (that is `updated_at`), and the
`triaged` edit **does not unblock it**. The sites that edit unblocks are `noted.co.uk`,
`loanandmortgagecalculator.co.uk`, `loancalculator.co.uk` — six rows, none of them
ai-agent-orchestration.com. The `triaged` change is still right on its own merits (two
producers of one `item_key` disagreed about status); only my evidence was wrong.
`WRONG_CALLS.md` 2026-08-10 has the full account and the one-line query that would have
caught it.

**Also correct the scale claim:** the status/dispatch mismatch is not build-only. The
council's own query shows `undeployed_asset` 86 (design), `phantom_internal_link` 18
(content), `unbuilt_internal_link` 17 (content), `image_url_404` 16 (design) on top of the
448 build rows. "Do not enable a fleet-wide sweep" is unchanged and stronger.

### The repair, once the revision is live

1. Queue `needs_composition` for `ai-agent-orchestration.com`, `status='triaged'`,
   handler `site-design-planner`, with `spec.allow_reinstall = true`.
2. **Rollback value: `3196d966-24ef-4415-9dc8-1afbc02166ca`** (its current shared collection).
3. Verify at the **artefact**, never the status — `complete` already lied once on this site
   (`f7ceba19`, 2 minutes, changed nothing): check a site-specific `palettes` row exists,
   `sites.style_collection_id` changed, `styles.css` `last-modified` moved, and served
   `--color-card-bg` is no longer `#ffffff`.
4. **AFTER measurement against the pre-registered prediction** — BEFORE was 58 failures on
   3 pages (38 white-card, 4 second light literal, 14 primary-as-ink, 1 over-image);
   **expect ~42 fewer, ~15 left.** A drop to near zero means something else changed too and
   this fix should not be credited.

### Commits this session (all on `087_towards_multiple_domains`)

`4bd0fb519` 113 correction + LANDMINE · `cfb05757a` 122 pointer · `4b28bc1cf` roll + census ·
`dca8b8084` repair queued + baseline · `0cd8404b0` no-op result + platform gap ·
`5c7b115c5` **the code change** (Council-Submitted: b8e341b9) · `2c24ed5f0` addendum 2 ·
`e1b8863e0` WRONG_CALLS · `a78640045` corrections to 113 + DES-082/083.

---

## ADDENDUM 4 (2026-08-11) — round 2 written and submitted. The repair is now ONE work item away

**Verification method changed under us.** CLAUDE.md retired `strings` on 2026-08-11 ("three
confidently wrong readings in one day"). Redone the sanctioned way and it holds:
chassis stamps **`bb534864…`** (binary probe, both replicas, bogus-sha control absent), and
`git merge-base --is-ancestor 5c7b115c5 bb534864` → **round 1 shipped.**

**Round 2 (`a36cbc6cb`, committed, NOT yet rolled)** answers the council's objection.
The flag is now read from **two** sources, both default false, both loud-fallback:
step config (fleet-wide) **and the work item's `spec`** (per-request). Prefer the latter —
setting the former on `site-design-planner` turns re-install on for every install.
Two new tests; mutation nil-ing the spec lookup fails the per-request test **alone**.
Council round 2 submitted under the same trail correlation **`b8e341b9-…`** — **verdict
pending, still owed.**

**RFC_022 scope check, enumerated not asserted:** `0` active agent definitions and `0` work
items name `allow_reinstall`. Opt-in, unsafe side default, no live consumer → not
architecture-scope under the 2026-08-11 ruling.

### THE REPAIR — one work item, once `a36cbc6cb` is in a rolled build

```sql
INSERT INTO site_work_items
  (site_id, source, item_type, item_key, severity, summary, spec,
   priority, handler_agent, status, created_by, pipeline)
SELECT s.id, 'manual-113-palette-repair', 'needs_composition', 'needs_composition', 'high',
       'Re-compose ai-agent-orchestration.com off the shared LIGHT seed palette (bugs_open/113)',
       jsonb_build_object('stage','composition','domain','ai-agent-orchestration.com',
                          'reason','shared_light_collection_on_dark_site',
                          'allow_reinstall', true),      -- <<< the per-request opt-in
       7, 'site-design-planner', 'triaged', '<your-lane>', 'build'
FROM sites s WHERE s.domain='ai-agent-orchestration.com';
```

- **status MUST be `triaged`** — `detected` is never claimed (`claim_work_item_action.go:102`).
- **Rollback: `UPDATE sites SET style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca'
  WHERE domain='ai-agent-orchestration.com';`**
- **Verify at the ARTEFACT, not the status.** `complete` already lied once on this site
  (`f7ceba19`: 2 minutes, changed nothing). Check: a `palettes` row with
  `source_domain='ai-agent-orchestration.com'` exists; `sites.style_collection_id` changed;
  `styles.css` `last-modified` moved; served `--color-card-bg` is no longer `#ffffff`.
- **AFTER measurement vs the pre-registered prediction:** BEFORE = 58 failures on 3 pages
  (38 white-card, 4 second light literal, 14 primary-as-ink, 1 over-image).
  **Expect ~42 fewer, ~15 left.** Near-zero means something else changed too.

### `[UNMEASURED]` — the one thing that could make the repair a no-op

`requestSpecFromCollected` handles `input_data.spec` and `input_data.body.spec`. **I have
not observed a live `needs_composition` dispatch's `collected_data` shape.** A third shape
means the flag silently does not arrive — the action then refuses (SAFE), but it will look
like a broken flag. **Check this first if the repair refuses:** find the run and enumerate
`jsonb_object_keys(collected_data->'input_data')`.

### Commits (all on `087_towards_multiple_domains`)

`4bd0fb519` · `cfb05757a` · `4b28bc1cf` · `dca8b8084` · `0cd8404b0` · **`5c7b115c5` r1 (LIVE)** ·
`2c24ed5f0` · `e1b8863e0` WRONG_CALLS · `a78640045` corrections · `73e7bd3da` ·
**`a36cbc6cb` r2 (awaiting roll)**.

---

## ADDENDUM 5 (2026-08-12) — THE REPAIR IS DONE AND VERIFIED. Nothing on this front is owed except reading one verdict

**Status in one line:** `ai-agent-orchestration.com` is repaired at the served artefact,
113's mechanism has no known remaining live instance, and the only open item is council
round 3's verdict.

### What was owed, and where each item ended

| owed (addendum 4) | outcome |
|---|---|
| read the round-2 verdict | **REVISE**, landed 2026-08-11 18:19:16Z, 8 objections. Answered in round 3 |
| confirm `a36cbc6cb` rolled | **LIVE** — chassis `v1.0.1290`, stamp `fa078ab3d`, binary probe both replicas + bogus-sha control, `merge-base --is-ancestor` true |
| fire the repair | **DONE** 13:48–13:56Z, verified at the artefact |
| the AFTER measurement | **DONE** — 58 → 40 against a pre-registered ~15. The miss is the finding |

### The repair, exactly

```sql
-- 1. composition  (item 57b9b3ff, complete in 103s)
--    spec: {stage, domain, reason, allow_reinstall: true}, handler site-design-planner, status 'triaged'
-- 2. render       (item ca5acb4b, complete in ~3m)   <<< NOT OPTIONAL, see below
--    spec: {domain, reason, stage: 'design'},        handler webdesign-agent,      status 'triaged'
```

**Rollback, still valid:** `UPDATE sites SET style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca' WHERE domain='ai-agent-orchestration.com';`

**THE TRAP THIS FRONT PAID FOR (now a LANDMINE):** the `needs_composition` half completes,
changes `sites.style_collection_id`, writes a site-specific `palettes` row — and **queues
nothing**. Every DB check reads green while the site serves the old stylesheet for ever.
`MissingStyleCollectionCheck` emits BOTH items for exactly this reason; a hand-written repair
that copies only the first half silently drops the half users see. Verify at
`curl -sI …/styles.css | grep last-modified`, never at the item status.

### Verified at the artefact

`--color-card-bg` `#ffffff` → **`#0D1117`** (byte-equal to `--color-surface`, on a palette that
now supplies no `card_bg` — `fillDarkSchemeSpecialisedSlots`' signature, i.e. **113's own fix
finally reaching this site**). `styles.css` last-modified 2026-08-11T16:22:21Z →
2026-08-12T13:56:26Z. Collection `3196d966` (shared seed) → `a0f1ac70`; palette `origin=seed`
→ `adopted` with `source_domain` set.

### The prediction missed, and that is why it was worth recording

**58 → 40**, predicted ~15. Two runs 25 min apart, identical. Attribution:

- **18** of the 38 white-card failures went — this repair, working as designed.
- **20** did not, and never could: `.team-member { background: #fff }`, a **hard-coded literal
  in a component template**. Exact, not indicative — `about.html` 12 elements / 12 failures,
  `index.html` 8 / 8. **This is a different defect** (D4's family / `features_open/026`).
- **14** primary-as-ink on the dark grounds — `bugs_open/122`, exactly as predicted, untouched.
- **4** on `rgb(248,249,250)` — **`[UNATTRIBUTED]`**. The cascade says `styles.css` (linked at
  line 142) should beat the page's inline `:root` (line 68) and it demonstrably does not. Ask
  `getComputedStyle` which declaration won; do not reason about it from the source order, which
  is what I did and it gave the wrong answer.
- **1** on the accent `rgb(240,165,0)` — new since the repair, unexamined.

**Had it landed on ~15 I would have credited this fix with 20 failures it cannot reach** and
the `.team-member` family would still be invisible. See `WRONG_CALLS.md` 2026-08-12: the BEFORE
table committed **trap #4 of this front's own six** — reading a route off a value.

### For a fresh session

1. **The only thing owed: read council `b8e341b9` round 3** and act on REVISE/REJECTED — the
   code is on the shared branch already.
   `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE kind='council_report' AND correlation_id='b8e341b9-4709-49ad-8b7b-f4c8894ba551' ORDER BY created_at;`
   Round 3's own answers to the 8 objections are in
   `COUNCIL_SUBMISSION_113_reinstall_r3_2026-08-12.json` — read it before re-arguing any of them.
2. **Do NOT re-open 113 for what remains on that site.** Two other families produce
   similar-looking damage there; the per-selector attribution above is the thing to carry
   forward, not the totals.
3. **The `.team-member` literal is unowned.** It is the biggest single remaining block on this
   site (20 of 40) and it is a component-template change, not a palette one. Same shape as D4,
   which is also still unowned.
4. **D2 (does the specialised-slot authority defect get its own bug file) is now moot for this
   site** — it was repaired by re-composition rather than by changing the shared seed. The
   *class* question (nothing refuses a light theme palette on a dark site spec — D3) is
   untouched and still open.

### ADDENDUM 5a (2026-08-12, later) — round 3 came back APPROVED. The front is closed except for one pod-verify

**`b8e341b9` round 3: APPROVED**, 2026-08-12 14:17:57Z, ~20 minutes after submission. 12
reviewers, `gated_by_truncation: false`, all 12 carrying a verdict, `decided_by` = *"approved
with 2 advisory objection(s) — none high-severity"*. **The trail is REVISE → REVISE →
APPROVED, and two of the three rounds found real defects** — which is the argument for
submitting rather than defending.

**The editquality advisory was right and is fixed (`9d4fbb4f7`).** My own round-3 code
returned a bare nil for BOTH "nothing at that path" and "something at that path that is not
an object", so the caller reported *"no `input_data.spec` in collected_data"* for a spec that
was demonstrably present — **the diagnostic the entire round existed to add would have lied
about the rarer case**, sending the next reader to the dispatch layer when the fault was
whoever queued the item. It now returns the reason and names the type with `%T`. Third test
sub-case added (the seat noted that gap too); mutation collapsing the two reasons fails that
sub-case alone.

**Two advisories deliberately NOT acted on, and they are the useful leftovers:**

1. **`bug_historian`** — `extractContentWithFallbacks` and page-content-writer's
   `extractSiteID` are the **sibling call sites on this same fault line** and still carry
   heuristic two-branch readers. It named the recurring shape exactly: *"one call site of a
   shared judgement gets the rigorous fix; the sibling stays heuristic."* Out of scope for a
   bug patch, genuinely unowned, and a good small piece of work for someone.
2. **`debug_historian`** — round 3 is committed, **not rolled**, so it is not pod-verified.
   **Owed after the next fleet roll**, with the same discipline rounds 1 and 2 got: obtain
   the stamp, `git merge-base --is-ancestor 9d4fbb4f7 <stamp>`, and grep the running binary
   for `resolved_from` with a bogus-sha control. Note the production behaviour already
   proven does NOT depend on this — the repair ran on round 2's code.

**Register updated** — DES-082 goes from *"live but not safely usable"* to **live and used**,
DES-083 from *"not yet observed emitting"* to **observed emitting and dispatching** (two
discovery-produced pairs completed 2026-08-11). The `improvement_guardian`'s high-severity
round-2 objection to DES-083 is recorded there as unsettled rather than closed by those runs:
they show the change works, not that the gate it removed was unwanted.

---

## ADDENDUM 5 (2026-08-12) — four owner instructions. Two done, one answered, one routed

### 1. "Make sure the repairs are done through the framework's own mechanism" — **DONE, verified at the artefact**

`ai-agent-orchestration.com` now has **its own palette** (`palette-ai-agent-orchestration-com`,
`origin=adopted`, `source_domain` stamped), it supplies **no `card_bg`**, and the served sheet
reads `--color-card-bg: #0D1117` **byte-equal to `--color-surface`** — the
`fillDarkSchemeSpecialisedSlots` signature. **The white cards are gone.** It went through
`needs_composition` → `site-design-planner` → `install_site_composition` with the per-request
flag; no hand-written SQL touched the site. Council trail `b8e341b9` reached **APPROVED** at
round 3 (`Council-Reviewed:` on `9d4fbb4f7`).

### 2. "Approval needed, but for now default that the human approves" — **BUILT** (`1fa86f5cc`, not yet rolled)

Every composition **replace** now records `reinstall_approved_by`. Order: step config →
work item spec `reinstall_approved_by` → spec `approved_by` (the column a real HITL flow
already fills, so wiring one later needs no code change here) → the sentinel
`default-grant/owner-2026-08-12`. Nothing blocks today — that is the "default that the human
approves" half.

**The sentinel is deliberately not a person-looking string**, because that is what makes the
eventual tightening measurable rather than a leap:
```sql
SELECT result->>'reinstall_approved_by', count(*)
  FROM site_work_items WHERE result ? 'reinstall_approved_by' GROUP BY 1;
```
**Do not reword the constant** — it is a stored value, so a tidy-up splits that population in
two. Registered **DES-084** (+ index row). Three tests, mutation-proven.

### 3. "Are these missing handlers?" — **NO. Measured, and the premise was wrong**

Every one of the 11 item types still sitting at `detected` has a **live handler agent**
(`agent_definitions` join, 11/11). And the handlers demonstrably work: `page_rerender` is
**2811 complete**, `undeployed_asset` 187 complete, and **84 items completed in the last six
hours**. The backlog I reported as 448 is now **21** fleet-wide.

What actually happened, and it is not a defect:
- The **driver** was the gap, not the handler. `TriageDetectedItemsAction` promotes
  `detected → triaged`; its three callers have one scheduled task between them
  (`improvement-sweep`), which is **`enabled=false`** on cost grounds and was run
  deliberately on 2026-08-11 — that run is what drained the queue.
- **226 `contrast_failure` items are `deferred` ON PURPOSE** (migration `389`, owner
  decision 2026-08-11), because promoting them through `css-patch-agent` would produce false
  closures. **0 have ever completed; `attempt_count = 0`.**

**So there is nothing to fix here in the "missing handler" sense.** The queue is cost-gated
by owner decision, and the one genuine blockage is (4) below. **The open question is a cost
question, not an engineering one:** whether `improvement-sweep` runs on a schedule or stays
a deliberate manual lever.

### 4. "Fix the issue broadly" — **ROUTED, not fixed, and deliberately so**

The broad contrast repair is gated on `bugs_open/213`, which `who-owns.py` reports **OWNED
and active**. I traced the blocker precisely and contributed it there (`fbc77f081`) rather
than starting a competing fix:

**`css-patch-agent` has no verification step at all.** Its workflow is
`ensure_site_record → load_current_css → check_has_css → plan_css_fix → save_css_to_db →
check_saved → deploy_css → complete|complete_no_css|complete_error` — three
`complete_workflow` terminals and **no `complete_work_item`/verification call**. So it is a
*third* case, distinct from the two producers in 213's title: not "the verifier implements
one predicate and not the other", but a handler that never enters the verification path, so
`out_of_scope` cannot fire for it either.

213's gate **is** working elsewhere — its `_verification` population grew 26 → 44 and is
recording `defect_persists` 9 and `error` 9. It simply never sees this agent.

**Nothing was unparked.** Releasing 226 items before `css-patch-agent` can be verified
reproduces exactly what migration `389` exists to prevent.

`[UNMEASURED]`, and it decides the size of the fix: what `check_saved` branches on. If a
no-op patch already lands on `complete_no_css` the repair is reporting only; if it lands on
`complete`, it is the false-closure defect itself.

---

## ADDENDUM 6 (2026-08-12, late) — COLD-START POINT FOR A FRESH CHAT. Sweep run, ownership answered, "nothing to patch" answered, and a correction

**Read this addendum and stop.** Addendums 1–5 are history; this is the state.

### The sweep — RUN, and turned back OFF

`improvement-sweep` was enabled at 16:15, fired **16:16:22**, and I disabled it again at
once. **It is `enabled=false` now — verify that before assuming, it is a live cost.**

What it did: **21 items promoted `detected` → `triaged`**, one improvement-loop
orchestration dispatched (`c9cb0d76`). Discovery also filed *new* work, so `detected` went
22 → 25 while `triaged` went to 24. **That is the sweep working, not failing** — the loop
discovers and triages in the same pass.

**Its gate is `LIMIT 1` — one site per 900s tick**, ordered by `sites.updated_at ASC`. So
draining six sites is ~90 minutes of it being enabled. Budget that before turning it on.
**The 226 parked `contrast_failure` rows are `deferred`, and triage promotes `detected`
only — the park is safe from the sweep.** Confirmed, not assumed.

### Which thread owns 213, and is it active — ANSWERED

**`bugfix_122_contrast_ink_slots`.** `who-owns.py`: OWNED, **ACTIVE, 27 commits/14d**, 50
mentions, handoff `HANDOFF_2026-08-12_continue_here.md` **dated today**. It committed on 213
today. **It is ahead of us and should not be competed with.**

Its `b2fca2f8f` (2026-08-12) is the thing to read: it costed the contrast_failure verifier
fork and concluded (a) the standing objection is at `verifier_coverage_test.go:199-201`,
three instances, and kills options 1–3; (b) `contrast_failure`'s exemption is **on the
record** at `verifier_coverage_test.go:156`, justified by an argument **RFC_017 refuted on
2026-08-08**, six days after it was written; (c) the answer is **discovery-path retraction**
via `resolveWorkItems` (`work_items_common.go:249`), blocked only on the audit reporting
*which* pages it covered, not how many.

### The "nothing to patch" case — ANSWERED, and it is already honest

`css-patch-agent`'s `save_css_to_db` is a guarded UPDATE (`length($2) BETWEEN 1 AND 8192`);
`check_saved` branches on `css_saved.count >= 1`, else `complete_error` — the config's own
description is *"Refuse to deploy unless the guarded append took a row (bugs_open/198)"*.
**A no-op patch cannot reach `deploy_css`.** My earlier `[UNMEASURED]` flag on this is
withdrawn.

**And it is moot in practice: `css-patch-agent` has NEVER processed a work item.** Its whole
footprint is the 226 `deferred` rows — 0 complete, 0 failed, `attempt_count = 0`.

### CORRECTION — my Addendum 5 item 4 was wrong on the mechanism

I wrote *"css-patch-agent has no verification step at all"*. **False.** Completion lives in
`build-dispatch-loop`'s `process_item` → `sub_workflow.steps.mark_complete` =
`complete_work_item`. I enumerated with `jsonb_each(default_config->'workflow'->'steps')`,
which is **top-level only** — the exact query LANDMINES.md warns about, on the table it
footprints. Corrected in `bugs_open/213` (`bd1ca99e4`) and logged (`71cdc11f3`).

**What is true:** `contrast_failure` has **no registered verifier** (twelve types are
registered; it is not one), so it reaches the verifying completion and finds no predicate —
which is 213's original title case, not a new one.

### Where this front now stands

- **113's own mechanism: repaired fleet-wide**, ai-agent-orchestration.com included, verified
  at served stylesheets. Council trail `b8e341b9` **APPROVED** at round 3.
- **`allow_reinstall`** live (r1/r2) + **approval recording** built (`1fa86f5cc`, DES-084,
  **not yet rolled** — pod-verify it after the next roll).
- **Broad contrast work belongs to `bugfix_122_contrast_ink_slots`, not here.** Do not
  unpark the 226; do not write a contrast verifier — that lane has costed it and chosen a
  different shape.

### If you pick this up cold, do these three things first

1. `grep -n "<table|symbol>" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` before any
   census — the SessionStart hook only matches **file** footprints, and the one that bit me
   twice is footprinted on a **table**.
2. `./scripts/who-owns.py <bug>` **before** filing a finding, not after. Mine duplicated a
   worse version of work the owning lane had done that morning.
3. Confirm `improvement-sweep` is still `enabled=false`.
