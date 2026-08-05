# HANDOFF — `bugs_open/098` unpublish primitive · **START HERE** · 2026-08-03

Cold-start for the next session. Read this, then `NOTES_unpublish_primitive.md` (the
missteps — they are the expensive part) and `RUNBOOK_unpublish_primitive.md` (every
command, with its gotcha).

---

## STATE 2026-08-05 — POPULATION ZERO, ACCEPTANCE COMPLETE; one small code task (debt 5b), then close

**Read this block first; everything below it is background.**

- **The batch retraction HOLDS.** 10 leopardess pages retracted 08-04 evening (run
  `e23b7257…`, owner-approved); re-checked 08-05 morning after BOTH refresh windows:
  all 10 still 404, robot-hands' original still 404, collateral 200, **0** new
  `page_rerender` rows. Two-part acceptance COMPLETE. Population of the bug: **zero**.
- **DECISION (owner-delegated, recorded in bug + RUNBOOK): archiving does NOT
  auto-retract.** Two-step procedure: archive, then 216 trigger with PAGE_IDS.
- **⚠ Debt 5's sibling-key half is REFUTED — do NOT trust `retraction_audit` in
  collected_data for awaited runs.** `persistAwaitingStateWithRetry`
  (coordinator.go:2052) parks a step onto a FRESH DB load carrying only awaited-request
  bookkeeping — every CollectedData mutation dies there. Proven live (strings-verified
  binary, absent key). The `agent_error_log` half (refusal rows, direct INSERT before
  dispatch) IS durable and proven. Full account: RFC_012 **addendum 2**, LANDMINES
  (corrected in place), WRONG_CALLS 08-04.
- ~~**THE ONE CODE TASK LEFT — debt 5b, small and specified:**~~ **DONE 2026-08-05**
  (committed, `Council-Submitted: 9a38c785-7733-4492-b614-2f67bf4e36c4`; live probe
  post-roll specified in the bug file's debt-5b block — one ACTIVE page id through 216,
  expect RETRACTION_AUDIT info + RETRACTION_REFUSED warning rows, zero dispatch).
  Original spec, for the record: in
  `retract_page_deployment_action.go`, extend `recordRetractionConditions` (or add a
  sibling call beside it) to ALWAYS write one `agent_error_log` row per real run —
  `error_code='RETRACTION_AUDIT'`, severity `'info'`, context = the full audit map —
  through the same proven INSERT (`insertRetractionConditionRow`). Keep the sibling-key
  attach (harmless; correct for non-awaited paths and for whenever RFC_012 lands). Test
  it at the DB layer the last one missed: sqlmock-expect the info row on clean runs, none
  on dry runs; and VERIFY LIVE by SELECTing the persisted row while a probe step is
  parked. One coherent task ⇒ one council round (`RESUBMIT_CORR` NOT needed — new task).
- **THEN: close 098** (owner has effectively invited it — "say the word") → move to
  `bugs_closed/` by `git mv` + BOTH paths named on the commit (see the git-mv landmine),
  update `MEMORY.md` + topic file + `MEMORY_closed.md`, 016b §10 index.
