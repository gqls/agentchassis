# 241 — The planner's URL canonicaliser silently moves every flat tool/guide URL on an adopted site

**Filed 2026-08-10.** Found while planning the framework rebuild of loancalculator.co.uk
(owner-directed), BEFORE the planner was run — so no damage occurred. The partial fix (the
representational half) is committed; the plumbing half is NOT.

**090 status:** not filed through the diagnosis loop. Substituted first-hand verification,
declared per the owner ruling of 2026-07-31: I read the deciding arm of the mechanism
directly (`page_canonical.go:152-173`, the tool/guide cases), read the overwrite site
(`site_db_actions.go:1142`, `url = EXCLUDED.url` unconditional on conflict), read the
consumer that makes it live (deployer takes the file path from `pages.url`), and measured
the blast radius on the live DB (query below). The claim is mechanical, single-file, and
was reproduced by unit test in the same session. What I did NOT verify live: an actual
planner run moving a URL — deliberately, since preventing exactly that was the point.

## The mechanism

`CanonicalisePage` (`platform/orchestration/datahelpers/page_canonical.go`) is the single
canonicalisation point for page identity (name, url, page_type) — doc 029/030 lineage,
deliberately convergent across adoption and planner shapes. For the nested roles it
returns:

- `role=tool`  → `/tools/<slug>/index.html`
- `role=guide` → `/guides/<slug>/index.html`
- `role=game`  → `/games/<slug>/index.html`

**No input produces `/tools/<slug>.html`.** The flat shape exists in the vocabulary
(blog-post and entity-page both emit `/<dir>/<slug>.html`) but the three nested roles
cannot reach it.

An adopted site that already serves flat URLs is therefore unrepresentable to the planner.
When the planner runs: `SyncPagesToDBAction` canonicalises every planned page
(`site_db_actions.go:281`), `upsertPage`'s `ON CONFLICT (site_id, name) DO UPDATE` sets
`url = EXCLUDED.url` **unconditionally** (`:1142`), and the deploy path derives the output
file path from `pages.url` (`git_deployer_actions.go:435`). Net: every flat tool/guide URL
is rewritten in place on the live rows, the next deploy publishes N new files, and the N
indexed URLs keep serving stale content. No error anywhere; every individual step is
working as designed.

## Measured blast radius (2026-08-10, live DB)

```sql
SELECT page_type, count(*), min(url) FROM pages
WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND status='active'
GROUP BY page_type;
-- guide 13  /guides/can-i-overpay.html
-- tool  11  /tools/application-tracker.html
-- content 1 /legal.html · landing 1 /index.html
```

**24 of 26 live URLs move** the moment a plan syncs. loanandmortgagecalculator.co.uk and
loancash.co.uk serve the same flat shape (they are verbatim ports today, but the moment
either is re-adopted into editable form and planned, the same rewrite fires). Guides have
a partial in-framework escape — plan them as `role=blog-post, parent_section=guides` —
but that lies about the page's type to preserve its address. Tools have no escape at all.

## Fix, half shipped

**Shipped (this commit):** `PageDescriptor.FlatURLs bool` — opt-in, default false, changes
only the URL arm of tool/guide/game via `nestedOrFlatURL()`. Twelve call sites unaffected
(zero value = old behaviour byte-for-byte; the pre-existing test corpus proves it). Five
new test cases cover the flat arms. Registered as **BLD-018** in the concept register,
same commit, per the 2026-07-29 ordering ruling.

**NOT shipped (owed):** the plumbing. A site-level flag — recommendation: `url_shape:
"flat"` in the site's `structure` spec aspect — read ONCE and passed to BOTH
canonicalisation surfaces:
- `write_site_plan_action.go:392`
- `SyncPagesToDBAction`, `site_db_actions.go:281`

⚠ Those two surfaces diverging is a **known, previously-shipped regression** — the comment
at `site_db_actions.go:245-254` records it (flat pages row vs nested plan row). One flag,
read once, passed to both, or don't ship it.

## How to verify

- Unit: `go test ./platform/orchestration/datahelpers/` (five FlatURLs cases + the whole
  pre-existing corpus as the default-unchanged proof).
- After plumbing ships: run the planner on a `url_shape:"flat"` site and assert
  `SELECT count(*) FROM pages WHERE site_id=$1 AND url LIKE '%/index.html' AND
  page_type IN ('tool','guide')` returns 0, and that the pre-plan URL set equals the
  post-plan URL set exactly.

## Relations

- Owner decision 2026-08-10 (loancalculator rebuild): "Fix the framework first, then
  rebuild" — this defect is why the rebuild is blocked on the plumbing.
- The rebuild lane's handoff carries the full context:
  `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-10_framework_rebuild_continue_here.md`

## STATUS 2026-08-11 — plumbing WRITTEN and on HEAD; roll + seed owed

The owed plumbing above is built, exactly to the one-flag-one-read-both-callers
constraint:

