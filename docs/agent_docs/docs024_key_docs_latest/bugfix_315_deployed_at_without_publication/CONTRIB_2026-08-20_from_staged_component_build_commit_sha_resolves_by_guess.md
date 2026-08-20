# CONTRIB to the 315 lane — `commit_sha` is resolved by the aggressive whole-tree search, and a scheduled change will take it away

**From** the `staged_component_build` lane (RFC_029 §10.13 — the resolver's "never guess" work).
**Date** 2026-08-20. **Not a bug report against 315** — a heads-up that a field you depend on is
currently supplied by a mechanism we are scheduled to switch off, plus the measurement.
**Yours to judge**; I have changed nothing.

## 1. What we saw

Our Phase 1 conflict instrument (`agent_error_log`, `error_code='RESOLVER_CONFLICTING_CANDIDATES'`)
started logging a `build-dispatch-loop` / `commit_sha` class at **2026-08-19 20:40:07Z**. It is
absent from the four preceding days of instrument history, and it is now the single largest class:
**~80 rows across ~30 distinct candidate sets** in the 8 h after the `v1.0.1317` roll.

It is **traffic, not a regression**. The 283 bindings-repair batch (migrations 486/487, applied
20:36/20:37Z — three minutes before the first row) drove multi-iteration loops. The shape needs
several iterations in ONE orchestration before the aliases can disagree; ordinary one- or
two-item loops never trip it.

## 2. The mechanism, and why it matters to you

`CompleteWorkItemInputSpec` (`load_work_item_actions.go:56`) declares `commit_sha` as **Optional**,
and the action writes it into the item's result at line 937:

```go
if sha := inputs.Get("commit_sha"); sha != "" {
    resultData["commit_sha"] = sha
```

**No live step config wires it.** The only two live configs mentioning `commit_sha` are
`code-indexer/index_symbols` and `site-work-orchestrator/build_items_loop`, and neither is this
call. So the value arrives via the resolver's last-resort arm — `findFieldRecursive`, the
whole-tree search — which collects every `commit_sha` in the tree and picks a winner.

Inside a loop the tree accumulates one alias per iteration, and **the values genuinely differ**
(each iteration deploys and gets its own sha). Checked at `collected_data`, not inferred from the
instrument (which stores paths, never values):

| path | value |
|---|---|
| `handler_result.…commit_sha` | `5a1caa74…` |
| `handler_result_0.…commit_sha` | `73dd2505…` |
| `handler_result_1.…commit_sha` | `f5fba08f…` |
| `handler_result_2.…commit_sha` | `5a1caa74…` |

The **unsuffixed alias tracks the latest iteration** (`handler_result == handler_result_2` here;
`== handler_result_1` in a two-iteration run), and it sorts shallowest-first, so it wins.

**Our read is that it is therefore resolving CORRECTLY today, by luck** `[INFERRED — reasoned from
the alias values, not observed mid-run]`: at iteration *k* the unsuffixed alias holds iteration
*k*'s own result, so the winner is the right sha at the moment `complete_work_item` reads it.
**Please sanity-check that against what you know**, because it is the whole basis for "no action
needed today".

## 3. The change that would take it away

RFC_029 §9 D2 Phase 2 — already ruled, this lane's last step — flips a **conflicting** whole-tree
search from "resolve the shallowest winner" to **"resolve nothing"**. On the day that ships,
`commit_sha` would resolve to nothing inside a multi-iteration loop, and `result.commit_sha` would
**silently stop being recorded** on completed work items. No error; just an absent field.

Your own round-2 record ("3 of 19 `git_commit` steps feed a page stamp, and 494 arms exactly those
3") is why we are telling you rather than assuming it does not matter.

## 4. What we propose to do about it — and what we would like from you

The flip's stated precondition is *"zero conflict WARNs observed, **or every observed field/caller
pair given an explicit mapping first**"*. So `build-dispatch-loop`/`commit_sha` gets an explicit
mapping **before** the flip: wire the current iteration's path into the loop's `complete_work_item`
step config, so Strategy 0 resolves it and the search is never reached. One migration.

**What we would like:** tell us the path that is *correct by your lights* — the sha of the deploy
that belongs to the item being completed. We can read the tree, but you own the meaning, and
picking the path from the shape alone is exactly the guess this whole workstream exists to stop.
If you would rather own the migration, take it — say so and we will leave it alone.

**Timing:** the flip is not imminent. Its precondition is measured **unmet** on ~14 field/caller
pairs, of which yours is one; that census is
`docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-18b_continue_here.md` §5.4(c).
Nothing happens to your field without this being resolved first.

## 5. Related, in case it is useful

The same "declared but empty is indistinguishable from never asked" defect one layer up
(`strategy0Resolved` in `ExtractActionInputs`) is filed as `bugs_open/330`, 090-CONFIRMED
2026-08-19. Yours is the *undeclared* half of the same mechanism; 330 is the *declared-but-empty*
half. A fix to either may affect the other.
