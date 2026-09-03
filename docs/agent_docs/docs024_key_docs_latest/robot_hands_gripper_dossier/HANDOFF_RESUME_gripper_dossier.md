# RESUME HERE — gripper dossier pilot

> # 👉 GO STRAIGHT TO THE BOTTOM: "✅ 2026-09-03 (midday) — THE GUARD IS WRITTEN, APPLIED AND COMMITTED".
> (It carries TWO CORRECTIONS to the 🔧 block below it — one of them would otherwise abort the seed.)
> (Supersedes every earlier block, all kept as history.)
> That block is the current ship state (the fix is committed and applied; the rerender is
> queued; the owner's browser is the last step), the exact commands to run, and every trap
> in what remains. Everything between here and there is history, kept for provenance —
> read it only if the bottom block sends you.

**Last updated 2026-08-25 — credentials are ON the island and verified; migration 436 is the next step and needs the owner to run it (this session's classifier refuses production DB mutations).** (body written 07-27; switch positions corrected 07-31
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

---

# ⭐ START HERE — 2026-08-25. Ship state: 2 of 7 steps done, step 1 is next and BLOCKED ON YOU

Everything above is history. This block is the current position. Ship checklist is
`docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/infra/island/RUNBOOK_island.md`
§"Tenant 2", steps 1–7 — **read its step 1 and step 2 notes, both were corrected on 08-25/08-20.**

## Where the work actually is

| # | step | state as of **2026-08-25** |
|---|---|---|
| — | route group built + tested | **DONE** (`f967d9307`, confirmed an ancestor of HEAD, so any build from HEAD carries it) |
| 2 | secrets → `/opt/island/.env` | **DONE 08-20**, owner-run. 7 `GRIPPER_*` vars, mode 600, 9 lines. Re-verified 08-25: still intact |
| — | outbound 465 from the island | **DONE 08-16.** Open; `220` Exim banner; `AUTH PLAIN LOGIN` offered to 176.126.243.183 |
| **1** | **migration 436 on the island** | **NOT DONE — NEXT, and blocked on you** (see below) |
| 3 | compose swap | not done — repo copy already carries the `GRIPPER_*` block; live copy does not (`grep -c GRIPPER_` = 0) |
| 4 | image build + swap | not done — island runs `v1.0.1216`, which **predates** the gripper code |
| 5 | verify at the container | not done |
| 6 | public smoke | not done |
| 7 | cluster: seed 208 + enable `report-request-pull` | not done |
| — | site widget + `/gripper-report/` page | not started (separate deliverable, DESIGN §2 "Site side") |

**Switch positions unchanged:** `report-dispatch` ON (self-gating, idle is free),
`report-request-pull` OFF, seed 208 NOT applied.

## The one command you need to run — step 1

I could not run it: this session's **auto-mode classifier refuses production DB mutations**
(it refused the same class of action for the git-tree write on 07-31). That is the guard
working, not a fault, and it should not be routed around. Everything it needs is verified:

```bash
cd /home/ant/projects/agentchassis
scp docs/agent_docs/sql_for_agents/436_tools_api_gripper_intake.sql root@toolsapisuk.vs.mythic-beasts.com:/opt/island/
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -v ON_ERROR_STOP=1 < 436_tools_api_gripper_intake.sql'
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -c "INSERT INTO island_migrations(filename, note) VALUES ('"'"'436_tools_api_gripper_intake'"'"', '"'"'gripper intake: 3 tables + 3 indexes + robot-hands.com site row'"'"') ON CONFLICT DO NOTHING"'
```

**Expect:** four `NOTICE … created` lines, then
`Migration 436 verified: 3 tables, 3 indexes, robot-hands.com deployed`, then `INSERT 0 1`.

**⚠ Run the three commands SEPARATELY, not chained with `&&`** — see the runbook's step-1 note.
The original one-liner ledgered into a column that does not exist, so the migration would have
applied and the ledger insert would have failed, leaving a schema change the ledger denies.

**Pre-checks already done for you (2026-08-25), so a failure means something new:**
- 436 is **purely additive** — zero `DROP`/`TRUNCATE`/`DELETE`; creates 3 tables + 3 indexes.
- **Idempotent** — every `CREATE` sits inside a `DO $$ … IF EXISTS … RAISE NOTICE 'skipping'`
  block, so a second run is a no-op (the build session also applied it twice on a throwaway PG).
- **Guarded against the wrong database** — raises if `gauntlet_rounds` or `sites` is missing,
  so it cannot be applied to `clients_db` by accident.
- **Verify block is `DO`/`RAISE EXCEPTION`, not bare `SELECT`s** — it can actually abort the
  transaction (the failure mode `LANDMINES` records for `SELECT`-only verify blocks).
- **The hardcoded site id is right**: `00ff3af5-dad8-4770-9f70-3edc267a3c92` matches
  robot-hands.com's real id in `clients_db`. A wrong one would stamp every intake row with a
  site the cluster does not have.

## After step 1, in order

3. **compose** — `scp` the repo copy of
   `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/infra/island/docker-compose.yml`
   (already carries the `GRIPPER_*` block with `:-` defaults, so it is safe to land before the
   image) and bump its `image:` tag.
4. **image** — commit first, then `make build-tools-api-ref` (builds from committed HEAD), bump
   `IMAGE_TAG`, `docker save … | gzip | ssh … 'gunzip | docker load'`, `docker compose up -d tools-api`.
   ⚠ **A same-tag rebuild serves the node's cached image** — bump the tag, always.
5. **verify at the container, never the tag** — `grep -a -c gripper/poller /tools-api` > 0 **and**
   a control literal that must read 0; then the log line `gripper route group mounted`.
6. **public smoke** — the four calls in the runbook (403 without Origin / 200 with; 401 without
   the key / 200 NDJSON ending `{"_meta":…}` with it).
7. **cluster** — apply the corrected seed 208 with `-v pull_key=<the value in
   /home/ant/.config/gripper-dossier/pull-key>`, then enable `report-request-pull`.

## Things that will mislead you if nobody tells you

- **`report-request-pull` must stay OFF until step 6 passes.** Enabled early it 404s every tick
  against an endpoint that does not exist yet.
- **The pull key exists in exactly two places and must match:** `/opt/island/.env`
  (`GRIPPER_PULL_KEY`, already there) and seed 208 on the cluster (not applied). The local copy
  is `/home/ant/.config/gripper-dossier/pull-key` (48 hex chars, mode 600) — **keep it**; a
  mismatch is a silent 401 on every pull tick, not an error anyone is shown.
- **`$` in the SMTP password is stored DOUBLED (`$$`) in `.env`, deliberately.** Compose
  interpolates `.env`, so a single `$` is silently eaten and the container gets a truncated
  password — measured on this box. Do not "tidy" it. Full entry: `LANDMINES.md`, *"A `$` in a
  secret written to a Compose `.env` is SILENTLY EATEN"*.
- **The island `sites` table holds only `vonc.com` until 436 runs.** Until then CORS answers
  **403** to anything from robot-hands.com — expected, not a bug. 436 seeds that row.
- **Liveness-probe the API with a real endpoint, never `GET /`** (the runbook's own note):
  `curl -X POST https://tools.apis.uk/api/v1/tools/gauntlet/round -H 'Origin: https://vonc.com' -d '{}'` → 200.
  A **wrong-origin 403 proves nothing** — I made exactly that mistake on 08-20 and nearly
  reported a guessed-domain 403 as "service healthy".
- **Long commands break when pasted** — they wrap and fail silently (step 2's original one-liner
  created no file at all). Prefer short commands calling a script; that is why
  `~/.config/gripper-dossier/append-env.sh` exists.

## Owed by the build session, deliberately NOT taken by this lane

**Council round 1 came back REVISE** (`623da25b-16d7-4836-8667-ffcd6352d6d6`, `editquality`,
a submission-wording defect: the rationale's field list omitted `accel_ms2` while the
`grounded_in` quote included it — both statements true, the query reads both, `accel_ms2` is
optional/defaulted and not collected by the chat). The fix is to state the full read-set in
`internal/tools-api/gripper/spec.go`'s package comment (collected vs defaulted) and resubmit
with `RESUBMIT_CORR=623da25b-16d7-4836-8667-ffcd6352d6d6`, after reading the FULL round-1
report. **That is the "tools api" build session's, with their own recipe already written down
in NOTES 08-16** — racing them to it is the compete-instead-of-contribute failure
`scripts/who-owns.py` exists to prevent. `[UNVERIFIED 2026-08-25]` I could not find the
correlation in `orchestration_states` when I looked; that may be age-out rather than absence,
and I did not chase it because it is not this lane's to close.

## Credentials — both issued, both verified, neither deployed beyond `.env`

- **Anthropic**: `/home/ant/.config/anthropic/gripper-dossier-api-key` (dotenv line, not a bare
  key), spend-capped, ~~verified live via the free `count_tokens` endpoint~~
  > **CORRECTED 2026-08-26:** that probe verified AUTH only — `count_tokens` is free and
  > succeeds on a zero-credit account. The account behind this key has NO CREDIT, found by
  > the first paid call (the ship-day smoke). See the ⭐⭐⭐ block + WRONG_CALLS 2026-08-26.
- **SMTP**: `/home/ant/.config/gripper-dossier/smtp.env`, verified by real `AUTH` from the dev
  box **and** by reachability + `AUTH PLAIN LOGIN` from the island.
- Both are **local dev-box files**, both mode 600. They live on the island only in
  `/opt/island/.env`. ⚠ **NOT a k8s Secret** — that wording was mine and it was wrong; this
  service runs on the island VM under docker compose, and the
  `deployments/kustomize/services/tools-api/` overlay in the repo is what makes the k8s reading
  look plausible. It is not where the live endpoint runs.
- **The SMTP password transited a chat transcript on 08-15** (owner pasted it). Not exploited
  and the file permissions are right, but it is worth rotating in cPanel at some point; if you
  do, update `smtp.env` **and** `/opt/island/.env` (with the `$$` escaping) together.


---

# ⭐⭐ START HERE — 2026-08-25 EVENING. Local prep ALL DONE; the owner runs two things; then this lane finishes 5–7

Supersedes the afternoon block above. Session "AI page 3" (evening) verified the island
cold — identical to the afternoon state — then did everything the classifier permits.

## What moved this evening (all verified, see NOTES 2026-08-25 evening for evidence)

- **Compose drift caught and defused.** The repo `docker-compose.yml` lacked the owner's
  07-31 `RATE_LIMIT_RPS: "2"` / `RATE_LIMIT_BURST: "20"` live-box tuning; step 3 as
  written would have silently reverted it. Merged back, proven additive-only, runbook
  Tenant 2 step 3 corrected, LANDMINES entry appended. Commit `644d07302`.
- **Image `docker.io/aqls/tools-api:v1.0.1340` BUILT and PROVEN** from committed ref
  `eef758543` (label `org.opencontainers.image.revision`; the BINARY is unstamped —
  `tools-api.dockerfile` never consumes `GIT_COMMIT`, so the label is the provenance,
  not a binary grep). Gripper symbols in the binary: `gripper/poller` ×2, control 0.
  HEAD tools-api tests all pass. Tag v1.0.1339 was burned by another session
  (uncommitted makefile bump, absorbed + declared in `644d07302`).
- **Archive staged**: `~/.config/gripper-dossier/tools-api-v1.0.1340.tar.gz`
  (19,933,240 B, md5 `1943e1c0dd517c880ac491cdaa352566`).
- **Classifier boundary mapped**: island mutations refused (scp), and CREATING a
  steps-3/4 ship script was refused twice (Bash heredoc + Write tool). The step-1
  script WAS permitted. Do not route around; steps 3–4 are owner-pasted commands.
- **`landmines-sync` is broken estate-wide** — delta logic wants to rewrite all 847
  entries, payload dies in the kubectl exec stream (`unexpected EOF`), 3 runs. Filed
  `bugs_open/402_HANDOFF_2026-08-25_landmines_sync_delta_wants_the_whole_corpus.md`.
  This lane's new landmine entry is in the FILE (system of record) but NOT delivered
  to `doc_notes`, and its verifier is NOT armed — re-run
  `./scripts/landmines-verify-dispatch.sh` once 401 is fixed.

  > **CORRECTED 2026-08-26 (closing session, not this lane): 402 is CLOSED and there
  > is NOTHING left for this lane to re-run.** Your entry
  > (`…the-repo-copy-of-a-box-deployed-config-file-drifts-behind-live-edits…`) was
  > delivered AND verified the same evening you filed — two landmine-verifier
  > verdicts in `doc_notes`, 2026-08-25 20:05:04Z and 20:08:20Z (both UNVERIFIABLE,
  > which is a verdict: the dispatch ran). Do not re-run the dispatch for it. Also:
  > the "wants to rewrite all 847" half was a misread of a mislabelled print
  > (`to insert/refresh:` showed the file's TOTAL, always) — the delta logic was
  > fine; the EOF was the 3MB body read-back, retried since `02c740616` and made
  > ~110KB by the closing commit. Full account:
  > `bugs_closed/402_HANDOFF_2026-08-25_landmines_sync_delta_wants_the_whole_corpus.md`.

## The owner's two runs

**1. Migration 436** — type in this chat:

```
! bash ~/.config/gripper-dossier/ship-step1-migrate.sh
```

Self-verifying (scp + md5, apply with ON_ERROR_STOP, ledger with the CORRECTED
`filename/note` columns, then reads back `gripper tables: 3 / ledger row: 1 / site
row: 1`). Idempotent; safe to re-run.

**2. Compose + image (steps 3–4)** — after 1 succeeds, paste these four, one at a time
(each is one line; the guard query first — expect `1`):

```
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -Atc "SELECT count(*) FROM island_migrations WHERE filename='"'"'436_tools_api_gripper_intake'"'"'"'
scp /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/infra/island/docker-compose.yml root@toolsapisuk.vs.mythic-beasts.com:/opt/island/docker-compose.yml
ssh root@toolsapisuk.vs.mythic-beasts.com 'gunzip | docker load' < /home/ant/.config/gripper-dossier/tools-api-v1.0.1340.tar.gz
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose up -d tools-api'
```

(The last command restarts the public tools-api — a blip of a few seconds on
tools.apis.uk. Expect `Recreated` and the gauntlet keeps working; the gripper group
mounts because `.env` already carries the key.)

## Then THIS lane owes, in order

5. **Verify at the container**: `docker compose ps tools-api` shows `v1.0.1340`;
   `docker inspect` label = `eef758543…`; log shows `gripper route group mounted` +
   `gripper/poller: started`. (Read-only ssh IS permitted for this session class.)
6. **Public smoke** — the four runbook calls (403/200 on Origin, 401/200 on the pull
   key — run the keyed one FROM the box so the key never transits a transcript), plus
   one real `/chat` turn (~1 Haiku call).
7. **Cluster half**: apply seed 208 with
   `-v pull_key="$(cat ~/.config/gripper-dossier/pull-key | cut -d= -f2)"`-style
   substitution (never echo it) — check the file's dotenv-vs-bare shape first — then
   enable `report-request-pull`, then watch ONE tick for
   `per_site → {"robot-hands.com": …}` with no `error`.
- **`report-request-pull` stays OFF until 6 passes.** Unchanged.
- ~~Site widget + `/gripper-report/` page: still a separate deliverable (DESIGN §2).~~
  **PAGE LIVE 2026-08-26 night** (`/gripper-report.html`, unlinked+noindex, caveat
  copy, seed 651 APPROVED corr `de0068fd`, advisories answered — NOTES night entry).
  **Widget bundle NOT yet served: every route is sitewide (owner decision A/B in
  NOTES), then the owner's browser click-test, then the soft-launch flip.**
- Council resubmit (`RESUBMIT_CORR=623da25b…`): still the tools-api build session's,
  deliberately not taken. Unchanged.


---

# ⭐⭐⭐ 2026-08-26 — LIVE ON THE ISLAND. Smoke 5/6. ONE owner item, then step 7 and done

**The ship happened this morning, owner-run, both halves clean.** Steps 1, 3, 4, 5
are DONE and verified at the artefact; step 6 is 5-of-6; step 7 is HELD (rule: only
after 6 fully passes). Evidence for everything: NOTES 2026-08-26 morning.

| # | step | state |
|---|---|---|
| 1 | migration 436 on the island | **DONE 08-26 ~08:50Z** — verify block passed, ledgered, read-back `3 tables / 1 ledger / 1 site row` |
| 3 | compose swap | **DONE** — merged copy over, `RATE_LIMIT 2/20` proven INSIDE the container (owner tuning survived) |
| 4 | image swap | **DONE** — `v1.0.1340` Up, revision label `eef758543…`, binary `gripper/poller`×2 control 0 |
| 5 | verify at container | **DONE** — `gripper route group mounted` + `gripper/poller: started tick=1m0s` + `listening` |
| 6 | public smoke | **5/6**: 403/200 CORS ✓, 401/keyed-200 `_meta` ✓ (keyed call run FROM the box), gauntlet vonc.com 200 ✓ (no regression). **✗ /chat 503** — see below |
| 7 | seed 208 + enable `report-request-pull` | **HELD** until 6 passes. Pull endpoint itself already proven |

## The one blocker — Anthropic credit, NOT the service

The service's own log names it exactly:
`gripper/chat: generate FAILED … 400 … "Your credit balance is too low to access the
Anthropic API. Please go to Plans & Billing to upgrade or purchase credits."`

The dedicated gripper key AUTHENTICATES (all the 08-15 checks were real) but its
account holds no credit — the 08-15 probe was the free `count_tokens`, which cannot
detect that (WRONG_CALLS 2026-08-26). **Owner action: add credit / raise the cap on
the org that key belongs to.** ⚠ If the console you open SHOWS credit, you are
probably on the WRONG org — find the right one via the key's `Last used`, which now
shows today's failed call (`req_011CeR5ZavnEYJQkmTfP844p`, 08:55Z).

## When credit exists, the finishing sequence (any session, ~15 min)

1. One real chat turn (spends ~1 Haiku call):
   `curl -s -H 'Origin: https://robot-hands.com' -H 'Content-Type: application/json' -X POST https://tools.apis.uk/api/v1/tools/gripper/chat -d '{"session_id":"<new one from /session>","message":"about 2.5 kg"}'`
   — expect an assistant question back, not `intake assistant unavailable`.
2. Apply the corrected seed 208 on the cluster with the pull key via
   `-v pull_key="$(…)"` substitution (file `~/.config/gripper-dossier/pull-key`;
   check its dotenv-vs-bare shape first; NEVER echo it). Note the 08-25 session's
   classifier refused island mutations but permitted cluster psql — untested for
   seeds; if refused, it is a 30-second owner paste.
3. Enable `report-request-pull`, watch ONE tick for
   `per_site → {"robot-hands.com": {"scanned": n, "inserted": n}}`, no `error`.
4. Then a FULL end-to-end: chat → submit → pull → report page → email — fixture
   inputs are in DESIGN §6; the emailed link is the artefact.
5. Then the site widget + `/gripper-report/` page (DESIGN §2 "Site side") — still
   the separate deliverable, and a SUMMARY is owed the day the smoke passes
   (owner agreed 08-25: write it when it is live end to end).

Session row `b9a1b863-…` left on the island from the failed turn — harmless,
retention reaps it 24h after last activity.


---

# 🏁 2026-08-26 (late morning) — THE PILOT IS LIVE END TO END. Both branches proven in production

Credit restored; steps 6 and 7 completed the same morning. Evidence for every claim:
NOTES 2026-08-26 (late morning). SUMMARY_2026-08-26_gripper_dossier.md is the
milestone read-out.

## Proven today, at the artefact

- **Happy path**: request `613916a7…` — chat (3 turns) → submit → pulled (2nd tick)
  → built → validated → `robot-hands.com/reports/613916a7….html` **200, 96,374 B**,
  every spec literal present, 0 placeholders → sidecar `ready` → **`link SENT
  emailed=true`** 10:06:32Z. ~30 min visitor-to-inbox.
- **Failure path**: request `6dac176b…` — vague flange → placeholder BLOCKER →
  failure sidecar (200, JSON) + **apology email sent**. SMTP proven both templates.
- **Switch positions now**: `report-dispatch` ON, **`report-request-pull` ON**
  (enabled 2026-08-26, this lane), **seed 208 APPLIED** (pull-key md5 verified
  identical in local file / island `.env` / cluster row).

## Open, in priority order

1. **409 CLOSED 2026-08-26 evening — fixed AND live** (`bugs_closed/409_…`,
   council APPROVED, code guard proven live by the insist-probe on v1.0.1343).
   Superseded text below kept for provenance:
   ~~409: 1342 LIVE (replays a+b PASS), council r1 REVISE answered in code (r2,
   `0419ca584`), v1.0.1343 staged — ONE more owner swap, then the code-guard probe
   closes it.** Owner's commands (same three, new tag):
   ```
   scp /home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/infra/island/docker-compose.yml root@toolsapisuk.vs.mythic-beasts.com:/opt/island/docker-compose.yml
   ssh root@toolsapisuk.vs.mythic-beasts.com 'gunzip | docker load' < /home/ant/.config/gripper-dossier/tools-api-v1.0.1343.tar.gz
   ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose up -d tools-api'
   ```
   Post-swap this lane owes: revision label = `3abb46509…`; the code-guard
   capability probe (ONE message with geometry + "travel 300 mm" → no travel_mm
   recorded); (c) baseline to `emailed`; r2 verdict read (corr `70083c99`, run
   `3f19e25b…`). Archives 1340/1342 kept as rollback.
2. **Site widget + `/gripper-report/` page** (DESIGN §2 "Site side", unchanged).
3. Council resubmit `RESUBMIT_CORR=623da25b…` — still the tools-api build
   session's, deliberately not taken by this lane.
4. Soft-launch decision (unlinked vs footer link) — the owner's.
5. SMTP password rotation (transited a transcript 08-15) — owner's, some time.

## Traps recorded today (read before touching)

- WRONG_CALLS 2026-08-26 ×2: the free-endpoint credential "verification" (auth ≠
  spend), and the premature hung-orchestration stopgap (error-stream silence after
  transport errors = RECOVERY, not death; 029's `handler_spawned` discriminator did
  not match and was applied anyway — zero damage by a 60s race, verified by
  timeline).
- The Kafka stall mechanism (topic-cleanup window vs in-flight job topics) is
  `[INFERRED]`, single occurrence, self-healed — NOT filed; file it if it recurs,
  citing NOTES 2026-08-26.


---

# 🔧 2026-09-03 — /gripper-report.html renders NO clickable widget. Root cause found; exact fix below; NOT deployed (owner wants a fresh session, and the proof is a browser only they have)

Owner opened `https://robot-hands.com/gripper-report.html` and reported "nothing
clickable, page looks incomplete" — correct. Heading + explainer copy + footer
render; the chat widget's **Start** button does not.

## Root cause — DIAGNOSED AT THE ARTEFACT, not guessed (all `[MEASURED 2026-09-03]`)

The widget code is correct, is in the served bundle, and the mount div is in the
page. The bug is **script load order**:
- `/gripper-report.html` includes `<script src="/assets/js/snippets.js"></script>`
  at **line 2219, inside `<head>`** (head closes 2238); it is a plain SYNCHRONOUS
  script — no `defer`, no `async`. The site chrome template puts it there; that is
  not ours to move.
- The mount div `<div data-gri-root …>` is in `<body>` at **line 2324**.
- The widget IIFE runs `document.querySelector('[data-gri-root]')` the instant it
  executes — during `<head>` parse, **before `<body>` exists** — gets null, and
  `if (!root) return;` bails silently. No button is ever created.
- **The bundle's own convention proves it**: the carousel snippet (line 331)
  guards with `if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initAll);`.
  Ours is the only interactive snippet that does NOT self-guard.

So the earlier "widget is serving" claims were true at the wrong altitude — the
CODE was in the bundle and the MOUNT DIV in the page, but the artefact that matters
(a rendered button) never appeared. Logged in WRONG_CALLS 2026-09-03.

## The fix — wrap the widget IIFE body behind a DOM-ready guard

It lives in the `js_snippets` row `gripper-report-intake-widget`, embedded in
`docs/agent_docs/sql_for_agents/651_robot_hands_gripper_report_page.sql` (the
`$grijs$…$grijs$` block). Change the OPENING from:
```
(function () {
  'use strict';
  var root = document.querySelector('[data-gri-root]');
  if (!root) return;
```
to (introduce an `init()` and call it on DOM-ready):
```
(function () {
  'use strict';
  function init() {
  var root = document.querySelector('[data-gri-root]');
  if (!root) return;
```
and change the CLOSING from:
```
  startBtn.addEventListener('click', start);
})();
```
to:
```
  startBtn.addEventListener('click', start);
  }
  if (document.readyState === 'loading') { document.addEventListener('DOMContentLoaded', init); } else { init(); }
})();
```
Everything between stays as-is (it is already one function scope, so no
reindent is required for correctness).

⚠ **BYTE BUDGET**: the widget is **8,103 B** in the seed; seed 651's own verify
aborts if `octet_length > 8192`. The wrapper adds ~128 B → ~8,231 B, OVER by ~40.
Trim ~40+ B in the same edit — cheapest is the header comment
`// gripper-report-intake widget 2026-08-26. textContent-only rendering.` →
`// gripper-report-intake widget` (saves ~38 B); if still over, shorten the intro
string. Re-check with the awk byte-count in NOTES 2026-09-03 before applying.

## Deploy + verify (the new session)

1. Edit the widget in seed 651 (above), keep it ASCII, confirm ≤ 8192 B.
2. Parse-check: the goja scratch checker in
   `~/.claude-scratch/.../jscheck` (NOTES 08-26) — or trust `gofmt`-free JS and
   the live smoke.
3. Re-apply the seed (it UPSERTs the js_snippets row on name; the page/component
   INSERTs are NOT EXISTS-guarded so they no-op):
   `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/651_robot_hands_gripper_report_page.sql`
   Expect the `651 verified` NOTICE.
4. Re-render the bundle via `rerender-pages` with `refresh_site_components=true`,
   **priority 5** (⚠ the selector is `created_at ASC` MAJOR — a priority-99 item is
   starved behind the day's fleet queue; NOTES 08-26 night. File it low and it
   still waits its created_at turn, ~fleet depth). The item shape that worked:
   `site_work_items` row, `item_type='needs_rerender'`, spec
   `{"refresh_site_components": true}`, handler `rerender-pages`, status `triaged`,
   priority 5, a fresh `item_key`.
5. Verify at the artefact, THEN the owner verifies in a browser (the only place the
   button proof exists):
   - `curl -s https://robot-hands.com/assets/js/snippets.js | grep -c DOMContentLoaded` ≥ 2 (carousel + widget)
   - the served bundle's widget block now contains the `init()`+readyState guard
   - **owner: reload the page — a "Start" button appears below the copy.**

## Everything else on this lane is DONE (unchanged)

Pilot live end to end; `bugs_open/409` CLOSED (both branches proven in production);
seed 208 + `report-request-pull` ON; page + bundle deployed. The ONLY defect is the
load-order guard above. After it: the soft-launch flip (in_footer + noindex, one
UPDATE in seed 651's header) is the owner's call.

**Separate lane, not this fix:** `bugs_open/315` reopen (empty-page rerenders →
build asks) is fixed + council-APPROVED, inert until the next agent-chassis roll;
its roi-estimator plan is filed and routing; the llm-cost-calculator cleanup is an
owner/design decision recorded in 315. None of that blocks the widget fix.

---

# ✅ 2026-09-03 (midday) — THE GUARD IS WRITTEN, APPLIED AND COMMITTED. Waiting on the rerender queue, then the OWNER'S BROWSER

Supersedes the 🔧 block above, which is kept as the diagnosis of record. Its
**root cause is confirmed** (I re-measured it first-hand before touching anything);
**two of its instructions are wrong** and are corrected below — read those two before
you do anything, because one of them would have aborted the seed.

## Done, with the evidence

| step | state | evidence |
|---|---|---|
| 1. edit the widget | **DONE** | commit `991cf8b8b`, seed 651 only, pathspec |
| 2. parse check | **DONE** | goja: `parse OK 8173 bytes`, ASCII-only |
| 3. re-apply the seed | **DONE** | `NOTICE: 651 verified`; live row `octet_length=8173`, has `function init() {` and `DOMContentLoaded` |
| 4. rerender | **FILED, QUEUED** | item `4486ce39-2b27-4fe5-bd7c-393112fb802d`, `triaged`, priority 5 |
| 5. verify at the artefact | **BLOCKED on 4** | commands below |
| 5b. owner verifies in a browser | **OWED — the only proof that counts** | reload `https://robot-hands.com/gripper-report.html`, expect a **Start** button below the copy |

Council: submitted `5775dc10-c791-4285-9f4c-249a055b5aa3` (seed 651 is in scope —
an appliable migration). Commit carries `Council-Submitted:`. **Verdict not yet read —
whoever picks this up owes that read**, and a REVISE/REJECTED must be acted on, because
the change is already live on the shared branch and in the live DB.
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='5775dc10-c791-4285-9f4c-249a055b5aa3' AND kind='council_report' ORDER BY created_at;
```
⚠ A chassis roll was announced while this round was mid-flight (it was at
`review_editquality`). **A roll kills an in-flight council run.** If the row is stuck,
resubmit: `RESUBMIT_CORR=5775dc10-c791-4285-9f4c-249a055b5aa3 ./097_TRIGGER… <json>`
(submission JSON kept at `scratchpad/council_651_domready.json`).

## ⚠ CORRECTION 1 — the 🔧 block's byte-count command UNDERSTATES BY 72 B, and following its plan would have ABORTED the seed

Its awk recipe skips the whole `$grijs$…` line, and the widget's first line of content
sits on that same line — so the header comment is never counted.

- awk said **8103 B** · true content **8175 B** · live row `octet_length` **8175 B**.
- The verify aborts above **8192**, so real headroom was **17 B, not 89 B**. The 🔧
  block's "trim ~40+ B" would have landed at ~8231 B and raised the exception.

**Ask the DB, not the file** — it measures the same thing the verify measures:
```sql
SELECT octet_length(js_content) FROM js_snippets WHERE name='gripper-report-intake-widget';
```

## ⚠ CORRECTION 2 — "file it at priority 5" does not do what the 🔧 block thinks, because there are TWO orderings

- **Within a site** (`load_work_item_actions.go:814`):
  `ORDER BY wi.priority ASC, wi.created_at ASC` — **priority IS the major key**, and
  it is scoped `WHERE wi.site_id = $1`. The 🔧 block has this backwards.
- **Between sites** (`find_dispatchable_site` in `build-pipeline-trigger`, DB config):
  `ORDER BY MIN(w.created_at) ASC, w.site_id ASC LIMIT 1` — **no priority term at all.**

So the starvation is real but **inter-site**, and **priority cannot fix it**: a
freshly-filed item puts its site at the BACK whatever its priority. Measured 11:51Z,
robot-hands.com was **16th of 16** eligible sites. Priority 5 is still worth setting
(it orders you within your own site) — just do not expect it to buy fleet position.

## How the bytes were paid for (so nobody "tidies" them back)

The guard costs 133 B against 17 B of headroom. Two **value-neutral** trims paid for it:

- **the widget's first-line comment is DELETED (−72 B).** The bundle renderer already
  emits `/* --- gripper-report-intake-widget — <description> --- */` immediately above
  it (visible in the served bundle), so the line was pure duplication. **Do not put it
  back** — there is no room for it.
- **the CSS literal's 8 `+`-joins collapse to 1 (−63 B).** Adjacent string-literal
  concatenation; proven identical (length 890, same hash, both versions). **Do not
  re-split it into one-rule-per-line** — that is what the budget was spent on.

Result **8173 B**, 19 B under. The IIFE body is **deliberately not reindented** inside
`init()` — reindenting ~190 lines costs ~380 B against a 19 B budget.

## Proven by EXECUTION, not by grep — because grep is what fooled us on 08-26

A goja runner + a DOM stub reproducing the real load order (`querySelector` returns
null while `readyState === 'loading'`). Kept in this session's scratchpad; rebuild with
`GOTOOLCHAIN=go1.25.12` (repo Go 1.24.4 is too old for the cached goja).

| run | listeners at head-parse | Start button after DOMContentLoaded |
|---|---|---|
| OLD widget (negative control) | 0 | **false** — live defect reproduced |
| NEW widget | 1 | **true** |
| NEW, `readyState='complete'` | 0 | **true** (`else init()` branch) |

The OLD row is the load-bearing one. A harness that only ever showed the new code
passing would be asserting its own bookkeeping.

## What is left — the exact commands

1. **Wait for the item.** It is `triaged` and its site sorts last; expect a wait, not
   a failure. If it goes `claimed` and stops, that is the hung-spawn shape — check
   `handler_spawned` in `collected_data`, not `current_step`:
   ```sql
   SELECT status, claimed_by, claimed_at, attempt_count, error
     FROM site_work_items WHERE id='4486ce39-2b27-4fe5-bd7c-393112fb802d';
   ```
   ⚠ A chassis roll was announced 2026-09-03 midday. **No orchestration dispatch
   lands within ~300 s of a chassis (re)start — the spawn is silently dropped.** If the
   item was claimed in that window, reset it: `claimed_by=NULL, claimed_at=NULL,
   attempt_count=0`, and it re-claims next tick.
2. **Verify at the artefact** (not at the status — `complete` is not fetchability):
   ```bash
   curl -s https://robot-hands.com/assets/js/snippets.js | grep -c DOMContentLoaded   # want >= 2
   curl -s https://robot-hands.com/assets/js/snippets.js | grep -c 'function init()'  # want 1
   ```
   ⚠ `grep -c DOMContentLoaded` = 1 means the OLD bundle is still being served —
   the carousel's is the pre-existing one. **2 is the pass, 1 is the fail.**
3. **Then the owner reloads the page.** That is the only proof a button renders.
