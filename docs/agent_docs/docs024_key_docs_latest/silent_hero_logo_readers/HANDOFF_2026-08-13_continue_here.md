# HANDOFF 2026-08-13 — continue here (silent_hero_logo_readers)

> **SUPERSEDED 2026-08-15 by `HANDOFF_2026-08-15_continue_here.md`** — 273 is CLOSED (live on
> `v1.0.1300`); this file remains the record of the 269 close and the 08-14 state updates.

**Written at the owner's request so this lane survives a fresh chat.** Supersedes
`HANDOFF_2026-08-12b_continue_here.md` for state; that file is still the record of how 261 was found.

**Read in this order:** this file → `bugs_open/269` (the live work) → `architecture_review/RFC_027`
(the open owner decision). Everything else is history.

> **STATE UPDATE 2026-08-14 — §1's two action rows are DONE; do not chase them.**
> - **269 is CLOSED** — the roll happened (chassis `v1.0.1297`, 2026-08-13T22:29Z), §9 verification
>   ran with controls, and the file moved: `bugs_closed/269_…` §11 holds the live proof and the one
>   caveat (collision halves test-proven only, no §6b file scoped yet). Owner approved the split.
> - **236 (hero/logo): the MECHANISM IS CONFIRMED** — the narrow one-function run
>   (`23f1cf9a…`, dispatched 08-14 after finding the 08-12 "cheapest next move" had NEVER run)
>   read the full body: **the park copies only AwaitedRequests/Status/LastActivity; the
>   `[CONTESTED]` fragment was an existence check, not a merge.** Occurrence witnessed live on the
>   run's own two parked children (awaited step's key ABSENT, earlier steps' present); ordering
>   verified at `coordinator.go:1795`/`:1839`. Full record = the 236 file's final contribution.
>   **236 stays OPEN: the fix is RFC_012 `(a)`/`(a′)` — an owner decision, not a patch.**
> - **Still open:** `RFC_027` owner ruling; `bugs_closed/261` §8 follow-ups 2/3; 267 §4b trend
>   re-read when bundle traffic is more than a handful.
>
> **SECOND UPDATE 2026-08-14 ~14:00Z — v1.0.1298 rolled (build point `bc39e7bf5`, probed both
> replicas with control); nothing of this lane's was gated on it. Re-read on 10 bundles: 0 bare
> handles, 0 new cap_only, collision witness still outstanding. THE NEXT WORK for a fresh session:
> `bugs_closed/261` §8 follow-up 2** — the ~10-signature per-file sibling cap
> (`siblingSignatures`, `diagnose_assemble_bundle_action.go`) hid the three functions a real run
> needed; a recorded harm case exists in 261 §8.2. Read 261 §8 + the LANDMINES entries on that
> footprint before touching it. Follow-up 3 (`knownScopeIdentities` omits `values`,
> `diagnose_route_action.go:541`, cosmetic) and follow-up 4 (the owed precedent check) queue behind
> it. Owner decisions RFC_012 (236's fix) and RFC_027 remain open — surface, don't work.
>
> **THIRD UPDATE 2026-08-14 (evening) — §8 follow-up 2 is DONE IN TREE: `bugs_open/273`.**
> Fix + 5 tests (mutation-proven) in `siblingSignatures`/`writeDeadEndTail`; council corr
> `ba3f6047-a2e5-4ce6-ac0e-edf0bb88c4e3` — **verdict READ: APPROVED first round, 12 min,
> 3 advisory objections, all discharged in 273 §8.** The one substantive recommendation
> (bug_historian: bound the tails' AGGREGATE, not just per-file — the 062 shape) was implemented
> same day: `siblingDeadEndTailTotalCap = 12000`, own test, own mutation proof. 273 needs a
> chassis roll, then its §5 verification (demand control: a zero proves nothing unless a
> dead-end file was scoped). Follow-up 4 was already discharged by 261 §9b. Remaining for a
> fresh session: 273's live proof after a roll; follow-up 3 (cosmetic); the RFC_012/RFC_027
> owner decisions — surface, don't work.

---

## 1. Where this lane is, in one screen

| item | state | needs |
|---|---|---|
| `bugs_closed/261` — code tier could not read its own index's spellings | **CLOSED, live** | nothing |
| `bugs_closed/267` — bundle advised an impossible whole-file re-read | **CLOSED, live on `v1.0.1295`**, council APPROVED `ac23f2f7` | nothing, except the §4b trend below when traffic resumes |
| `bugs_open/269` — sibling section offered bare method handles → **wrong body, silently** | **FIXED IN TREE, NOT LIVE. Council APPROVED first round** `e5809ca9-d718-44f6-8d27-6d8cd656dd28` (13/15, 2 advisory, both answered in its §8) | ① next chassis roll ② verify per its §9 — **needs a COLLISION file or it proves nothing** ③ then close |
| `architecture_review/RFC_027` — the handle grammar has no owner | **OPEN, needs an OWNER RULING** | a decision; "four bugs on something this small is acceptable" is a legitimate one |
| `bugs_open/236` — the case all of this was unblocking | **UNTOUCHED by this lane today** | someone to actually re-run the diagnosis now the tooling works |

**The one thing to do first if you pick this up cold:** nothing on the council side — 269 was **approved
first round** and both advisory objections are answered in its §8. **The next action is a chassis roll**,
then 269's §9 verification, then close it. If you want new work instead, go to §6: `bugs_open/236` has
never been re-run and both of its blockers are now live.

## 2. What shipped, and how to prove it rather than assume it

Live on chassis **`v1.0.1295`**, pods started `2026-08-13T13:53:19Z`. Commits, all pathspec-scoped:

```
139dcc3ca fix(267)              aade2842e landmine(267)        24f83ba90 revise round 2
17734b699 revise round 3        978a62be2 269 record           34cf44e2b 267 APPROVED
34e2d8e97 close(267) → bugs_closed/   6ceeaba1b landmine precision   a3fee59b8 fix(269)
957d083dd docs
```

**Proving a deploy on this service — the exact recipe, because I got it wrong twice getting here:**

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system logs $POD --tail=100000 | head -1          # PRECHECK: a startup line?
kubectl -n ai-persona-system logs $POD --tail=100000 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <the git_commit from the stamp> && echo IN
```

- **The precheck is not optional.** The stamp is a startup line and this service's log rotates; if
  `head -1` is not a startup line you are out of range and a miss means nothing. Worked at **7 minutes**
  after a roll; the landmine measured **0 hits at 44 minutes**.
- **The stamp is ONE commit — the build point — never yours.** Obtain it, then test ancestry. Grepping
  the binary for *your* sha returns absent on a binary that genuinely contains your change.
- **Always run a control**: a commit that must read ABSENT (any descendant of the stamp).

## 3. 269 — what to do next, in order

**① The verdict is read and recorded — APPROVED, no action owed.** Correlation
`e5809ca9-d718-44f6-8d27-6d8cd656dd28`, 13 seats approve, 2 advisory objections (`bug_historian`,
`prior_art_librarian`, both medium), answered in `bugs_open/269` §8a/§8b. The queries below are only if
you want to re-read it.

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'e5809ca9-d718-44f6-8d27-6d8cd656dd28'
 ORDER BY created_at DESC LIMIT 1;   -- want COMPLETED
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id = 'e5809ca9-d718-44f6-8d27-6d8cd656dd28' AND kind='council_report'
 ORDER BY created_at;
```
Submission file to revise from: `<scratchpad>/269_submission.json`; resubmit with
`RESUBMIT_CORR=e5809ca9-d718-44f6-8d27-6d8cd656dd28`.

⚠ **Filter by YOUR correlation.** `ORDER BY created_at DESC LIMIT 1` on `doc_notes` gets another lane's
verdict — there is a landmine for exactly that.

**② After the next roll, close it — the full recipe is `bugs_open/269` §9**, which carries the live-code
check (stamp + ancestry + precheck + control) and the behaviour check with its demand control. Two things
from it that are easy to skip and fatal to skip:
- **`bare_method_handles = 0` is evidence only if `bundles_in_window > 0`.** `bugs_closed/267` §9 records
  that exact zero being unreadable without its control, on 2026-08-13.
- **the bundle must have scoped a COLLISION file** or nothing was exercised. §6b names them;
  `discovery_checks/check_integrity.go` is the richest at six-way. A clean result over a collision-free
  file demonstrates nothing — which is exactly how this defect survived `bugs_closed/261`'s fix.

**③ `git mv` needs BOTH paths on the commit** — `git commit <old> <new>`. A single-path pathspec ships a
copy and leaves the original at HEAD. Verify with `git ls-tree -r --name-only HEAD | grep <slug>` → one line.

## 4. What 269's numbers mean, and the sentence not to round up

Measured 2026-08-13, **control first**: of `code_symbols` `kind='method'` rows, **1,175 of 1,175** store
the parenthesised `(Recv).Name` spelling and **0** do not — so the bare-name strip in the query below
does real work. A no-op strip would have produced a plausible wrong answer.

**17 collision groups · 48 of 1,175 methods · 4.1%**

```sql
SELECT count(*) AS groups, sum(n) AS methods FROM (
  SELECT count(*) AS n FROM code_symbols WHERE kind='method'
   GROUP BY repo, path, regexp_replace(symbol,'^\(.*\)\.','') HAVING count(*) > 1) x;
```

⚠ **4.1% is the FLOOR of the harm, not its rate.** In an n-way group a bare handle resolves to the first
and is wrong for the other n−1. Worst groups are **six-way** (`discovery_checks/check_integrity.go`,
`Name` and `Run`) → wrong 5 times in 6. **`pkg/diagnose/loop.go` is itself in the list**
(`(Outcome).String` / `(Tier).String`), so a diagnosis *of the diagnosis loop* — which 261 and 267 both
were — could have been handed the wrong body.

⚠ **Do NOT say "48 wrong bodies were served."** That is the population where the defect can fire, plus
the per-group odds. **Incidence is unmeasured** and needs a scan of `diagnosis_artifacts.body` for
bare-handle lines. Nothing in this lane's files claims otherwise; keep it that way.

## 5. The traps this lane paid for — read before touching anything here

1. **A markdown document handed to the shell is CODE.** An unquoted heredoc (`<<PY`) executed every
   backtick in a document body. Use `<<'PY'` and pass paths via the environment
   (`OLD=… python3 script.py`). Same family: a commit message mentioning a trailer name in prose tripped
   the trailer gate, reading the sentence as a trailer with value `.`.
2. **`git diff | grep '^-[^-]'` cannot see a removed markdown bullet.** Gate on
   `git diff --numstat <file>` first, then read removals with `grep '^-' | grep -v '^---'`.
3. **`LANDMINES.md` is append-only and shared.** Correct in place with a dated note; if you amend an
   `added:` line, carry its old content verbatim inside the new one, and expect the pattern check to flag
   the removal (it cannot tell an amendment from a deletion). Run `./scripts/landmines-sync.py --apply`
   after, `--check` to confirm.
4. **A council `REVISE` gated on your SKETCHES is not a code problem.** `editquality` raised HIGH twice
   on lines I had written but elided from the submission. Show disputed hunks verbatim; "a test asserts
   it" is not an answer to "show me the edit".
5. **Put `Council-Submitted:` on the commit even when you are about to submit seconds later.** The 269
   fix commit `a3fee59b8` lists as UNREVIEWED in the `098` coverage report for ever: I committed, then
   built the submission, so there was no correlation to write down. Forward-only forbids an amend and a
   cosmetic follow-up touch just to attach a trailer would be gaming the report. The code IS approved
   (`e5809ca9-…`, `bugs_open/269` §8d) — the report just cannot join it. The 267 commits all carry it and
   are all credited.
6. **A correction you offer a reviewer is a durable claim** and inherits none of the checking you did on
   your code. Both of my provenance claims were one measurement of a time-dependent check generalised
   into a property of the service — see `WRONG_CALLS.md` 2026-08-12, and its 08-13 narrowing.

## 6. What I did NOT do, and would do next

- **`bugs_open/236` has not been re-run.** It is the case this whole chain was unblocking: it failed
  twice for want of function bodies that 261 restored and iterations that 267 stopped wasting. **Both are
  now live.** This is the highest-value next move and it is not blocked by 269.
- **`bugs_closed/261` §8 follow-ups 2 and 3 are still open** — the per-file sibling cap of ~10 signatures
  (which hid the three functions a real run needed), and `knownScopeIdentities` omitting `values`.
  Follow-up 2 is the next thing to bite this same code path.
- **`RFC_027` needs an owner ruling**, not more work. It asks whether the `path:Symbol` handle grammar
  deserves one authoritative implementation after four bugs (`189`, `261`, `267`, `269`).
  `analysis.CanonicalSymbolName` is the first instalment; `code_symbols_actions.go:598` is the one
  remaining independent producer.
- **267 §4b's trend is unfinished, deliberately.** After traffic resumes, `cap_only` should stop growing
  from `2026-08-13 13:53:19+00` onward. **It must NOT go to zero** — the 6 historical iterations stay in
  the table for ever, so a zero means the query is wrong, not the bug fixed.

## 7. Files this lane owns

```
bugs_closed/261_…code_tier_cannot_read_the_symbol_spellings_its_own_index_produces.md
bugs_closed/267_…bundle_advises_a_whole_file_reread_that_can_never_fit….md   (§9 = live proof)
bugs_open/269_…sibling_signatures_render_methods_bare….md                    (§6b = measurement)
architecture_review/RFC_027_the_symbol_handle_grammar_has_no_owner….md
docs026_concept_register/register/diagnosis-loop.md                          (DIAG-043)
platform/orchestration/actions/diagnose_assemble_bundle_action.go
platform/orchestration/actions/diagnose_assemble_overcap_advice_test.go
platform/orchestration/actions/diagnose_assemble_sibling_spelling_test.go
internal/analysis/symbolbody.go                                             (SymbolSizes, FindFile, CanonicalSymbolName)
```
