# 208 — `page-rebuild` git-commits a regenerated `owned` page BEFORE the owned-page guard refuses it, so a live tool is destroyed and only the DB write is protected

**Filed 2026-08-06** by the `bugfix_201_page_content_writer_dispatch` lane, found in **pre-flight
for an owner-authorised full rebuild of `ai-agent-orchestration.com`** — the dispatch was **NOT
fired** because of this. **OPEN, unowned.**

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
