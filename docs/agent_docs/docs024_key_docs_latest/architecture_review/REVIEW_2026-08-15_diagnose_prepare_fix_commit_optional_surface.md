# REVIEW — `diagnose_prepare_fix_commit`'s accumulated optional surface (RFC_022 budget, third and last of the standing findings)

**Reviewed 2026-08-15 at the owner's direction**, under the 2026-08-14 rulings (N=10 on
shared actions; the surface is reviewed, never the reuse). Carried by **2 live agents**
(feature-implementer, fix-implementer) and declaring **11 optional keys**.

**VERDICT: ACKNOWLEDGE AT 11. No trims.** Baseline recorded in
`optional_key_budget_acks.json` and the cron's `ACKED_LEVELS` mirror.

## Method

As the worked example: implementation read with every key's read-site confirmed
(`diagnose_prepare_fix_commit_action.go:108-218`); live census via `--live-pairs`;
keys dated via `git log --follow` / `-S`.

## The strongest evidence first: this surface is FULLY EXERCISED

**All 11 optional keys are configured in live workflows** — the only one of the three
flagged actions where nothing at all rides on defaults alone. It is also the only one
of the three **opted in to unknown-config-key detection**, so a stale or misspelled key
on any of its steps is already reported by the validator. There is nothing here that
could even be suspected dead.

## The surface, decomposed — 11 keys are the commit pipeline's inputs, named

| concept | keys | read at |
|---|---|---|
| what to commit | `plan_field` (`:108`, missing = refusal), `files_field` (`:123`, missing = refusal), `originals_field` (`:142`) |
| why (provenance in the message) | `diagnosis_field` (`:216`), `council_field` (`:218`) |
| where | `repo_name` (`:178`), `base_branch`/`base_branch_field` (`:179`/`:185` — the one duality), `branch_field` (`:200`) |
| how described | `commit_message_field` (`:207`) |
| safety allowlist | `expected_symbols_field` (`:165`) |

Each key names one input of a commit: the approved plan, the implementation, the
original file bodies, the diagnosis and council rows quoted into the message, the
repo, the base branch, the branch, the message, and the expected-symbols allowlist
that is this action's stated reason to exist ("the implementer's allowlist safety
core"). None is a convenience alias of another.

## Growth history — three named, approved steps, finished 2026-07-18

Born `a4c6cc637` (2026-07-12) as the safety core; `expected_symbols_field` and the
stage-loop keys arrived with `c19b5d097` (2026-07-17, feature-builder delta 2,
"E1–E5 approved"); `base_branch_field` arrived `0dd750bcc` (2026-07-18, fixloop F1.2 —
a per-run base branch **replacing a stale literal**, i.e. the addition was itself a fix
for the stale-hardcoded-value class). **Zero additions in the four weeks since.** Every
post-birth key landed under a named, reviewed delta.

## What would reopen this

A twelfth key pages again (baseline 11), and under the ruled trigger that growth is
architecture-scope for a shared action; any authority-bearing addition is
architecture-scope regardless of count.
