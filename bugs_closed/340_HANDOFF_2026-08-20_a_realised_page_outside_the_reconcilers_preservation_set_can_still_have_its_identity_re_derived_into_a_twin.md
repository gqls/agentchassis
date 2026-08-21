# 340 — a realised page outside the reconciler's preservation set is invisible to every identity route, so its twin is still mintable; the mechanism is live and the exploitable population is currently ZERO

**Filed 2026-08-20** by the `bugs_open/215` same-name lane, at the council's
direction: the architecture seat's advisory objection on submission
`27cccfbd-3bf5-4744-a9f6-a5602e38cd30` (APPROVED round 1) was that this gap was
named as out-of-scope prose only, with no tracked item — *"it will be
re-discovered by another incident the same way this one was, unless it gets its
own tracked item now."* That is correct, and it is exactly how `215`'s own
same-name hole surfaced: as a paragraph in a fix account nobody had to act on.

**Verification route, declared per the 2026-07-31 owner ruling:** not through the
090 loop. Substituted first-hand evidence — the predicate is read at HEAD and
cited by symbol below, the mechanism is the one I had just finished building the
neighbouring fix for (so the code path was read line by line, not grepped), and
the population is measured live with the disconfirming case named. A 090 run on
the *adjacent* mechanism in this same file returned UNVERIFIABLE because its
bundle never surfaced `v3_site_actions.go` (`130f187c-7741-42ef-9f55-6a63cd7d1bba`),
so a second run against the same file has a known, specific reason to be
uninformative rather than merely a hopeful one.

## Mechanism

`reconcilePlanWithRealised` (`platform/orchestration/actions/v3_site_actions.go`)
does not reconcile against all realised pages. It first builds a **preservation
set** and then *replaces* its input with it:

```go
if noCurrentPlan || realisedPageCompositionIsPreserved(rm) {
    preserved = append(preserved, rp)
}
...
existingPages = preserved      // every downstream index is built over this
```

- `realisedPageCompositionIsPreserved` — `build_status ∈ {deployed, needs_rebuild}`,
  false when the column is absent.
- `noCurrentPlanFlag` — true only on a site's first plan after adoption.

So on a site that **already has a current plan**, a realised page whose
`build_status` is anything else (`planned`, `building`, `failed`, NULL) enters
**none** of the maps: not `realisedByName`, not `realisedByURL`, and none of the
three twin-identity keys. It is therefore unreachable by every route that mints
`identity_authority="realised"` — the Pass B url-exact snap, the three twin
layers, Pass A's union, and (as of 2026-08-19) the Pass B2 same-name stamp.

`realisedIdentityOf` returns ok=false for the plan entry, both write surfaces
re-derive the page through `CanonicalisePage` at the role's default hub, and
`upsertPage`'s `ON CONFLICT (site_id, name)` misses — **INSERT, not UPDATE.** A
second `pages` row for one page: the same phantom shape `215` documents, reached
by a different door. `honour_realised_identity` cannot help, whatever the site
has opted into.

**It is the same door for `filterOutRecomposePages`.** A page named in
`input_data.spec.recompose_pages` is removed from `existingPages` *before* the
reconciler sees it (`ValidateSitePlanAction`), which is deliberate — the operator
asked for it to be redesigned — but it has the identical identity consequence,
and nothing says so where recompose is documented.

## Severity: real mechanism, currently EMPTY exploitable population

**[MEASURED 2026-08-20, live DB.]** Active pages on sites that already carry a
current plan, whose `build_status` is outside the preserved set:

| | count |
|---|---|
| unpreserved active pages, fleet-wide | **40**, across **13** domains (adversecreditmortgage.co.uk 19, then 3 or fewer each) |
| of those, pages whose stored name is NOT its role prefix — i.e. the ones a re-derivation would actually RENAME | **0** |

**The second row is the finding, and it could have come out otherwise** — the
same query shape over the *whole* active population returns 64 such pages across
9 domains, so the predicate discriminates. Today every unpreserved page is
already a fixed point of the canonicaliser by name, so re-deriving it produces
the name it already has and the upsert UPDATEs. The defect is live in mechanism
and unexercised in fact.

