# HANDOFF — `bugs_open/357`, component identity — 2026-08-25 (evening)

**Cold-start for a fresh session. Read this, then
`HANDOFF_2026-08-25_continue_here.md` (this morning's — still accurate about phases 0/2/3
and the F2 guard, but its central open question is now ANSWERED here), then the bug file
`bugs_open/357_HANDOFF_2026-08-22_a_whole_tool_page_is_stored_in_a_slot_that_claims_to_be_a_hero_component.md`.**

> **PHASE 2 HAS FIRED IN PRODUCTION. That was the whole blocker this morning, and it is
> gone.** Two rows adopted, verified on every axis, live and serving. What replaced it as
> the lane's most important item is a **NEW, ESTATE-WIDE DEFECT** found on the way —
> unrelated to phase 2's correctness and much bigger than this lane.

---

## 1. State in one table

| | state | evidence |
|---|---|---|
| **Phase 0** — the provenance stamp reaches the DB | DONE, proven at volume | unchanged from this morning |
| **F2 guard** — a carried tool must not take the discarded render's stamp | PROVEN with demand | unchanged; `population_stamped` still **0** |
| **Phase 2** — stop the mislabelling at birth | **PROVEN IN PRODUCTION 2026-08-25 12:24Z** | 2 `adopted-fragment` rows, both regenerable + stamped, §3 |
| **Phase 3** — repair the 22 | **NOT RUN. Precondition 2 now MET; precondition 4 still UNTESTED** | §5 — this is the decision waiting |
| **The bug itself** | **OPEN. Population still 22** | re-measured 19:29Z |
| **NEW: the plan/route contradiction** | **DIAGNOSED, NOT FIXED** | §4 — the most valuable thing here |

Re-measured **2026-08-25 19:29Z**, after the chassis roll to `agent-chassis-669b45fdb4-*`:
`adopted_rows=2 · population=22 · population_stamped=0 · armed_carriers=6`.
**The roll changed none of it** — arming lives in `agent_definitions`, so it is config, not
code, and a roll cannot disarm it.

---

## 2. What was done today, in order

1. Ran **two site adoptions** on the owner's instruction — `cv1.co.uk` then `lampenkap.com`,
   via `082_submit_domain_unified.sh <domain> --from https://<domain>`, **no `--fidelity`**.
2. Both completed. `lampenkap.com` rebuilt and deployed (its index went to
   `page-build-handler`, 5 sections). `cv1.co.uk` routed **both** pages to
   `tool-recreation-handler` — the producer phase 2 was built for.
3. Both cv1 recreations produced perfect fragments and **both saves were refused whole** by
   `save_page_sections`' prune floor. That is §4, and it is the day's real finding.
4. Owner chose to **correct the two page plans** over relaxing the floor or waiting for the
   code fix. Applied as `OPERATION_2026-08-25_correct_cv1_tool_page_plans.sql` (+ `_ROLLBACK`).
5. Recreations re-queued → **both adopted**. Verified at the row and at the served artefact.
6. Canary rebuild fired to test precondition 4. **Killed by the chassis roll**, not by
   anything about 357. Re-fired; see §5.

---

## 3. THE RESULT — phase 2 is no longer an unproven mechanism

```
page                    slot_name           component         regenerable  stamped  bytes
cv1.co.uk/index         hero                adopted-fragment  t            t        17,595
cv1.co.uk/tool-example  generic-text-block  adopted-fragment  t            t        20,076
```

**`cv1.co.uk/index` IS `bugs_open/357`, and it is correct.** A 17,595-byte self-contained
interactive tool in a slot named `hero` — the exact configuration the bug was filed about —
and the row says `adopted-fragment`, with `content_data.body` reproducing `rendered_html`
byte for byte, carrying a real stamp. Before phase 2 that row would have declared itself the
shared hero component and joined the population.

- `slot_name` untouched on both (renaming it is what arms the carry-forward landmine).
- Both point at **the same** `component_versions` row `3301ef65-4d83-4ea5-aa7c-65cb38e83653`
  (template `{{.body}}`) — the "1.00 versions per component, not a log" property holding
  across two independent adoptions instead of being asserted from one.
- **At the artefact** (the only proof this estate accepts), both serve their tool:
  `cv1.co.uk/` 200/30,244 B and `/tools/example/index.html` 200/32,713 B, each carrying
  `tool-page` and **zero** `data-component="hero"`.
- All three STOP conditions checked. The one NULL-`component_id` row is
  `loanandmortgagecalculator.co.uk/tool-overpayment-priority`, created **31 minutes before my
  first dispatch**, on another lane's site, with **5 rows on its page** — so it came through
  the metadata path, `FallbackAdopted` was never set, and adoption was never offered it.
  ⚠ Its surface reading (no `<section>`, no `data-component`, no `component_id`) is exactly
  what a failed adoption looks like. **The row count is what discriminates.**

