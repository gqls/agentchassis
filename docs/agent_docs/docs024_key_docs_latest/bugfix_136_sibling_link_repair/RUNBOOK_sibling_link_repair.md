# RUNBOOK — `bugs_open/136` (section-editor slug), sibling link repair

Every command here had a gotcha. The gotcha is the point — change a command HERE when it
changes, not in your scrollback.

---

## 1. Who else is working this bug (run BEFORE anything else)

`scripts/who-owns.py <number>` reads **commits**, so a session mid-fix is invisible to it,
and on an ambiguous number it answers about the *other* bug. Both happened here. Use it,
then do the transcript sweep, which is what actually finds a live owner:

```bash
scripts/who-owns.py section_editor_and_three_siblings   # by SLUG — 136 is ambiguous

cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name '*.jsonl' -mmin -600 | grep -v <your-own-session-id>); do
  n=$(grep -c 'ApplySectionEditAction\|repairSectionLinks\|bugs_open/136' "$f" 2>/dev/null)
  [ "$n" -gt 0 ] && echo "$(basename $f .jsonl|cut -c1-8) $n $(date -r $f '+%H:%M')"
done
```

**Gotchas.**
- Grep for **code symbols**, not just the bug number: the session fixing `bugs_open/155`
  showed 119 hits on the number and 113 on `resolveStorageURIFromAsset`, and either alone
  would have found it — but a session that has not yet typed the number is only findable by
  the symbol.
- **≤8 hits is noise.** Every session in this fleet has the bug list and `CLAUDE.md` in
  context, so single-digit counts are incidental. The owner is the outlier (100+).
- `tail -c` on a `.jsonl` can abort with a uutils `tail` panic on some files; use `grep -c`
  over the whole file instead.

## 2. The census — unresolved internal links in stored `page_components.rendered_html`

**This is the query whose result I got wrong by reading a listing. Use the aggregate.**
It mirrors `datahelpers.NormalizePagePath` in SQL (lowercase, strip the `#`/`?` tail, strip
`index.html`, trim the trailing `/`) and `ClassifyLinkScope`'s exclusions.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<'SQL'
WITH pg AS (
  SELECT site_id,
         CASE WHEN rtrim(regexp_replace(lower(btrim(url)), 'index\.html$',''),'/') = ''
              THEN '/' ELSE rtrim(regexp_replace(lower(btrim(url)), 'index\.html$',''),'/') END AS npath
  FROM pages
  WHERE status NOT IN ('deleted','archived') AND url IS NOT NULL AND url <> ''
),
hr AS (
  SELECT pc.id AS pc_id, p.site_id, s.domain, p.name AS page_name,
         COALESCE(p.page_type,'') AS page_type, COALESCE(pc.slot_name,'') AS slot_name,
         m[1] AS href
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id,
  LATERAL regexp_matches(COALESCE(pc.rendered_html,''), 'href\s*=\s*"([^"]*)"', 'g') m
),
i AS (
  SELECT * FROM hr
  WHERE href <> '' AND href NOT LIKE '#%' AND href NOT LIKE 'http%' AND href NOT LIKE '//%'
    AND href NOT LIKE 'mailto:%' AND href NOT LIKE 'tel:%' AND href NOT LIKE 'javascript:%'
    AND href !~* '\.(css|js|png|jpe?g|gif|svg|webp|ico|pdf|woff2?|ttf|eot|mp4|zip|xml|txt|json)(\?|#|$)'
),
nm AS (
  SELECT i.*,
    CASE WHEN rtrim(regexp_replace(lower(btrim(split_part(split_part(href,'#',1),'?',1))),'index\.html$',''),'/') = ''
         THEN '/' ELSE rtrim(regexp_replace(lower(btrim(split_part(split_part(href,'#',1),'?',1))),'index\.html$',''),'/') END AS npath
  FROM i
)
SELECT COALESCE(CASE WHEN EXISTS (SELECT 1 FROM pg WHERE pg.site_id=nm.site_id AND pg.npath = nm.npath || '.html')
            THEN 'rewrite' ELSE 'UNLINK' END, 'TOTAL') AS repair_action,
       count(*) AS href_occurrences,
       count(DISTINCT nm.pc_id) AS components,
       count(DISTINCT nm.domain) AS sites
