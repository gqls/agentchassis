# NOTES — bugs_open/281 tool audit by instance (append-only, newest at the bottom)

## 2026-08-15 — pick, verify, research

- Picked 281 (`who-owns` → "(none identified)"). Live-transcript symbol grep found one session
  on `check_tool_health` — verified it was the 122 lane that FILED 281 (wrote the file 14:51Z,
  last activity 14:55Z). Re-ran the 30-min recency grep before first Write: 2 hits, both
  listing/LANDMINES noise.
- Bug valid at tree and live: `check_tool_health.go:68` `component_level='tool'`; live
  tool-auditor `load_tool` = seed 088 shape (`WHERE cc.id=$1 … LIMIT 1`, params component only).
- Census `[MEASURED]`: webdesign tool pages = 63 ported-page (section) + 4 tool-level;
  ported-page component `a7daa5c5…` = 115 instances / 2 sites; 0 tool PLANs
  (`doc_notes subject_type='tool' AND categories ? 'acceptance_criteria'`), 89 `needs_criteria`;
  improve_tool fleet: 196 unresolved / 45 complete; clause (b) matches 62 of 63 ported pages —
  the 63rd (`tool-ab-test-calculator`) carries a fork AND a ported instance, counted via (a).
- Three Explore agents' reports (audit machinery / ported+decomposition / governance) are
  summarised in the plan file; the load-bearing findings are restated in TL-042.

## 2026-08-15 — review pass (owner: "check over once more") — what it changed

- **Missed producer.** `check_tool_acceptance.go` already files improve_tool→tool-improver for
  ported instances. `component_versions` on the shared component: v1 (08-05, pre-edit
  snapshot = the 77-char passthrough) and v3 (08-14 18:48Z) both `changed_by=update_component_html`;
  the 08-14 trigger was `tool_acceptance:asset-formatter:<webdesign>` (complete). Bug 281's
  "none exist" is true only of `audit_tool` items. → edit 9 (gate that producer), and the new
  type is named for both.
- **D4 threshold from census** `[MEASURED]`: `update_component_html` has ONE live consumer
  (tool-improver). Multi-placed: 50 section-level (max 299 pages), 2 tool-level (2 pages /
  2 sites; `tool-llm-cost-calculator` rewritten 5× by tool-improver). Both incidents
  section-level → refuse `component_level<>'tool' AND >1 page`; WARN for multi-site forks.
- **section_edit not a viable per-instance fixer today**: owner-gate item spec is
  `{edit_type, page_name, slot_name, field_updates:{}}` — no instruction payload; ported-slot
  section_edit: 1 complete, 2 triaged.
- **Live latent hazard.** Shared template now 8,864 chars of asset-formatter markup
  (`{{.body}}` still present); all 115 instances `build_status='pending'` since 18:48:38Z.
  Verified NOT propagated: every `pc.updated_at` == 18:48:38.131 (the bulk flip), loancash
  `rendered_html` still loan content, pages deployed 08-12; no `component_template_corrupted`
  item since 08-14. Not repaired here (another lane's artefact); recorded in 281 + TL-042.
- **MISSTEP (mine).** First propagation census read "15 loancash instances carry the new
  template" from `rendered_html LIKE '%ported-page-section%'` — a class-name match with no
  control. The very next query (timestamps + a content check) refuted it. Cheap check that
  would have caught it: pair any "did X propagate" LIKE with a positive control on a row known
  untouched. → WRONG_CALLS.md.

## 2026-08-15 — implementation

- Go: `check_tool_health.go` rewritten on `toolEligibilityWhere`/`toolSubjectKeyExpr`
  (instance identity, split-scope cooldowns, `PageID` set, template checks gated by
  `templateIsFork`, ported branch → `ported_tool_fix`, Tier-2 cap 12); `check_tool_acceptance.go`
  ported branch → `ported_tool_fix`, cooldown split-scope; `tool_eligibility.go` header
  rewritten; `update_component_html_action.go` step 1b fence. `escapeJSON` had to stay — two
  sibling checks use it (build told me).
- Tests: `check_tool_health_test.go` (4 tests, incl. the Mind-Map-shaped fixture);
  `update_component_html_shared_fence_test.go` (5 sqlmock tests; **mutation-proven** — with the
  fence condition disabled the refusal test FAILS, restored → green); `verifier_coverage_test.go`
  gap entry + `liveItemTypes`. Whole `discovery_checks` package green; `go build ./platform/...` ok.
- Seeds 425/426: dry-run in aborted txn → pre-flight OK, post-condition OK. **Round-trip
  (apply body + ROLLBACK body, compare to `_before`) caught a real defect in both rollback
  files**: `to_jsonb('literal')` on an untyped string → "could not determine polymorphic type";
  fixed with `::text`; round-trip now `t` for both.
- Register: TL-042 written, TL-033 supersession note, index row.

## 2026-08-15 — commit, seeds applied, council

- Committed `25f92a967` (20 files, pathspec) with `Council-Submitted: 360ae540-8b64-41f9-94da-d7c316183398`;
  gofmt follow-up `a41d11e30`. WRONG_CALLS entry for this lane rode `e96055a03` (bugs_open/282
  session) as a same-file passenger — nothing lost, they flagged it in their message.
- Seeds 425/426 APPLIED 17:17Z (pre-flight re-proved 0 open items without spec.page_id at apply
  time); live rows verified by reading them back (params, gate condition, prompt needle, slot
  path); ledger rows present. Hand-ran the live load_tool query on the shared component with two
  different page_ids → two different tools (mind-map 18,242 chars / asset-formatter 9,222),
  `source_html == rendered_html` for a ported instance, display_name = subject key. The pin works.
- **Council round 1 died `complete_invalid` at `persist_submission` — my schema slip**: four edit
  `file` fields named two paths / carried whitespace ("a.go + b.go", "(+_ROLLBACK)"); the
  validator wants exactly one repo-relative path. Read `__step_error.failed_step` FIRST (RUNBOOK
  says so; it was the plan, not a seat). Fixed the four fields, moved companions into rationale,
  resubmitted under `RESUBMIT_CORR` (run orch `5464d7fd…`). Cheap check for next time: assert
  `re.fullmatch(r"[A-Za-z0-9_./-]+", edit["file"])` before firing — added to the build script.
- Advisory on the commit: "migration + platform code in one commit — needs a staged rollout
  order". Considered: config is live on apply, Go rides the roll; the seed headers state the
  ordering and why it is safe in either order (forks already carry spec.page_id; the fence rides
  the same image as the widening). Additive/opt-in shape → normal council scope per RFC_010 §1 /
  RFC_022, stated in the submission rationale.
