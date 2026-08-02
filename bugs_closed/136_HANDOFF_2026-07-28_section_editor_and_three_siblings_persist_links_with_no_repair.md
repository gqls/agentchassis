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

---

## Update 2026-08-02 — candidates 1 and 2 are BUILT and council-APPROVED at round 1. Still OPEN: inert until the next chassis roll.

**Taken by:** the `bugfix_136_sibling_link_repair` lane. Standing five in
`docs/agent_docs/docs024_key_docs_latest/bugfix_136_sibling_link_repair/`.
**Commit:** `66998d300` · **Council:** `0275f9c2-035f-4c9e-8a50-83836dfeffd9` —
**APPROVED at round 1**, 5 advisory objections, none high (`decided_by`: *"approved with
5 advisory objection(s) — none high-severity"*). Registered as **LNK-027**.

**Do not close this on the commit.** The bar is fixed AND live; Go is inert until an image
is rebuilt and rolled. The pre-roll pod-grep baseline is already measured, so the
after-check discriminates — see § Post-roll verification below.

### What was built

- **Candidate 1, `ApplySectionEditAction`** — and the file's framing was one level short.
  This action had **two** persist sites: `content_edit` returned its HTML for the caller to
  write, while `component_swap` wrote its own row *inside* `applyComponentSwap`. A repair
  before either would have been bypassed by the other — this bug's own shape, one level in.
  The swap's `UPDATE` moved to the caller; both branches now return a `sectionEditOutcome`
  and the action repairs once, then persists once. `errComponentLocked` semantics unchanged.
- **Candidate 2, `CreateReportPageAction`** — confirmed a real member: `renderReportSection`
  escapes ~25 deterministic fields with `html.EscapeString`, which reads as total, then
  embeds four LLM-authored prose fields with a bare `%s` (`create_report_page_action.go`,
  the `report-prose` loop). **Inert today: `SELECT count(*) FROM pages WHERE
  page_type='report'` → 0.**
- **The seam** — `repairComponentHTMLBeforePersist` in a new `component_link_repair.go`,
  a WRAPPER over `repairSectionLinks` with a one-element slice, exactly as this file
  proposed. One definition of the repair semantics, one test file.
- **The tail ticket, discharged** — `writeLinkRepairSkipLog` extracted. This change would
  have written the THIRD copy of that INSERT, which is when the `reuse_agent` seat's
  accepted deferral stops being defensible. Rows unchanged: each caller passes its own
  `agent_type`/`action`/message.
- **A pattern-check rule**, `check_unrepaired_component_write` — because the seat's
  objection was about the WRITER SET, not one writer, and the set moved under us:
  **`adopt_verbatim.go` became a writer of that column on 2026-07-30 (`e6a8bb63b`), two
  days after this bug was filed**, and no reader of this file would have known.

### Corrections to this file, from doing the work

1. **`RebuildBlogListingAction` is NOT a member of the class.** This file rates it "lower
   risk but not zero". It is lower than that, and the difference is measurable: its only
   anchor comes from the template it renders, and the live `content-listing` template
   carries **exactly one**, `href="{{.url}}"`, filled from `pages.url` — the same table the
   repair index is built from, under a *looser* predicate. Repair there could only ever
   no-op. Allow-listed with that evidence rather than guarded.
2. **The `[UNMEASURED]` volume is now measured** — and *this file's instinct to refuse a
   number until it was measured was right*. Fleet-wide, stored `page_components.rendered_html`
   holds **35 unresolved internal href occurrences** (18 rewritable, 17 unlinkable) in 13
   components on 13 pages across **6** sites. Query in the lane's RUNBOOK §2.
   ⚠ **It is [NOT ATTRIBUTABLE] to a writer** — a stored href does not record who wrote it,
   and 079's fix has been live since v1.0.1170. ⚠ Commit `66998d300`'s message quotes an
   **earlier, wrong** set of figures ("30 … 7 sites … 14/16"): those were read off a
   `GROUP BY` listing's row count instead of an aggregate. Logged in `WRONG_CALLS.md`.
3. **Neither guarded path has run in the retained window.** `orchestration_states` (2,469
   rows from 2026-07-13) holds **zero** runs owned by `section-editor` or `tool-improver`.
   So this is prevention on live, documented, reachable paths — not a bleed being stopped.
   Said plainly because the fix is easy to oversell.

### The council's five advisory objections, and what was done about each

| seat | severity | objection | disposition |
|---|---|---|---|
| `guardian` + `prior_art_librarian` | medium | "grep confirms one call site" for `applyComponentSwap` was asserted, not shown — if a second caller exists the swap's write silently disappears | **ANSWERED with the lookup:** `grep -rn "applyComponentSwap" --include=*.go .` → one call (`section_editor_actions.go:331`), one definition, five log/comment strings. Recorded in the lane's NOTES |
| `editquality` | medium | `adopt_verbatim.go` allow-listed as byte-preserving on an *unmeasured* assumption, while a landmine on that file warns `--fidelity high` is not a milder `locked` | **ANSWERED and the allow-list reason rewritten with citations:** it writes `content.RawHTML` verbatim (`:514`, `:533`) and stores `sha256(RawHTML)` in `content_data` (`:487`), so repair would invalidate the hash it exists to keep; it is reachable only under a strict binary (`apply_adoption_plan_action.go:426`). The cited landmine is about which PATH runs, not what this file writes |
| `guidelines` | soft | the config-lever test may prove inertness via `ExpectationsWereMet()` with nothing registered — the vacuous-negative pattern | **IT WAS EXACTLY THAT, AND IT IS FIXED.** With nothing registered that call returns nil unconditionally, and the HTML assert could not catch it either because the fail-open path also returns the input unchanged — it would have passed with the lever deleted. Both "negative" tests now REGISTER the call they must not see and require it to go unmatched. Mutation-checked: deleting the lever now fails |
| `bug_historian` + 3 others | medium/low | the enforcement is advisory-only, so the writer set stays open-ended; candidate 3 "should be tracked as a real ticket, not left implicit in a lint rule's silence" | **FILED as `architecture_review/RFC_008_a_mandatory_write_seam_for_page_components_rendered_html.md`**, with the argument against it stated fairly and the measurement that would settle it (does the advisory channel actually get read?) |
| `debug_historian` | medium | the orphaned-prose landmine (unlink → "text that should be a link") is carried to two new call sites without confirming downstream coverage | **ACKNOWLEDGED, not closed.** No downstream gate scans these two paths for that artefact: `bugs_open/116` records that the phantom-link discovery check has never run on any site. The repair's direction is unchanged from the three live callers, so this widens an existing gap rather than opening one — but it is a real residual and belongs to 116, not here |
| `compliance` | flagged, not an objection | the raw unescaped LLM prose in `create_report_page` is "XSS/injection-adjacent … worth another seat's eyes if reports go live" | **PASSED ON, not swallowed.** Out of this bug's scope and out of that seat's mandate. There is a partial guard already (the action refuses prose containing a `<script` tag). Whoever takes report pages live should treat this as a prerequisite |

### What is STILL OPEN on this file

1. **The roll, then the post-roll verification.** Until then the defect is reproducible.
2. **The tool-markup writers — `create_tool_component_action.go`, `deploy_tool_action.go`.**
   Genuine members: **7 of the 35** live unresolved hrefs sit in tool-shaped slots
   (`slot_name LIKE 'tool%'` or `page_type='tool'`), 5 components on 3 sites. Left out of
   this change on **collision** grounds, not merit — those files are in play across the tool
   lanes (146/149/154/126) and a pathspec commit still takes a same-file passenger.
   **They are deliberately NOT allow-listed in the pattern check, so it will keep firing on
   them.** Adding them to silence it would convert a live debt into a false all-clear.
3. **Candidate 3** — now RFC 008, not a line in this file.
4. **The standing stock is not cleared.** 17 unlinkable hrefs are live 404s on real pages
   today; this change stops the class recurring through two writers, it repairs nothing
   already stored. Detection belongs to `bugs_open/116`.

### Post-roll verification (the baseline is already taken, so this discriminates)

Measured on `agent-chassis-f8d46bd4c-6rtlj` **before** the build:

| grep | before | required after |
|---|---|---|
| `repaired dead internal links before persisting a single component` (new) | **0** | ≥1 |
| `failed to update page_component for swap` (**negative control** — this change removes it) | **1** | **0** |
| `SavePageSectionsAction: repaired dead internal links before persist` (079, live) | **1** | 1 |

Run all three in one exec, on **every** replica. The obvious marker "repaired dead internal
links" is **vacuous** — 079's live string contains it and greps 1 before anything ships.
Then re-run the RUNBOOK §2 census and confirm the stock is not growing.

---

## CLOSED 2026-08-02 — LIVE on v1.0.1229, pod-verified on both replicas with a negative control

**Bar met: fixed AND live.** Moved to `bugs_closed/`.

### The proof, one exec per replica, both replicas

| grep | before (v1.0.1228, taken 2026-08-02 ~11:00) | after (v1.0.1229) |
|---|---|---|
| `repaired dead internal links before persisting a single component` (new) | 0 | **1** |
| `Component link repair SKIPPED` (new) | 0 | **1** |
| `failed to update page_component for swap` (**negative control** — removed by this change) | 1 | **0** |
| `failed to persist section edit` (the replacement wording) | 0 | **1** |
| `SavePageSectionsAction: repaired dead internal links before persist` (079, **positive control**) | 1 | **1** |

Pods `agent-chassis-79479769b9-g7fbt` and `-n8nbj`, identical on both.

**The negative control is what makes this evidence rather than a hopeful grep.** A new
string appearing proves only that *something* shipped; a string this change DELETED going
1 → 0, under the identical pattern that returned 1 an hour earlier, proves the binary is
this code. `bugs_open/153` exists because a roll is not evidence.

### The live-path regression check, which the pod-grep cannot give you

This change edited `rerender_link_repair.go` — a live, proven path — to route its skip
record through the extracted `writeLinkRepairSkipLog`. Post-roll:

```sql
SELECT action, count(*) FILTER (WHERE occurred_at > '2026-08-02 10:30:00+00') AS since_roll,
       max(occurred_at) FROM agent_error_log WHERE error_code LIKE 'CONTENT_LINK_REPAIR%' GROUP BY 1;
-- rerender_page | 21 | 2026-08-02 14:25:07+00
```

**21 rows since the roll, `agent_type='page-rerender'`, `step_name='render_page'`, action
`rerender_page`** — correctly shaped, no blank or `unknown` labels. The rerender seam is
running on the new binary and still writing the rows every query downstream expects.

⚠ **Stated rather than glossed: the extracted `writeLinkRepairSkipLog` itself is NOT
exercised in production.** Those 21 rows are `CONTENT_LINK_REPAIR_DETAIL`, written by
`writeLinkRepairLog`, which this change did not touch. `CONTENT_LINK_REPAIR_SKIPPED` has
**zero** rows in the retained window — it only fires when the page-index read fails, which
has not happened. The extraction is covered by unit tests and by the fact that its three
callers compile against it; it is not covered by live traffic. A future thread should not
read "the refactor is proven" as covering that branch.

### What the newly-guarded call sites have done since going live: nothing, as predicted

Zero `CONTENT_LINK_REPAIR_*` rows for `apply_section_edit` or `create_report_page`. Both
paths remain dormant (still 0 `pages` with `page_type='report'`, still no `section-editor`
or `tool-improver` orchestrations). **This closed as prevention, not as a bleed stopped**,
and the file should be read that way.

### One thing the re-run of the census turned up, and it became its own bug

The census went 35 → 36 between the morning and the roll. The new row was
`vetcomparison.uk` / `tool-cma-obligation-checker`, `href="' + q.link + '"` — **not an
anchor at all**, but a JavaScript string concatenation that BUILDS one at runtime. Running
the shipping `RepairPageLinks` over those exact bytes deletes the anchor from the JS and
returns still-valid JavaScript:

```
IN : ' <a href="' + q.link + '" target="_blank" rel="noopener">See guide section</a>.</p>'
OUT: ' See guide section.</p>'
```

Filed as **`bugs_open/180`**. It is a defect in the shared repair, not in this fix — but it
matters to this file's own "next candidate", because the tool-markup writers are exactly
where JS-built anchors live. **Wiring the seam into `create_tool_component_action.go` /
`deploy_tool_action.go` must wait for 180, or it will delete working links from tools.**
That is now the ranked order: 180 first, then the tool writers.

Two consequences for the census figure itself, both corrections to my own earlier claim:
- the "35" includes at least one non-anchor, so it is an **upper bound** on real dead links,
  not a count of them;
- and the query's `href="..."` extraction cannot tell markup from a string literal, which is
  the same class of error as the repair's own regex.

### Still owed, and now living elsewhere

1. **`bugs_open/180`** — the JS-anchor corruption, and a prerequisite for item 2.
2. **The tool-markup writers** — 7 of the 35 hrefs sit in tool-shaped slots. Deliberately
   NOT allow-listed in `check_unrepaired_component_write`, so it keeps firing on them.
3. **`architecture_review/RFC_008`** — candidate 3, the mandatory write seam.
4. **The standing stock** — 18 unlinkable hrefs are still live 404s; detection is
   `bugs_open/116`'s, which has never run on any site.
