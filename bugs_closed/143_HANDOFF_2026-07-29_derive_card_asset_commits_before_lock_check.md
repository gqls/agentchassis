# 143 — derive_card_asset commits to git before its only lock check runs

> **CLOSED 2026-07-31 — FIXED and LIVE on `v1.0.1218`, pod-verified on BOTH replicas.**
> Fixed by session "bugfix 10" (`docs024_key_docs_latest/bugfix_143_asset_lock_before_commit/`).
> Council gate **APPROVED at round 1**, corr `b5ff41f1-84bc-477b-a881-83e3d2e8a805`
> (12 seats, 2 low + 1 medium objection, none high). Closure record at the bottom
> of this file.

Filed 2026-07-29 by the bugfix_131_og_card lane (session relojistas-5), confirmed while
answering a council objection on the sibling fix (trail `bfd73f71-…`, round 1 bug_historian
medium: "does any other action share this shape?" — yes, this one).

## The defect

`platform/orchestration/actions/derive_card_asset_action.go`: the git commit of a content
entity's card image runs at **:163** (`sendGitCommitRequest`), and the ONLY `locked_at`
reference in the file is the provenance upsert's `WHERE assets.locked_at IS NULL` at **:184**
— after the file in the site repo has already been replaced. An owner-approved (locked) card
row would keep its DB row and lose its artefact, exactly the class fixed for
`derive_brand_head_assets` in commit `e9e345464` + its round-2 revision (locks checked BEFORE
the commit, no status filter, fail closed).

Same family as `bugs_closed/058` (rebuild path ignored page_component locks) and `069`
(site_components writers ignored chrome locks): a writer bypasses a lock that lives elsewhere.

## Measured exposure (2026-07-29)

`SELECT count(*) FROM assets WHERE purpose='card' AND locked_at IS NOT NULL` → **0** (of 12
card rows). Latent, not live — the same state brand-head was in before the leopardess locks
were backfilled. The door is open; nobody is currently standing in it.

## Fix shape

This is the **third** call site needing "is this asset_key locked" (after `store_asset`'s
inline upsert guard and `lockedBrandHeadKeys`), which is the centralisation threshold the
architecture seat named on the `bfd73f71` trail: extract a shared helper (siteID + asset_key(s)
→ locked set, `locked_at IS NOT NULL`, **no status filter** — assets.status is unconstrained
text, a guard conditioned on it fails open), call it in `derive_card_asset` before the commit,
and swap `lockedBrandHeadKeys` to it. Refusal shape exists in-file (`derived:false`).

## How to verify

Lock a test entity's card row (`locked_at=now()`), run the derivation, confirm: no git commit
for that card, a visible refusal/skip in the return, row untouched. Unlock, re-run, confirm
normal derivation.

---

## Contribution 2026-07-31 (session "bugfix 8") — I collided with the owning session and stood down

I picked this bug up as UNOWNED and it was: `scripts/who-owns.py 143` reported no owning
workstream, and `git status` showed no working-tree edit to
`derive_card_asset_action.go`. **Both readings were true and both were stale within minutes** —
another session (`759437b9-…`, workstream
`docs024_key_docs_latest/bugfix_143_asset_lock_before_commit/`) was mid-fix and had written
nothing to disk yet. We wrote a shared asset-lock helper into the same package **in the same
minute**. Theirs (`asset_lock_guard.go` + `assetKeyLocked`) is on disk; mine
(`asset_lock_helpers.go`) has been deleted and my hunk in `derive_card_asset_action.go`
reverted, verified with `go build ./platform/...`. **Their fix is the one to review; nothing of
mine remains in the tree.**

Worth recording because the check that failed is one CLAUDE.md already names: *"It reads
COMMITS, so a session mid-fix is invisible — check the tree too."* I did check the tree. The
gap is that **neither signal exists during the interval between a session choosing a bug and
its first Write** — which for a research-first session is 20+ minutes. Logged in
`WRONG_CALLS.md`.

Independently reached the same design (fail-closed on bare `locked_at`, no status filter, no
expiry test, reuse `toPGTextArrayLiteral`), which is mild corroboration that it is the right
shape. **Three things from my pass that the shipped fix may not cover** — offered as candidates,
not objections, since I have not read their final diff:

