# PLAN — close the fragment blind spot (bugs_open/071, § "Related defect, same blind spot")

**Date:** 2026-08-06. **Thread:** bugfix_071_fragment_blindspot.
**Bug:** `bugs_open/071` — the residue its own 2026-07-27 triage named "the fragment
blind spot — unfixed, and now acknowledged in the code itself". The headline
detect-then-discard mechanism is closed (079's repair, live since v1.0.1174); the
renderer-default residue is `bugs_open/203`'s active lane; **the fragment half is
owned by nobody and is untouched since filing.**

## Why this bug, and the ownership check

Picked via the reference-heat ranking over live transcripts (memory:
`who-owns-is-blind-to-uncommitted-sessions`): 071 sat cold; its fragment-fix
symbols (`validateInternalLinks`, `element_refs`, `check_phantom_internal_links`,
`SplitFragment`) drew only listing-noise counts (1–4) in every live transcript,
re-run immediately before first Write. The one 27-hit transcript (d361e826) is
building a *different* discovery check (page-pairs); overlap is limited to the
shared coverage-test files — edits there are additive, re-read before editing.

Adjacent active lanes deliberately NOT entered: `bugs_open/203` (CTA renderer
default — its fix 880a405a6 landed 08-05; the `primary_cta_url`/`secondary_cta_url`
defaults map at `component_library.go:1136-1147` is still live and is recorded
here as **their** residue, not taken); `bugs_open/084` (asset_reference_404 check,
committed e526a5196, awaiting roll); `bugs_open/149` (checker-layer queue).

## The defect, verified at HEAD 2026-08-06 (090 substitute, stated)

Per the 2026-07-31 owner ruling: this cross-cutting claim was not put through 090
because it is not a NEW root-cause assertion — 071 established and measured the
mechanism two-sidedly on 2026-07-25/27, 016b §9 carries it ("`#frag` classifies as
`LinkScopeAnchor` and is skipped by phantom, misdirected and validate alike, with
nothing anywhere resolving a fragment against the page's ids"), and the code
acknowledges it (`datahelpers/link_repair.go:145-147`). What this thread did
instead, first-hand at HEAD:

- `datahelpers/links.go:113-114` — `#…` → `LinkScopeAnchor`; `:199-201` —
  `NormalizePagePath` drops `#fragment` before any comparison.
- `validate_page_content.go:910` — "external / anchor / mailto / asset scopes are
  not page links — skip."
- `check_phantom_internal_links.go` `accumulateLinkIssues` — same two scopes
  handled (`empty`, `page`); anchor scope skipped; page-scope fragments dropped by
  normalisation.
- No caller anywhere resolves a fragment against document ids
  (`grep -rn "fragment" platform/` — the only handling is `hrefSuffix` in
  link_repair.go, which *preserves* the fragment on a rewrite, correctly, and
  names this gap in its own comment).

## Live exposure, re-measured 2026-08-06 (it MOVED since filing)

071's 07-25 figure — 24 of 25 anchored links fleet-wide dead — **no longer
holds**. Measured today on rendered components of active, shipped pages:

- `path#fragment` links: **5**, all idea.uk (`/tools.html#audience-check` ×3,
  `/report.html#request-a-report` ×2) — **all resolve** (served pages carry the
  ids; HTTP 200; probed).
- bare `#fragment` links: **61** across 5 domains, dominated by skip-links
  (`#content` ×57 loanandmortgagecalculator + loancash — id present in stored
  rows AND on served pages), `#input-tokens`, `#gi-rules`, `#request-a-report`,
  `#audience-check` — every probed one resolves.

So: **live damage ≈ 0 today; the class is unguarded.** The estate got clean
through per-page repairs and the 092 writer constraints, not through any check —
071's own history shows the writer re-authoring dead anchors on *three*
consecutive rebuilds of one site. The next unconstrained writer run, tool port,
or template change reintroduces the class invisibly. Same posture as
`bugs_open/093` when it was fixed: "live exposure nil, structural gap real".

## Fix — detection in the canonical seam, prevention at the writer

Preferring the framework over the instance (owner conventions), and reusing the
two purpose-built shared mechanisms rather than adding a new surface:

### D1. Shared primitive: `SplitFragment` (datahelpers/links.go)
`SplitFragment(href) (path, fragment string)` — the one definition of "the
fragment of an href" (split on first `#`, fragment excludes `#`, query stays with
the path half per URL syntax `path?q#frag`). Gate, audit and repair currently
each reason about fragments ad hoc (`hrefSuffix`, `NormalizePagePath`'s
`IndexAny("#?")`); new consumers get one spelling.

### D2. Shared primitive: `DocumentSatisfiesFragment` (datahelpers/element_refs.go)
Extract the *presence* half of `OrphanElementRefs` into an exported helper:
"does this document contain or create an element with this id" — static
`id="x"`, dynamic `.id =` / `setAttribute('id',…)`, and the interpolated-id
loosening (a page that computes its ids satisfies any fragment that appears as a
quoted bareword). `OrphanElementRefs` is rewritten on top of it — one
definition, so the fragment check inherits every false-positive conservatism
that check paid for (css-filter-playground, 2026-07-29).

### D3. Detection arm in `check_phantom_internal_links` (ENABLED already)
New issue type **`dead_fragment_link`** inside the existing check — NOT a new
check, so it needs **no config change** and cannot land in the "built but never
enabled" state (`bugs_open/093`'s shape). The check already runs on
`completeness-discovery-agent` (live config verified 2026-08-06; work items
complete as recently as 08-05).

- bare `#x` in a surface → resolve against the CONTAINING page's full document
  (all its page_components + all site_components chrome — matching
  `OrphanElementRefs`' whole-page rule; skip if inside a data-runtime-fill
  shell; skip `IsNoopHref` forms, which are `dead_controls`' remit).
- `/path#x` whose path resolves to a real page → resolve `x` against the TARGET
  page's document + chrome. A `pages.url` that itself carries a fragment
  (idea.uk `tool-audience-check` → `/tools.html#audience-check`) already
  normalises to its path form on both sides, so target lookup is unchanged.
- Fragment on an UNBUILT target → already filed as `unbuilt_internal_link`; not
  double-filed. Target html NULL/empty → no judgement (absence of evidence).
- Severity **low** (an inert control, not a 404), priority below phantom;
  routed like its siblings by surface (page_component → page-build-handler /
  content; site_component → nav-link-fixer / build) — a rebuild re-runs the
  writer under D4's constraint, so the remediation actually converges.
- Verifier registered (`RegisterVerifier("dead_fragment_link", …)`): re-run the
  resolution on current stored html — resolved iff the href is gone or the
  fragment now resolves. Mechanical, same shape as `orphan_element_refs`'s.
- Both coverage guards (`verifier_coverage_test.go` sensor,
  `handler_coverage_test.go`) pass by construction: verifier registered, handler
  agents already registered. **Additive edits only** in shared test files (a
  concurrent lane is adding its own check).

### D4. Prevention at the writer (`prepare_link_context_action.go`)
One constraint sentence added to `link_constraint_text`: do not author `#fragment`
hrefs — bare or suffixed — unless the anchor is explicitly listed in the page
data; none are supplied today. Correct-or-absent (LNK-005's pattern): the writer
loses nothing (fragments it authors today are unverifiable inventions) and the
D3 remediation loop stops re-authoring what it just flagged.

### Considered and deliberately NOT in this round
1. **Gate-side fragment validation** (`validate_page_content`): the gate sees the
   writer's page_html WITHOUT chrome, so a bare-# targeting a chrome id would
   false-positive. Needs a chrome-aware id load at the gate; deferred, recorded.
2. **Repair extension** (unlink dead fragments at save): unlinking label-bearing
   anchors is the LANDMINES:1736 shape (label survives as bare text). Detection
   first, measure volume, then decide.
3. **Section-id emission** (make sections addressable so fragments CAN resolve —
   the capability half): changes every page's rendered HTML fleet-wide, with
   rerender/stale-page implications; architecture-adjacent; its own round if
   wanted. The detection arm does not depend on it.
4. **Anchor lists supplied to the writer**: needs per-page id censuses at
   write time; only worth it after (3).

## Blast radius, measured before submission (not delegated to reviewers)

- The arm's would-be findings on TODAY's estate: measured by running the
  shipping predicate over every fragment-bearing href (66 total) — expected
  **≈0 findings** (all probed fragments resolve; the dry harness in RUNBOOK
  confirms pre-roll, plus one induced positive control).
- Assembler adds no ids: served-vs-stored id diff on a probe page = ∅, so
  resolving against stored rows cannot false-positive on assembler wrappers
  (single-page probe; harness re-checks across all 66).
- New item type consumers: `page-build-handler` / `nav-link-fixer` — both already
  registered, both already handle these surfaces for `phantom_internal_link`.
  Item volume ≈0 at intro; no queue flooding.
- `verifier_coverage_test` / `handler_coverage_test`: satisfied in the same
  commit (sensor reads source).

## Registration & docs (same commit as the code, per the 2026-07-29 ruling)

- Concept register: new LNK entry (fragment-resolution primitives + arm), status
  "committed, inert until roll"; **correct LNK-009's stale status line
  visibly** (it says "deliberately not yet enabled"; the live config has
  `phantom_internal_links` enabled on completeness-discovery-agent — verified
  2026-08-06). Update index count.
- `bugs_open/071`: dated contribution — re-measurement, this plan, what remains
  with other lanes (renderer defaults map → 203; repair/gate/id-emission
  deferred candidates recorded).
- 016b §9 line 1709 family: point the entry at the fix once live.
- WRONG_CALLS: the `phantom_internal_links`-vs-`phantom_internal_link` spelling
  near-miss (a "zero items ever" claim from querying the check NAME as the item
  type; caught by reading the ItemType literal in the check source).

## Verification (owed at the roll)

1. Pod-grep a string only this change creates (`dead_fragment_link`) with a
   positive control (`phantom_internal_link`, live today) and a negative
   (invented string), one exec, every replica.
2. Induce the failing branch: insert a `#no-such-id` href into a scratch
   page_component (or run the harness against a doctored document), run the
   check, assert exactly one `dead_fragment_link` item with the right page,
   href and routing; then repair and assert the verifier closes it.
3. Assert the no-op case too (memory: check-the-no-op-case): a resolving
   fragment (`loancash #content`) files nothing.
4. Council: submit before/alongside the commit; `Council-Submitted:` trailer if
   the verdict has not landed.
