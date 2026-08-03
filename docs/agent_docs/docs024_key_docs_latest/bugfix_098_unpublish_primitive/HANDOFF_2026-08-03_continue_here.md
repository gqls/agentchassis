# HANDOFF — `bugs_open/098` unpublish primitive · **START HERE** · 2026-08-03

Cold-start for the next session. Read this, then `NOTES_unpublish_primitive.md` (the
missteps — they are the expensive part) and `RUNBOOK_unpublish_primitive.md` (every
command, with its gotcha).

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

1. ~~**Debt 5 first** — the audit's findings are discarded by the await.~~ **DONE, see above.**
2. Debts **3** (share the inbound-source logic with `check_orphan_pages` rather than copy)
   and **4** (consolidate the status predicate).
3. **Confirm with the owner before retracting more.** 11 pages still serve frozen artefacts;
   the owner deliberately scoped this run to one. 10 of the 11 are on
   leopardessconsulting.co.uk and are ordinary content pages, a lesser problem than the one
   just fixed — they advertise nothing broken.
4. **Fix 098's acceptance criterion** (see the top) before anyone tries to close on it.
5. Consider whether archiving should invoke retraction automatically. Today nothing archives
   a page in code, so the honest answer may be "no, and the runbook is the mechanism" — but
   that should be a decision, not an accident.
