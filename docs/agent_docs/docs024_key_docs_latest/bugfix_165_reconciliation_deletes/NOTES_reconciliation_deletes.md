# NOTES — bugs_open/165 completeness floors

Append-only, newest at the bottom. The missteps are the point.

---

## 1. Picking the lane up (2026-07-31 ~19:30 BST)

Handed the `bugfix_135_prune_floor` handoff. 135 itself is closed and live
(v1.0.1218); the remaining work is `bugs_open/165`, unowned.

Ownership check before touching anything, because `who-owns.py` reads commits and
cannot see a session mid-fix:

- `save_page_sections_action.go` — clean in the tree. **Free.**
- `site_db_actions.go` — clean. Free.
- `populate_nav_tables_action.go` — **dirty** (the `bugfix_149_nav_membership`
  lane). Site B is live territory; leave it.
- Grepped the live `.jsonl` transcripts for the target's code symbols. The 135
  lane's own session (`f0fe4678`) had 102 hits but its last entry was 18:22Z —
  it filed 165 and finished. Nobody else is on it.

Started with site A per the handoff.

## 2. Measuring before choosing cohorts

`165` is explicit that the cohorts must not be guessed, so nothing was written
until the distribution was read. Four things came out of it:

- **Pages are SMALL.** Of 409 pages with agent-writable rows: 178 have 1, 51 have
  2, 98 have 3, max 8. A per-class partition of a 3-row page is not a partition.
- **`slot_name` is 1:1 with a row** — 998 of 1,009 groups hold exactly one. This
  kills the per-slot cohort the bug file itself proposed. 89 legitimate shrinkages
  in 4.5 months would each have been refused by it.
- **`component_id` is the row count in a costume** — 365 of 409 pages have as many
  distinct components as rows. No independent signal.
- **The existing guards are blind on a third of the corpus.** The content-regression
  and interactivity guards both scope to `build_status='deployed'`, and **142 of
  426** pages with components have no deployed row.

Historic false-positive rate for the row cohort, from `page_component_history`
consecutive overwrite pairs: 2,620 transitions, 89 shrank, **4 below 0.5**
(0.15%), 1 below 0.34, 0 below 0.25. [PROXY — the "after" is the next event's
snapshot, which excludes empty-`rendered_html` rows and includes locked rows.]

## 3. THE MISTAKE THAT MATTERS: I read a count as damage

The plan-side cohort simulation said 3 pages would trip. One was
`idea.uk/index.html`: planned 6, "live 2". I wrote that down as evidence the
cohort was **finding real damage** — a homepage stripped to two sections is
exactly the defect this guard exists for.

It was not damage. Opening the rows instead of trusting the count:

```
hero · brief-explanation · tool-list · call-to-action · latest-news · info-card-grid
```

Six components for six planned sections. **Four of them are locked.** My
`live` count applied the agent-writable predicate, so it counted 2 — correctly —
and I read "2" as "the page has 2 sections" when it means "2 sections this save
may touch".

