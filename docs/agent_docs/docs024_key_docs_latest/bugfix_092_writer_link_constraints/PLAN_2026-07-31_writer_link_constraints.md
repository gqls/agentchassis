# PLAN — `bugs_open/092`: the page writer is never told which pages exist

**Opened:** 2026-07-31. **Bug:** `bugs_open/092_HANDOFF_2026-07-26_writer_never_receives_its_link_constraints.md`.
**Prior art:** concept register `LNK-017` (filed 2026-06-12, status `partial`) states this
same gap and names the same two candidates — so this is a **seven-week-old known defect**,
not a new discovery.

## The defect, restated

`page-content-writer`'s workflow runs `prepare_link_context` before the writer step, and the
prompt template interpolates its `link_constraint_text` under a `{{if}}` guard. The action
looks for its page list in `collected_data` at four speculative paths. **None of the four is
ever present on that orchestration**, so it returns an empty list, `buildLinkConstraintText`
returns `""`, the `{{if}}` elides the whole "## Internal Linking" block, and the model writes
links with no idea what exists.

## Re-verified live, 2026-07-31 (this is not a stale bug)

```
runs | zero_pages |            latest
  26 |         26 | 2026-07-31 15:36:26+00
```

`SELECT count(*), count(*) FILTER (WHERE (collected_data->'link_context'->>'page_count')::int = 0)
 FROM orchestration_states WHERE collected_data ? 'link_context';`

Still **100%**, and the latest failing run is from today. Neither
`prepare_link_context_action.go` nor `link_constraints.go` has moved since 2026-03-28.

## Design decisions, and their reasons

### D1 — the action queries the database itself (092 candidate 1)

Candidate 2 ("populate `db_sync.pages` on this path") was **ruled out on 2026-07-27** by the
092 filer: there is no array of pages anywhere on that orchestration to point `pages_field`
at. Re-confirmed today from the other side — of 26 `page-content-writer` orchestrations,
`db_sync` is present on **0**, `site_record` on **0**, top-level `site_id` on **0**, and
`input_data.site_id` on **26**.

### D2 — the DATABASE is authoritative; `collected_data` is the fallback, not the reverse

The invariant that matters is: **the writer's allow-list must equal the gate's accept-set.**
`validate_page_content`'s `loadValidPagePaths` decides what is a `phantom_link` by querying
`pages WHERE site_id = $1 AND status NOT IN ('deleted','archived')`. If the writer is
constrained from any *other* source, the two can disagree, and a writer obeying its
instructions gets its links flagged (or worse, un-flagged and shipped). Only reading the same
table under the same predicate makes that disagreement unrepresentable.

Nothing live is lost by inverting the precedence: the only producer of the configured field
(`db_sync.pages`, built by `LoadSiteForRebuildAction`) builds it **from this same table**, so
where both exist they agree. The configured field is still honoured whenever the DB path
cannot run, so no workflow loses its declared source.

### D3 — an empty page list must produce an EXPLICIT instruction, never silence

This is the actual fail-open. Today `buildLinkConstraintText` returns `""` for an empty list —
so the one input where guidance matters most produces **no guidance at all**, and the guard
then removes even the heading. The fix: an empty list emits "do not create internal links",
which is the safest available instruction and is correct in both of the states that produce it
(a brand-new site with no pages; a site whose page list could not be established).

### D4 — the two causes of "zero pages" must not look alike

*The site has no pages* and *I could not find out* have opposite remedies, so they are
reported differently: the first is a normal outcome with a stated `reason`; the second sets
`degraded: true` and writes a durable `agent_error_log` row (`LINK_CONTEXT_UNAVAILABLE`).
A `logger.Warn` is what let this run at 100% for seven weeks without anyone noticing.

### D5 — no URL is ever synthesised (092 trap 2)

`page.URL = "/" + page.Name + ".html"` hands the writer a plausible-but-wrong address — the
`bugs_closed/029` failure mode (an emitter that assembles URLs instead of citing real ones)
reintroduced one layer upstream. A page with no stored `url` is not a linkable target and is
dropped, counted, and logged. Dropping it is also what *causes* the DB fallback to fire, which
replaces the guess with the truth.

### D6 — `link_constraints.go` is deleted, not wired (092 trap 1)

`InjectLinkConstraints` is a near-duplicate of this action with **zero call sites** anywhere in
the tree, and it carries its own copy of the same URL synthesis (plus two extra guesses,
`/blog/` and `/tools/` prefixes). It is the standing landmine in `MEMORY.md` ("do NOT wire
`InjectLinkConstraints`"). Deleting it retires the landmine instead of restating it.

## Blast radius, measured before submitting (not left for a reviewer)

- Only consumer of the action: **`page-content-writer`** (1 row in `agent_definitions`).
- Every recorded writer run resolves a site id and would gain a real list:
  8 distinct runs in the retained window, all `robot-hands.com`, `site_id` resolvable on all,
  **31 linkable pages** available. So `0 → 31` on every run we can see, and the
  "no internal links" branch fires on **none** of them.
- `pages.url` is `NOT NULL` and **0 of 472 rows fleet-wide** are empty, so D5's drop path is
  currently unreachable in production — it is a guard, not a behaviour change.
- Only two `status` values exist fleet-wide (`active` 449, `archived` 23), so the gate's
  predicate and `loadActivePagesForLinkContext`'s `status='active'` are extensionally
  identical today. I use the **gate's**, for the D2 invariant.
- The visible change: the writer's prompt grows an "## Internal Linking" block it has not
  carried in living memory, on every site.

## Out of scope, deliberately

- The three hardcoded `/contact.html` defaults in `component_library.go` (`071`'s
  renderer-default class) — a different producer, downstream of the writer entirely.
- The fragment/anchor blind spot (`071`) — nothing emits section `id`s; unrelated mechanism.
- `resolve_internal_links`' `ctaFieldNames` coverage gap (`071`).
