# NOTES — bugfix 463 (append-only, newest at the bottom)

## 2026-09-03, session `463`

**Ownership, resolved before starting.** `scripts/who-owns.py 463` returns OWNED and points at
`gamedesign_uk_rebuild` — but that is because the lane FILED it. The bug file's §8 says
"Unowned … which owns the site and is not taking the fix", and no lane doc anywhere plans work
at Pass C. The ownership check reads commits, so a filer looks identical to a fixer. Resolved
by reading the file, not the tool's verdict.

**The bug is still valid.** Code reads as filed (`git blame` on the Pass C block shows only
comment and counter edits since `f026ad143`, 2026-05-21); gamedesign.uk's `articles-index` is
`deployed` with 0 children; 53 of 78 hubs fleet-wide childless across 21 sites.

**The file went dirty under me, mid-investigation.** An exploration pass reported
`v3_site_actions.go` clean; twenty minutes later `git status` showed 26 uncommitted lines in
`ValidateSitePlanAction` plus two untracked files (37 KB). It was the `428` lane building 463's
own fix candidate 2. Two consequences, both real: I must not build the observability half, and
I must not commit that file while their call site referenced an untracked callee — a pathspec
commit takes a same-file passenger, and that commit would have left HEAD not compiling for
every session. They committed (`eee40b554`, `91173c6d7`) before I needed to, so the constraint
dissolved; had it not, the plan was to do the `datahelpers` half first and wait.

**MISSTEP 1 — my asymmetry census encoded one branch of a two-branch function.** It returned
"5 asymmetric hubs", a believable number I nearly wrote into the plan as live blast radius.
Corrected: 0 of 83. Caught by reading the five rows rather than the count — all were
`/blog.html`-shaped, visibly the fallback branch. Full entry in `WRONG_CALLS.md`. The
uncomfortable part: `bugs_open/463` §3 records its own filer falling into the same class of
trap from the other side, and I had **read that warning** hours earlier. Reading it is not the
same as applying it.

**MISSTEP 2 — a test expectation that was simply wrong, caught by the test.** I asserted that a
re-proposed section index at the hub's own url survives under its proposed name. It does not:
Pass C exempts it by type, and then Pass B snaps it onto the realised identity, so it survives
under the STORED name. Not a Pass C drop at all. Rewritten as its own test asserting
`SnappedRename == 1` and `DroppedCollision == 0` — the distinction between "absent under the
name you looked for" and "deleted" is exactly what this bug is about, so it earned a test.

**MISSTEP 3 — a fixture that measured the wrong rule.** For the "realised entries are not
derived" guard I used `/guides/welcome.html`, which `ValidateRoles` rule 5 retypes to `guide`
before my derivation is reached. The test failed with a url I did not expect, which is how I
found out. Moved to `/articles/`, and pinned rule 5's three directories in their own test since
280+ live blog-post pages sit under `/guides/`.

**THE FINDING THAT CHANGED THE FIX.** Tracing what happens to a page that *survives* Pass C:
both write surfaces discard the planner's url and re-derive it from `CanonicalisePage`, whose
blog-post arm defaults the directory to `blog`. So the Pass C fix alone would have passed its
own tests and changed nothing on the served page — the hub would still resolve zero children
and 444's gate would still hold it. `bugs_open/463` §5 explicitly says this is "NOT about
`parent_section`". True of Pass C, false of the write path. Measured at the live
`agent_definitions` row: the prompt has no such field (32,191 chars, string absent), and 109 of
109 blog-post plan rows carry none. Corrected in the bug file in place.

**Peer review that changed the work, recorded because it did:**
- `428` warned that my "prove the guard by mutation, not by an absent expectation" obligation was
  real, and they were right — the test now asserts both sides of one input. They also declined my
  offer to shape the drop record for their bucket, on the grounds that naming the PASS goes stale
  when passes are renumbered while naming the STAGE does not. Good call; the two records stay
  independent.
- `Portfolio positioning` corrected my routing: 444's gate belongs to the `bugs_open/444` session,
  not to them. They also flagged copyonline.co.uk, released 15:49 today with no plan yet, whose
  brief plans a guides section — a live instance about to hit this exact pair.
