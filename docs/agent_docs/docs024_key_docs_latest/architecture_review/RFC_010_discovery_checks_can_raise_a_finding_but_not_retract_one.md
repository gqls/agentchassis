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
