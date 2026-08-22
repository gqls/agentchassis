# NOTES — bugfix 114 imagery wiring

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-22 — lane opened, bug re-validated, one mechanism filed to the loop

### Ownership check first

`scripts/who-owns.py 114` names three lanes that **cite** 114 —
`mortgagecalculator_couk_adoption`, `brochure_component_library`,
`bugfix_284_flag_only_items_promoted` — all contributors, none owning the fix. The only
in-flight `needs_diagnosis` item (`554ea3d2…`, status `diagnosing`) is the
OrphanPagesCheck lifecycle-axis bug, unrelated. `git log` on the bug file shows five
contributions 08-13 → 08-17 and no fix commit. Taking the fix.

### Validity

Re-measured everything the file asserts (queries in RUNBOOK). It is still valid and
bigger: 18 sites carry the poisoned default (was 10), 518/580 assets are entity-unlinked,
and only **23 of 94** content_hero assets fleet-wide are wired to a component.

The parked-queue half has changed shape since filing and the file does not say so: 40
`image_landed` items are now `complete` against 8 parked, where the file recorded 14
parked / 13 complete. The remaining 8 are pages with no built slots, not a queue nobody
drains — an emit-time guard (`flag_page_image_rebuild_action.go:147-160`) now skips
sectionless pages, and a revalidation pass stamped verdicts on 08-21.

### The finding that reframed the bug

`StoreAssetAction` writes `sites.content_data.<purpose>_url` on **every** asset store,
using `storage.BuildAssetPaths(purpose, ext)` — purpose-derived, so always
`/assets/images/hero.jpg`. The deployer commits under
`storage.DeployedAssetPath(assetKey, purpose)` — asset-key-derived. **Two derivation
rules for one artefact**, so every page-scoped hero generation re-stamps the site-wide
default with a path that may exist nowhere.

That explains something the bug file recorded as a mystery: fundamentallyai was repaired
on 07-29 and reads `/assets/images/hero.jpg` again today. It was not re-broken by a
person; the next hero store re-poisoned it. It also explains 10 sites → 18 sites.

And the workflow **already asks for this not to happen**: the imagery store step passes
`update_site_brand_assets: false`
(`docs/agent_docs/sql_for_agents/107_image_build_handler.sql:1163-1170`).
`grep -rn "update_site_brand_assets" --include="*.go" platform/ internal/` → **no hits**.
The key is dead. Somebody saw this coming and wrote a gate the action never read.

### The natural experiment, and what it rules out

Eight mcalc tool pages, one site, one day, one flow, all with an active
`content_hero_tool_*` asset. **Two wired, six on the fallback.** Full table in PLAN.

Ruled out by measurement:
- **race** — every asset was `active` 972–2650 s before its render, and `tool-affordability`
  rendered **2.2 days** later and still missed;
- **plan routes** — mcalc's current plan has no page-scope hero row for any tool page and
  no site-scope hero row at all (only `logo`), so routes 1 and 3 were empty for all eight;
- **asset shape** — all ten keys match the `ContentHeroKey` convention, purpose
  `content_hero`, status `active`; the Lane B predicate replays and matches **today**.

So Lane B works (23 wired components fleet-wide prove it) and something about *which
path each render took* differs. The 08-15 orchestration rows are purged.

**Filed to the diagnosis loop rather than guessed.** CLAUDE.md's rule is that a durable
mechanism claim goes through 090 before I assert it, and "it feels obvious" is exactly
the signal that is worthless here. Intake corr `23da0760-f2da-4095-967e-2bdd308aa7ea`,
run corr **`ea7dfeef-c11d-40c4-b24f-b8f42413b1ae`**.

### MISSTEPS

1. **Wasted a 090 intake on unescaped double quotes.** I wrote `assets["hero"]` in the
   symptom to name a map key. The trigger interpolates the symptom into a `$json$` SQL
   literal without escaping, so it aborted with
   `invalid input syntax for type json … Token "hero" is invalid`. Cheap check: write
   symptoms with **no double quotes at all**. Recorded in RUNBOOK with the error text so
   the next session recognises it. → `WRONG_CALLS.md`, and it is a LANDMINE candidate
   (a trap on a specific command where the failure is a confusing type error, not a
   "your quoting is wrong" message).
2. **Three round trips lost to guessed column names**: `site_work_items.attempts` (it is
   `attempt_count`), `orchestration_states.id` / `.agent_type` (they are
   `orchestration_id` / `owner_agent_type`), `site_discovery_rotation.last_run_at` (it is
   `last_selected_at`). CLAUDE.md says `\d <table>` before writing SQL and I did not.
3. **I nearly fired a live full page-build on another lane's active site as a "canary".**
   The approved plan called for it. On re-reading, an `image_landed` item routes to
   page-build-handler, which is the **full LLM build** — it would have rewritten
   mortgagecalculator's tool-repayment content while the mcalc adoption lane is active
   (92 commits/14d). The read-only census answered the same question better: it gave me
   eight cases instead of one, and the two positives are what made the experiment
   informative. **A destructive canary that yields one bit is worse than a census that
   yields eight.** Not run.

### Adjacent things found, deliberately not fixed here

- **The design-discovery rotation has been stalled since 08-11** (`site_discovery_rotation`:
  22 rows, newest `last_selected_at` 2026-08-11, while availability/completeness/quality/
  render-audit are all 08-21 or 08-22). That is the lane which runs
  `check_content_image_missing`, so the sweep-driven convergence this bug depends on has
  been dead for eleven days. The `site-discovery-staleness` CronJob (`bugs_open/230`)
  reports it **daily** — the design lane simply never appears in "stamps advanced last
  24h" — and nothing consumes the report. Belongs to 230's lane; contributing the
  measurement, not the fix. It is also the argument for making convergence event-driven
  rather than sweep-driven, which is this lane's Task B.
- `check_undeployed_assets.go:289-305` matches `rendered_html` against the **underscored
  purpose** (`content_hero.` / `content_hero-`) while deployed files carry the
  **hyphenated key** (`content-hero-tool-x.jpg`). `[UNVERIFIED]` — needs one query
  before filing; if it holds, that check cannot see any content hero as deployed.

### Stale claims in `bugs_open/114` itself (correction owed)

1. The ADDENDUM says the LLM-free rerender path "does not re-resolve fields", citing
   `flag_page_image_rebuild_action.go`. The header says that of the **terminal assemble
   leg**; `rerender_page_sections` **does** re-resolve (`:20-23`, `:459`).
2. Its `plan_sections_action.go:1608` citation is stale — that comment is now at
   `:2424-2432`.
3. Its merge-order claim (site-wide `hero_url` beats per-page `content_data.hero_url`) is
   contradicted by source: fresh `plan.ResolvedData` is merged **last** and wins
   (`rerender_page_sections_action.go:614-620`); the base only wins when the fresh data
   carries no hero.
4. "Why nothing caught any of it" quotes `check_image_url_404`'s old header. That check
   was fixed and closed (`bugs_closed/128`, live v1.0.1219) — it compares exact deployed
   paths now. The half of the claim that still stands is class (c): paths outside the
   `/assets/images/` prefix remain invisible to it.
