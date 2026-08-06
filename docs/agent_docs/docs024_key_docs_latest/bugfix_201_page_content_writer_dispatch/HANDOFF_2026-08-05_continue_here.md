# HANDOFF — `bugs_open/201` lane, 2026-08-05 21:00Z · read this first

**Fix-1 is committed (`37afbb847`), council APPROVED, and NOW DEPLOYED** on `v1.0.1254`
(both replicas, started 20:40:42Z / 20:41:08Z). **201 stays OPEN** — the fix is live but
**unproven**, and symptom 2 is untouched.

Do not re-derive anything below. Lane files: `PLAN` (the decision + three rejected shapes,
with a visible correction) · `RUNBOOK` (R1–R7b, the verification traps) · `NOTES` (evidence +
every misstep) · `README_where_we_are` (owner's plain prose).

---

## 1. What the bug is, in four lines

Three discovery checks named `page-content-writer` as their `HandlerAgent` and dispatched it
**directly**. The writer self-plans from `input_data.current_page.sections`; a discovery spec
has **no `sections` key** (measured: all 14 such items ever). So `plan_sections` early-returns
`ready_count: 0, reason: "no sections to plan"` (`plan_sections_action.go:867-875`) and the run
dies at `fail_no_ready_sections` — **11 of 11**, before writing anything.
**The cause is "no sections supplied", not "sections not ready".** The error text hides that.

## 2. What shipped

`check_literal_markdown.go`, `check_placeholder_contact.go`, `check_component_standards.go`
re-pointed to `page-build-handler`, which sources sections from `site_specs.site_plan` via
`load_page_sections_from_spec` and so cannot be broken by the caller's spec shape. Plus a stale
note in `verifier_coverage_test.go`. Completes a migration `check_empty_sections.go` and
`save_sections_claims_guard.go` already made.

Council **APPROVED** — `71523705-07d1-4067-9c5d-af371ba84b89`, *"approved with 5 advisory
objection(s) — none high-severity"*, 15 reviewers, 2 abstained. All five acted on; see NOTES.
The commit carries `Council-Submitted:`, which `098` credits automatically now it is approved —
**do not back-date a `Council-Reviewed:` trailer onto a later commit.**

## 3. ⚠ THE VERIFICATION — and the trap I already walked into for you

**A POD-GREP CANNOT PROVE THIS CHANGE. Do not attempt one.** RUNBOOK **R7** carries the full
explanation; the short form: the edit swaps one *pre-existing* string literal for another
(`"page-content-writer"` → `"page-build-handler"`), both of which were in the binary before and
after. It adds no string and removes none, so there is no control to construct. My first draft
of R7 told you to grep a **Go comment**; run against `v1.0.1254` it returned **0** and would
have read as "the fix did not ship". It had shipped. Corrected in place.

**The check that does work — R7b, newly-filed items only:**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT wi.item_type, wi.handler_agent, count(*), min(wi.created_at), max(wi.created_at)
FROM site_work_items wi
WHERE wi.item_type IN ('literal_markdown','placeholder_contact')
  AND wi.created_at > '2026-08-05 20:41:08+00'
GROUP BY 1,2 ORDER BY 3 DESC;"
```

- **PASS:** rows exist with `handler_agent = 'page-build-handler'`.
- **FAIL:** rows exist with `page-content-writer`.
- **NOT EVIDENCE:** zero rows — the checks have not fired. **Measured 20:48Z: zero rows.**

**When it will fire.** Both checks live on `quality-discovery-agent` (checks array:
`broken_nav_links, placeholder_contact, generic_theme, unverified_claims, voice_tells,
literal_markdown`). **22 runs all-time, last 2026-08-05 12:14Z.**

> ### ⚠ CORRECTED 2026-08-06 10:00Z — I wrote "the proof arrives on its next run". THERE IS NO NEXT RUN.
>
> `quality-discovery-agent` **is not on any schedule.** Its only `scheduled_tasks` row is
> `oneshot-quality-discovery-rh-20260730`, **`enabled = false`**, a one-shot from 07-30. And it
> is not an isolated case: **`SELECT count(*) FROM scheduled_tasks WHERE target_agent_type ILIKE
> '%discovery%' AND enabled` returns 0** — *no* discovery agent runs on a schedule, fleet-wide.
> All 22 runs were manual dispatches.
>
> **So R7b's zero rows will stay zero indefinitely, and waiting is not a plan.** Proving 201
> requires a *deliberate* dispatch. This is the estate's recurring shape — detection works,
> schedule and dispatch do not — and I reproduced it in my own handoff by inferring a cadence
> from a run count instead of reading `scheduled_tasks`. **Read `enabled` + `last_triggered_at`;
> a run count tells you an agent CAN run, never that anything WILL run it.**

**To get the proof, someone must fire a sweep** at a site with a live `literal_markdown` or
`placeholder_contact` instance. Trigger templates exist (`scripts/initial_messages/…
075_trigger_discovery.sh`, `290_design_discovery/082_fire_design_discovery_any_site.sh`) — read
before running; this lane has already found one committed trigger script that cannot execute.

⚠ **This is not a free action and should be an owner's call, not a verification convenience:**
- Known live instances are all **live customer sites** (`bugs_open/184`):
  `mortgagecalculator.co.uk` hero — **but that site is LOCKED**, see R6 trap 2 —
  `gaswholesalers.com` pricing, `webdesign.co.uk` news-listing.
- A discovery sweep **detects**; it does not repair. Items are filed at `status='detected'`, and
  `load_work_items` only loads `triaged`/`approved`, so a filed item does not immediately
  rebuild anything. **Verify that for yourself before dispatching** rather than taking this
  sentence on trust — if something does pick it up, the repair **regenerates the section and
  loses its prose** (§4).
- Some checks in that array (`unverified_claims`, `voice_tells`) are LLM-backed, so a sweep
  costs real spend on whatever site it targets.

**The other three traps are in RUNBOOK R6** and each is a false result waiting to happen:
1. the 14 existing rows still carry the **old** `handler_agent` — a re-arm must set it too;
2. `mortgagecalculator.co.uk` is **LOCKED**, and `load_work_items` returns *success with zero
   items* (`skipped_reason: site_locked`), indistinguishable from idle;
3. `complete` is not proof while symptom 2 stands — require `content_data` to change.

## 4. ⚠ WHAT THE REPAIR ACTUALLY DOES — expect prose loss; it is not the fix failing

`LANDMINES.md:4433` (root cause confirmed on `bugs_open/178`): `page-build-handler`'s writer
**never sees the page's own stored prose** unless `spec.mode="recreate"`, which none of these
checks sets. `load_existing_content_action.go:64-69` no-ops (`reason: "not_recreate"`) and
`load_page_record` carries only sections/title/page_type. **So the affected section is
rewritten from scratch and its prior prose is lost.**

**Do NOT "fix" that by setting `mode=recreate`** — that gate sources `research_results`, the
original adoption-crawl snapshot, i.e. *stale* content rather than none. There is today no
channel passing a page's LIVE stored section content to its own writer.

The decision stands only because the alternative is a permanent defect (11/11 hard fails; per
`184` the markdown reprints on every rerender) — **not** because regeneration is the right
repair. For `needs_content_page` (no sections at all) from scratch is exactly correct.

## 5. What is still OWED

1. **R7b after the next `quality-discovery-agent` run** — the only real proof. Zero rows is not
   a pass.
2. **Symptom 2** — `mark_complete` trusts `handler_result` blindly; an item reached `complete`
   having written nothing. Untouched **by 201 §2's own instruction** (fixing it first would make
   the broken route look repaired). This is now the next piece of work in this lane.
3. **`bug_historian`'s open objection, NOT closed:** the evidence for `page-build-handler` is
   *by analogy* (different item types succeeding — `content_rewrite` 19, `empty_section` 12),
   not these item types, and this trades a loud fail for a pipeline with filed history of silent
   partial success (016b §9). Hence trap 3 above.
4. **`check_component_standards`' edit is unverifiable from runtime data** — all 77
   `needs_content_page` items fleet-wide already carry `page-build-handler` (from
   `write_build_items`); that sub-check appears never to have fired. Say so; do not claim it
   passed. The `editquality` seat made this point.
5. **`RFC_014`** (filed) — the structural question: the guard checks an agent NAME exists, never
   that it can CONSUME the filed spec shape. Fifth recurrence. Owner call, three costed options.

## 6. Things that will mislead you

- **`who-owns.py 201` names `bugfix_194`.** That is an artefact of this lane writing 12 mentions
  of 201 into 194's files. The filer is the `bugfix_091`/`184` lane; the fix owner is this lane.
- **Two other sessions were live in `page-content-writer` code** on 08-05 (the `156` dedup lane
  in `save_page_sections_action.go`, and the concept-register drift lane). Neither mentions 201.
  `git log` alone will not show you a session mid-fix — grep live `.jsonl` transcripts.
- **A top-level `#>> '{workflow,steps,…}'` returns EMPTY** for `build-dispatch-loop`'s
  `claim_item` / `call_handler` — they are nested in a loop `sub_workflow`. Use
  `jsonb_path_query(…, '$.**.steps')`. An empty result reads exactly like "no such step".
- **Gate 2 is already checked, do not redo it.** `call_handler` forwards `domain` + `site_id`,
  both of `page-build-handler`'s required contract fields, item-type-agnostically;
  `spawn_handler` reads `agent_type_field: current_item.handler_agent` from the ROW.
  `scripts/audit-relay-gaps.sh`: 175 agents, **0 findings**.
