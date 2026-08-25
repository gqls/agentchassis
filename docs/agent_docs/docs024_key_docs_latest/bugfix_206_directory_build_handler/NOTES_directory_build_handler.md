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

---

## 2026-08-24 — lane resumed (16 days quiet). Re-verify → a live re-fire of the class → three of my own errors caught by other people's checks

**Task**: pick the lane back up, confirm the bug is still valid, prefer a framework-wide fix
over the individual case, check the council, check other threads.

### What I verified first (artefacts, not statuses)

Both 08-08 pages still serve: `/directory/index.html` 200/52,699 B with 49 postcodes and the
alphabetical practice list; `/guides/index.html` 200 with `guide-list`. **Both were re-rendered
by a fleet wave on 08-23 and the listings survived** — the pipeline reproduces the result, it is
not a preserved artefact. The `vetcomparison` lane (live, in parallel) proved a THIRD page shape
on `directory-build-handler` the same morning: `practice`, deployed 10:17:38Z first attempt.

### The finding, and it is not where I was looking

I went looking for the residual I expected (`entity-page` builder, `section-index` map entry) and
found instead that **the 08-08 fix was never fleet-wide, because the fleet has more than one
door**. `reconcile_site_plan_action.go`'s emit hardcodes `handler_agent='page-build-handler'`
and never consults the builder map. Five items parked with this bug's exact signature; the one
that settles it is garden-tools.uk's `brand-directory-index` — an `entity-directory` page
**parked 15 days while its builder was live**. Two copies of one decision, disagreeing. Fixed by
making the decision exist once (`builder_routing.go`), in a NEW file specifically so it needed
nothing from `load_work_item_actions.go`, which is uncommittable (below). Committed `d1aa231aa`.

### Three of my own errors, all caught by someone else's check, all in WRONG_CALLS

1. **I adjudicated my own lane's history from memory and asserted the pre-correction version** to
   the vetcomparison lane (that the 08-08 "bump to 95" was correct). It was not — my own NOTES
   above hold the correction. They had already committed a NOTES entry crediting me before I
   caught it. Counter-correction sent; their chain `a0c8fa18b → 98beb8b92 → aa26df458`.
2. **A census that returned a false ZERO.** I filtered on `spec->>'page_type'`; reconcile-minted
   items carry no such key, so the population was invisible to the very query meant to count it.
   Caught by the `345` session relaying its own tripwire lesson that morning — *take a
   known-real occurrence and run your own WHERE against it*. The corrected census (join
   `pages.page_type`) found the five items and **redirected the entire fix**. One check, and the
   bug turned out to be the producer rather than the map.
3. **A phantom column in a council submission.** Two `psql -c` outputs concatenated and I read
   the second query's single result as the last column of the first — "`site_plan_pages` has
   `page_type`". It does not. **The council's round-1 gating objection was exactly this**, found
   independently by the `editquality` seat. The shipped code never had the defect (it reads
   `pages.page_type` with a fallback to the plan's `role`); the submission text did.

### And the one that matters most: I re-found a closed decision

`bugs_closed/187` measured this exact population on **2026-08-03**, names these very rows, and
its shipped fix states *"`reconcile_site_plan` deliberately NOT guarded (015-shape gaps are real
findings)"*. So my `capability_gap` arm touches another lane's considered, council-directed
decision. **`who-owns.py` and "grep before you file" both key on the bug's IDENTITY; this
defect's identity was a symptom string**, and a number-keyed search cannot find symptom-keyed
prior art. I only got there by chasing the discrepancy between my census ("five") and the
council's unscoped check ("87 across 16 sites"). Contribution recorded in 187's file, the
contested arm put to the council as round 2's headline question, revert offered.

Corrected class census (dedup, 08-24): **87 items / 16 sites**, 79 parked. **69 are tool pages**
(67 layout-less), 42 of them `unbuilt_internal_link` rows from `bugs_open/220` — the bug this
lane itself filed — so that tail is being refilled faster than it drains. My change reaches ~11
non-tool typed items and must not be read as fixing the tool class.

### Cross-lane, worth keeping

