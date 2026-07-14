# HANDOFF — Product pages ship EMPTY, and the empty_section loop reports success without fixing them

**Created 2026-07-14. Found incidentally while wiring sprite bullets (imagery
workstream, Turn 36) — the gripper-detail page was evaluated as a candidate
surface and turned out to be a hollow shell. Not fixed; handed off deliberately.**
**Start a fresh chat from this file.** Owner: unassigned. Testbed:
robot-hands.com, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.

---

## 1. The defect, in one paragraph

robot-hands.com serves **planned, live, `status=active` / `build_status=deployed`
product pages whose product data is entirely absent**. The templates render, the
e-commerce chrome renders, and every *value* is blank. Worse: the platform
already **detected** this (an `empty_section` discovery check fired and created
work items) and the handler **marked them `complete`** — while the sections
stayed empty. So this is two bugs stacked: a **data bug** (required fields never
filled) and a **loop-integrity bug** (a fix loop that closes without fixing).

## 2. Live evidence (verified 2026-07-14)

`https://robot-hands.com/entities/gripper-detail.html` serves:
```html
<h1 class="pd-title"></h1>                 <!-- product name: empty -->
<span class="pd-price"></span>             <!-- price: empty -->
<span class="pd-meta-val"></span>          <!-- SKU: empty -->
<ul class="pd-features"><li></li><li></li><li></li><li></li></ul>   <!-- 4 empty bullets -->
```
…while still rendering **Add to Cart**, **Buy Now**, size/colour swatches, a
star-rating row and a discount badge — full e-commerce furniture on a site that
sells nothing. Same story on `/product-detail.html`. Both are **in the current
site plan** (roles `entity_page` / `content`), so they are NOT stale orphans —
they are pages the planner intends to exist. Both have `in_header=false,
in_footer=false`, so they're unlinked but publicly reachable and indexable.

**Counter-example that proves the pipeline is sound when data exists:**
dartsonline.com's `product-grid` renders **14 real product cards**. The machinery
works; it is the no-data path that fails open instead of failing safe.

## 3. Root-cause evidence gathered so far

**(a) The value fields were never generated.** For the `product-details`
instance (`page_components.id = 0cef6258-e41f-4627-aef6-4957103f36a5`),
`content_data` holds **49 keys** — but they are all *chrome*: `sku_label`,
`add_to_cart_label`, `size_option_1..4`, `rating_stars`, `main_image_url`,
`thumb_image_*`, plus site boilerplate (`domain`, `tone`, `nav_items`, `year`).
**Every value field the template needs is absent**: `product_name`,
`product_price`, `product_description`, `feature_1..4`, `product_sku`,
`product_category`, `availability_status`. The template's `{{.product_name}}`
etc. therefore resolve to empty strings — hence the hollow render.

**(b) The schema DID declare them as required.** `content_components.input_schema`
uses a **`fields` wrapper** (NOT JSON-Schema `properties` — a query against
`properties` returns nothing and will mislead you). `product-details_pre_037`
declares `feature_1` and friends as `{"type":"text","source":"llm",
"required":true,...}` **with no `on_missing` rule**.

**(c) `on_missing` enforcement is NOT the bug.** `plan_sections_action.go`
handles `skip_field | use_fallback | skip_section | needs_human_review | block`
in BOTH the optional branch (~line 1207) and the required branch (~line 1232),
and `default:` defers for safety. The machinery is sound. The open question is
why these `source: llm` fields never reached it — see §5.

**(d) A section that should have skipped instead persisted an LLM apology.**
`product-card-with-cta` (same page) has `content_data.result` containing LLM
**prose about its own inability to render**:
> "The data schema for this section requires product array data sourced from
> `query.affiliate_products`. Per the schema definition, this field is marked
> `required: true` with `on_missing: skip_section` — meaning without actual
> affiliate product data (names, prices, ratings…)"

