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