---

## 4. ⚠ THE NEW DEFECT — read this even if you only read one section

**`apply_adoption_plan` chooses the tool-recreation route for a page AND writes that same
page's multi-section plan, in one action, in one transaction — and the two contradict each
other.**

- Route: `apply_adoption_plan_action.go:719`, `if len(page.Features) > 0` →
  `needs_tool_recreation` → `tool-recreation-handler`.
- That handler declares `expects_no_sections_metadata`, so its save can only reach the HTML
  fallback, which **by construction emits exactly ONE section**.
- The same action writes `pages.sections` with 3 or 4 entries.
- `measurePageSectionCompleteness` (`save_sections_prune_floor.go:148`) divides 1 by that
  planned count, and `prune_floor_ratio=0.50` **refuses the whole save**. Nothing is written.
  The page keeps zero rows and a `save_refused_incomplete` item is parked for a human.

Measured on cv1 before the correction: `index` 1 of 4 = **25%**, `tool-example` 1 of 3 = **33%**.
The adoption's own analysis spec for `tool-example` reads **`"self_contained": true`** — on the
page it simultaneously planned with three sections.

**Any adopted interactive page planned with ≥3 sections is unsaveable.** 1/3 and 1/4 are both
below 0.50; 1/1 and 1/2 clear it (`Below` is `ratio < floor`, so exactly 0.50 passes).

**The cross-check that makes this the explanation for 357 itself** [MEASURED 2026-08-25]:

| planned sections on the page | rows in the 357 population |
|---|---|
| 1 | 1 |
| 2 | **20** |
| 4 | 1 |

**21 of 22.** The floor has been silently *selecting* which tool pages get a row at all:
≤2 planned → saved and mislabelled (that is 357); ≥3 planned → refused outright, page left
empty. So `adopted=0` was never evidence about phase 2 — it was evidence about a guard two
hundred lines below the binding.

**Estate-wide, not a cv1 curiosity:** **32** `save_refused_incomplete` items sit in
`needs_human_review` from 2026-07-31 to today across **~14 domains**, several named tool
pages — `webdesign.co.uk/tool-llm-cost-calculator` (1 of 4),
`fundamentallyai.com/tool-model-approach-selector` (1 of 3),
`mortgagecalculator.co.uk/game-fact-finder` (1 of 4). ⚠ Several older rows have EMPTY cohort
captures because the `planned sections` cohort postdates them — **a blank is a reason-string
format difference, not a different cause.**

⚠ **`site_work_items` is a rolling window.** Joining `needs_tool_recreation` to `pages` finds
only the two cv1 rows because the history is archived out. Do **not** read that as "only two
pages were ever routed to tool recreation".

