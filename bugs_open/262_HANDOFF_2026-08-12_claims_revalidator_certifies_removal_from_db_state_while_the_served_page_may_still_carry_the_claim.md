# 262 — the claims revalidator certifies a claim REMOVED from DB state, while the served page may still carry it

**Filed:** 2026-08-12 · **Lane:** `bugfix_168_deployed_asset_path` (retraction sweep)
**Found by:** council `debug_historian` (MEDIUM, advisory) in the round that **APPROVED** the change —
correlation `b67eb26a-14ef-45d7-b755-3e489fd57ef0`, verdict 2026-08-12 14:34:14Z.

> **How this was verified, per the owner ruling of 2026-07-31.** No `090` run. Substituted
> first-hand verification, and stated plainly because the rule requires it: the claim here is not a
> theory about a cause, it is two checked facts — a `grep` showing the columns are never read, and a
> live count showing rows already exposed. Both commands are below and both are one line. If the
> next session wants a mechanism claim beyond "these columns are not consulted", that **would** need
> the loop.

## The defect

`claims_unverified` is a finding about what a **live site asserts**. The check that files it is even
called `ScanDeployedClaims`. But the scan and both of its closing gates read **`page_components`
only** — the database — and never ask whether that content has actually been **served**.

```
grep -c "build_status\|deployed_at\|last_deployed" \
  platform/orchestration/actions/discovery_checks/check_unverified_claims.go   # 0
  platform/orchestration/actions/revalidate_unverified_claims.go               # 0
```

Both columns exist and are populated: `pages.build_status` (`planned` / `deployed` /
`needs_rebuild`) and `pages.deployed_at`.

So a component can be edited in the DB — bumping `updated_at`, changing what the scan reads — **before
the rerender/deploy has shipped it**. Both gates then pass on evidence that says nothing about the
served page:

- the **copy-changed** gate sees `updated_at > created_at` → "the page moved";
- the **claim-granular** gate sees the cited text gone from `ExaminedTextBySlot` → "the claim went".

and the item closes, certifying removal of a claim **the public page may still be serving**.

## Measured exposure — this is live, not hypothetical

```sql
WITH items AS (
  SELECT w.id, w.spec->>'page_id' AS page_id, w.status
  FROM site_work_items w WHERE w.item_type='claims_unverified' AND w.spec ? 'page_id')
SELECT i.status, count(*) AS items,
       count(*) FILTER (WHERE pc.newest > p.deployed_at) AS db_ahead_of_deploy,
       count(*) FILTER (WHERE p.deployed_at IS NULL)     AS never_deployed
FROM items i JOIN pages p ON p.id = i.page_id::uuid
LEFT JOIN LATERAL (SELECT max(updated_at) AS newest FROM page_components
                   WHERE page_id=p.id AND locked_at IS NULL) pc ON true
GROUP BY 1;
```

**2026-08-12: of 9 `complete` items, 2 sit on pages whose newest unlocked component update is LATER
than `deployed_at`.** (Open items: 1 of 19, plus 1 never deployed.)

⚠ **State that precisely.** This does **not** prove the live page still carries those claims. It
proves the closure's evidence **cannot show that it doesn't** — the fix may be sitting in the DB
unpublished. For a type whose entire subject is what the site says in public, "we cannot tell" is the
finding.

## Why it is not a regression, and why that makes it worse rather than better

**Predicate parity holds and was deliberate**: the emit side has exactly the same blindness, so the
revalidator judges by the same predicate the check files by — which is this lane's founding rule and
should not be broken casually.

**But the two ends are not symmetric in consequence:**

| side | blind to deploy state ⇒ |
|---|---|
| emit | files a finding about text that may not be served — a human reads it, low cost |
| **revalidate** | **closes** a finding while the claim may still be live — no human ever looks again |

So the same blindness is tolerable where it files and damaging where it **closes**. That asymmetry is
the argument for fixing only the closing end, and for not "fixing" the emit side to match.

## Fix candidates, ordered by what closes the door

1. **Refuse to close when the page has not deployed since the copy changed.** In
   `unverifiedClaimsVerdict`, after both existing gates: require
   `pages.deployed_at >= <newest examined component update>` and `build_status = 'deployed'`, else
   return `unknown` (non-terminal, item stays visible). Makes the bad state unrepresentable at the
   only place that matters, and costs one more column on a row the sweep already loads.
   ⚠ `parkedReviewItem` does not carry page fields — this needs the page row, so it belongs in
   `revalidateUnverifiedClaims`'s DB half, not in the pure ladder.
2. **Have `ScanDeployedClaims` return the page's deploy state** alongside `NewestComponentUpdate`, so
   both ends can see it and a later `voice_tells` adoption inherits it. More work; better shape;
   pairs naturally with `features_open/032`'s shared-helper request.
3. **Downgrade instead of refusing** — close but mark the closure provisional. Rejected here: the
   type exists because a machine must not silently settle a factual-claim question, and a
   provisional close is still a close.
4. **Do nothing, document it.** What we have today. Acceptable only while someone is watching the
   two exposed rows.

## Traps for whoever fixes this

- ⚠ **`build_status = 'deployed'` is not the same question as `deployed_at >= updated_at`.** A page
  can be marked deployed and still be serving a build older than the last content edit. Check the
  timestamps, and use the status only as a second condition.
- ⚠ **This narrows closure again.** The type already drains slowly (see the register entry's
  measured false-refusal cost); adding a third gate will slow it further. That is the safe direction,
  but say so with a number rather than discovering it later.
- ⚠ **Do not "fix" the emit side to match.** Predicate parity is what stops the two ends answering
  the same question differently; the asymmetry above is in consequence, not in predicate.
- The general form of this trap is already in `LANDMINES.md:1690` — *"A data repair RACES the sweep
  that publishes it — and the DB is not the website"*. This is that landmine, met by a mechanism that
  **closes human-review rows**.

## Related

- concept register **CQ-021** (both gates, with the measurements) and **CQ-020** (`voice_tells`,
  which has the same hole and neither gate)
- `features_open/032` — the `voice_tells` asymmetry and the shared-helper request
- `bugs_open/185` — detectors selecting on deployed state and missing live pages: adjacent, not the
  same (that one is about **selection**, this one about **closure**)
- Council verdict with the full objection:
  `docs024_key_docs_latest/bugfix_168_deployed_asset_path/VERDICT_2026-08-12_round7_*.json`
