# HANDOFF — 2026-08-18, fresh chat starts here: RFC_029 Phase 1 is **COMPLETE and both promises discharged** (417 PROVEN 26/26; the window caught `bugs_open/287`, whose fix took `field=result` to zero). Triage pass 2 is **ANSWERED** and hands us one buildable change. That change is the next work.

**Supersedes `HANDOFF_2026-08-16_continue_here.md`** (which superseded 15c). Nothing in this file
needs re-deriving; every figure below was measured 2026-08-17/18 and says where.
⚠ **Not to be confused with `HANDOFF_2026-08-16_gaswholesalers_tool_page_and_stray_logo.md`** —
that is a different sub-thread of this lane, another session's, untouched by any of this.

**Read in this order:** this file → RFC_029 **§10.10** (pass 2's answer — the next task's spec) →
§10.8/§10.9 (what is proven, and the figure I got wrong) → NOTES `## 2026-08-18`. The milestone
read-out for a human is `SUMMARY_2026-08-18_the_instrument_caught_its_first_real_bug.md`.

---

## 1. What is true now (all measured; nothing inferred in this section)

- **The recorder is live and has been since 2026-08-16 10:41Z.** Current build **v1.0.1308**,
  label revision `e7e5e4d53` **confirmed at `/proc/1/exe`** with `deadbeef1234…` absent as a
  negative control, digests matched. `53edef286` (the recorder commit) is an ancestor.
- **Council: APPROVED** (round 2, run `b5678c3a`, corr `75091072-9d65-433e-8a30-84719dc3f30f`).
  Commit `53edef286` carries `Council-Submitted:` and is credited automatically. Nothing owed.
- **Migration 417 is APPLIED (owner, 08-16 15:58:18Z) and PROVEN (08-18):** 26 image-build-handler
  runs, **26/26** asset-deployer children carry a bare `asset_id`, **0** carry a literal
  `asset_id!` key, **0** strict errors fleet-wide. Refusal branch still un-exercised — expected.
- **`bugs_open/287` (the `spawn_record` slug) is FIXED and we graded it.** Their WFA-017 +
  `!` flip took **`field=result` to zero**, holding across **193 loop runs / 12 h**. Our evidence
  is filed as 287 §10, §10a, §11d, §11e. **Do not reopen it.**
- **CTS-060's `verify-later` is discharged on both halves** (register updated).

## 2. THE NEXT WORK — one change, already specified (RFC_029 §10.10)

Pass 2 asked: every `work_item_id`/`current_page` reference in the dispatch loop is explicitly
mapped, so why do ~7 rows per run still hit the whole-tree search? **Answered at the code:**

```go
// action_inputs.go ~690-780
allFields := spec.Required + spec.Optional          // never pruned
// Strategy 0 resolves explicit dotted paths into result.Values
extracted := ExtractFields(collectedData, withoutStrict(allFields), logger)   // searches ALL of them
for k, v := range extracted {
    if _, alreadyResolved := result.Values[k]; alreadyResolved { continue }   // discarded AFTER the fact
```

**An explicit mapping never prevents the search — it only discards its answer. Only `!`
(`withoutStrict`) removes a field from the search at all.**

**THE TASK: prune the field list by what Strategy 0 already resolved, before calling
`ExtractFields`** (both the Strategy 1 `fieldNames` path and the Strategy 2 `allFields` path).

- **Why it is behaviour-equivalent** (state this in the submission, it is the crux): the merge
  discards exactly those keys today, and `ExtractFields`' only effect on a field nobody asked for
  is core-field recovery, which covers `domain`/`objective`/`model` **only**
  (`unified_extractor.go:387`) — none of them in this population. Strict fields are already
  excluded.
- **What it buys:** the bulk of the ~1,600 rows/12 h disappear; the platform stops doing a
  whole-tree search per already-resolved field per step on a hot path; and **every surviving row
  then means something** — a field the platform genuinely could not resolve explicitly, which is
  the real Phase 2 population.
- **Tests:** a field resolved by Strategy 0 produces NO conflict finding (use the fake recorder
  from `resolver_findings_test.go`); an unresolved field still searches and still records;
  strict-field behaviour unchanged; arm budgets unmoved (10/15 outer, 5/8 inner).
  Mutation-prove the prune (remove it → the first test fails).
