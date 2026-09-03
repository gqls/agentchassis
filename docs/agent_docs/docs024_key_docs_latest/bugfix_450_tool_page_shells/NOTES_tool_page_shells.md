# NOTES — bugs_open/450 tool page shells (append-only, newest at the bottom)

## (a) 2026-09-03 ~09:0xZ — lane opened; ownership, validity and the 090 verdict

**Ownership.** `scripts/who-owns.py 450` → OWNED/recently-active, naming
`docs024_key_docs_latest/portfolio_positioning` (87 commits/14d). Read its
`HANDOFF_2026-09-02_continue_here.md` addenda 2 and 3 before touching anything: that lane FILED
the bug and holds the **instance** work (owner ruling 2026-09-03, commit `b47b626c7`: build the 8
planned tools, keep cluster duplicates, chassis roll imminent). Its §7 answer and the 444 CONTRIB
are both addressed "to the fixing thread", and no `bugfix_450*` directory existed. Conclusion: the
class is unowned; this lane takes it. No competing fix — the instance/class split is clean.

**Validity re-check at the live DB** (the bug is a day old and other sessions move fast):

```sql
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.page_type='tool' AND p.status='active' AND p.deployed_at IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
                   WHERE pc.page_id=p.id AND pc.build_status<>'removed' AND cc.component_level='tool')
 GROUP BY 1 ORDER BY 2 DESC;
```

→ loanandmortgagecalculator 16 · webdesign.co.uk 14 · loanzy 11 · **seotools 7** · loancash 3 ·
idea.uk 3 · vonc 3 · leopardessconsulting 2 · cv1 1 · boxingonline 1 = **61 pages / 10 sites**,
identical to the filing census. Queue state the same day: `owned_page_review`
171 `needs_human_review`; `unbuilt_internal_link` **339 unresolved**, 158 failed, 88 HITL, 22
triaged. Bug is live and reproducing.

**090 verdict read** (the 444 CONTRIB made this a precondition). ⚠ It is NOT in `doc_notes` — the
bug file's own warning is right. The query that works:

```sql
SELECT result FROM site_work_items
 WHERE spec->>'dispatch_correlation_id' LIKE '96e97dc4%' AND item_type='needs_diagnosis';
```

44.7 KB of JSON; the `conclusion` field (4,916 chars) restates the chain with `[static]` grounding
on `owned_page_guard.go:176-190` and `[state]` reads of the seotools rows. Status **CONFIRMED**.
No re-run needed — the mechanism is settled; what this lane adds is the fix, not the diagnosis.

**Code files untouched since filing** — `git log --since=2026-08-30` on
`check_phantom_internal_links.go`, `owned_page_guard.go`, `save_page_sections_action.go` returns
nothing, and none is dirty in the tree. Safe to build on the filed mechanism.

## (b) 2026-09-03 ~09:2xZ — three findings that redirected the design

Exploration of the three doors (phantom-link routing / the owned-page guard / the 444 gate) turned
up three things that changed the plan rather than confirming it:

1. **`rebuild_policy` has no transition mechanism.** Zero `UPDATE … rebuild_policy` in Go, ever.
   Two INSERT-time writers only. The column is CHECK-constrained to `'generic'|'owned'`. So the
   bug's candidate 2 ("set the policy when the hold is filed") would have introduced the estate's
   first policy lifecycle *and* left no event to clear it when the tool lands. Redirected to a
   **derived** predicate, which self-clears by construction.
2. **Candidate 3's premise is only half true.** LNK-038 suppresses links to never-shipped pages at
   render, but (a) it states in its own source that it deliberately does **not** silence
   `check_phantom_internal_links`, which reads STORED html — so the items keep being minted; and
   (b) its predicate requires `build_status='planned' AND updated_at < NOW() - 48 hours`, and
   450's whole timeline is **under four hours** (plan 16:13Z → shells written 19:57–20:41Z). So
   LNK-038 refuses nothing on a fresh remake, and once the shell deploys the page leaves the
   predicate for ever. "The repair is redundant because LNK-038 hides the links" would have been a
   false premise to build on.
3. **The `owned_page_review` emitter is `ReconcileSitePlanAction`, not `validate_site_plan`.**
   Same workflow, later step, with `sync_pages` minting the page rows in between. Recorded as a
   correction in PLAN §5 and to be carried into the bug file.

Also confirmed, because the fix depends on it: **`deploy_tool_action.go` INSERTs the tool
component at :517 and raises the companion `needs_content_page` at :564** — component first. So a
derived predicate is already false by the time the companion item is written, and the
portfolio lane's imminent `add_tool` wave at the seven seotools shells cannot be parked by this
fix. This was the one ordering that could have made the fix harm a peer lane's live work.

## (c) 2026-09-03 — a claim of mine that was wrong, caught in the same session

I carried "578_retype_mislabelled_tool_rows_HOLD.sql retypes mislabelled `pages.page_type` rows"
into the design brief as evidence that mislabelled tool pages are a real population. **It retypes
`page_components`** (tool bytes in `hero` rows), not page rows. Nothing in the tree retypes
`pages.page_type`. The design decision it was cited for (D5, and the "loud misfire" argument)
does not depend on it, but the citation was false and would have gone into a council submission as
evidence. Caught by a subagent re-reading the migration I had only grepped. Logged in
`WRONG_CALLS.md`; the cheap check was `head -40` on the migration instead of trusting its filename.

## (d) 2026-09-03 ~11:0xZ — commit 1 landed EARLY and out of order, because HEAD was broken

`587666be8` — the derived refusal, its six call sites and seven test files.

