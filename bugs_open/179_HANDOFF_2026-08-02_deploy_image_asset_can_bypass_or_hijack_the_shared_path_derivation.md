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
guard covers. **Status:** ~~OPEN, unowned~~ → **CLAIMED 2026-08-04 by session "bugfix 100"**
(lane `docs024_key_docs_latest/bugfix_179_deploy_path_override/`), working **finding A**.
Finding B is already done. Taken after the filing lane's own newest handoff recorded it as
still unowned (`bugfix_168_deployed_asset_path/HANDOFF_2026-08-03b_continue_here.md:34`) and
no live session held it.

> **Fix direction: candidate 3 (delete the override), plus a refusal on explicit intent.**
> Chosen over the reframed candidate 2 ("record the override so readers can see it") for a
> measured reason the file did not have: **all six readers DERIVE the path and none reads the
> recorded row**, so a recorded override is still a path no reader resolves. Preferring a
> recorded path is `bugs_open/152`/`155`'s seam, not this one.
>
> **New finding, and it makes the override worse than the file describes: the override can be
> armed by a caller who never asked for it.** `ExtractActionInputs` resolves every declared
> field through a **depth-20 recursive search of the whole of `collected_data`**
> (`datahelpers/unified_extractor.go:440-489`), and `deploy_path` is a declared optional input
> on the live `asset-deployer` row. So a `deploy_path` key anywhere in a deploy
> orchestration — a nested sub-agent response, an echoed spec — silently redirects the commit.
> That is why the refusal is wired to **explicit sources only** (step config, the deprecated
> `deploy_path_field`, `input_data.deploy_path`) and deliberately NOT to
> `inputs.Get("deploy_path")`: refusing on the recursive hunt would turn a stray nested key
> into a false denial of a legitimate deploy, which this estate has already ruled is worse
> than the inert key it chases.

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

**Measured 2026-08-02, across THREE populations — and the third is the one whose omission
caused this file's finding-B error.** Counting only "no Go code sets it" is exactly the kind of
partial census that made me call finding B unreachable when it was not, so:

| population | count |
|---|---|
| work items carrying `deploy_path` in `spec` (**the standing queue**) | **0** |
| active agent definitions setting a `deploy_path` **value** | **0** |
| orchestrations with a `deploy_path` **value** | **0** |

⚠ **Match the JSON shape, not the bare word.** `collected_data::text LIKE '%deploy_path%'`
returns 9 — **all nine are this lane's own council submissions**, because a council run stores
the submission JSON and its rationale argues about `deploy_path` at length. Use
`LIKE '%"deploy_path":"%'`. Declared as an optional passthrough in the action's input spec and
in two SQL seeds (`044_asset_deployer.sql`, `107_image_build_handler.sql`).

So the risk set is empty *today*, on a census that now includes the queue — and nothing makes
it stay empty. Both the `architecture` and `bug_historian` seats pressed this at medium in
council round 3 (`abd9b119`, APPROVED), architecture noting pointedly that *"the round already
proved that 'currently unreachable' measurements on this exact mechanism"* can be wrong.
**Measuring it empty is not fixing it.** That is why finding A stays open while finding B is
struck through as done.

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

> ## CORRECTED 2026-08-02, hours after filing — FINDING B IS REACHABLE, AND IS NOW FIXED IN CODE
>
> **This section originally read "Why nothing is broken today", and it was WRONG.** It is kept
> below, struck through, because the way it was wrong is the useful part.
>
> **What I measured:** (a) that `check_undeployed_assets` no longer *raises* brand-head
> `undeployed_asset` items, and (b) that every variable-purpose *reader* resolves through
> `site_plan_imagery`, whose reachable set holds no brand-head purpose. Both true, both
> irrelevant to the question.
>
> **What I failed to measure: the STANDING QUEUE.** The predicate that stops new items being
> raised says nothing about items raised *before* it changed, and nothing sweeps a queue for
> items whose defining predicate has since moved.
>
> ```sql
> SELECT status, s.domain, spec->>'mode' AS mode, spec->>'purpose' AS purpose, created_at::date
>   FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
>  WHERE item_type='undeployed_asset' AND spec->>'purpose' IN ('og_card','favicon')
>    AND status NOT IN ('complete','cancelled','rejected');
> ```
>
> **11 rows. `mode` is NULL on every one. Two are at status `detected`, which is
> dispatchable** — `triage_detect_items` promotes `detected` into the build queue.
> dartsonline.com ×2 (2026-07-29), robot-hands.com ×9 (2026-07-18/19), all predating the
> `bugs_open/142` fix.
>
> And `asset-deployer`'s `check_mode` step only diverts `input_data.mode == "brand_head"`.
> These items have no mode, so they **fall through to `deploy_asset` → `deploy_image_asset`
> with `purpose=og_card`** and `asset_key` NULL. Under the pre-168 code that wrote
> `og_card.png` — litter. Under 168's unified derivation it writes **`og-card.png`: the live
> social card, replaced, by a git commit that runs before any lock or provenance guard.**
>
> **So the council's `bug_historian` (high) and `guardian` (medium) seats were right, and my
> round-2 rationale contained a false reachability claim.** The guard is shipped, not
> deferred: `deploy_image_asset` now REFUSES a brand-head purpose (finding B's candidate 1),
> before the storage-URI resolution and before any download or commit.
>
> **There is no exposure window,** which is the one thing that went right by accident: the
> path unification is not live either (pod-verified on `v1.0.1228`), so the clobber and its
> refusal ship in the **same image**.

~~**Why nothing is broken today — measured, not assumed (2026-08-02):**
`check_undeployed_assets` excludes brand-head purposes from its generic half
(`AND NOT (COALESCE(a.purpose,'') = ANY($2::text[]))`) and routes them to
`needs_brand_head_assets`, whose repair is **re-derivation**. And every variable-purpose
reader resolves through `site_plan_imagery`, whose reachable purpose set is
`hero`/`icon`/`illustration`/`sprite_sheet` — **163 rows, zero brand-head** (denominator run:
the same join unfiltered returns those 163). So no live path routes a brand-head purpose to
this action.~~

## Fix candidates, ordered by what closes the door

1. ~~**Make the deployer refuse a purpose it does not own.**~~ **DONE 2026-08-02, shipped in
   the same commit as the change that made it necessary** (council `abd9b119` round 2, gating
   objection from `bug_historian`). `deploy_image_asset` returns a refusal — a completed
   result carrying the reason, not an error, so the item resolves instead of retrying against
   a guard that will never let it through — when `storage.IsBrandHeadPurpose(purpose)`, before
   the storage-URI resolution and before any download or commit. Pinned by
   `TestDeployImageAssetRefusesBrandHeadPurposes`, which asserts the guard EXISTS, that it
   precedes `DownloadOptimizeAndPrepare` / `sendGitCommitRequest` / `DeployedAssetPath`
   (a guard that fires after the commit is not a guard), and that it returns a refusal rather
   than an error. Both properties mutation-proven: deleting the guard fails it, and *moving it
   after the download* fails it on ordering alone. This also gives `IsBrandHeadPurpose` — which
   168 had left with **zero production callers** — a real job again, closing the `reuse_agent`,
   `guardian` and `prior_art_librarian` notes about a dormant predicate.
   **This leaves the 11 queued items harmless rather than repaired**: they will now resolve
   with a reason instead of clobbering. Whether they should also be re-pointed at
   `mode=brand_head` (re-derivation, which is what they actually want) is a data question left
   for the owner — the code no longer depends on the answer.
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
