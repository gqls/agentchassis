# HANDOFF — 2026-08-18d, fresh chat starts here: four of the six owner decisions are shipped, and the interesting thing was a REVISE

**Supersedes `HANDOFF_2026-08-18c_continue_here.md`'s §6 path** — steps 1, 2, 3 and 5 are done.
Everything else in `08-18c` (the diagnosis behind decision 1, the canary reasoning, the 083 state)
stands and is not repeated. **Read this from disk, then `NOTES_required_fields_repair.md` from the
bottom.** All figures measured 2026-08-18 evening unless dated otherwise.

> **`08-18c` §5 carries one claim that is WRONG and worth striking as you read it:**
> *"⚠ `page_id` is **not** in the dispatch loop's `call_handler` input_mapping (verified live)"*.
> It IS there, as `page_id?` — the `?` drops it when the row's `page_id` is NULL, which is how the
> claim was probably formed. Verified against `build-dispatch-loop`'s live `sub_workflow`. The 300
> fix does not depend on it either way.

---

## 0. State

| thing | state | re-verify with |
|---|---|---|
| decision 2 — reclaim arm | **DONE** (mig `479`, `f95504674`). Never yet exercised on a real reclaim | scheduler log, `reclaimed` column |
| decision 1 Tier 1 — refusal status | **BUILT + COMMITTED `6aee22b00`; config LIVE (mig `480`); Go INERT until a roll.** Council **APPROVED at round 3, verdict READ** (`725b1f01`) — **both REVISEs found real defects**, see §1 | §1 |
| decision 4 — `bugs_open/300` | **BUILT + COMMITTED `42a4bf441`; INERT until a roll.** Council **APPROVED r1** (`203d858b`), **verdict READ**, 4 mediums answered; registered **WII-020** | §2 |
| decision 6 — gate cannot review config | **FILED, `bugs_open/314`** + 016b §9 pattern (`6826d2385`) | §3 |
| decision 3 — both canaries | left alone, as ruled. Unchanged | `08-18c` §4 |
| decision 5 — close `083` | **due ~2026-08-25**, unchanged | §4 |
| held pile | **15 rows / 4 pairs**, still all correctly held, nothing escalated yet | promoter log's `held` field |

**Nothing in §1 or §2 is behaving differently in production yet.** Both are Go changes waiting on a
chassis roll. The migration half of Tier 1 is live and is inert by construction until then.

---

## 1. DECISION 1 TIER 1 — shipped, and the council found something I had missed

### What it does

`SavePageSectionsAction`'s ownership refusal now leads its error with `ownedPageSkipReasonPrefix`
(`OWNED_PAGE_GUARD`). `update_work_item_status` gains ONE opt-in key,
**`owned_page_refusal_status`** — absent everywhere and therefore inert everywhere — which, when the
routed step error carries that marker, records that status instead of the configured one and stamps
`result->'owned_page_refusal'`. Migration `480` opts exactly one step in
(`page-build-handler.mark_item_failed → wont_fix`). A genuine save failure on the same step still
records `failed` and still counts, which is the load-bearing half: downgrading *every* failure would
blind the promoter's floor to real incompetence, which is worse than the bug.

Registered as **WII-019** (work-item-integrity) with the enumerated consumers and the blast radius.

### The REVISE, because it is the most useful thing here

`editquality` gated round 1 on a question I had not asked: **the handler writes the status, but the
dispatch loop runs afterwards — does the write survive?** It cited two `LANDMINES.md` entries saying
the loop replaces what a handler wrote. The entries existed; I had grepped LANDMINES for the symbol
I was *adding* (`wont_fix`) and not for the mechanism I was *writing through*.

The answer is a split, and it is measured:

