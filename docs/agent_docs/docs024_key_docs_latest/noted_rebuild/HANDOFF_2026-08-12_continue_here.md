# HANDOFF — noted.co.uk rebuild, continue here

**Written 2026-08-12 16:20. Supersedes `HANDOFF_2026-08-11_continue_here.md`.**
Standalone: you should not need to read anything else to start.

Then, in order of usefulness: `README_where_we_are.md` (plain-prose history for the
owner), `NOTES_noted_rebuild.md` (technical log incl. every misstep — the 08-12
entries are dense and worth reading before touching the CTAs or the box),
`PLAN_2026-08-10_noted_rebuild.md`, `RUNBOOK_noted_rebuild.md`.

---

## 1. What this is, in one paragraph

noted.co.uk is a note-taking app the owner hand-built: text, voice recordings,
photos, a little version history. It has been live since January out of the
framework's own B2 bucket. We are rebuilding it as a **fully decomposed** framework
site with a server-side backend, so a person can sign in and reach their notes from
any browser. **The legacy app still serves noted.co.uk** with a wind-down notice;
the new backend is live and public at `app.noted.co.uk`; the framework front end is
**built, deployed to the box, and not yet public**. No user data has moved and
nothing users depend on has broken.

## 2. THE FIRST THING TO DO

**Check whether the tool page finally assembled.** Everything else is done; this one
thing is what stands between us and a working migration path.

```sql
SELECT p.name, p.url, p.build_status
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='noted.co.uk' ORDER BY p.created_at;

SELECT wi.item_type, wi.status, wi.spec->>'page_name' AS page, wi.updated_at
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='noted.co.uk' AND wi.status NOT IN ('complete','cancelled','rejected')
ORDER BY wi.updated_at DESC;
```

> **RESOLVED 17:12 — the tool page is BUILT AND SERVING. Blocker 1 is closed.**
> Read this section for the trap it records, then go to §5 blocker 2, which is now
> the top of the list.

**State at 17:12 on 2026-08-12:** six pages `deployed` (five content pages +
`tool-legacy-rescue`); `tool-legacy-rescue-guide` still **`planned`**, blocked on
content validation (§5.2).

Getting there took two attempts, and the first is the lesson:

- **16:18:30 — `…_assemble`. Completed, and did the WRONG THING.** I built it by
  copying an existing `page_rerender` row and overriding only `spec`. The row
  carried the *contact* page in its **`page_id` column**, the handler read the
  column, and it re-deployed `/contact.html` while reporting `"success": true`.
  Harmless (contact is byte-identical) but it is why the page was still `planned`
  after a "successful" run. See §6.
- **17:07:12 — `…_assemble_v2`, `page_id` column set correctly. Completed 17:10:09**
  and the page flipped to `deployed` — which confirms the diagnosis rather than
  working around it.

`[MEASURED 17:12]` verified at the artefact, not the status:

| check | result |
|---|---|
| `gqls/vm-sites/noted.co.uk/tools/legacy-rescue/` | present |
| box `/var/www/noted.co.uk/tools/legacy-rescue/index.html` | present, landed 17:11:46 (one sitesync tick) |
| box-local `GET /tools/legacy-rescue/` | **`200 / 23425`** |
| rescue markup / `indexedDB.open` in the served page | 12 / 1 — the tool is really there |
| `tool-doc` header in the served page | **0 — stripped at deploy, exactly as the contract says** |
| `migrate.html` on the box links to the tool | 2 (hero + call-to-action) |
| shopfront control | `200 / 28075`, unharmed |
| live apex still the legacy app | yes — "being refreshed" still served from B2 |

**One deviation from PLAN §4.2, recorded not hidden:** the plan says a tool's JS
"deploys as a real asset (`/tools/assets/{fn}.js`)". It did **not** — `tools/assets/`
holds only `contact-form.js`, and the rescue JS is **inline in the page**. That is a
consequence of how I supplied it: I put the JS inside `html_content`, so
`content_components.js_content` is empty and there was nothing for the asset
emitter to extract. The tool works as served; if a separate asset is wanted, the JS
needs to go in `js_content`.

## 3. What is LIVE