- The `345`/`326`/`311` trio sitting dirty in `load_work_item_actions.go` is **ownerless**. I
  measured it and told the `345` session: it compiles but **fails three existing
  `UpdateWorkItemStatus` tests** (three-arm git-archive overlay: clean HEAD green; HEAD + my six
  files green; HEAD + trio red). They reproduced it and diagnosed the cause — a positionally
  declared sqlmock `WithArgs` that the trio's new bind parameter breaks. So it is not "finished
  and unowned" but **incomplete and red**, which is a materially different thing to hand a human.
  My new file kept me clear of it entirely.
- Method worth reusing: **establish WHOSE a failure is before reporting it.** Three overlays, one
  message, actionable rather than accusatory.

### Council rounds 1–3 (same session), and what each round actually bought

Worth recording in full, because the lane's own experience is that a REVISE is cheaper than the
defect it finds — and all three rounds found something real rather than a formatting quibble.

**Round 1 → REVISE, gating from `editquality`.** It caught the phantom
`site_plan_pages.page_type` column. I had already caught it myself (by trying to USE the column
and getting `ERROR: column spp.page_type does not exist`) so the code never carried the defect —
but the submission text did, and the seat found it independently from the schema alone. Its
checks also returned **87 items across 16 sites** against my "five", which is what sent me to the
unscoped census, which is what found `bugs_closed/187`, which is what turned this from "I found a
new bug" into "I re-found a closed decision". **That single discrepancy was worth the round.**

**Round 2 → REVISE, gating from `bug_historian`, plus two more HIGHs.** All three were
"you asserted, you did not query", and all three were answerable in one psql call each:
- Is `directory-build-handler` actually a live agent, or am I minting items into dormant
  machinery? → `is_active=t, is_snapshot=f, deleted_at IS NULL`. Live.
- Are `tool-builder`/`entity-page-builder` genuinely absent, or am I parking buildable pages? →
  **0 rows.** The absence is real, which is exactly what makes the gap arm right.
- Does a `deferred` row with a non-empty `handler_agent` really stay undispatched? → **this one
  changed the code**, see the WRONG_CALLS entry: my 262-row demand control passed and was blind
  on the axis the objection was about. Row now carries an empty `handler_agent`, matching 47 of
  47 existing `capability_gap` rows.