1. **The provenance upsert's suppression is silent, and the file's own header says it must not
   be.** `derive_card_asset_action.go`'s comment calls the entity-linked row "NOT best-effort:
   the entity link is what the query resolvers read, so a card that committed but never linked
   would stay invisible forever". But the upsert is an `ExecContext` whose `WHERE
   assets.locked_at IS NULL` suppression yields `RowsAffected()==0` **with no error**, and the
   action then returns `derived: true`. `0` is discriminating here (a plain INSERT or a
   permitted DO UPDATE both give 1), so checking it is cheap and closes the TOCTOU tail the
   pre-check cannot: a lock taken after the pre-check still lets the git commit through, and
   today that outcome is reported as a clean derivation.

2. **Two of the three call sites the bug names still hand-write the predicate**, so
   centralisation is only one-third done unless they were swapped too:
   `v3_site_actions.go:2642` (`StoreAssetAction`) and
   `derive_brand_head_assets_action.go:336` (`recordDerivedAsset`).

3. **`ingest_staged_asset_action.go` is a FOURTH asset-lock site** (4 references, added
   2026-07-29 by the asset_amend_path lane) and does **not** have this bug — it reads the lock
   before the upload *and* re-checks `FOR UPDATE` inside the transaction. Leaving it alone is
   defensible; leaving it *unmentioned* would make a later reader think the sweep was complete.

**On the expiry question, one measurement to save the next thread a query.** `assets` is one of
migration 115's four Pattern-A lock tables and carries `lock_type` + `lock_expires_at` +
`chk_assets_lock_type` + `idx_assets_timed_lock`, so the component-side expiry-aware predicate
is schema-applicable — but adopting it would weaken the guard, and register **LOCK-004** already
owns that as a named, outstanding project ("~11 bare `locked_at IS NULL` call sites", Go
follow-on to migration 115, still `partial`). Live state, measured 2026-07-31:

```sql
SELECT count(*) FROM assets
 WHERE locked_at IS NOT NULL
   AND lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW();
