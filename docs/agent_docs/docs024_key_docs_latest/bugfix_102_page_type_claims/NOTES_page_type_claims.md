# NOTES — bugfix 102, page-type-blind claims layer

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-28, session "bugsearch 7"

### Picking the bug

Ran `scripts/who-owns.py` across the unowned-looking candidates. 102 came back
with **no owning workstream** and no commit subject since it was filed on 07-27;
`site_work_items` had nothing open against it; the tree had no uncommitted work
under `datahelpers/claims*`. Its parent lane (`fabricated_stats_043/`) names it
under "**Still owed**" as *a blocker it is not working*, last commit 07-27 13:56.
Took it.

Rejected `104` (the previous bugsearch session's suggested next) because a
`bugfix_104_fleetwide_claim_patterns/` directory with a PLAN dated 07-28 already
exists in the tree — someone started it after that handoff was written. **The
handoff's "checked as genuinely unowned" had a half-life of under a day.**

### The measurement, and how the bug's own premise was too kind

The bug says **"Live exposure today: nil"**. That is true of webdesign.co.uk (no
`evidence_base` row, so its scans are off) and it is **not true fleet-wide**:
nine sites are armed and carry **124** live unregistered-number findings, **61**
of them on editorial page types, and every one of the 61 is a false positive.

The bug also frames the cost as noise ("trains a human to dismiss its findings").
It is worse than that: `unregistered_number` is **`error`** severity in the build
gate and `valid := blockerCount == 0 && errorCount == 0`, so those pages **cannot
be rebuilt**. gamesdesign.co.uk has 40 findings across four blog posts, all of
them probability worked examples.

| page_type | before | after |
|---|---|---|
| blog-post | 46 | 0 |
| content | 38 | 38 |
| report | 14 | 14 |
| adoption-tracker | 8 | 8 |
| tool | 7 | 0 |
| game | 4 | 0 |
| protocol-tracker | 3 | 3 |
| section-index | 2 | 0 |
| news-index | 1 | 0 |
| guide | 1 | 0 |
| **total** | **124** | **63** |

`comm` on the sorted finding lists: 61 suppressed, **0 newly appearing**.

### The correction I made to the bug's own fix candidate

Candidate 1 says "a guide/blog-post page's **prose** scan is either skipped or
graded a rung lower". Taken literally that regresses the case that motivated the
whole check: `check_unverified_claims.go`'s header records that its first live run
found "70+ agents across eight functional departments" **on a guide**. So only the
NUMBER scan reads the surface; `ScanBannedClaims` stays on everywhere. A banned
pattern is human-authored and has no false-positive problem to protect against.

### Missteps

1. **I wrote "verbatim live false positive" over two paraphrased fixtures.** The
   tool and game fixtures in `claims_surface_test.go` were shortened snippets I
   typed from the survey output rather than the live blocks. They failed the
   negative control immediately — they do not flag on ANY surface, so they could
   not have discriminated anything. **The negative control caught my own test
   fixtures, which is precisely the thing it exists for**; if I had only written
   the "editorial raises nothing" direction, both would have passed and pinned
   nothing. Replaced with the real blocks pulled from `rendered_html`.
2. **A test that could not fail.** My first
   `TestBannedClaimsAreStillCaughtOnEditorialPages` looped over seven page types
   calling `ScanBannedClaims(...)` — which takes no surface argument, so the loop
   variable was unused and every iteration was the same assertion. It looked like
   a per-page-type test and was one assertion wearing a costume. Rewritten as the
   discriminating pair: on a guide, one block yields the banned finding and **not**
   the number finding; on a content page it yields both.
3. **Counted a total by glob and got 187.** `survey/*.out` also matches
   `survey/*.fixed.out`, so "before" silently included "after". 187 = 124 + 63 and
   looked like a plausible number, which is what made it dangerous. Recounted per
   site from an explicit list.
4. **A truncated export I nearly compared against a complete one.** One site's
   re-export came back 89 rows against the baseline's 90 (`kubectl exec` printed
   `read message: unexpected EOF` to stderr and exited 0). Caught by diffing the
   row counts before diffing the findings. Had I not, the fix would have looked
   like it suppressed one more finding than it does.