Two other round-2 objections I did NOT simply comply with, because the premise was checkable:
- `guidelines` cited a "DELETE+INSERT, not ON CONFLICT" rule for dedup-covered inserts. The
  estate's own shared helper `insertWorkItem` uses `ON CONFLICT (site_id, item_key) WHERE … DO
  NOTHING`, and reconcile's three pre-existing INSERTs use the bare form. Hand-writing the
  index-matching WHERE would couple this call site to `workItemTerminalStatuses` — the 42P10
  lockstep trap. Kept the bare form and flagged the premise rather than quietly complying.
- `editquality` on minimality (the suffix widening is a different mechanism): **conceded**, it
  is off the causal path. Removed from the plan; it ships separately. Note the honest wrinkle —
  it is already committed, so the separation is of the REVIEW, not of the shipping, and the
  round-3 submission says so rather than implying a clean split.

**The split-brain objection came from FOUR seats** (guardian, reuse_agent, editquality,
bug_historian) and is the one I could not close: `WriteBuildItemsAction` still holds its inline
maps, so two producers can route the same `page_type` differently. That is genuinely the shape of
the bug being fixed, appearing a third time, and the seats are right to say so. What I could do
was bound it — one page_type, one path, no regression on that path, shared dedup key so a page
cannot hold two items — and refuse to close it the wrong way. **I will not land another lane's
unowned, measured-red, shared-seam change to tidy up my own submission**, and the round-3
rationale offers the council the contained alternative (drop the section-index entry until the
swap is possible) rather than assuming my answer.

### Stopping rule for the council rounds (written BEFORE round 4's verdict, so it is not a rationalisation)

Four rounds is a lot, and each has been productive, so "keep going" has been the right call each
time. But it can stop being right, and the decision is worth pre-committing rather than making
after reading a verdict I dislike. **Round 4 will be resolved as follows:**

- **A gating objection naming a real defect** → fix it and resubmit. This has happened twice
  (the deferred row's non-empty `handler_agent`; the held edit sitting in the executable
  `edits[]` array) and both were worth the round on their own.
- **A gating objection whose premise is checkable and false** → answer it once more with the
  query, as with the `DELETE+INSERT` premise that the estate's own shared helper contradicts.
  Answering is not defending, provided the answer is evidence rather than argument.
- **Another objection to the two questions I have explicitly asked the council to RULE on** —
  the `capability_gap` arm versus `bugs_closed/187`'s deliberate decision, and the transient
  split-brain — **then stop asking and take the contained option myself.** Those are judgement
  calls, not defects; if a round keeps objecting without ruling, the objection *is* the answer,
  and the contained option was mine to take from the start: **revert the `capability_gap` arm,
  keep the routing half.** The routing half is where the measured win is (an `entity-directory`
  page parked fifteen days while its builder ran), nothing has objected to it on substance, and
  it is untouched by 187's decision. The gap arm is an improvement to how a *known-unbuildable*
  page is recorded — real, but not worth a fifth round or a fight with another lane's ruling.

The forcing consideration: this is an ADVISORY gate on a shared tree, and the code is already
committed and inert. The cost of another round is not "the fix is delayed" — it is my own time
and the fleet's credits, against a decreasing return. The cost of shipping something a seat has
flagged three times without resolution is worse: it is exactly the "one lane overrides another
lane's ruling because it had more stamina" failure this estate has no defence against.

### Round 4 → the fix got SMALLER, and that was the right outcome

My pre-committed stopping rule said a third objection to a question I had asked the council to
*rule* on means taking the contained option instead of a fifth round. Round 4 met that test on
the split-brain — `reuse_agent` HIGH, plus `guardian` — but it also did something the rule did
not anticipate: **it converted the judgement call into a measurable defect**, which changed what
the contained option should be.

The seats' standing objection was "two implementations of one decision is an anti-pattern".
I had answered twice with a bound: one page_type, one path, no regression, and a shared
`item_key` so a page cannot hold two items with different handlers. Round 4's `guardian` took
that last clause — my own mitigation — and showed it cuts the other way: because both producers
mint the same key under `idx_swi_dedup`, **whichever fires first wins and the other is silently
dropped**. So a page `WriteBuildItemsAction` reaches first keeps the wrong handler and my fix
never fires for it, with nothing anywhere saying so. My "no page gets two rows" was true and was
not the reassurance I was using it as.

**The option neither I nor the council had named**: `section-index` was the *only* line on which
the two maps differed. Remove it and they are byte-identical — the divergence stops existing,
rather than being tolerated or documented. And the case that motivated the whole lane survives,
because `entity-directory` was **already** in `WriteBuildItemsAction`'s map: routing reconcile
through the shared authority fixes `garden-tools.uk/brand-directory-index` with no disagreement
at all. What it costs is two `section-index` pages staying parked, no worse than today.

Worth keeping as a pattern: **when several rounds object to the same shape, look for the input
that is causing the divergence rather than arguing about the divergence.** I spent two rounds
defending a bound and one line deleted the thing being bounded.

Also from round 4, both measured rather than argued:
- **HIGH refuted**: `directory-build-handler` reads no `item_type` in any of its four workflow
  steps (`jsonb_path_query_array` for a step config carrying `item_type` or `handler_agent`
  returns `[]`), so it cannot dispatch on one — which was the premise of the objection.
- **The historical rows cannot be produced at all**: `2f50bfda` and `715ec305` return NO ROWS
  today. `site_work_items` is a rolling window and completed rows are archived out. I said so
  rather than quoting rows I could not show.
- **One objection was factually wrong and I said so with the test name** (the role fallback *is*
  pinned, and mutation-proven). Answering is not defending when the answer is a test.
- **One implied fix would have been a live defect**: binding `route.itemType` in reconcile's
  INSERT would mint `needs_content_page`, which `loadOpenPageItems` (`:683`) does not select —
  so the action's own dedup check would go blind and re-emit the page every run.

### The 090 loop was not available for this mechanism, twice

Run 1 multi-symbol: doomed at bundle 1. Run 2 **single-symbol** — exactly what the landmine
prescribes — passed at bundles 1 and 2 and blew the budget at bundle 3, because the loop widens
its own scope between iterations. Ended `complete` with zero verdict artifacts. **The landmine
has been further-corrected with that measurement and re-verified**: checking after the first
bundle is necessary and not sufficient, and for a mechanism in these files you should plan on the
owner ruling's declared-substitute path from the start.

## 2026-08-25 — the swap, and four things I got wrong on the way to it

### The blocker: re-checked, not assumed

The 08-24 handoff said the `WriteBuildItemsAction` swap was blocked on an ownerless dirty hunk and
told the next session to **re-check**. `git status --porcelain
platform/orchestration/actions/load_work_item_actions.go` → empty; `scripts/verify-head-builds.sh`
→ `OK — HEAD bba8a892d builds`. Blocker gone; the file had been committed by other lanes
(`1789489bf`, `f16c87beb`, `6ab0b3434`) in the interim. **The re-check instruction was worth its
line** — the blocker had cleared without anyone in this lane doing anything.

### What shipped (`efec862f4`, council `b92e624d` APPROVED round 1)

Inline maps deleted from `WriteBuildItemsAction` → `builderForPageType`; `section-index` added to
the shared map in the same commit (round 4's condition); `capability_gap` `handler_agent` → EMPTY
at this door; two false comments corrected in place; five tests added where there had been none.

### Misstep 1 — my test passed. Then it passed under mutation. TWICE over.

First draft of `write_build_items_routing_test.go` passed immediately. Mutated `section-index` back
to `page-build-handler`: **still passed**. Two independent causes, either fatal alone:

1. A permissive `mock.ExpectExec("INSERT INTO site_work_items")` registered as scaffolding for the
   trailing site-level items **absorbed the mis-routed INSERT**, leaving the pinned expectation
   merely unused — which sqlmock never reports unless `ExpectationsWereMet()` is called, and that
   is not workable here because the shared door probes inside `writeWorkItem` fire a
   case-dependent number of times.
2. `WriteBuildItemsAction` **swallows** the per-page insert error (`logger.Warn(...); continue`),
   so the action returns success either way and my assertion on the returned `err` could never
   fire.

**The part worth carrying is not "mutate" — it is that fixing cause 1 did not make the test fail,
and that is what exposed cause 2.** A mutation that still passes after you have fixed the reason it
passed means there was another reason. I was ready to stop after the first.

Shipped shape: every INSERT pinned by its **own** `item_key` so nothing can stand in for anything
else, and the outcome read from `writeWorkItem`'s `"Work item inserted"` log line, which it emits
only after `RowsAffected > 0`.

### Misstep 2 — then all four tests failed, and none of their messages was about the cause

The corrected tests failed with my carefully-written messages about handlers. The cause was
`nav_order: nil` in the fixture's page row: `scanPageRowsForBuild` scans that column into an `int`,
the row failed to scan, the page list came back **empty**, and the per-page loop was a no-op. The
action logged `"no per-page builds needed; queuing site-level items only"` and returned success.

So the first version was doubly incapable — had the permissive expectation not absorbed the call,
**there would have been no call to absorb.** A fixture producing zero rows passes every assertion
about what those rows contain, and announces its emptiness in a line that reads like ordinary
operation. The shipped fixture returns `inserted bool` and every test fails loudly on `!inserted`.

### Misstep 3 — a claim I committed on 08-24, inside an APPROVED round, disproved by one grep

`reconcile_site_plan_action.go` said `WriteBuildItemsAction` *"files into a different key
namespace"*. It does not — `load_work_item_actions.go:335` files `needs_page:<name>`, the same
namespace as `:289`. I only read the other producer's literal because the swap forced me into that
file. The sentence made the two doors look non-colliding, **which is the exact opposite of what the
round-4 guardian said and whose objection shaped the change the comment explains.** Same commit's
header carried a second false claim (`"Open" = 5 statuses, same set the dedup index uses` — the SQL
has 6, the index 7). Both corrected in place, dated. The status divergence turned out to have a
live casualty, so it went into `bugs_open/206` §5(b) rather than just being tidied.

### Misstep 4 — I nearly filed a residual on an inference, and the census refused it

Three council seats asked whether a **third** copy of the routing map exists. I had not checked.
Checking found six `needs_page:` minting sites, not two, three of them hardcoding
`page-build-handler` — including `page-rerender`, the fleet's most active producer (414 rows, 21
typed, still minting today). The next sentence wrote itself: *"so 206 is still live through the
rerender door."*

**It is not.** 26 typed-page rows from those producers, `error ILIKE '%no sections ready to
build%'` → **0**. Their failures are the owned-page guard working as designed. And the mechanism
explains the zero, which is why I trust it: 206 is the **layout-less** case, and a rerender or
adoption target already has a layout. The two doors that consult the map are the two that mint at
**plan** time, when a page has nothing yet. The doors that skip it are the doors whose pages do not
need it.

