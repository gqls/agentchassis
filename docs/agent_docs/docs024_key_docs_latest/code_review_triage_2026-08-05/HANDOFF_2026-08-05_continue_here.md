# CONTINUE HERE — code-review triage lane, 2026-08-05 evening

**State: the work is DONE, LIVE and pod-verified. All three council verdicts APPROVED. What
remains is two dated re-checks and three items that are deliberately somebody else's.**

> **UPDATED 2026-08-06 morning.** Re-verified live on the **new** image `v1.0.1256` (§1).
> **2b is SETTLED** (§2b — by config census; a log grep cannot answer it). **2a is still due
> ~20:45Z today**, but its denominator has been taken early and its interpretation sharpened:
> the only caller that ran is `page-rerender`, 23 runs, all COMPLETED — the exact caller the
> stop condition names. **F13 re-checked: still another session's WIP, still not actionable.**
> Read §2a's correction block before running anything.

Read `NOTES_code_review_triage.md` §11–13 first if you only read one thing — §11 and §13 are
where a successor is most likely to repeat a mistake. **Then §14**, which corrects §13's
traffic check and records the evidence that would otherwise have expired by 20:45Z.

---

## 1. What shipped, and the proof

Nine of fifteen findings fixed across five commits, live on **`v1.0.1254`**:

| commit | findings |
|---|---|
| `79c713bff` | F15 — four root Go binaries added to `.gitignore` |
| `f887ed1ad` | F1, F2, F8, F14 — 195-lane cluster (all behaviourally inert) |
| `fa30062cc` | F7, F9, F3, F10 — 194-lane cluster |
| `6e607da1e` | F11, F12 |
| `1c6a3cab6` | corrections: the 3-of-6 census error, plus rename drift in LANDMINES + register |

> **RE-VERIFIED 2026-08-06 08:0xZ — the pod names below are DEAD.** The chassis rolled again
> overnight to **`v1.0.1256`** (pods `agent-chassis-7d4d7b9669-2r8f2` / `-6f2ps`, restarted
> 07:24Z) for reasons unrelated to this lane. `bugs_open/153` says a roll is not evidence your
> fix shipped; the converse also holds — **a later roll retires your proof's pod names.**
> Re-probed on both new replicas: `incoming_sections_with_content_data` → **1 / 1**,
> misspelled control → **0 / 0**. The fix is still live; the §12 gap (no removed-unique-literal
> negative) is unchanged. `RUNBOOK` R12.

**Pod-verified on both replicas** (`agent-chassis-d69d4467c-dvn8k` / `-fc8pq`), per
`bugs_open/153` — a roll is not evidence, the image carries no provenance:

```
refuse_save_without_sections_metadata    -> 1  / 1      (F7's rename)
incoming_sections_with_content_data      -> 1  / 1      (6e607da1e, the LAST code commit)
require_save_without_sections_metadata   -> 0  / 0      (wrong spelling; the probe can return 0)
```

**Read the caveat in NOTES §12 before citing this as a full verification.** There is no
removed-unique-literal negative available for this change set, so the greps prove the new code
is present, NOT that the old code is gone. The reasons are stated there and they are structural
(comments and Go identifiers are not string literals; the one removed SQL fragment has 51 other
occurrences). One `kubectl exec` per grep — batching times out on a binary this size, and the
image has no `strings`.

**Council:** `128d4fd1` (195 cluster), `cb575682` (194 cluster), `d0d2c97a` (F11/F12) — **all
approved**. Commits carry `Council-Submitted:` trailers, which `098` credits automatically now
the verdicts have landed; no amend is needed or permitted. The one substantive objection (two
reviewers: "'zero callers' is a session-side grep, not a Go-tooling proof") was answered by
renaming both functions and rebuilding — `RUNBOOK` R3.

## 2. TWO DATED RE-CHECKS — these are the actual outstanding work

### 2a. 2026-08-06 ~20:45Z — the `CONTENT_DATA_REGRESSION` read-out (24h post-roll)

This is the one that matters, because **F9 changed the predicate the check tests** and the
register carries a stop condition on it.

```sql
SELECT agent_type, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_code = 'CONTENT_DATA_REGRESSION' GROUP BY 1 ORDER BY 2 DESC;
-- occurred_at, NOT created_at
```

