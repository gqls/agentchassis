# 092 — the page writer is never told which pages exist, on every run

**Filed:** 2026-07-26, while fixing `bugs_open/079` (the deploy gate detected dead in-body
links and published them anyway).
**Severity:** high. This is the *upstream cause* of the invented links 079 has to clean up.
**Status:** OPEN — diagnosed with evidence, not fixed. Deliberately not folded into 079's
commit: different mechanism, different file, different agent's path, and it changes
content-generation behaviour fleet-wide, which nobody has measured.

## Symptom

`page-content-writer` invents links to pages that do not exist, on every site. Of the 15
unique phantom targets the deploy gate caught in its 13-day retained window, **15 were pure
inventions** — not one resolved to a real page in any form, with or without `.html`.

## Mechanism: the constraint step runs, finds nothing, and says nothing

The platform already has the machinery to prevent this, and it is already wired in.
`page-content-writer`'s workflow runs `prepare_link_context` before the writer step
(`k8s/bk_agent_definitions_backup.sql`, page-content-writer row):

```json
"prepare_link_context": {"action": "prepare_link_context",
  "config": {"enabled": true, "pages_field": "db_sync.pages", "max_links_per_section": 3},
  "next_step": "load_page_components", "output_field": "link_context"}
```

and the prompt template consumes it, guarded:

```
{{if .link_context.link_constraint_text}}\n## Internal Linking\n{{.link_context.link_constraint_text}}\n\n{{end}}
```

`PrepareLinkContextAction` reads its page list out of `collected_data` — configured field
`db_sync.pages`, with three fallbacks (`prepare_link_context_action.go:75-105`):

```go
	alternates := []string{
		"site_plan.pages",
		"pages_to_build.pages",
		"render_context.available_pages",
	}
```

**None of the four exists in that workflow's `collected_data`.** The writer orchestration
carries `render_context`, but its keys are brand/theme fields — `nav_items`, `services`,
`primary_color`, … — and there is no `available_pages`. There is no `db_sync` key at all.

So `extractPagesForLinking` returns an empty slice with only a `logger.Warn`
(`:102-105`), `buildLinkConstraintText` returns `""` on an empty list (`:151-153`), the
`{{if}}` guard elides the entire "## Internal Linking" block from the prompt, and the model
is left to guess. **The failure is silent at every layer**: no error, no work item, and the
prompt simply comes out shorter.

## Evidence

Live, 2026-07-26 — every writer run that recorded a link context:

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE (collected_data->'link_context'->>'page_count')::int = 0) AS zero_pages
FROM orchestration_states WHERE collected_data ? 'link_context';
```

```
 runs | zero_pages
------+------------
   20 |         20
