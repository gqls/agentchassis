# NOTES — bugfix 128, `image_url_404`

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-07-31 — picking the bug up

**Ownership check, because `who-owns.py` says OWNED and is wrong here.** The tool
names `brochure_component_library` (active, 12 mentions) — but that lane's own
`HANDOFF_2026-07-28b` lists 128 as *"read, still unowned"*, and the 07-29 diagnosis
section in the bug file says the same. `who-owns.py` reads **commits**, so a filing
workstream and an owning workstream look identical to it.

Did the check memory says is the real one: grep the live session transcripts for the
target's **code symbols**, not its number.

```
find ~/.claude/projects/-home-ant-projects-agentchassis -name '*.jsonl' -mmin -720 -size +5k \
  -exec grep -lE "check_image_url_404|loadKnownAssetPurposes|collectImagePathReferences" {} \;
```

33 sessions matched, but reading the surrounding context showed every one was a
directory listing or a data row, not work: the two with double-digit hits (`2c81fa62`,
`659b83fc`) were counting `image_url_404` **work items** while working `136`/`155` and
`083`/`033`. Nobody is in this file. Neighbouring bugs (`114`, `142`, `152`, `155`) ARE
actively worked, so the collision risk is subject-matter, not file-level — worth noting
before editing anything shared.

## Re-measuring before believing the file

The bug file's own numbers are two days old and it says in terms that two of its three
acceptance URLs went stale within a day. So: re-derived everything.

Population — all distinct `/assets/images/*` paths in deployed unlocked page
components, which is exactly the check's own scan:

```sql
SELECT s.domain, m[1]
FROM page_components pc JOIN pages p ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
CROSS JOIN LATERAL regexp_matches(pc.rendered_html,
  '(/assets/images/[a-zA-Z0-9_\-]+\.(?:jpg|jpeg|png|webp|svg|gif))', 'g') AS m
WHERE pc.build_status='deployed' AND pc.locked_at IS NULL
GROUP BY 1,2;
```

**127 paths, 13 sites.** Probed every one over HTTP.

**MISSTEP 1, caught by re-probing.** Four paths came back `000` and one `200` that I
expected to fail. `000` is a curl connection error, not a status — the bug file had
already warned about exactly this ("both 200 on retry") and I had read that sentence.
Re-probed all five with a cache-buster: **all 200.** Had I tallied the first pass I
would have reported four false negatives that do not exist, in a document whose whole
argument is about false negatives.

Cross-tab, after correction (ground truth = HTTP):

| predicate | reports a WORKING image | reports a BROKEN one | SILENT on a broken one |
|---|---|---|---|
| purpose/prefix skip (as shipped) | **21** | 11 | **6** |
| `storage.DeployedWebPath` | **1** | **17** | **0** |

The 07-29 diagnosis reproduces exactly. Its "79 of 95 masked" is not comparable
against today's 127-path population — the fleet moved — but the mechanism and the six
masked live 404s are identical.

## Why the 07-29 refutation of "compare paths" did not apply

That session tested a path predicate against `assets.url`/`filename` and refuted it,
correctly. Measured today over 267 active rows:

```
url LIKE 'https://s3.%'        152
filename = ''                  191
storage_path = ''              189
url LIKE '/assets/images/%'    115   ← of which 47 are the UNRESOLVED literal
                                       /assets/images/input-data.asset-key.jpg
```

So three quarters of the table cannot say where a file is served from. **The served
path is derived, not stored** — `storage.DeployedWebPath(asset_key, purpose)`, which
`deploy_image_asset` commits to via the shared `storage.AssetKeyFilename` and which five
writers render through. Not a new predicate: the inverse of the existing one.

## MISSTEP 2 — the fleet-wide false positive I nearly shipped, and the landmine that already existed

I wrote into the new loader's doc comment that `DeployedWebPath` *"already applies"*
`BrandHeadAssetPaths`. **It does not.** For `og_card`, `assetKey == purpose`, so the
`_`→`-` swap inside `AssetKeyFilename` is never reached and the helper returns
`/assets/images/og_card.png` for a file served at `/assets/images/og-card.png`.

Shipped, that would have reported **a 404 for the og card and the favicon of every site
in the fleet** — both referenced from the head, so on every page — inside a fix whose
entire purpose is removing false positives.

