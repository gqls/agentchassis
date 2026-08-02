# PLAN — 2026-08-02 — `bugs_open/136` (section-editor slug): the sibling writers of
# `page_components.rendered_html` that have no dead-internal-link repair

**Bug file:** `bugs_open/136_HANDOFF_2026-07-28_section_editor_and_three_siblings_persist_links_with_no_repair.md`
— resolve **by slug**, `136` is one of the ambiguous numbers (the other 136 is the
domain/pipeline rename).

**Ownership check before starting** (2026-08-02 ~10:20 BST):
- `scripts/who-owns.py section_editor_and_three_siblings` → the only commit whose subject
  is this bug is its own filing (`674f5dafc`). The workstreams it named
  (`bugfix_100_101_scrape_provenance`) match on the **number**, and their commits
  (`docs(136,144)`, `docs(136): OWNER RULING`) are all about the *other* 136.
- Live-session sweep of every `.jsonl` transcript touched in the last 10 hours for
  `ApplySectionEditAction|section_editor_actions|repairSectionLinks|save_sections_link_repair`:
  one session at 31 hits (`5fec9adb`) — it is the **137** lane (280 hits on
  `bugs_open/137`, whose fix is adjacent to link repair), not this bug. Everything else
  ≤8, i.e. incidental.
- **Not taken:** `bugs_open/155` (a live session has 119 transcript hits on it) and
  `bugs_open/093` (its own last update says it is no longer a code task — it is blocked
  on `bugs_open/083`, and 083 is hot in two sessions).

---

## What the bug is

`bugs_open/079`'s fix put dead-internal-link repair inside `SavePageSectionsAction` — the
chokepoint for the full-page section save. The council's `bug_historian` seat objected that
this was not shown to be the *only* writer of `page_components.rendered_html`, and it was
right: `ApplySectionEditAction` (targeted single-section edit, LLM-authored),
`CreateReportPageAction` (report dossier section, LLM-authored prose embedded raw) and
`RebuildBlogListingAction` all write that column with no repair at all.

The family is the one CLAUDE.md and `016b` §9 both name: **a mechanism made generic, then
guarded at one call site while the siblings stay open.**

## What I verified before writing any code (2026-08-02)

1. **Still valid.** `grep -rn "repairSectionLinks\|RepairPageLinks"` over `platform/` and
   `internal/` returns call sites in exactly three places: the build gate
   (`validate_page_content.go:412`), the outbound rerender seam
   (`rerender_link_repair.go:67`) and the section save (`save_sections_link_repair.go:77`).
   None of the sibling writers appears. Nothing has fixed this since it was filed.
2. **The report path really does embed LLM prose raw.** `renderReportSection`
   (`create_report_page_action.go:312`) escapes every deterministic field with
   `html.EscapeString`, but the four prose sections are written with a bare `%s`
   (`create_report_page_action.go:413`) — so an invented `<a href="/pricing">` in
   `summary_html` reaches `rendered_html` verbatim. The bug file's claim holds.
3. **The blog listing is NOT a member of the class, and this is measured, not asserted.**
   Its article hrefs come from `pages.url` (`blogPostsQuery` → `articles[].url` →
   `ensureArticleLinks`), which is the same table the repair index is built from, under a
   *stricter* predicate. And the live template it renders carries exactly one anchor whose
   href is `{{.url}}`:

   ```
   SELECT function, (SELECT count(*) FROM regexp_matches(html_template,'<a[^>]+href','gi')),
          (SELECT string_agg(m[1],' | ') FROM regexp_matches(html_template,'href="([^"]*)"','g') m)
   FROM content_components WHERE function='content-listing' AND is_active;
   -- content-listing | 1 | {{.url}}
   ```

   So a repair pass there could only ever be a no-op. **Excluded deliberately**, with the
   evidence above rather than a judgement.
4. **The volume the bug file marks `[UNMEASURED]` is now measured** — see
   `NOTES_sibling_link_repair.md` § census. Fleet-wide, stored `page_components.rendered_html`
   holds **30 distinct (page, href) internal links that do not resolve** against their own
   site's `pages` rows, across **7 sites**: 14 would be *rewritten* (a real `.html` target
   exists) and 16 would be *unlinked*. **What that number is NOT:** it is the standing
   stock, and it cannot be attributed to a writer — a stored link cannot say who wrote it.