| writer | guard | effect on `wont_fix` |
|---|---|---|
| `CompleteWorkItemAction` (`load_work_item_actions.go:1017-1025`) | `status NOT IN (…,'wont_fix',…)` | 0 rows matched — **preserved** |
| `failUnverifiedCompletion` (`complete_work_item_verification.go:428-429`) | identical list | **preserved** |
| `FailWorkItemAction` (`load_work_item_actions.go:1146-1160`) | **NONE** | **overwritten** to `triaged`/`failed` |

Of 115 owned-page refusals: **113 have `handled_by IS NULL`** (the handler's own write, untouched),
**2 went through `fail_work_item`**. So ~98% works; the other ~2% is today's behaviour unchanged, so
incomplete coverage rather than a regression. Positive control, because 113/115 alone reads as
"nothing overwrites anything": **122 rows at `needs_human_review` with
`handled_by='page-build-handler'`** — a handler-set protected status surviving this exact loop.

⚠ **This was UNOBSERVABLE before today.** Both paths write `failed`, so no query on the existing
population separates "the handler said failed and it stood" from "the loop overwrote it". The 2 rows
are visible only because `handled_by` distinguishes the writers. **The change is what makes the
defect visible** — which is an argument for shipping it, and means "I checked the live rows" would
have been false reassurance.

The `fail_work_item` gap is **contributed to `bugs_open/307`** (owned by `staged_component_build`,
active, same function) rather than fixed here: shared action, fleet-wide retry semantics, and
`failed` is itself in the sibling guard list so a naive copy would stop the retry ladder. The split
it probably wants is in the contribution.

### The SECOND revise, which retracted a claim in my own submission

Round 2 was gated by `prior_art_librarian`, **HIGH**: *"'nothing converts a detector's finding into
its `field_updates` payload' — a load-bearing absence claim used to justify punting the actual repair
to an unwritten RFC. The instructions' own worked example is a five-month-old, 3-run mechanism
declared nonexistent this same way."* **It was right; see the Tier 2 box in §4.** The claim is
retracted and corrected in four documents. Nothing in Tier 1 depended on it.

Seven other objections were answered by check rather than argument, and three of those checks are
worth keeping because they are re-runnable evidence about this change:

- **nothing string-matches the refusal's current message** — 0 hits across `platform/ internal/ pkg/
  cmd/ scripts/`, 0 live `agent_definitions`, and the 1 `scheduled_tasks` hit is prose in this lane's
  own remedy note, not a predicate. So re-wording the error was safe.
- **`page-build-handler` has exactly ONE definition row, version 1** — the documented
  two-active-rows landmine does not apply to this type. Checked independently of the migration's own
  guard, which is what `debug_historian` asked for.
- **the opt-in key is on exactly ONE live step**, by `jsonb_each` over every live definition rather
  than by assertion.

`guardian` also asked the right question about a shared action: *behaviourally inert is not
crash-safe*. **Seven cases added** driving the new block with a number / bool / empty / nil config
value and a `__step_error` that is a bare string, an array, or has a numeric message — each must
leave the configured status untouched and must not error.

### What is owed

**Post-roll verification, with BOTH controls** (query in `RUNBOOK`): an owned-page refusal must land
`wont_fix` carrying `result ? 'owned_page_refusal'`, **and** a genuine save failure on the same
handler must still land `failed` without the stamp. A zero on the control means no genuine failures
happened in the window, **not** that the split works — widen the window.

~~**Read the round-2 verdict** (`725b1f01`) and act on it.~~ **DONE — approved at round 3, verdict
read, advisories acted on.** The earlier commits carry `Council-Submitted:` and `098` credits the
correlation automatically; no amend, forward-only.

**Two residuals recorded rather than closed**, both from the round-3 advisories:
- **A `wont_fix` refusal is never re-validated.** It is excluded from retraction AND released by the
  dedup index, so if a page's `rebuild_policy` later flips `owned → generic`, nothing revisits the
  closed refusal. Harmless today (the finding re-raises and dispatches normally), recorded because a
  queue with no re-validation is a shape this estate has been bitten by.
- **⚠ Do not spread the marker trick.** This decides a terminal status by scanning error text for a
  prefix, accepted here because the seam offers no other channel. **If a second action wants the
  same thing, that is the point to give the coordinator structured error metadata instead** — two
  call sites make it a shared contract nobody declared. The `architecture` seat flagged it for
  exactly the next instance.

---

## 2. DECISION 4 — `bugs_open/300`, and its own figures had moved

`resolveStatusRepairComponent` resolves the drift finding's subject by `(page_id, slot_name)`,
reached by joining `site_work_items` through the unconditionally-mapped `work_item_id`. The stored
id is the **tiebreak within an ambiguous pair**; genuine ambiguity is refused rather than guessed;
and it falls back to the stored id when the pair yields nothing, so every existing caller is
unchanged. Every outcome now carries `resolved_by`.

**The exposure is bigger than the bug recorded, and the ageing is now observed.** All 82 lifetime
rows: `spec.page_component_id` resolves for **70** (12 dead, 15% — the bug said 1 of 20);
`(page_id, slot_name)` resolves for **82 of 82**. The bug recorded on 08-17 that all 16 deferred ids
still resolved — **11 do today**. Five died in a day, in a queue nobody touched.

⚠ **The tiebreak is not decoration, and this is the trap.** `(page_id, slot_name)` is NOT unique:
**17 such pairs on the estate carry more than one component, worst case 4.** None is a drift row
today — so resolving by the pair alone is correct on every row that exists now and silently wrong on
the first one that is not, and **no query against current data would ever object.** It is held by a
test, not by evidence.

**Residual, stated:** if both keys miss (the page is gone) it is still a hard error and still feeds
the floor. Candidate 1 is what was approved; softening a genuinely-missing subject is a different
judgement about what `complete` means here.

**Owed:** verification at the artefact after the roll, with the bug file's own two controls — a
dead-id row must close without failing **and** a genuinely-true drift must still be repaired, or "no
more failures" is equally consistent with having made the handler blind.

---

## 3. DECISION 6 — `bugs_open/314` filed

The gate's scope test is `SCOPE_RE='^(platform|internal|pkg)/'` — a **path** test standing in for a
**subject-matter** rule ("prose does not spend council credits"), and this estate's config ships as
SQL under `docs/agent_docs/sql_for_agents/`. **The gate is not declining to review config; it cannot
see that it is config.** [MEASURED, 14 days] 227 commits ship a numbered migration, **152 (67%)
carry no in-scope file** and are refused by construction; the other 33% pass on a Go half that may
be one line, because the test is `length > 0`.

Four candidates costed, including the cost of the preferred one (widening `SCOPE_RE` makes ~152
commits/fortnight newly eligible, and two seats always run). Verify needs a positive and **two**
negative controls.

⚠ **The prior art nearly stopped this being filed.**
`architecture_review/DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §8d cites the same
line and concludes the refusal is **correct** — and it is, *for prose*, which is what §8d was
arguing about. The transferable half is in 016b §9: finding prior art that endorses a rule is not
evidence about your case.

---

## 4. WHAT IS LEFT

1. **Read both verdicts** — `725b1f01` (Tier 1 round 2) and `203d858b` (300). Act on a
   REVISE/REJECTED: the code is already on the shared branch.
   `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;`
   then `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
2. **Verify both changes at the artefact after the next chassis roll**, each with BOTH controls
   (§1, §2). Neither is proven in production; the only evidence today is unit-level and
   mutation-gated.
3. **Decision 5 — close `083`** at ~2026-08-25 once `444`/`458`'s doors have held a week. Move with
   **both paths on the commit** (`git mv` landmine) and verify at HEAD with `git ls-tree`.
4. **Decision 1 Tier 2 — the detector→editor routing.** The large one, and **the only thing that
   actually repairs the ~134 findings queued on owned pages.** RFC first: it is a new shared repair
   route, so architecture-scope. Nothing above touches it.
   > ⚠ **`08-18c` §2a/§2c SIZE THIS WRONG, and I corrected it after the council caught me.** Both of
   > them say *"nothing converts a detector's finding into an editor's edit … someone must build the
   > step"*. **False, asserted with no query, gated HIGH by `prior_art_librarian` (corr `725b1f01`
   > round 2).** `copy-editor` is a live agent whose `run_copy_edit` step already returns
   > `{page_component_id, slot_name, field_updates, rationale}` — **exactly** `apply_section_edit`'s
   > input — built from each component's `content_data`, rendered HTML **and declared schema**, with
   > the type-preservation and link-preservation constraints a repair route needs already written
   > (its prompt enumerates the page's required links as data *"because a prose instruction to
   > preserve a set is not reliably followed"*).
   > **What is true, narrowed: nothing routes a specific detector FINDING to that producer** — it
   > takes a page, not a finding. So the RFC is *"aim an existing producer at one finding on one
   > component, and route the five refused types there"*, not *"build a converter per defect type"*.
   > ⚠ **And check the producer before building on it:** [MEASURED 2026-08-18] `copy-editor` has
   > **2 orchestration runs in all history, 0 work items, and no scheduled task drives it** — against
   > `section-editor`'s 227. A mechanism nothing drives is the shape this estate keeps mistaking for
   > a missing one, and it is also the shape that turns out to be broken on first real use. Full
   > account in `bugs_open/301`'s correction section.
   > ####
   > ⚠ **CORRECTED AGAIN the same evening, and this changes what to DO.** The council's
   > `prior_art_librarian` seat (round 3, medium) asked for those numbers as quoted query results
   > rather than my own `[MEASURED]` tag. Re-run, **two were wrong**: I compared orchestrations to
   > work items (like for like it is 2 vs 18 orchestrations, and 0 vs 227 work items — *no work
   > items at all* is the real point, nothing dispatches it); and **`copy-editor` is not dormant,
   > it is ONE DAY OLD** — seeded `2026-08-17 11:49`, updated `2026-08-18 17:59`, both runs today,
   > **owned by the `loanandmortgagecalculator_couk` lane** (migrations `447`/`462`, commit
   > `b04493b7b` *"stage 2 BUILT and PROVEN on its proof case"*).
   > **So do NOT open Tier 2 by drafting an RFC around it.** Talk to that lane first — the estate's
   > rule against competing with an owned mechanism applies, and a design written against a
   > `field_updates` contract that changed twice in two days would be stale before it was read.
5. **Two loose ends nobody owns**, both `[UNMEASURED]`, neither chased:
   - **`page-rerender` saves to owned pages 3,754 times without refusal** while `page-build-handler`
     is refused every time. Both go through the same guard. One of those facts needs explaining, and
     nobody should conclude the guard covers every save until somebody does.
   - **`08-18c` §4's follow-on**: a page whose name and URL are `tool-…` carrying
     `rebuild_policy='generic'` looks like a data defect. Nobody has counted how many.

> ⚠ **CORRECTED 2026-08-19 — `copy-editor` is owned by the `copy_quality_two_stage` lane, NOT `loanandmortgagecalculator_couk`.** I got the wrong lane from a `grep -rl "copy-editor"` hit in LMC's `README_where_we_are.md` — a *mention* — and read it as ownership. `scripts/who-owns.py` exists to separate those two, and I did not run it. The defining evidence is what the commits shipping migrations `447`/`462` actually touch: `docs024_key_docs_latest/copy_quality_two_stage/`. Register entry **CQ-024**. A CONTRIB is filed in their lane dir (`CONTRIB_2026-08-19_from_the_277_083_lane_…`, commit `7574482c7`).


## 5. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py` **by slug** for `083`,
`300`, `301`, `307` · grep live `.jsonl` for `save_page_sections|fail_work_item|301|307` · re-measure
§0 · then §4 step 1.
