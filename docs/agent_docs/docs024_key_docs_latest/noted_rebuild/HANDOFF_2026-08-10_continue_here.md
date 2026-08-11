# HANDOFF — noted.co.uk rebuild, continue here

Written 2026-08-10 end of session 1. **Read this first, then
`README_where_we_are.md` (plain prose history), then
`PLAN_2026-08-10_noted_rebuild.md` (design + the decomposition ruling).**

Chassis at the time of writing: **v1.0.1283** (fresh build deployed this
session; nothing in this lane depends on it yet — the engine is outside the
chassis entirely).

---

## 1. The one-paragraph state

noted.co.uk is a hand-built browser-only notes app that has been live since
January out of the framework's own B2 bucket, unknown to the framework. We are
rebuilding it as a framework site with a server-side backend so people can sign
in and reach their notes from any browser. **The legacy app is still what
noted.co.uk serves** (plus a wind-down notice we added). **The new backend is
built, tested and running on a VM but is not publicly reachable.** No user data
has moved, and nothing users depend on has changed.

## 2. What is LIVE right now

### On noted.co.uk (the legacy app, in `gqls/sites`, branch `master`)
- The original app, unchanged in function.
- **A wind-down notice** telling people the app is being rebuilt and to save a copy.
- **"Save everything"** — a backup that includes voice recordings, photos and
  version history. The old button saved *text only*; a note carrying a 4 KB audio
  blob and an 8 KB image produced a **334-byte** file. Round-trip tested
  (export → wipe IndexedDB → restore → media byte-identical incl. MIME type).
- Three WCAG-AA contrast fixes found by `scripts/render_audit.py`.
- The app is in version control for the first time.

### On `webdesign.vs.mythic-beasts.com` (176.126.243.62)
Key: `~/.ssh/webdesign_box_ed25519`. **Shared with the LIVE webdesign.uk
shopfront — always take a before/after control (RUNBOOK).**

- **Postgres 16.14**, `127.0.0.1:5432` only, database + role `noted`, creds in
  `/etc/noted/noted.env` (600). Cross-database `CONNECT` revoked from `PUBLIC`.
- **`noted-engine`** — Go binary at `/usr/local/bin/noted-engine`, systemd unit,
  own unprivileged user, `ProtectSystem=strict`, empty `ReadWritePaths=`,
  bound **`127.0.0.1:8090` only** (verified unreachable off-box).
- **nginx** site on `127.0.0.1:8082`, `/api/` → engine, doc root
  `/var/www/noted.co.uk` (**empty — deliberately**, see §4).
- **Backups**: nightly `pg_dump` 03:20 UTC → age-encrypted → B2
  `personae-noted-backups` (Object Lock, governance 30d). `webdesign-chat`'s
  `/var/lib/webdesign-chat` included. **Restore drill PASSED end to end.**

### Proven, not assumed
`box/smoke-test.sh` on the box: register → save a note → import the real
"Save everything" backup → **sign in on a fresh session and both notes are
there with the recording attached** → a second account sees **0** of them.

## 2b. UPDATE 2026-08-11 — the engine is PUBLIC

- **`https://app.noted.co.uk/api/health` → `{"status":"ok"}` from the open
  internet.** Full flow verified publicly: register (cookie `HttpOnly; Secure;
  SameSite=Lax`) → save a note → import the real backup format → sign in on a
  **new session** and see both notes with the recording.
- Owner authorised the `noted.co.uk` zone; DNS route created and verified at the
  **authoritative** nameserver, not from the command's success message.
- **Worker exclusion added:** `app.noted.co.uk/*` → *(no worker)*. The
  `*.noted.co.uk/*` route sends everything to `portfolio-sites-router`, which was
  answering before the tunnel was ever reached. **This will hit every tunnel
  hostname added to any of the 36 B2-fronted zones.**
- Orphan B2 key `…000c` **deleted**; `…000d` still deployed and **proven still
  working** by running the full backup afterwards.
- **The nightly timer fired unattended at 03:30** — first evidence the schedule
  works without a human. Three backups now in the bucket.
