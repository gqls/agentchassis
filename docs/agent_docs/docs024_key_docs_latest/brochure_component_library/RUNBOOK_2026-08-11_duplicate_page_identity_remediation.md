# RUNBOOK — remediating pages that exist twice under two spellings (bugs_open/215)

**Written 2026-08-11 by the 215 quiet-mode lane.** This covers the damage that
already exists. The code fix (PLAN-048) stops NEW twins being minted; it
deliberately refuses to resolve an existing pair, because choosing which name
owns a page is a content and SEO decision, not a reconciliation. **Every step
here is hand-run by a human or the sweep front. Do not automate it.**

## Why this cannot be a script (read before proposing one)

1. **The two sides are different builds, not copies.** [MEASURED 2026-08-11]
   robot-hands' three pairs carry 5/3/4 components on the bare side against 1
   each on the `tool-` side. Picking a survivor discards real content.
2. **The older, richer side is usually the flat/legacy one** — i.e. the one the
   framework's convention would retire. Convention and traffic point opposite ways.
3. **Archiving is not durable on its own.** [MEASURED 2026-08-11] Two rows
   hand-archived on 08-08 were rebuilt and re-deployed by the work-item pipeline
   today (`deployed_at` stamped 10:34:21 and 11:13:25) and now serve 200. Root
   cause is with the diagnosis loop, correlation
   `38099787-c7f9-46d4-b75e-3a1867fcaf41`. **Until that is fixed, assume any
   archive can be undone by the next replan-triggered build.**
4. **Archiving a row whose plan entry still stands gets it re-created.** The chain
   `site_plans → site_plan_pages → pages` regenerates. PLAN-017's cleanup removed
   the rows from the PLAN for exactly this reason.

## The population

Re-measure before acting — this is a point-in-time list, and a replan changes it.

```sql
-- Both spellings live: the pairs needing a decision.
WITH live AS (
  SELECT p.*, s.domain,
         regexp_replace(lower(p.name),'^(tool-|guide-|game-)','') AS stem,
         CASE WHEN p.url LIKE '%/index.html' THEN left(p.url, length(p.url)-11)
              ELSE regexp_replace(p.url,'\.html$','') END AS pathkey
  FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE p.status NOT IN ('deleted','archived'))
SELECT a.domain, a.name AS side_a, a.url, b.name AS side_b, b.url,
       (SELECT count(*) FROM page_components c WHERE c.page_id = a.id) AS comps_a,
       (SELECT count(*) FROM page_components c WHERE c.page_id = b.id) AS comps_b
FROM live a JOIN live b
  ON b.site_id = a.site_id AND a.name < b.name
 AND (a.stem = b.stem OR a.pathkey = b.pathkey)
WHERE a.deployed_at IS NOT NULL AND b.deployed_at IS NOT NULL
ORDER BY 1,2;
```

[MEASURED 2026-08-11] **7 pairs, 4 domains**: ai-agent-orchestration.com,
finetuning.uk, fundamentallyai.com (×2), robot-hands.com (×3). All 14 URLs
HTTP-tested 200 against a 404 control of 2697 bytes.

**Do NOT use the raw `planned` + never-deployed census as the phantom list.** It
returns ~28 rows, most of them legitimate: pool-\*.internal sites mid-first-build,
and catalogued-but-uncomposed pages (the bugs_open/050 population) that should
never go to zero. The twin-refined query above is the instrument.

## Per pair, in this order

1. **Decide the survivor — owner call.** Inputs: which side has content (component
   counts above), which side is indexed (the older flat URL usually is), which side
   the site's convention wants. Record the decision and its reason.
2. **Check which side the CURRENT plan names.**
   ```sql
   SELECT spp.name, spp.url FROM site_plan_pages spp
   JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current
   WHERE sp.site_id = '<site>' AND spp.name IN ('<side_a>','<side_b>');
   ```
   Archiving the in-plan side re-arms the refile loop until the next replan.
   **If both sides are in the plan (robot-hands, all three pairs) there is no
   loop-safe order** — fix the plan first (step 3).
3. **Remove the loser from the CURRENT plan** (`site_plan_pages` + its
   `site_plan_sections` rows), or run a replan on a site whose structure spec has
   `honour_realised_identity` set, so the converged plan stops carrying it. Skipping
   this is what makes step 5 temporary.
4. **Cancel open work items pointing at the loser**, with the reason recorded — the
   sweep front's 08-08 §2b precedent. `site_work_items.page_id` is **not**
   FK-constrained, so nothing else will notice them.
5. **Archive the loser**: `UPDATE pages SET status='archived' WHERE id = '<loser>'`.
   **Never `DELETE`** — three FKs onto `pages` are NO ACTION
   (`link_registry.target_page_id`, `redirects.source_page_id`,
   `page_component_history.page_id`) so a delete can fail rather than cascade.
6. **Retract the deployed file.** Both sides SHIPPED here, so archiving alone
   leaves the loser serving 200 for ever (bugs_closed/098's whole subject). Use
   `RetractPageDeploymentAction` (eligibility is `status <> 'active'`, so it must
   follow step 5), ALWAYS with explicit `PAGE_IDS`.
7. **Write a redirect** loser → survivor in `redirects`. [MEASURED 2026-08-11] zero
   `redirects.source_page_id` and zero `link_registry.target_page_id` rows
   reference any of the 14 twin rows today, so nothing conflicts — re-measure at
   execution.
8. **Verify at the artefact:** loser URL 404s, survivor URL 200s, and a collateral
   page still 200s (the control that proves the retraction was targeted). Then
   re-run the census in the *next* sweep to confirm the row did not come back —
   given finding 3 above, that second check is not optional.

## Coordination

The fundamentallyai sweep front owns that site's phantom cleanup and has done
this by hand before (`HANDOFF_2026-08-09_sweep_front_continue_here.md` §2b). Route
fundamentallyai pairs through its handoff rather than acting in parallel.

## Enabling the prevention switches (added 2026-08-11, from the council's round-2 questions)

All three live in the `site_specs` **`structure`** aspect, beside `url_shape`:
`honour_realised_identity`, `twin_identity_snap`, `stem_twin_snap`. All default
OFF; nothing changes for a site until its spec says otherwise.

- **Read the dark-launch evidence first.** With the gates off, the layers still
  record what they would have done:
  ```sql
  SELECT error_code, count(*), min(occurred_at), max(occurred_at)
  FROM agent_error_log
  WHERE error_code IN ('PLAN_PAGE_IDENTITY_TWIN_OBSERVED','PLAN_PAGE_STEM_TWIN_OBSERVED',
                       'PLAN_PAGE_STEM_TWIN_REFUSED','PLAN_PAGE_IDENTITY_SNAPPED')
  GROUP BY 1 ORDER BY 1;
  ```
  Each OBSERVED row is a second page identity that was about to be written. Zero
  rows before any replan has run is **not** evidence either way.
- **⚠ Re-adopting a site DROPS all three flags.** `WriteSiteSpecAction` deep-merges,
  so ordinary spec writes preserve them, but `apply_adoption_plan_action.go`
  replaces the structure spec wholesale. The failure is safe (a dropped key reads
  false = today's behaviour) but silent — **re-check the spec after any adoption run.**
- **Rollback.** Turning `honour_realised_identity` off after a pilot does not move
  live URLs; the next replan simply reverts to minting twins, i.e. the pre-fix bug.
  The kill switch costs you the defect back, not damaged pages.
- **Do NOT enable on the five decomposed sites** until `bugs_open/204` is fixed —
  their `pages.sections` hold positional slot names, which the snap carries verbatim.
