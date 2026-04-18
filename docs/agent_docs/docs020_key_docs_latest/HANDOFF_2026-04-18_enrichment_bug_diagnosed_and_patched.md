# HANDOFF — 2026-04-18: enrichment bug diagnosed and patched

Continues from `HANDOFF_2026-04-17_quality_scoring_js_separation_deployed.md`. That handoff left Track 1 (NULL `component_id` on rebuilt pages) as the headline unresolved issue. This session identified the root cause, validated it from live logs, and produced a patched `save_page_sections_action.go` ready to deploy.

## Status summary

| Track | Status |
|---|---|
| Component quality scoring | Deployed and working (unchanged) |
| JS separation | Working (unchanged) |
| **Track 1 — NULL component_id on rebuilt pages** | **Root cause identified. Fix written and syntax-checked. Order swap already in uploaded file. Belt-and-braces patch in outputs.** |
| Track 2 — component templates missing `{{.var}}` placeholders | Still open. Needs one more component-creator prompt revision |
| Track 3 — intermittent Kafka broker timeouts | Still open, low priority. Doesn't block pipeline completion |
| Track 4 — `pages` / `site_plan` drift on `membership`, `how-it-works`, `contact` | Still open, HITL decision |

## The root cause of Track 1

### What we saw

Every gauntlet diagnostic rebuild since yesterday showed the same pattern on page_components:
- `component_id = NULL`
- `content_data = f`
- `rendered_html` present (25KB, 15KB)

Yesterday we thought `save_page_sections` was early-skipping with `"reason": "no HTML content and no sections metadata"`. Today we confirmed that was only happening for some runs. The most recent run (13:07) actually ran the full path and returned `sections_saved: 2, skipped: null`. But `component_id` was still NULL.

### Why grep kept missing the enrichment logs

The log lines ARE being written by zap, but grep was scanning truncated views of them. The `msg` field often appears after the `action`, `caller`, `agent_id` fields and may be beyond the terminal-wrap point.

**Reliable grep pattern (works regardless of wrap):**
```bash
grep -E 'save_page_sections_action|enrichSectionsWith|validate_page_content\.go|plan_sections_action'
```
This matches on the `caller` field (source file path) which is short and always present. Matching on `msg` text or on `"action":"save_page_sections"` works but can miss lines that are truncated before `msg`.

### The actual bug

From the 13:07 capture, in order:
```
SavePageSectionsAction: Using structured metadata path  sections=2
enrichSectionsWithComponentIDs: invoked  section_count=2  db_nil=false
enrichSectionsWithPlannedNames: using planned section name  old_name="section" planned_name="gauntlet-interface"
enrichSectionsWithPlannedNames: using planned section name  old_name="section" planned_name="archetype-result-card"
enrichSectionsWithPlannedNames: enriched sections  enriched=2
SavePageSectionsAction: Complete  sections_saved=2
```

Note:
1. `enrichSectionsWithComponentIDs: invoked` fired
2. No `linked component` and no `no match found` logs followed
3. `enrichSectionsWithPlannedNames` fixed the names AFTER component IDs had been looked up

What happened inside `enrichSectionsWithComponentIDs`:
```go
for i := range sections {
    if sections[i].ComponentID != "" { continue }
    if sections[i].ComponentName == "" || sections[i].ComponentName == "section" {
        continue  // ← EVERY section hit this guard and skipped silently
    }
    // ... lookup never executed ...
}
```

The sections arrived with `ComponentName = "section"` because `extractSectionsFromMetadata` defaults to that string when `sections_metadata` from content-writer lacks a `component_name`/`function` field. Looking at the captured metadata:

```json
"sections_metadata": [
  {"rendered_html": "<style>..."},
  {"rendered_html": "<style>..."}
]
```

**Only `rendered_html`. No name field.** So `extractSectionsFromMetadata` falls back to the default `"section"`, the guard skips every section, component_ids never get looked up, INSERT runs with `componentIDPtr == nil`, and page_components.component_id is NULL.

### Why this was invisible in the old binary

`v1.0.972` added the `enrichSectionsWithComponentIDs: invoked` log, but the `linked component` / `no match found` logs are inside the function body after the guard. Because the guard skipped every section, neither log fired, so the function looked like it had produced no output. Combined with hard-to-grep truncated lines, the enrichment appeared to not run at all.

