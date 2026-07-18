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
- `/privacy` after cutover — tool or static? (default: tool; it's embedded in the binary.)
- ~~`/contact.html` form~~ **RESOLVED 2026-07-17** — mailto (owner's choice); staged at source
  (`sql/p1_07_contact_form_mailto.sql`), publishes on the next contact-page build. Also fixed a stale
  `idea-uk@leopardess.uk` in the form description. RUNNING_NOTES §Q.

## Errors parked for other chats
`aaa_fails_to_mend/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md` — (A) crash-looping runner replica,
(B) fleet-wide dead contact-form, (C) claim-timeout churn. `001_…replan_clobbers_built_pages_FIX.md`
— the planner bug.
