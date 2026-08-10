# RUNBOOK — noted.co.uk

Commands that were hard to get right, each with its gotcha attached. When one
changes, change it HERE.

---

## Reading the live site

**The live bytes are ground truth; the local `~/projects/domains/noted.co.uk/`
generations are NOT.** There are three local generations (`01`,
`02 - voice notes`, `03-sharing`) and the live site matches **none of them
exactly** — it sits between `02` and `03`.

```bash
export PATH="$HOME/.local/bin:$PATH"
b2 sync --no-progress b2://portfolio-sites/noted.co.uk ./live_noted
```

> **GOTCHA — do not conclude which generation is live from a `diff` alone.**
> That is exactly the mistake logged in `WRONG_CALLS.md` on 2026-08-10: in
> `diff live local`, a `<` line is present in **live**. Getting that backwards
> produced a confident, entirely fictional "the site is broken" finding.
> **Assert the thing directly instead** — `grep -c btn-share live_noted/index.html` —
> and confirm the file you are reading is the file being served:
> ```bash
> sha256sum live_noted/index.html   # must match what the browser receives
> ```

## Proving what the live app actually does

Reading the source is not enough; this app's defects live in what runs.
Playwright is at `/home/ant/.venvs/vonc_pw/bin/python`.

```bash
/home/ant/.venvs/vonc_pw/bin/python scratchpad/probe_backup.py
```

The reusable technique: **plant state through the app's own storage module,
then drive the app's own UI, then read what lands on disk.**

```python
planted = page.evaluate("""async () => {
    const { Storage } = await import('/js/storage.js');
    await Storage.saveNote({ id, title, content });
    await Storage.saveAudio(id, [new Blob([new Uint8Array(4096).fill(7)], {type:'audio/webm'})]);
}""")
```

> **GOTCHA — a Playwright download has no file extension.** `dl.value.path()`
> returns a temp artefact with no name, and `importer.js` dispatches on
> `.json` / MIME type, so a restore test silently imports **nothing** and every
> assertion after it fails for the wrong reason. Use
> `dl.value.save_as("something.json")`. This cost one confusing red run.

## Deploying a change to the LEGACY app

It deploys from the `gqls/sites` repo, **branch `master`** (not `main`).

```bash
git clone --depth 1 --branch master git@github.com:gqls/sites.git
# ... edit noted.co.uk/ ...
git add noted.co.uk && git commit noted.co.uk -m "..."
git push origin master
gh run list --repo gqls/sites --limit 3     # ~25s
```

The Action syncs only **changed** domains with
`b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"`, then
purges that domain's Cloudflare zone cache.

> **GOTCHA 1 — `--delete` means your directory must be COMPLETE.** Anything in
> the bucket but not in your commit is removed. Base every change on a fresh
> `b2 sync` down, not on a local generation.
>
> **GOTCHA 2 — another session will push between your clone and your push.**
> It happened within the hour on 2026-08-10 (an unrelated rerender). Check
> whether the incoming commits touch your files before rebasing:
> ```bash
> git fetch --unshallow origin master
> git log --oneline HEAD..origin/master -- noted.co.uk   # empty = safe
> git rebase origin/master
> ```

## Verifying a deploy — at the artefact, with a NEGATIVE control

A positive grep alone cannot distinguish "deployed" from "cache served me the
old page that happened to contain that string". Always pair it.

```bash
/home/ant/.venvs/vonc_pw/bin/python scratchpad/verify_live.py
```

- **Positive:** `rebuild-notice`, `noted.co.uk/full-backup` present.
- **Negative:** `Backup text notes`, `color: #777;`,
  `.danger { color: var(--danger); }` **absent** — strings the change REMOVED.
- **Behavioural:** plant a recording, click Save everything, assert the
  downloaded JSON contains it.

## Running the contrast auditor

`scripts/render_audit.py` (VIZ-010) is hand-run only — nothing invokes it.

```bash
cd sites_repo/noted.co.uk && python3 -m http.server 8779 --bind 127.0.0.1 &
/home/ant/.venvs/vonc_pw/bin/python scripts/render_audit.py \
    http://127.0.0.1:8779/ --width 390
```

