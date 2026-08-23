# NOTES — bugs_open/367 (running record, append-only, newest at the bottom)

## 2026-08-23 — re-verifying the bug before touching anything

Ran the live `classify` SQL (read from `agent_definitions`, both params bound) for item
`562788c3`. **`route=stale, component_id='', html_len=0`** — while component `0a1498b3`
sits at that (page, slot) with `build_status='pending'`, `locked_at` NULL, and **9,220
characters of `rendered_html`**. The finding is true: `n_still_empty=2` of `n_named=2`.
Bug confirmed still live.

Ownership checked: `scripts/who-owns.py 367` names `bugfix_342_absent_required`, but that
lane's own handoff banner says **LANE CLOSED**, names 367 as successor, and states *"Nothing
here needs picking up."* No competing session (`ListAgents` shows a `bugs_open/342` session
idle and a `bugs_open/344` session on a different residual).

## 2026-08-23 — MISSTEP 1: I measured `pages.sections` with the wrong shape and got a zero

To decide whether an "unresolvable target" route would ever be reachable, I asked how many
slots named in `pages.sections` have no `page_components` row. I unnested with
`jsonb_array_elements(...) s` and filtered `s->>'name' IS NOT NULL`.

**Result: 0 of 0.** I nearly wrote "the case is unreachable" into the design.

`pages.sections` is an array of **text** — elements look like `"hero"` — so `s->>'name'`
is NULL for every element and my filter discarded all 746 non-empty rows. The measurement
could not have come out any other way.

Redone with `jsonb_array_elements_text`, and vocabulary-checked before I trusted it (sampled
pages show `sections` entries and `page_components.slot_name` are the same vocabulary; 1,824
of 2,160 named slots DO resolve to a row, so the join is sound):

> **2,160** slots named in `pages.sections` on non-deleted pages; **336 (15.6%) have no
> `page_components` row at all**; 16 have a non-deployed, non-removed row.

So `count(comp)=0` is an ordinary state for about one planned slot in six — which is what
makes `close_stale`'s *"no longer exists"* a false statement quite apart from the
`build_status` narrowing. **Caught by:** the zero looked too round, so I ran a control on my
own query (`jsonb_typeof(sections->0)`) instead of on the world.

## 2026-08-23 — MISSTEP 2: I claimed a route "had fired exactly once ever" off a survivor artefact

I wrote that `stale` had been taken **once in the platform's history**, from
`SELECT min(created_at) FROM orchestration_states` reading 2026-07-19 — i.e. "the table goes
back a month, and there is only one `stale` in it".

**Refuted.** Retention is about **two days**: 08-22 has 1,324 rows, 08-23 has 3,299, and then
nothing at all until four days in July totalling **24 rows, every one `CANCELLED`** — stuck
rows the cleanup skips. There are **zero** rows for 2026-08-14→19.

**Caught by:** reading `bugfix_277_required_fields_repair/RUNBOOK_required_fields_repair.md`,
which records canary `332bb3f6` closing `complete/stale` on 2026-08-15 — a route my census
said had never happened. A minimum over a retained table is not a retention window.

The bug does not rest on this: the hand-run classify and the whole-population
re-classification are independent of the audit trail.

## 2026-08-23 — MISSTEP 3: I passed on a count that was my own `LIMIT`

I told an adversarial reviewer *"all 8 existing `content_rewrite:from_rfm:` conversions are
`failed`"*. The 8 was the `LIMIT 8` on my own sampling query. The real census is **31** — 28
`failed`, 2 `cancelled`, 1 `complete`. The direction held and the conclusion survived, but I
had passed a number off as a finding when it was an artefact of how I looked.

## 2026-08-23 — the design changed twice, and both changes came from measurement

I started on the bug file's own candidate 1 ("widen the `comp` CTE"). Two findings moved me
off it:

1. **Widening alone repairs nothing.** `file_rewrite` reads `spec.component_id`, `.page_id`,
   `.component_function`, `.reason`. Post-deploy items carry all four (62 of 62); render-time
   items carry none (0 of 3). Both `spec_paths` (`create_work_item_action.go:281,294`) and
   `item_key_suffix_field` (`:252-256`) are **deliberate hard errors** when unresolved. I had
   predicted a *degenerate colliding key*; the truth is better — it is a loud error, by a
   council-hardened design. Either way, widening buys a loud failure, not a repair.
2. **The repair arm is the wrong destination.** `content_rewrite`/`edit_live` runs at
   `page-build-handler`, whose `save_page_sections_action.go:823` DELETEs every
   agent-writable row on the page and `:1014` reinserts at `'deployed'`. 28 of 31 `from_rfm`
   conversions already fail there on the owned-page refusal (`bugs_open/333`, owned by the
   277 lane) — and the 367 page is `rebuild_policy=owned`.

So the rule became: **a disposer may close only on positive evidence of absence.** The
estate already states it at `revalidate_review_queue_action.go:684`.

## 2026-08-23 — the fix, proven before it was applied

Migration `574`. Everything below was run **inside a transaction that was then rolled back**,
so production was never used as the test rig:

- migration parses; its own verify block passes all 8 assertions incl. negative controls
- **C1** real item → `target_not_dispatchable`, `target_state=pending`, component RESOLVES
  (`0a1498b3`, `html_len=9220`, `n_still_empty=2`)
- **C2** retired slot (`tool-clip-path`/`ported-page`) → **still `stale`**,
  `target_state=component_retired`
- **C3** page that does not exist → **still `stale`**, `target_state=page_missing`
- whole population, all 65 items, old vs new: **exactly one route changes**
- **apply-then-rollback returns `default_config` BYTE-IDENTICAL**

Then applied for real, and re-verified by reading the query **back out of the live row**:
same three controls pass, same one-route delta. Council submitted first —
`d48c0a89-9ff8-4286-bfe9-2690dc13d5bc`.

Two SQL traps worth carrying: `snapshot_agent` is overloaded, so a bare literal gives
`function snapshot_agent(unknown) is not unique`; and `to_jsonb()` over adjacent
string literals is `unknown` — cast with `::text`.

## 2026-08-23 — what the fix does NOT do, stated so nobody reads more into it

The render-time population is now **visible and honest**, not repaired. It parks at
`needs_human_review` with the facts and the repair paths on the row. Repair needs
`bugs_open/333` plus a producer that writes the convert arm's read-set. Both named, neither
taken here. Do not let anyone write that 367 "made the render-time findings repairable".
