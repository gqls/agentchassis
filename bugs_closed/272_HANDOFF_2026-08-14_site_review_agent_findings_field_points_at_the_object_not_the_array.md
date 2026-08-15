# 272 — site-review-agent's `write_strategic_findings` step can never file a
# finding when the LLM returns the response shape its own prompt asks for

**CLOSED 2026-08-15 — fixed AND live AND proven end-to-end on v1.0.1301.**
Fix commit `2a3ea3e2c` (council APPROVED round 1, corr `5a79843a`); shipped in
`v1.0.1301` (image label `revision=0115f2b4`, ancestry + pod-binary probe with
controls); live verification run `b2c82a25-7803-4ecf-a7d7-1a25099fae40`
(mortgagecalculator.co.uk, 2026-08-15): `result=object`, `llm_found=5`,
`items_created=5`, `audit_source='site-review'`, error NULL — the exact pair
that read 5→0 at filing now reads 5→5, and the 5 work items are the FIRST
`site-review`-stamped rows in the estate's history. Full trail in the sections
appended below. The `parseAuditFindings` extraction was reused within a day by
the `bugs_open/213` lane (commit `a620912f5` adds a `recognised` return for
silence-vs-nonsense discrimination).

**Filed 2026-08-14.** Found as a side effect of verifying `bugs_open/264`'s fix
(the audit_source config was correct; the run still produced zero work items,
for an unrelated reason — this file).

## The defect in one line

`write_audit_findings`'s finding-parsing switch handles a raw JSON **string**
(with a fallback that unwraps `{"findings": [...]}`) and a raw JSON **array**,
but has no case for a raw JSON **object** — and `site-review-agent`'s prompt
asks for exactly an object: `{"overall_score": 1-10, "summary": "...",
"findings": [...]}`. When the LLM complies with its own instructions, findings
extraction silently returns zero.

## Measured, live, 2026-08-13/14

Dispatched one live `site-review-agent` run against `mortgagecalculator.co.uk`
(orchestration `5fe7ff0d-6fe5-411c-920f-85055832fa81`, `COMPLETED`, no error) as
part of `264`'s verification. `collected_data`:

```sql
SELECT collected_data#>>'{strategic_review,result}' ... -- shape: {"summary": "...", "findings": [...5 items...], ...}
SELECT jsonb_typeof(collected_data#>'{strategic_review,result,findings}'),
       jsonb_array_length(collected_data#>'{strategic_review,result,findings}');
--  array | 5
SELECT collected_data#>>'{strategic_findings_written,items_created}';
--  0
```
`[MEASURED]` The LLM produced 5 well-formed findings under `strategic_review.result.findings`.
The step's own config (`write_strategic_findings.config.findings_field =
"strategic_review.result"`) resolves to the **object** one level above the
array, not the array itself. `write_audit_findings_action.go`'s parse switch
(lines ~537-587) is:

```go
switch v := findingsRaw.(type) {
case string:
    // ...unmarshal, and on failure try wrapper["findings"] as a fallback...
case []interface{}:
    // ...build findings from each map item...
}
// no case for map[string]interface{} — findings stays nil, falls through to:
if len(findings) == 0 {
    return map[string]interface{}{"items_created": 0, "reason": "no valid findings"}, nil
}
```
`findingsRaw` here is a `map[string]interface{}` (the parsed `{"summary":...,
"findings":[...]}` object) — it matches neither case, so `findings` is never
populated, and `items_created` is silently `0`. **The result map in this branch
also carries no `audit_source` key** (confirmed live: `collected_data#>>
'{strategic_findings_written,audit_source}'` is empty) — this is why `264`'s
fix could not be end-to-end verified for `site-review-agent` via a real work
item: the write step never reaches the point where it would stamp one.

## Why this reads as inconsistent with `bugs_closed/150` — and isn't, quite

`bugs_closed/150` (2026-07-29) recorded `site-review-agent.write_strategic_findings`
promoting **3** items in one observed run — i.e. it has worked at least once.
`[UNVERIFIED]` — the likely reconciliation, not confirmed: an LLM instructed to
"Respond with ONLY a JSON object" does not always comply byte-for-byte; if it
occasionally emits a bare JSON array instead of the wrapped object, the
existing `case []interface{}` would catch it and items would be created — the
non-deterministic sibling of exactly the gap this file describes, not a
contradiction of it. Whether the prompt shape or the config was different on
2026-07-29 was not checked here (out of scope for this file — flagged, not
chased, per this codebase's own norm of naming what's unmeasured rather than
guessing). Either way, the CURRENT prompt (verified live 2026-08-13, unchanged
in intent by migration `340_site_review_agent_loads_the_premise.sql`, which
states explicitly "the finding vocabulary is UNCHANGED") asks for the object
shape, and the current code cannot extract findings from it when the LLM
complies.

## Blast radius

Checked: this is `write_audit_findings`'s own general-purpose parsing code, so
it is shared by all producers of `bugs_open/264`. The other three
(`brief-fidelity-auditor`, `content-quality-auditor`, `visual-design-auditor`)
all prompt for a **bare JSON array** directly (`"Respond with ONLY a JSON array
of UP TO N findings"`), which the existing `[]interface{}` case handles —
confirmed live 2026-08-13: all three produced real work items with correct,
distinct `audit_source` values in the same verification pass that surfaced this
file. **`site-review-agent` is currently the only producer that asks for an
object**, so this is a single-agent defect today, not (yet) a second instance
of `264`'s four-way class — but the parsing gap itself is generic, so any
future auditor whose prompt returns a wrapped object inherits it silently.

