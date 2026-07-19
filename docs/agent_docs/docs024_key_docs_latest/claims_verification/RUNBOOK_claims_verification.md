# RUNBOOK — Claims verification

Every command in this workstream that was hard to get right, with its gotcha
attached. Written retrospectively on 2026-07-19 (it should have existed from the
start — these were all recovered from scrollback, which is exactly the failure
the standing-five directive names).

`SITE` below is leopardessconsulting.co.uk =
`4851f6fc-71cf-4160-a270-e03d6d3e0732`. DB access is the CLAUDE.md one-liner:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## 1. Read the live evidence base

```sql
SELECT jsonb_pretty(data) FROM site_specs
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND aspect = 'evidence_base' AND is_current = true;
```

**Gotcha.** `evidence_base` is a *pinned* spec aspect and there is a unique index
on `(site_id, aspect) WHERE is_current` — so you never UPDATE it in place. Every
change is supersede-then-insert (see §2), and `pinned` must be carried forward or
the row silently loses its human-owned status.

## 2. Revise the evidence base (supersede + insert)

```bash
S=/path/to/scratch
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \
  -At -c "SELECT data FROM site_specs WHERE site_id='4851f6fc-...' AND aspect='evidence_base' AND is_current=true" > $S/eb.json
# edit $S/eb.json (jq or python), then:
{ echo "BEGIN;
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='4851f6fc-...' AND aspect='evidence_base' AND is_current=true;
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
VALUES ('4851f6fc-...', 'evidence_base', \$eb\$$(cat $S/eb.json)\$eb\$::jsonb,
        'hitl', '<what changed and why>', true, 'operator-claude', true) RETURNING id;
COMMIT;"; } | kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```

**Gotchas.**
- Use a **dollar-quoted tag** (`$eb$…$eb$`) — the JSON contains single quotes and
  regex backslashes that break ordinary quoting.
- Always `-v ON_ERROR_STOP=1` on a multi-statement pipe, or a failed UPDATE still
  lets the INSERT run and you get two `is_current` rows fighting the unique index.
- `jq -e . file.json` before sending. A malformed payload inside a dollar-quoted
  block fails deep in Postgres with a useless position number.

## 3. Scan live pages for claims (the operator CLI)

```bash
S=/path/to/scratch
# evidence base
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT data FROM site_specs WHERE site_id='4851f6fc-...' AND aspect='evidence_base' AND is_current=true" > $S/eb_live.json
# page components (unlocked only), base64 so HTML can't break the TSV
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At -F$'\t' -c \
 "SELECT p.name, COALESCE(pc.slot_name,''),
         replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n', '')
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id='4851f6fc-...' AND pc.rendered_html IS NOT NULL
    AND pc.rendered_html <> '' AND pc.locked_at IS NULL
  ORDER BY p.name, pc.position" > $S/components.tsv
# site-level chrome too (header/footer/head)
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At -F$'\t' -c \
 "SELECT '(site)', COALESCE(sc.slot_name,''),
         replace(encode(convert_to(sc.rendered_html,'UTF8'),'base64'), E'\n', '')
  FROM site_components sc WHERE sc.site_id='4851f6fc-...'
    AND sc.rendered_html IS NOT NULL AND sc.rendered_html <> '' AND sc.locked_at IS NULL" >> $S/components.tsv

go run ./cmd/claimscan -evidence $S/eb_live.json -components $S/components.tsv; echo "exit: $?"
```

**Gotchas.**
- `encode(...,'base64')` wraps at 76 chars — the `replace(..., E'\n', '')` is
  load-bearing or every row becomes many broken rows.
- Exit code is 1 when findings exist, 0 when clean: usable as a scripted gate.
- `go run` leaves no binary; if you `go build ./cmd/claimscan` instead, delete the
  stray `claimscan` binary from the repo root before committing.

## 4. Find where a banned claim actually lives

The rendered HTML is the symptom; `content_data` is the **render source** and is
what must be fixed, or the next re-render brings the claim back.

```sql
SELECT p.name, pc.slot_name, pc.position, pc.locked_at IS NOT NULL AS locked,
       pc.updated_at::date,
       (pc.content_data::text ~* '<pattern>') AS in_content_data,
       substring(regexp_replace(pc.rendered_html, '<[^>]*>', ' ', 'g')
                 from '.{0,80}<pattern>.{0,80}') AS snippet
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '4851f6fc-...' AND pc.rendered_html ~* '<pattern>';
```

**Gotcha — this one cost a whole round.** `regexp_matches(...)` without the `'g'`
flag returns only the FIRST match per row, so a sweep of the specs looked clean
after one pass and was not. Always:

```sql
SELECT aspect, (regexp_matches(data::text, '(?i)(.{0,160}<pattern>.{0,180})', 'g'))[1]
FROM site_specs WHERE site_id='4851f6fc-...' AND is_current = true;
```

## 5. Fix copy and redeploy (no LLM)

Edit **both** `content_data` and `rendered_html` in one transaction, then queue a
plain re-render:

```sql
INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity,
                             summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT '4851f6fc-...', '<page_id>', 'manual', 'build', 'page_rerender', 'high',
       '<why>', jsonb_build_object('page_name','<page>'), 25, 'page-rerender',
       'triaged', 'operator-claude', 'page_rerender:<page>'
WHERE NOT EXISTS (
  SELECT 1 FROM site_work_items w WHERE w.site_id='4851f6fc-...'
    AND w.item_key='page_rerender:<page>'
    AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled'));
```

**Gotchas.**
- `ON CONFLICT (site_id, item_key)` **fails** — `idx_swi_dedup` is a *partial*
  unique index (excludes terminal statuses), and ON CONFLICT cannot use it. Use
  the `WHERE NOT EXISTS` form above.
- Omit `spec.reason` deliberately. With `reason` ∈ {`image_landed`,
  `section_data_resolved`} the handler re-resolves fields; for a hand-corrected
  copy fix you want the plain assemble of the HTML you just wrote.
- Never route a copy fix through `needs_page` — that regenerates via the writer,
  which can reintroduce the claim if any spec still carries it.

## 6. Verify a deploy reached the live site

```bash
for url in faq.html technical-architecture.html; do
  html=$(curl -s "https://leopardessconsulting.co.uk/$url")
  echo "$url banned=$(echo "$html" | grep -icE '<pattern>') corrected=$(echo "$html" | grep -icE '<new wording>')"
done
```

Wait for propagation with an until-loop rather than a fixed sleep — commit →
GitHub Action → S3 takes a variable minute or two.

## 7. Run the discovery check / auditor by hand

Both are spawned through the generic entry point. Full kcat invocation is in
`NOTES_claims_verification.md`; the parts that bite:

- **One quoted JSON payload, single line.** kcat `-P` splits on newlines, so a
  pretty-printed body becomes many truncated messages.
- **No dispatch within ~300s of a chassis pod restart** — the spawn is silently
  dropped (CLAUDE.md).
- A **clean** check leaves no positive trace: checks emit findings only. Verify a
  clean pass by `status=COMPLETED` **plus** zero new work items, never by looking
  for a success line.

```sql
-- did the claims check produce anything?
SELECT summary, status, created_at FROM site_work_items
WHERE site_id='4851f6fc-...' AND item_type IN ('claims_unverified','stale_evidence')
ORDER BY created_at DESC;
```

## 8. Confirm the writer really received the whitelist (V2)

```sql
SELECT step_name, output_tokens, max_tokens,
       (prompt_rendered LIKE '%Verified Facts (the ONLY numbers%') AS had_whitelist
FROM llm_call_log
WHERE agent_type = 'page-content-writer' AND created_at > '<dispatch time>'
ORDER BY created_at;
```

**Gotchas.** The column is `prompt_rendered`, **not** `rendered_prompt`. And
`agent_error_log`'s timestamp is `occurred_at`, **not** `created_at` — both cost a
monitor restart apiece.

## 9. After the next image build (V4 activation)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c refresh_evidence_base'
```

Non-zero → apply `SEED_evidence_freshness_scheduled_task.sql`. Zero → the action
is not in the image; do **not** apply the seed (a seed naming an unregistered
action fails at runtime).

## 10. Submitting this workstream's code to the council

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

**Gotchas learned the expensive way** (all three cost a full round trip):
- `operation` must be one of **`modify | add | remove | config_change`**. A new
  file is `add` — **`create` is rejected**, server-side, after the client-side
  checks have already passed.
- A file may appear in **only one edit per stage**. Describing a new file with a
  `create` + a `modify` is two edits on one path and is refused.
- Track your run by **submission correlation, never by recency** — every council
  run is `owner_agent_type='generic'`, so "the most recent generic orchestration"
  will happily hand you another thread's run:

```sql
SELECT orchestration_id, status, current_step FROM orchestration_states
WHERE collected_data::text LIKE '%<SUBMISSION_CORR>%' ORDER BY created_at DESC;
```

- An absent state row means **not started yet**, not "dispatch lost".
- Two very different failures both land on `complete_invalid`; tell them apart:

```sql
SELECT collected_data->'__step_error' FROM orchestration_states WHERE orchestration_id='<id>';
```

`failed_step = persist_submission` → your JSON is malformed.
`failed_step = council_decide` → a reviewer was truncated (`bugs_open/019`); your
plan was fine and was fully reviewed.