That meta-commentary is now stored as page content. Two failures: `skip_section`
did not fire, and an LLM's refusal text was persisted as if it were content.
**A guard against meta-commentary-as-content is worth having fleet-wide.**

**(e) THE BIG ONE — the fix loop closed without fixing.** `empty_section` work
items exist for exactly these sections, handled by `page-build-handler`:

| summary | status | attempt | completed |
|---|---|---|---|
| Empty section 'product-details' on page gripper-detail | **complete** | 0 | 2026-07-10 23:55 |
| Empty section 'product-specs' on page gripper-detail | **complete** | 0 | 2026-07-10 23:54 |
| Empty section 'product-hero' on page gripper-detail | **complete** | 1 | 2026-07-10 23:59 |

They were marked `complete` on 2026-07-10 **and the sections are still empty on
2026-07-14.** The handler ran, produced a result payload, deployed, and reported
success — with the defect untouched. (Sibling items across the site sit at
`unresolved` / `needs_human_review` / `[stale: triaged 48h+]` — there is a large
`empty_section` backlog on robot-hands: ~36 items.) **This false-completion
pattern is the highest-value thing in this handoff** — it means the empty-section
loop's success signal cannot currently be trusted. Strong overlap with the
diagnosis→fix-loop workstream (`fixloop_eg_dartsonline/`): a bug that "dissolves"
but isn't fixed is exactly the graded-benchmark case.

**(f) Why the empty shell survives the assembly filter.** `getPageSections`
(`rerender_single_page_action.go:381`) drops visually-empty sections via
`sectionHasVisibleContent` (line 436) — which strips tags and keeps anything with
**>10 chars of text**. The product section is stuffed with static label text
("SKU:", "Category:", "Add to Cart", "Buy Now"), so it sails through the filter
despite having no real content. **The filter measures text, not *resolved data*.**

## 4. Scope — how fleet-wide is it?

Only two sites currently carry product components:

| site | product component instances | state |
|---|---|---|
| robot-hands.com | 6 | **broken — empty** (no product catalog exists) |
| dartsonline.com | 2 | fine (real affiliate product source) |

So today the blast radius is robot-hands. **But the mechanism is generic**: any
site whose planner places a component requiring a data source the site does not
have will ship an empty shell and (per §3e) may have that reported as fixed. As
the fleet grows this recurs. Fix the mechanism, not just robot-hands.

## 5. Open questions for the fixer (in priority order)

1. **Why did `page-build-handler` mark the empty_section items `complete`
   without filling them?** Read its result payloads (they're large, in
   `site_work_items.result`) — does it verify the section is non-empty after
   rebuilding, or does it complete on "deploy succeeded"? A fix-loop that
   completes on *deploy* rather than on *defect-gone* is the core bug.
2. **Who fills `content_data` for template components, and why were the
   `source: llm` fields skipped?** Start at `v3_site_actions.go` (it writes the
   `_sources_merged` marker present in these rows). Did these components bypass
   `plan_sections` entirely? (Their `content_data._built_at` is `2026-05-02`,
   predating the current pipeline — but the pages are still in the plan and were
   re-touched on 07-10, so "old artifact" is not a sufficient excuse.)
3. **Should the planner place product components on a site with no product
   catalog at all?** This is the upstream root cause. robot-hands is a
   spec/comparison site; `product-details` with Add-to-Cart is category-wrong for
   it regardless of data. Consider gating component selection on data-source
   availability (`content_components.data_sources` is currently empty for these).
4. **Should `sectionHasVisibleContent` measure resolved data rather than text?**
   A section whose schema-required value fields are all empty should be treated
   as empty, no matter how much static label text its template carries.

## 6. Suggested plan (not started — your call)

1. **Reproduce + trust nothing:** re-drive one `empty_section` item for
   gripper-detail (reset `status='triaged'`, **`attempt_count=0`** — see the
   re-drive lesson below) and watch what page-build-handler actually does.
