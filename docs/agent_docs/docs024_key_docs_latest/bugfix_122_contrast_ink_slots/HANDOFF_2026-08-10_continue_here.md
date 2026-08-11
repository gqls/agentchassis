# HANDOFF — bug 122, contrast / ink slots. START HERE.

**Written 2026-08-10.** Supersedes `HANDOFF_2026-08-07_continue_here.md`. That file's
§1 (engine pod-proof loop), §5 (how to read a `090` verdict) and §7 (open advisory
objections) still stand and are not repeated; its §2b and §6 are **superseded by this
file** — the propagation is done and the baseline has been re-run.

## The one-paragraph state

**The fix is delivered, verified at the artefact, and measured closed.** All 12 page
re-renders completed and were checked at their served URLs; the re-audit graded per
selector delivered **all 10 predicted closures**; a same-day regression on dartsonline
was found, fixed (migration `368`), re-rendered and re-measured — **that page now returns
`contrast=0`, the first clean page in this lane**. The missing half of the approved plan
is also done: migration `369` puts the render audit on a **weekly per-site cadence**, and
it fired within 70s of apply and found **171 firm findings** on robot-hands' interior pages, filing 34 of them (see §4a — the sweep and the drain BOTH capped).
What remains is not fixable by this lane's design: ~24 failures on component-painted
grounds (`bugs_open/212` §8, an owner decision), and the repair route those 34 new findings
flow into carries `bugs_open/213`'s false-complete defect.

## If you are starting cold, do these five things first

1. **Re-read `CLAUDE.md`** from disk — co-edited, changes weekly.
2. **`git log --oneline -15` and `git status`** — ~30 concurrent writers; your
   session-start snapshot is already stale.
3. **Re-prove the engine at the pod** if you are about to assert anything about it —
   loop is in `HANDOFF_2026-08-07` §1. Six rolls have happened under this lane, none ours.
4. **Read `SUMMARY_2026-08-10_contrast_ink_slots.md`** — the state in plain prose.
5. **Read `bugs_open/213` whole before touching the repair queue.** It is now
   load-bearing at scale (see §4).

## 1. DONE — propagation delivered and verified [2026-08-10]

All 12 `page_rerender` items completed 10:53–11:23Z. **All 12 verified at the served
artefact** (fetched at `pages.url` with a browser UA — never a constructed filename):
every page carries the new ink tokens.

Three pages (aao index, finetuning index + tools) show `Last-Modified` from **08-09**,
before our items ran. That is not a false complete: other lanes re-rendered them after
migration 338 applied, so they already carried the tokens and our render was
byte-identical. vonc index was re-rendered *again* at 13:54Z by another lane — tokens
still present. **On a shared fleet, "your item did nothing" and "your change was already
there" look identical at the status; only the artefact separates them.**

## 2. DONE — re-audit run, banked, graded per selector

`AFTER_2026-08-10_render_audit.txt` (15 pages). **All 10 predicted closures delivered:**

| site | expected | result |
|---|---|---|
| gaswholesalers | 6 | all 6 gone, `contrast=0` |
| robot-hands | 2 | both gone (only the over-image approximate `.H2` remains — discount per runbook) |
| finetuning | 2 | both gone, plus a 1.00:1 white-on-white and 5 broken images closed by other lanes |
| dartsonline | 1 | gone by component removal — then **regressed and re-closed**, §3 |

**§6 of the last handoff is now settled: vonc's `.stats-eyebrow` was [PREDICTED]
unreachable and is [MEASURED] unreachable** — 1.63:1, byte-identical after two same-day
re-renders. gamesdesign's likewise at 1.44:1. Both are `bugs_open/212` §8.

> **⚠ THE TOTAL WENT UP: 109 → 112, while every targeted failure closed.** Grade per
> selector against the banked before-state or you will conclude the opposite of the
> truth. The rise is other lanes' work on shared sites: aao +3 (a new
> 'Eight departments' section, `.H2` 1.12:1), vetcomparison 0→3 (marginal greys at
> 4.10–4.14), idea.uk +7 (more instances of its standing muted-on-cream class). **None
> are ours and none are in 122's scope** — but somebody should own them.

The three advisory-pin drifted slots (idea.uk border, dartsonline background,
mortgagecalculator secondary) introduced **no new failures**.

## 3. DONE — the same-day regression, and why it is the most important thing here

dartsonline removed `image-hover-card-grid` (closing its baseline row by removal) and
**replaced it with `info-card-grid`, which reintroduced the identical defect ×6** —
`__card-link` 1.06:1 ×5, `__eyebrow` 1.14:1. Two days after 338 closed the first one.

