# HANDOFF — bug 122, the 5.0 ink target + kill-switch. START HERE. Written 2026-08-15.

**This is the ink-DERIVATION thread's continuation**, following
`HANDOFF_2026-08-14_ink_derivation_continue_here.md`. The filing/coordination lane
(`bugsearch 8`) closed on token load 2026-08-15; its own handoff is the cold start for the
work-item side. This file is the code side plus the current gate state.

> ## ⚠ ONE DECISION IS OUTSTANDING AND IT IS THE OWNER'S. Everything below waits on it.
>
> **Proceed on `v1.0.1300`, or wait for a build carrying `e0f239118`?**
> They are **behaviourally identical for this rollout** (reasoning in §3). Recommendation:
> proceed. Do not re-derive this — read §3, then ask.

---

## 1. State, measured 2026-08-15

| thing | state |
|---|---|
| fleet binary | **`v1.0.1300`**, stamp `a2a691213`, pods up 2026-08-14T20:36Z, `generation == observedGeneration` (fully rolled) |
| newest image ANYWHERE | `v1.0.1300`, built ~12h ago. **No `v1.0.1301` exists**; makefile `IMAGE_TAG` is still `v1.0.1300` |
| `d4bbbf645` (5.0 + kill-switch, round 1) | **CARRIED** — `git merge-base --is-ancestor d4bbbf645 a2a691213` → true |
| `e0f239118` (council-approved round 2) | **NOT carried** |
| council `d60aab29-3590-474e-898c-cd5224c9a8ee` | **APPROVED** at round 2 (round 1 REVISE found a real defect) |
| held work item `829a8f3e-c3b6-4199-a3dd-9d7a973650c0` | `deferred`, `handler_agent='webdesign-agent'`, `claimed_by` NULL |

**⚠ A "fresh build was deployed" claim reached this lane on 2026-08-15 and was FALSE** — the
deployed build is still last night's `v1.0.1300`. Verified three ways: pod `startTime` unchanged,
`docker images` has no newer tag, deployment generation observed. **Always check the tag AND the
pod start time; "a build was deployed" is not evidence, and neither is a `make release` you did
not watch finish.**

### What is LIVE on real sites right now

The fix is **no longer dormant** — routine third-party traffic (`visual-design-audit`) re-rendered
two sites on the **4.5** binary before the roll:

| site | `--color-primary-ink` | `--color-accent-ink` | note |
|---|---|---|---|
| robot-hands.com | `#8a97bd` | `#f66e2f` | **4.5 values, LIVE** — the owner's best current look |
| cookly.uk | `#2C2C27` (unchanged branch) | `#af4625` | 4.5 values, LIVE |
| dartsonline.com | `#F0F2F7` | `#F0F2F7` | untouched, pre-repair |
| webdesign.co.uk | `#5c6b5d` (unchanged) | `#2b2b2b` | untouched, pre-repair |

All four `[MEASURED 2026-08-15 at the served stylesheet]`, each reproduced by an independent
implementation of the derivation **before** the served value was read.

**Since 20:36Z every NEW re-render emits 5.0.** robot-hands at 5.0 would be `#94a0c2` / `#f77f47`;
cookly accent `#a24122`; dartsonline `#94a0c2` / `#f18072`; webdesign accent `#915e2c` with primary
unchanged.

## 2. What shipped (all committed, all on the shared branch)

| commit | what |
|---|---|
| `d4bbbf645` | `inkMinContrast` 4.5→**5.0**; `inkFloorContrast` (4.5) split out and `pickInkOn` repointed at it; `inkPolicy` + `resolveInkPolicy` (opt-out kill-switch, clamped `[4.5,7.0]`); production pin in `palette_ink_policy_test.go` |
| `ec9a0ee2f` | pattern-check false-positive in a comment |
| `e0f239118` | **round 2**, from the council REVISE: `inkPolicy.resolved` — the zero value was the dangerous one |
| `99fa0a3fb`, `fe19fffbc`, `33f983452` | register VIZ-014 corrections |
| `c5b0a942e`, `c5638ee86`, `33f983452` | NOTES corrections |
| `5307fba98` | LANDMINES + WRONG_CALLS entries |
| `8b1ce6e8e` | round-2 council submission JSON |

