# 103 — every deployed tool page publishes its internal build brief as the public meta description

**Filed:** 2026-07-27 · **By:** gauntlet_dead_cta (found while surveying vonc.com's
Arena after the v1.0.1172 roll) · **Severity:** MEDIUM — public-facing on 16 live
pages across 6 sites; one of them tells search engines the page has no backend ·
> ## ⛔ BACKFILL ON HOLD — owner instruction, 2026-07-27
>
> **Do not run `docs/agent_docs/sql_for_agents/240_backfill_103_tool_meta_descriptions.sql`.**
> The owner has held it until the **Gemini content-provider thread has finalised its
> prompt** (`gemini_content_provider/`, `bugs_open/107`). The reason is sequencing, not
> doubt about the backfill: composed copy written now would be superseded by whatever the
> settled prompt produces, and rewriting six sites' public descriptions twice is worse
> than waiting.
>
> The code fix is committed and unaffected — it only governs pages created from now on,
> and it is inert until the next chassis roll regardless.
>
> **When the hold lifts:** re-run STEP 1 first. The row set moves (it was 16 → 15 → 17
> inside one day), so the count in that file is a starting point, not a target.

**Status:** OPEN — code fix **LIVE on v1.0.1177** (rolled 19:22:02Z; running-pod grep
`"rejected as a build brief"` → 2), council **APPROVED**
(`52241d09-287e-4d15-9010-400f78339298`, 10 reviewers, 0 unreadable, 5 advisory
objections, no veto), **inert until the next chassis roll**; the backfill is staged
and **deliberately not applied**. §3 of this file still holds — the code fix alone
changes nothing on the web, so the 17 live rows are untouched.

> ## Two corrections from taking this on, 2026-07-27 (bugs thread)
>
> ### 1. There is a SECOND call site, and this file names only one
>
> `create_tool_component_action.go:261-265` also creates a tool page and bound the same
> `description` into `meta_description`:
>
> ```go
> INSERT INTO pages (
>     id, site_id, name, url, title,
>     page_type, status, build_status, nav_order, meta_description
> ) VALUES ($1, $2, $3, $4, $5, 'tool', 'active', 'planned', 200, $6)
> `, pageID, siteID, pageName, pageURL, pageTitle, description)
> ```
>
> Found by grepping every writer of the column rather than trusting the filed call site.
> Fixing only `deploy_tool_action.go:341` would have left the class live on this path and
> **looked complete in review** — the exact shape of `bugs_open/093` (one guarded call
> site, rerender unchecked) and `bugs_open/112` (the second spawner).
>
> ### 2. The live count is 17, not 15 — this file's census undercounts
>
> §1's census uses `length > 400`. At **320** — the threshold the fix uses, chosen to sit
> clear of both populations — two more genuine briefs appear:
>
> | site | page | len | starts |
> |---|---|---|---|
> | gaswholesalers.com | `tool-fuel-cost-estimator` | 352 | "Allows fleet managers and fuel buyers to input weekly/monthly volume (gallons or litres)…" |
> | gamesdesign.co.uk | `tool-damage-formula-designer` | 390 | "Lets designers define a damage formula (flat, multiplicative, diminishing returns scaling)…" |
>
> Both were read individually, not inferred from the length: they are in the
> "Allows/Lets the user do X" register of a specification, not visitor-facing copy. So the
> repair set is **17 rows across 6 sites**, not 15 across 5.
>
> ### What was built
>
> - `datahelpers.PublicMetaDescription` / `MetaDescriptionLooksInternal` — one gate
>   between a string held internally and the text a search engine prints. It refuses a
>   brief-shaped **fallback** too, returning empty, so the escape hatch cannot
>   reintroduce the defect.
> - Both call sites routed through it, with a shared `composedToolMetaDescription` in the
>   register the companion guide page has always used.
> - Tests use the **real leaked string** as the internal fixture and the hand-fixed Arena
>   copy now live as the public one.
>
> ### What was deliberately NOT done
>
> - **The backfill is staged, not applied**: `docs/agent_docs/sql_for_agents/240_backfill_103_tool_meta_descriptions.sql`.
>   Dry-run verified (17 rows / 6 sites, negative control returns 1). It rewrites public
>   copy on six client sites, so it is an owner call. **§3 of this file still holds — the
>   code fix alone changes nothing on the web.**
> - **`meta_description` was NOT added to the tool page's `ON CONFLICT … DO UPDATE SET`**,
>   which would have made redeploys self-healing. vonc.com's Arena description was
>   repaired **by hand**; adding the column would clobber that hand-written copy with the
>   generic composed line on the next deploy. A conditional update — overwrite only when
>   the existing value is itself brief-shaped — is the obvious next step and deserves its
>   own review rather than riding along.

## 1. Symptom

`pages.meta_description` — the text search engines show under the result — is the
tool's **internal build specification**, verbatim. The worst case is live now at
`https://vonc.com/tools/arena/index.html`, 1,206 characters:

> "The Arena is Spark's competitive mode, v1 as a fully self-contained
> client-side experience (**no fetch calls, no backend**). Four elements, in
> order: (1) TODAY'S PROVOCATION — a bold prompt displayed prominently at the top
> (**embed 5 sample provocations in JS and pick one by day-of-date** so the page
> changes daily, e.g. …"

That is the instruction given to the generator, published as the page's own
description. It leaks implementation detail, it reads as machine output, and
Google truncates at ~155 chars so what actually renders is a fragment of a spec.

The other 15 are the same defect in a less embarrassing register — tool specs
written for an engineer, not a visitor:

| site_id (prefix) | page | len |
|---|---|---|
| `9ec3b9ee` (vonc.com) | `tool-arena` | **1206** |
| `4851f6fc` | `tool-process-automation-scorer` | 637 |
| `4851f6fc` / `2a8ebf9c` | `(tool-)agent-complexity-estimator` | 607 |
| `4851f6fc` / `1368e337` / `199733a8` / `2a8ebf9c` | `(tool-)llm-cost-calculator` | 543 |
| `1368e337` / `2a8ebf9c` / `4851f6fc` | `(tool-)ai-agent-roi-estimator` | 514 |
| `1368e337` | `tool-ai-data-risk-checker` | 508 |
| `199733a8` | `tool-model-approach-selector` | 505 |
| `5fe15466` | `tool-breakeven-volume-calculator` | 461 |
| `1368e337` | `tool-ai-readiness-quiz` | 460 |
| `5fe15466` | `tool-fuel-budget-forecaster` | 449 |

Census query (re-run it; do not trust this table's counts):

```sql
SELECT count(*) AS pages, count(DISTINCT site_id) AS sites
FROM pages
WHERE meta_description IS NOT NULL
  AND (meta_description ~* 'no fetch calls|elements, in order|embed [0-9]+ sample|client-side experience'
       OR length(meta_description) > 400);
-- 2026-07-27: 16 | 6
```

## 2. Root cause — cited

`platform/orchestration/actions/deploy_tool_action.go`. The tool page INSERT at
**line 321-343** binds `$9 = toolDescription.String` into `meta_description`:

```go
	SELECT name, display_name, function, category,
	       description, html_template, semantic_tags::text, input_schema::text
	FROM content_components
	WHERE id = $1 AND component_level = 'tool' AND is_active = true
```
```go
	INSERT INTO pages (… meta_description, sections, …)
	VALUES (…, $9, $10::jsonb, 'planned', 'active')
	…
	`, siteID, pageName, pageURL, pageTitle,
		navSection+" / "+toolDisplayName, maxNavOrder+1, inHeader, inFooter,
		toolDescription.String, sectionsJSON,   // <-- line 341
	).Scan(&pageID)
```

`content_components.description` for a `component_level='tool'` row is the
**build brief**, not marketing copy. Proven byte-identical, not merely similar:

```sql
SELECT cc.name, length(cc.description),
       cc.description = (SELECT p.meta_description FROM pages p
                         WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
                           AND p.name='tool-arena') AS identical
FROM content_components cc WHERE cc.id='faa69bcc-f94e-4ca0-ac76-14a03da4807c';
-- tool-arena-interface-vonc-com | 1206 | t
```

**The correct pattern already exists in the same function, ~100 lines below.**
The companion *guide* page (line 438-458) composes a human-facing description
instead of copying the brief:

```go
	fmt.Sprintf("A practical guide to %s — what it means, how it works, and how to use our interactive %s.",
		strings.TrimPrefix(toolDisplayName, "UK "), strings.ToLower(toolDisplayName)),
```

So this is not an unsolved problem — it is one of two sibling writes in one
function disagreeing, and the wrong one is the visitor-facing page.

## 3. It does not self-heal

The tool-page statement's conflict clause deliberately omits the column:

```go
	ON CONFLICT (site_id, name) DO UPDATE SET
		url = EXCLUDED.url, title = EXCLUDED.title,
		sections = EXCLUDED.sections, updated_at = NOW()
```

`meta_description` is **not** in the `DO UPDATE SET` list. Redeploying a tool
therefore leaves the bad description in place. **A code fix alone repairs nothing
already live** — the 16 existing rows need a backfill, and any fix that ships
without one will look correct in code review and change nothing on the web.

## 4. Fix candidates — ordered by what closes the door

1. **Stop the tool page from ever receiving the brief.** Give tools a distinct
   public-copy field (e.g. `content_components.meta_description`, or a
   `public_summary` in `input_schema`) and bind *that* at line 341, falling back
   to a composed sentence in the guide-page style — never to `description`.
   This makes "the brief is the SEO text" unrepresentable rather than merely
   unlikely. Prefer it.
2. **Compose at the call site**, mirroring line ~456 exactly: bind a
   `fmt.Sprintf` over `toolDisplayName`/`toolCategory` instead of
   `toolDescription.String`. Cheapest correct change; leaves the field free for
   a future editor to reintroduce the brief by hand.
3. **Backfill the 16 live rows** — required by §3 whichever of (1)/(2) lands.
   Compose from `display_name` + category; do NOT truncate the existing brief,
   which would leave a clipped spec rather than a description.
4. **Add a length/shape guard** where the page is written (>320 chars, or
   matching the census regex, is never a real meta description). This is what
   turns the class off for whatever writes a page next, not just this action.

Any of these is a change to `platform/`, so it goes through the council gate per
CLAUDE.md, and it is inert until a chassis image roll.

## 5. How to verify a fix

The green path proves nothing here — a *new* tool deploy is the failing branch.

1. Deploy a tool to a scratch site and assert its `pages.meta_description` is
   **not** equal to the source `content_components.description`:
   ```sql
   SELECT p.meta_description = cc.description AS still_broken
   FROM pages p JOIN content_components cc ON cc.id = '<tool_component_id>'
   WHERE p.site_id='<scratch>' AND p.name='<tool page>';
   -- must be f
   ```
2. Re-run the §1 census — must return `0 | 0` **after** the backfill, and the
   backfill must be verified on the served page, not just the row:
   `curl -s https://vonc.com/tools/arena/index.html | grep -o '<meta name="description"[^>]*>'`
3. Negative control: confirm the census regex still matches a deliberately
   bad row, or a `0` result is indistinguishable from a broken query.

## 6. Note for whoever takes this

Found while assessing what the v1.0.1172 roll changed for vonc.com. It is
**not** the Arena's main problem — that page also serves invented users with
handles (`@synthetix`, `@inkblot_vera`) and invented vote tallies from a
hardcoded `FLOOR_TAKES` array, with take-submission writing only to
`localStorage`. That is the `gauntlet_dead_cta` workstream's business and is
tracked in its PLAN, not here. This bug is only the meta-description leak, which
is fleet-wide and independent of it.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — the worst row was hand-fixed; the other 15 and the cause are untouched

Census re-run exactly as §1 instructs (do not trust the old table):

```
 pages | sites
-------+-------
    15 |     5
```

**16 → 15 / 6 → 5.** The single row that left is the vonc.com Arena — repaired by hand,
not by code. It now carries 147 characters of real copy and serves it:

```
curl -s https://vonc.com/tools/arena/index.html | grep -o '<meta name="description"[^>]*>'
-- content="Read today's provocation, browse the archive, then take a position into the
--          Gauntlet and defend it against an AI opponent on a twenty-minute clock."
```

The remaining 15 are the same rows this file listed, now topping out at 637 chars
(`leopardessconsulting.co.uk/tool-process-automation-scorer`), across leopardess,
ai-agent-orchestration, finetuning, fundamentallyai and gaswholesalers. Several duplicate
across sites (`llm-cost-calculator` × 4, `ai-agent-roi-estimator` × 3) — one component's
brief, published on every site that deployed it.

**The cause is exactly as filed.** `deploy_tool_action.go:341` still binds
`toolDescription.String` into `meta_description`, the `ON CONFLICT … DO UPDATE SET` at
`:333-337` still omits the column, and the correct composed pattern still sits ~110 lines
below at `:455`. No commits have touched the file. Confirmed against the running image:
there are no Go commits after `e96d42226`, which is in `v1.0.1174`.

**Quick-win assessment (this file was flagged for one).** It is *not* a sub-hour job, and
the reason is §3: candidate (2) is a two-line `fmt.Sprintf` swap, but shipping it alone
changes **nothing on the web** — the conflict clause means no redeploy repairs a live row.
The honest shape is: 15-row backfill (SQL, live immediately, ~20 min, fixes what the public
sees today) **plus** the code change (council gate + chassis roll, hours, stops the 16th).
The backfill is the quick win and it can be done first and independently. Compose from
`display_name` + category as candidate 3 says — do not truncate the brief.
