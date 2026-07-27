# 116 — the link-integrity audit has never run, on any site

**Filed 2026-07-27** (webdesign_couk thread). **Status: OPEN, unowned.**
Class: silent coverage failure. Nothing errors, nothing alerts, and the checks
themselves are correct — they simply never execute, so the platform reports a
clean bill of health it never took.

Distinct from `bugs_open/071` (the gate detects and then discards) and
`bugs_open/092` (the writer never receives its constraints), both of which are
about **detection logic**. This is about **whether the detector runs at all**, and
the answer is no.

## Observed

The owner found that no link on `webdesign.co.uk`'s home page worked — 10 of 13
hrefs returning 404 on a site that had been live and declared "98 pages, all 200"
since 2026-07-26. His question was the right one: *how did this get to be live
without being checked or flagged?*

Three checks exist for exactly this and all three are well written:
`phantom_internal_links`, `dead_controls`, `misdirected_cta`.

**They are enabled on exactly one agent, and that agent has never run.**

```sql
SELECT type,
       default_config::text ~ 'phantom_internal_links' AS has_phantom,
       default_config::text ~ 'dead_controls'          AS has_dead_controls,
       default_config::text ~ 'misdirected_cta'        AS has_cta
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type ILIKE '%discovery%';

 completeness-discovery-agent | t | t | t
 design-discovery-agent       | f | f | f
 quality-discovery-agent      | f | f | f
```

```sql
SELECT initial_request_data->'config'->>'agent_type' AS agent,
       count(*), max(created_at)
FROM orchestration_states
WHERE initial_request_data->'config'->>'agent_type' ILIKE '%discovery%'
GROUP BY 1;

 design-discovery-agent | 8 | 2026-07-27 21:31:02   <- the ONLY discovery agent ever to run
```

`orchestration_states` retains **13 days** (per `bugs_open/079`'s finding — not
24h), so this is "not once in the retained window", fleet-wide, across every site.

**webdesign.co.uk was checked** — at 2026-07-26 21:31, by `design-discovery-agent`,
which carries none of the link checks. It produced `undeployed_asset`,
`needs_rerender`, `evaluate_tools` and `capability_gap` work items, and no link
finding, because it never looked.

## Mechanism — two invocation paths, both closed

`completeness-discovery-agent` is reachable from exactly two places:

1. **`improvement-loop`** (the only `agent_definitions` row referencing it).
   **The owner confirms the improvement loop is switched off.**
2. **A scheduled task** — and there is one, which is worse than none:

```sql
SELECT name, enabled, target_agent_type, last_triggered_at
FROM scheduled_tasks WHERE target_agent_type ILIKE '%discovery%';

 oneshot-discovery-aao-20260726 | f | completeness-discovery-agent | 2026-07-26 13:39:48
 protocol-tracker-discovery     | t | adoption-researcher          | ...
 adoption-tracker-discovery     | t | adoption-researcher          | ...
 model-directory-discovery      | t | directory-researcher         | ...
```

`oneshot-discovery-aao-20260726` is **disabled**, a **one-shot**, and scoped to a
**different site** (aao = ai-agent-orchestration.com). There is **no recurring,
fleet-wide task that runs the link checks at all.**

So coverage is: whoever remembers to hand-create a one-shot for a specific site.
Every other site — including every site declared live — has never had its links
audited.

## Why it is dangerous out of proportion to its cause

The checks passing and the checks not running are **indistinguishable from
outside**. A site with no link findings looks identical whether it is clean or
unexamined, and the reasonable reading of "no findings" is the flattering one.
`webdesign.co.uk` was described as live with all pages returning 200 — true, and
irrelevant, because nobody had asked whether the pages *link* to each other.

It also silently devalues the work in `071` and `092`: a detector improved but
never scheduled produces exactly the same outcome as no detector.

## Fix candidates, ordered by what closes the door

1. **Run the checks on every build or change, not on a sweep** (the owner's own
   steer, 2026-07-27: *"whilst the improvement loop will return the checkers
   should run after every build or change I think"*). This makes coverage a
   property of the pipeline rather than of someone's memory, and it is the only
   candidate under which "no findings" means something. The deploy gate
   (`validate_page_content`) already runs per page write and shares the same
   datahelpers, so the seam exists; the gap is data-driven components and
   site components.
2. **Add the three checks to a discovery agent that actually runs.**
   `design-discovery-agent` is the only one with a track record. One config edit,
   restores coverage today, independent of the improvement loop. Cheap, and
   strictly better than the present state — but it still relies on discovery being
   dispatched.
3. **A recurring fleet-wide scheduled task** targeting
   `completeness-discovery-agent`. Cleaner separation than (2), but adds dispatch
   load, and a periodic sweep still cannot protect a build — it only reports after
   the fact.
4. **Re-enable the improvement loop.** Restores the original design, but the loop
   is off deliberately, and this bug should not be the reason to turn it back on.

(1) is the durable answer; (2) is what would make the fleet safer this afternoon.

## How to verify a fix

Do **not** verify by observing zero findings — that is the failure mode itself.
Induce the fault:

1. Point a link on a test page at a URL with no `pages` row.
2. Run whatever path the fix installs.
3. Require a `phantom_internal_links` work item to appear **for that page**.
4. Then remove it and require the finding to clear.

A green run with no seeded fault proves the agent executed, not that it detects.

## Related

- `bugs_open/071` — the gate detects every broken link then discards the finding.
  **Also carries, as of 2026-07-27, a new normalisation gap:** on sites whose pages
  are `dir/index.html`, `NormalizePagePath` strips `index.html` so `/tools`
  falsely matches `/tools/index.html` while returning 404 live. That is the one
  link of the ten on this home page that the audit would have passed even if it
  had run. Owned by two active workstreams — contribute, do not compete.
- `bugs_open/092` — the writer never receives its link constraints (the generation
  half). Owned by `bugfix_079_phantom_link_gate`, active.
- `bugs_open/049`, `083` — CTA/link integrity lineage.
- The immediate site repair is `docs/agent_docs/docs024_key_docs_latest/webdesign_couk/SQL_p10_fix_dead_homepage_links.sql`
  plus `gqls/sites` commit `b0dbe8358`. **That repair is an artefact, not a
  property** — it expires the next time the page is generated, because nothing
  upstream changed.