- **Restore drill PASSED from the owner's stored copy** at
  `~/.config/noted/noted-backup-age-identity.txt` (note: `.config`, not
  `.configs`), restoring the real 4-table schema into a throwaway container.
- Stray test account `smoke2@example.com` found in the live database and removed.
- Notice on noted.co.uk enlarged and made firmer, with the "next couple of days"
  timeframe. Ran through `claimscan`: **0 findings**, against a negative control
  that draws 4.
- **MIGRATION IS EASIER THAN PLANNED**: proven that a different page on the same
  origin reads the existing `NotedDB`. At relaunch everyone's notes are still in
  their browser, so `/legacy` can read them directly and the manual export is a
  fallback, not the route.

### ⚠ Key hygiene — outstanding
The age identity now exists in **four** places: `~/noted-backup-age-identity.txt`,
`~/.config/noted/`, `~/.ssh/`, and the owner's **email**. That key decrypts every
note in every backup. Email is the weak one (searchable archive, synced to
devices). Consolidate to one working copy plus one offline/password-manager copy,
and delete the rest.

## 3. DECISIONS WAITING ON THE OWNER

| # | Decision | Why it is blocking / urgent |
|---|---|---|
| 1 | **Delete the stray DNS record `app.noted.co.uk.webdesign.uk`** | I created it by accident (§5) and cannot remove it. Harmless but it is clutter in a live zone. |
| 2 | **Authorise the `noted.co.uk` zone for the tunnel** | **Hard blocker.** Until then the engine is reachable only from the box. Either `cloudflared tunnel login` selecting that zone, or a CF API token with DNS edit. |
| 3 | **Move the age identity off this workstation** | `~/noted-backup-age-identity.txt` is the **only copy in existence**. Lose it and every off-box backup is permanently unreadable, silently, until it matters. Needs password manager + one offline copy, then re-run the drill from the stored copy. |
| 4 | **Revoke the unused B2 key `…000c`** | A live credential with write access to the backup bucket that no system holds. |
| 5 | **What the privacy copy says** | The old site's pitch is "no server"; that becomes false at sign-in. The old wording is banned at the framework level so it cannot be copied forward. **A specific replacement I proposed was rejected — do not reintroduce it.** This is the owner's to write. |
| 6 | **Front-end route** (see §4) | Determines the whole next phase. |
| 7 | **Migration**: nobody's notes can be moved for them | No server-side copy exists and we cannot reach their browsers. Only path is a person exporting and importing. Affects how long the legacy app stays up. |
| 8 | **Whether media stays in Postgres** | I revised the plan (§6) — flag if you disagree. |

## 4. THE NEXT BIG PIECE — the front end, and why I stopped

**There is no user interface for the new app.** The engine has no HTML; the nginx
doc root is empty on purpose.

I deliberately did **not** hand-write one. Two owner rulings converge here:
"EVERY SITE GOES THROUGH THE FRAMEWORK, never hand-build one" (2026-08-04), and
this lane's own **decomposition** requirement — the site broken into the
framework's native parts, not delivered as opaque blobs. A hand-written SPA
would violate both and would be the fastest way to end up with exactly the
un-upgradable artefact we are replacing.

**So the next phase is the framework build**, and it is substantial:
1. Write the **experience patterns** for the app's behaviours before building
   (`authenticated-note-sync`, `local-first-capture`, `media-attachment-capture`,
   `backup-export-restore`) — contracts + `criteria_template`, so the checking
   layer can verify them. **This is what makes "the framework checks it" true.**
2. Mission brief → `082_submit_domain_unified.sh` → the relay builds the prose
   pages, decomposed into sections.
3. The editor as a `component_level='tool'` component, its JS deployed as real
   assets (`/tools/assets/{fn}.js`), calling `/api/` same-origin.
4. `rebuild_policy` stays `generic` everywhere. **No `owned` pages.**

**Read PLAN §5 first** — the criteria runner cannot order checks, has no
`expect_within_ms`, no retries, and cannot induce a failing dependency. The
behaviours most worth checking here are ordered and stateful, so some of this may
require a genuine platform contribution.

## 5. TRAPS SPECIFIC TO THIS LANE

