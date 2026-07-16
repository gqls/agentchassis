# RESUME HANDOFF — idea.uk VM site (start a fresh chat here)

**Updated 2026-07-16.** This is the single entry point to continue the idea.uk → VM workstream.
Read `SUMMARY_idea_uk_vm_site.md` for the plain-English state, then this for the operational detail.
Companions in this directory: `PLAN`, `RUNBOOK`, `RUNNING_NOTES`, and `sql/` (every change applied,
in order). The `HANDOFF_replan_clobbers_built_pages_FIX.md` here is a SEPARATE chassis-fix task.

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
   `github_repo`. Committed and **shipped in v1.0.1123**. NOT yet activated (see pending #1).
4. **`/request` hardened** — honeypot (`company_url`) + timing gate (`_elapsed` < 2500ms) + intake
   rate limit (`newIntakeLimiter`, 5/hr+15/day) + `mail.ParseAddress` validation + length caps + IP/UA
   capture on the `Order`; `request_hardening_test.go` (6 subtests PASS); `go build`/`vet` clean.
   NOT yet deployed to the box.
5. **Contact email** set to `idea.uk@contactforsales.com` across all sources the validator COALESCEs
   (`sites.email` is the canonical; `site_specs.identity.email` is what renders — both aligned; see
   `sql/p1_05` + `sql/p1_06`).
6. **Docs corrected** — `../idea_uk_section_data_missing/HANDOFF_spam_and_ip_blocklist.md` rewritten
   (it named the wrong process + datastore); `spam_read.sql` neutered as void.

## PENDING — next actions

### Owner (need box SSH / external access; can't be done from the chat sandbox)
- **ROTATE the exposed creds** (urgent): new AWS SES IAM key + `openssl rand -hex 32` for
  `INTERNAL_API_KEY`, update `/etc/idea/idea.env`, `systemctl restart idea`, re-issue operator `/op`
  links. RUNBOOK Phase 0. The scrub does NOT close this — values are in public history.
- **Deploy the hardened tool**: `cd …/idea.uk/golang_files && GOOS=linux GOARCH=amd64 go build -o idea .`,
  scp to `/opt/idea/idea.new`, mv, `systemctl restart idea`. RUNBOOK Phase 4 shipping note.

### Chat-doable next
1. **Guard the `vm-sites` Action THEN activate per-site deploy** (RUNBOOK §2b→§2c). The `gqls/vm-sites`
   Action rsyncs every changed domain to ONE `VM_HOST` secret = relojistas' box `167.233.33.159`.
   Before `UPDATE sites SET github_repo='vm-sites' WHERE domain='idea.uk'`, add a `deploy-targets.json`
   (domain→host) at the repo root so idea.uk isn't pushed to the wrong machine. Then the UPDATE
   activates the (already-shipped) per-site target.
2. **Provision pull-sync on the idea.uk box** (RUNBOOK §3a): systemd timer + sparse-checkout of
   idea.uk's own folder into `/var/www/idea.uk`, read-only deploy key. (Owner runs the box commands.)
3. **nginx cutover** (RUNBOOK §3b–3e): static root + proxy the **16** reserved tool paths
   (`service.go:527-543` — the full list is in the RUNBOOK; the old runbook's 7-path list would break
   the taster + operator flow). Prove `/stripe/webhook` through the new config BEFORE cutting over.
4. **Real-client-IP in nginx** (RUNBOOK §4a): idea.uk is behind Cloudflare but `setup.sh` never sets
   `set_real_ip_from`/`real_ip_header CF-Connecting-IP`, so nginx (and the new `/request` IP capture)
   would see Cloudflare's IP. Confirm the record is proxied (orange) first. Needed before any IP block
   list is meaningful.
5. **Remove existing spam** from `/var/lib/idea/orders.json` (owner-side; RUNBOOK §4c): back up, filter
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
- `/privacy` after cutover — tool or static? (default: tool; it's embedded in the binary.)
- `/contact.html`'s form posts to a dead `/contact` — repoint to `/request` or a `mailto:`?

## Errors parked for other chats
`aaa_fails_to_mend/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md` — (A) crash-looping runner replica,
(B) fleet-wide dead contact-form, (C) claim-timeout churn. `001_…replan_clobbers_built_pages_FIX.md`
— the planner bug.
