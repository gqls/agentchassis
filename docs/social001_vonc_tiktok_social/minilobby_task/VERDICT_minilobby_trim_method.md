# VERDICT — the supported path for the provocation-card mini-lobby trim

**Created:** 2026-07-09 · **Answers:** HANDOFF §4.0 ("settle it first") · **Status:** method settled, nothing written yet
**Evidence:** `/tmp/bundle_minilobby_trim.md` (code + docs + schema), the scoped Go sources read directly,
and read-only probes against `clients_db`.

---

## §0 Verdict in one paragraph

The trim is a **template edit propagated by the section-editor**, not an HTML patch and not a page rerender.
Direct SQL is the only writer for `content_components.html_template` (no action edits arbitrary templates —
`fix_component_template`'s `remove_element` cannot touch this component at all, see §2), so we UPDATE the
template with a hand-written `component_versions` snapshot first, exactly as `repair_template_slots` does.
We then propagate to the live instance with the **section-editor agent** (`edit_type: content_edit`), which
re-renders that one `page_component` from the changed template, rewrites `rendered_html`, reassembles the page
from the other five stored sections, and commits — the route `fix_component_template_action`'s own header
names. The loader is a separate, independent deploy: **direct SQL is the sanctioned writer** for
`js_snippets.js_content` (`render_js_snippets_for_site` is read-only and its generated banner says
"Do not edit directly — update `js_snippets` rows in DB"), followed by the asset-renderer trigger.
Doc 003's rejection of HTML patching does not bite, because we change the source and let the template
re-render — and in this specific case we can prove the two agree byte-for-byte (§3).

---

## §1 The five questions, answered

| # | Question | Answer | Evidence |
|---|---|---|---|
| 1 | Which action changes `content_components.html_template`? | **None, for an arbitrary edit.** Only `store_generated_component` (LLM component-creator) and `fix_component_template`'s `repair_template_slots` (a narrow `<no value>` → `{{.field}}` repair) write that column. Direct SQL + a manual `component_versions` snapshot is the supported route. | `fix_component_template_action.go:481-591`; `registry.go` |
| 2 | Is `remove_element` the reuse candidate? | **No — it cannot reach this component.** It `SELECT`s and `UPDATE`s **`site_components`** keyed by `(site_id, slot_name)` — header/footer/head only. It never touches `content_components` or `page_components`. The handoff's premise was drawn from the file header, not the code. | `fix_component_template_action.go:354-421` |
| 3 | Which action re-renders `page_components.rendered_html` from a **changed template**? | **`apply_section_edit` (section-editor), `edit_type: content_edit`** — single section. `rerender_page_sections` (light path) also re-renders from the template but does **all** sections. `rerender_single_page` only assembles stored HTML and would redeploy the old markup. | `section_editor_actions.go:598-641`; `rerender_page_sections_action.go:227-250`; `rerender_single_page_action.go` header |
| 4 | What happens to a Mode-B section with NULL/empty `content_data`? | Moot here. The light path escalates the **whole page** to `needs_page` if **any** section has `len(content_data) == 0` — which catches `{}` as well as NULL. **All six index sections carry non-empty objects**, so it would not escalate. But see §4: the light path is still the wrong tool. | `rerender_page_sections_action.go:167-179`, `:390-417` |
| 5 | How does a changed `js_snippets.js_content` reach `/assets/js/snippets.js`? | **Direct SQL is the only writer.** No production Go code INSERTs or UPDATEs `js_snippets` (only the action's own test file). `render_js_snippets_for_site` is a pure reader: it selects `is_active = true` snippets whose `applies_to` overlaps the site's component functions, concatenates, and hands `files{assets/js/snippets.js}` to `git_commit`. | `render_js_snippets_for_site_action.go:145-177, 237-259`; repo-wide grep |

**Bonus — do hand-added attributes survive a re-render?** Only if they are in the **template**. They are:
`data-runtime-fill` is present in `provocation-card`'s and `lobby-grid`'s `html_template`, not just their
`rendered_html`. So the marker survives, and the assembler keeps the section — the `≤10 chars of visible text`
filter exempts marked sections, and `assemblePage`/`sectionHasVisibleContent` are **shared** by
`rerender_single_page` and the section-editor, so the exemption holds on both paths.
(`rerender_single_page_action.go:429-452`.)

---

## §2 The finding that changes the shape of the task

**`provocation-card` and `lobby-grid` are not "Mode-B runtime-fill components". They are rendered outputs
stored as source templates.**

| component | `<no value>` | `</no>` | `{{.` slots | `schema_field_count` |
|---|---|---|---|---|
| provocation-card | 26 | 0 | **0** | 0 |
| lobby-grid | 37 | 0 | **0** | 0 |
| provocations-archive-list | 0 | 0 | 8 | 8 |
| hero / gauntlet-cta / brief-explanation / system-stats | 0 | 0 | 6 / 20 / 20 / 22 | 7 / 20 / 20 / 22 |

`RenderTemplate` strips every literal `<no value>` to `""` before returning
(`component_library.go:544-559`). So for these two components:

```
rendered_html  ==  html_template  with all '<no value>' removed
```

which the byte counts confirm exactly:

* provocation-card: `10300 − 26×10 = 10040` = current `rendered_len` ✓
* lobby-grid: `15677 − 37×10 = 15307` = current `rendered_len` ✓

**Consequences.**

1. The section renders empty **by accident**, not by design. Its emptiness — the thing the runtime-fill loader
   depends on — is a side-effect of `<no value>` stripping. It happens to be exactly the behaviour we want.
2. Their `content_data` (25 keys each) is **dead**. There are no `{{.}}` slots to consume it. Nothing on the
   page depends on it, which is why the trim is a pure template-text edit.
3. `repair_template_slots` **cannot** repair them: with zero `</no>` closing tags it returns
   `action: needs_regeneration` (`fix_component_template_action.go:540-553`). Any future component-standards
   audit will flag both. That is a real backlog item, not a blocker for this trim.
4. The render of these two components is a **pure function of the template**, so the post-trim `rendered_html`
   is predictable offline to the byte. That is what makes the section-editor's first production run verifiable.

---

## §3 Why doc 003's objection does not bite

003 rejects HTML patching because *"if we only patched `rendered_html`, the edit would be lost on the next
re-render"*. The rejection is of patching **instead of** changing the source. Here we change the source
(`html_template`) and let the sanctioned action re-render the instance from it. The edit is therefore
**reproduced**, not lost, by any future re-render — full path, light path, or section edit alike.

Because `rendered = template.replace('<no value>', '')` for this component (§2), the value the section-editor
will write is **exactly** the value we could have written by hand. The sanctioned path and the fallback path
converge on the same bytes; we take the sanctioned one, and use the prediction as the acceptance test.

---

## §4 Why NOT the light path (`rerender_page_sections`)

Three reasons, in order of weight:

1. **Blast radius.** It re-renders **every** section on the page. `brief-explanation`'s template was
   regenerated on **2026-07-06**; its vonc instance was last rendered **2026-07-03**. A light re-render would
   silently reshape that section too — an unreviewed content change to the live index, outside this task.
   (It would not break: `illustration_url` is `required:false, on_missing:skip_field` and the template guards it
   with `{{if .illustration_url}}`, so no empty `img`; and `RenderTemplate` cleans any `<no value>` to `""`.
   But it is still an unrequested change.)
2. **It is gated on a reason we do not have.** The `page-rerender` workflow's `check_rerender_mode` step only
   routes to `rerender_sections` when `input_data.spec.reason` is exactly `image_landed` or
   `section_data_resolved`; anything else falls through to the assemble-only `render_page`. Using it would mean
   passing a false reason, or widening the conditional.
3. **`page_rerender` is not the item type for a template change.** Nothing raises it for this cause.

Widening `check_rerender_mode` with a `template_changed` reason (a DB workflow row edit — no chassis rebuild)
is a reasonable **structural** follow-up, but it still re-renders all sections, so it does not replace the
section-editor for a single-section change.

`Categories:` (decision, root-cause)

---

## §5 The sequence

Two **independent** deploys — no interlock between them. If the loader still holds the `data.lobby` block while
the template has lost `.pc-card`, `section.querySelectorAll(".pc-card")` returns an empty NodeList and the loop
is a no-op; and vice versa. Either order is safe.

### A. Template + instance (one deploy)

1. `psql -f trim_minilobby.sql` — dated backups, `component_versions` snapshot, template UPDATE, loader UPDATE.
   Verify the two in-transaction `SELECT`s read as their column names say, **then** `COMMIT`.
2. `bash 086_section_edit_provocation-card_vonc.sh` — section-editor `content_edit` on
   `page_component a757434e-…`, re-render from the trimmed template, reassemble, commit `index.html`.
3. Verify by artifact:
   * `rendered_len` = **6488** (was 10040), `pc-card` absent, `<script` absent, `data-runtime-fill` present.
   * `curl -s https://vonc.com/index.html | grep -c 'pc-card'` → `0`
   * `curl … | grep -o 'data-component="[^"]*"'` → the same six.

### B. Loader (separate deploy)

4. (Already applied by `trim_minilobby.sql`.) `bash scripts/initial_messages/210_vonc_trigger/083_trigger-asset-renderer-vonc.sh`
5. `curl -s https://vonc.com/assets/js/snippets.js | head -3` → still `3 active snippet(s)`;
   `grep -c 'data.lobby'` → `0`.

### Numbers to expect

| artefact | before | after | delta |
|---|---|---|---|
| `content_components.html_template` | 10300 | **6618** | −3682 |
| `page_components.rendered_html` | 10040 | **6488** | −3552 |
| `js_snippets.js_content` (chars) | 4879 | **3365** | −1514 |
| `<no value>` in template | 26 | 13 | −13 |

The handoff's estimate of −1,200/−1,600 bytes was low: it counted only the grid markup and the CSS rule.
The trim also removes ~62 lines of `.pc-card*` CSS and the ~31-line dead inline `<script>`.

---

## §6 Risks, named

1. **The section-editor has never run in production.** `agent_definitions` has it `is_active = true` with a
   complete workflow (`ensure_site_record → spawn_deployer → load_edit_context → apply_edit → deploy_page →
   update_page_status → trigger_deploy → complete`), both its actions are registered in `registry.go`, and its
   reassembly reuses the proven `assemblePage`. But `orchestration_states` has **zero** rows for it. This is
   its first exercise. Recovery: restore `rendered_html` from `_vonc_pc_backup_20260709` and run
   `083_rerender-index-vonc.sh` (assemble-only) to redeploy. ~1 minute.
2. **`field_updates` must be non-empty.** `applyContentEdit` errors when neither `field_updates` nor
   `replacement_content_data` is supplied, and an empty `{}` may not survive `ExtractActionInputs`. The trigger
   therefore re-supplies `_built_at` with its exact current value — a genuine no-op merge that leaves
   `content_data` bit-for-bit unchanged. (`replacement_content_data` is deliberately avoided: the action's own
   comment warns that a nested lookup can bind `site_record.content_data` to it by mistake.)
3. **`provocation-card` is `forked_from IS NULL`** — a shared-library row by 003's definition, even though it is
   currently vonc-only (one instance, one row with that `function`). A direct SQL UPDATE bypasses
   `store_generated_component`'s field-set guard and its `component_versions` snapshotting; step 1 snapshots by
   hand to compensate. The change is a pure removal of Spark-specific markup from a Spark-specific component,
   so there is no cross-site exposure today — but the row should be forked if any other site ever adopts it.
4. **`provocations.json`'s `lobby` key becomes dead data** after this change. `today` / `arena` / `archive`
   remain live. No code reads `lobby` once the loader block is gone.
5. **Both templates remain rendered-artifact templates** (§2). Nothing here regresses that; the trim leaves 13
   `<no value>` in place, because they are what makes the shell empty for the loader to fill. Regenerating
   `provocation-card` or `lobby-grid` properly — as `provocations-archive-list` was — must consciously
   re-establish the empty-shell + `data-runtime-fill` contract, or the sections will ship with build-time copy.

`Categories:` (decision, next-task, root-cause)

---

## §7 OUTCOME — executed and verified 2026-07-09, CLOSED

Ran exactly the sequence in §5. Every prediction held.

**Writes.** `trim_minilobby.sql` committed (self-verifying `DO` block; `ON_ERROR_STOP=1`). Backups
`_vonc_pc_backup_20260709` (6 rows), `_vonc_cc_pcard_backup_20260709` (1), `_vonc_snippet_backup_20260709` (1);
`component_versions` snapshot written with `change_source = 'minilobby_trim_20260709'`.
Then `086_section_edit_provocation-card_vonc.sh` — the section-editor's **first ever production run**
(3 orchestrations, all COMPLETED). Then `083_trigger-asset-renderer-vonc.sh`.

**Verified by artifact, never by item status.**

| check | expected | actual |
|---|---|---|
| `html_template` md5 == offline TRIMMED file | equal | ✅ `adcbbcfd…` both sides |
| `js_content` md5 == offline TRIMMED file | equal | ✅ `9d0b5069…` both sides |
| `page_components.rendered_html` length | 6488 | ✅ 6488 |
| `md5(rendered_html) == md5(replace(html_template,'<no value>',''))` | true | ✅ **true** — §2's model exact |
| other five sections byte-identical to backup | all `t` | ✅ hero, gauntlet-cta, brief-explanation, lobby-grid, system-stats |
| `content_data` keys / `_built_at` | 25 / unchanged | ✅ 25 / `2026-07-03T12:56:49Z` (no-op merge worked) |
| live `index.html` `pc-card` | 0 | ✅ 0 |
| live `index.html` distinct `data-component` | 6 | ✅ 6 |
| live `role="listitem"` (4 pc-cards + 6 arena → 6) | 6 | ✅ 6 — lobby-grid's `article.lobby-grid-section__card` intact |
| live `snippets.js` header | `3 active snippet(s)` | ✅ 3, all three loaders present |
| live `snippets.js` `data.lobby` | 0 | ✅ 0; `pc-headline`/`pc-stat-value` fills retained |
| loader brace/paren balance | 0 | ✅ excised block self-contained; no JS runtime available, verified by lexer |

The three residual `1fr 1fr` on the live page belong to **gauntlet-cta** and **brief-explanation**
(their own legitimate two-column layouts), confirmed against the DB per section. provocation-card has none.
`data-component="lobby-grid"` still appears twice in the page because lobby-grid's own inline script contains a
`querySelector('[data-component="lobby-grid"]')` — the pre-existing double-occurrence that this trim has now
retired from provocation-card.

**Acceptance — ALL FIVE CONFIRMED, browser-verified 2026-07-09 19:10.** Six sections deploy ✅ ·
provocation-card fills ✅ (eyebrow "TODAY'S PROVOCATION", headline "AI will *never* be funny on purpose." with
`<em>` accent, body, both CTAs "File Your Position" / "See All Provocations", stats 1,284 · 3h 12m · 62%) ·
no `.pc-card` markup anywhere ✅ · lobby-grid's six arena cards unaffected ✅ · console clean apart from a
pre-existing `favicon.ico` 404, unrelated to this change ✅.

`Categories:` (milestone)

---

## §7a Cosmetic consequence of the trim — APPLIED and verified 2026-07-09

`.pc-body { max-width: 52ch }` (and the headline's `clamp(2rem, 5vw, 3.5rem)`) were tuned for the **left half**
of the old `1fr 1fr` grid. With the grid gone the container is a single `1fr` column at
`max-width: var(--container-max-width, 1200px)`, so the copy hugs the left and the right half of the section is
empty, while the `.pc-stat-strip`'s top border still spans the full 1200px. The section also keeps
`min-height: 60vh`, which reads sparse now that only one column fills it.

This was the visual half of what §4 of the handoff predicted ("the section will render one column of content
beside an empty one"). Fixed with one declaration: `.provocation-card-section .pc-container` now carries
`max-width: 820px` instead of `var(--container-max-width, 1200px)`, making it a centred single column.

Applied by `centre_provocation_card.sql` (snapshot `change_source = 'pc_container_centre_20260709'`) plus a
second run of `086_section_edit_provocation-card_vonc.sh`. Template 6618 → **6589**; `rendered_html`
6488 → **6459**, and `md5(rendered_html) == md5(replace(html_template,'<no value>',''))` held again — §2's
model has now predicted the exact output twice. Live within 12 s: `max-width: 820px` served once,
`pc-card` 0, six components, six `role="listitem"`, both `data-runtime-fill` markers.

**The `approved` defect (§8.1) recurred exactly as predicted** on this second section-editor run, and was
repaired with the same two-statement transaction. It is deterministic: **every** `apply_section_edit` needs the
`approved → deployed` repair until the action is fixed. All six sections are `deployed`, `schema_mode` NULL.

---

## §8 Defects found on the way — for the backlog

### §8.0 Fixes applied 2026-07-09 (defects 1 and 2) — **VERIFIED END-TO-END 2026-07-10**

Chassis `v1.0.1102` (commit `8f9fe537`) deployed 2026-07-10. Verification: a no-op section-edit
(correlation `78d283bf-82d1-4705-ac3a-599a6786a0b4`) ran through the section-editor on the patched binary —
the row landed at **`deployed` automatically**, `rendered_html` unchanged at 6459, no lock, no manual repair.
The spawned pod's zap log shows the full sequence 5 ms apart:
`UpdatePageStatusAction: Updated page` (09:42:22.491) → `UpdatePageStatusAction: Marked page_component
deployed` for `a757434e-…` (.496). Table-wide: **0 `approved` rows, 0 locked rows** (592 deployed / 22
pending). The manual `approved → deployed` repair is retired. (Log-hunting note: `apply_section_edit` /
`update_page_status` log in the **spawned `agent-section-editor-*` pod**, not the main `agent-chassis` pod.)

- **Go** (`platform/orchestration/actions/v3_site_actions.go`) — `UpdatePageStatusAction` gained an optional
  `page_component_id_field` config key: when `status=deployed` and the caller names a page_component, the action
  also sets that `page_components.build_status = 'deployed'`. Non-fatal on error (page row already committed).
  Reuses `execDB` / `ExtractNestedFieldString`; adds `page_component_updated` to the result. Builds + vets clean.
- **Go** (`platform/orchestration/coordinator.go`) — added `page_component_id_field` to `dataRefKeys` in
  `prefixConfigStepReferences`, so the new key gets loop-iteration prefixing like `commit_from`. (Not a
  whitelist — unknown keys already pass through; this is correctness for the loop case.)
- **DB** (agent_definitions) — section-editor's `update_page_status` step config gained
  `"page_component_id_field": "edit_result.page_component_id"`. Backup:
  `_section_editor_agentdef_backup_20260709`. Inert on the current binary (unknown keys ignored); activates when
  the patched chassis ships. `page_id_field` preserved.
- **DB** (`docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql`) — dropped the
  `auto_lock_on_deploy` trigger + function and normalised the one legacy `strict` row (gaswholesalers.com).
  Strict mode was stillborn: `schema_snapshot`/`content_snapshot` columns never created (so the sibling
  `lock_section_to_strict` / `unlock_section_for_redesign` functions would error if called), no Go reader of
  `schema_mode` / `strict_mode_trigger`, one firing ever. Reversible from
  `auto_lock_on_deploy.FUNCTION_BACKUP.sql`. This had to happen *before* the fix goes live, else every edited
  section would lock to `strict`. **Note:** with the trigger gone, the manual `approved → deployed` repair used
  twice today becomes a plain one-statement UPDATE (no lock to clear) until the chassis ships and the fix makes
  it automatic.

`Categories:` (fix, decision)

---

1. **`apply_section_edit` marks a row `approved` after successfully deploying it.**
   `updatePageComponentAfterEdit` (`section_editor_actions.go:946`) hardcodes `build_status = 'approved'`, and
   the workflow's `update_page_status` step only updates **`pages`**, never `page_components`. After our run,
   provocation-card was the **only `approved` row in 578**. Every discovery check filters
   `pc.build_status = 'deployed'` — `check_empty_sections`, `check_image_url_404`, `check_undeployed_assets`,
   `check_placeholder_image_in_use`, `check_component_standards` — so the section had gone **silently invisible
   to the entire audit surface**. This is the same shape as the `complete_error` ten-no-op bug: a successful
   path leaving a state nobody looks at. Restored by hand (see 2). **Fix:** `apply_section_edit` should set
   `deployed` after `deploy_page` succeeds, or `update_page_status` should carry the `page_component_id`.

2. **`auto_lock_on_deploy` is effectively dead — and nearly bit us.** vonc has
   `sites.strict_mode_trigger = 'first_deploy'`, yet all six index sections have `schema_mode IS NULL`: the
   trigger is `BEFORE UPDATE` only, and `save_page_sections` **INSERTs** rows already at `deployed`, so it never
   fires. Restoring `approved → deployed` by a plain UPDATE *would* have fired it and locked that one row to
   `schema_mode = 'strict'`, an anomaly among its five siblings. Repaired with a two-statement transaction:
   UPDATE to `deployed` (trigger locks), then clear `schema_mode`/`locked_at`/`locked_by` (guard
   `OLD.build_status <> 'deployed'` is now false, so it no-ops). **Decide:** make the trigger `BEFORE INSERT OR
   UPDATE`, or drop it.

3. **`js_snippets` has no `updated_at` column.** The handoff's §7 schema notes imply one; `SET … updated_at =
   now()` on that table is a hard error. Columns are `id, name, description, js_content, semantic_tags,
   applies_to, dependencies, created_at, is_active`.

4. **`fix_component_template`'s header oversells `remove_element`.** The header says it "does not change
   `page_components` `rendered_html` content because content changes go through the section-editor workflow",
   which reads as though it operates on page components at all. It does not: `remove_element`,
   `inject_nav_flex_css` and `responsive_fix` all read and write **`site_components`** only. Worth a header fix
   so the next reader does not plan around a capability that isn't there.

5. **`provocations.json` still ships a dead `lobby` key** (`generated_at, today, lobby, arena, archive`).
   Nothing reads it now. Drop it when the Phase-3 pipeline emits the file.

5a. **`brief-explanation` ships `<img src="">` on the live index** — the empty bordered box rendering its `alt`
   text ("Abstract data visualisation showing opinion splits…"). It is the **only `<img>` on the whole page**.
   Root cause is upstream of this task and **must not be patched here**: `assets` rows
   `illustration_game_master` / `illustration_gauntlet_cta` exist and are `active`, but their `filename` and
   `storage_path` are **NULL** and nothing is served under `/assets/images/` (404). Only a raw Backblaze S3
   `url` exists — and resolving to the presigned S3 url instead of the deployed git path is the exact bug fixed
   in `84f07d38`. So the asset has never been deployed into the `sites` repo.
   Re-rendering the section is **not** a fix either: the `{{if .illustration_url}}` guard wraps the `img`
   **and** the `brief-explanation__badge` ("Gauntlet Closes Soon"), and
   `.brief-explanation__container` is unconditionally `grid-template-columns: 1fr 1fr` at desktop — so dropping
   the image column trades a broken image for an empty right half plus a lost badge. This belongs to the
   imagery workstream (deploy illustrations to `/assets/images/`, then one `content_edit` setting
   `illustration_url`). Supersedes the vaguer backlog item 4 in the handoff.

6. **`provocation-card` and `lobby-grid` templates are rendered artifacts** (§2) and are **unrepairable** by
   `repair_template_slots` (zero `</no>` tags → `action: needs_regeneration`). Any component-standards audit
   will flag both. Their `content_data` is dead weight. Regenerating them must consciously re-establish the
   empty-shell + `data-runtime-fill` contract.

`Categories:` (backlog, root-cause)
