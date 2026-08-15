# HANDOFF 2026-08-15 — 268 closed; you are the CTA-resolution rollout session

Supersedes `HANDOFF_2026-08-14_canary_and_repair.md` (everything it asked for
is DONE — see `bugs_closed/268` §12 and NOTES 2026-08-14 night). Read order:
this file → NOTES 2026-08-15 session 3 → `bugs_closed/268` §12.

## 0. State, with evidence

- **268 is CLOSED**: fix `8f899cc8d` live (v1.0.1300, stamp `a2a691213`,
  probe-verified both pods, `git merge-base --is-ancestor` passes);
  canary-proven (`carried_fields` on plan items); 10/10 ever-held rows
  restored + re-rendered + live-verified; permanence proven (second rewrite
  kept the restored keys). Census 194/21 with the ever-held bucket at ZERO.
- **OWNER RULINGS 2026-08-15 (in chat, this lane's session 3):**
  (1) webdesign.uk's 8 emergency locks OFF — **DONE**
  (`ai_site_selling_automation/SQL_2026-08-15_unlock_cta_components.sql`,
  verify passed: 0 hero/cta locks, sibling chat-input-box lock untouched);
  (2) **re-run CTA resolution per site** for the ~194 never-resolved
  label-without-URL rows — **DISPATCHED IN FULL, your job is verification
  and the wrap-up.** Canary site dartsonline: 7/7 complete, **11/11 rows
  resolved, untouched rows byte-identical**; three bounded anomalies
  contributed into `bugs_open/248` (self-link on brands-index; duplicate
  target on barrel-weight; empty-label row stays invisible). Fleet batch:
  **119 items / 20 sites filed** (`ctaresolve_268_%`, backdated 08-11
  11:20), in queue as of ~09:5xZ. **248 clobber exposure measured ZERO
  pre-flight** (no dispatched page stores a CTA at a valid excluded-area
  page); full before-snapshot of all 110 url keys in this session's
  scratchpad (`canary/before_fleet_all_cta_urls.txt`) — recreate from
  `page_component_history` if the scratchpad is gone.
  **Remaining:** wait for the batch (expect some `failed` with fine work —
  `bugs_open/274`'s delivery defect); per-site matched-pair spot-checks;
  re-run the census + split; append a dated addendum to `bugs_closed/268`
  §12 with the final figures and the label-only residue list for the owner.

## 1. The mechanism (verified at the code — do not re-derive)

`page_rerender` item with `spec.reason='cta_links_stale'` →
`rerender_page_sections_action.go` CTA recompute arm (`:436-457`) →
`applyCTARecompute` (`:702-736`):
- existing label matching a REAL candidate page → that target wins (203's
  misdirection repair — some existing hrefs may legitimately CHANGE);
- valid, non-excluded, non-self stored target → kept;
- absent/invalid target → the site's positional hub target
  (`chooseCTATargets`: interactive pages first, then content hubs), when one
  exists; else the row is left as stored (a site with no hubs gains nothing —
  honest outcome, not a failure).
`section_data_resolved` does NOT do this — only `cta_links_stale` does.

## 2. The rollout recipe (canary site already dispatched)

Population query + per-site counts: NOTES 2026-08-15. 194 rows / 123 pages /
21 sites. Item shape (working example — dartsonline items
`ctaresolve_268_dartsonline.com_%`, filed this session):

```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary,
  spec, page_id, priority, handler_agent, status, created_by, item_key,
  pipeline, triaged_at, created_at)
SELECT DISTINCT ON (p.id) p.site_id, 'bugfix_268_cta_buttons_fleet',
  'page_rerender', 'high', '<summary>',
  jsonb_build_object('reason','cta_links_stale','page_name',p.name,
                     'page_id',p.id::text,'domain',s.domain),
  p.id, 20, 'page-rerender', 'triaged', 'bugfix_268_cta_buttons_fleet',
  'ctaresolve_268_' || s.domain || '_' || p.name, 'build', now(),
  '2026-08-11 11:15:00+00'          -- see the backdate caveat below
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<site>' AND p.status='active'
  AND pc.slot_name IN ('hero','call-to-action')
  AND (pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta')
  AND NOT (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')
ON CONFLICT DO NOTHING RETURNING item_key;
```

- **⚠ The backdated `created_at` is a queue-jump on our own rows** —
  dispatch selects the site holding the fleet's OLDEST eligible item
  (migration 284, cross-site priority unimplemented); a fresh item waits
  hours behind the standing backlog. The timestamps are SYNTHETIC; never
  read them as filing dates. If the backlog has drained, drop the override.
- Verify per site as a matched pair (invariant diff, RUNBOOK): before/after
  hrefs on hero/cta rows. Expect three outcomes per row: gained a target /
  legitimately retargeted (label-match) / left as stored (no valid hub).
- **A `failed` item is not failed work** — two `deploy_page` result-delivery
  failures already observed (recorded in `bugs_open/217`); check the row
  and the live page before believing a failure.
- After all sites: re-run the census + split (RUNBOOK), append the figures
  to `bugs_closed/268` §12 as a dated addendum, and report the residue to
  the owner (rows no resolution could fill = sites needing content/hubs, or
  accept label-only per row).

## 3. Falsifiers (check before trusting this file)

- `git log` on this directory + `bugs_closed/268`; `who-owns.py 268`;
  live-transcript grep (a hit can be a LANDMINES banner — read contexts).
- Chassis stamp per SERVICE (probe `/proc/1/exe` with candidates + controls;
  the startup provenance line scrolls in minutes).
- The dartsonline canary items (`ctaresolve_268_dartsonline.com_%`) may
  already be terminal — read their outcome AND the artefact before
  dispatching the other 20 sites; if the recompute misbehaved (mass
  retargeting of VALID links, prose loss), STOP and re-read
  `applyCTARecompute` — the keep-if-valid branch is the safety you assumed.
- webdesign.uk locks: expect **0** on hero/call-to-action now (the RUNBOOK
  query's old expectation of 8 is retired; chat-input-box still 1).
