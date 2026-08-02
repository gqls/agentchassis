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

---

## §8 — measuring a markup-writer's blast radius by RUNNING it, not by approximating it in SQL

Added 2026-08-03 for `bugs_open/180`. The rule (`bugs_open/093`) is that the census must be
the shipping function over real bytes; SQL can only nominate the population.

**8a. Dump the corpus — PER SITE, and check the ROW COUNT, not the byte count.**

```bash
# ⚠ A single COPY of the whole fleet TRUNCATED at ~2.8MB and exited 1 with
#   "Waiting for server to close stdin failed" — a partial dump. Chunk by site.
# ⚠ Use `-f -` with the SQL on stdin. `exec -i … -c "…"` leaves stdin open and is
#   what provokes the EOF error above.
> assembled.csv
for site in $(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c "SELECT domain FROM sites ORDER BY domain"); do
  cat > qs.sql <<SQL
COPY (
  SELECT jsonb_build_object('cid',p.id::text,'site',s.domain,'page',p.url,
           'html', string_agg(pc.rendered_html, E'\n' ORDER BY pc.position))::text
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE pc.rendered_html IS NOT NULL AND s.domain = '$site'
  GROUP BY p.id, s.domain, p.url
) TO STDOUT (FORMAT csv, QUOTE '"', FORCE_QUOTE *);
SQL
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -f - < qs.sql >> assembled.csv
done
wc -l assembled.csv    # must equal SELECT count(DISTINCT page_id) FROM page_components …
```

**Assemble by `position`, not per component.** `repairOutboundPageLinks` (LNK-023) is handed
the ASSEMBLED page, so a per-component census measures a different input — and for a
span-based guard the difference is real: an unclosed `<script>` in one component reaches
across every component after it.

**8b. Run the SHIPPING function and the FIXED one over it, in the package**, as a temporary
`probe_*_test.go` (delete before commit — the permanent tests are the ones that stay). The
three numbers worth printing, and the third is the one that matters:

| number | what it is | why |
|---|---|---|
| matches inside a non-markup span | the risk surface | fleet-wide, all spellings |
| components/pages the SHIPPED writer mutates | live damage today | the bug's real size |
| **legit matches LOST by the fix** | **the cost of the guard** | a guard that over-fires stops repairing real defects, silently — this must be **0** |

**8c. Nominate the awkward population in SQL, then run the function over it.** For a
span-based guard the awkward case is markup the scanner cannot finish reading:

```sql
SELECT count(*) FILTER (WHERE (length(lower(rendered_html))-length(replace(lower(rendered_html),'<script','')))/7
                          <> (length(lower(rendered_html))-length(replace(lower(rendered_html),'</script','')))/8) AS unclosed_script,
       count(*) FILTER (WHERE (length(rendered_html)-length(replace(rendered_html,'<!--','')))/4
                          <> (length(rendered_html)-length(replace(rendered_html,'-->','')))/3)                   AS unterminated_comment,
       count(*) AS total
FROM page_components WHERE rendered_html IS NOT NULL;   -- 2026-08-02: 9 | 0 | 1186
```

This is an APPROXIMATION (a `</script` inside a JS string counts), which is exactly why its
only job is to hand a row set to 8b.

## §9 — proving a guard by breaking it, when guards sit in SERIES

```bash
cp platform/orchestration/datahelpers/markup_spans.go /tmp/ms.bak
sed -i "s/const maskFiller = 0x00/const maskFiller = ' '/" platform/orchestration/datahelpers/markup_spans.go
go test ./platform/orchestration/datahelpers/ -count=1 2>&1 | grep -E "^--- FAIL|^ok"
cp /tmp/ms.bak platform/orchestration/datahelpers/markup_spans.go
```

**Gotcha, and it cost a wrong comment in this lane (see `WRONG_CALLS.md` 2026-08-02):** a
mutation that PASSES does not mean the code is fine or the test is weak — it usually means a
second guard sits underneath. Before rewriting the test, ask *what else guarantees the
property*, then build an input where **only the mutated guard could act**. And have the test
assert its own PREMISE (here: that `NonMarkupSpans` actually found the element), or it can
stop discriminating for a third reason and still pass.