**Seven mutation proofs**, each failing its own test, files restored byte-identical: M1 revert the
constant · M2 delete the kill-switch return · M3 delete the floor clamp · M4 repoint `pickInkOn` ·
M5 read the constant instead of `policy.minRatio` · M6 delete the `resolved` guard · M7 keep the
guard but drop its Error log.

### The config keys (live `agent_definitions` step config, no rebuild needed)

```
legible_ink_enabled            bool      global kill (default true)
legible_ink_disabled_site_ids  []string  per-site kill, site UUIDs
legible_ink_min_contrast       float64   retune, CLAMPED to [4.5, 7.0], clamp is logged
```
Read by `resolveInkPolicy` from the **`render_css_from_spec` step config on `webdesign-agent`**.
Disabled emits **nothing**, so consumers fall through their own `var()` fallback.

## 3. THE PENDING DECISION, with the reasoning done

**Why `v1.0.1300` is behaviourally identical to the approved `e0f239118` for this rollout:**
round 2 added `inkPolicy.resolved`, which defends against a *future* caller passing a bare
`inkPolicy{}`. In `d4bbbf645` there is exactly one call site and it passes a resolved policy (the
compiler enforces it — it is a signature change). So **the colours emitted are identical**.

**So the hold now protects almost nothing.** Yesterday it protected the owner's visual gate from
landing on 4.5 values he had overruled. Since 20:36Z any re-render emits 5.0, so an accidental
re-render now produces the *correct* colours. What remains is only the distinction between an
approved binary and a behaviourally identical unapproved one.

**Do not defend this hold harder than that.** Over-defended holds become folklore, and this lane
has written enough folklore-prevention this week.

## 4. NEXT, in order, once the owner rules

1. **Release the hold** (guard on the current status so you cannot race a claim):
   ```sql
   UPDATE site_work_items SET status='triaged', updated_at=now()
   WHERE id='829a8f3e-c3b6-4199-a3dd-9d7a973650c0' AND status='deferred'
   RETURNING id, status;
   ```
2. **File the second canary, webdesign.co.uk** (owner added it 08-14 because dartsonline barely
   shows the change — see §5). Same shape: `needs_design_review`, handler **`webdesign-agent`**,
   and **set `handler_agent` on the COLUMN, not only in `spec`** (see §6).
3. **Grade at the artefact, both checks — do not skip the second.**
   - Control D: dartsonline `--color-primary-ink` must read **`#94a0c2`** (`#8a97bd` = the 4.5
     binary; `#F0F2F7` = nothing shipped). Accent **`#f18072`**.
   - webdesign.co.uk: accent **`#915e2c`**; **primary comes back UNCHANGED (`#5c6b5d`) and that is
     CORRECT, not a failed rebuild** — it already clears 5.0.
   - Then confirm no NEW `contrast_failure` rows on that site after its next audit.
4. **Owner looks.** His "Go" gates the widening.
5. **Then the rest of the sites, one at a time**, reading the served hex each time.
6. **Only then the 168-component sweep** — designed and deliberately not started. `bugs_open/122`
   §7 and the approved plan hold the design. The eligibility rule still needs rebuilding on
   `fix_forced_text_colours_action.go:164-188`'s calibrated four-way `paintClass`; the naive
   "skip any block that paints its own background" refuses `system-stats .stats-eyebrow`, the one
   hand-made `--color-accent-ink` repoint, and **41 of 76 self-painted blocks (54%) are excluded
   wrongly**.

## 5. Blast radius, because it is lopsided and the canary understates it

