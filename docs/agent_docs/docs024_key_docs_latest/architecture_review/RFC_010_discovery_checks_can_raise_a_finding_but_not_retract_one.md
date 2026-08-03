# RFC 010 — a discovery check can raise a finding but not retract one, and the queue has no other way to learn a finding stopped being true

> ⚠ **`RFC_010` IS AMBIGUOUS — two unrelated papers carry it. Cite this one by SLUG.**
> The other is `RFC_010_who_may_answer_a_page_name_collision.md` (`bugfix_175` lane), which is
> **RATIFIED** and is the one **`CLAUDE.md`'s OWNER RULING 2026-08-02 refers to**. If you
> arrived here from that ruling, you want the other paper.
>
> Cause, stated because it is the fixable part: the number ledger in
> `PROCESS_architecture_review.md` had not been appended to since `RFC_002` — eight papers,
> including both of mine, claimed no number. Ledger restored 2026-08-02. Collisions left in
> place rather than renumbered, matching the `/bugs_*/` convention: never reassigned, resolve
> by slug.

**Filed:** 2026-08-02 by the `bugfix_168_deployed_asset_path` lane, at the owner's direction
("consider how this can be handled properly in the framework"), after a live near-miss whose
proximate cause was eleven work items that had been false for three days and were still
dispatchable.

**Status:** open, design question. **Nothing is proposed for implementation inside a bug
patch** — this changes what a mechanism shared by 50 checks guarantees, which is exactly the
shape `bugs_closed/124` was vetoed for.

---

## The concrete incident, because the abstraction is easy to wave away

`bugs_closed/168` unified the derivation of a deployed asset's path so that the writer and all
six readers resolve through one function. Correct change, council-approved, live on
`v1.0.1229`.

It also **changed what one input meant**. Before it, an `undeployed_asset` work item carrying
`purpose=og_card` would have deployed an image to `assets/images/og_card.png` — a path nothing
references, harmless litter. After it, that same item writes to `assets/images/og-card.png` —
**the live social card**, replaced, by a git commit that runs before any lock or provenance
guard.

Eleven such items were sitting in the queue. Two at status `detected`, which
`triage_detect_items` promotes into the build queue. They had been **false since 2026-07-31**,
when `bugs_closed/142` fixed `check_undeployed_assets` to stop raising them — and they were
false in the strongest sense: every artefact they named was serving HTTP 200 at the time.

The council caught it (`abd9b119`, round 2, gated at high). I had twice told the council the
exposure was unreachable, having measured (a) that the check no longer *raises* such items and
(b) that no *reader* passes such a purpose. Both true. **Neither could see an item that
already existed.**

> **A predicate change stops the tap. It does not empty the bath.**

## The structural claim

**49 of the platform's 50 discovery checks are monotonic: they can create a finding and have
no path to retract one.**

Each check computes, on every run, the current truth set for its site. It files what it finds
and **discards the complement** — the information that would let it close items it previously
raised that no longer reproduce. That complement is free: it is already computed.

There is exactly **one** exception, and it is the model to generalise from.

## Prior art — `check_backend_unreachable` already does this, correctly

`platform/orchestration/actions/discovery_checks/check_backend_unreachable.go:53-72`:

```go
if reachable {
    // Self-clear: resolve any open backend_unreachable item for this site.
    res, uerr := dctx.TX.ExecContext(dctx.Ctx, `
        UPDATE site_work_items
        SET status = 'complete', completed_at = now(), updated_at = now(),
            result = COALESCE(result,'{}'::jsonb)
                  || '{"resolved_by":"backend_unreachable","reason":"health recovered"}'::jsonb
        WHERE site_id = $1 AND item_type = 'backend_unreachable'
          AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved'])`,
        dctx.SiteID)
