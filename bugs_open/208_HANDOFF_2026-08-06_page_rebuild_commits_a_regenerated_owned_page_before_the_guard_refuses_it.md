# 208 — `page-rebuild` git-commits a regenerated `owned` page BEFORE the owned-page guard refuses it, so a live tool is destroyed and only the DB write is protected

**Filed 2026-08-06** by the `bugfix_201_page_content_writer_dispatch` lane, found in **pre-flight
for an owner-authorised full rebuild of `ai-agent-orchestration.com`** — the dispatch was **NOT
fired** because of this. **OPEN, unowned.**

---

## STATE 2026-08-07 — **LIVE on v1.0.1262 and pod-verified; the guard has NOT yet been observed to FIRE**

Rolled this morning. Verified in the running binary, not at the tag: `OWNED_PAGE_GUARD` **3**,
`OWNED_PAGE_GUARD_UNCHECKED` **1**, `include_owned` **1**; the string `f5710d6b0` **removed**
(`"guard standing down for this page"`) reads **0**, which is what proves this is the latest
commit rather than an older image sharing some symbols; a fabricated string reads **0**. All
**41** pods running this binary share one image digest, so that is the fleet by identity, not a
sample. Baseline re-checked: **13 of 14 served bodies byte-identical**; the 14th is the
never-built 404 (`planned`, 0 components, DB row untouched since June) whose hash moved because
the site's shared 404 template changed.

> **⚠ Still OPEN, and the reason is not bureaucratic.** Zero `owned_page_review` rows exist from
> `get_pages_to_build` or `assemble_page` — all 12 in the table are reconcile's and pre-roll. No
> dispatch has hit a site with owned pages since the roll, so **the guard has never been asked a
> question.** "Fix live + baseline clean" proves presence and absence of harm; it does not prove
> bite, and reading it as proof is 016b §9's silent-gate trap. The behavioural canary
> (synthetic page, never a real tool — see "How to verify") is what closes this.

## STATE 2026-08-06 (evening) — FIXED IN CODE. Taken by the `bugfix_208_owned_page_commit_before_guard` lane

