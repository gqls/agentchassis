# HANDOFF 2026-08-09 — webdesign.uk: the SITE IS BUILT, REVIEWED AND PARKED; next is the CHAT SERVICE (Phase 4)

> **SUPERSEDED same day by `HANDOFF_2026-08-09b_continue_here.md`** — Phase 4
> went from "blocked on the owner's key" to built, deployed and proven live
> through the tunnel within this session. Start from the `b` handoff; this one
> is history.

**Start here cold. Supersedes `HANDOFF_2026-08-08_continue_here.md`** (whose §1
drive-loop recipe is reproduced below because you will need it on the first
task). Read with: `PLAN_2026-08-04_webdesign_uk_vm_hosting.md` **§4 Phase 4 +
§5.1** (the chat service design and its controls — that is your brief) ·
`NOTES` 08-08 → 08-09 (five entries; the mechanisms, the wrong turns) ·
`SUMMARY_2026-08-08` (plain-prose milestone) · `README_where_we_are.md` (the
owner's log, newest at the bottom).

## 0. One paragraph

The five-page site is **built by the framework, serving, and the owner has
parked it** ("ok for now") after two review rounds: v1 was rejected, rebuilt
clean on 08-08, and on 08-09 the owner's three complaints were each verified at
the artefact and fixed — the missing home hero (a stranded asset, not a build
failure), the "hand-built" impression (false: full pipeline provenance), and
uncapped commercial terms (now **owner-confirmed: 2 revision rounds, 14-day
review window**, stated on every relevant page). **The next job is the CHAT
SERVICE** — Phase 4 of the VM plan, the one deliberately hand-written piece,
blocked only on the owner's scoped Anthropic key. After it: the input-box
section (Phase 5), which is what the 9 parked `unresolved_cta` items are
waiting for, then cutover (Phase 6) on the owner's approval.

## 1. State table — verified 2026-08-09, re-check before trusting

| thing | state | re-check |
|---|---|---|
| The site | **5 pages live on preview**: index, how-it-works, what-you-get, faq, contact. All assets resolve (5 heroes + logo + css + js) | `scripts`-free: run §6's verify script |
| Copy compliance | **0 visible-text hits across 18 ban patterns; 0 em dashes incl. titles** (verified at the SERVED pages, 08-09 midday) | §6 script |
| Commercial terms | **2 rounds / 14 days — OWNER-CONFIRMED 08-09**, `evidence_base.facts` sources say so, and **stated on index/how-it-works/what-you-get/faq** | curl + grep '14 days' |
| Box `webdesign.vs.mythic-beasts.com` (root, key `~/.ssh/webdesign_box_ed25519`) | live; nginx **127.0.0.1:8080 only**; ufw deny-in; sitesync 5-min; **8081 FREE for the chat service**; **NO Go toolchain on the box** (cross-compile + ship the binary) | `ss -tln`, `which go` |
| Tunnel `webdesign-box` `81f59f78…` | LIVE; `/etc/cloudflared/config.yml` routes apex+www+preview → 8080, preview overrides Host | `systemctl status cloudflared` |
| apex + www | **302 → webdesign.co.uk** via two host-scoped Page Rules (`6d4d5b67…`, `88794916…`). Restored 08-09 after a deliberate 108-min parking window | curl all three |
| CF token | works, expires **2026-09-01**; PageRules PATCH proven (zone `746f81e6…`). Session note: CF API writes trip the permission classifier — the owner approved them on 08-08 | pagerules GET |
| Work-item queue | **EMPTY except 9 `unresolved_cta` `needs_human_review`** — these ARE the input-box question, owner-gated; do not resolve mechanically | SELECT |
| Fleet dispatch | **still starved: no CronJob, fleet-FIFO, hundreds of older items.** Every stage is hand-driven (§2) | see §2 |

## 2. THE DRIVE LOOP — you will need this for anything pipeline-side

Nothing self-drives. Per stage: **find the item → claim it → orchestrate its
`handler_agent` directly → verify by payload → complete it**.

```sql
SELECT id, item_type, handler_agent, status FROM site_work_items
WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND status='triaged';
UPDATE site_work_items SET status='claimed', updated_at=now() WHERE id='<id>' AND status='triaged';
```
Envelope (kcat block exactly as `082_submit_domain_unified.sh`; `kcat -P -c 1`,
fresh UUIDs, `client_id=demo_client`, topic `system.agent.generic.requests`):
```json
{"action":"orchestrate","config":{"agent_type":"<handler_agent>"},
 "input_data":{"domain":"webdesign.uk","site_id":"1fcfa4f3-ec80-4010-878b-b971cd46711f",
   "work_item_id":"<id>","item_type":"<item_type>","source":"<source>",
   "spec":<spec>,"current_page":<spec>,"page_id":"<spec.page_id>","page_name":"<spec.page_name>"}}
```
Then `SELECT status, current_step FROM orchestration_states WHERE correlation_id='<corr>' AND parent_orchestration_id IS NULL;`
(**kcat exit 0 proves nothing**; and filter `parent_orchestration_id IS NULL` or
child spawns inflate every count). Blockers are in `agent_error_log`
(`error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`, `context->'issues'`), not in
orchestration rows. Worked examples with correlation ids: NOTES 08-08 (4) and
08-09 (1)–(5). Build scripts are in the scratchpad pattern; don't reuse a stale
envelope file inside a loop (cost me two wasted builds).

## 3. NEXT: Phase 4 — the chat service

**Brief: PLAN §4 Phase 4 + §5.1.** Summary of what is already decided, so you do
not re-litigate it:

- A small **Go, stdlib-first service on `127.0.0.1:8081`**, versioned in this
  lane's `box/` directory, deployed like the nginx/sitesync pieces already there
  (`box/setup-webdesignbox.sh`, `box/sitesync`, `box/webdesign.uk.nginx`).
  **The box has no Go toolchain** — cross-compile locally (`GOOS=linux
  GOARCH=amd64`) and ship the binary; decide and document the deploy step in the
  RUNBOOK as you go.
- **Sibling of `site-engine`, not an extension** (owner-level recommendation in
  the plan): it takes secrets and must be replaceable without touching the
  estate's capture binary. This is the SANCTIONED hand-written exception to the
  framework-only ruling — nothing generates backend code.
- Endpoints: `POST /api/chat`, `GET /health` (discovery check), later
  `POST /stripe/webhook`.
- **The §5.1 controls ship in the FIRST commit, not after:** (1) per-IP limit
  keyed on **`CF-Connecting-IP`** — through a tunnel `$remote_addr` is
  localhost, so a naive limiter is ONE GLOBAL BUCKET (`bugs_open/139`); prove it
  with the two-network `count(DISTINCT ip) > 1` check, because one machine
  cannot tell a constant from a working key. (2) hard turn cap per conversation.
  (3) per-day spend ceiling that **fails closed to the contact details**, not to
  an error page. (4) request log with tokens + cost per call. (5) transcripts as
  structured rows — they are the demand signal this phase exists to collect.
- Model **`claude-haiku-4-5`** (intake is not the product). The Fable pre-flight
  applies only when a paid build is wired, not here.
- **BLOCKED ON: the owner's scoped Anthropic key.** Ask for it before writing
  the call layer; everything else (controls, endpoints, deploy, health) can be
  built and tested against a fake provider first — and should be, so the key is
  needed only at the end.
- Then **Phase 5**: the input box as a **pinned section** in the site plan
  posting same-origin to `/api/chat`, with CTS-044 generation guards (external
  loader file, no inline script, `data-runtime-fill`). **Never ship it before
  the service exists.** Landing it resolves the 9 parked `unresolved_cta` items
  — that is their intended destination.

## 4. What is DONE — do not redo (and do not re-diagnose)

- **Provenance**: the site is framework-built. Every artefact has a pipeline
  commit and an orchestration; the claims gate blocked saves mid-build. If the
  owner says "looks hand-built" again, the tell is a silently-404ing asset
  (§5), not the pipeline.
- **Asset routing**: row-driven (`sites.github_repo='vm-sites'` →
  `resolveGitRepoNameDB`). The 08-04 assets went to `gqls/sites/webdesign.uk/`
  only because the row flipped after they deployed. **Tidy-up still owed:**
  delete that stale directory in `gqls/sites` — nothing serves it.
- **Caps**: mission_brief, identity, strategy, briefing, roadmap_brief,
  content_direction and `evidence_base` (facts + `writer_block` + 3 new bans)
  all carry the capped terms; the owner's ORIGINAL wording is preserved in
  superseded spec rows (`source='owner-caps-amendment-2026-08-09'` /
  `-confirmed-`). The `submission` spec deliberately still holds the original —
  it is the historical record of what was submitted.

## 5. Traps earned in this lane (all measured, all cost real time)

- **`evidence_base.facts[]` is bookkeeping; `writer_block` is the WIRE.** The
  writer's "Verified Facts" section renders from `writer_block` prose, NOT from
  `facts[]`. A fact registered only in `facts[]` never reaches a prompt, and
  nothing warns. Cost three full page rounds. **Check `llm_call_log.
  prompt_rendered`, not the orchestration's collected_data** — the spec rides
  along in memory and greps true while the prompt lacks it. (LANDMINES, filed.)
- **A page REBUILD resets the hero binding** (back to generic
  `/assets/images/hero.jpg`, a silent 404 with no broken-image icon) **and
  regenerates the TITLE** (re-introducing the banned em dash). After ANY page
  rebuild: re-check both, patch `page_components.content_data.background_image`
  + `rendered_html` and `pages.title`, then rerender. This is the class the
  owner spotted before I did.
- **A visitor check MUST extract `url(...)` from style attributes** — hrefs and
  srcs cannot see hero backgrounds. §6's script does.
- **Firecrawl caches 200s by default**; your own probe poisons what the pipeline
  then reads. The classifier's scrape now sends `max_age: 0` (kept). (LANDMINES.)
- **`build-dispatch-loop` under bare `action=orchestrate` no-ops silently** —
  COMPLETED, items untouched. Drive handlers directly instead. (LANDMINES.)
- **`asset-deployer` is DOWN**: both modes fail "storage client not available",
  which is why **no favicon has ever been produced** for this site. Use
  `image-build-handler` for images. The favicon needs the platform fix — a
  genuine open platform item, not lane work.
- The contact email is Cloudflare-obfuscated into `/cdn-cgi/l/email-protection`
  links that 404 under curl and work in a browser. Not a defect; don't chase it.

## 6. Verify like a visitor — the script that must pass before you call anything done

**`./docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/verify_served_site.sh`**
(committed 08-09; it reads the live `banned_claims` from the DB, so it does not
go stale as bans are added): fetch all five pages, sweep the **served visible
text** against every
live `banned_claims` pattern, check `<title>`s for em dashes, extract **every
`href`/`src`/`url(...)`** and curl each from the serving root. Expected clean
state today: 5×200, 0 ban hits, 0 title em dashes, only two known 404s
(`favicon.png`, `/cdn-cgi/l/email-protection*`). Anything else is new.

## 7. Owner ledger

**Owed by the owner:** the **scoped Anthropic key** (blocks the chat call
layer) · correction-fee number · written terms before live Stripe · final
review + cutover approval.
**Settled, do not reopen:** £1,200, no VAT · **2 revision rounds, 14-day review
window (08-09)** · Sonnet 5 on the writer · framework-only builds · one box per
trust class · box spec as billed · the site itself is "ok for now" (08-09) —
further copy polish is not the current job.

## 8. Access map

`~/.ssh/webdesign_box_ed25519` → root@box · `~/.config/cloudflare/token` (works,
expires 09-01; zone `746f81e6…`, rule ids `6d4d5b67…` apex / `88794916…` www) ·
`gh` = gqls, ADMIN on `vm-sites` · `b2` CLI · kubectl → cluster/DB
(`site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'`). **No Mythic panel, no
Stripe, no Cloudflare dashboard.**

---

> **CORRECTION APPENDED 2026-08-10 by the bugfix_239 lane (not this lane's thread — nothing
> above has been edited).**
>
> **The drive-loop `kcat` recipe in this file will silently do nothing if the JSON reaches
> stdin on more than one line.** `kcat -P` publishes **one message per line**, applying the
> same `-H` headers to each, so a pretty-printed heredoc envelope arrives as four to six
> invalid-JSON fragments — and the chassis (pre-fix) ran each one as the `generic` no-op and
> reported `COMPLETED` with an empty `execution_path`. That is the whole of `bugs_open/239`,
> whose root cause was found today from `chassis_intake_events` payload bytes: 8 of 8
> single-message sends resolved correctly, 10 of 10 fragmented sends did not.
>
> **So the `source`+`spec` trigger this lane bisected, and the "omit `source`" workaround,
> are both superseded** — omitting `source` worked only because it shortened the envelope
> enough to fit on one line.
>
> **Send it on ONE line:** `kcat ... <<<'{"action":"orchestrate",...}'` (here-string, single
> quotes), or pipe through `jq -c`. Never `<<JSON … JSON`.
>
> **Verify what arrived, not what you meant to send:**
> ```sql
> SELECT left(correlation_id::text,8), count(*) AS msgs FROM chassis_intake_events
> WHERE kind='request' AND correlation_id::text LIKE '<corr>%' GROUP BY 1;
> ```
> `msgs > 1` ⇒ fragments. And keep this lane's own rule: check `owner_agent_type` equals the
> agent you asked for, never just `status`.
>
> The chassis half is fixed (fail-closed refusal, `DISPATCH_FAIL_CLOSED`) but **inert until
> the fleet rolls** — `strings /app/agent-chassis | grep -c DISPATCH_FAIL_CLOSED` says
> whether it is live yet. The send-side rule above applies either way.
