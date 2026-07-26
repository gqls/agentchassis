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

---

# VERIFIED + SHARPENED 2026-07-21 (bugfix-029 tool-suggester session) — the code is read, the URL is fabricated (not "invented by the writer"), and it 404s even for tools that WERE built

Picked this up to fix it. The mechanism holds and is now read at the code, but the
framing above needs two corrections and the blast radius is measurably worse than
"tools that were never built".

## The diagnosis item DID run — it is now `status='complete'`

The 2026-07-19 note that "no orchestration row exists" was **queue latency**, not a drop
(the ~30-min council/diagnosis dispatch lag — same trap as the council gate). The
`needs_diagnosis` item is `complete` today. No durable verdict surfaced in `doc_notes`
(only unrelated council-gate notes match), so treat the primary evidence below as the record.

## CORRECTION 1 — the writer does not "invent a plausible URL". The **emitter** fabricates it and hands it over as an instruction + an acceptance test.

`platform/orchestration/actions/create_tool_cross_link_items.go:142`:

```go
toolPageURL := fmt.Sprintf("/tools/%s.html", toolFunction)   // suggestion-time, prefix intact
```

It reads `evaluation.result.suggestions` (`:88` — the same suggestion-time array
`create_items_loop` iterates), and bakes that fabricated URL into BOTH the rewrite
instruction and its acceptance test (`:173-186`). Verified in a live spec on leopardess:

> `suggestion`: "Weave a natural reference to 'Monitoring Coverage Gap Finder' (…). **Link to
> /tools/tool-monitoring-coverage-gap-finder.html.** …"
> `acceptance_test`: "Page contains at least one inline link to
> /tools/tool-monitoring-coverage-gap-finder.html …"

The live `page-build-handler` maps `rewrite_guidance? → input_data.spec.suggestion`
(pod config on v1.0.1144; the `k8s/bk_agent_definitions_backup.sql` copy is stale and says
otherwise — migration 072 threaded it). So `page-content-writer` is **obeying a phantom URL
it was told to use**, and the item's own acceptance test *requires* the phantom. This is not
an LLM confabulation; it is deterministic, and it is the emitter's.

## CORRECTION 2 — it is not "tools that were never built" that 404. It is **all of them**, because `/tools/{function}.html` never matches the real `pages.url`.

Fleet sweep, 2026-07-21 — every `source='tool-suggester'` `content_rewrite`, its constructed
`/tools/{tool_function}.html` left-joined to `pages.url` for that site:

```
24 items across 3 sites (leopardess 10, gamesdesign 13, ai-agent-orchestration 1)
0 of 24 constructed URLs resolve to a real page  — matched_page_url NULL on every row
```

Including a tool that **was** built: `tool-process-automation-scorer` deployed at
`/tools/process-automation-scorer/index.html`, but the item points at
`/tools/tool-process-automation-scorer.html`. The deployed-tool URL shape is
**non-deterministic across build paths** and cannot be reconstructed from the function name:

| build path | example | url shape |
|---|---|---|
| `deploy_tool_action.go:269-277` (library fork) | `tool-fuel-cost-estimator` | `/tools/fuel-cost-estimator.html` (**strips** `tool-`) |
| `create_tool_component_action.go:211-218` (novel, `CanonicalisePage`) | `tool-process-automation-scorer` | `/tools/process-automation-scorer/index.html` (`/index.html`) |
| observed fleet-wide | `tool-drop-rate-tuner`, `tool-loot-table-balancer` | `/tools/tool-…​.html` (**keeps** `tool-`) |

The emitter keeps the prefix and always appends `.html`. So it is wrong on **all three**
shapes. **The URL cannot be constructed; it must be looked up from `pages.url`, and the
lookup is only meaningful once the tool page row exists.** That is the whole bug in one line.

## Why nothing downstream caught it (three layers, all confirmed at code)

1. `internal-link-resolver` (`resolve_internal_links_action.go`) resolves only structured
   **CTA fields** (`ctaFieldNames`, `:98-105`), against real `pages.url`. It never looks at an
   arbitrary in-body prose `<a href>`, and it is **not wired onto content_rewrite output**.
