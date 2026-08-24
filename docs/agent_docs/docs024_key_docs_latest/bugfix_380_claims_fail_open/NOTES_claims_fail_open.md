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
