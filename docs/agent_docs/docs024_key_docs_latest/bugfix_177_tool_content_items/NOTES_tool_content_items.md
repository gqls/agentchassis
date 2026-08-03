# NOTES — bugfix 177 (append-only, newest at the bottom)

## 2026-08-03 ~11:35 — lane start, ownership, validation

- Session was named "bugfix 182" and pointed at 182 first. **Yielded**: live
  transcript `28533a86` was at the exit-plan-mode step of implementing 182's fix
  (re-measured the 78 slots itself, plan names `loadComponentSchemasByID` +
  the 095 treatment). Symbol grep, not bug-number grep, found it — the
  who-owns verdict line only named the FILING lane.
- Picked 177 by ascending reference-heat over 34 transcripts (last 4h):
  177 = 23, floor ~22, ceiling 402. Symbol grep on `tool_content:` found only
  the filing session (closed, last night) + two cold readers.
- Bug still valid, and WORSE than filed: 8 needs_human_review + 1 wont_fix
  (was 8 total at filing — `a5cabea0` minted 08-02, tool-cma-obligation-checker,
  vet site), and a5cabea0 has TWO triaged `content_rewrite` dependents
  (`9e9ec430`, `18bc832c`) — the 176 blocker shape, armed again.
- Root cause verified (bug file had it `[UNVERIFIED]`): create path declares no
  `pages.sections`; deploy path declares 4. All 9 items are the create path's.
  Control: `tool_guide` items (declared sections, same files) = 4 complete.
- Handler resolution read from live `agent_definitions` (page-build-handler →
  `load_spec_sections` → `load_page_sections_from_spec_action.go`): plan tables
  → spec aspect → pages.sections → sibling synthesis (plan-membership-gated).
- Edge case measured before designing the guard: 33 current-plan section rows
  for tool-named pages exist ⇒ the guard must consult plan sources, not just
  `pages.sections`. A parameter-trusting guard would wrongly skip those.
- The one deploy-path page with the full declared shape
  (`tool-ai-agent-roi-estimator`) is from 2026-04-23 — pre-dates the current
  work-item era; 1 slot; its item history is gone. Treated as no evidence
  either way about the handler building prose around tools today.
- **The no-op class is WIDER than 177's scope**: `error LIKE '%no sections
  ready to build%'` also matches 24+ `needs_page` rows from `image-build-handler`
  (11), `reconcile_site_plan` (9+), `page-rerender` (2), `json-leak-fix`,
  `gemini-p7-verification`. Different item type, different emitters, possibly
  legitimate deferrals — NOT swept, NOT fixed here (one coherent bug per fix).
  Named in the 177 close-out for a follow-up decision.

## 2026-08-03 ~12:00 — mid-implementation events

- Council submitted, `SUBMISSION_CORR=982507b0-2e18-4457-a354-85a809012bbd`
  (~11:55 BST). Budget ~30 min. SEQUENCING: the chassis roll must WAIT for the
  verdict — a roll kills an in-flight council run (kubectl-scale memory).
- Opus implementer dispatched with the final design (insertWorkItem routing per
  the gapPlanWorkItem precedent — the first draft's hand-rolled INSERT was
  revised away after the a5b70424 objections were found in research).
- **/tmp (16G tmpfs, shared by every session) hit 100% full mid-implementation.**
  Top consumers: other live sessions' scratchpad HEAD-check build contexts
  (447MB each; one session held 3.5GB of them). Not ours to delete. Mitigation:
  messaged the implementer to use `GOTMPDIR=/home/ant/tmp` and keep the
  headcheck off tmpfs. Appended the trap to LANDMINES.md (### heading — the
  sync REJECTS `##` entries with "cost DELIVERY"; six older entries in the file
  have that defect and were left for their lanes) and ran
  `landmines-sync.py --apply` clean.
- Register edits done (TL-003 conditional emission, TL-009 partial-shipped
  note); 016b §9 entry inserted ("unsatisfiable at birth"); RUNBOOK written.
