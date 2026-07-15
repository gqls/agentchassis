# HANDOFF — vonc.com / "Spark": provocations-index CLOSED; next = provocation-card mini-lobby trim

**Created:** 2026-07-09  •  **Revised:** 2026-07-09 (method corrected: HTML patching is the rejected mechanism; bundle added as §4.0)
**Supersedes nothing.** Companion to: `RUNNING_NOTES_vonc_v2.md` (chronological log),
`RUNBOOK_phase2_provocation_js.md` (operative runbook — its top block is the "YOU ARE HERE"),
`016b_debugging_guide_merged.md` (APPLY THIS to the project — merged/cumulative), and the per-tool
`PLAN_*` / `NOTES_*` docs.
**Purpose:** stand alone for a fresh chat. Read §1–§4, collect §5, then do the work.

---

## §1 First actions (do these three things)

1. **Read §2 (orientation) and §3 (what is already true).** Do not re-derive; the ids in §7 are
   copy-paste values — never hand-type a UUID (two typo incidents this thread).
2. **Take dated backups** before any UPDATE (§4, step T0). The previous backup names are dated and
   must not be reused (`CREATE TABLE IF NOT EXISTS` against an old name silently no-ops).
3. **Collect §5's four dumps** (template, instance, loader, current index HTML) and paste them.
   The generated DOM is authoritative over anything written in a spec — verified twice this thread.

Standing instruction from the user, in force: work in reasonable step sizes; check the schema before
writing SQL; prefer structural fixes to quick ones; reuse/alter existing functions before creating
new ones; never treat a `0 rows` result as decisive until the query itself is cleared; no summary
documents unless asked.

---

## §2 Orientation — what this system is

**agentchassis** builds multi-page websites from a domain name. Go actions + Postgres + Kafka
agents. Every agent is an orchestrator owning a workflow of steps that call actions; workflows stay
thin, complexity lives in Go. Deployment: agents commit to a git `sites` repo → GitHub Actions →
Backblaze B2 (public site).

**vonc.com / "Spark"** is an AI daily-provocation platform: one charged Provocation per day, users
file a position, "the Gauntlet" scores the room, users get an Archetype. It is the live test site.

**The runtime-fill mechanism (central to everything here).** Some sections are deliberately EMPTY at
build time and filled in the visitor's browser:

- The component template is a shell whose slots are blank, and its `<section>` tag carries
  `data-runtime-fill="true"`.
- The page assembler drops sections with ≤10 chars of visible text — the marker exempts them
  (patch shipped in `rerender_single_page_action.go`).
- A **loader** (an IIFE) lives as a row in the `js_snippets` table. The **site-asset-renderer** agent
  concatenates all active snippets into `/assets/js/snippets.js`, which every page loads.
- The loader fetches `/data/provocations.json` and fills the shell's selectors, failing gracefully.
- The data file is currently hand-committed to the sites repo; a Phase-3 pipeline will emit it.

Three loaders are live (`3 active snippet(s)` in the bundle header): `provocation-card-loader`,
`lobby-grid-loader`, `provocations-archive-loader`.

---

## §3 DONE state — what is true and verified (do not redo)

**Index page (`/index.html`) — live, six sections:** hero, provocation-card, gauntlet-cta,
brief-explanation, lobby-grid, system-stats. provocation-card and lobby-grid are runtime-filled and
browser-verified.

**Provocations archive (`/provocations/index.html`) — CLOSED 2026-07-08.** Was a 404 for two weeks —
the destination of the header nav and every primary CTA. Now live: eight archive rows fill from
`archive.entries[]`, empty state hides, ghost row fixed. Final numbers: instance `rendered_len` 7671,
live `curl … | grep -c 'data-archive-template] { display: none;'` → 2.

