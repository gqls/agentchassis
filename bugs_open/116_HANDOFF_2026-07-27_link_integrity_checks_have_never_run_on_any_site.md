# 116 — the link-integrity audit has never run, on any site

> ## STATUS 2026-08-03 (evening) — THE TITLE IS FALSE. The substance survives, and **every one of this file's four fix candidates is now owner-gated or forbidden by written policy.** Stays OPEN; it is NOT a coding task today.
>
> Taken by session "bugfix 100" as the next unowned bug, re-measured before being
> believed, and handed back deliberately unfixed. **The reason is the finding**, so
> read this block before writing any code against this file.
>
> ### 1. The checks DO run. They ran today.
>
> ```sql
> SELECT item_type, created_by, count(*), count(DISTINCT site_id) AS sites, max(created_at) AS newest
> FROM site_work_items
> WHERE item_type IN ('phantom_internal_link','dead_control','cta_names_unknown_destination')
> GROUP BY 1,2 ORDER BY 3 DESC;
> ```
> ```
>  cta_names_unknown_destination | completeness-discovery-agent | 60 | 6 | 2026-08-03 21:04:02
>  phantom_internal_link         | completeness-discovery-agent | 54 | 5 | 2026-08-03 21:04:02
>  dead_control                  | completeness-discovery-agent |  4 | 3 | 2026-08-03 21:04:07
> ```
>
> Filed **2026-08-03 21:03–21:04Z** against `gaswholesalers.com`, `vonc.com` and
> `gamesdesign.co.uk`. The audit chain that produced them is
> `improvement-loop` → `spawn_completeness_discovery` → `call_completeness_discovery`,
> gated on `audit_state.audit_due`.
>
> **Why this file said otherwise, and the trap to inherit:** the **check names are
> plural** (`phantom_internal_links`, `dead_controls`, `misdirected_cta`) and the
> **`site_work_items.item_type` values are singular and partly renamed**
> (`phantom_internal_link`, `dead_control`, and `misdirected_cta` files
> `cta_names_unknown_destination` plus a `page_rerender`). A query keyed on the
> check name returns zero rows and reads exactly like "never ran". There is **no
> mapping table** — each `ItemType` is a separate hardcoded literal inside the
> check (`check_dead_controls.go:160`, `check_misdirected_cta.go:318,:363`,
> `check_phantom_internal_links.go:315,:318,:325`).
> A different session hit the identical trap on this same file earlier today and
> recorded it (`bugfix_123_content_creator_claims/NOTES_content_creator_claims.md:34-50`).
>
> ### 2. What IS still true — the defensible claim
>
> **No enabled `scheduled_tasks` row targets any discovery agent, so audit cadence
> is whatever a human supplies.** Verified 2026-08-03: of 26 enabled tasks, none
> targets `completeness-`, `design-` or `quality-discovery-agent`, and none targets
> `improvement-loop`. Today's four audits were **hand-fired**
> (`finetuning_uk_repair/294_TRIGGER_improvement_loop_v1.sh`, and the pattern in
> `bugs_open/185:316-322`). Independently measured the same way on 2026-07-30 by
> `robot_hands_checker_gaps/NOTES_checker_gaps.md:95-96`, which had already
> corrected this file's framing: *"no enabled recurring task targets them, so
> cadence is whatever a human supplies. **Not 'the checks never run'**."*
>
> Per-site coverage is now durably recorded — `improvement-loop`'s `record_audit_pass`
> writes `sites.settings->'maintenance_profile'->'last_audit'` (migration 291). Read
> it rather than counting findings:
> ```
>  gaswholesalers.com | 2026-08-03 21:07:34     finetuning.uk | 2026-08-03 10:19:41
>  vonc.com           | 2026-08-03 21:07:29     … every other row NULL …
>  gamesdesign.co.uk  | 2026-08-03 21:07:21
> ```
> **4 of 37 site rows carry a stamp** (19 real sites once 17 `pool-*.internal` and
> `system.internal` are excluded), all four stamped today. **[UNVERIFIED] Do not
> restate this as "15 sites have never been audited"** — the key is written by a step
> introduced with migration 291, so a NULL means "not audited since the field
> existed", which is weaker than "never audited".
>
> ### 3. Why each of this file's four candidates is blocked — all four
>
> | candidate | status now |
> |---|---|
> | **1. Run the checks on every build or change** (this file's "durable answer", and the owner's own 07-27 steer) | **Forbidden today by the platform's own written policy.** The detected→triaged promoter (`TriageDetectedItemsAction`) lives *only* inside the stopped `improvement-loop`, so `detected` rows have no consumer: fleet census 2026-08-03 is **204 `detected` across 10 sites against 2 `triaged`** (`improvement-loop` register IMP-050). Adding a per-build detector now files more findings into that queue. `validate_page_content.go:644-650` already refused exactly this, in terms: *"This writes a work RECORD, not a work ITEM, and that is a deliberate choice. A `site_work_items` row would promise a repair that nothing performs."* Governing policy is **IMP-016** (`register/improvement-loop.md:130-136`): *"a discovery check should only be enabled once its handler agent actually exists — otherwise findings accumulate unconsumed."* |
> | **2. Add the three checks to `design-discovery-agent`** | **Explicitly warned against by the 149 lane.** `bugs_open/149:395-398`: double-seating interacts with `insertWorkItem` dedup on `item_key`, which is one of B2's three unexcluded candidate causes, so *"seating it before B2 is established could produce a change whose effect nobody can attribute"*. Also: every Group-B number predates the B4 roll and must be re-measured first (`149:780-782`). |
> | **3. A recurring fleet-wide scheduled task** | **Reverses an owner ruling.** This is **G1** and is an explicit separate owner go (`vigilant_designer_offer_analysis/PLAN_2026-08-02:17,144`). Migration `290` was written deliberately so that when G1 comes *"one flag flip is the whole change"*, and it *"deliberately does NOT touch `enabled`"*. |
> | **4. Re-enable the improvement loop** | **OWNER RULING 2026-07-29** (`bugs_open/136:32-35`): *"the improvement loop is stopped **DELIBERATELY** … a **decision, not a defect** — do not re-file them as dead scheduled tasks and **do not re-enable them**."* This file already said 4 should not be the reason to turn it back on; that is now a ruling, not a preference. |
>
> ### OWNER DECISIONS 2026-08-04 — the questions this file routed upward are ANSWERED
>
> - **D1: the improvement loop stays OFF for now** (owner, 2026-08-04). The 204 parked
>   findings stay parked; audit cadence stays manual. This re-affirms the 2026-07-29
>   ruling — do not re-litigate it from this file.
> - **D2: the per-build-checks steer is DEFERRED** (owner, 2026-08-04). The 2026-07-27
>   steer ("checkers should run after every build or change") is not retracted, but it
>   is not to be built while D1 stands. **Sessions should stop attempting this** — three
>   have now walked the same ground. The seam map for when it unblocks is in
>   `bugfix_116_link_check_coverage/` (PLAN §"What would change this plan").
>
> **Consequence:** this file stays OPEN as an accepted, owner-ruled state — a record
> that coverage is manual by decision, not a work item. The revisit trigger is D1
> changing, not time passing.
>
> ### 4. So what this bug actually needs
>
> **An owner decision, not a fix.** The ordering constraint is now visible and it is
> the opposite of what this file assumed: **detection cannot usefully be widened
> until the promotion gap (`bugs_open/083`) is answered**, because the platform's own
> policy forbids filing findings nothing drains. The owner-facing form of that
> question is already framed in
> `finetuning_uk_repair/SUMMARY_2026-08-03b_*.md:112-121` — the 204 parked findings
> across 10 sites, and whether the three-month disconnection gets its own answer.
>
> Until then the honest status of link-integrity coverage is the house phrase from
> `RFC_010:257-275`: **"built, approved, and undriven" — not "working".**
>
> **Diagnosis loop** filed on the mechanism rather than asserted from here:
> intake corr `aadb9c93-62af-4676-993f-b741310c2371`, run corr
> `54bf4506-5192-4528-8395-eb2c636a7fad`.
>
> **Workstream docs:** `docs024_key_docs_latest/bugfix_116_link_check_coverage/`.

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

