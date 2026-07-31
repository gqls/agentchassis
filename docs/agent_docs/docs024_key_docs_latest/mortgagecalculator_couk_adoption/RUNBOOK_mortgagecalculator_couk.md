# RUNBOOK — mortgagecalculator.co.uk adoption

Commands that were hard to get right, with the gotcha attached. Add to this file,
not to your scrollback.

## §1 B2 access

**The handoff says there are no B2 credentials. That is now WRONG** — as of
2026-07-31 ~21:56 local, `~/.config/b2/account_info` exists and authorises.

```bash
b2 account get          # confirms auth without mutating anything
b2 version              # 4.7.0 (b2sdk 2.12.0)  -- v4 CLI, see the flag gotcha below
```

**GOTCHA — the dry-run flag is `--dry-run`, not `--dryRun`.** The camelCase spelling
is b2 CLI v3. On v4 it exits **2** with a usage dump. That matters because a piped
`grep` over the failed output prints "(none)" for deletions and uploads, which reads
exactly like a clean no-op. **Print `${PIPESTATUS[0]}` and check it is 0.**

```bash
b2 ls --recursive b2://portfolio-sites/mortgagecalculator.co.uk/
```

## §2 Getting the true origin bytes — do NOT use curl

**GOTCHA — `curl https://mortgagecalculator.co.uk/robots.txt` is NOT the origin
file.** It returns **2,327 bytes**; the origin is **491**. Cloudflare injects a
`# BEGIN Cloudflare Managed content` block at the edge. The `x-amz-*` response
headers are still present, so the usual "it came from B2" tell does **not** catch
this. Committing the fetched copy would bake the injection into the origin and have
Cloudflare inject it a second time.

The handoff missed this because it ran `tail -5`, which lands past the injected
block. **When the bytes matter, take them from the bucket:**

```bash
b2 sync b2://portfolio-sites/mortgagecalculator.co.uk/ ./bucket   # no --delete = download only
```

Everything else survived the CDN untouched: 28 of 28 non-robots files were
sha256-identical live vs local, which is itself the evidence that Cloudflare
rewrites `robots.txt` specifically and not HTML generally.

## §3 Populating the deploy repo

```bash
# stage from the BUCKET download, dropping B2's directory placeholders
cd ./bucket && find . -type f ! -name '.bzEmpty' -exec cp --parents {} ~/projects/sites/mortgagecalculator.co.uk/ \;
```

`.bzEmpty` are zero-byte B2 folder markers. **Do not carry them**: the sibling
`loanandmortgagecalculator.co.uk` has **0** in its bucket against **52** repo files,
so 1:1 repo↔bucket is the established shape and the markers get swept on first sync.

`README.md` exists locally and is **404 live** — not in the bucket. Do not add it;
adding it publishes it.

**Commit with an explicit pathspec** — that repo carries other sessions' untracked
files (`.idea/`, `asset_key`, `kind=icon`, `scope_ref`, a stray `gamedesign.uk/…`):

```bash
cd ~/projects/sites
git add mortgagecalculator.co.uk
git commit mortgagecalculator.co.uk -m "..."
```

## §4 Proving the sync is safe BEFORE pushing

Run the workflow's own command as a dry run:

```bash
cd ~/projects/sites
b2 sync --delete --skip-newer --dry-run --no-progress \
  mortgagecalculator.co.uk b2://portfolio-sites/mortgagecalculator.co.uk
echo "EXIT ${PIPESTATUS[0]}"     # MUST be 0 -- see the --dryRun gotcha in §1
```

**Expected, and it is NOT a no-op:** 29 uploads + 35 deletes. Read the deletes
before panicking:

- **30 are `(old version)`** — B2 version pruning, each paired with a re-upload of
  the *same bytes*. `index.html (old version)` appears twice because the bucket
  holds two superseded versions.
- **5 are `.bzEmpty`** — the placeholders, as above.
- **0 are live content files removed without replacement.** That is the property
  that matters; check it with:

```bash
comm -23 <(grep -v '\.bzEmpty$' bucket_files.txt | sort) \
         <(cd ~/projects/sites/mortgagecalculator.co.uk && find . -type f | sed 's|^\./||' | sort)
# empty == nothing in the bucket is missing from the repo == sync deletes no content
```

**Why it uploads at all when the bytes are identical:** `b2 sync` compares mtime,
and freshly-staged files are newer than the Jan-2026 bucket objects. `--skip-newer`
skips only when the **destination** is newer. The GitHub runner does a fresh
`actions/checkout`, so its mtimes are newer too — same behaviour, and the same
behaviour every other domain already gets.

## §5 The outage chain this all exists to prevent

`deploy-to-b2.yml` derives changed domains from the diff, then:

```bash
for domain in $CHANGED; do
  if [ -d "$domain" ]; then
    b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"
  fi
done
```

A verbatim `page_rerender` **commits the rendered page into `gqls/sites`**
(`deploy_result.data.repo_url = https://github.com/gqls/sites`). If the domain is
absent from the repo, that commit *creates* the directory holding **one file**,
`[ -d "$domain" ]` becomes true, and `--delete` makes the bucket match one file.
**The run goes green.** Hence: populate the repo before anything can trigger a
rerender.

## §6 Adoption pre-flight (not yet run — from handoff §5a)

```sql
-- another lane already on this domain?
SELECT item_type, status FROM site_work_items
 WHERE status NOT IN ('complete','cancelled','rejected')
   AND (spec::text ILIKE '%mortgagecalculator%');

-- exactly ONE ported-page component must exist
SELECT id, name FROM content_components WHERE function='ported-page';

-- fidelity must survive the spawn boundary: input_mapping is an ALLOW-LIST
SELECT default_config->'workflow'->'steps'->'call_adopter'->'config'->'input_mapping'
  FROM agent_definitions WHERE type='site-adoption-orchestrator' AND is_active;
```

**Under `--fidelity high` the assertion in handoff §5d INVERTS.** It says
`needs_content_page` + `needs_tool_recreation` must be **zero**. That is the
`locked` assertion. On the recreate path those items are the *expected* output —
their absence would mean the run did nothing. Do not copy that check across.
