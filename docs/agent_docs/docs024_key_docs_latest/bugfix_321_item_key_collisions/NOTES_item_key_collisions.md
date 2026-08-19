# NOTES — bugs_open/321. Append-only, newest at the bottom.

## 2026-08-19 ~12:00Z — taking the bug; validity + ownership re-checked

- Session was named `bugs_open/321` by the owner but the prompt said 184; 184 is
  actively owned by a live session mid-canary (council-approved 08-18, canaries in
  flight at 10:42 today). Asked; owner chose 321. Did not touch 184.
- Bug re-validated live: `create_work_item_action.go:225-260` builds
  `<prefix>_<domain>`; `idx_swi_dedup` UNIQUE(site_id, item_key) over non-terminal
  rows; tool-suggester's two loop steps lack the suffix. The 10:25 run today:
  7 suggestions → 1 item (pre-fix baseline, on record).
- Fleet census: exactly 6 create_work_item steps inside loops; 4 lack the suffix
  (tool-suggester ×2 live-lossy, component-quality-auditor + internal-linker latent —
  neither has EVER filed an item). tool-auditor's 2 already carry `tool_data.page_id`.
- Discriminator evidence: 60 answers / 239 suggestions — 239 non-empty `function`,
  0 intra-answer duplicates. Same-site repeats across runs: 194 pairs ×1, 16 ×2,
  3 ×3, 1 ×4 → the two-strike brake reaches at most 4/214 pairs. ~10.5% steady-state
  duplicate-suggestion rate; component layer idempotent on function.

## 2026-08-19 ~13:00Z — the two fable reviews, and how each disagreement resolved

- **Latent two: fix now vs defer.** Risk agent said defer (continue_on_error loops
  skip silently on a wrong path). Design agent found both steps already hard-require
  the IDENTICAL path via `spec_paths` — verified live myself: CQA
  `spec_paths.component_id='current_component.component_id'`, linker
  `spec_paths.page_name='current_link.source_page'`. Zero new failure modes ⇒ fix all
  four. **The check that settled it was one query, not a judgement call.**
- **continue_on_error durable or quiet?** Contradiction settled by reading the code:
  `skipToNextLoopIteration` PERSISTS `{loop}_iter_{N}_error` to
  orchestration_states.collected_data (loop_error_handler.go:141-149,185-188) and
  loop_actions.go:505-511 folds status:"error" into `items_created` (a workflow output
  field). Durable. Flag ships with the suffix in one migration.
- **MISSTEP (mine + risk agent's), logged in WRONG_CALLS.md:** "zero snapshot rows
  exist for tool-suggester despite 484 calling snapshot_agent" — FALSE. Snapshots land
  in `agent_definitions_backup`, not `agent_definitions.is_snapshot=true`. Caught by
  reading `pg_proc.prosrc` after MY OWN 493 apply printed 'Snapshot captured' while my
  is_snapshot query returned 0. Cheap check that would have caught it: read the
  function body before asserting its effect is absent.
- Coordination: `bugs_open/313` session confirmed 490 (their internal-linker revival)
  and my 493 touch disjoint subtrees; proceed independently. Their sizing note:
  plan_links asks 1–3 links/plan ⇒ the collision costs up to ~2/3 of that agent's
  entire output from its first productive day. bugfix_275 lane (tool-suggester's 484,
  10:23Z) is closed; my pre-gate asserts their edits survive instead of racing them.
- A fleet-wide touch stamped updated_at=12:14:33 on 199 agent rows mid-research; my
  target fields and 484's edits verified unchanged after it (md5 re-checked).

## 2026-08-19 ~16:05Z — migration 493 APPLIED and proven

- Applied in one transaction: `OK 493: 4 suffixes set, continue_on_error set, 484
  controls intact`. Snapshots verified in agent_definitions_backup (3 rows).
- Induced-failure proof: second apply aborted at the md5 pre-gate
  (`create_items_loop subtree changed (md5 e16bad7c…)`) — the gate fires, and the
  file is not re-runnable after its own apply, by construction.
- Committed `4f6ddbebf` (migration + rollback + lane PLAN).

## 2026-08-19 ~17:20Z — detector built, deployed, and proven in production

- New mode `cmd/config-key-audit --loop-sitewide-item-keys` (loopitemkeys.go):
  WalkSteps single-pass, parent recovered from the qualified path; convicts
  loop-nested create_work_item with prefix and no HONOURED suffix (""/non-string
  convicted, flagged `suffix_declared_but_unhonoured`); acquits top-level,
  no-prefix, loops-over-sites, sub_workflow-under-non-loop. Registry-pinning test
  fails loudly if create_work_item renames/disowns either key.
- Proven by firing: verbatim pre-493 fixture → 2 findings; live export with ALL
  suffixes stripped → exactly the 6-step class; live fleet → 193 agents, 0 findings.
- Wrapper `scripts/audit-loop-sitewide-item-keys.sh`; CronJob 07:55 UTC daily,
  image pattern (Go binary, direct PG). Manual in-cluster run: complete, clean,
  doc_notes row written (source `loop_sitewide_item_key_check`).
- **Plan correction, visible:** my approved plan said "Python mirror + parity test"
  (single-owner-carriers pattern). That is the OLD pattern — shared-output-fields'
  cronjob header records why the image pattern replaced it (a mirror proves the
  mirror). Followed the image pattern; the parity property is carried by the
  registry-pinning test instead.
- Committed (detector + wiring). Makefile carried a named same-file passenger:
  IMAGE_TAG 1311→1315, another session's uncommitted bump matching the live release.
- Fleet release v1.0.1316 (owner, ~20:xx) built/pushed/retagged the check image
  unprompted — RELEASE_IMAGES membership doing its job; verified image exists and
  cluster CronJob repinned.

## 2026-08-19 evening — canary state

- No tool-suggester LLM run since 493 yet (llm_call_log empty after 16:05). The
  20:31 `add_tool_novel_webdesign.co.uk` item is a hand-filed rebuild row (webdesign
  lane), not the suggester loop — old-shape key expected there, not a regression.
- Next: Phase 3 (runtime Warn, platform/ scope → council), then canary dispatch.
