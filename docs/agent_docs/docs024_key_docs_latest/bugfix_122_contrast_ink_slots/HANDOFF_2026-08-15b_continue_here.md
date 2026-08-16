# HANDOFF — bug 122 ink work + spin-offs. START HERE. Written 2026-08-15 (evening).

**Supersedes `HANDOFF_2026-08-15_ink_5_0_continue_here.md`** for current state. That file is still
the record of the 5.0 canary gate (now closed) and its trap list; read it second, not first.

> ## ✅ NOTHING IS BLOCKED. NOTHING IS IN FLIGHT. No work item is queued, claimed or held.
>
> **UPDATED 2026-08-15 ~18:5xZ — the four owner decisions of §4/§5 are ALL ACTIONED and verified
> at the artefact. The estate is consistent. Only two genuinely-open items remain (§4a "why", and
> the `publish_site` parity failure in §5).**
>
> | decision | outcome |
> |---|---|
> | 1. Delete the duplicate plan line | **DONE.** Served homepage now shows the section **once** (heading count 1, was 2). Before-images of BOTH deleted rows banked in `migration_backups` under `bugs_open_278_delete_duplicate_info_card_grid` |
> | 2. Bring robot-hands + cookly to 5.0 | **DONE.** robot-hands `#94a0c2`/`#f77f47`; cookly accent `#a24122` — each the pre-computed 5.0 value, read at the served stylesheet |
> | 3. The unenforced bare-reference invariant | **LEFT, by owner decision.** Still true, still unenforced — §5 |
> | 4. Close the RFC_022 audit blind spot | **DONE — and it was far wider than the motivating case: 103 live actions were uncountable, 46 of them shared.** See §8 |
>
> **Fleet now on `v1.0.1303`** (pods restarted 18:45Z, stamp `5e075a6f9`, junk-sha control absent).
> Carries `e0f239118` **and** `263197ca2` — verified by ancestry with a working negative control.

---

## 1. What is DONE and LIVE (all verified at the artefact, not at a status)

| thing | state |
|---|---|
| **Ink derivation repair** | LIVE. `--color-primary-ink` is a lightness-shifted BRAND colour, not `--color-text` |
| **Target = 5.0** (owner ruling) | LIVE in `v1.0.1301`, stamp `0115f2b45` |
| **Per-site kill-switch** | LIVE, default ON, contrast knob clamped `[4.5, 7.0]` |
| **Council `d60aab29`** | **APPROVED** at round 2 (round 1 REVISE found a real defect — the zero-value `inkPolicy`) |
| **Migration 415** — `article-body` link repoint | **APPLIED + PROPAGATED to all 97 placements across 20 sites** |
| Canaries dartsonline + webdesign.co.uk | PASSED, re-graded from post-rebuild grounds |

**Verify any of this yourself in one command each** (never trust the table above after a roll):

```bash
# is the approved code running?  (ancestry, NEVER grep the binary for your sha)
git merge-base --is-ancestor e0f239118 <stamp from the pod>     # must be true

# does any page still SERVE the old link rule as the winning declaration?
# (the position() comparison is the load-bearing part — see §6 trap 3)
```
```sql
SELECT count(*) FROM page_components pc
WHERE pc.rendered_html LIKE '%article-body__content a{color:var(--color-primary,#1e40af)%'
  AND position('--color-primary-ink' in pc.rendered_html)
    < position('a{color:var(--color-primary,#1e40af)' in pc.rendered_html);
-- was 0 at 2026-08-15 16:0xZ
```

**Live colours, `[MEASURED 2026-08-15]`, each reproduced by an independent implementation:**
dartsonline `--color-primary-ink #94a0c2` (5.122) / `--color-accent-ink #f18072` (5.125) ·
webdesign.co.uk accent `#915e2c` (5.151), **primary UNCHANGED `#5c6b5d` and that is CORRECT** ·
robot-hands `#94a0c2` / `#f77f47` and cookly accent `#a24122` — **brought to 5.0 on 2026-08-15
by owner decision**; they had been left on the 4.5 values by a pre-roll third-party re-render.
**The whole estate is now on the 5.0 target.**

## 2. OWNER RULINGS — settled, do not re-litigate

1. **Brand-coloured links: YES.** (2026-08-15)
2. **Target 5.0**, not the bare 4.5 AA floor. *"As a default we only need to get to AA unless someone
   specifically says otherwise in the brief"* — and he specified otherwise for this.