Caught by re-reading my own comment against the helper's source before running anything.
Fixed with `storage.IsBrandHeadPurpose` / `storage.BrandHeadAssetPaths`, and pinned by
`TestImageURL404_BrandHeadPathsResolveThroughTheirOwnMap`.

**The part that stings: a landmine for exactly this was written HOURS earlier**, by the
`bugs_open/142` lane (`d671fb2b2`), with the measurement I needed. The `SessionStart`
hook only surfaces entries matching files **already dirty** in the tree, and
`platform/storage/url_helpers.go` was clean when my session began — which is precisely
the gap CLAUDE.md names when it says to grep the file yourself for symbol footprints.
One command, `grep -n DeployedWebPath …/LANDMINES.md`, would have returned the whole
trap. Logged in `WRONG_CALLS.md`.

## The discovery that changed the design — the routing branch was a duplicate

The masking test warned that unmasking activates `knownPurposeMapping` → `image-build-handler`,
"a dormant fleet-wide auto-regeneration path". I went looking for how to gate it safely,
and found the gate was unnecessary: `check_placeholder_image_in_use` **already does
exactly that job** —

| | `image_url_404` recognised branch | `check_placeholder_image_in_use` |
|---|---|---|
| paths | `hero.jpg`, `logo.png` (via basename root) | `/assets/images/hero.jpg`, `/assets/images/logo.png` |
| precondition | not masked ⇒ no asset of that purpose | `if hasAsset { continue }` |
| item types | `needs_hero_image`, `needs_logo` | `needs_hero_image`, `needs_logo` |
| handler | `image-build-handler` | `image-build-handler` |
| prompts | `loadImagePromptsForSite` | `loadImagePromptsForSite` |
| item key | `image_url_404:<basename>` | `placeholder_image_in_use:<purpose>` |

Both are enabled on the same agent (`design-discovery-agent` carries both names in
`default_config`). They differ only in the dedup key, so they would file **two work items
for one repair**. And neither has ever fired:

```sql
SELECT count(*) FROM site_work_items WHERE item_key LIKE 'placeholder_image_in_use:%';  -- 0
SELECT item_type, count(*) FROM site_work_items WHERE item_key LIKE 'image_url_404:%' GROUP BY 1;
-- image_url_404 | 13     (zero needs_hero_image, zero needs_logo)
```

So the branch is deleted, not gated. The fix adds findings and **no new autonomous
repair**. That is a better answer than the safety gate I was about to write, and I only
found it because the test told me to cost the activation first.

## Verifying with the real Go, not a re-implementation

My cross-tab was computed in Python mirroring `DeployedWebPath`. Two checks blind the
same way agree with each other, so I ran the **actual helpers** over the live rows via a
throwaway `cmd/` program in a `git archive HEAD` checkout:

```
reported 18   (page surface)     ← identical to the Python model
reported  2   (chrome surface)   ← identical
```

No disagreement. The 18 = 17 live 404s + webdesign's legacy `hero.jpg` (455KB, serves
200, backed by no asset row) — the one residual false positive, named in the check header
rather than hidden.

## MISSTEP 3 — the package would not build, and it was not my change

`go build ./platform/...` failed on `check_empty_sections.go:249: undefined: datahelpers`.
That file is another session's in-flight edit for `bugs_open/137` (added a
`datahelpers.HasRuntimeFillMarker` call, import not yet added). Verified against
`git archive HEAD` + my file only — which is the standing practice for exactly this
reason, and the reason it exists. All 10 new tests plus the full package suite pass
there.

**And that produced misstep 3b:** the `HEAD` checkout went into `/tmp`, which is a 16G
tmpfs shared by ~30 sessions and was already at 100%. The build died with "no space left
on device" *after printing BUILD_OK*, which reads like success. Moved to `~/.cache`.
There is a landmine for this too, added today by the `bugs_open/092` lane — the second
one I walked into without grepping for.

## Live state at hand-off

- Code committed `beff42809`, **inert until a chassis roll**.
- Council `99dca96a-413a-4bcb-b278-9577f920786d` submitted; running at `gate_adoption`
  when this line was written.
- Six existing `detected` rows keep the old extension-less dedup key and will not dedup
  against new ones; three `robot-hands` rows among the wider set were the old predicate's
  **false positives** and can be cancelled outright. Listed in the bug file.
