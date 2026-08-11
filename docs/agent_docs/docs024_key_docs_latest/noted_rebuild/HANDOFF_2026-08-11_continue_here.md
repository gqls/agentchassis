# HANDOFF — noted.co.uk rebuild, continue here

**Written 2026-08-11 17:30. Supersedes `HANDOFF_2026-08-10_continue_here.md`.**
Standalone: you should not need to read anything else to start.

Then, in order of usefulness: `README_where_we_are.md` (plain-prose history for
the owner), `PLAN_2026-08-10_noted_rebuild.md` (design + the decomposition
ruling), `RUNBOOK_noted_rebuild.md` (every command with its gotcha),
`NOTES_noted_rebuild.md` (technical log incl. every misstep).

---

## 1. What this is, in one paragraph

noted.co.uk is a note-taking app the owner hand-built: text, voice recordings,
photos, a little version history. It has been live since January out of the
framework's own B2 bucket, unknown to the framework. We are rebuilding it as a
**fully decomposed** framework site with a server-side backend, so a person can
sign in and reach their notes from any browser. The legacy app still serves
noted.co.uk with a wind-down notice; the new backend is **live and public** at
`app.noted.co.uk`; the framework build of the front end is **dispatched and
queued**. No user data has moved and nothing users depend on has broken.

## 2. THE FIRST THING TO DO

**Check whether the framework build has progressed.** It was dispatched
2026-08-11 17:00, correlation `59397ca9-c1c4-4938-8d2a-e78ffd7e045b`.

```sql
-- the spec cascade filling in
SELECT ss.aspect, COALESCE(ss.source_agent, ss.source) AS by, ss.created_at
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain='noted.co.uk' AND ss.is_current ORDER BY ss.created_at;

-- the relay
SELECT wi.item_type, wi.status, wi.handler_agent, wi.updated_at
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain='noted.co.uk' ORDER BY wi.created_at DESC;

-- pages
SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='noted.co.uk';
```

**State at 17:30 on 2026-08-11:** specs = `evidence_base`,
`imagery_style_guide` (both manual/pinned) + `submission`, `mission_brief`
(domain-submitter). `needs_domain_research` **`triaged`**, unclaimed. Pages: **0**.

> **⚠ DO NOT RESUBMIT.** `needs_domain_research` sat triaged for 20+ minutes
> because the fleet queue was **607 triaged / 3 claimed / 70 completed in 30
> min** — the pump is alive and this is queue depth. A missing row is latency.
> A retry costs a duplicate round.
>
> **⚠ Three items dated BEFORE the dispatch** (`needs_composition`,
> `needs_design`, `evaluate_tools`, all `detected`) are the **discovery rotation**
> picking the site up now a `sites` row exists — not this build. `detected` is
> not dispatchable, so nothing acts on them. Do not read them as build progress.

If the cascade has stalled rather than queued, the manual pump heartbeat is
`scripts/initial_messages/020_build_pipeline/076_trigger_build_pipeline.sh`.

> **CORRECTED 2026-08-11 18:30 — the instruction above stands; the reason was
> wrong, and the remedy does not apply to us.** At 90 minutes the state was
> unchanged (0 pages, still `triaged`, `attempt_count = 0`). It is **not** fleet
> queue depth. `build-pipeline-trigger`'s `find_dispatchable_site` step ends
> `ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1` — **one
> estate-wide line, oldest work item first, one site at a time.** Our item joined
> that line at 17:00 today.
> `[MEASURED 18:26]` **589 pending items older than ours across 19 sites**,
> draining at **~95/hour** (steady 13:00–17:00) ⇒ **ETA ~6 hours**. The trigger
> fires every ~2.5 min and is healthy (195 runs).
> - **Waiting is correct and the position can only improve**: ordering is by
>   `created_at`, so all new estate work sorts *behind* us. (Not absolute — an
>   older item currently `deferred`/`blocked`/`detected` would land ahead if
>   triaged.)
> - **Resubmitting cannot help** — now checked, not assumed:
>   `refreshOpenWorkItemSQL()` (`load_work_item_actions.go:1417`) sets
>   `updated_at` and **never `created_at`**, so a retry either no-ops or sorts
>   further back.
> - **Do NOT hand-fire `076` for this.** It runs the same trigger with the same
>   `ORDER BY`, so it re-picks the same globally-oldest site regardless of who
>   fires it. It is the remedy for a pump that is *not firing* — check
>   `orchestration_states` for recent `build-pipeline-trigger` rows first.
> - `wi.priority` is a **tiebreak within one `created_at`**, so it is not a lever
>   for jumping the queue either.
>
> Full working, plus two wrong turns of mine (an absent k8s CronJob that did not
> mean "unscheduled", and a drain figure that oscillated because a `claimed` item
> hides its whole site from the count): `NOTES_noted_rebuild.md`, 2026-08-11 18:30.

