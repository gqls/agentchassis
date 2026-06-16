# Runbook — BUG: gamesdesign.co.uk silent no-op rebuild

**One-line:** the index rebuild reports success but the live page stays stale,
because the page-content-writer regenerates far less content than exists
(~3,000 chars vs ~13,000–20,000), the content-regression guard CORRECTLY blocks
the overwrite with a hard error, yet the `page_rerender` work item is still
recorded `complete` — so a blocked, failed save presents as a successful
rebuild.

**Status:** diagnosis CONFIRMED from runtime evidence + the assembled code
bundle (2026-06-14). Two distinct faults, below. Fix not yet applied.

**How this was reached** (so the conclusion is auditable, not assumed): the task
sentence hypothesised "generated sections never reach save_page_sections". The
runtime evidence (`dbcontext -runtime-site gamesdesign.co.uk -runtime-page
index`) FALSIFIED that — sections do reach save; save fails on the regression
guard. The static code alone could not have told us which path fired (it shows
several success-with-zero early returns); the logs did. Method note for next
time: do not conclude the failing branch from code without the runtime trace.

---

## The evidence

**Live page is content-rich and intact** (so nothing has been lost — the page
is being PROTECTED, not clobbered):
```sql
SELECT pc.build_status, count(*), sum(length(pc.rendered_html)) AS html_len
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name = 'index' GROUP BY pc.build_status;
-- deployed | 5 | 34375
```

**Every recent rebuild fails the SAME way** (`agent_error_log`, 15 rows, agent
`page-build-handler`, step `save_sections`, action `save_page_sections`):
```
content regression blocked: new content has 2854 chars of text vs 13040 existing
content regression blocked: new content has 3323 chars of text vs 15072 existing
content regression blocked: new content has 3254 chars of text vs 14544 existing
… (every row ~2.8k–3.5k new vs ~13k–20k existing)
```

**Yet the work item completed** (`site_work_items`): the `page_rerender` rows
for index show `status = complete`, `0/3 attempts`, no error. A hard error from
the save step did not propagate to the work-item status.

---

## Fault 1 — the REAL bug: generation produces too little content

