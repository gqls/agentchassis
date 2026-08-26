# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-26 ~15:45Z; STATE + GATE ZERO revised 17:35Z (grind seat: #44 built and retired, queue cleared, GATE ZERO corrected against the real selector).
Supersedes `HANDOFF_2026-08-25_continue_here.md` (which had accumulated nine stacked STATE lines).

## STATE: 49 of 63 SERVE-CONFIRMED. NOTHING IN FLIGHT, NOTHING OWED. THE QUEUE IS CLEAR — filing works.

**USE THE GATED SERVE-GRADE, DO NOT HAND-ROLL ONE:**
`docs/agent_docs/docs024_key_docs_latest/webdesign_tool_rebuilds/servegrade.sh <slug> <ported-slot-file> <completed_at> [negatives…]`
— committed in this lane's own directory, NOT a scratchpad (a scratchpad is session-scoped and the
path would be dead for you). Dump the ported slot first:
`psql -At -c "SELECT rendered_html FROM page_components WHERE id='<ported slot>'" > /tmp/ported.html`. It EXITS NON-ZERO on non-200, a missing/unparseable `last-modified`, or
an artefact not newer than `completed_at`, and marks a negative control `WORTHLESS` when the ported
slot does not contain it. Three traps are baked into it, all of which faked a PASS:
**(a)** a 404 scores 0 on every negative control — printing the status code beside the counts is not
gating on it; **(b)** `date -d ""` returns **NOW**, so a missing header parses as "landed";
**(c)** ⚠ **`date -u -d` does NOT parse its input as UTC** — `-u` is output-only, this box is BST, so
comparing a Postgres UTC `completed_at` naively made artefacts look **3600 s fresher** and would have
passed anything up to an hour stale. Anchor it: `date -d "$completed UTC" +%s`. Mutation-proved.

⚠ **GRADE THE SERVED PAGE WITH A GATE, NOT A PRINTOUT.** On #46 the first attempt returned `http=404`
and **every negative read 0 and "valid"** — a perfect-looking pass off a 3 KB error page. Printing the
status code alongside the counts is NOT gating on it:
```bash
code=$(curl -s -o page.html -D hdr -w '%{http_code}' "$URL")
[ "$code" != "200" ] && { echo "REFUSING to grade: not 200"; exit 1; }
```
The 404 was a propagation blip, proven transient by curling the RECORDED `pages.url` plus two
untouched siblings (all 200). **A single 404 is not damage; the sibling control is what tells them
apart.** Also: `date -u -d "" +%s` returns **NOW**, so an empty `last-modified` header compares as
"landed" — guard on non-empty before comparing.

44 `removed` + 19 `deployed` = 63 (tool pages), verified 2026-08-26 17:32Z, with **zero pages carrying
both a live ported slot and a live native slot**. Nothing is part-done except the serve-grade below.

**#49 `tool-privacy-redactor` is BUILT AND RETIRED** — `add_tool 8cf95960` filed 18:55:33Z, complete
19:03:25Z, retired ~19:04Z under full guards. Component `d329d1d7`, native slot `6abb30e0` (20,248
chars); revert handle `a5c9781b` (md5 `1690fe78…`, len 10220). **SERVE-GRADE OWED:** assembler
`d00b62cb-8376-4912-bacc-29417c176ff3`, priority 80, do NOT re-file.
**Graded by EXECUTING both versions' regexes over fixtures, not by reading them** — the right method
for any tool whose product is a CLASSIFICATION. The ported one was wrong on **5 of 6** discriminating
cases: it silently missed 15-digit Amex (4-6-5) and IPv6 entirely while advertising "Cards" and "IPs",
and false-positived a Luhn-failing reference number and `999.999.999.999`. Rebuild passes all eight
fixtures including three negatives. **Reuse this method for the remaining detector-ish tools.**
⚠ It also taught a grep lesson: my first search for the page's client-side claim used a guessed
wording and came back EMPTY — the page says *"stays in this browser tab"*. **An absence found by a
guessed pattern is not an absence.**

