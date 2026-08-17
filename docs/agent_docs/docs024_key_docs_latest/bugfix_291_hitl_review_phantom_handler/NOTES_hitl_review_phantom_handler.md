# NOTES — bugfix 291 (append-only, newest at the bottom)

## 2026-08-17 — session start, verification, and design

- Bug 291 taken on. Ownership checked: `who-owns.py 291` → filed by the 284 lane, which
  closed 284 the same morning ("no handoff needed"); live-transcript grep (the lagging-check
  memory) found the only heavy `hitl-review` session WAS the 284 lane. Unclaimed. Queue empty.
- `[MEASURED 12:27Z]` 14 rows `status='blocked' AND error='Handler agent not registered:
  hitl-review'`, 2 sites, newest 06:55Z. Same count re-measured 12:04Z pre-090.
- `[MEASURED]` `hitl-review` appears at exactly ONE live-config path fleet-wide:
  `tool-auditor.default_config.workflow.steps.create_items_loop.config.sub_workflow.steps.create_review_item.config.handler_agent`
  (recursive jsonb walk over the live row).
- **CORRECTION to the bug file** `[MEASURED]`: its "Measure before choosing" line says the
  other `needs_human_review` rows "carry an empty handler". They do not — the 7 non-291 rows
  all carry `handler_agent='human-review'` (checkpoint producers: tool-recreation-handler,
  image-url-404-handler, generic). The EMPTY-handler idiom is real but lives in the Go
  discovery checks (`check_unverified_claims.go`, `check_voice_tells.go`) and in the fleet
  census at `refresh_evidence_fact_drift.go:698-703` (544 `''` vs 22 `'human-review'`).
  Correction to be written into the bug file visibly.
- **CORRECTION to the bug file**: `resolve_composition_layout_action.go:390` is NOT "the same
  mistake waiting to fire" in the dispatch sense — :391 sets `status:"needs_human_review"`
  explicitly, so its items are never claimed. The wrong HANDLER value spread by copying; the
  safe STATUS did not. The bleeding difference is tool-auditor's missing status key.
- Root cause pinned: `create_work_item_action.go:208-211` defaults status to `'triaged'`;
  seed `088_tool_auditor_agent.sql:158` names the handler, sets no status; no later migration
  (349/350/425/434) added one. `hitl-review` documented 2026-04-19 as "a convention, not a
  registered agent" — the handler was a roadmap row never built.
- Design stress-test (Plan subagent) found the FATAL flaw in the first draft: the live binary
  refuses empty `handler_agent` in `create_work_item` config (:184-187 hard error), so a
  config flip to `''` today would silently lose every finding under `continue_on_error:true`.
  Verified first-hand at :184-187 before accepting. → three-phase plan (see PLAN).
- Also from the stress-test, verified against the tree: migration slot 446 already taken
  (tool-suggester lane) → ours are 447/448; the INSERT widening trap at
  `load_work_item_actions.go:1396-1408` (conditional-append idiom is mandatory for the new
  `error` column); ~25 sqlmock test files expect the shared INSERT and the guard's probe
  needs adding where their paths now see it.
- `[MEASURED]` 13 of the 14 rows carry `spec->>'check'='tool_auditor'` + `spec.issue`; the
  2026-08-14 finetuning.uk row predates migration 434's spec rewrite and lacks both (and its
  item_key has no page-id suffix — page_id was empty). Repair keys on created_by + exact
  error text, NOT spec.
- All 14: `attempt_count=0`, `claimed_at IS NULL`, `approval_mode='auto'` — claim's
  not-registered branch nulls the claim fields, so the repair need not touch them.
- tool-auditor is driven as a work-item handler by the dispatch loop (~hourly runs observed
  2026-08-17 06:15→10:27 on webdesign.co.uk); the bleed is live, hence config-first phasing.
- 090 filed 12:03Z: `RUN_CORRELATION_ID=3555b514-ca8f-4f31-9f55-e105ce73e961` (dispatch-loop
  correlation, the artifact key). Verdict pending at time of writing.
