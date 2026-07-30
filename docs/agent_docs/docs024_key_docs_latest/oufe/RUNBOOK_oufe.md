# RUNBOOK — oufe.com / oxenunity.com

Commands for this workstream, each with the gotcha that made it hard to get right.
**Status markers:** `[PROVEN]` = run here, output seen. `[UNPROVEN]` = written from
the source/another workstream's runbook, not yet executed on these domains.

DB shell used throughout:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```
Templates DB (agent definitions) swaps `clients_user/clients_db` for the templates
pair — check the target before assuming; `agent_definitions` lives in templates_db,
everything site-shaped lives in clients_db.

---

## 1. oxenunity.com — hand-authored single page

The site is one file, authored not generated. It goes to the shared sites repo and
the GitHub Action does the rest.

```bash
cd /home/ant/projects/sites
git add oxenunity.com/index.html
git commit oxenunity.com/index.html -m "oxenunity.com: single-page placeholder"
git push origin master
```

**Gotchas**
- The sites repo default branch is **`master`**, not `main`; the deploy workflow
  only fires on pushes to `master`.
- The Action derives changed domains from `git diff --name-only HEAD~1 HEAD` and a
  top-level `dir.tld/` pattern, syncs `b2://portfolio-sites/<domain>/`, then purges
  Cloudflare **by looking up a zone whose name equals the directory name**. If the
  zone does not exist the lookup yields null and the purge is **silently skipped** —
  a green Action does not mean the site is reachable.
- One shared bucket (`portfolio-sites`), keyed by hostname prefix. There is nothing
  to create on B2 for a new domain.
- This repo is shared with every other session's rerender commits. Commit by
  explicit pathspec, never `git add -A`.

Verify:
```bash
b2 ls b2://portfolio-sites/oxenunity.com/
curl -sI "https://oxenunity.com/?cachebust=$(date +%s)"
curl -s  "https://oxenunity.com/?cachebust=$(date +%s)" | grep -c "OXEN UNITY"
```
An unbusted GET tells you what a cache thinks, not what is deployed.

---

## 2. oufe.com — seed before you submit

`ensure_site_record` upserts on `domain`, so pre-creating the row is safe and it
wins the race against the classifier for the aspects we want to control.

### 2.1 Site row
```sql
INSERT INTO sites (domain, name, network_id, status, email)
SELECT 'oufe.com', 'oufe.com', id, 'active', 'oufe@contactforsales.com'
FROM networks ORDER BY created_at LIMIT 1
ON CONFLICT (domain) DO UPDATE SET email = EXCLUDED.email;
```
**Gotcha** — `status='active'` is what `upsertSite` writes, but it is **not in the
validated vocabulary** (`draft/building/review/published/deployed/archived/error`).
Never scope a query by `sites.status='active'` and expect it to mean anything.
**Gotcha** — set the email. `bugs_open/063`: the hallucinated-email check **fails
open** when a site has no contact email, and a fabricated address reached
production for hours on another site that way.

### 2.2 evidence_base aspect (zero facts, real bans)
`evidence_base` is **not a table** — it is a `site_specs` aspect. There is a partial
unique index on `(site_id, aspect) WHERE is_current`, so **never UPDATE in place**:
supersede then insert, in one transaction.

```sql
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND aspect = 'evidence_base' AND is_current;
INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, pinned)
SELECT id, 'evidence_base', $eb$ { ... } $eb$::jsonb, 'manual', 'oufe-workstream', true, true
FROM sites WHERE domain='oufe.com';
COMMIT;
```
**Gotchas**
- **Dollar-quote the JSON** (`$eb$…$eb$`). It contains single quotes and regex
  backslashes; ordinary quoting mangles both.
- Run with `-v ON_ERROR_STOP=1`. Without it a failed supersede followed by a
  successful insert leaves two `is_current` rows fighting the unique index.
- `jq -e . <file` before sending.
- Carry `pinned` forward on every supersede or the row silently loses
  human-owned status.
- **Read the existing row before superseding.** A blanket supersede on 2026-07-24
  clobbered a structured register unread on another site; the fix was a merge.

### 2.3 imagery_style_guide aspect
Same supersede-then-insert shape, `aspect='imagery_style_guide'`. Without one,
`content_hero` generates unstyled on a fresh site (`bugs_closed/027`).

---

