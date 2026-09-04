# RUNBOOK — bugfix 475, the delivery email's instructions promise

Every command here had a gotcha attached when it was got right. Fix commands **here**, not in
scrollback.

---

## Read the live email template (the bug's ground truth)

The copy is **step config on an agent definition**, not code and not a site spec. That is why it is
fixable without a roll — and why two separate price/copy censuses have missed it (mig `726`'s header
explains the £200 case).

```sql
SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
FROM agent_definitions
WHERE type='delivery-email-sender'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

⚠ **Timestamp your read, and check what else has touched the path today.** Migration `776` was
applied to this exact jsonb path at 12:05:25Z on 2026-09-04 for an *unrelated* false-promise clause.
"That template was migrated today" reads as "the bug is fixed" and is not evidence either way.

```bash
ls docs/agent_docs/sql_for_agents/ | grep -iE 'delivery|email' | sort   # what has aimed at this row
```

Narrow check for this bug's specific clause, which is what to cite:

```sql
SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
       LIKE '%ZIP comes with instructions%' AS clause_still_present
FROM agent_definitions WHERE type='delivery-email-sender'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Prove `webdesign.uk` is a framework site — at the artefact, with a control

The DB row is **not** sufficient. A page count says nothing about whether the route serves.

```sql
SELECT s.domain, s.id, count(p.id) AS pages
FROM sites s LEFT JOIN pages p ON p.site_id=s.id
WHERE s.domain='webdesign.uk' GROUP BY 1,2;
-- 1fcfa4f3-ec80-4010-878b-b971cd46711f, 18 pages [MEASURED 2026-09-04]
```

⚠ **`pages` has no `slug` column** — it is `name` and `url`. A `SELECT p.slug` errors out.

Then the check that actually decides it (a parked host 200s every path, so the **404 control is the
load-bearing half**):

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://webdesign.uk/guides/tool-css-variables-guide.html  # expect 200
curl -s -o /dev/null -w '%{http_code}\n' https://webdesign.uk/guides/this-does-not-exist.html       # expect 404
```

## Who else is in this file / this row, before you edit

Three lanes were simultaneously live on the delivery email on 2026-09-04. **Ask; do not infer.**

```bash
python3 scripts/who-owns.py 475          # reads COMMITS — blind to a session mid-fix
git log --since=1.day --format='%ad %h %s' --date=format:'%m-%d %H:%M' -- platform/delivery platform/orchestration/actions/send_delivery_email_action.go
git status --porcelain platform/delivery platform/orchestration/actions/send_delivery_email_action.go   # a same-file passenger
```

Then `ListAgents` and message the lane. **A live session answering is the only current ownership
signal on this tree**; `who-owns.py` cannot tell "owns the bug" from "is busy nearby", and it reported
475 as OWNED when the lane had done copy only.

Known neighbours as of 2026-09-04: **`site_delivery_and_editor`** (the copy), **`bugs_open/477`**
(`handover.go`, `site_deliveries`, mig 778, the follow-up letter), **`stripe`** (every price in that
same `body_template`).

## The ZIP — what is in it, and the assertion that will bite

```bash
# list a delivered zip by extension, which is the only honest census
unzip -l <zip> | awk '{print $4}' | sed 's/.*\.//' | sort | uniq -c | sort -rn
```

⚠ **Do NOT grep the listing for `readme|instruction|guide|how|start|help`.** On a site that *has* a
guides section it returns three hits that are all site content (`hero-guides.jpg`, `icon-guides.jpg`,
`guides/index.html`). **A needle matching the subject matter as well as the artefact type finds the
subject matter.** Count by extension: there are no document files at all.

⚠ **`zip_deliverable_action.go:259-261` asserts `len(zr.File) == len(files)`.** Adding a `README.txt`
without teaching that assertion **makes the action fail**. Teach it the synthesised entries **by
name**, never by loosening the count — a count-only relaxation lets a missing README and an extra
site file cancel out.

## Deploying the owner's screenshots (phase 3) — the purpose-collision trap

Source: `/home/ant/Downloads/idea_uk_netlify/` (ten captures).

⚠ **`deploy_image_asset` resolves its source by PURPOSE, not by the `asset_id` you pass** — the
second same-purpose asset on a site silently deploys as the first. Ten screenshots will share a
purpose, so this is the worked case, not a near miss.

- Supply `spec.s3_uri` explicitly as a real `s3://bucket/key`, derived from the asset's own
  **`storage_path`**. **Not its `url` column** — that may be a stale presigned link or a local path
  post-deploy (`bugs_open/152`). A non-empty `s3_uri` is consulted *before* the buggy site-wide cache.
- **Verify at the artefact:** `sha256sum` the ten deployed files and assert they **differ**, then open
  one. `success: true` ten times over is exactly what the collision looks like.

## Ordering — the one that wedges a delivery if you get it backwards

**Config is live on apply; Go is inert until a roll.** So the template must **not** name
`{{instructions_link}}` until a chassis image carrying the vocabulary has rolled. A template naming a
token the running binary lacks leaves the literal in the body, trips the post-fill `{{` scan — **and
that scan fires AFTER `delivery.Claim` has stamped the handover.** Result: a stamped, undeliverable
handover needing the operator re-mint recipe.

Prove the roll before raising the migration — **ask the pod, do not grep the binary**:

```sql
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```

```bash
git merge-base --is-ancestor <your-commit> <the git_commit above> && echo SHIPPED
```

⚠ Filter by `pod_name`, not the `service` column (other pods share the image). ⚠ It is a **two-hour
window**, not a history — fine for "has my commit rolled yet", useless for dating anything older.

## Build and verify