- **`sites.github_repo='vm-sites'` is LOAD-BEARING SAFETY.** The default repo
  routes to `gqls/sites` → `b2 sync --delete` → **the prefix the live app serves
  from**. A framework build on the default repo would delete the running
  application. Do not "tidy" it. (Also in LANDMINES.md.)
- **This box serves a live commercial shopfront.** Before/after control every
  time; the **box-local** check is the real one (`Host: webdesign.uk` →
  `127.0.0.1:8080`, baseline `200 / 28419 B`) because `webdesign.uk` externally
  302s to `webdesign.co.uk` **by design** (owner-confirmed).
- **`webdesign-chat` binds `*:8081`** — not loopback; ufw is its only control.
  Not ours to fix, but never copy the pattern. `noted-engine` binds loopback.
- **`pg_dump -f` fails into the 700 root backup dir**; dump to stdout and let
  root's shell redirect. **`pg_restore` as `postgres` reports a valid dump as
  corrupt** because it cannot read a 600 root file — verify as root.
- **`writeFiles` on a B2 key includes HIDE.** A hidden backup vanishes from
  `b2 ls` entirely, so hidden and never-uploaded look identical. Always list
  `--versions`. Never set a short `daysFromHidingToDeleting`.
- **`cloudflared tunnel route dns` silently appends the cert's zone** to a
  cross-zone hostname and reports success. Check what it created.
- **Test suites here skip loudly without `NOTED_TEST_DSN`.** A silent skip prints
  `ok`. Run with `-v` and confirm the tests actually executed.

## 6. THINGS I CHANGED MY MIND ABOUT (do not re-derive)

- **Media lives in Postgres, not B2** — revises PLAN §3a-i. The disk-coupling
  concern was right; a hard per-account quota enforced in the same transaction as
  the insert *bounds* growth where a different backend only *relocates* it, and it
  keeps media inside the nightly encrypted dump. Migration path (`storage_key`
  column) left open in `schema.sql`. Revisit past a few GB.
- **The webdesign.uk 302 is NOT a defect** — owner-confirmed intentional. An
  earlier NOTES entry said otherwise and is corrected in place. **The actual
  shopfront breakage remains uncharacterised; ask what the symptom is.**

## 7. MY OWN ERROR RATE THIS SESSION — read before trusting anything above

Six recorded wrong calls in one session (`WRONG_CALLS.md`), in three families:

1. **A backwards `diff`** → concluded the live site was crashing. Refuted by
   running the page. *Confidence tracked how interesting the conclusion was.*
2. **Four false negatives from my own instruments** — `pg_restore` as the wrong
   user, a JSON path typo, a deny-test that "passed" because the subcommand did
   not exist, and grepping for expected error strings. *A check that fails for
   its own reasons is indistinguishable from the thing under test failing.*
3. **Joining an unexplained observation to an unlocated fault** (the 302).
   *Two unexplained things are not evidence for each other.*

The habits that caught them: print the whole output before grepping it; verify by
a differently-shaped second route; make every negative control fail once on
purpose. The mutation test in `engine_test.go` and the induced hide in the backup
monitor exist because of this.

## 8. FILES

| Path | What |
|---|---|
| `README_where_we_are.md` | plain-prose history, owner's document, append only |
| `PLAN_2026-08-10_noted_rebuild.md` | design, decomposition ruling, phasing |
| `PLAN_2026-08-10_box_backup.md` | backup design + the B2 capability probe |
| `RUNBOOK_noted_rebuild.md` | every command, each with its gotcha |
| `NOTES_noted_rebuild.md` | technical log incl. every misstep |
| `SUMMARY_2026-08-10*.md` | two milestone read-outs |
| `SEED_2026-08-10_noted_site_and_specs.sql` | applied; site row + specs |
| `box/noted-engine/` | the Go engine + tests |
| `box/*.sh`, `box/*.service`, `box/*.timer` | backups, drill, smoke test |

**Deployed sites row:** `noted.co.uk`, `github_repo='vm-sites'`,
`deploy_config={"target":"vm","capabilities":["backend"]}`, `evidence_base`
(7 bans, 0 facts, pinned) + `imagery_style_guide` (pinned).