3. **Build the kill-switch** (reversing this lane's earlier decline).
4. **Wait for the council-approved binary** before the visual gate — I recommended proceeding on the
   behaviourally-identical earlier build and was overruled. Honour this shape in future: he wants the
   reviewed artefact, not merely the equivalent one.
5. **Widen the `article-body` repoint to all 20 sites** — *"it only needs to go to 5.00 - the
   framework default."* The fix pins no number; it points at `--color-primary-ink`.

## 3. The mechanism, in one paragraph (so nobody re-derives it)

`buildLegibleInkDefaults` (`palette_specialised_slots.go`, step 12 of `RenderCSSFromSpecAction`)
emits `--color-primary-ink` / `--color-accent-ink`: the palette colour moved in **HSL lightness
only**, hue and saturation preserved, smallest change that clears `inkMinContrast` (**5.0**) against
**four** grounds — background, surface, and each composited under the renderer's own 5% white
`--section-surface` overlay. Where the source already clears, it is returned **unchanged** (the no-op
branch — this is why webdesign's primary does not move). Consumers opt in with the two-level form
`var(--color-X-ink, var(--color-X))`, never bare. `inkFloorContrast` (4.5) is a **separate** constant
read by `pickInkOn` for the `--color-<x>-text` slots (labels ON filled controls) — **do not re-merge
them**, or a ruling about links silently retunes every button in the fleet.

## 4. OPEN — two spin-offs, both from the owner looking at webdesign.co.uk

### (a) `bugs_open/278` — duplicated section on webdesign.co.uk's homepage. LOCATED, not diagnosed.

`site_plan_sections` carries `info-card-grid` **twice** for the `index` page, `ordering` 1 and 2,
both stamped `2026-07-25 16:26:18.203946+00` — one transaction. **So the page is composed exactly as
planned; the save/render path is exonerated.** It is a true duplicate: same four card titles in the
same order.

**UNKNOWN and deliberately unasserted: why the plan contains it.** I have not opened the planner.
**`090` is required before anyone asserts a structural cause**, because that would predict
recurrence and the census says **N=1 fleet-wide**.

⚠ **The likely fix is a plan edit + re-render, not code — but do not apply it until the "why" is
known**, or the plan regenerates it. §8 of that file banks the evidence a fix would destroy.

### (b) ~~The AI-sounding copy~~ — **RESOLVED 2026-08-15 ~19:15Z, and the ordering constraint HELD**

`[MEASURED 2026-08-15 19:2xZ]` The webdesign.co.uk homepage now carries **three** components
(hero · **one** `info-card-grid` · call-to-action) and the h2 reads **"Tools and guides, sorted by
the problem you're solving"**. The *"A workbench, not a sales pitch"* construction is **gone from the
served page** — 0 occurrences, was 2.

**The sequence worked exactly as the two lanes agreed**, and that is the part worth keeping: the
composition fix landed first (~18:2xZ, duplicate plan line deleted), the voice rewrite ran after
(~19:15Z, all three components re-rendered in one pass). Had the rewrite gone first it would have
either collapsed the duplicate — destroying `278`'s diagnostic state before it was banked — or
rewritten only position 2 and left position 3 serving the old copy. **The ordering was not luck; it
was negotiated between the lanes and honoured by both.**

Original description, kept because the *why* still matters for the next site:

Their h2 is literally *"A workbench, not a sales pitch"*. That lane confirmed the "X, not Y" shape is
already codified as an AI tell (`voicetells.go`) and banned by the v2 house voice; webdesign.co.uk
has **no voice spec and no `voice_gate`**, so nothing caught it. Their remedy is a voice-only
`content_rewrite` through the framework.

⚠ **ORDERING CONSTRAINT, agreed with that lane:** composition before voice. A page-level rewrite
could collapse or re-duplicate the section and destroy 278's diagnostic state; a section-scoped
rewrite of position 2 leaves position 3 serving old copy. **Ping that lane when 278 closes** — their
session ended; `copy_quality_two_stage/HANDOFF_2026-08-15` is the pickup.

## 5. Optional tidy — none of it urgent

- ~~robot-hands + cookly still serve 4.5 inks.~~ **DONE 2026-08-15** — both rebuilt and verified at
  the served stylesheet. Keep the general fact: **only a `webdesign-agent` run regenerates a
  stylesheet, and nothing schedules one**, so any future ink change stays dormant per site until
  something re-renders it.
- **`bugs_open/122` §11** — a `var(--x, fallback)` whose `--x` is defined but of the **wrong type**
  (a gradient in a `color:` slot): the fallback is dead code while the source reads as if it has a
  safety net. Evidence confirmed. Needs its own bug number.
- **The no-bare-reference invariant is unenforced.** The kill-switch's emit-nothing path is safe only
  because **46 ink references across four surfaces are all two-level** (measured, with a control).
  A future migration adding a bare one silently arms a declaration-dropper. A periodic check is the
  real fix and is not built.
- ~~RFC_022's optional-key audit cannot see `render_css_from_spec`.~~ **CLOSED 2026-08-15 — and the
  hole was far wider than that one action. Full finding in §8.** What remains open there is the
  **pre-existing `publish_site` cron-parity failure**, which belongs to another lane.
- **The `webdesign.uk` lane was never told** its one page was re-rendered in the widening; that
  session had closed. Not a homepage, rows unlocked, only a `color:` inside `.article-body__content a`.

## 6. TRAPS — the expensive ones, all paid for already

1. **`page_rerender` with a free-text `reason` is a NO-OP for a template change.**
   `check_rerender_mode` matches `input_data.spec.reason` **exactly** against
   `{image_landed, section_data_resolved, cta_links_stale}`; anything else takes `else_step:
   render_page` = *"assemble stored HTML"*. Four items reached `complete` with
   `page_components.updated_at` unchanged.
2. **`rerender_page_sections` REQUIRES `spec.page_name`, and its absence still reports COMPLETED.**
   `Required: [target_site_id, page_name]`. Without it the step fails instantly while the workflow
   completes. **Already documented in
   `brochure_component_library/scripts/rerender_page_sections_direct.sh`'s header** (measured
   2026-07-26, incl. two failures from another session). I found it after failing twice. **Grep the
   scripts dir before hand-rolling a dispatch envelope.**
3. **A substring census cannot answer a CASCADE question.** A repoint ADDS the new rule beside the
   old, so "does the old text appear?" always over-reports. Two pages still contain the raw string
   and render correctly because the new rule comes last. Use the `position()` query in §1.
4. **A binary carries only its OWN stamp.** Grepping `/proc/1/exe` for your sha returns *absent* on a
   correct binary. Use `git merge-base --is-ancestor`, and always run a junk-sha negative control.
5. **`make release` does NOT bump `IMAGE_TAG`.** A same-tag rebuild ships the node's cached binary —
   reports success, changes nothing. Use `make release IMAGE_TAG=v1.0.xxx`.
6. **`--apply` on the real migrations dir sweeps in every other lane's pending files** (~8 were
   pending). Scope it: `MIGRATIONS_DIR=<dir with only your file> ./scripts/migration/run-migrations.sh --apply`.
