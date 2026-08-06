# HANDOFF 2026-08-06c — both fixes LIVE at v1.0.1259, 189 behaviourally PROVEN; ONE gated command closes both

Supersedes `HANDOFF_2026-08-06b_…` (written before the v1.0.1259 roll). Standing
owner brief: `HANDOFF_2026-08-05_next_bug_pickup.md`.

## The single thing left

The auto-mode classifier blocks writes to `agent_definitions`, so an **operator
must run one command** (`!` prefix puts its output in the chat):

```
! kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < <(tail -45 docs/agent_docs/sql_for_agents/023_page_content_writer_agent.sql)
```

It sets `slot_name_from: "current_section.name"` on `page-content-writer`'s
`render_section` and `render_from_template` steps. **Until it runs, 189's BUILD
half is inert and both bugs must stay open.** It is self-verifying — a
`DO $verify$ … RAISE EXCEPTION` refuses the COMMIT unless both keys read back
(expect `NOTICE: slot_name_from present on both render steps`). A validated
JSON backup was taken first (20,339 B, both keys confirmed absent):
`scratchpad/pcw_default_config_backup_20260806.json`; rollback is the same
UPDATE from that file. Do **not** re-run the seed's full-workflow UPDATE block —
it is stale against live and would revert a later `prompt_template` patch that
exists only as a pasted transcript.

Then: confirm both keys, fire 204's canary on the BUILD path
(`guide-how-loans-are-calculated`, per 204 §How to verify — prose must change,
zero `needs_new_component` filed), and record the closure evidence in **both**
bug files.

> **⚠ CORRECTED 2026-08-06 — do NOT move either file to `bugs_closed/`.**
> An owner direction of the same day (recorded in memory as
> `owner-keeps-fixed-bugs-in-bugs-open`) says *"please leave the bugs that
> you've found in bugs_open not in the closed bug file"*, given after another
> thread moved `126` and `181` on the fixed-AND-live bar and had both directed
> back (`669ca58c5`). **This overrides CLAUDE.md's `/bugs_closed/` bar and it
> overrides the earlier handoffs in this series, including 06b, which tell you
> to `git mv`.** Finish the work normally — fix, council, pod-verify, induce —
> write the evidence into the file as a dated section, and leave the file where
> it is. Ask before moving any bug file until the owner says the general rule is
> back.

## State

**`bugs_open/204`** (build path could not resolve a positional slot name) — fix
`13252f714`, council APPROVED (`d3e232b8`). **LIVE since v1.0.1257.** Verified:
pod-grep of three added strings on one pod per chassis-image deployment with a
fabricated-string control at 0; plus read-only proof that **57 of 57**
loancalculator sections unresolvable by name ARE resolvable by the stored-id
route. Open only for the build-path canary, which needs the config above.

**`bugs_open/189`** (a resolving save renames a positional slot, so the
locked-row guard misses and duplicates it) — fix `92e14493b` + `1d11827c1`,
council APPROVED (`87444080`), registered **PBP-035**. **LIVE at v1.0.1259 and
BEHAVIOURALLY PROVEN on the re-render path** (full table in the bug file):

- `stored_slot_name` greps 1 on all three deployments; it was **0** at
  v1.0.1257, so that measured zero is the negative control.
- Induced `section_data_resolved` on `tool-loan-vs-savings` (item `b4de13fb`,
  orchestration `b807b035`): **4 rows, not 5**; slot names `prose-0, prose-1,
  tool-2, prose-3` unchanged (the old code renamed the prose rows to
  `ported-prose`); locked `tool-2` kept row id `10be4f71` **and** its
  2026-08-02 `updated_at`; served page still 4 `<section>` blocks.
- **Proof the save actually ran** (a no-op would look identical): the three
  unlocked prose rows have NEW row ids stamped `11:36:54`, and the action
  reports `rerendered=4, carried=0`.

Open only because the BUILD route is still armed without the config.

## Method notes worth keeping

1. **A green row count can be a no-op.** 189's pass condition (4 rows, names
   intact) is satisfied equally by "the fix worked" and "nothing happened". The
   discriminator was **row identity** — DELETE+INSERT gives new ids — plus the
   action's own `rerendered`/`carried` counters. Always pick a check that
   distinguishes the fix from inaction, not just from the damage.
2. **`approved` is a dead status for `page_rerender`.** Filing the work item
   did nothing; every other recent item went straight to `complete` because its
   creator dispatched it. Filing is not dispatching — publish to
   `system.agent.generic.requests` and confirm at the DB, because `kcat -P`
   exits 0 having sent nothing.
3. **`spec.reason` selects the branch.** Without `reason:
   section_data_resolved` the rerender takes the assemble-only path and never
   reaches the save, so it cannot grade this bug.
4. **Pods were 46 minutes old** — checked deliberately against the ~300s
   post-restart window in which a dispatch is silently dropped. Worth checking
   every time, and especially right after a roll.
5. **A verification item finishing `blocked` is not a failure** — it is the
   platform recording that a lock prevented a full overwrite. Grade the rows,
   not the item status.

## Earlier lessons from this arc (carried forward)

- **Read the verdict you cite.** I called 182's semantics "council-reviewed";
  only its SUBMISSION exists — no `council_report` for corr `80fbbe7d`. The
  guardian seat caught it. In WRONG_CALLS, with the one-query check.
- **"Sole consumer" is a query, not a memory.** `plan_sections` has **two** live
  consumers, not one.
- **After making a dead path live, trace where it terminates.** 204's fix armed
  189's trap on a second path. Found by tracing compile→save after the council
  round.
- **A pathspec commit cannot exclude a same-file edit, either direction.** My
  `v3_site_actions.go` half of 189 was swept into another session's
  `1d11827c1`; nothing lost, and both commits say so.
- **The trailer gate is right.** `Council-Submitted: pending` was refused —
  the trailer is a join key and forward-only forbids an amend. Submit first.
- **Standing architecture question** (`doc_notes d9d67807`): the tri-state
  id-resolution judgement is now inline at two call sites. A THIRD needs a
  shared helper first.
