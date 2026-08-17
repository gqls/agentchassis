# RFC_033 — "the page's section list" is re-derived by eight readers and only one of them knows about locks; should it be one entrypoint?

**Status: RULED 2026-08-17 (owner) — OPTION 2, BUILT.** Owner's words, on being shown the three
options and this lane's recommendation: *"go with your recommendation"*. So: the source-scan
lockstep, **not** a single mandatory entrypoint. Live as
`platform/orchestration/actions/section_list_reader_coverage_test.go` (council `02cb2134`).
**Two corrections this file owes its own population section**, both found by running the
detector the ruling asked for rather than re-reading the register: the readers of
`site_plan_sections` number **seven, not the eight this title claims** (the eight came from the
register's named list, which mixes plan readers with cache readers) — and one of the six,
`discovery_checks/check_sectionless_pages.go`, **was missed by the hand census in §C below**,
because that census grepped `FROM` and the file uses `JOIN`. The test failed on its first run
naming exactly that file, which is the shortest possible argument for building it.

*(Original status, retained: OPEN — filed 2026-08-16 by the `bugfix_285_lock_blind_section_list` lane, at the
council architecture seat's direction.** The seat raised `ARCHITECTURE_SIGNAL: needs_rfc`
(medium) on the 285 fix's round-2 submission (corr `79f70435`, APPROVED 15:31Z) and explicitly
recommended *"proceeding but opening the RFC in parallel rather than gating this round"*. The
fix shipped; this is the parallel half. **Nothing here asks for the 285 fix to change.**)*

## The seat's objection, verbatim

> `MergeLockedPageSlots` + `LockedPageSlotsSQL` become a cross-package contract (loader + drift
> check + future readers named in the header) with no single enforced call path — a future
> section-list consumer can still be written lock-blind exactly as the original bug arose. Worth
> an RFC to decide whether this becomes a single mandatory entrypoint rather than a convention.

and, in its notes:

> disciplined accretion of a load-bearing cross-package contract is still accretion — an RFC
> would let you decide once whether "section list" should be a single function all readers call,
> rather than a merge convention three files must independently remember.

**The uncomfortable part, which is the reason to take it seriously: this is the shape of the bug
it is objecting to.** `bugs_open/285` (the section-list case) happened because one reader of "the
page's section list" did not know a rule every other part of the system assumed. The fix taught
that reader the rule. It did not stop the next reader being written without it — the only thing
standing between us and a second instance is a file header, a register entry and this document.

## The population (measured 2026-08-16, not asserted)

**A. The one assembler that now merges locked rows.**
`LoadPageSectionsFromSpecAction` (`platform/orchestration/actions/load_page_sections_from_spec_action.go`)
— four plan tiers, then `datahelpers.MergeLockedPageSlots`, then one jsonb-compared write to the
`pages.sections` cache. Live consumer census: **`page-build-handler.load_spec_sections` is the
only agent-definition step that calls the action.**

**B. The one other reader taught the same rule.**
`discovery_checks/check_section_source_drift.go` — compares plan-vs-cache through the same merge,
because without it the fix's own day-one effect would have been 13 false drift items.

**C. The readers that re-implement the tier order and stay lock-blind by design (7).**
`page_section_satisfiability.go`, `ensure_page_section_layout_action.go`,
`datahelpers/plan_section_counts.go`, `tool_content_item.go`, `flag_page_image_rebuild_action.go`,
`discovery_checks/check_sectionless_pages.go`, `discovery_checks/check_literal_markdown.go`.
Each carries its own copy of "site_plan_sections, else the aspect, else pages.sections" — several
say so in a comment that names the loader as the original. For each of them, "this page has no
sections" and "this page has only locked sections" are today the same answer.

**D. The writers of the same state (`pages.sections`).**
`apply_gap_plan_action.go` (×3, incl. its INSERT), `ensure_page_section_layout_action.go`, the
loader itself, and seven INSERT-a-new-page paths (`page_role_upsert`, `site_db_actions`,
`create_blog_posts`, `adopt_verbatim`, `apply_adoption_plan`, `create_tool_component`,
`cmd/webdesignport`). The INSERT paths are **benign by construction** — a page being created has
no `page_components` rows, so there is no locked row to omit. The other two are transient: the
next build re-merges. "Transient" is doing real work in that sentence, and it is only true while
every save path goes through the loader (see E).

**E. The save paths, which is where a lock-blind list becomes damage.**
Six agents call `save_page_sections`. Over 30 days only **two ran at all**: `page-build-handler`
(98 runs, all through the loader) and `page-rerender` (625 runs, none). `page-rerender` **cannot**
reproduce the class — `rerender_page_sections` composes its proposal from the stored
`page_components` rows (`loadStoredSections` → `carryStoredSection`), so a locked row is in its
proposal by construction. That asymmetry is visible in the data: the fleet's `lock_blocked_change`
population is **38 `overwrite` and 7 `remove`**. `page-rebuild`, `pageflow-builder`,
`site-work-orchestrator`, `tool-recreation-handler`: **0 runs in 30 days** — and `page-rebuild` is
the one that would matter if it woke, because it plans from `current_page.sections` with no loader
in its chain.

**So the live blast radius today is one path, and the estate is one dormant agent away from two.**

## The question for the owner

Should "what sections does this page have?" become **one function every reader must call**, or
stay **a convention that eight places may each re-derive**?

Three shapes, costed:

1. **Convention + register entry (today).** Cost: nothing new. Risk: exactly the objection — the
   ninth reader is written lock-blind, and nothing fails when it is. Note this is *weaker* than it
   sounds, because the seven C-readers are not merely unenforced, they are **deliberately** blind
   and correct to be so in at least one case (`ensure_page_section_layout` fires only when every
   source is empty, which is the one case the loader also declines to merge).
2. **A source-scan lockstep test** (the estate has this pattern already — `asset_lock_guard`'s
   `sendGitCommitRequest` file-set test, and the `html_template` writer census from the OTHER 285).
   A new file that reads `site_plan_sections`/`pages.sections` fails the build until it is
   declared either a merge-caller or deliberately-blind-with-a-reason. Cost: one test, one
   allow-list with reasons. Buys: the ninth reader is a build failure instead of a bug, and the
   deliberate blindness of C becomes *stated* rather than *inherited*. **This is the cheapest
   thing that would have caught the original defect.**
3. **One mandatory entrypoint** — `datahelpers.PageSectionList(ctx, db, siteID, pageName)`
   returning the merged list plus its source label, with every C-reader migrated onto it. Cost:
   eight call sites, several of which want subtly different answers (counts vs names vs
   "is it empty"), and one of which is a discovery check that must NOT merge. Buys: the concept
   has an owner. Risk: a single function with eight callers and three legitimate variants is how
   a parameter-boolean grows.

**This lane's recommendation: (2), and not (3) yet.** The measured population says the divergence
that actually bit us was a *writer of the list* (one call site), while the seven re-derivers are
readers whose blindness is currently harmless and in one case correct. A lockstep test makes the
next divergence loud without pretending eight readers want one answer. Revisit (3) if `page-rebuild`
wakes, or if a second *writer* of a saved section list appears — either event moves a C-reader
into the E column, which is the transition that turns a convention into a defect.

## The tripwire, stated as a rule, in force meanwhile

**A new reader or writer of a page's section list must state, in its own file, whether it honours
locked rows — and if it does not, why that is safe for what it answers.** "It follows the loader's
tier order" is not a statement about locks; the loader's tier order is exactly what was lock-blind
until 2026-08-16. If you cannot say why blindness is safe, call `datahelpers.LoadLockedPageSlots`
+ `MergeLockedPageSlots`.

## Relations

- `bugs_open/285_HANDOFF_2026-08-15_section_list_assembly_is_lock_blind_…` — the defect; its
  "Implementation" and "LIVE" sections carry the code and the artefact evidence.
- Concept register **LOCK-008** (`register/locks.md`) — the mechanism, its landmines, the three
  advisory residuals from the approving round, and the same census as §D/§E here.
- `bugs_closed/058` — the ROW half (the write guard). This RFC is about the LIST half's *reach*.
- **RFC_027 / RFC_028** — the same shape one layer down ("the symbol-handle grammar has no owner
  and this is its fourth bug"; "the input-resolver precedence chain has no owner"). If those are
  ruled toward single owners, the argument for (3) here strengthens by precedent.
- `031_locks.md` L156 ("don't merge locks across multiple tables in a downstream reader") — the
  rule the 285 fix had to argue past, and the reason option (3) needs care: a single entrypoint
  that answers "is this locked" for everyone is precisely what that line warns against; the merge
  is defensible because it answers "what sections exist".
