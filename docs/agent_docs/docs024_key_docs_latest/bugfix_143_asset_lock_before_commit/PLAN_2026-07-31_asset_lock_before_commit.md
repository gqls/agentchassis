# PLAN — bugs_open/143: an asset derivation must check the lock BEFORE it commits

**Opened 2026-07-31.** Case file: `bugs_open/143_HANDOFF_2026-07-29_derive_card_asset_commits_before_lock_check.md`
(filed by the `bugfix_131_og_card` lane, which explicitly deferred the fix:
"`bug_historian` still wants the lock check centralised (that is `bugs_open/143`'s job)"
— `bugfix_131_og_card/NOTES_og_card.md:566`).

## What is actually wrong (verified first-hand, 2026-07-31)

`platform/orchestration/actions/derive_card_asset_action.go`:

- **:163** `sendGitCommitRequest(...)` — the card JPG replaces the file in the site repo.
- **:184** `WHERE assets.locked_at IS NULL` — the ONLY lock reference in the file, on the
  provenance upsert, which runs **after** the artefact has already been replaced.
- The upsert is `ExecContext` and its `RowsAffected` is **discarded**, so when the lock
  guard suppresses the `DO UPDATE` the action still returns `derived: true`. The failure
  is not merely mis-ordered — it is invisible.

So an owner-approved (locked) card keeps its DB row and loses its artefact, and the
caller is told it worked.

`derive_brand_head_assets` had exactly this shape and it was fixed on the og-card lane
(`e9e345464` + round-2 revision `a22010eaa`): `lockedBrandHeadKeys` reads the locked set
**before** the commit, the locked artefact is left out of the `files` map entirely, and
the partial-lock case is reported at the call boundary via `skipped_locked`.

## Scope: what is in the population, and what I nearly wrongly added

Every `sendGitCommitRequest` call site in the tree (4):

| call site | writes an `assets` row for the bytes it commits? | in population |
|---|---|---|
| `derive_card_asset:163` | yes — `INSERT INTO assets … ON CONFLICT` | **YES — the bug** |
| `derive_brand_head_assets:175` | yes — `recordDerivedAsset` | **YES — already fixed, to be de-duplicated** |
| `emit_sprite_css:145` | no — commits CSS; only *reads* the sprite_sheet asset | no |
| `deploy_image_asset:236` | no — commits bytes the row already points at | **no, and this is a real distinction** |

**`deploy_image_asset` is NOT a sibling instance, and I nearly recorded it as one.**
It has no lock check anywhere and its `UPDATE assets SET url=… WHERE id=$1` (:255-261)
carries no `locked_at` guard — which looks like the same defect and is not. Deploying a
locked asset is **publication of the approved artefact**, not replacement of it: the
bytes come from the row the caller named, and recording the deployed local URL on a
locked row is desirable (without it the locked row keeps pointing at an expiring
presigned URL). Guarding that `UPDATE` would be a regression, not a fix. The 143 class
is narrower and precise: **a derivation that regenerates an artefact from a source and
overwrites the approved output.**

(`bugs_open/155` — `deploy_image_asset` resolves its source by `purpose` not `asset_id` —
CAN route the wrong bytes over a locked artefact's path, but that is 155's defect and
155's fix removes the trigger. That lane committed on that file today, so this fix does
not touch it: a same-file passenger is the one thing a pathspec commit cannot prevent.)

## The fix, ordered by what closes the door

1. **One shared helper, not two careful edits** — `asset_lock_guard.go`:
   `lockedAssetKeys(ctx, db, siteID, keys...) → map[string]bool`, plus
   `assetAgentWritableSQL()` for the writer's own `WHERE` clause so the decision is
   race-free. This is the third call site for "is this asset_key locked", which is the
   centralisation threshold the architecture seat named on the `bfd73f71` trail, and the
   same shape 016b already prescribes for a duplicated predicate ("one exported list plus
   a lockstep test — not two careful edits").
2. **`derive_card_asset` checks before the commit** — before storage is even required,
   mirroring brand-head's ordering — and refuses visibly (`derived:false`, `locked:true`,
   reason names the lock).
3. **The upsert's suppressed write stops being invisible** — read `RowsAffected`; 0 means
   the lock guard fired between the pre-check and the write (TOCTOU), and that is
   reported and logged at ERROR rather than returned as success.
4. **`lockedBrandHeadKeys` becomes a two-line delegation** to the shared helper, so the
   two surfaces cannot drift.
5. **A lockstep test that fails when a NEW producer appears** — the door for the "next
   producer" the architecture seat asked about on 145. It enumerates the
   `sendGitCommitRequest` call sites from the package source and fails if the set changes,
   forcing a deliberate classification instead of a silent omission.

## Semantics: deliberately NOT changed

`assets` carries `lock_type` and `lock_expires_at` (schema read 2026-07-31), and the
platform's canonical component predicate `pageComponentAgentWritableSQL` is
**expiry-aware**. Every existing *asset* lock check is not: `StoreAssetAction` (:2642),
`ingest_staged_asset` (:177, :297) and `lockedBrandHeadKeys` (:239) all test bare
`locked_at IS NOT NULL`.

Making the shared helper expiry-aware would **weaken three existing guards** — i.e.
change what the mechanism guarantees, which is RFC territory under the owner ruling of
2026-07-29 (§1: an addition to a shared vocabulary needs an RFC when it changes what the
shared mechanism GUARANTEES). So the helper preserves the existing asset semantics
exactly — bare `locked_at`, **no status filter** — and the expiry question is recorded as
an open note, not resolved by a side effect of this fix.

Live check backing "no live impact either way" (2026-07-31): 5 locked asset rows
fleet-wide, `lock_expires_at` NULL on all 5, `lock_type` set on 1.

## How it will be verified

- `go test ./platform/orchestration/actions/ -run 'AssetLock|CardAsset|BrandHead'`
- The lockstep rule proven to FIRE, not merely to pass, by feeding it a synthetic
  unguarded producer (a quiet test passes when the rule is gone).
- Council gate before/alongside the commit (scope is `platform/`).
- Live: no locked `card` rows exist (0 of 13), so this is latent-closure — the door is
  shut before anyone stands in it. The pod-grep proves the binary carries the guard.