### Why the HTML-fallback path (which works correctly) made us think the bug was elsewhere

When `sections_metadata` is absent, save_page_sections falls back to parsing `data-component="..."` out of the assembled HTML. That path populates `ComponentName` with the real component name (`gauntlet-interface`), the guard doesn't trigger, enrichment succeeds. Yesterday's runs that saved successfully were probably hitting the HTML fallback. Today's runs all went through the metadata path — thus hit the bug.

## The fixes

Two changes. Both in `save_page_sections_action.go`.

### Fix 1 — enrichment call order (already applied to uploaded file)

Lines 207-208, in `SavePageSectionsAction`:
```go
// BEFORE:
enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger)
enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)

// AFTER (already in uploaded file):
enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)
enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger)
```

`enrichSectionsWithPlannedNames` reads planned names from `pages.sections` and fills them in over generic/empty values. Running it FIRST means that by the time component_ids lookup runs, each section has a real name to look up.

Swap alone fixes today's bug. But it's fragile: a page without a populated `pages.sections` row, or a position-count mismatch, would still fail.

### Fix 2 — belt-and-braces inside `enrichSectionsWithComponentIDs` (in outputs as `save_page_sections_action.go` and `enrich_fix.diff`)

Three changes, all inside the function:

**(a) HTML-attr extraction promoted BEFORE the "section" guard.** The function already contained code to extract `data-component="..."` from the rendered HTML — but that code ran AFTER the guard had already skipped. Now extraction happens first; if `ComponentName` is missing or generic, the HTML value is adopted as the name. Only skips if neither source has a usable name.

**(b) Non-ErrNoRows query errors now log as Warn.** Previously a context cancellation or a connection drop looked identical to "no rows found" because the error was discarded. All three `QueryRowContext` sites now log non-ErrNoRows errors explicitly, making connection/timing issues diagnosable.

**(c) `no match found` log now includes `candidates_tried`.** So future name-matching failures show exactly which names were attempted in one log line, without needing a source read.

Both files are in `/mnt/user-data/outputs/`:
- `save_page_sections_action.go` — full patched file (gofmt-clean)
- `enrich_fix.diff` — 94-line unified diff against the file user uploaded

## Recommended deploy sequence

**Option A:** Deploy the order swap alone first (file as-uploaded by user).

Queue one gauntlet diagnostic, check `page_components.component_id`:
- Populated → swap alone fixes it. Deploy belt-and-braces as separate commit later.
- Still NULL → swap is insufficient, proceed to Option B.

**Option B:** Deploy order swap + belt-and-braces together (`/mnt/user-data/outputs/save_page_sections_action.go`).

Slightly more code in one deploy but more robust to:
- Pages with no `pages.sections` row
- `site_plan.pages` positions not aligning with saved positions
- Future changes to what content-writer emits in `sections_metadata`

### Expected log signature after either deploy

```
SavePageSectionsAction: Using structured metadata path
enrichSectionsWithPlannedNames: using planned section name  old_name=section planned_name=gauntlet-interface
enrichSectionsWithPlannedNames: enriched sections  enriched=2
enrichSectionsWithComponentIDs: invoked  section_count=2
enrichSectionsWithComponentIDs: linked component  slot_name=gauntlet-interface  matched_by=exact:gauntlet-interface
enrichSectionsWithComponentIDs: linked component  slot_name=archetype-result-card  matched_by=exact:archetype-result-card
SavePageSectionsAction: Complete  sections_saved=2
```

### Validation query

```sql
WITH vonc AS (SELECT id FROM sites WHERE domain = 'vonc.com'),
     g AS (SELECT id FROM pages WHERE site_id = (SELECT id FROM vonc) AND name = 'gauntlet')
SELECT pc.position, pc.slot_name,
       CASE WHEN pc.component_id IS NULL THEN 'null'
            WHEN cc.id IS NULL THEN 'dangling'
            WHEN cc.is_active = false THEN 'inactive'
            ELSE 'live: ' || cc.function END AS link_state,
       pc.updated_at
FROM page_components pc
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = (SELECT id FROM g)
ORDER BY pc.position;
```

Both rows should show `link_state = 'live: gauntlet-interface'` and `'live: archetype-result-card'`.

## The structurally correct long-term fix (not in this session's patch)

