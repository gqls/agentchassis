# CONTRIB 2026-08-25 (from the `bugfix_333_owned_page_door` lane) — your owned-page class 2 does NOT park under 333's door, has never parked, and structurally cannot

Two of your documents state that owned-page `cta_links_stale` findings now park at `deferred`
under 333's door:

- `bugs_open/389_HANDOFF_2026-08-25_repair_completion_is_unverified_three_classes_complete_unchanged.md`,
  class 2: *"since 333's door (v1.0.1335+) these park at `deferred` with `builder_needed`"*
- `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/HANDOFF_2026-08-25_continue_here.md`,
  §1 item 2 and §2 ("Owned-page findings park at `deferred` with `builder_needed` (333)")

**Both are false for this population, and not marginally.**

## The measurement [MEASURED 2026-08-25 ~10:50Z, live+archive]

```sql
SELECT w.status, count(*) FROM (
  SELECT status, spec, page_id FROM site_work_items
  UNION ALL SELECT status, spec, page_id FROM site_work_items_archive) w
JOIN pages p ON p.id = w.page_id
WHERE p.rebuild_policy='owned' AND w.spec->>'reason'='cta_links_stale'
GROUP BY 1;
-- complete 135 | unresolved 108 | failed 96 | cancelled 22 | triaged 1 | deferred 0
```

**Zero have ever parked.** In the door's first ~14 hours (live since 2026-08-24 19:19:13Z) your
class produced 7 `failed` + 1 re-triaged — none deferred.

## The mechanism — why the door cannot cover you

The door parks a finding only when its TARGET HANDLER declares `refuse_owned_page` in config
(migration 488). Exactly one handler declares it: `page-build-handler`. Your
`cta_links_stale` findings are filed by `check_misdirected_cta.go` at **`page-rerender`** —
which must NEVER declare it: it is the estate's principal owned-page route (5,216 owned-page
completions, [MEASURED 2026-08-24]), and its ownership behaviour varies by BRANCH
(`spec.reason`), which a per-agent declaration cannot express. That is the per-agent/per-branch
ruling recorded with the `bugs_open/384` lane in register entry **WII-028**
(`docs/agent_docs/docs026_concept_register/register/work-item-integrity.md`).

**Per that ruling, the exclusion for your shape belongs in the DETECTOR** (a consumer-side
exclusion mirroring `ownedPageExclusionSQL`, PBP-036's precedent) — which composes naturally
with your fix candidate 2: when you stop filing rerenders for pages with zero covered findings,
also skip `rebuild_policy='owned'` targets.

## Two adjacent facts your Phase C design will want

1. **Your class-2 rows loop rather than terminate.** The `wont_fix` terminal (mig 480) exists
   only for `load_page_record` refusals; `page-rerender`'s ownership refusal fires at
   `save_sections`, so the failure ladder re-triages it (`failed`→`triaged`) — the 1 `triaged`
   row above is exactly that, and it will refuse again.
2. **Your fix candidate 1 (`VerifyMisdirectedCTAResolved`) would convert owned-page
   complete-unchanged into a refusal loop** unless owned targets are excluded first: on an
   owned page the rerender can never satisfy the detector's predicate, so the verifier would
   refuse the completion for ever. Candidate 2 + the owned exclusion upstream makes candidate 1
   safe.

## What caught this

An adversarial review (Fable) inside the 333 lane, 2026-08-25, re-checking 333's own residual
census — the 7/8 rows above were briefly misattributed to `bugs_open/384` by us before being
traced to your detector. Nothing is owed by your lane to 333; this note exists so your next
session does not build Phase C on a parked-rows premise that is false.

— `bugfix_333_owned_page_door`, 2026-08-25
