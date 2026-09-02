# RUNBOOK — gamedesign.uk rebuild

Commands this lane had to get right, each with its gotcha. `[PROVEN]` = run here, output seen.
`[UNPROVEN]` = written from the worked example, not yet executed on this domain.

DB shell throughout:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```
⚠ `kubectl exec -i` inside a `while read` / `for` loop **eats the loop's stdin**. Every call in a
loop carries `</dev/null`. One of this lane's own loops silently processed fewer rows before this
was added.

---

## 1. Probe the served site, with the control that makes a 200 mean something `[PROVEN]`

```bash
# the invented URL MUST be non-200, or the domain is a catch-all and every 200 is void
curl -s -o /dev/null -w '%{http_code}\n' https://gamedesign.uk/this-path-does-not-exist-9z8x7.html   # 404
for p in / /about.html /tools.html; do printf "%-16s %s\n" "$p" "$(curl -s -o /dev/null -w '%{http_code} %{size_download}b' https://gamedesign.uk$p)"; done
```

## 2. Measure body content — and STATE THE METRIC `[PROVEN, then corrected]`

```python
# maintext.py — text inside <main>, with <script> AND <style> stripped, whitespace collapsed
import sys,re
h=sys.stdin.read()
m=re.search(r'<main.*?</main>',h,re.S)
if not m: print('NOMAIN'); sys.exit()
t=re.sub(r'<(script|style).*?</\1>','',m.group(0),flags=re.S)
t=re.sub(r'<[^>]+>',' ',t); print(len(re.sub(r'\s+',' ',t).strip()))
```
**Gotcha (WRONG_CALLS 2026-09-02):** the first version stripped tags but not `<style>` blocks
sitting INSIDE `<main>`, so "5,977 chars" was ~60% CSS. Same page by this metric: 2,473. The
direction (→0) never depended on it; the comparability did.

**Gotcha:** `curl … | python3 maintext.py` on a URL that 404s measures the 404 page. Check the
code first (§1) or you will record a "content" figure for a page that does not exist.

## 3. Bisect when a page emptied — the sites repo IS the serving surface `[PROVEN]`

```bash
cd ~/projects/sites
git log --format='%h %ad %s' --date=short -- gamedesign.uk/index.html | head
for c in $(git log --format=%h -15 -- gamedesign.uk/index.html); do
  printf "%s %s main=%s\n" "$c" "$(git log -1 --format=%ad --date=short $c)" \
    "$(git show $c:gamedesign.uk/index.html 2>/dev/null | python3 maintext.py)"
done
```
**Gotcha (the wrong call this lane made):** if you compute a "before" sha with
`git log -1 --before=<date> -- $f` and the file had **no commit before that date, `$c` is
EMPTY**, `git show :$f` errors to stderr, python reads empty stdin and prints **0** — an ABSENT
file measured as an EMPTY file. Guard it: `[ -n "$c" ] || { echo ABSENT; continue; }`.

## 4. The discriminating control — was it this site, or the fleet? `[PROVEN]`

For every site directory with commits on the day, for every `.html` touched, compare `<main>`
text at the parent vs the commit. The disconfirming result is any content→0 on ANOTHER site.
```bash
cd ~/projects/sites
git log --since=2026-04-16 --until=2026-04-17 --name-only --format='' \
  | grep -oE '^[a-z0-9.-]+\.(uk|com|co\.uk)/' | sort | uniq -c | sort -rn
# then per site: parent-vs-commit <main> length for each touched .html (see scratchpad/fable/daycontrol.py for the full version)
```
Result 2026-04-16: gamedesign.uk 4/11 emptied; six other sites 0/139. On 04-18: gamedesign.uk 3
emptied + 4 born-empty; four other sites 0 emptied but **13 born-empty** (fleet-wide that day).

## 5. Which commit is the chassis actually running? `[PROVEN]` — the CLAUDE.md log grep is OUT OF RANGE

```bash
# CLAUDE.md's recipe returns NOTHING on agent-chassis even at --tail=20000 (startup line scrolled).
# Use the capabilities table instead:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F' | ' -c "
SELECT service, left(git_commit,12), count(DISTINCT pod_name), max(last_seen_at)::timestamp(0)
FROM service_binary_capabilities WHERE service ILIKE '%chassis%' AND last_seen_at > now()-interval '2 hours'
GROUP BY 1,2 ORDER BY 4 DESC;" </dev/null
# then, in the agentchassis repo, with BOTH controls:
git merge-base --is-ancestor d777cb4d2 <stamp> && echo guard-1-live
git merge-base --is-ancestor 6579e9ae1 <stamp> && echo guard-2-live
git merge-base --is-ancestor HEAD <stamp>    # expect FAIL (HEAD is ahead of any build)
git merge-base --is-ancestor <stamp> HEAD    # expect PASS
```
**Gotcha:** the table's key column is `service`, not `service_name`. During a roll two stamps
coexist (2026-09-02 16:19: 231 pods on `ebf27c60`, 346 still on `a2732c72`) — a per-fleet answer
is two answers; read the pod you mean.

## 6. Pre-seed the site row BEFORE submitting `[UNPROVEN on this domain — shape from oufe]`

Per `oufe/SEED_2026-07-25_oufe_site_and_specs.sql`. Three reasons it comes first: the
hallucinated-email check FAILS OPEN with no email (`bugs_open/063`); the claims layer no-ops until
`evidence_base` exists; `content_hero` renders unstyled with no `imagery_style_guide`
(`bugs_closed/027`).
```sql
\set ON_ERROR_STOP on
BEGIN;
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES ('gamedesign.uk', 'gamedesign.uk', '00000000-0000-0000-0000-000000000002',
        'active', '<OWNER-SUPPLIED EMAIL>', 'gamedesign.uk')
