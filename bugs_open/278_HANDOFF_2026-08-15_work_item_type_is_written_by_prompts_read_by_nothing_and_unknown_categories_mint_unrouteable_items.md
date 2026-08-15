# 278 — audit prompts demand a `work_item_type` nothing reads, and `classifyFinding` mints an unrouteable item type for every category outside its hardcoded sets

**Filed 2026-08-15** by the `bugs_open/272` session, as the second half of the
LANDMINES entry *"`write_audit_findings` drops a JSON OBJECT silently and routes
on `category`, not on the `work_item_type` your prompt carefully asks for"*
(half 1 — the object-shape parse gap — is `bugs_closed/272`, fixed and live on
v1.0.1301). **Status: OPEN, researched and evidence-complete, fix NOT started —
this file is the cold-start handoff for the fixing session (see the final
section).**

**No `090` diagnosis run — first-hand verification substituted (stated per the
owner ruling of 2026-07-31):** every claim below is either a direct code read of
`classifyFinding` (this session, at HEAD `2a3ea3e2c`+), a live `agent_definitions`
query, or a live `site_work_items` census, each quoted inline. The mechanism has
also been independently read twice before: by the vigilant_designer session that
wrote the LANDMINES entry (2026-08-14, reading the action before wiring
`offer-analyser` to it) and, from the symptom side, by `bugs_open/115` (2026-07-27).

## The defect, three legs

### Leg 1 — `work_item_type` is a false affordance: demanded from the LLM, discarded unparsed

Two live prompts require it in EVERY finding (`agent_definitions`, 2026-08-15):

- `site-review-agent`: `"work_item_type":"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure"` — "Each finding MUST include ALL of these fields"
- `content-quality-auditor`: `"work_item_type":"content_rewrite|needs_content_page|tone_shift|cta_improvement"`

`[MEASURED]` Nothing reads it, checked at three layers:
- The `auditFinding` struct (`write_audit_findings_action.go:171-184`) **has no
  such field** — the value is dropped at json.Unmarshal / map-extraction, before
  routing is even reached.
- Go, all spellings: `grep -rn "work_item_type\|WorkItemType\|workItemType"
  --include=*.go platform/ internal/ pkg/` → one hit, and it is a WRITER
  (`create_tool_cross_link_items.go:270` stamps the key into a spec map);
  zero readers. (Spelling caveat honoured: three variants searched.)
- Agent configs: `default_config::text LIKE '%work_item_type%'` over live
  `agent_definitions` → exactly the two prompt occurrences above; no
  `query_database` step, no reader.

Routing is decided entirely by `classifyFinding`
(`write_audit_findings_action.go:205-435`) from `category` + page existence.
The LLM's choice costs tokens on every finding of two auditors and does
nothing; worse, a prompt maintainer who "fixes routing" by editing the
`work_item_type` vocabulary changes nothing and gets no signal (the
[[writes-the-field-is-not-reads-the-field]] shape, and the same defect class as
`bugs_closed/264`: a config value that looks load-bearing and is dead).

### Leg 2 — the Rule-6 fallback mints a novel item type for any unknown category, silently

`write_audit_findings_action.go:423-434` (verbatim at time of filing):

```go
// ── Fallback: unknown category → content-gap-planner for triage
spec["page_name"] = pageName
return classifiedFinding{
    ItemType:     "audit_finding_" + category,
    HandlerAgent: "content-gap-planner",
    ...
    Priority:     priority + 10,   // deprioritised on top of being unrouteable
```

No vocabulary guard, no log line, no counter in the result map. The minted type
is registered nowhere: `bugs_open/115` measured `audit_finding_brief_fidelity`
appearing in the whole Go tree only in a coverage test, and in zero
`agent_definitions` rows — so `detected` is terminal for it in practice.

`[MEASURED]` Live census 2026-08-15 (`item_type LIKE 'audit_finding_%'`): **6
rows, 3 distinct minted types' worth of producers**:

