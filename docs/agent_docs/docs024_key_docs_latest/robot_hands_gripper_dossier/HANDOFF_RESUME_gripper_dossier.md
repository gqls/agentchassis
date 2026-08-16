# RESUME HERE — gripper dossier pilot

**Last updated 2026-08-16 — the route group is BUILT and tested, NOT shipped; see the 08-16 block at the end and NOTES 08-16.** (body written 07-27; switch positions corrected 07-31
08:15Z; fixture 4 result 07-31 10:45Z; cleanup complete 07-31 15:42Z; `bugs_open/160`
CLOSED+LIVE 07-31 21:10; mailer adoption re-checked 08-04, found NOT self-contained; route-group
proposal drafted 08-05; **both owner-supplied credentials (Anthropic key + SMTP) issued and
verified live as of 08-15** — see below). Read
this first, then `NOTES_…md` (bottom
up) for the technical log and `README_where_we_are.md` for the owner's account.
Design of record: `DESIGN_2026-07-24_gripper_dossier_pilot.md` — **but read its
§2 CORRECTION before building anything public-facing.**

---

## State in one line

**The cluster half is BUILT, LIVE and PROVEN end to end.** All three DESIGN §6
fixtures pass on the real site, including the failing branch. The public-facing
half is not built and its design was rewritten on 07-26.

## What is proven, with the evidence

| fixture | result | evidence |
|---|---|---|
| 1 — success | **PASS** | `robot-hands.com/reports/d1a371be-04a5-4ee6-b744-d64c6fd9e7c4.html` HTTP 200, 43,049 B, carries the substituted formula literal `(2.5 × 12 × 2) ÷ (0.15 × 2)`; negative control `(9.9 × 99 × 9)` absent; sidecar `ready`; 0 refs on the homepage |
| 2 — honest no-match | **PASS** | `…/reports/29c3f8aa-3246-4a81-be8a-1e6b237cc467.html` HTTP 200, 41,670 B, item **`complete`** (a no-match is a SUCCESS), carries *"No gripper in this index meets the requirement"* verbatim, zero Match/Marginal |
| 3 — induced failure | **PASS** | `mass_kg="not-a-number"` → `score_grippers` hard error (never a guessed default) → `handle_failure` → `publish_failed` → `fail_workflow`; item **`failed`**; `…/reports/edd863e8-445b-494e-a5d7-5ebdeb6d68cb.json` serves `{"status":"failed"}` HTTP 200; **0 HTML pages created** |

Live on chassis **v1.0.1175**. Seeds applied: **204, 207, 209, 210**.

## Current switch positions — ~~everything is OFF~~ **DISPATCH IS ON since 2026-07-30 22:13Z**

> **UPDATED 2026-07-31 08:15Z — this section said "everything is OFF" and that is no
> longer true.** The owner read the two fixture pages, found two rendering defects in
> the headroom chart, and instructed: enable dispatch and regenerate. Both happened.
> The strikethrough below is kept because "nothing runs until someone enables them" is
> still the right mental model for `report-request-pull`.

- ~~`report-dispatch` **disabled**~~ → **`report-dispatch` is ENABLED** (owner
  instruction 2026-07-30). 90s interval, target `report-dispatch-loop`, topic
  `system.agent.scheduled.requests` — confirmed consumed via the LIVE deployment's
  `EXTRA_REQUEST_TOPICS`, and proven end-to-end by a sibling scheduled task on the same
  topic in the same window. **`report-request-pull` remains disabled.** Enable order is
  `report-dispatch` first, always.
  - **It is correctly self-gating, so an idle ON is free.** Its `pre_query` ends
    `HAVING count(*) > 0`, so with nothing at `awaiting_report` the scheduler logs
    *"Pre-query found no rows — task ran with nothing to do"* and **publishes no
    message at all**. Do not "fix" that query — and do **not** read it through
    `left(pre_query,120)`, which cuts the `HAVING` off and makes it look unfiltered
    (that misreading is logged in `WRONG_CALLS.md`, 2026-07-30).
