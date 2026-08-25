# 392 — a DB timeout while loading link context publishes pages with NO internal links, and nothing notices afterwards

**Filed** 2026-08-25 by the `bugs_open/358` lane, on the owner's ruling of 2026-08-25 (decision 4:
commission a reader). **Routed at the `bugfix_092_writer_link_constraints` lane** (quiet 14d as of
filing — `scripts/who-owns.py 092`), whose own detector is what made this visible.

**Status** OPEN. Not diagnosed via `090` — no causal theory is asserted: the mechanism below is the
writer's own recorded statement of what it did, and every figure is `[MEASURED 2026-08-25]` with the
query inline.

## 1. The defect, in plain terms

When the page-content writer starts, it loads the list of pages it may link to
(`prepare_link_context_action.go`). If that load fails, the run **degrades deliberately**: the
writer is instructed to emit **no internal links at all**, and a durable row records it
(`LINK_CONTEXT_UNAVAILABLE`, born of `bugs_closed/092`'s fix — degrade-not-fail was that lane's
deliberate design, and this file does not argue with it).

The gap is **after** the degrade: the page ships without internal links, the row that says so
expires on the retention clock, and **no mechanism retries, re-renders, or even counts**. The
degradation is permanent-until-coincidence — the page stays link-less until something unrelated
happens to rerender it.

## 2. Evidence

```sql
SELECT occurred_at, agent_type, severity, context->>'site_id' AS site,
       context->>'failure' AS failure, context->>'outcome' AS outcome, context->>'degraded' AS degraded
  FROM agent_error_log WHERE error_code='LINK_CONTEXT_UNAVAILABLE' ORDER BY occurred_at;
```

`[MEASURED 2026-08-25]` — 2 rows, both 2026-08-24 14:21Z, **two distinct sites**
(`0a538b4a-803c-4f82-b298-d916f893fe8e`, `a998349c-6a55-45d5-8558-c0e6b63d915b`), both
`page-content-writer`, both `severity=error`, `degraded=true`, failure
`page query failed: query pages: FATAL: query timeout (SQLSTATE 08P01)`, outcome
*"writer instructed to emit NO internal links"*. Two sites hit within 24 seconds — the failure is a
shared-resource event (DB load), so **bursts are the expected shape**: one bad minute can degrade
every page being written across the fleet at that moment.

Detection lineage: the code was caught **undeclared** by the `finding-code-registry-check` CronJob
on its first live run (2026-08-24, ~2h after the first row — `bugs_open/358`), which is the only
reason anyone read these rows at all.

## 3. What is asked for — a READER, not a redesign

The write side is correct and stays: degrade-not-fail, plus the durable row. Commissioned
(owner ruling 2026-08-25): **something automated that selects `LINK_CONTEXT_UNAVAILABLE` rows and
acts.** Fix candidates, ordered by what closes the door:

1. **Heal**: a consumer in the `cmd/content-loss-check` family (the estate's proven
   reader-with-writer exemplar, `bugs_open/355` A2/A3) that, per unresolved row, checks whether the
   page(s) written in that run now carry internal links, and files a `page-rerender` work item if
   not — then resolves the row (**extract first: `resolved=true` halves remaining life to 14 days**).
   The site_id is in `context`; the writer could also be extended to record the page id, which would
   make the join exact instead of site-wide.
2. **Retry at the source**: one bounded retry of the page query inside `prepare_link_context`
   before degrading. Cheap, halves the incidence, does not remove the class (the second timeout
   still degrades silently) — a mitigation alongside 1, not instead.
3. **Count**: alarm on rows/day. Weakest — it tells a human, it does not fix a page.

**Acceptance**: a new `LINK_CONTEXT_UNAVAILABLE` row leads, without human action, to the affected
page(s) carrying internal links again (candidate 1), verified at the served artefact, not at the
work item's status. Registry follow-up: flip the code's entry to `consumed` with
`reader: <file:line>` and `reader_sink: agent_error_log` in the same commit that ships the reader —
the checker verifies both (`DBG-075`).

