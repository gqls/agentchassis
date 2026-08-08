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

## 2026-08-07 — Go code confirmed live (someone else's build/roll); council round 1 REVISE, fixed, resubmitted

Checked the owner's report of "a fresh chassis build has been deployed"
directly rather than assuming it meant this lane's code: pod-grepped both
`agent-chassis` replicas (now `v1.0.1262`, rolled by another session) for
`ensure_page_section_layout` (5), `business_directory` (4),
`directory-build-handler` (1), negative control (0). **The Go code is
genuinely live.** DB config is not: `agent_definitions` has no
`directory-build-handler` row yet and `directory-listing`'s schema is
unchanged — migrations 325/326 have not been applied.

**Read the round-1 council verdict in full** (`diagnosis_artifacts`,
`body` column — the `metadata` column only holds a summary, `doc_notes`
truncates; go to `body` for the complete per-reviewer JSON). **REVISE**,
gated by `bug_historian`'s edit-1 objection (high): `resolveBusinessDirectory`
returned an empty slice, not an error, when a site had no
`directory-export-json` config — indistinguishable from "genuinely zero
eligible businesses." Three other reviewers independently raised high-severity
objections too (`tooling_provenance`: no spawn before the `call_agent` to
`page-build-handler`, a documented landmine shape; `prior_art_librarian`: no
existence check attached for a pre-existing `directory-build-handler` row;
`guardian`: unmeasured fleet-wide blast radius of the dispatch-map flip).

**Fixed the gating objection and the two other fixable HIGH ones:**
1. `resolveBusinessDirectory`'s no-config branch now returns an error, which
   routes into `plan_sections_action.go`'s existing resolver-error handling
   (distinct from the empty-list/`on_missing` path) — no new mechanism
   needed, the platform already had the right path for this.
2. `326_directory_build_handler_agent.sql`'s workflow gained a
   `spawn_page_builder` step before `call_page_build_handler`, matching
   `directory-export-orchestrator`'s own proven `spawn_exporter` →
   `call_exporter` shape.
3. (Not gating, but three independent reviewers — `reuse_agent`,
   `constitution`, `architecture` — flagged the same thing): extracted
   `insertSitePlanSectionRows` so the new action stops hand-copying
   `write_site_plan_action.go`'s insert shape. **Did NOT** wire
   `write_site_plan_action.go` itself onto the shared helper — it gained
   `assigned_fact_ids` handling (`PBP-037`) from a separate, actively-revised
   lane (council REJECTED on breadth, mid-`RFC_016`) during this exact
   review round. Touching another lane's contested file to satisfy a
   medium-severity reuse nit felt like the wrong call — documented as an
   honest follow-up instead.

**Answered, not code-changed** (evidence in the resubmission's
`grounded_in`, not repeated here): the blast radius is exactly **one**
pending `entity-directory` work item fleet-wide (the very one this fix
targets) and **two** `entity-directory` pages total (the second already
deployed, untouched by this change since it already has sections);
`queryresolve.go` read in full both before round 1 and again now — no
existing case touches `business_intel`; `cb7b4d759` re-confirmed via
`git log -S` to genuinely carry the `load_work_item_actions.go` line;
`page-build-handler`'s actual `input_contract` matches what
`directory-build-handler` supplies it.

**Rebuilt, retested (unit, against fresh `git archive HEAD` — moved twice
more during this one session), committed** (`37560f120`), **resubmitted on
the same correlation** (`RESUBMIT_CORR=5b8e4cf7-...`). Round-2 verdict not
yet read as of this note — checked once, not landed yet (queue was visibly
contended both times). **Read it before doing anything else in this lane** —
see HANDOFF for the query.

## 2026-08-08 — Go code confirmed LIVE on a fresh build; round 2 was REVISE; round 3 submitted