2. **Close the loop-integrity hole:** make the handler's completion conditional on
   a post-fix verification (section now has resolved data), else `failed` /
   `needs_human_review`. This is the fleet-wide win.
3. **Fail safe on missing data:** honour `skip_section` for these components, and
   add a guard that refuses to persist LLM meta-commentary as content.
4. **Decide robot-hands' product pages:** either give them a real data source
   (gripper catalog → `query.*`), or remove the product components from the plan
   and delete the two pages. Do not leave Add-to-Cart furniture on a spec site.
5. **Add a discovery check:** component instance whose schema-`required` value
   fields are absent/empty in `content_data` → work item (sibling of
   `unresolved_cta` / `image_source_unsatisfiable`).

## 7. Code map

| what | where |
|---|---|
| field resolution, `on_missing`, required branch, deferred items | `platform/orchestration/actions/plan_sections_action.go` (~1109–1290, `createDeferredItems`) |
| fills `content_data` / writes `_sources_merged` | `platform/orchestration/actions/v3_site_actions.go` |
| page assembly + empty-section filter | `platform/orchestration/actions/rerender_single_page_action.go:381` (`getPageSections`), `:436` (`sectionHasVisibleContent`) |
| discovery checks incl. `empty_sections` | `platform/orchestration/actions/discovery_checks.go:72`, `discovery_checks/` |
| handler for `empty_section` items | agent `page-build-handler` (`agent_definitions`) |

**Prior art — read before designing:** `idea_uk_section_data_missing/`
(`needs_section_data` items are emitted at `needs_human_review`; `query.*`
resolution lives in `actions/queryresolve/`, and only `pages_where_type:<type>`
is actually implemented), `imagery/FUTURE_section_data_handler_1_.md` (superseded
but explains the resolution machinery), and `fixloop_eg_dartsonline/` (the
false-completion pattern belongs there as a benchmark).

## 8. Mechanisms you will need (hard-won, from the imagery workstream)

- **DB:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U
  clients_user -d clients_db` (read-only SELECT auto-approves; mutations prompt).
- **Re-driving a work item:** ALWAYS reset `attempt_count=0` alongside
  `status='triaged'` and clear claim metadata. At `attempt_count >= max_attempts`
  the item is excluded from `find_dispatchable_site` and sits `triaged` forever —
  it looks like dead dispatch but is correctly idle.
- **Zombie claims:** an item stuck `claimed` >10 min blocks its ENTIRE site from
  dispatch. Standing unstick: `UPDATE site_work_items SET status='triaged',
  claimed_by=NULL, claimed_at=NULL WHERE status='claimed' AND claimed_at < now()
  - interval '10 minutes';`
- **Page assembly reads `page_components.rendered_html` DIRECTLY** — it does NOT
  re-render from `content_data` + template. Changing `content_data` alone changes
  nothing on the live page until something rewrites `rendered_html`. **This is
  probably relevant to §5.1** — a handler could "fix" content_data and deploy,
  and the page would still serve the old empty markup.
- **Manually-inserted work items are NOT auto-triaged** — insert with
  `status='triaged'`, `triaged_at=now()`, `attempt_count=0`.

## 9. Copy-paste queries

```sql
-- The empty product component instances and their (chrome-only) data keys
SELECT s.domain, p.url, cc.function,
       (SELECT string_agg(k,', ' ORDER BY k) FROM jsonb_object_keys(pc.content_data) k) AS keys
FROM page_components pc
JOIN content_components cc ON cc.id=pc.component_id
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE cc.name ILIKE '%product%';

-- What the schema actually requires (NOTE: 'fields' wrapper, not 'properties')
SELECT cc.name, f.key, f.value->>'source', f.value->>'required', f.value->>'on_missing'
FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') f
WHERE cc.name ILIKE '%product%';

-- The false completions (§3e)
SELECT summary, status, attempt_count, completed_at
FROM site_work_items
WHERE item_type='empty_section' AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
ORDER BY created_at DESC;
```
