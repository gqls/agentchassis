# 304 — Retracting the LAST page of a site cannot unpublish it: the git delete empties the directory, the deploy skips it, and both halves report success

**Filed:** 2026-08-18, `loanzy_uk_example_site` lane · **Status:** OPEN, reproduced first-hand
today · **Live instance at filing:** `https://loanzy.uk/about.html` serving **200** with
`cf-cache-status: DYNAMIC` (so the bucket really holds it, it is not an edge cache), ~70
minutes after a retraction that every mechanism called successful.

> **On the 090 diagnosis loop (owner ruling 2026-07-31):** not run, and this states why, as
> the ruling requires. The cause is not inferred from a symptom — I read the deploy
> workflow's source, read the run log where it announces the skip in its own words, listed
> the bucket, and probed the edge. Every link in the chain was observed rather than
> reconstructed, and the mechanism was already independently documented by another lane
> (LANDMINES, 2026-08-08) from a different starting case. What is NEW here is the
> interaction with `retract_page_deployment`, which is first-hand.

## What happens

1. A page is retracted (`page-retraction` → `retract_page_deployment`). It deletes the file
   from `gqls/sites`. Real commit, correct behaviour: `Retract 1 retired page(s) from
   loanzy.uk (bugs_open/098)`, `14:06:49Z`.
2. That file was the **only** file under `loanzy.uk/`. Git does not track empty directories,
   so the directory ceases to exist in the tree.
3. `Deploy to B2` fires. `Get changed domains` correctly identifies `loanzy.uk` (the diff
   still lists the deleted path). `Sync to B2` then guards each domain on `[ -d "$domain" ]`,
   which is now **false**, so it takes the `else` branch:

   ```
   Changed domains: loanzy.uk
   WARNING: loanzy.uk in changed set but no directory — skipped
   ```

   `b2 sync --delete` — the only thing that removes a file from the bucket — never runs.
4. The Cloudflare purge step runs anyway, purging a cache in front of an object that was
   never removed. Run conclusion: **success**.
5. The page keeps serving. `pages.status='archived'`, `build_status='deployed'`, the repo is
   correct, the bucket is wrong, and nothing anywhere reports a fault.

## Why this is worth a bug and not just the existing landmine

The mechanism is already in `LANDMINES.md` ("Deleting a whole domain DIRECTORY from
`gqls/sites` deploys NOTHING", 2026-08-08, `bugfix_071_fragment_blindspot`, measured on run
`31266031734`). This bug is the **interaction**: it means the platform's own unpublish
primitive — built to close `bugs_closed/098`, *"archiving a page does not retract it from the
deployed site"* — is **structurally unable to unpublish a site's last page**. 098 is closed;
for this case its remedy does not reach the artefact. A single-page site, a site being
retired page-by-page, and any site whose only deployed page is being pulled all land here.

It also defeats the one case where unpublishing matters most. The reason this was found is
that the page in question was a compliance problem (a build that classified itself as a UK
credit broker — see `loanzy_uk_example_site/SUMMARY_2026-08-18_…`), i.e. exactly the
situation where "the retraction reported success" is the sentence you must not believe.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Never let a site directory become empty (best).** When `retract_page_deployment` removes
   what it can see is the last file under `<domain>/`, have it write a `.keep` in the same
   commit. The directory then survives, `[ -d ]` holds, and `b2 sync --delete` reconciles the
   removal through the normal path. No destructive verb is added anywhere, and the empty-
   directory state — the thing the deploy cannot express — stops existing.
2. **Teach the deploy to express prefix removal.** Replace the silent `else` with a genuine
   removal: when a domain is in the changed set and has no directory, `b2 rm --recursive
   "b2://portfolio-sites/$domain"`. This is the only way a whole *site* retirement can ever
   work, so (1) and (2) are complements, not alternatives. Guard it hard — it is a
   destructive verb keyed on an absence — and at minimum require that the push's diff for
   that domain contains deletions and no additions.
3. **Make the WARNING an error.** A green run that silently skipped a changed domain is a
   lie by omission; `exit 1` on that branch (once 1 or 2 exists) turns a discoverable-only-
   by-grep condition into a failed deploy.
4. **Verify at the artefact.** `retract_page_deployment` returns `success:true` for a git
   delete. It could re-probe the page URL after the deploy window and report the truth. This
   is detection, not prevention — list it last, and note it overlaps `bugs_open/236` (nothing
   on the platform asks whether a site SERVES).

## How to verify a fix

Retract the only page of a throwaway site and then, **at the URL**, confirm it 404s — never
at the action's return value, never at the workflow's conclusion, and never at the repo
(the repo is already correct today). Grep the run for the tell:

```bash
gh run view <id> --repo gqls/sites --log | grep -E "Changed domains:|Syncing .* to B2|no directory"
```
A domain in `Changed domains:` with no matching `Syncing` line is an orphaned prefix.

## Removing the orphan today (the manual remedy, and a correction)

`b2 rm --versions -r "b2://portfolio-sites/<domain>/"` — dry-run first, it prints exactly the
keys it would remove. **The landmine entry says these credentials "live only as GitHub
secrets"; that is out of date.** `[MEASURED 2026-08-18]` the b2 CLI on this box (v4.7.0) is
authorised against `portfolio-sites`: `b2 ls b2://portfolio-sites/loanzy.uk/` returned the
single orphaned key, and `b2 rm --dry-run --versions -r` listed exactly it. Note for agent
sessions: both the direct `b2 rm` and a `gh api` write to `gqls/sites` were **refused by the
session harness's auto-mode classifier** as destructive/outward-facing, so an agent may need
the owner to run the removal or grant the permission — budget for that rather than assuming
you can clean up after a retraction.

## Note from the 429 fix (2026-09-02, bugfix_429_mirror_unpublish lane)

The MIRROR side of this bug's shape is now explicit rather than accidental:
`b2worker.Publish` (the `publish_site` hosted-copy seam, DGH-008) REFUSES an
empty file set outright, naming this bug — so "last page retracted ⇒ empty
origin ⇒ the mirror silently keeps serving the whole site" is at least a
recorded refusal in `publish_site`'s result (`origin tree … is EMPTY but a
hosted copy is still standing`), not a silent skip. Whole-site unpublish remains
THIS bug's decision. When it is made, the mirror half has a ready hook:
`publish.Request.AllowBulkUnpublish` / the `allow_bulk_unpublish` input lifts
the sweep's bulk floor, so a deliberate teardown can be expressed as an explicit
hand dispatch rather than a new destructive verb.