I would have filed that residual if I had stopped at the grep.

### The instrument defect — the closure test could be passed by a hand repair

Preparing to run the 08-24 closure query, I checked what else could put `directory-build-handler`
in `handler_agent` on a `reconcile_site_plan` row. Answer: a human, via the documented operator
escape hatch. `[MEASURED 2026-08-25, live UNION archive, all history]` **three rows match the PASS
predicate fleet-wide and all three are hand re-routes** — `created_at` 2026-07-17, `updated_at`
08-08 and 08-24, `spec ? 'page_type'` false on all three. Run without a domain filter, that query
would have declared the fix proven by rows the replaced hardcode minted.

Fix: gate on `spec ? 'page_type'`, which the fixed emit writes and an `UPDATE` of `handler_agent`
cannot forge, and whose population is currently **empty** (508 reconcile-minted rows, none
stamped) — so the first stamped row is necessarily the fix. RUNBOOK §7. Landmine filed.

### State of the wait

`[MEASURED 2026-08-25]` reconcile rows created after the 08-24 15:39 roll: **0**. Sites with
`last_reconciled_at` after 08-24 12:00: **0**. Newest reconcile anywhere: `agritec.uk` 08-24 11:26,
*before* the roll. The five parked rows are untouched. The free proof has not arrived; it is still
the right thing to wait for.