7. **`deferred` is the ONLY parking state for a work item.** A `blocked` one un-parks itself within
   600s via the live `feasibility-recheck` task — *unless* `handler_agent` is empty, when it jams for
   ever. Full entry in `LANDMINES.md`.
8. **Set `handler_agent` on the COLUMN, not only in `spec`.** The router reads the column.
9. **Page render ≠ stylesheet render.** A page deployed after a roll still serves the stylesheet's
   older inks.

## 7. Key identifiers

`d60aab29-3590-474e-898c-cd5224c9a8ee` council (APPROVED r2) · commits `12cf55015` `8ad05d01a`
`d4bbbf645` `e0f239118` · migration `415_article_body_link_ink_repoint.sql` (+ `_ROLLBACK`) ·
stamp `0115f2b45` = `v1.0.1301` · `bugs_open/122`, `bugs_open/278`

**Rollback routes for the ink work, cheapest first:** per-site kill-switch
`legible_ink_disabled_site_ids: ["<uuid>"]` on `webdesign-agent`'s `render_css_from_spec` step config
(config flip, no rebuild) → global `legible_ink_enabled: false` → `415_..._ROLLBACK.sql` → code revert.


## 8. RFC_022's optional-key counter had a blind spot covering half the fleet — CLOSED 2026-08-15

**Found because the audit said nothing about `render_css_from_spec` when three optional keys were
added to it.** Its silence was the gate not looking, and it was not cited as a pass.

**The cause, and why nobody had found it: the code's own comment asserted coverage that did not
exist.** `censusOptionalKeys` iterates `ListActionInputSpecNames()`, so an action that never
registered an `ActionInputSpec` is skipped — it cannot be over budget because it cannot be counted.
The comment said such actions are "omitted too (`--unregistered-actions` owns that class)". **It does
not.** That mode reports actions **absent from `GlobalActionRegistry`** — steps rejected on every
message — which is the opposite population. A registered, working action with no input spec was
reported by **neither** audit.

`[MEASURED 2026-08-15]`

| | |
|---|---|
| actions the budget **can** count (spec with optional keys) | **119** |
| actions it **cannot** (no spec at all) | **103**, of which **46 are shared** (≥2 carriers) |

Uncountable and heavily shared: `execute_llm_prompt` (64 carriers), `query_database` (41),
`spawn_agent` (41), `call_agent` (41), `conditional` (40).

**The fix is REPORT-ONLY by design** (`censusUncountedActions`, commit `263197ca2`): a new
"NOT COUNTED — the optional surface is UNKNOWABLE, not zero" section. Registering 103 specs is work
most of them will never need, and paging the estate over a documentation gap teaches readers to
ignore the check. What changed is that a reader can now distinguish **"zero"** from **"unknown"**,
which the old output could not express. The registry import is **named** (main.go carries the same
package blank) precisely to tell "does not exist" from "exists, declares nothing".

> ⚠ **OPEN, and NOT this lane's:** `optional_budget_cron_parity_test` **fails at HEAD** — confirmed
> by reverting my change and re-running, so it is pre-existing. `check.py`'s
> `OPTIONAL_KEY_COUNTS["publish_site"] = 0` while the registry declares **3**. Another lane added
> optional keys to `publish_site` without regenerating the cron literal, so **the DEPLOYED check is
> stale for that action**. The test names the regeneration command. Reported rather than fixed
> because it is that lane's surface; it wants doing.
