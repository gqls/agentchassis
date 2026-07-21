# BUG 044 — plan_sections carries an empty-schema component by NAME HEURISTIC, silently

**Filed:** 2026-07-20 · travelling-docs thread · sibling of `bugs_open/024`
**Severity:** medium — latent today, silent when it fires, and it fires on a
component that does not exist yet (so nothing will surface it at the time).
**Status:** ✅ CLOSED 2026-07-21 — fixed AND live on chassis `v1.0.1146`.

> ⚠️ **Duplicate number.** Two unrelated cases share `044`. THIS one is the
> `plan_sections` slug (now closed). The OTHER — *no capability inventory /
> dormant agents undetectable* — is a different case and stays in `/bugs_open/`.
> Resolve 044 by slug, never by number.

---

## CLOSURE — 2026-07-21 (fixed AND live)

**Fix candidate 1 (preferred) was taken.** `planSection`'s empty-schema branch
now calls the SAME predicate the rerender escalation guard uses,
`isSelfContainedSection(comp)` (`component_level == 'tool'` AND empty
input_schema), BEFORE the function-name heuristic. A self-contained tool is
exempted by its explicit marker and returns `"ready"`; the name heuristic can no
longer carry a tool to `deferred`. Two call sites of the "is this emptiness
legitimate?" judgement now share ONE predicate, so they cannot drift apart — the
exact residual `bug_historian` flagged.