## 4. Traps for whoever picks this up

- **The rows expire.** 365 days under migration 567 (the code is declared `human-evidence` today),
  but a consumer that resolves rows drops them to the 14-day clock — extract what you need first.
- **A rerender is not proof.** Verify links at the served page; a rerender that hits the same
  timeout re-degrades and re-records, which is the loop working, not failing.
- The registry entry (`finding_code_registry.json`) carries the disposition and this bug's number;
  keep them in step or the daily check will say so.

---

> **CORRECTED 2026-08-25 (session `bugs_open/387`, lane `bugfix_392_link_context_unread`) — three
> claims above are wrong or stale, and the bug is re-scoped by owner decision. §1–§4 left exactly
> as written; read them with this on top.** Appended before any code was written.
>
> **1. The severity rests on damage that no longer exists.** Both motivating rows are healed or
> never published, verified 2026-08-25 through a join this file does not mention (below):
> agritec.uk `seaweed-and-the-carbon-question` was rewritten ~1h after the degrade and carries 3
> prose links; remortgagecalculator.uk `what-your-number-means` FAILED mid-build and has no
> `page_components` at all. §1's *"permanent-until-coincidence"* stands as a CLASS property —
> and the coincidence is exactly what happened here. **This is a latent class defect, not a live
> incident**, and it should not be read as three pages currently broken.
>
> **2. The row already joins to the exact page — §3's "the writer could also be extended to
> record the page id, which would make the join exact" understates what is there.** `context`
> holds only `site_id`, true; but the row's first-class column `orchestration_id` IS populated
> (filled by `actionJoinIdentity`, `platform/orchestration/actions/log_action_error.go:99-112`),
> and it resolved the exact page for **both** rows:
> ```sql
> SELECT orchestration_id,
>        collected_data->'input_data'->'current_page'->>'id'   AS page_id,
>        collected_data->'input_data'->'current_page'->>'name' AS page_name,
>        collected_data->'input_data'->'current_page'->>'url'  AS url
> FROM orchestration_states WHERE orchestration_id IN (
>   '405ee425-0d64-4db3-887b-0a4963d391e2','f8543e0c-9034-46a9-ae28-a25f8200ec5e');
> ```
> ⚠ that table's primary key is `orchestration_id`, not `id`, and the page is at
> `input_data.current_page.id`, not `input_data.page_id`. Recording `page_name` in `context` at
> write time is still worth doing, for a reason this file did not have: orchestration rows are
> reaped, the log row lives 365 days, so the join has an expiry the row does not.
>
> **3. FIX CANDIDATE 1 IS WRONG AS WRITTEN: a `page-rerender` work item cannot restore prose
> links.** Verified against LIVE `agent_definitions` 2026-08-25 (not the seed files):
> `page-rerender` neither spawns `page-content-writer` nor knows `edit_live`;
> `page-build-handler` does both; and `page-content-writer` is the **only** agent fleet-wide
> whose workflow contains `prepare_link_context`. A re-render regenerates the page FROM
> `content_data`, and `content_data` is where the missing links aren't. The working route is
> `item_type='content_rewrite'` with `spec.mode='edit_live'` at `page-build-handler`, which
> re-runs the writer under a fresh allow-list — ~30 such items/day today.
>
> **4. §4 trap 1 and §3 candidate 1's "extract first" are STALE.** Both say `resolved=true` halves
> a row's remaining life to 14 days. Migration `567_finding_codes_outlive_the_plumbing.sql`
> deleted that arm and `580_database_cleanup_comment_stops_lying_about_retention.sql` corrected
> the live sweep's own comment: **`resolved` no longer shortens a row at all**, and this code is
> not in 567's short-retention list, so its rows live 365 days either way. A consumer may resolve
> freely. The same stale sentence sits in the registry `why` for this code and in
> `cmd/content-loss-check/main.go:36-38`; both are flagged for correction alongside.
>
> **RE-SCOPED (owner decision, 2026-08-25): an artefact check, cause-agnostic.** A reader keyed on
> this code alone would address ~1% of the visible symptom — the code has 2 rows ever, while
> `[MEASURED 2026-08-25]` **411 of 736** deployed active pages carry zero writer-authored prose
> anchors (by page_type, zero/total: blog-post 31/164, guide 59/98, content 97/158, tool 159/200,
> landing 17/34, entity-page 15/17, section-index 20/44). For `tool` and `entity-page` zero is the
> NORM, so the check gates by page type and must never equate "zero body links" with "392 damage".
> Of the 187 on prose types, **48 of 48 owned pages are link-less** — reported, never repaired
> (filing there burns an LLM run then dies `wont_fix` at `OWNED_PAGE_GUARD`), and passed to the
> `bugs_open/396` and `bugs_open/333` lanes as their finding, not ours.
> ⚠ **Instrument:** count in `page_components.content_data`, never `rendered_html` — templates
> inject nav/hero/CTA so a degraded page still reads 2–3 there (that instrument gives 140, not
> 411). Hero/CTA links are structured fields (`cta_url`, `link_url`) carrying no `href=` and are
> excluded by design. See `WRONG_CALLS.md` 2026-08-25 — I made this error myself before catching it.
>
> **Plan, evidence and the queries:**
> `docs/agent_docs/docs024_key_docs_latest/bugfix_392_link_context_unread/{PLAN_2026-08-25_392,RUNBOOK_392,NOTES_392,README_where_we_are}.md`.
> Bug stays **OPEN** and is now **OWNED** by that lane.

