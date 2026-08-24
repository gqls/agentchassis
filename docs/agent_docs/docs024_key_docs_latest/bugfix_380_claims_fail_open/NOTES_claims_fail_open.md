# NOTES — bugfix_380_claims_fail_open (append-only, newest at the bottom)

## 2026-08-24 session 1 (bugs_open/380 session) — adoption through first live proof

**Done and LIVE (all verified at the artefact):**
- Migrations **597** (auditor fail-closed + cold arm + receipts + cap 12000) and **598** (planner
  arms) applied ~16:20Z/16:21Z, recorded in schema_migrations (record-only, notes cite the
  DO/RAISE verifies + snapshots 78714f54 / f263eaa1). ROLLBACKs committed alongside. NOT yet
  committed to git — files sit in the tree with 599_HOLD, 600, the v5 plaintext, the seed.
- **599** (writer arm) written as `_HOLD` + v5 plaintext for the owner's read
  (`brochure_component_library/sql/page_content_writer_prompt_v5_2026-08-24.txt`).
  ⚠ MISSTEP (logged for WRONG_CALLS): I first generated that plaintext from migration 330's
  base64 — the LIVE template is 1,718 chars different (two later insertions, House Voice tail
  gone). PBP-035's exact warning. Caught by length diff; regenerated from live.
- **600** (rotation, 3600s/7-day, owner decision D2) written, NOT applied — held for the
  hand-dispatch proof per plan.