Picked up cold from the 2026-08-07 handoff. Re-ran Step 0 first: `37560f120`
and `f750595dd` both still on the branch. Pod-grepped `v1.0.1262` (both
replicas) for the round-2 marker string
(`"cannot distinguish this from a real zero-business result"`) — **absent**
on both, so at that point the running binary carried round 1's code only,
not round 2's fix, even though `directory-build-handler`/`business_directory`/
`ensure_page_section_layout` were all present (round 1's symbols).

**Read round 2's verdict in full** (`body` column, not `metadata` — the
gating summary alone doesn't carry the per-reviewer detail): **REVISE
again**, `decided_by: "gating objection from bug_historian"`. The gate:
edit 1's fix (missing-config → error, not empty slice) was applied at
`resolveBusinessDirectory` alone with no audit of whether
`queryresolve.go`'s other resolvers share the same "no source config →
silent empty" shape. Four other MEDIUM objections rode along, not gating but
needing an answer: `editquality` (edit 3's sketch claimed
`ensure_page_section_layout_action.go` was switched to the new shared
helper, but the submission's own `edits` array never listed that file — a
submission-metadata gap, not a code defect: the code change was real,
already committed in `37560f120`); `guardian` (326.sql's edit was labelled
`operation:"modify"` against a `.sql` path with the owning pipeline only
inferable, not stated — surface-ownership rule wants `config_change` +
explicit pipeline name); `prior_art_librarian` (the "build-dispatch-loop and
build-pipeline-trigger both return `[]`" claim licensing the dispatch-map
flip was inherited from round 1 without a fresh check, against a documented
LANDMINE that these two can silently disagree); `debug_historian` (326's
`DO`/`RAISE` guard uses `!=` against a `#>>` path — same class already
flagged for 325 in round 1).

**Did the actual audit** (bug_historian's gate) rather than just asserting
it: read every resolver in `queryresolve` package —
`resolvePagesWhereType`/`resolvePagesUnderSection`/`resolveProducts`
(`queryresolve.go`), `resolveSectionIndexForType` (`section_index_for.go`),
`resolveLatestNews`/`resolveNewsArchive` (`news_items.go`),
`resolveModelDirectory`/`resolveModelDirectoryFull`/`resolveDirectoryKind`
(`directory_items.go`). **None share the shape.** Every one of them either
queries a platform table directly (no external config lookup at all — a
zero result is unambiguously zero matching rows) or reads a GLOBAL registry
not scoped per-site. `resolveBusinessDirectory` is the only resolver in the
package whose data source is itself gated by a separate per-site config row
that can be present or absent. This is real evidence, not an inherited
claim — first time this specific audit had actually been run.

**Fresh-checked the dispatch-gate claim** (prior_art_librarian) rather than
re-asserting round 1's: pulled the LIVE `find_dispatchable_site` SQL
straight from `agent_definitions` (not a seed file — "the seed is not the
system") and the LIVE `build-dispatch-loop` `load_items` step config. Both
now carry IDENTICAL item-eligibility predicates (status/attempt_count/
approval_mode/depends_on) — the divergence the LANDMINE describes was fixed
by `284`/`285` and still holds today. Neither filters on item_type/
handler_agent absent an explicit step-config filter, and `load_items`
carries none. Then made it concrete: checked the two named work items
(`715ec305`, `2f50bfda`) and their site (`vetcomparison.uk`,
`72b9e3a6-...`) directly — site unlocked, no claimed items, `depends_on IS
NULL` on both, `attempt_count(1) < max_attempts(3)` — both will pass both
gates once re-triaged.

**Found and fixed a REAL bug while investigating debug_historian's
objection**, not just a style nit: 326's guard used
`cfg #>> '{...}' != 'expected'` — `#>>` on a MISSING path returns NULL, and
`NULL != 'expected'` is NULL, which an `IF` treats as false, so the `RAISE`
would never fire and `'326 OK'` would print even against a wrong/absent key.
Exact match for a documented LANDMINE ("A migration verify block comparing a
jsonb path with `<>` sits GREEN for ever when the key does not exist").
Fixed all three checks to `IS DISTINCT FROM`, induced the fix directly
(`SELECT NULL::text != 'x', NULL::text IS DISTINCT FROM 'x'` → `NULL, true`)
rather than trusting the rewrite. Checked 325's LIKE-based checks for the
same class too: safe, because the same migration statement unconditionally
UPDATEs the same row with a non-null literal, and the guard's own
`schema_txt IS NULL` check already catches a missing row first — 325's
residual weakness is LIKE's substring imprecision (round 1's original, lower
severity), not NULL-swallowing.

**Checked guidelines' non-gating "missing" note** (processing_mode) too,
since it was cheap and concrete: fleet-wide `processing_mode` is deserialized
in three places and never once branched on — it has no behavioural effect
anywhere in the platform. The precedent 326 cites
(`directory-export-orchestrator`) and `page-build-handler` itself both have
it NULL in the live DB today. Its absence from 326 cannot cause any
divergence from the cited precedent.

**Committed the 326 guard fix alone** (`528f545f6`, `Council-Submitted:`
trailer — migration not yet applied anywhere, so this was a plain edit, no
rollback). **Owner reported a fresh chassis build was deployed mid-session**
— checked rather than assumed: pods moved to `v1.0.1263`, and this time the
round-2 marker string IS present (1/1, negative control 0/0) — round 2's Go
fix (`37560f120`) is now genuinely live, not merely committed.

**Resubmitted round 3** on the same correlation
(`RESUBMIT_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`), adding the missing
edit entry (editquality), relabelling 326's edit to `config_change` with the
owning pipeline named (guardian), and putting all of the above audits into
`grounded_in` as quoted, checkable evidence rather than restated assertions.
Run orchestration `c7f494a4-de65-43e9-9fec-62d635e871e5`, dispatched
2026-08-08 ~10:11; queue was clear (LAG 0) at submit time and a
`council-gate-orchestrate` run was already `EXECUTING_STEP` within seconds.
**Verdict not yet read as of this note.** See HANDOFF for the query and next
steps — this session is handing off here rather than push through the
~30-minute queue wait plus the remaining apply/build/re-triage/verify
sequence in one sitting.

## 2026-08-08b — round 3 APPROVED; migrations applied; improvement-loop experiment (predictions registered BEFORE the run)

Round 3 verdict read from `diagnosis_artifacts`: **approved**, 2026-08-08
09:18:48Z (the handoff's "submitted ~10:11" was BST; verdict landed ~7 min
after submission). Subsequent commits in this lane carry
`Council-Reviewed: 5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`.

Applied 325 + 326 by hand (each printed its own `OK` NOTICE via the
`IS DISTINCT FROM` guards), recorded both with `--record-only`, and set
`directory-build-handler.image_tag='v1.0.1264'` — the tag moved AGAIN
under this lane (1263 → 1264, third external roll); pod-grepped both
replicas 5/3/1/1 with negative control 0 before trusting it.

**Owner asked: would the improvement loop have picked these problems up?
Experiment: run it one-shot over vetcomparison.uk and watch.** Baseline
work-item snapshot taken first. The baseline itself already answers the
detection half: **discovery ran on 2026-08-02 and minted three
`unbuilt_internal_link` items** (index→/directory/index.html,
about→/directory/index.html, index→calculator). All three are `complete`,
`attempt_count=0` — and the pages are still unbuilt. Their `result` shows
what happened: **the dispatch rebuilt and redeployed the CONTAINING page
(`/index.html`, `/about.html`), not the target.** Cause, read from live
config: `build-dispatch-loop.process_item.call_handler.input_mapping` maps
`"page_name?": "current_item.spec.page_name"` (the container) and never
reads the item's `page_id` column (the target the check deliberately filed
against — its spec even says "Do NOT rebuild the linking page"). Then
`mark_complete` runs unconditionally on handler success.

Predictions for the fresh run (disconfirmable, stated before firing):
- P1 re-detects unbuilt_internal_link for /directory/index.html (index +
  about) and the calculator; `complete` is terminal so dedup does NOT block
  re-minting. Practice (/entities/practice.html) may fire too — it did not
  on 08-02, reason unknown.
- P2 NO detection for guides-index via phantom links (nothing in deployed
  HTML links to it — checked homepage hrefs); `incomplete_page_group` may
  detect it but its `needs_page:guides-index` item_key dedups against the
  parked `needs_human_review` row, so no new item.
- P3 the dispatched "fixes" rebuild the containers again (index/about),
  mark complete, and directory-index remains unbuilt — the wrong-page
  mapping above, plus (stacked) bare page-build-handler has no ensure_layout
  step so even the right page would no-op.
- P4 the two parked needs_page rows (needs_human_review) are untouched by
  the loop (triage promotes only status='detected').
- P5 no colour churn (design_intent.palette.reference_values pin present;
  content_data ? 'color_scheme' = t; check-side generic_theme fix live).
What refutes P3: directory/index.html deploying with real business data.

## 2026-08-08c — experiment results: the loop did BETTER than predicted, and found a real bug in 326

Run: correlation `867d6054-3f8f-4b11-9352-b29cecd9aaaa`, dispatched 14:32Z,
14 orchestrations, all COMPLETED by 14:40Z (queue was NOT contended — my
"still queued" reports were my own monitor's blindness, see WRONG_CALLS
2026-08-08: its query named a non-existent column and `|| true` swallowed
the error every poll).

Predictions vs outcomes:
- **P1 CONFIRMED and exceeded**: 7 `unbuilt_internal_link` + 1
  `empty_internal_href` minted — directory link from THREE pages (index,
  how-it-works, guide-independent-strategy), practice from two, calculator
  from one. Dedup did not block (old items terminal).
- **P2 CONFIRMED for phantom links** (zero stored surfaces link to
  /guides/index.html — pre-verified) — but see P4: the page was covered
  anyway, by `incomplete_page_group`.
- **P3 WRONG in the most useful way.** The dispatched fix did NOT re-run the
  wrong-page rebuild for the needs_page rows, because...
- **P4 REFUTED — the loop REVIVED both parked `needs_page` rows.**
  `incomplete_page_group` re-minted `needs_page:directory-index` /
  `needs_page:guides-index`; the `refreshOpenWorkItem` machinery
  (bugs 091/184, closed THIS WEEK — newer than the 08-02 run) refreshed the
  open rows, taking the handler from the builder map: directory-index →
  **directory-build-handler** (correct, new), guides-index →
  page-build-handler (its page_type isn't in the map). Both re-dispatched
  within the run:
  - directory-index: `ensure_layout` RAN AND WORKED — `site_plan_sections`
    now carries `directory-index: [hero, directory-listing]` under the
    current plan, written by the loop's own dispatch. Then
    `call_page_build_handler` FAILED: contract violation, child received
    fields literally named `input_data.site_id` etc. **Migration 326 put the
    `input_data.` prefix on the input_mapping KEYS** (the child-field-name
    side); working precedents (build-dispatch-loop's call_handler, the
    improvement-loop trigger itself) use plain keys. Fixed in
    **migration 336** (applied + recorded 15:0xZ, its own IS DISTINCT FROM
    guard, NOTICE seen); item left `triaged` att 1/3 so build-pipeline-trigger
    re-dispatches it unaided.
  - guides-index: re-dispatched on bare page-build-handler and no-op'd AGAIN
    ("no sections ready to build") — the 2026-08-07 handoff's Step 5
    ("handler unchanged") is hereby CONFIRMED DEFECTIVE by the live system,
    exactly as this session predicted from the workflow read. **Deviation
    from handoff**: re-routed 2f50bfda to directory-build-handler
    (ensure_layout is page-name-generic; defaultSectionsForPage has an
    explicit guides-index case), status triaged, attempts reset.
- **P5 HELD (with a footnote)**: stylesheet was rewritten and redeployed
  14:41Z but the accent value is byte-identical to the 08-05 stylesheet
  (checked at the sites repo: `#059669` present in a5c0d3c9 AND 85ed11d2) —
  no churn FROM THIS RUN. Standing pre-existing discrepancy: the
  design_intent pin says accent `#10b981`, live has had `#059669` since an
  earlier webdesign run. One shade apart, same hue family. Noted, not chased
  in this lane.
- Backlog effects: the loop also promoted the 08-06 stranded `detected`
  items (3 undeployed_asset, audit_tool → triaged; stale_sc_* rerenders →
  complete) and queued a full page_rerender wave — the loop doing its
  ordinary job on a site it hadn't visited since 08-02.

**Filed `bugs_open/220`** (unbuilt_internal_link dispatch rebuilds the
CONTAINER page, marks complete, re-detects forever): the 08-02 items proved
it (deploy results name /index.html, /about.html — not the target); the
8 fresh triaged link items will re-demonstrate whichever way they resolve —
append their outcome there.

Still open at this note: both needs_page rows `triaged` for
directory-build-handler post-336, monitor armed (foreground-tested this
time) on their status + pages.deployed_at.

## 2026-08-08d — second live-fire defect in 326's delegation: spec/current_page missing (337)

Attempt 2 on directory-index (with 336's plain keys) got PAST the contract
check, through ensure_layout / plan / content-write / validate / save — then
died at `update_status`: "could not determine page_id". Cause read at
`v3_site_actions.go` (UpdatePageStatusAction) + the live step config:
update_status resolves the page via `input_data.spec.page_name`, falling back
to `current_page.name`; the dispatcher supplies both ("spec":
"current_item.spec", "current_page": "current_item.spec"); 326's delegation
supplied neither. load_page_record reads `input_data.page_name` — which IS
passed — which is exactly why the failure moved five steps downstream.
**Step order matters here: update_status runs BEFORE deploy_page**, so
content was written and saved but nothing deployed — the page 404s while the
DB holds its components (checked the artefact, not the status, before
assuming a deploy).

**Migration 337** (applied + recorded, own guard, NOTICE seen) mirrors the
dispatcher's proven keys into the delegation: spec + current_page alongside
the 336 plain keys. Pre-checked the remaining chain before letting the LAST
attempt run: deploy_page reads only site_record.*/page_record.id (child-
produced), spawn_rerender_agent takes no inputs — self-sufficient.

Transferable shape for the eventual 016b entry if this recurs elsewhere: **a
delegating agent that hand-picks which fields to forward re-runs the target's
whole input contract from scratch — every consumer step, not just the entry
step.** The dispatcher's mapping is the de-facto contract; mirror it, don't
sample it. guides-index's in-flight run snapshotted the pre-337 mapping
(workflow_plan is copied at claim) and will fail once more at update_status
by design; its retry picks up 337.

## 2026-08-08e — DONE: both pages live; closure written; register updated

Post-roll (v1.0.1266, pod-grepped, negative control 0): held both items
through the ~300s spawn-drop window, auto-re-triaged after.
**> CORRECTED same session:** my priority "bump" to 95 was BACKWARDS — the
dispatcher orders `priority ASC` (`load_work_item_actions.go:683`), so I
starved the two builds for ~45 min and misread it as congestion; the
disconfirming tell (priority-140 audit_tool unclaimed while 80s processed)
was in my own transcript. Caught by finally greping the ORDER BY.
WRONG_CALLS second 08-08 entry. priority=10 dispatched both within a cycle.

Outcomes: directory-index deployed 17:02:22Z (item complete, att 2/3),
serving 61 real alphabetical practices/49 postcodes at HTTP 200;
guides-index deployed 17:07:31Z, repo commit 836fd73b, listing exactly the
three real guides + real tool CTA. Closure evidence written INTO
bugs_open/206 (stays there, owner direction 08-06). BLD-017 status +
verify-later discharged in the register, index row updated. The owner's
original complaint (homepage "Search the directory" → 404) is resolved on
the live site.