```
→ **0**. Fleet-wide there are **5** locked asset rows; **4 carry `lock_type IS NULL`** (which the
conservative rule reads as permanent) and one is `permanent`. No Go writer stamps `lock_type`
onto an assets row at all, so a timed asset lock can only arrive by hand today. **The two
readings are therefore indistinguishable on live data — the choice has to be made on the
guarantee, not on the measurement**, and deferring to LOCK-004 is the right call.

---

## CLOSURE 2026-07-31 (session "bugfix 10") — fixed, reviewed, live, pod-verified

**Commits** (branch `087_towards_multiple_domains`):
`3aa7a5d17` the fix + register + landmine · `56d97a885` the standing five + WRONG_CALLS ·
`6f20acd78` council round-1 answers · `3dba831cf` the write-predicate de-duplication ·
(`c7380f57c` is an unrelated hook fix this work happened to expose).

### What shipped

1. **`platform/orchestration/actions/asset_lock_guard.go`** — one definition of "may
   automation overwrite this asset?" (concept register **LOCK-007**). The read predicate
   is DERIVED from the write predicate (`assetLockedSQL = NOT(assetAgentWritableSQL)`),
   never re-typed. Returns an `assetLockSet` carrying `locked_by`/`lock_type`/`locked_at`
   plus `Describe(key)`, so a refusal names what an operator must clear — live `locked_by`
   values are free-text sentences, so they are reported verbatim, never classified.
   Callers must treat an error as LOCKED; both do.
2. **The ordering fix.** `derive_card_asset` reads the lock immediately after page
   resolution — before the S3 download, the decode and the git commit — and refuses
   visibly (`derived:false`, `locked:true`, reason naming the holder and date).
3. **The invisible half, which this file did not name and which is why nobody noticed.**
   The provenance upsert was `_, err = ExecContext(...)`: a `DO UPDATE ... WHERE` the lock
   suppresses yields **no error and no row**, so the action returned `derived: true` on a
   run that had just destroyed an approved artefact. `RowsAffected()` is now read; 0 means
   a lock was taken inside the TOCTOU window, and that is reported at the call boundary
   and logged at ERROR, because the artefact and its row genuinely disagree at that point.
4. **De-duplication, both directions.** `lockedBrandHeadKeys` is a one-line delegation
   (the READ side), and all three hand-typed `WHERE assets.locked_at IS NULL` clauses —
   `derive_card_asset`, `recordDerivedAsset`, `StoreAssetAction` — now call
   `assetAgentWritableSQL("assets.")` (the WRITE side). Byte-identical SQL; what it buys
   is **LOCK-004's predicate sweep having ONE line to edit instead of four.**
5. **Two lockstep tests for the next producer**, which is what the architecture seat
   actually asked for. One fails when a NEW file joins the `sendGitCommitRequest` path
   (forcing a deliberate in/out classification); one fails when any writer hand-types the
   predicate again. Both `t.Fatal` on an empty scan — a blind classifier reports a clean
   tree for ever — and both are paired with a synthetic-case test proving the rule FIRES.

### Semantics deliberately NOT changed (and pinned so a later change is a decision)

No status filter, and **not expiry-aware** — bare `locked_at`, unlike the expiry-aware
`pageComponentAgentWritableSQL` (`lock_helpers.go:44`, confirmed). Adopting the
expiry-aware form would WEAKEN three existing guards, i.e. change what the mechanism
guarantees, which is register **LOCK-004**'s named outstanding Go predicate sweep and not
a bug fix. This file's own measurement — 5 locked asset rows fleet-wide, `lock_expires_at`
NULL on all 5 — shows the two readings are **indistinguishable on live data**, which is
exactly why the choice was made on the guarantee. `TestAssetLockPredicateIsBareLockedAt`
rejects `status`, `lock_expires_at` and `lock_type` in the predicate.

### The three candidates from the bugfix 8 contribution above — all answered

1. **Silent suppression of the provenance upsert** — fixed, see (3). They were right that
   `0` is discriminating (a plain INSERT and a permitted DO UPDATE both give 1).
2. **"Centralisation is only one-third done unless the other writers were swapped too"** —
   **correct, and I had shipped exactly that gap.** `assetAgentWritableSQL` had no callers
   but its own negation and its own test. Fixed in `3dba831cf`, with a test that keeps it
   fixed. This contribution is why it was checked rather than assumed.
3. **`ingest_staged_asset` is a fourth site that does NOT have this bug** — confirmed
   (`:177` pre-check before the upload, `:247` `FOR UPDATE` re-check in-transaction). Left
   alone deliberately, and now recorded on LOCK-007's `verify-later` so the boundary
   cannot quietly become a fifth undocumented duplicate.

### Scope: what was deliberately NOT fixed, with the reason

**`deploy_image_asset` has no asset-lock check and must not be given one.** It commits
bytes the named row *already points at*, so deploying a locked asset is publication of the
approved artefact, not replacement of it; guarding its `UPDATE assets SET url ... WHERE
id = $1` would leave a locked row pointing at an expiring presigned URL. `emit_sprite_css`
commits CSS and only reads the sprite asset. The 143 class is narrower than "commits to a
site repo": **regenerates an artefact from a source AND upserts the row describing it.**
That reasoning lives in `inPopulation()` in the test, not in a doc nobody greps.

### Live verification (the only evidence that counts)

`v1.0.1218`, both replicas, new symbol + positive control in the same exec:

```
agent-chassis-776f55c5f9-bjfhq  new143=1 sharedquery=1 control=3
agent-chassis-776f55c5f9-g9vqc  new143=1 sharedquery=1 control=3
```

(`new143` = `"provenance write SUPPRESSED by the asset lock"`, a string this fix created;
`sharedquery` = `"SELECT DISTINCT ON (asset_key)"`; `control` = `"approved assets are
never overwritten"`, which predates the fix and proves the grep and the binary are sound.)
The tag was NOT trusted: `v1.0.1216` existed locally before these commits, so per
`bugs_open/153` only the pod-grep decides. **Caveat stated rather than glossed:** the
write-predicate de-duplication (`3dba831cf`) landed after this roll and is not in
`v1.0.1218`. It produces byte-identical SQL, so the defect this file describes is fully
closed by what IS live; that commit is maintainability, and rides the next roll.

**Still latent, not a rescue:** 0 of 13 live `card` rows carry `locked_at`. The door is
shut before anyone stood in it.

### Where the rest of it lives

- Workstream (plan, notes, runbook, owner log, summary):
  `docs/agent_docs/docs024_key_docs_latest/bugfix_143_asset_lock_before_commit/`
- Concept register: **LOCK-007** in `docs026_concept_register/register/locks.md`
- Landmine: *"Guarding an asset's provenance UPSERT is not guarding the asset"* in
  `LANDMINES.md` (synced to `doc_notes`)
- Transferable pattern: 016b §9, *"A guard on the ROW is not a guard on the ARTEFACT"*
- `WRONG_CALLS.md`: the three-lane collision, written from the causing side