## Scope decision, and what is deliberately left out

| writer | in this change? | why |
|---|---|---|
| `ApplySectionEditAction` | **yes** — both edit types | the bug's candidate 1; the documented targeted-edit path, reachable from `tool-improver` |
| `CreateReportPageAction` | **yes** | the bug's candidate 2; LLM prose embedded raw. **Inert today** — `SELECT count(*) FROM pages WHERE page_type='report'` → 0 |
| `RebuildBlogListingAction` | **no** | not a member of the class — hrefs come from `pages.url` (measured above) |
| `create_tool_component_action.go`, `deploy_tool_action.go` | **no, and it is the next candidate** | the census puts 8 of the 30 live rows in tool slots (`tool-cta`, `tool-*`). Left out on *collision* grounds, not merit: those two files are in play across the tool lanes (146/149/154/126) and a pathspec commit still takes a same-file passenger. Recorded in the bug file as the ranked next step |
| candidate 3, a `persistSectionHTML` every writer must call | **no** | the bug file says do not smuggle it in, and CLAUDE.md's platform-seam ruling makes it architecture-scope: it touches nine files and several live proven paths |

## The design

One new file, `platform/orchestration/actions/component_link_repair.go`, holding a single
entry point for the **one-component** case:

```go
func repairComponentHTMLBeforePersist(ctx, params, siteID, domain, pageName, pageURL,
                                      actionName, html string, logger) string
```

- It **wraps the existing pure seam** `repairSectionLinks` with a one-element
  `[]SectionData` — exactly the shape the bug file proposed — so the repair SEMANTICS
  keep one definition and one test file. No second copy of the rules.
- Same reversal lever (`repair_internal_links`, default ON), same fail-open on an
  untrustworthy page index, same `CONTENT_LINK_REPAIR_DETAIL` record with `ActionName`
  discriminating the path, so "which path stopped repairing" stays answerable — the
  question `bugs_open/097` exists to keep answerable.

**Repair at the persistence point, not the render point.** In `ApplySectionEditAction` the
swap branch used to persist inside `applyComponentSwap` while the content-edit branch
persisted in the caller — two persist sites, so a repair at either would leave the other
open, which is this bug's own shape one level in. So the swap's `UPDATE` moves out to the
caller: one repair call, immediately before one `switch` that persists. A future third edit
type cannot bypass it without deleting the call.

**The third skip-log insert is why the shared one gets extracted now.** The bug file's own
tail ticket ("unify the two skip-log inserts") is discharged here rather than deferred: the
alternative was to write a *third* near-identical `agent_error_log` INSERT, which is the
duplication the `reuse_agent` seat objected to, made worse. Each caller keeps its exact
`agent_type` / `action` / message so the rows stay byte-identical to what queries already
match.

## What this does NOT fix, stated plainly

- **It does not clear the standing stock.** The 30 live rows above are already stored and
  already deployed; nothing here rewrites them. A sweep is a separate, riskier change
  (unlinking is destructive) and `bugs_open/116` — the phantom-link discovery check has
  never run — is the proper home for detection.
- **It repairs `rendered_html`, not `content_data`.** Same limit as 079's fix: a re-render
  from `content_data` can reintroduce the href in storage. The deployed artefact is still
  covered, by the outbound rerender seam (`repairOutboundPageLinks`, 097).
- **It cannot see a link to a page that is planned but not yet created** — that link is
  unlinked, exactly as the build path already unlinks it. `bugs_open/097` owns that
  question; this change does not widen it.

## Verification plan

1. `go build ./...` and `go test ./platform/orchestration/actions/...`.
2. New unit tests: the single-component seam rewrites/unlinks/no-ops; the config lever
   turns it off; an untrustworthy index ships the HTML unchanged (fail-open); **and a
   mutation check** — the guard is proven by breaking it, not by a green run
   (`mutate-the-code-to-prove-the-guard`).
3. Council gate submission (advisory) before/alongside the commit.
4. Pod-grep with a discriminating marker + a positive and a **negative** control after the
   next chassis roll — `bugs_open/153`: a roll is not evidence your fix shipped.
