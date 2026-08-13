# HANDOFF — bug 122 lane. START HERE. Written 2026-08-12 (evening), updated 2026-08-13.

> **STATUS IN ONE PARAGRAPH.** The retraction is **built, council-APPROVED, and LIVE on both
> services** since the 2026-08-13 roll. It has **never run**, and cannot until **2026-08-17
> ~14:54Z**, because the render-audit rotation has a 7-day per-site window and the fleet was
> stamped 08-10 — that is correct, do not force it. Everything owed is now a **dated
> verification**, not a build: on 08-17, `robot-hands.com` is first up, and §1c states exactly
> which rows must close and which must stay open. Read §1c, then §2 (the one correction), then §3.

Supersedes `HANDOFF_2026-08-12b_continue_here.md` for **state**: its §3 task is **DONE and
committed** (`5639a1103`). That file stays the reference for **why the design is shaped this way**
(§3.1's two constraints, §3.3's four hazards) — all of it still true, and §3.2 has **one correction**
recorded below that you must read before touching the code. `HANDOFF_2026-08-12` remains the
reference for the owner's two decisions and the 📅 2026-08-16 action, **which is unchanged and
still owed**.

**Nothing in this lane is on fire.** The queue is empty, no site is locked out, the 226 are parked
and stable, and the code that just shipped **does nothing at all** until two services roll.

---

## 1. State, measured 2026-08-12 evening

| thing | state |
|---|---|
| the retraction | **BUILT + COMMITTED `5639a1103`**, 8 files, HEAD verified to build and test green in a clean `git archive` tree |
| council | **APPROVED** — trail corr `a43b63d6-da35-4136-9471-88ec6ace799a` (round 1 killed by a roll, §1a; round 2 approved 20:01Z). **13 reviewers, 4 abstained, `unreadable: 0`, `gated_by_truncation: false`** — the approval is on substance, not on seats that failed to render. 5 advisory objections, none high, all checked — §1b |
| is it live? | **YES, BOTH HALVES, as of the 2026-08-13 roll — but NOT YET EXERCISED, and it cannot be until 08-17.** Both services stamped `69612d692`, which carries `5639a1103`. See §1c |

### 1a. The roll of 2026-08-12 19:13Z — what it did and did not do

**It killed the council round.** Last `updated_at` **19:13:18Z**; chassis pods started
**19:13:54Z** / **19:14:16Z** — frozen 36 seconds before the first new pod, then stuck at
`review_tooling_provenance` for 25 minutes. This is a known trap with a written remedy
(`LANDMINES.md`, *"A chassis roll KILLS an in-flight council"*), which was followed: compare
`updated_at` to pod age, wait out the ~300s post-restart window, **resubmit with
`RESUBMIT_CORR=`**. Resubmitting on the **trail** rather than fresh is not cosmetic — the commit
was already pushed carrying `Council-Submitted: a43b63d6`, and a fresh correlation would strand
that trailer on a run that can never produce a verdict, which forward-only forbids amending.

**It did NOT ship this lane's change**, and the useful general lesson is that those are
independent facts: the roll you watch happen is not evidence about your commit. Both
`agent-chassis` and `browser-runner-adapter` came up on the same commit seconds apart — so **a
fleet release does roll the adapter with the chassis, and step (b) below is ONE release, not two
separate rolls.** That is measured, and it is the one piece of good news here.

### 1b. The verdict's 5 advisory objections, and what was done about each

None were high-severity and none blocked. All were checked rather than waved — two produced real
work, and one of those was a genuine gap in my own evidence.