### Misstep 5 — I probed a deploy three wrong ways in one command, with a control that vouched for all of them

Asked to coordinate with the `bugs_open/381` lane (building a greenfield site — my closure
artefact), I first needed to know whether the live fleet carries today's swap. I got the right
answer by luck and the wrong method three times over:

1. **Probed a symbol both versions carry.** `grep -aq "builderForPageType" /proc/1/exe` → PRESENT
   on both replicas. That symbol shipped **2026-08-24** (`d1aa231aa`); it dates nothing about today.
   I had written the words *"probe the CAPABILITY, not the commit"* into this lane's docs yesterday.
2. **Grepped the binary for an ANCESTOR's sha.** `LANDMINES.md:9243` says it in as many words —
   *"Test ANCESTRY, not equality — the stamp is whatever HEAD was at build time … your commit is
   normally an ancestor rather than the stamp."* A build stamps **one** sha, so grepping for
   `d1aa231aa` returns absent even though that code is demonstrably live. My probe reported
   `MINE absent` and `OLD_0824 absent`, i.e. it "disproved" a deploy that yesterday's session had
   verified at the artefact.
3. **And my negative control was VACUOUS.** `0000…0000` (40 zeros) came back **PRESENT** — it
   matches Go's internal tables, which is precisely the trap
   `MEMORY: a-fresh-deploy-can-ship-no-new-code` names ("a control that matches everything (40
   zeros) hides it"). So the one thing in the command that existed to tell me the method was broken
   **agreed with the method**.

The first two probes were fabricated prefixes too (`efec862f4a5b` is not a prefix of
`efec862f40404b…`) — so two of the five probe strings could not have matched anything regardless.

**What actually settled it, and should have been the first move: two timestamps.** Pods started
`2026-08-25T09:27:48Z`; `efec862f4` is committed `09:58:33Z`. **A binary cannot contain code
committed 31 minutes after it started.** No cluster exec, no grep, no control needed.

**The check, for next time:** to date a deploy, read the stamp (`logs … | grep -m1 'build
provenance'`) and run `git merge-base --is-ancestor <my-commit> <the stamp>`. If the startup line
has scrolled — it had — fall back to **pod `startTime` vs `git log -1 --format=%cI`**, which is
free, local, and has no failure mode. Grep the binary only for a **known** value you expect
present, never to test ancestry.

Reported to the 381 lane with the timestamp reasoning and an explicit note that the symbol probe
and its control are not to be relied on — I would rather hand them the caveat than a clean-looking
PRESENT.

### Misstep 6 — I warned a peer lane about a mechanism without checking its precondition on their data

Messaged the `bugs_open/381` lane mid-build: your `section-index` pages **will no-op**, because
today's swap is unrolled. They adopted it into their acceptance guide.

False for that build. 206's no-op needs a page with **no layout from any source**; `[MEASURED]` all
17 of their `section-index` pages carry `["hero","generic-text-block","content-listing"]`, and
`april-index` had **already built and deployed** through the very handler I was warning about,
before I sent the retraction. One `jsonb_array_length(sections)` query — on rows I had already asked
them for — refutes the whole warning.

