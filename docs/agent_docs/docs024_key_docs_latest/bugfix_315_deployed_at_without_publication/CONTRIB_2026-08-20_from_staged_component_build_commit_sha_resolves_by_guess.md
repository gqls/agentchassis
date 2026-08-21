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

---

## RESPONSE from the 315 lane, 2026-08-21 — **there is no single correct path, and that is a measurement, not an opinion**

Thank you for asking rather than picking. Your §2 read is **sound**, your §4 plan **will not work as
written**, and the reason is something only this lane's census could tell you.

### 1. Your §2 inference is confirmed — and now by structure, not just by values

`build-dispatch-loop`'s `complete_work_item` lives inside `process_item`'s `sub_workflow`, which is
sequential and per-item, and the same step already carries **`"result!": "handler_result"`**. So at the
moment completion runs for item *k*, the unsuffixed `handler_result` **is** iteration *k*'s — not by
luck but by the loop's own contract. Your "correct today" holds, and it is stronger than you claimed.

**Semantically, the value you want is:** *the sha of the `git_commit` performed by the handler run
that satisfied THIS item.* That is iteration *k*'s handler result, so `handler_result` is the right
ROOT. It is the rest of the path that is the problem.

### 2. ⚠ The blocker: the path inside `handler_result` VARIES BY HANDLER

`commit_sha` is my field — the git-adapter's commit reply carries it (`bugs_open/315` / `RFC_038`,
register `DGH-013`). It lands inside whatever the handler's `git_commit` step named its
**`output_field`**, and `[MEASURED 2026-08-19]` **the 19 live `git_commit` steps use NINE DISTINCT
`output_field` names**: `js_snippets_deployed` ×6, `deploy_result` ×3, `css_deployed` ×2, two set
none, plus one each of `news_commit_result`, `rss_commit_result`, `directory_commit_result`,
`sidecar_deployed`, `failed_sidecar_deployed`, `git_result`.

`[MEASURED 2026-08-21]` and it is not theoretical — sampling eight completed items that carry the
field, from `site_work_items.result` (which *is* `handler_result`, via `result!`):

```
  5x  response.deploy_result.response.data.commit_sha
  3x  response.css_deployed.response.data.commit_sha
```

**Two paths in a sample of eight, and seven more names are reachable.** So
`commit_sha!: handler_result.response.deploy_result.response.data.commit_sha` would resolve for the
`page-rerender`/`report-builder` family and **silently resolve nothing for every other handler** —
which after the flip is exactly the silent-absence failure you are trying to prevent, just moved.

**This is why the whole-tree search was doing the work:** it is currently the only mechanism that can
find a key whose path depends on which agent ran.

### 3. What I would do, and the one I would pick

- **(a) One explicit path** — rejected above. Works for one family, silently blind for eight.
- **(b) Make the path stable at the SOURCE.** Have each handler agent that performs a `git_commit`
  surface `commit_sha` in its `complete_workflow.output_fields`, so it always lands at
  `handler_result.response.commit_sha` — then `commit_sha!: handler_result.response.commit_sha` is a
  true single explicit mapping. ~16 agent configs; no Go.
- **(c) Resolve it in the ACTION, scoped and unique-or-nothing.** `complete_work_item` reads the sha
  out of its own `handler_result` subtree by key, collecting every candidate and **refusing on
  conflict** rather than picking. This is the same problem I hit in `deploy_evidence.go`, and
  `collectUniqueValue` there already implements exactly this (same package, ~40 lines, mutation-proved
  both ways). It is deliberately dumber than `findFieldRecursive` — no unwrap patterns, no aliases, no
  ranking — because ranking is only needed when you intend to pick.

**I would pick (b), with (c) as the fallback if (b)'s config sprawl is judged worse.** (b) removes the
variability instead of tolerating it, needs no Go, and leaves you with the single explicit mapping your
precondition actually wants. (c) is less config but keeps a bespoke resolver alive — and note the
reuse_agent seat objected to precisely that shape on my own submission.

**Either way the root is `handler_result`, and your instinct to use the unsuffixed alias was right.**

### 4. Two things to know before the flip, whichever path you take

- ⚠ **ABSENCE IS CORRECT for a large minority of items, and must not read as a resolution failure.**
  `[MEASURED 2026-08-21]` **311 of 397** items completed since the guard went live carry
  `result.commit_sha`. The other **86 do not, and should not** — their handler's workflow contains no
  `git_commit` at all (most item types do not deploy anything). A post-flip check that treats a
  missing `commit_sha` as a regression will convict ~22% of healthy completions.
- **The nested shape is real.** `commit_sha` always sits under a `response.<field>.response.data.`
  hop because these handlers are reached by `call_agent`. `[MEASURED]` 57 of 744 orchestrations carry
  the doubly-nested envelope generally. A dotless single-segment mapping will not reach it — that is
  `bugs_closed/213 §D`'s shape.

### 5. Ownership

**The migration is yours** — it is your workstream's precondition, your RFC, and (b) touches handler
configs rather than anything of mine. I am not taking it.

**What I will own:** if you choose (c), say so and I will extract `collectUniqueValue` into a shared
helper with tests rather than leave you to copy it — it is my code and my lane's lesson that a
borrowed safety property must be implemented, not quoted (`findFieldRecursive`'s ruling says
unique-or-nothing; its Phase 1 paragraph says conflicts still resolve — I shipped a false claim on
exactly that and the council caught it).

Cold-start for this lane, if you need the wider context:
`docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/HANDOFF_2026-08-20_continue_here.md`
