# RUNBOOK — bugfix 128, `image_url_404`

Every command that was hard to get right, with its gotcha attached. Change it HERE
when it changes.

## Ownership: `who-owns.py` is wrong on this bug, in a knowable way

```bash
./scripts/who-owns.py 128        # says OWNED by brochure_component_library
```

**It reads COMMITS, so a FILING workstream and an OWNING one look identical.** The
lane it names lists 128 as "read, still unowned" in its own handoff. The check that
actually works is a grep of live session transcripts for the target's **code symbols**
— never its number, which appears in every `ls bugs_open/`:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis
find . -name '*.jsonl' -mmin -720 -size +5k \
  -exec grep -lE "check_image_url_404|loadKnownAssetPurposes|collectImagePathReferences" {} \;
# then READ the context of the hits — most are directory listings, not work:
grep -oE ".{120}image_url_404.{120}" <session>.jsonl | tail -5
```

## The population the check actually scans

Mirror the check's own query, or you are measuring a different set:

```sql
SELECT s.domain, m[1] AS path
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
CROSS JOIN LATERAL regexp_matches(pc.rendered_html,
  '(/assets/images/[a-zA-Z0-9_\-]+\.(?:jpg|jpeg|png|webp|svg|gif))', 'g') AS m
WHERE pc.build_status = 'deployed' AND pc.locked_at IS NULL
GROUP BY 1,2 ORDER BY 1,2;
```

Gotchas: `build_status='deployed' AND locked_at IS NULL` are both part of the check's
predicate — drop either and your numbers will not match what the check can see. The
chrome surface is a **second** query against `site_components` (no `build_status`
column worth filtering; it is the stored artefact the whole site renders):

```sql
SELECT s.domain, m[1]
FROM site_components sc JOIN sites s ON s.id = sc.site_id
CROSS JOIN LATERAL regexp_matches(COALESCE(sc.rendered_html,''),
  '(/assets/images/[a-zA-Z0-9_\-]+\.(?:jpg|jpeg|png|webp|svg|gif))', 'g') AS m
GROUP BY 1,2;
```

## Probing HTTP ground truth — and the trap in the result

```bash
curl -s -o /dev/null -w "%{http_code}" --max-time 25 "https://$domain$path"
```

**`000` is NOT a status.** It is a curl connection error, and on this fleet it happens
often enough that a single pass will hand you several. **Re-probe every non-200 before
tallying**, with a cache-buster, or you will report false negatives in a document about
false negatives:

```bash
curl -s -o /dev/null -w "%{http_code}:%{size_download}" --max-time 25 "https://$d$p?cb=$RANDOM"
```

Five of 127 needed a re-probe on 2026-07-31; **all five were 200.**

## Deriving the served path — where the answer actually lives

Not in the `assets` table. Measured 2026-07-31 over 267 active rows: `url` is a
presigned S3 link on 152, `filename` empty on 191, `storage_path` empty on 189, and of
the 115 whose `url` looks local, **47 are the unresolved literal**
`/assets/images/input-data.asset-key.jpg`.

```sql
SELECT count(*), count(*) FILTER (WHERE url LIKE 'https://s3.%'),
       count(*) FILTER (WHERE COALESCE(filename,'')=''),
       count(*) FILTER (WHERE url LIKE '%asset-key%')
FROM assets WHERE status='active';
```

The served path is **derived**: `storage.DeployedWebPath(asset_key, purpose)` —
**except** for `favicon`/`og_card`, which must go through
`storage.BrandHeadAssetPaths` (`og_card` serves as `og-card.png`; the helper returns
`og_card.png`). `storage.IsBrandHeadPurpose(p)` is the branch. Getting this wrong
reports a 404 for every site's og card and favicon.

## Running the REAL Go helpers over live data (not a re-implementation)

A Python mirror of a Go function agrees with itself, which proves nothing. Extract the
rows with `psql -At -F'|'`, then drive the actual helpers:

```bash
T=$HOME/.cache/bugfix128_headtree            # NOT /tmp — see below
rm -rf $T && mkdir -p $T && git archive HEAD | tar -x -C $T
cp platform/orchestration/actions/discovery_checks/check_image_url_404*.go \
   $T/platform/orchestration/actions/discovery_checks/
mkdir -p $T/cmd/xcheck128 && cat > $T/cmd/xcheck128/main.go <<'EOF'
# ... reads assets.txt + rendered_paths.txt, prints paths no asset deploys to
EOF
cd $T && TMPDIR=$HOME/.cache/gobuildtmp go run ./cmd/xcheck128 assets.txt rendered_paths.txt
```

## Building and testing when the shared tree does not compile

The working tree carries other sessions' in-flight edits — on 2026-07-31
`check_empty_sections.go` referenced `datahelpers` with no import (another lane's
`bugs_open/137` work) and the whole package failed to build. That is not your change:

```bash
cd $HOME/.cache/bugfix128_headtree     # git archive HEAD + your files only
TMPDIR=$HOME/.cache/gobuildtmp go test ./platform/orchestration/actions/discovery_checks/ -run TestImageURL404 -v
TMPDIR=$HOME/.cache/gobuildtmp go test ./platform/orchestration/actions/discovery_checks/
```

**Do NOT put the checkout in `/tmp`.** It is a 16G tmpfs shared by ~30 concurrent
sessions and it hit 100% during this work. The failure is misleading: `go build` printed
`BUILD_OK` and *then* died with "no space left on device", because the error is in
output capture, not the command. `df -h /tmp` first; `~/.cache` otherwise.

## sqlmock expectations must match the query ORDER

The check issues three queries in a fixed order and `sqlmock` is ordered by default:

```
1. FROM page_components     2. FROM site_components     3. FROM assets
```

Query 3 is skipped entirely when the first two find no reference — arm it anyway (an
unfulfilled expectation is only checked if you call `ExpectationsWereMet`).

## After the roll — the acceptance set

```sql
SELECT s.domain, w.item_key, w.severity, w.spec->>'kind', left(w.summary,80)
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.item_type='image_url_404' AND w.created_at > '2026-07-31'
 ORDER BY 1,2;
```

Expect ~18 page findings + 2 chrome findings. Must be present:
`vonc.com/…/hero.jpg` (was masked), `idea.uk/…/og-card.png` (chrome, severity `high`),
`finetuning.uk/…/case-study-legal-rag.jpg` (regression guard),
`ai-agent-orchestration.com` `image_url_404:empty-src`. Must be **absent**:
`fundamentallyai.com/…/brand-illustration.jpg` (200) and the three
`robot-hands` `content-hero-tool-*` rows (all 200).

## Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
# SUBMISSION_CORR printed; find the run by PAYLOAD, never by the printed id:
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='<CORR>' AND kind='council_report' ORDER BY created_at;
```
