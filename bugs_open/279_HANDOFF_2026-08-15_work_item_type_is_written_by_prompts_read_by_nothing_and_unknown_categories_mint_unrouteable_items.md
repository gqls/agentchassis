# 279 — audit prompts demand a `work_item_type` nothing reads, and `classifyFinding` mints an unrouteable item type for every category outside its hardcoded sets

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

---

## STATUS UPDATE 2026-08-15 (fixing session) — fix built, submitted, committed; Go half awaits the next chassis roll

**Council submission:** `Council-Submitted: 925d7759-ddd7-4504-a752-4550e0f32220`
(APPROVED/REVISE verdict not yet read at commit time — resolve via the
`diagnosis_artifacts` query in the runbook).

**Decisions taken, per the candidates above:**

- **Candidate 1: the `capability_gap` fork, not refuse-and-count.** The fallback
  now files the platform's "found work I have no handler for" shape
  (`bugs_closed/077`), mirroring `CapabilityGapItem` (`discovery_checks/remit.go`)
  exactly: status `'deferred'`, **empty** `handler_agent`, severity `low`,
  priority 200, `spec.builder_needed`/`capability`/`not_dispatchable`/`gap_kind`
  (= `rule_missing`), the finding's own severity kept as `spec.finding_severity`.
  `diagnose_triage`'s roadmap view (`item_type='capability_gap' OR
  status='deferred'`) reads these, so they surface instead of rotting. Refusing
  was rejected because it discards the observation the estate has a ratified
  vocabulary for. Dedup key `capability_gap:unrouted_audit_category:<category>` —
  one open row per site per unknown category. The result map gains
  `unrouted_categories` (only when non-empty) and each unrouted finding draws a
  Warn — the silence is closed.
- **A finding it took the tests to surface:** an unknown category on an
  **existing** page never reaches the fallback — Rule 4's `default:` branch
  routes it to `content_rewrite`. The fallback fires only when the page does not
  resolve (free-prose page names — exactly how the production rows were minted).
  Rule 4 unchanged, deliberately; now documented in the test.
- **The CI gate:** `classifyEmittableItemTypes` in
  `write_audit_findings_verifier_join_test.go` declares all 13 emittable types
  with their consumers; `TestClassifyFindingEmitsOnlyDeclaredItemTypes` drives
  every category set + Rule 4/5/6 literals + off-vocabulary samples through every
  page situation. **Mutation-verified**: re-adding the minting line → 17
  failures; restored → green (full `platform/orchestration/...` suite passes
  against a clean `git archive HEAD` overlay — the working tree itself doesn't
  compile due to another session's WIP `publish_site_action.go`).
- **Candidate 2a (delete the dead prompt field): migration `416`**
  (`sql_for_agents/416_auditor_prompts_drop_dead_work_item_type.sql` +
  `_ROLLBACK` sidecar). Replace literals dry-tested read-only against the live
  rows before writing the file (2→0 occurrences each, jsonb cast holds, steps
  intact). Applied + recorded — see below for verification.
- **Shared-path control:** the guardian-seat sqlmock test from the 213 round
  caught the INSERT's status parameterisation on first run (its job); updated to
  pin that a routed finding still inserts `'detected'`.

**Verification state, against the "How to verify" list above:**

1. Unit — DONE (see mutation check above).
2. Coverage join assertion — DONE (`TestClassifyFindingEmitsOnlyDeclaredItemTypes`).
3. Live brief-fidelity dispatch — **BLOCKED until the next chassis roll**: the Go
   half is inert until an image ships (check with the build-provenance stamp,
   then dispatch one run against site 62b5978e… and confirm zero new
   `audit_finding_%` rows, findings landing as `capability_gap` with
   `audit_source='brief-fidelity-audit'`).
4. Prompt leg — verified live post-apply (query + result recorded below).

**Still needing an OWNER decision (unchanged from candidates 3/4):**
- the 4 `detected` `audit_finding_brief_fidelity` rows' disposition (115's
  evidence — not bulk-cancelled by this fix, deliberately);
- whether brief-fidelity-auditor gets a real route + a live caller (candidate 3;
  product-shaped). Until then its findings will surface as `capability_gap`
  roadmap rows, which is the honest interim.

## VERDICT 2026-08-15 — APPROVED, round 1 (corr `925d7759`), 3 advisory objections, all dispositioned

**`Council-Reviewed: 925d7759-ddd7-4504-a752-4550e0f32220`** — verdict read in full;
the code commit `d6d56e540` carries `Council-Submitted:` and is auto-credited by 098.
12 seats returned (4 abstained); approvals from editquality, guidelines,
tooling_provenance*, guardian*, diagnosis_guardian, improvement_guardian, compliance,
debug_historian*, constitution, mission, prior_art_librarian, architecture
("point_fix... convergence onto existing architecture, not accretion"). Starred seats
objected advisorily; every objection was verified live this session rather than filed:

- **debug_historian (medium)** — migration 416's `WHERE type=` shape matches the
  multi-version `agent_definitions` landmine (dormant second active row). **CHECKED,
  does not apply:** both agents have exactly ONE definition row in the whole table
  (version 2, active, field gone — query in scrollback recorded here: all-rows census,
  no type/status filter). The migration's own precondition (`count=2` across both
  types) would have refused the multi-row case. Its second (low) point — no post-apply
  verify — was a plan-sketch artefact: the applied file HAS a DO/RAISE postcondition
  block.
- **guardian (low ×2)** — dedup widening is unconditional: **CHECKED**, exactly one
  live `deferred` row matches this action's key family (a `design-audit`
  `needs_content_planning` from 08-04); for it the widening converts a would-be
  unique-index insert error into a counted skip — no outcome change, `idx_swi_dedup`
  already blocked the row. Byte-identity for all 5 producers: **CHECKED LIVE** —
  visual-design-auditor prompts `colour|spacing|typography|dark_section|responsive`
  (all designCategories), offer-analyser hard-requires its seven routable values (its
  prompt even warns against the minting bucket), brief-fidelity-auditor hardcodes
  `brief_fidelity` (the intended-repair case). Its medium (migration filed as
  operation 'add' not 'config_change') is a submission-labelling point — noted for
  next time, nothing to change post-approval.
- **reuse_agent (medium ×2)** — hand-mirrored capability_gap shape vs a shared
  constructor, and `classifyEmittableItemTypes` as a second registry. **Answered in
  the doc_notes decision record** (subject `action/write_audit_findings`):
  `CapabilityGapItem` builds its own spec for the discovery-check insert path and
  cannot carry the audit finding's payload; extraction waits for a third producer
  (the WII-016 third-adopter pattern). The two registries answer different questions
  (emittable-by-this-action vs verifier coverage estate-wide) and live in different
  packages; drift between them is caught by the coverage test's own union rule.
- **tooling_provenance (medium)** — no doc_notes row for the shared action.
  **DISCHARGED:** decision record written to `doc_notes`
  (`subject_type='action', subject_key='write_audit_findings'`,
  categories `["council-gate","decision-record"]`) covering the fallback shape, the
  Rule-4 nuance, the closed-set rationale, the constructor decision, and the
  brief-fidelity interim.

Remaining to close this file: the Go half goes live on the next chassis roll →
run verify step 3 (one brief-fidelity dispatch, zero new `audit_finding_%` rows,
findings landing as `capability_gap`), then move to `bugs_closed/` — plus the two
owner decisions above, which survive the closure as 115/candidate-3 items.

---

## CONTRIBUTION 2026-08-15 from the `bugfix_213` lane — the blocked filter can suppress the very rows Leg 2 now files, and it is BLIND TO BOTH producer and category

**Not a competing claim and not a request.** `who-owns.py 279` says this file is yours and
active, so this is filed here rather than as a new bug. It is one mechanism downstream of Leg 2's
fix, found while auditing who else in the estate reads an *absent* finding as meaningful. **Take
it, reshape it, or reject it — I am not working it.** Everything below is first-hand: a code read
at HEAD plus a live census, both quoted.

### The mechanism

`write_audit_findings_action.go` runs a second suppression check on every finding, after the
dedup-key check and before the insert (the "Broader blocked check", ~line 793):

```sql
SELECT EXISTS(
    SELECT 1 FROM site_work_items
    WHERE site_id = $1 AND status = 'blocked'
      AND item_type = $2
      AND ($3::uuid IS NULL OR page_id = $3::uuid)
)
```
`$3` is the NEW finding's `PageID`. When it is NULL the third clause is **TRUE for every row**,
so the predicate collapses to *"does this site have ANY blocked row of this item_type"*.

**Leg 2's fallback always sets `PageID: nil`** (verified at source, the `return classifiedFinding{…}`
block: `ItemType: "capability_gap", HandlerAgent: "", Status: "deferred", PageID: nil`). So for
`capability_gap` the NULL branch is not an edge case — **it is the only branch that ever runs.**

The consequence is that the filter defeats the dedup key sitting three lines below it. That key is
per-category —
`DedupKey: fmt.Sprintf("capability_gap:unrouted_audit_category:%s", category)` — and its comment
states the intended design in terms: *"One open row per site per unknown category — the missing
thing is a rule for the CATEGORY, however many findings or producers hit it."* The blocked filter
reduces that to **zero rows per site for EVERY category, once any one `capability_gap` on that
site is blocked.**

### It is producer-blind too, and that is the sharper half

`capability_gap` is a **co-filed** item_type. The blocked rows are not yours — every one was filed
by the discovery/remit path, not by `write_audit_findings`:

```
 domain                   | status  | created_by                   | item_key
 idea.uk                  | blocked | completeness-discovery-agent | capability_gap:content_duplication_rewrite
 mortgagecalculator.co.uk | blocked | design-discovery-agent       | palette_contrast
 vetcomparison.uk         | blocked | design-discovery-agent       | palette_contrast