**Commit `cb7b4d759`.** Council **SUBMITTED**, verdict not yet read — corr
`5d1dcb10-7929-431e-b9e5-496992ce3229` (the commit carries `Council-Submitted:`, not
`Council-Reviewed:`). Registered as **PBP-036** in the same commit as the seam. Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_208_owned_page_commit_before_guard/`.
**Stays OPEN: Go-only, so it is inert until an image carries it.** No DB config half, so a roll
is the whole of the deployment.

**Ownership note for the next reader:** `scripts/who-owns.py 208` said OWNED within hours of
filing, because the *filing* commit touches the file. It was not owned — the filing lane's own
transcript says "208 is handed off". Live `.jsonl` transcripts are the only source that sees an
uncommitted session; that is what cleared it.

### The filing was right, and three things were bigger than it said

1. **Three pipelines, not one.** `pageflow-builder` and `site-work-orchestrator` run the same
   `assemble_page → deploy_page (git_commit) → save_sections` order. `pageflow-builder` also
   selects `planned`.
2. **14 pages over 6 domains, not 2 over 1** — including `tool-arena-interface` and
   `tool-gauntlet` on **vonc.com**, i.e. the site of the very clobber migration 164 was written
   for. Plus an `include_all` branch with no status filter, which would reach the ~189 owned
   pages at `deployed`.
3. **The refusal was aborting the whole batch.** `continue_on_error` is unset on all four build
   loops, so `save_page_sections`' hard refusal failed the entire workflow; with selection
   ordered `nav_order, name`, every page after the owned one never rebuilt either.

### Fix candidate 1 was the right instinct; the fix is that plus the arm it cannot reach

- **Layer 1, selection** — `queryPagesForBuild` excludes `owned` in **both** branches, behind a
  new `include_owned` step-config key defaulting false. **Both Go callers** inherit it, so
  `write_build_items` also stops minting `needs_page` items for owned pages. This answers the
  filing's own ⚠ about other callers: they are `page-rebuild`, `pageflow-builder` and
  `write_build_items`, all named, and none legitimately wants an owned page.
- **Layer 2, composition** — `AssemblePageAction` returns the action's **existing**
  `{skipped:true, skip_reason}` shape for an owned page. `git_commit` already honours it
  (`checkUpstreamSkipped`), so **no agent config changed anywhere**. Threaded through
  `save_page_sections` and `update_page_status` so the skip survives the iteration.

**Fix candidate 2 ("move the guard earlier") is what Layer 2 is**, but at the measured seam:
`assemble_page` has exactly the three exposed consumers and nothing else. **Candidate 3
(reorder) stays rejected** for the reason the filing gives, and the file was right that it is
tempting — note `page-rerender` and `page-build-handler` already save-before-commit, so the
reorder is directionally correct and still the wrong instrument. **Candidate 4 (de-queue the 14
pages) was not done**: it leaves the trap armed, and the pages are now safe where they sit.

> **⚠ The filing's closing question — "worth asking what else has a DB-row guard behind a git
> commit" — has an answer that changes the fix, so read it before touching this.** `git_commit`
> looks like the right place for the guard and is **not**: `page-rerender` and `section-editor`
> commit pages too, and those are how owned pages *legitimately* deploy — migration 164 says
> that path "is deliberately NOT gated". A guard there would stop tool pages deploying at all.
> Recorded in `LANDMINES.md`.

### What is NOT established, stated plainly

- `[UNDETERMINED, EVIDENCE REAPED]` Whether this has ever actually fired. **11 of the 14 fix
  items `site-work-orchestrator` has consumed targeted `owned` pages on webdesign.co.uk
  (2026-08-04) and all 11 failed** — the exact signature of commit-then-refuse. But all 11 pages
  serve working tools today, and terminal `orchestration_states` are reaped at ~24h, so whether
  those runs reached `deploy_page` **cannot be recovered. Do not assert it.**
- `[MEASURED]` The damage is **not** realised: all 13 live owned pages still serve working tools
  (`BASELINE_2026-08-06_owned_pages_served.txt`, sha256 per body — the control set for the
  post-roll re-check).
- **The `needs_page → page-build-handler` route was never the dangerous one** (all 158 rows
  fleet-wide route there; it saves before it deploys). A future fix aimed at it is aimed at a
  path that already works.

### How to verify after the roll (the filing's own recipe, plus a negative control)

1. Pod-grep `OWNED_PAGE_GUARD` — **0 on v1.0.1261**, so a ready-made negative control; add a
   fabricated string in the same exec to prove the pipeline rather than the spelling.
2. **Synthetic** canary only, never a real tool page: a throwaway `owned` + `needs_rebuild` page
   on a low-value site, then dispatch `page-rebuild`. Assert its file's git history untouched,
   exactly one `owned_page_review` item (a second dispatch files none), the run **completed**
   rather than failing at `save_sections`, and — the control — a sibling **generic**
   `needs_rebuild` page in the same run did rebuild.
3. Re-run the baseline sweep: all 13 bodies byte-identical by sha256.

### Follow-up deliberately NOT included

`UpdatePageStatusAction` also stamps `deployed` after an **ordinary content-failure** assembly
skip. That is a real defect of the same family, but un-stamping it changes retry behaviour on
the fleet's main build path (the page would be re-selected every run), a wider blast radius than
this bug. Left out on purpose, and pinned by `TestUpdatePageStatus_OrdinarySkipStillStamps` so a
future widening is a decision rather than a side effect.

---

> **On the "diagnosis before debugging" default (owner ruling 2026-07-31):** `090` was **not**
> run. Substituted first-hand verification, declared rather than omitted: I read every link in
> the chain directly — the selection SQL, the live step graph and both step configs from
> `agent_definitions`, and the refusal in Go. Each is quoted below with its line or config. No
> link is inferred. What I have **not** done is *induce* the failure (that would mean destroying
> a live tool page), so the damage is **predicted from the mechanism, not observed** — flagged
> as `[INFERRED]` where it matters.

## The chain, each link read directly

**1. Selection ignores `rebuild_policy`.** `page-rebuild`'s `get_pages_to_rebuild` step is
`get_pages_to_build` with config `{"build_statuses": ["needs_rebuild"], "include_all": false}`.
`queryPagesForBuild` (`get_pages_to_build_actions.go:120-130`) is:

```sql
SELECT ... FROM pages
WHERE site_id = $1 AND status = 'active'
  AND COALESCE(build_status, 'planned') IN (...)
```

**There is no `rebuild_policy` clause anywhere in that file** — `grep -n rebuild_policy
get_pages_to_build_actions.go` returns nothing. So an `owned` page at `needs_rebuild` is selected
exactly like a `generic` one.

**2. The loop commits before it saves.** `build_pages_loop`'s order, from the live step graph:

```
plan_sections → write_page_content → review_page_content → check_review_approved
   → assemble_page → deploy_page → save_sections → update_page_status → complete_page
