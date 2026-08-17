# HANDOFF — restore /services.html and finish the 08-12 findings (leopardessconsulting.co.uk)

**Purpose.** Start a fresh session from exactly here. Self-contained: everything below was
measured on 2026-08-12/14 by the session that wrote it, with dates on each claim. The wider
site context is `HANDOFF.md` §11 (same directory) — read §11 first if you have never
touched this site; read only this file if you have.

**The one-paragraph version.** Between 2026-08-08 and 08-12 the fleet's automated loops
ran across this site (~190 orchestrations on 08-11/12 alone). They built useful things
(five guide pages, tool card imagery) and they regressed `/services.html` in **five ways**,
all silent, all found by checking the artefact rather than by any alert. This handoff is
the repair list for that page, then the rest of the 08-12 findings in priority order.
**Nothing has been repaired yet** — the 08-12 session was documentation-only, and the
08-14 session only added the forensics below.

**Site:** `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Page:** `/services.html` · `page_id = ebc2c413-61e2-465e-b22b-9aab0167abc9`
**Pre-repair snapshot (take your own diff against it when done):**
`scripts/SNAPSHOT_2026-08-14_services_pc_pre_restore.json` — all 4 slots' `content_data`
as of 2026-08-14, before any repair.

---

## 0. Read these THREE warnings before running anything

1. **Every figure in this file decays.** This site is inside the automated improvement
   loops now; the 08-11 pass is proof they rewrite content. Re-verify each regression at
   the served page before repairing it — one may have been repaired (or worsened) by
   another pass between this file being written and you reading it.
   `curl -s https://leopardessconsulting.co.uk/services.html` is the instrument; the
   specific assertions are inline below.
