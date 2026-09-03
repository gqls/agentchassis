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