**The provocations arc (worth understanding — it explains several standing rules):**
planned 2026-06-22, then **ten builds "completed" having built nothing**. Root cause was two layers:
(a) the planner emitted the page row but **no sections** — for exactly the two non-standard roles,
`blog-post` (legitimate; the blog pipeline builds those) and `section-index` (the defect);
(b) page-build-handler routes zero-sections to a step named **`complete_error`** — a
`complete_workflow` with `output_fields [page_content, site_record]` and success_message "Content
writer skipped — page has no sections defined". The writer never runs, so `page_content` is absent,
and **a work-item result carrying only `site_record` is the diagnostic signature of this path** (a
healthy run's `complete` emits `[sections_saved, deploy_result]`). An error path implemented as a
successful completion; the message never surfaces.
**Section sources for a page build, in order:** `site_specs` aspect `site_plan` (authoritative in
code) → `pages.sections` (fallback) → same-role sibling layout synthesis. The **`site_plan_sections`
table is NOT read by this path.** vonc has no `site_plan` aspect, so `pages.sections` serves it. The
unblock was `UPDATE pages SET sections = '["provocations-archive-list"]'`.

**The archive component (`70d6662a`) is the reference build.** Generated with the thread's lessons
baked in: `data-runtime-fill` emitted **at generation** (no post-hoc string surgery), **no inline
script at all** (the loader is external), llm header fields (nothing can defer), one hidden
clone-template row, visible empty state. Verified `has_marker = t`, `has_inline_script = f`. This is
the pattern for all future dynamic sections.

**Shared-component check — CLOSED, no contamination.** Components with `forked_from IS NULL` are a
cross-site shared library keyed by `function`. `brief-explanation` (`58363894`) is shared by vonc +
idea.uk + robot-hands.com. Our regeneration was safe because `store_generated_component`'s field-set
guard blocks drops/renames on shared components (the `fdd92ad4` incident: renaming a field in place
silently emptied every dependent) and ours was a pure 0→20 field addition. The shared base is
**neutral** — 18 llm fields, 2 generic static CTA labels, no Spark vocabulary. Rule: regenerate a
shared base **only** for neutral, purely-additive improvements; site-specific voice must **fork**
(`forked_from = base_id`).

**Plan-storage question — deferred, deliberately.** A table-first alteration of
`load_page_sections_from_spec` was drafted and then **withdrawn**: design doc 029 shows plan storage
is its **open Q1** ("site_specs aspects vs new table", lean = partitioned `site_plan_*` aspects +
a `reconcile_site_plan` action). Three shapes coexist in production: the legacy singular `site_plan`
blob aspect (five sites), 029's partitioned aspects (apparently unimplemented), and the
`site_plans` tables (the vonc-generation planner writes these AND dual-writes `pages.sections`).
The decision belongs to the planner/reconciler thread, made once. **User is reverting the repo file
with `load_page_sections_from_spec_action.ORIGINAL.go`** — note the revert only reaches the cluster
on the next chassis build/push. Evidence to contribute to Q1: the table path now exists in
production, post-dating the doc's lean.

`Categories:` (milestone, root-cause, decision)

---

## §4 NEXT TASK — provocation-card mini-lobby trim

**Why.** The index currently shows the same idea twice: provocation-card renders its own four-card
"mini-lobby" *and* lobby-grid renders the six-card arena below it. The agreed role split is
lobby-grid = the arena grid; provocation-card = the single headline provocation + stats + CTAs.
This is the last live redundancy on the index. Decision confirmed by the user on 2026-07-04.

**Scope (three artefacts change).**

1. **Template** `content_components.html_template` for provocation-card `6163ff14-9f94-4962-aa19-d2718eabdeb1`:
   remove the `div.pc-card-grid` block (it contains four `article.pc-card` elements, each with
   `span.pc-card-icon` / `h3.pc-card-title` / `p.pc-card-desc` — all empty, loader-filled).
2. **Loader** `js_snippets.js_content` where `name = 'provocation-card-loader'`: remove the
   `data.lobby` fill block (the `if (Array.isArray(data.lobby)) { … }` section). Keep the header,
   body, CTA and stat-strip fills.
3. **The live index instance** `page_components.rendered_html` for that component on page
   `b4d24f8e-fccd-49df-9dad-aa56a0b20a68` — because an index rerender is **assemble-only**: it
   reassembles from stored `rendered_html` and will happily redeploy the old markup. (This is exactly
   the trap that made the archive redeploy ship the unfixed HTML: `rendered_len` was unchanged.)

**Two consequences to handle deliberately, not discover later:**

- **CSS layout.** `.provocation-card-section .pc-container` is `grid-template-columns: 1fr 1fr` at
  `min-width: 900px`, because the card grid was the second column. With it removed, the section
  will render one column of content beside an empty one. The trim must also drop or change that
  media-query rule (to `1fr`) in the same edit.
- **The template's own inline script** (`querySelector('[data-component="provocation-card"]')` +
  `.pc-card` hover activation) becomes a no-op over an empty NodeList. Harmless, but it is dead
  code: either remove it in the same edit, or leave it and note it. (Related backlog item: the
  inline-script *extraction* pattern is already live for `gauntlet-interface`, `latest-news`,
  `tool-archetype-taster-quiz` — they carry `js_content` plus a `<script src=…>` reference and no
  raw inline block. provocation-card and lobby-grid cosmetics are candidates for the same treatment.)

