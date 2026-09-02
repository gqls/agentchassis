# RUNBOOK — mortgagecalculator.co.uk adoption

Commands that were hard to get right, with the gotcha attached. Add to this file,
not to your scrollback.

## §1 B2 access

**The handoff says there are no B2 credentials. That is now WRONG** — as of
2026-07-31 ~21:56 local, `~/.config/b2/account_info` exists and authorises.

```bash
b2 account get          # confirms auth without mutating anything
b2 version              # 4.7.0 (b2sdk 2.12.0)  -- v4 CLI, see the flag gotcha below
```

**GOTCHA — the dry-run flag is `--dry-run`, not `--dryRun`.** The camelCase spelling
is b2 CLI v3. On v4 it exits **2** with a usage dump. That matters because a piped
`grep` over the failed output prints "(none)" for deletions and uploads, which reads
exactly like a clean no-op. **Print `${PIPESTATUS[0]}` and check it is 0.**

```bash
b2 ls --recursive b2://portfolio-sites/mortgagecalculator.co.uk/
```

## §2 Getting the true origin bytes — do NOT use curl

**GOTCHA — `curl https://mortgagecalculator.co.uk/robots.txt` is NOT the origin
file.** It returns **2,327 bytes**; the origin is **491**. Cloudflare injects a
`# BEGIN Cloudflare Managed content` block at the edge. The `x-amz-*` response
headers are still present, so the usual "it came from B2" tell does **not** catch
this. Committing the fetched copy would bake the injection into the origin and have
Cloudflare inject it a second time.

The handoff missed this because it ran `tail -5`, which lands past the injected
block. **When the bytes matter, take them from the bucket:**

```bash
b2 sync b2://portfolio-sites/mortgagecalculator.co.uk/ ./bucket   # no --delete = download only
```

Everything else survived the CDN untouched: 28 of 28 non-robots files were
sha256-identical live vs local, which is itself the evidence that Cloudflare
rewrites `robots.txt` specifically and not HTML generally.

## §3 Populating the deploy repo

```bash
# stage from the BUCKET download, dropping B2's directory placeholders
cd ./bucket && find . -type f ! -name '.bzEmpty' -exec cp --parents {} ~/projects/sites/mortgagecalculator.co.uk/ \;
```

`.bzEmpty` are zero-byte B2 folder markers. **Do not carry them**: the sibling
`loanandmortgagecalculator.co.uk` has **0** in its bucket against **52** repo files,
so 1:1 repo↔bucket is the established shape and the markers get swept on first sync.

`README.md` exists locally and is **404 live** — not in the bucket. Do not add it;
adding it publishes it.

**Commit with an explicit pathspec** — that repo carries other sessions' untracked
files (`.idea/`, `asset_key`, `kind=icon`, `scope_ref`, a stray `gamedesign.uk/…`):

```bash
cd ~/projects/sites
git add mortgagecalculator.co.uk
git commit mortgagecalculator.co.uk -m "..."
```

## §4 Proving the sync is safe BEFORE pushing

Run the workflow's own command as a dry run:

```bash
cd ~/projects/sites
b2 sync --delete --skip-newer --dry-run --no-progress \
  mortgagecalculator.co.uk b2://portfolio-sites/mortgagecalculator.co.uk
echo "EXIT ${PIPESTATUS[0]}"     # MUST be 0 -- see the --dryRun gotcha in §1
```

**Expected, and it is NOT a no-op:** 29 uploads + 35 deletes. Read the deletes
before panicking:

- **30 are `(old version)`** — B2 version pruning, each paired with a re-upload of
  the *same bytes*. `index.html (old version)` appears twice because the bucket
  holds two superseded versions.
- **5 are `.bzEmpty`** — the placeholders, as above.
- **0 are live content files removed without replacement.** That is the property
  that matters; check it with:

```bash
comm -23 <(grep -v '\.bzEmpty$' bucket_files.txt | sort) \
         <(cd ~/projects/sites/mortgagecalculator.co.uk && find . -type f | sed 's|^\./||' | sort)
# empty == nothing in the bucket is missing from the repo == sync deletes no content
```

**Why it uploads at all when the bytes are identical:** `b2 sync` compares mtime,
and freshly-staged files are newer than the Jan-2026 bucket objects. `--skip-newer`
skips only when the **destination** is newer. The GitHub runner does a fresh
`actions/checkout`, so its mtimes are newer too — same behaviour, and the same
behaviour every other domain already gets.

## §5 The outage chain this all exists to prevent

`deploy-to-b2.yml` derives changed domains from the diff, then:

```bash
for domain in $CHANGED; do
  if [ -d "$domain" ]; then
    b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"
  fi
done
```

A verbatim `page_rerender` **commits the rendered page into `gqls/sites`**
(`deploy_result.data.repo_url = https://github.com/gqls/sites`). If the domain is
absent from the repo, that commit *creates* the directory holding **one file**,
`[ -d "$domain" ]` becomes true, and `--delete` makes the bucket match one file.
**The run goes green.** Hence: populate the repo before anything can trigger a
rerender.

## §6 Adoption pre-flight (run 2026-07-31 — clean)