```
`write_audit_findings` stamps `created_by = auditSource` and keys as
`capability_gap:unrouted_audit_category:%`; **no blocked row has either shape.** So a row blocked
because a *discovery check* could not route `palette_contrast` now suppresses *your* path from
recording an unrelated unrouted audit category on that site.

This is the co-filing trap that `write_audit_findings_retraction.go` refuses **structurally** —
its header: *"the item_type does not name its producer … a candidate is only in scope when its own
`spec.audit_source` equals the source of the run doing the observing. A producer's silence says
nothing about another producer's findings."* The retraction helper has that guard. **The blocked
filter, reading the same shared item_type, does not.** Relevant to the owner's 2026-08-02 ruling
on converging N producers onto one `item_type`: the ruling's condition is that the producer set
and key shape be stated — this is a second mechanism reading that shared type with neither in view.

### Live blast radius [MEASURED 2026-08-15]

```sql
SELECT item_type, count(*) AS blocked_rows,
       count(*) FILTER (WHERE page_id IS NULL) AS page_id_null, count(DISTINCT site_id) AS sites
FROM site_work_items WHERE status='blocked' GROUP BY item_type ORDER BY blocked_rows DESC;
```
| item_type | blocked rows | page_id NULL | sites |
|---|---|---|---|
| `image_url_404` | 40 | 40 | 15 |
| `capability_gap` | 18 | 18 | 14 |

All 18 `capability_gap` blocks carry the same cause — `"No handler_agent set — item cannot be
routed to any agent"` (`claim_work_item_action.go:162`), i.e. something claimed a row Leg 2
deliberately files as non-dispatchable. **So the filter is ARMED on 14 of ~22 sites**, including
gamesdesign.co.uk, mortgagecalculator.co.uk and oufe.com.

