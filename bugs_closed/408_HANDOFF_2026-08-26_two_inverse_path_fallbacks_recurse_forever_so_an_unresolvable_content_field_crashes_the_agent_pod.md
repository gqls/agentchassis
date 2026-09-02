# 408 — two inverse path fallbacks recurse forever, so an unresolvable `content_field` crashes the agent pod with a stack overflow

**Filed 2026-08-26 by the `bugs_open/357` lane**, found while testing 357's phase-3
precondition. **Not a 357 defect** — 357 is only how I reached a missing field. This is a
general fault in the fleet's page-composition path and it takes the pod down.

---

## 1. The defect

`extractFieldValue` (`platform/orchestration/actions/multipage_actions.go:1185`) has two
fallbacks for an unresolvable path, and **they are exact inverses of each other with no depth
bound and no visited set**:

```go
// A — path CONTAINS ".response." -> strip it and recurse          (:1207-1214)
if strings.Contains(fieldPath, ".response.") {
    fallbackPath := strings.Replace(fieldPath, ".response.", ".", 1)
    return extractFieldValue(data, fallbackPath, logger)
}

// B — path does NOT contain ".response." -> add it and recurse    (:1215-1223)
parts := strings.Split(fieldPath, ".")
if len(parts) >= 2 && !strings.Contains(fieldPath, ".response.") {
    responsePath := parts[0] + ".response." + strings.Join(parts[1:], ".")
    return extractFieldValue(data, responsePath, logger)
}
```

A removes `.response.`; B puts it back. Each call satisfies the other's condition, so:

```
page_content_0.response.page_html  --A-->  page_content_0.page_html
page_content_0.page_html           --B-->  page_content_0.response.page_html
… forever, until the goroutine stack passes 1 GB and the runtime kills the process
```

**Every cycle emits a `Warn`**, so the pod also floods its own logs on the way down.

## 2. Evidence [MEASURED 2026-08-26]

Pod `agent-page-rebuild-147e9f90-4xf8j`, container terminated `exitCode: 2`, `reason: Error`:

```
runtime: goroutine stack exceeds 1000000000-byte limit
runtime: sp=0xc0215b8360 stack=[0xc0215b8000, 0xc0415b8000]
fatal error: stack overflow
```

**The whole log of that container is one repeated line — 12,654 of 12,654:**

```json
{"level":"warn","caller":"actions/multipage_actions.go:1202","msg":"Field not found in path",
 "step_name":"build_pages_loop_iter_0_assemble_page","action":"assemble_page",
 "field":"page_html","full_path":"page_content_0.response.page_html"}
```

`grep -c 'Field not found in path'` = **12654**; total lines = **12654**; distinct payload = **1**
(`"field":"page_html"`). The `full_path` is the ORIGINAL on every line, which is the A→B→A
cycle returning to its start every second hop.

**Three orchestrations were destroyed by it** — all abandoned mid-flight at
`build_pages_loop_iter_0_assemble_page` and only released hours later by the stale reaper
(`"reaper: stale EXECUTING_STEP for >4h"`): `e0c2d505-9875-4347-a718-a852f32ec6b7`,
`5a0cad41-fe0c-4636-9b2d-9c942486019c`, `bf29ec85-8ef9-457a-9366-1ca121a95810`. Pod restart
count reached **2** in 27 minutes.

## 3. Why the caller is innocent — and why that makes the fix easy

`AssemblePageAction` (`multipage_actions.go:17`) calls it at `:106` and **already handles the
missing case correctly**:

```go
content := extractFieldValue(params.CollectedData, contentField, params.Logger)
if content == "" {
    // No content found - treat as skipped rather than error
    // This allows the loop to continue with other pages
```

The contract the caller relies on is *"returns the value, or `""` if it cannot be found"*. The
function honours that on every path **except** the two fallbacks, where instead of returning
`""` it never returns at all. **Nothing about the call site needs to change.**

## 4. How it is reached

Any `assemble_page` whose `content_field` does not resolve, where the path has ≥2 parts. The
route I hit: a page whose content writer returned `skipped: true`,
`reason: "no sections defined for page"`, `section_count: 0` — so
`page_content_0.response.page_html` was legitimately absent. **A page the writer skips is a
NORMAL outcome the caller was written to tolerate**, and it kills the pod.