Baseline at the roll: **0 rows in all history.** Interpretation decided in advance (NOTES §13),
so read it against this and not against what you would like it to say:

- `page-build-handler` rows only → the widening is catching the non-deployed states F9 was
  about. Intended.
- **any `page-rerender` row → `PBP-031`'s stop condition fires: the report's predicate is
  misconceived and the per-caller opt-in MUST NOT proceed.** First question to settle: is it a
  genuine non-deployed loss, or is the F9 widening over-firing? Both are possible now and they
  were not before.
- still zero → deployed-but-unexercised. **Record that, do not round it up to "verified".** The
  report has never fired in any version, so zero discriminates nothing.

~~Also worth pairing: `... WHERE action='save_page_sections' AND occurred_at > <roll>` to confirm
the path ran at all. At the roll it had not (0 rows in 25 min while the fleet did 128), so a
zero could have meant "no traffic" rather than "no regression".~~

> **CORRECTED 2026-08-06 — that pairing check cannot work, and the denominator is already
> taken.** `agent_error_log` records **errors**; a successful `save_page_sections` writes no
> row, so 0 there means "no errors", never "no traffic". **Do not re-run it as a traffic
> check.** The working denominator was measured this morning (NOTES §14b/§14c, `RUNBOOK` R10),
> and deliberately taken early because `orchestration_states` reaps terminal rows at ~24h —
> the evidence of *which caller ran* ages out at ~20:54Z today, i.e. as you read this:
>
> - **`page_components`, roll → 02:21Z: 35 rows inserted, 10 distinct pages, 35 of 35 carrying
>   `content_data`.** The path ran, and every incoming save carried structured content — so
>   silence is correct for a demonstrable reason, not for want of traffic.
> - **The only caller that ran is `page-rerender`** — 23 orchestrations, all `COMPLETED`,
>   20:54:05Z → 02:21:07Z, fingerprinted by `rerender_sections.sections_metadata`. That is
>   *the exact caller this section's stop condition is about*, so tonight's read lands on the
>   case the rule was written for.
> - **Positive control passes at HEAD `61df92ff0`**: `TestShouldReportContentDataLoss` case 1
>   asserts the 194 signature returns `true`. The predicate can fire. (Scope: the decision
>   function only — not predicate → INSERT → `agent_error_log` end to end.)
>
> **So the "still zero" branch below is now stronger than it was written.** It is no longer
> "deployed-but-unexercised": it is *23 page-rerender runs with no regression, on a predicate
> proven able to fire*. Still short of end-to-end proof — say that, and do not round it up.
> Known limit: only the last save per page survives in `page_components` (23 runs → 10 pages),
> so ~13 intermediate saves left no trace and the detector is their only witness.

### 2b. The F3 warning has never fired, by design — **DONE 2026-08-06, and by config not logs**

`grep` the pod logs (or `agent_error_log` is not where it goes — it is a `zap.Warn`) for
`caller names html_field but declares nothing about sections_metadata`. Expected: **nothing,
ever**, until someone adds a `save_page_sections` step with `html_field` and neither metadata
key. If it does fire, that caller is one config line from silently taking another step's reply
as its own content — see NOTES and the comment at the warning site.

> **SETTLED 2026-08-06.** Both pods grep **0** — but they had restarted 35 minutes earlier, so
> that grep is nearly worthless as evidence and **a log grep is the wrong instrument for this
> check**: the property is a config invariant, and config cannot age out. Re-measured with the
> nested walk (`RUNBOOK` R11): all **six** callers still explicit, zero would-warn rows. The
> warning is structurally unreachable until a seventh caller is added, so this needs no
> recurring check — only a re-run if someone adds a `save_page_sections` step.

## 3. Three items that are NOT mine to close

- **F13 — confirmed, left for its owner.** `check_endpoint_health_action.go`'s comment cites
  `:215`/`:216`; the real locations are `markTaskComplete` at **:235** and the
  `config["task_name"]` read at **:237** (the comment's own 21-line insertion shifted them:
  `216 + 21 = 237`). **Not fixed because that file carried 25/9 lines of another session's
  UNCOMMITTED work and the comment is on their `+` side.** A pathspec cannot exclude a same-file
  passenger. If that work has since landed, this is a two-number edit — check
  `git diff --numstat` on the file first (`RUNBOOK` R9).
  **Re-checked 2026-08-06 08:2xZ: STILL NOT ACTIONABLE.** `git diff --numstat` reports
  **25/4** (was 25/9 — so that session is still actively editing, not finished). Last commit on
  the file remains `bcb6afbe8` (2026-03-25). Leave it; re-check the numstat before touching.