The belt-and-braces works around a deeper issue: **`compile_page_sections` in page-content-writer emits `sections_metadata` entries with only `rendered_html`**. It should also emit `component_name` (or `function`). If it did, `extractSectionsFromMetadata` would have a real name to use and none of the enrichment gymnastics would be needed.

This is in a different agent (`page-content-writer`) and touches its workflow. Deferred to a separate session.

## Architecture fact captured this session

### Spawned pod labels (for future log-streaming work)
- Persistent deployments: `app=agent-chassis`
- Spawned dynamic pods: `app=dynamic-agent`, `agent-type=<type>`
- Combined selector: `-l 'app in (agent-chassis,dynamic-agent)'`
- Pod names `agent-<type>-<8char>-<5char>` (e.g. `agent-page-build-handler-1c5d960b-s6bnn`)
- Idle timeout 600s — pods are often Completed/gone when we look for them

### save_page_sections config shape (unchanged but captured)
```yaml
save_sections:
  action: save_page_sections
  config:
    html_field: validation_result.clean_html
    sections_metadata_field: page_content.response.sections_metadata
    site_id_field: site_record.site_id
    page_name_field: input_data.spec.page_name
```

Early-skip when both `html_field` and `sections_metadata_field` resolve to empty → returns `skipped: true, reason: "no HTML content and no sections metadata"`.

### Reliable grep for chassis logs
```bash
# Match on the source-file path in the `caller` field
grep -E 'save_page_sections_action|enrichSectionsWith|validate_page_content\.go|plan_sections_action'

# Or match on short action-name fields
grep -E '"action":"save_page_sections"|"action":"validate_page_content"|"action":"plan_sections"'
```

The `caller` field is short and never gets truncated below the visible terminal width. The `msg` field often does.

## False starts noted (avoid repeating)

- **"Stale binary"** — binary is fine, v1.0.972 is running
- **"Wrong pod selector"** — spawned pods are `app=dynamic-agent`, not `app=agent-chassis`
- **"Content-writer is returning empty"** — it wasn't. Content-writer returns rich HTML; the issue was only with the metadata shape it emits alongside
- **"Enrichment is skipping because DB unreachable"** — DB is fine, the guard just skips based on ComponentName
- **"SavePageSectionsAction isn't running"** — it is, grep just couldn't find the logs due to truncation

## Files generated or modified

| File | Status |
|---|---|
| `/mnt/user-data/outputs/save_page_sections_action.go` | Patched (swap + belt-and-braces), syntax-checked with gofmt, ready to deploy |
| `/mnt/user-data/outputs/enrich_fix.diff` | 94-line unified diff against user-uploaded file |
| `/mnt/user-data/outputs/HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md` | This handoff |
| `/mnt/project/002_system_architecture.md` | Needs minor update (next section) |
| `/mnt/project/016_debugging_guide_v2.md` | Could use a new entry for silent-enrichment failure class (next section) |

## Suggested doc updates (not applied — user should review first)

### 002_system_architecture.md

Around line 314, the post-processing chain reads:
```
save_page_sections (store rendered sections in page_components)
```
Consider expanding to:
```
save_page_sections (store rendered sections in page_components, enriching
  slot_name from pages.sections planned order and linking component_id
  via content_components.function lookup)
```

Around line 734 (Layer 1 QA checks), the "Is component_id NULL?" check deserves a pointer to what that NULL means:
```
Layer 1: Structural Checks (algorithmic, no LLM)
  → "Is component_id NULL?" — indicates save_page_sections couldn't match
    the section's slot_name to a content_components.function. Either
    the component doesn't exist (library gap) or the name didn't make it
    through save's enrichment (extractSectionsFromMetadata fallback + HTML
    attr parse should now cover most cases).
```

### 016_debugging_guide_v2.md

Add a new row to the symptom table:

| Symptom | Cause | Diagnosis |
|---|---|---|
| `page_components.component_id = NULL` after rebuild, but `rendered_html` populated | save_page_sections ran, but enrichSectionsWithComponentIDs skipped every section — ComponentName was `"section"` (the extractSectionsFromMetadata default) because sections_metadata from content-writer lacks a component_name field | Check logs for `enrichSectionsWithComponentIDs: invoked` followed by the ABSENCE of `linked component` or `no match found`. After the belt-and-braces patch, look for `adopted data-component as name` logs. Long-term fix: make compile_page_sections emit component_name in each metadata entry. |

Both changes are small and pure-additive. Not applied because user should decide on wording.
