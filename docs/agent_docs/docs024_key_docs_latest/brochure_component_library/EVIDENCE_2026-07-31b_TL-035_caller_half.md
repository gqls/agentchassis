# EVIDENCE — TL-035 caller half: the council's five objections, answered with checks

Council gate, submission `2c895dd1-adae-4f8e-8acf-4592b8ca3981`, **APPROVED round 1**
("approved with 1 advisory objection(s) — none high-severity"). Eleven seats approved,
one objected. Five objections in total, one medium and four low.

**One of them changed the code.** The other four were diligence requests — "you asserted
this, I could not check it from here" — and every one of them is answerable with a
command, so this file runs the commands rather than restating the assertion. Same
discipline as the adapter half's `EVIDENCE_2026-07-31_TL-035_capture_renders.md`, for the
same reason: answering an evidence objection in prose reproduces the fault.

---

## 1. bug_historian, MEDIUM — the parser FAILS OPEN on an envelope it does not know

> *"`extractShotList`'s four-envelope fallback returns silently empty if the adapter's
> actual reply shape doesn't match any of the four guessed paths … The plan reuses rather
> than fixes this trait and now rides a second, less-tested key (`renders`) on top of it
> with no logging/error when zero paths match, so a future adapter envelope change would
> silently drop renders with no signal distinguishable from 'not requested'."*

**Conceded, and fixed in code rather than in a log line.** The objection is exactly right
about the shape: renders are the one key whose empty state is *normally* correct, so an
empty renders list carries no information at all. Logging "found nothing" would fire on
every default-off run and mean nothing on any of them.

The fix ties the shot lists to the envelope `results` was actually found in:

```go
raw, envIdx := extractAtFirstMatch(collected, envelopePaths(field, "results"))
...
v.Shots   = extractShotList(collected, field, "screenshots", envIdx)
v.Renders = extractShotList(collected, field, "renders", envIdx)
```

`extractShotList` tries `paths[envIdx]` first and then the rest, so it can only ever find
more than the old independent order did — no regression is possible. What it buys:

**an adapter shape this code does not recognise can no longer drop renders silently
WITHOUT ALSO hiding results, and empty results is already a hard error one caller up** —
`judge_acceptance_results: no results at %q (or its response fallbacks)`. The silent drop
is now unrepresentable rather than merely reported, which is the stronger of the two
answers and the one the "order fix candidates by what closes the door" rule asks for.

The four path literals also stopped being three copies (`results`, `screenshots`,
`renders`) and became one `envelopePaths(field, key)` — keeping three lists in step was
itself the drift risk the objection describes.

**Proven, not asserted.** `TestShotListsFollowTheEnvelopeResultsCameFrom` builds a reply
carrying results at the *flat* path and a stale renders list at the *more specific*
`response.data` path — the case the old order got wrong — and asserts the flat one wins.
Mutant (`if envIdx >= 0` → `if false`, with `_ = envIdx` so it still compiles):

```
--- FAIL: TestShotListsFollowTheEnvelopeResultsCameFrom (0.00s)
    renders must come from the envelope results came from (flat), got "s3://b/nested.png"
```

## 2. bug_historian, low — the ordering landmine is documented, not enforced

> *"Risk #4 names a real ordering landmine (DB config write racing the chassis roll) and
> the plan's own mitigation is 'documented in the register, not enforced in code' — this
> is the documented recurring shape of a recorded decision with no enforcement point."*

**Accepted as characterised, and deliberately not fixed here.** The seat itself notes the
failure mode is silent-*inert* rather than silent-*error*: setting the key early costs a
config row that does nothing until the roll, and the roll fixes it with no intervention.

Enforcing it in code would mean the chassis refusing or warning on a config key its binary
does not understand — which is a **general** mechanism about config/binary version skew,
applying to every action and every key, not to this one. That is architecture scope by the
2026-07-29 ruling (it would change what the shared config mechanism guarantees), and
smuggling it in under a wire-connection fix is precisely the pattern `bugs_closed/124` was
vetoed for. Recorded as a landmine on TL-035 and in the RUNBOOK's ordered procedure, and
named here so the next person to want it knows it is a real, separate piece of work.

## 3. editquality, low — is "one per PASSING (url, profile)" exact?

> *"The 'one per PASSING (url,profile)' semantics for Renders is inferred from a single
> quoted line of run_checks_action.go (:400); worth confirming the adapter can't also emit
> a render for a failing profile."*

**Confirmed. It cannot.** The routing is a single boolean returned by the capture and
consumed by its only caller:

```
$ grep -n "isFailure\|failing bool" internal/adapters/browserrunner/run_checks_action.go
386:  req RunChecksRequest, results []CheckResult, profile, url string, urlIdx int) (ref ScreenshotRef, failing bool, ok bool)
403:  isFailure := len(failed) > 0
424:  isFailure, true                       <- the returned `failing`
```

`isFailure` is `len(failed) > 0` over that run's own results, and `Execute` files on it:
`if failing { out.Screenshots = … } else { out.Renders = … }`. A run with any failing check
therefore cannot reach `Renders` by any path. The adapter half already carries a test for
exactly this, which the seat could not see from its side:

```
$ grep -n "func TestFailingRunGoesToScreenshotsEvenWithRendersOn" internal/adapters/browserrunner/run_checks_action_test.go
843:func TestFailingRunGoesToScreenshotsEvenWithRendersOn(t *testing.T) {
```

## 4. guardian + bug_historian — is `tool-acceptance-agent` really the only caller?

> guardian: *"Plan asserts tool-acceptance-agent is the only live agent referencing
> request_browser_run, but this is stated from a prior manual query, not verified this
> round — worth confirming since it's the blast-radius boundary the whole approval rests
> on."*
>
> bug_historian, missing: *"Whether any OTHER agent … could reach request_browser_run via a
> generic/spawn path not visible to a default_config LIKE scan (the SQL check only covers
> agent_definitions.default_config text, not ad-hoc spawn_agent calls with inline
> input_mapping)."*

**Re-run this round, and widened past my own original query in both directions.**

**(a) Every definition row, not just the active non-snapshot ones** — my submission's query
filtered on `is_active AND NOT is_snapshot AND deleted_at IS NULL`, which is the narrow
world the guardian is right to distrust:

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) AS snap, deleted_at IS NOT NULL AS deleted
  FROM agent_definitions WHERE default_config::text LIKE '%request_browser_run%';
--      type          | is_active | snap | deleted
-- tool-acceptance-agent |    t    |  f   |    f      (1 row)
```

One row across the **whole table**. No snapshot, no soft-deleted, no inactive definition
carries it.

**(b) The ad-hoc spawn path bug_historian names.** A `LIKE` scan of `default_config` cannot
see an inline `input_mapping`, so the repo was searched with no file-type filter:

```
$ grep -rn "request_browser_run" --include=* . | grep -v "^./docs/" | grep -v "\.git/"
platform/orchestration/actions/tool_acceptance_actions.go:70:  RegisterActionInputSpec("request_browser_run", …)
… (9 further hits, ALL inside tool_acceptance_actions.go / its test: the registration,
   the doc header, and its own error strings)
```

Every non-docs hit is the action's own definition. No workflow JSON, no seed, no Go call
site constructs a step naming it. The docs hits are its two seed files (`145`, `147`) and a
rename note.

**(c) What has actually run, which neither seat asked for and settles it best.** A caller
invisible to both checks above would still leave orchestration rows:

```sql
SELECT owner_agent_type, count(*) AS runs, min(created_at)::date, max(created_at)::date
  FROM orchestration_states
 WHERE orchestration_name ILIKE '%acceptance%' OR owner_agent_type='tool-acceptance-agent'
 GROUP BY 1;
--  tool-acceptance-agent | 4 | 2026-07-30 | 2026-07-31
```

Four runs, all history, one agent. **The blast-radius claim holds on definitions, on code,
and on run history.**

**The honest limit, stated rather than glossed:** (b) proves no caller exists *in this
repo at this commit*. A step could still be written into `agent_definitions` by hand
tomorrow — which is true of every action and is what the default-off design protects
against: such a caller gets exactly today's behaviour unless it also sets the key.

## 5. prior_art_librarian, low — does `ab21beac` exist and say what the register quotes?

> *"Register update claims TL-035's adapter half was 'APPROVED r1 as submission ab21beac' —
> worth confirming this council round exists and said what's quoted, since a false REGISTER
> claim quoted as contract is the named adjacent defect (bugs_closed/031)."*

**A good objection to raise on principle, and the answer is yes.**

```sql
SELECT correlation_id, metadata->>'decision', created_at
  FROM diagnosis_artifacts WHERE correlation_id LIKE 'ab21beac%' AND kind='council_report';
-- ab21beac-b5cd-43a8-a66f-c73ef33b6d49 | approved | 2026-07-31 08:19:45+00
```

The same seat also flagged, as `missing`, that it could not run the precedent check on
whether that round had already litigated merging `renderLine` into `evidenceLine`. **It had
not** — `renderLine` did not exist in that round; the adapter half stopped at the wire
format (`CaptureRenders`, `Renders`) and wrote no note bodies. The separate-function
decision is made and argued for the first time here, which is why it is stated at length in
the code comment rather than cited.

## 6. reuse_agent, low — same precedent gap, and it points at something real

The seat approved "contingent on the precedent check finding nothing that contradicts the
separate-function choice". Per §5 it does not. Worth recording the seat's framing though,
because it is the useful reading: two functions that differ only in a string literal are
normally a reuse failure, and this one is defensible **only** because the strings are
claims about different things — `evidenceLine` asserts a failure, `renderLine` explicitly
disclaims one. If a future change makes both lines say the same thing, they should merge,
and that is the test for whether the split is still earning its keep.

---

**Trailer:** the commit that shipped the caller half (`9cc63c775`) carries
`Council-Submitted: 2c895dd1-adae-4f8e-8acf-4592b8ca3981`, written before the verdict
landed. It is now APPROVED and `098` resolves the correlation at report time, so the commit
is credited without an amend. The follow-up commit answering objection 1 carries
`Council-Reviewed:` — that verdict has been read, and this file is the reading.
