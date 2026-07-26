# RUNBOOK — component CSS coupling (bugs_open/072)

Every command here was needed to get something right, with its gotcha attached.

## Measure the symptom (the only bar that counts)

Over HTTPS, on the rendered page — not the DB row, not the git file.

```bash
for d in ai-agent-orchestration.com relojistas.com gaswholesalers.com robot-hands.com; do
  html=$(curl -s "https://$d/"); css=$(curl -s "https://$d/assets/css/styles.css")
  printf '%-28s uses=%s css_rules=%s inline=%s\n' "$d" \
    "$(printf '%s' "$html" | grep -c 'class="news-card"')" \
    "$(printf '%s' "$css"  | grep -c 'news-card')" \
    "$(printf '%s' "$html" | grep -c 'news-card {')"
done
```

**Gotcha:** the two *styled* sites are the control, and the check is worthless without
them. A run that only looks at the two broken sites cannot tell "added the missing CSS"
from "overwrote three sites' news design".

## Is the stylesheet stale? — the question that cracked it

```bash
for d in ai-agent-orchestration.com relojistas.com gaswholesalers.com robot-hands.com; do
  printf '%-28s css=%s\n' "$d" \
    "$(git -C ~/projects/sites log -1 --format='%ad %h' --date=iso -- "$d/assets/css/styles.css")"
  git -C ~/projects/sites log --reverse --format='  markup first: %ad %h' --date=iso \
      -S'class="news-card"' -- "$d/index.html" | head -1
done
```

**Gotcha 1:** `~/projects/sites` is a *checkout*, and it lags the live sites — relojistas'
homepage had no `news-card` in the 07-25 checkout while the live page emitted six. Treat a
zero here as "not in this checkout", never as "not on the site".

**Gotcha 2:** `page_components.created_at` is **useless** for dating a component's arrival
— the rows are recreated on every re-render, so everything reads as today. Use the sites
repo's git history, which records what actually shipped and when.

## Which snippets does a site match, and did they land?

```sql
WITH f AS (
  SELECT DISTINCT cc.function FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain = 'relojistas.com' AND p.status = 'active'
)
SELECT s.name FROM css_snippets s
WHERE EXISTS (SELECT 1 FROM jsonb_array_elements_text(s.applies_to) e JOIN f ON f.function = e.value);
```

**Gotcha:** `p.status = 'active'` is the only correct filter. The historical
`status IN ('deployed','published','draft','planned')` matches **nothing** — no page has
those values — and that filter is what produced the empty component lists in the first
place (`design_actions.go:340-345`).

Then look for the marker in the deployed CSS: a stylesheet that received snippets has
`/* === Component-specific styles === */`. Its **absence** is the tell, and it is a much
sharper signal than grepping for one class.

## The survey that sized the whole problem

```sql
WITH inuse AS (
  SELECT DISTINCT cc.function, cc.html_template LIKE '%<style%' AS has_style
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE p.status = 'active' AND cc.function IS NOT NULL AND cc.function <> ''
)
SELECT has_style, count(*) FROM inuse GROUP BY has_style;
```

86 true / 8 false on 2026-07-26. This is the query that turned "components have no CSS
mechanism" into "components already ship their own CSS and two stragglers do not" — it
reframed the fix from inventing a convention to finishing one.

## Prove a `<style>` block will actually be stored

`saveSectionsExtractFromHTML`'s regex captures style blocks **only ahead of** script
blocks. Putting the block in the wrong place drops it silently.

```bash
# dump the templates, then run the REAL regex over them
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -t -A -c "SELECT html_template FROM content_components WHERE function='latest-news';" > ln.html
```

then a throwaway Go program using the regex copied verbatim from
`save_page_sections_action.go`. Measured: correct placement stores 4,566 of 4,598 bytes;
appending the block at the end instead stores **1,190** and drops all 3,355 characters of
CSS with no error anywhere.

**Gotcha:** `go run` needs a `go.mod` — `go mod init` in the scratchpad dir first, or it
fails with "go.mod file not found".

## Apply one migration without sweeping everyone else's

```bash
./scripts/migration/run-migrations.sh                       # dry run + probe, always first
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/222_*.sql
./scripts/migration/run-migrations.sh --record-only 222_news_components_carry_their_own_css.sql \
  --note "applied by hand <date>; <what you checked>"
```

**Gotcha:** `--apply` applies **every** pending file. There were 14 pending on 2026-07-26,
13 of them other threads'. Direct `psql` + `--record-only` is the sanctioned way to apply
only yours.

**Gotcha:** re-grep the migration number immediately before writing the file. This one was
written as 217, was 221 by the time it was saved, and shipped as 222 — three collisions in
one afternoon.

## Verify a Go change is actually live

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "data-component-css"'
```

**Gotcha:** grep a string the change **created**. Grepping `news-card` or `css_snippets`
would hit pre-existing code and pass before the roll.
