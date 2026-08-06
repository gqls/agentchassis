# HANDOFF 2026-08-06 — cold start for bugs 152 + 155. Read this first.

**State: the fix is LIVE and PROVEN AT THE BINARY. Both bugs are still OPEN, and
they are open for one reason only — the behavioural proof has not been run.**
Everything below is either done-and-evidenced or a precise instruction. You do not
need to re-derive anything.

## Do not redo these — they are done and evidenced

| thing | state | evidence |
|---|---|---|
| Code fix (7 files, one derivation) | committed `1d11827c1` | `go build`/`vet`/`test` green, incl. against a clean `git archive HEAD` tree |
| Council gate | **APPROVED round 1** | `c055840a-9edc-4f9a-8a4a-b23ac4cad02a`, 8 advisory, none high; both checkable mediums acted on in `bb53326a8` |
| Live on the fleet | **YES, chassis `v1.0.1259`, both replicas** | negative control `"Resolved s3_uri from site content_data via asset_id"` → **0**; positive `AssetSourceRef` → 2; nonsense control → 0 |
| Migration (backfill `storage_path`) | **APPLIED**, 205 rows | presigned-with-no-`storage_path` now **0** fleet-wide |
| Register IMG-068 | entry + index row, status corrected to live | `imagery.md`, `000_concept_index.md` |
| LANDMINES entry for 155 | **retired on its own written test** | dated correction in place, synced to `doc_notes` |
| Standing five | PLAN, NOTES, RUNBOOK, README_where_we_are, SUMMARY | this directory |
| Wrong calls | 2 logged | `WRONG_CALLS.md` (schema-from-summary; a test that couldn't observe its own property) |

**My commits, in order:** `1d11827c1` (fix) · `d20d81a19` (docs+landmine+wrongcall) ·
`1ec6361c5` (test strengthened after a mutation caught it) · `bb53326a8` (council
responses, migration renumbered 321→323) · `bef973cb3` (verdict recorded) ·
`20d5f29c4` (passenger disclosure) · `ef13c1425` (live proof + SUMMARY).

## The ONLY thing standing between here and closing both bugs

### Task A — 155's closure test (deploy 2+ same-purpose assets by `asset_id` alone)

The bug's own recipe, and nothing weaker counts: **`success:true` and distinct
destination paths were BOTH already true while it was shipping identical bytes.**

Target: **dartsonline.com**, 20 active `icon` assets (site id from
`SELECT id FROM sites WHERE domain='dartsonline.com'`). Pick 2–3 of them.

```sql
SELECT a.id, a.asset_key, left(COALESCE(NULLIF(a.storage_path,''),a.url),70) AS source
FROM assets a JOIN sites s ON s.id=a.site_id
WHERE s.domain='dartsonline.com' AND a.purpose='icon' AND a.status='active'
ORDER BY a.asset_key LIMIT 3;
```

Dispatch each via the spawn+call wrapper in
`scripts/initial_messages/180_adoption/081c_direct_asset_deployer.sh` — but change
the `input_mapping` to pass **`asset_id`** and NOT `s3_uri`. Passing `s3_uri`
bypasses the entire code path under test and would prove nothing; that is the July
workaround, not the test.

```
"input_mapping": {"domain":"input_data.domain","asset_id":"input_data.asset_id",
                  "purpose":"input_data.purpose","asset_key?":"input_data.asset_key"}
"input_data": {"domain":"dartsonline.com","asset_id":"<uuid>","purpose":"icon",
               "asset_key":"<key>"}
```

⚠ **Publish via the container COMMAND, not piped stdin** — `kubectl run -i --rm |
kcat -P` silently drops ~4 in 5 messages at exit 0 (the `kcat-publish-silently-drops`
landmine). Use `--command -- sh -c "... && echo PUBLISH_OK"` and confirm `PUBLISH_OK`.

**PASS** = the deployed files have **different** sha256s AND at least one, opened and
looked at as an image, matches its own `origin_prompt`. Download from the site's
served URL (`storage.DeployedWebPath` form: `/assets/images/<key-hyphenated>.jpg`).
**FAIL** = identical sha256s — which would mean the fix is wrong, not that the test is.

Expected-and-fine outcome: the bytes match what is already deployed there, because
the July workaround already put the correct icons in place by hand. **The test is
about whether the six deploys differ FROM EACH OTHER**, not whether the site changes.

### Task B — 152's closure test (a real derivation off a recovered logo row)

Any one of the five sites whose logo row was previously unresolvable and now resolves
via `storage_path`: webdesign.co.uk, gaswholesalers.com, finetuning.uk,
vetcomparison.uk, leopardessconsulting.co.uk. Run `derive_brand_head_assets`
(`mode=brand_head` on `asset-deployer`) and confirm favicon/og-card are produced and
serve 200. Before the fix this errored at *"could not derive storage key from logo
url"*; that error is now impossible for these five.

⚠ Check `locked_at` first — the deriver refuses locked brand-head artefacts by
design, and a refusal is a `derived:false` **result**, not an error. That is correct
behaviour, not a failure of this fix; pick a site whose favicon/og_card are unlocked
or you will read a refusal as a negative result.

### Task C — close both files

Only after A and B. `git mv` to `bugs_closed/` and **name BOTH the old and new paths
on the `git commit` pathspec** — a pathspec commit after a bare `git mv` ships a
COPY, leaving the file in both dirs at HEAD. Verify at HEAD, not at the tree:
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep -E '15[25]'`
should return exactly one line each. Then update `016b` §10's index rows, and
consider a §9 entry for the transferable pattern (*an identity reconstructed from a
shared cache instead of read from the row it belongs to* — this is its second
instance after `bugs_closed/168`'s destination-side twin, which makes it a class).

## Three things that will mislead you if nobody tells you

1. **`bugs_open/152`'s own opening section describes a symptom that can no longer
   occur.** The readers PARSE `assets.url`, they do not fetch it, so signature expiry
   is irrelevant. The live half is different and is written up in the 2026-08-06
   progress section. I nearly fixed the dead half.
2. **The 49 stranded rows are NOT a blocker and must not be "repaired".** Their
   source is genuinely unrecorded (hand-repairs and derived-card upserts). Inventing
   a plausible `storage_path` for them would convert an honest loud failure into a
   silent wrong-file fetch — the exact defect this lane just removed.
3. **`resolveReferenceAssetURIs` now WARNS where it used to skip silently.** If you
   see `"reference asset row cannot name its source object — anchor skipped"` in the
   logs after this rolled, that is the fix reporting a pre-existing condition, not a
   new fault. It was always happening; it was just invisible.

## If you are picking a different bug instead

Both bugs are safe to leave in this state — the defect is gone from production; only
the paperwork proof is outstanding. `who-owns.py` will say OWNED on 152/155 because
of these commits, which is correct: contribute into the bug files, do not re-fix.