## 3. Tier-3 submission (mission brief + roadmap brief)

`082_submit_domain_unified.sh` accepts `--mission` / `--mission-file` only. It has
**no `--roadmap-file` flag**, although `domain-submitter` persists both `roadmap`
and `roadmap_brief` and `build-site-planner` treats the roadmap brief as
authoritative ("build ONLY the pages listed… do NOT invent additional pages").
So a Tier-3 submission means hand-rolling the envelope from the script's own
`kcat -P` block.

Shape (see the script for the exact headers and pod invocation):
```json
{"action":"orchestrate",
 "config":{"agent_type":"domain-submitter"},
 "input_data":{"domain":"oufe.com","fidelity":"medium",
   "email":"oufe@contactforsales.com",
   "mission_brief":{"text":"…"},
   "roadmap_brief":"…",
   "roadmap":{…}}}
```
**Gotchas**
- Briefs must be **plain prose, single-lined, no nested JSON** — they are read by
  prompts, not parsed.
- **No figures in either brief.** A number in a spec is a given and outranks every
  writer-side rule (see PLAN §5).
- `--fidelity` is recorded only; the plumbing that would act on it is not wired.
- kcat needs `-P -c 1` and a heredoc for a single-line envelope.
- Do not dispatch within ~300s of a chassis pod restart — the spawn is silently
  dropped.

---

## 4. Monitoring the build

```sql
-- orchestration
SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='<CID>'::uuid;

-- the spec cascade filling in
SELECT aspect, source_agent, is_current, created_at
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='oufe.com' ORDER BY created_at;

-- work items
SELECT item_type, status, handler_agent, LEFT(summary,60)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='oufe.com' ORDER BY priority;

-- anything stuck
SELECT item_type, handler_agent, LEFT(error,120)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='oufe.com' AND status IN ('blocked','failed');

-- pages
SELECT name, page_type, build_status, jsonb_array_length(sections) AS n_sections,
       in_header, in_footer
FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com');
```

**Gotchas**
- A `validate_content` blocker reason is **not recoverable from the DB** — the work
  item's `result` is empty. Watch the chassis log live during
  `validate_page_content` or you will not learn why a page died. Five of nine pages
  died this way on the last fresh domain.
- An absent orchestration row usually means **queued**, not dropped. Dispatch
  latency of 16–30 minutes is normal under load. Check whether *other*
  orchestrations started in the meantime before concluding anything.
- New sites are **not enrolled in the discovery/audit sweeps** — they start
  invisible to the immune system. Do not read "no discovery items" as "no problems".

---

## 5. Turning the news feed off (deliberate)

After classification lands, deep-merge `content_features.news_feed.recommended =
false` onto the `classification` aspect (supersede-then-insert as above). The
improvement loop re-runs `enrich_news_feed` every cycle, so check it holds and
re-apply if it flips back — if it does, that is a finding worth recording, not a
retry.

---

## 6. Evidence research (V5)

Spawn `evidence-researcher` with `{site_id, domain, research_query}` on the generic
entry point. Chain: web_search → prepare_urls → batch_webscrape → extract atomic
claims → `verify_and_register_citations`.

**What the verdicts mean** — do not conflate them:
- quote found → citation stands, `accessed` bumped
- HTTP 200 but quote absent → **`citation_lost`**: the claim now rests on nothing
- fetch failed / 403 / 5xx / PDF → **`fetch_error`**: unknown, never treated as
  drift. A paywall going up is not evidence a fact is wrong.
- PDFs are refused rather than half-read. For a PDF-only source the honest route is
  a human-attested fact marked non-reverifiable.

**Gotchas**
- Quote matching is forgiving on presentation (entities, curly quotes, nbsp,
  thousands separators) and strict on content: `411` matches `411&nbsp;`, never
  `412`.
- `ai_service` must be set at **step** level only; a root-level one shadows it.
- V5 has never completed end-to-end since its blocker was fixed. Treat the first
  run as an experiment and record what happens either way.

---

## 7. Claims checks before deploy

```bash
go run ./cmd/claimscan -evidence <eb.json> -components <components.tsv>
# exit 1 = findings, exit 0 = clean
```
**Gotcha** — exporting components for the scan needs
`replace(encode(...,'base64'), E'\n','')`; base64 wraps at 76 characters and the
unwrapped form is what the tool expects.

