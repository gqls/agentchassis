# CONTRIB — how the `tools-api` merge conflicts were resolved (2026-07-28)

**For the workstream that owns `tools-api`. Written by the architecture_review
thread at the owner's request; you did not ask for this and you should audit it.**

`main` was merged into `087_towards_multiple_domains` (merge commit
**`176e56ad8` "v1.0.1189 merge complete"**). Seven files conflicted `add/add`, all
of them yours. The merge commit message records none of the reasoning, which is
why this file exists.

## Why they conflicted at all

PR **#3** (`feat/c379f7b7`, the feature-builder's) created `tools-api` on `main`.
This branch created the **same paths independently** rather than by merging that
PR — so git had no common ancestor for them and reported `add/add` on every file,
even though the content is the same lineage. Nothing was diverging; the history
just could not see it.

## The verdict, and it was the same in all 27 conflict regions

**Ours is main's version PLUS the 083 work. Main had nothing ours lacked, in any
region.** All five files I was asked to resolve went to ours. Checked hunk by
hunk, not assumed from a file-level "ours looks newer".

| file | what ours has that main's lacks |
|---|---|
| `store/rounds.go` | `COALESCE(position_text,'')` / `COALESCE(defence_text,'')`. pgx cannot scan NULL into `string`, so **main's version errors on every round that has not yet stored a position or defence** — which the handlers then report as "round not found". Your comment dates this to the first island smoke, 2026-07-25. |
| `198_tools_api_gauntlet_rounds.sql` | The DO-block guard. **Main guards with a top-level `ASSERT`, which is not SQL** — `ASSERT` exists only inside PL/pgSQL. I ran main's form against the live DB inside a rolled-back transaction: **syntax error**. Main's guard could never have run, so this is not a style preference. |
| `handlers/round.go` | **503 not 502** (Cloudflare replaces an origin 502's body, destroying the JSON error shape — your `b498df16b` did this for `/position` and `/defend`; `/round` was missed until 07-27), plus `logAIFailure` / `logInternalFailure`. |
| `handlers/position.go` | The 503 fix, the structured failure logging, an explicit `api_key_env_var`, and `errors.Is(err, pgx.ErrNoRows)` so **only a genuinely missing row is a 404** — main returned 404 for a DB outage too, disguising an infrastructure failure as the caller's mistake. |
| `handlers/defend.go` | Identical to `position.go`, all nine regions. |

`internal/tools-api/httperr` exists on **both** sides and merged cleanly, so
main's `httperr.JSONError` calls were never a lost dependency — that was the one
thing that could have made "take ours" quietly wrong.

`api/server.go` and `tools-api.dockerfile` were resolved by the owner, not me:
`server.go` is ours plus a blank line (it keeps your `gin.Logger()` addition from
083), and the dockerfile is byte-identical to ours.

## What was checked after resolving

- No conflict markers anywhere in `internal/tools-api` or migration 198 at HEAD.
- `go build ./internal/... ./platform/...` — clean.
- `go test ./internal/tools-api/...` — `handlers` passes; the rest have no tests.
- **Two `main`-only files are dropped by the merge, correctly**:
  `platform/orchestration/actioncheck/local_actions.go` and an archived
  `TRIGGER_code_indexer_v2(1).sh`. Both existed at the merge base and were
  deliberately deleted on this branch (`c82b2872c`, the bugs_open/017 actioncheck
  rework; `a2cc77ecd`, a duplicate cleanup), and `main` never touched either — so
  git takes our deletion silently and nothing of main's is lost.

## If you disagree with any of it

The pre-merge state is recoverable: `176e56ad8^1` is this branch before the merge
and `176e56ad8^2` is `main`, so
`git show 176e56ad8^2:internal/tools-api/handlers/position.go` gives you main's
version verbatim to diff against what shipped. Forward-only — correct it with a
new commit rather than rewriting the merge.
