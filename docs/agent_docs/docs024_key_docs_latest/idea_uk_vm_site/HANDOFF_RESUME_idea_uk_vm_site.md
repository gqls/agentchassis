# RESUME HANDOFF — idea.uk VM site (start a fresh chat here)

> ## ▶ START HERE — state as of 2026-07-26 (supersedes everything below)
>
> **The content pipeline is BUILT AND LIVE, the box is healthy, and the extended report is now
> PROVEN END TO END** (2026-07-26: owner submitted, received and reviewed a real report in the new
> format — evidence below). **The top job is now the second deploy** (automatic order expiry), so
> the queue can never silently close itself again.

> ## ⚠️ CONCURRENT-SESSION NOTICE — added 2026-07-26 18:45 UTC (session "idea.uk vm site 6")
>
> **1. THE SECOND DEPLOY IS DONE — TWICE.** I built and deployed at **18:29 UTC**, before reading
> this block: I was working from the 15:34 version of this file and did not see the 15:57 rewrite.
> That binary carried the expiry work plus `bugs_closed/089`. A **second deploy at 18:44 UTC**
> added `bugs_closed/090`, which did not exist as a fix at 18:29 — I found it at 18:32 *while
> verifying 089 on the box*.
>
> **That near-miss is the useful part**: re-probing after the 18:29 deploy still returned the
> forged address, which is the only reason 090 was not written up as shipped when it was not.
> **The deploy that closes a bug is not the deploy that happened to precede it.** So check the
> marker each fix introduces, never the fact that "a deploy happened":
>
> ```bash
> grep -ac "refused fake=1" /opt/idea/idea   # 089 → 1
> grep -ac "X-Real-IP"      /opt/idea/idea   # 090 → 1   (0 in BOTH earlier binaries)
> ```
>
> Rollbacks kept: `/opt/idea/idea.prev-2026-07-25` (pre-everything) and
> `/opt/idea/idea.prev-2026-07-26-089only`. Orders backed up before each
> (`orders.json.bak-2026-07-26-predeploy`, `…-predeploy2`). Verified after both restarts: unit
> active, queue answering, all orders intact.
>
> **Not yet deployed:** three copy defects in the report renderer, fixed and committed after the
> above (doubled full stop, a score line reading "out of 5 —" with no number, a sentence-cased
> field spliced mid-sentence). They run at engine time, so they need a deploy to take effect but
> nothing is urgent about them — let them ride the next one.
>
> **2. ~~AN ENGINE RUN IS IN FLIGHT~~ — LANDED 18:40 UTC, restarting is safe again.**
> `ord_1785090638951163875` ran in **9.5 minutes** (18:31:13 → 18:40:46, 10,166 chars), not the
> 20–30 this file predicted. It is now `awaiting_payment` with a live Stripe `cs_live_` session
> and **is waiting on the owner to pay £29** — that is the last leg of the chain, and the only
> one that has still never run in production. The order survived a second restart intact
> (status, report and session id all verified after it).
>
> **The warning was real while it stood**, and the reason it stood is worth keeping: fulfilment is
> an in-memory goroutine, so a restart strands an order in `running`, and `ExpireStale`
> deliberately skips `running` — the slot then leaks permanently and only a hand-edit frees it.
> See item 4.
>
> **3. THIS RUN IS PARTLY DUPLICATIVE AND I OWN THAT.** The 12:40 run already proved the report
> FORMAT; I fired a second one before the 15:57 rewrite reached me. It is not wasted — the 12:40
> order was **declined**, so `approve → pay link → Stripe payment → delivery` has still never run
> in production. This one is being taken through exactly that leg, which is the only part of the
> chain still unproven. Logged in `WRONG_CALLS.md`.
>
> **4. Residual worth someone's attention:** `ExpireStale` sweeps `awaiting_review` and
> `awaiting_payment` but skips `running` — correct while a run is genuinely in flight, wrong after
> a restart, because no goroutine survives one. An order left `running` by a restart holds a slot
> for ever, which is the exact failure the expiry work was written to end. Cheap fix: at startup
> (before the first sweep), any order still `running` cannot be, so reset it to `requested` or
> `expired`. NOT built — deliberately out of scope while a run was in flight.

### What idea.uk is, today (all curl-verified 2026-07-26)

