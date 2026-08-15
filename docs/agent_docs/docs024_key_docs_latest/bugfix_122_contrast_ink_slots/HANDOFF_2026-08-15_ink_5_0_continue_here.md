# HANDOFF — bug 122, the 5.0 ink target + kill-switch. START HERE. Written 2026-08-15.

> ## ✅ GATE SATISFIED 2026-08-15 ~10:38Z — v1.0.1301 carries `e0f239118`. §4 steps 1–2 EXECUTED.
>
> The owner ran the release. Verified before acting, all three ways: pods restarted
> 2026-08-15T10:14Z on image `v1.0.1301`; stamp `0115f2b4528b0063fd01e7af275ccefe9c5a991d` read
> from the chassis pod's own full log AND probed into `/proc/1/exe` (junk-sha negative control
> absent) AND matching on `render-audit-adapter`; `git merge-base --is-ancestor e0f239118
> 0115f2b45` → **true** (retraction `5639a1103` also still carried). Dispatch waited out the
> 300s post-restart drop window (acted at ~10:40Z, 26 min after pod start).
>
> **Done at ~10:41Z, §4 order, before-states re-read fresh first** (dartsonline
> `#F0F2F7`/`#F0F2F7`, webdesign `#5c6b5d`/`#2b2b2b` — both match the banked shas' values):
> 1. Hold released: `829a8f3e` → `triaged` (guarded UPDATE returned 1 row).
> 2. Second canary filed: `4cdceebc-72e6-4ee6-9a3c-2c1f9fe066f0`,
>    `css_rerender_ink_round2_webdesign_co_uk_20260814`, `triaged`, `handler_agent` on the
>    COLUMN, collision guard 0 open webdesign-agent items at filing time.
>
> **§4 step 3 GRADED 2026-08-15 ~11:0xZ — BOTH CANARIES PASS, at the artefact, against the
> pre-written table:**
> | site | slot | expected | served | full-file diff vs banked before |
> |---|---|---|---|---|
> | dartsonline | primary-ink | `#94a0c2` | **`#94a0c2`** | exactly 2 lines changed (the 2 inks) |
> | dartsonline | accent-ink | `#f18072` | **`#f18072`** | — |
> | webdesign.co.uk | accent-ink | `#915e2c` | **`#915e2c`** | exactly 1 line changed (accent-ink) |
> | webdesign.co.uk | primary-ink | unchanged `#5c6b5d` | **`#5c6b5d`** | the correct no-op branch |
>
> Both items ran first attempt (`claimed` by `build-dispatch-loop` ~7 and ~25 min after filing,
> `complete`, `attempt_count` 0). No page-level ink override on either homepage (the layer that
> wins was checked). The `#94a0c2` branch of the three-way confirms the 5.0 binary ran — not
> `#8a97bd` (4.5) and not `#F0F2F7` (nothing shipped).
>
> **Independent re-grade, 2026-08-15 10:42–10:46Z (second thread, arrived at the same verdict by a
> different route — recorded because the ROUTE differs, not the answer).** The table above compares
> served values against a table written BEFORE the rebuild. That is sound, but it cannot tell a
> correct value from a value that is right for the wrong palette: a rebuild can move
> `--color-background`/`--color-surface`, and a prediction against stale grounds agrees with itself.
> So this pass **re-transcribed the grounds from the POST-rebuild stylesheet** and recomputed:
>
> | site | slot | served | recomputed from post-rebuild grounds | worst-of-4 |
> |---|---|---|---|---|
> | dartsonline | primary-ink | `#94a0c2` | `#94a0c2` ✓ | **5.122** |
> | dartsonline | accent-ink | `#f18072` | `#f18072` ✓ | **5.125** |
> | webdesign | accent-ink | `#915e2c` | `#915e2c` ✓ | **5.151** |
> | webdesign | primary-ink | `#5c6b5d` | `#5c6b5d` ✓ *(no-op branch)* | 5.324 |
>
> Grounds were **unchanged** by both rebuilds, so the pre-written table was valid — now established
> rather than assumed. All four clear 5.0 with margin; the thinnest is 5.122.
>
> **The control that makes this non-vacuous:** at 4.5 the same inputs give `#8a97bd`/`#ef7060` on
> dartsonline — different hexes. The check could have come out otherwise.
>
> **webdesign's primary cannot discriminate anything** — `#5c6b5d` is both the pre-repair value and
> the 5.0 value, because it already clears. What proves the rebuild ran is **accent moving
> `#2b2b2b` → `#915e2c` in the same file.** Never grade webdesign on primary.
>
> **Baseline banked for the owed audit check: dartsonline had 17 open `contrast_failure` rows at
> 2026-08-15 10:44Z** (`status NOT IN (complete, verified, rejected, wont_fix, cancelled, failed)`).
> Compare after its next audit; a RISE is the regression signal. Round 1's emission would have
> passed the colour table above and failed exactly here — which is why this half is not optional.
>
> **Owed-check status, measured 2026-08-15 13:40Z: NOT YET GRADABLE — no audit has run since the
> rebuild, and neither site is due before ~08-18.** Today's zero is "no audit yet", not "audited
> clean": zero `contrast_failure` rows created after 10:44Z on either site (counted arrivals by
> `created_at`, any status — not just open survivors), open counts unchanged (dartsonline 17 =
> baseline; **webdesign baseline banked now: 7 open**), and the only post-rebuild orchestrations
> on the two sites are the canary rebuild chains themselves plus availability-discovery
> (12:14/12:31Z). The filer of `contrast_failure` is **`render-audit-agent`, dispatched by the
> `site-render-audit-rotation` scheduled task** (hourly tick, per-site 7-day cadence stamped in
> `site_discovery_rotation` — the 08-11 stamps match each site's newest `contrast_failure` row to
> within 3 minutes). Next due: **dartsonline 2026-08-18 ~00:58Z, webdesign 2026-08-18 ~02:59Z**.
> "Due" makes a site eligible; the rotation takes ONE due site per hourly tick, so the actual run
> can lag due-time by hours when several sites come due together. **Do not grade the owed check on
> webdesign's 08-14 `visual-design-auditor`/`content-quality-auditor` runs** — different audit
> family, and pre-rebuild anyway. (robot-hands' render-audit, §7's Monday canary: due 08-17
> ~14:54Z.)
>
> **Remaining: §4 steps 4–5 — THE OWNER LOOKS.** dartsonline.com (one eyebrow label on the
> homepage) and webdesign.co.uk (every in-prose link). His "Go" gates widening, one site at a
> time, reading the served hex each time. Then §4 step 6 (the 168-component sweep) stays
> not-started until after that. Also owed: confirm no NEW `contrast_failure` rows on the two
> canary sites after their next audits.