That is why this is filed rather than fixed, and why it is filed rather than left
as prose: the population is one `build_status` transition away from non-empty, and
nothing watches it.

```sql
-- re-run before assuming this is still harmless; the interesting number is the THIRD column
SELECT s.domain,
       count(*) FILTER (WHERE COALESCE(p.build_status,'') NOT IN ('deployed','needs_rebuild')) AS unpreserved_active,
       count(*) FILTER (WHERE COALESCE(p.build_status,'') NOT IN ('deployed','needs_rebuild')
                          AND p.page_type IN ('tool','guide','game')
                          AND p.name NOT LIKE p.page_type || '-%')                            AS would_be_renamed
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status NOT IN ('deleted','archived')
  AND EXISTS (SELECT 1 FROM site_plans sp WHERE sp.site_id = p.site_id AND sp.is_current)
GROUP BY 1 HAVING count(*) FILTER (WHERE COALESCE(p.build_status,'') NOT IN ('deployed','needs_rebuild')) > 0
ORDER BY 2 DESC;
```
The prefix test is an approximation of "not a fixed point" (it misses URL-only
divergence and the homepage/section-index families). The exact test calls
`datahelpers.CanonicalisePage` through the descriptor `write_site_plan_action.go`
actually builds — the LMC lane's harness for that is the worked example, and a
census that does not go through the real canonicaliser will disagree with the
write path.

## Fix candidates, ordered by what closes the door

1. **Give the identity routes a wider input than the composition routes.** The
   preservation set answers *"whose COMPOSITION must a re-plan not redesign?"* —
   `215`'s identity question is a different one, *"who owns this NAME?"*, and it
   has no reason to inherit the narrower answer. Build the identity maps
   (`realisedByName` and the three twin keys) over **all** active realised rows
   while leaving the composition passes on `preserved`. Makes the collision
   unrepresentable rather than unlikely. Cost: the identity routes would then
   fire for pages the reconciler currently ignores entirely, so the flag-off
   inertness argument has to be re-made for that wider set (it should hold — the
   stamped fields are the same — but it must be re-argued, not assumed), and
   `snapPlanPageOntoRealised` carries composition, so the split has to be real
   rather than nominal.
2. **Stamp identity for unpreserved same-name rows only** — the narrow version of
   (1), reusing the 08-19 stamp with a second, wider name index. Smaller blast
   radius, closes only the same-name door, leaves the twin layers as they are.
3. **Refuse instead of re-deriving**: when a plan entry's name matches ANY active
   `pages` row the reconciler did not consider, record it and leave the entry
   alone. Detects rather than fixes, but it is the only candidate that needs no
   inertness argument at all.
4. Do nothing and watch the query above. Defensible **only** while the third
   column is 0, which is not a property anyone is currently maintaining.

Candidate 1 is the structural answer and the one to cost first. Note the
adjacent decision it must not quietly make: an unpreserved page is often an
unbuilt one, and the twin layers already **refuse** a never-shipped stem twin on
the stated ground that a shared stem is not evidence two rows are one page. Any
widening has to keep that refusal, or it trades a phantom for a wrong snap.

## How to verify a fix

Unit: a same-name pair where the realised row is `build_status='planned'` on a
site with a current plan — today no marker is stamped and the entry is
re-derived; after a fix `realisedIdentityOf` returns the stored triple. The
`215` lane's `TestReconcile_SameNameStamp_*` fixtures are the ready template
(they use `deployed` / `needs_rebuild`; this needs the third case). Live: the
third column of the census above stays 0 **and** a deliberately-created
unpreserved non-fixed-point page survives a re-plan without acquiring a twin —
without that second half the census cannot tell a fix from an empty population,
which is the mistake `215`'s own first verification step made.

## Relations