FROM nm
WHERE NOT EXISTS (SELECT 1 FROM pg WHERE pg.site_id = nm.site_id AND pg.npath = nm.npath)
GROUP BY ROLLUP(1) ORDER BY 1;
SQL
```

Measured **2026-08-02**: `rewrite` 18 / `UNLINK` 17 / **TOTAL 35** href occurrences, 13
components, **6** sites. Tool-shaped (`slot_name LIKE 'tool%' OR page_type='tool'`): 7.

**Gotchas.**
- **`(N rows)` at the foot of a detail listing is not a count of anything real.** Mine was
  a `GROUP BY (domain, page, slot, href, action)`, so it collapsed repeats: the listing said
  30, the aggregate said 35. Ask for `count(*)`; name the unit ("href occurrences").
- The `rewrite`/`UNLINK` split is **only** correct if the two `EXISTS` clauses match
  `RepairPageLinks`' arms: it rewrites when `npath || '.html'` exists and unlinks otherwise.
  Change one arm in Go and this query silently reports the wrong split.
- `regexp_matches(...,'g')` in a `LATERAL` **drops rows with no match**, which is what you
  want here (a component with no anchors is not a finding) but is a silent inner join.
- It reads `rendered_html` only. `content_data` can hold the same href, and a rerender from
  `content_data` reintroduces it — see the `rerender_link_repair.go` landmine.

## 3. Is the repair path even exercised? (the honest-exposure query)

```sql
SELECT owner_agent_type, count(*) FROM orchestration_states
WHERE owner_agent_type IN ('section-editor','tool-improver') GROUP BY 1;   -- 0 rows, 2026-08-02
SELECT count(*) FROM pages WHERE page_type='report';                        -- 0, 2026-08-02
SELECT count(*) AS retained, min(created_at)::date FROM orchestration_states; -- 2469, 2026-07-13
```

**Gotcha:** `orchestration_states` has **no `agent_type` column** — it is `owner_agent_type`.
And it is retention-clocked (~20 days here), so "zero runs" means *zero in the window*, never
"never ran". Record the window with the number or the claim rots.

## 4. Does the blog listing need the repair? (the query that said no)

```sql
SELECT function, name,
       (SELECT count(*) FROM regexp_matches(html_template,'<a[^>]+href','gi')) AS anchors,
       (SELECT string_agg(m[1],' | ') FROM regexp_matches(html_template,'href="([^"]*)"','g') m) AS hrefs
FROM content_components WHERE (function='content-listing' OR name='article_grid') AND is_active;
-- content-listing | content-listing | 1 | {{.url}}
```

**Gotcha:** the table is **`content_components`**, not `components` (which does not exist).
`RebuildBlogListingAction` loads its template from there with `html_template LIKE '%range%'`,
so a template without a `range` is not the one it renders — check the same predicate.

## 5. Verifying the Go change

```bash
go build ./... && go test ./platform/orchestration/actions/ -count=1
gofmt -l platform/orchestration/actions/          # pre-existing offenders live here; check YOUR files are absent

# HEAD, not this tree — a green local build can be green on someone else's uncommitted work
T=$(mktemp -d); git archive HEAD | tar -x -C $T && (cd $T && go build ./...)
```

**The mutation check, because a green run proves nothing on its own:**

```bash
# A: make the repair a no-op  -> 3 tests must fail
#    insert `if true { return html }` at the top of repairComponentHTMLBeforePersist
# B: delete the fail-open     -> the skip test must fail
#    change `if !indexOK {` to `if false {`
go test ./platform/orchestration/actions/ -run TestRepairComponentHTML -count=1
```

## 6. The pattern-check rule, and its four controls

`check_unrepaired_component_write` in `scripts/pattern-check.py`. **Do not trust it until you
have run all four controls** — the `strip_comments` landmine (LANDMINES.md, 2026-07-31) is
live on exactly this file:

```bash
python3 - <<'PY'
import importlib.util
spec = importlib.util.spec_from_file_location("pc", "scripts/pattern-check.py")
pc = importlib.util.module_from_spec(spec); spec.loader.exec_module(pc)
for name, files in {
  "(a) unguarded writer must FIRE": ["platform/orchestration/actions/deploy_tool_action.go"],
  "(b) guarded must be QUIET":      ["platform/orchestration/actions/section_editor_actions.go"],
  "(d) allow-listed must be QUIET": ["platform/orchestration/actions/adopt_verbatim.go"],
}.items():
    f=[]; pc.check_unrepaired_component_write(files, None, f); print(name, len(f))
PY
```

Control **(c)** is the one that catches a wrong stripper: copy an unguarded writer, add a
COMMENT naming `repairComponentHTMLBeforePersist`, and confirm it STILL fires. It does,
because the guard is matched as a **call on a non-comment line of the RAW body** — searching
the stripped body can delete a guard and invent a finding; searching the raw body lets a
comment silence a real one.

## 7. Pod-grep after the next chassis roll (a roll is NOT evidence — `bugs_open/153`)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ NEW code only — 0 before, 1+ after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'repaired dead internal links before persisting a single component'"
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'Component link repair SKIPPED'"
# ✓ NEGATIVE control — a string this change REMOVED; must go 1 -> 0
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'failed to update page_component for swap'"
# ✓ POSITIVE control — 079's marker, live since v1.0.1170; must stay 1
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'SavePageSectionsAction: repaired dead internal links before persist'"
```

**Gotchas.** Run it on **every replica**, not `deploy/agent-chassis` (that reads one pod of
N). The obvious marker "repaired dead internal links" is **vacuous** — 079's live string
contains it, so it greps 1 before anything ships; the discriminating phrase is
"…before persisting a single component". And `grep -c` is case-sensitive: a mis-cased
pattern reads exactly like "not shipped".
