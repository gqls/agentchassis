# CONTINUE HERE — code-review triage lane, 2026-08-05 evening

**State: the work is DONE, LIVE and pod-verified. All three council verdicts APPROVED. BOTH dated
re-checks are now CLOSED. What remains is three items that are deliberately somebody else's — and
one of those has moved.**

> **UPDATED 2026-08-07 01:1xZ — the owed 20:45Z read is DONE (4.5h late) and the verdict is
> STILL ZERO. `PBP-031`'s stop condition did NOT fire; the per-caller opt-in is not blocked.**
> Full read-out and its limits in **NOTES §17**; recorded in the register at `PBP-031`.
> Headline: **48 runs across THREE callers** (page-rerender 44, page-build-handler **3** — a new
> third caller, and the one F9's widening was actually about — page-rebuild 1), **55 saves over 16
> pages, 55 of 55 carrying `content_data`**, zero regression rows. Live on **`v1.0.1261`**
> (5th pod generation; re-probe, never quote a pod name). **Say "no regression in the retained
> window since the roll", NOT "verified" and NOT "never fired"** — §17c explains why this table
> has no "all history" to speak of.
>
> Two corrections this read produced, both in **NOTES §17**: the ~24h `orchestration_states`
> reaper is **confirmed and now characterised** (`COMPLETED`/`FAILED` only, on **`updated_at`**,
> hourly `database-cleanup`; boundary matched its own last run minus 24h to within 30s) — **and my
> first check of it said "24 days" and was blind in exactly the way `RUNBOOK` R6 warns about**.
> Second: §16's `domain` census is a **≤30-day window, not all history**, because clause 1 of the
> same task reaps this table — a fact **§4 of NOTES already recorded yesterday**.
>
> **2b was SETTLED 2026-08-06** (§2b — by config census; a log grep cannot answer it).
> **F13 re-checked 2026-08-06 11:2xZ: still another session's WIP (25/4), still not actionable.**
> **The `domain` item was RE-DIAGNOSED 2026-08-06 and is now FILED for diagnosis** — see §3.

Read `NOTES_code_review_triage.md` §11–13 first if you only read one thing — §11 and §13 are
where a successor is most likely to repeat a mistake. **Then §14–17.** §14–15 correct §13's
traffic check and §14c's own caller query; **§16 re-diagnoses the `domain` finding (it was
mis-framed three ways); §17 is the closing read plus two corrections to this lane's own
measurements.**

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

> **RE-VERIFIED TWICE ON 2026-08-06 — the pod names below are DEAD, and so are their
> replacements.** The chassis rolled to **`v1.0.1256`** overnight (probed 08:0xZ) and again to
> **`v1.0.1257`** at **09:52Z** — pods `agent-chassis-5b9fd84984-hqc5d` / `-qvzkg`. **Three pod
> generations in under three hours**, none of them this lane's doing. Re-probed on the current
> replicas [10:0xZ]: `incoming_sections_with_content_data` → **1 / 1**, misspelled control →
> **0 / 0**. The lane's source files are untouched at HEAD since `1c6a3cab6`, so 1257 is a
> rebuild of the same code. `bugs_open/153` says a roll is not evidence your fix shipped; the
> converse bit us twice — **a later roll retires your proof's pod names.** The §12 gap (no
> removed-unique-literal negative) is unchanged. **Re-probe before citing; never quote a pod
> name without its date and image tag.** `RUNBOOK` R12.

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

## 2. TWO DATED RE-CHECKS — ~~these are the actual outstanding work~~ **BOTH NOW CLOSED**

### 2a. 2026-08-06 ~20:45Z — the `CONTENT_DATA_REGRESSION` read-out (24h post-roll) — **DONE 2026-08-07 01:1xZ, verdict STILL ZERO**

> **CLOSED. Read 2026-08-07 01:1xZ, 4h31m after the due time. `CONTENT_DATA_REGRESSION`: 0 rows.**
> That is **branch 3** of the three decided below — not branch 1 (no `page-build-handler` rows) and
> **not branch 2, so `PBP-031`'s stop condition has NOT fired and the per-caller opt-in is not
> blocked by this read.**
>
> | | |
> |---|---|
> | regression rows | **0** |
> | runs since the roll | **48**, all COMPLETED, across **three** callers |
> | — `page-rerender` | 44 (29 durable to 08:48:25Z + 15 to 20:15:33Z) |
> | — `page-build-handler` | **3** (11:51Z → 20:39Z) — new, and the caller F9's widening was about |
> | — `page-rebuild` | 1 (08:32Z) |
> | saves | **55 rows / 16 pages / 55 of 55 carrying `content_data`** |
> | on `v1.0.1261` alone | 11 saves / 3 pages / 11 of 11 |
> | pod probe, both replicas | added literal **1 / 1**, misspelled control **0 / 0** |
>
> **The sentence to quote, and not a word stronger:** *48 runs across three callers and 55 saves
> over 16 pages produced no regression row, on a predicate proven able to fire at the unit level;
> every one of those saves carried `content_data`, so the silence has a demonstrable cause rather
> than being absence of traffic. NOT proven end to end — nothing exercised predicate → INSERT →
> `agent_error_log`.* Do **not** write "verified", and do **not** write "never fired": §17c shows
> this table is reaped at 14/30 days, so it has no "all history".
>
> **The lateness cost exactly one thing** — `orchestration_states` had reaped further, so the
> caller counts had to be reassembled from the durable 10:00Z record plus a non-overlapping
> increment rather than read off in one query (NOTES §17b). The early-denominator decision in §14c
> is thereby vindicated. Full read-out, the reaper characterisation, and two corrections to this
> lane's own measurements: **NOTES §17**. Recorded in the register at `PBP-031`.

