# RUNBOOK — `bugfix_305_negation_gate`

Every command here was hard to get right once. The gotcha is attached to the command, not filed
somewhere else.

## 0. The regex rule that governs every query in this lane

**Postgres has no `\b`.** There `\b` is a *backspace character*; the word boundary is `\y`. A Go
pattern pasted into psql silently matches nothing and returns a confident zero
(`LANDMINES.md:4219`). Go's RE2 spells it `\b` and has no `\y`.

**So: prove the pattern before you quote the count.** One query, both arms:

```sql
SELECT (SELECT count(*) FROM regexp_matches(
          'not a demo, but a product', '\ynot (?:just )?[^.;:]{2,50},\s*but\y', 'gi')) AS must_be_1,
       (SELECT count(*) FROM regexp_matches(
          'we ship on Kubernetes and Kafka',   '\ynot (?:just )?[^.;:]{2,50},\s*but\y', 'gi')) AS must_be_0;
```

## 1. The distribution census (what the gate will cost)

`1,503` rows is one week of `page-content-writer`. Needs `SET statement_timeout` — the default kills it.
⚠ **Do not run this while a `090` diagnosis run is in flight** — `bugs_open/305 §4` records a loop's
data request dying with `statement_timeout` while its filer ran a heavy `llm_call_log` query.

```sql
SET statement_timeout='150s';
WITH c AS (
  SELECT id,
    (SELECT count(*) FROM regexp_matches(response_text, '[a-z\)"''],\s+(?:not|never)\s+(?:just\s+|merely\s+|simply\s+|only\s+)?[a-z]', 'gi')) AS x_not_y,
    (SELECT count(*) FROM regexp_matches(response_text, '\ynot (?:just |merely |simply |about )?[^.;:]{2,50},\s*but\y', 'gi')) AS not_x_but_y,
    (SELECT count(*) FROM regexp_matches(response_text, '\yrather than\y', 'gi')) AS rather_than,
    (SELECT count(*) FROM regexp_matches(response_text, '[.!?]\s+(?:It|This|That|They|These|We)\s+(?:doesn.t|does not|isn.t|is not|won.t|will not|can.t|cannot|aren.t|are not|don.t|do not)\y', 'g')) AS neg_reveal
  FROM llm_call_log
 WHERE agent_type='page-content-writer' AND success AND created_at >= '2026-08-13')
SELECT count(*) AS calls,
       count(*) FILTER (WHERE x_not_y>=1)                                   AS ge1_xny,
       count(*) FILTER (WHERE x_not_y>=2)                                   AS ge2_xny,
       count(*) FILTER (WHERE not_x_but_y>=1)                               AS ge1_nxby,
       count(*) FILTER (WHERE rather_than>=1)                               AS ge1_rt,
       count(*) FILTER (WHERE neg_reveal>=1)                                AS ge1_reveal,
       count(*) FILTER (WHERE x_not_y+not_x_but_y+rather_than >= 2)          AS ge2_family
  FROM c;
```

## 2. Verify at the ARTEFACT, never at `updated_at`

A rerender bumps `page_components.updated_at` without regenerating anything — that is what made the
complained-of copy look five days newer than it was (`bugs_open/305 §3`).

```sql
SELECT p.url, pc.slot_name,
       pc.content_data::text ~* '\w,\s+(not|never)\s+\w' AS x_not_y,
       pc.content_data::text ~* '\yrather than\y'         AS rather_than
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'ai-agent-orchestration.com'
   AND p.url ~ '(model-directory|adoption-tracker|protocol-tracker)'
 ORDER BY 1, 2;
```

⚠ `pages` has **no `slug` and no `path`** column — it is `name` and `url`. Two agents wasted a round
trip each on that.

## 3. Read the live writer workflow (before anchoring a migration on it)

`jsonb_pretty` minus the 12 KB prompt template, or the output is unreadable:

```sql
SELECT jsonb_pretty(
         (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content}')
         #- '{config,prompt_template}')
  FROM agent_definitions
 WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

⚠ `jsonb - 'key'` fails with *"operator is not unique: unknown - unknown"* when both sides are
untyped literals in a `-Atc` one-liner. Use `#-` with a path array, as above.

## 4. Is the gate live? (three levels, in this order)

```bash
# a) the binary says what built it — per SERVICE, and the line SCROLLS
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <my-commit> <the-sha-from-that-line> && echo SHIPPED
```

```sql
-- b) the marker: did the gate actually run, and what did it do?
SELECT count(*) FILTER (WHERE collected_data::text LIKE '%__copy_gate%') AS runs_with_marker,
       count(*) AS orchestrations
  FROM orchestration_states
 WHERE created_at > '<roll timestamp>' AND collected_data::text LIKE '%page-content-writer%';

-- c) the retry rows. NOTE: the marker is present on SUCCESSFUL retries too, so filter failures on
--    success=false, never on a non-empty error_message (the bugs_open/119 precedent).
SELECT count(*) FROM llm_call_log WHERE error_message LIKE 'RETRY (bugs_open/305%';
```

⚠ **An empty `build provenance` grep means "scrolled out of range", not "unstamped"** — on a busy
service the line is gone within hours. Fall back to the binary probe with a control:
`kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe` and in the same
breath a sha that must be ABSENT. Never `strings` (not in the image).

## 5. Migration

`497` is written as `_HOLD` on purpose: it rewires the writer's step chain, so it must be applied
**after** the image carrying `rewrite_negations` is live, or the chain names a step the binary cannot
run. `SIDECAR_RE` excludes `_HOLD.sql` from the auto-apply sweep and still lists it.

```bash
./scripts/migration/run-migrations.sh              # DRY RUN — always first, per session
./scripts/migration/run-migrations.sh --apply      # takes EVERY pending file, not just mine
```

⚠ `--apply` applies **every** pending migration in the directory, not the one you are thinking about.
Read the dry-run list before running it, and expect other sessions' files in there. `_HOLD` is
excluded by `SIDECAR_RE` (`scripts/migration/run-migrations.sh:65`) and still listed, which is the
whole point of the suffix: it holds an ordering-critical file without hiding it. Renaming `497` out of
`_HOLD` is the deliberate act that says "the image is live".

## 6. Tests

```bash
go test ./platform/orchestration/datahelpers/ -run 'Negation|Voice|Strawman' -v
go test ./platform/orchestration/actions/ -run 'Negation|CopyGate|RewriteNegations' -v
go build ./... && go vet ./platform/... ./cmd/...
```