⚠ **This is not rare-path-only.** `content_field` is operator-set config
(`"content_field": "page_content.response.page_html"` on `page-rebuild`'s `assemble_page`), so
a typo in any agent's step config produces the same crash rather than a skipped page.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Give the recursion a bounded, monotone shape.** Try each path form at most once — e.g.
   resolve against an explicit ordered candidate list (`original`, `stripped`, `with-response`)
   in a loop with no self-call at all. This makes non-termination structurally impossible
   rather than merely unlikely, and is the smallest diff.
2. **Pass a depth/visited parameter.** Correct, but it keeps a recursive shape whose safety
   depends on every future editor preserving the guard — weaker than (1) for the same effort.
3. **Return `""` from the second fallback instead of recursing.** One line, kills the cycle,
   but silently drops the one genuinely useful case (a path written without `.response.` that
   needs it). Acceptable only as an emergency stop.
4. ~~Raise the stack limit / add a pod memory limit~~ — **rejected**: it converts a fast crash
   into a slow one and leaves the log flood and the 4h orchestration abandonment intact.

**Also worth fixing regardless of which is chosen:** the `Warn` inside the loop. A log line per
recursion turned one failed lookup into 12,654 lines. Log the failure once, at the top-level
call, not per attempt.

**This is `platform/` Go, so the fix wants the council gate** (and it is inert until an image
is built and rolled).

## 6. How to verify

Unit-level, and it needs no cluster:

```go
// must RETURN, and must return "" — today it never returns
got := extractFieldValue(map[string]interface{}{"page_content_0": map[string]interface{}{
    "response": map[string]interface{}{"skipped": true},
}}, "page_content_0.response.page_html", zap.NewNop())
// want: "" — and the test must have a TIMEOUT, because the failure mode is non-termination,
// not a wrong value. A plain assert would hang the test binary rather than fail it.
```

⚠ **A test asserting the return value alone cannot fail today — it never gets to assert.**
Run it under `go test -timeout 30s` and treat the timeout as the failing signal, or the suite
reports a hang rather than a defect.

End to end: rebuild a page whose writer skips, and assert the pod's restart count is unchanged
and the orchestration reaches a terminal state rather than sitting at `assemble_page`.

## 7. Relations

- **`bugs_open/357`** — how it was found. It is currently **blocking 357's phase 3**: the
  precondition that an adopted row survives a rebuild cannot be tested while every rebuild of
  that page crashes the pod. See
  `docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/HANDOFF_2026-08-26_continue_here.md`.
- **`bugs_open/406`** — the other defect the same investigation turned up; unrelated mechanism.
- The stale-orchestration reaper is the only thing that releases the wedged runs, at **>4h**.

## 8. Addendum 2026-08-26 evening (357 lane) — caller census, and what fixing this exposes

- **`extractFieldValue` has exactly 1 caller as of 2026-08-26** — `AssemblePageAction` at
  `multipage_actions.go:106` (`grep -rn 'extractFieldValue(' platform/ internal/ pkg/`). So
  candidate 1 cannot break another call site: the `""` contract is already what the sole
  caller handles.
- **Fixing this converts 357's canary crash into a clean full-chain skip, and the 357
  handoff's canary pass condition then passes VACUOUSLY**: assemble returns the skip shape →
  `git_commit` skips (`checkUpstreamSkipped`) → `save_sections` runs but exits at
  `len(sections)==0` (`save_page_sections_action.go:344`/`:401`) before the Layer 2
  carry-forward → the run COMPLETES with the row untouched. Whoever fixes this: do not let a
  green cv1 canary be read as precondition-4 evidence for migration 578 — see
  `docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/HANDOFF_2026-08-26b_continue_here.md`.
- **v1.0.1345 (rolled 2026-08-26 ~20:36Z) does NOT carry a fix** — no commit touches this
  file after `c4baa53e7`, and the recursion is present at HEAD (`:1213`/`:1223`).

## 9. FIX COMMITTED 2026-09-02 — `6e2d4a039`, Council-Submitted `3918db52` — OPEN until live at the pod

**Shape:** candidate 1, generalised.** Unrolling the old recursion showed the §5 sketch's flat
three-element list `[original, stripped, with-response]` was NOT semantics-preserving in the
general case: on multi-occurrence (`a.response.b.response.c`) and non-first-segment
(`a.b.response.c`) shapes, the old code tried MORE forms before cycling and terminated
successfully if any resolved — a flat list would have silently dropped those routes. The
implemented candidate builder generates the old recursion's exact terminating tried-set
(original; each `.response.` stripped in turn; the fully-stripped form with `.response.`
re-inserted after the first segment, skipped when identical to the original). Caught at plan
time by tracing the recursion, before any code was written; logged in `WRONG_CALLS.md`
2026-09-02.

- New pure walk helper `walkFieldPath` (three-valued: resolved / key-missing → next candidate /
  non-map-mid-path → end outright, preserving the old default-branch short-circuit), the
  family convention (`traverseNestedPath` / `traverseFieldPath` / `traverseFieldPathGeneric`).
- One `Warn` per failed lookup (message and `field`/`full_path` keys preserved, `paths_tried`
  added); per-candidate attempts at Debug. The dead `ExtractStepData` branch not ported
  (only reachable with a nil argument, which `ExtractStepData` returns as-is — severable if a
  reviewer disagrees).
- 15 tests in `extract_field_value_termination_test.go`, run under `go test -timeout`;
  **mutation control done**: the exact crash input against the old function verbatim FAILS by
  stack overflow in 3.7s, so the guard is proven able to fail. `verify-head-builds.sh --with
  --test` green at `38db61b28`.

**Diagnosis-loop substitution statement (owner ruling 2026-07-31):** this file's root cause
was not routed through the 090 loop; the substitute is the first-hand evidence in §2 — the
mechanism is directly OBSERVED (the two inverse fallbacks read at source, the crash captured
at the pod with exit code and the 12,654-line log, three wedged orchestrations by correlation),
not hypothesised, so a 090 run would re-read the same fifteen lines. Stated here per the
ruling's named escape hatch rather than left implicit.

**Still owed before this moves to `bugs_closed/`** (fixed AND live bar): a chassis image at or
after `6e2d4a039` built (bumped `IMAGE_TAG`), rolled, verified at the pod
(`git merge-base --is-ancestor 6e2d4a039 <stamp>`); then the end-to-end check of §6 — rebuild
a writer-skipped page, pod restartCount unchanged, orchestration terminal with
`skip_reason: "no content found at …"`, `grep -c 'Field not found in path'` = 1 for that
lookup. **That green run is NOT precondition-4 evidence for 357/578** (§8, and the 357
handoff 2026-08-26b Finding 1 — the whole chain skips; the row survives because nothing
touched it).

**Separate track recommended, not done here:** the estate carries four `.response.`-fallback
resolvers (**4** as of 2026-09-02: multipage, hitl `getNestedFieldValue`,
`resolveFieldPathForSpawn`, `resolveFieldPathCallAgent`) with three DIFFERENT candidate
orders, atop the ~10–14 near-clone dot-path walkers in LANDMINES' dotted-path-resolver entry.
A concept-register census entry (each walker, its order, its failure signal, dated) is the
cheap discoverability win; consolidation would change resolution semantics for somebody and is
RFC-scope if ever actually needed. No lane owns this today.

**Roll-verification instrument (from the webdesign-tool-rebuilds lane, 2026-09-02, baseline
taken on the running binary):** the chassis rolled 2026-09-02 12:28Z (`v1.0.1352`) BEFORE this
fix's 13:53Z commit, so `6e2d4a039` is NOT aboard — verified current. Baseline: `paths_tried`
(a string unique to this fix — single source hit) is **ABSENT** from `/proc/1/exe` on the
running pods, with the shared `"Field not found in path"` string **PRESENT** as the positive
control — so the instrument discriminates. **The cheap check at the next roll: probe the
binary for `paths_tried`, expect PRESENT.** Prefer this capability probe over the provenance
stamp here: the startup line had rotated out of both pods' logs (~9.5h) and the tag-bump
commit is not the build stamp (`bugs_open/249`'s straddle).

**Second-order consequence for delivery lanes (from the webdesign-tool-rebuild lane,
2026-09-02), with a path caveat they could not have seen:** post-fix, a no-content assembly
completes CLEANLY — so a work item whose delivery depends on the rebuild *landing* reads
`complete` while the page was never reassembled and keeps serving what it served before,
every status green. **Check `collected_data->'assembled_page'->>'skip_reason'` (non-null ⇒
skipped; re-polling the artefact is futile) before concluding a rebuild ran.** This is the
`a-complete-work-item-is-not-a-repaired-artefact` shape, one level up. ⚠ **The caveat: that
key exists only on the `assemble_page` ACTION path** (`AssemblePageAction`,
`multipage_actions.go` — consumers `page-rebuild` / `pageflow-builder` /
`site-work-orchestrator`, **3** as of 2026-09-02). The `page-rerender` agent's path
(`assemblePage`, `rerender_single_page_action.go`) contains **zero** occurrences of
`skip_reason` / `extractFieldValue` / `assembled_page` [MEASURED 2026-09-02] — it neither
carries 408's defect nor writes the discriminator, so a gate keyed on that field is silent
(not false, VACUOUS) for rerender-path deliveries. Know which function your chain reaches
before trusting the key — the two functions eleven characters apart are an existing
LANDMINES entry.

## 10. COUNCIL APPROVED r1 (2026-09-02, corr `3918db52`) — three advisories, all ACTED ON in `b8bf40694`

9 seats reviewed, 7 abstained, none high-severity. Dispositions:

- **bug_historian (medium) — "the fix trades a loud failure for a quiet one":** ACCEPTED and
  closed. When `content_field` resolves empty and the upstream step did NOT declare
  `skipped: true`, the caller now writes `agent_error_log` code
  **`ASSEMBLE_CONTENT_FIELD_UNRESOLVED`** (high) — so a mis-set `content_field` is countable
  (`SELECT count(*) FROM agent_error_log WHERE error_code='ASSEMBLE_CONTENT_FIELD_UNRESOLVED'`),
  while a declared writer skip stays a routine skip. Behaviour unchanged either way.
- **bug_historian (low/missing) — "only 3 of ~14 walkers proven non-recursive":** DONE, with
  a control this time. **20** walker function bodies censused as of 2026-09-02 (comment-stripped
  body extraction; positive control = the old recursive function, correctly flagged). ONE
  self-recursion in the repo: `FindByPath` (`datahelpers/content_search.go`) — its two arms
  each strictly decrease a finite measure (arm 1 shortens the path; arm 2 descends the data
  with the first segment preserved, so arm 1 cannot re-fire) and cannot cycle on acyclic
  JSON-derived data. **Not the 408 shape; the crash class is closed across the family.**
- **guardian (low) — "the timeout lives in the invoker":** DONE — every table row now runs
  under a 30s in-test watchdog; a reintroduced hang fails the test itself, a stack-overflow
  crash fails the suite on its own.

**Fleet context for the quiet-skip consequence (webdesign-tool-rebuild lane, measured
2026-09-02):** the RERENDER path's own skip (`"page has no component rows"`, key
`rendered_page`) fired **36 times in 1,146 runs over 7 days** and predates 408 entirely. ⚠ On
that path the `skipped` key is **ABSENT on normal runs** (36 rows carry it; all 36 are skips)
— a check testing `skipped='false'` matches nothing and looks clean; **test for `'true'`**.
Their lane's WRONG_CALLS has the fuller account (a mutation test that proved a branch
EXECUTES without proving its query CAN MATCH).