**Where the fix lives.** The Go change (`plan_sections_action.go`, the
`isSelfContainedSection(comp)` exemption in `planSection`'s empty-schema branch)
was swept into HEAD as part of the `v1.0.1146` build commit
(`fe2ba5e52`, a multi-session sweep). It now sits just after the concurrent
`bugs_open/026` change that reads fields via `schemaContentFields(...)`; the two
compose correctly — `schemaContentFields({})` returns `(nil,false)`, so the
empty-schema branch is still entered for a tool and the exemption still fires.

**Live verification (2026-07-21).** Running pod
`agent-chassis-55bbccfdbc-xrkv6`, image `docker.io/aqls/agent-chassis:v1.0.1146`:
```
strings /app/agent-chassis | grep -c "self-contained tool component"   # 1  (the NEW Reason literal, unique to this fix)
strings /app/agent-chassis | grep -c "needs regeneration with content fields"  # 1  (positive control — pre-existing)
strings /app/agent-chassis | grep -c "self-contained tool section"     # 1  (rerender's DIFFERENT wording — negative control, distinct string)
```
The discriminating literal `self-contained tool component` is created only by
this change (the rerender path uses `self-contained tool section`), so its
presence proves THIS fix is in the pod, not merely that the file was touched.

**Regression test** (`plan_sections_tool_defer_test.go`, committed `8ef063e6a`).
Exercised at the `planSection` level the "How to verify" section below asks for.
Proven discriminating against an isolated tree built from `git archive HEAD`:
- WITH the exemption → the tripping-name tools (`tool-content-planner`,
  `tool-blog-outliner`, `tool-body-copy-scorer`, `tool-article-summariser`,
  `tool-text-fitter`) all return `"ready"`. PASS.
- With the exemption stripped (026 kept) → they return `"deferred"`. FAIL,
  reproducing the bug exactly.
- The benchmark tool stays `"ready"` (unchanged); a genuine schemaless SECTION
  with a content name (`article-body`, `content-block`, `text-body`) still
  `"deferred"` — the `bugs_closed/004`/`005` blanked-article protection is
  preserved (it keys on `component_level != 'tool'`, so sections are untouched).

**Latency claim re-checked live (2026-07-21).** Of the active
`component_level='tool'` components, the 14 with an empty/NULL input_schema are
exactly the set that reached this branch, and **none** trips the five substrings
— so there was never a live reproduction, and the fix changes no current tool's
observable status (they now reach `"ready"` via the marker instead of by falling
through the name check). The 15 tools that DECLARE fields never enter this
branch at all.

**Fix candidate 2 (observability) NOT done** — deliberately out of scope. Making
the *consequence* legible (surface when a carried section's component template is
newer than its stored render) is a broader change to the re-render path and is
the same idea as 024's fix candidate 4; candidate 1 structurally prevents the
tool mis-carry, which is what closes this case. If the class recurs on a
non-tool path, that observability is the follow-on worth building.

**Pattern recorded.** The transferable shape — *"one call site of a shared
judgement gets the rigorous fix; the sibling stays heuristic"* — is added to
016b §9.

*Original filing below, unchanged.*

---

## Why this exists as its own file

`bugs_open/024` fixed the empty-`content_data` escalation guard in
`rerender_page_sections_action.go` by keying an exemption on the EXPLICIT
`component_level='tool'` marker rather than on a heuristic about field shape.

A **second, independent** piece of code makes the *same judgement* — "is this
component's emptiness legitimate?" — one function away, and still decides it by
**pattern-matching the component's function name**.

The council's `bug_historian` seat objected to 024's fix on exactly this ground
(round 5, MEDIUM, verdict REVISE):

> "one call site of a shared 'is this emptiness legitimate' judgment gets the
> rigorous explicit-marker fix, while a second call site making the identical
> judgment stays heuristic and will silently mis-carry a future self-contained
> component whose Function name happens to avoid those five substrings … that
> residual should at minimum be tracked as its own work item before this is
> treated as done, not merely disclosed in prose."

It also noted this platform has had a second independent silent-drop path for
the same content class **before** (the page assembler's visible-content filter
alongside the rebuild's own filter). This is that pattern recurring. Filing it
rather than leaving it in a `risks` paragraph is the whole point.

## The mechanism

`plan_sections_action.go:1141-1160` (the `planSection` empty-schema branch),
verified against the working tree 2026-07-20. For a component whose
`input_schema` declares no `fields`, the section is marked **`deferred`** —
which downstream means it is **CARRIED unchanged, not re-rendered** — when
`strings.ToLower(comp.Function)` contains any of:

```go
needsContent := strings.Contains(funcLower, "article") ||
    strings.Contains(funcLower, "content") ||
    strings.Contains(funcLower, "body") ||
    strings.Contains(funcLower, "text") ||
    strings.Contains(funcLower, "blog")
```

A carried section keeps its stored `rendered_html`. So for any component that
matches, a durable template fix is computed and discarded — the *identical*
end-state to `bugs_open/024`, reached by a different route.

> **CORRECTED 2026-07-20, before this file was first committed.** My council
> submission and the first draft of this file both cited
> `plan_sections_action.go:1090-1108` and called the deferral "silent". Both
> were wrong, and I caught it only by opening the function to check the line
> numbers before filing. (a) The line numbers were stale — carried forward from
> the round-4 submission and then shifted further by my own `toolTemplateValid`
> edit in the same file. (b) It is **not silent at the deferral itself**: it
> logs `plan_sections: content component has empty schema, deferring` at
> **Warn**, with the function and section, and sets
> `item.Reason = "component has empty input_schema — needs regeneration with
> content fields"`. What is invisible is the **consequence** — that a deferred
> section is carried, so a template fix is discarded — not the decision. Fix
> candidate 2 below is correspondingly narrower than I first wrote.
> Cheap check that would have caught both: open the function.

## Why it is latent rather than live

`tool-loot-table-balancer` matches none of the five substrings, so it returns
`"ready"` and re-renders. Verified 2026-07-20 against the live population: of
the **27 active `component_level='tool'` components**, none has a function name
containing any of the five tokens.

So there is no reproduction available today. That is precisely what makes it
worth filing rather than fixing-and-forgetting: **the first component that
trips it will be a new one**, and the failure is silent — no error, no work
item, no failed status. It will present as "the fix didn't take", which is
exactly how 024 presented and cost three cycles.

A `tool-content-planner`, `tool-blog-outliner` or `tool-body-copy-scorer` would
trip it on the day it is born.

## Fix candidates

1. **Preferred — key it on the same explicit marker.** `isSelfContainedSection`
   (`rerender_page_sections_action.go`) already encodes the judgement:
   `component_level == 'tool'` AND empty `input_schema`. `planSection` has the
   component in hand, and `component_level` is already SELECTed by
   `loadSectionComponents` and carried on `componentInfo.Raw`, so this needs no
   new query. Two call sites, one predicate.
2. **Make the CONSEQUENCE legible, not the decision.** The deferral already
   logs at Warn (see the correction above), so the gap is downstream: nothing
   says "a template fix for this component was discarded". The re-render path
   is where that belongs — a carried section whose component template is
   *newer* than the stored render is the signal worth surfacing, and it would
   have caught `bugs_open/024` on the first cycle rather than the third.
   Note this is the same idea as 024's fix candidate 4.
3. **Narrower, if 1 is judged too broad:** keep the name heuristic but require
   `component_level != 'tool'`, so a tool can never be name-matched into a
   carry.

## How to verify a fix

There is no live reproduction, so it needs a constructed one:

```sql
-- Confirm the latency claim still holds (no active tool trips the heuristic):
SELECT name, function FROM content_components
WHERE component_level = 'tool' AND is_active = true
  AND (function ILIKE '%article%' OR function ILIKE '%content%'
    OR function ILIKE '%body%'    OR function ILIKE '%text%'
    OR function ILIKE '%blog%');
-- 2026-07-20: 0 rows. If this ever returns rows, the bug is LIVE for them.
```

Then a unit test at the `planSection` level: an empty-`input_schema` component
with `component_level='tool'` and `function='tool-content-planner'` must return
`"ready"`, not deferred.

## Relationship to other records

- **`bugs_open/024`** — same end-state (a computed render discarded), different
  route. 024's Go fix shipped in v1.0.1140; this one is untouched.
- **`bugs_closed/004` / `bugs_closed/005`** — the blanked article bodies. The
  name heuristic here is plausibly a descendant of that fix, which is why it
  matches on `article`/`body`/`content` in the first place. Any fix must keep
  genuine article-class sections protected.
- **016b §9** — the transferable pattern is *"one call site of a shared
  judgement gets the rigorous fix; the sibling stays heuristic"*. Worth an entry
  once this is fixed, because the council found it by looking for exactly that
  shape rather than by reading the code.
