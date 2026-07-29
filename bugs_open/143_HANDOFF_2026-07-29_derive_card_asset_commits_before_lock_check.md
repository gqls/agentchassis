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