| | |
|---|---|
| **Guides** | 9, live, in journey order on a self-populating hub: creating-ideas → building-it → testing-it → user-acceptance → feedback-loops → patents → copyright → funding-ways → funding-sources |
| **Tools** | 4 cards on `/tools.html`: the £29 Verified Idea Report · "Should you patent it?" (free) · "Which funding route fits?" (free) · Free Audience Check |
| **Paid tool** | Extended and DEPLOYED (binary built 07-25 15:11): the report now leads with an assessment of the submitted idea, carries "Check it yourself" source links, discloses AI use in the report itself, and renders an honest "too early to assess" outcome |
| **Box** | `116.203.204.115`, service `idea` restarted 07-26 13:19, queue **open (0/5)**, `OPERATOR_EMAIL=idea-uk@leopardess.uk`, no `CONTACT_EMAIL` line (correct — see §Email) |
| **Locks** | 27 authored sections locked; both hub listings deliberately unlocked so they keep deriving |

Site id `1244516d-014d-421c-88c6-090bb1e9552a`. SQL applied this arc: `sql/p4_01`…`p4_19`.

### ✅ DONE 2026-07-26 — the new report format is PROVEN in production

Owner submitted a real idea, confirmed it, received the draft, judged it good, and declined it
(declining is the right way to close a test without self-charging; it releases the slot and emails
the requester politely at no cost).

Verified from the stored order, not from impression — `ord_1785069609860726188`, 13,227 chars of
text / 20,207 of HTML, every new-format marker present and the old-format giveaway absent:

| Marker | |
|---|---|
| `YOUR IDEA, ASSESSED` / `Your idea, assessed` (text + HTML) | ✅ |
| `Check it yourself` source lists | ✅ — 16 links in the HTML |
| `A considered next step` | ✅ |
| `We use AI to research…` (disclosure in the report, not just the T&Cs) | ✅ |
| `FURTHER IDEAS WORTH PURSUING` (ideation half, retitled) | ✅ |
| Old intro `"You asked us to find AI product ideas for…"` | ✅ **absent** |

`too early to assess` did not appear — correct, the submitted idea was assessable. That branch is
covered by tests (`service_test.go`), not by this run; a deliberately vague submission would
exercise it if you ever want the live proof.

**So the whole chain works**: request → operator confirm → step-0 assessment with live web search
→ cross-vendor cut → verify → score → draft to operator → decision. Two things worth watching on
the next few real runs: wall-clock (two long search passes now) and spend (~2× per report).

### TOP JOB NOW — a deploy is written, tested and waiting

Committed, `go build`/`go vet` clean, full suite green — **inert until the owner builds and
deploys** (the tool has no CI). Contents:

- **Automatic order expiry** — `Store.ExpireStale` + `App.sweepStale()` at startup, hourly, and
  before `/capacity` answers. `STALE_REVIEW_DAYS` (14) / `STALE_PAYMENT_DAYS` (7); `0` disables.
  New terminal status **`expired`**, distinct from `declined`, retaining the row.
- **`OPERATOR_EMAIL` as single source of truth** — `reportContact()` no longer reads an env var
  directly with a hardcoded fallback; it is wired from config at `NewApp`.
- Tests: `expire_stale_test.go` (3 cases: releases the right ones only, CreatedAt fallback for
  legacy rows, disabled thresholds are a no-op).

```bash
cd docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
GOOS=linux GOARCH=amd64 go build -o idea .
scp idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'systemctl stop idea; mv /opt/idea/idea.new /opt/idea/idea; systemctl start idea; sleep 2; systemctl is-active idea; curl -s http://127.0.0.1:8080/capacity'
```

### THE TRAPS — read these before touching anything

1. **NEVER edit `/var/lib/idea/orders.json` under a running service.** It is read ONCE at startup
   and rewritten wholesale from memory on every order change. An edit while running is invisible
   to the process AND gets clobbered by the next request. Cost two failed attempts on 07-26.
   Always: `systemctl stop idea` → edit → `systemctl start idea` → `curl /capacity`.
   Corollary: **`systemctl start` on an already-active unit is a no-op** — use `restart`, or
   stop-then-start. This is what made the second attempt look like the first had failed.
