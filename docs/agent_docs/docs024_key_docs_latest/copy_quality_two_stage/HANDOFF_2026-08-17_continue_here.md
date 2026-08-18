# HANDOFF 2026-08-17 — continue here

**Lane:** `copy_quality_two_stage`. **State: DONE and APPLIED. Stage 2 exists, is live,
passed its proof case on its first run, and — owner-approved the same evening — the edit
has been applied to the live page and verified at the artefact. The lane has no
outstanding build work. **2026-08-18: the harder page has now been run too — it found two
real mechanism faults (both fixed) and left ONE proposal awaiting the owner.**

> **UPDATED 2026-08-17 evening — the proof case is CLOSED.** `gate_page_links.py` now exits
> 0 ("all 16 required links present"); the six guides have had a homepage link since
> 21:52:55Z. Applied via `section_edit` item `e02190da` → orch `b747199d` → COMPLETED
> 21:53:18Z. Every must-not-change item held (H1, 31 class attrs, 12 cards, 2 tool-grids,
> no banned phrase; words 629 → 669). Verdict banked in
> `loanandmortgagecalculator_couk/acceptance/RESULT_2026-08-17_index_link_gate_PASSING.txt`
> beside the 08-12 FAILING baseline. Both work items closed with evidence on the row.
> **The write path is proven end to end. The `on_approve` → `section-editor` handoff is
> still NOT** — approval today is a human running two commands, and `bugs_open/033` is why.

> **UPDATED 2026-08-18 — RUN 2 (the harder page) BROKE STAGE 2 TWICE, both faults now fixed.**
> `ai-agent-orchestration.com/index` (8 components, 78 KB, a DIFFUSE register fault) made the
> agent attempt a whole-page rewrite; it truncated at `stop_reason=max_tokens`. **So
> 08-17's restraint was a property of a LEGIBLE DEFECT, not of the design** — the table
> below is corrected accordingly. Fixed by migration **462**: a THREE-EDIT BUDGET plus
> `max_tokens` 32,000, because an edit set bounded at the source cannot truncate. Re-run
> passed (8,181 tokens, 3 edits) and its judgement is what page-scoped read exists for —
> the same pitch restated in FIVE sections and one resource under FOUR names.
> **My own gate then failed that proposal wrongly, twice** (a volume floor that could not
> tell de-duplication from gutting; array fields unchecked beyond their type while
> reporting "1 of 1 type-checked"). Both fixed, controls re-proven. Full record: NOTES
> 2026-08-18. **Run 3's proposal `d2378b77` is PARKED and needs the owner** — its edits
> DELETE live copy, which is the class of change a human should see first.

Everything the 08-15 handoff called "NEXT WORK" is delivered:

- **Stage 2 (`copy-editor`) is seeded and live** — `sql_for_agents/447_copy_editor_stage_two.sql`,
  register **CQ-024**. **Config-only: no Go, no roll, no council round** (that gate's scope
  is `platform/`/`internal/`/`pkg/`). Live the moment 447 applied.
- **Phase 4 acceptance gates are built and inducible** — `gate_stage2_edit.py`, five induced
  controls plus a dialect control, all six fire.
- **The proof case passed** — orch `18e0d79e`, proposal item **`6dce90f1-bbc7-43b3-a71c-ebfa48cf9afe`**,
  gate exit 0. Full numbers in NOTES 2026-08-17.

## ~~THE ONE THING WAITING ON THE OWNER~~ — DONE 2026-08-17 evening (kept: the recipe is reusable)

**The owner approved, and it is applied.** What follows is the procedure that ran, which is
the same procedure for the NEXT proposal — the review queue still has no surface, so every
approval is a human running these steps.

**Two corrections earned in the doing, both already applied to the tooling:**

- **`--item` was broken and this handoff recommended it.** It read a flat
  `review_data.page_component_id`; a real proposal nests `review_data.edits[]`, one entry
  per component. It would have failed for anyone who followed this file. Fixed in
  `8a45a3e3c` — found by USING it, not by reading it.
- **File `triaged`, then claim it yourself before direct-publishing.** `section_edit` claim
  latency is avg 1,695 s with a 21,757 s tail (n=172), so waiting is not viable inside a
  session; claiming first is what stops the loop double-dispatching a payload already in
  flight. Check `ai_endpoint_health.healthy` first — a stopped queue reads green everywhere
  else (LANDMINES, 2026-08-17, webdesign lane).

**What the case was** (kept, because it is the worked example): the six missing guide links
on `loanandmortgagecalculator.co.uk/index` were held unrepaired from 2026-08-12 by the
owner's ruling (*"leave it for stage 2 as proof"*). Stage 2 proposed six `<li>` entries added
to the Guides list and **nothing else** — no prose rewritten, no reordering, no markup
changed — the gate passed it, the owner approved, and it went live the same evening.

