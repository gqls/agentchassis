# RUNBOOK — bugs_open/390

Every command here had a gotcha attached. When one changes, change it HERE.

## 1. Census the population — MUST union the archive

`site_work_items` is a rolling window; closing a row archives it out of the table you queried.
Any "how many ever" question that reads only the live table understates and reads as authoritative.

```sql
WITH allrows AS (
  SELECT site_id, spec, status, created_at FROM site_work_items        WHERE item_type='contrast_failure'
  UNION ALL
  SELECT site_id, spec, status, created_at FROM site_work_items_archive WHERE item_type='contrast_failure'
) SELECT status, count(*) FROM allrows GROUP BY 1 ORDER BY 2 DESC;
```

## 2. The damage figure: repaired, then filed again with the SAME colours

⚠ Exclude the arm-1 invented selectors (`TAG.TAG`, where the class equals the tag) or you count
`bugs_closed/352`'s damage as this bug's.

```sql
WITH allrows AS (
  SELECT site_id, spec, status, created_at FROM site_work_items        WHERE item_type='contrast_failure'
  UNION ALL
  SELECT site_id, spec, status, created_at FROM site_work_items_archive WHERE item_type='contrast_failure'
), k AS (
  SELECT s.domain d, a.spec->>'selector' sel, a.spec->>'page_name' pg,
         a.spec->>'fg' fg, a.spec->>'bg' bg, a.status, a.created_at,
         (split_part(a.spec->>'selector','.',1) = split_part(a.spec->>'selector','.',2)) invented
  FROM allrows a JOIN sites s ON s.id=a.site_id
)
SELECT count(*) AS identical_refilings_after_complete
FROM k later JOIN k earlier
  ON later.d=earlier.d AND later.sel=earlier.sel AND later.pg=earlier.pg
 AND later.fg=earlier.fg AND later.bg=earlier.bg AND later.created_at > earlier.created_at
WHERE earlier.status='complete' AND NOT later.invented;
```
**97 as of 2026-08-25.** Date every re-run: this grows by addition and reads as current for ever.

## 3. Which sites can css-patch-agent even reach

The agent resolves its target as `sites → style_collections → css_themes`. A site with
`style_collection_id IS NULL` parks at `css_no_theme_198` and never reaches the LLM.

```sql
SELECT s.domain, s.style_collection_id IS NOT NULL AS linked, ct.origin, ct.version,
       length(ct.css_content) AS css_len,
       (SELECT count(*) FROM sites s2 JOIN style_collections sc2 ON sc2.id=s2.style_collection_id
         WHERE sc2.css_theme_id=ct.id) AS site_count
FROM sites s
LEFT JOIN style_collections sc ON sc.id = s.style_collection_id
LEFT JOIN css_themes ct        ON ct.id = sc.css_theme_id
WHERE s.domain = '<domain>';
```
⚠ **Read migration 542's gate against these two numbers before predicting an outcome**: the agent
refuses unless `css_len >= 4096 AND site_count <= 1`. A site on a shared seed theme parks at
`css_base_integrity_guard_198` however reachable its stylesheet looks from the browser. Getting
this wrong is exactly the misstep in NOTES §2026-08-25(c).

## 4. Have the appended repairs survived?

```sql
SELECT s.domain,
       (length(ct.css_content) - length(replace(ct.css_content,'css-patch-agent','')))/15 AS markers_left
FROM sites s JOIN style_collections sc ON sc.id=s.style_collection_id
JOIN css_themes ct ON ct.id=sc.css_theme_id ORDER BY 2;
```
⚠ **Zero markers on a site with completed repairs is not "it never patched"** — it is usually
migration 543's `persist_css_to_theme` having overwritten the row at the site's next design run.
Check `css_themes.updated_at` against the repair's `updated_at`.

## 5. At the artefact — always with a control in the same breath

A parked domain 200s every path, so an invented URL that 404s is what makes a 200 mean anything.

```bash
curl -sS -L "https://$D/index.html"            -o p.html -w "page HTTP %{http_code}\n"
curl -sS -o /dev/null -w "control HTTP %{http_code}\n" -L "https://$D/index-390-control.html"   # must be 404
curl -sS -L "https://$D/assets/css/styles.css" -o t.css  -w "css  HTTP %{http_code}\n"
```
⚠ The stylesheet link can 404 while the page 200s (cv1.co.uk does). Check the CSS fetch's own
status; a 404 body is ~2.7 KB of HTML and greps as "no rules found", which reads like a clean file.

Find the declaration that actually wins, and where it lives:
```bash
python3 - <<'PY'
import re
h=open('p.html',encoding='utf-8',errors='replace').read()
c=open('t.css', encoding='utf-8',errors='replace').read()
TOKEN='<the class from the filed selector>'
def hits(text):
    for r in re.finditer(r'([^{}@]+)\{([^}]*)\}',text):
        sel=re.sub(r'/\*.*?\*/','',r.group(1),flags=re.S).strip()
        if TOKEN in sel and re.search(r'(^|;|\s)color\s*:',r.group(2)):
            yield sel.replace('\n',' ')
for m in re.finditer(r'<style[^>]*>(.*?)</style>',h,re.S):
    for s in hits(m.group(1)): print('PAGE  @%d  %s' % (m.start(), s))
for s in hits(c): print('THEME       %s' % s)
print('stylesheet link at byte', [m.start() for m in re.finditer(r'/assets/css/styles\.css',h)][:1])
PY
```
⚠ This is a **static approximation** — it cannot see media queries, cascade layers or
`getComputedStyle`. It is good enough to classify a population and NOT good enough to be the
contract; that is why the fix measures in the browser.

## 6. Did the repair actually take? — the only observation that settles it

```sql
SELECT status, result->>'resolved_by', result->>'resolved_at'
FROM site_work_items WHERE id = '<item id>';
```
A **retraction** (`resolved_by='render_audit'`) is the platform's own recorded definition of
"repaired". A re-filing with byte-identical `fg`/`bg` is the disconfirming result.

⚠ **The re-audit window is 3 days, not 7** — read it live, never from WII-016's status line,
which is stale on exactly this constant:
```sql
SELECT pre_query FROM scheduled_tasks WHERE name='site-render-audit-rotation';
```

## 7. Applying the migrations

Dry-run per session and after every roll; `--apply` takes EVERY pending file, so scope the dir.
The council gate covers appliable migrations — submit before or alongside the commit.