**Read a clean result correctly.** The deterministic scan does not see currency
amounts or finance vocabulary at all (PLAN §C2c). On this site a clean claimscan
means "no banned pattern matched", not "no invented numbers".

The V3 LLM auditor (`claims-auditor`) is **not on any schedule** — dispatch it by
hand, one call per pass.

---

## 8. Editing live copy, and re-rendering one page

`content_data` edits alone change nothing a visitor sees: the assemble stitches
the **stored `rendered_html`**. To make an authored edit stick you must re-render
the sections from content_data.

```bash
./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_rerender_page.sh <page_name> <domain> [reason]
```
Default reason `section_data_resolved` re-renders every section from stored
content_data through the current template with **no LLM call**.

**Gotchas, all of them paid for on 2026-07-26:**

- **`049b_deploy_single_page.sh`'s `section_data_resolved` branch is broken.** It
  sends `{page_id, site_id, domain}`, but `rerender_page_sections` declares
  `Required: []string{"target_site_id", "page_name"}`
  (`rerender_page_sections_action.go:80`) and nothing derives `page_name` from
  `page_id`. You get:
  `step rerender_sections failed: ... missing required fields: [page_name]`.
  Its assemble-only branch never touches that action, which is why the gap
  survived — it only bites on the branch you need after editing content_data.
  The trigger above supplies `page_name`.
- **`slot_name` must be the component's function name, not `'main'`.** On every
  working page, `page_components.slot_name` equals the entry in `pages.sections`
  (`hero-about`, `about-content`, `generic-text-block`). A hand-inserted row with
  `slot_name='main'` matches no section, renders nothing, and the run reports
  **`COMPLETED | complete_skipped`** — a success-shaped non-event. The adjacent
  known trap is a NULL slot_name; a *wrong* one fails the same silent way.
  Symptom to check first: `page_components.rendered_html` still NULL and
  `rerender_single_page` returning `skipped: true, reason: "no components found
  for page"` (`rerender_single_page_action.go:105-118`).
- **A page at `build_status='planned'` is skipped too.** Set `needs_rebuild`
  before re-rendering a page that was never built.
- **Never re-render chrome with `refresh_site_components` after hand-editing it** —
  that regenerates header/footer from the template and discards the edit. Chrome
  lives in `site_components.rendered_html`; `pages.rendered_header/footer` were
  NULL on this site, so the masters are the only place to patch.
- Check for NULL `content_data` on ANY section first: one NULL escalates the whole
  page to the content writer and **regenerates the copy**, silently discarding
  authored text. The trigger refuses up front on this.

## 8b. Rendering an OWNED page (every tool page is one)

`TRIGGER_rerender_page.sh` defaults to `section_data_resolved`, which runs
`save_page_sections`. That action **hard-refuses an owned page**:

```
step save_sections failed: page tool-recovery-waterfall is rebuild_policy=owned
(tool/widget-owned): a generic section save would clobber it. Use
apply_section_edit for targeted edits or the tool pipeline for rebuilds.
Refusing to overwrite.
```

The guard is right — the action DELETEs and reinserts `page_components`, which is
the TL-001 clobber — and **every tool page is `rebuild_policy='owned'`**, set at
creation. So the path documented in §8 for publishing a copy edit can never
publish a tool.

For a standalone tool the render is the identity function: the component is
`render_mode='template'` with no `{{ }}` placeholders and a NULL `input_schema`,
so the rendered HTML is the template. Populate it directly, then assemble-only:

```sql
UPDATE page_components pc
   SET rendered_html = cc.html_template, build_status='deployed'
  FROM pages p, content_components cc
 WHERE p.id = pc.page_id AND cc.id = pc.component_id
   AND p.site_id = (SELECT id FROM sites WHERE domain='<domain>')
   AND p.name = '<tool-name>'
   AND cc.html_template NOT LIKE '%{{%';   -- refuse if it DOES have placeholders
```
```bash
# assemble-only: no reason argument, so it takes the render_page branch and
# never calls save_page_sections
./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
  <page_id> <site_id> <domain>
```

The `NOT LIKE '%{{%'` guard matters. If the template *does* carry placeholders,
copying it verbatim ships `{{ .field }}` to a reader, so that case needs the tool
pipeline rather than this shortcut.

