# RUNBOOK — webdesign.uk build service

Commands that were hard to get right, with the gotcha attached. When one changes,
change it **here**.

---

## Grounding queries

DB access (from CLAUDE.md):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

**Gotcha:** there is no `postgres` role — `psql -U postgres` fails with
`FATAL: role "postgres" does not exist`. Use `clients_user` for everything,
including `\l`.

### Fleet size (the §12 figure — re-run before quoting any scale claim)

```sql
SELECT count(*) total,
       count(*) FILTER (WHERE status='deployed') deployed,
       count(*) FILTER (WHERE domain LIKE 'pool-%') pools
  FROM sites;
```

**Gotcha:** `count(*)` alone answers a different question — 17 of the 32 rows are
empty `pool-*.internal` shells and 1 is `system.internal`. A bare total overstates
the fleet by more than 2×.

### What a new site actually needs seeded

```sql
SELECT aspect, is_current, created_at::date FROM site_specs
 WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com')
 ORDER BY created_at;
```

Read `oufe/SEED_2026-07-25_oufe_site_and_specs.sql` alongside it — its preamble
explains *why* three of those aspects must exist **before the first page is
written** (`evidence_base` gates the whole claims layer and silently no-ops if
absent; a missing site email makes the hallucinated-email check fail open;
a missing `imagery_style_guide` makes `content_hero` generate unstyled).

### Agent definitions

```sql
SELECT type, category, status FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
 ORDER BY type;
```

**Gotcha:** the column is `display_name`, **not** `name` — `SELECT type, name`
fails. Run `\d agent_definitions` first, as CLAUDE.md says.

### Model cost of a build (do this before pricing anything)

```sql
\d llm_call_log
SELECT count(*), min(created_at)::date, max(created_at)::date FROM llm_call_log;
-- 45,205 rows, 2026-03-25 → 2026-07-28  (checked 2026-07-28)
```

**Gotcha, inherited from idea.uk:** a per-run figure is only a **floor** unless
every call in the run is logged. `EVIDENCE_2026-07-27_ai_unit_economics.md`
records exactly this — usage logging was gated on cache activity, so calls whose
system prompt fell under the cacheable minimum (512 tokens on Opus 5, 1024 on
Sonnet 5) logged nothing, and 3 of 5 calls were invisible. Check the denominator
before quoting a total.

### Which model is the fleet actually on

```sql
SELECT model, count(*), max(created_at)::timestamp(0) latest FROM llm_call_log
 WHERE created_at > now() - interval '4 days' GROUP BY 1 ORDER BY 3 DESC;
```

**Gotcha:** run this before acting on any statement of the form "we're on model
X". On 2026-07-29 the answer was 1,468 Sonnet 5 and **zero Fable 5** while the
session itself ran Fable — the phrase covered the session and an intention, not
the fleet. DB model config is live on write, so the distinction decides whether
the pre-flight checks below are still owed.

### Fable-5 pre-flight (PLAN §7b) — in this order

1. **Org data retention ≥ 30 days.** A ZDR org gets `400 invalid_request_error`
   on *every* Fable request, and the error names the request, not the setting.
2. **Grep the chassis LLM call layer** for params Fable rejects with a 400:

```bash
grep -rn "temperature\|top_p\|top_k\|budget_tokens\|\"thinking\"" \
  --include="*.go" platform/ internal/ | head -40
```

   **Gotcha:** a clean grep of Go source is not the whole answer — model params
   also live in `agent_definitions.default_config` as live DB rows, so check
   both. (Precedent: all 16 council seats set `max_tokens=8000` from config.)
3. **Then** measure one real Fable build from `llm_call_log`.

## DNS

```bash
dig +short webdesign.uk A
dig +short webdesign.uk NS
```

Empty is **expected** as of 2026-07-28 — the owner has not pointed it yet. Empty
is therefore not evidence about registration either way.

## Deploy path facts worth not re-deriving

- Default artefact repo is `sites` (→ GitHub Action → B2). VM-hosted sites
  override to `vm-sites` via `sites.github_repo`; resolution is
  `resolveGitRepoNameDB` and `git_repo_resolution_test.go` documents the bug it
  prevents.
- Only `idea.uk` and `relojistas.com` are on `vm-sites`. Only `relojistas.com`
  has a non-empty `deploy_config`.

## Not yet written

Everything operational — provisioning the box, the engine build, Stripe test
mode, the preview vhost. Those arrive with P1; this file gets them then, not
before.
