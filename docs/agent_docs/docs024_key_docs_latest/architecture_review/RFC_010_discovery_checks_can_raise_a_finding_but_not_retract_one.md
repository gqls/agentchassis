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
