# 215 — two emitted pages canonicalise to ONE name and the whole replan write dies on the unique index; which plans fail is decided by LLM emission variance

**Filed 2026-08-08** by the brochure_component_library lane, from a live failure
the same morning (fundamentallyai replan, corr
`1cb17b11-fd10-4d52-836c-36e2fa246ff6`):

```
step write_site_plan failed: ... insert site_plan_pages for "tool-llm-cost-calculator":
ERROR: duplicate key value violates unique constraint "idx_site_plan_pages_name" (SQLSTATE 23505)
```

**Verification route, declared per the 2026-07-31 owner ruling:** not through the
090 loop; substituted first-hand evidence from the failing run's own
`collected_data` plus the code path read at HEAD — the raw emission contains
both colliding names (query below), the canonicalisation site and the
unguarded insert are cited by line, and the day-before run on the SAME site
shows the differential (one variant emitted → no collision).

## Mechanism

1. The planner LLM may emit a page under its stem name AND its canonical name
   in one plan. This run: `llm-cost-calculator` (page_type `tool`, 3 sections)
   **and** `tool-llm-cost-calculator` (page_type `tool`, 0 sections — a stub),
   plus the same pattern in waiting: `tools` (2 sections) and `tool-tools`
   (0 sections). Read them from the failed run:
   ```sql
   SELECT p->>'name', p->>'page_type', jsonb_array_length(p->'sections')
   FROM (SELECT jsonb_array_elements(collected_data->'llm_plan'->'result'->'pages') p
         FROM orchestration_states
         WHERE correlation_id='1cb17b11-fd10-4d52-836c-36e2fa246ff6') x
   WHERE p->>'name' LIKE '%llm-cost-calculator%' OR p->>'name' LIKE '%tools%';
   ```
2. `WriteSitePlanAction` canonicalises each page independently
   (`datahelpers.CanonicalisePage`, called at
   `write_site_plan_action.go:277`): a `tool`-typed `llm-cost-calculator`
   becomes `tool-llm-cost-calculator`. **Nothing dedups the page list after
   canonicalisation**, so two entries now carry one name.
3. The per-page insert (`write_site_plan_action.go:379`) hits
   `idx_site_plan_pages_name` on the second one and the action errors. The
   write is transactional — verified in the same incident: no new `site_plans`
   row, previous plan still `is_current`, zero orphan rows — so there is **no
   data damage**, but the entire replan is lost.
4. Whether any given replan fails is therefore decided by **whether the LLM
   happened to emit both spellings that run**. The previous morning's replan
   of the same site emitted only the stem variant and succeeded; this one
   emitted both and died. A retry may pass. That makes this a low-frequency,
   zero-signature reliability hole in every `build-site-planner` run
   fleet-wide — the failure names neither the canonicaliser nor the LLM, and
   an operator's likeliest wrong conclusion is the one the error suggests
   (a stale unique index or concurrent write).

## Why the stub pages exist at all

The prompt's context shows the site's existing pages (which include the
canonical `tool-*` names) while the planning instructions talk about tool
pages by stem. The model, asked to enumerate every page (rule 17's
every-page requirement arrived the same morning — seed 333 — and plausibly
raised the odds of exhaustive enumeration), lists both spellings. Emission
variance, not a prompt defect as such — the write path must be safe against
it regardless.

## Fix candidates, ordered by what closes the door

1. **Dedup after canonicalisation, inside `WriteSitePlanAction`** (the only
   door): group the validated+canonicalised page list by final name; merge
   duplicates keeping the entry with sections (a stub loses to a composed
   entry; two composed entries = keep first, log loudly with both section
   lists). ~20 lines before the insert loop; makes the collision
   unrepresentable regardless of what the LLM emits.
2. **Planner prompt: name pages canonically** (tell it tool pages are always
   `tool-<stem>` and never to emit both). Reduces the odds; closes nothing —
   the write path would still be one emission away from dying.
3. Retry-on-23505 at the orchestration layer — treats the symptom, hides the
   defect, explicitly NOT recommended.

## How to verify a fix

Feed `WriteSitePlanAction` a page list containing a stem + its canonical
variant (unit: the failed run's raw pages array is a ready fixture) — the
write must succeed with ONE `site_plan_pages` row for the name, the composed
sections must win over the stub, and a log line must name the merge. Then
re-run the census: no plan write failure with `SQLSTATE 23505` on
`idx_site_plan_pages_name` in `orchestration_states.error` after the fix
ships (query `error LIKE '%idx_site_plan_pages_name%'`).

## Relations

Found while executing `bugs_open/151` candidate 1 Slice B (RFC_016 §3a option
(a) compliance replan — the failure cost that observation run);
`bugs_open/204`/`214` (the same wire's other positional/naming traps);
`datahelpers.CanonicalisePage` (the canonicaliser itself is correct — the gap
is the absent post-canonicalisation dedup).