> **GOTCHA — it defaults to a 390px MOBILE viewport.** Desktop found 3 failures
> where mobile found 1, because the sidebar small print is hidden on mobile.
> **Run both widths**, and both pages (`/` and `/guides/about.html`).

## The seed

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  < docs/agent_docs/docs024_key_docs_latest/noted_rebuild/SEED_2026-08-10_noted_site_and_specs.sql
```

Applied out of band, **not** via the migration runner — per-site setup, not a
schema change. Idempotent (`ON CONFLICT`, supersede-then-insert).

> **GOTCHA — `github_repo='vm-sites'` is load-bearing safety, not tidiness.**
> With the default repo, the first framework build of this domain would
> `b2 sync --delete` over the **live application**. See the seed file header.

Check the bans actually match the sentences they were written for — a ban that
matches nothing is the silent failure (an invalid regex degrades to a literal
substring):

```sql
SELECT b->>'reason',
       'we can''t see your notes, read your text, or listen to your recordings'
         ~* (b->>'pattern') AS catches_it
FROM site_specs s, LATERAL jsonb_array_elements(s.data->'banned_claims') b
WHERE s.site_id = (SELECT id FROM sites WHERE domain='noted.co.uk')
  AND s.aspect='evidence_base' AND s.is_current;
```

## Traffic — do not try to measure it from B2

`personae-access-logs/portfolio-sites/` **stops on 2026-05-10** (6,752 objects,
first 2026-03-14), and every row has `-` for user-agent because Cloudflare is
the client. Human traffic is **not measurable from here**; most of what is
logged is scanner noise (`wp-signup.php`, `secrets.json`). Do not report a user
count from this source, and do not read the absence as "no users".

---

## The box: `webdesign.vs.mythic-beasts.com` (176.126.243.62)

```bash
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com
```

Ubuntu 24.04.4, 2 cores, 7.8 GB RAM, 50 GB disk. Shared with the **live
webdesign.uk shopfront** — nginx on `127.0.0.1:8080`, `webdesign-chat` on
`*:8081`, `cloudflared` tunnel out. Postgres 16.14 added 2026-08-10 for noted.

> **ALWAYS take a before/after control when you touch this box.** It serves a
> live commercial front door that is not this workstream's. The cheap pair:
> ```bash
> curl -sL -o /tmp/wd_before.html https://webdesign.uk/ && sha256sum /tmp/wd_before.html
> ssh … 'curl -s -o /dev/null -w "%{http_code} %{size_download}\n" -H "Host: webdesign.uk" http://127.0.0.1:8080/'
> ```
> The **box-local** one is the real control — `https://webdesign.uk/` currently
> 302s to `webdesign.co.uk` (a different site, served from the bucket), so an
> external check can pass while the box is broken. Baseline as of 2026-08-10:
> external `200 / 25970 B / sha 26293b6d…`; box-local `200 / 28419 B`.

### Postgres on this box

Cluster `16/main`, **`listen_addresses=localhost`**, bound `127.0.0.1:5432`,
verified unreachable from outside. Database `noted` owned by role `noted`;
credentials in `/etc/noted/noted.env` (mode 600, root) as `NOTED_DATABASE_URL`.
The password was generated on the box and has never left it.

```bash
psql "$(grep NOTED_DATABASE_URL /etc/noted/noted.env | cut -d= -f2-)"
```

> **GOTCHA — Postgres grants `CONNECT` to `PUBLIC` on every database by
> default**, so a freshly created service role can open *any* database on the
> cluster and read its catalogue. This box is now shared, so that was closed:
> `REVOKE CONNECT ON DATABASE postgres, template1 FROM PUBLIC`. **Do the same for
> every database added here**, and verify with a negative control — the `noted`
> role must be *refused* on `postgres`:
> ```bash
> PGPASSWORD=… psql -h 127.0.0.1 -U noted -d postgres -c 'SELECT 1'
> # expect: FATAL: permission denied for database "postgres"
> ```
> This failed the first time it was checked; the fix is only trustworthy because
> the check was re-run and *changed answer*.

### Backups

`/usr/local/sbin/noted-pg-backup.sh`, run by `noted-pg-backup.timer` daily at
03:20 UTC + up to 15 min jitter. `pg_dump -Fc` to `/var/backups/noted/`
(700 root), 14-day retention, pruning only after a size-checked good dump.

