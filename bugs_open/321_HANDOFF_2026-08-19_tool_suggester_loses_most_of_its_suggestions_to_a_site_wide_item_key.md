# 321 — the tool-suggester's suggestions collide on a site-wide `item_key`, so ~72% of them are silently dropped

**Filed 2026-08-19** by the `bugfix_275_silent_row_caps` lane, found by the run that confirmed
`bugs_open/319`'s fix. **Same class as `bugs_open/275`** — work is done, then discarded where nothing
can see it — and it sits directly downstream of 275's fix: the whole point of showing the model 80
tools instead of 30 was to get better suggestions, and most of them never become anything.

## The defect

`create_items_loop` files one `add_tool` work item per suggestion, with:

```
"item_key_prefix": "add_tool"          (library forks  -> tool-deployer)
"item_key_prefix": "add_tool_novel"    (novel builds   -> tool-generator)
```

`create_work_item_action.go:234` builds the key as **`<prefix>_<domain>`** — with nothing identifying
*which tool*. Every novel suggestion for a site therefore produces the identical key
`add_tool_novel_gamesdesign.co.uk`, and `idx_swi_dedup` is `UNIQUE (site_id, item_key) WHERE status
NOT IN (…terminal…)`. The second suggestion in the same run collides with the first and **one request
is simply lost**.

**The code already documents this exact failure and ships the remedy** (same function, lines 236-259):

> `'<prefix>_<domain>'` is SITE-wide: two components fixed close together on one site collide on
> `idx_swi_dedup (site_id, item_key)` and **one request is simply lost**. A step that knows what it is
> acting on names it here.

The remedy is the `item_key_suffix_field` config key. **The tool-suggester's two loop steps do not set
it.**

## Measured 2026-08-19 — suggestions made vs work items created

Paired from `llm_call_log.response_text` against `site_work_items` written within 5 minutes of each answer:

| answered | suggested | items created |
|---|---|---|
| 2026-08-19 10:25 | **7** | **1** |
| 2026-08-15 20:29 | **4** | **0** |
| 2026-08-15 18:20 | 1 | 1 |
| 2026-08-15 18:18 | 5 | 3 |
| 2026-08-15 00:01 | **8** | **1** |
| 2026-08-14 23:44 | 6 | 2 |
| 2026-08-12 17:24 | 4 | 1 |
| 2026-08-12 01:26 | 5 | 2 |

**40 suggestions → 11 work items. ~72% lost.** One run lost all four.

⚠ **It is a RACE, not a clean cap** — note the 5→3 row. A colliding key is only blocked while the
earlier item is non-terminal, so whether suggestion #3 survives depends on how fast #1 completed. That
is why the loss rate varies and why no one number describes it. **Non-determinism is the tell that this
is a collision rather than a designed throttle**; a deliberate "one tool per site at a time" rule would
not sometimes let three through.

## Why nobody noticed

- **The loop reports success.** Each iteration's `create_work_item` either files a row or hits the
  unique index; the workflow completes either way and `orchestration_states` reads COMPLETED.
- **The evidence is split across two stores.** The suggestions live in `llm_call_log.response_text`;
  the items live in `site_work_items`. Neither alone shows a gap — you have to pair them, which is the
  query in the table above.
- **1–2 items per run looks like a reasonable answer**, not like a truncated one. It is the same
  "plausible either way" signature as `bugs_open/275` itself.

## Fix candidates, ordered by what closes the door

1. **Set `item_key_suffix_field` on both loop steps** — `current_suggestion.function` is the natural
   discriminator (it is kebab-case, unique per tool, and already in the payload). This is a config
   change on `agent_definitions`, live on apply, and it is what the field exists for. ⚠ **It is a HARD
   ERROR if the path does not resolve** (deliberately — two council seats argued the fallback to a
   site-wide key silently reinstates this very bug), so verify the path against a real
   `current_suggestion` before applying, or every suggestion fails instead of colliding.
2. **Then re-measure the pair.** The disconfirming arm is free and already recorded above: today 7→1.
   After the fix a 7-suggestion answer must produce 7 items, and if it does not, the residue is a
   different defect.
3. **Consider whether 7 tools per site per run is wanted.** This bug says the loss is *silent and
   accidental*; it does not argue that every suggestion should be built. If a throttle is desired,
   make it an explicit one that says so in the item — not a key collision.

## ⚠ Interaction with `bugs_open/275` and `bugs_open/319`

These stack, and the order matters for anyone sizing the work:

- **275** widened the menu 30 → 80 tools (fixed, live, proven).
- **319** raised the answer budget so the model can actually reply (fixed, live, proven).
- **321** is the reason most of that reply still goes nowhere.

Fixing 321 will therefore *increase* the number of tools actually built per run for the first time — 7
rather than 1 on the run measured today. That is the intended behaviour of the pipeline, but it is a
real change in build volume and spend, so it is an owner call rather than an obvious win.

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly.** The mechanism is not inferred — it is written in the
source that implements it (`create_work_item_action.go:234-259`, including the comment describing this
exact collision and naming the remedy), plus the index definition, plus a paired measurement over eight
runs with a varying (and therefore non-tautological) result. Grepped both bug directories: nothing
covers this step or this key.

## Related

`bugs_open/275` (the class, and the fix upstream of this) · `bugs_open/319` (the answer budget, fixed
in the same session) · `create_work_item_action.go` `item_key_suffix_field` (the designed remedy and
the council rounds that made its failure loud) · MEMORY `a-complete-work-item-is-not-a-repaired-artefact`.
