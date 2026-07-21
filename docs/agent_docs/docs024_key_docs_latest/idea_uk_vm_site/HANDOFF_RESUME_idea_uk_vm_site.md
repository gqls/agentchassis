# RESUME HANDOFF — idea.uk VM site (start a fresh chat here)

> ## ▶ START HERE — state as of 2026-07-21
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
> **THE PLATFORM FIX is in COUNCIL REVIEW, round 3 pending — this is the one live thread.**
> The chrome renderer's schema-blindness is a CLASS (not just idea.uk), so it went to the council gate
> as submission **`SUBMISSION_CORR=7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5`**. Rounds 1 & 2 = REVISE; the
> round-1 void was `bugs_open/019` (fixed & now live). Round 3 submitted 2026-07-21 (`RUN_ORCH_ID=
> 0e4e5f26-5967-4343-bb89-5dedf9a5931d`), verdict pending. **FIRST THING IN THE NEW THREAD: read the
> verdict** (queries below). Round 3 = the OBSERVABILITY version (fixes the shared blanking mechanism
> in `component_library.go`, the chrome resolver, AND `bugs_open/041`'s dead-JS bug; loud field-named
> Error on dead controls). The council wanted it to BLOCK/ESCALATE too; **owner ruled 2026-07-21: ship
> observability now, do block/escalate as `bugs_open/054` (a follow-on, NOT started).**
>
> **Read the verdict:**
> ```
> SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
>  WHERE correlation_id='7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5' AND kind='council_report' ORDER BY created_at;
> SELECT body FROM diagnosis_artifacts WHERE correlation_id='7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5'
>  AND kind='council_report' ORDER BY created_at DESC LIMIT 1;   -- full reviews JSON
> ```
> If APPROVED → the fix is a Go change (`platform/`), so it must be BUILT into a chassis image and
> rolled, then verified in-pod, then commit carries `Council-Reviewed: 7152c7cf-…`. The plan (6 edits,
> all in the submission JSON in this dir) is a SKETCH — someone implements the real diff. If REVISE
> again → the objections come back with the reviewers' own checks answered; the submission JSON has
> every measurement already attached, so a 4th round is wording, not new evidence.
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

**Updated 2026-07-21.** This is the single entry point to continue the idea.uk → VM workstream.
Read `SUMMARY_idea_uk_vm_site.md` for the plain-English state, then this for the operational detail.
Companions in this directory: `PLAN`, `RUNBOOK`, `RUNNING_NOTES` (execution log — newest at the
bottom, §X.1–§X.7 cover this thread), `README_where_we_are.md` (owner's plain-prose log),
`council_submission_chrome_schema_driven.json` (the live council submission), and `sql/` (every DB
change applied, in order — `p3_01`…`p3_04` are this thread's). The
`HANDOFF_replan_clobbers_built_pages_FIX.md` here is a SEPARATE chassis-fix task.

**Where to pick up (new thread, in order):**
1. Read the council verdict for `7152c7cf` (queries in START HERE). That is the one live decision.
2. If APPROVED → implement the 6-edit plan for real, build+roll a chassis image, verify in-pod, commit
   with the `Council-Reviewed:` trailer. If REVISE → the objections are wording now, not evidence.
3. `bugs_open/054` (block/escalate) is the owner-scheduled next platform piece — but it overlaps
   `bugs_open/023` fix #3 (the same consumer); coordinate, don't build a parallel handler.
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