- **Process:** platform code on a shared seam → **council gate before or alongside the commit**
  (`097`, one run, `RESUBMIT_CORR` not needed — new task). Argue it as an RFC_029 amendment, not
  a new RFC: it changes no guarantee, it makes the existing "explicit beats search" rule actually
  skip the search. Build/test from `git archive HEAD` + your files — the tree does not compile.

## 3. Then, in order

1. **After the fix rolls, re-read the window** (RUNBOOK, "RFC_029 observation window"). Whatever
   survives IS Phase 2's precondition list — and it will be a fraction of today's.
2. **Triage what survives, pair by pair.** For each: is the winner the value the step needs? Yes →
   mark it `!` (**not** "write an explicit mapping" — see §2; for an already-mapped field the
   mapping is already there and does nothing). No → that pipeline was living on the search.
   Today's population, for reference (6 h, 142 loop runs): `build-dispatch-loop` `current_page`
   656 / `work_item_id` 299; `page-content-writer` `current_page` 58; `page-rerender` 39;
   `page-build-handler` `current_page` 8 / `page_type` 3.
3. **Phase 2** (conflicts resolve NOTHING) — its own council-gated task, gated on step 2 reaching
   zero or fully marked, **not on a date**. It is now much less risky than §10.5 feared: for any
   field Strategy 0 resolved, the search's answer is already discarded, so Phase 2 changes nothing
   for it. Flip sites are marked in code and in `unified_extractor_search_test.go`'s header.

## 4. Traps and corrections a fresh session must not re-learn

- ⚠ **A figure I published was WRONG; do not re-quote it.** "−73%, 3.4 rows per run" (RFC_029
  §10.7, 287 §11d) came from an **11-run, 1.3-hour** sample. Matched-window on 193 runs:
  **17.7 → 8.4 per run, −53%.** Corrected at §10.9 and 287 §11e; `WRONG_CALLS.md` entry.
  **Write the denominator inline with any per-unit figure, or wait for it.**
- ⚠ **"A fresh build was deployed" is not evidence** — MEMORY `a-fresh-deploy-can-ship-no-new-code`
  (a same-tag rebuild shipped 203 commits of nothing on 08-17). **Ask the image label first:**
  `docker inspect aqls/<svc>:<tag> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`,
  then confirm at `/proc/1/exe`. Never guess candidate shas (cost ~22 execs on 08-17). The
  negative control must be capable of being absent — 40 zeros match every binary; use a fake sha.
- ⚠ **Never arm `!` on a field whose reference does not yet resolve** — it hard-fails the step.
  287's `mark_complete` was safe only because the suffix fix landed first. Order is load-bearing.
- ⚠ **A `*_TRIGGER_*` script publishes on EVERY run** — re-running `097` to re-read its output
  cost a duplicate council round. `| tee` the first run. (LANDMINES + WRONG_CALLS.)
- ⚠ **A replayed migration takes a SECOND snapshot** whose reason says `pre-update` but whose
  content is post-change — so "the latest snapshot" is the wrong pre-image. Predicate on the
  CONTENT. (LANDMINES; 417's header carries the corrected check.)
- **`kubectl exec -i` with nothing piped hangs.** The working tree does not compile (other lanes).
- **Rows carry no `orchestration_id`/`step_name` by design** (pod-level attribution;
  `context.identity_scope` says so). Do not file that as a bug. If step attribution is ever
  needed, §10.7a(d) notes the action name IS cheaply reachable at the bypass site.

## 5. Not ours, but nobody appears to own it

While proving 417: **14 of the 26 asset-deployer children FAILED** — `failed to get latest
commit/base tree for branch "master"`, a git-adapter/repository error well after input
resolution. They received their `asset_id` correctly; this is not RFC_029's. But that is a **54%
failure rate on asset deploys** with no visible owner. Worth a pointer to whoever owns the deploy
path, or a `090`. Recorded here and in RFC_029 §10.8 so a "26/26 proven" headline never buries it.

## 6. Session-start checklist

1. `git log --oneline -10`; re-read this file from disk (it goes stale in hours here).
2. Verify the running build by LABEL first (§4), then read the window (RUNBOOK).
3. Build §2's change; council-gate it; commit per task with an explicit pathspec and a
   `Council-Submitted:` trailer if the verdict has not landed.
4. Everything else in this lane is closed. RFC_029 §10.4–§10.10 is the implementation record;
   NOTES holds the missteps; the 08-18 SUMMARY is the human read-out.
