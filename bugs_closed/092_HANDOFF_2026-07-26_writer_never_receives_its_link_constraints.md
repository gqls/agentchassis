# 092 — the page writer is never told which pages exist, on every run

**Filed:** 2026-07-26, while fixing `bugs_open/079` (the deploy gate detected dead in-body
links and published them anyway).
**Severity:** high. This is the *upstream cause* of the invented links 079 has to clean up.

> **STATUS 2026-07-31 — FIXED IN CODE, COMMITTED (`2e1bfb39e`), NOT YET LIVE.**
> Fix candidate 1 implemented, plus candidate 3 and both traps. Council submission
> `4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d` (`Council-Submitted:` trailer — verdict pending
> at commit time). **This file stays OPEN until the chassis rolls and a fresh writer run
> records `page_count > 0`**: the defect is reproducible until it ships, and a commit is
> not a deploy. Workstream:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_092_writer_link_constraints/`.
> What was verified before the fix (not assumed): the diagnosis below **still held on
> 2026-07-31** — 26 of 26 runs at `page_count 0`, latest 15:36 UTC the same day, and
> neither action file had moved since 2026-03-28.

**Status (original):** OPEN — diagnosed with evidence, not fixed. Deliberately not folded
into 079's commit: different mechanism, different file, different agent's path, and it
changes content-generation behaviour fleet-wide, which nobody has measured.

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

---

> **NOTE ON THE PARAGRAPH BELOW, 2026-07-31:** the 07-28 correction stands and is the
> reason this bug was worth fixing rather than deferring — there is no downstream
> mitigation, so the writer's inventions reach live pages. The fix at the bottom of this
> file is prevention, and it does not repair a single already-deployed page.

> **CORRECTED 2026-07-28 (brochure_component_library thread):** the paragraph above —
> *"invented links are being removed before they ship"* — is **false**. `079`'s repair
> runs and its output is **discarded at persistence**: `save_page_sections` takes the
> structured `sections_metadata` path and never reads `validation_result.clean_html` on
> the primary build plan. Proven in production twice on 2026-07-28 (fundamentallyai
> capabilities: repair logged 10:45:01.347Z, unrepaired components saved 400ms later,
> all 9 targets 404 on the live page; vonc /about.html same shape). `bugs_open/079` is
> **REOPENED** with the mechanism cited. Consequence for sizing THIS bug: there is
> currently **no downstream mitigation at all** — the writer invents destinations (and
> `src` paths, which were never in the repair's remit) and they ship. The urgency
> discount in the paragraph above is withdrawn; this bug is the upstream cause of live
> 404s on deployed pages today, not of paragraphs that lose their links.

---

# THE FIX, 2026-07-31 — committed `2e1bfb39e`, inert until the next chassis roll

Workstream: `docs/agent_docs/docs024_key_docs_latest/bugfix_092_writer_link_constraints/`.
Council: `Council-Submitted: 4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d`.

## Re-verified before touching anything (the bug was still live)

```
runs | zero_pages |            latest
  26 |         26 | 2026-07-31 15:36:26+00
```

and, from the other side — over **all** `page-content-writer` orchestrations, not only
those that recorded a link context:

```
writer_runs | has_input_site_id | has_site_record | has_toplevel_site_id | has_db_sync
         26 |                26 |               0 |                    0 |           0
