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

## §6 Adoption pre-flight (run 2026-07-31 — clean)

> **CORRECTED 2026-07-31, same session.** This section first carried the handoff's
> `spec::text ILIKE '%mortgagecalculator%'` phrasing. **Do not use it.** It matched
> **41 `page_rerender` rows that belong to `loanandmortgagecalculator.co.uk`** —
> the sibling domain *contains* ours as a substring — and 41 is a plausible page
> count for our site, so it reads as "another lane is mid-adoption here, stop".
> Every affected query below is now an exact match or a join. See NOTES and the
> `LANDMINES.md` entry on nesting domain names.

```sql
-- another lane already on this domain? JOIN and match EXACTLY; group by domain so
-- a foreign site announces itself instead of hiding inside a total.
SELECT s.domain, w.item_type, w.status, count(*)
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.status NOT IN ('complete','cancelled','rejected')
   AND s.domain = 'mortgagecalculator.co.uk'
 GROUP BY 1,2,3;

-- exactly ONE ported-page component must exist (matters on the locked path;
-- checked anyway — a second one is a fleet-wide problem either way)
SELECT id, name FROM content_components WHERE function='ported-page';

-- find YOUR run by payload, exact match, never by the printed id
SELECT orchestration_id, owner_agent_type, current_step, status,
       collected_data->'input_data'->>'fidelity' AS fidelity
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'destination_domain' = 'mortgagecalculator.co.uk'
 ORDER BY created_at DESC;
```

**Result on 2026-07-31:** our domain had **0 `sites` rows, 0 runs, 0 work items**.
The `fidelity` plumbing check (migration 274 / the `call_adopter` `input_mapping`
allow-list) is **moot on this path** — a dropped `fidelity` is indistinguishable
from `high`, because `082` itself defaults adopt-mode to `high`:
`FIDELITY="${FIDELITY:-high}"` (line 124). It is worth checking only for `locked`.

**Do not dispatch within ~300s of a chassis pod restart** — the spawn is silently
dropped. Check for a roll in progress *and* pod age before submitting:

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,READY:.status.containerStatuses[0].ready,START:.status.startTime'
```

Two replicaset name prefixes in that output means a roll is in flight; wait for it.

**Under `--fidelity high` the assertion in handoff §5d INVERTS.** It says
`needs_content_page` + `needs_tool_recreation` must be **zero**. That is the
`locked` assertion. On the recreate path those items are the *expected* output —
their absence would mean the run did nothing. Do not copy that check across.

## §7 Positioning specs — two refinements to what the handoff says

Both verified in code today, before relying on them.

**`audience` is not consumed, but it is not un-referenced.** Handoff §6 says the
aspect "is read by NOTHING". Practically true — no agent, prompt or pipeline path
reads its data. But `grep` finds exactly one hit and it is not a contradiction:

```go
// internal/core-manager/admin/spec_admin_handlers.go:226
if aspect == "identity" || aspect == "tone" || aspect == "audience" {
    scopeQuery += " AND COALESCE(p.page_type, '') NOT IN ('blog-post')"
}
```

That uses the aspect **name** to decide which pages an admin operation covers. It
never reads the spec body. Expect this hit; it does not mean the aspect is live.

**`content_direction.formatted` regenerates ITSELF — but only on the action path.**
Handoff §6.2 warns that a hand-written spec which does not regenerate `formatted`
is invisible. True, and worth being precise about *when*:

```go
// platform/orchestration/actions/site_spec_actions.go:211-217
// This runs for every content_direction write — classifier, adoption, HITL.
if aspect == "content_direction" {
    formatted := datahelpers.FormatContentDirection(specMap)
    if formatted != "" { specMap["formatted"] = formatted }
}
```

It runs **before the transaction**, unconditionally, for any write that goes through
`site_spec_actions`. So:

- writing the spec **through the platform's own action** → `formatted` is correct by
  construction, and the handoff's trap cannot fire;
- writing it with **raw SQL** (which is what the sibling's `set_divergence_specs.py`
  does) → `formatted` is whatever you put there, and a spec that looks applied
  changes nothing.

**Prefer the action path.** If raw SQL is unavoidable, reproduce
`datahelpers.FormatContentDirection` exactly and gate it as handoff §6.3 describes —
compare as a **multiset of lines**, not as a string, because Go map iteration order
is random so the stored section order is arbitrary.

Schema reminders that cost time otherwise: the JSONB column is **`data`**, not
`spec_data`; `idx_site_specs_current` is `UNIQUE (site_id, aspect) WHERE is_current`,
so the supersede must precede the insert.
