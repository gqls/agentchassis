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