> **CORRECTED 2026-07-31, same session.** This section first carried the handoff's
> `spec::text ILIKE '%mortgagecalculator%'` phrasing. **Do not use it.** It matched
> **41 `page_rerender` rows that belong to `loanandmortgagecalculator.co.uk`** —
> the sibling domain *contains* ours as a substring — and 41 is a plausible page
> count for our site, so it reads as "another lane is mid-adoption here, stop".
> Every affected query below is now an exact match or a join. See NOTES and the
> `LANDMINES.md` entry on nesting domain names.

```sql
-- another lane already on this domain? JOIN and match EXACTLY; group by domain so
-- a foreign site announces itself instead of hiding inside a total.
SELECT s.domain, w.item_type, w.status, count(*)
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.status NOT IN ('complete','cancelled','rejected')
   AND s.domain = 'mortgagecalculator.co.uk'
 GROUP BY 1,2,3;

-- exactly ONE ported-page component must exist (matters on the locked path;
-- checked anyway — a second one is a fleet-wide problem either way)
SELECT id, name FROM content_components WHERE function='ported-page';

-- find YOUR run by payload, exact match, never by the printed id
SELECT orchestration_id, owner_agent_type, current_step, status,
       collected_data->'input_data'->>'fidelity' AS fidelity
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'destination_domain' = 'mortgagecalculator.co.uk'
 ORDER BY created_at DESC;
```

**Result on 2026-07-31:** our domain had **0 `sites` rows, 0 runs, 0 work items**.
The `fidelity` plumbing check (migration 274 / the `call_adopter` `input_mapping`
allow-list) is **moot on this path** — a dropped `fidelity` is indistinguishable
from `high`, because `082` itself defaults adopt-mode to `high`:
`FIDELITY="${FIDELITY:-high}"` (line 124). It is worth checking only for `locked`.

**Do not dispatch within ~300s of a chassis pod restart** — the spawn is silently
dropped. Check for a roll in progress *and* pod age before submitting:

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,READY:.status.containerStatuses[0].ready,START:.status.startTime'
```

Two replicaset name prefixes in that output means a roll is in flight; wait for it.

**Under `--fidelity high` the assertion in handoff §5d INVERTS.** It says
`needs_content_page` + `needs_tool_recreation` must be **zero**. That is the
`locked` assertion. On the recreate path those items are the *expected* output —
their absence would mean the run did nothing. Do not copy that check across.

## §7 Positioning specs — two refinements to what the handoff says

Both verified in code today, before relying on them.

**`audience` is not consumed, but it is not un-referenced.** Handoff §6 says the
aspect "is read by NOTHING". Practically true — no agent, prompt or pipeline path
reads its data. But `grep` finds exactly one hit and it is not a contradiction:

```go
// internal/core-manager/admin/spec_admin_handlers.go:226
if aspect == "identity" || aspect == "tone" || aspect == "audience" {
    scopeQuery += " AND COALESCE(p.page_type, '') NOT IN ('blog-post')"
}
```

That uses the aspect **name** to decide which pages an admin operation covers. It
never reads the spec body. Expect this hit; it does not mean the aspect is live.

**`content_direction.formatted` regenerates ITSELF — but only on the action path.**
Handoff §6.2 warns that a hand-written spec which does not regenerate `formatted`
is invisible. True, and worth being precise about *when*:

```go
// platform/orchestration/actions/site_spec_actions.go:211-217
// This runs for every content_direction write — classifier, adoption, HITL.
if aspect == "content_direction" {
    formatted := datahelpers.FormatContentDirection(specMap)
    if formatted != "" { specMap["formatted"] = formatted }
}
```

It runs **before the transaction**, unconditionally, for any write that goes through
`site_spec_actions`. So:

- writing the spec **through the platform's own action** → `formatted` is correct by
  construction, and the handoff's trap cannot fire;
- writing it with **raw SQL** (which is what the sibling's `set_divergence_specs.py`
  does) → `formatted` is whatever you put there, and a spec that looks applied
  changes nothing.

**Prefer the action path.** If raw SQL is unavoidable, reproduce
`datahelpers.FormatContentDirection` exactly and gate it as handoff §6.3 describes —
compare as a **multiset of lines**, not as a string, because Go map iteration order
is random so the stored section order is arbitrary.

Schema reminders that cost time otherwise: the JSONB column is **`data`**, not
`spec_data`; `idx_site_specs_current` is `UNIQUE (site_id, aspect) WHERE is_current`,
so the supersede must precede the insert.

## §8 Positioning — `divergence_rule` is INERT; use `target_audience` (corrected 2026-08-03)

> **CORRECTION to handoff §6 and to this lane's own plan.** The handoff says the
> sibling "now carries an explicit `divergence_rule`" and that we should mirror it.
> **Do not.** `divergence_rule` has **no consumer anywhere in the platform**:

```bash
grep -rn "divergence_rule" .    # docs + one .py script only — no Go, no SQL, no prompt
```

Worse, it is *fragile as well as inert*: the only step that serialises a whole
identity spec into a prompt is the classifier's own `classify_and_extract`
(`{{.site_specs.specs.identity | toJSON}}`), and the classifier's output schema has
no such key — so the next classifier run **drops it**.

**What actually carries divergence is `identity.target_audience`**, and that is what
the live sibling uses. Read it and match its shape:

```sql
SELECT data->>'target_audience' FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE s.domain='loanandmortgagecalculator.co.uk' AND ss.aspect='identity' AND ss.is_current;
-- "...Not the single-subject researcher. A visitor who only wants a mortgage
--  repayment figure ... is served better by a single-subject site..."
```

`target_audience` is a real consumer: a step-contract scalar in
`v3_site_actions.go:1233`, plus `component_library.go:116` and
`internal/agents/contentcreator/agent.go:53`.

**Second lever: `content_direction.things_to_avoid`** — reaches `build-site-planner`,
`tool-recreation-handler`, `blog-content-planner`, `content-gap-planner`.

### GOTCHA — editing `content_direction` by raw SQL silently changes nothing

The content writer reads **one** field: `{{.site_specs.specs.content_direction.formatted}}`.
`FormatContentDirection` regenerates it only on the **action** path. Raw SQL must
update the array **and** `formatted` in the same statement, or the spec looks applied
and steers nothing.

Format is exact and reproducible: `HumaniseKey` replaces `_` with spaces and
uppercases **only the first character** — so `things_to_avoid` renders as
`Things to avoid:` followed by `- ` items. Sections join with a blank line, and Go
map order is random, so **appending a section at the end is a valid output**; there
is no canonical order to preserve.

The verification that matters is array↔blob agreement, not length:

```sql
SELECT count(*) AS items, count(*) FILTER (WHERE position(item in fmt) > 0) AS in_formatted
  FROM (SELECT jsonb_array_elements_text(data->'things_to_avoid') AS item,
               data->>'formatted' AS fmt
          FROM site_specs WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
            AND aspect='content_direction' AND is_current) t;
