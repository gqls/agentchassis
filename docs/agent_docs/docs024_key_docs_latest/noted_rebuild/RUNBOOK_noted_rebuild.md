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