1. **`guardian` — "you enumerated CONSUMERS of the response but never PRODUCERS of the item_type."**
   The best objection of the round, and correct: the seam's landmine says adopting retraction on a
   **co-filed** type silently closes the other producer's findings (the reason
   `check_undeployed_assets` was rejected as WII-009's first adopter). **Measured afterwards, and it
   could have come out otherwise:** all 226 rows are `source='render-audit'` /
   `created_by='render-audit-agent'`, **one** distinct spec key-set, no `audit_source` label; one Go
   file files the type. ⚠ **The census sees producers that have FILED — a never-fired producer is
   invisible to it. Re-run it before anything else starts filing `contrast_failure`.**
   > **2026-08-13 — FULLY RE-RUN, both halves, and nothing has changed.** Code side at HEAD: still
   > exactly one minting file, `write_render_audit_findings_action.go` (`:296`, `:304`, `:535`,
   > `:547`, `:604`). Row side over all 226: **one** `source` (`render-audit`), **one** `created_by`
   > (`render-audit-agent`), `spec ? 'audit_source'` **false** on every row, and exactly **one**
   > distinct spec key-set. No second producer has appeared in code or in the rows. Re-run it again
   > before anything else starts filing `contrast_failure` — a never-fired producer stays invisible
   > to the row half, which is the whole reason the caveat exists.
2. **`bug_historian` — the URL-shape contract.** A key built from today's URL shape is compared
   against one a PAST filing wrote, and nothing pinned which way a mismatch fails. **Fixed in code:**
   `TestWriteRenderAuditFindings_ShorterPageDoesNotPrefixMatchALongerOne` pins that `/pricing` never
   prefix-matches a key belonging to `/pricing.html` — the dangerous direction. Removing the `#`
   from the prefix makes it RED.
3. **`reuse_agent` — "why not extend `revalidate_review_queue` / `reviewRevalidators`?"** A fair
   Step Zero I had not shown. **Answered, and the answer is decisive on two independent grounds:**
   that action runs **in-chassis, which has no Chromium**, so it structurally cannot take a contrast
   measurement (the whole reason this lives on the audit path); and its selector is bounded by
   `workItemRevalidatableStatuses` = {`needs_human_review`, `unresolved`}, which excludes `deferred`
   — widening it is a fleet-wide change to a shared selector whose own comment records `failed`
   being deliberately excluded after measuring the blast radius.
4. **`bug_historian` — "this is a symptom fix."** Accepted, unfixed, and already stated: it patches
   ONE type around the generic hole that `GetVerifier` completes any unregistered type untouched.
   `dark_section_audit` keeps the identical exposure. `bugs_open/213`'s call, §4 item 3.
5. **`architecture` (low) — the third instance rule.** This is the **second** producer to hand-roll
   "still-failing set built before locks/caps, retract via `resolveWorkItems`" inline. Two is not a
   pattern; **a third must extract a shared helper instead of copy-pasting.** Written into WII-016
   so whoever writes the third finds it. Its verdict recorded `ARCHITECTURE_SIGNAL: point_fix` and
   explicitly declined to call the wire addition architecture-scope under RFC_022's narrow
   exception — *because* the consumers were enumerated rather than asserted.