- The `feed lane` verified all three of my claims first-hand before acting, then added a producer
  I had not censused: `create_blog_posts_action.go` passes a two-field `PageDescriptor` and never
  threads `ParentSection` at all. Verified: true, and there is no url at that call site to derive
  from, so it is a different fix in a producer dormant since 2026-04-24 (`bugs_open/460`).
  Recorded as a residual in 463 §9 rather than taken.

**`bugs_open/467` filed** — `truncatePreservingRealised` drops every net-new page once the
preserved set reaches 20. 26 of 42 sites are already there. It means this fix must be verified
on a site *under* the cap; gamedesign.uk (9 pages) qualifies.

**Not done, and the fix is not finished without it:** verification at the artefact. Go changes
are inert until the chassis image rebuilds and rolls. The bar is `bugs_open/463` §7's step
boundary (`proposed = survived`) **plus** the children landing at `/articles/<slug>.html` rather
than `/blog/<slug>.html` — the second assertion is the one a Pass-C-only fix would fail, and a
served-page check cannot tell the two apart.

**Tooling note, not a finding:** the Fable model ran out of account credits three times during
this session (HTTP 429), so the two adversarial-review agents I launched on it never returned.
The adversarial work was done by hand instead — a side-by-side simulation of the old and new
Pass C rules across every URL form, the branch-count census above, and the mutation tests.

## 2026-09-03, later — council round 1: REVISE, and the objection was about my submission, not my code

Verdict `revise`, gating objection from `editquality` on edit 1, 16 minutes after dispatch
(submitted 16:51:29Z, report 17:07:05Z — the ~30-minute budget in CLAUDE.md was about right).

> *"Correctness of the narrowed collision test rests entirely on `datahelpers.PagePathKey`'s
> semantics, which are never shown or evidenced. For the fix to both (a) preserve the ratified
> `bugs_closed/141` collision (flat `/news.html` vs `/news/index.html` must still be dropped)
> and (b) NOT collide a child (`/articles/foo.html` vs hub `/articles/index.html`),
> PagePathKey must fold trailing `/index.html`-style hub …"*

**Accepted in full, and it is a fair hit.** The code is right and I had the evidence — the
behaviour table, the tests, the mutation run — but I put a *call* to `PagePathKey` in the sketch
and its *semantics* in prose. The runbook's own rule is that reviewers judge the sketch and a
claim in the rationale is not inspectable, and I broke it on the single function the entire
correctness argument rests on. The reviewer could not verify (a) or (b) without leaving the
submission.

Round 2 puts `PagePathKey`'s verbatim body in the sketch, and adds to `grounded_in`: its own doc
comment, its pre-existing test, both halves of the objection computed on both sides, the full
old-vs-new table across every URL form, and the argument that the table is exhaustive rather
than illustrative (the two predicates partition the space). Nothing else in the plan changed —
resubmitted on the same correlation so the trail accumulates.

**The lesson, which is not "write more prose":** when a fix's correctness reduces to one
helper's behaviour, that helper's BODY belongs in the sketch even though the diff does not touch
it. The diff and the evidence are not the same set of lines.

Worth noting the round cost about fifteen minutes and produced a materially better submission,
which is the standing argument for revising rather than defending. It also did something I did
not expect: the seats ran their own read-only checks and independently reproduced my
109-of-109 `parent_section` measurement, so that figure now has a second, non-me source.

## 2026-09-03, close of session — state, and the one thing still owed

**Committed and green.** `9b540c2e6` (both code halves + tests), `244651c03` (the phantom-hub
and mirror tests), plus the doc commits. `scripts/verify-head-builds.sh` passes on committed
HEAD. Note HEAD was *already* red when I started, from another lane's `83407cd37` — two
declaration tests, unrelated — so run HEAD alone first before attributing a failure to yourself.

**Council: round 1 REVISE (accepted, submission fixed), round 2 in flight** at
`review_reuse_agent` as of 17:17Z on correlation `9f6c6374-1b76-4094-9b4c-e04808d8428c`. The
commit carries `Council-Submitted:`, so `098` credits it automatically once approved; no amend
is needed and none is possible. **Read the verdict and act on it** — the code is already on the
shared branch, so a REVISE or REJECTED is a live obligation, not a formality.

