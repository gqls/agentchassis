# HANDOFF — 079 platform fix, candidate 1: repair at the persistence point

**Written 2026-07-28 evening (brochure thread, owner-directed). Status: DESIGNED, NOT
IMPLEMENTED.** This is a cold-start document — everything an implementing session needs
is here or one hop away. Read `bugs_open/079_HANDOFF_2026-07-26_…md` REOPENED banner
first (mechanism + evidence), then this file. All file:line facts below were verified
against the working tree on 2026-07-28 (branch `087_towards_multiple_domains`, pushed).

## Session-start checks (do these before writing code)

1. `git log --oneline -10 -- platform/orchestration/actions/save_page_sections_action.go
   platform/orchestration/actions/validate_page_content.go` — another session may have
   moved. `scripts/who-owns.py 079` too (it reads commits; a mid-fix session is
   invisible, so also `git status` the two files).
2. Re-verify the two load-bearing line references below with a quick grep — the tree
   moves ~1,500 commits/week; the SHAPES are stable, the numbers may not be.
3. Check the Anthropic lane is alive before planning the council round (it died and was
   revived twice on 07-28): recent `agent_error_log` rows with `LLM_API_ERROR`.

## The defect, one paragraph

`validate_page_content` repairs dead in-body links in `clean_html` and returns it
(`validate_page_content.go:355-358`), but on the primary build plan `save_page_sections`
takes the structured path — `sections_metadata_field` (`save_page_sections_action.go:166`)
— and reads `html_field` only when the metadata yields zero sections (`:190`). So the
repaired string is dead config whenever metadata is present, which
`require_sections_metadata: true` makes always. The repair is COMPUTED and durably logged
(`CONTENT_LINK_REPAIR_DETAIL`) and never SAVED. Proven on three sites; 090 verification
verdict UNVERIFIABLE (no refutation), corr `954d8da9`.

## The fix

**Apply `datahelpers.RepairPageLinks` to each section's HTML inside
`SavePageSectionsAction`, immediately before the guards/insert.** Persistence becomes the
enforcement point: no build path can save an unrepaired section, whatever the workflow
config says.

**Why here and not in validate (candidate 2):** `page-rerender`'s structured save path
(`034_page_rerender_agent.sql:197-202`, `sections_metadata_field:
rerender_sections.sections_metadata`) has **no validate step at all** — a gate-side
repair can never cover it. The save action covers every body-section persister:
page-build-handler, page-rerender, page-rebuild, site-work-orchestrator,
pageflow-builder, tool-recreation-handler. It also cleans stored HTML that the
interactive-tool preservation block carries forward (`:286-357` can inject stored DB html
at `:339`) — which is exactly how gamesdesign's stored `href=""` finally gets repaired.

### Where, precisely

In `save_page_sections_action.go`, AFTER `enrichSectionsWithPlannedNames` /
`enrichSectionsWithComponentIDs` (`:236-239`) AND after the interactive-tool preservation
block (`:286-357`), BEFORE the content-regression guard (`:363`) — so the guards measure
what will actually be persisted. Unlink keeps inner text, so the stripped-text
measurements barely move.

### Building blocks (all already exist, same package)

- `loadValidPagePaths(ctx, db, siteID, logger) (datahelpers.PageURLIndex, bool)` —
  `validate_page_content.go:1129`. Fail-open `(nil, false)` semantics already correct
  (empty-set-means-everything-is-a-phantom is the catastrophic case; the bool means
  "trustworthy"). `siteID` is parsed at `:109` of the save action; `params.DB` is
  guaranteed non-nil past the `:46-54` early skip.
- `datahelpers.RepairPageLinks(html, index) (string, []LinkRepair)` —
  `datahelpers/link_repair.go:139`. Fragment-safe: pure string transform, byte-identical
  when clean, touches only page-scope/empty hrefs. Its `data-runtime-fill` exemption is
  whole-document by comment but the marker is emitted per-section in practice — applied
  per-section it exempts exactly the runtime-fill sections (a narrowing; correct).
- Logging: reuse `writeLinkRepairLog` (`validate_page_content.go:599`) with
  `linkRepairOrigin{AgentType: params.AgentType, StepName: <current step>, ActionName:
  "save_page_sections", PageName: pageName, PageURL: <see note>}`. Keep the SAME error
  code `CONTENT_LINK_REPAIR_DETAIL` — origin fields discriminate (precedent:
  `rerender_link_repair.go:77-85` did exactly this for the rerender seam). When the index
  load is NOT ok, write the `CONTENT_LINK_REPAIR_SKIPPED` row and skip repair, mirroring
  `rerender_link_repair.go:52-60` — **fail open; never block a save on a failed index
  load.** PageURL note: the pageID lookup (`:1118-1124`) reads only `pages.id`; either
  extend it to fetch `url` or pass `""` — origin fields are best-effort.
