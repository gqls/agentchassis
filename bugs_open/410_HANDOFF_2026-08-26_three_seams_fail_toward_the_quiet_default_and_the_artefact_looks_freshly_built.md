# 410 — three independent seams fail toward the quiet default, complete green, and ship nothing

Filed 2026-08-26 by the `dartsonline_traffic` lane, at the `news_editorial` lane's suggestion —
they found the third instance and declined to fix it inside a feature commit, correctly (035 §6.1's
own scope veto). Two of the three are mine. Nobody owns the seam, which is why it is filed rather
than carried.

**This is a pattern file, not a new investigation.** Each instance is documented where it was found;
what is new is the direction they share and what that implies about which defaults are safe.

## The three, one week, three lanes

| # | seam | what happens | where |
|---|---|---|---|
| 1 | a listing re-rendered in assemble mode after its card image lands | the stored array is re-rendered **verbatim**, empty image fields included. Three re-renders, no change, item `complete` | `bugs_open/384` |
| 2 | a `page_rerender` carrying a reason the Go reader does not know | `keyReason=""`, `scoped=false` → unscoped, **assemble-only**, completes green | `bugs_open/404` |
| 3 | `loadStoredSections`' row scan fails | `logger.Warn(...); continue` — the function returns **fewer sections, or none**, with **no error**, and the page renders empty | verified this filing: `rerender_page_sections_action.go:1206`, scan branch at +32 |

## The property they share, and why it is worse than three bugs

**Every one fails toward assemble/skip — the quiet default — and the artefact is left looking
freshly built.** Not blank, not erroring, not obviously stale: *rebuilt*. A completed work item, a
new deploy stamp, a page that renders.

**That default is correct, and that is the problem.** Assemble-only is the right ordinary behaviour
— it cannot escalate a page to the content writer, and it cannot destroy hand-placed in-body imagery
(a loss this estate has already paid for). `rerender-pages` has produced **6,428** `page_rerender`
items of which **3** carry a reason at all `[MEASURED 2026-08-26]`: assemble is the overwhelming
norm and should be. So **the estate's safest mode is also its silent-failure mode**, and every drift,
every unknown value and every skipped row lands in it. Failing toward *re-resolve* would be
self-announcing — too many re-renders, someone notices. Failing toward assemble announces nothing by
construction.

**The one confirmed instance of items actually taking the silent branch**: 7 of 19 `literal_markdown`
work items predate migration 473, which is what taught the gate that reason. They took
`else_step: render_page` and completed green (`bugs_open/404`, 2026-08-26). Whether those pages were
later repaired by another route is **not** established.

## What this is NOT

Not "add more checks". All three seams sit *downstream* of correct detection — in 384 the asset was
right, in 404 the item was filed correctly, in 3 the query returned rows. **The defect is in the
handling, and the handling reports success.** A checker that watches the artefact catches these only
after the fact and only if it is enabled, which is `bugs_open/399`'s finding one seam along.

## Fix candidates, ordered by what closes the door