**This is the ink-DERIVATION thread's continuation**, following
`HANDOFF_2026-08-14_ink_derivation_continue_here.md`. The filing/coordination lane
(`bugsearch 8`) closed on token load 2026-08-15; its own handoff is the cold start for the
work-item side. This file is the code side plus the current gate state.

> ## ⚠ RULED 2026-08-15: **WAIT for a build carrying `e0f239118`.** The owner chose to wait.
>
> I had recommended proceeding on `v1.0.1300` (the emitted colours are identical — §3). **The
> owner ruled the other way: the visual gate runs on the binary the council approved.** Recorded
> here because it reverses this file's own first draft, and because the reasoning in §3 is still
> correct and must not be used to re-argue a settled decision.
>
> ### What that needs, exactly — and the likely reason a "fresh build" appeared to change nothing
>
> **`make release` does NOT bump `IMAGE_TAG`.** It builds at whatever the makefile says, currently
> `v1.0.1300`, and its own usage line spells out the form: `make release IMAGE_TAG=v1.0.xxx`.
> A same-tag rebuild ships **the node's cached binary** — it reports success and changes nothing,
> which is exactly the shape of the 2026-08-15 "a fresh chassis build has been deployed" claim
> (pods never restarted; no image newer than `v1.0.1300` exists anywhere).
>
> **So the release is:**
> ```bash
> make release IMAGE_TAG=v1.0.1301        # owner runs this; releases are whole-fleet
> ```
> Builds from **committed HEAD**, and `e0f239118` is in HEAD's history `[VERIFIED 2026-08-15, with
> both a positive and a discriminating negative control]`, so a HEAD build carries it.
>
> **Then, before anything else, gate on the stamp — never on "a release ran":**
> ```bash
> git merge-base --is-ancestor e0f239118 <the new stamp>   # must be true
> ```
> Read the stamp **per service** (`bugs_open/249`), and remember a binary carries only its OWN
> stamp — grepping `/proc/1/exe` for `e0f239118` returns *absent* on a correct binary. Always run
> a junk-sha negative control.
>
> **The hold stays until that returns true.** Verified still intact 2026-08-15:
> `deferred | handler_agent=webdesign-agent | claimed_by NULL`.

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