**Two other things that bit here before this worked:** `slot_name` must equal the
component function name and must also appear in `pages.sections`
(`bugs_open/095` — the prepared insert carried `'main'`), and a section with NULL
`content_data` makes the trigger refuse, so a standalone tool wants `{}`.

## 9. Tool acceptance (Tier 4)

The sweep `tool_acceptance_due` raises an `acceptance_run` item per active tool
with a deployed page and current criteria; `tool-acceptance-agent` drives headless
Chromium and writes an `acceptance-run` doc note.

**Gotchas**
- Criteria live in a fenced ```criteria block inside the tool's travelling PLAN.
- **Do not include an `asset_loads` check** — it asserts a JS-extraction path that
  was designed and never built, and failed every tool on its first sweep.
- Never invent a selector: copy it from the generated HTML.
- Interaction steps are `fill | click | select` only. No navigation, fresh browser
  per URL — a single-page tool is testable, a multi-page journey is not.
- A Tier-2 static pass can confirm a selector exists but can never refute one.
  "The tool works" is a Tier-4 claim only.

## 9. Rendering a component you have LOCKED (the trap that follows from locking)

`save_page_sections` **preserves** locked rows rather than rendering them
(`loadActiveLockedRows`, `save_page_sections_action.go:697`; a row is
agent-writable only if unlocked or on an expired timed lock). So a row inserted
WITH a permanent lock and an empty `rendered_html` renders as **nothing, for
ever**, and the page still reports success. Locking at insert time is the natural
thing to do and it is the wrong order.

Two ways out. Either lock after the first successful render, or write
`rendered_html` by hand in the same statement — the migration-182 pattern. When
writing it by hand, render the component's OWN template rather than
reimplementing it, or the stored HTML drifts from what the renderer would emit:

```bash
# pull the template and the data, then execute them with Go's html/template
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT html_template FROM content_components WHERE name='<component>';" > tmpl.html
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT pc.content_data FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.name='<page>' AND pc.slot_name='<slot>';" > data.json
go run render.go tmpl.html data.json "<component-id>" > out.html   # see NOTES 2026-07-28
```

Then check the render before storing it: `grep -c '{{' out.html` must be **0**.
A template referencing a function that is not in the render funcmap fails to
PARSE rather than degrading — `inc` and `add` are NOT registered (the only
FuncMaps are `render_css_from_spec_action.go:238` and
`compute_component_quality.go:354`), so use a CSS counter for numbering.

**Check the locks on the OTHER sections before any re-render.** On this site the
Thames prose was 7,896 bytes of grounded, audited content sitting unlocked on a
`rebuild_policy='generic'` page, one re-render away from being regenerated.

## 10. `TRIGGER_rerender_page.sh` — the empty-reason trap

`REASON="${3:-section_data_resolved}"` uses `:-`, which treats an **empty string
as unset**. Passing `""` to get assemble-only silently gives you
`section_data_resolved` instead. The run prints the reason it actually used —
read that line, do not assume:

```
corr=... page=thames-water (...) domain=oufe.com reason=section_data_resolved nulls=0
```

## 11. Running claimscan against a live site

The scanner takes the register and a TSV of base64 component HTML. `encode(...)`
wraps base64 at 76 columns and the TSV is line-delimited, so the newlines must be
stripped or every component after the first line is truncated:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT p.name||E'\t'||pc.slot_name||E'\t'||replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','')
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND pc.rendered_html IS NOT NULL;" > components.tsv

kubectl ... -tAc "SELECT data::text FROM site_specs
  WHERE site_id='<site>' AND aspect='evidence_base' AND is_current;" > evidence.json

go run ./cmd/claimscan -evidence evidence.json -components components.tsv
```

It exits non-zero on findings, so `EXIT=$?` after a pipe reports the pipe's last
command and not the scan. Confirm the component you care about is in the list it
prints, because a component with NULL `rendered_html` is silently absent from the
scan rather than reported as unscanned.

## 12. Getting past a 403 on a public-sector or corporate source

Ofwat, Parliament's research briefings and similar sites return **403 to plain HTTP
clients** (curl, WebFetch) while serving the same page normally to a browser. This
is **bot protection keyed on the client, not authentication** — no account, login
or API key is involved, and the documents are public.