- Reversal lever: step-config bool `repair_internal_links`, default **true**, read via
  `configBoolOrDefault` — same name and rationale as `validate_page_content.go:212-218`
  (DB config is live-immediately, so the behaviour can be withdrawn fleet-wide without an
  image roll; off-by-default would be inert and the bug still live).
- Seam function: extract the per-section pass as something like
  `repairSectionLinks(sections []SectionData, index datahelpers.PageURLIndex, indexOK
  bool) (rewritten, unlinked int, repairs []datahelpers.LinkRepair)` mutating
  `sections[i].HTML` — pure, no DB — so it is testable exactly the way
  `rerender_link_repair_test.go` tests its seam (ActionParams with nil DB, index passed
  directly). Do NOT re-test repair semantics — `datahelpers/link_repair_test.go` (14
  tests) owns those.
- Log line: give the applied-repairs Info log a distinctive compiled string (it is the
  pod-grep marker), e.g. `"SavePageSectionsAction: repaired dead internal links before
  persist"` with zap fields for counts.

### What NOT to do

- Do NOT touch the rerender outbound seams (`rerender_single_page_action.go:160-162`,
  `rerender_pages_actions.go:213-215`) — they deliberately repair the OUTBOUND string
  only. After this fix they become belt-and-braces; leave them.
- Do NOT remove `html_field: validation_result.clean_html` from the page-build-handler
  config — it is the live fallback when metadata is empty.
- Do NOT wire `InjectLinkConstraints` (092's dead duplicate; zero call sites, stays dead).
- Double-repair is fine: tool-recreation persists already-repaired `clean_html`; a second
  pass is a byte-identical no-op — do not special-case it.

## Tests

1. Seam test (new, `package actions`, no DB): three sections — one with a phantom link,
   one with a valid link, one with `href=""` — assert phantom unlinked with inner text
   kept, `href=""` unlinked, valid section **byte-identical**; plus `indexOK=false` →
   all sections byte-identical (fail-open).
2. `go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...`
   — the shared tree may not compile for unrelated reasons (another session's WIP): test
   against `git archive HEAD` + your files if so, never edit their code.

## Deploy + live verification runbook

1. Commit by pathspec (the action file + test file only), message names 079.
2. Council submission (platform change; advisory but the norm): rationale = this file's
   "The defect" + "Why here"; plan ≤8 edits; `grounded_in` quotes from the two action
   files. Budget ~30 min queue. Read `decided_by` per round; ~11% of REVISEs are one
   seat's unparseable JSON (`unreadable`, not `abstained`). Trailer
   `Council-Reviewed: <SUBMISSION_CORR>` only on APPROVED.
3. `make build-agent-chassis` (builds committed HEAD), **bump IMAGE_TAG** (makefile ~:16),
   push + deploy. **A retag is not a rebuild** — check image `.ID` + `.CreatedAt`.
4. Pod-grep the new compiled log string (positive marker) + an existing string as positive
   control. No orchestration dispatch within ~300s of the pod (re)start.
5. **Live proof, zero LLM spend:** fire a `page_rerender` work item (reason
   `section_data_resolved`) at gamesdesign `bayesian-ranking`
   (site `e33263f4-74f8-494f-b191-546845dbbddf`, page `b3c2da23-d867-4bc3-8641-80d3c8775067`,
   `/tools/bayesian-ranking.html`). The rerender feeds the STORED sections — which carry
   `href=""` since 07-21 — through `save_page_sections`. After the fix the saved
   `page_components.rendered_html` must have lost `href=""` and the served page must be
   clean. **The INSERT must include `handler_agent='page-rerender'`** or the item
   hard-blocks (template in `cta_link_integrity/NOTES…` 07-28 correction; also needs
   `source` and `created_by`, both NOT NULL). Verify the PERSISTED rows, then crawl the
   SERVED page — never the action's return map (016b §9, landmine L18).
6. Second repro if wanted: vonc `/about.html` (page-build route). fundamentallyai
   capabilities is NO LONGER a repro (reverted clean 07-28).
7. Close-out: 079 → `bugs_closed/` only when fixed AND live AND repro pages verify; 016b
   §9 already carries the transferable pattern; update memory
   `bugfix-079-phantom-link-gate.md`; 092 keeps its own file (upstream cause, still open).

## Config follow-up (optional, cheap, AFTER the roll)

`page-rerender`'s and `page-build-handler`'s save steps need no config change (repair
defaults on). If the council wants an explicit declaration, a seed adding
`repair_internal_links: true` to the save step configs is cosmetic — the in-code default
carries it.