```bash
systemctl start noted-pg-backup.service   # run now
journalctl -u noted-pg-backup.service -n 5 --no-pager
systemctl list-timers noted-pg-backup.timer --no-pager
```

> **GOTCHA 1 — `pg_dump -f <file>` FAILS here, at 03:20, silently, nightly.**
> The dump directory is `700 root:root` (a dump contains every note), but
> `pg_dump` runs as the `postgres` user via `sudo`, so *it* cannot open the
> output file: `could not open output file … Permission denied`. Write to
> **stdout** and let this script's own root shell do the redirect:
> `sudo -u postgres pg_dump -Fc -d noted > "$OUT.partial"`.
> Caught only because the service was **run immediately** rather than left to the
> timer. Enabling a timer is not evidence it works.
>
> **GOTCHA 2 — the same permission wall inverts on restore, and it fakes a
> CORRUPT BACKUP.** `sudo -u postgres pg_restore -l /var/backups/noted/*.dump`
> fails — not because the dump is bad, but because `postgres` cannot read a
> root-only file. It reports as though the archive were unreadable. Verify as
> **root** (`pg_restore -l "$D"`), and restore by redirecting on **stdin**:
> ```bash
> sudo -u postgres pg_restore -d <target> < /var/backups/noted/<file>.dump
> ```
> `pg_restore -U postgres -h /var/run/postgresql` as root does NOT work either —
> peer authentication maps to root, not postgres.

**Proven restore, 2026-08-10** — do this again after any change to the script:

```bash
psql "$URL" -c "CREATE TABLE probe(id int primary key, v text);
                INSERT INTO probe VALUES (1,'survives a restore');"
/usr/local/sbin/noted-pg-backup.sh
sudo -u postgres createdb noted_restore_probe
sudo -u postgres pg_restore -d noted_restore_probe < $(ls -t /var/backups/noted/*.dump | head -1)
sudo -u postgres psql -d noted_restore_probe -tAc "SELECT v FROM probe WHERE id=1;"
# -> survives a restore
sudo -u postgres dropdb noted_restore_probe
```

> **THE BACKUPS ARE ON THE SAME DISK AS THE DATABASE — this is NOT disaster
> recovery.** It covers a bad migration or a `DELETE` without a `WHERE`. If the
> box is lost, the dumps go with it. An off-box copy is **required before real
> user notes land here**; it needs a B2 application key scoped write-only to one
> backup prefix (it dials OUT, so it fits this box's no-inbound posture), and
> issuing one is the owner's call. Do not let the first real user sign up before
> this is resolved.

### The restore drill

```bash
docs/agent_docs/docs024_key_docs_latest/noted_rebuild/box/restore-drill.sh [identity-path]
```

Quarterly, and after any change to the backup scripts. Seven steps: identity
present and mode 600 → newest object (listed with `--versions` so a hide is
visible) → download with an **admin** key → assert the `age-encryption.org`
header → decrypt → assert a valid pg archive → restore and count rows.

> **It deliberately does NOT fall back to the box's postgres.** The scenario
> being rehearsed is *"the box is gone"*, so borrowing its database would
> rehearse the one case that cannot happen. With no local postgres client it
> spins up a throwaway `postgres:16` container instead; with neither, it FAILS
> loudly rather than passing on weaker evidence.

**Last run 2026-08-10 — DRILL PASSED**, entirely off the box: 4191 bytes down,
age header confirmed, decrypted to 3991, `PostgreSQL custom database dump
- v1.15-0`, restored in a container, 1 table / 1 row read back.

> **NEGATIVE CONTROL, run the same day:** the drill with a freshly generated
> decoy identity **fails at step 5** — `age: error: no identity matched any of
> the recipients`. Step 5 is therefore a real check and not a formality. Re-run
> this whenever the drill is changed: a decrypt step that has never refused a
> wrong key has not been shown to be able to.

> **⚠ The drill has NOT yet been run from the owner's stored copy of the
> identity.** It currently reads `~/noted-backup-age-identity.txt` on the
> operator workstation, which is the *only* copy in existence. The thing a drill
> is for is the recovery path as it will really exist, and the step that fails in
> a real recovery is always the key nobody could find. Re-run it against the
> password-manager copy once that exists, then delete the workstation file.