### noted.co.uk — the legacy app (repo `gqls/sites`, branch **`master`**)
Unchanged this session. The original app plus a wind-down notice, a working
**"Save everything"** full backup (text + recordings + photos + history), and four
WCAG-AA contrast fixes. **This is still what the public sees at the apex.**
Verified 2026-08-12: `x-amz-*` headers (B2), body still contains "being refreshed"
and "Save everything".

### app.noted.co.uk — the engine
`{"status":"ok"}` from the open internet, unchanged this session. Accounts,
sessions, notes, media, and an import accepting **exactly** the "Save everything"
format. On `webdesign.vs.mythic-beasts.com` (176.126.243.62), key
`~/.ssh/webdesign_box_ed25519`. Backups nightly 03:20 → age-encrypted → B2
(Object Lock 30d); restore drill passes.

### The framework front end — NEW, live on the box, not public
Five pages **deployed and served from the box**: `index`, `how-it-works`,
`migrate`, `about`, `contact`. Box-local `Host: noted.co.uk` → **`200 / 26402`**.
Not reachable publicly: the apex still routes to B2. Cutover is still a deliberate
future step.

## 4. What changed on 2026-08-12 (this session)

### The build landed on its own
`needs_domain_research` completed 23:55 on 08-11 — **6h55m** after filing, against
a measured ~6h estimate. No bypass was fired. The whole cascade then ran unattended
overnight (classifier → vertical research → strategy → briefing → site plan → pages
→ imagery → rerender), finishing 04:37.

### Delivery to the box was silently broken; fixed AND hardened
Pages were reaching `gqls/vm-sites` correctly and **never reaching the box**.
`/var/www/noted.co.uk` was empty and nginx served 403, while `sitesync.service`
exited **0/SUCCESS every 5 minutes**.

Cause: the box's clone is a `--filter=blob:none --sparse` partial clone whose
**cone contained only `webdesign.uk`**. `noted.co.uk` was in `ls-tree origin/main`
but never materialised, so sitesync's missing-folder guard skipped it in silence.
The 08-11 change added the domain to `DOMAINS` in `/usr/local/bin/sitesync`; the
cone is set in `setup-webdesignbox.sh:70`, which only runs at provisioning. **Two
lists that had to agree, in two files, one never re-run.**

Fixed (`sparse-checkout add`) and then hardened: **sitesync now derives the cone
from `DOMAINS` on every run** and the one silent guard is split into two loud ones.
Box backup at `/usr/local/bin/sitesync.bak-20260812`; repo copy is
`noted_rebuild/box/sitesync` and matches the box.

### CTA destinations set (owner decision: primaries → app.noted.co.uk)
The build wrote every CTA's *text* and **no URL at all**, so all six hero /
call-to-action slots rendered **zero anchors** and nothing linked to the product.
Now: `index` and `how-it-works` primaries → `https://app.noted.co.uk/`; secondaries
follow the copy already written; `migrate` → the rescue tool. Six of the original
`unresolved_cta` items closed.

### `/legacy` — the rescue tool — BUILT and TESTED
`docs/.../noted_rebuild/legacy_tool/noted-legacy-rescue.html`. Reads the previous
app's `NotedDB` (same origin) and hands back notes, recordings, photos and history
as a file byte-compatible with the old app's own export. Created through the
framework via `create_tool_component`; component `tool-legacy-rescue`, page
**`/tools/legacy-rescue/index.html`**, plus an auto-created companion guide.

**Tested against real IndexedDB** — `legacy_tool/test_legacy_rescue.py`, Playwright,
24 checks, three cases. **Both load-bearing guards mutation-verified**: removing the
`abort()` fails "leaves no database behind"; ignoring the legacy `{blob}` record
shape fails the recording count and rescue checks.

Commits: `fc27a74e0`, `70732301c`, `e5d664f97`, `714c1d65c` (plus `23f1229f0`,
`3ce4da7a9` from the queue investigation).

## 5. THE BLOCKERS, in priority order