**Residual risk nobody has closed** (`guardian`'s "missing"): two CONCURRENT render-audit runs on
one site are not reasoned about. `batch_id` only protects a run from its own rows, so run B can
retract a row run A filed seconds earlier if B did not re-measure that pairing. Defensible ("B
looked and did not find it") but untested and unmeasured.
| 226 `contrast_failure` items | still **parked** `deferred`, `parked_by='migration_389'`, `max(attempt_count)=0`, **0 ever completed** — unchanged, and re-measured this session |
| register | `WII-016` added (work-item-integrity), index row added. Nothing to drop from `102_coverage_ratchet.txt` |
| `LANDMINES.md` | one entry added + synced: *a parked (`deferred`) work item is retractable* |
| `site-render-audit-rotation` | enabled, weekly per site, zero LLM spend — this is what will drive the retraction |
| `site-discovery-rotation-quality` | enabled, **inert until 2026-08-16 09:49Z** — correct and waiting, do not "fix" it |

---

### 1c. 2026-08-13 — IT IS LIVE, and the first exercise is a dated, falsifiable prediction

**Live, verified at the artefact with controls** (not at the tag, not at git). Both `agent-chassis`
and `browser-runner-adapter` are stamped `69612d692a4a07d61eea3f648e1152e0fd36fd0a`, and
`git merge-base --is-ancestor 5639a11036752603dc0cfb036ff83e947fd11a44 69612d692…` → **true**, with
a positive and a negative control both behaving. Chassis startup line had scrolled, so it was the
binary probe: `69612d692` PRESENT, yesterday's `7a1887e31` ABSENT (so the probe discriminates).
⚠ **Grepping the binary for MY OWN commit returns `absent` and that is CORRECT** — only the build's
own sha is stamped, never every ancestor. Ancestry is a git query, not a grep. Both traps are now
in `LANDMINES.md`.

> **RE-VERIFIED 2026-08-13 evening, on the v1.0.1295 pods — and this needed redoing, it was not
> paranoia.** The fleet was released again at 13:53Z, *after* the check above, and a release rolls
> every service — so the morning's verification had expired as evidence. Re-probed the running
> chassis binary (`/proc/1/exe`): `69612d692…` **PRESENT**, `7a1887e31` **absent** (the control, so
> the probe still discriminates); `browser-runner-adapter` and `render-audit-adapter` both print
> `69612d692…` in their own provenance line; `git merge-base --is-ancestor 5639a1103 69612d692…` →
> **true**. **A newer release did not displace the retraction.** Re-run this after any roll — "it
> was live this morning" is a claim about this morning.

**⚠ IT CANNOT FIRE UNTIL 2026-08-17 ~14:54Z, AND THAT IS THE MECHANISM WORKING.** Measured
2026-08-13 14:10Z: `site-render-audit-rotation` is enabled, hourly, `last_triggered_at` 13:27Z —
and **0 sites are due.** Its `pre_query` selects sites whose `last_selected_at` is older than
**7 days**, and the whole active fleet was stamped 08-10 by the audits that produced our 226
findings. A fire with no due site dispatches nothing and advances `last_triggered_at` exactly like
a working one (`enabled` + a fresh tick ≠ ever ran — there are **no** `render-audit-agent`
orchestrations in 10 hours). **Do not "fix" this. Do not force a run to see it work.**

**THE CANARY IS FREE, so debug_historian's "count before you write" gate costs nothing.** The
`pre_query` is `LIMIT 1`, so the first exercise IS a single-site run by construction. Baseline
captured 2026-08-13 **before** any pass: `retracted_so_far` = **0**, all 226 still `deferred`, 0
carrying `batch_id`.

**The prediction — write the outcome against this, and treat a mismatch as a scope bug:**

| | |
|---|---|
| first site up | **`robot-hands.com`**, last audited 2026-08-10 14:54:23Z, **due 2026-08-17 14:54:23Z** |
| its ceiling | **34** open contrast rows across **21** pages. A first pass may close AT MOST 34, and only on pages it measured |
| next two | `loancalculator.co.uk` (08-17 15:54Z, **0** open rows — a pass that retracts anything there is wrong) · `cookly.uk` (08-17 16:55Z, 7 rows / 3 pages) |

**And the single sharpest test, which is free and comes with its own control.** `/selection-guide.html`
on that site holds exactly **three** open rows:

- `contrast_failure:/selection-guide.html#A.info-card-grid__card-link` → **must RETRACT** if migration `368` fixed it
- `contrast_failure:/selection-guide.html#SPAN.info-card-grid__eyebrow` → **must RETRACT**, same reason
- `contrast_failure:/selection-guide.html#A.cta-btn` → **must STAY OPEN.** `368` did not touch it

Same page, same run, opposite required outcomes. That distinguishes "retraction works" from
"retraction closes everything on any page it looked at" — which no count of closures alone can do.
If all three close, the scope is wrong: read §1b(1) and stop.

**⚠ READ THIS BEFORE YOU GRADE THE CANARY (added 2026-08-13 evening).** Two of those three rows
are `--color-primary-ink` consumers, and **another live thread is changing what that variable
computes to.** Verified in the served page, not the template:

```
robot-hands.com/selection-guide.html
  :554  .info-card-grid__eyebrow    color: var(--color-primary-ink, …)   <- ink consumer
  :648  .info-card-grid__card-link  color: var(--color-primary-ink, …)   <- ink consumer
  :805  .cta-btn-primary            color: var(--color-cta-bg, …)        <- NOT an ink consumer
```

Session `581eb30a` has an uncommitted contribution in `bugs_open/122` proposing to repair
`legibleInkFor` so the ink becomes a lightened relative of the source instead of `--color-text`
(verified real: 3 of its 16 sites re-measured at the artefact, plus the deciding line
`palette_specialised_slots.go:350` — see `NOTES` 2026-08-13 §3). **If that lands AND robot-hands
re-renders before 14:54:23Z on 08-17, the two "must RETRACT" rows stop testing retraction and start
testing the new derivation** — and a mismatch would read as "retraction is broken" when nothing of
ours moved. It takes BOTH a roll and a per-site re-render, so it is not likely; it is just silent
if it happens.

- **The `A.cta-btn` control is immune** — different variable — so the over-closure half of the test
  survives either way. Grade that one with full confidence regardless.
- **If robot-hands has re-rendered:** re-fetch `/selection-guide.html`, re-read the two `-ink`
  declarations and the value `--color-primary-ink` now resolves to, and re-state the two positive
  predictions against what is actually served **before** the audit fires. Do not grade them against
  this table.
- Check cheaply: `curl -s https://robot-hands.com/assets/css/styles.css | grep -- '--color-primary-ink'`
  → `#E2E8F0` (== `--color-text`) is the **unchanged** state this table assumes. A navy means the
  derivation fix landed.

**And `A.cta-btn` now has a mechanism, not just "368 did not touch it"** (`NOTES` 2026-08-13 §5).
`.cta-btn-primary` sets `color: var(--color-cta-bg, var(--color-primary))`, and `--color-cta-bg` on
that site is a **`linear-gradient(...)`** — not a valid `<color>`. The variable *is* defined, so the
sane fallback (`--color-primary`, `#1A1F2E`) is **never reached**; the declaration is invalid at
computed-value time and `color` **inherits** `#ffffff` from `.cta-section`, over a `#ffffff` button.
**CONFIRMED at the instrument, not inferred** — the token came back mid-session and the row's own
spec says `fg: "rgb(255, 255, 255)"`, `bg: "rgb(255,255,255)"`, `ratio: 1`, `text_sample:
"Run MatchMatrix"` (which is the **primary** button's `href`, so it is the right element). Any other
colour pair would have refuted it.

`[MEASURED]` **16 of 17** filed `%cta-btn%` rows fleet-wide are this exact defect — robot-hands 10,
finetuning.uk 4, ai-agent-orchestration.com 2 — and the 17th is the control: leopardess's
`A.tool-cta-btn-primary` at **2.27:1**, a valid token used validly and simply too pale. The
effective `--color-cta-bg` is a gradient on **5 of 10** sites sampled, which is why it survived.

⚠ **Resolve that token at the layer that WINS, not the one that is easy to `curl`.** My first pass
read only `/assets/css/styles.css`, reported "3 of 8", and was then apparently refuted by
ai-agent-orchestration.com (site-level `#0D1117`, yet white-on-white rows). The refutation was my
measurement: that site's **page `<style>` block** redefines the token to the gradient. Same literal
`linear-gradient(135deg,#1e40af,#1e3a8a)` appears page-level on 4 unrelated sites — a shared
template default, `bugs_open/113`'s shape, not this lane's to chase. Full working: `NOTES` §5/§5a.

No ink-slot fix and no repoint can close this row; it needs the consumer to stop reading a
background-typed token into a colour slot.

---

## 2. ⚠ THE ONE CORRECTION TO `12b` §3.2 — read this before editing the retraction

`12b` §3.2 step 1 said to build the still-failing set from *"exactly the keys already computed at
`:266`"*. **That is wrong, it was not followed, and the code does the other thing.**

`:266` is reached only by findings that survive two filters above it: the **locked-component skip**
(we can see the fault but a human has locked that component, so we never file it) and the
**`max_items` cap** (default 60, worst-ratio-first, remainder dropped). A finding removed by either
was **measured and is still failing**. Building the retraction set from the FILED items reads "not
filed" as "fixed" and closes those tickets — the false completion the park exists to prevent.

So the set is built from `payload.Contrast` **before every filter that decides what to file**, and
`over_image` approximations count as still-observed (an unknown backdrop is "I could not tell", not
evidence of health). **If you change one thing in this function, do not change that.** It is pinned
by `TestWriteRenderAuditFindings_MeasuredButUnfiledFindingsAreNotRetracted`, and reverting it to
the handoff's spec makes that test RED — which is how it was proven rather than argued.

---

## 3. NEXT, in order. Do not skip to the unpark.

**(a) ~~Read the council verdict~~ DONE — APPROVED, and the objections are actioned in §1b.**
Nothing is owed here. Do **not** hand-write a `Council-Reviewed:` trailer onto `5639a1103`:
forward-only forbids the amend, and `098` credits it automatically from its `Council-Submitted:`
trailer now the correlation is approved.
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'a43b63d6-da35-4136-9471-88ec6ace799a'
 ORDER BY created_at DESC LIMIT 1;                        -- COMPLETED, not EXECUTING_STEP
SELECT created_at, body FROM doc_notes WHERE categories ? 'council-gate'
 AND body LIKE '%a43b63d6%' ORDER BY created_at DESC LIMIT 3;
```
Two things the submission asks the council to rule on explicitly, so read for these: **(1)** that
retraction closes PARKED rows, and **(2)** the §2 correction above. If the verdict turns APPROVED,
`098` credits the commit automatically — **do not write a `Council-Reviewed:` trailer by hand.**

**(b) ~~Get it into a release~~ DONE 2026-08-13 — both services live on `69612d692`.** Nothing owed.
Confirmed twice over that a fleet release rolls the adapter with the chassis (08-12: same commit 29
seconds apart; 08-13: same commit again). Recorded because the pre-08-13 version of this handoff
treated "roll both services" as two things to coordinate, and it is one.
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
kubectl -n ai-persona-system logs -l app=browser-runner-adapter --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 5639a1103 <the stamp>   # "did my fix ship?" is a QUERY
```
An empty grep means **"not in range"** (it is a startup line and scrolls), never "unstamped" — fall
back to the binary probe with a known-present and a known-absent control. A release can straddle
sessions' commits, so read the stamp of the service you actually mean (`bugs_open/249`).

**(c) COUNT BEFORE YOU LET IT WRITE — DONE for the first pass; re-run it if 08-17 slips.** Raised by
the `debug_historian` seat and it was right: this closes 226 already-parked production rows and the
plan had no read-only count of what the first pass would take. **Baseline and per-site prediction
are captured in §1c**, taken before the mechanism could fire. Re-run the query below if you arrive
after 08-17 and want the ceiling as it then stands:
```sql
-- the ceiling: open contrast rows, by site and by how many distinct pages they span.
-- A retraction can only ever touch rows whose page the run actually measured, so
-- per-site "pages with open rows" bounds what one site's audit could close.
SELECT site_id,
       count(*)                                                   AS open_rows,
       count(DISTINCT split_part(substr(item_key, 18), '#', 1))   AS pages_spanned
FROM site_work_items
WHERE item_type='contrast_failure' AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
GROUP BY site_id ORDER BY open_rows DESC;
```
Then let **ONE** site's audit fire and reconcile its `retracted` / `retracted_parked` against that
site's row before the rotation runs fleet-wide. A first pass that closes far more than the site's
`open_rows`, or that closes anything on a site whose audit did not run, means the scope is wrong —
stop and read §1b(1). Backup already exists at
`scratchpad/backups/backup_park_contrast_failure_20260811.tsv`.

**(d) Watch that audit, then grade it at the artefact.** The first retraction will come from
a site's own `site-render-audit-rotation` fire. Grade the ROWS, never the run's status:
```sql
-- did anything retract, and did the mechanism say so?
SELECT item_key, status, result->>'resolved_by', result->>'reason', result->>'resolved_at'
FROM site_work_items WHERE item_type='contrast_failure' AND result ? 'resolved_at'
ORDER BY result->>'resolved_at' DESC LIMIT 20;
-- the park, which will now drain on its own
SELECT status, count(*), count(*) FILTER (WHERE result ? 'resolved_at') AS retracted
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY status;
```
The action's own result carries `retracted`, `retracted_parked`, `retraction_scope_pages`, and
`retraction_unavailable:true` when the adapter is too old to say what it measured. **If you see
`retraction_unavailable`, the browser-runner-adapter has not rolled** — that is the version-skew
branch working, not a bug.

**(e) THEN unpark, and only then.** One UPDATE at the foot of migration `389`, predicated on
`spec->>'parked_by' = 'migration_389'`. Row-level backup at
`scratchpad/backups/backup_park_contrast_failure_20260811.tsv`. Expect the remainder to be
**smaller than 226** by the time you get here — that is the point, not a discrepancy.

**📅 2026-08-16 — unchanged and still owed.** Price the discovery-rotation ramp: calls AND tokens
(baseline ~248k input tok/h idle; the driven sweep was ~806k/h). Queries at the foot of
`sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql`.

---

## 4. Also still open in this lane

| | item | status |
|---|---|---|
| 1 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **Owner's.** Architecture, not a bug patch. Unchanged |
| 2 | `bugs_open/242` — the silent 25-page cap | **DONE, live `v1.0.1288`.** It was the PRECONDITION for this session's work: it shipped the page COUNTS, this shipped the IDENTITIES |
| 3 | `dark_section_audit` | Straddles the same hole with the same unsound reason, which is now the ONLY entry still making that argument — the cross-reference in `verifier_coverage_test.go` was corrected without deciding it. **`bugs_open/213`'s call or the owner's, not ours** |
| 4 | Free cross-check | if a lane re-renders robot-hands `/selection-guide.html`, the audit filed `info-card-grid__card-link` + `__eyebrow` failures and migration `368` should close both. **Grade at the next audit, never at the item status** |

---

## 5. Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted failure closed.
- **A filed count is not a found count.** "34 findings" was 171 firm — 111 dropped by a cap. And
  **226 is a FLOOR, not a census**: the audit was capped at 25 pages until `v1.0.1288`.
- **Read the selection before asserting it excludes your rows.**
- **A `file:line` in a handoff is a pointer, not a quotation** — this fired TWICE in two days now.
  Yesterday on `:171` (stale line), today on `:266` (right line, wrong SET — see §2). Open the file.
- **A pathspec commit still takes a same-file passenger.** Expect it; verify at HEAD, not the tree.
- **`pages.sections` is an array of plain strings**; an object-shaped census returns 0 rows silently.
- **Never run `run-migrations.sh --apply` on this tree.**
- **A call count does not price an LLM loop.** Threshold the expensive unit.
- **Put a row COUNT you could be wrong about in your post-check.**
- **NEW — a passing mock cannot assert a negative; MUTATE.** All four retraction guards were proven
  by reverting each to its plausible-but-wrong form and confirming the matching test went RED. The
  first mutation IS the handoff's own spec, which is how §2 was established rather than argued.
- **NEW — `deferred` is in NEITHER status list.** Not terminal (holds its dedup slot), not closed
  (any retraction may close it). "Parked" does not mean "nothing will touch this". See LANDMINES.