- Three ~~`source='manual-test'` work items~~ → **four**: fixtures 1–3 (terminal since
  07-27) plus **FIXTURE 4**, `4ccc73d7-c467-480f-9a39-0b327b383870`, queued
  2026-07-31 08:15Z. It re-runs **fixture 1's exact inputs** under a new
  `request_id` so the chart fix is directly comparable — expected page
  `/reports/bf3765d6-befe-43a8-b1cd-ca5c210f39e9.html`.
  - **UPDATED 2026-07-31 10:45Z — FIXTURE 4 IS `complete` AND THE FIX IS VERIFIED BY EYE.**
    8m34s end to end (08:16:03 → 08:24:37Z), page **HTTP 200, 43,546 B**, no `verify_prose`
    violation this run. All three defects gone against fixture 1 on identical inputs: the
    capped-bar label reads `6.42× (Insufficient data)` **whole** (was clipped mid-word), the
    two reference captions are on **separate lines** at `y=364`/`y=376` (were overprinted
    into illegible mush at a shared `y=364`), and both capped bars **end in a point** while
    the uncapped 2.45× stays flat (pre-fix, 6.42× and 7.60× were flat rects indistinguishable
    from a true 3× bar). Evidence, geometry and method: NOTES, 07-31 "later" entry.
    **The label is full text with no ellipsis** — so this is the computed gutter working, not
    the truncation fallback masking it, which is the distinction that made a mutation test
    read falsely green during the fix.
  - **One residual, cosmetic and PRE-EXISTING: value labels overlap the dashed reference
    lines** on five rows. Identical in fixture 1, so the fix neither caused nor addressed it.
    Flagged to the owner; deliberately **not** filed as a bug.
  - **CLEANUP COMPLETE 2026-07-31 15:42Z, owner instruction.** Both halves done and each
    verified at the artefact, not the status. **DB** (this session, ~15:10Z): `pages` (3
    rows) + cascaded `page_components`, plus `site_work_items` (4 rows, `source='manual-test'`)
    — deleted after checking every FK onto both tables for dependents (zero everywhere).
    **Git/CDN** (owner-run via `!`, this session's verified command, 15:42Z): commit
    `c47bbfab6` on `gqls/sites` `master` deleted `robot-hands.com/reports/` (8 files — the 3
    fixture html+json pairs, fixture 3's failure sidecar, and one untracked stray
    `fixture-1-success.json` nobody had been tracking). Note for next time: `sites.github_repo`
    is empty for this domain, so the real deploy target is the fallback repo literally named
    `sites`, branch `master` — **not** `main`, which is what the DB's `github_branch` column
    says; a live discrepancy, harmless here, worth knowing before trusting that column.
    **Verified, not assumed:** the new tree has no `reports` entry and every sibling path is
    untouched; `gh run list` shows the triggered `Deploy to B2` run completed successfully in
    25s; all five retracted URLs return **404**, and two unrelated live pages still return
    **200** (no collateral). Full detail in NOTES, 07-31 "later still" and "final".
    **Nothing further owed on this cleanup.** The `gh api … -X POST` git-tree write was
    refused twice by this session's own auto-mode classifier before the owner ran it
    directly — a live-repo mutation guard working as intended, not a repo problem; don't read
    that as a reason to route around it next time either.
- ~~**Cleanup is still owed** and now covers more: the `manual-test` rows, the two 07-27
  pages, fixture 3's `…edd863e8….json`, and fixture 4's page once the owner has seen
  it. Do not clean up unasked — the owner asked to see them.~~ **DONE 2026-07-31 15:42Z** —
  see the entry above under fixture 4.
- **Seed 208 is committed but deliberately NOT applied.** Its `base_url` points at
  `…/api/gripper/v1`, a path the island Caddy allowlist 404s. Re-seed it to
  `https://tools.apis.uk/api/v1/tools/gripper` when the route exists.

## Next actions, in order

1. **Do NOT write `cmd/gripper-intake/`.** It would be the estate's fourth VM fork.
   The public half is a route group **inside the existing `tools-api`**
   (`internal/tools-api/`), which already has per-request CORS from the island's
   own `sites` table, a rate limiter, an input cap and a key. `tools-api` is the
   **gauntlet thread's** ("vonc 6"); `bugs_open/083` (the *other* 083 — resolve by
   slug, this is `detected_findings_never_reach_a_handler`, not the gauntlet-503 one,
   which closed) — coordinate before editing (`scripts/who-owns.py`).
   - **`platform/mailer` has NO reachable consumer until this route group exists**
     (re-checked 08-04 — its two named callers are this route group and idea.uk's
     paid report, and the latter is a separate VM deploy outside this build). Do
     not "adopt" it by wiring an import with nothing calling it — that is the
     helper-with-no-callers trap. The real next action is coordinating this route
     group with the gauntlet lane, same channel as the httpguard `ClientIP` fix
     (write the ask here / in NOTES, owner routes the priority) — not a solo patch.
   - **DRAFTED 08-05**: `PROPOSAL_2026-08-05_gripper_route_group_in_tools_api.md`
     — grounded against the live `internal/tools-api` source (not the design doc's
     summary of it), including a correction DESIGN §2 needed: the router is NOT
     yet multi-tool despite what §2 claims, it's one hardcoded `/gauntlet` group.
     Ready to route to the gauntlet lane whenever the owner does so.
   - **Owner-supplied credentials (proposal §8): BOTH DONE as of 08-15.**
     Anthropic key: `/home/ant/.config/anthropic/gripper-dossier-api-key` (dotenv-style
     `GRIPPER_ANTHROPIC_API_KEY=...`), permissions fixed 664→600, verified live via free
     `count_tokens`. SMTP: `/home/ant/.config/gripper-dossier/smtp.env` (same dotenv
     shape — `GRIPPER_SMTP_HOST`/`_PORT`/`_USER`/`_PASS`/`_FROM`), same 664→600 fix,
     verified live via `smtplib` AUTH-only (no message sent). **Both are local dev-box
     files, not deploy artefacts** — ~~need adding to the `tools-api-secret` k8s Secret~~
     **WRONG, corrected 2026-08-16: the deploy target is the ISLAND, not the cluster.**
     `tools-api` runs on the island VM (`/opt/island`, docker compose), so these go into
     **`/opt/island/.env`** (root, 600, written ON the box — never through a transcript),
     per `RUNBOOK_island` "Tenant 2" step 2. I wrote the k8s wording in the 08-05 proposal
     and repeated it here on 08-10/08-15; the 08-16 build session caught it. There is a
     `deployments/kustomize/services/tools-api/` overlay in the repo, which is what made
     the k8s reading look right — but the live public endpoint is the island's Caddy, and
     that is where the route group actually ships.
     **08-16: outbound 465 CHECKED FROM THE ISLAND** (my 08-15 check was dev-box-only,
     which the runbook rightly flagged as insufficient): open, `220` Exim banner,
     `AUTH PLAIN LOGIN` advertised to the island IP. Runbook step 2's precondition is
     discharged; EMAIL-002 narrowed in the register (its "cloud boxes can't do 465" is
     Hetzner-specific and false here).