`[MEASURED]` **9 templates opt into the ink vars: 4 `content_components` + 5 `layouts`.** The five
layouts carry `a { color: var(--color-accent-ink, var(--color-accent)) }` — **every in-prose link**.

- **dartsonline's served CSS has ZERO live ink references.** Its only consumer is
  `.info-card-grid__eyebrow`, in `page_components.rendered_html` on `/index.html`. **One small
  uppercase label.**
- **webdesign.co.uk serves the layout `a` rule** — every article link changes.

**Approving on dartsonline alone approves a great deal nobody has seen.** That is why the owner
added the second canary.

## 6. Traps this thread paid for — read before touching anything here

- **Only `webdesign-agent` regenerates a stylesheet.** `render_css_from_spec` appears in exactly
  one agent definition. `page-rerender`/`rerender-pages`/`rerender-site`/`rerender-chrome` have
  **zero** `render_css` steps; `rerender_pages_actions.go:558` writes only the `<link>` tag. **"Any
  re-render ships the change" is FALSE** and I published it before checking — a warning broad
  enough to forbid harmless work is one people stop reading.
- **`deferred` is the ONLY parking state.** `blocked` un-parks itself within 600s via the live
  `feasibility-recheck` scheduled task — *unless* `handler_agent` is empty, when it jams for ever.
  Repairing an item's handler therefore silently removes a park. Full entry in `LANDMINES.md`.
- **Enumerate promoters by QUERY, not memory.** Two lanes independently published lists of "every
  path that can un-park this", both missing `feasibility-recheck`, because it lives in a
  `scheduled_tasks` **row** and no Go grep finds it:
  `SELECT name, enabled, interval_seconds, pre_query FROM scheduled_tasks WHERE pre_query ILIKE '%site_work_items%';`
  **A table with a "can it?" column reads as a census.** Say which substrates you searched.
- **A binary carries only its OWN stamp.** Grepping `/proc/1/exe` for your commit returns *absent*
  on a correct binary. `git merge-base --is-ancestor <your-sha> <the stamp>` is the test. Always
  run a junk-sha negative control.
- **Page render ≠ stylesheet render.** A page deployed after a roll still serves the stylesheet's
  older inks. Do not read a page's deploy time as evidence of its ink values.
- **A work item's `handler_agent` must be on the COLUMN.** The router reads the column
  (`load_work_item_actions.go:676`); a spec-only handler yields an item that claims then
  immediately `blocked`s. The 08-14 filing had exactly this defect.
- **Fixture inputs must be transcribed from the artefact, and the test must name an OUTPUT.**
  A probe with invented grounds yields a figure indistinguishable from a measured one.
- **`--color-<x>-ink` used to BE `--color-text`** on every site. That is fixed; the register entry
  that still said otherwise is corrected. Do not re-derive the old belief from a stale doc.

## 7. Open, not started

- **`bugs_open/122` §11** — a `var(--x, fallback)` whose `--x` is defined but of the wrong type
  (a gradient in a `color:` slot): the fallback is dead code while the source reads as if it has a
  safety net. Evidence confirmed, `[INFERRED]` lifted. Needs its own bug number.
- **The no-bare-reference invariant is unenforced.** The kill-switch's emit-nothing path is safe
  only because **46 ink references across four surfaces are all two-level** (measured, with a
  control). Nothing stops a future migration adding a bare one, which would make the switch a
  declaration-dropper. A periodic check is the real fix and is not built.
- **RFC_022's optional-key audit cannot see `render_css_from_spec`** — no `RegisterActionInputSpec`
  is registered, so `scripts/audit-optional-key-budget.sh` does not list it. Its silence is the
  gate not looking. General hole, not specific to this action.
- **Monday's canary**: robot-hands' weekly audit. It now grades on the `#8a97bd` branch; both
  must-retract rows still clear. Ceiling is 33 (one item was cancelled by a page archival, not by
  a lane).
