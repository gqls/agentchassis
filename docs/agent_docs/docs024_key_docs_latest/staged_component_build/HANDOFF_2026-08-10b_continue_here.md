# HANDOFF — 2026-08-10 (b), fresh chat starts here: two storage fixes await their proofs, and batch 8 has two clean subjects left

**Supersedes `HANDOFF_2026-08-10_continue_here.md`** for state and work-list. Still
binding from the chain and not repeated: the 08-09 handoff's §0 (shared-228) and §2 (the
two rerender traps), and the 08-08 handoff's §3 (the interactive-fence line). The 08-10
handoff's §2 correction box (the batch-8 requalification) is carried forward below in
condensed form — read the original if you touch the loancalculator or gamesdesign tools.

## 1. State (verified 2026-08-10 ~22:15Z)

- **57 subjects proven end-to-end: 54 sections + 3 tools.** `tool-setup-builder` went the
  full line on 08-10: fence 15/15 live, 11/11 mutants, PLAN persisted at
  `subject_type='tool'` byte-identical, S6 in-cluster 15/0/9 (all skips profile-gated,
  0 unimplemented), no work item raised. NOTES `## 2026-08-10 (second session…)` entries
  hold every table.
- Fleet: **v1.0.1283** (chassis pods up 21:43Z 08-10). Re-grep at session start.
- Naming-contract check after the persist: PASS — 26 testable / 12 backlog / 0 BROKEN.
- **The tree carries another session's compile-breaking WIP** (`save_page_sections_action.go`
  modified + untracked `save_sections_decision_gate.go` — `undefined: pq`, a type
  mismatch). `go build` on the working tree FAILS through no fault of this lane's files;
  build against `git archive HEAD` (RUNBOOK practice) until they commit.

## 2. THE TWO OPEN PROOFS — do these before any new work

### 2a. bugs_open/243 (vision/storage for tool acceptance) — fix APPROVED + rolled, proof owed

`tool-acceptance-agent` was added to `storageAgents` (`spawn_actions.go`), commit
`543206039`, council **APPROVED round 1** (`5eb4ad58`, 2 advisory objections both
answered by query — see the bug file). v1.0.1283 was built after the commit **but a
pod-grep cannot prove this change shipped** (no unique literal; rodata dedupes; no VCS
stamp). The proof is behavioural:

```sql
SELECT correlation_id, processing_node, current_step,
       collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE owner_agent_type='tool-acceptance-agent' AND created_at > '2026-08-10 21:43+00'
 ORDER BY created_at;
```

- PASS = a run on an `agent-tool-acceptance-agent-*` node reaching **`complete`** (not
  `complete_no_look`), empty step error, **and the first-ever vision rows in
  `llm_call_log`** (0 all-history as of 08-10; that also finally exercises MDL-040's
  provider path — read its register entry before quoting it).
- A run on an `agent-chassis-*` node proves NOTHING here (the inline path is
  deliberately unfixed — owner decision still open, candidates 2 and 3 in the bug file).
- `complete_no_look` on a POST-roll SPAWNED run = fix absent or wrong — re-open loudly.
- The sweep historically runs ~19:00Z–04:00Z; if nothing has spawned by your session,
  the due-sweep may simply not have raised items — check
  `site_work_items WHERE item_type='acceptance_run'` before concluding anything.

### 2b. bugs_open/245 (chassis must not carry B2 credentials) — CODE HALF committed, verdict + roll + overlay half owed

The spawn block's four `os.Getenv` credential value-copies are now `secretKeyRef` against
`personae-storage-secrets` (commit `e7e3b4e3c`, trailer
`Council-Submitted: c45c6412-20aa-45ab-b5ae-38fcc2bd7887`). References are REQUIRED —
missing key = visible `CreateContainerConfigError`, replacing the bug's silent
first-use failure. Owed, in order:

1. **Read the council verdict on `c45c6412`** (a poll was running when this handoff was
   cut; find by payload: `collected_data->'input_data'->>'fix_correlation_id'`). Act on
   REVISE/REJECTED — the code is already on the shared branch (forward-only).
2. **After the NEXT roll** (1283 does NOT carry it): verify a spawned storage pod
   (`kubectl get pod <agent-…> -o yaml` → the four env vars must be `valueFrom:
   secretKeyRef`, not `value:`), then prove a real storage OPERATION at the artefact —
   the same-lane CONTRIBUTION in the bug file (bugs_open/248 context) sharpened this: a
   pod that STARTS is the bug's failure mode, not its refutation. `deploy_image_asset`
   succeeding end-to-end is the canonical proof.
3. **Only then** delete overlay lines 77–98
  (`deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml`)
  — and re-run BOTH greps first (`os.Getenv` for the four names — now expect 0 in Go —
  and `AccessKeyEnvVar`), then candidate 3's post-removal checks (bug file).
4. Note for whoever rolls: under the new binary the chassis credential lines are inert
   (nothing reads them), so the window between roll and overlay-removal is safe in that
   direction only.

## 3. Batch 8 — what actually remains

The 08-10 morning pool of "17 ready tools" requalified to (full evidence: NOTES
`## 2026-08-10 (second session)`; correction box in the superseded handoff):

- **DONE**: `tool-setup-builder`.
- **2 clean, fork-free, resolvable subjects left**: `tool-grip-force-friction-calculator`
  and `tool-matchmatrix` — both robot-hands, both pages `rebuild_policy='owned'`. **Ask
  the robot-hands lane before firing S6** — a failing verdict dispatches `tool-improver`
  at their owned pages (RUNBOOK §11 note). The fences themselves can be authored and
  mutation-proven without dispatch; only the S6 run needs their word.
- `tool-llm-cost-calculator` — resolvable but has FOUR forks sharing the one
  `doc_plans` key (templates differ up to 3.3 KB). Author fork-aware or defer.
- **9 unresolvable under the Tier-4 lookup** (8 loancalculator naming mismatches — their
  lane's call; `tool-bayesian-ranking` — the §11 two-row rename restores gamesdesign's
  own convention). A PLAN authored for any of these lands in BROKEN-A.
- Blocked/skip unchanged: fuel-budget-forecaster (gaswholesalers logo 404, 7+ days),
  gas-unit-converter (known-broken page, owner call).
- The tool line for the next subject: HANDOFF_2026-08-10 §3 (the generator now takes
  `subject_type`/`kind`/`batch`/`mutants_file` per entry — `manifest_batch8.json` is the
  worked example, commit `8fa701849`+`40c0f17f2`).

## 4. Standing defect list

Items 1–8 unchanged from `HANDOFF_2026-08-09` §4. Item 9 (vision/storage) → now
`bugs_open/243`, fix in flight (§2a). Item 10 (batch-8 naming gate) stands. NEW since:
`bugs_open/245` (§2b) and the same-lane parallel session's `bugs_open/248` (deploy
defects around `deploy_image_asset` — read it before touching asset deploys; it also
carries the note that the chassis's missing `IMAGE_BUCKET` makes the generic-topic
deploy route fail outright, which is 243's inline-path sibling).

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK. The tree is co-edited and was
   carrying a compile-breaking WIP at cut time (§1).
2. Pod-grep chassis (`request_component_browser_run` ≥1 + negative control) and
   browser-runner (three long markers, RUNBOOK §4). No dispatch within 300s of a restart.
3. Run §2a's query (the 243 proof), then read `c45c6412`'s verdict (§2b item 1).
4. Re-run the census + `CHECK_naming_contract.sh` before quoting any batch-8 figure.
5. `who-owns.py` + live-transcript grep before writing at robot-hands, loancalculator,
   or anything touching 248's deploy surface.