5. **`grep -c` went silent on two output files** — en-dashes in the snippets make
   grep treat them as binary. `LC_ALL=C grep -ac`. (Already in MEMORY as a
   fleet-wide landmine; I hit it anyway, which is what "already known" is worth.)

### The shared tree, mid-change

`bugfix_104_fleetwide_claim_patterns` is editing **the same four files** live —
it makes the banned-claims scan fleet-wide (`datahelpers/claims_global.go`,
`ScanAllBannedClaims`). By 22:57 our edits had interleaved in
`validate_page_content.go`, `check_unverified_claims.go` and `cmd/claimscan/main.go`,
and the combined tree builds and passes.

**Neither change can be committed alone without breaking `HEAD`'s build:** their
half calls `ScanAllBannedClaims`, defined in a file that is still untracked; my
half changes `ScanUnregisteredNumbers`'s signature, which their call sites now use.
`make build-<service>` builds from committed HEAD, so a non-compiling HEAD breaks
every session's build, not just ours. Recorded here because the lesson generalises:
**two sessions editing one file is survivable; two sessions editing one file where
one has an untracked new file is a build-breaking commit for whoever goes first.**

### Council

Round 1 submitted at 22:02Z, `SUBMISSION_CORR=de4a19f5-8f03-4e74-92cb-c23c10ab829d`.
Queue depth at submission: 2 rows in `review_editquality`, positions 5 and 15.

### Council round 1: APPROVED with 3 advisory objections — and two of them changed the code

Verdict at 23:2x: **approved with 3 advisory objection(s) — none high-severity**, 13 seats,
corr `de4a19f5-8f03-4e74-92cb-c23c10ab829d`. An approval is not a reason to skip reading the
objections; two were right and one was already covered.

**1. edit-quality: `blog-index` is an unmeasured extrapolation.** Correct, and it caught the
exact thing my own code comment forbids ("do not widen this from intuition"). I had put
`blog-index` in the editorial set by analogy with `blog-post`. It has **3 pages fleet-wide and
raised zero findings even scanned against an EMPTY register** — so there is no evidence either
way. **Dropped.**

**2. guardian + compliance: check the misclassification risk instead of naming it.** I had
written risk #2 as "a site filing marketing copy under `guide` would go unscanned — check
whether any site does" and then not checked it. So I checked, and it is real one type over:

```
gamesdesign.co.uk | about-index   | section-index
gamesdesign.co.uk | contact-index | section-index
```

**Two of the twenty `section-index` pages are an ABOUT page and a CONTACT page.** An index page
type whose body is marketing copy fails the second half of my own membership bar. Its entire
contribution was 2 false positives (one quoted market-share sentence on robot-hands).
**Dropped** — a blind spot over an about page is the worse trade, and I would rather carry a
known false positive than an unknown blind one.

Re-measured with the narrowed set of five (`guide`, `blog-post`, `news-index`, `tool`, `game`):

| | before | 7 types | **5 types (shipped)** |
|---|---|---|---|
| fleet total | 124 | 63 | **65** |
| suppressed | — | 61 | **59** |

The two that come back are exactly robot-hands' `section-index` pages, as predicted.

**3. debug-historian: no pod-grep verification step.** It was in the RUNBOOK but not in the
submission, so the objection is fair on what the council could see. **And the RUNBOOK's marker
was wrong anyway** — I had named `section-index`, a string that appears in at least four other
Go files (`page_growth_budget.go`, `v3_site_actions.go`, `apply_gap_plan_action.go`,
`populate_nav_tables_action.go`) and which I then removed from the code entirely. It would have
returned a confident `1` on a binary that did not have the fix. Corrected to `resolvePageType`
(a symbol only this change introduces: **0** on the live pod today) with `scanComponentClaims`
(**2**) as the positive control proving the grep method works on that binary.

**Objections I did not act on, and why:** the `reuse_agent` seat asked whether another action
already resolves `page_type` through a fallback chain — grep for `page_record.page_type` outside
`load_page_record_action.go` returns nothing, so there was nothing to reuse. The `guidelines`
seat asked whether the four collected-data reads need an `input_contract` entry; they are
step-to-step data inside one workflow, already covered by `load_page_record`'s own wiring, and
no workflow JSON changes here. The `compliance` seat's residual — *ongoing* monitoring for
future page-type misclassification — is a genuine gap and a genuinely different mechanism;
filed in the bug file as a named follow-on rather than accreted onto this change.