-- must be N of N   (was 14 of 14 on 2026-08-03)
```

**GOTCHA — `substring(col from 'literal(.{0,400})')` fails**: Postgres regex
quantifier bounds cap at **255** ("invalid repetition count(s)"). Use
`substr(col, position('needle' in col), 320)`. Cost me one query; it is already in
016b §9.

### Supersede before insert, and do NOT rely on `pinned`

`idx_site_specs_current` is `UNIQUE (site_id, aspect) WHERE is_current`, so the
supersede must be a **separate statement before** the insert — a data-modifying CTE
doing both in one statement risks tripping the index. Column is `data`; `source` and
`created_by` are NOT NULL.

**`pinned` does not protect a spec** — `write_site_spec` never reads it and the
replacement row defaults to false (see `LANDMINES.md`). The durable hold is
`sites.locked_at`.

## §9 Holding the whole site — `sites.locked_at`

One reversible switch that stops **all** automated dispatch for a site, honoured by
the dispatcher (`load_work_item_actions.go:126-138`, returns zero items with
`skipped_reason: site_locked`) and by the 213 dispatch gate:

```sql
UPDATE sites SET locked_at=NOW(), locked_by='<lane>: why'
 WHERE domain='mortgagecalculator.co.uk';
-- release:
UPDATE sites SET locked_at=NULL, locked_by=NULL WHERE domain='mortgagecalculator.co.uk';
```

It gates **dispatch**, not completion — an in-flight orchestration finishes and
writes its spec; only the follow-on item is held. Preferred over deferring items one
by one once a **chain** is involved, because each handler creates the next item and
you cannot win that race at a 120-second tick.

**The build chain, so you know what you are holding back:**
`needs_domain_research` (classifier) → `needs_vertical_research`
(vertical-exemplar-researcher) → `needs_strategy` (domain-strategist) →
`needs_briefing` (build-briefing-agent) → **`needs_site_plan` (build-site-planner)**.
Everything up to briefing writes only specs. `build-site-planner` is where pages
start being planned — that is the line to hold if the live site must not change.

---

## §10 Rebuilding pages — the whole chain, with the gotchas attached (2026-08-03/04)

**Order is composition → design → pages, and it is not advisory.** A page renders its
chrome at build time; a stylesheet or a header arriving later does **not** retro-fit it.
Proven both ways on this site (08-03 canary).

### §10a Where chrome comes from — do not chase the `pages` columns

```sql
-- VESTIGIAL. Empty for all 562 pages fleet-wide. Do not diagnose from these.
SELECT rendered_header, rendered_footer, rendered_head FROM pages WHERE ...;

-- THE REAL STORE. Zero rows here IS the "no chrome" signature.
SELECT slot_name, build_status, length(rendered_html), updated_at
  FROM site_components WHERE site_id='62b5978e-4271-4589-8e00-4baebfc0447c';
```

Create them (`populate_nav_tables → render_site_components → create_rerender_items`):

```bash
./docs/agent_docs/docs024_key_docs_latest/bugfix_149_nav_membership/TRIGGER_nav_rebuild.sh mortgagecalculator.co.uk
```

⚠ **It files a `page_rerender` per page — 26 here, homepage included.** Defer them straight
after. What actually protects an unbuilt page is that it has **zero `page_components`**
(`rerender_single_page_action.go:565`, `:168-209` → `skipped:true`, no deploy) — *not*
`build_status`, and *not* `include_statuses`, which filters `pages.status` where the value
`deployed` never occurs.

### §10b Deploy ONE page, assemble-only, with the site still locked

```bash
./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
  <page_id> 62b5978e-4271-4589-8e00-4baebfc0447c mortgagecalculator.co.uk