- **`siteUsesFlatURLs`** (`platform/orchestration/actions/site_url_shape.go`, new) — the
  ONE reader: `site_specs` aspect `structure`, key `url_shape`; `"flat"` → true; absent
  spec/key, any other value, nil DB or read error → false (nested default, all existing
  sites byte-identical).
- Both surfaces thread it into their `PageDescriptor` (`WriteSitePlanAction`,
  `SyncPagesToDBAction`), and `TestFlatURLFlagReachesBothCanonicalisationSurfaces`
  (`site_url_shape_test.go`) pins the pair mechanically — a surface that drops the flag,
  or reads it around the shared helper, goes red. Six sqlmock cases pin the helper.
- `go build ./platform/...` + both test suites green **at committed HEAD via
  `git archive`** (not the dirty shared tree), 2026-08-11.

**Commit provenance, stated plainly:** the code reached HEAD in **`7a066dba1`** — the
`bugs_open/215` lane's commit — as a **declared same-file passenger** (both lanes edited
the same two surfaces; their commit message names the sweep and the reason: they could
not commit their hunks without taking these, and the untracked helper+test were required
for HEAD to compile). Forward-only holds; nothing was lost; there is no
`Council-Submitted:` trailer on that commit because it was not this lane's commit.

**Council round 2 (the plumbing): corr `70256656-4ada-465e-b959-096ae7225eb9`**,
submitted 2026-08-11. Round 1 (the field): APPROVED, corr `6fdb9ce6-…`, and its
editquality objection ("the plan reads as a fix but is only a prerequisite capability")
is what this round completes. The round-1 advisory questions are answered in the round-2
submission with the queries run: the plan-sync write path is `site_db_actions.upsertPage`
(`url = EXCLUDED.url` unconditional; the other two upsert helpers are not in the planner
path), and the "second flat mechanism" is verbatim adoption (fidelity=locked), which
preserves rather than synthesises — it is how the flat URLs got in, not a rival.

**Still owed before this is fixed-and-live:**
1. Roll the chassis image carrying `7a066dba1`+ (do NOT roll while a council run is in
   flight — a roll kills it), pod-grep `url_shape` positive + a negative control.
2. Seed `url_shape:"flat"` into loancalculator.co.uk's `structure` spec
   (supersede-then-insert). ⚠ Seeding on any OTHER site first requires measuring that its
   live shape IS dir-flat — the flag cannot express root-flat or arbitrary crawl shapes.
3. The verification block above ("How to verify", post-plumbing) — unchanged, still the
   test.

## STATUS 2026-08-11 (later) — round-1 code LIVE and MEASURED; round 2 REVISE; round 3 submitted + committed

- **The round-1 plumbing is LIVE without a roll of ours:** `7a066dba1` is an ancestor of
  `038211dd8`, the commit the 215 lane artefact-verified into `v1.0.1288`. Probed on BOTH
  replicas (grep -ac /proc/1/exe): `siteUsesFlatURLs`=3, `url_shape`=2; near-miss controls
  `siteUsesFlatURLt`/`url_shapf`=0. Owed item 1 of the previous block is DONE.
- **The seed is applied and verified** (`SEED_2026-08-11_url_shape_flat.sql`): current
  structure row has `url_shape='flat'`, adoption keys intact, 27-entry pages list, one
  current row. Owed item 2 DONE. First run aborted on my wrong 26-entry expectation — the
  DO/RAISE guard was right; corrected + annotated in the seed.
- **Round 2 (corr `70256656`) returned REVISE** — gating: the LANDMINES 2026-08-11 entry
  "Re-adopting a site silently drops the structure spec's opt-in flags" names this very
  key; adoption's supersede+INSERT would drop `url_shape` on any re-adoption and the bug
  would return. Also: wire the other live-site URL synthesisers; the seed cannot rely on
  `pinned`; prior_art asked whether `siteUsesFlatURLs` pre-existed (no — the landmine
  postdates the code; timeline in the round-3 submission).
- **Round 3 submitted (same trail corr `70256656`) and committed `19acfc895`:**
  `carryForwardStructureSpecKeys` (adoption preserves ALL unknown structure-spec keys —
  covers PLAN-048's three gates too, 215 lane notified in
  `brochure_component_library/NOTIFY_2026-08-11_readoption_flag_drop_fixed_by_241_lane.md`);
  flag wired into apply_gap_plan new_page, create_tool_component, deploy_tool
  resolveToolPageIdentity (new-tool arm; stored identity still wins — pinned);
  contract test widened to per-file exact descriptor counts with the blog-post
  exclusions stated. **Round-3 code is in NO image yet** — probe
  `carryForwardStructureSpecKeys` + a near-miss control at the next roll.
- Deliberately NOT done, registered on BLD-018: consolidating the two typed readers of
  the structure row (`siteUsesFlatURLs` / `siteIdentityPolicyFor`) — that is the 215
  lane's file, hours old and council-approved; theirs to absorb.
