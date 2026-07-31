# 143 — derive_card_asset commits to git before its only lock check runs

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