```
No `reason` ⇒ assemble-only ⇒ **no LLM, authored copy untouched**. Bypasses the queue, so
the site can stay locked. Prereq: the page must already have `page_components`.

### §10c Building pages that do NOT yet have components — needs the QUEUE

`needs_content_page` (handler `page-build-handler`) is what creates components, and it only
dispatches through the gate, so **the site must be unlocked**. Documented subset procedure:

```bash
# 1. capture what is ALREADY armed, so the backstop never touches another lane's items
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT string_agg(quote_literal(id::text), ',') FROM site_work_items
 WHERE site_id='62b5978e-4271-4589-8e00-4baebfc0447c' AND status IN ('triaged','approved');"

# 2. arm ONLY your items;  3. unlock;  4. run the backstop every 15s
```

⚠ **The backstop must NOT be a bare "defer everything that is not my items".** Each handler
creates the NEXT item in the chain (`needs_content_page` → `page_rerender`), so a naive
backstop defers the follow-on and the page never deploys. Guard on the page name too:

```sql
UPDATE site_work_items SET status='deferred', updated_at=NOW()
 WHERE site_id='62b5978e-…' AND status IN ('triaged','approved')
   AND id NOT IN (<pre-existing armed>, <your items>)
   AND COALESCE(item_key,'') !~ 'guide-remortgaging|guide-buy-to-let|guide-negative-equity'
   AND COALESCE(summary,'')  !~ 'guide-remortgaging|guide-buy-to-let|guide-negative-equity';
```

**Queue position is FIFO on `created_at`, priority is only a tiebreak.** Check where you
are rather than assuming a stall — this site's `needs_content_page` rows are dated
2026-07-31, which made them the OLDEST in the fleet and they dispatched immediately.

### §10d New URLs cannot overwrite the original site — verify per page

Rebuilt guides go to **directory form** (`/guides/remortgaging/index.html`); the originals
are **file form** (`guides/remortgaging.html`). Different paths, both serve. **`/index.html`
is the ONE exception** — same URL in and out, so it overwrites live content. Confirm before
arming anything:

```sql
SELECT name, url FROM pages WHERE site_id='62b5978e-…' AND name IN (<your pages>);
```

### §10e Restoring a page the platform overwrote

```bash
cd ~/projects/sites
git log --format='%h %ci %s' -- mortgagecalculator.co.uk/index.html   # find the last good sha
git show <sha>:mortgagecalculator.co.uk/index.html > mortgagecalculator.co.uk/index.html
git commit mortgagecalculator.co.uk/index.html -m "..." && git pull --rebase && git push
```
**Rebase, never merge** — the deploy derives changed domains from `git diff HEAD~1 HEAD`,
and a merge makes your commit the first parent, dropping the domain while the run still
goes green. Live in ~40 s. Restore point for the original homepage: **`825a36994`**
(11,125 bytes, 28 links).

**Then remove the components the overwrite created**, or the next rerender re-does it:
```sql
DELETE FROM page_components WHERE page_id='<page>';   -- restores "assembles to nothing" protection
```
Back them up first (`SELECT json_agg(row_to_json(t)) …`); the rendered artefact is
recoverable from git regardless.

### §10f Verifying — the only check that counts

```bash
cd ~/projects/sites/mortgagecalculator.co.uk
while read -r f; do rel="${f#./}"
  curl -sf -o /tmp/v.tmp "https://mortgagecalculator.co.uk/$rel" || { echo "FETCH-FAILED: $rel"; continue; }
  [ "$(sha256sum "$f" | cut -d' ' -f1)" = "$(sha256sum /tmp/v.tmp | cut -d' ' -f1)" ] || echo "differs: $rel"
done < <(find . -type f ! -path './.git/*')
# expect exactly one line: robots.txt (Cloudflare)
```

⚠ **Use `curl -sf` and branch on failure.** With a bare `curl -o`, a missing output
directory makes every `sha256sum` compare against a non-existent file and **every file
reports "differs"** — which reads exactly like the site being destroyed. That happened on
08-04 (the session scratchpad had been relocated by another lane) and cost a genuine scare.

⚠ **Spot-check at the FULL `/index.html` form, never the bare directory URL** (added
2026-08-08). This host resolves a directory to its `index.html` at the ROOT only:
`/` serves, but `/tools/repayment/` and `/tools/repayment` both 404 while
`/tools/repayment/index.html` serves 200. A bare-URL spot-check therefore reads as
"every rebuilt tool is down" — which briefly diagnosed as a site-wide regression on
08-08. The sweep above was never at risk (it fetches full file paths); the trap only
bites hand-typed checks. The tell that it is THIS and not an outage: `/` still serves
the original homepage byte count (11,125B).

### §10g A backstop OUTLIVING its batch defers the NEXT batch's follow-ons (added 2026-08-05)

The §10c backstop loops on a timer (240×15s). If a second batch starts while an earlier
backstop is still looping, the old backstop's keep-list does not contain the new batch's
page names — so it silently defers the new batch's follow-on `page_rerender` items.
Measured 2026-08-05: 12 tool recreations went `complete`, only 2 pages deployed; the 10
missing deploys were `page_rerender` rows deferred by the GUIDES batch's backstop, which
had ~10 minutes of life left when the tool batch began. The two that deployed were filed
after it died.

Rules: **kill the backstop the moment its batch completes** (don't let the timer run
out); one backstop at a time; recovery is cheap — re-arm the deferred rerenders and let
them drain (`UPDATE ... SET status='triaged' WHERE item_type='page_rerender' AND ...`).
And re-check "did my batch actually DEPLOY" at the pages table, never at the work items:
12/12 `complete` with 2/12 `deployed` is exactly what this failure looks like.

## §11 Checking the legislation watch (evidence-freshness) on this site

Seeded 2026-08-09: `site_specs` aspect `evidence_base` (pinned), 4 SDLT facts
citing GOV.UK. **UPDATED 2026-08-10: now 13 facts** — one per band edge and per
rate (A1), so an oracle can read a band table off the register instead of
parsing prose; see §13 for how the quotes are lifted. The daily
`evidence-freshness` scheduled task re-fetches each citation URL and requires
the stored verbatim quote to still appear (normalised:
case/whitespace/curly-punctuation/thousands).

```sql
-- Did the sweep run, and what did it think of our facts?
SELECT f->>'id' AS fact, f->>'verified_at' AS verified_at
FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
     jsonb_array_elements(ss.data->'facts') f