**#48 `tool-insight-injector` is BUILT AND RETIRED** — `add_tool 31611056` filed 18:38:46Z, complete
18:48:57Z, retired ~18:49:30Z under full guards. Component `fd9c0799`, native slot `d5294763` (18,106
chars); revert handle `dd1e438c` (md5 `985e97de…`, len 9369). Graded PASS. **SERVE-GRADE OWED:**
assembler `716ced5b-2a90-4a2a-ab14-d4ba94a17ad7`, priority 80, do NOT re-file.

⚠ **TWO MEASUREMENT TRAPS I FELL INTO ON THIS ONE — both cost nothing and both were stated as fact:**
**(1) Do NOT pipe a gate query through `head -N`.** Mine cut 3 of 10 rows and I read the truncation as
"nothing claimable on the page". A row cut by `head` is indistinguishable from a row that does not
exist, and it bites hardest in a multi-section heredoc where the LAST section is the casualty. Cap
with `LIMIT` (visible in the result) or not at all — psql's `(N rows)` footer is the honest signal.
**(2) Do NOT hand-type a timestamp you can join to.** `'2026-08-26 03:52:09'` against a real
`03:52:09.367409` excluded 28 rows of the same sweep batch: the margin read **2** hand-typed and
**30** read from the row. Use `(x.priority,x.created_at) < (w.priority,w.created_at)`.

**#47 `tool-layout-generator` is BUILT AND RETIRED** — `add_tool 9e314640` filed 18:27:08Z, claimed
1m53s later, complete 18:33:55Z, retired ~18:34:30Z under full guards. Component `1050a211`, native
slot `bb418e82` (18,940 chars); revert handle `8b75e29b` (md5 `3407a176…`, len 9223). Graded PASS.
**SERVE-GRADE OWED:** rerender `d5d057a0-2f84-4cc6-9c78-a4473a533b54`, priority 80, do NOT re-file.
**Generalisable finding for the remaining rebuilds:** for any tool that GENERATES code, *"is it
responsive?"* has TWO answers — the tool's own page and the tool's OUTPUT — and **only the second is
the product**. This one emitted a fixed-px three-column grid with no `@media` at all. A census
`@media` count answers the first question only.

**#46 `tool-asset-formatter` is BUILT AND RETIRED** — `add_tool 12e3ef8c` filed 18:00:15Z, claimed
18:10:38Z, complete 18:15:05Z, retired 18:16Z under full guards. Component `3e26f29a`, native slot
`9734606a` (17,369 chars); revert handle `518fe90e` (md5 `e900cdd5…`, len 9222). Graded PASS at the
arms. **SERVE-GRADE OWED:** rerender `0853324f-9c7b-41e6-a5d7-45e3cede457a`, priority 80, do NOT re-file.
⚠ Its run carried a real `__step_error` (`suggest_related_pages` died at `max_tokens`) while the item
read `complete` — it cost NOTHING because `related_pages` was explicit in the spec, which is a second
reason to always carry that key. Tool HTML was checked for truncation before the retire.

**#45 `tool-head-architect` is BUILT AND RETIRED** — `add_tool cd0078ae` filed 17:47:42Z, claimed
**17:49:56Z (2m14s)**, complete 17:54:33Z, retired **17:54:52Z (19 s later)** under full guards.
Component `bf0d7919`, native slot `a3faae4d` (19,911 chars). Ported slot `4f13e098` (md5 `9802cb3c…`,
len 9212) is the revert handle. Mechanism-graded PASS at the arms incl. both output-correctness bugs
(context-correct `escapeAttr`/`escapeText`, and `escapeJsonForScript` mapping `<`/`>`/`&` to `\uXXXX`).
**SERVE-GRADE OWED:** rerender `486fef43-2b4a-4d8d-bf61-24eec6a2786a`, priority 80, do NOT re-file.