2. **Never lock a section whose component schema has a `query.*` source.** Locks are row-granular;
   `SavePageSectionsAction` re-attaches locked rows verbatim, so locking a derived listing freezes
   it while every render still reports success. Nearly killed the self-populating guides hub on
   07-25. Every lock script since carries a guard that refuses. **This now bites for real**:
   `bugs_open/058` (the lock gate) went CLOSED & LIVE on **v1.0.1165** today — locks are enforced,
   so editing a locked page means unlocking first.
3. **`page_components.slot_name` must equal `content_components.function`** — the renderer keys
   its component lookup on `slot_name`, not `component_id`. NULL ⇒ every section is "carried"
   ⇒ nothing renders ⇒ **the job still reports COMPLETED**. See RUNBOOK Phase 5.
4. **`pages.sections` is not written by the rerender path** — backfill it, or the page is invisible
   to `ListedPageEligibilitySQL` and the imagery sweep.
5. **"No orchestration row" means QUEUED, not dropped.** Check whether *anyone's* orchestrations
   are starting before re-firing. Latency ranged from <1 min to ~12 min this week.
6. **Verify against the live page by curl, never the job status or the DB.** VM sitesync is a
   5-minute timer on top of the render. Every "verified" claim in these docs names the curl.
7. **A schema field with `source: static` AND a `fallback` is UNOVERRIDABLE** — the fallback is
   written into resolved_data unconditionally and resolved_data merges last. Hit twice
   (`guide-list`, `tool-list`); both fixed by dropping the fallback after verifying no-op.

### Email — settled, do not re-litigate

`OPERATOR_EMAIL=idea-uk@leopardess.uk` is **correct for the tool**. The site and the tool
deliberately use different addresses; an earlier claim that leopardess was "stale" was a site fact
wrongly widened to the tool (logged in `WRONG_CALLS.md`). The `CONTACT_EMAIL` line has been
removed and that is right: today's binary falls back to the correct hardcoded address, and the
queued deploy makes it resolve from `OPERATOR_EMAIL` properly.

### Open decisions for the owner

- **Margin**: each report now makes 6 model calls including two long web-search passes — roughly
  double the previous spend, at the same £29. Worth revisiting after a few real runs.
- **Ceiling**: `MAX_ACTIVE_ORDERS=5` is a deliberate throttle on spend and operator attention.
  Once the funnel produces real volume, is 5 still right?
- **Report copy** now *undersells* slightly (it mentions both halves but the further-ideas half
  briefly). Safe direction; revisit only if conversion suggests it.

### Optional housekeeping, no urgency

- ~60 spam `requested` rows from 2026-06-11 (the "elBd" injection strings) still sit in
  orders.json. They hold no slots. The /request hardening (honeypot, timing, limiter) went live
  with the 07-25 deploy, so the flood should not recur.
- Two SES custom MAIL-FROM DNS records outstanding since 07-18 (deliverability, not blocking).

### Where the record lives

- `RUNNING_NOTES_idea_uk_vm_site.md` §X.12–§X.18 — the technical log for this arc, including every
  misstep and correction.
- `README_where_we_are.md` — the plain-prose owner log.
- `SUMMARY_2026-07-25` → `SUMMARY_2026-07-26` → `SUMMARY_2026-07-26b` — how the understanding moved.
- `AUDIT_2026-07-25_paid_tool_vs_copy.md` — what the paid tool does vs what the page claimed, the
  owner's ruling, and what was built.
- `RUNBOOK_idea_uk_vm_site.md` **Phase 5** — the repeatable recipe for adding a guide or tool.
- `features_open/014` — the pipeline vision and its build log; `015` — the fleet-wide ladder.

### If you are adding the next guide or tool

Follow RUNBOOK Phase 5 verbatim; it is now proven ten times. The hubs list new pages
automatically (`query.pages_where_type:guide|tool`) — no hub edit needed, just a re-render after
the page ships. Content policy from `014`: stages 6–9 (patents/copyright/funding) stay
hand-authored until claims-verification V5 is live; stages 1–5 may take generated copy. Any new
tool where one answer can be decisive must **gate before it scores** — see both existing finders.