| item_type | status | audit_source | created |
|---|---|---|---|
| audit_finding_brief_fidelity | cancelled | design-audit | 2026-07-24 |
| audit_finding_audience | **complete** | design-audit | 2026-08-11 |
| audit_finding_brief_fidelity ×4 | **detected** | brief-fidelity-audit | 2026-08-13 |

The one `complete` row completed same-day with **no `resolution` in its spec**
— `[UNVERIFIED]` whether content-gap-planner actually did anything with it
(memory: a `complete` work item is not a repaired artefact; check what closed
it before citing it as "the planner can handle these").

### Leg 3 — one auditor is unrouteable BY CONSTRUCTION

`[MEASURED]` The four live producers' prompt category vocabularies against
`classifyFinding`'s sets (`designCategories`/`metadataCategories`/
`componentCategories`/the Rule-4/5/6 names):

| producer | prompt categories | routable? |
|---|---|---|
| site-review-agent | structure, content, gap, cta, differentiation | all ✓ |
| visual-design-auditor | colour, spacing, typography, dark_section, responsive | all ✓ (designCategories) |
| content-quality-auditor | tone, gap, cta, differentiation, content | all ✓ |
| **brief-fidelity-auditor** | **`brief_fidelity` — hardcoded, the only value** | **✗ — in NO set** |

So **100% of brief-fidelity-auditor's possible output takes the minting
fallback** — which is why `bugs_open/115`'s three correct findings died in
`detected`. Compounding (from `bugs_closed/264` §12, still true then): the
auditor also has **no live caller** — no workflow references it, no
`scheduled_tasks` row; its 08-13 rows came from a manual verification dispatch.
A reliable check (115: "right about everything, three days early") that nobody
runs, whose output nothing can route.

## Reconciliation with the neighbouring files (read these before fixing)

- `bugs_open/115` — the SYMPTOM side (open, unowned): the three 07-24 findings
  that died. Its fix candidate 1 ("route the item type, or refuse to write it")
  IS this file's writer-side fix; its candidates 2-4 (terminal-state audit on a
  cadence; run the audit on a cadence; surface open findings) are the
  family-level detector and stay with 115/083. **This file owns the writer
  mechanism; 115 keeps the routing/cadence candidates. Cross-ref appended
  there.**
- `bugs_open/083` (the "detected findings never reach a handler" one — number
  is ambiguous, resolve by slug) — the fleet-wide family; 115 contributed the
  298-items-no-handler census there. A fix here removes ONE producer of that
  family, not the family.
- `bugs_closed/077` — the precedent this estate already ratified: "found work I
  have no handler for" files as **`capability_gap`**, which is read as a roadmap
  and cannot silently rot.
- `bugs_closed/213` — the verifier/producer-join precedent on this same action
  (dark_section got its OWN item_type because a shared one re-ran the wrong
  verifier predicate), plus `write_audit_findings_verifier_join_test.go`, which
  structurally asserts the designItemTypes↔verifier join — the natural home for
  a new "every mintable item_type has a registered consumer" assertion.
- `bugs_closed/272` + the LANDMINES entry (grep `drops a JSON OBJECT silently`)
  — half 1, and the live table of which sibling parse switches still lack an
  object case.

## Fix candidates, ordered by what closes the door