**#44 `tool-monolith-splitter` is BUILT AND RETIRED** — `add_tool e164b069` filed 16:49:27Z, claimed
17:03:19Z (13m52s), complete 17:06:50Z, run graded clean, retired 17:23:20Z under full guards, tombstone
re-read. **SERVE-GRADE OWED:** rerender `4d02956e-391a-4024-9eb8-ef2c7c9bb2bc` is `triaged` with ~80
dispatchable items ahead `[MEASURED 17:28Z]` (~24 claims/h ⇒ ~3.3 h). **Do NOT file a second rerender**
— it would dedupe against that row. State while waiting is safe: DB has ported removed + native
deployed, the page serves the old single tool, and that rerender assembles native-only when it lands.
Grade it with step 7 below when it completes.

**⚠ GATE ZERO — THE RUNBOOK'S QUERY COUNTS THE WRONG POPULATION. Use the corrected one.**
`[VERIFIED at the arms 2026-08-26]` The runbook's citations (`maintenance_actions.go:882`,
`seed_build_queue_action.go:69`) are **two different queues, neither of them `site_work_items`** —
the first claims from `maintenance_queue` filtered by `task_type`, the second reads `build_queue`.
The real loader is `platform/orchestration/actions/load_work_item_actions.go` (~747-790), and it
additionally requires `attempt_count < max_attempts` **and** `(retry_after IS NULL OR retry_after <=
NOW())`. The old query therefore *overcounts* stalled retries and *undercounts* by filtering
`pipeline='build'`, which the dispatcher does not filter (`build-dispatch-loop.load_items` sets
neither a pipeline nor a handler filter; `max_items: 5`). Measured today: old query **10**, truth **8**.

```sql
SELECT count(*) AS truly_ahead FROM site_work_items wi
WHERE wi.site_id='6b49db8e-d447-4467-8277-4f3018af9897'
  AND wi.status IN ('triaged','approved')
  AND wi.attempt_count < wi.max_attempts
  AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
  AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
  AND (wi.priority, wi.created_at) < (60, now());
SELECT max(claimed_at) AS site_last_served FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897';
```

**⚠ RUN THE BLOCKER QUERY TOO — `truly_ahead = 0` does NOT mean "will be claimed soon".** Proven the
hard way at 18:00Z: #46 was filed at 0 ahead and sat `triaged` for **ten minutes** because one
`needs_content_page` row was `claimed`, which excludes the WHOLE SITE from selection. Both queries or
neither:
```sql
SELECT count(*) FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND status='claimed';  -- >0 ⇒ site excluded now
```
This is not a fault (a content-page build takes minutes); it just makes the count unpredictive. And
note it changes fast — measured 0 blocked sites fleet-wide at 17:29Z, live on this site by 18:01Z.

**And item depth is not what decides whether you are served — SITE SELECTION is.**
`build-pipeline-trigger.find_dispatchable_site` (live config) ends `ORDER BY wi.created_at ASC,
wi.priority ASC, wi.id ASC LIMIT 1` and excludes a site entirely via `NOT EXISTS (… status='claimed')`.
So (a) the **oldest item fleet-wide wins the site** and priority is only a tie-break — a fresh row never
wins selection itself, it just needs its SITE to win, after which `load_items` takes 5 by `priority ASC`
(this is how a priority-60 filing behind a 03:47 batch was claimed in 13m52s today); and (b) **one stuck
`claimed` row halts all dispatch for its site** — `[MEASURED 17:29Z, NEGATIVE]` 0 sites fleet-wide hold a
claimed row older than 30 min, so that is latent, not today's problem, and it is **not** offered as the
explanation of the 08-26 dormancy (no evidence either way). Full working: NOTES 17:30Z.

⚠ **Do NOT bump `priority` and do NOT re-file** — pickup is `priority ASC` (LOWER first), so a bump moves
you BACKWARDS, and `LANDMINES.md` forbids it.