The rest of this section is kept as written, because the decision rule was set in advance and the
verdict must be readable against it. This was the one that mattered, because **F9 changed the
predicate the check tests** and the register carries a stop condition on it.

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
> - **`page_components`, roll → 08:47Z: 47 rows inserted, 13 distinct pages, 47 of 47 carrying
>   `content_data`.** The path ran, and every incoming save carried structured content — so
>   silence is correct for a demonstrable reason, not for want of traffic.
> - **Callers that ran, via `owner_agent_type` (re-measured 10:0xZ, supersedes the 08:00Z
>   figures):** `page-rerender` **29** runs (20:54:05Z → 08:48:25Z) and `page-rebuild` **1**
>   (08:32:35Z), all `COMPLETED`. page-rerender is *the exact caller this section's stop
>   condition is about*, so tonight's read lands on the case the rule was written for.
>   **Identify callers with `owner_agent_type`, NOT by fingerprinting
>   `sections_metadata_field`** — that value is shared by four definitions, and a `count(*)`
>   over the step walk counts step occurrences rather than runs (`RUNBOOK` R10, corrected).
> - **Positive control passes at HEAD `61df92ff0`**: `TestShouldReportContentDataLoss` case 1
>   asserts the 194 signature returns `true`. The predicate can fire. (Scope: the decision
>   function only — not predicate → INSERT → `agent_error_log` end to end.)
>
> **So the "still zero" branch below is now stronger than it was written.** It is no longer
> "deployed-but-unexercised": it is *23 page-rerender runs with no regression, on a predicate
> proven able to fire*. Still short of end-to-end proof — say that, and do not round it up.
> Known limit: only the last save per page survives in `page_components` (30 runs → 13 pages),
> so the intermediate saves left no trace and the detector is their only witness.

**If you are the 20:45Z reader, this is the whole job — four commands, ~5 minutes.**

```bash
# 1. Which pods/image are live NOW? (three rolls happened on 08-06; never reuse a pod name)
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime'
# 2. Still carrying this lane's code? one exec per grep; expect 1 / 1 then 0 / 0
kubectl exec -n ai-persona-system <pod> -- grep -ac "incoming_sections_with_content_data" /app/agent-chassis
kubectl exec -n ai-persona-system <pod> -- grep -ac "require_save_without_sections_metadata" /app/agent-chassis
```

```sql
-- 3. THE READ ITSELF (occurred_at, NOT created_at)
SELECT agent_type, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_code='CONTENT_DATA_REGRESSION' GROUP BY 1 ORDER BY 2 DESC;

-- 4. THE DENOMINATOR, refreshed. Do NOT use agent_error_log for this (see the block above).
SELECT count(*) AS rows_inserted, count(DISTINCT page_id) AS pages,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS with_content_data,
       min(created_at), max(created_at)
FROM page_components WHERE created_at > '2026-08-05 20:45:00Z';

SELECT owner_agent_type, status, count(*), min(created_at), max(created_at)
FROM orchestration_states
WHERE created_at > '2026-08-05 20:45:00Z'
  AND jsonb_path_exists(workflow_plan, '$.**.steps.*.action ? (@ == "save_page_sections")')
GROUP BY 1,2 ORDER BY 3 DESC;
-- ^ this one is REAPED at ~24h. If it returns fewer runs than the 29+1 recorded at 10:00Z,
--   that is retention, NOT a drop in traffic. The 10:00Z figures above are the durable record.
```