```

Three things it gets right, and they are the design constraints for any generalisation:

1. **It retracts on a POSITIVE observation, never on an absence.** It clears because the probe
   *succeeded*, not because the check "found nothing". This is the whole safety property. A
   naive "close everything not in this run's findings" would be catastrophic: a check that
   errored, or that was silently blinded by a bug, returns an empty finding set that is
   indistinguishable from a healthy one — the standing landmine that *a gate's 0 findings has
   TWO causes with opposite fixes*.
2. **It is transactional** (`dctx.TX`), so a retraction cannot half-apply against a run that
   then fails.
3. **It records who retracted it and why** in `result`, so the row explains itself later.

What it does **not** get right is shared with everything else here — see the `unresolved` hole
below.

## Measured scale (2026-08-02)

| population | count |
|---|---|
| open, non-terminal work items fleet-wide | **909** |
| of those, older than 14 days | **497** |
| in `detected` (i.e. promotable and dispatchable) | **206** |
| distinct `item_type`s involved | 33 |

Oldest still-dispatchable items by type include `page_rerender` (56, back to 2026-07-14),
`needs_rerender` (21), `phantom_internal_link` (18), `empty_internal_href` (7).

**This is not a claim that 497 items are false.** It is a claim that **nothing in the platform
can tell you how many of them are**, and that the cost of being wrong is set by whatever the
handler does when the item is finally dispatched — which, as 168 showed, is not fixed at the
time the item was raised.

## The `unresolved` hole, which is a second and independent defect

`idx_swi_dedup` excludes `unresolved` from its uniqueness predicate:

```
UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL
  AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved','cancelled'])