2. **Before that, land two shared pieces** (`features_open/024` A2/A3), or the
   estate forks at exactly this point:
   - a **mailer** in `platform/` — `grep -rn "net/smtp" --include=*.go platform/
     internal/ cmd/` returns **nothing**; the only working one is idea.uk's VM app,
     outside the build;
   - **`platform/httpguard`** — one per-IP limiter, one CORS policy, one honeypot +
     timing gate. The public API's current limiter is the weakest of the three that
     exist.
3. **Then generalise `score_grippers`** into a config-driven engine with its rule
   table in `site_specs` (owner ruling 07-27: *finish the pilot first, generalise
   after*). Pattern to copy: `CHVerticalProfile`
   (`companies_house_vertical_profiles.go`) — a table, not Go per site. Nine of 296
   registry entries currently serve two of ~1,000 sites.

## Landmines that cost real time here

- **`scheduled_tasks.target_topic`'s column DEFAULT (`system.agent.generic.requests`)
  is a topic NOTHING CONSUMES.** 18 of 18 enabled tasks use
  `system.agent.scheduled.requests`. It fails **silently and looks healthy**: the
  scheduler logs *"Successfully produced message"* and *"Triggered task"* and stamps
  both timestamps — that is the **normal** fire-and-forget path
  (`cmd/scheduler/main.go:287-296`), so equal timestamps prove nothing either way.
  Discriminating evidence is downstream only: zero `orchestration_states` rows for
  the target agent type, and zero mention of the correlation_id in the chassis log.
