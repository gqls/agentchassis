# 218 — the placeholder scan convicts JavaScript, and tool-recreation's failure path then discards the finished tool while the item reads `complete`

**Filed 2026-08-08 by the mortgagecalculator-adoption lane.** Two defects, one
incident shape. Defect A is FIXED (committed `201350e23`, inert until a
post-`201350e23` roll; council `Council-Submitted: a9ffed15`). Defect B is
UNFIXED and needs a deliberate decision.

## The incident

12 `needs_tool_recreation` items ran on mortgagecalculator.co.uk on 2026-08-05.
All 12 read `complete`. Three produced nothing: `tool-overpayment` and
`game-fact-finder` saved zero `page_components` and serve 404; a third
(`tool-portfolio`) was held by the bug-020 fabrication gate (correct behaviour,
separate thread of work — its `needs_human_review` item of 12:55:38 is the
record, but the detector's signals died with the purged orchestration row and
rotated pod logs, so it needs a fresh run to re-judge).

## Defect A — `checkPlaceholderPatterns` scans code as if it were prose

`platform/orchestration/actions/validate_page_content.go` lowercases the WHOLE
artefact HTML — inline `<script>`/`<style>` included — and runs plain substring
matches from `placeholderPatterns`. The bracket entries (`[name`, `[client`,
`[company`) are substrings of idiomatic code. Every hit is `severity=blocker`.

The three live convictions (all 2026-08-05, `agent_error_log`,
`step_name='validate_content'`, issue `type='placeholder_text'`,
`value='[name'`):

| when | domain | the "placeholder" |
|---|---|---|
| 01:00:58 | idea.uk | `querySelector('input[name="' + KEYS[i] + '"]:checked')` — CSS attribute selector |
| 12:21:33 | mortgagecalculator.co.uk | `if (fields[name]) fields[name].classList.toggle(...)` — object indexing |
| 12:59:20 | mortgagecalculator.co.uk | `rows.map(([name, val]) => ...)` — array destructuring |

`input[name=` appears in essentially every form-handling script, so any
self-contained tool with inline JS is convictable. The pattern list's own
comment records this exact class once before: bare `placeholder` was removed
for firing on `<input placeholder="...">` attributes and CSS class names.

**Fix — two rounds, read both:**

