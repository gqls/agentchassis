# 054 — list components `{{range .items}}` with no `{{if}}` guard: no empty-state, fleet-wide

**Filed 2026-07-21** (relojistas thread, from a council objection on the bugs_open/027 rework).
**Status: OPEN.** Low severity — a completeness/UX gap, not a crash. Filed because it is the
generic form of the defect the 027 rework fixed in only two of seven places, and the council
(bug_historian seat) correctly flagged that fixing two call sites leaves the mechanism
generic and the other five exposed.

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
