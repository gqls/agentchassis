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