### ⚠ What I did NOT establish — and the control matters more than the finding

**Whether the suppression has ever actually FIRED is unknown.** No run reports
`items_skipped_blocked > 0`. **Do not read that as reassurance** — I nearly did:

```sql
SELECT count(*), min(created_at)::date, max(created_at)::date
FROM orchestration_states WHERE collected_data->'findings_written' ? 'items_skipped_blocked';
-- 9 runs, oldest 2026-08-15, newest 2026-08-15
-- (orchestration_states itself: 5,282 rows back to 2026-07-13)
```
**The field exists in 9 runs, all from today, against a table holding two months of history** —
because the counter arrived with your own fix. Five of those nine are my lane's manual audits.
The meter is hours old, so the zero measures the instrument's age, not the estate's behaviour.
Two candidate reads remain open and this evidence cannot separate them: *the filter never fires
because unrouted categories are rare on blocked sites*, or *it fires routinely and nothing has
ever counted it*.

**The cheap way to settle it, if you want it settled:** `unrouted_categories` is already in the
action output (line ~888) and is counted **before** the filters, while `items_skipped_blocked` is
counted after. A run where `unrouted_categories` is non-empty and no matching `capability_gap` row
appears is the suppression, caught in the act, with no new instrumentation needed.

### Why it may be worth a line in your fix rather than a separate file

Leg 2's fix is what makes this acute, and it is worth being explicit that this is **not** an
argument against it. Before the fix, unknown categories minted `audit_finding_<category>` — a
*distinct item_type per category* — so an item_type-scoped blocked filter could only ever suppress
the one category that was blocked. Converging them onto a single `capability_gap` is the right
call for every other reason, and it is what turns one blocked row into a **site-wide, all-category,
all-producer** mute. The convergence did not create the defect; it widened its blast radius, and
the filter is where the fix belongs.

Candidate shapes, ordered by what closes the door (yours to judge):
1. Make the blocked check match the **dedup key**, not `item_type` + a NULL-collapsing page clause
   — it is the identity the row already has, and it is per-category by construction.
2. Failing that, scope it by producer as the retraction helper does (`spec.audit_source`), so a
   discovery-filed block cannot mute an audit-filed finding.