2. **Your repair may not survive, and that is expected — verify anyway, then re-verify
   after the next fleet roll.** The damage class is `bugs_open/238` (a regeneration
   replaces `content_data` wholesale and drops resolver-sourced keys). Its fix ("the
   carry") **has rolled** (live since chassis `v1.0.1291`; current is ≥ `v1.0.1297`) —
   **but 238 §9.2 (2026-08-12) shows a live regeneration lost renderer-sourced keys
   anyway**: `source:"renderer"` fields short-circuit `sourceResolver.resolve` to
   `(nil, true)`, so they never look "missing" and the carry never runs. A `090`
   diagnosis of that hole was filed (`97ef39f0`). Until it closes, any regeneration of
   this page can re-drop what you restore. That is a reason to verify and to date your
   verification, not a reason to skip the repair.
3. **On webdesign.uk, restoring `content_data` + a `page_rerender` was NOT sufficient**
   (238 §9.2, same date) — the buttons only came back after hand-splicing
   `rendered_html`. **Do not assume that transfers here**: this site's proven path is
   different (hand-edit `content_data` + a no-LLM `section_data_resolved` rerender via
   `scripts/rerender_page_safe.sh`, which worked for exactly these fields on 07-31). If
   your rerender does NOT surface the restored fields: **stop, record it as a
   confirmation of 238 §9.2's hole on a second site, and contribute it to
   `bugs_open/238` before deciding anything** — that observation is worth more than the
   repair. If you then hand-splice `rendered_html` to ship, say so in the bug file and
   in RUNNING_NOTES: it re-arms the `bugs_open/229` divergence detector by construction.

---

## 1. The five regressions on /services.html — evidence and repair, in execution order

All four components on this page have `updated_at = 2026-08-11 18:15:23Z`. The trigger
was a `page_rerender` work item ("1 misdirected CTA(s) on services"), which evidently
escalated to full regeneration — the teaser items' keys, prose and CTA all changed, which
a merge-only rerender cannot do. `[MEASURED 2026-08-12]`, mechanism attribution
`[INFERRED]` — do not re-assert it as fact without checking `bugs_open/238`/`268` first.

### Order matters

Do all `content_data` edits (1a–1e below) **first**, run the escalation-guard pre-checks,
then **ONE** rerender, then verify everything at the served page in one pass. Not one
rerender per fix — each rerender is a fresh chance for the loops' next pass to interleave.

### 1a. Six images gone (worst visible damage)

**Evidence** `[MEASURED 2026-08-12, re-check before repairing]`: the served page has
exactly **one** `<img>` tag. All six `teaser-reveal-panel` items carry `image_url: ""`;
item 6 (`news-credibility`) has lost `image_url`/`image_alt`/`open_label` entirely.

**The images are NOT lost.** All six files are live and distinct
`[MEASURED 2026-08-14]`:

| path | bytes |
|---|---|
| `/assets/images/icon-service-monitoring.jpg` | 41,443 |
| `/assets/images/icon-service-orchestration.jpg` | 37,372 |
| `/assets/images/icon-service-oversight.jpg` | 26,535 |
| `/assets/images/icon-service-verification.jpg` | 30,605 |
| `/assets/images/icon-service-toolbuild.jpg` | 26,155 |
| `/assets/images/icon-service-siteops.jpg` | 46,716 |

(The `assets` table's rows for these six all carry the placeholder URL
`/assets/images/input-data.asset-key.jpg` — that is `bugs_open/248` (the
**undeployed_asset slug** — ⚠ the number is shared with an unrelated CTA bug, resolve by
slug) plus `bugs_open/152`. The **files** deployed correctly on 07-31; only the DB
metadata is wrong. Do not "fix" the asset rows as part of this repair — contribute the
leopardess numbers to 248 instead, §3.)

**The complication: the regeneration also rewrote the item KEYS**, so the icons no longer
map 1:1. Current items `[MEASURED 2026-08-12]`: `verification-pipeline`,
`hierarchical-orchestration`, `human-oversight`, `decision-record`, `model-routing`,
`news-credibility`. Four map cleanly:

| item | icon |
|---|---|
| `verification-pipeline` | `icon-service-verification.jpg` (drawn: a token emerging double-outlined past a register) |
| `hierarchical-orchestration` | `icon-service-orchestration.jpg` |
| `human-oversight` | `icon-service-oversight.jpg` |
| `decision-record` | `icon-service-monitoring.jpg` — plausible but **LOOK AT IT FIRST** |

`model-routing` and `news-credibility` have no natural icon — the 07-31 set was drawn for
`toolbuild`/`siteops` items that no longer exist. **Recommended:** wire the four that fit,
generate two new icons for the two new items via the proven Route-A recipe (scope-less
`needs_imagery` → image-build-handler, `kind:'icon'` → Banana; the exact prompts and the
style constraints that were needed — "outlines only, no tick marks, no cross or X marks" —
are in `RUNNING_NOTES.md` 2026-07-31). **The site rule is absolute: look at every
generated image by eye before wiring it.** On 07-31, two of six were rejected on sight
(a confirmation drawn as an X; solid-white fills fighting the gold linework).

**Do NOT wire anything to `/assets/images/input-data.asset-key.jpg`** — several distinct
assets resolve to that one path; what it serves today is whatever was deployed to it last
(probably an 08-08 tool hero, not an icon).

Repair SQL shape (adjust items 5–6 to your generated asset paths; `image_alt` must never
restate the `hook` — schema rule):

```sql
UPDATE page_components SET content_data = jsonb_set(content_data, '{items}', (
  SELECT jsonb_agg(
    CASE f.elem->>'key'
      WHEN 'verification-pipeline'      THEN f.elem || '{"image_url":"/assets/images/icon-service-verification.jpg","image_alt":"<write one>"}'::jsonb
      WHEN 'hierarchical-orchestration' THEN f.elem || '{"image_url":"/assets/images/icon-service-orchestration.jpg","image_alt":"<write one>"}'::jsonb
      WHEN 'human-oversight'            THEN f.elem || '{"image_url":"/assets/images/icon-service-oversight.jpg","image_alt":"<write one>"}'::jsonb
      WHEN 'decision-record'            THEN f.elem || '{"image_url":"/assets/images/icon-service-monitoring.jpg","image_alt":"<write one>"}'::jsonb
      WHEN 'model-routing'              THEN f.elem || '{"image_url":"<new asset>","image_alt":"<write one>","open_label":"<check siblings>"}'::jsonb
      WHEN 'news-credibility'           THEN f.elem || '{"image_url":"<new asset>","image_alt":"<write one>","open_label":"<check siblings>"}'::jsonb
      ELSE f.elem END ORDER BY f.idx)
  FROM jsonb_array_elements(content_data->'items') WITH ORDINALITY f(elem, idx)))
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name='teaser-reveal-panel';
```

Item 6 also needs `open_label` restored — read a sibling item's value for the expected
shape before inventing one.

### 1b. The Mistral claim is back (worst-in-kind: verified false, previously removed)

**Evidence**: item 5 (`model-routing`) body reads *"A workflow step can call Claude,
Gemini, Mistral or another provider, or run down a self-hosted Ollama path…"*.
`platform/aiservice/factory.go:23-33` supports **anthropic, ollama, gemini**; `openai` is
a stub returning "not yet implemented"; everything else hits `default:` unsupported.
`[MEASURED 2026-08-12; the factory re-read 2026-08-12]`. This same claim was removed from
this same page on 07-31 (`AUDIT_verified_facts.md` "Re-measurement 2026-07-31").

Replacement body, keeping the item's true second half:

> A workflow step can call Anthropic's Claude or Google's Gemini, or run down a
> self-hosted Ollama path that never leaves the cluster. That distinction matters when a
> step touches data you would not send to an external provider.

(Adjust the tail to whatever the live body's second sentence actually says when you read
it — the quote above splices today's fragments; read the full body first.)

⚠ **Check the whole regenerated payload while you are in there.** The regeneration also
wrote *"checked more than 2,000 businesses"* (true — 3,419 on 07-31, durably phrased) and
*"over 9,545 items collected, more than 8,297 credib[le]"* (item 6 — **unverified**; the
07-31 measurement had 7,990 items / 6,794 scored, so 9,545 is plausible growth but was
never checked; the table is `content_feed_items`). Verify or re-phrase durably before the
rerender ships them again. Rule from `AUDIT_verified_facts.md`: no claim without a row.

**Do NOT "fix" the guide page the same way.** `/guides/llm-cost-calculator-guide.html`
says the *calculator covers* OpenAI/Claude/Gemini/Llama/Mistral/Cohere **pricing** — that
is a claim about what the tool compares, not about what the platform can call
`[MEASURED 2026-08-14]`. It is only false if the tool does not actually cover them: drive
the tool and check its provider list; fix whichever of the two is wrong.

### 1c. The CTA is misdirected again — label says conversation, target is a tool

**Evidence** `[MEASURED 2026-08-14]`, live anchors:

```
<a href="/tools/tool-agent-complexity-estimator.html" class="cta-btn cta-btn-primary">Book an architecture conversation</a>
<a href="/tools/automation-savings-estimator/index.html" class="cta-btn cta-btn-secondary">Not ready to talk? Use our AI Automation Time Savings Estimator to see how this applies before you get in touch.</a>
```

On 07-31 this was authored as primary "Get in touch" → `/contact.html`, secondary → the
ROI estimator with a label naming it. A "book a conversation" label pointing at a
calculator is `bugs_open/248`'s **CTA** case (a CTA recompute silently overwrites an
authored `/contact.html` link — again: two unrelated bugs share number 248, resolve by
slug) — and, pointedly, the work item that triggered the 08-11 rerender was itself
*"1 misdirected CTA(s) on services"*. Whether the recompute failed to fix it or created
it is `[UNVERIFIED]` — worth one paragraph contributed to the 248-CTA file either way.

```sql
UPDATE page_components SET content_data = content_data || jsonb_build_object(
  'primary_cta',       'Get in touch',
  'primary_cta_url',   '/contact.html',
  'secondary_cta',     'Not ready to talk? Estimate the return on an AI agent first.',
  'secondary_cta_url', '/tools/ai-agent-roi-estimator.html')
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name='call-to-action';
```

(Secondary label is a suggestion — anything that names the tool it reaches. The 07-31
rule: a label must not promise a different interaction than its target delivers.)

### 1d. One card link 404s

**Evidence** `[MEASURED 2026-08-14]`: `info-card-grid` card index 1 ("Data checked before
it's trusted") links `/case-study-automated-intelligence-pipeline.html` — a `pages` row
created 2026-08-11 16:21 with `build_status='planned'`, never deployed, **404 live**.
`RepairPageLinks` cannot strip it: the row exists, so the link passes its test.

Repoint to `/case-studies.html` (the page that actually describes the Companies House
verification pipeline — same reasoning as the 07-31 repointing):

```sql
UPDATE page_components SET content_data =
  jsonb_set(content_data, '{cards,1,link_url}', '"/case-studies.html"')
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name='info-card-grid';
```

⚠ The planned page row is still there, and `bugs_open/266` (archived pages rebuilt by
four producers, none reads `page.status`) means it may deploy later, and the internal
linker that repointed these cards on 08-11 may repoint them again at its next pass. Your
repair is correct **now**; note in RUNNING_NOTES that if the case-study page ever
deploys, the card becomes a candidate to point there deliberately.

### 1e. The carousel flag

**Evidence** `[MEASURED 2026-08-12]`: `carousel` is absent from the `info-card-grid`
instance's `content_data`; the served page has no `data-hcc-*` markup. The canonical
template **still carries the carousel arm** (`data-hcc-carousel` present,
`[MEASURED 2026-08-14]` — its md5 has moved since 07-31, `204a3975…` vs the L9 file's
`f99b791c…`, so another lane touched the template; gate on the arm's presence, **not** on
md5 equality with the L9 file). `snippets.js` is still 13,781 bytes with both snippets,
and `js_snippets['hero-card-carousel'].applies_to` should still contain
`info-card-grid` — verify with one SELECT; if some lane reverted it, re-run L9 §2 and
re-fire `site-asset-renderer`.

```sql
UPDATE page_components SET content_data = content_data || '{"carousel": true}'::jsonb
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name='info-card-grid';
```

---

## 2. The rerender and the verification — the part that is NOT optional

**Pre-rerender gate — run BOTH queries from `scripts/L9_services_carousels.sql` §4**
(every slot a non-empty object; every required `source:"llm"` field present). Both must
return 0 rows **or the whole page escalates to the LLM writer and your hand-authored
repairs are regenerated away in the same breath.** This is not theoretical: it is
probably how the 08-11 damage happened.

Then ONE rerender:

```
./scripts/rerender_page_safe.sh 4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk services
```

(Takes the **page_id/site_id**, not names, and confirms its Kafka publish — the old
inline kcat route lost ~4 in 5 attempts silently. Do not regress to it.)

**Verify at the served page — poll on MARKUP, never on a class name** (a class name
matches the component's own inline `<style>` and returns instantly against a stale page —
this cost a wasted round on 07-31):

```bash
U="https://leopardessconsulting.co.uk/services.html?cb=$(date +%s%N)"
curl -s "$U" > /tmp/svc.html
grep -c '<img'                                       /tmp/svc.html   # expect 7 (6 icons + hero) — was 1
grep -c 'data-hcc-carousel'                          /tmp/svc.html   # expect ≥1 — was 0
grep -c 'href="/contact.html" class="cta-btn cta-btn-primary"' /tmp/svc.html  # expect 1
grep -c 'case-study-automated-intelligence-pipeline' /tmp/svc.html   # expect 0
grep -ci 'mistral'                                   /tmp/svc.html   # expect 0
grep -c 'info-card-grid__card-link'                  /tmp/svc.html   # expect 6 (survivors of link repair = targets real)
```

Then the real-gesture probe: drive the page in headless Chromium **with
`--force-prefers-reduced-motion`** (a smooth scroll under `--virtual-time-budget` never
advances and reports a working carousel as dead — this exact mistake was made and caught
on 07-31), click both carousels' arrows, and run the no-init **mutant** to prove the probe
can fail. The working probe script pattern is in `RUNNING_NOTES.md` 2026-07-31
("Verification: 19 served-page assertions, then a real-gesture probe with two mutants").

**Date the verification in RUNNING_NOTES.** On this site a pass is a pass *on a date*,
not a durable state — that is the §11.7 standing question, unchanged.

---

## 3. After the page is whole — contributions owed, then the backlog

In order:

1. **Contribute to `bugs_open/248` (undeployed_asset slug)**: leopardess holds 15 asset
   rows on the placeholder URL (6 `icon_service_*` 07-31, 6 `content_hero_tool_*` 08-08,
   +3 older); fleet-wide 76 rows / 12 sites, 2026-01-28 → 2026-08-11
   `[MEASURED 2026-08-12]`. The bug is CONFIRMED (090) and OPEN — contribute, do not
   re-file, and check `scripts/who-owns.py` first.
2. **Contribute to `bugs_open/248` (CTA slug)**: the 1c case — an authored
   `/contact.html` primary replaced by a tool link on a page whose trigger item was
   itself "misdirected CTA". Also read `bugs_open/268` (a `content_rewrite` drops CTA
   destination keys, 214 buttons fleet-wide, filed 08-12) before writing — your case may
   belong there instead; 268's mechanism is explicitly NOT established yet, and a clean
   second-site observation is exactly what it needs.
3. **Contribute to `bugs_open/152`**: two leopardess asset rows have regressed to
   presigned S3 URLs since the 07-29 cleanup (`hero_case_studies` re-created 08-08,
   `content_hero_tool_automation_savings_estimator` 08-11) `[MEASURED 2026-08-12]` —
   confirms 152 recurs on every generation, as filed.
4. **Sweep the stale figures off `/case-studies.html`** (HANDOFF.md §11.5): live page
   says 143/56 agent definitions (actual 193/187 on 08-12), "75,061 orchestration state
   records" (table is pruned hourly at 24h; held 5,997 — same defect class as the 90,790
   claim fixed on 07-31; **no cumulative total belongs on any page, rephrase durably**),
   "eight live sites" (definition-dependent; re-derive from the 07-31 audit's definition,
   do not just count `sites` rows — that is 40 and means something else). Update
   `AUDIT_verified_facts.md` in the same pass; then drain the 3 `claims_unverified`
   items rather than leaving them at `needs_human_review` for another fortnight.
5. **The voice work** (HANDOFF.md §11.6): 33 `voice_tells` items unactioned; ~12 pages
   still carrying the owner-banned "honest". Method, measurement and traps:
   `docs/agent_docs/docs024_key_docs_latest/fleet_copy_quality/CONTRIB_2026-08-12_the_honest_ban_and_the_voice_gate_nobody_opted_into.md`.
   Its key trap: assert on **shape** after replacement (double commas, dangling
   articles), not just on the word being gone.
6. **`tool-process-automation-scorer` acceptance failure**: 7 passed / 2 failed
   (`submit-shows-error` at both viewports, `fix_cycles_spent: 0`), surfaced 08-08,
   untouched. Fire `scripts/tool_acceptance_run.sh` (rewritten 08-11 — read its header:
   three things must line up, two fail quietly) to reproduce before fixing.
7. **sitemap.xml** still lacks the 4 trust-series pages (unchanged since §10.4).

---

## 4. Quick reference

| thing | value |
|---|---|
| site_id | `4851f6fc-71cf-4160-a270-e03d6d3e0732` |
| /services.html page_id | `ebc2c413-61e2-465e-b22b-9aab0167abc9` |
| info-card-grid component_id | `fc56f085-8e9a-4f6b-8e8d-600f9a1381e2` |
| teaser-reveal-panel component_id | `22c12251-73aa-4232-bd67-ef9edcfe8061` |
| current site_plan spec id (aspect = section authority; `site_plan_sections` has 0 rows here) | `8439e6b2-e671-44c1-a63a-c90175ed59c6` |
| pre-repair snapshot | `scripts/SNAPSHOT_2026-08-14_services_pc_pre_restore.json` |
| the 07-31 recipe this repair re-runs | `scripts/L9_services_carousels.sql` (+ its in-file gotchas: md5 not length(), the three-place placement rule, the pre-rerender gate) |
| DB access | `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db` |
| backups | `bak_*` tables before ANY change (site rule); note `page_components.id` is NOT stable across rerenders — back up and re-query by `(page_id, slot_name)` |

**Done-when:** all six §2 grep assertions pass on a cache-busted fetch; both carousels
move under a real click and the no-init mutant kills both; the three contributions (§3
items 1–3) are written into their bug files; RUNNING_NOTES and README_where_we_are carry
dated entries; the commit is pathspec-scoped and names this file.


---

## EXECUTED 2026-08-14 ~18:15–19:00Z — do not re-run this repair

All of §1 (1a–1e), §2 and §3 items 1–3 were carried out and verified on 2026-08-14 by the
services-restore session. All six §2 grep assertions pass (note: the `card-links` grep
counts **17 lines** because the component's inline CSS matches the string — the real
anchor count is 6, which is what the assertion means); both carousels move under a real
click and the no-init mutant kills both; the three contributions are in their bug files;
`bak_leo_services_pc_20260814` holds the pre-repair rows (alongside the JSON snapshot).
Two new icons were generated for `model-routing`/`news-credibility`
(`icon_service_routing`/`icon_service_credibility`, both eyeballed and accepted). The
§1b sweep also removed one claim this file did not enumerate: item 4's pruned-table
count + "weeks after the fact" retention promise (see `AUDIT_verified_facts.md`
2026-08-14).

**Still open from this file:** §3 items 4–7 (case-studies stale figures, voice work,
process-automation-scorer acceptance, sitemap). **Standing re-check:** the §0.2 warning
stands — re-run the §2 assertions after the next fleet roll or any regeneration touching
this page; whether the 268 fix (`8f899cc8d`, committed 09:13 BST 08-14) is in the
running chassis was not provable from this session (v1.0.1299; sha-probe negative,
provenance line scrolled). Full narrative: `RUNNING_NOTES.md` 2026-08-14.

> **UPDATE 2026-08-16:** re-check done after the roll to v1.0.1303 — the repair survived
> untouched (served page byte-identical to 08-14, all six assertions pass) AND the 268 fix
> (`8f899cc8d`) is proven in the running chassis (ancestor of the pod's stamp `5e075a6f9`,
> stamp sha present in `/proc/1/exe` with a random-hex absent-control). The §0.2 warning is
> retired for this page from that roll onward. Remaining from this file: §3 items 4–7.

> **CONTINUE HERE (2026-08-16 ~11:30Z, session checkpointed on usage limit).** §3 items 1–4
> DONE and live (item 4: `/case-studies.html` swept against the LIVE `evidence_base` register,
> infographic re-drawn number-free as `_v4`, both `claims_unverified` items handled — detail
> RUNNING_NOTES 2026-08-16, commits `8fc8c9acb` `4ac4b799f` `3ca88ae8d`). Remaining, in order:
> 1. ~~**§3 item 7 — sitemap.xml**~~ **DONE AND LIVE 2026-08-16 ~15:52Z** — 27 → **36** `<loc>`,
>    served file byte-identical to the generated one, 36/36 probe 200, negative control (the
>    never-deployed `/case-study-automated-intelligence-pipeline.html`) 404 and absent, 0 archived
>    leaks. Detail: RUNNING_NOTES 2026-08-16 (afternoon).
>    > **CORRECTED 2026-08-16 — "No platform generator exists" was TRUE at turn 18 (2026-07-17)
>    > and FALSE from 2026-07-28.** It is `scripts/site-discovery-files.py`, concept register
>    > **SEO-002**: `robots.txt` + `sitemap.xml` + `llms.txt` for any site from the `pages` table,
>    > dry-run by default, probes every URL and lists only 200s. Do not hand-roll the turn-18
>    > recipe. `retract_page_deployment_action.go` was a dead end — its only mention of
>    > sitemap.xml is a comment saying it must NEVER touch it. Logged in `WRONG_CALLS.md`.
>    Two carry-forwards, both deliberate, neither done: the tool also emits **`llms.txt`**
>    (6,970 B, built from each page's own `<h1>`/first sentence) — a NEW file for this site, so
>    out of that item's scope and available as a costed next step; and **`robots.txt`** is
>    Cloudflare-managed here (the tool detects the merge; ClaudeBot/GPTBot/CCBot et al. are
>    disallowed at the edge), so shipping ours changes nothing until a dashboard setting is off
>    — an owner's call, not a session's.
> 2. ~~**§3 item 6 — `tool-process-automation-scorer` acceptance**~~ **DEFECT ALREADY FIXED —
>    PROVEN BY HAND 2026-08-17.** The premise was stale: the failing check was Tier 2's STATIC
>    arm (`check_tool_acceptance.go` only confirms an `interaction`'s anchors exist in the
>    served HTML), the `improve_tool` item closed 2026-08-11, and `#pas-error` /`.pas-submit`
>    are both on the deployed page today. Driven live in headless Chromium: empty form + click
>    → the error shows; all nine answered + click → it hides and a score appears (opposite
>    result on the opposite branch = the probe discriminates). Probe: `scripts/cdp.py` +
>    `scripts/probe_pas.py`, a stdlib CDP client — there is no node/puppeteer/websocket module on this
>    box. The platform's own Tier-4 run is RAISED AND QUEUED as work item
>    `fcfbdfd5-a1f5-427d-962f-8caaf82ea145`; **leave it, it claims itself when the endpoint
>    below recovers.** All three `tool_acceptance_run.sh` preflights were checked, including
>    the pod-grep (browser-runner `v1.0.1305` carries the arms, negative control absent).
> 3. **§3 item 5 — voice work: MEASURED, and STOPPED AT AN OWNER QUESTION — do not just
>    grind the queue.** Re-measure before acting; the handoff's "33 items / ~12 pages with
>    'honest'" is stale. Today `[2026-08-17]`: **34 open items / 210 findings**, of which
>    **banned_phrase is 104**; and running the site's own 14 regexes over all 36 served pages
>    gives **138 of 145 hits on the single pattern `\btrust(ed|worthy|s)?\b`** — which now
>    flags the site's OWN product name ("The AI Vendor Trust Checklist"), quoted research
>    titles and other people's statistics, because the trust content pillar was built AFTER
>    the 2026-07-18 rule. Much of the queue is unsatisfiable by construction. **That is the
>    owner's rule to narrow, not a session's.** Census: `scripts/voice_census.py`, method
>    in `fleet_copy_quality/CONTRIB_2026-08-12_the_honest_ban…md` (§4: a mechanical ban-list
>    is a smell, not a crime), full working in RUNNING_NOTES 2026-08-17.
>    The part needing no ruling is **written and NOT applied**:
>    `scripts/VOICE_2026-08-17_banned_phrases_ready.sql` (2 × "earns its keep", 1 ×
>    self-labelling "honestly"), with backups, a `DO`/`RAISE` verify block, the three
>    rerenders and the served-page assertions + control. ⚠ Its whole reason for existing:
>    `use-cases-list.use_cases` is `source: site_specs.portfolio.use_cases`, so that edit
>    must go to the ASPECT — a `content_data` edit reads back fine and is reverted by the
>    rerender you fire to publish it (LANDMINES "A `site_specs.<aspect>.<path>`-sourced
>    field…"; that entry's verifier run `05d0454c` — check `doc_notes` categories ?
>    'landmine-verification').
>
> **⚠ BLOCKING EVERYTHING FLEET-WIDE, 2026-08-17 11:08Z — the Anthropic endpoint hit its
> ACCOUNT SPEND CAP:** `ai_endpoint_health` says `healthy=f`, *"You have reached your
> specified API usage limits. You will regain access on 2026-09-01 at 00:00 UTC"*, confirmed
> independently in `llm_call_log` (4 real failures, `council-gate` + `landmine-verifier`).
> `claim_work_item` then releases every claim whose handler uses that endpoint, so items sit
> at `triaged` with `attempt_count 0`, the dispatch loop COMPLETES every ~90s, and there is
> **no error anywhere** — the reason is only in `collected_data->'claim_result'`. 26 items
> blocked within the hour. **Owner action (provider console); nothing here retries into it.**
> New LANDMINES entry covers it, plus the second-order effect that
> `find_dispatchable_site` is `ORDER BY created_at ASC … LIMIT 1`, one site per tick, so one
> unclaimable item starves the whole fleet (webdesign.co.uk took 18 of 18 runs in an hour).
> **Do not fire rerenders or dispatches until `ai_endpoint_health.healthy` is true again:**
> `SELECT endpoint_url, healthy, error FROM ai_endpoint_health;`
>
> **New in this directory:** `scripts/commit_site_file.sh` — commit any file into a site's
> directory in the deploy repo via the git-adapter, with a `PUBLISH_OK` receipt and the repo
> read from `sites.github_repo`. Use it in preference to `commit_brand_assets.sh`, whose
> `kubectl run -i … kcat -P < file` publish drops ~4 in 5 at exit 0.