Classified at the template, not guessed: `info-card-grid` hardcodes nothing; it inks
`color: var(--color-primary)` on the **page** grounds, so it is engine-reachable (NOT
212's class) and is literally 338 §4's `tool-list` shape. **Migration `368`** applied
14:47Z — 2 foreground uses wrapped, 4 non-targets asserted intact, 27 placements / 14
sites measured first (no-op where primary is already legible; `var()` fallback where the
slots are absent). Re-rendered 15:09Z. **Measured: `contrast=0`.**

**dartsonline `brands` was deliberately NOT enqueued** — its only page_component has NULL
`content_data` and fails the batch's own pre-check. It also places `info-card-grid`, so it
will pick 368 up whenever its owning lane next renders it. Left for that lane on purpose.

**Why this is the headline and not a footnote:** the class returns on its own, and it
returned inside the window in which this lane happened to be looking. That is the whole
argument for §4.

## 4. DONE — edit 8, the cadence, LIVE-PROVEN (migration `369`)

`site-render-audit-rotation`: hourly tick, 7-day due window (= **weekly per site**), one
site per fire, skips mid-build sites, stamped no-op when none due; own concurrency group
`render-audit` cap 1 (the audit has a dedicated pod), timeout 1800s < interval 3600s.
Clones the proven `site-discovery-rotation-*` mechanism rather than inventing one.

**Proven at the artefacts, not at `enabled` + a fresh tick:** rotation stamped
robot-hands.com 14:54:23Z → orchestration `b30943e4-440c-4f7c-8221-48ded2c6a562` step
`audit` → `COMPLETED` 14:57:29Z → **171 firm findings, 34 `contrast_failure` items filed**,
born `detected`, deduped `contrast_failure:<page-path>#<selector>`. See §4a for the other 137.

**Two consequences you must not miss:**

1. **The homepage-only baseline was hiding most of the problem.** Those findings are on
   *interior* pages — tool pages, guides, `/about`, `/selection-guide` — which the
   15-page survey never fetched. Expect that scale per site as the rotation works round.
   Among them: `info-card-grid__card-link` + `__eyebrow` on `/selection-guide.html`,
   which **migration 368 will close when that page next renders** — a live cross-check of
   both of today's changes, free, if you go and look.
2. **`bugs_open/213` is now load-bearing.** These items route to `css-patch-agent`, which
   213 shows can stamp `complete` with nothing written. The detection half is now
   excellent and the repair half is known-defective. **Grade repairs at the NEXT audit,
   never at the item status** — and treat 213 as this area's highest-value open bug.

### 4a. TWO caps bite on a real site, and only one of them admits it — `bugs_open/242`

Found while verifying the cadence's second fire, so the sample is **2 of 2 rotation runs**:

| | robot-hands (fire 1) | loancalculator (fire 2) |
|---|---|---|
| deployed pages | **31** | **27** |
| pages actually swept | 25 (`max_pages`) | 25 |
| **never rendered** | **6** | **2** |
| firm findings on what WAS swept | 171 | 2 (both in locked components — correctly not filed) |
| items filed | 34 | 0 |
| dropped by the drain's `max_items` | **111**, and it SAYS so | 0 |

**The drain's cap records its own bite** (`findings_capped: true`, `findings_dropped: 111`
in `collected_data->'findings_written'`). **The sweep's cap does not** — `truncated` is
computed at `request_render_audit_action.go:157`, logged at `:160`, returned in `Metadata`
at `:251-259`, and is **absent from `collected_data->'render_audit'`**, whose only keys are
`response`, `response_status`, `response_received_at`. So a partial sweep is
indistinguishable from a complete one in the stored artefact, and the missed pages are the
*same* ones every week — the tail never rotates into view.

**Do not read "the audit runs weekly" as "the backlog is being worked off."** robot-hands
alone carries ≥171 firm findings against a ≤60/site/week ceiling, and re-finds them each
cycle until they are actually repaired — which `bugs_open/213` says may not be happening.
Filed as **`bugs_open/242`** with the fix candidates ordered; the cheapest closes the door
by putting `pages_total` in the summary, which is only parity with the drain one step down.

## 5. What is left, honestly scoped

**Nothing in 122's own scope is blocked.** What remains is two other people's questions:

- **`bugs_open/212` §8 — component-painted grounds.** ~24 failures plus the two
  `.stats-eyebrow` rows. The renderer knows two grounds; a component that paints its own
  has no correct token to ask for. **Repointing components one at a time works** (that is
  what 338 and 368 are); whether the renderer should learn about component-painted grounds
  is an architecture question with an owner decision pending. Do not "fix" this inside a
  bug patch — that is exactly the scope veto in CLAUDE.md's platform-seams section.
- **`bugs_open/213`** — above.
- **Unowned drift found in passing** (§2): aao +3, vetcomparison +3, idea.uk +7. Not ours.

**122 itself is closable to the extent the engine allows.** The bug file has been updated
with today's evidence and what it does NOT cover; it stays in `bugs_open/` per the owner's
06-08 ruling that a finished bug stays there.