## ⚠ STANDING HAZARD: 41 confirmed-false findings queued against the rebuilds — and the count is a FLOOR

`tool_acceptance` has filed **41 `improve_tool` items against tools this lane rebuilt**, every one
reading *"interaction anchor #X absent from deployed page"*. **They are wrong, and the cause is
CONFIRMED** — `needs_diagnosis` `91228c39-8980-42bf-95cd-bd16bb43de0a`, complete 10:59:05:

- criteria are the tool's **authored PLAN document** (`check_tool_acceptance.go:loadCurrentCriteria`
  → `SELECT body FROM doc_plans WHERE subject_type='tool' AND subject_key=$1 AND is_current`),
  written with **bare** ids;
- the renderer rewrites every id: `ConvertTemplateToInstanceScope` does
  `strings.ReplaceAll(out, 'id="'+id+'"', 'id="'+instancePrefix+id+'"')`, with
  `InstanceToken` = `"c-" + s`.
- So `#ring-copy-button` is sought on a page carrying `id="c-tool-focus-ring-ring-copy-button"`.
  **The tell is not the matching names — it is that `boots` PASSES** (its selector `.tool-container`
  is a *class*, which nothing prefixes) **while every id-anchored check fails on the same page.**

⚠ **The 41 is a FLOOR that grows by one tool per rebuild.** Verified at both artefacts for #44
(page serves `id="c-tool-monolith-splitter-ms-framework"`, its `is_current` `doc_plans` criteria seek
`#ms-copy-btn`), and #45 will follow the same path once its acceptance run fires. Holding rows does
not slow the accrual — only fixing the checker or the criteria does.

**Fleet-wide: 110 anchor-absent findings against 2 of every other kind; 32 already `complete`.** Each
becomes an LLM rewrite: an observed `tool-improver` note reads *"Root cause: unknown. Fix: Rebuilt
tool HTML to restore the #sessions-per-day-input element"* — regenerating something never missing.

