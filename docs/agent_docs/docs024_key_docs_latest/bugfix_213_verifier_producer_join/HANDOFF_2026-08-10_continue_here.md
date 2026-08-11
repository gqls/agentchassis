# HANDOFF 2026-08-10 — bugfix 213, continue here

**Cold-start for this lane.** Read this, then `NOTES_verifier_producer_join.md` for the
missteps and the mutation matrix. `PLAN_2026-08-10_verifier_producer_join.md` has the
design and why the rejected candidates were rejected.

---

## STEP 0 — DONE. Verdict read; here is what it said

**APPROVED, round 1** — 14 seats, 0 unreadable, 3 abstained, `gated_by_truncation:
false`, 4 advisory objections, none high-severity. Recorded with
`Council-Reviewed: c9c7c83f-…` on `5d482297e`, which also **acts on** the two
objections that were real: the guardian's "the shared branch is untested" (now
`TestOnlyTheOptedInVerifierCarriesAScopeTest`, mutation-proven) and the
prior_art_librarian's literal-item_type consumer sweep (clean three ways — Go,
`reviewRevalidators`, and the claim-timeout exclusion list). Full account in NOTES.

**The one follow-on worth carrying forward**, from the `architecture` seat, which
signalled `insufficient` on the class while agreeing no RFC is needed: `Grades` is
opt-in, so the NEXT converging producer on any of the other 10 verified item_types
reproduces this bug unless a human remembers to write one. It asks for a periodic
check flagging a verified item_type that accumulates rows with more than one
spec-shape/`audit_source` and no `Grades`. **Not built.** That is the closure of the
defect *class*, as opposed to this instance.

⚠ **Do not read a verdict with the CLAUDE.md `LIMIT 1` query** — it is
correlation-blind and returned another lane's REVISE. See the RUNBOOK.

<details><summary>original STEP 0 instructions, kept for the resubmit path</summary>

**Read the council verdict.** Submitted 2026-08-10, correlation
`c9c7c83f-d706-48b0-b433-55de51d88f9f`. At the time of writing it was still running
(`current_step = gate_tooling_provenance`, `EXECUTING_STEP`), so **no verdict existed
yet and none is recorded anywhere in this lane.**

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='c9c7c83f-d706-48b0-b433-55de51d88f9f' AND kind='council_report'
ORDER BY created_at;

-- still running? (a missing row is latency, not a dropped dispatch — do NOT resubmit)
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c9c7c83f-d706-48b0-b433-55de51d88f9f';

-- the human-readable note
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

- **APPROVED** → record it in the bug file and NOTES. Do **not** amend `2d151c41f` to add
  a `Council-Reviewed:` trailer — forward-only forbids it, and it is unnecessary: the
  commit carries `Council-Submitted: c9c7c83f…`, and the 098 report resolves the verdict
  at report time and credits it automatically.
- **REVISE / REJECTED** → **the code is already on the shared branch**, so this is a
  follow-up commit, not a hold. Resubmit with `RESUBMIT_CORR=c9c7c83f-d706-48b0-b433-55de51d88f9f`
  so the trail accumulates.

</details>

## What is done

| | state |
|---|---|
| Half A — `dark_section` → its own item_type `dark_section_audit` | **committed** `2d151c41f` |
| Half B — `VerifierPolicy.Grades` remit contract + gate enforcement | **committed** `2d151c41f` |
| Coverage-guard classification (both maps) | **at HEAD**, but see the passenger note below |
| Tests (join guard, over-correction guard, out_of_scope reason) | **committed**, mutation-proven |
| WII-013 register entry + index row | **committed** `3c72619fc`, `3895be34e` |
| Two LANDMINES entries + `landmines-sync.py --apply` | **committed** `3c72619fc`, synced |
| Standing five (PLAN/NOTES/RUNBOOK/README) | **committed** |
| Council | **APPROVED r1**, verdict read, all 4 advisory objections actioned or answered (`5d482297e`) |

**Build state, verified:** `go build ./...` clean and
`go test ./platform/orchestration/actions/...` green against a clean
`git archive HEAD` extraction, run after the final commit. ⚠ **Do not judge this by
building the live working tree** — it was concurrently broken by another session's
in-flight `save_page_sections_action.go` / `save_sections_decision_gate.go`, which are
nothing to do with this lane.

> **Passenger note, so `git show` does not mislead you.** `2d151c41f`'s message
> describes seven files and the commit contains **six**. My edits to
> `discovery_checks/verifier_coverage_test.go` were swept into another session's commit
> `d644723b8` ("RFC_015 round 1 revisions") before I committed — the documented
> same-file-passenger hazard. **Nothing is lost**: the content at HEAD is exactly the two
> intended edits (verified, 4 occurrences of `dark_section_audit`, correct text) and
> forward-only means it stays where it is. Find it with
> `git log -S dark_section_audit -- <path>`.

## What is NOT done — in priority order

