# PLAN 2026-08-22 — wiring generated imagery to the pages it was made for (`bugs_open/114`)

Lane opened 2026-08-22. Bug: `bugs_open/114_HANDOFF_2026-07-27_generated_imagery_is_deployed_and_never_referenced.md`.

## What this lane is for

Imagery is planned, generated, deployed — and then nothing points a page at it. Five
lanes have contributed evidence to 114 since it was filed; none took the fix. This lane
takes the **framework-wide** half: the writers and the convergence, not the per-site
repair.

## Validity re-check (2026-08-22, before any work)

All `[MEASURED]` today unless marked. Queries in `RUNBOOK_imagery_wiring.md`.

| claim | measured today |
|---|---|
| fixture unwired | mortgagecalculator's 10 `content_hero_tool_*` assets: active, serving 200, **0 rendered references**, entity links NULL |
| entity links absent fleet-wide | **518 of 580 as of 2026-08-22** `assets` rows have `entity_type` NULL. Last 14 days: only `card` (45) linked; `hero` 66, `content_hero` 45, `icon` 19, `favicon` 7, `og_card` 7, `logo` 5 all NULL |
| the poisoned default spread | `sites.content_data.hero_url='/assets/images/hero.jpg'` on **18 sites** (10 when 114's 07-29 contribution measured it) |
| the 07-29 repair was undone | fundamentallyai was repaired to `hero-home.jpg` on 07-29; it reads `/assets/images/hero.jpg` again today |
| that path is dead on some sites | HTTP probe: relojistas.com **404**, vonc.com **404**, fundamentallyai.com **404**; gamesdesign/idea/oufe/webdesign.co.uk/mortgagecalculator **200** |
| the queue mostly drains now | `image_landed`: 40 complete / **8 needs_human_review** / 7 cancelled / 3 wont_fix / 1 failed (was 14 parked / 13 complete at filing) |
| the class is fleet-wide | content_hero assets vs wired components: dartsonline 20/12, robot-hands 16/7, **gamesdesign 14/0, finetuning 14/0**, mcalc 10/2, leopardess 7/0, fundamentallyai 6/0, ai-agent-orchestration 4/0, idea.uk 3/2 — **23 of 94 wired, as of 2026-08-22** |

**Verdict: still valid, and larger than filed.**

## Root causes, and which are proven

### GAP 1 — `store_asset` poisons the site-wide fallback. PROVEN, code-level.

`StoreAssetAction` (`platform/orchestration/actions/v3_site_actions.go:3453-3464`) writes
`sites.content_data.<purpose>_url` on **every** store, deriving the value with
`storage.BuildAssetPaths(purpose, ext)` — a **purpose**-derived path, always
`/assets/images/hero.jpg` for `purpose=hero`. The deployer commits the file under the
**asset-key**-derived name (`deploy_image_asset_action.go:403` →
`storage.DeployedAssetPath(assetKey, purpose)` → `/assets/images/hero-about.jpg`,
`/assets/images/content-hero-tool-repayment.jpg`).

So **every page-scoped hero generation re-stamps the site-wide default with a path that
may exist on no site.** That is the mechanism behind the 10 → 18 site spread and behind
fundamentallyai's repair being undone.

The workflow already tries to prevent it: the imagery store step passes
`update_site_brand_assets: false` (`docs/agent_docs/sql_for_agents/107_image_build_handler.sql:1163-1170`).
**That key is dead** — `grep -rn "update_site_brand_assets" --include="*.go" platform/ internal/`
returns nothing. A config key that nothing reads, exactly the class in
`MEMORY.md`'s `a-config-key-that-nothing-reads`.

### GAP 2 — the entity link is never written at generation. PROVEN, exhaustive.

The **only** writer of `assets.entity_type`/`entity_id` in the codebase is
`derive_card_asset_action.go:214-228`, and it hardcodes `purpose='card'`.
`StoreAssetAction`'s two INSERTs (`v3_site_actions.go:3371-3373`, `:3419-3421`) omit both
columns and expose **no config key** to supply them. Both readers require the link and
`purpose='card'`: `queryresolve.go:370-372` (`pageImageJoins` — every listing card on the
fleet) and `check_content_image_missing.go:219-221` (the sweep's `ca` join).

`check_content_image_missing`'s GENERATE arm has the page id in hand
(`check_content_image_missing.go:236`, used at `:269`) and omits it from the work-item
spec (`imageryplan.BuildSpec`, `platform/orchestration/imageryplan/imageryplan.go:194-219`).

### GAP 3 — convergence is sweep-driven and the sweep lane is dead. PROVEN.

The check's design (its own header, `:23-25`): pass 1 GENERATE hero → pass 2, a later
sweep, sees no card and files DERIVE → the derive writes the entity link → pass 3 silent.

`site_discovery_rotation`, measured today: **design-discovery-agent's newest
`last_selected_at` is 2026-08-11** (22 rows, 08-09..08-11) while every sibling lane is
current (availability 08-22, completeness/quality/render-audit 08-21). The
`site-discovery-staleness` CronJob (`bugs_open/230`, SCH-025) has been reporting the
design lane's absence from "stamps advanced last 24h" **daily, and nobody reads it**.

So a one-shot generation never converges. mcalc has **0** card assets and its 10 heroes
have sat unlinked since 08-15.

> The rotation stall itself is `bugs_open/230`'s lane — **report, do not fix**. What this
> lane owes is a convergence that does not depend on a sweep running at all.

### GAP 4 — the resolver fires for some pages and not others. NOT YET EXPLAINED.

Lane B (`plan_sections_action.go:426-448`) looks up the asset by
`imageryplan.ContentHeroKey(pageName)` with no plan row and no entity link needed. It
**does** work — 23 components fleet-wide carry a `content-hero-*` path.

The natural experiment, on one site, one day, one flow:

| page | asset created | component rendered | gap | outcome |
|---|---|---|---|---|
| tool-repayment | 19:31:38 | 19:47:50 | 972 s | fallback |
| tool-equity-release | 19:33:13 | 19:50:24 | 1031 s | **WIRED** |
| tool-portfolio | 19:36:33 | 19:53:52 | 1039 s | fallback |
| tool-fee-analyser | 19:38:20 | 20:20:02 | 2502 s | fallback |
| tool-bridging-loan | 19:39:57 | 20:22:20 | 2543 s | fallback |
| tool-overpayment | 19:41:28 | 20:25:02 | 2614 s | **WIRED** |
| tool-rate-forecaster | 19:43:30 | 20:27:41 | 2650 s | fallback |
| tool-affordability | 19:37:54 | **2026-08-18** 00:20 | 2.2 days | fallback |

Ruled out by measurement, not by argument:
- **not a race** — every asset was `active` 16+ minutes before its render, and the
  2-days-later one still missed;
- **not the plan routes** — mcalc's current plan has **no** page-scope hero row for any
  tool page and **no** site-scope hero row at all (only `logo`), so routes 1 and 3 are
  empty for all eight;
- **not the asset shape** — all ten keys are `content_hero_tool_*`, purpose
  `content_hero`, status `active`; the Lane B predicate replays and matches today.

That leaves *which code path each render took* and *what pageName each saw*. The
08-15 orchestration rows are purged, so it cannot be settled from history.

**Filed to the diagnosis loop rather than guessed** (CLAUDE.md: file before asserting a
mechanism). Intake corr `23da0760-f2da-4095-967e-2bdd308aa7ea`, **run corr
`ea7dfeef-c11d-40c4-b24f-b8f42413b1ae`** (the run correlation is the artifact key).

### GAP 5 — nothing detects the state. PROVEN by reading the checks.

- `check_image_url_404` — fixed by `bugs_closed/128`, compares exact deployed paths now.
  It answers "is this path unbacked?", not "is the right image on this page?"
- `check_placeholder_image_in_use` — narrowed 2026-08-12 to the canonical `asset_key`.
  When it does fire it files `needs_hero_image`, i.e. **generates a new site hero**
  rather than wiring the page-scoped one that already exists.
- `check_content_image_missing` — sweeps only `page_type` `blog-post` and `tool`
  (`:129-146`), so `content` pages are outside every producer (114's 08-17 finding).

Nothing anywhere asks: *this page has a page-scoped asset, deployed and serving, and the
page points at the generic site fallback instead.*

## The plan

Three sequenced tasks, each a coherent council submission. **Go image first, then config
and migrations** — config is live the moment it is applied, Go is inert until a roll.

### Task A — stop the poisoning (GAP 1), then repair the 18 sites

1. `StoreAssetAction`: write the site-wide `<purpose>_url` **only when
   `asset_key == purpose`** (a page-scoped asset must never set site-wide brand state);
   derive it with `storage.DeployedAssetPath(assetKey, purpose).RelativeURL` (the path the
   deployer actually commits — one derivation rule, the principle `bugs_closed/168`
   established); and make `update_site_brand_assets: false` **actually skip the write**.
   Default absent = write, so pageflow-builder's brand stores are unchanged.
2. Test + **mutation proofs**: remove the asset-key gate → a test must fail; ignore the
   config gate → a test must fail. A guard nobody has watched fail is not a guard.
3. Concept-register entry in the same commit; `LANDMINES.md` entry for the
   `BuildAssetPaths` vs `DeployedAssetPath` divergence, then
   `./scripts/landmines-verify-dispatch.sh`.
4. **After the roll is proven at the artefact**: the repair migration, held
   (`_HOLD.sql`) because running it before the gate is live just invites the next store
   to re-poison. Per site: keep `/assets/images/hero.jpg` where an active canonical
   `hero` asset makes it correct; otherwise repoint to the site's real deployed hero, or
   NULL it (the resolver skips an empty value). Idempotent; `DO`/`RAISE` verify block,
   never bare `SELECT`s — a non-empty result cannot stop a `COMMIT`.
5. Verify with a **demand control**: a page-scoped store that must NOT move
   `content_data`, and a canonical store that must.

### Task B — the entity link at generation, and event-driven convergence (GAPs 2+3)

1. `StoreAssetAction`: optional `entity_type` / `entity_id_field` config, mirroring the
   existing `purpose` / `purpose_field` pair. Opt-in, absent = today's NULL behaviour.
   RFC_022 posture: opt-in, unsafe default OFF, no live consumer until the held config
   migration lands — not architecture-scope, and the consumer enumeration goes in the
   submission rather than being asserted.
2. `check_content_image_missing`'s GENERATE arm: carry `entity_type`/`entity_id` in the
   spec (spec keys, added at the call site — do not widen the shared `BuildSpec`, whose
   plan-driven callers have no page id).
3. `flag_page_image_rebuild`: when the landed item is a `content_hero`, file the DERIVE
   item then and there, in the exact `contentImageSpecJSON` shape and under the existing
   `contentImageItemKey` for dedup. Pass 2 stops waiting for a sweep.
4. Held config migration mapping the new keys into image-build-handler's store step.
5. Acceptance is the mcalc fixture, end to end, **at the artefact**: entity link set,
   derive item filed with no sweep, card asset created and linked, page carrying the
   content-hero path on the **served** page, and a re-run of the check filing nothing.

### Task C — detection, hygiene, corrections

1. A flag-only check for GAP 5: a page-scoped asset exists, deployed, and neither the
   page's `content_data` nor its `rendered_html` references it.
2. The 8 parked rows: 5 unsatisfiable (reuse the `300_sweep_187…` pattern), 3 marked
   "satisfiable now" by the 08-21 revalidation → re-triage.
3. The remaining fixture pages re-fired once GAP 4 is understood.
4. A dated correction appendix on `bugs_open/114` (it carries four stale claims — see
   NOTES).
5. Report, do not fix: the rotation stall (`bugs_open/230`), and
   `check_undeployed_assets`'s underscore-vs-hyphen pattern (verify first, then file).

### Owner decisions — flagged, not taken

- Widening `check_content_image_missing`'s surface to `page_type='content'` is
  fleet-wide image-generation spend.
- Whether the 5 unsatisfiable parked rows should be cancelled or their pages built.

## Out of scope, with owners

`bugs_open/230` rotation stall · in-body imagery durability (238-class,
`inline_guide_imagery` lane) · mcalc's deferred `tools-index` imagery item (214's
residual — **must not be undeferred**, it would mint a fresh 114 instance) ·
`bugs_open/235`/`236`/`209`/`152` (adjacent asset defects) · component-schema fallback
literals.

---

## REVISION 2026-09-02 — resumption: what the 08-22 plan got right, what is superseded, and the remaining work

Evidence for every claim here: `NOTES_imagery_wiring.md`, 2026-09-02 entry.

### Corrections to this plan, dated, not edited away

- **GAP 4 is substantially explained and it was not a resolver defect.** The natural
  experiment's cohort was all `page_type='tool'` — pages whose `hero` component row is a
  misidentified tool fragment (`bugs_open/357` / RFC_046): the row declares `hero` and
  stores the tool shell, so wired-vs-fallback measured at `content_data` was noise on rows
  that render neither. The UNVERIFIABLE diagnosis verdict was right to refuse it. Non-tool
  wiring failures are `bugs_open/412`'s account (owned, active lane).
- **Task B's remaining halves are RETIRED as designed, not "not yet built".**
  (1) `StoreAssetAction` entity-link config: NOT building it. The convergence it was for
  is achieved event-side — 193/193 natural emitter firings produced entity-linked cards —
  and the config keys would be two more optional keys against the RFC_022 budget with no
  live consumer. The entity link stays owned by `derive_card_asset`, single-writer.
  (2) GENERATE-arm spec keys: same reasoning; the emitter carries `entity_id` in its spec
  already (`ContentImageSpecJSON`), which is what the acceptance test consumed.
- **Task C1's "flag-only check" survives but its design changed twice today**: per-page
  flags would flood a review queue with no working surface (`bugs_open/033`), and a
  capability gate on GENERATE was rejected (it would trade away the card-derivation
  benefit). The check below files **one rollup item per (site, state)** instead.

### Remaining work, in order

**C1 — `check_unrendered_page_imagery` (the missing detector, GAP 5).** DB-only,
flag-only. Population: active assets under `imageryplan.ContentHeroKey(page)` for the
site's pages (the exact class this bug measures). A page whose deployed content-hero path
appears nowhere in its deployed components' `rendered_html` is classified:
  - `unwired` — an image-capable component exists and is not a fragment: the asset is
    deliverable; remedy exists (412's deploy-time wiring / a rerender once that lands);
  - `no_image_slot` — no component template on the page carries an image branch: the
    composition gap (`editorial_design_uplift`'s 189-page census, made standing);
  - `fragment_slot` — the capable component's row is a 357 fragment: counted, cited to 357,
    NOT actionable here.
One `needs_human_review` item per (site, state), no handler, dedup key
`unrendered_page_imagery:<state>`, spec carries count + measured-at date + up to 12
examples + the re-runnable census query pointer. Uses `Resolved` (narrow, ItemKey) to
retract a state's rollup when it empties — positive observation only.
Tests + mutation proofs; register entry same commit; council submission before/alongside.

**C2 — residual poisoned keys migration.** Delete `icon_url`/`content_hero_url`/
`illustration_url`/`sprite_sheet_url` from `sites.content_data` (27 site-key pairs, zero
canonical assets behind any, no Go reader, resolver fallback reads only
hero_url/logo_url). 562's pattern: per-row backup, DO/RAISE guard, idempotent.
`logo_url` untouched (26/26 legitimate).

**C3 — council debt.** Resubmit 562 (`RESUBMIT_CORR=4145fcdc…` — verdict never produced,
no orchestration row). New submission for C1+C2.

**C4 — queue hygiene.** 7 parked `image_landed` rows (4 robot-hands tool pages —
almost certainly 357-shaped, reval `unknown`; 3 `still_holds`). Disposition recorded,
not silently cancelled. The 7 `failed` page_rerenders (08-27..31) belong to their sites.

**C5 — communications** (fixes not in isolation):
  - `bugs_open/412` (finetuning lane): contribute the 357-connection + the
    undeployed_asset backlog evidence; state that C1 cites their fix candidate 1 as the
    unwired-state remedy and ask whether they want the deploy-time wiring built here or
    kept (it is their candidate; our IMG-073 is its sibling).
  - `mortgagecalculator_couk_adoption`: their §2 "diff the render path" question is
    answered — 357, not a render-path divergence; saves them the dig.
  - `editorial_design_uplift`: their one-shot 189-page census becomes C1's standing
    `no_image_slot` state.
  - `inline_guide_imagery`: the illustration_url deletion touches nothing they resolve
    (site_assets.illustration reads plan/assets, not content_data) — stated so they need
    not re-derive it.
  - `bugs_open/384` lane: no action needed; their PBP-048 re-resolve is receiving the
    emitter's cards (193 naturally-derived cards flowed through it).

### The closing bar, restated against today

1. natural landing files the derive item without a sweep — **MET** (193, evidence in NOTES);
2. resulting card entity-linked + file serves on unhurried probe — **MET** (193/193; two probes 200);
3. no new poisoned `content_data.<purpose>_url` — **MET** (census + apis.uk discriminator);
4. detection check exists or the decision recorded — **OPEN → C1 above.**

When C1 is live (or its verdict recorded), 114 moves to `bugs_closed/` with the residual
states explicitly handed to 357 (fragment slots), 412 (unwired delivery), and the owner
(composition/no-slot population).

### ADDENDUM 2026-09-02 late — two tracked follow-ups out of the council round, and the 412 build

- **C1 amendment:** the detector was REVISE'd round 1 (a landmine cited unread — see
  NOTES) and resubmitted round 2 with 709 split out; 709 travels standalone, hardened.
- **NEW follow-up (architecture seat, advisory): narrow or retire
  `check_undeployed_assets` half 1** once IMG-077 rollups are live — two overlapping
  mechanisms must not coexist indefinitely, and its parked backlog grows uncapped
  (1,662 at the council's re-measure). The backlog's DISPOSITION is an owner decision;
  the check's narrowing is this lane's follow-on work.
- **DONE under `bugs_open/412` (their §10 handover): candidate 1 built** — IMG-078,
  opt-in default OFF, migration 710 held. 412 §11 is the build record; acceptance
  protocol excludes finetuning.uk and runs on IMG-077 `unwired` sites.
- **The deferred half of 412 (candidate 2, light-vs-heavy emission) remains theirs/ours
  jointly AFTER the floor is proven** — deliberately not in any current submission.

### ADDENDUM 2026-09-02, last — RFC_063 RULED (option B) and this lane CLAIMS the imagery seeding step, with its preconditions stated

The owner ruled tonight: the six plan-less sites converge into the plan tables, with a
scoped hand-insert exception (closed backfill only — the ruling's verbatim scope is in
RFC_063's OWNER RULING appendix, `01a3b96ac`). **This lane claims the IMAGERY SEEDING
step of the conversion, and only that step:**

- **What it is:** per converted site, `site_plan_imagery` page-scope hero rows seeded
  FROM THE ASSETS TABLE (`key = existing active asset_key` under the ContentHeroKey
  convention) — never from page enumeration, per this lane's own spend-trap finding
  (`check_unfulfilled_imagery_plan` generates for any plan row whose key lacks an
  active asset). Four sites have assets to seed (ai-agent-orchestration 17,
  finetuning 14, loancash 9, lampenkap 6); gaswholesalers and cookly seed ZERO rows.
- **Sequenced strictly after:** (1) the ruling's unwaived one-site reconciler-skip
  proof, and (2) the composition half creating each site's current `site_plans` row
  (imagery rows hang off `plan_id`). Both belong to whoever takes the composition
  half — NOT claimed here.
- **Why this lane:** it wrote the two-row minimal-materialisation spec, owns the
  route-1 consumer (IMG-078), and the seeded rows are the wiring round 2's plan-less
  arm arriving "for free" exactly as the RFC contribution predicted. The seeding SQL
  will go through the council as a migration like everything else here.
- **finetuning.uk stays excluded from IMG-078 ACCEPTANCE** even once converted — the
  664/649 attribution overlap (412 §11) is unchanged by the ruling.
