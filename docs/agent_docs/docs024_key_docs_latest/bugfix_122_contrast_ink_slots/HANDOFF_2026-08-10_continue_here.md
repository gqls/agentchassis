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
it fired within 70s of apply and filed **34 real findings** on robot-hands' interior pages.
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
`audit` → `COMPLETED` 14:57:29Z → **34 `contrast_failure` items filed**, born `detected`,
deduped `contrast_failure:<page-path>#<selector>`.

**Two consequences you must not miss:**

1. **The homepage-only baseline was hiding most of the problem.** Those 34 are on
   *interior* pages — tool pages, guides, `/about`, `/selection-guide` — which the
   15-page survey never fetched. Expect that scale per site as the rotation works round.
   Among them: `info-card-grid__card-link` + `__eyebrow` on `/selection-guide.html`,
   which **migration 368 will close when that page next renders** — a live cross-check of
   both of today's changes, free, if you go and look.
2. **`bugs_open/213` is now load-bearing.** These items route to `css-patch-agent`, which
   213 shows can stamp `complete` with nothing written. The detection half is now
   excellent and the repair half is known-defective. **Grade repairs at the NEXT audit,
   never at the item status** — and treat 213 as this area's highest-value open bug.

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
