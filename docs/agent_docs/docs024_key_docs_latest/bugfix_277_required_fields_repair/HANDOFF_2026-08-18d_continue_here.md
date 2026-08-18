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
| decision 1 Tier 1 — refusal status | **BUILT + COMMITTED `6aee22b00`; config LIVE (mig `480`); Go INERT until a roll.** Council corr `725b1f01`: round 1 **REVISE**, round 2 resubmitted | §1 |
| decision 4 — `bugs_open/300` | **BUILT + COMMITTED `42a4bf441`; INERT until a roll.** Council corr `203d858b`, verdict pending | §2 |
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

### What is owed

**Post-roll verification, with BOTH controls** (query in `RUNBOOK`): an owned-page refusal must land
`wont_fix` carrying `result ? 'owned_page_refusal'`, **and** a genuine save failure on the same
handler must still land `failed` without the stamp. A zero on the control means no genuine failures
happened in the window, **not** that the split works — widen the window.

**Read the round-2 verdict** (`725b1f01`) and act on it. The commit carries `Council-Submitted:`;
`098` credits it automatically if it turns approved. Do **not** write `Council-Reviewed:` on an
unread verdict.

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
   route, so architecture-scope. Shape in `08-18c` §2c. Nothing above touches it.
5. **Two loose ends nobody owns**, both `[UNMEASURED]`, neither chased:
   - **`page-rerender` saves to owned pages 3,754 times without refusal** while `page-build-handler`
     is refused every time. Both go through the same guard. One of those facts needs explaining, and
     nobody should conclude the guard covers every save until somebody does.
   - **`08-18c` §4's follow-on**: a page whose name and URL are `tool-…` carrying
     `rebuild_policy='generic'` looks like a data defect. Nobody has counted how many.

## 5. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py` **by slug** for `083`,
`300`, `301`, `307` · grep live `.jsonl` for `save_page_sections|fail_work_item|301|307` · re-measure
§0 · then §4 step 1.
