# HANDOFF 2026-08-25 — the invented selector lane: `352` CLOSED, arm 2 split to `bugs_open/390`

**This supersedes `HANDOFF_2026-08-24_continue_here.md`** (kept for the trail; two of its figures are
struck through and corrected there). Read this, then `NOTES_invented_selector.md` (newest at the
bottom) and `RUNBOOK_invented_selector.md`. The bug file's own banner is
`bugs_open/352_HANDOFF_2026-08-22_contrast_findings_name_a_selector_that_matches_nothing.md`.

---

## 0. STATE — **352 IS CLOSED. Arm 2 is now `bugs_open/390`.** (owner ruling, 2026-08-25)

> **This section was rewritten hours after it was first written.** It originally said *"Can we close
> it? NO"* and named the split as an owner call this handoff would not make. The owner then made it:
> **split arm 2, close 352 against arm 1.** Both are done.

**`bugs_closed/352_HANDOFF_2026-08-22_contrast_findings_name_a_selector_that_matches_nothing.md`**
— arm 1: fixed (`ffa6e1c3d`), council-APPROVED (`acadbe8b`), live since `v1.0.1334` and still carried
on `v1.0.1337`, proven at the artefact, and it has now held a day of real traffic. Migration `587`
withdrew the 73 unexecutable legacy rows. **Nothing about arm 1 is outstanding.**

**`bugs_open/390_HANDOFF_2026-08-25_a_correct_contrast_selector_still_loses_the_cascade_so_the_repair_is_authored_and_inert.md`**
— arm 2: a *correct* selector whose appended rule is **outranked**, so the repair is authored,
deployed, `complete`, and inert. **Live, reproducible, and not designed.**

⚠ **390 is not a copy of 352's arm-2 sketch — the mechanism was verified first-hand before filing and
the verification CORRECTED it.** Three things 352 had wrong or missing:

1. **It is not "equal specificity loses on source order". It is LOWER specificity.** The offender
   `.ported-page-section .ported-page-content a` is **(0,2,1)**; the filed selector
   `.ported-page-content A` is **(0,1,1)**, and the agent's own prompt instructs it to repeat the
   filed selector verbatim. The rule loses *before* source order is consulted — and would lose on
   source order too.
2. **The offending VALUE is reachable even though the DECLARATION is not.** `--color-primary: #e8f5ee`
   is defined in the editable theme, and `#e8f5ee` is exactly the `fg` the finding recorded. **So
   352's proposed precondition — "if the declaration is not in `css_themes`, refuse and park" — would
   park a repairable finding.** 390 ranks that candidate last and restates the test.
3. **A pale-green link on a pale-green background is a PALETTE defect**, not a cascade one. Beating
   the cascade would paper over a bad token.

**What is left in THIS lane after the split: nothing that blocks anything.** One dated check (§4(2),
due 2026-08-28), one owed item (§4(3)), one thing that is not ours (§4(4)). The lane's docs stay here
as 352's working record and as 390's provenance.

## 1. The bug, in three sentences

The render audit recorded a class-less element's **tag name** in a field called `Class`, so the
orchestrator composed selectors like `H3.H3`, `P.P`, `A.A` — which select elements carrying
`class="H3"`, of which there are none. `css-patch-agent` faithfully wrote a CSS rule against that
selector, deployed it, and marked the work item `complete`; the text stayed unreadable.
**[MEASURED 2026-08-25 09:40 UTC] 111 such rows were already `complete`** — repairs recorded that
could never have applied. That 111 is the permanently-quotable damage figure; 587 never touches it.

## 2. ⚠ READ THIS BEFORE TOUCHING ARM 2 — the obvious fix is a REGRESSION

The bug file's own candidate (1) says to omit the class component so `H3.H3` becomes `h3`. **No.**
Today `p.P { … }` matches nothing and is therefore *harmless*. Corrected to `p` it matches — and one
stylesheet per site means it recolours **every paragraph on the site**. `P.P` and `A.A` were 121 of
the 181. The shipped fix composes the selector **in the page** (class → own id → nearest ancestor
with an id or class → bare tag) and **asserts it selects the very element that was measured**; a
bare tag is refused and counted. The invariant is *"prove it"*, not *"stop lying"*.

