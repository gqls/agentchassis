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

Migration `323` applied 2026-08-06 (roll-independent: write-only, and the old readers
ignore `storage_path`). 205 rows backfilled, verify block silent, re-measured after:
presigned-with-no-storage_path is now **0**, and all five at-risk logo rows resolve
via `storage_path`. The 49 stranded rows are deliberately untouched — their source is
genuinely unrecorded, and inventing one would be worse than failing loud.

## 2026-08-06 (later) — the council's two mediums, and a passenger I took

**APPROVED round 1**, `c055840a-9edc-4f9a-8a4a-b23ac4cad02a`, 8 advisory
objections, none high. Both mediums that were checkable turned out to be right,
which is worth recording because my instinct on reading them was that they were
generic reviewer caution:

- *"the census cannot see dynamic/queue-built callers"* — correct. Widening it to
  `site_work_items.spec` and `scripts/` found `081c_direct_asset_deployer.sh`,
  an operator crib printing `content_data->>'{hero,hero_about,logo}_uri` as the
  URI to paste into a hand deploy. **My three-way census was three views of the
  same surface** (Go source, agent configs, two repo dirs) and every one of them
  was blind to a shell script in a fourth dir. Three checks that share a blind
  spot agree with each other — the memory line for it is
  `two-blind-checks-agree-with-each-other`, and I had cited that very lesson when
  designing the census.
- *"a migration number is not yours until the ledger says so"* — correct, and
  faster than I would have believed: I listed the directory, took `321` as free,
  and a concurrent session committed its own `321` (`b14609e05`) inside the same
  hour. `322` had gone too. Now `323`.

**A passenger, disclosed.** My commit `bef973cb3` names `LANDMINES.md` on its
pathspec for a one-line renumber, and took **another session's two new landmine
entries** with it (19 added, 1 mine) — the same-file case CLAUDE.md says no hook
can prevent. Nothing is lost and forward-only forbids an amend, so it is recorded
here instead. The lesson is narrower than "be careful": **a pathspec protects you
per FILE, and a shared append-only file is exactly where that protection is
worth least.** Reading `git diff --numstat` before committing would have told me
(it did — 19/2 for a 1/1 edit) if I had read it as a *gate* rather than as the
append-only check I ran it for.

## 2026-08-06 (later still) — the closure test refuted my scope claim, which is the best thing it could have done

Ran 155's own closure test rather than handing it off. Three `asset_id`-only deploys
at dartsonline icons, all published with `PUBLISH_OK` confirmed (the first attempt
printed nothing because `--quiet` is not a `kubectl run` flag — the
`kcat-publish-silently-drops` discipline caught a *different* silent-publish failure
than the one it was written for, which is the argument for the discipline).

All three COMPLETED and all three **skipped**: `"no storage URI found for icon"`.
`asset_id` was in the child's `input_data`; the action never saw it. Cause:
`asset-deployer`'s `deploy_asset` step declares `input_fields` without `asset_id`,
and `ExtractActionInputs` Strategy 1 extracts only listed names
(`action_inputs.go:441`). That config dates from **2026-02-20** — so it was already
true when 155 was filed, and **the branch I deleted could not have produced 155's
reported symptom.**

The surviving candidate is `findStorageURI` Priority 2 — top-level `{purpose}_uri`
in `collected_data`, written in-run by the same two writers, read *before* the
asset_id path. I had read that function. I cite it in this very PLAN as the reason
to keep the `collected_data` key. And I still wrote "the wrong-bytes state becomes
unrepresentable" into the commit, the council submission and the register.

**The transferable bit, which is not "be careful":** I was reasoning about the
branch I was editing, and writing a claim about the outcome. Those are different
scopes and nothing in my process forced me to reconcile them. The check that would
have: *one query over live `agent_definitions` asking which steps invoke this action
and what they declare* — reachability, before claiming a fix. Two minutes at plan
time; it is now R8 in the RUNBOOK.

Recorded as a correction in `bugs_open/155`, in `WRONG_CALLS.md`, and as a scope
correction on the register entry (which council seats read as ground truth). The
bug stays OPEN with a revised, three-step closure list. The fix itself is unchanged
and still good — it is one arm of two, and now says so.