> **UPDATE 2026-08-13 evening — blockers 2, 3 and 5 are CLOSED; 4 remains.**
> The privacy copy is written (owner-approved), registered, and live on
> `/privacy.html` **verbatim — 22/22 sentences** (the writer only receives the
> `writer_block` STRING, so the copy now travels inline in it; `supplied_copy`
> alone reaches nobody). The guide's blocker was never the copy: it was the ban
> firing on a TRUE sentence about the old app — and the deeper defect was that
> `writer_block` itself said "The old site had no server at all", teaching the
> phrase it banned. Reworded (owner ruling: describe what the old app DID);
> the guide rebuilt clean first try and its CTAs point at the tool. The site's
> queue is EMPTY bar the three inert `detected` rows. Still genuinely open:
> **blocker 4** (degraded states by hand, pre-launch), the owner's
> deletion-vs-backup-retention disclosure decision, and cutover.
> Full trail: NOTES 2026-08-13.

| # | Blocker | What is known |
|---|---|---|
| ~~1~~ | ~~**Tool page is `planned`**~~ **CLOSED 17:12** — built, deployed, serving `200 / 23425` on the box, and migrate's button now resolves | Nothing promoted it automatically because `page_rerender` normally files items only for pages already `deployed`. Fixed by hand-filing one with the **`page_id` COLUMN** set — see §2 for the attempt that got that wrong |
| 2 | **The companion guide failed content validation** | `step validate_content failed … 1 blockers, 0 errors`, escalated to `needs_human_review`, `will_retry: false`. `[INFERRED]` it hit the `evidence_base` — 7 banned claims, **0 registered facts**, and a `writer_block` that explicitly forbids an agent writing privacy wording for this product. The specific blocker text is **not** recorded in the orchestration row. If that inference is right, the gate worked as designed |
| 3 | **The privacy copy is the owner's to write** (carried forward) | The old "no server" wording is banned in `evidence_base` so no agent copies it forward. A replacement I proposed previously was rejected — **do not reintroduce it.** Blocker 2 is probably a direct consequence of this gap |
| 4 | **Degraded states unverified** (carried forward) | The platform cannot induce a failing dependency. "Save fails loudly, text survives" must be exercised **by hand** before launch. It is the clause protecting the unrecoverable thing |
| 5 | Guide page has 2 `unresolved_cta` of its own | Same class as the ones already fixed; fix with the 074 script once the guide's content is unblocked |

Lower priority, carried forward: `voicescan` cannot run (needs a voice spec);
`cloudflared.service` has no `Restart=` (webdesign lane's file — raised, not
changed); off-box backup of `webdesign-chat` now rides in noted's backup and that
lane may not know; the owner accepted the key copies as they are — **do not reopen**.

Also present and inert: three `detected` discovery items from 08-10
(`needs_composition`, `needs_design`, `evaluate_tools`) — `detected` is not
dispatchable, nothing acts on them. One `failed` `needs_page` from 04:15
("Re-render index after its image asset landed") which **self-resolved** via the
04:12 rerender batch; it is a dead row, not an open failure.

## 6. TRAPS — read before touching anything

**New this session, and each one cost real time:**

- **A `complete` work item is not a repaired artefact.** Five `page_rerender` items
  reported `"success": true` with `deploy_result`s naming commits, and
  `rendered_html` changed by **zero bytes** (the assembled page was byte-identical,
  so git had nothing to commit and the adapter reported success anyway).
- **`page_rerender` re-assembles a page from EXISTING component HTML.** It does not
  re-render a component from `content_data`. The action that does is the section
  editor's `content_edit` (`section_editor_actions.go:215`) — use
  `scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh`.
- **A failed step shows `COMPLETED` with `error` NULL.** Read `__step_error`:
  ```sql
  SELECT key, left(value::text,400) FROM orchestration_states os, jsonb_each(os.collected_data)
  WHERE os.owner_agent_type='page-build-handler' AND os.current_step='complete_error'
    AND key LIKE '%error%' ORDER BY os.created_at DESC;
  ```
- **"Leave `cta_url` unset and the template renders no button" is FALSE.** A
  **render-time resolver** supplies one. Omitting it produced
  `<a href="/contact.html">Save everything</a>` — the notes-rescue button pointing
  at a contact form. Suppressing a CTA needs an explicit destination or the
  `cta_text` removed. **A destination that is not a real page renders no button**,
  which is how migrate sat safely while `/legacy` did not exist.
- **A new tool's `<script>` must open with a `tool-doc` header**
  (`/* === tool-doc ===` … `=== /tool-doc === */`, `platform/content/tool_doc_header.go`).
  `create_tool_component` refuses without it and creates nothing.