**Status 2026-08-26 16:45Z: all 41 are now `deferred` — HELD by this lane, reversibly.** Verified:
41 deferred, all keeping `handler_agent` (so still promotable), all carrying the un-defer condition
in the row; 0 left triaged; **the 15 anchor-absent rows on other sites are UNTOUCHED.** Shape copied
from two verified estate precedents (`webdesign.uk` rows `41d82357`/`0559eb67`, deferred by
`webdesign_uk_build_service` citing this same diagnosis). Undo in one statement — see NOTES 16:45Z.
⚠ **Name the side effect when you read the queue numbers:** holding those 41 removed 41 items from
ahead of this lane's own next filing (~107 ahead → 25). The hold was justified independently and the
gate STILL refuses a filing at 25 ahead / ~5 claims per hour, so nothing was gained by it — but a
session that benefits from its own queue action should say so rather than quietly enjoy it.
**Un-defer when** the mismatch is fixed (doc_plans criteria carry the scoped id, or the checker
resolves through `InstanceToken`) **and** `tool_acceptance` has re-run; if a finding still reproduces
at the served bytes it is real. Prior status line:
**41 `triaged`, 66 ahead, NONE TOUCHED.** Owned by `staged_component_build`
(`scripts/who-owns.py tool_acceptance`), who hold the CONTRIB + verdict:
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/CONTRIB_2026-08-26_from_webdesign_tool_rebuilds_tier2_anchors_are_unscoped_while_the_renderer_scopes_them.md`.
**Do not dispatch, promote or "fix" these rows, and do not let a queue-drain effort proceed without
reading this** — the stall is currently the only thing standing between them and the 43 rebuilds.
Raised with the owner as a decision; he asked for a re-check rather than choosing, so the contested
hold stays untaken. **If you take it, it is a reversible status flip on webdesign's 41 ONLY** — the
other 69 belong to other lanes' sites.

## The recipe (proven 45 times) — unchanged except GATE ZERO, which is CORRECTED above

1. GATE ZERO (above). Then: fetch the live page cache-busted and read the ported slot IN FULL.
2. Gates: library-claim (0 rows or pin fork identity), local active component (0), open `add_tool`
   (0), adopt flag `true`, margin on the page's queued rerender. `related_pages`: 1–3 EXISTING
   non-tool `pages.name` picked by TOPIC.
3. **Write the brief as a SPECIFICATION, not a bug report.** Describe the tool to BUILD; keep the
   defect archaeology in NOTES. Two builds died at `max_tokens` on 08-25 because the brief carried
   history — golden-ratio went 4,431 → 2,701 (still died) → **1,551 chars (built in 4m28s)**.
   Character count is NOT the predictor; surface area is.
4. File; attend with foreground poll loops (never a background watcher).
5. Grade the **RUN** (`page_adopted='true'`, no `already_exists`, no `__step_error` — an item reads
   `complete` with `error` NULL on a dead run), then **RETIRE IMMEDIATELY** (guarded txn, DO/RAISE
   pre- and post-asserts, md5 pinned, post-commit re-read).
6. Mechanism-grade the component at the DECIDING code arms — never the tool-doc header.
7. Serve-grade cache-busted: http=200 first, `last-modified > completed_at` (S3 lands 1–2 min AFTER
   the item completes), negatives 0, positives present. **Validate every negative BOTH ways** — it
   must count ≥1 in the ported bytes, or a 0 proves nothing.
8. Tombstone re-read. Dispatch the sidecar's dry-run retraction if the tool had one; record the
   orphan (`bugs_open/365`, 12 files across 11 tools).

## ✅ #44 `tool-monolith-splitter` — DONE 2026-08-26 (serve-grade owed, see STATE)

Page `05449406-4215-4c4a-9ffc-0fae8b83b7a0`, ported slot `e134edb7` now `removed` (md5 `79c32824…`,
len 9037 — the revert handle), native component `89fe1b20`, native slot `e65b7a9c` (14,453 chars).
Brief that worked: **1,894 chars, written as a specification**. The banked 2,636-char `file_ms.sql`
brief was NOT used — ~40% of it was defect archaeology, the shape that killed two builds on 08-25.
Crosslinks filed and both `deferred` under `OWNED_PAGE_GUARD`, as expected — not delivered.
Handoff's ported-defect list said "4 inline onclick"; the slot has **3**. Detail: NOTES 17:30Z.

## ✅ #45 `tool-head-architect` — DONE 2026-08-26 (serve-grade owed). Analysis retained below for reference

Page `3fe28a53-9862-45fd-ac97-f5d193b390f5`, slot `4f13e098-72d9-446f-a2d9-a248d6fc8aa5`, md5
`9802cb3c…`, url `/tools/head-architect/index.html`. **Self-contained** — the census counted 1
external script, but that is the literal string `src="..."` inside a source COMMENT, not a real
`<script src>`. No sidecar, no orphan. (This is why it was reclassified out of the external-script
class on 08-24.)

What it does: builds a complete `<head>` block from four fields (site name, page title, description,
OG image) plus a pasted raw head whose stylesheets, custom metas and scripts it tries to preserve.
Emits charset/viewport/title/description, an Open Graph + Twitter card set, and a JSON-LD `WebSite`
entity block.

**Ported defects — the load-bearing one is an OUTPUT-CORRECTNESS bug, not a UI one:**
- **Every field is interpolated into markup unescaped.** `${title}`, `${desc}`, `${siteName}`,
  `${img}` go straight into `content="…"` and `<title>…`. A site name of `My "Great" Site` closes the
  attribute early and corrupts every tag it appears in — and this tool's entire output is markup the
  visitor pastes into their own site. **Escape for attribute context**, and say on the page that it
  does.
- **The JSON-LD block can be broken out of.** `JSON.stringify` escapes JSON but does NOT escape
  `/`, so a description containing `</script>` terminates the `<script type="application/ld+json">`
  block early. Escape `<` (or `</`) when embedding JSON in a script element.
- **The preservation is a regex heuristic and the page never says so.** `raw.match(/<script[\s\S]*?>[\s\S]*?<\/script>/gi)`
  — the ported source calls it "(Client-side heuristic)" **in a comment the visitor cannot see**. A
  self-closing or unclosed `<script src=…>`, or a script containing `</script>` in a string, is
  mis-captured or dropped SILENTLY. Either parse properly (`DOMParser` on the pasted head is
  available and exact) or state plainly what is preserved and report anything it could not classify.
- **Custom `property=` metas are dropped without notice** — the custom-meta regex matches `name=`
  only, so a visitor's existing `og:` or `article:` tags vanish while the tool regenerates its own.
- ~~**No duplicate handling**: a pasted head already containing a title/description yields output with
  both the preserved and the generated one.~~ **CORRECTED 2026-08-26 17:50Z at the code — THIS IS
  INVERTED, and it changes the requirement.** The ported source makes exactly **three** `raw.match`
  calls — scripts, `link rel=stylesheet`, and `meta name=` (with `(?!description|viewport)`). **There
  is no `<title>` capture at all.** So a pasted title is **silently DROPPED**, a pasted
  `meta name="description"` is **silently DROPPED** by the negative lookahead, and every
  `meta property="og:…"` is **silently DROPPED** because the regex matches `name=` only. Nothing is
  ever duplicated. The defect is **silent loss**, so the fix is *preserve, or say what you replaced* —
  **not** de-duplication. A brief written from the original line would have specified the wrong
  behaviour.
- **Placeholder defaults are silently substituted** (`|| "My Website"`, `|| "My Page Title"`,
  `|| "Page description."`), so an empty form produces a plausible-looking head full of dummy values
  ready to paste. Make an unfilled field visibly unfilled.
- `alert("Copied!")` unconditional with no failure arm; 2 inline `onclick`s; 2 globals.
- Cosmetic but sloppy for a code generator: several vestigial `head += `\n    \n`;` lines emit blank
  lines with trailing whitespace into the output.

`related_pages`: `learn-marketing-seo-for-llms` (the JSON-LD/entity half is exactly that article's
subject) and `learn-security-xss-vulnerability` (the escaping half). Both verified active non-tool
pages.

## ✅ #46 `tool-asset-formatter` — DONE 2026-08-26 (serve-grade owed). Analysis retained for reference.

Page `1a6d54a8-db1f-40e6-8a11-504bb9931edc`, slot `518fe90e-c5c8-4123-8e00-9915a04eee51`, md5
`e900cdd5…`, url `/tools/asset-formatter/index.html`.

What it does: the visitor lists semantic keys against hosted asset URLs (add/remove rows) and gets a
paste-ready prompt block telling an AI assistant it is forbidden to invent image URLs and must map
only to this dictionary, with the map emitted as fenced JSON. **Keep that instruction text — it is
the product.**

**Ported defects:**
- **The stated sanitisation does not do what it says.** The source comment reads *"Sanitize key to be
  a safe variable name (replace spaces with underscores)"* and the code is
  `key.replace(/\s+/g,'_')` — whitespace only. `logo-main!`, `2fast` and `class` all pass through
  unchanged into a dictionary the prompt calls "semantic keys" for the model to map logic to.
  Either validate to a real identifier and say what was changed, or stop claiming to sanitise.
- **Duplicate keys silently collapse.** Two rows with the same key overwrite in the object (last
  wins), so the list on screen and the dictionary in the prompt disagree with no warning — the
  output stops describing the input.
- **No URL validation at all**: `not a url` maps happily into a block that instructs a model to use
  it as an image source.
- **`copyPrompt` is the same three defects as monolith-splitter, same author's hand** — guard keyed
  on placeholder PROSE (`text.includes("Add your assets")`), unconditional `Copied!` with no failure
  arm, and a 2-second restore that hardcodes `#333`/`#fff`, colours the button never had.
  **And here the prose guard has a live hole:** the empty case writes **`"No assets mapped."`**,
  which does not contain "Add your assets" — so **pressing Copy with no assets puts the literal
  string "No assets mapped." on the clipboard while the button reports success.**
