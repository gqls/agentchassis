# RUNBOOK — webdesign.co.uk

Every command that was hard to get right, with its gotcha attached. When one
changes, change it **here**.

Shorthand used throughout:

```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
```

---

## 1. Is the domain wired?

```bash
curl -sS -m 10 https://webdesign.co.uk/ | head -c 300
```

**Verified 2026-07-25:** returns the B2 worker's own JSON —
`{"error":"B2 returned error","objectKey":"webdesign.co.uk/index.html","status":404,…NoSuchKey}`
— from a Cloudflare edge IP. That output is the good news: registration, DNS,
the CF zone and the edge→B2 worker route all already exist, and the **only**
missing thing is content in `b2://portfolio-sites/webdesign.co.uk/`.

A plain connection error or a registrar parking page would have meant real
owner-side work. It did not.

**Deploy needs no new config.** `~/projects/sites/.github/workflows/deploy-to-b2.yml`
discovers changed top-level `<domain>/` dirs by git diff and looks the Cloudflare
zone up **by name** at purge time — there is no static zone map to extend. And
`resolveGitRepoNameDB` (`platform/orchestration/actions/helpers.go:220-249`)
defaults to repo `sites` when `sites.github_repo` is empty, which is the
gqls/sites → B2 path. **So leave `sites.github_repo` empty.**

**[OWNER-CHECK]** the deploy Action's `CF_API_TOKEN` must be able to see the
webdesign.co.uk zone. Non-blocking; the symptom of failure is only a stale cache.
On the first deploy, read the Action log: a null `ZONE_ID` means the token needs
the zone added to its scope.

---

## 2. Submission and the parked cascade

**Start the watcher first.** The cascade emits each next item at `status='triaged'`
(`create_work_item_action.go:141`), which is immediately dispatchable, and
`build-pipeline-trigger` fires every 120s.

```bash
cd docs/agent_docs/docs024_key_docs_latest/webdesign_couk/scripts
./watch_park_webdesign.sh webdesign.co.uk &        # polls every 5s
```

Then submit (fresh mode — no `--from`):

```bash
./082_submit_domain_unified.sh webdesign.co.uk \
  --email <owner-email> \
  --mission-file docs/agent_docs/docs024_key_docs_latest/webdesign_couk/MISSION_webdesign_couk.txt
```

**Gotcha — `--mission-file` is single-lined.** The script does
`sed 's/\\/\\\\/g; s/"/\\"/g' | tr '\n\t' '  ' | sed 's/  */ /g'`, so the brief
becomes one JSON string. Paragraph structure is lost; write the brief so it reads
correctly as continuous prose. Do not put literal `"` or `\` in it beyond what
you want escaped.

**Gotcha — dispatch queues.** Publish→run-start was measured at ~29 minutes under
normal load. A missing `orchestration_states` row is almost always latency, not a
drop. Find the run by payload, never by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'domain' = 'webdesign.co.uk'
ORDER BY created_at DESC;
```

Lock the site the moment the row exists (belt — see the watcher header for why it
is not braces):

```sql
UPDATE sites SET locked_at = now(), locked_by = 'webdesign-couk-standup'
WHERE domain = 'webdesign.co.uk';
```

### The airlock (release one leg at a time)

```bash
# 1. see what is parked
$PSQL -c "SELECT id, item_type, handler_agent, status, priority
          FROM site_work_items
          WHERE site_id = (SELECT id FROM sites WHERE domain='webdesign.co.uk')
          ORDER BY created_at;"

# 2. allowlist it, then release it
echo "<item-uuid>" >> scripts/park_allowlist.txt
$PSQL -c "UPDATE site_work_items SET status='triaged' WHERE id='<item-uuid>';"

# 3. watch progress in orchestration_states — NOT site_work_items.updated_at,
#    which is not maintained (bugs_open/035)
```

Legs in order: `needs_domain_research` → `needs_strategy` → `needs_briefing` →
`needs_site_plan` → (`needs_composition`, then `needs_design` which `depends_on` it).

**The design pin goes in after the classifier leg and before the design leg** —
see §3. There is no time pressure inside a park; that is the whole point.

---

## 3. The design pin

`SQL_p3_design_intent_pin.sql` — supersede + merge, never overwrite. Modelled on
`robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql`.

Why the ordering matters: the classifier has a `write_design_intent_spec` step,
and `WriteSiteSpecAction` (`site_spec_actions.go:120-310`) deep-merges NEW over
OLD **with no `pinned` guard** — `site_specs.pinned` is honoured only by
evidence_base code. A pin written before the classifier runs gets partially
overwritten.

Why a pin is needed at all: webdesign-agent's `analyze_design` step renders only
the **structured** `design_intent.palette` / `.typography` blocks into its prompt.
Free-text `colour_mood` / `style_direction` is never rendered, so a site with only
prose intent falls into the LLM's "no design intent exists → invent" branch and
re-rolls its palette every run.

Verify before releasing anything downstream:

```sql
SELECT data->'palette'->'reference_values', data->'typography'->'reference_values'
FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain='webdesign.co.uk')
  AND aspect='design_intent' AND is_current;
```

`idx_site_specs_current` is UNIQUE on `(site_id, aspect) WHERE is_current` — so
the write must supersede the old row in the same statement. The column is `data`,
not `specs` (the `site_specs.specs.<aspect>` path is the *prompt's* keying).

---

## 4. Colour lands in four places

A composition change must move all four, or the site drifts:

```sql
SELECT p.colours, ct.color_palette, sc.color_palette, l.scheme, l.name
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct        ON ct.id = sc.css_theme_id
JOIN palettes p           ON p.id  = ct.palette_id
JOIN layouts l            ON l.id  = ct.layout_id
WHERE s.domain = 'webdesign.co.uk';
```

`style_collections.color_palette` is the one component rendering actually reads.
Expect layout `tool-portal-light` / `scheme='light'`.

---

## 5. Verify the artefact, never the status

`complete` is not proof the work happened. Standing checks:

```bash
curl -s https://webdesign.co.uk/tools/index.html | grep -c '#0055ff'   # expect 0
curl -s -o /dev/null -w '%{http_code}\n' https://webdesign.co.uk/tools/assets/site-header.js
```

```sql
-- a complete item that actually failed
SELECT id, item_type FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='webdesign.co.uk')
  AND status='complete' AND result->'response'->>'status'='failed';
```

---

## 6. Ops rules inherited from other workstreams

- **Never parallel-dispatch `build-dispatch-loop`.** Single-flight: dispatch only
  when nothing on the site is `claimed`. (`find_dispatchable_site` already skips a
  site with a claimed item — do not defeat that by hand.)
- **No orchestration within ~300s of a chassis pod restart** — the spawn is
  silently dropped.
- **kcat payloads must be one line**; `kcat -P -c 1` with a heredoc.
- **Watch batches with an absolute `created_at`/`completed_at` cutoff**, never
  `now() - interval`, and never `updated_at` (bugs_open/035).
- **`handler_agent IS NULL`** ⇒ the item sits `triaged` forever with no error.
- **Order children before their index page** — `query.pages_where_type` filters
  `pages.status IN ('active','deployed')`, so an index built first lists nothing.

---

## 7. The port pipeline (repo-local, no cluster needed)

```bash
go run ./cmd/webdesignport transform \
  --sites ~/projects/sites \
  --port  docs/agent_docs/docs024_key_docs_latest/webdesign_couk/port \
  --out   build/output/webdesign_couk
```

`build/output/` is gitignored — only the inputs (`port/*.json`, `port-compat.css`)
and the Go code are committed.

*(import / verify invocations get added here as they are built)*