```

`deploy_page` is `action: git_commit`, config `{"page_field": "current_page", "domain_field":
"site_record.domain", "content_field": "assembled_page.html"}`, `next_step: save_sections`.
So the **regenerated HTML is committed to the site repo** — and the site deploys from that repo —
one step *before* the guard runs.

**3. The guard is in the step that runs second.** `save_page_sections_action.go:140-156`:

```go
// A rebuild_policy='owned' page belongs to a tool/widget or is a …
SELECT COALESCE(rebuild_policy, 'generic') FROM pages WHERE id = $1
… "page %s is rebuild_policy=owned (tool/widget-owned): a generic section save would destroy …"
```

It refuses correctly. It refuses **the database write**. The file in the repo is already replaced.

**Net:** an `owned` tool page swept into a rebuild is fully recomposed by an LLM, its generic
prose committed over the working tool, and then the run errors at `save_sections` — leaving the
**live page replaced** while `page_components` still describes the tool that is no longer served.
An inconsistent state that no rerender repairs, because `content_data` still describes the old page.

## Why it is live right now, and what nearly happened

`ai-agent-orchestration.com` has **two `rebuild_policy='owned'`, `page_type='tool'` pages sitting
at `build_status='needs_rebuild'`**: **`agent-complexity-estimator`** and **`password-entropy`**.

`page-rebuild` **sweeps in every `needs_rebuild` page on the target site regardless of which pages
the operator named** — documented by the `feature_021` lane in its own handoff and in
`scripts/rebuild_pages.sh`'s header (note 2). So **any** real dispatch of the operator bulk
rebuild at this domain, with any page list, takes both tools.

**[INFERRED — the mechanism is verified, the damage is not observed]** I did not fire it. The
prediction is that both tool pages would be replaced by generic regenerated prose on the live
site. Given the cost of being right, the correct order is to fix or exclude first and test after.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Exclude `owned` at SELECTION** — add `AND COALESCE(rebuild_policy,'generic') <> 'owned'` to
   `queryPagesForBuild` (or as an opt-in config field defaulting to excluding). **Preferred:** an
   owned page is by definition not pipeline-owned (`adopt_verbatim.go:612`), so it should never
   enter a generic rebuild loop at all. Makes the state unreachable rather than caught late.
   ⚠ Check the other callers of `get_pages_to_build` before changing the shared default —
   `page-build-handler` and the build pipeline use it too, and a fresh `planned` page is a
   different case from a `needs_rebuild` owned one.
2. **Move the guard earlier** — check `rebuild_policy` before `deploy_page`, e.g. a conditional at
   the top of the loop body that skips to `complete_page`. Cheaper to reason about, but it leaves
   the LLM spend (the page is still fully recomposed before being discarded).
3. **Reorder `save_sections` before `deploy_page`** so the refusal precedes the commit. Tempting
   and probably wrong: it changes the commit/save ordering for *every* page, and `deploy_page`'s
   output (`page_deployed.commit_sha`) is consumed downstream by `update_page_status`.
4. **Operational stopgap, not a fix:** take the two pages out of `needs_rebuild` before any
   dispatch. Leaves the trap armed for the next site and the next operator.

## How to verify a fix

Do **not** verify by rebuilding a real tool page. Take a site with an `owned` page at
`needs_rebuild`, run `get_pages_to_build`'s query with the fix applied, and assert the owned page
is absent from the returned set. Then a real dispatch on a site with a *generic*-only
`needs_rebuild` set, checking the owned page's `updated_at` and its served HTML are both
**unchanged**.

## Related

- **Same shape, different table — this is a known class:** `LANDMINES.md` (asset lock, ~:1789):
  *"That guard is real and it works. It also protects **only the DB row**, and in every one of
  these actions the `sendGitCommitRequest` that replaces the file in the site repo runs
  **first**."* That entry is about `assets`; this is the same defect in `pages`. Worth asking
  what else has a DB-row guard behind a git commit.
- `bugs_closed/037` — `needs_rebuild` pages unprotected by the **re-plan** guard. Same theme
  (`needs_rebuild` is an under-protected state) but a different mechanism (`realisedPageIsBuilt`
  in `v3_site_actions.go`), so this is not a reopening.
- `features_open/021` / `feature_021_operator_bulk_page_rebuild` — the operator entry point whose
  first real dispatch surfaced this. Its handoff already documents the sweep-in; what it does not
  say is that a swept-in **owned** page is damaged rather than skipped.
- `bugs_closed/084` — owned pages and browser verification (different concern, same population).