```

So the moment an item goes `unresolved`, dedup stops protecting it and **every subsequent
detection creates another copy**. Of the eleven items, **nine were duplicates under two
`item_key`s** — five `favicon`, four `og_card`, all robot-hands.com.

And `check_backend_unreachable`'s self-clear excludes `unresolved` too. So `unresolved` is a
state that:

- is **not terminal** (it still matches "open" in most queries, and it is not in
  `idx_swi_completed`),
- is **not deduplicated** (copies accumulate),
- is **not retractable** (the one self-clearing check skips it),
- but **is** excluded from dispatch by most selectors — so it accumulates silently.

It behaves as a landfill. Whatever is decided about retraction, `unresolved` needs a decision
of its own: either it is terminal (and belongs in the dedup and completed predicates) or it is
open (and must be reachable by retraction). It is currently neither.

## Options, ordered by what closes the door

1. **Give `CheckResult` a way to say what it observed HEALTHY, and let the runner do the
   retraction.** A check populates something like `Resolved []ResolvedFinding{ItemType,
   ItemKey, Reason}`; the runner performs one uniform, transactional `UPDATE` and records
   `result.resolved_by`. **Opt-in**: nothing changes for the 49 until each is edited, so there
   is no flag day and no check is forced to make a claim it cannot support. The runner owning
   the SQL is the point — otherwise the status vocabulary drifts 50 ways, which is the drift
   class this repo keeps paying for. Cost: one field on a shared struct, one runner change,
   then per-check adoption at each check's own pace. **This is `backend_unreachable`'s
   pattern, promoted from an ad-hoc query to a contract.**
2. **Re-validate at dispatch.** Before handing an item to a handler, re-run the predicate that
   raised it and drop it if it no longer holds. Closes the door hardest — a stale item can
   never be *acted on*, which is the actual harm — but checks are not individually addressable
   as predicates today, and it puts a synchronous check-run on the dispatch path.
3. **Detector-version stamping.** Items record the check identity and a version; a version bump
   marks prior items for re-evaluation rather than dispatch. Honest and general, but it
   requires a version discipline that nobody will maintain by hand, and a bumped version is
   itself a claim someone has to remember to make.
4. **A TTL / expiry sweep.** Cheap and generic; also blunt. It would close genuinely-unfixed
   old items, converting a "stale finding" problem into a "silently dropped real defect"
   problem, which is worse.
5. **Manual repair.** What was done for the eleven
   (`bugfix_168_deployed_asset_path/SQL_2026-08-02_retire_stale_brand_head_undeployed_items.sql`).
   Does not scale to 909, and each repair is a fresh opportunity to cancel the wrong rows.

**Recommendation: option 1, with option 2 noted as the thing option 1 does not achieve.**
Option 1 makes retraction *possible and uniform*; it does not make a stale item *undispatchable*,
because a check only clears what it has positively observed healthy. That residual is worth
stating in whatever ships, not discovered later.

## OWNER RULINGS, 2026-08-02

**Decision 1 — how the queue learns a finding stopped being true: OPTION 1 NOW, OPTION 2 LATER.**
A check gains a way to report what it positively observed HEALTHY; the runner performs the
retraction, uniformly and transactionally. Re-validation at dispatch (option 2) follows later.
The two compose: option 1 makes retraction *possible*, option 2 makes a stale item *unactionable*.

**Decision 2 — what `unresolved` means: IT IS OPEN.** Not terminal. So retraction and
deduplication must both be able to reach it, rather than it remaining not-terminal,
not-deduplicated and not-retractable at the same time.

**Standing condition on both:** retraction fires **only on a positive observation of health,
never on an absence of findings**. A check that errored or was silently blinded returns an empty
result indistinguishable from a clean site; getting this backwards would quietly close real
defects fleet-wide.

## What Decision 2 actually costs — measured 2026-08-02, and it is not an index swap

Making `unresolved` open means removing it from `idx_swi_dedup`'s exclusion list. Two findings
change the shape of that job, and both were found by pre-flight rather than during it.

**(a) 87 duplicate rows block the index outright.** The new predicate would newly cover rows
that already violate it:

| | |
|---|---|
| colliding `(site_id, item_key)` pairs | **48** |
| rows involved | **135** |
| rows that must be collapsed first | **87** |

Concentrated in `undeployed_asset` (47 rows / 20 keys), `needs_internal_links` (15/7),
`page_rerender` (15/3), `deactivated_component` (14/3), `needs_sprite_css` (10/1). A
`CREATE UNIQUE INDEX` against this population **fails**. The cleanup is a prerequisite, not a
follow-up — and it is the same "which copy do I keep, and does discarding the others lose a true
finding?" judgement the eleven brand-head items needed, at eight times the scale.

**(b) The index and the Go list must move in a specific order, and one order breaks the fleet.**
`insertWorkItem` interpolates `workItemTerminalStatuses`
(`platform/orchestration/actions/work_items_common.go:37`) into
`ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (…)`. Postgres
infers the target index by requiring the `ON CONFLICT` clause to **imply** the index predicate.
That file's own comment records that this has already broken the fleet once, when migration 157
added `cancelled`.

Working the implication both ways gives an asymmetry that decides the sequencing:

| order | inference | what happens |
|---|---|---|
| **Go first** (drop `unresolved` from the list, index unchanged) | `NOT IN (6)` does **not** imply `NOT IN (7)` — a row could be `unresolved` | ❌ **42P10 on every keyed insert, fleet-wide.** This is the breakage of 157, repeated. |
| **Index first** (drop `unresolved` from the index, Go unchanged) | `NOT IN (7)` **does** imply `NOT IN (6)` — inference still succeeds | ⚠️ Inference is fine, but Go no longer treats an `unresolved` row as a conflict target, so an insert colliding with one **raises 23505 instead of deduping** |

So Go cannot move first. Index-first is inference-safe but opens a window in which precisely the
inserts this change exists to fix (a new detection colliding with an `unresolved` row) hard-fail
instead of upserting — and those are not rare: `undeployed_asset` alone has 20 such keys today.

**Therefore Decision 2 is: collapse the 87 → change the index → roll the binary promptly**, with
a known and deliberately short window. It is a coupled schema+binary change on the insert path of
every work item in the estate. It wants the council gate, a migration reviewed on its own terms,
and someone watching the roll — not an improvised evening.

## DECISION 1 IS IMPLEMENTED — council `846f4f3d`, APPROVED at round 1

Shipped 2026-08-03: `CheckResult.Resolved`, `resolveWorkItems`, runner wiring,
`check_backend_unreachable` converted, four guards (five mutation proofs). Registered as
**WII-009**. Approved "with 5 advisory objection(s) — none high-severity"; four were checkable
and were checked rather than filed.

**`editquality` (medium) was right about an overstated claim.** My submission said a sqlmock
test "pins the query's three load-bearing predicates". It did not — it matched only
`UPDATE site_work_items` and the arguments, so dropping the status filter, the narrow/wide
switch or the batch guard would all have left it green, none of them being an argument. The
test now asserts the predicates in the **query text**, and both the status list and the batch
guard are mutation-proven.

**`prior_art_librarian` (medium) asked for the inventory behind "49 of 50".** Attached, and it
is now stronger than the claim: at `7efc891f6~1`, **exactly one** of 50 check files contained
`UPDATE site_work_items` (`check_backend_unreachable.go`); after the conversion, **zero** do —
every retraction goes through the seam.

**`guardian` (low) asked whether `CheckResult` crosses a wire boundary.** It does not — no
`json` tags, never marshalled. Worth recording the near-miss the seat's question exposed:
`internal/adapters/browserrunner/run_checks_action.go` declares an **unrelated type of the same
name** which *is* serialised (`Results []CheckResult \`json:"results"\``). A future reader
grepping `CheckResult` will find a wire type that is not this one.