```bash
scripts/verify-head-builds.sh --with <file> --test      # against HEAD, BEFORE committing
scripts/verify-head-builds.sh                           # after committing
```

Never hand-roll `git archive HEAD | tar` — that recipe is why this machine keeps filling up.
`/tmp` is a 16 GB tmpfs (RAM); a full one presents as
`link: mapping output file failed: no space left on device`, which reads like a compiler fault.

## Getting content INTO a framework page (the instructions page, phase 3)

The copy lane established there is **no verbatim-content spec field**. Verified here independently
against the live `page-build-handler` row rather than taken on trust — `grep -o 'spec[a-z_.]*'` over
`default_config` returns exactly: `spec.mode`, `spec.page_id`, `spec.page_name`, `spec.suggestion`,
`spec_sections{,.sections,.section_subjects,.section_facts}`.

⚠ **The `{{…}}` interpolations are ESCAPED in the stored JSON, so `grep -oE '\{\{...\}\}'` over the
row returns ZERO and reads as "nothing is interpolated".** Grep the bare key names instead.

**Three channels reach the writer, not one.** The copy lane named the first; the other two are
structured and are the better home for exact quotes:

| channel | wired as | reaches the writer as |
|---|---|---|
| `spec.suggestion` | `rewrite_guidance?` on the content-writer call | free text, the only prose channel |
| `spec_sections.section_subjects` | `plan_sections.section_subjects` | `PlanItem.Subject` → `current_section.subject`; the v5 prompt renders it only when non-empty |
| `spec_sections.section_facts` | `plan_sections.section_facts` | `assigned_fact_ids` → **`AssignedWriterBlock`, composed from ONLY the assigned facts** (`FactsScoped=true`) |

**So for text a customer must see VERBATIM — the Netlify error strings — `section_facts` is the
designed channel**, because the writer is handed the assigned facts as its evidence block instead of
the whole-site one, rather than being asked in prose to please quote something exactly.

`[MEASURED 2026-09-04]` `webdesign.uk` already has a live `evidence_base` — **27 facts**, all
carrying `source`, **none** using `attested_by`. So the register is armed on this site; an attested
fact would be first-of-kind here.

> ⚠ **DO NOT promise that registering them buys automatic staleness detection. It does not, for
> these strings.** `refresh_evidence_base` re-runs a fact's own `source.sql`, re-proves an
> `artifact_check` against a stored artefact, and re-proves a citation against its URL — none of
> which can re-derive *what a third party's signup screen says today*. `attested_by` (the right shape
> for the owner's performed run) has no automatic re-check at all. A URL-citation source pointed at
> Netlify would be worse than nothing: the estate has already rejected a Cloudflare-fronted source
> for **perpetual false drift** (`maps.org.uk`, 2026-09-02).
>
> **What registering them DOES buy is that the writer cannot paraphrase them away.** That is worth
> it on its own. Staleness stays a human job, which is what `bugs_open/475`'s rot rule already says.

**Also note:** `spec.sections` really is a list of section NAMES (verified by the copy lane against a
completed item, an 11-element array beginning `hero`), so prose cannot travel that way. But
`section_subjects` and `section_facts` are PARALLEL arrays aligned to it by index — the alignment is
enforced (`load_page_sections_from_spec_action.go:624`, `len(specSectionFacts) == len(specSections)`
or the key is omitted entirely: *"aligned or absent, never guessed"*).

## The instructions page: placement DECIDED, and the row must exist before the item

**`page_type = 'content'`, URL `/putting-your-site-online.html`** (decided by this lane 2026-09-04;
the copy lane holds the words and fires the filing).

`[MEASURED 2026-09-04]` `webdesign.uk` derives page URLs from `page_type` alone, no site-area:

| `page_type` | URL shape | occupants |
|---|---|---|
| `content` | `/<name>.html` | contact, faq, how-it-works, what-you-get (4) |
| `blog-post` | `/guides/<name>.html` | the six tool guides (6) |
| `tool` | `/tools/<name>/index.html` | the six tools (6) |
| `landing` | `/index.html` | index (2 rows, one superseded) |

**Why `content` and not `blog-post`, recorded because the reason outlives the decision:** this URL
goes into a `README.txt` inside a ZIP **that can never be corrected once a customer has it**. Every
other customer link is correctable (the page) or expiring (the tokens). So **stability dominates
tidiness.** `/guides/` is 100% tool guides — a coherent marketing/SEO section, and coherent sections
get reorganised (a `nav_restructure` capability gap is already queued). Top-level `content` sits
beside `how-it-works` and `what-you-get`, which have been stable and describe what the customer is
buying. Leave `noindex` FALSE: a customer who lost the email should be able to find it.

### ⚠ `page-build-handler` CANNOT CREATE A PAGE — the row must pre-exist

`load_page_record` is a **SELECT only** (`SELECT id, name, title, page_type, sections, url,
build_status … FROM pages`) — no INSERT anywhere in the action. The next step is a conditional on
`page_record.found == true` whose **`else_step` is `complete_error`**, described as *"audit findings
for new pages will skip here"*.

**So: create the `pages` row first** (`name`, `url`, `page_type`, `build_status='planned'`), **then**
file the `needs_content_page` item pointing at it with `spec.page_id`. Filing the item alone
completes with an error and builds nothing.

⚠ **There is no `page_role` column** — `information_schema` returns **0** for it on `pages`. A spec
key of that name reaches nothing (`page-build-handler` interpolates only `spec.mode`,
`spec.page_id`, `spec.page_name`, `spec.suggestion`, `spec_sections`). **`page_type` is not settable
from the spec at all: you set it when you create the row**, so a wrong one is not a re-file away —
it is baked in, and it determines the URL.