The consequence was in the design, not the prose: my plan denominator was the raw
planned count, so a **perfect** rebuild of that page (6 sections handed over, 4
swallowed by locks, 2 written) scores 2/6 = 33% and is **REFUSED**. A guard that
blocks every rebuild of precisely the pages a human cared enough about to lock.
That is not a false positive to tune away later — it is the failure mode `165`
warns about in terms ("a guard that cries wolf gets deleted by the first person it
blocks"), aimed at the most curated pages on the estate.

Fixed: denominator is `planned − suppressed − locked`. Re-measured: **0 trips on
238 reachable pages** (the 2 remaining are `rebuild_policy='owned'`, refused ~370
lines earlier). Pinned by `TestPageSectionFloorPassesAHealthyRebuildOfALockedPage`,
and mutation M1 confirms that test fails the moment the locked term is removed.

**What would have caught it cheaply:** the same thing that did — reading the six
rows rather than the one number. A count that has a predicate in it answers the
question the predicate encodes, not the question you asked it.

## 4. Smaller wrong turns, recorded because they cost time

- **The step-level consumer census returned 0, then 3, when the answer is 6.**
  `step->>'action_type'` is the wrong key (it is `action`), and even corrected the
  top-level census misses `pageflow-builder` and `page-rebuild`, which nest the
  step inside a loop. The `086/087` landmine says exactly this and I hit it anyway.
  Matching the literal key text `"action": "save_page_sections"` gives 6, which
  agrees with the claims guard's own comment ("six live agents persist sections
  through here") — a cross-check I should have looked for first.
- **First mutation-test pass was invalid and looked fine.** The backup path was
  wrong, `cp` silently failed to restore, and M2/M3/M4 each ran on top of the
  previous mutant. Every one "failed as expected", which is exactly what a valid
  run looks like. Re-ran from a fresh baseline each time; results held, but they
  had not been *evidence* until then.
- **The shared tree would not build,** and it was not mine:
  `discovery_checks/check_empty_sections.go:249: undefined: datahelpers`, another
  session's in-flight edit. The 135 handoff warns about exactly this. Worked around
  with a 12MB `git archive HEAD go.mod go.sum platform internal pkg` tree rather
  than touching another session's file.
- **`/tmp` (shared 16G tmpfs) hit 100% mid-session** and commands began failing
  with ENOSPC. A full `git archive HEAD` needs ~350MB and died half-way. Asked the
  owner rather than deleting other sessions' scratch unilaterally; cleared 74
  scratch dirs belonging to sessions idle >6h (~4.2GB) on his say-so.

## 5. What shipped

`ecf738002` — `save_sections_prune_floor.go` (+17 tests), wired into
`save_page_sections_action.go`. Council `a54172b6-9756-4abc-a9e0-f173ad4de779`,
committed with `Council-Submitted:` since the verdict had not landed.

Not yet proven live: the code is inert until the next chassis roll, and **a green
run proves nothing** — the floor is inert on healthy input by design. Both branches
still need inducing in production, per `165`'s own verification bar.

## 6. Council verdict — APPROVED round 1, and one objection was a real defect

`a54172b6-9756-4abc-a9e0-f173ad4de779`: **approved**, 15 reviewers, 2 abstained,
7 advisory objections, none high. Dispositions, because "approved" is not a reason
to skip reading them:

**FIXED — `recurrenceExpected`, raised as medium by FOUR seats** (`editquality`,
`tooling_provenance`, `debug_historian`, and `guardian` adjacently). They were
right and I had walked straight past a landmine written the same day. The new
emitter reuses a per-page `item_key` through `insertWorkItem`, so it inherited the
anti-churn heuristics: a refusal arriving within 3h of a terminal predecessor is
**dropped entirely**, and the third is born `unresolved` — terminal, and off the
human-review queue the row exists to reach. Both silent, both firing exactly when
a chronically-thin page most needs looking at.

The landmine's own decision rule is "detected defect → leave the flag off; action
request → set it", and applying it honestly settles it: there is **no handler**
(`handler_agent` NULL, status `needs_human_review`), so "the fix is not working"
has no referent. Each refusal is a fresh event about a fresh rebuild. Flag set,
stated at the call site as the landmine requires, and pinned by
`TestPageSectionRefusalSurvivesATwoStrikeHistory` — which supplies a two-strike
history that WOULD brand the item and requires the INSERT to still carry
`needs_human_review`. Verified by clearing the flag and watching it fail (M5).
The emitter now returns its error (call sites discard it with an explicit `_ =`)
purely so the branding is observable from a test at all — unobservable is how the
vacuous-mock landmine on this very helper got written.

**ANSWERED BY MEASUREMENT — `prior_art_librarian`, medium, and the sharpest of
them.** It cited a standing landmine: `pages.sections` is a materialised CACHE and
"the build reads `site_plan_sections`", so my ratchet-breaking cohort might be
measured against stale data. Checked rather than argued:

- Of 115 pages resolvable in both, **114 agree**; **0** have the cache HIGHER than
  upstream. Only cache-higher can cause a false refusal, so the drift that exists
  (1 page, cache lower) is in the permissive direction and harmless.
- `site_plan_sections` resolves for **115** pages against `pages.sections`' **425**.
  Switching the denominator would make the cohort inert on most of the estate.

So `pages.sections` stays, now for a measured reason rather than an assumed one.
Recorded here because the objection was correct to demand the check even though
the check cleared it.

**NO CHANGE — `guidelines`, medium, on `ON CONFLICT` vs `DELETE+INSERT`.**
Verified: `insertWorkItem` does use `ON CONFLICT (site_id, item_key)`
(`load_work_item_actions.go:1230`), so the prose was accurate. The seat's concern
is a property of the shared helper this change reuses, not something introduced
here; the file already carries a comment tying that `ON CONFLICT` clause to
migration 012's index predicate, which is the drift this repo tracks separately.

**OPEN, and honestly open — `guardian`, medium.** "Is a failed `save_page_sections`
step fatal for the six consumers, or swallowed?" A refusal that the pipeline marks
`complete` anyway would add a row nobody reads. Measured what is visible:
`page-build-handler` has `error_step: mark_item_failed`; `page-rerender` and
`tool-recreation-handler` have none; the other three nest the step inside a loop
and do not appear in a top-level census at all. That is not enough to assert
either way, and the induction already planned answers it empirically — induce a
refusal and see whether the orchestration reports failure or reports complete.
**Added to the acceptance test rather than answered here.**

**ALREADY DONE — `bug_historian` and `architecture` both asked that
`bugs_open/165` stay OPEN and tracked** so sites B and C do not fall off the radar
once the highest-stakes one is fixed. It is open, its table row records site A as
guarded-but-unproven, and the `architecture` seat's warning is worth repeating
here: if a future round closes B or C by **copying this cohort shape without its
own measurement**, that repeats the mistake 165 was filed to stop.

**LOW, unactioned — `reuse_agent`** asked whether an equivalent refusal
`item_type` already exists under another name, and `guardian` noted it has no
record of the 2026-07-29 owner ruling cited in the risks (it is in CLAUDE.md,
verifiable by any thread).

## 7. LIVE INDUCTION — both branches proven on v1.0.1223 (2026-07-31 ~22:20 BST)

**Provenance first, because a roll is not evidence.** Both replicas pod-grepped in
one exec each: three markers my change added (`save_refused_incomplete` ×2,
"returned too few sections to replace what is stored", "completeness floor could
not be measured") plus two controls invariant under my own diff (a `lock_helpers`
literal and 135's `index_code_symbols: prune REFUSED`). All present on both.

The second commit (`bb51f5c6e`, `recurrenceExpected`) added **no string literal**,
so it is invisible to a pod-grep — a real provenance hole. Closed indirectly: the
binary carries a literal from `9cc63c775` (committed 20:37 BST), which POSTDATES
`bb51f5c6e` (20:33 BST), and since `make build-*` builds from committed HEAD, my
fix is necessarily an ancestor of what was built. Worth remembering — **a fix
consisting only of a struct field or a control-flow change cannot be pod-grepped
for directly; you have to date it against a neighbouring commit that did add a
string.**

### Four dead ends before the guard could even be reached

The induction itself was easy; *arriving* at `save_page_sections` took five
firings. Recorded because the next person converting sites B and C will hit the
same wall:

1. **`check_rerender_mode` skipped the whole section path.** A plain page-rerender
   goes `check_rerender_mode → render_page` (assemble stored HTML) and never
   touches `save_sections`. The sections path needs
   `input_data.spec.reason ∈ {image_landed, section_data_resolved, cta_links_stale}`.
2. **`check_escalated` stopped it before `save_sections`.** `rerender_page_sections`
   escalates the WHOLE page to the writer if any section's `content_data` is
   missing a schema-required `source:"llm"` field. Two candidate pages escalated.
   The condition is NOT "content_data IS NULL" — I had to mirror
   `missingRequiredLLMFields` in SQL to find a page that would pass.
3. **`save_page_sections` returned `{"skipped": true, "reason": "no page name"}`** —
   it bails at the top, before every guard including mine, when it cannot resolve
   the page name. page-rerender's config reads `input_data.spec.page_name`, which
   a hand-fired payload does not carry.
4. **My first induction idea was structurally wrong.** I planned to add synthetic
   `page_components` rows to inflate the *stored* side. But
   `rerender_page_sections` loads ALL rows for the page and regenerates from them,
   so synthetic rows inflate the numerator too and the ratio stays 1.0 — the floor
   would never have fired. This is "a repro is destroyed by the render" in a new
   costume. Inflating `pages.sections` instead is what works, and it is also
   safer: it touches one jsonb column and no content.

### The results

**REFUSAL branch** — plan inflated 7 → 20, so a healthy 7-section rebuild scores
35%:

```
orchestration: FAILED @ save_sections
save_page_sections: overwrite: REFUSED for page "enterprise-reference-deployment"
  — this run re-confirmed too little of what is stored (prune_floor_ratio=0.50):
  planned sections 35% (7 of 20). NOTHING was deleted; …
site_work_items: save_refused_incomplete | needs_human_review | high
page_components: all 7 rows byte-identical to baseline (md5 + updated_at 2026-07-27)
```

**Three things that only a live run could establish:**

- **The PLAN cohort is what refused. The rows cohort read 7 of 7 = 100% and
  passed.** That is the ratchet scenario, live: with only the row cohort — the
  obvious single-cohort design — this run would have been waved straight through.
  The second cohort is not belt-and-braces; it is the one that fired.
- **`recurrenceExpected` works.** The item landed as `needs_human_review`, not
  branded `unresolved`. That is the four-seat council objection, closed with
  evidence rather than with a unit test alone.
- **The `guardian` seat's open question is ANSWERED: the consumer treats it as
  fatal.** page-rerender has no `error_step` on `save_sections`, and the
  orchestration went to **FAILED**, not `COMPLETED`. So the refusal genuinely
  stops the pipeline rather than adding a row nobody reads. That was the one
  thing I could not answer from config, and it is now answered for this consumer.
  **Still unmeasured for the other five** — see the bug file.

**PASS branch** — plan restored to 7, re-fired:

```
orchestration: COMPLETED
completeness_status: passed   sections 100% (7 of 7), planned sections 100% (7 of 7)
sections_saved: 7             writable_rows: 7   planned_sections: 7   locked_rows: 0
```

The numbers are reported on the SUCCESSFUL save, which was candidate (3) of 135 —
the denominator published beside the count, so `orchestration_states` can answer
"was that rebuild thin?" after the fact.

Cleanup verified: all three pages I touched restored to their exact baseline
plans, row counts unchanged, and **0 pages fleet-wide carry a stray induced
marker**.

### A DEFECT THE INDUCTION EXPOSED, not yet fixed

The refusal text ends: *"the rows this run did not confirm are retained and a
later run that sees the whole corpus will prune them."* **That sentence is false
at this call site.** It is inherited verbatim from `prune_floor.go`'s shared
`Reason()`, where it is true — 135 refuses only the prune, so a later healthy run
does clean up. Here the WHOLE SAVE is refused, nothing is pruned later, and the
page simply does not get rebuilt until someone acts. An operator reading it could
reasonably conclude the situation self-heals. It does not.

Not fixed in this pass **deliberately**: the sentence lives in the shared rule, so
changing it changes 135's message too, and making the clause caller-supplied is a
signature change to a shared mechanism — exactly the class CLAUDE.md says to route
on its own merits rather than ride along inside a bug fix. Filed as a follow-up in
`bugs_open/165`.

## 8. SITES B AND C — 2026-07-31, a contributing lane (not this lane's session)

Written into your NOTES rather than a competing lane directory: you own 165 and
did site A including its induction. Code is in `983e4b0a2`, council
`c69e935a-7134-45c1-81c3-2f1da7831827` (submitted, verdict pending at time of
writing). **Neither site is live and neither is induced.**

### What I got wrong, in order, because that is the useful part

**1. I read a cartesian product as a count, and nearly designed the cohorts on it.**
Sizing nav per site I wrote `FROM sites s LEFT JOIN site_nav_groups g ON
g.site_id=s.id LEFT JOIN site_nav_items i ON i.site_id=s.id … count(i.id)`. It
returned leopardess 92, finetuning 75, ai-agent-orchestration 72 — against a
whole-table total of **184 rows across 16 sites**. Every figure was items ×
groups (leopardess: 4 groups × 23 items). Two child tables joined off one parent
multiply each other and the output still looks exactly like a table of counts.
Caught only because the per-site figures summed past a total I happened to have
taken in the same batch. **An inflated denominator makes a completeness floor read
as passing when it should refuse** — the one direction this guard must not be
wrong in. Logged in `WRONG_CALLS.md`; the fix is a scalar subquery per child.

**2. My first "second unit" for B was wrong, and the data said so twice.** I
started from `pages WHERE (in_header OR in_footer)`, which matched stored nav
items exactly on 14 of 16 sites — and I nearly took that as the denominator. The
two misses were the loan-calculator sites (1 item, 0 flagged pages), and the
reason is that **legal pages bypass the flag check entirely** in
`classifyPagesForNav` (`if legalNames[nameLower] || isLegalPage(nameLower) {
legal = append(...); continue }` runs *before* the `in_header`/`in_footer`
branches). Replaying the full rule — `NOT system AND (legal OR in_header OR
in_footer)` — matched **16 of 16**. A 14/16 match looked like a good enough
signal and was a wrong rule.

**3. I planned to narrow the DELETE, and the data stopped me.** Discovering that
`site_nav_groups` holds a `tools` group no Go code writes, my first instinct was
that the site-wide DELETE was over-broad and should be scoped to the three groups
this action recreates. That is wrong: robot-hands' `tools` item is
`/tools/gripper-safety-factor-calculator`, and `exp_utility` for that site is 9
against 8 stored — the **current classifier already places that page in
`utility`**, so preserving the `tools` group would render it in the nav twice.
The unscoped DELETE is correct; it is a full rebuild. Cost me a design detour and
is now a `LANDMINES.md` entry so the next person does not repeat it.

### The measurement that decided B's cohorts

The bug file asked for per-nav-group cohorts. Replaying the membership rule in
SQL (see RUNBOOK R-B1) gives expected-vs-stored per group:

```
 robot-hands.com  | tools   |  0 |  1 | *** WOULD REFUSE ***
 …all 37 other (site, group) pairs | ok
```

`classifyPagesForNav` **re-homes pages between groups** as a matter of course, so
a per-group cohort scores a legitimate re-homing as a 100% loss of one class and
would refuse robot-hands' nav rebuild **for ever**. Group membership is a
classifier OUTPUT, not an independently-losable class.

Note this is a **different** failure from A's per-`slot_name` refutation. A's
cohorts were too SMALL (998 of 1,009 groups hold one row). B's partition KEY is
not stable. Two distinct ways a partition can be wrong; both invisible without
measuring.

What shipped: `pages seen` (loaded vs existing under `navPageScopeSQL`, the
loader's own predicate, now a shared constant so the two cannot drift) and `nav
items` (to write vs the DELETE's exact complement). **Expected == stored on 16 of
16 sites**, so every cohort reads 100% today and the measured false-positive rate
is 0.

### The generalisation, and why B has no ratchet cohort

You needed a plan cohort for A because `page_components` is **AUTHORED** — a
truncating writer's output becomes the stored baseline, so the row cohort reads
2/2 = 100% for ever. `site_nav_items` is **DERIVED**, recomputed from the page
corpus on every rebuild, so a wrongly-truncated nav is repaired by the next
healthy run. **A derived table self-heals; an authored one ratchets.** That is the
test for whether a future consumer needs a second unit at all, and it is now in
CTXA-025.

### C, and what it honestly cannot prove

One cohort, partition deferred, with the query to run written into the file
header. `link_registry` is **0 rows all-history** (re-verified independently —
matches the `bugfix_092` contribution in the bug file). There is no distribution
to partition on, and guessing one from the SQL's shape is what your own PLAN
decision 3 warns against. Consequence to be honest about: **C's floor is inert by
construction today** (`Stored=0` reads as fully confirmed) and **cannot be proven
by live induction** while the table is empty and `ExtractAndSyncLinksAction` has
0 orchestrations in the retained window. Whether that is closable on tests + roll
is your call, not mine.

### A correction to this lane's own note

Your update block in `bugs_open/165` told the next converter that "`site_nav_items`
carries the same lock columns" as `page_components`. **It does not** — the live
schema is `id, site_id, group_id, parent_item_id, label, url, page_id, item_type,
position, status, metadata, created_at, updated_at`, and a `%lock%`/`%owned%`/
`%writable%` search across the three tables matches only `page_components`.
Corrected in the bug file and in CTXA-025. Flagging *how* it happened rather than
just that it did: it is the one claim in an otherwise thoroughly measured block
that was an inherited plausibility rather than a schema read.

## 8. Another session finished B and C while I was proving A (2026-07-31 ~22:40)

Picked the lane back up to find `983e4b0a2` — sites B and C converted by another
session, who read site A's code and **generalised my private refusal emitter into
a shared `emitPruneRefusalWorkItem` in `prune_floor.go`**, carrying the
`recurrenceExpected` fix and reusing my `floorMockDB`/`expectRefusalItem` test
helpers. That is the system working: they contributed into the case rather than
competing, and said so in their commit.

**First thing I did was check for a regression in the shared rule I depend on**,
because they added 181 lines to `prune_floor.go` and my fix is live and induced.
My whole suite passes at HEAD; `Reason()` is untouched. No regression.

### What I did: retired my duplicate emitter (`21a3f24b1`)

Two near-identical emitters in one package is the drift class the council reviews
for, and `reuse_agent` flagged the risk on round `a54172b6` **before** the
duplicate existed. Theirs is a faithful generalisation, so mine goes; the prose
stays site-specific, which is the split their helper was shaped for.

**The part that is not tidying:** `TestPageSectionRefusalSurvivesATwoStrikeHistory`
now points at the SHARED function, so it guards all three call sites. Proven by
dropping `recurrenceExpected` from the shared emitter in an isolated tree:

```
--- FAIL: TestPageSectionRefusalSurvivesATwoStrikeHistory
     (their TestPruneRefusalWorkItemRoutesToHumanReview stayed GREEN)
```

That asymmetry is the whole point and is worth stating: **their routing test
cannot pin the flag.** With `recurrenceExpected` set the two-strike COUNT is never
issued, so a test that supplies no branding history passes whichever way the flag
goes — the vacuous-mock landmine again, in a new place. Mine supplies the history
that *would* brand the item and pins the INSERT's `status`. So the test had to
MOVE, not be duplicated.

### What I did NOT do, and why

The misleading refusal sentence is now false for **three of four** consumers, not
one — I verified each caller by reading it (`populate_nav_tables_action.go:149`,
`site_db_actions.go:461`, both `return nil, err`), so the original reason for
deferring it has evaporated. But `prune_floor.go` was live territory: that session
committed it at 22:33 and was still writing at 22:41. A same-file collision is the
one failure mode no hook catches — "whoever commits takes both edits". So the
finding went into `bugs_open/165` with the table, the fix shape, and the warning
not to delete the load-bearing half of the sentence.

**The lesson worth keeping: an ownership check has a half-life of minutes here.**
When I started this lane, B and C were unowned and `populate_nav_tables_action.go`
was dirty for the 149 lane. Three hours later B and C were done by a third
session, and the file I most wanted to edit had changed hands twice. Re-run the
transcript grep immediately before the Write, not once at the start — I did, and
it is the only reason this ended as a contribution rather than a collision.

## 9. Closing the guardian's question for all six consumers

The last open item on site A. Answered by extracting each consumer's step config
(text window — the step-level census misses `pageflow-builder` and `page-rebuild`,
which nest it in a loop) and checking history:

- **Four have no `error_step`** → the refusal fails the orchestration. Proven live
  for `page-rerender`; the other three are **[INFERRED]** from the identical
  config shape and the same engine. Marked as inferred rather than stated flat,
  because that is the difference between this section and a claim.
- **Two route on error.** `page-build-handler` → `mark_item_failed` →
  `complete_error`; `tool-recreation-handler` → `complete_workflow`. Measured:
  `page-build-handler`'s one retained run carrying `__step_error` ended
  **COMPLETED**. So the guardian's worry is real for those two — the orchestration
  status does not reflect the refusal.

**But the conclusion is not "two consumers are unprotected", and the distinction
matters.** The floor returns *before* the DELETE, so content is protected in all
six regardless of error handling; and the durable `site_work_items` row is written
*before* the error is returned, so the refusal survives a run that reports
COMPLETED and outlives the ~2-day `orchestration_states` retention.

That is the payoff for a decision that looked like belt-and-braces when I made it.
The refusal's visibility was deliberately NOT left to the step status — 135's
`recordPruneRefusal` reasoning, carried over — and this is the case that shows why:
**for a third of the consumers the step status would have been the wrong place to
look, and nothing in the config would have told a reader that.** Pipeline error
routing is not something a shared guard can assume; making the guard carry its own
surface is what makes it independent of six different workflows' opinions.

Remaining gap, stated so it is not mistaken for done: the three `[INFERRED]` rows,
and confirming the work item lands on the COMPLETED-reporting path. One induction
each, recipe in the RUNBOOK.

## 10. The C caveat resolved itself, and the [INFERRED] rows became facts (2026-08-01)

**My staleness caveat closed the right way.** I flagged the site-C row as "already
stale — re-check it" because another session was mid-refactor of
`enforceLinkRegistryFloor` from `error` to `bool`. It landed (`d1f6a9426`), the
council called it a workaround, and they **reverted** it (`f4825c9ca`), filing
`bugs_open/173` for the underlying disproportion instead. So C is back to `error`
semantics and my original table stands: **3 of 4 consumers fail the whole action.**
Worth noting the caveat cost me two lines and saved a wrong row from propagating
for the twelve hours it was in doubt.

**The three `[INFERRED]` rows are now verified — and NOT by the three inductions I
had written down as the plan.** Reading the coordinator gave the engine's rule:

```
step error -> (1) loop iteration + continue_on_error -> SKIP iteration, continue
           -> (2) error_step (step-level, then config)   -> route
           -> (3) otherwise                               -> failWorkflow
```

Exit (1) was the real hazard and I had not considered it: all three consumers nest
`save_sections` in a loop, so if the loop carried `continue_on_error` a refusal
would skip one page and let the build report success — the silent outcome the
guard exists to prevent, arriving through the guard's own error. Measured: the flag
is UNSET on all three loops, so they take exit (3).

**This is the better verification, and the reason is worth keeping.** Three
inductions would have produced three green samples and told me nothing about a
fourth consumer or about tomorrow's config. Establishing the rule covers every
consumer including ones not yet written, and it would catch a config change that
three passing inductions would sail past. **Prefer the mechanism to the sample when
the mechanism is readable** — the sample is only better when the mechanism is not.

### It also found something I was not looking for

Because `continue_on_error` is unset on those loops, a refusal on ONE page fails
the ENTIRE multi-page build. That is precisely `bugs_open/173`, filed hours earlier
from site C by the B+C lane — and it is not confined to their loop. Four loops
across four agents have the shape, not one. Contributed there rather than
re-diagnosed here.

I also answered the fleet-wide question 173 explicitly listed as *not* done —
whether any live loop sets `continue_on_error: true` and swallows failures it
should not. Census: 20 live loops, 10 unset / 9 true / 1 false; all nine `true`
loops are fan-out dispatch (`call_agent`, `spawn_agent`, `create_work_item`), and
**none of them wraps any of the four floor-guarded actions.** So no floor refusal
is being swallowed anywhere today.

**Measurement gotcha that produced a wrong first answer:** a loop's body lives in
`config.sub_workflow.steps` (18 of 20 loops), not `config.substeps` (2 of 20).
Counting off the wrong key returns `0` for nine loops and reads exactly like "these
loops are empty, nothing to worry about" — the same shape as every other entry in
this file's §2: **the count answered the question my key encoded, not the one I
asked.**

## 9. Sites B and C — council outcome, and an operational datum about the gate itself

**APPROVED at round 3**, `c69e935a-7134-45c1-81c3-2f1da7831827`, 2026-08-01 08:05
UTC. 12 approve, 2 advisory objections, none high. Rounds 1 and 2 both REVISE.

**Round 1** (gated by `debug_historian`): the floor guards NEW loss and cannot see
loss that already happened and stabilised — accepted as a standing limit and
written into `nav_prune_floor.go` rather than argued away. Also `editquality` on
the "no ratchet" wording, `bug_historian` asking for an independent sweep rather
than trusting the header's self-authored candidate list (**run: 18 `DELETE FROM`
statements, exactly 4 are reconciliation deletes, all now guarded — no fifth
site**), and `guardian` asking which workflows call these and whether they handle
a failed step (**answered from the DB and the coordinator source: no `error_step`
anywhere, so `routeToErrorStepOrFail` falls through to `failWorkflow` — the
orchestration FAILS, it does not mark complete**).

**Round 2** (gated by `bug_historian`) is the one worth reading. See the bug file
for the full account; the short version is that I had diagnosed site C's real
problem correctly — its only caller is nested in a loop with no
`continue_on_error`, so a refusal fails the whole site build — and then **routed
around it in the same breath** by making the action never error. Four seats
rejected that. Round 3 reverted to the uniform error contract and filed the gap as
`bugs_open/173`.

**An operational datum for whoever next budgets for the gate.** Round 3's first
run **hung 4 hours in `review_prior_art` and was killed by the reaper**
(`error` column: `reaper: stale EXECUTING_STEP for >4h; step=review_prior_art`),
producing no verdict. A second lane's run died the same way in the same window
(`review_debug_historian`, identical stale timestamp — the reaper batch-stamped
both). A straight resubmit of the *identical* file then completed in ~8 minutes
and approved. So: **a council run that goes quiet for hours is not necessarily
deliberating — check `orchestration_states.error` for the reaper string before
assuming your submission is the problem, and just resubmit.** Note this is a
different failure from `bugs_open/138`'s truncation (`gated_by_truncation` was
false on every round here).

**Still owed, unchanged:** the roll, then both branches induced for B. C cannot be
induced while `link_registry` is empty.

**One loose end left deliberately for this lane rather than raced:**
`save_sections_prune_floor.go` still spells its own `detail` map inline —
`pruneFloorDetail` in `prune_floor.go` is the shared one now (sites B and C use
it). Your `21a3f24b1` retired the duplicate *emitter*; the reporting block is the
remaining half. Not done here because you were active in that file at the time.

## 11. Site B induced live — and the pass branch was free (2026-08-02 ~09:20-10:00 BST)

Picked the lane back up. Two things had changed underneath it: `prune_floor.go`
was committed and the tree clean (`983e4b0a2`), so the other lane's live
territory was released; and the fleet had rolled to **v1.0.1228**, five tags past
site A's.

**Pod-grep first, both replicas.** B and C markers present, site A's present,
`ZZZ_NOT_A_REAL_SYMBOL_ZZZ` at 0. Recorded honestly that `983e4b0a2` is purely
additive so **no same-diff negative control exists** — the discriminating pair
(A's marker would survive a stale image, B/C's would not) does that job instead.

### The design work was choosing the lever, and one option turned out not to exist

I assumed both of B's cohorts were inducible. They are not. `pages seen` compares
the loader's returned rows against a count built from **the same predicate**, and
every nullable column in that SELECT is `COALESCE`d while `name`/`id` are NOT
NULL. So no SQL I can write makes a row unscannable: cohort 1 diverges only on a
genuine driver fault or a race. **That is a property of the design, not a hole in
the test** — but it is exactly the kind of thing that would have read as "I only
bothered to test one cohort" if I had not gone and looked. `[VERIFIED]` against
`information_schema.columns` and the loader's SELECT at
`populate_nav_tables_action.go:249-282`.

### Two safety properties I checked BEFORE inserting anything

1. In `nav-updater`, `refresh_nav_tables` is step 2 and `render_site_components`
   — which bakes nav into header/footer HTML — is step 3. A refusal fails the
   orchestration before anything renders, so synthetic rows **cannot** leak into
   served HTML through this path.
2. `site_nav_items` is fully derived. A guard that failed to fire would delete my
   synthetics and rebuild the correct nav — self-healing, not destructive.

Target chosen on measurement, not convenience: oufe.com had **0 orchestrations in
24h** and no open nav work item. (Checking the pod does not check the queue.)

### Result — refusal branch, first attempt

`FAILED @ refresh_nav_tables`, `nav items 33% (8 of 24)`, all 8 real rows
byte-identical to baseline md5s, all 16 synthetics still present, one
`nav_rebuild_refused_incomplete` item at `needs_human_review` with no handler.
`pages seen` read 9 of 9 and stayed silent — the cohort that catches a classifier
collapse fired alone, which is the one the bug file's proposed partition would
have missed.

Swept for the marker **fleet-wide**, not on the target site — the site A lesson
about abandoned inductions left inflated.

### The thing I nearly got wrong: paying for a pass branch I already had

My plan said "clear the induction, re-run, watch it pass". That would have run
`nav-updater` to completion: re-render site components, commit JS, and create one
rerender work item **per page** — real fleet LLM cost on a live site, for a
*synthetic* pass.

Before spending it I asked whether the pass had already happened. It had:
`site-adoption-agent` orchestration `dcf88c1c-…`, COMPLETED 2026-08-01 09:04 on
loancash.co.uk, `completeness_status: passed`, both cohorts with raw numbers. A
genuine build clearing the guard is **better** evidence than an induced one —
it proves the floor is reached and inert on real traffic, which a synthetic run
cannot.

This is the same move as §10's: prefer the evidence the system already produced
to the evidence you were about to manufacture. §10 read the mechanism instead of
inducing three consumers; this read the record instead of inducing a pass. Cost
one query each time.

### Two findings that came out of the run rather than the plan

- **The misleading refusal sentence is now a production artefact.** oufe's
  operator-facing text ends *"the rows this run did not confirm are retained and
  a later run that sees the whole corpus will prune them"* — false for B, where
  the whole rebuild was refused and nothing is tidied later. Previously filed as
  reasoning; now demonstrated.
- **`100% (1 of 0)`** appears on the loancash pass — the `Stored=0`
  first-appearance case. Correct by design, reads oddly. Cosmetic, recorded so
  nobody diagnoses it as a bug.

### Site C: the induction is impossible, and that is the finding

Re-measured rather than inherited (the claim was two days old): `link_registry`
is **still 0 rows, 0 pages, NULL max(created_at)**, all history, fleet-wide. Its
only live consumer still has no orchestrations. So neither branch is reachable —
`Stored=0` means the floor cannot refuse, and an action that never runs cannot
pass. C is **blocked on `bugs_open/092`**, not skipped, and it carries no risk
meanwhile precisely because it is provably inert.

## 12. The refusal-sentence fix, taken now that its file was released (2026-08-02 ~10:15-10:45 BST)

I had recorded this as blocked: *"not done in this session because `prune_floor.go`
was another session's live territory at the time (they committed it at 22:33 and
were still writing at 22:41)"*. Re-checked before assuming that still held —
`983e4b0a2` is in, the working tree is clean for that file, and the two other live
sessions with `prune_floor` hits in their transcripts turned out to be an
auto-loaded memory index line and an unrelated `rebuild_policy` grep. Released.

**The design decision worth recording is REQUIRED vs OPTIONAL.** The obvious
implementation is a second method, or a default argument, so the four existing
call sites keep compiling. That reintroduces the exact defect: a consumer that
says nothing inherits somebody else's aftermath. The failure mode here is **not a
caller who forgets — it is a default that is silently wrong**, so the parameter is
required and the compiler asks the question. Empty degrades to the shorter
`"NOTHING was deleted."`, which is true for every possible consumer; it never
falls back to a borrowed clause. Ranked by what makes the bad state
unrepresentable rather than by what is smallest, which is the standing rule for
ordering fix candidates.

**The census is complete rather than sampled, and that is a property of the code
not of my diligence:** `Reason` is a method on the *unexported* `pruneFloorVerdict`,
so it is unreachable outside the package and `grep -rn '\.Reason('` cannot miss a
consumer. Four call sites, all updated in the one commit, compile-enforced.

**Mutation, from a fresh baseline each time** (the 07-31 lesson about a wrong
restore path making mutations accumulate while every one "fails as expected"):
M1 restoring the fixed borrowed clause fails BOTH new tests; M2 making the empty
case fall back to a consumer's clause fails exactly the empty-case test. Restore
verified by grep count, package green after. **The negative assertion is the
load-bearing half of the first test** — asserting only that the caller's own
clause appears would still pass if the borrowed sentence were appended beside it.

Submitted to the gate as corr `22cdef56-da93-42f7-b9f9-71ff82abcdf6`, committed
`56365d86b` with `Council-Submitted:` (verdict still in flight at `gate_diagnosis`
as of 10:42 BST).

### Two process things I got wrong in this round, both cheap and both worth the line

**1. I registered the seam one commit late.** The platform-seam rule asks for
concept-register registration *in the same commit that ships the seam*; I shipped
`56365d86b` and registered in `2910f1e0a`. Recorded in the register itself rather
than backdated — saying so is the entire point of the rule.

**2. The `pre-commit` architecture-signal hook fired, and it is a false positive
here.** "exported symbol removed/changed — `-func (v pruneFloorVerdict) Reason(op,
subject, configKey string) string`". `Reason` is a method on an unexported type,
so nothing outside the package can call it. Worth knowing the heuristic keys on
the *identifier's* capital letter and not on the receiver's reachability, because
the advisory reads as an RFC trigger and is not one. (The genuine RFC test, per
the 2026-07-29 owner ruling, is whether the mechanism's GUARANTEE changed. It did
not: every allow/refuse outcome, the cohort arithmetic, the clamping and the
`Stored=0` rule are untouched. This changes what it SAYS.)

### And one I got wrong in the ledger, which became a fleet landmine

Appending to `LANDMINES.md`, I checked my diff was pure-append with
`git diff | grep -c '^-[^-]'` and got `0`. On the concept-register commit minutes
earlier the same check said `0` while `--stat` said **1 deletion** — because a
removed markdown bullet is `-- **thing**` in a diff, two hyphens, which that
pattern excludes by construction. It fails OPEN, and `0` is exactly what a genuine
clean append prints, so there is no tell.

Then, writing that up, I grepped the diff for a foreign symbol to prove nothing
rode along — and counted **context** lines, so an untouched neighbour three lines
above my append read as a passenger. Same trap, opposite direction (false alarm
rather than false all-clear), ten minutes apart.

Filed as a landmine rather than a wrong call: nothing had gone wrong yet, and the
check was being run *because* the work was careful. `LANDMINES.md`, entry
`git-diff-grep-cannot-see-a-deleted-markdown-bullet…`, synced to `doc_notes`.

**A process gotcha found the hard way while filing it:** entries are `###`, not
`##` (`HEADING_RE` in `scripts/landmines_lib.py`). A `##` heading is not parsed as
an entry, so my text was absorbed into the PRECEDING entry and my
`- **footprint:**` line overwrote *theirs* in `doc_notes` — I briefly corrupted
another lane's row. Caught by reading the sync's own output instead of trusting
"applied: 700 rows", fixed by correcting the heading level and re-running
`--apply`, and verified with `--check` showing that entry back on its own
footprints.

## 13. APPROVED — and the medium objection was worth more than the approval (2026-08-02 ~11:00 BST)

Corr `22cdef56`: **approved**, 11 reviewers, 6 abstained, `decided_by: "approved
with 1 advisory objection(s) — none high-severity"`. Full dispositions in
`bugs_open/165`; what belongs here is what the objections taught.

**`editquality` (MEDIUM) caught a real weakness in my evidence, not in the code.**
Its point: my mutation run happened in a package other sessions edit
concurrently, and there **a mutant that breaks the BUILD reads exactly like one
correctly caught**. My run was in fact sound — the failures printed the assertion
messages, not compiler errors — but I had not *established* that, and "it happened
to be fine" is not the same claim as "it was verified". Re-ran in an isolated
`git archive HEAD` tree with each mutant compiled before its failure was trusted.

**The genuinely new evidence was the specificity control**, which I had not
thought to run: mutate the *unrelated* `Disabled` branch and confirm my two tests
still PASS while only the disabled-branch test fails. M1 and M2 show the tests
fail when the thing they guard breaks; **only M3 shows they do not fail when
something else breaks.** Without it "these tests fail on mutation" is compatible
with "these tests are brittle". That is the shape to reuse: a mutation suite
needs at least one mutation it should *survive*.

**A false alarm inside that work, recorded because it looked like a regression.**
The first isolated baseline FAILED — ten `BindSiteExperience`/`DocSubject` tests
plus a panic. Not a broken HEAD: my `git archive` took only `go.mod go.sum
platform internal pkg`, and `doc_subjects_common_test.go` reads
`../../../docs/agent_docs/sql_for_agents` and the experience-register harvest.
A missing fixture, diagnosed in one read of the panic text. **An isolated tree is
only isolated from the right things if you archive what the tests read.**

**`editquality` (LOW) — replaced a grep with the compiler.** "Exactly four call
sites" rested on `grep -rn '\.Reason('`. Better: revert the signature to three
arguments in the isolated tree and let `go build ./...` enumerate the callers. It
named four, repo-wide. Scope stated rather than glossed: `go build` does not
compile `_test.go`, so that covers production callers and the test build covers
the other nine. **When a claim is "these are all of them", prefer the tool that
must be exhaustive over the one that must be spelt right.**

**`bug_historian` — and the check I used to answer it was itself wrong first.**
It asked whether this was a resubmission of an already-adjudicated diff. It is
not: `git log -G` over the signature line returns exactly two commits ever, 135's
and mine. But my first attempt used `-S`, which returned **only 135** — because
`-S` counts *occurrences*, and adding a parameter preserved the searched string,
so the count never moved. Had I stopped there I would have told the council
"nothing has ever changed this signature" while looking at my own commit from an
hour earlier. Third instance today of the same family, now all three on one
`LANDMINES.md` entry: **a check written in a hurry answers a neighbouring
question, and the failure is silent exactly when it fails open.**
