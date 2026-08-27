# RUNBOOK — bugs_open/414

Every command here was hard to get right, and the gotcha is attached to it. Change a command HERE,
not in your scrollback.

## 1. The retraction sweep — the one query this whole bug is about

Run it at RETRACTION time, over **every** aspect and all three surfaces, keyed on the CLAIM (a
distinctive phrase) and never on the key you removed. This is what found the residual that kept 414
alive; the daily mechanised form is `spec_supplies_claim` (CLM-030).

```sql
SELECT 'spec' AS surface, s.domain, ss.aspect FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE ss.is_current AND ss.data::text LIKE '%checked against the FCA handbook%'
UNION ALL SELECT 'component', s.domain, COALESCE(pc.slot_name,'') FROM page_components pc
 JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.content_data::text LIKE '%checked against the FCA handbook%'
    OR pc.rendered_html LIKE '%checked against the FCA handbook%'
UNION ALL SELECT 'work_item', s.domain, w.item_type FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.status NOT IN ('complete','cancelled','rejected')
   AND (w.summary LIKE '%checked against the FCA handbook%' OR w.spec::text LIKE '%checked…%');
```

- ⚠ **Run it over spec HISTORY the first time** (drop `is_current`): two dated rows ten days apart is
  what tells you an AGENT copied the instruction rather than a human writing it twice. Here it gave
  `content_direction` 08-02 (manual) and `strategy` 08-12 (`domain-strategist`).
- ⚠ **`_` IS A SQL WILDCARD.** `data::text LIKE '%acceptance_marker%'` also matches the prose
  "acceptance marker", so a KEY census silently becomes a text census. Escape it
  (`'%acceptance\_marker%'`) **or** assert the two forms as separate columns — which is better,
  because the escaping trap bites in both directions (escaping one manufactured 38 false findings on
  2026-07-31; see `LANDMINES.md`).

## 2. Stripping a phrase from a spec, guarded, history intact

Never a bare `UPDATE`: assert the exact text first, in a `DO` block that `RAISE`s, inside a
transaction. **A verify block of `SELECT`s cannot stop a `COMMIT`.**

```sql
BEGIN;
DO $$
DECLARE v_row uuid; v_cs text; v_tail text := $marker$<the exact sentence, leading space and all>$marker$;
BEGIN
  SELECT ss.id, ss.data->>'<field>' INTO v_row, v_cs
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='<domain>' AND ss.aspect='<aspect>' AND ss.is_current;
  IF v_row IS NULL THEN RAISE EXCEPTION 'guard: no current row'; END IF;
  IF right(v_cs, length(v_tail)) <> v_tail THEN
    RAISE EXCEPTION 'guard: field does not END with the exact sentence; tail is %', right(v_cs, length(v_tail));
  END IF;
  -- trim, jsonb_set, then assert the phrase survives NOWHERE in the row
  UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id=v_row;
  INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by) VALUES (…);
END $$;
COMMIT;
```

- **The unique partial index `idx_site_specs_current` is on `(site_id, aspect) WHERE is_current`** —
  supersede the old row BEFORE inserting, in the same transaction.
- **INDUCE THE GUARD FIRST.** Copy the script, corrupt the expected text (`sed 's/marker/MARKER/'`),
  run it, and watch it abort with nothing changed. A guard you have not seen fire is a comment.
- Use dollar-quoting (`$marker$…$marker$`) for the literal: the sentence contains single quotes.

## 3. Dispatching a copy repair through the framework

```sql
INSERT INTO site_work_items
 (site_id, source, item_type, severity, summary, spec, page_id, priority, handler_agent, status,
  created_by, item_key, pipeline, approval_mode, max_attempts, triaged_at)
VALUES (…, 'content_rewrite', 'high', …,
  jsonb_build_object('mode','edit_live', 'page_name', …, 'page_id', …,
                     'suggestion', <the rewrite guidance>, 'acceptance_test', …),
  …, 10, 'page-build-handler', 'triaged', …, 'content_rewrite:bug414:<page>', 'build', 'auto', 2, now());
```

- **`spec.mode='edit_live'` is the load-bearing key.** Without it `page-build-handler` regenerates
  the page's sections from scratch — measured at `bugs_open/178` as 4,439 → 1,806 chars on one page.
  It is read as `input_data.spec.mode` by the `load_current_section_content` step.
