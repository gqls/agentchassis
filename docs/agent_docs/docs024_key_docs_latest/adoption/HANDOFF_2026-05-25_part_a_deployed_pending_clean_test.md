# HANDOFF — 2026-05-25 — Part A deployed, awaiting a clean first-plan test

## One-paragraph status

The adoption→build pipeline flattens section-index pages: an adopted
`games-index`/`tools-index` ends up as a flat `games`/`tools` page at
`/games.html` instead of a section hub at `/games/index.html`. Root cause is
fully traced (below). A fix — **Part A**, a name-suffix rule in `ValidateRoles` —
is written, unit-tested green, and **deployed**, but has **never been tested on a
clean run**: every attempt got muddled by stale plans, mid-cascade reads, an
undeployed image, or a teardown that deleted the wrong site row. The next chat's
job is simple and specific: **clean-adopt once, on confirmed-Part-A code, and read
one query** to see whether `games-index`/`tools-index` survive as section indexes.

---

## The confirmed root cause (do not re-litigate this — it's evidence-backed)

The LLM is faithful and the convergence works. Corruption happens in
`WriteSitePlanAction` via a two-step interaction in
`datahelpers/page_canonical.go`:

1. The plan-builder LLM emits e.g. `{name:"games-index", page_type:"content"}`
   with **no `slug`, no `url`, no `parent_section`** (confirmed in the
   `plan_site` `response_text`; strategy_notes even says "preserving all 20
   existing pages exactly").
2. `ValidateRoles` derives the slug from the name, stripping `tool-`/`guide-`/
   `game-` prefixes **and** the `-index` suffix → slug `games`. With no url and
   no declared parent, none of its old rules (2 declared-parent, 3 URL-pattern,
   4 nested-URL) fire, so the role stays the raw `content`.
3. `CanonicalisePage(role="content", slug="games")` re-adds a prefix only for
   `tool`/`game`/`guide`/section-index roles — **not** for `content`/`blog_post`.
   So `content` + `games` → flat `games` at `/games.html`. Permanent strip.

Why `tool-*`/`game-*` survive: their role re-adds the prefix
(`ttk-calculator` → `tool-ttk-calculator`), so the strip round-trips.
Why `guides-index` survives: typed `blog-index`, it hits the section-index
branch. The single upstream origin is the **`page_type` assigned at adoption**:
`games-index`/`tools-index` come through as `content` (should be a section-index
type). `analyze_site` (adoption LLM) typed them `content`; everything downstream
faithfully keeps it.

Recorded in `016_debugging_guide_v2_21_.md` ("Adoption faithfulness: LLM +
convergence are faithful; WriteSitePlanAction strips identity for
content/blog_post types"). The genre is `ARCHITECTURAL_TENSIONS.md` #1.

---

## Part A — what's deployed, and what it does

**Files (deployed):** `page_role_validator.go` + `page_role_validator_test.go`
(in `/mnt/user-data/outputs/`). One new rule in `ValidateRoles` (Rule 2):

```go
} else if strings.HasSuffix(v.Name, "-index") && !isLeafRole(rawRole) {
    correctedRole = "section-index"
}
```

plus an `isLeafRole` helper (`tool`/`guide`/`game`/`blog-post`/`entity-page`).
A page whose **name** ends in `-index` becomes a section hub — reading the signal
the LLM emits reliably (the name) instead of the url/parent it omits. Guard stops
an explicit leaf with an odd `-index` name being clobbered.

**Verified mechanics:** role-only fix is sufficient. `ValidateRoles` → role
`section-index`, slug `games` → `CanonicalisePage` (via `normalisePageType`
kebab→snake → `section_index` → `isSectionIndexRole` true) → rebuilds
`games-index` + `/games/index.html`. Tests: 4 pre-existing pass (their inputs
carry urls/parents so the new branch doesn't fire — no regression), 3 new pass
(production-shape recovery, leaf guard, idempotency). `gofmt`/`vet` clean.

**Part A does NOT fix:** guides de-prefixing (`guide-rng-design` → `rng-design`)
— no `-index` suffix; that's a separate product call (faithful URL vs canonical
blog slug). Nor the hard-coded vertical vocabulary (`tools`/`guides`/`games`) in
`nestedRoleFromURL` — that's Part B.

---

## THE NEXT STEP (this is the whole task for the next chat)

A clean first-plan test of Part A. Every prior attempt failed for an avoidable
reason; here is the exact recipe that avoids all of them.

### 1. Clean teardown — pin to `site_id`, run with NOTHING in flight

The previous teardowns keyed on `domain` and (a) deleted the wrong row when two
`gamesdesign.co.uk` rows existed, and (b) raced a live cascade so the verify
lied. Before tearing down, confirm quiet and single-row:

```sql
-- must return 0 rows (nothing running)
SELECT status, current_step, owner_agent_type, updated_at
FROM orchestration_states
WHERE site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND status NOT IN ('COMPLETED','FAILED')
ORDER BY updated_at DESC LIMIT 5;

-- note the exact id(s); if >1 row, delete each by explicit id
SELECT id, created_at FROM sites WHERE domain='gamesdesign.co.uk';
```

Then tear down using the explicit UUID(s) — replace every
`WHERE domain='gamesdesign.co.uk'` / `WHERE site_id IN (SELECT … domain=…)` in
the cleanup script with `site_id = '<uuid>'`. Keep `source_domain=` for the
library tables (style_collections, css_themes, palettes, typography_sets,
layouts). Verify:

```sql
SELECT
 (SELECT count(*) FROM sites WHERE domain='gamesdesign.co.uk') AS sites,
 (SELECT count(*) FROM site_plans sp JOIN sites s ON s.id=sp.site_id
    WHERE s.domain='gamesdesign.co.uk') AS plans;
```
Both 0 → clean. (User is cleaning down the DB before the next chat, so this may
already be done. Re-verify regardless.)

### 2. Confirm Part A is in the running image BEFORE triggering

The image must contain the deployed `page_role_validator.go`. Spawned jobs run
"latest" only *after* the build+push completes — a planner that spawns
mid-build runs the old image. Confirm the deploy landed (the chassis/agent pods
post-date the build) before the trigger.

### 3. Adopt once: `https://gamedesign.uk` → `gamesdesign.co.uk`

### 4. WAIT for the planner to actually run — do not read early

The planner runs well after adoption writes the 20 pages. Reading the plan early
shows empty/stale data (this fooled us repeatedly). Poll until a **fresh**
`site_plan` claim appears with a timestamp after the trigger:

```sql
SELECT item_key, status, claimed_at
FROM site_work_items
WHERE site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND item_key='site_plan_gamesdesign.co.uk'
ORDER BY claimed_at DESC;
```

### 5. The test — read the AUTHORITATIVE plan table

```sql
SELECT spp.name, spp.role, spp.url
FROM site_plan_pages spp
JOIN site_plans sp ON sp.id=spp.plan_id AND sp.is_current
WHERE sp.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND (spp.name LIKE '%-index' OR spp.name IN ('games','tools','guides'))
ORDER BY spp.name;
```

- **PASS:** `games-index` and `tools-index` present, `role=section-index`, urls
  `/games/index.html` and `/tools/index.html`; **no** flat `games`/`tools`.
  → Part A holds in production. Move to the parked items below.
- **FAIL (still flat `games`/`tools`) on confirmed-Part-A code:** Rule 2 didn't
  fire. The LLM name is reliably `-index`-suffixed (confirmed), so look at:
  build didn't include the file / planner path doesn't call this `ValidateRoles`
  / `v.Name` isn't the raw name. Pull the planner's `plan_site` `response_text`
  (confirm the LLM still emits `games-index`) and the `WriteSitePlanAction`
  correction logs.

---

## Reading-discipline rules (we broke these repeatedly — keep them)

1. **The authoritative output is `site_plan_pages` (joined to current
   `site_plans`).** Work-item names, pod `-w` snapshots, and mid-cascade empty
   `site_plans` are NOT the plan. Three wrong conclusions this session came from
   reading those instead. (Debug guide: "recurring debugging trap, part 2".)
2. **Confirm the run is complete before diagnosing** — the relevant `*_<domain>`
   work item `complete` with a post-trigger timestamp, or the orchestration
   `COMPLETED`.
3. **Filtering `llm_call_log` by date catches multiple runs.** `ORDER BY
   created_at DESC LIMIT n` can return the *previous* run's response. Match on
   the run's actual planner timestamp.
4. **Teardown by `site_id`, never `domain`, and never with a cascade running.**

---

## Parked findings (do NOT start these until Part A is verified; capture only)

Each is real, separate, and was deliberately NOT chased to avoid tangling threads.

1. **`guides-index` renders with no list component.** The LLM emitted
   `guides-index` with `sections: []`; games-index/tools-index got
   `game-list`/`tool-list` because the LLM gave them those sections. Component
   lists are driven by the page's `sections` array → `section_type` → selector
   (`queryresolve.go` resolves `query.pages_where_type:<type>` off the
   `page_type` column; confirmed from `page_components`: games/tools hubs have
   lists, guides hub has only `hero`). So **list rendering keys off `sections`,
   not the hub `page_type`** — which means the flavour-collapse (Tension #2) is
   **cosmetic for rendering** and deprioritised. The real gap is the planner
   emitting empty `sections` for the guides hub. Candidate fix (constrain-the-
   source, Tension #1): a deterministic "section-index hub gets its matching
   list component" rule rather than relying on the LLM to remember it.

2. **Tool-page hero CTAs point to non-existent pages.** Deployed
   `tools/ehp-calculator/index.html` has buttons "Launch EHP Calculator" /
   "Read the EHP Guide" linking to `/contact.html` and `/services.html` — the
   hero schema's `cta_url`/`secondary_cta_url` fallbacks, because their sources
   (`pages.contact`, `pages.services`) don't exist on this site. Silent
   fallback to vertical-inappropriate defaults → broken links on a deployed
   page. Same genre as Tension #1 (silent fallback). NOTE: the latest plan now
   includes a `contact` page, which may change this — re-check after a clean run.

3. **Dispatcher reliability.** `c90b7ce4`'s orchestration history showed many
   `build-dispatch-loop` `FAILED` on `process_item_iter_*` and a long run of
   `page-build-handler: complete_error` (13:59–16:38 on 2026-05-24), plus
   multiple adoption/strategy cascades stacking on one site row. Echoes the
   HANDOFF_2026-04-23 dispatcher-stall (Bug 1). Noise for the Part A test (which
   only needs the planner output) but a real reliability thread.

4. **Stale comment.** `render_js_snippets_for_site_action.go`'s header still
   describes `loadComponentCSSSnippets` as using `applies_to && $1::jsonb` and
   being "latently broken". User reports that `jsonb && jsonb` bug was fixed
   earlier — so the comment now misleads. 30-second cleanup: update/remove it.

5. **Tension #2 companion (flavour collapse) — confirmed cosmetic, withdrawn.**
   The originally-proposed "merge the two role-normalisers" is WRONG: `normaliseRole`
   (validator, routing-collapsed) and `normalisePageType` (canonicaliser,
   flavour-preserving) are intentionally layered. And finding (1) shows rendering
   doesn't depend on the hub flavour. No code needed unless a future case proves
   otherwise.

---

## Larger work, design-dependent (not scoped yet)

- **Part B — de-hard-code `nestedRoleFromURL`** (`tools`/`guides`/`games`
  baked in; breaks for vet/recipe/etc. verticals). Needs the generic
  section-leaf-role decision (`entity-directory`/`entity-page`). Belongs with
  the Tension #1/#2 work.
- **The page_type root cause at adoption.** `games-index`/`tools-index` typed
  `content` originates in `analyze_site`. Prior art exists:
  `patch_analyze_site_add_game*.sql` (v1–v3) in outputs — relevant to fixing the
  type upstream rather than only correcting it in `ValidateRoles`. Part A is the
  downstream safety net; fixing the source type is the structural complement.
- **Adoption-faithfulness 90-day lock window.** Currently the minimal lock-free
  first-plan route is live (faithful first plan, but the *next* planner run is
  free to restructure — no 90-day window). The full timed-lock project (053
  lock_expiry, 054-full, write_site_plan patch) is designed and held in outputs,
  NOT applied. Resume only after adoption beds in. See
  `FOCUS_adoption_faithfulness_via_locks.md`, `PLAN_lock_coherence.md`.

---

## Document index (current versions in /mnt/user-data/outputs/)

- `016_debugging_guide_v2_21_.md` — incident-level; has the confirmed root cause
  and the two debugging-trap entries.
- `ARCHITECTURAL_TENSIONS.md` — genre-level; #1 (trust LLM structural labels),
  #2 (identity derived in multiple places), corrected to current understanding.
- `page_role_validator.go` / `_test.go` — Part A (deployed).
- `FOCUS_planner_ignores_adopted_state.md` — the divergence mechanisms.
- `FOCUS_adoption_faithfulness_via_locks.md`, `PLAN_lock_coherence.md` — the
  deferred lock work.
- `cleanup_gamesdesign_co_uk.sql` — teardown (domain-keyed; pin to UUID per §1).
- `patch_analyze_site_add_game_v3.sql` — prior art for the upstream page_type fix.

## Standing context for the next chat

Pipeline: Go chassis, Kafka/saga agents. Post-adoption cascade: adoption →
classifier → strategist → briefing → **build-site-planner** → composition →
webdesign → page-build → rerender (doc 007 v4). Spawned jobs run latest code
post-build. Namespaces: `ai-persona-system`, `kafka`. Conventions: simple
workflows / complexity in Go; spawn sub-agents not subworkflows; ALWAYS check DB
schema before SQL; reuse existing functions; no `logger.Debug` (won't show);
think before prescribing, don't conclude from partial/mid-flight signals.