### 1. The roll, then prove it at the artefact
The fix is **NOT in `v1.0.1283`** (that image predates the commit). After the next roll:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ NEW — expect 0 before the roll, 1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verifier_scope_mismatch'"
# ✓ POSITIVE CONTROL — live since RFC_017, must read 1 in the SAME exec
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verification_unavailable'"
```
Repeat on **every replica** — a label greps a subset (`bugs_open/153`). The control is
what proves the grep and the binary rather than your spelling.

Then the behavioural half, which no grep gives: a `hardcoded_section_colors` item whose
spec carries **no** `check` key must land `triaged`/`failed` with `error` beginning
*"completion blocked: the verifier registered for this item_type does not grade this item"*.

```sql
SELECT status, attempt_count, error, result->'_verification'
FROM site_work_items WHERE result->'_verification'->>'status'='out_of_scope'
ORDER BY updated_at DESC;
```

### 2. Migration `374`, AFTER the roll (not before)
Defensively re-type any still-open producer-B rows filed between the commit and the roll:

```sql
UPDATE site_work_items
   SET item_type='dark_section_audit',
       item_key=replace(item_key,'_hardcoded_section_colors_','_dark_section_audit_')
 WHERE item_type='hardcoded_section_colors'
   AND spec->>'audit_source' IS NOT NULL
   AND status NOT IN ('complete','failed','rejected','wont_fix');
```
**0 rows qualify today — re-measured after the verdict, still 0.** The guardian seat
raised exactly this as a state-transition side effect; the population is empty, so
**do not ship an empty migration**. Re-check after the roll and skip if still 0. Pre-roll it would be pointless
(an audit could immediately re-file old-type rows); post-roll it is idempotent and final.
A pre-roll-filed row that reaches completion post-roll is caught by Half B anyway —
blocked loudly, not silently closed. Next free number was 374; **re-check, it moves.**

### 3. Grade the 11 closed producer-B items — the part needing judgement
**A `complete` count is NOT a false-complete count.** Two verdicts already exist:
gamesdesign.co.uk confirmed **FALSE** at the served artefact (bug §3); relojistas.com
measures **CLEAN** in `bugs_open/122`'s `BASELINE_2026-08-06_render_audit.txt`. Nine
unknown. Enumerate them with the bug file's §4 query filtered to
`spec->>'audit_source'='design-audit' AND status='complete'`.

Grade each against **its own `spec.acceptance_test`** at the served artefact (browser /
computed styles — the 122 lane's render-audit machinery). **Do NOT re-run the verifier**:
it will pass again, for the same correct reason (bug §6).

For each confirmed-unrepaired, **insert a fresh row** (`item_type='dark_section_audit'`,
`status='detected'`, spec copied plus `reopened_from=<old id>` and
`reopen_reason='bugs_open/213 false complete'`, item_key recomputed under the new type so
dedup and two-strike apply). **Leave the historical rows' status alone** — `complete` is
the honest record of what the machine did.

**Acceptance measure for the whole fix:** the 11-vs-0 asymmetry must become capable of
disappearing — after one post-roll audit cycle, a producer-B item must be able to reach
`unresolved`/`failed`/`detected`. Until one does, the fix is deployed but unexercised,
and those are different claims.

### 4. Two things deliberately left open, on the record
- **`designRouting["dark_section"]` still points at `color-variable-fixer`**, whose
  transform provably cannot touch producer B's typical defect (an already-`var()`
  fallback — that is *why* gamesdesign's item passed). That is a routing decision for the
  design-audit route owner, not this bug. Named in WII-013 so it is not lost.
- **`spec.acceptance_test` is still read by nothing** on the completion path (grepped:
  zero consumers outside the `improve_tool` family). The candidate verifier for
  `dark_section_audit` is `criteria_check` (RFC_002) over that field. That is the
  follow-on that would turn a declared gap into real verification.

## Traps that cost me time today

1. **Every ownership check reads COMMITS.** `who-owns.py`, `git log`, "no workstream dir
   exists", and the session-start `git status` snapshot all share one blindness. Run
   `git status --short <the bug's cited paths>` **first**. An untracked new file beside a
   bug's paths is the strongest ownership signal there is. This cost ~1h on `bugs_open/214`
   (full account in `WRONG_CALLS.md`, distilled into LANDMINES).
2. **Mutate one half at a time.** Reverting Half A alone left the guard GREEN, because
   Half B independently covers the route. Had I mutated both at once and stopped, I would
   have claimed "the test guards the fix" when it guards *at least one half of it*.
3. **`git archive HEAD` must be given the repo** — `git -C <repo> archive`, not a bare
   `git archive` after `cd`-ing into the scratchpad, which silently leaves you with an
   almost-empty directory that then "builds fine".
4. **Council schema**: `.plan.summary` is required and is a different field from
   `rationale`; `operation` must be `modify|add|remove|config_change` (`create` is
   rejected — a new file is `add`). The `commit-msg` trailer gate correctly blocks
   `Council-Submitted: pending` — submit first, commit with the real correlation.
5. **`site_plan_pages` vs `site_plan_sections`** (from the 214 detour, but general): a
   page with no sections looks like a missing page. Resolve page existence against the
   table the *consumer* reads.

## The 214 debt I incurred and discharged

I investigated `bugs_open/214` for an hour before discovering its owner mid-fix, and stood
down. I contributed the measurements rather than competing — appended to that bug file
2026-08-10. **If anyone picks 214 up: the filed census is section-scope only and
understates it roughly fourfold.** Page scope has 28 unresolvable refs of 162, its
consumer join is `scope_ref = $page` *exactly* (not the section join's tolerant `LIKE`),
19 of 22 current-plan orphans have active generated assets, and gamesdesign.co.uk's about
page serves two `<img src="/assets/images/hero.jpg">` that **404** while the commissioned
`hero-about.jpg` sits deployed at 202,259 B. That is not this lane's work — do not adopt
it here.
