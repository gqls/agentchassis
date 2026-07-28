# Handoff — four writers of page_components.rendered_html have no link repair at all

**Filed 2026-07-28 evening. Status: OPEN, unowned. Found by the council gate, not by a
crawl** — the `bug_historian` seat raised it as a medium objection against the
`bugs_open/079` fix (submission `7c24776e-07f8-4c2e-b1b6-ad3e73c6023c`), and the check it
asked for turned up a real gap. Its words: *"The plan does not establish that
SavePageSectionsAction … are the ONLY writers of page_components.rendered_html fleet-wide —
only that they are the only ones with a 'save_page_sections' STEP NAME in current
agent_definitions."* It was right.

## The defect

`bugs_open/079`'s fix puts dead-internal-link repair inside `SavePageSectionsAction`, which
is the chokepoint for the **full-page section save** — 6 live agent types. That is where
LLM-authored body prose normally enters, and it closes the door on that route.

It is **not** the only writer of `page_components.rendered_html`. Measured 2026-07-28
against the working tree (`grep -rln "INTO page_components\|UPDATE page_components"
--include="*.go" .` then narrowed to those that actually SET the column):

| writer | writes | repairs links? |
|---|---|---|
| `save_page_sections_action.go` | full section set | **yes**, after 079 |
| `section_editor_actions.go:1038` (`ApplySectionEditAction`) | one section, LLM-authored | **NO** |
| `create_report_page_action.go:186,218` | report page section | **NO** |
| `rebuild_blog_listing_action.go:322,357` | blog listing section | **NO** |
| `create_tool_component_action.go:300`, `deploy_tool_action.go:374` | tool markup | **NO** |
| `fix_forced_text_colours_action.go:359`, `fix_harcoded_colours_action.go:236` | colour-only rewrite of existing html | n/a — cannot introduce an href |
| `internal/core-manager/admin/page_admin_handlers.go:336` | admin API, human-driven | n/a |
| `cmd/webdesignport/import.go:216,231` | one-off import CLI | n/a |

Verified none of the three prose writers has any repair:

```
grep -n "RepairPageLinks\|loadValidPagePaths\|repairOutbound\|validateInternalLinks" \
  platform/orchestration/actions/section_editor_actions.go \
  platform/orchestration/actions/create_report_page_action.go \
  platform/orchestration/actions/rebuild_blog_listing_action.go
# -> no matches
```

## Which one actually bites

**`ApplySectionEditAction` is the one to fix first.** It is the targeted-section-edit path —
an LLM rewrites one section and it is written straight to `page_components.rendered_html`,
with exactly the same freedom to invent `/pricing` or `/how-it-works` that produced 079's
evidence. It is also the path `bugs_open/117`/`118` route chrome repairs through, and the
path CLAUDE.md's own guidance points sessions at for targeted edits ("targeted edits go
through `apply_section_edit`" — `save_page_sections_action.go:156`). So the more carefully a
session follows the documented practice, the more reliably it bypasses the 079 fix.

`create_report_page` and `rebuild_blog_listing` are lower risk but not zero: the blog
listing's hrefs are template-rendered from real `pages` rows, so they are structurally
sound; report-page section HTML is LLM-authored and is not.

**[UNMEASURED]** No count of phantom links actually shipped through these paths. The 079
evidence is all from the build route. A cheap first query: `agent_error_log` has no repair
rows for these actions *by construction* (they never repair), so the volume has to come from
scanning `page_components.rendered_html` written by them against `pages.url` — i.e. the
`check_phantom_internal_links` discovery check, which `bugs_open/116` says has never run on
any site. **Do not state a volume until that is measured.**

## Fix candidates, ordered by what closes the door

1. **Repair inside `ApplySectionEditAction` before its UPDATE**, reusing `repairSectionLinks`
   from `save_sections_link_repair.go` (already a pure seam taking `[]SectionData`; a
   single-section caller wraps one element). Same fail-open, same `repair_internal_links`
   lever, same `CONTENT_LINK_REPAIR_DETAIL` code with `ActionName: "apply_section_edit"` so
   the origin field keeps discriminating. Smallest change, covers the path that bites.
2. Same for `create_report_page_action`.
3. **The structural option**: a single `persistSectionHTML` helper that every writer of
   `page_components.rendered_html` must call, making an unrepaired write unrepresentable
   rather than merely unlikely. Correct in principle; it touches nine files and several live
   proven paths, so it is an architecture-review change, not a bug patch — see CLAUDE.md's
   platform-seam ruling. **Do not smuggle it in as candidate 1's implementation.**

## Related

- `bugs_open/079` — the build-route half, fixed and awaiting deploy. This file is the
  sibling-path half its own council review surfaced.
- `bugs_open/092` — the upstream cause (the writer never receives its link constraints).
  Fixing 092 would reduce the input to every one of these paths at once.
- `bugs_open/097` — the same shape one layer up: a repair with one call site, siblings
  unprotected. This is that pattern recurring at the persistence layer.
- `bugs_open/116` — the phantom-link discovery check has never run on any site, which is why
  this gap has no measured volume.
- 016b §9 — "fix applied to one representation, the other left stale".

## Also owed, smaller: unify the two skip-log inserts

The `reuse_agent` seat objected (medium) in the same round that
`writeSaveSectionsRepairSkipLog` (`save_sections_link_repair.go`) is a deliberate near-twin
of the insert at `rerender_link_repair.go:52-63` — two independently-maintained INSERTs
against `agent_error_log` with the same shape and intent. The duplication was disclosed and
reasoned about in the submission rather than hidden, and the council accepted the deferral,
but asked it be ticketed rather than forgotten. **This is that ticket.** Extract one shared
`writeLinkRepairSkipLog` on its own merits, not as a passenger on a content fix.