**It landed before its council submission and before its register entry, which is not the
practice.** The `bugs_open/427` lane committed one line in
`rerender_page_sections_action.go` (their 454 fix) with a correct explicit pathspec; my
half-finished rename was dirty in that same file, so their pathspec took it. HEAD then called
`pageRefusesGenericBuild`, `refusalToolPending` and an 8-arg `emitOwnedPageReviewItem`, none
committed — and `make build-*` builds from HEAD, so every session's next image build was broken.
They measured the minimal closure (six files), verified it with `verify-head-builds.sh --with`,
and **messaged me rather than committing my in-flight work under their name.** That was the right
call and I have said so.

Sequence after that: `gofmt -w` on the new test file (the pre-commit pattern check caught it),
commit refused once by the **trailer gate** — I had written `Council-Submitted: pending-post-roll`
as a placeholder and the gate correctly refuses a non-UUID join key, since 098 resolves it to
nothing and forward-only forbids fixing it by amend. Dropped the trailer (the submission did not
exist yet; committing before a submission needs no trailer at all), committed, and
`verify-head-builds.sh` reads **OK — HEAD 587666be8 builds**.

**The misstep is mine and it is not the passenger.** I held a shared-package RENAME dirty across
a long design phase. A rename breaks the package for everyone from the first save until the last
call site lands, so the window I left open is the whole defect. It cost two other lanes before
either touched my code: the 440 lane's mutation re-proof read `build failed` three times on my
half-committed symbols and drew a wrong conclusion about its own tests. Logged in
`WRONG_CALLS.md` (2026-09-03, "I held a shared-package refactor dirty…").

## (e) 2026-09-03 ~11:1xZ — council submitted; what the submission concedes

Corr **`2b236e83-ffd1-4911-b73f-1c17249064c1`** (`council_submission_450_r1.json`). `DRY_RUN=1`
admission passed first, free, before spending anything.

Three things stated in the submission rather than left for a seat to find:

1. **The `578_retype_mislabelled_tool_rows_HOLD.sql` citation is WITHDRAWN in the submission
   itself.** I had it as evidence that mislabelled tool PAGES are a real population; it retypes
   `page_components`, not `pages.page_type`. Better to withdraw a citation in the submission than
   have a seat refute it — and the design decision it was cited for does not rest on it.
2. **"Is there a live population of mislabelled tool-typed pages whose generic rebuild is
   actually wanted?" is marked `[UNMEASURED]`** rather than argued away. It is the one question
   that could turn this fix into a false-refusal source, and I have not run the census.
3. **The out-of-order process** is declared in the rationale, not hidden. The commit carries no
   `Council-Reviewed:` trailer and makes no review claim.

## (f) 2026-09-03 — a trap the peer lane handed me, worth more than the fix it came with

**Do not verify this fix with a re-render.** Since 2026-09-02 a light re-render renders the
page's own stored `content_data` back at itself — clean run, healthy count, nothing delivered
(`bugs_open/454`, fix committed `9831e9ab4`, NOT yet in a rolled image). Had my post-roll check
been "re-render a shell and look at it", I would have been reading a mirror and would have
concluded something about my own guard from it. RUNBOOK §8b now says so. This is the
`a-plausible-external-cause-is-when-to-doubt-your-instrument` shape arriving as a gift instead of
as a lost day, and it came from telling a peer what my change did to a function they own.

## (g) 2026-09-03 ~11:4xZ — the repair wave's numbers, and the one instance that is interesting to THIS lane

From the `portfolio_positioning` lane, on the instance half (their measurement, not mine):
all **7 seotools tools built 09:30–09:54Z, no retries**, every one **adopting its existing page
at the existing URL** (`page_adopted: true`, no duplicate rows). Nothing published yet — all 7
`page_rerender` items sit `triaged` behind older site backlogs and they are not jumping the
queue. So the 61-page census should shrink by 7 shortly, by 8 once websitepromotion's
`tool-channel-prioritiser` builds.

**The 8th is the one worth watching from here.** It is the SECTIONLESS variant: its planned page
had no `site_plan_sections`, so the link repair parked all 7 of its `unbuilt_internal_link` items
at `mark_no_ready_sections` (HITL) instead of writing a shell. That fork is the reason the
plan-side arm (candidate 1) is designed to hold **empty-sectioned tool pages too**: "no shell" is
not "harmless" — it converts a served-prose bug into a recurring HITL tax on a page row no
producer will ever fill. This instance is the live evidence for that design decision, and it is
worth re-reading when the arm is written rather than trusting my summary of it.