```

That second query is the one that decided the fix. It says the configured field and all
three fallbacks are unreachable **and** that `input_data.site_id` is the only identity the
path carries — which the package's shared `extractSiteID` does not look at. Everything
else followed from it.

## What changed

1. **The database is the source** (`loadLinkablePages`), under *exactly* the predicate
   `validate_page_content.go:1252` `loadValidPagePaths` uses to decide what is a
   `phantom_link`. The invariant that matters is **the writer's allow-list equals the
   gate's accept-set**: any other source can disagree with the gate, and a writer flagged
   for obeying its own instructions is the drift class this codebase keeps paying for.
   `collected_data` remains the fallback, so a workflow that declares a list keeps it.

2. **An empty list now says so.** `buildLinkConstraintText` returned `""`, and the
   consuming template's `{{if}}` guard then dropped the entire "## Internal Linking"
   section — so the writer with the *least* information was given the *least* instruction.
   It now emits "do NOT create any internal links", which is true and safe in both states
   that produce it.

3. **The two causes of an empty list no longer look alike.** `degraded` is *"the list
   could not be established"*, never *"the list is empty"*; a brand-new site with no pages
   is a correct empty list. `db_consulted`, `source` and `reason` are in the output, and an
   unreadable list writes a durable `agent_error_log` row, `error_code`
   `LINK_CONTEXT_UNAVAILABLE`. Candidate 3 ("fail loudly") — a `logger.Warn` and an elided
   prompt section is *why this ran at 100% for seven weeks*.

4. **Trap 2 closed:** no URL is ever synthesised from a page name. A url-less page is
   dropped and counted.

5. **Trap 1 closed by deletion:** `link_constraints.go` is gone — 173 lines, 5 symbols,
   **0 call sites** (grepped per symbol). It carried its own copy of the synthesis plus two
   extra guesses (`/blog/`, `/tools/` prefixes). The standing landmine "do NOT wire
   `InjectLinkConstraints`" is retired rather than restated. `page-content-writer`'s
   `default_config` still carries a dead `link_constraints` block; it is now *provably*
   unread, and is left alone deliberately (a config change with a separate risk profile).

## Measured before submitting, not left for a reviewer

| claim | measurement |
|---|---|
| one consumer | 1 `agent_definitions` row references `prepare_link_context` (`page-content-writer`) |
| the fix reaches every run | 8 distinct writer runs in window, site id resolvable on all, **31** linkable pages each — `0 → 31` |
| the predicate choice is safe | `pages.status` fleet-wide is **only** `active` (449) / `archived` (23), so the gate's predicate and `loadActivePagesForLinkContext`'s `status='active'` are the same set today |
| the synthesis path was unreachable | `pages.url` is `NOT NULL` and **0 of 472** rows are empty |
| the prompt cap is inert | largest site has **99** linkable pages, mean 30; cap is 200 |

## How to verify it went live (do NOT verify from git or the tag)

```sql
-- 1. a fresh writer run now carries a real list
SELECT created_at,
       collected_data->'link_context'->>'page_count' AS pages,
       collected_data->'link_context'->>'source'     AS source,
       collected_data->'link_context'->>'degraded'   AS degraded,
       length(collected_data->'link_context'->>'link_constraint_text') AS text_len
FROM orchestration_states
WHERE collected_data ? 'link_context'
ORDER BY created_at DESC LIMIT 5;
-- pre-fix: pages=0, text_len=0 on every row. Post-fix: pages>0, source='database'.

-- 2. the loud arm, if anything cannot resolve
SELECT occurred_at, severity, error_message, context
FROM agent_error_log WHERE error_code='LINK_CONTEXT_UNAVAILABLE'
ORDER BY occurred_at DESC LIMIT 10;
```

Pod-grep the running chassis for a string this change ADDED, plus a positive control in
the same exec (a roll is not evidence your fix shipped — `bugs_open/153`):

```
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "LINK_CONTEXT_UNAVAILABLE"; \
   strings /app/agent-chassis | grep -c "PrepareLinkContextAction"'