- **`create_report_page` requires `request_id` to be a real UUID** — it becomes the
  page's public URL. An invalid one also silently disables the failure sidecar,
  because `handle_failure` builds it from the same field.
- **`complete` / `deployed_at` is not fetchability** (`bugs_open/098`). Fixture 2 was
  **404 for ~2 minutes** after the item said `complete`, then 200. Poll the URL.
- **Verify against the pod that is running NOW.** The chassis rolled twice during
  one session (v1.0.1173 at 13:45, v1.0.1175 at 18:00, both other threads). The pod
  I first pod-grepped no longer existed an hour later.
- **`bugs_open/029-hung-spawns` is live and roll-adjacent.** One run here hung at
  `spawn_handler` for 4m45s. Stopgap: mark the orchestration `FAILED`, reset the work
  item (`claimed_by=NULL, claimed_at=NULL, attempt_count=0`); it re-claims next tick.
  A hang has **`handler_spawned` ABSENT** in `collected_data` — check that field
  rather than sampling `current_step`, which shows the same value mid-flight.

## Owner-gated

- Anthropic key **issued 07-27** (capped per project, not per key — accepted for
  now). Gates only the island/public half.
- Cleanup of the two live fixture pages awaits the owner reading them.
- The soft-launch decision (unlinked → footer nav link) is still theirs.


---

## 2026-08-16 — ROUTE GROUP BUILT (not shipped). Resume from RUNBOOK_island "Tenant 2"

- **Code**: `internal/tools-api/{gripper,store/gripper.go,handlers/gripper.go,middleware/{bands,internalkey}.go,api/server.go,config/config.go}` + `cmd/tools-api/main.go`; DB `sql_for_agents/436_tools_api_gripper_intake.sql`; seed 208 base_url corrected; island `docker-compose.yml` env block added (opt-in, `:-` defaults). Register **PUB-005**; LANDMINES "two gin groups"; council submitted with the commit.
- **Proven locally**: unit + integration tests (real Postgres, 436 applied twice, store lifecycle), real-process smoke of all four routes both directions, SIGTERM graceful stop. NOTES 08-16 has the list.
- **Three corrections to carry forward**: the chat spec speaks the CLUSTER's field names (measured against the live `report-builder` query — DESIGN §5.3's "island stays dumb" is not how the pipeline works); the island `sites` has no `deploy_config` (pull key = `GRIPPER_PULL_KEY` env, must equal seed 208's); seed 208's base_url is now the corrected `…/api/v1/tools/gripper`.
- **NEXT, in order (RUNBOOK_island "Tenant 2" steps 1–7)**: 436 on the island → secrets into `/opt/island/.env` (owner/authorised; check 465 FROM the island) → compose + image swap (bump `IMAGE_TAG`, `make build-tools-api-ref`, `docker save|ssh load`) → verify at the container → public smoke (403/200, 401/200) → seed 208 on the cluster with the SAME key → enable `report-request-pull` → then the site widget + `/gripper-report/` page (DESIGN §2 "Site side", unchanged). Email copy (`gripper/email.go`) wants an owner read before launch.
- **Council round 1 = REVISE (`editquality`, wording: my rationale's field list omitted `accel_ms2` while the grounded_in quote had it — both true; the query reads both, accel_ms2 is optional/defaulted, not collected).** Read the FULL round-1 report first, then fix `spec.go`'s package comment to list the whole read-set (collected vs defaulted) and RESUBMIT with `RESUBMIT_CORR=623da25b-16d7-4836-8667-ffcd6352d6d6`. NOTES 08-16 has the detail.
- **Switch positions unchanged**: `report-dispatch` ON, `report-request-pull` OFF, seed 208 NOT applied.
