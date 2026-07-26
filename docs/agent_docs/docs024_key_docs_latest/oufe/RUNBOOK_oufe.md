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
