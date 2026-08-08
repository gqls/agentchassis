# NOTES — bugfix 220 (append-only, newest at the bottom)

## 2026-08-08 ~18:45Z — lane opened, ownership swept

- who-owns 220: commits are the filing (206 lane) + a contrib (116 lane). Transcript
  grep for FIX-SITE symbols (`load_page_record|availableBuilders|check_phantom_internal_links`),
  not the bug number: b5a58a2b = 206 lane, closing message says 220 is a follow-up it
  is NOT taking; 63f97914 = fragments lane (reads 220, works dead_fragment_link);
  005ad3d6 = 201/RFC_017 lane — touched `complete_work_item_verification.go` TODAY
  (fail-closed flip, committed). Conclusion: free. [MEASURED — greps + tail reads]
- Re-verified the three mechanism legs live (see PLAN § Validity). The vetcomparison
  instance itself is healed (206 lane built directory-index 17:02Z) — do NOT re-use it
  as the acceptance case without re-checking for a fresh unbuilt target.

## 2026-08-08 ~19:00Z — the finding the bug file does not state

`load_page_record_action.go` resolves page_name BEFORE page_id (`:7-10`, `:174`).
Candidate 1 as written in the bug file ("map page_id and have load_page_record prefer
it") would be INERT: spec.page_name (container) is always present for this item type,
so the name always wins. Confirmed by reading the live step config:
`page-build-handler.load_page_record.config` = `{page_id: input_data.spec.page_id,
page_name: input_data.spec.page_name, site_id: site_record.site_id}` — note BOTH
config paths point at the CONTAINER for this item type (spec.page_id is the container;
only the COLUMN carries the target). Fix shape per RFC_010 §2: opt-in field
`authoritative_page_id`, default absent = today's behaviour.

- Precedent found for the mapping half: `site-work-orchestrator.call_handler` already
  maps `"page_id?": "current_fix_item.page_id"`. [MEASURED — live agent_definitions]
- Zero live agents map `current_item.page_id`; `input_data.page_id` is read only by
  `page-retraction` + `deduplicate-sections`, both with 0 work items ever. [MEASURED]
- Only build-dispatch-loop maps `current_item.spec.page_name` (jsonb_path query over
  live rows). The sibling loop does not share the defect. [MEASURED]
- `ExtractActionInputs` ignores undeclared config keys at runtime (declared-field
  iteration; UnknownConfigKeys is the offline audit) → no image/config ordering
  constraint. [VERIFIED — read action_inputs.go:198-247]

## 2026-08-08 ~19:30-20:00Z — implemented, tested, submitted

- Go: `authoritative_page_id` on load_page_record (body refactored into shared
  `queryPageRecordRow` so the id path cannot drift from the name path);
  `VerifyUnbuiltInternalLinkResolved` + registration; coverage-map entry removed.
- A guard I did not know existed caught the missing half:
  `TestRegisteredVerifiersMatchClaimTimeoutExclusion` — registering a verifier obliges
  the claim-timeout lockstep (declared list in sql_for_agents/220 + live column). Done
  both halves (220 edit same commit, mig 341 on 331's template; live list read first:
  8 entries, no drift). [VERIFIED — test failed, then green]
- Tests green on the dirty tree AND against a clean `git archive HEAD` overlay (the
  tree carries another session's WIP in this same package). [MEASURED]
- MISSTEP (full entry in WRONG_CALLS.md 2026-08-08): ran `landmines-sync.py --apply`
  directly after appending the LANDMINES entry — the documented wrapper
  `landmines-verify-dispatch.sh` should have been used; recovered by hand-firing
  `trigger-landmine-verifier.sh`, correlation `f70fb3af`. Check the verdict later:
  `SELECT created_at, left(body,120) FROM doc_notes WHERE subject_key LIKE
  'LANDMINES.md#loadpagerecord%' AND categories ? 'landmine-verification';`
- Council: submitted r1, correlation `def4441c-df3a-460a-b2ce-208da04f4023`
  (submission JSON in this dir). Committing with `Council-Submitted:` per the
  2026-07-30 trailer rule; budget ~30 min for the verdict, find the run by payload
  not by printed id.