**Their finding that sharpens our own file:** the position-2 collision is not incidental —
`create_tool_component` inserts the widget at a **hardcoded position 2** ("same as
deploy_tool_action") without consulting what the page already holds, so every repaired shell ends
with the tool AND the old prose block sharing position 2, ordering decided by whatever the
renderer's `ORDER BY` leaves. Theirs to decide; ours only to note that after the tool attaches
our refusal has lifted, so nothing of ours is holding that still for them.

## (h) 2026-09-03 — council r1 dispatched and RUNNING; the DB is slow enough to be worth a note

Corr `2b236e83-ffd1-4911-b73f-1c17249064c1`, observed at `review_constitution|EXECUTING_STEP`.
⚠ The cluster DB was slow enough this session that a plain `psql -c` **timed out repeatedly at
100–120 s** while the fleet was busy. That is load, not a dropped dispatch, and the CLAUDE.md
warning applies exactly: **do not retry the trigger on that evidence** — it costs a duplicate
round. Poll in the background instead of blocking, and find the run by payload
(`collected_data->'input_data'->>'fix_correlation_id'`), never by the printed `RUN_ORCH_ID`,
which is not the id the chassis assigns.

## (i) 2026-09-03 ~12:3xZ — council APPROVED round 1, and the four mediums were all worth answering

Corr `2b236e83-ffd1-4911-b73f-1c17249064c1`: **APPROVED, 4 advisory objections, none high**, 16
seats, 3 abstained. Approval does not make the objections wrong, and three of the four named a
check I had ASSERTED rather than run. All are now measured. **This is the "a REVISE round is
cheaper than the defect it finds" lesson arriving on an APPROVED verdict.**

**1. `prior_art_librarian` (medium) — and it caught a FALSE CLAIM of mine.** It objected that
"nothing in this estate has ever UPDATEd `rebuild_policy`" rested on my own grep, and that if an
UPDATE existed the design argument collapses. It was right to ask. **SIX hand-run migrations SET
that column** — 164 (the seed backfill), 195, 367, 377, 667, 668 — two of them (667/668, terms
and privacy locking) more recent than anything I had looked at. What is true is the narrower
claim: **zero Go UPDATEs, so no AUTOMATED transition** `[MEASURED 2026-09-03]`. The design
conclusion survives and is arguably stronger — a per-page lifecycle whose only clearing event is
a human writing a migration is not one a planner can rely on — but **the claim as written was
false and is corrected in the code comment**, not just here. WRONG_CALLS logged.

**2. `debug_historian` (medium) — "settle risk #1 before merge, not after".** The mislabelled
-`page_type` population I had carried as `[UNMEASURED]`. Now measured, and it reframes the whole
change: of the **67 pages** the predicate matches, **48 are ALREADY `rebuild_policy='owned'`** and
were refused by the old guard too. The genuinely NEW refusals are **19 pages** — 18 under
`/tools/`, and **exactly ONE elsewhere: `idea.uk` `/report.html`**, typed `tool`, six components,
no tool. That is the entire mislabelled-suspect population: one page, failing loud. My submission's
risk section implied a possible unmeasured class; it is a single row.

**3. `guardian` (medium) — the cost on the universal insert path was "asserted, not measured".**
`EXPLAIN (ANALYZE)` on the door's read: for a non-tool page the EXISTS subplan reads
**`(never executed)`** — Postgres short-circuits on the `page_type` test, exactly as the comment
claimed — with the whole read at **~2.2 ms**, in line with the ~2.7 ms the 333 door already
documents for the plain policy read it replaces. On a tool page the subplan adds ~2.3 ms, on a
small minority of writes. Both figures are now in the code.

**4. `reuse_agent` (medium) — does `check_missing_tools.go` already watch these pages?** Read it:
**no overlap.** It counts tool COMPONENTS site-wide and files one per-SITE `evaluate_tools` item
at `tool-suggester` ("should this site get tools?"). Ours is per-PAGE, different item_type,
different handler, different question. There is no second disposition path for one defect — the
seat's founding failure mode does not apply, but the question was the right one to ask.

**5. `architecture` (medium) — do parked `tool_pending` items reactivate, or pile up?** Honest
answer, and it is a real limitation rather than a clean pass: `deferred` is in **neither**
`workItemTerminalStatuses` nor `workItemClosedStatuses`, so a parked row **holds its dedup slot**
(re-finds collapse onto it instead of stacking) and is **retracted normally** when its detector
stops reproducing it. But nothing PROMOTES it back to dispatch when the tool arrives — by design,
since promoting it would dispatch work the handler refuses. So the row waits for its detector's
next pass to retract it. That is the same mechanism the owned-page door has used since `333`, and
it is **not** the "hold with no consumer" this bug complains about (those never dedup or retract),
but the distinction is thinner than I would like and is worth stating rather than glossing.

**6. `editquality` (medium) — a submission-quality miss, not a code one.** My sketch narrated
`emitOwnedPageReviewItem`'s signature change and the `censusExcludedOwnedPages` rename in a
trailing comment instead of showing them as edits. The code is correct; the SUBMISSION was
under-specified, and the seat could not see the change it was being asked to approve. **Reviewers
judge the sketch — the RUNBOOK's own warning, which I read and then did exactly this.** Next
submission: every signature change gets its own diffed edit.

## (j) 2026-09-03 ~13:xxZ — the supply half built; the census correction that arrived mid-build

`5e6fee47b` (Go, inert) + `681190083` (migration 729, **committed unapplied**) + BLD-029.

**Why the migration is held, and it is not the usual caution.** Its prompt text asserts
*"validation holds back tool pages whose tool does not exist"* — false until a chassis carrying
`5e6fee47b` rolls. The KEY is order-safe early (old binaries ignore it); the SENTENCE is not,
because it would describe a validation that is not running. Both wait. RUNBOOK §10 carries the
preconditions in order.

**Rehearsed rather than trusted** — 720's lane found its own guard arithmetic wrong exactly this
way. Inside a rolled-back transaction: apply passes every guard and its verify block
(`enforce_tool_sources=true`, 720's flag still `true`); the apply→ROLLBACK round trip returns the
template to md5 `85b9821d6d75e8142245552c8986d38b`, **byte-identical**, key ABSENT. The verify
block also defends 720's sentence, 720's flag, 433's directory rule and 718's imagery surface,
because three lanes edit that one row and eating a neighbour's sentence would look like a clean
apply.

**Mutation results, all three arms killing DIFFERENT sets** (this is what makes the suite worth
anything): resolver always-producible → the five Held/receipt tests fail, **no Kept test does**;
removing the page-name candidates → exactly the three name-based tests fail, section-name and Held
stay green; census fail-closed → only `CensusErrorFailsOpen` fails. The last is the one that
matters: fail-closed there would starve every fresh build on a transient DB error, which is the
deadlock 444 warned about arriving by accident rather than by design.

**One real collision, caught by the estate's own tooling:** my first test file declared
`containsString`, which already exists in `v3_site_actions.go`. `verify-head-builds.sh --test`
caught it against HEAD; I deleted mine and used the package's. Worth noting because the local
`go test` could NOT have caught it at that moment — another session's in-flight
`datahelpers/unified_extractor.go` was broken in the tree, so the package would not compile
locally at all. **On a shared tree, `verify-head-builds` is not a formality — it was the only
working oracle for several minutes.**

**A correction arrived mid-build and changed two numbers in my own docs.** The portfolio lane
re-ran the shell census after its repairs and found seotools had vanished from it while 0 of 7
pages were published. Chasing that turned up a second, worse blind spot that was mine: **the
census did not test `cc.is_active` and my guard does.** Corrected figures: **67 pages / 16 sites**,
of which **48 were already `owned`** — so the guard's genuinely NEW population is **19**, and the
mislabelled-`page_type` risk I had carried as `[UNMEASURED]` is exactly **one page**
(`idea.uk` `/report.html`). The general lesson is in WRONG_CALLS and now in the RUNBOOK: **run
your fix's predicate as the census** — a fix and its denominator disagreeing is invisible while
both look reasonable, and re-running the old query reproduced its number and read as confirmation
when it only reproduced the question I had encoded.

## (k) 2026-09-03 — gate council APPROVED, and the guardian found a test I had claimed but not written

Corr `4e7497ed`: **APPROVED**, 9 objections across the seats, none high. Two mediums actioned in
`5bfc016d7`; one low left open on purpose.

**The one that matters, and it is a defect in my own work rather than in the design.**
`v3_site_actions.go` said *"Pinned by TestToolGateRunsBeforeListingGate"* — **and that test did
not exist.** The guardian seat objected that the ordering should be pinned by an explicit test
"naming both keys armed together, not deferred to 'worth a reviewer's eye'", reasoning from my
prose without any way of knowing the named test was absent. It landed exactly on the gap.
Now written, plus `TestListingGateFirstWouldKeepTheEmptyHub` as the **control**: run the gates the
wrong way round and the hub survives on the strength of children about to be removed, so the
ordering test cannot pass for any order. WRONG_CALLS logged — **a comment naming a test is a
citation, and citations get checked.**

**`reuse_agent` (medium) — why not call `resolveToolPageIdentity`?** Answered in the code so it is
not re-raised: it runs the same naming rule in the OPPOSITE direction (function → page row, one
function at a time, canonicalising to decide where a page WOULD live), whereas this takes a
planned page with no row yet and asks which functions could fill it, from a batch census.
Inverting it costs one query per planned page and a canonicalisation that is meaningless before
the row exists. The shared part is the RULE, bound by lockstep test.

**`architecture` (low) — a pointer from the listing gate's side.** Added, comment only. ⚠ It
breaches the "zero edits to `listing_item_sources.go`" I promised the 444 lane, so I told them
directly rather than letting them find it in a diff. Zero behaviour changed; the promise was about
not touching their mechanism, and a comment does not — but a promise is a promise and the honest
move is to say so.

**LEFT OPEN, deliberately — `bug_historian` (low), and it is the sharpest thing in the verdict:**
nothing PINS the §7 assumption itself. This gate's entire licence to hold pages is "nothing reads
planned tool pages", which is a NEGATIVE finding living in a code comment, not a check. A future
producer that starts reading them would starve **silently**. The one-line disarm (the key) is
today's mitigation; the real answer is a periodic "has a reader appeared" check. **Named as
follow-up work rather than quietly accepted** — and it is the thing to build next if this lane
continues.

## (l) 2026-09-03 — a rehearsal recipe of mine went stale within the hour, which is the point

My RUNBOOK §10 originally recorded a hardcoded "before" md5 (`85b9821d…`) as proof the migration
round trip was byte-exact. Re-rehearsing an hour later produced `d2e44c19…` — **not a regression:
another lane had landed a prompt migration, so the live baseline had moved.** Had I trusted the
recorded value I would have read a healthy round trip as a failure.

Recipe replaced with a **self-baselining** form: capture the baseline into a temp table inside the
same transaction, apply, unwind, and let the query answer `byte_exact` directly. It cannot go
stale because it never records an absolute. Proven twice, at two different baselines. The general
shape — **a check that compares against a recorded constant expires; a check that computes both
sides in one breath does not** — is worth more than the specific recipe, and I had warned the 443
lane about exactly this an hour before doing it myself.

## (m) 2026-09-03, pre-roll — THE DEMAND-CONTROL BASELINE, captured before it becomes unreconstructable

Owner: "a fresh chassis is being built and will be deployed in the next hour." All three commits
(`587666be8`, `5e6fee47b`, `5bfc016d7`) confirmed ancestors of HEAD, so a build from current HEAD
carries both halves. Pods pre-roll: ReplicaSet `75b987cbd7`, started 08:57–08:58Z.

**Captured NOW because after the roll it cannot be reconstructed** — and because a post-fix zero
is not evidence without it (`a-post-fix-zero-needs-a-demand-control`):

| # | metric `[MEASURED 2026-09-03 pre-roll]` | value |
|---|---|---|
| A | `owned_page_review` rows carrying `spec.refusal_class` | **0** |
| B | OPEN `unbuilt_internal_link` items **at a tool-shell page** — *the demand* | **16** |
| C | OPEN items routed to `page-build-handler` at a tool-shell page, all types | **59** |
| D | tool-shell pages (the guard's own predicate) | **67** |
| E | `capability_gap:tool:*` rows (plan-side gate; 729 unapplied) | **0** |

**Why this baseline is worth more than the post-roll numbers.** A is 0 *by construction* — the
`refusal_class` key did not exist before `587666be8`, so ANY non-zero reading afterwards was
caused by this change and nothing else. B and C are the control that makes A's reading meaningful:
there are **16 items actively pointing at shell pages, and 59 across all producer types**, so work
WILL be claimed against these pages after the roll. Without B and C, a post-roll `A = 0` would be
ambiguous between "the guard held nothing because nothing tried" and "the guard did not fire" —
the exact ambiguity that makes an absence unreadable. With them, `A = 0` can only mean the second.

**What to expect, stated in advance so it cannot be rationalised afterwards:**

- **A should become non-zero**, with `refusal_class='tool_pending'`, as queued items are claimed.
- **B and C should FALL**, as those items terminate `wont_fix` rather than rebuilding pages.
- **D should be flat or fall slowly** — this fix repairs nothing; D only falls when the
  portfolio lane's real tools attach (7 of its 8 are built and queued to publish).
- **E stays 0** — 729 is deliberately unapplied until the roll is confirmed.
- ⚠ **No dispatch for ~300 s after the pods restart** — spawns in that window are silently
  dropped, so an immediately-post-roll reading of A is expected to be 0 and means nothing.

**If A stays 0 once dispatch resumes and B/C have not moved, the guard did not fire** and the
first thing to check is whether the image actually carries the commit (§7), not whether the
predicate is right.

## (n) 2026-09-03 12:0xZ — THE ROLL LANDED AND BOTH HALVES ARE LIVE, proven at the artefact

Pods rolled 12:05:45–12:07:16Z onto `v1.0.1358`; `rollout status` reports successfully rolled out.
⚠ Three ReplicaSets were briefly visible at once (`554857f96f`, `5987fcb597`, `59cff58d6b`) — that
was mid-rollout churn, not three versions in service, and reading pod state during it would have
been misleading. Waited for `rollout status` before concluding anything.

**Proven at the artefact, not the tag** — the binary's own statement:

```
build provenance","git_commit":"d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85"
```

```
git merge-base --is-ancestor 587666be8 d0252fd4d  → SHIPPED   (the door)
git merge-base --is-ancestor 5e6fee47b d0252fd4d  → SHIPPED   (the plan-side gate, still keyless)
git merge-base --is-ancestor 5bfc016d7 d0252fd4d  → SHIPPED   (the actioned objections)
```

So "did my fix ship?" was a query, not an inference — exactly what `bugs_open/153`'s stamping
bought. **The tag alone would not have told me this**: `v1.0.1358` names an image, and one release
can straddle several commits.

**⚠ MIGRATION 729 IS STILL NOT APPLIED — and now for a NEW reason.** Its documented preconditions
are both met (council APPROVED `4e7497ed`; a chassis carrying `5e6fee47b` is live). The apply was
**refused by the session's permission classifier**, which is a correct guard on a live-database
write and was NOT worked around. Surfaced to the owner for a decision. Until it applies, the
plan-side gate remains inert: `enforce_tool_sources` is unset, so `enforceToolItemSources` never
runs, and `capability_gap:tool:*` stays at 0. **The door half is unaffected and is live now.**

**Verification deliberately NOT read yet.** At 12:09Z we were inside the ~300 s post-restart
no-dispatch window, where spawns are silently dropped — so a reading of "0 receipts" would have
been an artefact of the window, not a result. I said that in advance in (m) precisely so I could
not rationalise it afterwards. A watch is armed that waits the window out, then reads
receipts / open-link-items / open-build-items against the pre-roll baseline **0 / 16 / 59**.

**What would falsify the fix, stated before the reading arrives:** receipts stay at 0 while the 16
open link items are claimed and the shell pages get rebuilt anyway. The first thing to check then
is NOT the predicate but whether the running binary really carries the commit — which the stamp
above already answers, so a null result would point at the declaration probe
(`refuse_owned_page` on `page-build-handler`) or at those items not being claimed at all.

## (o) 2026-09-03 12:1xZ — MY OWN DEMAND CONTROL WAS WRONG, and it is the third measurement of this shape today

First post-window reading: **0 / 16 / 59**, identical to baseline (m). Before treating that as a
null result I checked whether those items *can* be claimed — and they mostly cannot.

Status breakdown of the "59 open" build items at tool-shell pages:

| status | n | claimable? |
|---|---|---|
| `unresolved` | 26 | **NO — `unresolved` is in `workItemTerminalStatuses`** |
| `needs_human_review` | 24 | NO — the human queue, never auto-claimed |
| `failed` | 8 | **NO — terminal** |
| `triaged` | 1 | yes |

> **CORRECTION to (m): the demand was never 59, and "16 open link items" was the same error.** My
> filter was `status NOT IN ('complete','verified','rejected','wont_fix','cancelled')` — five of
> the six terminal statuses. It **admitted `unresolved` and `failed`**, which ARE terminal
> (`work_items_common.go`, `workItemTerminalStatuses`), and admitted the HITL queue on top. The
> genuinely dispatchable demand `[MEASURED 2026-09-03 12:1xZ]` is **6**, not 59.

**Why this matters more than the arithmetic.** A demand control exists to make a zero readable.
Mine counted rows that will never be dispatched, so it would have licensed exactly the conclusion
it was built to prevent: "59 items were waiting and none was refused, therefore the guard failed."
The truth is that ~53 of them were never going anywhere, guard or no guard.

**Third instance today of one shape** — the census floor (`deployed_at`/`is_active`), the stale
`[MEASURED]` position, and now this. In each case the filter I wrote did not match the concept I
named it after, and in each case the number looked reasonable enough to repeat. The tally is the
point (WRONG_CALLS): **before quoting a count, say the predicate out loud and check it against the
noun** — "open" is not `NOT IN (five statuses)`, it is "a handler will pick this up".

**The corrected control, and what it now licenses:**

- **6** claimable items at shell pages — real demand, small.
- **11** work items created since the roll fleet-wide, **0** of them at a tool-shell page — so the
  write-time door has had **no opportunity to fire yet**. Its zero is not evidence either way.
- **0** `deferred` rows created since the roll — consistent with the above, not a failure.

So the honest statement right now is **"the guard is live and has not yet been exercised"**, not
"the guard did not fire". The watch continues on receipts, which remains the right signal; if the
6 claimable items are picked up and rebuild their pages without a receipt, THAT is falsification.

## (p) 2026-09-03 — the census gap reconciled, and `cc.is_active` found TWICE from opposite directions

The 427 lane verified the narrowing at the source (`save_page_sections_action.go`: one added
condition, `refused && class == refusalOwned`, and the tool arm still live at `load_page_record`,
`multipage_actions` and the rerender escalation), then reported a census disagreement — theirs
58/53/9, mine 67/54/10 — and marked it `[UNRECONCILED]` rather than adopting either. Reconciled
here rather than left standing, because it was my number:

| encoding | pages | sites |
|---|---|---|
| guard predicate, **WITH `cc.is_active`** (mine, and the guard's) | **67** | 16 |
| same, **WITHOUT `cc.is_active`** (theirs) | 58 | 12 |
| **the gap: a tool component EXISTS but is INACTIVE** | **9** | 5 |

58 + 9 = 67, closed exactly. The nine each hold one tool row with `is_active=false`:
ai-agent-orchestration.com ×2, finetuning.uk ×3, gaswholesalers.com ×1,
leopardessconsulting.co.uk ×1, robot-hands.com ×2.

**Which number is right depends on the question, and only one of them is about the guard.**
`toolShellPredicateFor` carries `cc.is_active = true`, so those nine ARE pages the guard refuses;
their predicate would have predicted no refusal for them. For "how many pages does this guard act
on", 67. For "how many pages carry a tool component of any kind", 58.

**The part worth keeping: this same distinction was found twice today, from opposite directions,
by two different lanes.** This morning the portfolio lane's report exposed that my ORIGINAL census
lacked `is_active` while the fix had it — a floor, corrected to match the fix. This afternoon the
427 lane bumped into the corrected version from the other side. Two independent arrivals at one
seam is the strongest evidence available that the guard should keep `is_active` — and it is also
the predicate `create_tool_component` uses for its own "does this site already have this tool"
probe, so the gate and the tool writer agree about what *having a tool* means. Nothing changed on
the strength of it; the point is that the encoding is now corroborated rather than merely chosen.

Their 53-vs-54 half I DID adopt: `build_status='deployed'` versus any non-`removed` row, and the
more inclusive reading is the honest one for "would be refused a repair", since a live-but-not-yet
-deployed row is a page mid-maintenance.

## (q) 2026-09-03 12:4xZ — the drain rate: what IS established, what is NOT, and where I stopped

The 427 lane offered a bound rather than a correction: their two readings 40 min apart did not
move (66/15 both times, db clock 12:00–12:40Z) although tool attachments were running, so the
coupling between attachment rate and shell-set size is looser than "pages leave one at a time".
I tested it instead of accepting it, and the test went three rounds before I stopped it.

**ESTABLISHED, and it is the only part I needed:**

- The shell set is **STABLE at 66–67** across ~40 minutes, on **two independent readers'**
  measurements. So the harm metric's denominator is not moving under me, which is the one property
  the 1.15 writes/hour base rate depends on.
- **ZERO shell pages were CREATED in the last 12 h.** Only 7 `page_type='tool'` pages were created
  at all and all 7 were co-created with their tool. So the set is not being replenished, and
  bug 450 is not currently minting new shells — consistent with the remakes having paused, not
  with anything my fix did (the door half does not stop a page ROW being created; only 729 would).
- **ZERO writes to any shell page since the roll**, which is the harm metric itself.

**NOT ESTABLISHED, and marked rather than resolved: the drain rate.** I measured "39 genuine
repairs in 12 h", then "12 since 12:07Z" — against a set that fell by ONE. That does not close,
and the reason is my own predicate again:

> **`pc.created_at - p.created_at >= 1 hour` does not mean "the page was a shell before this
> insert".** It cannot distinguish a FIRST tool arriving on an old page (a real departure) from a
> tool being REGENERATED on a page that already had one (`create_tool_component_regenerate` —
> never a shell, never in the set). Both look like an old page receiving a tool row. So "39" and
> "12" are upper bounds on departures, not departures.

Measuring it properly needs the page's component state *immediately before* each insert —
`page_component_history` or a time-travel join — and **I stopped rather than build it**, because
nothing depends on the answer: neither lane needs the drain rate, and the denominator stability I
DID need is established directly by two readings of the set itself.

**This is the fourth time today the same shape has caught me** (census `deployed_at`, census
`is_active`, demand control `open`, and now `repair`), and the third time inside one hour of
adopting the 427 lane's rule for exactly it. Writing the rule down does not install it. The tell
each time was identical and cheap: **the arithmetic did not close.** 39 repairs against a set that
fell by 1 is not a subtle discrepancy — it is a factor of 39, and it was visible in the first
result rather than the third. **When two of your own numbers cannot both be true, stop measuring
and go read the predicate.**

⚠ And a smaller one worth copying from them: they caught themselves stamping a UTC reading
`~14:00` — BST written as UTC, on the very number they had just corrected. Every timestamp in this
lane's docs is UTC from the database clock (`now()` returns `+00`), never local.

## (r) 2026-09-03 12:4xZ — the harm metric, properly powered: condition on ACTIVITY, not wall-clock

The portfolio lane reported all 8 tools repaired and added *"your guard has not taken effect yet —
zero `refusal_class='tool_pending'` rows, and the 19 `unbuilt_internal_link` rows touched since
12:00Z are all still `triaged`. Either `587666be8` did not ride that roll or the wave has not
started. Worth checking at the stamp."* Right to say so, and their second branch is the true one:
**the stamp was already checked** — `d0252fd4d` carries `587666be8` (see (n)) — so the code is
live and the wave has not reached these pages.

But their observation exposed a hole in my instrument. If items sit `triaged` and nothing
dispatches, a zero harm reading means "the fleet is idle", not "the guard held". **So measure
whether the fleet is building at all** `[MEASURED 2026-09-03 12:45Z, 0.63 h after the roll]`:

| | since the roll |
|---|---|
| orchestrations started (fleet-wide) | **199** |
| work items status-changed | 234 |
| `page_component_history` writes, ALL pages | **63** |
| ...of those, to a tool-shell page | **0** |

The fleet is busy, so the zero is not idleness. **That converts the test from wall-clock to
activity, which is the right denominator** — it is robust to the fleet being quiet or frantic,
where a per-hour rate is not:

- historical share, 10 d before the roll: **275 shell writes / 17,205 total = 1.60%**
- expected in this window at that share: 63 × 0.0160 = **1.01**
- observed: **0** → p(0 | λ=1.01) ≈ **0.36. STILL UNINFORMATIVE.**
- to reach p < 0.05 needs λ ≥ 3, i.e. **188 fleet writes** — about **125 more**, ≈75 min at the
  current ~100/h.

**So the honest reading is unchanged and the improvement is in the instrument, not the result.**
The watch is re-armed on the activity-conditioned form: it fires immediately on a falsifying shell
write, on a `tool_pending` receipt, or when fleet writes pass 188 — at which point a zero finally
carries information. ⚠ Even then it will read as *"consistent with the guard holding, AND with
nothing having tried"*, because a quiet window cannot distinguish those two; only a receipt can.

## (s) 2026-09-03 — the sectionless fork repairs CLEANER, which inverts my own argument in its favour

The portfolio lane's 8-of-8 completion produced a finding I did not predict and which is a
**stronger** case for holding empty-sectioned tool pages than the one I wrote:

- websitepromotion's `tool-channel-prioritiser` — the SECTIONLESS fork, which parked at HITL and
  never got a shell — now carries **exactly one component: the tool, 29,859 B. No leftover
  `generic-text-block`, no position-2 collision.**
- The seven seotools pages — the SHELLED fork — each carry the tool **and** the shell's prose
  sharing position 2. Debris the repair cannot remove.

> **CORRECTION to my own framing in (m)/PLAN D8.** I argued the sectionless fork is *not*
> harmless — 7 HITL items plus a `needs_content_page` per remake — and used that to justify the
> gate holding empty-sectioned tool pages too. That is right about the cost BEFORE repair and
> wrong about the cost AFTER it. **The variant that looked worse at the time is the one that heals
> cleanly; the shelled variant serves a lie to the public AND leaves permanent debris.** So the
> plan-side gate saves cleanup as well as saving face, and the argument for holding sectionless
> pages is stronger than the one I made — for the opposite reason to the one I gave.

Also from them, and it closes my one open risk: **none of the 8 looks like the `[UNMEASURED]`
case** — a page wrongly typed `tool` whose generic rebuild was genuinely wanted. All 8 were real
tools the briefs asked for. The single fleet-wide candidate for that class remains `idea.uk`
`/report.html`, unchanged.

## (t) 2026-09-03 — the [UNMEASURED] risk is CLOSED, and the answer is not the one I expected

The portfolio lane cautioned that *"six components and no tool is also what a page looks like
after a tool was removed or deactivated, not only what a mis-typed page looks like"* — a
distinction I had not made. `idea.uk` `/report.html` was my sole fleet-wide candidate for the
council's `[UNMEASURED]` misfire class, and I had called it "mislabelled" on a predicate that only
established "typed tool, no tool-level component". Different claims.

**Settled at the rows and then at the artefact:**

- `page_component_history`: **no tool-level row has EVER been attached.** So not a stripped page.
- Its six components: `hero`, `Generic Text Block` ×2, `info-card-grid`, `call-to-action`, and
  **`report-request-form` — a form, at SECTION level.**
- **At the served body `[MEASURED 2026-09-03]`: 1 form, 8 inputs, 2 buttons, 100,896 B.** Control
  in the same run, a known-real tool: advertise's `ab-test-calculator`, 1 form / 11 inputs.
  **Indistinguishable from a real tool by this lane's own test.**

**So the answer to the council's risk is: ZERO pages where the refusal causes harm.** The one
candidate is a fully working interactive page, and refusing a generic rebuild of it is *correct* —
a rebuild would clobber that form, which is migration 164's argument one level along.

> **BUT THE RECEIPT WAS LYING, and that is the real finding.** The old wording said the generic
> builder *"would publish prose about a tool that is not there"*. For this page that is **false**:
> the tool IS there, at the wrong level. An operator reading that receipt would go hunting for a
> missing tool and find a working one. Fixed: the summary and the advice now state the MEASURED
> fact (no tool-**level** component) and leave the consequence conditional, and the advice tells
> the reader to **check the served body before assuming the page is empty**, naming this page as
> the worked example. The test that pinned the old wording asserted the falsehood, so it now
> asserts the opposite — that the summary must NOT claim the tool is absent.

**The shape, for the fourth or fifth time today:** I measured one thing (`no tool-level
component`) and named it another (`mislabelled`, then `prose about a tool that is not there`). The
predicate was right; every sentence I wrapped around it was an inference I had not tested. What
caught it was a peer asking what else could produce the same reading — which is the cheapest
possible form of the check and the one I keep not running on my own claims.

## (u) 2026-09-03 13:3xZ — FALSIFIED. The guard did not refuse a generic write, and I am handing it off undiagnosed.

The watch fired exactly as it was built to. **36 `page_component_history` writes across six of the
seven canonical seotools shells, 13:05:14Z–13:24:36Z**, producer `needs_content_page` →
`page-build-handler`, on pages with `tool rows ever: 0` and `policy: generic`. **Zero
`owned_page_review` rows of any class in that window.** The guard was live: those pods were
`v1.0.1358` / stamp `d0252fd4d`, which carries `587666be8`.

**I checked whether my own metric had gone stale before believing it, and it had not.** After
`29b40e8bc` a re-render write to a shell page is *expected*, so "writes to a shell page" could
have been measuring the wrong thing. It was not: the producer is a generic content builder, not a
re-render, and these are the exact pages this bug exists for. **This is the real thing.**

**What I established, and where I stopped:**

- `page-build-handler` **does** declare `refuse_owned_page: true` — the arm is configured.
- The items were minted **2026-09-03 by `rerender_single_page_action` and `tool-generator`** — a
  producer this lane never accounted for, and NOT `tool-deployer`.
- The writes split **18 with a resolvable `source_item_id` / 18 without**, so there are probably
  **two write paths**, and I have identified neither.

**I did not diagnose it, deliberately.** Both the `load_page_record` arm and the (then-live)
`save_page_sections` arm should have refused; either standing down implies a fail-open I have not
located. Guessing between "the page lookup returned Nil", "the policy read errored", "the write
never crosses `save_page_sections`" and "the claim path skips `load_page_record`" would be exactly
the untested-inference habit that has cost this lane six WRONG_CALLS entries today. **The handoff
states all four candidates as candidates.**

**The consequence that matters most, and it is uncomfortable:** `29b40e8bc` removed the tool arm
from `save_page_sections` on the argument that *every generic path is caught earlier*. §1 is
evidence that argument may be false. If the earlier seams did not catch a `needs_content_page`
build, then I have just removed the backstop that would have. **That is the first thing the next
session must settle, and it is a self-inflicted risk, not an inherited one.**

⚠ Also of note: a new chassis **`v1.0.1359`, stamp `3043885191…`**, rolled at 13:28Z and carries
everything through `b1a3107e6` — including the narrowing. So the window in §1 is the LAST image,
and the current one behaves differently. Re-measure against the current stamp; do not reuse the
13:05Z reading as a statement about what is running now.

## (v) 2026-09-03 — handoff written

`HANDOFF_2026-09-03_continue_here.md` in this directory. §1 is the falsification and leads the
document; §2 is what is live/committed/blocked; §3 separates the numbers that can be trusted from
the ones that cannot; §4 the peer dependencies; §5 the ordered open list; §6 the traps.

## (w) 2026-09-03 14:3xZ — the served-body reading RETRACTS two findings I had already adopted, and the writes turn out to have IMPROVED the pages

The `portfolio_positioning` lane delivered the reading it owed, as observed rather than reasoned,
and it falsifies two of its own earlier findings — **both of which I had taken into NOTES (s) and
acted on.** Recording the whole chain rather than quietly reverting, because the chain is the
lesson.

**What was measured, by byte offset in the served HTML, on all seven seotools pages:**
`hero-tool` ≈51,0xx → `generic-text-block` ≈54,6xx → the tool ≈61,000–63,500. **Hero, prose, tool.
Consistent across all seven.**

**RETRACTED #1 — "the tool and the prose compete for position 2, in an order nothing declares."**
False. Two rows sharing database `position` do not compete in the rendered output; the order is
stable. The error was inferring SERVED order from STORED position — their words, and exactly the
class this lane has been paying for all day, one table along.

**RETRACTED #2 — "debris the repair cannot remove."** False, and the opposite is true. Read on
`tool-robots-txt-tester`: *"The Robots.txt Tester reads that file the way a crawler does, then
shows which of your URLs are blocked, which are allowed, and where two rules contradict each
other… Paste in a URL or the robots.txt content itself…"* That is **accurate copy about a tool
that is present, sitting directly above it.**

**Which means the 13:05–13:24Z writes IMPROVED these pages.** The original prose was written
2026-09-02, before any tool existed, and promised a tool that was not there — this bug's exact
symptom. The rewrite ran *after* the repair put a real tool on the page, so the builder had the
tool in context and described it truthfully. **The only casualty was the `component_id`**, since
repaired.

> **CORRECTION to (s), which was itself a correction.** In (s) I recorded their
> "sectionless-heals-cleaner" finding as *inverting my own argument in its favour* and said so in
> a commit message and to two lanes. **That inversion is now itself retracted.** The shelled fork
> does not leave debris; its prose became an asset once a tool existed to describe. What survives
> is only the narrow point that the sectionless page needed no prose rewrite at all — which is a
> much weaker claim than the one I propagated. **The argument for holding empty-sectioned tool
> pages therefore rests where it originally did: the recurring HITL tax, not post-repair
> cleanliness.**

**The third-order lesson, and it is the one worth carrying.** I corrected my position on the
strength of a peer's measurement (s), propagated the correction, and the measurement was then
retracted by a better one from the same peer. Nothing here was careless — each step was the best
reading available at the time. **The failure mode is propagating a correction as confidently as a
measurement.** A peer's finding is evidence, and evidence gets superseded; my (s) entry stated
their inference as settled fact and pushed it into a commit message and two other lanes within
minutes. The cheap discipline is the one they modelled here: say *observed* or *reasoned* on every
claim that travels, and when you adopt someone else's finding, adopt its status too.

**Also closed:** all eight tools verified serving at 14:32:42Z by instance-scoped id. The instance
work is done on their side; §1's outstanding repair is the two remaining orphans
(`idea.uk/tool-funding-fit`, `loanzy.uk/tool-loan-vs-savings`), not seotools.
