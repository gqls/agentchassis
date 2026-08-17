# HANDOFF 2026-08-17 — continue here

**Lane:** `copy_quality_two_stage`. **State: the build is DONE. Stage 2 exists, is live,
and passed its proof case on its first run. The lane is now waiting on ONE owner decision
(apply the proposal) and otherwise has no build work outstanding.**

Everything the 08-15 handoff called "NEXT WORK" is delivered:

- **Stage 2 (`copy-editor`) is seeded and live** — `sql_for_agents/447_copy_editor_stage_two.sql`,
  register **CQ-024**. **Config-only: no Go, no roll, no council round** (that gate's scope
  is `platform/`/`internal/`/`pkg/`). Live the moment 447 applied.
- **Phase 4 acceptance gates are built and inducible** — `gate_stage2_edit.py`, five induced
  controls plus a dialect control, all six fire.
- **The proof case passed** — orch `18e0d79e`, proposal item **`6dce90f1-bbc7-43b3-a71c-ebfa48cf9afe`**,
  gate exit 0. Full numbers in NOTES 2026-08-17.

## THE ONE THING WAITING ON THE OWNER

**Apply the proof-case proposal, or not.** The six missing guide links on
`loanandmortgagecalculator.co.uk/index` were kept unrepaired since 2026-08-12 by the owner's
own ruling (*"leave it for stage 2 as proof"*). Stage 2 has now proposed the repair and the
gate passes it. **The page is still unchanged** — by D2 the agent cannot write, and it did
not.

The proposal: six `<li>` entries added to the Guides list, **nothing else touched** — no
prose rewritten, no reordering, no markup changed.

To apply once the owner says yes (the `on_approve` → `section-editor` path is declared but
**unexercised**, since it depends on the dashboard `bugs_open/033` is about — so do it by
hand):

```bash
# 1. re-grade first: the page may have changed since 08-17
python3 docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/gate_stage2_edit.py \
        --item 6dce90f1-bbc7-43b3-a71c-ebfa48cf9afe
# 2. file the section_edit item for section-editor, spec shape:
#    {"domain":…, "page_name":"index", "edit_type":"content_edit",
#     "page_component_id":"d6c9198b-1e8c-4299-ac84-57c6a950e0f0", "field_updates":{…}}
#    born 'triaged' (a hand-filed 'detected' row fails silent unless its
#    (type,handler) pair is known-good — SCH-026 holds new pairs for a human)
# 3. prove it at the artefact, not at the status:
python3 docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/gate_page_links.py
```

⚠ **Re-grade before applying.** The proposal was computed against the component as it stood
on 08-17; if a rebuild has rewritten `prose-0` since, the stored `field_updates` would
overwrite the newer copy wholesale (`content` is the component's ONLY field, so the edit is
a full-field replacement). The gate reads the CURRENT row, so a re-grade catches it.

## What is proven, and what only looks proven

| claim | status |
|---|---|
| stage 2 produces a gate-passing edit | **MEASURED** — one page, one run, 0 checks failing |
| it leaves good sections alone | **MEASURED once** — 6 added lines, nothing else |
| the gate can fail | **MEASURED** — 6 controls, all fire |
| the four safety rules hold | **ASSERTED AT APPLY TIME** by guarded `DO` blocks, not by comments |
| the agent cannot write to a page | **STRUCTURAL** — no step in it can; the migration RAISEs if one is added |
| the apply path works | **NOT PROVEN — never run.** `on_approve` → `section-editor` is declared and unexercised |
| it works on a subtler defect | **UNKNOWN.** Six absent links is a legible defect; restraint on a vaguer page is untested |
| `field_updates` narrows blast radius | **VACUOUS HERE** — `ported-prose` declares one field, so it is a full-field write either way |

## Next work, in the order that closes doors

1. **Apply the proof case** (owner's call, above). It is the only thing that exercises the
   write path end to end.
2. **A second page, deliberately chosen to be harder** — a multi-component page with a
   multi-field component, so `field_updates` and the type gate are doing real work rather
   than standing by. The interesting question is whether "leave good sections alone"
   survives when the defect is register rather than a missing set.
3. **Dispatch.** Nothing routes to `copy-editor` today, by choice. Wiring
   `content-quality-auditor`'s findings to it is the obvious next step and is exactly the
   `css-patch-agent` shape the PLAN cites — but it should not be wired until (1) and (2)
   are done, and a new (item_type, handler) pair is held for a human canary anyway.
4. **`bugs_open/033`** (another thread's) still gates ROUTINE operation — a queue nobody
   reads is where proposals go to park. It does not gate one-off proof runs, which is what
   decision 4 established.

## Standing cautions that survive from the last handoff

- **Re-verify the chassis stamp** before trusting instrumented rows. At this writing:
  `v1.0.1305`, commit `6a782274b`, both replicas, binary-probed with a negative control
  (the startup line had already scrolled). Mode-split ancestry holds.
- **LMC:** never fire `run_improvement_sweep_once.sh` (promotes all `detected` rows). Check
  lane activity before any write to their site.
- **Concurrent sessions are fast here.** Re-verify "X does not exist" from these docs
  against the live DB before building X — the whole lane's history is that lesson, and it
  paid again this session (the webdesign case was closed three hours before the last
  handoff called it blocked).
- The capture-only arm harness (`loancalculator_couk/voiceh_rewrite_v3.sh` + `SRC_ITEM`)
  still works and is still the way to test a prompt without dispatching a build. Stage 2
  did not need it: its workflow cannot write, so a live run IS the safe run.

## The five living docs

PLAN (§11 records delivery + three corrections) · NOTES (2026-08-17: the re-verification,
the webdesign correction, the build, the first run) · README_where_we_are (the owner's
plain-prose account, 2026-08-17) · SUMMARY_2026-08-12 / 08-14 / 08-15 / **08-17** (the
series) · this HANDOFF.