WHERE s.domain='mortgagecalculator.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;

-- Any drift/staleness items raised?
SELECT item_type, status, left(summary,80) FROM site_work_items swi
JOIN sites s ON s.id=swi.site_id
WHERE s.domain='mortgagecalculator.co.uk' AND swi.item_type ~ 'stale_evidence|citation'
ORDER BY swi.created_at DESC LIMIT 5;
```

Gotchas: a `citation_lost` on the FIRST sweep after seeding means the stored
quote does not survive `VisibleTextFromHTML` extraction (my python extraction
vs the Go extractor) — fix the quote, do not read it as legislation moving.
The sweep task row's `last_completed_at` covers the WHOLE fleet, not this
site — proof this site was swept is `verified_at` moving on OUR facts (or an
item naming us), never the task timestamp.

## §12 Filing a tool-recreation item by hand (added 2026-08-10)

Three sessions have now filed these and none of them wrote the SQL down. It is
here, with the two things that cost time.

```sql
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, priority,
   handler_agent, status, created_by, item_key, approval_mode, pipeline, triaged_at)
VALUES ('62b5978e-4271-4589-8e00-4baebfc0447c', 'adoption', 'needs_tool_recreation',
        'medium', '<what and why>',
        $json${ "mode":"recreate", "source":"adoption", "page_name":"tool-stamp-duty",
                "page_type":"tool", "interactive_features":[ … ] }$json$::jsonb,
        '<pages.id for that page>', 8, 'tool-recreation-handler', 'triaged',
        '<lane>: <task>', 'needs_page:tool-stamp-duty', 'auto', 'build', NOW())
RETURNING id, status, created_at;
```

- **`$json$…$json$` dollar-quoting, not apostrophe-doubling.** The spec text
  contains apostrophes and `"` (an id contract quoting option values needs
  both); dollar-quoting takes them literally. Parse the JSON with python before
  running the file — a malformed spec inserts happily and fails at the agent.
- **`item_key` reuse is allowed once the previous row is terminal** (the dedup
  index only covers non-terminal rows), so re-filing the same page is a plain
  INSERT — no need to invent a new key.
- **`status='triaged'` + `triaged_at` is what makes it dispatch.** `detected`
  rows sit for ever; the site's ~40 of them are why an unlocked site is not
  already churning. Pickup is `build-pipeline-trigger` (120s), claimed by
  `build-dispatch-loop`. A whole recreation — analyse, rebuild, validate,
  fabrication-check, save, deploy — took **~5 minutes** end to end on 08-10.

**GOTCHA — `complete` on the work item is NOT success, and a fast `complete` is
the tell.** On 08-10 an item read `complete` **52 seconds** after filing. Its
orchestration had ended at `complete_error`, and `result.response` held the
*site record* (the output of an early step) rather than a tool. Read the
orchestration, not the item:

```sql
SELECT orchestration_id, current_step, status,
       jsonb_pretty(collected_data->'__step_error') AS step_error
FROM orchestration_states
WHERE collected_data::text LIKE '%<work item uuid>%' ORDER BY created_at DESC;
```

That day's `__step_error` was `AI endpoint unavailable … status 400 … "You have
reached your specified API usage limits."` — a fleet-wide Anthropic budget
window (failures 14:51–17:02Z, healthy again by 18:12). It is worth knowing what
this looks like, because **it presents as a work item that completed**. The
artefact is the arbiter: the failed run left `page_components` untouched
(same row id, same `updated_at`), which is how it was ruled a clean no-op.

```sql
-- did a NEW component actually land?
SELECT id, build_status, length(rendered_html), updated_at
FROM page_components WHERE page_id='<pages.id>';
```

## §13 Lifting citation quotes with the REAL extractor (added 2026-08-10)

A quote is only evidence if the daily sweep can still find it, and the sweep
uses Go (`datahelpers.VisibleTextFromHTML` + `QuoteFoundInText`). Extracting
with python and hoping the two agree is the day-one `citation_lost` trap in
§11. Remove the class instead: run the Go functions.

A ~60-line `main.go` in a scratch module with