## 3. What is LIVE

### noted.co.uk — the legacy app (repo `gqls/sites`, branch **`master`**)
The original app, working, plus:
- **A large wind-down notice**: "Noted is being refreshed in the next couple of
  days", stating that **we hold no copy** and that if the notes are lost there is
  nothing to restore them from. Deliberately does **not** say the notes will be
  deleted — that would be a false claim about our own behaviour.
- **"Save everything"** — a backup including recordings, photos and history. The
  old button saved **text only** (a note with a 4 KB audio blob and an 8 KB image
  produced a **334-byte** file). Round-trip tested.
- Four WCAG-AA contrast fixes found by `scripts/render_audit.py`.

### app.noted.co.uk — the engine (live, public, HTTPS)
`{"status":"ok"}` from the open internet. Accounts, sessions, notes, media, and
an import accepting **exactly** the "Save everything" format already on people's
disks. Verified publicly end to end: register → save → import → **sign in on a
new session and the notes are there with the recording**.

On `webdesign.vs.mythic-beasts.com` (176.126.243.62), key
`~/.ssh/webdesign_box_ed25519`:
- `noted-engine` — Go binary, own unprivileged user, `ProtectSystem=strict`,
  empty `ReadWritePaths=`, bound **`127.0.0.1:8090` only**.
- Postgres 16.14, `127.0.0.1:5432` only, db+role `noted`, creds
  `/etc/noted/noted.env`. Cross-database `CONNECT` revoked from `PUBLIC`.
- nginx on `127.0.0.1:8082`; doc root `/var/www/noted.co.uk` (**empty until the
  framework build publishes** — deliberately).
- `sitesync` pulls **both** `webdesign.uk` and `noted.co.uk` from `gqls/vm-sites`
  every 5 min (generalised to a list 2026-08-11; backup at `sitesync.bak-20260811`).
- Backups: nightly 03:20 → `pg_dump` → **age-encrypted** → B2
  `personae-noted-backups` (Object Lock, governance 30d). Includes
  `webdesign-chat`'s data. **Timer fired unattended at 03:30 — proven.**
  **Restore drill PASSES** from `~/.config/noted/noted-backup-age-identity.txt`.

### In the framework
`sites` row with `github_repo='vm-sites'`, `deploy_config={"target":"vm",
"capabilities":["backend"]}`; `evidence_base` (7 bans, 0 facts, **pinned**);
`imagery_style_guide` (**pinned**); two **experience patterns** (§5).

## 4. THE NEXT BIG PIECE — the front end