> ## ▶ PREVIOUS STATE (2026-07-22) — superseded by the block above, kept for history
>
> **The migration is DONE and LIVE, and `/bugs_open/018` (the broken chrome) is FIXED AND VERIFIED
> LIVE.** idea.uk now serves the chassis static site with full navigable chrome and a working free
> tool, on one origin. Nothing on the live site is currently broken by a defect this workstream owns.
>
> **What was fixed since the last handoff (all live, all verified against the deployed page — not the
> work-item status):**
> 1. **`018` — chrome links all `href=""` → 0.** Root cause was NOT the theory 018 guessed: the chrome
>    renderer (`render_site_components_action.go`) fills templates from a HARDCODED value map and never
>    reads `input_schema`; idea.uk's two per-site components declared other field names, so every one
>    resolved empty, and the templates had no `{{if}}` gates so each empty became a visible dead link.
>    Fixed by `sql/p3_01` (rewrote both templates against the real vocabulary, gated every anchor, set
>    `sites.logo_url`) + `sql/p3_02` (promoted the stuck rerender item). **NOT fleet-wide — idea.uk was
>    the only affected site.** Verified: `curl https://idea.uk/ | grep -oE '<a href="[^"]*"'` → 0 empty.
> 2. **The free taster ("no chrome, just text" + "POST only").** `/audience-check` is an AJAX fragment
>    endpoint by design; the form was seeded as a native POST with no JS, so the browser navigated to
>    the bare fragment. Same defect `p2_02` fixed for the report form, never applied here. Fixed by
>    `sql/p3_03` (JS interceptor + result div; corrected the pointer-page URL that fed the POST-only
>    cards) + `sql/p3_04` (forced a SECTION rerender — a plain `rerender-pages` cannot apply a template
>    edit; see the landmine below). Verified: real POST returns the 2537B fragment; taster runs in place.
>
> **THE PLATFORM FIX is SHIPPED AND LIVE — the council thread is CLOSED (2026-07-22). Not the live
> thread any more.** The chrome renderer's schema-blindness (a CLASS, not just idea.uk) went to the
> council gate as `SUBMISSION_CORR=7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5`. **All 3 rounds = REVISE**
> (round-1 void was `bugs_open/019`, since fixed & live). The base observability fix (schema resolver +
> named dead-control Errors + `bugs_open/041`'s dead-JS UNION) was committed & rolled during rounds 2–3
> by concurrent sessions and is **live in v1.0.1146+**. Round 3 (verdict 2026-07-21 11:17, REVISE,
> 11/13 approve, non-veto) surfaced three REAL objections the handoff had wrongly predicted were
> "wording": (a) the missing-field detector used a control-flow-blind **regex** that false-flags
> `{{range}}`/`{{if}}`-nested fields — ~30 active components would log false Errors; (b) a second
> silent-drop **sibling** (`RenderTemplateWithMap`) the fix hadn't touched; (c) `RenderTemplate`'s
> caller set is large/diverse (8 sites, 5 pipelines). **Owner ruled 2026-07-22: ship the fix, no round
> 4** (council is advisory, stays at REVISE, **no `Council-Reviewed:` trailer** — same posture as
> `bugs_open/053`).
>
> **Shipped `78482c86b` (2026-07-22), VERIFIED LIVE on v1.0.1149:** rewrote `missingBareFields` as a
> scope-aware `text/template/parse` walk (only ungated root-scope fields reported; regex kept as the
> unparseable-template fallback) + routed the sibling `RenderTemplateWithMap` through the same detector.
> Pod-grep (symbols created by ONLY my commit): `missingBareFieldsRegex`/`bareFieldName`/
> `scanTemplateFuncs` all present in `agent-chassis-7d4ff8b54-cm786`. **Caveat:** the sibling half is
> **dead code today** (`RenderTemplateWithMap`=0 in the binary — its only caller `rerenderContactInfo`
> has no callers; linker-eliminated), so it is *correct-if-revived*, not doing runtime work now.
>
> **`bugs_open/054` (block/escalate) is the remaining platform follow-on — OWNED BY ANOTHER SESSION,
> not started as of this handoff.** Number collision: `054_…_chrome_unresolved_field_escalation_and_
> consumer` is OURS; `054_…_unguarded_range_items…` is the relojistas thread's. My parse-tree fix
> SHARPENS the exact `inURLAttr` signal 054 will escalate on (no more false positives) — coordinate,
> don't duplicate. Do NOT re-submit `7152c7cf`; the council thread is done.
>
> **NEW BUGS THIS THREAD FILED (all real, none started):**
> - `bugs_open/030` — the dispatch queue: ONE partition, ONE consumer, so every session's trigger
>   serialises. Measured latency 16–36 min, and under load it DIVERGES (lag 21→161 in 2h). A council
>   review costs an hour+ before it starts. **Cheapest fix: print the lag at publish time** (snippet in
>   030). **Check lag before submitting anything** (`kafka-consumer-groups.sh --describe --group
>   generic-requests-group`), and NEVER re-fire a queued dispatch — it double-spends and lands further back.
> - `bugs_open/041` — chrome component JS is never published (`collectJSAssets` reads `page_components`
>   only); idea.uk's mobile menu is dead on every page (`/tools/assets/site-header.js` 404s). Fixed IN
>   the council submission (edit 5), so it lands with the platform fix.
> - `bugs_open/054` — the block/escalate follow-on (above).
> - `bugs_open/006` C addendum — a claim that dies BEFORE doing work stalls indefinitely (no
>   `claimed_at` requeue predicate exists in `platform/`); operator reset is in the addendum.
> - `bugs_open/024` — got a 2nd reproduction: `rerender-pages` sets no `spec.reason`, so it can NEVER
>   apply a template edit, for any component, on any site, while reporting success.
>
> **Still owed on the tool (owner box-side, unchanged from before, NOT this thread's work):** deploy the
> tool binary (hardened `/request` + email-subject fix + it should emit `/report.html#request-a-report`
> so `p3_03`'s client-side `#request` retarget stopgap can be deleted); prove Stripe through the new
> nginx; confirm `proxy_read_timeout`; purge Cloudflare; two SES bounce DNS records.
>
> **Four rules this workstream learned the hard way:**
> 1. **Verify against the deployed artefact, never the work-item status.** `complete` is not proof:
>    a rerender reported 9/9 complete, deployed real files, published a JS asset — and changed nothing
>    on the page (`§X.4`). Read the work item's `result` JSON for what actually deployed.
> 2. **A rerender has TWO modes and the default cannot see template edits.** `rerender-pages` sets no
>    `spec.reason` → assemble-from-stored-HTML. To apply a `content_components.html_template` edit you
>    need `reason='section_data_resolved'` (or `image_landed`/`cta_links_stale`) — insert the
>    `page_rerender` item by hand (`sql/p3_04` is the template). Guard it: that path escalates to the
>    LLM content writer (rewrites live copy) if any section has NULL `content_data`.
> 3. **A schema `fallback` is not a safe default — on a URL field it is a fabrication licence.** Applying
>    `header-bold-gradient.cta_url`'s `/contact.html` fallback on a miss re-creates the phantom-CTA bug
>    LNK-007 killed. Correct-or-absent (LNK-005): leave it unset, let the gated template render nothing.
> 4. **A rendered value that looks hand-authored may be a resolved query.** The tool-card URLs came from
>    `source: query.pages_where_type:tool` (the pointer page), not the stored `content_data`. Fix the
>    source, not the copy.

**Updated 2026-07-22.** This is the single entry point to continue the idea.uk → VM workstream.
Read `SUMMARY_idea_uk_vm_site.md` for the plain-English state, then this for the operational detail.
Companions in this directory: `PLAN`, `RUNBOOK`, `RUNNING_NOTES` (execution log — newest at the
bottom, §X.1–§X.8 cover this thread; §X.8 is the round-3 close), `README_where_we_are.md` (owner's
plain-prose log), `council_submission_chrome_schema_driven.json` (the council submission — CLOSED, do
not resubmit), and `sql/` (every DB change applied, in order — `p3_01`…`p3_04` are this thread's). The
`HANDOFF_replan_clobbers_built_pages_FIX.md` here is a SEPARATE chassis-fix task.

**Where to pick up (new thread, in order):**
1. **The council thread (`7152c7cf`) is CLOSED — shipped on the owner's ruling, live in v1.0.1149. Do
   NOT read its verdict expecting a live decision, and do NOT resubmit.** (See START HERE + NOTES §X.8.)
2. `bugs_open/054` (block/escalate — make an unresolvable render field ESCALATE, not just log) is the
   remaining platform follow-on. **OWNED BY ANOTHER SESSION — check `scripts/who-owns.py 054` and read
   their docs before touching it.** It consumes the `inURLAttr` signal my parse-tree fix just made
   accurate; it also overlaps `bugs_open/023` fix #3 (same consumer). Coordinate, don't build a
   parallel handler. Number collision: resolve `054` by slug (chrome-escalation is ours).
4. Side-finding to close: `sites.content_data` for idea.uk still holds the stale
   `idea-uk@leopardess.uk` (a reviewer's check surfaced it; the p1_05/p1_06 sweep missed this column).
   Not rendering today, but a live wrong address one code path away.

## Goal
Make idea.uk one complete site behind the VM's nginx: the chassis-built static pages **and** the live
£29 tool, on one origin. Today they're disconnected — static → B2 (invisible; DNS → the VM), and the
VM's nginx serves only the tool.

## Working rules
Go not Python. British English. **Schema first** (`\d <table>`, read the function before changing it).
Structural fixes over patches. Reuse existing functions. `logger.Info` not `.Debug`. A 0-row result
isn't decisive until the query is cleared. Go changes are **inert until a chassis image rebuild**;
DB/workflow config is live immediately. The idea **tool** is a separate stdlib-only Go module
(`docs024_key_docs_latest/idea.uk/golang_files/`, `module idea`) with **no CI** — ship by building a
linux/amd64 binary, scp to the box, `systemctl restart idea`.

## Key facts
- idea.uk site_id `1244516d-014d-421c-88c6-090bb1e9552a`.
- Box: Hetzner (Nuremberg) `116.203.204.115`, `ssh root@116.203.204.115`. Tool: systemd `idea`,
  `127.0.0.1:8080`, orders in `/var/lib/idea/orders.json` (a FILE, **no DB**), env `/etc/idea/idea.env`.
- **Not this box:** `167.233.33.159` is relojistas' box. `setup.sh` has takeover semantics — never
  point it at the live idea.uk box.
- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
- Deployed chassis image at last check: **v1.0.1123** (built from local filesystem, not git — verify
  behaviour against the pod, not commits).

## DONE (this workstream)
1. **Credential scrub** — `idea.env.example` real AWS SES creds + `INTERNAL_API_KEY` replaced with
   placeholders; `scripts/check-secrets.sh` + `.githooks/pre-commit` guard installed
   (`git config core.hooksPath .githooks`). Stripe/Anthropic there were placeholders — never exposed.
   ⚠️ **Guard + hook are UNTRACKED** — commit them so other clones get the protection.
2. **Site completed** — 9 coherent pages, all `deployed`. `guides-index` + `news-index` composed
   (were 404). `tool-audience-check` is a **pointer page** (`url=/audience-check`, `build_status=
   deployed` pinned to the current plan so reconcile skips it, 0 sections); the tool-list cards on
   `index`+`tools` link straight to the live `/audience-check`.
3. **Per-site deploy target wired** — `resolveGitRepoName` (`helpers.go:206`) now called by
   `GitCommitAction` + `deploy_image_asset_action`; `upsertSite`/`EnsureSiteRecordAction` surface
   `github_repo`. Committed and **shipped in v1.0.1123**. **ACTIVATED 2026-07-16** — see #7–#9.
4. **`/request` hardened** — honeypot (`company_url`) + timing gate (`_elapsed` < 2500ms) + intake
   rate limit (`newIntakeLimiter`, 5/hr+15/day) + `mail.ParseAddress` validation + length caps + IP/UA
   capture on the `Order`; `request_hardening_test.go` (6 subtests PASS); `go build`/`vet` clean.
   NOT yet deployed to the box.
5. **Contact email** set to `idea.uk@contactforsales.com` across all sources the validator COALESCEs
   (`sites.email` is the canonical; `site_specs.identity.email` is what renders — both aligned; see
   `sql/p1_05` + `sql/p1_06`).
6. **Docs corrected** — `../idea_uk_section_data_missing/HANDOFF_spam_and_ip_blocklist.md` rewritten
   (it named the wrong process + datastore); `spam_read.sql` neutered as void.
7. **§2b DONE (2026-07-16)** — `gqls/vm-sites` Action guarded: `deploy-targets.json` allowlist
   (relojistas.com → its box only; unmapped domains skipped), `VM_HOST` secret retired. Verified
   LIVE: idea.uk skip proven 3×, relojistas deploy green through the map. Secrets guard + hook now
   tracked (were swept into `.gitignore` by bulk commits).
8. **vm-sites runner EXISTS now** — the Action had NEVER run (no runner on the repo; runner image
   had no ssh/rsync — silent exit-127). Fixed: image `aqls/github-actions-runner:v1.0.1126`
   (+openssh-client +rsync) + new `github-actions-runner-vmsites` deployment
   (`deployments/kustomize/services/github-actions-runner-vmsites/`). RUNNING_NOTES §N.
9. **§2c DONE + repo seeded** — `sites.github_repo='vm-sites'` for idea.uk (rollback: set NULL);
   `gqls/vm-sites` seeded with the built artefact from `gqls/sites` (8 pages + assets, 4cbaf2a).
   ⚠️ RUNBOOK §3b corrected: static `terms.html`/`refund-policy.html` DO exist and are footer-linked
   with `.html` — cutover config needs 301s to the tool's canonical legal pages.

## PENDING — next actions

### Owner (need box SSH / external access; can't be done from the chat sandbox)
- ~~ROTATE the exposed creds~~ **DONE 2026-07-17** — old SES user deleted, new SMTP user verified,
  `INTERNAL_API_KEY` rotated, service restarted healthy. Leaked history values are dead. `/op` links:
  issue fresh on next use (old ones no longer verify).
- **Deploy the hardened tool**: `cd …/idea.uk/golang_files && GOOS=linux GOARCH=amd64 go build -o idea .`,
  scp to `/opt/idea/idea.new`, mv, `systemctl restart idea`. RUNBOOK Phase 4 shipping note.

### Next (all remaining steps run on the box — owner's hands; the chat prepares/verifies)
1. ~~Provision pull-sync on the idea.uk box~~ **DONE 2026-07-18** — `/var/www/idea.uk` holds all 8
   pages, `sitesync.timer` syncs every 5 min, read-only deploy key accepted, nginx untouched.
   Traps fixed en route: `ssh` ignores `$HOME` (`/bugs_open/016`); `scp -r` nests on an existing
   destination (RUNBOOK §3a). RUNNING_NOTES §S.
2. **nginx cutover** (RUNBOOK §3b–3e): static root + proxy the **16** reserved tool paths
   (`service.go:527-543` — the full list is in the RUNBOOK; the old runbook's 7-path list would break
   the taster + operator flow) **+ the three `.html→` 301s for the legal pages (§3b correction)**.
   Prove `/stripe/webhook` through the new config BEFORE cutting over.
3. **Real-client-IP in nginx** (RUNBOOK §4a): idea.uk is behind Cloudflare but `setup.sh` never sets
   `set_real_ip_from`/`real_ip_header CF-Connecting-IP`, so nginx (and the new `/request` IP capture)
   would see Cloudflare's IP. Confirm the record is proxied (orange) first. Needed before any IP block
   list is meaningful.
4. **Remove existing spam** from `/var/lib/idea/orders.json` (owner-side; RUNBOOK §4c): back up, filter
   the all-`test` rows, restart. No DB, no `Delete` method — edit the file.

## Landmines (do not relearn the hard way)
- **Never re-run `build-site-planner` to compose missing pages** on a non-adoption-locked site — it
  silently regresses built pages and can't fill empty ones. Full write-up:
  `HANDOFF_replan_clobbers_built_pages_FIX.md` + memory `replan-clobbers-built-pages`. To compose one
  page, drive its build; don't re-plan.
- **git-adapter `createOrGetRepo` makes repos PUBLIC** — create any target repo by hand.
- **`page-build-handler` claim-timeout churn**: a build can succeed (page deployed) yet its work item
  reverts and re-runs. Verify via `page_components`/deployed HTML, and mark a verifiably-complete item
  `complete` to stop the churn. Logged in `aaa_fails_to_mend/006`.

## Open decisions (none blocking)
- ~~`/privacy` after cutover~~ **RESOLVED 2026-07-18: the tool keeps all three legal pages**; the
  static `.html` copies 301 onto them (already staged in `box/idea.uk.nginx`).
- ~~`/contact.html` form~~ **RESOLVED 2026-07-17** — mailto (owner's choice); staged at source
  (`sql/p1_07_contact_form_mailto.sql`), publishes on the next contact-page build. Also fixed a stale
  `idea-uk@leopardess.uk` in the form description. RUNNING_NOTES §Q.

## Errors parked for other chats
`aaa_fails_to_mend/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md` — (A) crash-looping runner replica,
(B) fleet-wide dead contact-form, (C) claim-timeout churn. `001_…replan_clobbers_built_pages_FIX.md`
— the planner bug.