**Then write the verdict against the three branches above — and record it in the concept
register at `PBP-031`, which is what carries the stop condition.** If it is still zero, the
honest sentence is *"30 runs across two callers, no regression, on a predicate proven able to
fire at the unit level; not proven end to end"* — not "verified".

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
- **The `agent_error_log` `domain` divergence — real, fleet-wide, nobody's. RE-DIAGNOSED
  2026-08-06 11:2xZ; read NOTES §16, not this paragraph's earlier version.** An unset domain is
  stored as `''` rather than NULL by **three of twenty** writers, and those three are one
  copy-paste family sharing a byte-identical 13-column block —
  `agenterrors/agenterrors.go:94`, `store_generated_component_action.go:1358`,
  `component_write_guard.go:320`. The other **17** store NULL (10 via `NULLIF($n,'')`, 7 by
  omitting the column), **so the convention exists and these three break it** — this is
  clone-and-drift, not a missing convention, and the fix is three characters at three sites.
  Live [MEASURED 11:2xZ]: 4,688 real domain, **9,949 `''`**, 128 NULL of 14,765 — so
  `WHERE domain IS NULL` sees **1.3%** of the no-domain rows, an under-report of **79×** (it was
  26× yesterday; the ratio drifts because the broken family contains the coordinator's generic
  writer, which produces most of the table). Confirmed by an exact `error_code` partition with
  zero overlap (NOTES §16d). **Use `COALESCE(domain,'') = ''` meanwhile.** Still NOT fixed: it is a
  stored-shape change on the fleet's highest-volume error writer and wants a council round.
  **⚠ Those figures are a ≤30-DAY WINDOW, not all history** — clause 1 of `database-cleanup` reaps
  this table at 14 days resolved / 30 unresolved (NOTES §17c; §4 of NOTES already knew, and I still
  wrote "all history"). The 79× and the partition hold for the population a reader can query, which
  is what the advice is about; totals-since-the-beginning are not available from this table.
  **FILED FOR DIAGNOSIS 2026-08-07 01:1xZ** via `090`, per the 2026-07-31 owner ruling that a
  cross-cutting root-cause claim is not "filed" until it has been through the loop.
  Intake correlation `94144fbc-3c01-4ed4-982a-bae8ac6caea8`; **the key the artifacts are written
  under is the RUN correlation `a7b1e113-8857-4161-ad2b-f3b7387e33e9`** — use that one:
  ```sql
  SELECT kind, metadata->>'verdict', left(body,4000) FROM diagnosis_artifacts
  WHERE correlation_id='a7b1e113-8857-4161-ad2b-f3b7387e33e9' ORDER BY created_at DESC;
  ```
  **VERDICT READ 2026-08-07 01:3xZ — `UNVERIFIABLE`, and it does NOT weaken the claim. NOTES §18.**
  The loop re-derived the mechanism independently from the code it could see — *"the mechanism (bare
  `$2` for domain) is real, but at a different symbol than named"* — and then could not check my
  **location** claim, because **it cannot see code written after 2026-07-28**: zero index hits for
  `package agenterrors` (created 08-06 by RFC_012), and zero-match symbol searches for the
  NULLIF-wrapping siblings, which are also post-07-28 files. It read the pre-RFC_012
  `agent_error_log.go:LogAgentError`, which genuinely did its own INSERT — **both descriptions are
  right about different trees.** The loop flagged its own staleness in its `code_requests`
  (`bugs_open/108`'s fix working as designed: it reports stale, it does not claim fresh).
  **⚠ DO NOT RESUBMIT until the code index moves** — same run, same answer, another round wasted.
  The index ref is pinned by migration 252 to `086_experience_loop` and wants `'main'`.
  **⚠ AND: the verdict was nearly LOST — read the new `LANDMINES` entry before you run any `090`.**
  The item finished `complete`, all three orchestrations `COMPLETED`, and `diagnosis_artifacts` held
  **5 rows, every one `kind='bundle'`, zero reports**. The verdict survived only in
  `collected_data->'verdict'`, which the 24h COMPLETED reaper deletes. Cause: the reply failed
  `message validation` and `coordinator.go` correctly converted it to a parent failure — but nothing
  propagated that into the work item or the artifact table.
  **Also corrects §16's volume claim (NOTES §18c):** the `''` population is **88% one live vet-lane
  incident** (12,090 of 13,783, from 08-04, still firing), not steady-state generic traffic. So the
  ratio is an **incident metric** — 26× on 08-05, 79× on 08-06, **~109× now** — and the defect's
  severity must be argued from the mechanism and the reader-side blindness, **never from the row
  count**. Argue the fix (three characters, three sites, converging on the majority) on that basis.
  **Two traps this re-diagnosis paid for**, both in NOTES §16: the `NULLIF` on the sibling
  columns is **compelled** by their `::uuid` cast (`''::uuid` errors — so that INSERT's internal
  asymmetry is not evidence of intent about `domain`); and **a refactor is not a review of what it
  moves** — RFC_012 relocated this exact INSERT on 2026-08-06 and carried the defect over verbatim.

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
