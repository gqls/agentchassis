# 262 — the claims revalidator certifies a claim REMOVED from DB state, while the served page may still carry it

> ## ✅ CLOSED 2026-08-12 (evening) — `ce8733262`, **LIVE on `v1.0.1293`**. Bar met: fixed AND live.
>
> **Probed at the artefact, not at git, the tag or a status.** 98 Running `agent-chassis` pods on ONE
> digest `sha256:4717bcb3`, tag `v1.0.1293`, rolled 2026-08-12T19:13:54Z. `grep -a` on `/proc/1/exe`,
> exit code captured beside every count:
>
> ```
> NEEDLE_262=1 rc=0   ("…the database is not the website")
> NEEDLE_262b=1 rc=0  ("an unreadable page row is not evidence that anything shipped")
> NEGCONTROL_oldselect=0 rc=1   (a string ea18664f3 REMOVED — proves the binary is NEWER than it)
> CONTROL_pos=1 rc=0            ("register moved, not the page" — the owner's gate)
> CONTROL_absent=0 rc=1         (fabricated — proves grep can return 0)
> ```
>
> ⚠ **This commit has no removed-string control of its own.** Its only deletions in the shipped file
> are two function signatures, which leave no greppable literal. `NEGCONTROL_oldselect` therefore
> proves the binary is newer than `ea18664f3`, **not** newer than `ce8733262`; what proves *this*
> commit shipped is the two present needles, both unique to it (`grep -rn` over non-test Go → 1 hit).
> Say it that way rather than letting a borrowed negative control vouch for a later commit.
>
> ### ⚠ The gate is LIVE but has NEVER BEEN ASKED — and that is not the same as "it approved"
>
> **The sweep has not run since the roll.** `review-queue-revalidate-daily` (daily, enabled) last
> fired **2026-08-12T08:44:33Z**; the published gate was committed at **17:42Z** and rolled at
> **19:13Z**. So its `0` refusals is *"never asked"*, structurally — not *"asked and found nothing"*.
> The same holds for the claim-granular gate (committed 12:56Z, also after that run).
> **The next daily run, ~2026-08-13 08:44Z, is the first that can exercise either.**
> 21 open items will be re-examined then.
>
> ⚠ **Do not read the run history off the revalidation stamp.** `result->'revalidation'->>'at'` is
> **last-write-wins** — every open item carries the SAME `2026-08-12T08:44:34Z`, because each run
> overwrites it. Grouping by it looks like a run log and is not one; it shows the last run that
> touched each row, so a run whose items were all re-examined later is invisible. I inferred
> "08-11 was skipped" from exactly that and it was **unsupported** — `scheduled_tasks` is where the
> cadence lives. To ask *"has my gate ever been asked?"*, compare `max(stamp)` against the ROLL time,
> never against the refusal count.
>
> Fix candidate 1 was taken. `resolved` now also requires that the page was **published after its copy
> last changed**. Four refusal arms, all non-terminal and all spelled differently, because *"I could
> not look"* must never share a spelling with *"it has not shipped"*: the page row could not be read ·
> the page has never been deployed · `deployed_at` precedes the newest examined component update ·
> `build_status` is not `deployed`.
>
> **`build_status` and the timestamp are checked SEPARATELY** — neither implies the other, which is
> the trap this file warned about. **Equal timestamps RESOLVE**: publishing in the same instant as the
> edit is a real ordering, and refusing it would strand items whose clocks agree. That boundary has its
> own test.
>
> Placed **last** in the ladder: *"the words are still there"* and *"the copy never moved"* are more
> informative refusals, so this arm is reached only once the content evidence is otherwise sufficient.
>
> All five behaviours **mutation-tested** — disabling each of the four arms, plus making the comparison
> exclusive — each fails exactly the test written for it. Every pre-262 test now passes a
> `publishedAfterEdit()` fixture so it still isolates the arm it was written for.
>
> **Ownership was checked before starting:** `who-owns.py` clean, **0** open work items matching, and
> the three live transcripts naming this file were all incidental (two `ls` listings, one `git status`).

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