- **SEED_claims_auditor.sql** regenerated from the post-597 live row (first seed ever).
- **First live cold audit ever ran** (TRIGGER_claims_audit.sh, receipts asserted):
  garden-tools.uk corr 86ef3e17 (cold arm rendered=t, findings returned, `claims_llm_*` work
  item filed — the FIRST in that step's life) + leopardessconsulting.co.uk control corr
  be39ddba (roster arm, found a REAL drift finding: "22 sites" vs verified floor 25).
  doc_notes receipts written for both (pipeline/claims-audit).
- Attestation-only registers fleet-wide = **0** `[MEASURED 2026-08-24 ~17:00Z]` — the
  numeric-scan guard in the Go slice is a live no-op (the enumeration the council needs).
- Owner decisions D1-D4 recorded in PLAN. Go slice (claims_practice.go family, warning default,
  NOT unioned) is being written by a subagent — NOT yet landed/committed.

**OPEN DEFECT FOUND BY THE PROOF (unresolved, do not lose):** the auditor's `load_page_text`
extraction is broken in a way the one-run history never surfaced. Evidence, garden-tools run
16:52:14Z: all 6 page rows ARE in `prompt_rendered` (positions found for every name) but
`how-we-assess`'s rendered row does NOT contain its `faq` component's text — the owner's
"we buy the tool at the same price" sentence is in `page_components.rendered_html` (slot faq,
5,409 chars, has_sentence=t) and ABSENT from the prompt ('we buy' occurs 0 times in the whole
prompt). Also `index` strips to **1 char** of text and `care`/`seasonal-planner` to a few
hundred, from ~12-14k of html each. Theories tested: components locked (no — 0 locked),
page predicate (no — 6/6 in prompt), rerender raced the run (no — components last updated
08-23 21:44Z), unbalanced per-component `<style>` pairing across string_agg order (no —
opens=closes for every component). NEXT THEORY untested: the `<style[^>]*>.*?</style>`
non-greedy strip still pairs ACROSS components because `string_agg` has NO ORDER BY — even
with balanced tags, `<style A>...</style> ... <style B>...</style>` collapsed in a different
aggregation order could interleave; OR the CSS/text contains a literal `<` run; OR
`{{.page_texts}}`'s Go-map render (`map[columns:... rows:[...]]`, itself ugly but harmless)
is truncating. ALSO NOTE: `{{.page_texts}}` renders the whole query_database result object as
a Go map — the model reads it fine but it wastes tokens and the first row is duplicated at
top level. The fix candidate: give `load_page_text` a deterministic
`ORDER BY pc.position`/`pc.slot_name` inside string_agg AND per-component style-stripping
(strip BEFORE aggregation), or extract text in Go. **The audit found real findings anyway**
("We test and research carefully", affiliate-disclosure, severity high) so the mechanism is
proven; the extraction defect UNDERCOUNTS. File as a follow-up bug or fix in this lane before
the rotation (600) amplifies it.

**Cross-session state:** 381 session coordinated (numbers: theirs 591-595, mine 597-600 —
596 was taken by the 243 lane mid-plan); 364 session HOLDING claims.go edits until my Go
commit lands — I owe them the sha the moment it lands. loanzy session took my two corrections
into the bug file (33/48 not 29/48; auditor UNDRIVEN not leaky) and caught that my 597 header's
"ONE llm_call_log row ever / claims_llm% = 0 rows" figures were staled by my own verification —
PIN THOSE PREDICATES with `AND created_at < '2026-08-24 16:00Z'` in the header BEFORE
committing 597. The DB lock convoy (orphaned webdesign.co.uk COPY, 85 waiters) was diagnosed
here and terminated by webdesign-tool-rebuild — LANDMINES entries owed (ClientWrite: cancel
insufficient/terminate; shell `timeout` orphans the psql backend — use statement_timeout).

**Still to do (the plan file is authoritative):** commit everything per the commit map
(pathspec, Council-Submitted trailers); council submissions S-config + S-go via 097; apply 600
after the extraction defect is at least filed; docs (bug file §5/§6 fix record, CLM-024/025/026,
LANDMINES ×6, WRONG_CALLS ×2, CONTRIB notes ×9, RFC_003 owner answers); message 364 the sha;
owner reads the v5 plaintext then 599 applies.

## 2026-08-24 session 1, part 2 — everything shipped; what the proof and the council found

**Shipped (all 2026-08-24):**
- Commit `856d0e1fd` — config slice: 597, 598, 599_HOLD (+ v5 plaintext from LIVE), 600, 601, seed,
  lane docs, TRIGGER script. `Council-Submitted: e684fc8d` → **APPROVED round 1, 5 advisory
  objections, none high** (19:23Z). 597/598/600/601 applied + recorded (record-only, notes cite the
  verify blocks). 599 HELD for the owner's plaintext read (D4).
- Commit `c9cd817d9` — Go slice: `claims_practice.go` (+test), `claims.go`, `validate_page_content.go`,
  `check_unverified_claims.go`, `cmd/claimscan`. `Council-Submitted: 1d87615f` (verdict pending).
  INERT until an image rolls. HEAD build/test proven for its packages; `cmd/config-key-audit` is red
  AT HEAD 111ee2314 on `TestBudgetCronCountsLiteralMatchesTheRegistry` (`create_tool_component` 4 vs 5,
  `deploy_tool_to_site` 3 vs 4) — another lane's Optional keys without a `check.py` regeneration; I
  added no Optional key. The 364 session is now rebasing on c9cd817d9 (told them).
- Pattern-check advisory on the Go commit: `logged-model-output` at `validate_page_content.go:1247` —
  `practiceClaimsSeverity` logs a variable named `raw`; it is STEP CONFIG, not model output. Rename to
  `configured` next time the file is touched (cosmetic; not worth a fourth commit tonight).

**The proof, and the defect it found (601):** first cold run at garden-tools (corr 86ef3e17) returned
findings but the owner's sentence was NOT in `prompt_rendered` (2,702 input tokens). Traced: not
locks, not the page predicate, not a rerender race, not unbalanced tags (opens = closes per component).
Cause: **PostgreSQL ARE takes the greediness of the FIRST quantifier**, so `<style[^>]*>.*?</style>`
over an unordered `string_agg` eats first-`<style`-to-last-`</style>` across components (how-we-assess
3,732 → per-component 8,269; `index` → 1 char). 601 strips per component, lazy, ordered; verify block
ran the motivating extraction live (8,266 chars, sentence present). Re-dispatch (corr bcf23316,
6,313 input tokens): first two findings = the owner's two sentences, severity high.
⚠ **MISSTEP — I re-derived a mechanism already documented THREE DAYS EARLIER**: migration **518**
(2026-08-21, `bugs_open/320` lane) found the identical defect ("noted.co.uk/index measured as ONE
character"), fixed the visible-text MEASURE per component (517 = `page_visible_text_len(uuid)`), and
measured 349 of 693 pages losing >half their text. My fleet census grepped the LITERAL regex (1 hit)
instead of the mechanism (`string_agg(` + `regexp_replace`) — the council's bug_historian seat called
exactly that narrowing. Re-census (mechanism, live configs, 2026-08-24): `.*?` in a Postgres
`regexp_replace` on **meta-description-backfiller.load_pages_missing_meta**, **webdesign-agent.
load_decisions** (+ unordered `string_agg`), unordered `string_agg` over `rendered_html` on
**internal-linker.load_target_page** and **visual-design-auditor.load_design_context**. Not fixed
here (other lanes' agents); recorded in LANDMINES addendum + WRONG_CALLS. Follow-on worth doing:
ONE shared `page_visible_text(uuid)` TEXT function with 517's length derived from it, so 601's
formulation is not a third copy.

**Council advisories (config slice) and the answers:**
1. editquality / bug_historian — mechanism (b) (the writer arm) is HELD, so one of three paths is
   uncorrected at source; detection-only interim. TRUE and disclosed; owner decision D4. Closes when
   the owner reads `page_content_writer_prompt_v5_2026-08-24.txt` and 599 applies.
2. bug_historian — fleet audit for the `conditional_branch → else_step: complete` skip-as-success
   shape. DONE `[MEASURED 2026-08-24]`: **38** conditionals fleet-wide route a branch to a
   `complete_workflow` step; most are legitimate nothing-to-do exits (no pages/candidates, dry_run,
   approve/reject routing). The 380 SHAPE — a missing TARGET or absent EVIDENCE completing as success —
   is ~6: `site-adoption-agent.check_crawl_content` (content_quality none → complete),
   `tool-auditor.check_tool_found` / `tool-improver.check_tool_found` (tool missing → complete),
   `internal-linker.check_target_found`, `page-build-handler.check_page_found` /
   `tool-recreation-handler.check_page_found` (the bugs_closed/299 "skip is a third state" case that
   `bugs_open/354`'s `error_route_completion.go` now distinguishes). Handed to the 354 lane (owner of
   the third-state seam) rather than patched here.
3. bug_historian — the 601 census was literal-string, not mechanism. TRUE; re-censused above.
4. reuse_agent — why a separate LLM rotation (600) rather than a check TYPE in the existing
   discovery loop? Answer, recorded: the discovery loop is deterministic Go checks on a 3-hour
   `quality-discovery-agent` cadence; the auditor is ONE Sonnet call per site (the claims layer's V3
   "judgement lane" by design, `SPEC_claims_verification.md`), and an LLM call inside the 3h loop
   would cost ~8× the 7-day cadence the owner ruled. The deterministic practice family (CLM-026) IS
   headed for the discovery check (slice 2b) — that is the "check type" the seat wants, and it is
   where it belongs. 590's rotation shape is the estate's template for a per-site LLM agent on a clock.
5. reuse_agent — the TRIGGER script duplicates a pattern. Partly true: it uses the shared
   `kafka-publish-lib.sh` (the 200 racing copies do not) and adds the never-create-a-site refusal;
   parameterising a generic single-agent trigger is a fair follow-on for whoever owns the triggers.

**Still open in this lane:** owner read of the v5 plaintext → apply 599; Go council verdict
(1d87615f) → act on REVISE; image roll → verify a `practice_claim` warning appears in a
validate_page_content result on garden-tools (and re-run claimscan against the served pages);
slice 2b (discovery-check wiring) with the 12/1,867 number in hand; the shared visible-text
function; the LLM item's missing `page_id`.

**Go slice council verdict (19:29Z): APPROVED round 1, 1 advisory objection, none high** (corr 1d87615f).
Seats and what they said: editquality (medium) — my claimscan SKETCH omitted the third call-site guard
(`if eb.HasScannableRegister()` before `ScanUnregisteredNumbers`); the CODE has it (c9cd817d9,
`cmd/claimscan/main.go`) — a submission-writing omission, not a code gap. guardian (medium) — the
`ParseEvidenceBase` widening is a shared parse path and the consumer enumeration should be re-verified by
a human, not asserted: the list is in the submission's risks §4 (gate, save floor, section editor,
negation rewriter, meta-description action, discovery check, tool_backend_provision, content-duplication
check); each either scans banned claims (nil-safe; an attestation adds no patterns) or is now guarded by
`HasScannableRegister()`; `tool_backend_provision` keys on `data ? 'facts'`. Owner: worth a read when
the image rolls. guardian (low) — the mutation test passes `nil` into `ScanAllBannedClaimsWithSuppressed`:
compiles and passes (nil-safe receivers). architecture (low) — CLM-026 should list every nil-rule
widening: it names both (`regulated`, `operating_history`). Docs commit 171ffed55; index rows for
CLM-025/026/027 added in the next commit (the pattern check caught the missing rows).