1. **Make "I did not understand this" a refusal, not a fallback.** An unknown reason, an unparseable
   row, a mode that cannot satisfy the request — each should fail closed and file, not degrade to
   assemble. The cost is real (a refusal on the fleet's busiest pipeline is loud) and that is the
   point: instance 3 sits on `rerender_page_sections`, so this needs its own review, not a feature
   commit. **This is the door-closing fix and the expensive one.**
2. **Parity between a vocabulary and its readers, asserted at commit time** — `bugs_open/404`'s
   candidate 0, reading the *live* condition rather than a pasted copy, and proven by adding a value
   to the fixture and requiring the test to go red. Narrower than 1, needs no design decision.
3. **Make the silent skips countable.** Instance 3 returns fewer rows than it selected and says so
   only in a log. Returning a count, or erroring when `scanned < selected`, converts an invisible
   loss into a number a caller can assert on. Cheapest of the three; catches the class rather than
   the instances.

## Verification, and the trap in it

Whatever ships, **prove it can fail**: add an unknown reason to a fixture and require a refusal;
make a scan fail and require an error. A test asserting over today's known values passes on the day
it is written and can never do anything else — the shape all three lanes logged separately this week.

## 090 substitution, stated

Not run through the loop. Instance 3 read first-hand at the file:line above; instances 1 and 2 are
documented in their own files with their own evidence, and this file deliberately points rather than
restates. What is NOT established: whether instance 3 has ever fired in production — I verified the
branch exists and reports nothing, not that a scan has failed.

## Relations

`bugs_open/384` (instance 1, fixed at the framework) · `bugs_open/404` (instance 2, latent, candidate
0 waiting) · `bugs_open/399` (the same "detectable rather than unrepresentable" argument one seam
along) · `features_open/035` §6.1 (the scope veto that correctly kept instance 3 out of a feature
commit) · `LANDMINES.md` "a stale PAGE holds every improvement since it rendered"

---

## CORRECTIONS 2026-08-26, hours after filing — an attribution, a citation, and a real reproduction

**1. The 6,428 / 3 figure is NOT mine and my `[MEASURED 2026-08-26]` marker implied it was.**
It was measured and supplied by the **`bugs_open/384` lane**; I took it from a message and stamped it
with the estate's own "I measured this" marker. That marker is supposed to distinguish a figure I
checked from one I relayed, and using it on a relayed number is the precise failure the marker rule
exists to prevent — a reader auditing this file would have come to me for the query and I do not have
it. **Corrected: the control is the 384 lane's, dated 2026-08-26, and I have not independently
re-run it.** The argument it supports — that assemble is the correct and overwhelming norm, which is
exactly why every drift lands there — is unaffected; its provenance is.

**2. `rerender_page_sections_action.go:1206` has already expired. Cite the symbol.**
When I filed this, `:1206` was the `rows.Scan` branch. It is now the `loadStoredSections` **function
signature**; the Warn-and-continue is at `:1238`, moved ~32 lines by `bd811fa93` (*"035 P1:
loadStoredSections reads the composition pair"*) **the same afternoon**. Verified just now.

> **Cite it as `loadStoredSections`' `rows.Scan` error branch** (`logger.Warn("rerender_page_sections:
> row scan failed"); continue`). This file is under active edit from at least two lanes, so any line
> number written here expires — including the one I just corrected it to.

Same family as this repo's standing rule against citing `HEAD~1`: **a reference that moves is not a
reference.** Cheap check before quoting a line: `grep -n "<the distinctive string>" <file>`.

**3. The third seam now has a real reproduction, not a hypothetical one.**
`bd811fa93` added two columns to that SELECT. **Six tests went red reporting *"expected exactly one
section, got 0"* — and not one said *"scan mismatch"***. So a genuine change to a genuine query
presented as **an empty page rather than an error**, and the tests encoded the symptom while losing
the cause. That is the seam firing under a routine, correct edit, which is stronger than the
argument the file otherwise makes from construction alone. Supplied by the `news_editorial` lane.

**What this does to candidate 3** (make the silent skips countable): it moves from *cheapest* to
*best evidenced*. Had `loadStoredSections` returned `scanned < selected` as an error, or even a
count, those six tests would have named the cause on the first run instead of six times reporting
its symptom.

---

## 2026-08-26 (later) — candidate 3 is not a proposal: this package already implements it, for a different reader

Two additions, both verified here rather than relayed — the first correcting how I pitched the
reproduction, the second removing the need to argue for the fix at all.

**1. The seam is invisible to the tests that cover it, not merely unnamed by them.** I wrote that the
six tests "encoded the symptom and lost the cause". The `news_editorial` lane's sharper version, which
I checked: those tests assert **section count and content**, never *rows-in equals rows-out*. So they
would have **passed** on a column change that was wrong in a way that still scanned. Confirmed by
grep across the package's `_test.go` files: the loader's covering tests assert `len(sections) != 1`
and similar, and **no test in that area asserts scan completeness**. That is not "the error message
was unhelpful" — it is a check that cannot come out the other way, the family all three lanes logged
separately this week.

**2. And the guard already exists, ten files away, arrived at the same way.**
`TestAnUnreadableSectionRefusesComponentGrainForTheWholePage`
(`validate_page_content_surface_test.go`) protects `collectPageSections` with **exactly** candidate
3. Its own comment states the defect in terms this file could have borrowed wholesale:

> *"A section whose HTML the canonical reader cannot resolve must NOT be silently dropped from the
> scan while the rest of the page is judged and reported as scanned. That is a partial, invisible
> loss of coverage… **The guard is a COUNT**: fewer sections back than the metadata held means
> something was dropped, whatever the reason."*

and its failure message is this file's whole thesis in one line — *"that page would be scanned in
part and reported as scanned in full"*.

**Three things follow.**

- **Candidate 3 stops being a design proposal.** It is *"apply the count guard this package already
  uses on `collectPageSections` to `loadStoredSections`"* — the same reframing that strengthened
  `bugs_open/399`'s candidate 1. A reviewer asking "why now" gets a precedent, not an argument.
- **It was found by a council objection, not by an incident** (bug_historian, round 1, correlation
  `3ed2b792`). So the estate has already reasoned its way to this guard once *before* it cost
  anything — and the loader simply never inherited it. That is a stronger case for propagating it
  than three post-hoc instances.
- **Copy its mutation check too, verbatim in spirit.** That test carries: *"delete the
  `len(sections) != len(items)` guard and this test must fail. If it still passes, the guard is not
  what is producing the refusal and the coverage hole is back."* A count guard added to
  `loadStoredSections` without that discipline is a check nobody has proven can fail — which is the
  defect this file is about, reintroduced as its own fix.

### Implementation note for candidate 3 — WHICH count, decided before anyone writes it

A count guard on `loadStoredSections` has a wrong version that will get itself removed. Raised by
the `news_editorial` lane from having just done the adjacent change; **verified here**, the loader's
own `WHERE` clause:

```sql
FROM page_components
WHERE page_id = $1
  AND build_status IS DISTINCT FROM 'removed'
ORDER BY position ASC, id ASC
```

**The function drops rows legitimately, in SQL, before Go ever sees them.** So:

- **WRONG: `len(out)` vs "rows in `page_components` for this page".** That comparison counts
  tombstones the query deliberately excluded, so the guard fires on **every page carrying a removed
  component** — a large and entirely healthy population. A guard that fires constantly on correct
  input gets loosened within a week, and a loosened guard is a dead one. This is the failure mode
  that matters, because it ends with the coverage hole reopened *and* a test that looks like it is
  protecting something.
- **RIGHT: rows the cursor yielded vs rows successfully scanned.** Increment a counter per
  `rows.Next()`, compare to `len(out)` at the end, and error (or return the delta) when they differ.
  It needs **no second query**, so it also cannot race a concurrent write the way a re-count would —
  which matters on a shared tree where another lane may be writing the same page.

The distinction in one line: **the guard's job is "did I lose anything the query gave me", not "does
the page have as many components as I expected".** Only the first is knowable inside the function,
and only the first is invariant to legitimate filtering.

Pinned here rather than left to implementation because the wrong version is the intuitive one, is
cheap to write, passes its own tests on a clean fixture, and fails only against real data carrying
tombstones — by which point it is in production and the first response is to relax it.

---

## 2026-08-26 (later still) — CANDIDATE 3 IS SHIPPED for instance 3, with the ratchet that closes the class

Picked up by the `bugs_open/410` lane. This file was filed deliberately unowned (*"Nobody owns
the seam, which is why it is filed rather than carried"*) and the `news_editorial` lane had
declined instance 3 on purpose under 035 §6.1's scope veto. Both correct; neither was competing.
Commit `7c443aac6`, council correlation `c8385154-17b4-43f5-94b2-41f552f43867` (submitted, verdict
pending at time of writing). Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_410_silent_scan_loss/`.

### What shipped

`datahelpers.ScanShortfall(offered, kept, subject)` — one implementation of the pinned count
(rows the **cursor yielded** vs rows **kept**; never a second query). Applied **strictly** in
`loadStoredSections`: any loss is an error. Plus a per-file **count ratchet** over a shared
baseline — a blocking Go source sensor over `platform/orchestration/actions/**` and an advisory
tree-wide twin in `scripts/pattern-check.py`, both reading the same baseline file.

### Four corrections to this file, all verified here rather than relayed

**1. The class is 207, not 225 — and my own 225 was wrong the same way this file warns about.**
`[MEASURED 2026-08-26]` **207** unmarked cursor-loop scan swallows in **127 files** tree-wide;
166 in `platform/orchestration/actions/**`, 41 outside it. My first census counted any `.Scan(`
whose error branch continues, which also catches `db.QueryRow(q, u).Scan(&id)` inside a loop over
some other collection — a single-row lookup where `continue` is control flow over an *item*, not
loss of a *row the database handed us*. **Had that number stood, the ratchet would have pinned
`QueryRow` sites and then fired on them, naming a remedy that is meaningless there** — which is
this file's own "fires constantly on correct input → gets loosened → dies" trap, arriving through
the detector rather than through the data.

**2. `save_page_sections` REPLACES THE PAGE'S ROWS WHOLESALE, so this is destruction, not
degradation — and that decides the strict-vs-graded question.** Verified at
`save_page_sections_action.go:898`: `DELETE FROM page_components WHERE page_id = $1 AND …`. A
section missing from `loadStoredSections`' slice is not merely unrendered; **its row is
deleted**. That is why this reader is strict where `scanBlogArticles` is graded. This file argued
for candidate 3 without this fact, and it is the strongest single argument available for it.

**3. A THIRD precedent, and it is closer than the one §"candidate 3 is not a proposal" cites.**
`scanBlogArticles` (`rebuild_blog_listing_action.go`) already implements this guard **on this
exact `rows.Scan` shape**, with a graded response — `attempted, scanFailures := 0, 0`, per-row
Warn and continue, and an error when every offered row failed (*"refusing to report an empty
listing as 'no posts'"*). And per the `bugs_open/384` lane, its graded branch was **forced by a
gating council objection** (`170147b4`, bug_historian) after a first cut that logged-and-skipped
unconditionally and documented the exposure in prose without closing it. So candidate 3 is not
merely *"a guard this package already uses for a different reader"* — it is **a guard a review
seat has already required twice on this pattern**, which the loader never inherited.

**4. THE SHAPE IS NOT THE DEFECT, which changes what the ratchet can be.** `scanBlogArticles` is
*in* the baseline: it has the swallow shape **and** a correct guard, because the guard lives after
the loop. A ban-style ratchet — the obvious design, and the one the minting ratchet could use
because its population was zero — **would have convicted this estate's best precedent on this very
pattern**. Hence per-file counts plus an at-site `// scan-loss:accepted: <reason>` opt-out.

### The blast-radius question, answered before a seat asked it

`news_editorial` warned that guardian had objected at HIGH severity that morning to touching
`rerender_page_sections` at all without a canary or fast-revert path, with architecture,
render_guardian and debug_historian independently agreeing. Their sharpest form: *"your guard's
first live effect could be to start failing rerenders that today quietly half-work."*

**It is zero, and by schema rather than by luck** `[MEASURED 2026-08-26]`. Every column the loader
projects is structurally incapable of failing a scan on today's data: `id uuid NOT NULL`,
`position integer NOT NULL`, and every other column COALESCEd to non-NULL text or scanned into a
NULL-safe `[]byte` (`content_data`, NULL on 54 live rows, scans to nil). Control:

```sql
SELECT count(*), count(*) FILTER (WHERE position IS NULL), count(*) FILTER (WHERE content_data IS NULL)
FROM page_components WHERE build_status IS DISTINCT FROM 'removed';  -- 2194 | 0 | 54
```

So the guard **cannot** convert a currently-working rerender into a failing one. It speaks only on
a projection/destination divergence — a code defect, introduced by an edit — and it refuses
*before any write*, so the page keeps serving its last good render. ⚠ **This expires if the
projection or the schema changes**; a nullable un-COALESCEd column added to that SELECT falsifies
it. The `news_editorial` lane, who are adding columns there for 035 P1, have accepted this as a
constraint on their change.

### The verification trap this file warns about, discharged by execution rather than by claim

Every guard's mutation was **run**: neuter `ScanShortfall` → its test RED; delete the loader's
call and restore `return out, rows.Err()` → **both** refusal tests RED, and their failure text
reproduces the original defect verbatim (*"returned 1 section(s) with no error"*, *"(0 sections,
nil error)"*); remove a baseline line → *"NEW silent scan loss"*; inflate a count → *"FELL … now
ratchet down"*. The classifier carries its own `StillBites` test, because a source sensor whose
pattern is neutered matches nothing and passes for ever — which looks exactly like a clean tree.

### RESIDUALS — stated here so they do not become this file's own quiet default

1. **CONTENT loss is not covered.** `loadStoredSections` still does
   `_ = json.Unmarshal(cdJSON, &s.contentData)`: on corrupt JSON it **keeps the row and empties
   it**, so `offered == kept` and no count guard can see it. A second silent-loss class in the
   same loop, on a different axis. Commented at the site; needs its own decision about whether an
   unparseable section may render as an empty one.
2. **The 41 sites outside `platform/orchestration/actions` are advisory-only** — no blocking cover.
3. **The ratchet tracks the SHAPE**, so it cannot tell whether an existing count is acted on.
4. **Classifier parity is verified, not enforced** — no test runs the Python classifier from Go.
5. **Candidates 1 and 2 remain open.** Candidate 1 (unknown → refusal) is still the expensive
   door-closing fix and still needs its own review. Candidate 2 is `bugs_open/404`'s candidate 0,
   confirmed unclaimed by the 384 lane and **not** taken here — a parity check would not notice a
   swallowed scan, and a scan ratchet would not notice a reason the DB knows and Go does not.

**This bug stays OPEN**: one of three instances is fixed, the Go half is inert until the chassis
rolls, and the verdict has not landed.

---

## 2026-08-26 (evening) — VERDICT: APPROVED round 1, advisories actioned, and ⚠ the number 410 is now AMBIGUOUS

**`c8385154-17b4-43f5-94b2-41f552f43867` → APPROVED, round 1**, *"4 advisory objection(s) — none
high-severity."* `7c443aac6` is credited via its `Council-Submitted:` trailer. Full dispositions,
the reuse_agent convergence adjudication (one TRUE sibling — `scanBlogArticles`, converge on next
touch; one FALSE sibling — `collectPageSections`, degrade-not-refuse over an in-memory array, and
forcing it onto the helper would be false convergence), the full-column blast-radius measurement
that refuted the debug_historian objection (`2295 | 2295 | 1064 | 10 | 0 | 0` — the NULL-heavy
columns are exactly the COALESCE'd ones), and the after-the-roll verification recipe are in the
lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_410_silent_scan_loss/` (NOTES §evening,
RUNBOOK §12).

⚠ **A SECOND, UNRELATED bug now shares this number** (CLAUDE.md ambiguous-number list, 2026-08-26):
`410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_…` — the content-feed phase-lock case,
different files, different lane, being fixed by its own session. **Resolve by slug; `git log` the
FILE PATH, not the number.** A commit message saying "410" may mean either.

Status unchanged otherwise: **OPEN** — instance 3's fix rides the next chassis roll (verify via
RUNBOOK §12), instances 1–2 and candidates 1–2 belong to their own lanes, and the content-loss
residual (`_ = json.Unmarshal` keeps the row and empties it) still needs its own decision.

---

## 2026-08-26 (post-roll) — INSTANCE 3 IS FIXED AND LIVE, verified at the artefact

Fresh chassis roll, both `agent-chassis` pods probed per the lane RUNBOOK §12 three-way form
`[MEASURED 2026-08-26]`: the guard's capability literal (`refusing the partial result`) is
**PRESENT in both pods on both nodes**, the must-present control passed, the must-absent control
stayed absent. Zero refusals since the roll — with zero rerender traffic in the window, so that
zero is recorded as undemanded rather than as proof; the demand control is the mutation-proved
test suite, which fires the guard on every build.

**Scoreboard for this pattern file, as of 2026-08-26 post-roll:**

| instance | state |
|---|---|
| 1 (`bugs_open/384`, listing never re-rendered) | fixed at the framework — that lane's own record governs |
| 2 (`bugs_open/404`, unknown reason → assemble) | **latent, candidate 0 unclaimed** — confirmed unowned by the 384 lane 2026-08-26 |
| 3 (`loadStoredSections` scan swallow) | **FIXED AND LIVE** — `7c443aac6` + `b93622995`, council APPROVED r1 (`c8385154`), probe-verified in both pods |
| the class (207 sites, 2026-08-26 census) | pinned: blocking ratchet live in every build, advisory twin live on every commit |

**Still open, and why the file stays in `bugs_open/`:** candidate 1 (unknown → refusal, the
door-closing fix, needs its own review), 404's candidate 0, the content-loss residual
(`_ = json.Unmarshal` keeps the row and empties it — invisible to any count guard), and the 41
advisory-only sites outside the blocking package.

---

## 2026-08-31 — the content-loss residual is CLOSED at the worked site, and the guard's zero is now DEMANDED

**The residual** (`_ = json.Unmarshal(cdJSON, &s.contentData)` in `loadStoredSections` — keeps
the row, EMPTIES its content, `offered == kept`, invisible to the count guard) **is fixed**: the
branch now Warns with the row id and DROPS the failing row, so the existing `ScanShortfall`
refuses the whole load. One refusal mechanism, no new error literal. The open question the
comment posed ("may an unparseable section render as an empty one?") was answered by measurement,
not preference `[MEASURED 2026-08-31]`: `content_data` is **jsonb**, so the only reachable decode
failure is a non-object value — and **0 of 2,751** non-NULL values fleet-wide are non-objects
(`SELECT jsonb_typeof(content_data), count(*) FROM page_components GROUP BY 1` → object 2751,
NULL 56). So the refusal fires on **no page the database can currently produce** and exists for
the first writer that changes that. The boundary is exactly as strict as the parse and no
stricter: **55 loadable rows carry SQL NULL** content_data today and stay loadable as nil-map
sections (a live, legitimate population), and jsonb `null` decodes to the same nil map — both
pinned by `TestLoadStoredSections_NullContentDataStaysLoadable`. Both mutations run and killed:
restore the `_ =` form → the new refusal test red with "(2 sections, nil error)"; delete
`ScanShortfall` → all three refusal tests red. Classifier parity re-verified after the edit:
still 207 sites / 127 files, 0 disagreements, the rerender file still 0 unmarked.
Council: **APPROVED r1** (`a69d82f2-9859-4c33-98d9-e791fade2974`, 2026-08-31, all reviewers —
two low advisories, both accepted as trades stated in the submission; the four-seat precedent
ask answered at the artefact: `7c443aac6` itself shipped the deferral comment and bullet).

**The guard's live zero graduated from UNDEMANDED to DEMANDED** `[MEASURED 2026-08-31]`: 209
orchestrations since 2026-08-30 carry the `rerender_page_sections` action in their plan, **87
carry the `rerender_sections` step's output** in `collected_data` (77 COMPLETED + 10 FAILED —
every failure a downstream `OWNED_PAGE_GUARD` refusal or a timeout, none a scan refusal), and the
guard literal probed PRESENT in both pods of today's ReplicaSet `6d6856d8d5` with both controls
clean. So the guard's code path has run ≥87 times at the live tier with zero refusals: scans are
completing, which is the success state. Measurement gotcha for the next reader, recorded so it is
not re-derived: **`orchestration_states.execution_path` is `[]` for this pipeline even on runs
that provably failed at `save_sections`** — count step execution by the step's `output_field` key
in `collected_data`, never by `execution_path`.

**Scoreboard delta vs 2026-08-26:** instance 2 / 404 candidate 0 was picked up by the 404 lane
itself (`ef4236b4d`, 2026-08-26 — one vocabulary definition in `platform/livespec`, all four
readers name it, undeclared reasons now LOUD); the content-loss residual is **closed at the
worked site** (the CLASS — a decode swallowed after a successful scan elsewhere — remains a
stated ratchet blind spot). Still open, and why the file stays in `bugs_open/`: candidate 1
(unknown → refusal fleet-wide, needs its own design round, unclaimed) and the 41 advisory-only
sites outside the blocking package.