```

The second is the control: it is present in every build, so `0 0` means the grep is wrong,
while `0 N` means the image predates this commit.

**Do not verify by reading the prompt template.** It is correct and always has been; the
data it interpolates is what was missing.

## What this fix does NOT do

- It repairs **no already-deployed page**. This is a write-path fix; the existing 404s on
  live sites are `bugs_open/071`'s and `097`'s.
- It does not touch the three hardcoded `/contact.html` defaults in `component_library.go`
  (`071`'s renderer-default class) — a different producer, downstream of the writer, which
  a writer-side fix structurally cannot reach.
- It does not address the fragment/anchor blind spot (`071`): nothing emits section `id`s.
- **Residual found while fixing this, deliberately not widened into:**
  `build_render_context`'s `sources` map takes `available_pages` from `db_sync.pages` too,
  so `render_context.available_pages` is empty on the same 26 runs, for the same reason.
  Nothing in Go and no live prompt reads it by that name, so it is inert rather than
  harmful — recorded here so the next thread does not rediscover it as a second bug.

---

## Council verdict, 2026-07-31 — APPROVED at round 1, and what the objections were worth

`4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d` → **approved**, "6 advisory objection(s) — none
high-severity", 6 seats abstained on relevance. Three mediums were worth acting on; the
rest are recorded here because a verdict nobody reads is the same as no verdict.

**Acted on in code** (`9a57d2395`):

- **`reuse_agent`, medium — and it was right.** I named `loadValidPagePaths` as the reuse
  target and then *copied its predicate*, with a comment as the only thing holding the two
  in step: *"the precise failure mode the founding incident describes, just one layer more
  sophisticated (documented duplication vs. blind duplication)"*. The predicate is now a
  shared constant used by both queries, so divergence is unrepresentable rather than
  discouraged. Not the whole query — the two callers need different projections (urls
  indexed by normal form vs. title/description for the prompt), so factoring it would force
  one shape on both. Pinned by `TestGateAndWriterShareOnePageEligibilityPredicate`.
- **`debug_historian`, low.** Stated explicitly that `pages.status` is load-bearing (dispatch
  and `validate_page_content` both branch on it), *unlike* the `sites.status` informational
  column whose landmine warns that filtering on it silently blinds a query. The analogy is
  inverted here, and it was being left to the reader.

**Answered with a measurement instead of an edit:**

- **`editquality`, medium** — the deletion of `link_constraints.go` rested on a code-search
  absence claim *"that cannot be confirmed from SQL"*. It can, and now is: **no
  `agent_definitions` row anywhere references `InjectLinkConstraints` or
  `inject_link_constraints` in its config**, and `link_constraints` is not a registered
  action name. The only surviving reference is the dead top-level config block on
  `page-content-writer` (`{"enabled": true, "max_internal_links_per_section": 3}`), which is
  data no code reads — now provably so.
- **`guardian`, medium** — *"is `params.DB` a NEW field on `ActionParams`?"* No. It is
  pre-existing plumbing (`types.go:54`); this lane never touched `types.go`. A fair question
  to gate on, and the answer is one `git status` away.
- **`guidelines`** — asked whether `page-content-writer` declares `input_contract` /
  `output_contract`. **It does not, and neither does any other agent: 0 of 185 active
  definitions declare either field.** So the DECLARED CONTRACTS guideline is *inert
  fleet-wide*, which is what that seat itself suspected ("this reads as the GUIDELINE being
  inert/stale here rather than the plan being wrong") — now measured rather than suspected.
  Worth knowing for every future submission that draws this objection.

### `bug_historian`, medium — the sibling audit it asked for, done

> *"the plan's own rationale names a sibling silent-empty exposure on the same shared
> mechanism and leaves it unaudited … should at minimum file a follow-up work item naming
> which of the five callers can hit an unresolved site_id and silently no-op."*

Audited. Of `extractSiteID`'s five callers, **three fail loudly** and are safe —
`site_db_actions.go:229` and `:533` return `fmt.Errorf`, `get_pages_to_build_actions.go:38`
likewise. **Two are exposed:**

| caller | on unresolved `site_id` |
|---|---|
| `UpdateSiteTimestampsAction` (`site_db_actions.go:490`) | `logger.Warn` + `{"updated": false}` — a silent no-op |
| `ExtractAndSyncLinksAction` (`site_db_actions.go:396`) | **no check at all**; returns `{"links_extracted": N, "persisted": false}` — a *success-shaped* result where the link registry is never written |

The second is the worse shape: a caller reading `links_extracted: 12` has every reason to
think the registry was populated.

**And the fact that stops this becoming a claim:** `link_registry` holds **0 rows,
all-history** (`count(*) = 0`, no `max(created_at)`). Both exposed actions run on exactly
one agent, `multipage-website-builder` — and there are **0 `multipage-website-builder`
orchestrations in the retained window**.

> **[UNDETERMINED] — I cannot tell from a zero-run window whether the registry is empty
> because this exposure fires, or because the agent simply never runs.** Those are the two
> causes with opposite fixes, and asserting either would be exactly the error this bug is
> made of. What is measured: the exposure exists in the code, and the table has never had a
> row. What is not: which one explains the other.

Not filed as a new bug: `bugs_open/165` already owns `link_registry` (`site_db_actions.go:1474`,
the reconciliation delete) and is actively worked, so the measurement is contributed there
rather than competed with. `LNK-001`'s standing `verify-later: link_registry population`
is the register-side home for it.

---

# LIVE on v1.0.1219, 2026-07-31 19:09Z — proven at the pod, in both directions

Chassis rolled to `v1.0.1219` (pods `agent-chassis-59cb674798-t7dgn` 19:09:31Z,
`…-z84n8` 19:09:52Z). **A roll is not evidence your fix shipped** (`bugs_open/153`) — the
image may predate the commit — so this is a string check on the running binary, on **both
replicas**, with a positive control in the same exec and both directions tested:

```
                                    t7dgn   z84n8