**Status:** filed through the diagnosis loop — intake `f2fa4b9e-28b6-4f45-9ffa-2627c2031af0`,
**RUN_CORRELATION_ID `fbdaca97-a97e-41e6-b422-2475521e6a6c`**. It returned **UNVERIFIABLE**
(`scope-not-narrowing`), **not REFUTED**, naming two gaps — **both since closed first-hand**
and written up in NOTES (the work item's `source=adoption` / `created_by=site-adoption-agent`
/ `item_key=needs_page:tool-example`, and the floor's denominator read from the function body).

**Not fixed, and NOT mine to fix quietly.** The durable remedy is for `apply_adoption_plan` to
write a one-entry plan for a page it routes to tool recreation. That is a code change to a
shared seam, so it wants the council gate. **The owner has been told and has not commissioned
it** — he chose the per-page correction for today. It is the top candidate for a new bug file.

---

## 5. Phase 3 — what is met, what is not, and why 578 has NOT been run

`578_retype_mislabelled_tool_rows_HOLD.sql` targets all 22 rows.

| precondition | enforced by the file? | state |
|---|---|---|
| 1. phases 0/2 built, rolled, verified | no | **MET** |
| 2. an organically adopted row carrying a stamp exists | **YES — RAISEs** | **MET** (2 of them) |
| 3. 577 applied and carriers armed | **YES — RAISEs** | **MET** (6 armed, re-checked post-roll) |
| 4. **a canary rebuild ran on an adopted page and preserved it** | **NO — prose only** | **UNTESTED** |
| 5. re-census on the day | **YES — by predicate** | automatic |

**Precondition 4 is the only thing outstanding, and it is the one the file does not enforce.**

- Canary #1 (`e0c2d505-9875-4347-a718-a852f32ec6b7`) **FAILED** — reaped with
  `"reaper: stale EXECUTING_STEP for >4h; step=build_pages_loop_iter_0_assemble_page"`.
  The chassis rolled underneath it. `save_result` was never set: **`save_page_sections` never
  ran**, so it tested nothing.
- Canary #2 (`5a0cad41-fe0c-4636-9b2d-9c942486019c`) fired 19:30Z on the fresh build,
  `index` only, with **`tool-example` deliberately held untouched as a control**.

**⚠ READ THE CANARY WITH ITS DEMAND CONTROL OR IT WILL LIE TO YOU.** The step order in
`page-rebuild`'s sub-workflow is **not** intuitive:

```
plan_sections -> write_page_content -> review_page_content -> check_review_approved
   -> assemble_page -> deploy_page (git_commit) -> save_sections -> update_page_status
   -> complete_page
```

> **CORRECTED 2026-08-25 evening, before anyone acted on it.** An earlier draft of this
> section put `save_sections` immediately after `assemble_page` and `deploy_page` last. That
> was inferred from the orchestration sitting at `assemble_page` with no `save_result`, which
> is consistent with the true order and does not establish it. Traced properly from each
> step's own `next_step`, **`deploy_page` is a `git_commit` action and it sits BETWEEN
> assemble and save.** The load-bearing claim is unchanged — the save runs after assemble, so
> `save_result` is still the demand control — but the git commit in the middle is very likely
> why BOTH canaries stalled at `assemble_page`, and that is not something you would look for
> if you believed the save came next.

`save_sections` runs **AFTER** `assemble_page`. I read a pre-save state as a clean pass and
nearly wrote it into a summary: 1 row, right md5, right component, still stamped — a perfect
match against the pinned baseline, and **meaningless**, because the only step that could have
changed anything had not run. **The discriminating check is
`collected_data ? 'save_result'`.** Not the orchestration being alive, not the page being
flagged, not elapsed time.

Baseline to compare against, pinned in `scratchpad/canary_before.txt`:

```
index         pos 1  slot hero                md5 26f484f2744ab3e9cd19e50f600a52b8  17,595 B
tool-example  pos 1  slot generic-text-block  md5 291b88d876e182a32a4a538c514878d2  20,076 B
both: component 9d4b922b-a548-4ca2-987c-ecacc7904b1f, version 3301ef65-4d83-4ea5-aa7c-65cb38e83653
```

**Pass = `save_result` present AND row count still 1 AND md5 unchanged AND component still
`adopted-fragment`.** A row count of **2** is the carry-forward landmine firing and means
**STOP** — do not run 578.

### Then, and only then

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/578_retype_mislabelled_tool_rows_HOLD.sql
```

Apply **by hand** — never through the migration runner (`_HOLD` means exactly that). It
re-censuses by predicate on the day, backs every affected row into
`page_components_backup_357_20260823`, prints the six `owned` pages by name, and RAISEs if
bytes/`slot_name`/`position` moved or a repaired row is not reproducible from its own
`content_data`. Rollback: `578_..._ROLLBACK.sql`.

**Safety check already done for you:** 578's predicate keys on `cc.name='hero'`, and the
adopted-fragment component's `name` is **`"Adopted Fragment"`** — so it structurally cannot
sweep in the two new adopted rows. Verified by running the predicate against them: **0**.

**Owner's standing instruction (2026-08-25):** *continue to phase 3 on a clean verification.*
The adopted rows verify cleanly. **I did not run 578** because precondition 4 is still
untested — the canary died to the roll, and re-typing 22 live rows into a shape whose
durability under rebuild is unproven is exactly what that precondition exists to prevent.
**That is a deferral on evidence, not a refusal**; finish the canary and the instruction
stands.

---

## 6. Traps recorded today (all fleet-wide, all committed)

`LANDMINES.md`:
1. **`--fidelity locked` on an adoption SKIPS the entire build cascade.**
   `apply_adoption_plan_action.go:486` returns early into `adopt_verbatim.go`; the routing is
   at `:708`. The run completes, a real site appears, and **the code under test is never
   reached** — every counter reads exactly as before. This would have made the whole day
   silently pointless. Verifier: dispatched.
2. **An adopted page's INTERACTIVE classification does not track its HTML.** Verifier came
   back **STILL_VALID**, 7/7 checks matched.

`WRONG_CALLS.md`: **I predicted both sites' routing from markup and was wrong about both, in
opposite directions.** `lampenkap.com/index` carries a working lux calculator and went to the
STATIC builder; `cv1.co.uk/index` has **zero** `<script>` tags and zero controls and went to
tool recreation. I told the owner lampenkap was "the surest single shot"; he chose to run both,
which is the only reason there is a result at all.

Also in NOTES, not yet landmines:
- **spawned agent pods are ephemeral** (`agent-<type>-<hash>`) and reaped within minutes, so a
  run's `adopt fragment:` log lines are gone. Capture live or rely on the DB. My log monitor
  watched `-l app=agent-chassis`, the **wrong pod set**, and its silence meant nothing until I
  ran a control that must have matched and got **0**.
- **`HEAD~1` is not your commit's parent on this tree.** Chasing a ledger line my commit
  removed, three attempts diffed `HEAD~1..HEAD` and came back **empty** because another
  session had committed in between — an empty diff reads exactly like "no removal happened".
  Pin the sha.
- **Same-file passengers went BOTH ways in one hour:** my commit carried the 386 lane's
  in-place `LANDMINES.md` correction (nothing lost — they replaced a line, which is the
  prescribed form), and my `WRONG_CALLS.md` entry was committed by another session under
  their message. Both fine, both worth knowing.

---

## 7. Commands and keys

```bash
# state, one line
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='adopted-fragment') AS adopted,
       (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0) AS population;"

