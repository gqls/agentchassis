# NOTES — directory-build-handler (`bugs_open/206`)

## 2026-08-06/07 — diagnosed, designed, implemented, submitted; not yet live

Picked up after the owner asked (via `features_open/021`'s chat) to build
`vetcomparison.uk`'s practice-directory page through the framework, and to
file whatever the framework can't do as a bug rather than hand-build it.

**Diagnosis (`bugs_open/206`, filed first, before any code):** read
`load_work_item_actions.go` at HEAD and found `entity-directory` explicitly
named in `unavailableBuilders` — a builder the original authors reserved the
name for and never built. Checked the closest prior claim on record
(`vetcomparison/PLAN_2026-07-26_site_strength.md`'s "the entity-page
machinery is proven live") directly against the DB and falsified it: the two
cited sites' `entity-directory`/`entity-page` pages use unrelated generic
components (`archetype-grid`, `content-block-about`), and
`p.sections @> '"directory-listing"'::jsonb` returns 0 rows fleet-wide.
Corrected that claim in place in the PLAN file rather than silently.

**Design.** Read `FOCUS_directory_builder_and_list_components.md` in full —
the `query.*` resolution mechanism (`queryresolve` package) is already
proven live (`guide-list`/`tool-list` deployed on 4 real sites: found by
direct query, not by trusting `usage_count`, which turned out to be a dead
counter — 0 on every listing component including ones confirmed live).
So the actual gap was narrow: one new resolver (`business_directory.go`)
reading a site's own `directory-export-json` config to query
`business_intel.businesses`, plus one new action
(`ensure_page_section_layout`) to fill a genuinely-empty page's plan, plus a
thin new agent (`directory-build-handler`) chaining the two into the
EXISTING generic `page-build-handler`. No new content-writing logic
anywhere. Full reasoning and what was deliberately left out (client-side
search across all 2,337 businesses; `entity-page`/practice pages, still on
hold pending P1's crawl) is in `PLAN_2026-08-06_directory_build_handler.md`.

**Implementation.** All Go code built and unit-tested clean against
`git archive HEAD` + the diff (this repo's own shared-tree discipline — the
working tree had an unrelated broken file, `component_write_guard.go`, mid-edit
by another session at the time). Two migrations written
(`325_directory_listing_binds_to_business_directory_query.sql`,
`326_directory_build_handler_agent.sql`) following the
`scripts/migration/run-migrations.sh` convention — **NOT yet applied**, see
HANDOFF.

**A same-file passenger fired, exactly as documented.** My uncommitted edit
to `load_work_item_actions.go` (moving `entity-directory` from
`unavailableBuilders` to `availableBuilders`) got swept into `cb7b4d759`
(`fix(208)`, an unrelated owned-page fix that happened to touch the same
file) before I could commit it myself. Confirmed via
`git log -S "directory-build-handler ensures the page's plan"`. Functionally
harmless — forward-only holds, nothing was lost — but it means that one hunk
reads as part of the `208` commit in `git blame`, not this lane's. Noted in
the eventual commit message and the concept register entry so the paper
trail is honest about it.

**Committed** `f750595dd` (everything else: the two new Go files, the
`defaultSectionsForPage`/`registry.go`/`queryresolve.go` diffs, both
migrations, the PLAN doc, concept register BLD-017). Submitted to the
council gate first — `SUBMISSION_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`
— and committed with `Council-Submitted:` rather than waiting, per this
repo's own timing norm. Verdict not read yet as of this note — **read it
before treating this as approved**, and if REVISE/REJECTED, this is already
on the shared branch and needs a follow-up commit, not a hope it goes away.

**Not done this session, and why:** did not build the image, did not roll,
did not re-triage the two named `site_work_items` rows
(`715ec305` directory-index, `2f50bfda` guides-index), did not fire anything
against `vetcomparison.uk`. All deliberate — the session was long enough
that pushing through image build (several minutes) + the council's own
queue latency (this repo's own documented ~30 min, not ~2) + roll + pod
verification + live re-triage + end-to-end page verification risked doing
that work carelessly rather than with the same rigour as everything above.
See `HANDOFF_2026-08-07_continue_here.md` for the exact remaining steps.