The page-content-writer (the LLM step that produces the page's sections, feeding
`save_page_sections`' `CollectedData`) is regenerating ~3,000 chars of text for a
page that legitimately holds ~13,000–20,000. The guard's own comment
(save_page_sections_action.go L328–330) names this exact cause:

> "Refuse to overwrite content-rich pages with empty template shells. This
> prevents LLM failures (credit exhaustion, timeouts, empty responses) from
> wiping good content."

So the guard is a SYMPTOM-CATCHER working as designed; the defect is upstream.

**The guard logic (confirmed, L341–361):** if existing deployed text > 200 chars
and `newTextLen < existingTextLen/4`, it returns a hard error
(`return nil, fmt.Errorf("content regression blocked …")`). ~3k < ~13k/4 trips
it every time.

**Fix surface — investigate the generation, not the save:**
- The page-content-writer agent/step that fills the section content (NOT in this
  bundle's scope — assemble a second bundle scoped on the generation side; the
  in-scope `plan_sections_action.go` is the planning step and is the bridge to
  find it).
- Likely lines of enquiry (each falsifiable against logs/DB):
  - is the LLM returning truncated/empty content (credit/timeout/empty
    response) — the guard's hypothesised cause? Check the content-writer's own
    log lines + token usage for those orchestrations.
  - is only a SUBSET of sections being generated (e.g. 1 of 5), so the total is
    short but each is fine? Compare section count generated vs the 5 deployed.
  - is the writer being handed a stale/empty brief or the wrong page context for
    a RECREATE/re-adoption rebuild specifically (the task's trigger)?
- DO NOT "fix" by loosening the guard. The guard is correct; loosening it would
  let the short content overwrite the good page — the worse outcome the guard
  exists to prevent.

## Fault 2 — the VISIBILITY bug: a failed save rolls up to a complete work item

`save_page_sections` returns a hard error (Fault 1's guard), logged as
`step save_sections failed` in `agent_error_log` — but the `page_rerender`
`site_work_items` row is `complete`. So the rebuild PRESENTS as success. This is
the "invisible without inspecting stored components" half of the symptom.

**Fix surface — the status-propagation path:**
- the `page-build-handler` workflow: how a step error in `save_sections` maps to
  the work-item terminal status. A step that errors should drive the work item
  to a failed/blocked status (or at minimum a distinct "blocked" state), not
  `complete`.
- check whether the handler swallows the step error (treats any return as
  done), or whether the work-item completion is written before/independently of
  the step outcome.
- this fault is SEPARATE from Fault 1 and worth fixing regardless: even once
  generation is fixed, a future genuine regression-block should be VISIBLE as a
  non-complete work item, not silently `complete`.

---

## Reproduce / inspect

```bash
CK=docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit
# (run from inside contextkit: use ./cmd/…; from repo root: ./$CK/cmd/…)

# 1. The evidence bundle (this bug, save side) — already produced:
go run ./cmd/bundle -analysis analysis6.json -root ~/projects/agentchassis \
  -constitution <path>/thin_slice_constitution.md -step debug \
  -task "gamesdesign index rebuild: save blocked by content regression (~3k vs ~13-20k); work item still completes. Confirm the generation-side shortfall and the status rollup." \
  -scope platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction \
  -scope platform/orchestration/actions/plan_sections_action.go \
  -include platform/orchestration/actions/registry.go \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables page_components,pages,site_work_items \
  -runtime-site gamesdesign.co.uk -runtime-page index -capabilities \
  -out /tmp/bundle_gamesdesign.md

# 2. The Fault-1 bundle (generation side) — the next step. Find the
#    content-writer agent/step first (grep the workflow / registry for the step
#    that produces section content feeding save_page_sections), then:
#    -scope <content-writer action>.go  -scope plan_sections_action.go
#    plus the agent_definitions row for the page-content-writer
#    (-doc a dbcontext -rows dump of its workflow prompt).
```

### How to build the Fault-1 (generation-side) bundle — step by step

This bundle (`save_page_sections` + `plan_sections`) shows the guard firing; it
does NOT contain the content-writer, so it can't show WHY the input is short.
But it contains the BRIDGE to find it. The chain is named in the code:
`plan_sections_action.go` L497–498 — *"This sits between load_page_record and
spawn_content_writer in the page-build-handler workflow"* — so the workflow is
`load_page_record → plan_sections → spawn_content_writer`, and the generating
agent is **page-content-writer**.

The mechanism that makes generation short (confirmed from the bundle,
`plan_sections_action.go` L922–962): `plan_sections` emits a per-section plan;
each `sectionPlanItem` has a `Status` (`ready | deferred | skipped`) and
`LLMFieldSpecs` (the `source:"llm"` fields the writer must produce). The
**page-content-writer only generates the fields it is handed.** So if the plan
marks most index sections `deferred`/`skipped`, or hands over few
`LLMFieldSpecs`, the writer is correctly asked for little — and produces the ~3k
that trips the guard. That makes the section-plan status the prime suspect, and
it is checkable BEFORE building the next bundle.

**Step A — the cheapest diagnostic first (read-only): what did plan_sections
decide for the index?** The plan is what the writer acts on; if it's mostly
`deferred`/`skipped`, Fault 1 is in planning, not the writer.
```sql
-- section rows + their status for the gamesdesign index (adjust table/column
-- names to your schema — check with: dbcontext -schema site_plan_sections,pages):
SELECT sps.section_name, sps.status, sps.component_id
FROM site_plan_sections sps JOIN pages p ON p.id = sps.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name = 'index'
ORDER BY sps.position;
```
- Many `deferred`/`skipped`, or far fewer rows than the 5 deployed components →
  the defect is in `plan_sections` (why is it not marking sections `ready`?) →
  scope the next bundle on `plan_sections` + the readiness inputs it reads.
- All `ready` with full field specs → planning is fine; the writer itself is
  returning short/empty content → scope on the writer + its prompt.

**Step B — find the content-writer step and its action.** Trace the workflow
from the bridge:
```bash
# the spawn step named in plan_sections:
grep -rn 'spawn_content_writer\|page-content-writer\|content_writer' \
  platform/orchestration/actions/ platform/ | grep -iv '_test' | head
# the agent definition + its workflow (this is the prompt that decides how much
# the LLM writes):
#   dbcontext -rows "SELECT type, version, default_config FROM agent_definitions
#                    WHERE type='page-content-writer' ORDER BY version DESC LIMIT 1"
```

**Step C — assemble the Fault-1 bundle** (one command; same wrapper, different
scope + the writer's prompt as a doc). Run from inside contextkit (`./cmd/…`) or
repo root (`./$CK/cmd/…`):
```bash
# first capture the writer's workflow prompt as a doc:
go run ./cmd/dbcontext \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -rows "SELECT default_config FROM agent_definitions WHERE type='page-content-writer' ORDER BY version DESC LIMIT 1" \
  > /tmp/content_writer_prompt.md

go run ./cmd/bundle -analysis analysis6.json -root ~/projects/agentchassis \
  -constitution <path>/thin_slice_constitution.md -step debug \
  -task "gamesdesign index: regenerated content is ~3k vs ~13-20k existing, tripping the content-regression guard. Determine why the page-content-writer produces so little — is plan_sections handing over few ready sections / LLMFieldSpecs, or is the writer returning short output?" \
  -scope platform/orchestration/actions/plan_sections_action.go \
  -scope <the content-writer action file found in Step B>.go \
  -include platform/orchestration/actions/registry.go \
  -doc /tmp/content_writer_prompt.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables site_plan_sections,page_components,pages \
  -runtime-site gamesdesign.co.uk -runtime-page index \
  -out /tmp/bundle_gamesdesign_generation.md
```
This bundle has: the planning code, the writer's code, the writer's actual
prompt (the doc), the plan/section schema, and the runtime trace — i.e. the
whole generation path in one artefact. That is what tells you why the generation
is short, which is the fix for Fault 1.

```bash
# (Fault-2 status-rollup investigation is separate — see Fault 2 above. Its
#  bundle scopes the page-build-handler completion path, not the writer.)
```

**Cheap confirmations (read-only):**
```sql
-- Did this rebuild snapshot the (good) existing content to history before the
-- block? The guard returns BEFORE the snapshot/DELETE (L357 is above L372/L399),
-- so the deployed components are untouched — confirm none were deleted:
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id=(SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND p.name='index' AND pc.build_status='deployed';
-- expect still 5 (the guard fired before any DELETE — page intact).
```

---

## Verification once fixed

- **Fault 1 fixed** when a rebuild of the gamesdesign index generates text within
  range of the existing ~13-20k (not ~3k), the guard does NOT fire, and the page
  re-deploys with refreshed content (`page_components` rendered_html changes;
  timestamps advance with NEW content, not stale).
- **Fault 2 fixed** when a save step that errors (force a regression-block in a
  throwaway test, or observe a real one) drives the `page_rerender` work item to
  a failed/blocked status, NOT `complete`.
- Both: the symptom — "success reported, page stale" — cannot recur, because
  either the content is genuinely refreshed (F1) or the failure is visible (F2).

## What NOT to do

- Do not loosen or remove the content-regression guard — it is correct and is
  the only thing currently protecting the live 34k-char page from a 3k overwrite.
- Do not conclude which generation failure mode (truncation vs subset vs stale
  brief) without the content-writer's own logs/token usage — the same
  don't-guess-the-branch discipline that corrected the original diagnosis.
