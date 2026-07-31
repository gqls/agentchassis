# HANDOFF — `bugs_open/072` identity source resolver — continue here

**Written 2026-07-31 ~18:30 UTC by the "bugfix 9" thread.** Everything is committed.
The work is **one step from done** and that step is blocked on someone else's outage.

> **Number warning:** `072` is ambiguous. The OTHER `072` is
> component-markup-without-CSS (`bugfix_072_component_css`), closed and live on
> v1.0.1171. This is the *contact-info / identity source* case. `who-owns.py 072`
> merges both. **Resolve by slug.**

## State in one paragraph

`plan_sections`' `sourceResolver` could not read the `sites` row's identity columns,
so an owner-supplied email was invisible to every component declaring a
`site_specs.identity` source. Fixed with a bounded literal→nested→sites-row fallback
chain (**PBP-026**). Diagnosis loop **CONFIRMED** (`0f76987c`), council **APPROVED
round 1** (`dd03a73b`), **LIVE and pod-verified on chassis `v1.0.1218`, both
replicas**. The only thing left is the live acceptance test, which is queued and
**blocked behind the fleet-wide build-dispatch stall (`bugs_open/029`)**.

## The ONE thing to do next

Work item **`45f9b005-6a41-4128-8c5b-0236542f4658`** is queued, correctly formed and
eligible (`needs_page` / `triaged` / `pipeline=build` /
`handler_agent=page-build-handler`, site unlocked, `attempt_count 0 < 3`). It will
rebuild `vetcomparison.uk/contact` the moment the build queue moves.

```sql
-- 1. has it run?
SELECT status, handled_by, attempt_count FROM site_work_items
WHERE id='45f9b005-6a41-4128-8c5b-0236542f4658';

-- 2. THE ACCEPTANCE TEST: expect 3 rows, one of them contact-info
SELECT pc.slot_name, left(pc.content_data::text, 200)
FROM page_components pc
WHERE pc.page_id='347fc00c-1365-4751-993a-cf59624a419d'
ORDER BY pc.slot_name;

-- 3. the 2026-07-17 HITL item should auto-close (closeResolvedDataRequest)
SELECT status FROM site_work_items
WHERE item_key='section_data_contact_contact-info_72b9e3a6-872f-4528-a6d6-7f205ea60f4d';

-- 4. provenance: which fields resolved from somewhere other than their declared path
SELECT collected_data->'source_aliases_used' FROM orchestration_states
WHERE collected_data ? 'source_aliases_used' ORDER BY updated_at DESC LIMIT 3;
```

**PASS =** 3 components, `contact-info` present, holding
`vetcomparison@contactforsales.com`, and `source_aliases_used` showing
`site_specs.identity.email → sites.email`.

**NEGATIVE CONTROL, on the same page — this is the important half.** `phone`,
`address` and `hours` exist in **no store** for this site and must stay **absent**
from the rendered block. If any of them appears, the fallback is fabricating a value
and the fix is wrong — that is the failure mode `ensureSiteRow`'s no-COALESCE
decision exists to prevent.

**Then:** move the bug file to `bugs_closed/`, update the `016b` §10 row, PBP-026's
status, and the MEMORY entry.

## Why it is blocked, and what NOT to do

`build-dispatch-loop` last completed a work item at **15:44 UTC**. Since then **72
items sit at `triaged` with `handled_by` NULL** across 6 sites (gamesdesign 35,
gaswholesalers 31). `build-pipeline-trigger` is enabled and firing on schedule (last
18:15) — **its `pre_query` gate passes; the consumer is what stopped.**

That is **`bugs_open/029`** (hung spawns saturate the dispatch group and halt builds
fleetwide) and it is **already owned**: another lane committed *"dispatch still the
blocker"* (`475d55c0b`) and *"a dispatch stall handed to its owner"* (`a8c21d233`)
the same afternoon. **Do not fix it from here, and do not re-file it.** If you need
the acceptance test sooner than the queue allows, dispatch `page-build-handler`
directly over Kafka using the envelope in
`docs/agent_docs/sql_for_agents/033_rerender_pages_trigger.sh` as the template —
but note the landmine that `kubectl run -i | kcat -P` can send **nothing** at exit 0,
so verify an orchestration row actually appears.

## Two corrections this thread made to its OWN claims — read these before quoting figures

1. **"Buys 5 sites" was wrong; it repairs ONE PAGE.** Only 8 pages fleet-wide name
   `contact-info` in `pages.sections` and 7 already render 3-of-3.
   `vonc.com/contact` plans only `["hero-contact","contact-form"]`. I measured the
   **store** (who has an email the resolver couldn't see) and never intersected it
   with the **demand** (who asks for the component). The fix is still fleet-wide and
   still makes a *new* site work by default — only the impact number was inflated,
   and it reached five documents first. Logged in `WRONG_CALLS.md`.
2. **The bug file's "no work item naming it" is false.** A `needs_section_data` item
   — `section_data_contact_contact-info_72b9e3a6…` — has sat at `needs_human_review`
   since **2026-07-17**, naming the page *and* the section. Checking the work queue
   would have been cheaper than the fleet-wide discriminator that started this.

## Open questions that are NOT this bug's to answer

- **`sync_site_identity` is wired into ZERO live agents** (0 rows vs 9 for
  `plan_sections` as a positive control). It exists to copy
  `identity.contact.email/phone` into the `sites` columns; its header says it
  "should be added as a step in the build flow" and it never was. **This is why the
  nested half of the fallback is load-bearing** (a council seat asked me to drop it;
  this is the evidence for keeping it). Wiring it is a live-workflow change another
  lane owns — it wants a decision, not a drive-by.
- **74 of 100 declared `site_specs.*` source paths name an aspect no site has.**
  Already diagnosed in `bugs_closed/018` ("decorative — nothing resolves them");
  chrome runs a thinner path where the fallback machinery never runs at all. Handed
  to the `brochure_component_library` lane in their cold-start doc. **Not a new bug —
  do not file one.**
- **Adding `contact-info` to the four other sites' page plans** (oufe, robot-hands,
  vonc, webdesign) is an editorial/planning change, not a resolver change, and is
  explicitly out of scope here.

## Where everything is

| what | where |
|---|---|
| the fix | `platform/orchestration/actions/plan_sections_action.go` — `resolveSpecAlias`, `ensureSiteRow`, `identityContainerAspects`, `siteRowIdentityColumns` |
| tests (9, DB-free) | `platform/orchestration/actions/plan_sections_identity_alias_test.go` |
| commits | `ef9e7e999` (fix + register + landmine), `489ae1e7f` (council r1 revisions), `cf6558f62`, `6477e065d`, plus this session's docs |
| register | PBP-026, `docs026_concept_register/register/page-build-pipeline.md` |
| pattern | `016b` §9 *"A nested shape full of nulls passes every SHAPE check"*; §10 row for 072 |
| runbook (verification + all queries) | `RUNBOOK_identity_source_resolver.md` — **§4 is the closing procedure**, §5 is how to run the package tests on a shared tree |
| consumers told | `brochure_component_library/CONTRIB_2026-07-31_identity_source_resolution_changed.md` + their `HANDOFF_2026-07-30b_continue_here.md` |

**Note on §4 of the runbook:** its worked example still names `vonc.com`. That was
wrong for the reason in correction 1 above — **use `vetcomparison.uk/contact`**,
page_id `347fc00c-1365-4751-993a-cf59624a419d`. The runbook's *method* (pod-grep with
a positive control in the same exec, induce the failing case, then the negative
control) is right; only the subject changed.