- **F6 — unactionable, and it is a gap in the original triage.** Named only in the handoff's
  ownership table, described in none of its verdict sections, and the raw `/code-review` output
  was never saved (this directory held only the handoff). There is no record of what F6 claimed.
  Do not guess. If the reviewer can be re-run, that is the only recovery.
- **`LogAgentError`'s `domain` asymmetry — real, fleet-wide, nobody's.** The INSERT writes
  `domain` as a bare `$2` while `site_id` gets `NULLIF($1,'')::uuid`, so an unset domain is
  stored as `''`, not NULL. Live: 4,155 rows with a real domain, **3,189 with `''`**, 122 NULL —
  so `WHERE domain IS NULL` under-reports "no domain" by ~26×. Found while measuring F5;
  outside what any finding asked; NOT fixed because changing a shared writer's stored shape is a
  seam change needing its own review. Worth a lane if anyone is counting rows by domain.

## 4. Traps this lane hit — do not re-pay for these

1. **Ownership expires in minutes.** The triage's premise ("every cluster belongs to an active
   lane") was true at 11:02 and false at 11:03. Re-read `ls bugs_open/ bugs_closed/` and grep
   the lane dir; do not trust a table.
2. **Blame the LINES a finding cites, not the file.** `git log -1 -- <file>` mis-assigned F11
   and F12 to the open 156 lane; `git blame -L` puts them in the closed 194 lane. Two lanes
   touched that file 26 minutes apart.
3. **`jsonb_each` over `{workflow,steps}` finds 3 of 6 `save_page_sections` callers.** The step
   is nested in a loop `sub_workflow` for the rest. Use
   `jsonb_path_query(default_config, '$.**.steps')`. `LANDMINES.md` documents this under a
   footprint naming the very keys I was censusing, and I ran the wrong query anyway — **grep
   LANDMINES for your symbol BEFORE the census**, because the SessionStart hook only matches
   footprints against files already dirty. Full account: `WRONG_CALLS.md`.
4. **A retention figure that could not have come out otherwise is not evidence.** "Oldest row is
   30 days old" is produced identically by a working 30-day reaper and by none at all. Test the
   boundary (`RUNBOOK` R6), and grep for the reaper in the language it is written in — this one
   is SQL in a `scheduled_tasks.pre_query` column, invisible to `--include=*.go`.
5. **`count(domain)` counts the empty string.** See §3.
6. **Read a council verdict BY YOUR CORRELATION.** `doc_notes ... ORDER BY created_at DESC
   LIMIT 1` returns whichever session landed last — it handed me another lane's verdict.
   `diagnosis_artifacts` column is `body`, not `content` (`RUNBOOK` R8).
7. **The working tree may not build even when HEAD does.** On 2026-08-05 evening
   `tool_acceptance_actions.go` was another session's mid-edit WIP and broke `go build ./...`;
   committed HEAD built and tested clean via `git archive`. Check which you are looking at
   before diagnosing.

## 5. Lane files

`PLAN_2026-08-05` (decisions + 4 corrections to the inherited brief) · `NOTES_` (evidence and
every misstep — §11 the census error, §12 the live proof and its gap, §13 the owed check) ·
`RUNBOOK_` (9 checks, each with its gotcha) · `README_where_we_are` (owner's prose) ·
`SUMMARY_2026-08-05` · `HANDOFF_2026-08-05_code_review_triage.md` (the inherited triage, with
five corrections appended, NOT edited in place).

Fleet-wide: `WRONG_CALLS.md` ×2 (the reaper misread; the 3-of-6 census) · `LANDMINES.md` (key
rename, corrected in place with a dated note; `landmines-sync.py --apply` run) · concept
register `PBP-031` + `000_concept_index` + `link-management` (rename, and the same conflation
F7 caused there).
