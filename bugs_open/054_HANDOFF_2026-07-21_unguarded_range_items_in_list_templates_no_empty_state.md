# 054 — list components `{{range .items}}` with no `{{if}}` guard: no empty-state, fleet-wide

**Filed 2026-07-21** (relojistas thread, from a council objection on the bugs_open/027 rework).
**Status: OPEN.** Low severity — a completeness/UX gap, not a crash. Filed because it is the
generic form of the defect the 027 rework fixed in only two of seven places, and the council
(bug_historian seat) correctly flagged that fixing two call sites leaves the mechanism
generic and the other five exposed.

> **UPDATE 2026-07-21 (cta_link_integrity thread) — the empty-state SWEEP is FIXED & LIVE.
> Bug stays OPEN, narrowed to the resolver root-cause (fix-candidate 2).**
>
> - **Fix-candidate 1 (template guards): DONE & LIVE.** `migration 185`
>   (`docs/agent_docs/sql_for_agents/185_list_empty_state_guards.sql`, applied + ledgered
>   `2026-07-21 11:46Z`, config change so live immediately — no image roll). Wraps all five
>   unguarded templates (`archetype-grid`, `game-list`, `guide-list`, `tool-cta`, `tool-list`)
>   in `{{if .items}}…{{else}}<p class="…-empty">…</p>{{end}}`, matching the news pair. The
>   empty-state copy is a **new `source:llm` `empty_state_text` field** (translatable, per
>   bugs_open/026) with an English template fallback. Verified three ways: (a) the bug's own
>   audit query below now returns `has_if_guard=t` for **all 7** range components; (b) Go
>   `text/template` parse+render of every transformed template — populated → cards & no
>   empty-state, empty → empty-state & no cards, `empty_state_text` override honoured;
>   (c) dry-run with ROLLBACK before the real apply. Commit `f8ef83133`.
> - **Fix-candidate 3 (standing lint): DONE.** `scripts/check_list_empty_states.py` — advisory,
>   flags any active component that `{{range .items}}` without an items guard, exit 1 on a hit.
>   Currently clean (`OK: all 7 …`). DB-only content, so it is an operational check, not a
>   `pattern-check.py` (diff-linter) entry. Same commit.
> - **Fix-candidate 2 (resolver root-cause): still OPEN, now precisely located.** The empty
>   render is reachable because **`plan_sections_action.go:1288-1321` sets
>   `resolvedData[field]=value` and `continue`s for any `source:query.*` field before the
>   `required`/`on_missing`/`min_items` branch (`:1333-1432`) ever runs.** An empty slice is
>   not `nil` in Go, so the `items {required:true, min_items:1}` all five carry (and every
>   query-sourced array in the library) is **silently ignored** — the schema says "≥1 item
>   required" and nothing enforces it. The comment at `:1285-1287` even claims `on_missing`
>   applies on empty; the code does not implement it. `[UNVERIFIED]` whether any *downstream*
>   stage (a content-validation action) re-checks `min_items` after this resolver — I did not
>   trace past `plan_sections`. This is a data-integrity gap beyond empty-state UX and wants
>   its own diagnosis run before anyone changes the resolver (masking a genuine "resolver
>   errored" case as "empty" is the trap the original fix-candidate-2 flags).

> **UPDATE 2026-07-22 (085_debug_and_feature_loops thread) — fix-candidate 2 diagnosis
> COMPLETED; the `[UNVERIFIED]` is now resolved and the feared trap does NOT apply.
> Bug stays OPEN pending an owner direction call (below) + the Go change + an image roll.**
>
> **Root cause, fully traced and cited (working tree, 2026-07-22):**
> 1. `queryresolve.go:359` — `resolvePagesWhereType` (and the sibling array resolvers)
>    build `items := make([]map[string]interface{}, 0)` and `return items, nil` on zero
>    rows. That is a **non-nil empty slice**, so at `plan_sections_action.go:1313` the
>    `else if value != nil` branch is taken → `resolvedData[field]=value; continue`
>    (`:1314-1315`). The `required`/`on_missing`/`min_items` logic below is never reached
>    for a query field. Confirmed.
> 2. `min_items` is **read** into `fieldMinItems` (`:1260-1262`) but a fleet-wide grep
>    (`min_items|MinItems|minItems` over `platform/ pkg/ internal/`, non-test) shows it is
>    only ever copied into the `missingField` **metadata** struct (`:1367,1395,1409,1421,
>    1433`) — it is **never length-compared against a resolved slice anywhere**. So the
>    contract `{required:true, min_items:1}` is dead for query arrays.
> 3. **`[UNVERIFIED]` resolved:** the same grep proves **no downstream Go stage** re-checks
>    `min_items` after this resolver (only `plan_sections_action.go` and
>    `component_schema_fields.go` mention it; the latter just normalises `minItems`→
>    `min_items`). Nothing enforces it, full stop.
>
> **Blast radius (live DB, 2026-07-22) — every `source:query.*` field across active
> components, by declared `on_missing`:**
>
> | on_missing (declared) | required | fields | components | effect of honouring it on an EMPTY array |
> |---|---|---|---|---|
> | `skip_field` | false | 10 | 5 | field omitted → `{{if .items}}` false → **empty-state** (already the outcome; candidate 1) |
> | *blank* → defaults to `skip_field` (`:1251-1253`) | true | 7 | 6 | same — omit → empty-state. **Includes all 5 of this bug's list components.** |
> | *blank* → `skip_field` | false | 4 | 4 | omit → empty-state |
> | `skip_section` | true | 8 | 8 | **section dropped from the build** — the only behaviour-changing bucket |
>
> The 8 `skip_section` set (all `component_level='section'`): `category-listing`,
> `content-listing`, `directory-listing`, `filtered-result-grid`, `gripper-spec-sheet`,
> `product-card-with-cta`, `product-grid` (all array fields), + `featured-content`
> (a `text` scalar — unaffected by an array/min_items fix). These 7 array listings are a
> **separate exposure** from the 7 `{{range .items}}` components candidate 1 guarded: they
> range over `.products`/`.entries`/`.results`/`.articles`, carry **no** `{{if}}` guard,
> and today render a genuinely blank section when empty. Their authors declared
> `on_missing=skip_section`, which the short-circuit silently ignores.
>
> **Why honouring `on_missing` is SAFE (the handoff's traps do not fire):**
> - The `skipped` status is aggregated at `:733-735` into neither `ready` (not built) nor
>   `deferred` (no HITL/data-request item) — a skipped section is **simply omitted** and
>   re-evaluated on the next re-plan/render, so it **self-heals** when data arrives.
> - A skipped section never reaches `check_empty_sections`, so honouring `skip_section`
>   **shrinks** the empty-section false-positive family (016b §9), not grows it.
> - "Mask resolver-errored as empty" cannot happen: the code already routes `qerr != nil`
>   to fallback (`:1307-1312`) separately from the empty-success path; a fix touches only
>   the latter.
> - No query field declares `block` or `needs_human_review`, so no fix can start blocking
>   or HITL-flooding a page.
>
> **The one genuine decision (fleet-facing, so it's the owner's):** for an empty
> `skip_section` array listing (e.g. a `product-grid` with zero products), should the
> section be **dropped** (honour the declared intent — recommended) or instead get an
> **empty-state** like the 5 list components? The 5 list components are unaffected either
> way (they already render an empty-state).
>
> **Fix shape (Go, inert until image roll):** in the `query.*` branch, when the resolved
> value is an empty array that fails its declared contract (`required` or `min_items≥1`),
> do NOT short-circuit — route it through the SAME `on_missing` handling the generic path
> uses (factored into one shared helper so the two paths cannot drift — the drift class
> `bugs_closed/044` was closed for). Council-gate before commit.
>
> **STATUS 2026-07-22 — fix-candidate 2 BUILT, TESTED & COMMITTED (`7e60627ef`); INERT
> until the next chassis image roll → bug stays OPEN.** The owner chose **honour
> `skip_section` (drop)** for empty listings. The Go change is exactly the fix shape above:
> a `handleMissingField` closure shared by the generic + `query.*` paths, plus pure DB-free
> helpers `queryListBelowContract`/`queryResultLen`, with unit tests
> (`plan_sections_contract_test.go`). `go build` + `go test ./…/actions` pass, incl. the
> pre-existing `planSection` defer tests (refactor preserved). Council-gate submitted:
> `SUBMISSION_CORR=958d52f4-af5e-4414-9caa-82655577bda6` (verdict pending; trailer added
> only on APPROVED). **To close:** verify live after roll — an empty `product-grid`/
> `directory-listing` etc. is dropped (not blank) and re-appears on the next render once
> data exists; the 5 list components still render the empty-state.
>
> **COUNCIL ROUND 1 = REVISE** (bug_historian drove it; 4 approve / 4 object). Objections
> and how R2 addressed each — all verified against the live code/DB, not re-asserted:
> - **bug_historian #1 (the empty-slice convention is generic across callers):** code-check
>   `grep queryresolve.Resolve` → exactly **2 consumers**; the other
>   (`reconcile_section_data_action.go:171`) already guards with `!hasItems(val)` (a length
>   check), so **no other `value!=nil` trap site exists**. Not left generically exploitable.
> - **bug_historian #2 (silent permanent disappearance of a never-populated listing):** added
>   a **Warn log** at the below-contract branch so a dropped required listing is
>   operator-visible. A fuller staleness/HITL tripwire is flagged as an **owner fast-follow**.
> - **reuse / prior_art:** consolidated the duplicate `hasItems` type-switch onto the shared
>   `queryResultLen` primitive; code-checks confirm **no** pre-existing on_missing handler or
>   min_items/length comparator existed to reuse.
> - **guardian (layer choice):** stated explicitly — the fix must live in `planSection` because
>   the field's `required`/`min_items`/`on_missing` metadata is only in scope there;
>   `queryresolve.Resolve` gets only `(name, siteID, limit)` and there is no post-resolve
>   validation stage. Caller-safety: `shouldSkip`/`shouldDefer`/`missingFields` are locals.
> - **debug_historian:** added a **pod-grep verification step** (post-roll,
>   `strings /app/agent-chassis | grep -c queryListBelowContract`) to the plan.
> - JSON-type check: `required` is a JSON boolean and `min_items` a JSON number, so the Go
>   `.(bool)`/`.(float64)` reads genuinely fire (a top-level `input_schema->>'required'` is
>   NULL by design — the contract is per-field; that is what the R1 reviewer's query hit).
>
> R2 revisions committed `6e9f06ecf`; resubmitted on the same corr (R2 run `2904a344`).
>
> **COUNCIL ROUND 2 = REVISE, but 7 approve / 3 object** (up from 4/4; `debug_historian` and
> `prior_art_librarian` flipped to approve — the pod-grep step and the quoted code-checks
> landed). **Loop STOPPED here** per the runbook ("one council run per coherent task, not
> per iteration"; "seats contradict — pick one, record why, move on"). The 3 remaining
> objections, and the disposition of each:
> - **editquality (low) — the `hasItems` reuse is scope-creep against minimality.** This
>   directly contradicts R1's `reuse_agent` ask (which wanted exactly this consolidation, and
>   approved it in R2). Per the runbook's "seats contradict across rounds" rule: **kept**, because
>   it removes a genuine duplicate type-switch and the reuse seat approved it twice. Recorded here
>   as the deliberate call.
> - **editquality (low) — call out the `continue`→`return`.** It only affects the optional
>   path, which was already the last statement before the loop's end, so `return` from the
>   closure and `continue` are behaviourally identical; the pre-existing `planSection` defer
>   tests pass unchanged. Documented, no further change.
> - **guardian (medium) — name the pipelines that invoke `plan_sections` (runtime blast
>   radius).** Answered from the live DB: the `plan_sections` action is invoked within the
>   **build pipeline** — `page-build-handler` and `build-site-planner` (also referenced by
>   `page-content-writer`/`image-build-handler`/`diagnose-agent`). Bounded set, all the
>   page-build path; not an unbounded fan-out.
> - **bug_historian (medium) — a required listing that NEVER recovers data should fail-loud
>   into a tracked `site_work_items` row, not just a Warn log.** Principled, but it
>   **re-litigates the owner's own decision**: the 7 fields declare `on_missing=skip_section`
>   ("drop if empty"), and the owner chose (2026-07-22) to honour that. Creating a tracked item
>   instead would OVERRIDE the declared `skip_section` intent — a different decision from the one
>   made. Left as an **explicit owner call** (below), because the honest schema-faithful behaviour
>   and the fail-loud behaviour genuinely diverge here.
>
> **Bug remains OPEN** (fix inert until roll). **Open owner decisions:**
> 1. **Accept the fix as-is** (honours `skip_section` literally: empty listing dropped + Warn
>    logged) **OR** build `bug_historian`'s structural fail-loud guard — an empty *required*
>    `skip_section` list emits a tracked `site_work_items` (needs_human_review-style) row so a
>    never-recovering listing is surfaced, not just logged. The latter overrides the schema's
>    declared `skip_section` for the required+empty case.
> 2. The two fast-follows: a standing lint (future `Resolve` consumers length-check, not
>    `value!=nil`); the staleness/HITL tripwire (subsumed by decision 1 if the guard is built).

## What

Seven active `content_components` render a query-sourced list with `{{range .items}}`. After
the 027 rework, **two** (`latest-news`, `news-listing`) wrap it in
`{{if .items}}…{{else}}<placeholder>{{end}}`. The other **five do not**:

```sql
SELECT function,
       html_template ~ '\{\{if [^}]*items\}\}' AS has_if_guard
  FROM content_components
 WHERE html_template LIKE '%{{range .items}}%' AND is_active;
-- archetype-grid | f
-- game-list      | f
-- guide-list     | f
-- tool-cta       | f
-- tool-list      | f
-- latest-news    | t   (fixed by the 027 rework)
-- news-listing   | t   (fixed by the 027 rework)
```

## Why it is LOW, not the silent crash the framing suggests

`{{range}}` over a nil or empty slice in Go `text/template` is a **no-op**, not an error — it
iterates zero times and produces empty output. So an unguarded `{{range .items}}` on an empty
list does not fail; it renders an **empty container with no "nothing here yet" message**.
(001 §7's silent-failure lesson is specifically about an *LLM prompt* rendering empty and the
model getting nothing — a different consequence of the same nil; for an HTML listing the
consequence is a blank section, not a crash.)

So the real defect is a missing graceful empty-state, and an inconsistency: the two news
components now degrade gracefully, the other five render blank. On a freshly-built site whose
guides/games/tools/entities have not populated yet, those five sections are silently empty
with no explanatory copy — which also feeds the `check_empty_sections` false-positive family
(a runtime/query-fill section reads as empty to a server-side check).

## Fix candidates

1. **Add the `{{else}}` empty-state to the five templates**, matching the news pair
   (`{{if .items}}{{range}}…{{end}}{{else}}<p class="…-empty">…</p>{{end}}`). Mechanical,
   per-component, low-risk. Each needs a component-appropriate message (and ideally a
   translatable llm-sourced field like news-listing's `loading_text`, so it is not
   hardcoded English — see bugs_open/026 for the English-on-a-Spanish-site version of this).
2. **Root-cause option the council raised: the generic `text/template` `missingkey=zero`
   posture.** There is no central guard that a query-sourced field always resolves to a
   non-nil `[]`; each template carries the risk. A shared render-time default (empty slice
   for any declared `type: array` schema field whose source did not resolve) would make the
   `{{if}}` guard unnecessary everywhere. Larger change; verify it does not mask a genuine
   "resolver errored" case as "empty".
3. **A `content_components` audit check** (discovery): flag any active section component whose
   `html_template` contains `{{range .items}}` without a matching `{{if .items}}`. Turns this
   from a one-off sweep into a standing lint.

Recommend (1) as the immediate sweep and (3) so it does not recur; (2) only if the
empty-state message is genuinely wanted centrally rather than per-component.

## How to verify a fix

The audit query above returns `has_if_guard = t` for every row that ranges over items. Then,
per component, render it against an empty list and confirm a visible empty-state message
(from `curl`, not a JS-filled browser).

## Related

- `bugs_open/027` — the 027 rework added the guard to the two news components; this is the
  same guard owed to the other five. Do not re-open 027 for these — they are separate
  components.
- `bugs_open/026` — the English-hardcoded placeholder is the same "empty-state copy is not
  translatable" problem on the news component specifically.
- `016b §9` "a false positive is a LOCATION" — the empty sections these render feed the
  `check_empty_sections` false-positive family.