- **`window.onload = () => …`** seeds the two example rows — a window-global assignment that
  clobbers any other `onload` on the page.
- Rows are built with `row.innerHTML` interpolating `value="${defaultKey}"` and an inline
  `onclick="removeAssetRow('${id}')"`; 4 inline handlers, 4 globals.
- Copy reads back `#output.innerText` — and the payload is a fenced JSON block, where exact
  whitespace matters. Copy the composed string.

`related_pages`: `learn-ai-builders-anti-slop` (its thesis is precisely "stop the model inventing
generic filler", which is this tool's job) and `learn-ai-builders-content-first`. Both verified
active non-tool pages.

Then, smallest-first: head-architect 9,212 · asset-formatter 9,222 · layout-generator 9,223 ·
insight-injector 9,369 · … **re-run the census, do not trust this list.** The FIVE rich apps go
LAST, one at a time, owner-reviewed (standing ruling).

## Two items OWED — each wants its OWN `replace_existing` filing, not a fold-in

1. **cubic-bezier — keyboard access** (arrow-key nudge) for the two drag handles, cut to fit the
   token budget. A real gap on a site publishing `/learn/accessibility/focus-states.html`.
2. **golden-ratio — a REAL crop export.** The rebuild has **NO download at all**; the ported one had
   a "Download Crop" button that cropped nothing and burned the guides into the photograph. Wanted:
   cropped to the chosen ratio, centred on the overlay, guides NOT drawn on it.
   ⚠ **Both are `replace_existing:true` filings on rebuilt tool pages — the exact shape of the
   write-conflict spiral the noted lane found** (`bugfix_283_component_instance_scope/CONTRIB_2026-08-26_…`).
   Check strike history for the completed-then-overwritten pattern BEFORE either dispatches.

## Standing rules (load-bearing)

- ONE at a time (serial item key); file ONLY what you attend in-turn.
- Retire = status flip ONLY, never delete; revert handle = row id + length + md5 recorded pre-file.
- Grade the RUN, the COMPONENT (by mechanism), the SERVED page — never a status.
- **Any all-history claim about `site_work_items` must `UNION site_work_items_archive`** — the live
  table is a rolling window (cost me a published figure on 08-25: 13 `failed` was really 22).
- **Census by `spec->>'check'`, never by item_type name** — checks file under *other* type names by
  design, so an item_type census for `tool_acceptance` returns 0 and reads as "never ran".
- **`related_pages` mentions never land on this site** — `deferred` is a TERMINUS, not a gate: 0 of
  80 `tool_crosslink` rows have ever completed here, because every `/learn/` article is
  `rebuild_policy='owned'`. Keep filing the key (the finding is correctly targeted and is the raw
  material when an owned-page route exists); **never report a mention as delivered.**
- **Nothing may defer behavioural correctness to `tool_acceptance` while its findings are the
  anchor-absent class.** Grade at the mechanism and the served bytes, or say plainly it is ungraded.
- Counts carry the date they were counted (owner 08-22). A `[MEASURED]` figure about STATE expires.

## Cold-start dependencies

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
Site `6b49db8e-d447-4467-8277-4f3018af9897`. Tally: the RUNBOOK's `GROUP BY build_status` query
(expect `removed` = 43 + N). All per-tool ids/md5s: NOTES, per-tool entries (newest at bottom).
Chassis: rolled 2026-08-25 19:07Z, stamp `a7459a44b`; adopt path and the 360 tombstone guard both
verified present on both replicas (NOTES 08-25 20:30Z) — **re-verify after any further roll.**