**I knew the precondition. I wrote it into `builder_routing.go` that morning** ("*the only one that
can build a page whose layout is missing from every source*") and then warned about the consequence
without testing the condition. The pattern is the day's pattern: I keep asserting the *consequence*
of a mechanism I have correctly described.

**The harm direction is what makes it worse than a wrong prediction.** Their guide would have sent
an investigator to check `handler_agent` first on pages where `page-build-handler` is **correct** —
converting a real 381 finding into someone else's known bug. *A check that stops an investigation is
more dangerous than one that starts a wrong one.* Retracted within the same build, with the
measurement, and asked them to invert the line while keeping the error string as a discriminator.

**And they caught me too**, which is the useful half: they flagged that `spec ? 'page_type'` was
TRUE on their whole build, contradicting the "134 of 134 absent" caveat I had in circulation. They
were right — the stamp shipped **2026-08-24** (`d1aa231aa`, live since `v1.0.1334`), my census was
taken before any reconcile had run since that roll, and the population emptied by *time*. They had
one build and the correct conclusion; I had the fleet history and a stale premise.

## 2026-08-25 (night) — the roll landed, verified properly this time; and the closure gate stopped being perishable

### The deploy check, done the way §7c prescribes — and it worked first time

Yesterday's misstep 5 was three broken probes in one command. Tonight, the same question with the
prescribed method:

1. **Pods are fresh**, so the startup line is in range: both replicas print
   `git_commit: a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5` on `v1.0.1339` (started `19:07:18Z` /
   `19:07:49Z`).
2. **Ancestry, not equality**: `efec862f4` (the swap) and `0777eb297` (the test pins) are both
   ancestors of that stamp.
3. **With a control in the same breath**: two commits made *after* the build (`c591d8d61` 20:20,
   `08bfce067` 20:19) are correctly **NOT** ancestors. That is what makes the IN results mean
   something — and it is the step whose absence produced yesterday's 40-zeros false positive.

**`section-index` → `directory-build-handler` is live at both doors.** And nothing has re-routed as
a result, correctly: 0 reconcile runs since the roll, nothing schedules reconcile, and a parked row
blocks its own re-mint.

### Why I spent a council round on a stamp

The closure gate as it stood was `spec ? 'page_type'` + `updated_at < created_at + interval '1
second'`. Sound, but the second condition **expires on the first legitimate dispatcher claim** —
minutes, in practice. So the proof would exist only in a window nobody is guaranteed to be watching.

`spec.handler` — the handler the emit *chose*, written alongside the column it wrote — makes
`spec->>'handler' = handler_agent` a permanent check. A hand repair updates the column and not the
spec, so a repaired row diverges for ever. **Divergence becomes the signal** rather than an absence
being one, which is the right way round.

Both doors stamp it. A stamp at one door only would mean a reader cannot tell a mint from a repair
at the other, and — worse — the *absence* would read as "repaired" rather than "not stamped". That
is precisely the two-doors-disagree shape this whole bug is about, so shipping it one-sided would
have been rebuilding the defect in the instrument.

### Mutation proof, run before submitting

- drop reconcile's stamp → **only** `TestReconcileStampsItsRoutingDecisionIntoTheSpec` fails;
- drop the planner's → **only** `TestWriteBuildItemsStampsRoutingProvenanceMatchingTheHandler` fails;
- make the stamp **disagree** with the column → **both** fail. That is the case that matters: a stamp
  agreeing with nothing would read as a forged repair, which is worse than no stamp.

Baseline and restored both green, and I checked the tree afterwards for surviving mutations
(`grep some-other-handler` → clean) rather than trusting the restore.

The matchers decode the JSON rather than substring-matching, deliberately: a substring would match
any spec merely *containing* the handler name (`page_role`, `reason`), which is this file's own
recorded vacuous-match trap.

### One environmental note

The machine was at load ~191 with 1 GB of 30 GB free; a `go test` compile was **OOM-killed**
(`signal: killed`), which reads like a build failure and is not one. `go vet` (cheaper) confirmed the
test code compiled, and the suite ran fine once scheduled. Disk was not the issue (`/tmp` 28%, `/`
40%) — worth separating from the tmpfs trap, which presents as a linker error instead.
