# 470 — `image_url_404` findings can never close: the check emits no retraction, so a repaired image keeps its finding for ever

**Filed** 2026-09-03 by the improvement_loop lane
(`docs/agent_docs/docs024_key_docs_latest/improvement_loop/NOTES_improvement_loop.md` §(xx)).
**Status:** OPEN, unowned. Local, self-evidencing, one file (plus one key in a sibling).

**Diagnosis-loop substitution, stated per the 2026-07-31 owner ruling:** this file asserts a
LOCAL cause, not a structural one — a single check has no retraction path — and it is
self-evidencing: `grep -c ResolvedFinding` on the file reads 0, the sibling control reads 5,
and the closer census reads 0 rows ever closed by the check. `090` was not run; nothing here
depends on a cause outside the file. If a fix turns out to need the shared seam (candidate 3
below), that half goes through `090` and the council first.

## 1. Symptom

`[MEASURED 2026-09-03 17:3xZ]` — the standing pile of `image_url_404` rows, live table:

| status | rows |
|---|---|
| `detected`, handler-less | **87** across **34** sites |
| `complete` (live + archive, ever) | 8 |
| `cancelled` (ever) | 3 |

Oldest open week 2026-07-20. Age profile of the 87: 07-20 5 · 07-27 7 · 08-03 22 · 08-10 4 ·
08-24 31 · 08-31 18. **None of the 8 `complete` rows was closed by the check** — it cannot
close anything (§2). Every one of the 87 is a claim of unknown truth: the image may have been
repointed, regenerated or removed weeks ago and the row would read exactly the same.

This is a different defect from "nobody drains the flag-only pile" (the improvement_loop
lane's plan item 4). That pile is findings nothing acts on. These are findings that stay
**after they stop being true**.

## 2. Cause — one file, read

`platform/orchestration/actions/discovery_checks/check_image_url_404.go` (568 lines, last
touched 2026-08-03, `d915d2f1b`). `Run` (`:203`) files three shapes, all `Status: "detected"`,
all handler-less by design (header: *"a stale reference is repaired by removing or repointing
it, which no image generator can decide"*):

| shape | ItemKey | filed at |
|---|---|---|
| `unbacked_path` | `image_url_404:<filename>` | `:244–272` |
| `empty_src` | `image_url_404:empty-src` | `:291–308` |
| `bare_token_src` | `image_url_404:bare-token-src` | `:323–346` |

**It never appends to `result.Resolved`.** `grep -n Resolved` on the file returns one hit, a
comment at `:68` ("resolved from a schema"). The control — `check_site_structural_validity.go`,
which files the same class of flag-only finding — has **5** `ResolvedFinding` emissions and
closes `head_essentials_missing` the moment its page re-probes clean (`:1116`); the
improvement_loop lane watched 343 of those close themselves in a day.

The closing machinery is generic and would work here unchanged:
`ResolvedFinding{ItemType, ItemKey | AllOfType, Reason}` (`discovery_checks/registry.go:121`)
→ `resolveWorkItems` (`actions/work_items_common.go:559`) sets `status='complete'` by
`site_id + item_type (+ item_key)` for any row not already closed. **It does not filter on
`handler_agent`** — so flag-only rows are retractable; this check simply never asks.

**Same hole, one key over:** `check_asset_reference_404.go` retracts only the
`unresolvable_reference` key (`:339`); its `empty_src` key (`assetRefItemKey("empty_src","")`,
`:373`) has no retraction path.

Contrast the check's own header, which is careful about the OPPOSITE direction (false
positives — 1 of 127 measured) and says nothing about false persistence. The measurement that
condemned the old predicate compared "reports a WORKING image | SILENT on a broken one"; it
never had a column for "still reports an image that has since been fixed".

## 3. Why it went unnoticed

The improvement loop reports `complete_clean` over the top of a flag-only pile (improvement_loop
PLAN §3), and `insertWorkItem` writes with `dropOnConflict` (`load_work_item_actions.go:1787`),
so a re-run that re-finds the same key is a silent no-op and a re-run that does NOT re-find it
is… also silent. Nothing distinguishes "still broken" from "fixed" at the row. The only
consumer of these rows counts them (`countUnroutableDetected`, `work_items_common.go:359`) and
nothing reads the count.

## 4. Fix candidates, ordered by what closes the door

1. **Per-key retraction at the end of `Run` (the door).** Load this site's open
   `image_url_404` keys (pattern: `loadOpenArchivedItemKeys`,
   `check_archived_page_still_serving.go:618`, status `NOT IN` the closed set), subtract the
   keys this run filed, and emit one `ResolvedFinding{ItemKey, Reason: "re-scanned: <path> is
   now backed by asset <key> / no longer referenced"}` per survivor. Covers all three shapes.
   The `empty-src` and `bare-token-src` keys are site-singletons, so "not re-filed this run" is
   exactly "no longer observed". Mutation to prove it: delete the retraction block, the new
   test must fail. Same edit, same shape, for `check_asset_reference_404.go`'s `empty_src`.
2. **`AllOfType` when the scan files nothing** — three lines, honest, and insufficient: it
   clears a site only when EVERY image is fixed, so a site with one stubborn 404 keeps its
   twenty repaired ones open. Acceptable as a first step only if candidate 1 is committed to.
3. **The class fix, so the next check cannot make this mistake:** a registry-level test that
   every check whose `Run` can append to `result.Findings` for an item_type also has a path
   that appends a `ResolvedFinding` for it — or names itself in an allowlist with a reason
   (`check_site_unreachable` has one: routing). The producer census of 2026-09-03 found 13
   flag-only producers with **no shared constructor** and a copy-pasted comment at 11 sites;
   a contract test is the only place a rule about all of them can live. Shared seam →
   `090` + council.

Order recommended: 1 (with the sibling key), then 3. Not 2 alone.

## 5. How to verify it is fixed AND live

Go change → inert until the chassis image rolls. Then, per CLAUDE.md, ask the pod what it runs
(`build provenance` line, or the binary probe with both controls). Then:

```sql
-- rows the CHECK closed (its Reason lands in result.reason); expect > 0 within one rotation
SELECT count(*), min(completed_at)
  FROM site_work_items
 WHERE item_type = 'image_url_404' AND status = 'complete'
   AND result->>'reason' LIKE 're-scanned%';
-- and the standing pile should fall from 87 for reasons the rows themselves state
SELECT count(*) FROM site_work_items
 WHERE item_type = 'image_url_404' AND status = 'detected' AND (handler_agent IS NULL OR handler_agent = '');
```

⚠ **A falling count alone is not the proof** — the work-item-archiver and hand closes also
move rows. Read `result.reason`, and check `completed_at` against the roll (the
improvement_loop lane credited another lane's batch once; NOTES §(oo)).

## 6. Not this bug

- Whether the 87 are TRUE today. Probably most are (the check's false-positive rate is 1/127)
  — but that is the point: nobody can tell from the row, and this bug is about the row.
- Who should ACT on a true `image_url_404` — that is plan item 4 of the improvement_loop lane
  (the flag-only pile has no consumer; `livespec/unarmed_completers.go:85` records that
  `image-url-404-handler` escalated 3/3 to a human).