**THE RECIPE, for the next proposal** (the `on_approve` → `section-editor` path is declared
but **unexercised**, since it depends on the dashboard `bugs_open/033` is about — so this is
by hand):

```bash
# 0. is the queue even claiming? a stopped queue reads green everywhere else
#    SELECT healthy, last_checked FROM ai_endpoint_health WHERE name='claude';
# 1. STALENESS, then re-grade against the LIVE row (never the payload's own "before")
#    SELECT pc.updated_at > swi.created_at AS payload_is_stale FROM site_work_items swi
#      JOIN page_components pc ON pc.id=(swi.spec->'review_data'->'edits'->0->>'page_component_id')::uuid
#     WHERE swi.id='<review item>';
python3 docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/gate_stage2_edit.py \
        --item <review item>
# 2. file the section_edit, spec COPIED BY SQL from the proposal row (never retyped):
#    {"domain":…, "page_name":…, "edit_type":"content_edit",
#     "page_component_id":…, "field_updates":{…}}
#    born 'triaged', then claim it yourself (UPDATE … status='in_progress',
#    claimed_by='<session>') before direct-publishing, or the loop may dispatch it too
# 3. publish to section-editor on system.agent.generic.requests (input_data needs
#    domain, site_id, page_id, page_name, slot_name, page_component_id, edit_type,
#    field_updates, work_item_id). kcat -P exits 0 having sent NOTHING — the
#    orchestration row is the only proof of dispatch
# 4. prove it at the artefact, NOT at the status: check_edit_skipped routes a lock- or
#    decision-gated REFUSAL to 'complete' as well, so 'complete' alone means nothing
python3 docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/gate_page_links.py
# 5. close both items with the evidence on the row — nothing in the platform will
```

⚠ **Re-grade before applying, always.** A proposal's `field_updates` is a full-field
replacement frozen when it was written; if anything rewrote that field since, applying it
reverts the newer copy and every status still reads success. On a single-field component
(`ported-prose` declares only `content`) that means reverting the WHOLE component. The gate
reads the CURRENT row, so a re-grade catches it. Now a LANDMINE entry in its own right.

## What is proven, and what only looks proven

| claim | status |
|---|---|
| stage 2 produces a gate-passing edit | **MEASURED** — one page, one run, 0 checks failing |
| it leaves good sections alone | **REFUTED as a general claim, 2026-08-18.** True on a legible defect (6 added lines); on a diffuse one it attempted the whole page and truncated. Now BOUNDED at 3 edits by config (462) rather than trusted |
| the gate can fail | **MEASURED** — 6 controls, all fire |
| the four safety rules hold | **ASSERTED AT APPLY TIME** by guarded `DO` blocks, not by comments |
| the agent cannot write to a page | **STRUCTURAL** — no step in it can; the migration RAISEs if one is added |
| the apply path works | **PROVEN 2026-08-17** end to end: proposal → gate → `section_edit` → render → deploy → six links live. ⚠ but the `on_approve` → `section-editor` HANDOFF is still unexercised — a human filed the item |
| it works on a subtler defect | **YES, once bounded** — run 3 found five-fold cross-section restatement and one resource under four names, and proposed 3 removals. ⚠ unapplied: the owner has not reviewed it |
| `field_updates` narrows blast radius | **REAL on run 3** — 3 named fields across 3 components, one of them an `array` (`features`, 10 items → 9), leaving 58- and 22-field components otherwise untouched |

## Next work, in the order that closes doors

1. ~~**Apply the proof case**~~ **DONE 2026-08-17** — gate exits 0, six links live.
2. ~~**A second page, deliberately chosen to be harder**~~ **DONE 2026-08-18 — and it found two mechanism faults.** What remains of it: the owner reviewing `d2378b77`, whose three edits DELETE copy.
3. **A THIRD page, to test the 3-edit budget's ceiling** — run 3 said five sections restate each other and edited three of them, so the budget is now a known under-fix on a page like that. Does a second pass on the same page find the remaining two, or does it re-propose the same three?
4. ~~old item 2~~ **superseded** — a multi-component page with a
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

- **Re-verify the chassis stamp** before trusting instrumented rows — it moved between the
  two halves of this file. ~~`v1.0.1305` / `6a782274b`~~ → **2026-08-18: `v1.0.1308`,
  commit `e7e5e4d53`**, binary-probed on `-dvscb` with a negative control (the startup line
  had already scrolled). Mode-split ancestry still holds. **A stamp in a handoff is a
  dated observation, never a current fact.**
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