`bugs_open/215` — same phantom shape, different door; its 2026-08-19 same-name
stamp closes the preserved-set case and this file owns the remainder. Register
**PLAN-048** (`docs026_concept_register/register/site-plan-and-reconciler.md`)
names this as the first of its three stated limitations. `bugs_open/204`
(positional slot names on the six decomposed sites — `normaliseRealisedToPlanPage`
carries `sections` verbatim, so a widening must not opt those sites in
implicitly). `bugs_open/266` (archived pages rebuilt by four producers — a
neighbouring "the status predicate is not where you think" defect on the same
table). `features_open/012` recompose (the same identity consequence by operator
request, undocumented there).

---

# CLOSED IN CODE 2026-08-21 — fixed by candidate 2, NOT candidate 1, and the reason is a coupling this file only half-saw

Owner instruction, 2026-08-21: close it properly rather than leave it against a
population that is currently empty. Council submission
**`97542c8c-b628-4947-8d39-c9f7a40dcfb6`**.

## What was built

One new name index, `realisedByNameAll`, built over **every** realised row before
`existingPages = preserved` discards the rest — and consulted by the **same-name
identity stamp alone**, on the arm where the preserved index misses. Pass B2's
existing block becomes an if/else; the preserved arm is byte-unchanged, so every
Pass B2 contract test still exercises the same path it always did.

## Why NOT candidate 1, which this file called "the structural answer"

Candidate 1 was to widen all four identity maps. Writing it exposed the reason not
to, and this file's own text contains the clue without drawing the conclusion: it
warns that *"`snapPlanPageOntoRealised` carries composition, so the split has to be
real rather than nominal"* — and then recommends the option where the split is
hardest to make real.

**The twin layers snap the WHOLE entry.** A hit routes through
`snapPlanPageOntoRealised` → `normaliseRealisedToPlanPage`, which carries the
realised row's `sections`, title, nav and meta. Widen their input to unpreserved
rows and a plan entry can inherit the **empty composition of a page nobody has
built yet** — trading a phantom for a silently emptied page, which is exactly the
damage the 08-20 canary caused by a different route and which this estate has now
paid for twice.

The same-name stamp has no such coupling: it writes identity fields and never
touches `sections`. So widening only what **it** can see is safe in precisely the
way widening the snap layers is not. That is candidate 2, and it closes the door
this file is about. Candidate 1 remains available for the twin layers if anyone
ever measures a population that needs it; it is a different change with a
different argument, and it should not be smuggled in behind this one.

## Inertness is unchanged, which is the claim to attack