> **CORRECTED 2026-08-08 (council REVISE, round `a9ffed15`):** the first fix
> (`201350e23`) added a regex `stripScriptAndStyle` helper. The council's
> `reuse_agent` and `prior_art_librarian` seats caught what I missed: this
> file's claims checks already own a prose scope — `datahelpers.
> ExtractAssertionText`, called two lines below the placeholder scan — so the
> stripper was a second mechanism for a solved problem. Logged in
> `WRONG_CALLS.md`; the cheap check was one grep of `datahelpers/` before
> writing a new helper.

The landed fix (`b75f36601` + gofmt `f51ac6af8`): the placeholder scan reads
`ExtractAssertionText(html)` blocks (real HTML parse; script/style/code/pre/
head and attributes excluded) — which also dissolves the literal-`</script>`
edge case and stops code-sample convictions. One exemption: `<no value>` is
markup-shaped and parses away, so it keeps scanning the raw document (a test
pins this — moving it would have made the pattern silently inert). Every other
check still reads full HTML. Tests carry the three convicted snippets plus
guard-survival cases; a mutation run (scan raw HTML) fails the ignore-tests.
**Council trail (all under `a9ffed15`):** round 1 REVISE (reuse, answered by the
code above) · round 2 REVISE (plan-record issues only: verification recipe
mislabelled + two landmines; consumer reliance asserted, not measured; defect B
"filed" ≠ routed) · round 3 REVISE at 9/11 approve — **its editquality objection
was a real code catch**: the narrowing dropped `<head>`, so a `<title>`/meta
placeholder escaped. Fixed `35889819c`: `headProseBlocks` scans title +
description/og/twitter meta alongside body blocks; JSON-LD deliberately unread
(code-shaped — the collision class being fixed); mutation-pinned. Round 3's
gating seat escalated to "routing defect B is not a fix" — answered in round 4
with the process fact (defect B under ACTIVE diagnosis `c56b691d`; the 090
coverage rule forbids a parallel patch; the save-anyway-vs-cannot-complete call
belongs to that run). Round 4: **APPROVED 2026-08-08** ("2 advisory objections, none high-severity").
The `Council-Submitted: a9ffed15` trailer on all four code commits resolves to
this approval at 098-report time — nothing owed, no amends. Advisories worth a
future reader's minute: (a) compliance notes JSON-LD is now a declared blind
spot for THIS check while being exactly where fabricated prices/ratings would
surface machine-readably — a structured-data detector is its own design, not a
prose-scan rider; (b) debug_historian asks whether the mutation runs were
behavioural — they were: both sed mutations compiled and ran (scan-raw-HTML,
drop-head-append), failing 3 and 6 assertion cases respectively.
The consumer-reliance question is now MEASURED: the check's entire recorded
history is 46 convictions (2026-07-15→08-07); 43 are visible-prose contexts the
new scope preserves, the only 3 non-prose ones are the JS false positives fixed
here — zero attribute- or code-context true positives ever, for any consumer.

**All four live consumers inherit the change** (from live definitions):
`page-build-handler/validate_content`, `content-reviewer/validate_content`,
`tool-recreation-handler/validate_tool`, `report-builder/validate_page`.
Stated narrowing: placeholders living only in attributes, code samples or
script-emitted strings are no longer caught by this check.

**Verify live** after a post-`f51ac6af8` roll — two landmines are baked into the
commands (the council caught them in an earlier draft of this recipe): enumerate
the pods, never trust a label selector (one image serves many services); and
`grep -c` prints NOTHING and exits 1 on a zero count, so defuse it:

```bash
for p in $(kubectl get pods -n ai-persona-system -o name | grep agent-chassis); do
  echo -n "$p  strip="; kubectl exec -n ai-persona-system ${p#pod/} -- sh -c \
    'strings /app/agent-chassis | grep -c stripScriptAndStyle || echo 0'
done
# required: strip=0 on EVERY replica (round 2 REMOVED the symbol — a 1 means the
# image carries round 1 only). Positive control that the binary is readable:
# grep -c ExtractAssertionText || echo 0 must be non-zero.
```

Then a re-run of the two dead recreations saves components
(`SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE
p.name IN ('tool-overpayment','game-fact-finder')` goes 0 → >0).

## Defect B — `validate_tool`'s failure routing is dead config, so a failed validation discards the recreation and completes the item

> **ROUTED 2026-08-08, not just filed** (the council's gating seat was right that
> a prose filing rots): 090 intake `315f7f88-ca07-4ab0-b3de-0ecd11b0edee`,
> claimed by `diagnose-dispatch-loop` as run `c56b691d-43c3-4ad2-ab48-f043689100ea`.
> Read that run's verdict before starting independent work here.

The seed (`docs/agent_docs/sql_for_agents/099_tool_recreation_handler.sql:153`)
puts `"error_step": "save_sections"` at STEP level on `validate_tool` — the
stated intent is that this validation is advisory on the recreation path.
The LIVE `agent_definitions` row instead carries `error_step` nested INSIDE
`config`, where the engine never reads it (`platform/messaging/processor.go:433`
reads `error_step` from the step map). Step-level routing is therefore absent:
a `validate_tool` failure falls into `routeToErrorStepOrFail`'s fail path, the
recreate output (a full, paid-for LLM response — 9–12k output tokens each, per
`llm_call_log`) is discarded, and the work item still ends `complete`.

The live row has no snapshots and its `updated_at` (2026-08-08 08:54) was
bumped by this morning's roll, so who moved the key and when is not
reconstructable. [INFERRED: the misplacement is drift, not design — the seed
and the live row disagree and only one can be the intent.]

**Fix candidates, ordered by what closes the door:**
1. Make a workflow run that discards its output incapable of marking the item
   `complete` (engine-level; kills the whole "complete with nothing" class —
   the standing `complete ≠ artefact` landmine made structural).
2. Restore `error_step` to step level per the seed (one-key UPDATE on the live
   row; live immediately; but it makes a REAL validation failure save anyway —
   which is the seed's stated intent, and defensible only while Defect A's fix
   keeps validation honest).
3. Do nothing and rely on Defect A's fix — validation stops false-failing, so
   the dead routing stops mattering except for genuine failures, exactly the
   case where discard-vs-save needs a human decision anyway.

## Diagnosis provenance (per the 2026-07-31 owner ruling)

Run through the 090 loop before filing: intake `0de6e0e4`, run `86721efd`,
verdict **UNVERIFIABLE (stopped: iteration-cap)** — not refuted. The loop
independently confirmed the workflow-shape sub-claims (the `validate_tool`
step exists, runs `validate_page_content` over `completeness_check.clean_html`;
the two pages sit `needs_rebuild`/0 components while their items read
`complete`) and named exactly two items of still-needed evidence, both of
which this file supplies first-hand: (a) the `validate_tool` failure rows —
`agent_error_log` 12:21:33 work_item `67baa1cb` (tool-overpayment) and
12:59:20 work_item `ece6449d` (game-fact-finder), both
`agent_type='tool-recreation-handler'`, `step_name='validate_tool'`,
"content validation failed: 1 blockers, 0 errors"; (b) the paired
`validate_content` rows whose `context.issues` carry `placeholder_text`
`value='[name'` with the JS location snippets quoted above.

## Oddity noted in passing, unowned

The `result` payloads stored on the three dead recreation items describe the
WRONG artefacts (tool-overpayment's result describes a stamp-duty calculator;
game-fact-finder's describes a "legal-disclaimer" page proposal) — visible in
diagnosis bundle `86721efd` iteration 5's data_request output. [UNMEASURED:
whether this is result cross-wiring in the dispatch loop or reuse of the item
row by a different producer.] Worth its own look before anyone trusts
`site_work_items.result` on this path.
