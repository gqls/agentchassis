# HANDOFF 2026-09-04 — bugs_open/114: the bug changed shape. Read this before the 09-03 one.

**Supersedes `HANDOFF_2026-09-03_continue_here.md`.** That handoff said 114 was one observation
from closable and the remaining work was arming a built mechanism. **Both halves are now wrong.**
Read this, then the 2026-09-04 entries at the bottom of `NOTES_imagery_wiring.md`, then the
`PLAN_2026-08-22…md` addendum dated 2026-09-04.

## ⚠ THREE THINGS THAT WILL MISLEAD YOU IF YOU READ THE OLD HANDOFF FIRST

1. **DO NOT ARM MIGRATION 710.** The owner's 2026-09-04 words — *"let's use the hero images
   somehow, we don't need a stop gap though"* — answered a binary (hand-wire four pages vs arm
   the built mechanism). He rejected the stop-gap and asked for the OUTCOME. **He has never seen
   710, the council REVISE (`bd78490d`), or the key name** `[SOURCED from the finetuning lane,
   not from the owner]`. 710 stays **HELD on a live objection**, not overdue. A roll is exactly
   when someone notices a held migration whose carrying image has shipped and assumes its moment
   has come; the `inter thread comms` lane is propagating that warning fleet-wide.
2. **`unwired` is NOT "pages showing nothing".** IMG-077's `referenced` test asks only whether
   the page renders *the content-hero path*, never whether it renders *some* hero.
   `leopardessconsulting.co.uk`'s 2 `unwired` pages **serve their planner hero at 200** and hold
   a redundant SECOND `content_hero_*` asset — duplicate generation, not undelivered imagery.
3. **The closure gate in the 09-03 handoff is satisfied and it no longer means what it meant.**
   `unwired` has been observed on 5 sites (websitepromotion 1, loanzy 3, finetuning 4,
   leopardess 2, fundamentallyai 3); 15+ sites swept, 19 rollups. The detector works. **But
   closing 114 on that basis would close it on the delivery reading, which is not the bug.**

## State in one paragraph

114 is **an `on_missing` CONTRACT VIOLATION**, not a hero-delivery gap. `[MEASURED 2026-09-04,
null-safe]` all **874** deployed rows declaring an image-typed `site_assets.*` field declare
`on_missing: use_fallback` with a non-null fallback — total, no policy mix. **123 rows across 32
sites hold no value**, which the contract forbids. **86 paint the site-wide brand hero or the
legacy literal, 37 paint nothing.** The producer is `save_page_sections_action.go:938`'s
page-wide DELETE plus the wholesale INSERT at ~1130 (**122 of 123 rows sit inside that DELETE's
lock-only predicate**), which is why migration 664's repair decayed **9→3 in eight days**. A
three-phase fix is at the council: **`74ffbb5b-609c-4949-a4ba-5142140a71d3`**.

## What to do next, in order

1. ~~Read the council verdict before writing any Go.~~ **DONE — APPROVED**, corr
   `74ffbb5b-609c-4949-a4ba-5142140a71d3`, 15 reviewers, `unreadable` 0,
   `gated_by_truncation` false. **But SEVEN objections are accepted and change the design, so
   the approved plan is NOT the plan to implement** — read the adjudication in NOTES
   (2026-09-04, "Council verdict") before writing anything. The one that matters most:
   **my own fallback branch was a SILENT FILL** — it wrote the generic value with no work item
   and nothing distinguishing "resolved cleanly" from "papered over", which is this bug's whole
   history repeated one level along. Fix that before implementing, not after.
   Also accepted: extract `carryStored`'s precedence into ONE shared helper rather than a second
   implementation; check the two reads already on the destructive path before adding a third;
   put the read and the destructive write in one transaction; state the seal's known narrowness
   in the code header; and carry the phase split in `doc_plans`/`doc_notes`, not only in the
   submission JSON.
2. **TRACE ONE LIFECYCLE FIRST — this is the only unproven link.** Nobody has shown a single
   `page_id`'s full `page_component_history` going from an unresolved BIRTH row to the
   `save_page_sections_overwrite` that added the value. The census, the lock arithmetic and the
   `source` tag make the mechanism the strong candidate; that trace makes it proven.
   ⚠ **The tag is in `source`, NOT `application_name`** — that column is NULL on these rows, and
   this lane's own lineage carried the wrong one into a `090` symptom.
3. **Make the create-only vs rebuild split** before writing the guard. Several wholesale INSERTs
   write a row's FIRST content (`create_tool_component`, `deploy_tool`, probably `adopt_verbatim`
   and `create_report_page`), where wholesale is harmless. The in-repo precedent is
   `page_component_writer_coverage_test.go`'s `exemptWriters` map — reason per entry, plus a
   second test that fails when an exemption goes stale. **Reuse it; do not invent one.**