1. **Make the unrouteable state unrepresentable at the writer.** In
   `classifyFinding`'s fallback: stop minting `audit_finding_<category>`.
   File the finding as **`capability_gap`** (the 077 precedent) with the
   original category in spec, OR refuse it and count it in the result map
   (`unrouted_categories: {...}`) the way 272's fix added
   `findings_field`/`findings_type` to the zero path. Either way the result map
   and a Warn log must say it happened — today it is silent. Add the
   coverage-test assertion (213's test file) that every item_type this action
   can emit has a registered verifier/consumer, so the NEXT unrouteable type
   fails CI instead of shipping.
   ⚠ Decision to weigh (2026-07-29 owner ruling #1): "refuse to write unknown
   types" arguably changes what this shared action GUARANTEES (from "always
   writes something" to "writes only registered types"). Probably still normal
   council-gate scope (it strengthens, not weakens, and the consumer set is
   enumerated above), but say so in the submission rather than leaving the seat
   to raise it.
2. **Resolve `work_item_type` one way or the other — a field demanded and
   discarded is the worst of both.** Either (a) delete it from the two prompts
   (cheapest; stops paying tokens for a false affordance; routing stays on
   category, which for these two auditors is fully routable today), or (b) parse
   it and honour it as a routing override when it names a verifier-known type.
   (b) changes live routing for two currently-working auditors — if chosen, it
   wants its own careful diff of what would have routed differently
   (`spec->>'category'` vs prompt `work_item_type` on existing rows). Candidate
   (a) is the recommendation; (b) needs a reason.
3. **Give brief-fidelity-auditor a real route** — a deliberate mapping for
   `brief_fidelity` (most findings are page-scoped content deviations from the
   brief: plausibly Rule-4 `content_rewrite` semantics, or `needs_spec_update`
   when site-wide), decided WITH its wiring gap (no live caller) rather than
   before it: routing an auditor nobody dispatches fixes half a defect. That
   wiring decision is product-shaped — flag to the owner rather than assume.
4. NOT this file: the cadence/terminal-state audit (115/083's candidates) and
   the four `detected` rows' disposition (they are 115's evidence; do not
   bulk-cancel them).

## How to verify a fix

1. Unit: with a finding whose category is outside every routing set, the action
   must NOT insert an `audit_finding_*` row — it inserts `capability_gap` (or
   refuses, per the chosen candidate) AND the result map names the category.
   Mutation-check the guard (272's file shows the pattern).
2. Coverage: the new join assertion fails if any emittable item_type lacks a
   registered consumer (prove by temporarily re-adding the minting line).
3. Live: dispatch one `brief-fidelity-auditor` run (site 62b5978e…, the
   mortgagecalculator rows show it works mechanically); confirm zero new
   `audit_finding_%` rows and that its findings land as routable types or
   `capability_gap`, with `audit_source='brief-fidelity-audit'` intact.
4. Prompt leg: after candidate 2a, confirm the two prompts no longer contain
   `work_item_type` (live `agent_definitions`, not the seed — the seed is not
   the system).

## COLD-START for the fixing session

Read, in order (≈15 min): this file · `classifyFinding` +
`WriteAuditFindingsAction` (`platform/orchestration/actions/write_audit_findings_action.go:205-435,469-`)
· `bugs_open/115` · the LANDMINES entry (grep `drops a JSON OBJECT silently`) ·
`bugs_closed/213` §on the verifier join + `write_audit_findings_verifier_join_test.go` ·
`bugs_closed/272` (the shape of the previous fix on this action: council round,
mutation-tested table test, artefact-level deploy verification — reuse all
three patterns).

Then: (1) re-run the census above (other sessions move fast here — 272 went
filed→fixed→live→closed inside 26 hours, and the parser gained a third return
from the 213 lane within a day); (2) check `who-owns.py 115` and grep live
transcripts for `classifyFinding` before starting; (3) decide candidate 1's
refuse-vs-capability_gap fork and candidate 2a, submit to the council gate
(`097` trigger — budget ~30 min, commit with `Council-Submitted:`), commit by
pathspec, and leave the Go change to ride the next chassis build. Config/prompt
edits (candidate 2a) are live immediately via migration — image-first ordering
does NOT bind here since nothing new is read, but keep the prompt edit in the
same review. The 4 `detected` rows and the wiring of brief-fidelity-auditor
need an owner decision — put both in the README/where-we-are when you get
there.
