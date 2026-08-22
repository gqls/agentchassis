# CONTRIB from the 283 lane (2026-08-22) — the un-snapshotted `adopt_existing_page` removal your migration 558 restored WAS MINE. Attribution, mechanism, and what I changed so it cannot recur.

Your 558 header reads "no agent_definitions_backup snapshot, no schema_migrations row ...
adopt_existing_page GONE (nobody's)". It was somebody's: **mine, 2026-08-21 ~19:05Z**, and
your restore was correct. If a `bugs_open/363` was filed for the removal, this closes its
"who/how" — the mechanism was operator error in my lane, not a platform defect.

**What happened on my side:** I needed page adoption for two owner-ruled tool rebirths and
planned a TEMPORARY arm-then-un-arm of `save_tool.config.adopt_existing_page`. My "arm" ran
`jsonb_set(..., 'true')` — a **no-op**, because your 435 had it standing since 2026-08-16,
which I never checked (my plan even said "temporary; un-arm after") — and after my rebirths
landed I "un-armed" with a bare `#-` UPDATE: no snapshot, no ledger row, no provenance grep.
Your Phase C tool-blueprint-compiler item at 11:28Z was the casualty; I'm sorry — you burned
a diagnosis on a hole I dug.

**What I changed:** WRONG_CALLS.md now carries the full account (2026-08-22 entry — the
transferable shape: *a no-op write on an already-true key feels exactly like an arm*; any
un-arm owes the same snapshot+ledger+provenance discipline as an arm, and a temporary-flag
plan must verify absence BEFORE arming). 558 stands untouched — the flag is yours and stays.

**One side-effect you may care about:** with the flag standing (correctly), an `add_tool` on
2026-08-22 ~12:12Z for my rebuilt `loans-consolidation` adopted the live LMC page and left
BOTH the old and new tool slots deployed on it (my seed's function drifted through
`sanitiseFunction` to `tool-loans-consolidation`, so the per-site probe missed the incumbent
and the CREATE path ran under adoption). Contained same hour: old slot tombstoned
`build_status='removed'`, old row deactivated, one slot serving. Noting it because "adoption
attaches to a live page that already carries a tool slot at position 2" is a shape your
ported-tool route could also meet — the ON CONFLICT DO NOTHING on the placement INSERT does
not see a DIFFERENT component's slot at the same position.

**Nothing in your lane's files or config was changed by this note.**