```
replace github.com/gqls/agentchassis => /home/ant/projects/agentchassis
```

imports `platform/orchestration/datahelpers` directly, fetches with the same
User-Agent and Accept headers as `fetchCitationDocument`, and offers `dump`
(print the visible text) and `check <quote>…` (verify). Keeping it OUT of the
repo tree matters on a tree this many sessions share — nobody can sweep it into
a commit.

Then **lift the quotes out of the dumped text with a regex, never retype them**,
and use `.` where the source has a currency symbol or a curly apostrophe so no
non-ASCII is typed at all (this channel rewrites some characters on emission).
Require exactly one match per quote, so a page edit that duplicates or drops a
phrase fails loudly.

**Induce the red before trusting the green.** Ask the checker for a quote you
know is wrong — `Up to £126,000 Zero` — and confirm it reports `NOTFOUND` and
exits non-zero in the same run as the real one. Thirteen FOUNDs prove nothing
until one NOTFOUND proves the check can fail.

## §14 Writing tool PLANs (the acceptance fences) — added 2026-08-10

Three scripts, in this order. All live in `acceptance/`.

```bash
# 1. EMIT — drive the live rebuilt page, record what it answers.
#    One tool per invocation: every page here is .../index.html, so the emitter
#    would write index.criteria.json each time and silently overwrite.
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
for slug in simple repayment stamp-duty overpayment bridging-loan \
            equity-release fee-analyser rate-forecaster portfolio; do
  python3 toolgolden.py --emit-criteria /tmp/crit/tmp \
      https://mortgagecalculator.co.uk/tools/$slug/index.html
  mv /tmp/crit/tmp/index.criteria.json ../mortgagecalculator_couk_adoption/acceptance/criteria/$slug.criteria.json
done

# 2. RE-DERIVE — nothing is pinned until something other than the page agrees.
cd ../mortgagecalculator_couk_adoption/acceptance
python3 verify_criteria.py                                   # must be 0 MISMATCH
python3 verify_criteria.py --mutate sdlt-ftb-relief-cap=625000 stamp-duty
                                                             # must EXIT 1

# 3. INSTALL — supersede + insert one doc_plans row per tool.
python3 install_fences.py            # dry run first, always
python3 install_fences.py --apply
```

**THE SUBJECT KEY IS NOT THE PAGE NAME, and getting it wrong fails silently for
ever.** Both tiers derive it (`discovery_checks/tool_eligibility.go`):
`cc.function` when the component is `component_level='tool'`, otherwise
`regexp_replace(p.name,'^tool-','')`. Our recreated pages carry a SECTION
component, so `tool-stamp-duty` is keyed **`stamp-duty`**. A PLAN under the wrong
key produces no error — Tier 2 goes on writing `needs_criteria`, Tier 4 goes on
emitting nothing, and the site looks exactly as it did before you started.
`install_fences.py` reads the key out of the eligibility query rather than
building it in python, so it cannot drift from the Go. To check by hand:

```sql
SELECT CASE WHEN cc.component_level='tool' THEN cc.function
            ELSE regexp_replace(p.name,'^tool-','') END AS ladder_subject_key,
       p.name, p.page_type, cc.component_level, cc.function
FROM pages p JOIN sites s ON s.id=p.site_id
JOIN page_components pc ON pc.page_id=p.id
JOIN content_components cc ON cc.id=pc.component_id
WHERE s.domain='mortgagecalculator.co.uk' AND p.page_type='tool' ORDER BY p.name;
```

**Confirm it against the platform's own opinion before installing** — the
`needs_criteria` notes name the key the sweep is actually looking for:

```sql
SELECT subject_key, created_at FROM doc_notes
WHERE subject_type='tool' AND categories ? 'needs_criteria' ORDER BY created_at DESC;
```

**Only three of the twelve recreated tools are ineligible, and it is structural**
— `tool-affordability` has two components (the eligibility rule wants a SOLE
component on a `page_type='tool'` page), `game-fact-finder` is `page_type='game'`,
`investor-index` is `page_type='section-index'`. A PLAN for any of them is a row
nothing loads.

### The check that must run BEFORE `--apply`, every time

Installing a PLAN switches Tier 2 on for that tool, and Tier 2 appends **three
built-in shell failures independent of your fence**
(`check_tool_acceptance.go:478-500`): `shell-doc-header`,
`shell-template-residue`, `shell-dead-controls`. A failure raises `improve_tool`
carrying `spec.component_id` — for these pages the **shared `hero` component,
252 pages across 18 sites**. `no_auto_fix` does NOT cover this path; it governs
Tier 4 only.

Run the three checks with the platform's own functions, never a re-implementation
(`DeadControlAnchorsOutsideRuntimeFill` carries a per-anchor runtime-fill
exemption, `bugs_open/137`, that a rewrite will get wrong). The scratch module is
the §13 pattern — a `go.mod` with
`replace github.com/gqls/agentchassis => /home/ant/projects/agentchassis`, kept
OUT of the repo tree so no session can sweep it into a commit:

```go
strings.Contains(html, content.ToolDocOpen)              // shell-doc-header
strings.Contains(html, "<no value>")                     // shell-template-residue
datahelpers.DeadControlAnchorsOutsideRuntimeFill(html)   // shell-dead-controls
```