- **Tool page URLs are canonical, not chosen.** `CanonicalisePage(role="tool")` →
  `/tools/<bare>/index.html`, nested rather than flat because `siteUsesFlatURLs`
  reads `site_specs` aspect `structure` — **which this site does not have** — and
  defaults to nested. So content pages are flat (`/about.html`) and the tool page is
  nested. That is the framework's decision; **do not "tidy" it.**
- **Enumerate jsonb keys before joining on one.** `unresolved_cta` specs have
  `page_name`/`section_name` and **no `page_id`**; a draft of my SQL joined on
  `page_id`, would have matched zero rows and **committed successfully**. Every
  hand-written status change in this lane's SQL now asserts `ROW_COUNT`.
- **`site_work_items` has NOT NULL columns you will not guess** (`source`,
  `created_by`, …). To hand-file an item, copy an existing row's whole shape:
  `CREATE TEMP TABLE _tpl AS SELECT * FROM site_work_items WHERE id=<template>`,
  UPDATE the few fields, `INSERT INTO site_work_items SELECT * FROM _tpl`.
- **⚠ …and if you copy a row, RESET `page_id` — the handler reads the COLUMN, not
  your `spec`.** `site_work_items` has real `page_id` and `component_id` columns
  (`idx_swi_page`). I filed an assemble item for the tool page by copying the
  *contact* page's row and overriding only `spec`; the item ran, reported
  **`"success": true`**, and **re-deployed `/contact.html`**. Nothing said the target
  was wrong — the summary said "tool page", the spec said "tool page", and the
  deploy said `contact.html`. The tell is one join:
  ```sql
  SELECT wi.status, p.name AS page_the_column_points_at
  FROM site_work_items wi LEFT JOIN pages p ON p.id = wi.page_id
  WHERE wi.item_key = '<your key>';
  ```
  Harmless here (contact re-deployed identically), but on a destructive item type
  it would not be.
- **An unquoted heredoc will not survive a JSON/Python body.** Pass values through
  the environment and quote it (`<<'PYEOF'`), as `075_…` does.

**Carried forward, still true:**

- **`sites.github_repo='vm-sites'` is LOAD-BEARING SAFETY.** The default routes to
  `gqls/sites` → `b2 sync --delete` → **the prefix the live app serves from**. A
  build on the default repo would delete the running application.
- **A wildcard Worker route owns every subdomain.** `*.noted.co.uk/*` was answered
  before the tunnel was ever reached while every check on the box passed. Fixed with
  `app.noted.co.uk/*` → *(no worker)*.
- **`systemctl kill -s HUP cloudflared` TERMINATES it**, and the unit has no
  `Restart=`.
- **This box serves a live commercial shopfront.** Before/after control every time.
  The **box-local** check is the real one (`Host: webdesign.uk` → `127.0.0.1:8080`)
  because `webdesign.uk` externally 302s to `webdesign.co.uk` **by design**.
  ⚠ **Take the baseline yourself immediately before**: the handoff's recorded
  `28419 B` now reads `28015 B`, because that lane ships continuously. That drift is
  not a fault and is not ours.
- **`webdesign-chat` binds `*:8081`**, not loopback. `noted-engine` binds loopback.
- **B2 `writeFiles` includes HIDE** — always `--versions`.
- **The shell working directory persists between tool calls.** Use absolute paths.

## 7. Verify at the artefact — the checks that actually settle things

```bash
# the box (does it have the pages, and is the shopfront unharmed?)
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  'ls /var/www/noted.co.uk/ | head; \
   curl -s -o /dev/null -w "noted: %{http_code} %{size_download}\n" -H "Host: noted.co.uk" http://127.0.0.1:8082/; \
   curl -s -o /dev/null -w "shopfront: %{http_code} %{size_download}\n" -H "Host: webdesign.uk" http://127.0.0.1:8080/'

# the sparse cone MUST list every domain in sitesync's DOMAINS
ssh … 'git -C /var/lib/sitesync/repo sparse-checkout list'

# the live apex must still be the LEGACY app until cutover
curl -s -o /dev/null -D - https://noted.co.uk/ | grep -i 'x-amz-version-id'
curl -s https://noted.co.uk/ | grep -io 'being refreshed'

# the engine
curl https://app.noted.co.uk/api/health

# the rescue tool's probe (re-run after ANY edit to the html)
/home/ant/.venvs/vonc_pw/bin/python \
  docs/agent_docs/docs024_key_docs_latest/noted_rebuild/legacy_tool/test_legacy_rescue.py

# re-render CTA components from content_data (all | <page> | <page>|<slot>)
./scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh migrate
```

