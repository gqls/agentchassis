# 179 — `deploy_image_asset` can still bypass the shared path derivation (`deploy_path`), and can now write to another writer's artefact (brand-head)

**Filed:** 2026-08-02 by the `bugfix_168_deployed_asset_path` lane, discharging the
council's `bug_historian` seat objection at round 1 (`abd9b119`): *"No work item /
follow-up filed for the disclosed `deploy_path` passthrough exposure — should be tracked so
a future caller doesn't silently reintroduce the divergence this plan just closed."* The
seat was right that a disclosure in prose is not tracking, which is the same failure mode
that produced `bugs_open/168` itself (the 128 lane wrote "its own item" and did not create
one).

**Severity:** Low **today**, on a measured empty risk set — exactly like 168, and filed for
exactly that reason. **Class:** a shared mechanism with a documented escape hatch that no
guard covers. **Status:** OPEN, unowned.

**Depends on:** `bugs_open/168`'s fix being live (commit `4035455ae`, chassis image pending).
Both findings below are *about* the contract that change introduces; neither is reachable
before it ships.

---

## Finding A — `deploy_path` overrides the derivation and is invisible to every reader

`storage.DeployedAssetPath` is, as of 168, the one derivation the writer and all six readers
resolve through. But `deploy_image_asset_action` applies it and *then* lets an input override
the result outright:

```go
derived := storage.DeployedAssetPath(assetKey, purpose)
processed.Paths = derived
// ... then:
deployPath := inputs.Get("deploy_path")          // or config["deploy_path"]
if deployPath != "" {
    processed.Paths = storage.AssetPaths{FilePath: deployPath, ...}   // wins
}
```

A caller that sets `deploy_path` publishes the file somewhere **no reader can predict**,
because readers only ever see `(asset_key, purpose)`. That is the same shape as the defect
168 closed — a writer whose output the readers cannot derive — reintroduced through a
supported input rather than through duplicated code.

**Measured, 2026-07-31 (128 lane) and unchanged 2026-08-02:** no Go code sets it, and it
appears in **zero orchestrations in history**. It is declared in the action's input spec and
in two SQL seeds (`044_asset_deployer.sql`, `107_image_build_handler.sql`) as an optional
passthrough. So the risk set is empty *today* — and nothing makes it stay empty.

## Finding B — a brand-head purpose routed here now OVERWRITES the deriver's artefact

Raised by the council's `guardian` seat in the same round, severity medium: *"no guard stops
a future caller of `deploy_image_asset` from doing so and silently overwriting favicon/
og-card paths for a site."*

This is a **genuine behaviour change introduced by 168**, and the ticket should say so
plainly rather than let it live only in a commit message:

| | before 168 | after 168 |
|---|---|---|
| `purpose=og_card` via `deploy_image_asset` | writes `assets/images/og_card.png` | writes `assets/images/og-card.png` |
| consequence | a file **nothing references** — inert garbage | **overwrites** the artefact `derive_brand_head_assets` published and the site head points at |

Neither is good, and 168's version is the *correct* one for a path-agreement contract — the
old behaviour was a silent no-op that looked like a successful deploy. But the failure mode
moved from "harmless orphan" to "clobbers an owner-visible brand asset", and the clobber is
a **git commit that runs before any provenance guard** — see LANDMINES.md § *"Guarding an
asset's provenance UPSERT is not guarding the asset — the git commit already ran"*, and
`bugs_open/143`.

**Why nothing is broken today — measured, not assumed (2026-08-02):**
`check_undeployed_assets` excludes brand-head purposes from its generic half
(`AND NOT (COALESCE(a.purpose,'') = ANY($2::text[]))`) and routes them to
`needs_brand_head_assets`, whose repair is **re-derivation**. And every variable-purpose
reader resolves through `site_plan_imagery`, whose reachable purpose set is
`hero`/`icon`/`illustration`/`sprite_sheet` — **163 rows, zero brand-head** (denominator run:
the same join unfiltered returns those 163). So no live path routes a brand-head purpose to
this action.

## Fix candidates, ordered by what closes the door

1. **Make the deployer refuse a purpose it does not own.** `deploy_image_asset` returns a
   refusal (not an error) when `storage.IsBrandHeadPurpose(purpose)` and no explicit
   override is present, naming `derive_brand_head_assets` as the writer. Makes the bad state
   unrepresentable at the only moment anyone is watching, and gives
   `IsBrandHeadPurpose` — which 168 left with **zero production callers** — a real job again.
2. **Route `deploy_path` through the derivation instead of around it**: keep the override but
   require it to be *recorded* where readers can see it (`assets.storage_path` /
   `assets.filename` already exist and are populated on 78 of 267 active rows). Then a
   reader can prefer the recorded path over the derived one, and the escape hatch stops
   being invisible. This is the same "record what you wrote instead of re-deriving it" fix
   that `bugs_open/152` and `bugs_open/155` want, so it should probably be designed with
   them rather than separately.
3. **Delete `deploy_path`.** Zero callers in history. Smallest surface, but it removes an
   escape hatch someone may have been relying on out-of-tree, and an unused mechanism that
   is *documented* is cheaper to keep than to re-derive later.
4. Leave both, and rely on nobody doing it. **This is the status quo for A**, and for B it is
   a status quo that only became a clobber on 2026-08-02.

## Scope

Candidate 1 is a behaviour change to a live action's contract (a new refusal path) — council
gate, not a silent bug patch. Candidate 2 is architecture-scope and overlaps two other open
lanes; **check `bugs_open/152` and `bugs_open/155` before starting**, and read
`scripts/who-owns.py` on both — 155 is a defect in *this same file*
(`resolveStorageURIFromAsset` resolves the source image by purpose rather than by the
`asset_id` passed).

## Verify a fix

- A `deploy_image_asset` step with `purpose=og_card` and no override must not commit a file
  (candidate 1), and the refusal must be a *result*, not an error — the platform's convention
  for a declined action.
- The `deploy_path` census must stay at zero, or the recorded-path reader must exist:
  `SELECT count(*) FROM orchestration_states WHERE collected_data::text LIKE '%deploy_path%';`
- `IsBrandHeadPurpose` regains a production caller, or is deliberately retired — a helper
  with no callers looks exactly like a finished refactor.

## Related

`bugs_open/168` (the change that created finding B and disclosed finding A),
`bugs_open/152` + `bugs_open/155` (the same "identity reconstructed rather than read" root;
155 is in this same file), `bugs_open/143` (asset locks vs the git commit that precedes
them), council `abd9b119` round 1 (`bug_historian` + `guardian` objections, both discharged
by this file), IMG-067 / IMG-066 in the concept register.
