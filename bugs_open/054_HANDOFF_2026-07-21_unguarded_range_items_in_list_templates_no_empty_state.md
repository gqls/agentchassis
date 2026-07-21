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
