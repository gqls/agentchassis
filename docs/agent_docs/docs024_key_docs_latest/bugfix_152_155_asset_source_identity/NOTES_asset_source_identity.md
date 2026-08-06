# NOTES — asset source identity (bugs 152 + 155). Append-only, newest at the bottom.

## 2026-08-06 — bug selection, and why these two together

Swept `bugs_open/` against `who-owns.py` and against the live `.jsonl` transcripts
of all 12 active sessions (the ownership script reads COMMITS, so a session mid-fix
is invisible to it — memory `who-owns-is-blind-to-uncommitted-sessions`). Owned or
active: 033, 071, 083, 084, 085, 093, 096, 107, 113, 114, 122, 151, 152?, 178, 181,
182, 184, 196, 197, 201, 204. `152` showed OWNED only because `who-owns` matched the
*number* inside other lanes' docs; no session's transcript mentions
`deploy_image_asset`, `resolveStorageURIFromAsset` or `assets.url` beyond passing
citations, no open work item touches them, and no commit has touched the four
implicated files since `fd0516b18` (2026-08-04, the 179 lane, which CLOSED and
explicitly handed this seam over). Took 152+155 as one lane on 179's own advice.

## The bug I filed for turned out to be half-dead — and the live half is different

`152` as filed says the two derivation call sites *fetch* `assets.url` and get a 401
on an expired signature. **That is no longer true and I nearly wrote a fix for it.**
Read the code: both sites PARSE the url to an object key and download with the
client's own credentials (`presignedURLToS3URI` → `ExtractKeyFromS3URI` →
`s3Client.Download`). An expired signature is irrelevant to a parse. Had I trusted
the bug file's own summary I would have "fixed" a symptom that cannot occur.

What IS live is the mirror image: `deploy_image_asset` now FLIPS `url` to the
deployed local path and preserves the source into `storage_path` (IMG-053 "Edit F",
which landed after 152 was filed) — and **nothing reads `storage_path` back**. So
the readers are stranded by the platform's own repair. Census 2026-08-06 before any
change:

```
presigned url, no storage_path   205   (resolvable by parse — the case that still works)
local web path, storage_path      107   (recoverable, and NO reader looks)
local web path, no storage_path    49   (stranded: hand-repairs + derived-card upserts)
```

Five sites' **active logo rows** carry a non-presigned url (four of them the
unrendered template literal `/assets/images/input-data.asset-key.jpg`), so the next
`derive_brand_head_assets` on each fails at "could not derive storage key from logo
url" — a live, unfired landmine, not a hypothetical.

A **fourth** reader nobody had listed: `resolveReferenceAssetURIs`
(`imagery_style_guide.go`) reads `url` only and skips SILENTLY, so imagery style
anchors evaporate the moment the referenced asset is deployed. No log line, no
error, weaker generations.

## Measuring the cache's readers — three ways, each disconfirmable

Before deleting `sites.content_data->>'{purpose}_uri'` I had to know nothing else
read it. One grep would have been a claim about one spelling
(`a-grep-proves-absence-only-for-its-spelling`):

1. Go grep for key construction (`+ "_uri"`, `"_uri"`) across `platform/`,
   `internal/`, `cmd/` — writers plus the one buggy reader.
2. Regex over LIVE active `agent_definitions` configs:
   `default_config::text ~ '(hero|logo|icon|content_hero)_uri'` → **0 rows**.
   Positive control in the same run: `LIKE '%_uri%'` → 7 agents, every hit
   `s3_uri`/`image_uri`/`dataset_uri`. Without that control the 0 would have been
   worthless (`two-blind-checks-agree-with-each-other`).
3. Grep over `sql_for_agents/` and `internal/core-manager/` → 0.

Then the trap I nearly fell into in the other direction: the in-run
`collected_data[{purpose}_uri]` copy DOES have a reader (`findStorageURI`
Priority-2, legacy pageflow). I kept it. It cannot leak the DB value across runs
because `ExtractNestedFieldString` is a strict dot-path with a `.response` unwrap
(`data_helpers.go:1199-1234`), NOT the depth-20 recursive hunt that
`ExtractActionInputs` uses — the distinction 179's own refusal guard turns on.

## Misstep: I asserted the council submission schema instead of reading it

Built the submission JSON with `plan` as an ARRAY of edits, from CLAUDE.md's
abbreviated description ("a `plan` (≤8 edits, each with file/operation/rationale/
sketch)"). The trigger refused: `ERROR: .plan missing`. The real schema is an
OBJECT — `{summary, edits[], grounded_in[], risks}` — and it is written out in full
in the script's own header, which I had not opened. Cost one refused run (client-side,
no credits). Logged in `WRONG_CALLS.md`. The generalisable half: `risks` is a
first-class field I would have skipped entirely, and it is where a reviewer is told
what to check — writing it forced me to state the bare-key/ReferenceFetcher split and
the COALESCE precedence question, both of which are genuinely arguable.

## What shipped, and what the fresh build does NOT contain

Commit `1d11827c1`, council `c055840a-9edc-4f9a-8a4a-b23ac4cad02a` (submitted, verdict
pending — `Council-Submitted:` trailer, so 098 credits it automatically on approval).

**The chassis build deployed this morning is `v1.0.1257`, started 09:52Z — BEFORE
this commit — so it does not carry the fix.** Proven at the binary rather than
assumed, on both replicas, with both controls:

```
AssetSourceRef                                       -> 0  (new symbol: not shipped)
"Resolved s3_uri from site content_data via asset_id" -> 1  (positive control: the
                                                             OLD buggy branch is live)
```

That positive control is the interesting one: it is the log line of the very branch
this change deletes, so the same command that proves my code is absent proves the
defect is present. After the next roll the pair inverts — that is the closure test.

Migration `321` applied 2026-08-06 (roll-independent: write-only, and the old readers
ignore `storage_path`). 205 rows backfilled, verify block silent, re-measured after:
presigned-with-no-storage_path is now **0**, and all five at-risk logo rows resolve
via `storage_path`. The 49 stranded rows are deliberately untouched — their source is
genuinely unrecorded, and inventing one would be worse than failing loud.