No field is written that the stamp did not already write under corr `27cccfbd`:
`identity_authority` (no reader outside `realisedIdentityOf`, whose two call sites
are both inside `if identityPolicy.HonourRealisedIdentity` — enumerated
exhaustively in that round at the `editquality` seat's request), `url` (both write
surfaces derive and overwrite it when not honouring), `page_type` (written back
only where it already equals the writers' own `firstNonEmptyField` derivation,
which the stamp's type-equality precondition guarantees). No new config key, no
writer edit, no new DB read, no change to the twin layers, no change to any
composition pass.

## Mutation evidence — run, not predicted

| mutation | test that went red |
|---|---|
| the wider index is never consulted (340 reverted) | `…_ReachesAnUnpreservedRealisedRow` |
| the wider index is built over `preserved` instead of every row — **the subtle wrong fix** | `…_ReachesAnUnpreservedRealisedRow` |
| the unpreserved arm also imposes composition | `…_UnpreservedRowGetsIdentityButNOTComposition` |
| the from-scratch early return is removed | `…_FromScratchBuildIsStillUntouched` |

Suites green on a `git archive HEAD` shadow tree with only the changed files copied
in — necessary, not ceremonial: another session's in-progress edit to
`render_site_components_action.go` (using `agenterrors` before adding its import)
breaks the working tree's build and would otherwise have masked the result.

## The residue, stated rather than silently closed

When **nothing** is preserved the function still takes its from-scratch early
return and hands the plan back untouched — so a site on which no page is
`deployed` or `needs_rebuild` keeps the pre-340 behaviour. That is deliberate:
every page on such a site is unbuilt, so the twin is a twin of nothing anyone has
served, and the early return is what makes a from-scratch build cheap. **Pinned by
`TestReconcile_SameNameStamp_FromScratchBuildIsStillUntouched`**, so it cannot be
closed by accident or "fixed" without someone deciding to.

## Population, re-measured on the day

Unchanged from the filing: **40** unpreserved active pages across **13** domains on
sites with a current plan, of which **0** would actually be renamed today — against
64 across 9 domains for the whole active population, so the predicate discriminates
and the figure could have come out otherwise. **The defect was live in mechanism and
empty in fact, which is the argument for fixing it now rather than after the
population became non-empty.** The census query in this file remains the way to
re-check.

**Status: fixed in code, council submitted, INERT UNTIL THE CHASSIS ROLLS.** This
file stays open until the roll, per the CLAUDE.md bar — a fix committed but not
shipped is still reproducible. Verify then with the artefact probe (a positive
literal, a one-letter near miss, and an instrument positive) and the census above.

## COUNCIL: APPROVED, corr `97542c8c-b628-4947-8d39-c9f7a40dcfb6` — and four objections answered with evidence rather than prose

*"approved with 1 advisory objection(s) — none high-severity"*, 6 seats abstained.
`bug_historian` objected on record without blocking; `editquality`, `guardian` and
`debug_historian` approved with advisories. All four were checkable, and one of them
improved the file.

**1. `debug_historian` (medium) — "the population measurement scopes blast radius by
`pages.status='active'` without first enumerating `GROUP BY status/build_status`;
`status='active'` is a documented informational-column trap."** The right criticism, and
the enumeration it asked for both **corrects and strengthens** the figure.
[MEASURED 2026-08-21, every page on a site with a current plan]:

| status | build_status | pages | of which a re-derivation would RENAME |
|---|---|---|---|
| active | deployed | 522 | **58** |
| active | needs_rebuild | 36 | 0 |
| **active** | **planned** | **41** | **0** |
| archived | deployed | 28 | 1 |
| archived | needs_rebuild | 14 | 0 |
| archived | planned | 26 | 0 |

Two things follow. The unpreserved-active population this fix newly protects is **41**, not
the 40 I filed — a small correction, and the `would_be_renamed` figure of **0** is unchanged,
so "live in mechanism, empty in fact" stands on a better measurement than it did. And the
`active` scoping turns out to be right **for a reason I had not stated**: `load_existing_pages`
itself ends `WHERE p.site_id = $1 AND p.status = 'active'`, so an archived row never reaches
the reconciler at all. That is a code fact about the loader, not an assumption about what the
column means — which is exactly the distinction the objection was pressing. The 58 in the
first row are the population the *original* same-name stamp already protects.

**2. `editquality` (low) — "if `realisedByName` is built with lower-cased keys and the new
map is not normalised the same way, the wider arm silently never hits and the fix is a
no-op."** The precise shape of a dead guard, and worth checking rather than waving off.
Both maps are keyed on the raw stored `name` and the lookup uses the raw `lm["name"]` —
no `ToLower` on either side, verified by reading all three sites. The seat also noted this
"should self-surface via the pinned `ReachesAnUnpreservedRealisedRow`" test, and it would:
that test asserts the stored triple is honoured, which is only reachable through the wider
map.

**3. `guardian` (low) — "the else-branch runs on every plan reconciliation for every site,
not just the population the bug names."** True, and the bound is worth stating: the branch
is one map lookup on a miss of another map, and it can only *do* anything where a plan
entry's exact name matches a realised row the preservation rules excluded. Where it fires
it writes the same three identity fields the stamp already wrote under corr `27cccfbd`, all
of which are inert to a writer that is not honouring. So the new cost is a hash lookup per
plan page, and the new *behaviour* is bounded by `honour_realised_identity`, which three
sites carry.

**4. `guardian` (low) — "does helper `twinRealisedPage` already exist?"** It does,
`v3_site_reconcile_identity_test.go:22`, and the suites compile and pass; the new
`unpreservedRealisedPage` wraps it.

**5. `bug_historian` (low) — the closed precedents in this exact subsystem were not cited.**
Correct, and they are the right neighbours. `bugs_closed/037` (needs_rebuild pages
unprotected by the replan guard) and `bugs_closed/038` (replan rebuilds every deployed page
and regenerates its content) are both **boundary defects of the same preservation set** this
fix routes around for identity; `bugs_closed/050` (replan may compose pages another
subsystem renders) is the reason Pass B2 has its three-way section branching at all, and so
is the direct reason this fix stamps identity **without** touching composition. Cited here
rather than only in the submission, because this file is what the next reader opens.

**Its medium objection is not an objection to the change**, and is quoted in full because it
states the mechanism better than the filing did: *"Stamping identity onto an unpreserved
realised row, without reconciling composition, makes the (site_id,name) UPSERT correctly
match the real existing page instead of minting a twin. That is the fix's whole point."*

---

# LIVE 2026-08-21 on chassis `v1.0.1322` — verified by ASKING THE BINARY, because this fix adds no string to grep for

The fix rode the roll. **CLOSED.**

## How it was verified, and why the usual probe was useless here

Every previous close-out in this family greps `/proc/1/exe` for a literal the change
introduced. **This change introduced none.** It adds a map (`realisedByNameAll`), a loop and
an `else` arm — Go keeps function names for stack traces but not local variable names, so
there is nothing in the binary to find. Grepping for `realisedByNameAll` would have returned
absent on a pod that *does* carry the fix, which is the worst possible outcome: a false
negative that looks like a rigorous check.

So the route is the one this estate built for exactly this case — **the binary states the
commit it was built from** (`bugs_open/153`, register **BLD-019**), and "did my fix ship?"
becomes a query rather than an inference. Both pods, started 16:54Z, probed 16:59Z while the
startup line was still in range:

```
{"msg":"build provenance","git_commit":"bac1899216fc6406f46cfcf8710f6a74c24276e0"}   (both pods, identical)

git merge-base --is-ancestor a16bd9aea bac189921   -> YES   the 340 fix is in the running build
git merge-base --is-ancestor 6ce422600 bac189921   -> NO    post-stamp commit, correctly not an ancestor
```

**Cite as "live on `v1.0.1322` as at 2026-08-21".**

> **A misstep inside the verification itself, recorded because it is this lane's own
> recurring shape.** My first negative control was a commit I *assumed* postdated the build
> — it did not, so `--is-ancestor` returned true and I briefly read it as the control
> failing. **A control that cannot come out the intended way is not a control**, and I had
> picked one without checking which side of the stamp it fell on. The fix took one command:
> `git rev-list $STAMP..HEAD` names commits that are genuinely after the build (59 of them,
> this tree moving as it does), and any of those discriminates. Caught within the same
> minute and nothing was published on it, so it is recorded here rather than in
> `WRONG_CALLS.md` — but the check is worth stealing: **derive the negative control from the
> stamp, never from your memory of the ordering.**

## Status

**Fixed, council APPROVED (`97542c8c`), live and verified.** The population it protects was
**0** on the day it shipped and the point was always to close the door before it opened —
the mechanism was live and the exploitable set empty, which is the cheapest moment to fix
anything.

**Not behaviourally proven, and deliberately not chased.** No replan has run through it, and
proving it would mean firing one at a site holding an unpreserved page — which buys nothing
here: the same-name stamp it extends *was* proven in production on 2026-08-20 (19 phantoms
→ 0), and this change alters only which rows that proven stamp can see. The residue below
is pinned by tests, not by a live run.

## The stated residue, unchanged

A site on which **nothing** is preserved still takes the from-scratch early return and keeps
the pre-340 behaviour. Every page on such a site is unbuilt, so the twin is a twin of
nothing anyone has served. Pinned by
`TestReconcile_SameNameStamp_FromScratchBuildIsStillUntouched` so it cannot be closed by
accident.

**Re-check the population** with this file's census query before assuming it is still empty:
[MEASURED 2026-08-21, corrected by the council round] **41** unpreserved active pages across
13 domains on sites with a current plan, **0** of which a re-derivation would rename.