**Method — SETTLE IT FIRST (§4.0); my earlier draft named the rejected mechanism.**
Doc 003 (Contracts) states: *"content_data is always the source of truth … if we only patched
rendered_html, the edit would be lost on the next re-render … **this is why HTML patching was rejected
as an edit mechanism**"*. So hand-editing `page_components.rendered_html` is a bridge at best — a
future re-render regenerates it from the template. (Our archive/marker edits touched **both** template
and instance, so nothing we shipped is at risk; but it is not the sanctioned path.)

Doc 003 names **two** re-render paths, and neither is the assemble-only rerender we have been
triggering:
- **full path** — `needs_page` → page-build-handler → page-content-writer (LLM regenerates copy);
- **light path** — `rerender_page_sections`, behind a **`page_rerender`** item: re-renders every
  section from the **existing `content_data`** plus freshly re-resolved fields, **no LLM**, through
  the same `RenderComponentAction` template path. *"A section with NULL `content_data` cannot be
  re-rendered this way, so the light path escalates the whole page to the full rebuild."*
- `rerender_single_page` (what `rerender-index-vonc.sh` triggers) **only assembles pre-rendered
  components** — confirmed in its own header. A template-only edit will not appear through it.

Also: **`fix_component_template_action.go` already implements a `remove_element` fix type** ("removes
HTML elements matching a pattern") — the exact operation this task needs. Its header says it does not
change `page_components` `rendered_html` content because *"content changes go through the
section-editor workflow"*. Reuse before create: that action, or the section-editor, may be the
sanctioned route.

**Open question that decides the sequence:** provocation-card is Mode-B (empty shell, runtime-filled),
so its `content_data` may be empty or NULL — in which case the light path escalates to a full rebuild.
Check before choosing (query in §5).

**§4.0 — run the bundle first.** `bundle_minilobby_trim.sh` (outputs) is primed to settle exactly
this: which action changes the template; which action re-renders instances from a changed template;
what a Mode-B/NULL-`content_data` section does; which work-item type raises each path; whether
hand-added attributes survive a re-render; and how a changed `js_snippets.js_content` reaches
`/assets/js/snippets.js`. It is read-only (dbcontext `\d` + capped SELECTs, then the pure assembler);
it triggers nothing. Run it, read the verdict, **then** pick the method below.

**Fallback method, only if the bundle says no supported path exists.** Dump the template offline, edit
it, `UPDATE content_components SET html_template = $tpl$…$tpl$` with the complete new text (no
multi-line SQL `REPLACE` of nested markup), verify by **length delta** (roughly −1,200 to −1,600 bytes
for the removed grid plus the CSS rule change), keep `data-runtime-fill` intact, and propagate with a
`page_rerender` item rather than by hand-patching the instance. Take dated backups first.

**Redeploy path (two separate deploys — they are not the same thing):**

- Loader change → `UPDATE js_snippets` → `bash scripts/initial_messages/210_vonc_trigger/083_trigger-asset-renderer-vonc.sh`
  → check `orchestration_states` COMPLETED → `curl -s https://vonc.com/assets/js/snippets.js | head -3`
  (still `3 active snippet(s)`) and confirm the `lobby` fill is gone from the bundle.
- Template + instance change → `bash scripts/initial_messages/210_vonc_trigger/rerender-index-vonc.sh`
  (assemble-only) → verify.

**Verification (artifact, never item status):**

```sql
-- instance shrank and kept its marker
SELECT LENGTH(rendered_html) AS rendered_len,
       (rendered_html LIKE '%data-runtime-fill%') AS marker_kept,
       (rendered_html LIKE '%pc-card-grid%')      AS grid_gone_expect_f
FROM page_components
WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'
  AND component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
```

```
curl -s https://vonc.com/index.html | grep -c 'pc-card-grid'          # expect 0
curl -s https://vonc.com/index.html | grep -o 'data-component="[^"]*"' # expect the same six
```

Browser: provocation-card still fills (eyebrow, headline, body, two CTAs, three stats), no four-card
grid, lobby-grid's six arena cards unchanged, footer intact.

**Acceptance:** six sections still deploy; provocation-card fills; no `.pc-card` markup anywhere on
the index; lobby-grid unaffected; no console errors.

`Categories:` (next-task, decision)

---

## §5 Data to collect first (paste these)

```sql
-- T0. DATED backups (never reuse an old backup name — IF NOT EXISTS silently no-ops)
CREATE TABLE _vonc_pc_backup_20260709 AS
SELECT * FROM page_components WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid;
SELECT COUNT(*) AS pc_rows FROM _vonc_pc_backup_20260709;              -- expect 6

CREATE TABLE _vonc_cc_pcard_backup_20260709 AS
SELECT * FROM content_components WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';
SELECT COUNT(*) AS cc_rows FROM _vonc_cc_pcard_backup_20260709;        -- expect 1

CREATE TABLE _vonc_snippet_backup_20260709 AS
SELECT * FROM js_snippets WHERE name = 'provocation-card-loader';
SELECT COUNT(*) AS snip_rows FROM _vonc_snippet_backup_20260709;       -- expect 1
```

```sql
-- D0. THE DECIDING PROBE: does the index instance have content_data?
--     NULL/empty => 003 says the light re-render path escalates to a full rebuild.
SELECT pc.slot_name, pc.schema_mode,
       (pc.content_data IS NULL)                     AS content_data_null,
       COALESCE(jsonb_typeof(pc.content_data),'-')   AS cd_type,
       COALESCE(pc.content_data::text,'')            AS cd,
       LENGTH(pc.rendered_html)                      AS rendered_len
FROM page_components pc
WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
  AND pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';

-- D1. the template (authoritative DOM)
SELECT html_template FROM content_components
WHERE id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';

-- D2. the live index instance
SELECT LENGTH(rendered_html) AS len, rendered_html
FROM page_components
WHERE page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
  AND component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';

-- D3. the loader as stored
SELECT LENGTH(js_content) AS js_len, js_content
FROM js_snippets WHERE name = 'provocation-card-loader';

-- D4. sanity: is provocation-card still vonc-only? (shared-library check before editing a base)
SELECT COUNT(DISTINCT p.site_id) AS sites
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE pc.component_id = '6163ff14-9f94-4962-aa19-d2718eabdeb1';        -- expect 1
```

D4 matters: `forked_from IS NULL` means the row is in the shared library. It was vonc-only on
2026-07-04; confirm before editing, because a direct SQL UPDATE bypasses `store_generated_component`'s
guards and `component_versions` snapshotting entirely.

---

## §6 Commands & triggers

Database:
`kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db`

Kubernetes: main namespace `-n ai-persona-system`; Kafka `-n kafka`, cluster
`personae-kafka-cluster` (bootstrap `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`).

Triggers (in `scripts/initial_messages/210_vonc_trigger/`):

| script | what it does |
|---|---|
| `083_trigger-asset-renderer-vonc.sh` | rebuild `/assets/js/snippets.js` from active `js_snippets` and commit |
| `rerender-index-vonc.sh` | assemble-only rerender + deploy of `/index.html` |
| `085_rerender-provocations-vonc.sh` | same for `/provocations/index.html` |
| `084_create_provocations-archive-list_vonc.sh` | component-creator trigger (the working dual-placement pattern) |

Watch a triggered orchestration (the scripts print the id):

```sql
SELECT status, current_step FROM orchestration_states
WHERE correlation_id = '<CORRELATION_ID>'::uuid;
```

Watch a work item:

```sql
SELECT item_key, status, attempt_count, claimed_by, completed_at,
       substring(COALESCE(error,''),1,200) AS err
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
ORDER BY created_at DESC LIMIT 5;
```

---

## §7 Key ids and schema notes

| thing | id |
|---|---|
| site vonc.com | `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74` |
| current site plan | `77493277-f510-47ea-aa27-8fca415743d6` |
| page: index | `b4d24f8e-fccd-49df-9dad-aa56a0b20a68` |
| page: provocations-index | `e4b3b195-919f-45ad-854e-201d3e846ea8` |
| component: provocation-card | `6163ff14-9f94-4962-aa19-d2718eabdeb1` |
| component: lobby-grid | `9304f14d-e19b-4ce1-b3fd-f6a315aec6ed` |
| component: brief-explanation (SHARED ×3 sites) | `58363894-9db9-4d2f-81ac-c47b54d97fc3` |
| component: provocations-archive-list | `70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37` |
| plan section row (provocations-index) | `e006a35e-9258-49f0-a99a-f11eba409bc9` |
| site: idea.uk | `1244516d-014d-421c-88c6-090bb1e9552a` |
| site: robot-hands.com | `00ff3af5-dad8-4770-9f70-3edc267a3c92` |

Schema notes learned the hard way:

- `pages` has **`name`**, not `page_name`; also `sections jsonb DEFAULT '[]'`, `build_status`,
  `deployed_at`, `last_built_at`. **`last_built_at` is never written** by the build *or* the rerender
  path — a dead column (backlog item).
- `js_snippets` has **no `site_id`** — snippets are global; each loader self-checks for its section
  (`if (!document.querySelector('[data-component="…"]')) return;`) so it is inert on other sites.
- `page_components`: `rendered_html`, `content_data`, `build_status`, `component_id`, `page_id`;
  trigger `trigger_auto_lock_on_deploy` fires on UPDATE.
- `content_components`: keyed by `function`; `forked_from IS NULL` ⇒ shared library row.
- `site_plan_sections` has `UNIQUE (plan_id, page_name, ordering)`; existing vonc rows carry NULL
  style ids.

---

## §8 Standing rules & gotchas (all paid for in this thread)

**Debugging discipline**
- A `complete` work item proves nothing. Verify by artifact: DB row, `curl`, browser. Ten items said
  complete while building nothing.
- Never accept `0 rows` as an answer until the query is cleared of blame (wrong column, wrong id,
  wrong spelling). Two of this thread's dead ends were query faults.
- Pod logs are ephemeral across rollouts; verify deploys by artifact. `logger.Debug` does not appear
  in logs. Grep zap logs by message + JSON field, never `field=value`.
- `agent_error_log` filtered by `orchestration_id` outlives the pod and names the failing step.
- Copy full UUIDs; never hand-type them.

**Editing stored HTML**
- A marker/attribute `REPLACE` must anchor on the **opening tag**, not a bare attribute: the string
  `data-component="X"` also appears inside the section's own inline `querySelector`, and a plain
  replace corrupts it into an invalid two-attribute selector (`SyntaxError`; happened twice —
  `fix_marker_selector.sql`). Better still: emit markers at generation.
- The `hidden` attribute maps to a **UA-stylesheet** `display: none` and loses to **any author**
  `display` rule on the same element. A clone template inside a `display: grid` item class renders as
  a ghost row. Fix with `[data-…-template] { display: none; }` (`fix_archive_template_display.sql`).
- An index/page rerender is **assemble-only**: it redeploys stored `rendered_html`. Template-only
  edits will not appear. Edit the instance too, or run a content rebuild.
- Multi-line block removal: dump → edit offline → UPDATE full text → verify by length delta.

**Components & agents**
- Reuse before create; alter before replace. Read the real code before asserting behaviour.
- Regenerating a **shared** base is only safe for neutral, purely-additive changes; site-specific
  voice forks. Direct SQL UPDATEs bypass the field guard and `component_versions` snapshots.
- The component-creator trigger pattern that works: **dual placement** of `section_type` (top-level
  *and* inside `spec`) and a **quote-free description** (name attribute values in prose) to survive
  the kcat/JSON pipeline. See `084_create_provocations-archive-list_vonc.sh`.
- Bake guards into generation: emit `data-runtime-fill`; forbid inline `<script>` elements in dynamic
  components; make header copy llm-sourced so nothing can defer.
- Stored `js_len` runs ~6 bytes under the source file (paste drift). The bundle and the browser are
  the ground truth for completeness.

**Working practice**
- `/home/claude` **resets between sessions**; `/mnt/user-data/outputs` persists. Re-seed working
  copies **from outputs** before any `>>` append or `cp` toward outputs — a `>>` on a missing file
  creates a fragment that then clobbers the real doc (happened once).
- Always quote heredoc delimiters (`<< 'EOF'`) when the body contains backticks, `$`, or markdown —
  an unquoted one executed the document's code spans as shell commands.
- Never leave a bare angle-bracket tag in markdown prose; backtick-wrap it. An unwrapped `<script>`
  token broke the user's markdown reader mid-document.

---

## §9 Backlog / open threads

1. **Framework preventions** (from the ten-no-op investigation — the highest-value items):
   - Planner invariant at plan-store: every planned page whose **role** is built by
     page-build-handler must have ≥1 section. Make the role→pipeline mapping explicit
     (`blog-post` is legitimately section-less; `section-index` was not).
   - `complete_error` on the zero-sections path must **fail loudly** (own message, or raise
     `needs_plan_sections`) — never report success.
   - Auditor rules: page `active` + linked (nav / `link_registry` / CTA specs) + `build_status` in
     (`planned`,`needs_rebuild`) beyond a threshold → work item; post-deploy URL-presence check per
     active page; drift check `pages.sections` vs the current plan.
   - `last_built_at` is never written — decide: write it or drop it.
2. **Store-validation hardening:** reject an unclosed `<script>` at store time (a truncated inline
   script once shipped and broke a page).
3. **`PATCH_validate_input_contract.go`** — drafted, not deployed.
4. **Resolver illustration support:** assets `illustration_game_master` / `illustration_gauntlet_cta`
   exist but `ensureAssets` surfaces only hero/logo. brief-explanation's illustration and its two
   `href="#"` CTA placeholders remain unwired.
5. **CTA-map pass (parked, user chose Option B).** There is no arena page; the interactive surface is
   the Gauntlet tool at `/tools/gauntlet/index.html`. The CTA graph is circular: hero → archive;
   archive → home; gauntlet-cta → archive; only nav/footer reach the tool. Retarget SQL is in the
   notes for when the real take-filing arena exists. These URLs are baked into rendered sections, so
   a proper refresh is a section rebuild, not string surgery.
6. **Autonomy design** (`PLAN_dynamic_sections_and_loaders.md`): section descriptor
   `{role, kind, data_feed}` on the plan; component-creator "Tier E" runtime-feed shells;
   a **loader-builder** agent (sibling of `tool-generator`, which forbids `fetch`); Phase-3 pipeline
   emitting `provocations.json {today, arena, archive}`. The two hand-built loaders are its reference
   implementations.
7. **029 Q1 evidence hand-off** to the planner/reconciler thread; confirm with
   `SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'` and a repo grep for
   `reconcile_site_plan`.
8. **Cross-site tidy (other threads own these):** idea.uk and robot-hands.com each have one
   `brief-explanation` instance still `pending` on stale rendered HTML, awaiting their own content
   rebuild. No contamination — they inherited a neutral base.
9. **Other-page audit:** the `provocation` blog page is absent from the plan's page list yet was
   built (evidence the `pages.sections` fallback works); check the remaining pages for the same
   silent-no-op signature.
10. **Apply `016b_debugging_guide_merged.md`** to the project — the project copy forked across chats;
    this one is cumulative (includes the silent-no-op entry, its 2026-07-06 amendment, marker
    anchoring, and the `hidden`-vs-author-CSS entry).

---

## §10 File inventory (`/mnt/user-data/outputs/`)

**Read first:** `RUNBOOK_phase2_provocation_js.md` (operative; top block = current position) ·
`RUNNING_NOTES_vonc_v2.md` (chronological log) · `016b_debugging_guide_merged.md` (**apply this**) ·
`PLAN_dynamic_sections_and_loaders.md` (autonomy design + the plan-storage decision and its
supersession).

**Per-tool docs:** `PLAN_provocation-card.md` / `NOTES_provocation-card.md` (← the next task) ·
`PLAN_lobby-grid.md` / `NOTES_lobby-grid.md` · `NOTES_provocations-index.md` ·
`SPEC_provocations-archive-list.md` · `PLAN_brief-explanation.md` / `NOTES_brief-explanation.md` ·
`TOOL_DOCS_convention.md`.

**Loaders (source of record; they ship via `js_snippets`, not as files):**
`provocation_card_loader.js` · `lobby_grid_loader.js` · `provocations_archive_loader.js`.

**Data:** `provocations.sample.json` (v3 — `today` / `lobby` / `arena` / `archive`; the `lobby` key
becomes dead after the trim).

**SQL:** `fix_archive_template_display.sql` · `fix_marker_selector.sql` · `lobby_grid_install.sql` ·
`provocations_archive_install.sql` · `phase2_step3_insert_snippet.sql`.

**Bundle:** `bundle_minilobby_trim.sh` — primed `cmd/bundle` request that settles the trim's method
(read-only; run it before editing anything). Contains Step-0 resolvers for the four paths that cannot
be asserted from outside the repo, and a trailing block of row-state SQL the bundle does not carry.

**Triggers:** `084_create_provocations-archive-list_vonc.sh` · `trigger-asset-renderer-vonc.sh` ·
`rerender-index-vonc.sh` · `make_085_rerender_provocations.sh` (generates the provocations rerender
by sed-copying the index one).

**Go (reference / pending):** `load_page_sections_from_spec_action.ORIGINAL.go` (the revert target) ·
`load_page_sections_from_spec_action.WITHDRAWN.go` (provenance stub — do not deploy) ·
`rerender_single_page_action.go` (patched + deployed: the `data-runtime-fill` exemption) ·
`PATCH_validate_input_contract.go` (pending) · `PATCH_section_visible_content.go` ·
`store_generated_component_action.go` (reference — read its shared-component guard).

`Categories:` (docs, handoff)