ON CONFLICT (domain) DO UPDATE SET email = COALESCE(sites.email, EXCLUDED.email);
-- evidence_base + imagery_style_guide: copy the oufe file's two blocks, supersede-then-insert,
-- never UPDATE in place (partial unique index on (site_id, aspect) WHERE is_current).
COMMIT;
```
**Gotcha (positioning lane, 2026-09-02):** `ensure_site_record` scans `name`+`network_id`
**without `COALESCE`** — a row with either NULL stalls the build at `needs_site_plan` with a Scan
error. Set both explicitly. Broke a sibling-flow release the same day.
**Gotcha:** `status='active'` is what `upsertSite` writes and is NOT in the validated vocabulary
(`draft/building/review/published/deployed/archived/error`). Never scope a query by it.

## 7. Dispatch the FRESH build `[UNPROVEN — not yet run; owner review of the brief first]`

```bash
cd /home/ant/projects/agentchassis
./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh gamedesign.uk \
  --email '<OWNER-SUPPLIED EMAIL>' \
  --mission-file docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/MISSION_2026-09-02_gamedesign_uk.txt
```
No `--from`: adopting from gamedesign.uk ingests the empty shells; adopting from the sibling
reproduces the sibling, which D2 forbids. The brief contains no `"` or `\` so 082's
`sed|tr|sed` fold is lossless (checked by running its exact pipeline over the file).
**Gotcha:** no orchestration dispatch within ~300 s of a chassis pod (re)start — the spawn is
silently dropped. The fleet was mid-roll at 16:19 on 2026-09-02; check §5 first.
**Gotcha (bugs_open/438, measured 2026-09-02):** a COMPLETED submitter does NOT mean the brief
landed. `persist_mission` fails on every fresh submit (082 sends `mission_brief`, not
`mission`), and the step that does carry it, `persist_mission_brief`, fails on 3 of 12 sites.
Verify before trusting the classifier's output:
`SELECT length(data->>'text') FROM site_specs WHERE aspect='mission_brief' AND is_current AND site_id='<id>';`
— expect the brief's length (2,892 here). Three `agent_error_log` rows from the submitter
(`persist_mission`/`persist_roadmap`/`persist_roadmap_brief`, "missing required fields:
[spec_data]") are the NORMAL fingerprint of a fresh submit, not a problem with your build.
**Gotcha:** the classifier queues behind the fleet. Find the run by payload, not by the printed
id: `SELECT current_step,status FROM orchestration_states WHERE collected_data->'input_data'->>'domain'='gamedesign.uk' ORDER BY created_at DESC LIMIT 3;`

## 7b. BEFORE retracting a tree: census inbound links from our other sites `[PROVEN — late]`

```sql
SELECT s.domain, p.name, (regexp_matches(pc.rendered_html, 'https?://<domain>[^"'' <)]*', 'g'))[1]
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html LIKE '%<domain>/%' AND s.domain <> '<domain>';
```
**Gotcha (2026-09-02):** I cleared gamedesign.uk at 17:05 and found at 18:45, from the sibling's
message, that ONE of their pages deep-linked into the tree I removed. Adopted sites inherit
absolute links to their source. Run this first and tell the owning lane BEFORE the push.

## 8. Verify at the ARTEFACT, never at a status `[PROVEN 2026-09-02 ~18:00Z]`

After the cascade reports done:
```bash
cd ~/projects/sites && git pull --ff-only
for f in $(find gamedesign.uk -name '*.html'); do printf "%-60s %s\n" "$f" "$(python3 maintext.py < $f)"; done | awk '$2==0 || $2=="NOMAIN"'
# expect NO lines. Then the served copy, with the control:
curl -s -o /dev/null -w '%{http_code}\n' https://gamedesign.uk/no-such-page-4k2.html   # 404
for p in / /privacy.html /terms.html /sitemap.xml; do printf "%-16s %s\n" "$p" "$(curl -s -o /dev/null -w '%{http_code}' https://gamedesign.uk$p)"; done
curl -s https://gamedesign.uk/ | grep -c 'href="mailto:"'    # expect 0
```
**Gotcha:** an unbusted GET tells you what Cloudflare's cache thinks. Append `?cachebust=$(date +%s)`
or read `last-modified`. And a `complete` work item is not a rendered page — `bugs_open/432` §3a
shows eight `complete` rerenders on a page that is still empty.