## Fix candidates, ordered by what closes the door

1. **Change `site-review-agent`'s `findings_field` config to
   `"strategic_review.result.findings"`.** Config-only, no roll, smallest fix —
   but only repairs this one producer, and does nothing for the next agent that
   returns an object.
2. **Add the missing `case map[string]interface{}:` to the parse switch**,
   mirroring the string branch's existing `wrapper["findings"]` unwrap: if the
   map has a `"findings"` key, parse that; otherwise fall through to the
   existing zero-findings path. This is the candidate that makes the bad state
   unrepresentable — closes the gap for `site-review-agent` AND any future
   object-returning auditor, and needs a Go change + roll.
3. **Both — config fix as a config-only stopgap while candidate 2 goes through
   review**, given `site-review-agent` is dispatched by `improvement-loop` (a
   recurring pipeline) and has therefore been producing zero (or
   near-zero/nondeterministic) findings for an unknown but plausibly long
   window.

## How to verify a fix

Dispatch one `site-review-agent` run (`config:{agent_type:"site-review-agent"}`,
`input_data:{domain:"<any deployed site>"}` via the generic dispatch topic) and
check:
```sql
SELECT collected_data#>>'{strategic_findings_written,items_created}',
       collected_data#>>'{strategic_findings_written,audit_source}'
FROM orchestration_states WHERE orchestration_id = '<new run>';
```
`items_created` must be > 0 (when the LLM does return findings) and
`audit_source` must read `site-review` — both empty today.

## Related

- `bugs_open/264` — the audit_source-resolution fix this was found while
  verifying; unaffected by this file (its fix is correct and independently
  confirmed live for the three producers that DO create items).
- `bugs_closed/150` — the earlier, once-off sighting of this same step, from
  the opposite angle (it fired unexpectedly and broke a triage assumption,
  rather than not firing at all).

---

## FIX APPLIED 2026-08-14 (second session) — candidate 2, code only; candidates 1 and 3 REJECTED

**Taken up 2026-08-14 by the `bugs_open/272` session** (distinct from the filing
session, which moved on to an unrelated lane — checked its live transcript
before starting). Re-verified the bug first:

