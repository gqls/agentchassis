# Handoff — tool-suggester makes pages link to tools that do not exist

**Filed 2026-07-19.** Owner-visible: a visitor clicked a link on leopardessconsulting.co.uk and
got a blank 404. Found while fixing `/bugs_open/001`, which had this damage attributed to it
in error — see the correction in that file. **This is the bug that actually caused it.**

## Symptom

leopardessconsulting.co.uk `/services.html` carried a link to
`/tools/tool-monitoring-coverage-gap-finder.html`. That page has never existed:

```sql
SELECT name, url FROM pages
WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND (name ILIKE '%monitoring%' OR name ILIKE '%coverage%');
-- 0 rows
```

## Mechanism

`tool-suggester` proposes a tool and, **in the same batch**, writes `content_rewrite` work
items telling the content writer to reference that tool on existing pages. It writes them at
**suggestion** time — not after the tool is built — so the instruction names a tool that has no
page, and the writer invents a plausible URL for it.

On leopardess, 2026-07-18 02:32:56, ten items in one batch:

```sql
SELECT status, summary, spec->>'page_name' AS page, left(spec->>'suggestion',80)
FROM site_work_items
WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND source='tool-suggester'
ORDER BY created_at;
```

- one `add_tool` — *"Process Automation Suitability Scorer"* (status `complete`)
- nine `content_rewrite` — *"Add \<tool\> tool reference to \<page\> page"*, spec `suggestion` =
  *"Weave a natural reference to '\<tool\>' (…)"*, across three different tools.

The `add_tool` item and the nine reference-weaving items are **independent and unordered**.
Nothing sequences a rewrite behind the build of the tool it references, and nothing checks at
rewrite time that the tool has a page. Of the three tools referenced, only *Process Automation
Suitability Scorer* was ever built (07-18 07:53); *Monitoring Coverage Gap Finder* and *Data
Quality Risk Estimator* were referenced on five pages between them and never built at all.

The item that produced the 404 is `complete`:

> `Add Monitoring Coverage Gap Finder tool reference to services page` — status `complete`

Confirmed at the orchestration: run `615bee1d-e487-44d6-9692-82e9074a529f`, 2026-07-18
07:44:54, `input_data.spec` = `{"page":"services","source":"tool-suggester","page_name":
"services","suggestion":"Weave a natural reference to 'Monitoring Coverage Gap Finder' (Aimed
at visitors who need to track what is publishe…"}`. The `services` `page_components` rows all
update at 07:50:41.

## Why it matters beyond one link

1. **It is autonomous.** Unlike `/bugs_open/001`, which needs someone to deliberately fire a
   re-plan, this path runs on its own schedule. That is the property that made the leopardess
   damage look fast enough to "outrun a fix".
2. **It rewrites human-reviewed copy.** The rewrite regenerates the page's content wholesale.
   On leopardess that reintroduced material a person had audited out. `page_components` has
   `locked_at` / `lock_type` / `lock_expires_at`, and the discovery checks **do** honour them
   (`check_empty_sections.go`, `check_unverified_claims.go`, `check_placeholder_contact.go` all
   filter `locked_at IS NULL`) — but every `page_components` row on the rewritten leopardess
   pages had `locked_at` NULL, so nothing was engaged. Whether the rewrite path honours the
   lock at all is **unverified** — I did not read that path. Check before assuming either way.
3. **It manufactures a claim.** On a site whose governing rule is "no claim ships without an
   evidence row", "we have a Monitoring Coverage Gap Finder" is a fabrication with a clickable
   URL attached.

## Fix candidates (not implemented — none of this is written)

- **Sequence it.** Emit the `content_rewrite` items only once the tool's page exists — e.g. let
  the tool-build path emit them on success, the way `deploy_tool_action.go` /
  `create_tool_component_action.go` already emit follow-on items.
- **Gate at consumption.** Before a `content_rewrite` whose `spec.suggestion` names a tool runs,
  resolve the tool to a `pages` row; if it is missing, park the item rather than write the copy.
  Cheaper and catches the general case (any rewrite referencing a not-yet-real target), but
  leaves the bad item in the queue.
- **Gate at the link layer.** `resolve_internal_links_action.go` already exists to resolve
  internal destinations; a writer-emitted href to a non-existent page is exactly what it is for.
  Worth reading first — this may be a case it should already have caught, which would make the
  real defect "the writer bypasses the resolver" rather than anything in the suggester.

Read all three before choosing: the third would make this a symptom of a wider gap, and
`/bugs_open/023` (CTA/link integrity — "button label and URL are unrelated schema fields,
nothing pairs them") looks adjacent. **Grep 023 before starting; this may be the same family.**

## How to verify a fix

1. Find a site with an open `add_tool` item and its sibling `content_rewrite` items.
2. Assert no `content_rewrite` naming a tool runs while that tool has no `pages` row.
3. After the tool builds, assert the reference items then run and the emitted href resolves to
   a real page (200, not a soft 404).
4. Fleet sweep for existing damage — every page linking to a `/tools/…` URL with no `pages` row:

```sql
SELECT s.domain, p.name, pc.rendered_html
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ILIKE '%/tools/%'
-- then resolve each href against pages.url for that site
```

This existing damage is **not** cleaned up by fixing the emitter. Leopardess `/services.html`
is still in its rewritten state.

## Related

- `/bugs_open/001` — the re-plan clobber, which this was wrongly filed under. Its "FRESH
  EVIDENCE" section describes this bug's damage; read the correction at the end of that file.
- `/bugs_open/023` — CTA/link integrity; likely the same family, check before starting.
- A `needs_diagnosis` item was filed (`needs_diagnosis:tool-suggester-writes-content-rewrite-wo`,
  correlation `a8b483ff-55af-463d-9622-837c73780e48`) but **never dispatched** — no
  orchestration row exists for it. Everything above is primary DB evidence, not a loop verdict.
  Re-firing the 090 trigger on this symptom would still be worth it.