The platform already has a browser: `playwright-go` is in `go.mod` and Chromium is
installed for the acceptance runs. A ~40-line program that opens the page and reads
`InnerText("body")` gets **200** where curl gets 403. For a PDF, plain `curl` with a
browser User-Agent is usually enough, because the block is on the HTML routes:

```bash
curl -sL -o doc.pdf --max-time 90 \
  -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36" \
  "<url>"
pdftotext -layout doc.pdf doc.txt     # -layout is load-bearing, see below
```

**`pdftotext -layout` preserves column positions, and that is what makes a chart
readable.** A figure's data labels and its axis labels land on different lines; the
column offset is the only thing tying a value to its category. Without `-layout`
they arrive as an unordered list of numbers.

**Never trust a column mapping on its own — corroborate it.** For Ofwat's Figure
1.1 the mapping was confirmed by arithmetic stated elsewhere in the same document:
588 − 436 = 152, and the prose says "increase average household bills by £152". A
mapping that reproduces an independently stated total is evidence; a mapping that
merely looks plausible is not. Also check each value appears **exactly once** in the
document, or you cannot tell which label it belongs to.

**Beware catastrophic regexes on the extracted text.** `grep -oE "[^.]{0,180}kw[^.]{0,180}\."`
over a 68KB extraction was OOM-killed. Split on `.` in Python and filter instead.

## 13. `context_terms` must match how a value will be LABELLED, not just written