**Induce the red in the same run**: a fixture carrying `<no value>`,
`<a href="#">…</a>` and `/* === tool-doc ===` must report all three. Twelve
passes from a checker that has never failed are not evidence of anything.

### Why an emitted value is not yet an expected value

`--emit-criteria` records what the tool prints today. Pinning that is F3
(`PLAN_2026-08-09` §0) and `run_checks_action.go:775-781` says so in the code
that does it. `verify_criteria.py` recomputes every value from a source that is
not the page's script, at three labelled strengths (DEFINITION / REGISTER /
CONVENTION), and `install_fences.py` **pins only what was re-derived** — which
drops containers, prose and echoed inputs without needing a heuristic.

Two traps that cost time here:

- **A sign can sit OUTSIDE the currency symbol.** `<U+2212>£961.61`. A
  `re.search(r"-?\d…")` needs the sign adjacent to the first digit, so it matches
  at the `9` and returns a POSITIVE number. Use `re.sub(r"[^0-9.\-]","",s)` and
  map U+2212/U+2013 to ASCII first. A parser used on both sides of a comparison
  agrees with itself and writes the error into the record.
- **A one-month disagreement on a duration is usually a CONVENTION, not a bug.**
  The pages stop amortising at `remaining > 0.005`; `oracles.amortise` stops at
  `1e-9`. Do not pin a duration until that threshold is stated somewhere.

### The scheduler will not sweep most of these

`check_tool_acceptance_due.go` gates on `PageHasShippedPredicateFor` =
`NOT (deployed_at IS NULL AND build_status <> 'deployed')`. ~~Seven of this site's
eight fenced tools have `deployed_at IS NULL` and `build_status='needs_rebuild'`
**while serving HTTP 200** (`links.go:304-308` records the same nine pages).~~

> **CORRECTED 2026-09-02 — this paragraph is STALE and now says the opposite of the truth.**
> Re-run today, **all 18** `page_type='tool'` pages read `build_status='deployed'` and
> `invisible_to_due_sweep = f`. Nothing here is hidden from the due sweep any more, and the
> "fire a run by hand" instruction below is no longer a workaround for this cause (it remains
> a valid way to force a run). **The blocker moved, it did not clear** — see the next block.
Check before assuming a fence is live:

```sql
SELECT name, build_status, deployed_at,
       (deployed_at IS NULL AND build_status <> 'deployed') AS invisible_to_due_sweep
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='mortgagecalculator.co.uk' AND page_type='tool' ORDER BY name;
```

Until that is resolved, fire a run by hand — the due sweep's own item shape:

```sql
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, priority,
   handler_agent, status, created_by, item_key, approval_mode, pipeline, triaged_at)
VALUES ('62b5978e-4271-4589-8e00-4baebfc0447c', 'discovery', 'acceptance_run', 'low',
        'Tier-4 acceptance run: <key>',
        $json${"function":"<ladder subject key>","component_id":"<cc.id>",
               "page_id":"<pages.id>","page_name":"<pages.name>",
               "check":"tool_acceptance_due"}$json$::jsonb,
        '<pages.id>', 90, 'tool-acceptance-agent', 'triaged',
        '<lane>: <task>', 'acceptance_run:<key>:62b5978e-4271-4589-8e00-4baebfc0447c',
        'auto', 'build', NOW())
RETURNING id;
```

Claimed ~3 minutes after insert; the neighbouring lane's runs took **~30 minutes**
created→complete. The verdict is a `doc_notes` row, NOT the work item:

```sql
SELECT categories, left(body,400) FROM doc_notes
WHERE subject_type='tool' AND subject_key='<key>' ORDER BY created_at DESC LIMIT 1;
```

### Verify the fence at the artefact, not at the status

`position('```criteria' in body)` and a `LIKE '%computed_values%'` prove the row
is non-empty, nothing more. Read each fence back OUT of the database, parse it,
and compare to the source file — 68 of our 80 assertions carry non-ASCII (the
U+2212 minus, currency symbols) and a quoting or encoding fault would land
exactly there. `computed_values` compares the EXACT text; whitespace is the only
latitude.

## §15 When a triaged build item is not picked up: the fleet queue starves this site; dispatch by hand (added 2026-08-11)

A correctly-filed `triaged` item on this site can sit for HOURS untouched while
the dispatch machinery visibly completes work on other sites. That is not a
fault in the filing and not a dead queue — it is the site-selection rule:
`build-pipeline-trigger` (every 120s) runs `find_dispatchable_site`, which picks
**ONE site per tick, ordered by the fleet's globally OLDEST dispatchable item**
(`ORDER BY wi.created_at ASC … LIMIT 1`, read from the live agent definition).
Any site holding older triaged items outranks yours on every tick. Measured
2026-08-11 ~15:30Z: 7 sites held ~273 dispatchable items older than a
just-filed item here — 81 of them from **18 days earlier** on one site — so this
site's ETA was "when everyone else's backlog drains". `priority` is irrelevant
to site selection (and note the per-site pickup inside the loop orders
`priority ASC` — LOWER number first — so copying a low number is fine).

**Diagnose before bypassing** — all three must hold, or the bypass just hides a
real fault:

