# HANDOFF — the 223 / 254 / RFC_022 lane, cold start for a fresh chat

Written 2026-08-15. **Supersedes `HANDOFF_2026-08-14_continue_here.md` in full** — every
item it listed as open has since been ruled and executed. Read this file, then `NOTES`
(newest at the bottom, the 2026-08-15 block). Everything below is verified unless
marked otherwise.

---

## 0. WHAT IS ACTUALLY LEFT (read this and you can stop)

**One owner decision, two watch items, zero pending builds.**

1. **OWNER DECISION — the two remaining standing reviews.** The daily
   `optional-key-budget-check` (06:50 UTC) truthfully reports **2 findings every
   morning**: `analyse_repo_local` (12 optional keys, carriers code-indexer +
   diagnose-agent) and `diagnose_prepare_fix_commit` (11 keys, feature-implementer +
   fix-implementer). Each owes ONE review of its accumulated surface, then an ack
   entry (see §2 mechanics). The owner's standing choice (2026-08-14) was "wait for
   natural contact"; the visible cost, now that the clock runs, is a permanently red
   morning report — which habituates readers, the exact decay RFC_022 was filed
   about. The worked example to copy is
   `architecture_review/REVIEW_2026-08-14_append_doc_note_optional_surface.md`
   (method: read the action + every key's read-site; census live keys with
   `--live-pairs`; date the keys with `git log --follow`; verdict trim-or-acknowledge;
   ack both sides; re-apply the overlay).
2. **WATCH — the as-of note's first live rendering** (`bugs_closed/254` residual).
   Renders only on EMPTY code-lookup answers; measured 2026-08-15, **zero verify runs
   have happened at all** since 08-14 12:00, so it is unobservable for want of demand.
   Check: `SELECT count(*) FROM llm_call_log WHERE step_name='verify' AND
   prompt_rendered LIKE '%as-of: this answer describes commit%';` — a zero means "no
   empty answer yet", NEVER "note missing". When it first fires, note it in
   `bugs_closed/254`'s closure section.
3. **WATCH — kind-shaped misexplanations post-fix.** `[UNMEASURED]` fleet-wide (the
   both-ways proof is n=1). Settles itself as landmine verifications accumulate;
   compare any wrong verdict's reason against the as-of note in the same prompt.

## 1. State in one paragraph

Both bugs are CLOSED, live, and in `bugs_closed/` (`223`: the index states what it
cannot represent + var/const indexed; `254`: the indexed commit travels with every
empty answer and dates every persisted verdict — live since v1.0.1297, verified through
v1.0.1300 by ancestry). RFC_022 is CLOSED end-to-end: interim (migrations 381/383),
counter built (register **WFA-013**), budget **RULED N=10** (404/405 put the ruled
trigger in both rosters, 377 breakpoint intact), and the counter runs daily
(`optional-key-budget-check`, Python-mirror shape at the owner's direction, first
scheduled run proven 2026-08-15: 186 agents, 2 findings, ack honoured). The first
standing review (`append_doc_note`) concluded ACKNOWLEDGE AT 11, no trims — 10 of its
11 keys are the birth-day schema mirror of `doc_notes`, the 11th is 223's own
council-approved suffix. **Owner framing, binding:** sharing is estate design; a budget
finding reviews the accumulated SURFACE, never the reuse (auto-memory:
`shared-actions-are-estate-design-not-a-smell`).

## 2. The machinery this lane leaves behind (do not rebuild any of it)

| thing | where | proof |
|---|---|---|
| answer-site as-of note + dated evidence line | `diagnose_code_lookup_action.go`, `diagnose_load_runtime_action.go` | live ≥1297; persisted verdict `16f0475d` |
| counter + census | `cmd/config-key-audit --optional-key-budget [N] [--acks f]`, wrapper `scripts/audit-optional-key-budget.sh` (defaults N=10, passes acks) | live fleet runs 08-13/14/15 |
| acknowledged baselines | `architecture_review/optional_key_budget_acks.json` (source of truth) → Go `--acks` + cron `ACKED_LEVELS` mirror | first scheduled run: ack honoured |
| daily check | `deployments/kustomize/services/optional-key-budget-check/` (06:50 UTC; one doc_notes row per run incl. clean; TWO rows on a red day — retry) | `subject_key='optional-key-budget'` rows |
| parity guards (FOUR) | `cmd/config-key-audit/optional_budget_cron_parity_test.go` | each mutation-proven; traversal fixture picks its subject dynamically |
| roster clauses | migrations 381/383 (interim) + 402/403 (counter named) + 404/405 (N=10 ruled) — all applied + recorded | 405 verify: 17 seats marked, 1 shared prefix |

## 3. The four things a fresh session will most likely get wrong

1. **Editing `check.py` (acks or counts) without re-applying the overlay.** The
   configmap is generated from the file; parity tests force the repo halves together
   at BUILD time but nothing forces the APPLY. After any edit:
   `kubectl apply -k deployments/kustomize/services/optional-key-budget-check/overlays/production/uk_001/`.
2. **Reading the daily red as noise to silence.** The 2 findings are truthful until
   the two reviews land. Never quiet them by raising N or hand-editing ACKED_LEVELS
   without a review — the ack file's `review` field is the licence.
3. **Probing a pod binary behind `2>/dev/null`.** A deleted pod's `NotFound` is
   indistinguishable from "unstamped" — this lane produced thirteen false absences
   that way. Digest-match first, stderr visible, always a positive AND negative
   control. The fleet rolled 1295→1300 in ~36h; read the service you mean.
4. **`099_SYNC_gate_roster.py --apply` is still suspended** (would revert 377's 68%
   caching saving). Roster edits are surgical anchored migrations — worked pairs
   381/383, 402/403, 404/405.

## 4. Standing traps carried forward

Indexer reads the REMOTE tip (`git ls-remote`, never `rev-parse origin/…`) ·
`landmines-sync.py --apply` consumes the NEEDS_VERIFICATION signal (use
`landmines-verify-dispatch.sh` to verify-on-sync) · no orchestration dispatch within
~300s of a chassis restart · `verify_unverifiable` has still never executed ·
`code_symbols` has `line_start`/`line_end`, no `indexed_at` · never
`ORDER BY created_at DESC LIMIT 1` on a shared table · `go run` folds exit 2 into 1 —
wrappers discriminate refusal by EMPTY STDOUT · migration numbers 402 (and bug numbers
223 et al.) are double-assigned across lanes — resolve by slug/filename.

## 5. Where the paperwork is

This directory: PLAN, RUNBOOK, NOTES (append-only), README_where_we_are (owner's log —
his rulings of 08-14 are recorded there in his own framing), SUMMARYs 08-10/08-11/08-14
(the series is the record), council submissions, superseded handoffs (08-10, 08-14).
Fleet-wide: `bugs_closed/223`, `bugs_closed/254`, `architecture_review/RFC_022_*`
(STATUS: CLOSED, rulings verbatim), `REVIEW_2026-08-14_append_doc_note_*`,
`optional_key_budget_acks.json`, 016b §9 ×2 + §10 rows, `WRONG_CALLS.md` ×3,
`LANDMINES.md` (staleness entry + 08-11 correction), register WFA-013 + DIAG-036/042,
auto-memory `shared-actions-are-estate-design-not-a-smell`.