- **Owner decisions still open (in `SUMMARY_2026-08-04`, §Where we're going):** RFC_012
  options A/B/C (filing thread recommends B, DB-backed — amended by addendum 2); RFC_011's
  deferred destructive-verb question (only when the next verb arrives).

## STATE 2026-08-04 (end of the debts session) — ALL FIVE DEBTS PAID; what remains is OWNER DECISIONS

| | |
|---|---|
| Debt 5 (audit lost to the await) | **PAID + LIVE** — chassis `sha256:0e99ace…`, both replicas; pod-grep `retraction refused for page`=1, `retraction_audit`=1, pre-existing control =1. Audit → `collected_data.retraction_audit` (sibling key) + `agent_error_log` rows (`RETRACTION_REFUSED`/`RETRACTION_STRANDED_TARGETS`, warning, written BEFORE dispatch); `conditions_recorded/_lost` in the audit |
| Debts 3+4 (copied census / bespoke predicate) | **PAID** — debt 3 as a LOCKSTEP (`datahelpers.InboundLinkSurfaces` + tests both sides; queries deliberately NOT unified, opposite safe-failure directions); debt 4 as the family's missing lifecycle member `datahelpers.PageWantedLivePredicateFor` (12 sites after the census; byte-identical SQL so liveness is unobservable BY DESIGN) |
| Council | **BOTH rounds APPROVED**: `5a965452…` (debt 5; 8 seats, round 1 died to a fleet roll — resubmit with `RESUBMIT_CORR` keeps trail + trailer) and `37593214…` (debts 3+4; its census objection was RIGHT — found 5 more sites, migrated) |
| RFC | `architecture_review/RFC_012_the_await_overwrite_destroys_action_findings.md` — filed at the architecture seat's direction; **already gained an addendum from the bugfix_192 lane** (the SAME write unguarded in `storeActionResult`, cost a fleet-wide outage). The RFC is becoming the class's home. **Needs the owner: rule on options A/B/C (filing thread recommends B)** |
| Commits | `e35e549a8` (debt 5) · `b99f883e3` (roll-kill note) · `9e0274f53` (verdict + RFC 012) · `6a7ab87a8` (debts 3+4 + PBP-029) · `185bb98d9` (census follow-up) · `28ff21fcd` (verdict records) — all with resolving council trailers |
| Register | PBP-029 (+ index row 1,717 → 1,718) |
| Landmine | LANDMINES.md gained "an action that RETURNS findings and AWAITS a response loses the findings" — synced to doc_notes |

**Known-and-skipped, NOT missed** (pick up when their lanes are clean): 2 lifecycle
sites in `tool_acceptance_actions.go` (was another session's dirty WIP on 08-04) and 1 in
`check_page_canonical_collision.go` (080 lane's fresh PLAN-047 code, and its COALESCE `=`
form is NULL-identical anyway).

**NEXT — every remaining item is an owner call, none is code:**
1. Retract the remaining **11** frozen-artefact pages? (10 = leopardess ordinary content;
   measured by curling all 14 on 08-03 — never derive this count from `deployed_at`.)
   Mechanism is proven: `sql_for_agents/216_TRIGGER_page_retraction.sh`, ALWAYS with
   `PAGE_IDS`; acceptance is the TWO-PART test in the bug file.
2. Should archiving invoke retraction automatically? (Nothing in code archives a page
   today, so "no — the runbook is the mechanism" may be the honest answer, but decide it.)
3. RFC_012 options A/B/C — now urgent-adjacent, since the 192 addendum shows the class
   biting elsewhere.
4. The bug stays OPEN until 1/2 are decided — do not close on the one retraction.

---

## STATE IN SIX LINES

| | |
|---|---|
| Bug | `bugs_open/098` — **STILL OPEN**, do not close |
| Resurrection fix | **LIVE**, chassis digest `sha256:5da2888…`, both replicas |
| Unpublish primitive | **BUILT + LIVE**; council REJECTED → **RFC 011 DECIDED, option B** (verb kept, dropped from the generic allowlist) |
| Council correlation | `4a7f0877-4149-4431-97d4-318d093570a4` (rounds 1 and 2) |
| Routed to | `architecture_review/RFC_011_git_adapter_action_vocabulary_and_the_unpublish_verb.md` |
| Retraction performed | **ONE, done and proven** — robot-hands `/learning-center/index.html` 200 → 404 via the `page-retraction` agent. **13 pages still serve frozen artefacts** and the class grows ~1/day |

**Commits (all carry `Council-Submitted: 4a7f0877…`; the verdict is REJECTED, so `098`
will bucket them and that is correct — do NOT write `Council-Reviewed:`):**
`43c259f46` primitive · `5b66615d4` resurrection fix + graph audit · `6d2c3006b` bug
correction + seed · `8f73e7279` cousin fix · `89b6d7963` verdict + RFC 011.

---

## THE ONE THING THAT WILL BITE YOU FIRST

**This bug's own closing criterion does not work, and it is the bug's own lesson.**
`SELECT count(*) FROM pages WHERE status='archived' AND deployed_at IS NOT NULL` — stated in
the file as "after the fix: 0" — **did not move when the page was successfully retracted**.
It cannot: the retraction deliberately does not clear `deployed_at` (that is a shared-column
semantic change, deferred with a census). The count measures *archived-and-once-deployed*.
To measure what is actually still served, ask the artefact: for each archived+stamped page,
check whether its derived path exists in **its own** deploy repo (`sites` or `vm-sites`).

**Do not close 098 on the one retraction.** Archiving still does not retract *by itself* —
nothing in the codebase archives a page, so nothing invokes the retraction either. 13 pages
still serve frozen artefacts and one more arrived today from another lane.

## WHAT THE VETO ACTUALLY SAID

Not "this is wrong". The `architecture` seat endorsed the design in terms worth quoting
back: *"expressing delete-as-null-sha inside the existing CommitToRepo path is the right
reuse, inheriting retry, prefixing and atomicity for free rather than bolting on a
parallel write path. On that axis the plan is sound and I'd carry it."*

It vetoed **how it reached production**: `delete_file` is a new reserved verb on a shared
adapter's vocabulary, added inside a bug fix — `bugs_closed/124`'s shape. The allowlist
edit is the sharp end: `gitAdapterActions` gates what *any* workflow may ask the git
adapter to do, and its own comment used to say destructive verbs were unreachable through
it.

**CLAUDE.md is explicit: a scope veto is NOT answered by resubmitting with better
measurements.** Do not re-run the council on this. RFC 011 costs four options and records
a preference (option B: keep the verb, drop it from the generic allowlist).

## OWED — correctness debts on code that is ALREADY LIVE

These are not packaging arguments; they are defects. All four land on
`retract_page_graph.go`, the part added fastest.

**DEBTS 1 AND 2 ARE PAID** (commit `946fe4280`) — 2 was a real defect, not a theoretical
one: the widened scan found a referrer the markup-only scan missed
(`gripper-cycle-time-estimator/hero`, a CTA url in `content_data`; `/contact.html` went
4 → 5 referrers). **NEW DEBT 5, and it is the important one: the graph audit's findings
never reach the durable record** — the step awaits the adapter response and the await
machinery overwrites `output_field`, so refusals survive only in ephemeral pod logs. A
retraction that refuses a page would today refuse it SILENTLY, which is the opposite of the
stated design. Fix by persisting the audit before dispatching, or emitting a work item.

Original list, for the record:

1. ~~**HIGH — the link census does not use `linkablePageStatusPredicate`.**~~ **DONE** A landmine says
   an offline census over `pages` must, or an archived page makes a correct result look
   wrong. Verify and fix. *(This one is in the auto-memory index; I built the thing it
   warns about.)*
2. ~~**MEDIUM, the sharpest catch of the round — inbound references also live in
   structured `content_data` fields.**~~ **DONE, and it was real** — see above.
3. **MEDIUM — the inbound-source logic is COPIED from `check_orphan_pages.go`** with only a
   comment warning of drift. Extract and share it. (My argument that the two ask different
   questions — what IS unreachable vs what BECOMES unreachable — is in the file header and
   is not wrong, but a comment is not a mechanism.)
4. **MEDIUM — `status='active'` is a third bespoke spelling** rather than a consolidation
   onto a canonical helper. The census half of this was answered after submitting: the only
   other selector with the defect was `queueDirectoryPageRerenders`, fixed in `8f73e7279`.

## WHAT IS SETTLED, SO YOU NEED NOT RE-DERIVE IT

- **The far side already reconciles.** `b2 sync --delete` + a Cloudflare purge per changed
  domain, on every push. Removing a file from the repo genuinely retracts the page. 098's
  "make deploys reconciling" option was half-built already.
- **GitHub's deletion semantics, probed live** (POST `/git/trees` — creates an unreferenced
  object, no ref moves, no workflow fires): null sha on an existing path → **201**; on an
  absent path → **422 `GitRPC::BadObjectState`**. The existence filter is required.
- **Path derivation** (round 1's gating objection, answered by measurement): the real
  `PageFilePathFromURL` over all 13 candidates against **each page's own deploy repo** —
  11 of 12 `sites`-repo pages have a file at exactly the derived path, the 12th is
  genuinely absent and 404s, the 13th deploys to `vm-sites` and is correctly absent there.
  **Zero mismatches.**
- **The primitive works**, proven live: write → delete → idempotent re-delete on a scratch
  path the deploy workflow ignores, no-op logged, branch head unmoved.
- **`page-retraction` agent is APPLIED and PROVEN** (`sql_for_agents/215`), dispatched by
  `sql_for_agents/216_TRIGGER_page_retraction.sh`. One real retraction has run through it
  end to end. **Always pass `PAGE_IDS`** — omitting it puts every non-active stamped page on
  the site in scope.
- **The retraction itself**: 200 → 404, file gone, deploy workflow green, six neighbouring
  live pages unaffected, and the live audit agreed with the hand measurement (0/0/0).

## TRAPS THIS SESSION HIT (all in NOTES, worth 30 seconds)

- **Comments are not in the binary.** I pod-verified a chassis image by grepping for
  strings I had added *as Go comments* — read 0 for everything, which looks exactly like
  "the fix did not ship". Use a string literal the change introduced: SQL, a log line, an
  error message.
- **A control is code, and unreviewed code is where errors live.** Three separate
  verification checks were wrong this session while the code was right, including a
  hand-rolled "negative control" regex that fired on the fixed binary.
- **`sites.github_branch` says `main`; `gqls/sites` has no `main`** (it carries `master`,
  its default, and `750start`), and the deploy workflow triggers only on `master`. Never
  pass that column to a commit or retraction.
- **Check each page against ITS OWN deploy repo.** relojistas is `vm-sites`; a sweep
  hardcoding `gqls/sites` reports a stale leftover as the live artefact.
- **A link census that reads one table answers a question about that table.** My first
  orphan check read only `page_components` and confidently reported 10 of 16 pages would be
  stranded. The nav and site chrome are where most links live; the real answer was zero.
- **`go build` cannot parse SQL, and `[]string` is not a driver type here.** PREPARE against
  the live schema (caught `site_components.component_type`, which does not exist — it is
  `slot_name`) and pass arrays via `datahelpers.PGTextArrayLiteral` with `::text[]`.
- **All-zero guard results look identical to a broken query.** Always re-run against a case
  you know is positive: on robot-hands, `/contact.html` gives 4 body / 2 chrome / 1 nav.

## THE 090 DIAGNOSIS RUN RETURNED NOTHING

Filed as required for a structural claim (`5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb`). It
**completed and wrote three bundles and no verdict** — no confirmation, no refutation. NOTES
records the six-step first-hand chain standing in for it, per the owner ruling of
2026-07-31 that a substitute must be *declared*, not silently omitted. Do not cite that run
as corroboration.

## NEXT ACTIONS, IN ORDER

> **CORRECTED 2026-08-03 (late): item 1 is DONE — committed, not yet live** (commit with
> `Council-Submitted: 5a965452-a9a0-40a6-a990-410f14ac32b0`; the chassis build in flight at
> commit time predates it, so verify against the NEXT build: pod-grep the added literal
> `retraction refused for page`). The audit now lands in `collected_data.retraction_audit`
> (sibling key — the await overwrite only touches the step-name and output_field keys) and
> every refusal becomes an `agent_error_log` row (`RETRACTION_REFUSED` /
> `RETRACTION_STRANDED_TARGETS`, severity `warning`), written before dispatch on real runs.
> Details + the two design points in NOTES (2026-08-03 late entry). Also note the serving
> population is **11, measured by curling all 14** — the "13" below was stamp arithmetic,
> corrected in the bug file the same evening.

1. ~~**Debt 5 first** — the audit's findings are discarded by the await.~~ **DONE, see above** —
   and **council APPROVED 2026-08-04** (8 seats, no high-severity; the coordinator-level
   question went to `architecture_review/RFC_012` at the architecture seat's direction).
2. ~~Debts **3** (share the inbound-source logic with `check_orphan_pages` rather than copy)
   and **4** (consolidate the status predicate).~~ **DONE 2026-08-04, APPROVED same day**
   (`37593214…`, 4 advisory objections none high; its census ask found 5 more sites,
   migrated in the follow-up `185bb98d9` — 2 known-and-skipped, see NOTES): debt 3 as a
   lockstep — `datahelpers.InboundLinkSurfaces` + tests both sides, queries deliberately
   NOT unified; debt 4 as the family's missing lifecycle member
   (`datahelpers.PageWantedLivePredicateFor`), seven sites, byte-identical SQL. Register
   PBP-029; details in NOTES 2026-08-04.
3. **Confirm with the owner before retracting more.** 11 pages still serve frozen artefacts;
   the owner deliberately scoped this run to one. 10 of the 11 are on
   leopardessconsulting.co.uk and are ordinary content pages, a lesser problem than the one
   just fixed — they advertise nothing broken.
4. **Fix 098's acceptance criterion** (see the top) before anyone tries to close on it.
5. Consider whether archiving should invoke retraction automatically. Today nothing archives
   a page in code, so the honest answer may be "no, and the runbook is the mechanism" — but
   that should be a decision, not an accident.