```sql
-- 1. the item is actually dispatchable (status/attempts/approval/depends_on)
SELECT status, attempt_count, max_attempts, approval_mode, depends_on
FROM site_work_items WHERE id='<item>';
-- 2. the site is unlocked and has no claimed item hogging the per-site slot
SELECT locked_at FROM sites WHERE id='62b5978e-4271-4589-8e00-4baebfc0447c';
SELECT count(*) FROM site_work_items
WHERE site_id='62b5978e-4271-4589-8e00-4baebfc0447c' AND status='claimed';
-- 3. the trigger itself is alive (else the problem is bigger than this site)
SELECT last_triggered_at FROM scheduled_tasks WHERE name='build-pipeline-trigger';
```

**The bypass**, precedent
`scripts/initial_messages/180_adoption/081b_trigger_dispatch_gamesdesign.sh`:
publish one `orchestrate` request for `build-dispatch-loop` pinned to this
site (kcat pod in the kafka namespace; swap SITE_ID/DOMAIN in a copy of that
script). Blast radius = exactly the `triaged`/`approved` items on THIS site,
because that is all the loop's item query selects — check what those are first.
Used 2026-08-11 for item `97f4d0ab` (correlation `5125e6b6…`): claimed within a
minute of the publish, after 80+ minutes starved.

**kcat exit 0 is NOT delivery** (LANDMINES: `kcat -P` can silently drop) — the
only check is the result: the item flips to `claimed` and an orchestration row
appears. If nothing moves in ~3 minutes, treat the message as never sent.

## §15 Is the ladder even LOOKING at this tool? (added 2026-09-02)

§14 tells you how to write and install a fence. It does not tell you how to check the ladder can
reach the tool at all, and on this site **9 of 18 tool pages it cannot** — while seven of those nine
hold a current PLAN that nothing loads. Run the platform's own predicate, never a paraphrase of it
(`discovery_checks/tool_eligibility.go`, `toolEligibilityWhere` + `toolSubjectKeyExpr`):

```sql
SELECT DISTINCT p.name AS page,
       CASE WHEN cc.component_level='tool' THEN cc.function
            ELSE regexp_replace(p.name,'^tool-','') END AS ladder_subject_key,
       cc.component_level
FROM pages p
JOIN page_components pc ON pc.page_id=p.id
JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='62b5978e-4271-4589-8e00-4baebfc0447c'
  AND COALESCE(pc.slot_name,'') <> 'removed'
  AND cc.is_active AND p.status='active'
  AND ( cc.component_level='tool'
     OR ( p.page_type='tool'
          AND NOT EXISTS (SELECT 1 FROM page_components pc_t
                            JOIN content_components cc_t ON cc_t.id=pc_t.component_id
                           WHERE pc_t.page_id=p.id AND cc_t.component_level='tool' AND cc_t.is_active)
          AND (SELECT count(*) FROM page_components pc_n
                 JOIN content_components cc_n ON cc_n.id=pc_n.component_id
                WHERE pc_n.page_id=p.id AND cc_n.is_active) = 1 ))
ORDER BY 1;
```

**A page drops out of this set by GAINING a component.** The sole-component clause is the one that
carries our adopted pages, and adding a `generic-text-block` beside the calculator silently ends
their eligibility — the PLAN keeps its key, the page keeps serving, and the verification just stops.
That is how seven fences here became orphans. **Cross-check the answer against the fences:**

```sql
SELECT subject_key, created_at::date FROM doc_plans
 WHERE subject_type='tool' AND is_current AND body LIKE '%```criteria%' ORDER BY 1;
```
Any `subject_key` in the fence list but NOT in the eligibility list is a fence nothing reads.

## §16 Two extractor traps that manufactured false findings here (added 2026-09-02)

Both produced a confident wrong answer in one session; both are one word to avoid.

- **SQL `substring(x from '…')` returns the FIRST match only.** Reading
  `substring(rendered_html from '<img[^>]+src="([^"]+)"')` per row said `tool-cta` renders the same
  stamp-duty card on all four pages. It renders **six distinct** cards; the first is stamp-duty
  every time. Use `regexp_matches(…, 'g')` with `count(*)` **and** `count(DISTINCT …)`, and read the
  count before you read a value.
- **`LIKE '%background-image%'` matches the CSS PROPERTY NAME in a `<style>` block.** It reported
  every `hero-tool` row as carrying a background image; they carry a solid colour. Test for the
  value, not the property: `rendered_html ~ 'url\(''?/assets'` — or extract the URL and look at it.

## §17 A demand control for "is this asset referenced anywhere?" (added 2026-09-02)

Before recording a zero from a `page_components.rendered_html LIKE '%<file>%'` query, **prove the
query can return non-zero for that kind of file.** `snippets.js` is referenced by all 42 pages and
returns **0** from this query — it is injected by the chrome assembler and never stored in a
component, so it is a *worthless* control that agrees with your target. The working control is a
script that genuinely lives in a stored component:

```sql
-- must return 1 — if it returns 0 your query is blind, not your target absent
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
 WHERE p.site_id='62b5978e-4271-4589-8e00-4baebfc0447c'
   AND pc.rendered_html LIKE '%mortgage-lender-directory-listing.js%';
```
Enumerate what the column actually holds first — `SELECT substring(pc.rendered_html from
'src="([^"]*\.js)"'), count(*) … GROUP BY 1` — and pick the control from that, not from what the
live page serves.