---

## OWNER DECISION 2026-08-06 — D3: OPTION 4. Staged, supervised, per-site loop runs; full re-enable is the destination, not the starting point.

The owner has answered the question §4 routed upward ("the 204 parked findings
across 10 sites, and whether the three-month disconnection gets its own answer").
The decision, from the four options presented:

> **Option 4: run the improvement loop deliberately, one site at a time,
> supervised** — the `294_TRIGGER` pattern (`finetuning_uk_repair/`). Each run
> audits, triages and drains that one site's findings with a human watching.
> Sites with the most parked findings first. **Full fleet re-enable (option 2)
> is the destination** once a few supervised runs have shown the repairs are
> sane — it is not authorised yet; each per-site run is.

What this changes, and does not:

- **D1 stands**: the `improvement-sweep` scheduled task stays OFF. Nothing
  recurring is enabled by this decision. Per-site runs are hand-fired.
- **D2 stands**: no per-build detectors while the queue has no automated
  consumer. The revisit trigger is now concrete — after supervised runs have
  drained the parked backlog and demonstrated sane repairs, D1/D2 get
  re-asked with evidence.
- **Bulk promotion (option 3) is explicitly rejected** — triage is the step
  that decides what is worth doing, and it is not to be skipped on a queue
  where 235 items have already failed at least once.
- Executing sessions: fire ONE site per run, most-parked-findings first, watch
  the repairs land, record each run's outcome here or in a lane doc. Do not
  parallelise sites until the owner widens the authorisation.

This file stays OPEN as the record of the staged programme; it closes when
either the backlog is drained and G1 (the recurring task) gets its own owner
go, or the owner rules the manual cadence permanent.