ADDED    LINK_CONTEXT_UNAVAILABLE     1       1     <- new error code
ADDED    "There are NO pages ..."     1       1     <- the explicit empty-list instruction
REMOVED  "## INTERNAL LINKS"          0       0     <- link_constraints.go, deleted
REMOVED  "## Internal Links"          0       0     <- the old duplicate heading
CONTROL  PrepareLinkContextAction     8       8     <- present in every build, incl. pre-fix
```

The control is what makes the zeros mean something: `0 0` across the board would say the
grep is broken, and `0 N` would say the image predates the fix. `N N` with the removed
strings at zero is the pass, and the **removed** rows are the half a one-directional check
would miss — they prove this is the new binary rather than merely *a* binary containing the
new string.

> **Note on the two commits.** `2e1bfb39e` (the fix) and `9a57d2395` (the round-1 review
> answer) cannot be distinguished by pod-grep, and that is not a gap: the second is a
> refactor — extracting the shared `linkablePageStatusPredicate` — that Go **constant-folds
> into byte-identical SQL**. It changes no runtime string and no runtime behaviour; its
> value is that drift becomes unrepresentable, and that is enforced by a test, not by the
> binary. Do not go looking for a marker that cannot exist.

## CLOSED — the induced run, 2026-07-31 19:16Z

Induced deliberately rather than waited for: writer runs are irregular (26 today, in bursts),
and the queue would not reliably produce one. Dispatched `page-build-handler` directly by
kcat (`corr dc7d9e77-ae07-4633-8eef-1fdd647b48b2`), **with the `PUBLISH_OK` receipt** — the
`kubectl run -i | kcat -P` pattern silently produces nothing about four times in five, so a
publish without a receipt is not a publish.

**The target was chosen so the induction could not write anything.**
`loancalculator.co.uk/guide-can-i-overpay` is `rebuild_policy='owned'`, on a site whose
status is `active` not `deployed`, and the page is undeployed. `save_page_sections` refuses
`owned` pages — a pre-existing guard, so *nothing this run produced could reach the page* —
while `prepare_link_context` runs long before that refusal. Confirmed after the fact:

```
page-build-handler     complete_error   COMPLETED   <- refused at save, as intended
page-content-writer    complete         COMPLETED   <- the writer ran in full
internal-link-resolver complete         COMPLETED

pages.updated_at for the target = 2026-07-30 22:10:19+00   -> NOT touched by the run
```

### The decisive row

```
        created_at         | pages |  source  | db_consulted | degraded | text_len |                  reason
---------------------------+-------+----------+--------------+----------+----------+------------------------------------------
 2026-07-31 19:16:21.96+00 |    27 | database | true         | false    |     2739 | 27 linkable page(s) read from the pages table
```

Against **`0` / `null` / `0`** on all 26 pre-fix runs. `source: database` is the fix's own
path saying so; `db_consulted: true` with `degraded: false` is the authority confirming it
was actually read, not merely defaulted.

### The list is real `pages.url`, not synthesised

All **27 of 27** listed addresses match a stored `pages.url` for that site exactly
(`position(p.url in <constraint text>) > 0` for every active page). The block now reaching
the writer begins:

```
When creating internal links, ONLY link to these pages:

- /guides/can-i-overpay.html (Can I Overpay My Loan? UK Rules & ERCs Explained)
- /guides/car-finance-explained.html (PCP vs HP Car Finance | loancalculator.co.uk)
- ...
```

— and note it starts at the instruction, not at a heading: the duplicate `## Internal Links`
is gone, because the prompt template supplies `## Internal Linking` on the line above.

`agent_error_log` holds **0** `LINK_CONTEXT_UNAVAILABLE` rows, which is the correct result —
nothing degraded, and it confirms the loud arm is not firing spuriously.

**Fixed AND live AND proven ⇒ `bugs_open/` → `bugs_closed/`.**

---

## The `[UNDETERMINED]` above is RESOLVED — 2026-08-02, by the `165` lane

The sibling audit ended on an honest refusal to guess:

> *"I cannot tell from a zero-run window whether the registry is empty because
> this exposure fires, or because the agent simply never runs. Those are the two
> causes with opposite fixes."*

**It is the second: the agent never runs.** Measured today, and the discriminator
is cheap — if the exposure had ever fired, the action must have executed at all,
and it has not:

