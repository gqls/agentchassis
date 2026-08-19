# CONTRIB from the bugfix_311 lane — RFC_036 §9.3 is BUILT, council-APPROVED and LIVE on v1.0.1316; your Phase D (the 2 blocked tools) is unblocked, and it is the fix's first real test

For `webdesign_tool_rebuilds/HANDOFF_2026-08-19_continue_here.md` §"Phase D" and the line
"RFC_036 §9 — nobody has built it". Written 2026-08-19 20:40 UTC.

**What shipped.** `create_tool_component` now looks up a LIBRARY claim on the function —
`function=$1 AND component_level='tool' AND forked_from IS NULL AND is_active` (the
`idx_cc_tool_function_unique` predicate verbatim, so check and index cannot drift) — and when
one exists the INSERT carries `forked_from = <library id>`, so the partial index no longer
fires and `save_tool` completes. No claim → byte-identical to before. Lookup read error →
today's loud index refusal (fail-open, never a silent guess). Code:
`platform/orchestration/actions/create_tool_component_action.go:249-285`. Commit
`e24bc9c0f`; council `ceae30f2` APPROVED round 1; the owner ruled it half of a
precondition pair with 311's section fix (`17d883333`, also live).

**Proof it is in the binary you are dispatching at** (measured 20:35Z, not inferred from the
tag): both v1.0.1316 replicas stamp `07eeba4a1…`; `git merge-base --is-ancestor e24bc9c0f
07eeba4a1` is TRUE; the literal `library tool claims this function` is PRESENT; a fake sha
is ABSENT on the same pod. (Your handoff says "verify by DIGEST, not a binary probe" — the
digest proves the IMAGE changed; only ancestry of the stamp proves THIS commit is in it.)

**What your Phase D run should assert, because it is the first real exercise of this code:**
1. A NEW row, `component_level='tool'`, name `tool-ab-test-calculator-webdesign-co-uk`,
   **`forked_from = '8c9a6e06-e2b2-4f21-baf6-651585375f0c'`** (the library row).
2. `save_tool` COMPLETES — no SQLSTATE 23505 in the item `error` or the child orchestration.
3. The Info log line `CreateToolComponentAction: library tool claims this function — new
   component forks from it (RFC_036 §9.3)` on the chassis pod, `function=tool-ab-test-calculator`.
4. The library row `8c9a6e06` UNTOUCHED: html md5 `8673be08f969504f5a9ceb46e45d7656`, schema
   md5 `688e1188b91ccef0674cd527daa05ec3`, `updated_at` 2026-05-06 (pinned 20:38Z).
5. Your usual artefact checks (`{{\.` 0, `onclick=` 0, the page gains the slot).

**One thing to read before you fire it:** the action's "already exists" check
(`create_tool_component_action.go:228-237`) returns `already_exists` early if an ACTIVE
tool-level component with function `tool-ab-test-calculator` is linked to any page of the
site — so whether the ported slot is still active at build time decides whether the fork
branch is reached at all. Today's rows for that function: library `8c9a6e06` (active),
forks `cd60486c` (…-webdesign-co-uk, INACTIVE) and `58da6570` (…-idea-uk, active).

**Known follow-up, tracked not hidden (RFC_036 §11):** `deploy_tool_to_site`'s existing-fork
lookup matches `forked_from=$1 AND name=$2` with ITS name shape, so a later deploy of the same
library tool to the same site would not recognise the generator's fork and would fail loudly at
`pages_site_id_name_key` — not silently. Not in scope for your run; if it bites, that is the
signal to do the §11 widening.

Please report the outcome (pass or fail) into
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md` or in your
own NOTES with a pointer — it closes the owner's precondition pair. I did not fire it from my
lane: your session is actively dispatching at webdesign.co.uk (`page_rerender` items 20:36Z)
and the trigger takes one site per tick.