### Live, verified, closed

The owner deployed a fresh chassis: **v1.0.1196**, image id `d8a4a6f0b560` (distinct from
1195's `98ae7405f91b`, so a real rebuild and not a retag). Both pods:

```
resolvePageType        = 2     <- marker, this change's new function; 0 before the roll
ProseNumbersAreClaims  = 5     <- the policy method
scanComponentClaims    = 2     <- POSITIVE CONTROL, pre-existing: the grep method works here
```

**One thing stated as inference, because it is one.** That the *narrowed* five-type set is live
rests on the image being built 23:30:05 BST, **14 minutes after** `955832067` (23:16:12), plus
builds taking committed `HEAD`. I tried to prove it by symbol and could not: the two dropped
page types (`blog-index`, `section-index`) are strings that appear in four other Go files
regardless, and the binary carries **no `vcs.revision` stamp** — `make build-*` uses `git
archive` into a clean context, so there is no `.git` for the toolchain to stamp from. Presence
of the fix is proven; which variant is inferred, and the bug file says so in those words.

`bugs_open/102` → `bugs_closed/102`. Register CLM-016 updated to the shipped numbers, 016b §10
row rewritten, and 016b §9 gained the shared-tree/HEAD pattern from this session.

**Three residuals were filed rather than absorbed**, so the closure is not overstated: the
`report` model-number class, candidate 3 (tutorial framing), and the compliance seat's real
gap — **no ongoing control for future page-type misclassification**, only this one-time check.

### After the close: the misstep that was not about this bug at all

The owner asked whether the page-type misclassification check should be part of the
check-and-fix system fired from the improvement loop. **I had already answered that question
wrongly, unprompted.** Asked how to make my one-time check ongoing, I recommended a bespoke
scheduled SQL sweep, on the grounds that the natural home — a discovery check — "would never
run, because the improvement-sweep has been disabled since 2026-05-02".

Both halves of that were wrong, and the second is the one that matters:

1. **A disabled scheduler is not a dead subsystem.** Measured only after being pushed:
   `completeness-discovery-agent` had produced 144 work items and `design-discovery-agent` 108,
   the latest on **2026-07-25 — three days earlier**, fired by routes other than the sweep,
   including a one-shot `scheduled_tasks` row aimed straight at a discovery agent
   (`oneshot-discovery-aao-20260726`). What is genuinely dead is far narrower and I could have
   said it exactly: **`claims_unverified` items number zero, ever.** I took CLM-004's phrase
   "effectively never runs" — scoped to one check's route — and widened it to the package.
2. **`discovery_checks/` has build-enforced invariants I did not know existed**:
   `handler_coverage_test.go`, `verifier_coverage_test.go`, and `remit.go`'s
   detector-wider-than-handler residue rule. A parallel sweep with its own item type is
   invisible to all three. And `IMP-016` already states the policy for my exact situation — a
   check is enabled once its handler exists, observe-only ahead of that — so I proposed routing
   around a written policy I had not read. My check needs no handler: it is HITL-terminal, the
   same `HandlerAgent: ""` shape as the two checks either side of it.

**The uncomfortable part.** I spent the whole session editing `check_unverified_claims.go`,
which lives in that directory, and never listed its 69 siblings or opened `registry.go`.
**Editing one file in a package is not knowing the package** — the invariants live in the files
you have no reason to open for your own change. Logged in `WRONG_CALLS.md` with the two cheap
checks: `ls` the directory you are standing in, and `grep` the concept register for the
subsystem's own category before proposing anything beside it.

Corrected recommendation, now grounded: a discovery check on
**`completeness-discovery-agent`** (structural lane, ran three days ago) rather than
`quality-discovery-agent` (thematic lane, 7 items ever, nothing since 07-17) — and
`page_type` misclassification is genuinely structural, not claims-specific, since
`page_growth_budget.go`, `apply_gap_plan_action.go` and `populate_nav_tables_action.go` all key
on it too. `IMP-020` is the warning attached: a check written and never added to an agent's
`checks` array has literally never fired.