```sql
-- has the action EXECUTED, ever, on any agent?
SELECT count(*) FROM orchestration_states WHERE collected_data::text LIKE '%links_extracted%';
-- 0

-- has its only carrier run?
SELECT count(*) FROM orchestration_states WHERE owner_agent_type='multipage-website-builder';
-- 0

-- what the live build pipeline actually is, same window:
--   build-dispatch-loop 588 · build-pipeline-trigger 587 · page-rerender 22 · page-build-handler 1
--   multipage-website-builder: absent
```

`extract_and_sync_links` is carried by exactly one agent
(`multipage-website-builder`, `is_active=true`, two rows), and that agent does not
appear in the live build pipeline at all while the real builders ran ~1,200 times
in the same window.

**Scope this honestly — "never" is bounded.** `orchestration_states` is
retention-clocked: its oldest surviving row is **2026-07-13**, so the true claim
is *"has not run in the ~20-day retained window"*, not "never in history". What
IS all-history is the other half: `link_registry` has **0 rows and NULL
`max(created_at)`** with no retention job on that table, so the registry has never
held a row since it existed. The two together are what make the conclusion safe —
the table has never been written, and the only thing that could write it has not
run in the entire observable window.

**So the exposure at `site_db_actions.go:396` is real code on a dormant path.**
It should still be fixed — a success-shaped `{"links_extracted": N, "persisted":
false}` is exactly the shape this bug is made of — but it is not the reason the
registry is empty, and fixing it will not populate anything.

### And the reference was circular, which is why this sat unowned

This file said *"not filed as a new bug: `bugs_open/165` already owns
`link_registry` … so the measurement is contributed there rather than competed
with."* `165`'s site-C section said the opposite: *"that is `bugs_open/092`'s
territory, not this floor's."* **Each deferred to the other, so neither owned it,
and both then closed.** The 165 lane repeated the pointer today — writing
"blocked on `bugs_open/092`" into a bug file, the 016b index and the concept
register — without checking that 092 had been closed on 2026-07-31, the same day
the pointer was written. Paths corrected in all three.

The transferable bit: **a deferral names a destination, and nobody re-checks that
the destination accepted it.** Two "contributed there rather than competed with"
notes, written in good faith a few hours apart, produced an orphan. When you defer
an item to another case, say so *in that case's file*, not only in your own.

### What is actually left, and it is an owner call rather than a bug

`multipage-website-builder` is `is_active=true`, carries the only copy of
`extract_and_sync_links`, and has not run in the observable window while a
different set of agents does the building. Either it is **retired** (in which case
the action, its exposure and site C's floor are all dead code, and `is_active`
is lying) or it is **dormant-but-intended** (in which case the exposure is a live
landmine waiting for its first run). That is a decision about the platform's
build path, not a defect to diagnose — raised rather than filed.

Consequence for `bugs_closed/165` site C: its floor cannot be induced live until
that decision goes one way, and it is inert and risk-free meanwhile.


> **CORRECTION 2026-08-02 (later the same day) — the retention figure above is WRONG, and the conclusion needs a different source.**
>
> Everything above bounds "0 orchestrations" with *"`orchestration_states` is
> retention-clocked (oldest row 2026-07-13), so that is ~20 days"*. **It is ~24
> HOURS.** `COMPLETED` rows are reaped after about a day — measured: 2,504
> COMPLETED rows, oldest **24.7h**; FAILED oldest **25.4h** — and the whole-table
> `min(created_at)` reads 2026-07-13 only because `CANCELLED` (24), `RUNNING` (4)
> and `INITIALIZED` (2) are **not** reaped. A handful of stragglers in statuses the
> census was not about set a floor twenty times too long. Caught by watching a row
> I had quoted at 09:40 (`dcf88c1c…`) vanish by 10:40 while the table grew
> 2,454 → 2,546 and its oldest row never moved.
>
> **So the orchestration evidence never supported "the agent never runs" — only
> "not in the last day".** The conclusion is still correct, but on a different and
> much stronger source: **`site_specs` has no retention job and goes back to
> 2026-02-25** (1,874 rows, 36 sites), and across all of it the only
> `recommended_builder` ever recorded is **`pageflow-builder`** — 1,216 rows, 14
> sites. `multipage-website-builder` was never chosen, not once, in five months.
>
> Right answer, wrong reason, and the reason is what was published. Fleet landmine
> filed: "`orchestration_states` keeps terminal rows ~24 HOURS — and
> `min(created_at)` says 20 days". **Any other claim resting on an
> `orchestration_states` census needs re-bounding per status.**