**What is still owed, and it is the only thing:** verification at the artefact. Go changes are
inert until the chassis image rebuilds and rolls, and that is a fleet-wide action this lane does
not take. Both waiting lanes (`gamedesign.uk`, `designblog.co.uk`) have independently confirmed
the live chassis stamp is `30438851…` and that `9b540c2e6` is NOT an ancestor of it, and both are
holding their re-plans until told. The bar, in order:

1. `proposed = survived` at the plan/validate step boundary (463 §7);
2. **the children present in `site_plan_pages` at `/articles/<slug>.html`, not `/blog/<slug>.html`**
   — a Pass-C-only fix passes (1) and fails (2), and the served page cannot tell them apart;
3. only then the served hub, which also depends on `bugs_open/457` (another lane's, in flight).

Use a site UNDER `bugs_open/467`'s 20-page cap. gamedesign.uk has 9.

**Peer outcomes, all resolved.** `428` took my baseline warning into their bug as a blocking
caveat with a demand control, and declined my offer to shape the drop record for their bucket —
rightly: they record the STAGE, which survives passes being renumbered, and I record the PASS.
The `feed lane` filed `bugs_open/468` for the `create_blog_posts` gap rather than leaving it as a
residual in mine, on the argument that a residual stops being read the day its host bug closes;
the two now cross-reference. `designblog.co.uk` corrected their own migration 732 rationale after
my exemption finding — it had cited Pass C exposure for a proposed `tools-index`, which
`isSectionIndexType` exempts, so their surgical route was right for one reason rather than two.
`portfolio_positioning` corrected my routing and is watching copyonline.co.uk, released 15:49Z
with no plan yet, as a possible live instance of the pair.

## 2026-09-04 — the fix is LIVE, and proving it needed a capability probe, not a sha

Council came back **APPROVED at round 2** (17:26:38Z on 2026-09-03), so the round-1 REVISE cost
about twenty minutes and produced a better submission. Nothing further owed there — the commit's
`Council-Submitted:` trailer means `098` credits it automatically.

**Establishing liveness was the interesting part, and every sha-based route failed.**

- The chassis ReplicaSet `agent-chassis-ffc9ddff9` rolled at **2026-09-03T22:07:19Z**, ~5h after
  the fix commit. Pods 13h old, zero restarts.
- **The image tag did not change** — `v1.0.1360` before and after. On this estate that is the
  same-tag-rebuild shape that normally means *no new code*, so the tag is no evidence either way.
- **The `build provenance` startup line had scrolled** — the earliest retained log line was six
  minutes old. An absent grep there means "out of range", not "unstamped", and reading the whole
  log OOM-killed the tool.
- **Every sha probe came back ABSENT**, including the stamp `3043885191b…` a peer had reported
  hours earlier. With only absent results and no present-control, that is indistinguishable from a
  blind probe.

What worked was probing the **capability** with a present-control in the same breath: the
pre-existing literal `dropped flat page colliding with realised section index` is PRESENT (so the
binary is greppable and the probe can see), and the literal this fix adds,
`path collides with realised section index`, is **also PRESENT**. That is the fix, in the running
binary, established without knowing what commit built it.

**A peer's liveness reading EXPIRED rather than being wrong.** Both `gamedesign.uk` and
`designblog.co.uk` independently confirmed on 2026-09-03 that `9b540c2e6` was not an ancestor of
the live stamp, and both have been holding re-plans on it ever since. They were right when they
looked; the 22:07Z roll happened afterwards. The lesson is not "they were careless" — it is that a
liveness check is a measurement with a shelf life, and on this estate the shelf life is one roll.

**Still unexercised.** `[MEASURED 2026-09-04 11:05Z]` **zero** orchestrations carrying a
`validate_plan` step have run since the roll, so the fixed path has not executed once. Verification
needs a deliberate re-plan; it is gamedesign.uk's dispatch to make, and they have been cleared.