**`guardian` (low) asked for the call-path exposure, and the answer is larger than "one
check".** Five active agents invoke `run_discovery_checks`: `completeness-discovery-agent`,
`design-discovery-agent`, `quality-discovery-agent`, and — unexpectedly — `council-gate` and
`fix-proposer`. The retraction loop runs on all five call paths. Its *data* effect remains nil,
which is the next finding.

## THE SEAM SHIPS INERT, AND SAYING SO IS THE POINT

Measured 2026-08-03, and it corrects my own framing: I told the council the blast radius was
"one check", which implies one check will exercise it.

**Zero will.** `backend_unreachable` — the seam's only adopter — is enabled by **0** active
agent definitions and has produced **0** work items in all history, against 1 VM-target site.
So nothing retracts anything until a check is enabled or a second one adopts.

That is not a defect: opt-in with an unsafe default OFF is what the owner's ruling asked for,
and a mechanism with no callers is the *expected* first state of an opt-in seam. But it means
the honest status is **"built, approved, and undriven"**, not "working" — and this estate has a
standing lesson that a silent mechanism is usually UNDRIVEN rather than missing, which is
exactly the trap a future reader would fall into on seeing `items_resolved: 0`.

**So the next step for Decision 1 is adoption, not more mechanism.** The cheapest honest first
adopter is a check that already computes a positive observation and throws it away. Choosing
one is a separate piece of work, and it should measure the retraction on a real site before
claiming the seam works.

## DECISION 1 HAS A FIRST REAL ADOPTER — `check_empty_sections`, 2026-08-03

Committed `2287606d1` + hardening `27891fab8`; council `97923026-2b2d-4925-b9a3-de6f70c49d2b`
**APPROVED at round 1** (15 seats, 3 advisory objections, none high). **LIVE AND PROVEN on `v1.0.1243`, both replicas,
2026-08-03.** A `completeness-discovery-agent` sweep of `leopardessconsulting.co.uk`
(correlation `4401d952`) **retracted 4 `empty_section` items raised 2026-04-14 / 04-23** — over
three months stale, and unclosable by anything in the platform until now. Fleet-wide
`result ? 'resolved_at'` went **0 → 4**. **And the control that matters: the same sweep left 6 of
that site's 10 empty_section rows open** (3 `unresolved`, 2 `needs_human_review`, 1 `detected`),
so it closed what it had evidence for and nothing more. The paper's central claim — that the
complement is free information the checks were throwing away — is now demonstrated rather than
argued.

**OWNER RULING 2026-08-03 — the two-strike interaction is ACCEPTED AS-IS, and tracked as a
follow-up rather than fixed here.** A retraction writes `complete` onto an existing row and so
feeds `insertWorkItem`'s two-strike counter identically to a handler fix. Three council seats
(`guardian` medium, `improvement_guardian` low, `bug_historian` low) independently asked for a
human decision. Ruled: accept, because it is measured empty today (0 of 17) and `insertWorkItem`
sits on the insert path of every work item in the estate — changing it from inside a check
adoption is exactly the `bugs_closed/124` shape the guardian seat exists to veto. Exempting
retractions from the counter remains arguably correct (the counter breaks discover/fix loops,
and a retraction is evidence the loop is NOT stuck) and is now open question **Q1** below.

`check_empty_sections` now retracts `empty_section` findings whose slot it has positively
re-observed as rendering content. It reuses `emptySectionVerdict` — the pure, unit-tested
predicate already written for the completion gate — so there is one answer to "is this section
empty", not two. Enumeration is from the **item** side: walk the slots that `empty_section` items
name for this site and ask what occupies each one now. Every retraction is a statement about a
row that was read.

**Measured on the live queue before any code was written**, and the split is the result that
matters:

| bucket | items | disposition |
|---|---|---|
| every deployed component in the slot renders content | **17** (6 sites, 15 `unresolved`) | retract |
| still empty | 19 | leave open |
| slot holds **no deployed component** | **10** | leave open — ambiguous |
| **mixed** slot (one of several still empty) | **1** | leave open — conservative |

**The absence rule this RFC forbids would have closed the 10 + 1 = 11 items it has no evidence
about.** That is the standing condition's cost, finally denominated.

### Three findings the adoption produced, which the paper did not anticipate