---

> **CODE SHIPPED 2026-08-25 — `924079c94`, council `Council-Submitted: 7d923ff6`. Bug stays OPEN:
> the fix is committed but INERT, and by this estate's bar (fixed AND live) it cannot close yet.**
>
> **What was built.** One new in-process discovery check,
> `platform/orchestration/actions/discovery_checks/check_missing_prose_links.go`, registered
> LNK-039. Two arms:
> - **artefact** — a deployed prose page (`blog-post`/`guide`/`content`) with zero
>   writer-authored internal prose anchors gets one `content_rewrite` at `page-build-handler`
>   with `spec.mode='edit_live'`, `PageID` set, key `no_outbound_links:<page>:<site>`, capped at 2
>   per site visit;
> - **reader** — unresolved `LINK_CONTEXT_UNAVAILABLE` rows are resolved to their page, re-checked
>   and marked `resolved` on a **healed verdict only**. That is what flips this code from
>   `human-evidence` to `consumed` in `finding_code_registry.json`, done in the same commit with
>   `reader` file:line and `reader_sink`, as DBG-075 requires.
>
> **Nothing new is deployed** — no service, no image, no cron slot, no handler, and **no change to
> any live agent definition**. The repair route already runs ~30 items/day. `prepare_link_context`
> now also records `page_name`/`page_id`/`domain` on its own row, so a reader is not dependent on
> an orchestration row that is reaped long before the log row's 365 days are up.
>
> **Verification actually performed** (see `bugfix_392_link_context_unread/NOTES_392.md` for the
> commands): full `platform/orchestration/...` suite green, with the test *names* checked because
> a `-run` regex matching nothing also prints `ok`; **11 mutations applied, compiled, tested and
> reverted, none surviving**; `verify-head-builds.sh` post-commit reports HEAD builds. ⚠ The
> `--test` form is NOT a gate on this machine — HEAD fails 22 ways on its own (Kafka down locally,
> DB-dependent suites, other lanes' build failures), so the control is the DIFF: the failure set
> with this change is identical to bare HEAD's.
>
> **What still has to happen before this can close, in order:**
> 1. a fleet roll, then prove the check registered (`service_binary_capabilities`, positive +
>    negative controls — a name that must be present and one that must be absent);
> 2. add `missing_prose_links` to a live `run_checks.config.checks` array **by hand** (image
>    first, then config — an unregistered name is warn-and-skip, i.e. silently inert while reading
>    as enabled);
> 3. read the first pass **report-only** before it files anything;
> 4. the canary: induce the degrade on ONE page of a low-stakes site via a **site-scoped opt-in
>    hook**. ⚠ NOT via a bogus `site_id` — that corrupts the very column the reader joins on, so
>    the reader would false-pass by never firing; NOT `LOCK TABLE pages` (fleet-wide stall); NOT a
>    hand-inserted synthetic row (proves the reader, not the chain);
> 5. verify at the **served** page with `scripts/probe-page-url.sh`, never a composed URL.
>
> **Two live tripwires that nothing watches, stated rather than left implicit.** The remit's
> safety is not name-based — 103 deployed `tool-%` pages are typed `blog-post`, and the page-type
> rule is acceptable only while it agrees with a structural "carries a tool component" rule, which
> it does today and could stop doing silently. And the repair route fails or is refused ~21% of
> the time, so a filed item is not a repaired page.

---

> **COUNCIL APPROVED 2026-08-25 (round 2, corr `7d923ff6`) — 15 seats, `architecture` among those
> approving, so not architecture-scope by ruling rather than by assumption. Six advisory
> objections, none high-severity; one of them corrected a published number in this file.**
>
> ⚠ **CORRECTION to the re-scoping block above: "411 of 736" was measured with the WRONG
> PREDICATE.** The `debug_historian` seat objected that the census never stated its predicate —
> the documented trap for audit-shaped checks — and it was right. I censused with the hand-rolled
> `status='active' AND build_status='deployed'`; the check itself takes the shared family,
> `PageWantedLivePredicateFor` **AND** `PageHasShippedPredicateFor`. Re-measured with the check's
> own predicate: **444 of 773**, and the gap is **51 pages visible to the check and invisible to
> my census, zero the other way** — pages carrying a `deployed_at` but currently flagged
> `needs_rebuild`, exactly what that predicate exists to catch (`bugs_closed/037`). **The code was
> already right; the number published beside it was not.** In-remit: content 107, guide 59,
> blog-post 31 = **197**, of which **145** also clear the 1,500-char prose floor. So the honest
> day-one actionable figure is **145**, not the 139 stated earlier.
>
> **The other five, and why none required a code change:**
> - `editquality`: flipping the registry to `consumed` overstates the live state while the reader
>   is inert. **Accepted** — the registry note now says plainly that the reader is committed and
>   NOT yet running, and names both steps that must land before it is.
> - `bug_historian`: re-invoking the writer over an otherwise-good page is `bugs_closed/012`'s
>   shape and wants a content-preservation check. **Already exists and is on by default** —
>   `save_sections_shrink_guard.go`, wired into `SavePageSectionsAction:673`, per-slot floor 0.5,
>   built for `bugs_open/178`, which is the same class.
> - `reuse_agent`: `resolve_internal_links_action.go` was never named as a rejected reuse
>   candidate. Fair. It augments **CTA-bearing** sections' `resolved_data` at build time — the
>   structured-link surface this check's instrument deliberately excludes. Different surface,
>   different moment.
> - `tooling_provenance`: asked whether the check registers with the discovery registry. It does
>   (`init() → Register`); the submission sketch simply did not show it.
> - `prior_art_librarian`: the exclusivity claims were narrative rather than queried. The queries
>   are in `bugfix_392_link_context_unread/RUNBOOK_392.md` §3.
>
> Commit trailer `Council-Reviewed: 7d923ff6-3810-4f2b-9000-e02df68a6b9e` — written only now, on a
> verdict actually read.
