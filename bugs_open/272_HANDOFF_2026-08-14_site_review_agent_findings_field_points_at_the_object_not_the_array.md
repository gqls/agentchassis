# 272 — site-review-agent's `write_strategic_findings` step can never file a
# finding when the LLM returns the response shape its own prompt asks for

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