**1. The seam's addressing unit is coarser than the producer set — and this is the first thing an
adopter must check.** `Resolved` closes rows by `(site_id, item_type, item_key)`. Where several
producers deliberately converge on one key — which the owner's ruling of 2026-08-02 (`RFC_010_who
_may_answer_a_page_name_collision`) explicitly encourages — **one producer's positive observation
closes another producer's finding**. 13 item types have ≥2 Go producers today.

The worked case is worse than "they might disagree". `undeployed_asset` was the obvious first
adopter (95 open items, and a switch arm that already observes "Deployed." and discards it). It is
filed by BOTH `check_undeployed_assets` and `write_render_audit_findings_action` under the same
key, by design. The render audit's finding is *"this image serves broken on a real page"*; the
check's healthy signal is *"a deployed component's HTML references the filename"*. **You cannot
have a broken `<img>` unless the HTML references its src** — the two are positively correlated, so
adopting there would have retracted every render-audit 404 finding on the next sweep. It was
rejected on that ground and the hazard is now in `LANDMINES.md`.

*This does not change what the seam guarantees, so it is not an RFC-scope change under the
2026-07-29 ruling — but it is a precondition every adopter must discharge, and it belongs here
rather than in one check's comments.*

**2. Adopting by appending to the end of `Run` ships an inert mechanism.** Most monotonic checks
open with `if len(findings) == 0 { return }` — correct for a check that can only file, exactly
backwards for one that can retract, because a site with zero findings is the only site that guard
fires on and precisely the site whose stale items need closing. It is green in every test that has
a finding. Also in `LANDMINES.md`; mutation-proven in this adoption.

**3. Retraction interacts with the two-strike rule, and the interaction is real though empty
today.** `insertWorkItem` counts `status IN ('complete','failed')` rows created within 7 days and
brands the third item on a repeated `item_key` as `unresolved`. A retraction writes `complete`
onto an existing row, so it can burn a strike — and at two strikes the next genuine detection is
born `unresolved` and undispatchable, i.e. the landfill, created by the mechanism meant to drain
it. Measured: of the 17 items that retract, **0** were created within 7 days and **0** keys are
already at 2 strikes. Stated as a live interaction rather than dismissed, because this lane's
round-2 error was exactly a "measured as currently unreachable" claim that was not. Whether "the
problem went away by itself" and "a handler fixed it" should feed that counter identically is a
question for a human, not a measurement.

## What this RFC deliberately does not do

It does not implement anything. `CheckResult` is shared by 50 checks and the runner is on the
dispatch path; changing either from inside a bug lane is the `bugs_closed/124` shape, and the
guardian seat would be right to veto it. It also does not propose closing any of the 497 aged
items — that number is a measurement of *ignorance*, not of falsehood, and acting on it before
the mechanism exists would be exactly the blunt-instrument failure of option 4.

## Related

`bugs_closed/168` (the change that made a stale item dangerous), `bugs_open/179` (the residual
`deploy_path` escape hatch on the same action), `bugs_closed/142` (the fix whose predicate
change orphaned the eleven), `bugs_open/152` (robot-hands' `assets.url` template literal, still
real, deliberately untouched by the repair), `WRONG_CALLS.md` 2026-08-02 ("I told a council
twice that a clobber path was unreachable"), the `Dedup index ↔ Go list lockstep` landmine, and
`RFC_009` — whose closing section names the same underlying habit from the other direction: the
platform reconstructs state from metadata instead of recording what it did.

## Open questions

**Q1 — should a retraction count toward `insertWorkItem`'s two-strike counter?** Opened by the
owner ruling of 2026-08-03 (above), which accepted the current behaviour and asked for the
question to be tracked rather than closed. The counter's stated purpose is to break
discover/fix loops: *"a discover agent that keeps re-finding an issue after the fix agent
reports `complete` would loop forever if we only counted 'failed'"*. A retraction is a different
signal — the finding stopped reproducing without a handler doing anything — so counting it
identically is at least arguable, and the failure it risks is the one this whole paper is about:
at 2 strikes the next genuine detection is born `unresolved`, i.e. the landfill, created by the
mechanism meant to drain it. **Measured 0 of 17 affected on 2026-08-03**, and the measurement is
what makes deferring it safe rather than merely convenient. Whoever picks this up: it changes
shared insert-path behaviour for every item type in the estate, so it wants its own council
round, and `workItem.recurrenceExpected` is the existing lever that already exempts a class of
items from the same counter — read it before inventing a second one.