## 6. Commits from this session

All pathspec-scoped, all carrying the lane's standing `Council-Reviewed: c4d9c841`:

- `0d9e555ec` mig 368 — info-card-grid opts into the legible ink slots
- `4b924895f` mig 369 — render audit gets its cadence
- plus the docs commit carrying this file, NOTES, README, SUMMARY, the banked
  `AFTER_2026-08-10_render_audit.txt`, the register updates and the LANDMINES entry.

Register: **VIZ-013 corrected** (it said "not yet observed live" and was stale by at least
a week — it is live and filing), **VIZ-014 updated** to fully-live plus a new landmine
recording the two-grounds limit, **VIZ-015 added** for the cadence.

New LANDMINES entry: a `pages.sections` placement census returns 0 rows for a component
the page visibly renders — it is an array of **plain strings**, not objects. I hit it
sizing 368 (MISSTEP 15 in NOTES). A zero-placement answer reads as "nothing uses this",
which is the most dangerous wrong answer available when you are about to edit a shared
component.

---

## 7. UPDATE 2026-08-11 — the cadence proved itself, and surfaced the OWNER DECISION

**Engine re-proven on the fresh build, `v1.0.1284`**, both replicas, same five-symbol
table as §1, negative control 0. VIZ-014 survived the roll. Do not carry this forward
past the next one.

**The rotation swept the entire fleet overnight — 19 sites, one per hour, 14:54 on 08-10
through 09:02 on 08-11.** It now goes quiet until ~08-17 (7-day due window, stamped
no-op when nothing is due). The mechanism works exactly as designed and needs nothing.

**It filed 220 `contrast_failure` items. All 220 are in `detected`.** Top sites:
vonc 38, robot-hands 34, idea.uk 27, mortgagecalculator 22, lendzy 18, aao 17,
dartsonline 17.

### The finding, and my two wrong turns getting to it

> **MISSTEP 18 — I twice wrote down a cause that the next query refuted.**
> **(a)** "Nothing promotes detected items; the promoter is disabled." Refuted:
> **2,827 items have gone to `complete` since 08-04**, and `page_rerender` alone shows
> 2,169 progressed. Promotion demonstrably works.
> **(b)** "Then `contrast_failure` must be excluded by type — no dispatcher names it."
> Refuted twice: `page_component_status_drift` is *also* named by no dispatcher and
> progresses 98-to-20; and reading the action settles it —
> **`triage_detected_items` has NO type filter at all**
> (`triage_detect_items_action.go:162-173`: `WHERE site_id = $1 AND status = 'detected'`).
> Both wrong causes were plausible, both were built on a query I had already run, and
> **each took one further query to kill.** The check that would have saved both: before
> asserting a selection excludes your rows, READ THE SELECTION — it is one file.

**What is actually true:** triage is **site-scoped and type-blind**. Every detected item
on a site is promoted the moment `improvement-loop` runs *for that site*. So
`contrast_failure` is not blocked, it is **queued behind whatever makes improvement-loop
run** — and `improvement-sweep`, the scheduled task that drives it, has been
`enabled = false` since **2026-05-02**. Promotion today therefore happens only when some
other lane happens to trigger the loop for a site, which is why the backlog is uneven:
**776 detected items across 20+ types, oldest 2026-07-24.**

`[UNVERIFIED]` — **what triggers improvement-loop when it does run.** I did not establish
it. That is the one question to answer before acting, and it decides between the options
below.

### The decision (owner's, not this lane's)

| option | what it does | cost / risk |
|---|---|---|
| **A. Re-enable `improvement-sweep`** | the designed path; promotes everything, everywhere | **releases all 776 detected items fleet-wide at once**, incl. 193 `page_rerender` and 86 `undeployed_asset`. Off for 3 months — nobody knows why. **Find out why before flipping it** |
| **B. Targeted promoter for `contrast_failure`** | a scheduled task promoting only this type | contained and reversible, but it is a second promoter beside 286's "single owner" — a shared-mechanism change, so RFC/council territory |
| **C. Leave it** | findings accumulate as a measured, honest backlog; repaired when a lane touches the site | costs nothing, decays nothing — the items are the record. **Defensible while `bugs_open/213` is open** |

**The confound that makes C respectable rather than lazy: `bugs_open/213`.** The repair
route these items feed can stamp `complete` without writing anything. Promoting 220 items
into a verifier with a known false-complete defect converts an honest backlog into 220
false closures — and a `complete` row is much harder to find later than a `detected` one.
**Fix 213 before choosing A or B.** That ordering is the actual recommendation.

### Also still true from yesterday, unchanged

`bugs_open/242` (the sweep's silent 25-page cap) is unfixed — every one of those 19 sweeps
measured at most 25 pages and none of them said whether that was all. The 220 findings are
a floor, not a census.
