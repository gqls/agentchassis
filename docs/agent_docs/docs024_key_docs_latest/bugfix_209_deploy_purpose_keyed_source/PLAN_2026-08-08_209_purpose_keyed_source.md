# PLAN — 209: `deploy_image_asset` resolves a source by PURPOSE

Started 2026-08-08 (late), picked up from the 221 lane's handoff.
Design, phasing, decisions **and their reasons**. Corrections to the originating
brief live here, marked as corrections.

---

## The brief, as inherited

`bugs_open/209` (filed 2026-08-06 by the `bugfix_152_155_asset_source_identity`
lane, as 155's second arm) says: `findStorageURI`'s Priority 2 reads a top-level
`{purpose}_uri` from `collected_data` — a purpose-keyed, last-write-wins slot used
as a per-asset source — and it is consulted *before* the `asset_id` path. Severity
medium. Fix candidates ranked 1) delete Priority 2, 2) key by asset, 3) guard on
"exactly one same-purpose asset".

## CORRECTION 1 (2026-08-08) — the fix ranking is inverted, and candidate 1 is harmful

> The bug file ranks **first**: *"Delete Priority 2 and make the asset_id path the
> only DB-free route."* **Measured, that fix would cause the class of failure it is
> meant to prevent.**

The two workflows that store and deploy in one run (`pageflow-builder`,
`site-work-orchestrator`) carry **no `input_fields`** on their deploy steps. They
resolve inputs through `ExtractActionInputs` Strategy 2 → aggressive recursive
search, and `findFieldRecursive` iterates a Go map, whose order is randomised.
Running the real helper 400× on identical input, the **logo** deploy step resolved
the **hero's** `asset_id` in **344/400 runs (86%)**.

Priority 2 is therefore load-bearing, not cruft: those two workflows' assets differ
*exactly by purpose*, so a purpose key is the correct discriminator for them.
Removing it swaps a correct lookup for an 86%-wrong one.

Evidence: `NOTES_209_purpose_keyed_source.md`, and the four characterisation tests
in `platform/orchestration/actions/deploy_image_asset_purpose_source_test.go`.

## CORRECTION 2 (2026-08-08) — the defect is wider than "Priority 2"

Priorities **3–7** of `findStorageURI` are *also* purpose-keyed. A fix phrased as
"delete Priority 2 so asset_id is the only DB-free route" does not do that; it
falls through to five more purpose-keyed lookups. Any real fix must address the
function's *keying*, not one branch of it.

## CORRECTION 3 (2026-08-08) — `[UNMEASURED]` is now a measured negative

The bug file left open whether any live workflow can trigger this. **Measured: none
can.** Census of every live definition carrying `deploy_image_asset` (three:
`asset-deployer`, `pageflow-builder`, `site-work-orchestrator`), plus
`image-build-handler`, which is the file's named suspect:

- the hero+logo pair use **different purposes**, so their slots never collide;
- `image-build-handler`'s two same-`purpose:"hero"` stores are on **mutually
  exclusive branches**, and it has no deploy step — it delegates with
  `s3_uri: asset_stored.image_uri`, i.e. by identity;
- `asset-deployer`'s `collected_data` contains **no `{purpose}_uri` key**, and the
  reader is a strict path walk, so the branch is unreachable there even on the
  store-failure path (it degrades to a safe skip).

**Status change proposed: 209 is LATENT, not live.** Real defect, no reachable
instance. This is a downgrade in urgency and *not* a closure — the door is still
open for the next workflow that stores two same-purpose assets in one run.

## Decision taken this session: diagnose and pin, do not fix yet

Reasons, in order:

1. **No live exposure.** Changing a shared action's resolution order to fix an
   unreachable defect spends council and roll risk against zero current damage.
2. **The obvious fix is the harmful one.** Having just measured that, shipping
   *any* reordering tonight without the council would be the exact mistake the
   platform-seams ruling exists to prevent.
3. **The cheap durable win is available now**: characterisation tests that record
   *why* Priority 2 was kept, so the next thread that reads the bug file's ranking
   and reaches for candidate 1 trips over the 86% figure instead of shipping it.

## Where a fix should go when it is taken (recommended ranking, replacing the file's)

1. **Key the in-run value by asset identity, additively** — the bug file's
   candidate 2. Write `asset_uris.<asset_key>` alongside the existing
   `{purpose}_uri`, and have the reader prefer `asset_key` when the caller supplies
   one (`asset-deployer` already does; `image-build-handler` already passes
   `asset_key?`). Keeps every current caller working, makes identity available, and
   nothing is deleted. **This is a platform seam** — a new reserved key namespace on
   a shared action — so per the 2026-07-28/29 rulings it needs a concept-register
   entry *in the same commit* and a council submission alongside, and the other
   consumers must be **told**, not merely measured.
2. **Give the legacy deploy steps explicit `input_fields`** so they stop resolving
   by recursive search at all. This is the fix for the *general* hazard rather than
   this bug, and it is config-only (live immediately, no roll).
3. **Only then** consider retiring the purpose-keyed priorities, once no caller
   depends on them.

Candidate 3 from the bug file (guard on "exactly one same-purpose asset in the
run") stays last: it is a guard against a state the code can still express.

## Open questions for whoever takes the fix

- ~~`[UNVERIFIED]` Are `pageflow-builder` / `site-work-orchestrator` reachable at all?~~
  **ANSWERED — OWNER RULING 2026-08-09:** *"pageflow-builder and site-work-orchestrator
  are not dead, but not being worked on. If we need to diverge from them then we can
  use new actions and workflows as necessary."* So: alive, frozen, must keep working
  untouched. The fix is divergence, not surgery on the legacy pair.
- Does the recursive-`asset_id` instability bite any *other* action with a
  no-`input_fields` step whose spec field name occurs twice in `collected_data`?
  That is a shared-helper question, wider than 209.

## DECISION PENDING (2026-08-09) — fix shape under the owner ruling, awaiting go-ahead

The ruling makes the recommended ranking concrete. Two divergence shapes were put
to the owner (plain-prose version in `README_where_we_are.md`, same date):

1. **(Recommended) Opt-in field on `deploy_image_asset`, unsafe default OFF.**
   e.g. `resolve_by_identity: true` in step config. OFF = today's behaviour bit
   for bit (legacy pair never sets it). ON = resolve by explicit `s3_uri` input or
   `asset_id`/`asset_key` only — **no purpose-keyed fallback, no recursive
   search** — and fail loudly (skip-with-reason) rather than guess. Modern callers
   (`asset-deployer`; `image-build-handler`'s `call_asset_deployer` mapping) opt
   in; they already pass identity. Follows the 2026-08-02 owner ruling ("new
   authority on a shared seam ships as an OPT-IN FIELD, not a documented
   contract"). The characterisation tests shipped 08-08 are the proof harness for
   the OFF position.
2. **New action, old one untouched.** Zero edit-risk to the frozen pair; costs a
   duplicated download-optimise-commit path and a second maintenance surface.
   Only if the owner prefers the legacy code literally unedited.

Either shape still wants the additive writer half — `asset_uris.<asset_key>`
recorded alongside `{purpose}_uri` — which is the platform seam: concept-register
entry in the same commit, council submission alongside, consumers told.

**No code yet.** The owner asked for the problem explained clearly before moving
further; building starts on their pick of shape 1 vs shape 2.