2. `validate_page_content.go` `validateInternalLinks` (`:540-582`) **does** extract every
   in-body href and check it against real `pages.url` — and files a phantom as
   `Severity:"warning"` (`:571`), which is **non-blocking** (`valid := blockerCount==0 &&
   errorCount==0`, `:257`). So the page deploys with the 404, by design ("the improvement loop
   resolves it" — it doesn't; cf. 023/033/049 detected-but-not-delivered family).
3. Post-deploy `check_phantom_internal_links.go` exists but has the coverage/durability gaps
   documented in `049`.

## Fix chosen: emit from the tool-BUILD success path, using the real page URL (candidate 1, made concrete)

The race is removable **by construction**: the `add_tool` item spec already carries
`related_pages` (the suggester writes `spec_data: current_suggestion`, verified live), and the
build actions already create the tool page and emit follow-on items (`needs_content_page`,
companion guide) carrying the real `page_id`/`tool_page_url`. So:

- Emit the `content_rewrite` cross-link items from `deploy_tool_action.go` /
  `create_tool_component_action.go`, **after** the page row exists, using the **real
  `pageURL`** they just computed; read `related_pages` from the incoming `add_tool` spec.
- Extract the spec-builder into a shared helper (reuse `create_tool_cross_link_items.go`'s
  body, delete `:142`'s fabrication, take `realURL` as a parameter).
- Remove the suggester's `create_cross_links` workflow step (098) so nothing emits at
  suggestion time any more.

Result: a cross-link item only exists for a tool whose page row exists, and it carries a
resolvable URL. **Residual, deliberately deferred to `049`:** if the tool page is created
(`planned`) but its content build never deploys, the link still 404s — that is 049's
mechanism 2 (planned-but-unbuilt page linked), a broader class, not this emitter's defect.

Existing damage (the 24 items + their woven links on live pages) is **not** cleaned up by the
emitter fix — that is a separate sweep, coordinated with `049`.

---

# FIXED 2026-07-26 (bugfix-029 session) — emitter moved to the tool BUILD paths, config half live, Go half awaiting the next image

## What changed

**Go (`platform/orchestration/actions/`)** — the URL is never constructed again:

- `create_tool_cross_link_items.go` — `:142`'s `fmt.Sprintf("/tools/%s.html", toolFunction)` is
  gone. The file now holds `emitToolCrossLinkItems`, a shared emitter that **takes** the tool
  page's real `pages.url` and refuses to emit if handed anything that is not an absolute path,
  plus `resolveToolPageURL` (reads the URL via `page_components → content_components.function`,
  falling back to `pages.name`; both READ `pages.url`).
- `deploy_tool_action.go` / `create_tool_component_action.go` — call the emitter after the page
  row and its `needs_content_page` item exist, with the `pageURL` they just wrote and
  `related_pages` from the `add_tool` spec (new optional action input). `deploy_tool_to_site`
  also emits on its already-deployed early return, so **re-running the deployer is the supported
  way to backfill cross-links** for a tool deployed before this fix.
- The suggestion-time action is **kept registered and made fail-safe** rather than deleted: an
  unregistered action named in config invalidates a workflow (`bugs_closed/017`), and config can
  come back from a stale backup. It now resolves a real page and emits nothing when there is
  none — it cannot fabricate from anywhere.

**Config (`211_tool_crosslink_emit_at_build.sql`, applied + recorded 2026-07-26)** — deletes
tool-suggester's `create_cross_links` step (`create_items_loop → complete`) and wires
`related_pages` into both build steps. **This half is LIVE now**, so no new phantom items can be
created regardless of when the image ships. Parts 2/3 are inert on the deployed binary and
activate with it; the Go side also reads `input_data.spec.related_pages` directly, so the halves
can roll in either order.

## Beyond the diagnosis: the items are GATED on the tool page going live

The "VERIFIED + SHARPENED" section above deferred *"tool page created (`planned`) but its content
build never deploys → link still 404s"* to `049`. **That deferral is withdrawn for this emitter.**
It is 049's class only while the emitter runs at suggestion time with no relationship to any
build; once the emitter sits inside the build path, it is this code's own remaining failure mode,
and it reproduces exactly the damage this bug is about. It is also not rare: 19 of 33 live
`needs_content_page` items are parked in `needs_human_review`.

So `emitToolCrossLinkItems` emits immediately only when the tool page is already
`deployed`/`needs_rebuild`; otherwise it attaches `depends_on` = the open `needs_content_page`
item for that page, and if there is no open item (or it failed terminally) it emits **nothing**.
The loader already enforces this (`load_work_item_actions.go:562-571`: an item is selected only
when every `depends_on` row is `complete`/`verified`). **Cost, stated:** a tool page whose content
build never completes leaves cross-link items parked in `triaged` instead of writing a dead link;
parked items age and may be swept by `bugs_open/070`. That is the intended direction.

## Evidence re-grounded 2026-07-25 (the 07-21 figures were stale, and the bug had grown)

R1 re-run: **27 items across 4 sites** (was 24 / 3 — fundamentallyai.com joined), **0 of 27
resolve**. The emitter kept firing for the four days between diagnosis and fix.

## How to verify (post-roll) — RUNBOOK R6/R8

1. Pod-grep a string the change CREATED, with a control:
   `strings /app/agent-chassis | grep -c "emitToolCrossLinkItems: refusing to emit without a real tool page URL"`.
2. Trigger a tool build on a test site; the new `tool_crosslink:%` row's `spec->>'tool_page_url'`
   must equal that tool page's `pages.url` (R6 joins on exactly that).
3. Config half is checkable now, independent of the image: `create_cross_links` absent,
   `create_items_loop.next_step='complete'`, `related_pages` wired on both build steps.

## Still open, deliberately — NOT closed by this fix

**The existing damage.** 27 items and the links already woven into live pages
(leopardessconsulting.co.uk `/services.html` among them) are untouched by an emitter fix. That is
a content sweep to coordinate with `049`. Pre-fix rows are identifiable by
`item_key LIKE 'tool_crosslink:%' AND spec->>'tool_page_url' IS NULL`.
