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
