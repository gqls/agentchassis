# RUNBOOK — bug 122 contrast / ink slots

Every command here was needed to get an answer right, with its gotcha attached.
Newest additions at the bottom of each section.

---

## Measure what a visitor actually sees

```bash
python3 scripts/render_audit.py https://site.example/ [more urls...]
```

- Headless Chromium, **every** element (`body *`), computed style, alpha-composited
  backdrop. This is the ONLY thing that answers the contrast question — see the
  gotchas below for the three cheaper methods that all give wrong answers.
- `(over an image — ratio approximate)` findings are NOT firm. The tool reports them
  rather than asserting them; discount them when counting.
- ~30–60s per page. Ten homepages is about 5 minutes, so background it.

**GOTCHA — the audit output goes to stdout and a long run's head scrolls out of a
tool result.** Redirect to a file in the scratchpad and grep it, or you get only the
last site's detail and a total:

```bash
python3 scripts/render_audit.py <urls> > audit_$(date +%F).txt 2>&1
grep -E '^(FAIL|ok)' audit_2026-08-06.txt      # per-site summary
```

## Fetch a served stylesheet

```bash
curl -fsS -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36" \
  https://<domain>/assets/css/styles.css -o css_<domain>.css
```

**GOTCHA — a bare `curl`/`urllib` gets `403 Forbidden` on every site.** The origin
rejects a non-browser user agent. A whole 19-site census returned `403` on every row
and read exactly like a routing failure. Send a browser UA.

**GOTCHA — the path is `/assets/css/styles.css`, not `/styles.css`.** The short path
404s. Read it off the page rather than guessing:
`curl -fsS https://<domain>/ | grep -oE '<link[^>]*stylesheet[^>]*>'`

## The three cheaper measurements that are all WRONG

Recorded because each one looked authoritative and this bug's own history was
distorted by the first:

1. **A regex over `styles.css`** cannot resolve the cascade (which `a` rule wins),
   cannot see ancestors, alpha or gradients, and cannot know which variable a given
   element actually uses. This produced 122's superseded table, in which
   dartsonline "scored" 1.11 because the audit compared its background against
   itself.
2. **A palette row / `AuditPalette`** cannot see a literal that is in no palette —
   which is where the hardcoded inks live.
3. **Comparing a foreground against the PAGE background** instead of the element's
   own flags the header, logo and nav as failing when they sit on their own
   `--color-header-bg` and are fine at 17:1. "Fixing" those breaks what works.

## Find which artefact owns a failing CSS rule

The audit names a selector; this says which of the four surfaces defines it.

```sql
SELECT 'layouts' AS src, name FROM layouts WHERE css_template LIKE '%<selector>%'
UNION ALL SELECT 'css_snippets', name FROM css_snippets WHERE css_content LIKE '%<selector>%'
UNION ALL SELECT 'content_components', name FROM content_components
  WHERE html_template LIKE '%<selector>%' AND COALESCE(is_active,true);
```

Then pull the rule itself — `substring(... from '\.<selector>\s*\{[^}]*\}')`.

**GOTCHA — component CSS lives INSIDE `html_template`, not in a CSS column.**
`content_components` has no `css_styles` column (I assumed one and got
`ERROR: column "css_styles" does not exist`). `css_snippets.css_content` is a
different, much smaller surface: 21 rows, and **0 of them mention `--color-primary`
at all** — so a census run only against `css_snippets` will report zero and look
like a clean bill of health.

## Census: is a palette colour used as an ink or as a fill?

```sql
SELECT name,
       (SELECT count(*) FROM regexp_matches(css_template, '[^-]color:\s*var\(--color-primary[),]', 'g')) AS ink_uses
FROM layouts WHERE css_template ~ '[^-]color:\s*var\(--color-primary[),]' ORDER BY 2 DESC;
```

**GOTCHA — the `[^-]` prefix is load-bearing.** Without it, `background-color:` and
`border-color:` match as inks and every count is inflated.

## Is a proposed new CSS variable already taken? (do this BEFORE naming one)

```sql
SELECT 'content_components' AS src, count(*) FROM content_components WHERE html_template LIKE '%<--var-name>%' AND COALESCE(is_active,true)
UNION ALL SELECT 'layouts', count(*) FROM layouts WHERE css_template LIKE '%<--var-name>%'
UNION ALL SELECT 'css_snippets', count(*) FROM css_snippets WHERE css_content LIKE '%<--var-name>%'
UNION ALL SELECT 'site_components', count(*) FROM site_components WHERE rendered_html LIKE '%<--var-name>%'
UNION ALL SELECT 'page_components', count(*) FROM page_components WHERE rendered_html LIKE '%<--var-name>%';
```

All five surfaces, because a variable can be consumed by an already-rendered artefact
that no source template mentions. `--color-accent-text` returns 0/0/0/0/0 — it is
derived by the platform and consumed by nothing, which is the dead-config shape.

## Does a layout actually ASK for a palette slot? (the LANDMINE check)

```sql
SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "<slot>"%';
```

A palette reaches the stylesheet **only** through `{{palette "X" "literal"}}` calls in
a layout template. A slot no layout names is never emitted, so adding it to
`darkSchemeDerivations` compiles, logs success, and changes nothing. Measured:
`primary_text` / `cta_text` / `header_text` / `footer_text` are declared by **18 of
18** layouts; `card_bg` 18; `surface_alt` 3; `icon_chip_bg` **0**.

## Prove the render-audit chain is live

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{.items[0].metadata.name}')
for s in write_render_audit_findings scanStoredStatClaims zzzInventedControlXyz; do
  printf "%-40s " "$s"
  kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'"
done
# v1.0.1257: 11 / 2 / 0  -- the 0 proves the grep discriminates; the 2 proves
# I am reading the binary I think I am.
```

Then the config half, which the binary cannot tell you:

```sql
SELECT jsonb_object_keys(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='render-audit-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- site, audit, write_findings, complete, complete_error
```

And the cadence half, which is the one that is missing:

```sql
SELECT name, enabled, target_agent_type FROM scheduled_tasks WHERE enabled;
-- 28 rows, none targeting render-audit-agent
SELECT item_type, status, count(*) FROM site_work_items
WHERE item_type='contrast_failure' GROUP BY 1,2;   -- 4, all complete, all 2026-08-04
```

**GOTCHA — `SELECT ... FROM orchestration_states WHERE owner_agent_type=...`
returning 0 rows does NOT mean "never ran".** Terminal rows are reaped at ~24h. It
means "not in the last day". The cadence question is answered by `scheduled_tasks`,
which has no reaper.

## Schema, before any of the above

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c '\d <table>'
```

**GOTCHA — I skipped this twice in one session and paid for it twice**:
`site_components` has no `is_active` (the lifecycle columns are `locked_at` /
`lock_type` / `build_status`), and `content_components` has no `css_styles`.
`site_specs` has no `resolved_composition` either — it is a single `data` jsonb keyed
by `aspect`. CLAUDE.md says schema first; these are the tables it is worth obeying on,
because each query *reads* perfectly.