4. **Reuse `datahelpers.SchemaContentFields`** to read the schema — it is what
   `missingRequiredLLMFields` uses. Re-parsing `input_schema` will draw the reuse seat.
5. **Phase 2 and Phase 3 are NOT in the submission.** Phase 2 widens
   `check_required_fields_missing` (its two exclusions — non-`required` and image-typed — ARE
   this blind spot, deferred 2026-08-11 on unmeasured volume; **the volume is now 123**).
   Phase 3 is the table-level default, **architecture-scope**, and goes to
   `architecture_review/` as an RFC.

## Why Phase 3 is not optional, which is the afternoon's most consequential finding

`[editorial_design_uplift, 2026-09-04]` **~25 SQL migrations under `sql_for_agents/` write
`page_components.content_data`**, and no application-layer guard can constrain a `psql` session.
Most are additive (`jsonb_set`: 664, 230, 229, 287) or key-preserving (`regexp_replace`: 231,
232); `043_section_editor.sql:330` is wholesale (`content_data = NULL` on one hero row by uuid —
**checked: that uuid no longer exists in `page_components`, so it accounts for none of the 123**).
**A hand-repair IS a migration.** So the writer class behind the measured decay lives where
Phases 1–2 cannot reach by construction, and only the table boundary closes the door.

Wholesale writers outside `platform/orchestration/actions/`:
`internal/core-manager/admin/page_admin_handlers.go:343` (admin API, dynamic `setClauses` — and
it **locks on edit**, which is why exactly 1 of 123 is locked), `cmd/webdesignport/import.go`,
`cmd/content-data-recover/sql.go`. ⚠ **Census classified by SQL FORM, not by traced call path —
a stated gap, not a proof.**

## Traps this session hit — all four were instruments that could not come out otherwise

1. **A COMPOSED page URL.** `/can-you-trust-ai-with-your-data` → 404; the recorded `pages.url` is
   `/blog/can-you-trust-ai-with-your-data.html` → **200**. Curl the RECORDED url.
2. **`rendered_html LIKE '%background-image%'`** — matches every component's own `<style>` block.
3. **`position('background-image') > position('</style>')`** — `position()` returns the FIRST
   occurrence, always the style block. **Its control inverted** (708 of 753 rows holding a value
   read as "empty") and that is the only reason it was caught.
4. **`NOT (content_data ? field)`** — `jsonb ?` propagates **NULL**, so 4 NULL-`content_data` rows
   vanished from every count. Published **121** to three lanes before reconciling to **123**.
   Use `NULLIF(col->>'k','') IS NULL`, which is total.

**The working instrument, and validate it before trusting it:** does `rendered_html` contain the
stored value **verbatim**? Control arm **752 of 753**. And **grep LANDMINES by TOKEN, never by
sentence** — the files are hard-wrapped and a phrase grep returns a false absence (cost me a
"not documented" reading on an entry that existed).

## Blast radius, enumerated `[MEASURED 2026-09-04]` — and the consumers have been told

`save_page_sections` is invoked by **three** live pipelines, one step each:
**`page-build-handler`**, **`page-rerender`** (the fleet's highest-volume page path) and
**`tool-recreation-handler`**. `seal_declared_field_contract` appears in **0** live
`agent_definitions` rows against an existing surface of 21 config keys on those steps, so
declaring it cannot retroactively arm the unknown-key detector. Per the owner's 2026-07-29 §3
ruling the three consumers were **told, not merely counted** — broadcast requested through the
`inter thread comms` lane, stating what changes for a consumer: nothing until armed, and once
armed a pipeline **relying on a wholesale rebuild to CLEAR a stale `content_data` key would lose
that clearing.** If a lane depends on that, it must be heard before the key is armed anywhere.

## Cross-lane state

- **`finetuning`** — handed this over; holds finetuning.uk's 4 unwired pages as a **diagnostic
  witness only** (EXCLUDED as acceptance evidence either way, 412 §11). Undertook not to touch
  its hero components' `content_data` while this is in flight. Will probe at the served page when
  there is something to measure. Their human-reported *"case studies page is missing a hero"* is
  the one datum showing a visitor can see this.
- **`editorial_design_uplift`** — contributed the producer candidate and both census passes.
  Owns none of the 7 components (they are shared library rows — I wrongly implied otherwise and
  corrected it).
- **`475`** — clean edges (`platform/delivery/`); gave the two round-warnings that shaped the
  submission.
- **`imagery`** — different lane (`docs024_key_docs_latest/imagery/`), doc-only footprint.
- ⚠ **HEAD is RED and it is not this lane's:** two tests in `platform/orchestration/actions`
  from `fail_work_item_message_template.go` (`83407cd37`, the 440 lane). **Establish the
  with/without control before reading any red as yours.**