There is **no UI yet**, on purpose. Do not hand-write one: it would violate both
"EVERY SITE GOES THROUGH THE FRAMEWORK" (owner, 2026-08-04) and this lane's
**decomposition** requirement (owner, 2026-08-10: *"locked can mean many things.
I just want it all controlled fully by the framework — upgrades, maintenance,
tools checking, everything"*). That rules out `082 --fidelity locked` /
`adopt_verbatim`, which stores each page as one `ported-page` component at
`rebuild_policy='owned'` — deployed and monitored but never re-planned.

When the build lands:
1. **Bind the two experience patterns** to the built components
   (`site_experiences`, `proposed → bound → verified`).
2. The editor becomes a `component_level='tool'` component; its JS deploys as a
   real asset (`/tools/assets/{fn}.js`) calling `/api/` same-origin.
3. **Build `/legacy`** — see §6, this is the migration.
4. `rebuild_policy` stays `generic` everywhere. **No `owned` pages.**
5. Cut over noted.co.uk from the bucket to the box, keeping the legacy app
   reachable for a grace period.

## 5. The experience patterns (applied, `status=draft`)

| pattern | contract | degraded | checks |
|---|---|---|---|
| `authenticated-note-sync` | 3 | 2 | 12 |
| `legacy-local-data-adoption` | 3 | 2 | 6 |

Source: `EXPERIENCES_2026-08-11_noted_patterns.sql`.

The load-bearing clause: **the editor may say "Saved" only as a consequence of a
successful server response.** The legacy app set that indicator locally — a false
"Saved" is the worst lie this product can tell, because it is the moment someone
stops worrying and closes the tab.

**Three things the runner genuinely cannot check** (recorded in `_unsupported`
at each check — do not read a green run as covering them):
- `sign_in_round_trip` is **flaky by construction**: no `expect_within_ms`, the
  runner asserts 300 ms after the step, a sign-in is a network round trip. A
  failure is not proof of a broken sign-in. Do not promote it to a gate.
- `notes_survive_a_reload` **cannot be ordered** after sign-in. The real
  assertion (same notes in a second independent session) is inexpressible today
  and lives in `box/smoke-test.sh`.
- The legacy page's actual behaviour is **inexpressible**: the runner cannot seed
  IndexedDB. Covered by a Playwright probe.

**An unsupported check type is INERT, not an error.** The verify query at the
bottom of the SQL file is what catches a typo; run it after any edit.

`funnel_stage` is CHECK-constrained to `awareness|consideration|conversion`, so
neither `retention` nor `onboarding` exists. Used `conversion`; it is the least
wrong, not the right answer.

## 6. THE MIGRATION IS EASIER THAN PLANNED — build `/legacy`

`[MEASURED, browser-verified 2026-08-11]` A **different page on the same origin**
reads the existing app's `NotedDB` — notes, content, and audio, all four stores.

IndexedDB is keyed by **origin**, and the origin is `https://noted.co.uk`
whether the bytes come from the bucket or the box. **So at relaunch, everyone's
notes are still in their browser and the new site can read them directly.** The
manual export/import is a **fallback**, not the migration.

`/legacy` must: open `NotedDB` **without a version number** (naming a version
triggers `onupgradeneeded` — a migration against someone's only copy); be
**read-only, never delete**; show counts **before** asking for anything; offer a
download route **without an account**; and tolerate the database being absent.

**Re-run the probe at cutover.** Origin is scheme+host+port; a hostname change, a
`www.` redirect or a scheme change silently invalidates this.

## 7. OPEN ITEMS

| # | Item | Note |
|---|---|---|
| 1 | **The privacy copy** | Owner's to write. The old "no server" wording is banned in `evidence_base` so no agent copies it forward. **A replacement I proposed was rejected — do not reintroduce it.** Owner asked for copy that avoids where anything is stored and focuses on what the site does for people. |
| 2 | **`voicescan` cannot run** | Needs a voice spec; noted has none until the classifier runs. Re-check once the build lands. |
| 3 | **Degraded states unverified** | The platform cannot induce a failing dependency. "Save fails loudly, text survives" must be exercised **by hand** before launch. It is the clause protecting the unrecoverable thing. |
| 4 | **`cloudflared.service` has no `Restart=`** | Any signal that stops it is an outage until a human notices. webdesign lane's file — raised, not changed. |
| 5 | **Off-box backup of `webdesign-chat`** | Now included in noted's backup. That lane may not know. |
| 6 | **Owner accepted the key copies as they are** | Four copies incl. email. Closed at his direction; do not reopen. |

## 8. TRAPS — read before touching anything

- **`sites.github_repo='vm-sites'` is LOAD-BEARING SAFETY.** The default routes to
  `gqls/sites` → `b2 sync --delete` → **the prefix the live app serves from**. A
  build on the default repo would delete the running application. Do not "tidy" it.
- **A wildcard Worker route owns every subdomain.** `*.noted.co.uk/*` →
  `portfolio-sites-router` answered before the tunnel was ever reached, while
  *every* check on the box passed. Fixed with `app.noted.co.uk/*` → *(no worker)*.
  **This will hit every tunnel hostname added to any of the 36 B2-fronted zones.**
- **`systemctl kill -s HUP cloudflared` TERMINATES it.** It does not reload, and
  the unit has no `Restart=`. This cost a ~40s outage on the live shopfront.
- **This box serves a live commercial shopfront.** Before/after control every
  time. The **box-local** check is the real one (`Host: webdesign.uk` →
  `127.0.0.1:8080`, baseline `200 / 28419 B`) because `webdesign.uk` externally
  302s to `webdesign.co.uk` **by design** (owner-confirmed — it is NOT a fault).
- **`webdesign-chat` binds `*:8081`**, not loopback; ufw is its only control.
  Never copy that. `noted-engine` binds loopback.
- **`pg_dump -f` fails** into the 700 root backup dir (it runs as `postgres`);
  dump to stdout and redirect as root. **`pg_restore` as `postgres` reports a
  valid dump as corrupt** because it cannot read a 600 root file — verify as root.
- **B2 `writeFiles` includes HIDE.** A hidden backup vanishes from `b2 ls`
  entirely, so hidden and never-uploaded look identical. Always `--versions`.
- **`cloudflared tunnel route dns` silently appends its cert's zone** on a
  cross-zone hostname and reports success. Check what it created, at the
  **authoritative** nameserver.
- **Test suites skip loudly without `NOTED_TEST_DSN`** — a silent skip prints
  `ok`. Run with `-v`.
- **The shell working directory persists between tool calls.** Two commands here
  failed because a relative `cd` from an earlier call was still in effect. Use
  absolute paths.

## 9. MY ERROR RECORD — read before trusting anything above

Seven recorded wrong calls across two sessions (`WRONG_CALLS.md`), in four
families:

1. **A backwards `diff`** → concluded the live site was crashing. Refuted by
   running the page.
2. **Four false negatives from my own instruments** — `pg_restore` as the wrong
   user, a JSON path typo, a deny-test that "passed" because the subcommand did
   not exist, and grepping for expected error strings.
3. **Joining an unexplained observation to an unlocated fault** — I called the
   webdesign.uk 302 the shopfront's breakage. It is intentional. *Two unexplained
   things are not evidence for each other.* **The actual shopfront breakage
   remains uncharacterised — ask the owner what the symptom is.**
4. **Every check passing while the thing was broken** — the Worker case. Passing
   checks *bound* a fault; they do not locate it. All three of mine tested the
   box, and the fault was in front of it.

The habits that caught these: print whole output before grepping; verify by a
differently-shaped second route; make every negative control fail once on
purpose. The mutation test in `engine_test.go`, the induced hide in the backup
monitor, and the decoy-identity run of the restore drill all exist for this reason.

## 10. COMMANDS

```bash
# the box
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com

# engine health, publicly and locally
curl https://app.noted.co.uk/api/health
ssh … 'curl -s http://127.0.0.1:8090/api/health'

# full stack smoke test (register → save → import → second session)
ssh … 'bash /tmp/smoke.sh'      # source: box/smoke-test.sh

# backups
ssh … 'systemctl start noted-pg-backup.service'
docs/.../noted_rebuild/box/check-noted-offsite-backups.sh      # off-box monitor
docs/.../noted_rebuild/box/restore-drill.sh ~/.config/noted/noted-backup-age-identity.txt

# engine tests (needs docker for a throwaway postgres)
cd docs/.../noted_rebuild/box/noted-engine
NOTED_TEST_DSN="postgres://postgres:test@127.0.0.1:55432/postgres?sslmode=disable" go test -v ./...

# rebuild + ship the engine (no Go on the box)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o noted-engine .
scp -i ~/.ssh/webdesign_box_ed25519 noted-engine root@webdesign.vs.mythic-beasts.com:/tmp/
ssh … 'install -m755 /tmp/noted-engine /usr/local/bin/noted-engine && systemctl restart noted-engine'

# copy gate (what the platform will flag)
go run ./cmd/claimscan -evidence /tmp/noted_evidence.json -components <tsv>

# contrast (hand-run only; defaults to a 390px MOBILE viewport — run both)
/home/ant/.venvs/vonc_pw/bin/python scripts/render_audit.py <url> --width 390
```

## 11. FILES

`docs/agent_docs/docs024_key_docs_latest/noted_rebuild/`

| file | what |
|---|---|
| `HANDOFF_2026-08-11_continue_here.md` | **this file** |
| `README_where_we_are.md` | owner's plain-prose log, append-only |
| `PLAN_2026-08-10_noted_rebuild.md` | design, decomposition ruling, phasing |
| `PLAN_2026-08-10_box_backup.md` | backup design + the B2 capability probe |
| `RUNBOOK_noted_rebuild.md` | commands, each with its gotcha |
| `NOTES_noted_rebuild.md` | technical log incl. every misstep |
| `SUMMARY_2026-08-10*.md` | two milestone read-outs |
| `SEED_2026-08-10_noted_site_and_specs.sql` | applied — site row + specs |
| `EXPERIENCES_2026-08-11_noted_patterns.sql` | applied — the two patterns |
| `MISSION_2026-08-11_noted.txt` | the brief that was dispatched |
| `box/noted-engine/` | the Go engine + tests + schema |
| `box/*.sh, *.service, *.timer, *.nginx, sitesync` | everything on the box |