- `[MEASURED]` Live config unchanged: `findings_field = "strategic_review.result"`;
  live prompt still ends `Respond with ONLY a JSON object: {"overall_score": 1-10,
  "summary": "one paragraph", "findings": [UP TO 5 findings]}`.
- `[MEASURED]` All 6 surviving `site-review-agent` orchestrations (1 on 08-13,
  5 on 08-14 — the improvement-loop is dispatching these daily) show
  `jsonb_typeof(result)=object`, `jsonb_typeof(result.findings)=array`,
  `items_created=0`, `reason='no valid findings'`.
- `[MEASURED]` `SELECT count(*) FROM site_work_items WHERE
  spec->>'audit_source'='site-review'` → **0**, all history. (Only proves the
  post-264 window — pre-264 items were stamped `design-audit`.)

**No `090` diagnosis run — first-hand verification substituted (stated per the
2026-07-31 owner ruling):** the claim is local to one function, not
structural/cross-cutting; the filing session measured the live run and this
session independently re-verified all three legs (live config+prompt, 6/6 live
runs, the code at `write_audit_findings_action.go`) before changing anything.

### What shipped

`platform/orchestration/actions/write_audit_findings_action.go`: the inline
parse block is extracted into unexported, unit-testable functions —
`findingsFromList` (former array-case body, verbatim), `findingsFromString`
(former string-case body, verbatim) and `parseAuditFindings`, which adds the
missing **`case map[string]interface{}`**: it unwraps `v["findings"]`
(mirroring the string case's existing wrapper fallback), handling both an
array value and a JSON-string value. The zero-findings return keeps
`reason: "no valid findings"` byte-identical and **adds** `findings_field` and
`findings_type` (`%T` of the unmatched value), plus a `logger.Warn` with the
same keys — this defect was invisible precisely because the zero path said
nothing anywhere.

New `write_audit_findings_parse_test.go` pins every shape (wrapped object —
fails pre-fix; bare array — 150's regression guard; bare/wrapped/fenced JSON
strings; object-with-string-findings; no-findings-key; garbage; scalar) plus a
field-completeness test. **Mutation-verified**: repointing the map case at a
wrong key fails both tests; restoring passes.

### Why candidates 1 and 3 (repoint `findings_field` to `strategic_review.result.findings`) were REJECTED

1. `ExtractNestedField` cannot walk the segment `findings` into a **bare
   array**, so the repoint breaks the one shape that has ever produced items
   (`bugs_closed/150`'s 3 items were the LLM *disobeying* its prompt). Config
   `.result` + the code fix handles **both** shapes; `.result.findings` handles
   only the compliant one.
2. Applied migration `399_four_auditors_audit_source_resolves_to_a_real_value.sql:181-183`
   **guards** `findings_field IS DISTINCT FROM 'strategic_review.result'` with a
   RAISE — a repoint makes 399 read as drift to any future probe/replay. (The
   filing session did not surface this; it is the strongest reason the
   config-only stopgap was a trap.)
3. The code fix needs no migration at all, and no revert-migration later.

Cost accepted: zero-item runs continue until the next chassis roll (findings
are re-derived every run and deduped, so nothing is permanently lost).

### Council + commit

- Council submission: `Council-Submitted: 5a79843a-c6d7-46e7-a8a2-df4482bd716c`
  — verdict **APPROVED, round 1** (2026-08-14, "approved with 2 advisory
  objection(s) — none high-severity", 6 seats abstained). The commit predates
  the verdict under the `Council-Submitted:` trailer, which `098` resolves to
  approved automatically at report time — no amend.
- Commit: `2a3ea3e2c` (code + test + this file + 016b §9).
- **Advisory 1 (editquality, medium) — ANSWERED WITH EVIDENCE, no change:** it
  read the `{items_created:0, parse_error:...}` return (no `reason` key) as a
  behavioural fork introduced by this fix. It is not:
  `git show 2a3ea3e2c~1:platform/orchestration/actions/write_audit_findings_action.go`
  lines ~552-560 show the pre-fix code returning exactly that shape on the
  string-parse-failure path — the refactor moved it verbatim. The two zero-item
  shapes (`reason` vs `parse_error`) are a pre-existing contract; merging them,
  as the seat suggested, would be the actual behaviour change. Fair residue
  accepted: the new tests assert the extracted parser's returns, not the
  action's final result map (asserting that needs DB scaffolding out of
  proportion to a parse fix).
- **Advisory 2 (reuse_agent, medium) — RECORDED AS A NAMED FOLLOW-UP, not done
  here:** `create_blog_posts_action.go:101-146` carries a near-identical
  three-way shape switch, and a third copy invites drift (one gets a case the
  other doesn't — precisely this bug's genesis). A shared normalizer was looked
  for before duplicating (none exists: `ParseLLMJSON` is a lenient
  string-parser, `tryUnwrapMapPatterns` is unexported and scoped to
  `ExtractFields`) — the seat is right that the submission did not show that
  search. Consolidating into a `datahelpers` "LLM result → list, unwrapping a
  named key" helper is a real candidate, but it is a new shared seam touching
  N call sites — per the 2026-07-29/08-02 rulings that wants its own scoped
  change with the consumers named, not a rider on a bug fix. The sibling
  switches are enumerated in 016b §9's new entry for whoever picks it up.
- 016b §9 gains the transferable pattern (*an LLM-result consumer that
  type-switches on shape silently drops every shape it doesn't name*), with the
  sibling switches that still lack an object case listed.

### POST-ROLL CHECK 2026-08-15 — v1.0.1300 does NOT carry the fix; still OPEN

The overnight roll (pods `6c68fcc549-*`, 2 replicas, image `v1.0.1300`) was
checked and the fix **missed the build by 83 minutes**:

- `[MEASURED]` Image label: `docker image inspect aqls/agent-chassis:v1.0.1300
  --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'` →
  `a2a691213`, created 2026-08-14 21:21:49 BST. The fix commit `2a3ea3e2c` is
  22:44:42 BST. `git merge-base --is-ancestor 2a3ea3e2c a2a691213` → NO (and
  `a2a691213` IS an ancestor of the fix — consistent).
- `[MEASURED]` Pod binary probe, both directions of control clean:
  `2a3ea3e2c…` (full sha) absent from `/proc/1/exe`; positive control
  `write_audit_findings` present; negative control `deadbeef…` absent. (The
  `build provenance` startup line had scrolled out of retained logs — 11h-old
  busy pods — so the label + binary probe substituted, per LANDMINES.)
- Do **not** dispatch a verification run against 1300 — it will produce
  `items_created=0` from the old binary and reads exactly like a regression.

**The fix ships with the first agent-chassis build cut from a HEAD at or after
`2a3ea3e2c` (i.e. any build after 22:44 BST 2026-08-14 / v1.0.1301+).** The
fastest stamp read is the docker image label above; the pod-side check is
`git merge-base --is-ancestor 2a3ea3e2c <revision label or provenance stamp>`.

### Verify fixed-AND-live (the bar for moving this to bugs_closed/)

1. Confirm the fix shipped **on agent-chassis specifically**:
   `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor <this fix's commit> <the stamp>`.
   (Startup line scrolls — fall back to the known-sha binary probe with a
   control, per LANDMINES.)
2. Dispatch one `site-review-agent` run (any deployed site) and check, per the
   original section above: `items_created > 0` and
   `strategic_findings_written.audit_source = 'site-review'`. The run's
   findings should then appear:
   `SELECT item_type, status FROM site_work_items WHERE spec->>'audit_source'='site-review';`
3. If the LLM returns zero findings legitimately, the zero result now carries
   `findings_type` — `map[string]interface {}` there with a non-empty
   `result.findings` array would mean the fix did NOT ship (wrong binary), not
   that the bug regressed.