```

**20 of 20. A 100% failure rate**, and `length(link_constraint_text) = 0` on all of them.
Sampled `collected_data` keys from one such run (`bbcc1186-381c-42ea-8a09-0a35da69bac6`):

```
action, agent_config, build_render_context, input_data, link_context,
link_resolver_info, load_site_specs, prepare_link_context, render_context,
researcher_info, resolved_links, resolve_links, sections_for_render,
select_sections, site_specs, spawn_link_resolver, spawn_research_agent
```

— no `db_sync`, no `site_plan`, no `pages_to_build`; `render_context` present but without
`available_pages`.

**Gotcha for whoever verifies this:** `prepare_link_context` runs inside the
page-content-writer's OWN orchestration, not the page-build-handler's. Query the child rows
(`collected_data ? 'link_context'`), or you will look at the parent, not find the step, and
conclude it never ran.

## Two traps for the fixing thread

1. **`InjectLinkConstraints` is NOT the missing piece.** It is defined in
   `platform/orchestration/actions/link_constraints.go:37` and has **zero call sites**, and
   `page-content-writer`'s `default_config` even carries a dead `"link_constraints":
   {"enabled": true, "max_internal_links_per_section": 3}` block that no Go code reads. It
   is a near-duplicate of `prepare_link_context`, which already runs. Wiring it would give
   the platform two competing implementations of the same prompt block. Delete it or make
   it the single implementation — do not run both.
2. **`prepare_link_context` synthesises URLs** when a page has a name but no url
   (`prepare_link_context_action.go:128-134`):
   ```go
   		if page.URL == "" && page.Name != "" {
   			if page.Name == "index" || page.Name == "home" {
   				page.URL = "/index.html"
   			} else {
   				page.URL = "/" + page.Name + ".html"
   			}
   ```
   A hardcoded `.html`, not `NormalizePagePath`, and not the stored `pages.url`. Fixing the
   plumbing without fixing this would hand the writer plausible-but-wrong addresses — the
   `bugs_closed/029` failure mode (an emitter that assembles URLs instead of citing real
   ones) reintroduced one layer upstream. Whatever fix lands should read `pages.url`
   directly.

## Fix candidates (none implemented)

1. **Have the action query the database.** It has `params.DB` and a `site_id`; digging four
   speculative paths out of `collected_data` is why it silently finds nothing. One query —
   the same one `loadValidPagePaths` and `loadResolverPageSet` already use — gives it the
   real `pages.url` values and removes the synthesis in trap 2. Largest blast radius is that
   the writer's prompt grows a section it has not had in living memory.
2. **Populate `db_sync.pages` on this path** so the configured field resolves. Smaller, but
   it leaves a silent-empty failure mode in place for the next workflow that forgets.
3. **Fail loudly.** Whatever else lands, an empty page list on a site that demonstrably has
   pages should not be a `logger.Warn` and an elided prompt section. It is the reason this
   went unnoticed long enough to be measured at 100%.

## How to verify a fix

`page_count > 0` and a non-empty `link_constraint_text` on a fresh writer run (query above),
then a build on a site with a known page set: the emitted hrefs must all resolve. Do **not**
verify by reading the prompt template — it is correct and always has been; the data it
interpolates is what is missing.

## Relates to

- `bugs_open/079` — the deploy-gate backstop, FIXED 2026-07-26. It repairs or removes the
  links this defect causes. Prevention here would make that repair mostly a no-op.
- `bugs_open/071` candidate 4 — "stop the writer inventing link targets. The prompt should be
  given the site's real page list and told to link only within it." That candidate assumed
  the machinery needed building. It does not: it exists, it runs, and it is fed nothing.
- `bugs_closed/029` — cite the real `pages.url`, never a constructed one. See trap 2.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — still 100%, and candidate 2 is now RULED OUT

Verification sweep, not a fix. The diagnosis above holds unchanged.

**Re-measured live**, same query as § Evidence:

```
 runs | zero_pages |            latest
------+------------+-------------------------------
   16 |         16 | 2026-07-27 14:27:31+00
```

Still **100%**, and the latest failing run is from today. The denominator fell 20 → 16 only
because `orchestration_states` is on a retention clock — that is a shrinking window, not
improvement. No code has moved: `prepare_link_context_action.go` and `link_constraints.go`
have no commits since 2026-03-28, and `InjectLinkConstraints` still has **zero call sites**
(trap 1 intact — do not wire it).

### The new finding: fix candidate 2 cannot work, so this needs a chassis roll

Candidate 2 is "populate `db_sync.pages` on this path so the configured field resolves" —
attractive because config is live immediately and needs no image. **There is nothing on that
orchestration to point the field at.** Every top-level key of the latest writer run was
inspected for a page list:

```sql
SELECT k, jsonb_typeof(v), left(v::text,120) FROM orchestration_states o,
LATERAL jsonb_each(o.collected_data) e(k,v)
WHERE o.collected_data ? 'link_context' AND (k ILIKE '%page%' OR v::text ILIKE '%"url"%')
  AND o.created_at = (SELECT max(created_at) FROM orchestration_states WHERE collected_data ? 'link_context');
```

The hits are page **HTML** (`compile_page`, `page_content`, `complete`), render context, and
section plans. `input_data.site_plan` is present and is literally `{}`. There is no array of
pages anywhere on the run, so repointing `pages_field` has no valid target and would fail the
same silent way.

**Therefore candidate 1 (have the action query the DB) is the only real option**, and this bug
is a Go change → council gate → image roll, not a config tweak. Size accordingly: it is one
query in `PrepareLinkContextAction` (it already holds `params.DB` and a `site_id`, and
`loadValidPagePaths` is the query to copy), plus deleting the trap-2 URL synthesis so it emits
stored `pages.url`. Small diff, real blast radius — the writer's prompt grows a section it has
not carried in living memory, on every site.

**Worth knowing before sizing the value:** `bugs_open/079`'s deploy-gate repair is now live
and exercised in production (`agent_error_log` error_code `CONTENT_LINK_REPAIR_DETAIL`,
dartsonline.com 2026-07-27, 1 rewrite + 1 unlink in one build). So invented links are being
*removed* before they ship. That lowers the urgency and does not remove the case: an unlinked
phantom is a paragraph that lost its link, which is still a worse page than one whose writer
was told what exists.