# the adopted rows, full verification
# (regenerable + stamped + slot_name untouched — the RUNBOOK's WORKED query)

# canary 2 — MUST read save_ran, not just status
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT status, current_step, (collected_data ? 'save_result') AS save_ran
FROM orchestration_states WHERE correlation_id='5a0cad41-fe0c-4636-9b2d-9c942486019c'::uuid AND owner_agent_type='page-rebuild';"
```

| what | key |
|---|---|
| cv1.co.uk adoption | corr `468cb727-d2c7-4299-b332-3fc36c0996c6` · site `8c3e9118-2455-4f0d-b01a-5dcde13dcf99` |
| lampenkap.com adoption | corr `a3e1a948-0979-4b0f-8592-cfbd979d9899` |
| diagnosis of the new defect | intake `f2fa4b9e-…` · **run `fbdaca97-a97e-41e6-b422-2475521e6a6c`** |
| canary #1 (failed, reaped) | `e0c2d505-9875-4347-a718-a852f32ec6b7` |
| canary #2 (in flight) | `5a0cad41-fe0c-4636-9b2d-9c942486019c` |
| adopted-fragment component | `9d4b922b-a548-4ca2-987c-ecacc7904b1f` ("Adopted Fragment") |
| its version row | `3301ef65-4d83-4ea5-aa7c-65cb38e83653` |

**Commits today:** `b0cf6e501` (the refusal finding) · `0d0cd6595` (passenger note) ·
`d7a0ff938` (diagnosis gaps closed) · `6419c5f63` (the plan correction + rollback) ·
`b5884518e` (phase 2 fired) · `2c720e9af` (the canary near-miss).
No council trailer on any: all are docs + a site-specific data correction, none in council scope.

---

## 8. What the next session should do, in order

0. **⚠ FIRST, THE CONTROL — because canary #2 stalled too.** As of 19:41Z canary #2 has sat
   at `build_pages_loop_iter_0_assemble_page` for 10m37s with `save_ran = f`, exactly where
   canary #1 was reaped. **Two attempts, same page, same step.** Before re-firing a third,
   flag a **non-adopted** cv1 page (`request-index` or `how-it-works-index`) as
   `needs_rebuild` and run the same rebuild:
   - it also stalls → the stall is the rebuild path (or the `git_commit` immediately after
     assemble), **nothing to do with 357**, and precondition 4 needs a different vehicle;
   - it reaches `save_result` → the stall is specific to the page carrying a 17.5KB adopted
     fragment, **which IS a 357 finding** and bears directly on whether phase 3 is safe.

   I did not run it because a third concurrent rebuild on the same site would muddy both.
   [MEASURED 19:40Z] the rolling window holds only **2** `page-rebuild` runs, both mine, so
   "no page-rebuild ever reaches the save" is true of the window and worthless as a claim —
   the denominator is what makes that zero unreadable.

1. **Read canary #2's `save_ran` first.** If it is `f`, the canary tested nothing again —
   re-fire; do not interpret the unchanged rows.
2. If it passed: **run 578**, then verify at the artefact per the migration's own afterword
   (curl each repaired page, re-run the population query, let ONE rebuild run and compare row
   counts and per-row md5, confirm the false `required_fields_missing` items stop being re-filed).
3. If it failed with a row count of 2: **STOP**, and the carry half needs work before any repair.
4. **Either way, file §4 as its own bug.** It is estate-wide, it has 32 parked victims, and it
   is not 357 — 357 is the mislabelling, this is the refusal. They share a cause and deserve
   separate files.
5. 357 closes only on the standing bar: **fixed AND live**. The population is still 22.