## 11. CLOSED 2026-09-02 — fixed AND live AND the crash input exercised in production

- **Live at the binary:** `v1.0.1354`, `paths_tried` PRESENT at `/proc/1/exe` against the
  webdesign-tool-rebuilds lane's pre-proven discriminating baseline (ABSENT on `v1.0.1352`,
  positive control PRESENT both times).
- **§6 end-to-end RUN, 2026-09-02 ~18:1x-18:3xZ:** cv1.co.uk flagged and rebuilt through the
  real pipeline (canary corr `6e84a4e3-726d-4ddc-a497-717c0469e238`, receipt-asserted
  publish). Iteration 0 hit THE EXACT CRASH INPUT — an adopted page whose writer skipped, so
  `page_content_0.response.page_html` resolved on no candidate form — and produced:
  `assembled_page.skipped=true, skip_reason="no content found at
  page_content_0.response.page_html"` (the FIXED code's own message), save ran and exited
  clean, `update_page_status` reached, **the whole orchestration chain COMPLETED**
  (`0711ef20`, parent `557052ef`) in minutes, **both chassis pods restartCount 0**
  (baseline pinned pre-dispatch). Under the old code this input never returns — three
  orchestrations died on it 2026-08-26, each wedged 4h. Both adopted rows byte-identical
  after (md5s `26f484f2…`/`291b88d8…`, 1 row each).
- **Not captured, stated honestly:** the one-Warn-vs-12,654 log-line count — the spawned
  rebuilder pod's log stream returned empty at capture time (the ephemeral-pod trap this
  file's own relations note). The count claim rests on structure instead: a 12,654-line
  flood ends in a dead pod, and the pods lived.
- **The green canary is NOT precondition-4 evidence for 357/578/701** — §8/§10 stand.
- Verified how the closing bar asks: at the binary, at the orchestration record, and at the
  artefact rows — not at a status.
