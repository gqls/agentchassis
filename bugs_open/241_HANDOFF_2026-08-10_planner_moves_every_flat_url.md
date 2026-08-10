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