## 3. WHAT IS DONE (do not re-derive; do re-check the dates)

### 3a. Arm 1 — live on every roll since, and proven at the artefact

| | |
|---|---|
| fix commit | `ffa6e1c3d`, council **APPROVED** round 1 (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`) |
| first live | `v1.0.1334`, 2026-08-24 15:39 UTC |
| still live | `v1.0.1337`, 2026-08-25 09:27 UTC, all three services stamping `4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2` |
| ancestry | `merge-base --is-ancestor ffa6e1c3d 4c996e1b5` → YES, with `HEAD` as a control that correctly returns NO |

**The artefact-level proof** (§3b of the old handoff has it in full): two scheduled audits straddle
the first roll — 15:31:50 filed 47 rows with **3** invented and no `selector_scheme`; 17:33:16 filed
10 rows, **0** invented, `verified/v1` on every one. Then settled in the page with an independent
stdlib parser: `.ported-page-content A` counted **15** and **8** on two live pages, matching the
producer exactly, all 23 being class-less `<a>`s; two pre-roll rows' `SPAN.SPAN` / `LABEL.LABEL`
counted **0** against 22 real `<span>`s and 6 real `<label>`s. Controls throughout.

### 3b. It has held for a day of real traffic [MEASURED 2026-08-25 09:40 UTC]

| | |
|---|---|
| rows filed since the 15:39 roll | **15** |
| of those, invented `TAG.TAG` | **0** |
| carrying `selector_scheme` / `matches` | **15 / 15** |

### 3c. Migration 587 — applied, 73 withdrawn

Applied by hand **2026-08-24 19:11:22 UTC**, `UPDATE 73`. Post-checks: `open_invented = 0`,
`withdrawn = 73`, all carrying `pre_352_status`, `falsely_completed = 0`.

⚠ **Every open-population census in this lane now returns ZERO by design** — staleness by
SUBTRACTION, which reads as *"this never happened"*. **RUNBOOK §10** carries the which-side-of-587
query and the recovery query that keeps returning **73** (deferred 58 + unresolved 15, 13 sites) for
ever. Use that one.

## 4. WHAT IS LEFT — in order of value

### (1) ~~ARM 2 — the only thing keeping 352 open~~ → **MOVED to `bugs_open/390`, 2026-08-25**

**Do not work arm 2 from this file.** `bugs_open/390` supersedes everything below in this item: it
carries the first-hand verification, the reproduction commands, four fix candidates ordered by what
closes the door, and an explicit list of what is *not* done (blast radius `[UNMEASURED]`, fix
undesigned, `090` named as the right next spend). The sketch below is kept only because 390 §2
records where it was wrong, and a reader may want the original.

#### The superseded sketch

Live, reproducible, **not designed**. Sketch only, from the old handoff §7: `css-patch-agent`'s
workflow gains a **measurable precondition** — grep `css_themes` for a declaration governing the
filed selector's property; if the offending declaration is not in the file the agent can edit,
**refuse and park** with a `parked_by` marker (198's `mark_base_unsafe` shape) rather than append a
rule that cannot win. And completion should consult the spec's own `acceptance_test` at the
`checks.GetVerifier` / `verifyBeforeComplete` choke point — which
`write_audit_findings_verifier_join_test.go:85` confirms **nothing reads today**.

**Read `bugs_open/296` §10.5 first** — it reaches the same finding from the other end.
**This needs a real design pass and probably a `090` diagnosis run**, not a patch.

### (2) A DATED CHECK THAT COMES DUE 2026-08-28 — does the withdrawal actually re-detect?

587 freed 73 dedup slots on 13 sites on the promise that still-failing pairings return under
verified selectors. **[MEASURED 2026-08-25 09:40 UTC] 0 of 56 withdrawn (site, page) pairings have
returned — and that is expected, not a failure:** all 13 sites were last selected by the rotation
*before* 587 applied, and **0 have been re-audited since**.

The rotation's live `pre_query` window is **3 days** (not the 7 that WII-016 said — corrected there
2026-08-25). Earliest of the 13 due **2026-08-26 21:20 UTC**; all 13 by roughly **2026-08-27 21:30
UTC**. **So from 2026-08-28: any of the 13 with no re-filed `contrast_failure` and a visible contrast
fault is a defect in this promise.** The query is in §5.

### (3) THE TWO COUNTERS STILL HAVE NO READER — owed, and honest about it

`skipped_unverified_selector` / `skipped_unanchored_selector` ride the action's result map and its
log line and **nothing surfaces them**. Raised by the council's `bug_historian` seat (medium) and not
closed. They now certainly *fire* — the composition path is live and producing — and the cost is
demonstrated: the 17:33 audit's `write_render_audit_findings: complete` line, counters and all, was
gone from the chassis logs within the hour when the fleet rolled. **A counter whose only sink is a
log line on a service that restarts is not bookkeeping, it is a hope.**

### (4) NOT OURS, UNFILED ON PURPOSE — the render audit's timeout rate

[MEASURED 2026-08-24 19:08 UTC] **11 of 20** `render-audit-agent` runs **in one day** ended
`complete_error`, every one on `Request timed out (code: TIMEOUT)` at almost exactly 3 minutes, and
that rate **predates this lane's change**. ⚠ **That figure is not reproducible** —
`orchestration_states` is pruned to ~24 h and those rows are gone. Re-measure before quoting.

Prior art searched 2026-08-24: nothing in `/bugs_open/` or `/bugs_closed/` (nearest is `296`, a
different subject) and no open `needs_diagnosis`. **Left unfiled deliberately**: a symptom count is
not a cause, and CLAUDE.md's rule is that a cross-cutting root cause goes through `090` *before* it
is asserted. Whoever picks it up should run `090_TRIGGER_needs_diagnosis_v1.sh`, not write a
mechanism from a count.

### (5) UNPROVEN, and narrower than it sounds

[UNPROVEN] The locked-component interaction is narrowed, not proven closed — I never established
that the old substring check dropped a real finding; the mechanism permitted it.

## 5. THE QUERIES THAT DECIDE THINGS

```sql
-- WHICH SIDE OF 587 AM I ON? Run this FIRST.
SELECT count(*) FILTER (WHERE result->>'cancelled_by'='migration_587') AS withdrawn_by_587,
       max((result->>'cancelled_at')::timestamptz)                     AS applied_at
  FROM site_work_items WHERE item_type='contrast_failure';
-- >0 → post-migration; an open-population census returning 0 is the SUCCESS condition.

-- ARM 1 STILL HOLDING? (the roll was 2026-08-24 15:39 UTC)
SELECT count(*) AS rows_since_roll,
       count(*) FILTER (WHERE item_key ~ '#([A-Z][A-Z0-9]*)\.\1$') AS still_invented, -- MUST be 0
       count(*) FILTER (WHERE spec ? 'selector_scheme')            AS scheme_stamped,
       count(*) FILTER (WHERE spec ? 'matches')                    AS carries_matches
  FROM site_work_items
 WHERE item_type='contrast_failure' AND created_at > '2026-08-24 15:39:00+00';
-- ⚠ rows_since_roll = 0 means no audit found anything, NOT that the fix works.

-- THE 73, FOR EVER (the open census correctly returns 0 now)
SELECT result->>'pre_352_status' AS status_before_587, count(*), count(DISTINCT site_id) AS sites
  FROM site_work_items
 WHERE item_type='contrast_failure' AND result->>'cancelled_by'='migration_587'
 GROUP BY 1 ORDER BY 2 DESC;   -- deferred 58, unresolved 15 — 73 across 13 sites

-- ITEM (2), FROM 2026-08-28: have the withdrawn pairings come back?
WITH withdrawn AS (
  SELECT DISTINCT site_id, split_part(split_part(item_key,':',2),'#',1) AS page_path
    FROM site_work_items
   WHERE item_type='contrast_failure' AND result->>'cancelled_by'='migration_587')
SELECT (SELECT count(*) FROM withdrawn) AS withdrawn_pairings,
       (SELECT count(*) FROM (
          SELECT DISTINCT w.site_id, w.page_path FROM withdrawn w
            JOIN site_work_items n ON n.site_id=w.site_id AND n.item_type='contrast_failure'
             AND n.created_at > '2026-08-24 19:11:22+00'
             AND split_part(split_part(n.item_key,':',2),'#',1)=w.page_path) q) AS returned;

-- IS THE ROTATION IDLE OR STALLED? Never read liveness off last_triggered_at.
SELECT count(*) FILTER (WHERE s.status IN ('active','deployed')) AS active_sites,
       count(*) FILTER (WHERE s.status IN ('active','deployed')
         AND COALESCE(r.last_selected_at,'-infinity'::timestamptz) < now() - interval '3 days') AS due_now
  FROM sites s LEFT JOIN site_discovery_rotation r
    ON r.site_id=s.id AND r.agent_type='render-audit-agent';
-- due_now = 0 with an advancing last_triggered_at is IDLE BY DESIGN, not broken.
```

## 6. ⚠ THE TRAPS THIS LANE HIT — every one cost real time, none is obvious

- **A `[MEASURED <date>]` marker certifies the NUMBER and says NOTHING about the POPULATION.** Four
  figures were published wrong in two days and **every one was a population error, none arithmetic**:
  the wrong table (a live table read as all-history while an archive existed), the wrong slice (a
  `LIMIT`), the wrong sub-population (retracted rows counted as all filers), and the wrong window
  (`interval '7 days'` on a table pruned to 24 h). See `WRONG_CALLS.md`, four entries dated
  2026-08-24/25.
- **`site_work_items` is a ROLLING WINDOW and `site_work_items_archive` exists** (25,281 rows back to
  2026-02-22). Any "ever / never / all-history" question must union it.
- **`orchestration_states` is pruned to ~24 hours.** A 7-day filter on it silently answers a
  one-day question, and the figure is unreproducible afterwards.
- **A scheduled task that dispatches nothing looks identical to one that did** — `last_triggered_at`
  and `last_completed_at` both advance. Ask the `pre_query` how many rows are due.
- **`idx_swi_dedup` is `(site_id, item_key)` with NO `item_type`** — the key namespace is global
  across types. Landmine filed; baseline **20** cross-type pairs live ∪ archive, as a ratchet.
- **A retraction arm has no producer predicate** — `resolveWorkItems` closes by
  `(site_id, item_type, item_key)`, so on a shared type it closes other producers' live requests.
  Landmine filed. `contrast_failure` is safe because it has **exactly one** producer (VIZ-016).
- **A line number in a landmine has a shelf life of days** — `resolveWorkItems` moved ~48 lines in
  one day. Anchor on the string.
- **BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES while both controls pass** — use the
  NUL-split pipeline, and note it is slow: my probe on 2026-08-25 **timed out at 120 s before its
  controls ran**, which is not a probe.
- **The council gate is WORKING** (a sibling lane's handoff says it is down; that is stale).

## 7. OTHER LANES — all told, nothing owed back

- **`brochure_component_library`** (`bugs_open/296`): the 73 left their durable count on 2026-08-24;
  CONTRIB updated with how to tell a returning row (`spec ? 'selector_scheme'`) from a new fault, and
  the 108→111 correction.
- **`bugfix_122_contrast_ink_slots`** (`bugs_open/211`): the "six `.H3` headings" correction stands;
  the canary aimed at their site never ran and they were told.
- **`bugs_open/384`**: a long, productive exchange — they found a trap in a shared index, I found
  their key space was safe by measurement rather than convention, and their unprompted reversal of
  their own `Resolved` arm produced the retraction-authority landmine. Closed out.
- **`bugs_open/198`**: filed this bug and released it. Closed out.

## 8. WHERE EVERYTHING IS

- lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/`
  (`PLAN_2026-08-24`, `RUNBOOK` — **§10 is the one to read**, `NOTES` newest at the bottom,
  `README_where_we_are`, `SUMMARY_2026-08-24`)
- bug file: `bugs_open/352_HANDOFF_2026-08-22_contrast_findings_name_a_selector_that_matches_nothing.md`
- migration: `docs/agent_docs/sql_for_agents/587_retire_invented_contrast_selectors_HOLD.sql`
  (+ `_ROLLBACK`, + `_VERIFY`) — **applied 2026-08-24 19:11:22 UTC**
- register: **VIZ-016** (the selector contract and the shared `item_key` shape) and **WII-016** (the
  retraction seam; carries two dated corrections from this lane)
- landmines filed by this lane: the `_HOLD`-migration status one, `idx_swi_dedup` has no
  `item_type`, and the retraction-arm-has-no-producer-predicate one