- **Priority is ASCENDING urgency: 10 beats 110.** lendzy's queue sat at 35–140; at 110 the item
  would have waited behind 16 others.
- **The handler regenerates ALL ready sections on the page, not just the slot you care about.**
  There is no slot filter on `plan_sections`. Snapshot every component's `content_data` md5 on the
  page first, so `edit_live`'s preservation of the untouched slots is checked rather than assumed.
- **The dispatcher is `build-pipeline-trigger` (30s) → `build-dispatch-loop`, `status IN
  ('triaged','approved')`, `max_items 5` per site by priority.** lendzy's previous claim had been 9
  hours earlier (selector starvation, `bugs_open/413`); these items were claimed within ~15 min.
- ⚠ **`resolve_links` / `spawn_content_writer` time out often** (fleet-wide on 2026-08-27, multiple
  lanes). `CHILD_ORCHESTRATION_FAILED` on attempt 1 is the handshake race, not your payload —
  `max_attempts 2` gives it a second run. Read `orchestration_states.collected_data->>'__step_error'`
  for `failed_step`, because a FAILED step can show COMPLETED with `error` NULL.

## 4. Exporting the live corpus for a claimscan dry run — the recipe that works

⚠ **Do NOT stream 15 MB through `kubectl exec`.** Two attempts failed differently and both looked
like success at first:

```bash
# ATTEMPT 1 — one big stream: "unexpected EOF", 2,283 of 2,585 rows, exit 0 on the pipeline.
# ATTEMPT 2 — a per-domain loop: `kubectl exec -i` ATE THE LOOP'S STDIN (pattern-check's
#             check_stdin_eater), so it processed one domain and stopped at 51 lines.
# WHAT WORKS — write inside the pod, then copy the file out:
kubectl -n ai-persona-system exec postgres-clients-0 -- bash -c "psql -U clients_user -d clients_db \
  -At -o /tmp/corpus.tsv -c \"SELECT s.domain || '::' || p.name || E'\t' || COALESCE(pc.slot_name,'') \
  || E'\t' || replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','') || E'\t' \
  || COALESCE(p.page_type,'') FROM page_components pc JOIN pages p ON p.id=pc.page_id \
  JOIN sites s ON s.id=p.site_id WHERE pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' \
  AND pc.locked_at IS NULL;\" && wc -l /tmp/corpus.tsv" </dev/null
kubectl -n ai-persona-system cp postgres-clients-0:/tmp/corpus.tsv ./corpus.tsv
```

**COUNT THE ROWS THREE WAYS AND MAKE THEM AGREE** — pod file, local file, and the DB:
`SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE pc.rendered_html IS
NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL;`. A short corpus scans CLEAN and reads
as a clean fleet — the failure mode a 2026-08-24 dry run hit when it silently dropped 414 rows.

Then: `go run ./cmd/claimscan -components ./corpus.tsv -show-suppressed` (exit 1 = findings).
`BANNED` lines are refusals; `PRACTICE` lines are warnings and are **excluded from the exit code**,
so grep for them explicitly or you will read a practice finding as clean.

## 5. Council submission

The schema is **nested** and the flat form is rejected client-side:
`{rationale, submitter, plan:{summary, edits:[{file,symbol,operation,rationale,sketch}], grounded_in:[], risks}}`
with `operation` one of `modify|add|remove|config_change` (**not** `create`).

```bash
DRY_RUN=1 ./docs/…/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh submission.json   # free
./docs/…/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh submission.json             # ~30 min
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## 6. Verifying at the artefact

```bash
scripts/probe-page-url.sh lendzy.co.uk about tool-affordability-complaint-checker-guide
```
Reads the recorded `pages.url` (never composes one) and enforces an invented-URL control, so a
parked/catch-all domain cannot read as healthy. Exit 2 is a REFUSAL, never a pass.
Then the §1 sweep expecting 0, and curl-and-grep by body with a fabricated-URL control on the same
domain plus a byte-delta check. **`complete` is not proof:** a lock- or decision-gated refusal
completes too — check `edit_result.skipped` / `locked` / `decision_gated`.