3. At minimum, stop `capability_gap` being claimed at all — it is filed `deferred` with an empty
   handler *deliberately*, and 18 rows say something claims it anyway. That is arguably the root:
   no block, no mute. (`bugs_open/077` is cited in Leg 2's own comment for this.)

**Cross-refs:** `bugfix_213_verifier_producer_join/NOTES_…md` 2026-08-15 batch 3 (why this lane
was auditing absence-keyed consumers) · `write_audit_findings_retraction.go` header (the
producer-scope guard this filter lacks) · owner ruling 2026-08-02 §1 (converging producers onto
one `item_type`).

---

## STATUS UPDATE 2026-08-15b — OWNER DECISIONS EXECUTED (round 2, `Council-Submitted: 336d1549-6d85-4c33-ab09-e8faaac5aae4`)

The owner ruled on candidates 3/4 and added a directive; all three executed this
session, ground re-checked first (chassis still v1.0.1302, pods predate
`d6d56e540` → the Go half is STILL not live; the 4 rows were still `detected`;
the 213-lane contribution below was the one thing that had changed).

1. **The 4 `detected` rows: CANCELLED** (owner decision), reason inline in each
   row's `error`; re-run post-roll. `bugs_open/115` updated.
2. **"Stop agents writing triage items with labels that don't exist":**
   `work_item_type_minting_ratchet_test.go` — a source-scan ratchet banning
   dynamically-CONSTRUCTED ItemType/itemType values (concatenation/Sprintf) in
   `actions` + `discovery_checks`, comments stripped before matching, with a
   self-test that the pattern still bites (4 must-match shapes, 5 must-not).
   Measured zero construction sites remain at this HEAD. **Stated residual:**
   config-minted types (`create_work_item` takes `item_type` from step config;
   `section_edit` has no Go literal) are invisible to source scans — claim-time
   handler resolution is their only backstop; a write-time check needs a
   vocabulary registry that does not exist.
3. **brief-fidelity-auditor promoted — but NOT via a Go route.** The route-to-
   `content_rewrite` sketch (candidate 3's own text) was **refuted by the
   auditor's four real findings**: three of four were design-intent violations
   ("Animations beyond hover states", "Rounded corners beyond 12px") that a
   content rebuild cannot repair. Instead the auditor now **speaks the router's
   vocabulary** (migration `417`, APPLIED + recorded: eleven routable categories
   chosen by repair shape, offer-analyser's warning discipline, exact-page-name
   instruction, explicit off-vocabulary escape hatch → `capability_gap`), while
   `audit_source` carries the brief-fidelity identity — category = who repairs,
   audit_source = who found. No new item_type, no closed-set change. Wiring into
   the improvement loop is migration **`419_HOLD`** (splices
   `spawn_brief_fidelity`/`call_brief_fidelity` between `call_offer_analyser`
   and `record_audit_pass`, mirroring the offer pair incl. `error_step`
   continuing the sweep) — **held until the chassis stamp carries `d6d56e540`**,
   apply commands + guards in the file's own header.
4. **The 213-lane contribution (blocked-filter mute): FIXED, candidate 2.** The
   broader blocked check is now producer-scoped (`spec->>'audit_source' = $4`) —
   the same structural guard the retraction helper carries for the same
   co-filing trap. Pinned by regex+args in the guardian control; mutation-
   verified (deleting the clause fails the control). Its candidate 3 (make
   `capability_gap` unclaimable) is the deeper root and is **filed as
   `bugs_open/284`** — the claimer is unidentified (18 blocked rows, recurring
   every few days), so per the 2026-07-31 ruling it carries the diagnosis gap
   stated, not a guessed cause. The 18 rows are NOT repaired yet (they would
   re-block until 284's mechanism closes).

### POST-ROLL CHECKLIST (whoever sees the next chassis roll)

1. Confirm the stamp: `git merge-base --is-ancestor d6d56e540 <stamp>` (and this
   round's commit, which supersedes it).
2. Apply `419_HOLD` per its header (rename → apply → `--record-only`).
3. Dispatch one improvement sweep (or one manual brief-fidelity run) against
   site `62b5978e…`; verify: zero new `audit_finding_%` rows; findings land as
   routed types under `audit_source='brief-fidelity-audit'`; any off-vocabulary
   category shows as `capability_gap` (`deferred`, empty handler) and in the
   result map's `unrouted_categories`.
4. Then: `115` closable (fixed AND live), and THIS file moves to `bugs_closed/`.

> **CORRECTED 2026-08-15c:** the wiring migration referenced above as `418_HOLD` is now
> **`419_improvement_loop_gains_brief_fidelity_HOLD.sql`** — a concurrent session took 418
> (and a second 417) while this lane's round-2 work was in flight; the council's
> architecture seat caught the collision. This lane's applied `417` keeps its filename
> (renaming a recorded migration breaks its ledger checksum); its header's "418_HOLD"
> pointer is deliberately stale and the renamed file's own header explains.
