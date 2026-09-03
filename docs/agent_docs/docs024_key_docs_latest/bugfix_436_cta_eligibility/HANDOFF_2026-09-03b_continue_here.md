# HANDOFF — 2026-09-03b. **START HERE** (supersedes `HANDOFF_2026-09-03_continue_here.md`).
`bugs_open/436` — built, approved, rolled, enabled, **and now verified live except one blocked step**.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_436_cta_eligibility/` · register **LNK-041**
(`docs026_concept_register/register/link-management.md`) · council **APPROVED round 2**, corr
`9faa2a23-f3bc-464e-8c3a-9d3d44759cc0` · predecessor lane `bugfix_389_cta_relevance/`
(`bugs_closed/391` — the damage; 436 is the cause).

## 0. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 436` · **re-run the RUNBOOK's
prediction census** (a census goes stale by addition — do not quote this file's four sites) · grep
LANDMINES for any symbol you are about to trust (two new entries were added by this lane).

## 1. WHAT IS DONE

Everything in the previous handoff's §1 (column `714`, ranking + label bindings live, check registered
on the fleet, `715_HOLD` applied, council APPROVED r2) — plus, on 2026-09-03 mid-morning:

| | state | proof |
|---|---|---|
| The check RUNS, provably, per run | **YES** | `collected_data.run_checks` carries `checks_run` / `checks_unregistered` / `checks_failed`. First post-enablement pass (idea.uk 09:26:06Z): ran, 46/46, `[]`, `[]`. **Use this, not the runner's log line** |
| "0 items" explained | **rotation coverage, not a healthy fleet** | hand-mirror census `[MEASURED 2026-09-03 10:05Z]`: exactly **4** sites fossil-shaped fleet-wide |
| The check FIRES where predicted | **2 of 4 induced, both exact** | vetcomparison.uk + cv1.co.uk filed `needs_human_review` quoting the census's own pages and numbers |
| The check is SILENT where predicted, for the stated reason | **YES — the disconfirming control** | idea.uk's rank-1 is also below the default and is correctly absent (lead 7 = curated ladder) |
| Detector corroborated OFF the database | **YES** | cv1.co.uk serves `<a href="/tools/example/index.html" class="header-cta">` (target 200, invented URL 404) |
| **The lever binds the ranking, both directions, live** | **VERIFIED** | cv1.co.uk: opt out → retraction reason *"only 2 eligible interactive candidate(s)"*; opt back in → detail *"among 3 candidates"*. **3 → 2 → 3 is the assertion.** Control: vetcomparison.uk's item untouched across all four runs |
| Fleet state after the canary | **restored** | `eligible_as_cta_target=false` count fleet-wide = **0**. No data decision taken |

Full evidence, queries and the wrong prediction I made: lane `NOTES` (2026-09-03 entry). Recipes:
lane `RUNBOOK` (three new sections). Owner read-out: `README_where_we_are.md`.

## 2. ⏳ WHAT IS LEFT — one blocked verification and one owner decision

### 2a. ⛔ BLOCKED — the header button at the SERVED bytes
This is the last unproven thing, and it is **not** roll-bound any more; it is a session-permission
block. It needs a chrome re-render + redeploy:

```bash
# rerender-pages, refresh_site_components:true — LOAD-BEARING (without it the run reassembles from
# the STORED site_components.rendered_html and the header cannot move)
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},
 "input_data":{"site_id":"8c3e9118-2455-4f0d-b01a-5dcde13dcf99","domain":"cv1.co.uk",
               "refresh_site_components":true}}
```
Publish it via `scripts/kafka-publish-lib.sh` (a ready caller is in the RUNBOOK). **This session's
permission classifier refused that dispatch in every form it was offered; nothing was published and
no retry is pending.** If your session can dispatch it: opt `tool-example` out, dispatch, wait for the
deploy, and assert the served `class="header-cta"` href moved to
`/tools/job-search-readiness-checker/index.html`; then flip back, re-dispatch, assert it returns.

⚠ **cv1.co.uk is the only site that can show this** as of 2026-09-03 — the header reaches the ranking
only where no footer-group nav item is labelled `contact`, and cv1.co.uk was the only such site that
is also fossil-shaped. See the new LANDMINES entry before picking a different one.

**What is genuinely unproven is narrow:** that the header *caller* re-reads the ranking. The ranking's
response to the lever is proven (§1), the header's pick is observable on the wire, and the call shape
is unit-pinned. **Do not record 436 as verified-in-full until this is dispatched.**

### 2b. The STORED half — not blocked, just not showable by a rerender
`applyCTARecompute` KEEP #2 holds any valid stored destination (PLAN, "what this deliberately does
NOT do"), so no rerender will move a stored CTA field. It needs a full page rebuild, which regenerates
copy. The **7 of 10** stored `tool-example` destinations on cv1.co.uk are that stated limit, live and
measurable — including both *other* tools' guide pages, which send readers to the example page.
Decide whether that is worth a rebuild or whether the limit simply stands as documented.

### 2c. OWNER DECISION — and it now has real substance
The previous handoff guessed "plausibly nothing to do". That was wrong: **four live sites** currently
point their primary button at a fossil-ranked page, one of them at a page called `tool-example`.
Re-run the census, then ask. Two remedies per site, and the difference matters: `eligible_as_cta_target
=false` changes nothing but CTA candidacy; demoting `nav_order` **also moves the visible menu** where
`in_header=true`. The read-out is written for him at the foot of `README_where_we_are.md`.

### Then: CLOSE
Bar is fixed AND live AND verified. §1 satisfies "live"; 2a is the outstanding half of "verified".
On close: move the bug file to `bugs_closed/`, update `MEMORY_workstreams.md` + `MEMORY_closed.md`,
write the lane's **first** SUMMARY (the close is the milestone), and add the transferable pattern to
016b §9 only if something new surfaces — the lock-in pattern is already in 391/LNK-040/LNK-041.

## 3. Traps — the previous handoff's §3 still stands, plus three from this session
- ⛔ **Never run `scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh`** — its
  tail hard-codes finetuning.uk and mass-triages *that* site's `detected` items whatever domain you
  passed, while your own run succeeds normally. **New LANDMINES entry.**
- **The header CTA reaches the ranking only where no footer nav item is labelled `contact`** (matched
  on `site_nav_items.label`, NOT `pages.in_header`). A canary chosen on tool count proves nothing.
  **New LANDMINES entry.**
- **A retraction does not poison the item key against a genuine re-fire.** I predicted the flip-back
  would be deduped (`bugs_open/326`) and it filed a fresh item. The check's header comment describes a
  human *dismissal*, which behaves differently. Do not skip direction 2 on dedup grounds.
- `orchestration_states` has no `agent_type` column — it is **`owner_agent_type`**.

## 4. Who was told what
Unchanged from the previous handoff (399 CONTRIB, `cta_target_content_pass` CONTRIB, 114 two-way, 384
deliberately not contacted). Nothing new was routed this session.

## 5. The one-query status board
```sql
SELECT count(*) FILTER (WHERE NOT eligible_as_cta_target) AS opted_out, count(*) AS pages FROM pages;
SELECT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly'
FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
SELECT name, count(*) FROM service_binary_capabilities WHERE kind='discovery_check'
  AND name IN ('misdirected_cta','cta_rank_anomaly','no_such_check_zz') GROUP BY name;
SELECT s.domain, w.status, w.summary FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='cta_rank_anomaly' ORDER BY w.created_at DESC;
```