`numberSupported` only lets a fact support a number when one of its `context_terms`
appears in the surrounding text window. Terms written for prose ("average household
bill") do not appear in a chart's terse label ("2024-25, where it started"), so
claimscan correctly reports a **registered** value as unregistered.

Two ways out, and the right one is usually the second:

- widen `context_terms` to include label vocabulary; or
- **make the chart label carry the term** — which is nearly always clearer copy as
  well. "2024-25 baseline, before the review" says more than "where it started" and
  happens to satisfy the guard.

Loosening or removing `context_terms` is the wrong fix: they exist to stop one fact
blanket-supporting every similar number on the page.

## 14. Dispatching a render audit (the whole-site contrast/images/overflow sweep)

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
BODY='{"action":"orchestrate","config":{"agent_type":"render-audit-agent"},"input_data":{"domain":"oufe.com","site_id":"a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39","target_site_id":"a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39"}}'
echo "CORR=${CORR}"
kubectl -n kafka run "kcat-raudit-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '${BODY}' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=${CORR} -H orchestration_id=${ORCH} \
  -H request_id=$(cat /proc/sys/kernel/random/uuid) -H message_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_type=request -H client_id=demo_client -H action=orchestrate \
  -H from_agent_type=cli -H from_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```

No `PUBLISH_OK` → nothing sent (the recorded kcat-stdin landmine — payload MUST be
in the container command). Completes in ~60–90s for 8 pages. Read the verdict:

```sql
SELECT jsonb_pretty(collected_data->'render_audit'->'response'->'summary')
FROM orchestration_states WHERE correlation_id='<CORR>'::uuid;
-- full findings: ->'response'->'contrast' / 'broken_images' / 'overflow'
-- firm failures only: jsonb_path_query_array(..->'response', '$.contrast[*] ? (@.over_image == false)')
```

**Gotchas, each paid for:**
- `pages_failed` counts pages with a **firm** contrast failure, not unreachable
  pages — those land in `unreachable` and are worse.
- It audits rows `build_status='deployed'` ONLY. A live page whose row says
  `needs_rebuild` is invisible to it — contact.html was, for a day, because its
  planned `contact-info` section was never built and the partial-build guard
  (correctly) kept refusing the deploy stamp. Fixed by removing the never-built
  section from `pages.sections` (the component would have fabricated phone/hours
  — `bugs_open/140`).
- **If a dispatch produces no orchestration row**: do NOT re-fire. One query
  first — `SELECT kind, status, convert_from(payload,'UTF8') FROM
  chassis_intake_events WHERE correlation_id='<CORR>';` A `response` row seconds
  after the `request` means the chassis REJECTED it and replied to a topic
  nobody reads; `body.error.message` names the cause. That is how the
  `initial_step`-for-`start_step` seed defect hid for a day (016b §9,
  2026-07-29).

---

## Claims blast-radius scan, fleet-wide (added 2026-07-30)

The measurement to run before changing anything in the claims layer, and the one
that sized `bugs_open/149` C1. It uses the gate's own engine, so it predicts what
the platform will actually do.

**Every step below exists because it went wrong once.** Do not shorten it.

```bash
S=/tmp/scratch; mkdir -p "$S/corpus"
go build -o "$S/claimscan" ./cmd/claimscan

# 1. Site list WITH the expected component count — you need it to detect truncation.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.id, s.domain,
  (SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id=s.id AND pc.rendered_html IS NOT NULL
     AND pc.rendered_html <> '' AND pc.locked_at IS NULL) AS components
FROM sites s ORDER BY 3 DESC;"

# 2. Export per site. TWO traps in this loop:
#    - kubectl exec -i EATS the loop's stdin -> read from fd 3 (or < /dev/null).
#    - the stream TRUNCATES on large exports, leaving a well-formed SHORT file
#      with only 'unexpected EOF' on stderr. Assert the count and retry.
while IFS='|' read -r sid domain want <&3; do
  for attempt in 1 2 3; do
    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
      "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
              replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n', '') ||
              E'\t' || COALESCE(p.page_type,'')
       FROM page_components pc JOIN pages p ON p.id = pc.page_id
       WHERE p.site_id = '$sid' AND pc.rendered_html IS NOT NULL
         AND pc.rendered_html <> '' AND pc.locked_at IS NULL" \
      < /dev/null > "$S/corpus/$domain.tsv" 2>/dev/null
    [ "$(wc -l < "$S/corpus/$domain.tsv")" = "$want" ] && break
    echo "SHORT $domain — retrying"
  done
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
    "SELECT data FROM site_specs WHERE site_id='$sid' AND aspect='evidence_base' AND is_current" \
    < /dev/null > "$S/corpus/$domain.evidence.json"
done 3< sites.txt

# 3. Scan each site against ITS OWN register. Omit -evidence for an unarmed site:
#    that scans exactly as the platform scans it (fleet-wide patterns only).
#    The 4th TSV column is page_type — WITHOUT it every page reads UNKNOWN and you
#    get the editorial false positives the platform itself no longer raises.
"$S/claimscan" -evidence "$S/corpus/$d.evidence.json" -components "$S/corpus/$d.tsv"

# 4. Read the results with `command grep -a`. NOT bare grep: in the Claude Code
#    shell grep is a ugrep wrapper with -I, and one non-UTF-8 byte in site copy
#    makes it return zero matches AND PRINT NOTHING — not even 0. LC_ALL=C does
#    not help. A blank where a count should be is the tell.
command grep -ac "^BANNED" "$S/scan_all.txt"   # blockers — these REFUSE a save
command grep -ac "^NUMBER" "$S/scan_all.txt"   # errors  — recorded, allowed
```

**Reading it:** `BANNED` lines are what the build gate AND (since `CLM-018`) the
persistence floor will refuse. `NUMBER` lines only ever record. `banned_claim` is the
JSON value and appears nowhere in the output — grepping for it returns 0 on every
site, a false all-clear.

**Result 2026-07-30, for comparison only — re-measure, never quote:** 949 components
/ 14 sites → 3 BANNED (webdesign.co.uk `tool-blueprint-compiler`; robot-hands.com
`how-it-works` and `gripper-catalog`), 59 NUMBER on the 4 armed sites. The surface was
908 on 07-28 and 919 on 07-29 — it moves every day.

## Verifying a chassis deploy carries your change (added 2026-07-30)

`bugs_open/153`: an `IMAGE_TAG` bump does not imply a rebuild, and `verify-agent-images`
prints all-green on a retag. So prove it in the binary, with controls:

```bash
docker run --rm --entrypoint sh docker.io/aqls/agent-chassis:<tag> -c '
  strings /app/agent-chassis | grep -c "<a string YOUR change added>"      # expect >0
  strings /app/agent-chassis | grep -c "CONTENT_LINK_REPAIR_DETAIL"        # positive control, expect >0
  strings /app/agent-chassis | grep -c "<a string that exists nowhere>"'   # negative control, expect 0
```

**Pick an ASCII-ONLY marker.** `strings` splits a Go literal at every non-ASCII byte,
so any marker containing an em dash — which house style makes likely in error messages
— greps to **0** in an image that demonstrably contains it. That reads exactly like a
failed build. Measured on `v1.0.1208`: em-dash form 0, ASCII fragment of the same
literal 1.