```sql
-- do the CTA buttons actually exist in the artefact?
SELECT p.name, pc.slot_name,
       (length(pc.rendered_html)-length(replace(pc.rendered_html,'<a ','')))/3 AS anchors
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='noted.co.uk' AND pc.slot_name IN ('hero','call-to-action')
ORDER BY p.name, pc.position;
```

## 8. What to do next, in order

1. **Resolve blocker 1** (tool page assembled?) and verify the tool page at the
   artefact — in the repo, on the box, and by re-running the Playwright probe
   **against the deployed URL**, not just the local file.
2. **Take blocker 3 to the owner** — the privacy copy. Blocker 2 is very likely
   waiting on it, and no agent may write it.
3. **Bind the two experience patterns** to the built components (`site_experiences`,
   `proposed → bound → verified`). Untouched this session.
4. **Exercise the degraded states by hand** (blocker 4) before any cutover.
5. **Re-run the IndexedDB origin probe at cutover.** Origin is scheme+host+port; a
   hostname change, a `www.` redirect or a scheme change silently invalidates the
   entire migration premise, and the failure is silent.
6. Cut over noted.co.uk from the bucket to the box, keeping the legacy app reachable
   for a grace period.

**Still true and still the rule:** `rebuild_policy` stays `generic`, **no `owned`
pages**, and every page goes through the framework — no hand-written HTML, however
small (owner ruling 2026-08-04).

## 9. Files

`docs/agent_docs/docs024_key_docs_latest/noted_rebuild/`

| file | what |
|---|---|
| `HANDOFF_2026-08-12_continue_here.md` | **this file** |
| `README_where_we_are.md` | owner's plain-prose log, append-only — **his document** |
| `NOTES_noted_rebuild.md` | technical log incl. every misstep; the 08-12 entries matter |
| `PLAN_2026-08-10_noted_rebuild.md` | design, decomposition ruling, phasing |
| `RUNBOOK_noted_rebuild.md` | commands, each with its gotcha |
| `CTA_2026-08-12_noted_cta_destinations.sql` | applied — CTA destinations, with a stale reasoning note corrected in-file |
| `legacy_tool/noted-legacy-rescue.html` | the rescue tool (markup + JS + tool-doc header) |
| `legacy_tool/test_legacy_rescue.py` | its Playwright probe, 24 checks, mutation-verified |
| `EXPERIENCES_2026-08-11_noted_patterns.sql` | applied — the two experience patterns |
| `box/` | everything on the box (`sitesync` is current with the box) |

Scripts added this session:
`scripts/initial_messages/001_assemble_all_pages_rerender/082_trigger_rerender_site_noted.sh`,
`scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh`,
`scripts/initial_messages/140_tool_suggester/075_create_noted_legacy_rescue_tool.sh`.

## 10. My error record this session — read before trusting anything above

Four wrong calls, all corrected in `NOTES` where they were made:

1. **"No k8s CronJob ⇒ the pump is undriven."** Wrong: `build-pipeline-trigger` had
   195 runs, driven by a `scheduled_tasks` row. An absent CronJob is not an absent
   scheduler.
2. **Measured drain with the *eligible* count**, which oscillates because a
   `claimed` item hides its whole site from the query. Implied 21 minutes against a
   true ~6h. The stable quantity is pending items **older than yours**.
3. **Filed a LANDMINES entry without grepping first** — the mortgagecalculator lane
   had already documented that exact trap, better. Removed my duplicate and merged
   back only what was new.
4. **"Leave `cta_url` unset and no button renders."** False, and it failed in the
   one direction I claimed to be guarding against. See §6.

The habits that caught these: measure the quantity that gates *you*, not fleet
health; verify at the artefact rather than the status; enumerate keys instead of
reading the one you expect; and mutate the code to prove a guard can fail.
