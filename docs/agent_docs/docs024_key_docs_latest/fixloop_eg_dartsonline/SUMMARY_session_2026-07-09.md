# SUMMARY — Diagnosis→Fix Loop session, 2026-07-09

Session recap, reproduced verbatim as requested. Companion documents:
RUNBOOK_diagnosis_fix_loop(10).md (task, phases, gotchas),
PLAN_fixloop_pilot.md (slices, criteria, scoring rubric),
NOTES_running_fixloop(10).md (turn-by-turn evidence trail and decisions).

---

## What happened, in short

I diagnosed the pilot bug before building anything — and it turned out not to be a guides bug at all. Then I landed the whole F0.1 plumbing layer: the artifacts table, the bundle write-through, the intake route, and automatic dispatch. Everything is verified against the live database; one piece (the Go write-through) is code-complete but waits on your image build.

## The diagnosis

The mandatory pre-check cracked it in about a dozen queries. Ten of dartsonline's fifteen pages were never built — only `content` and `landing` deployed. The discriminator is `pages.sections`: an exact partition, five with sections deployed, ten without still `planned`, no exceptions.

Three causes chain. The planner authored `sections` for only five of fifteen plan pages. Then `page-build-handler` routes sectionless pages from `check_has_ready_sections` into `complete_error` — which is an `action: complete_workflow`, a **success** terminal — so the work item is stamped `complete` for a page that was never built. And `populate_nav_tables_action.go:243` filters on `pages.status` (a lifecycle column defaulting to `'active'`) rather than `build_status`, so it cheerfully publishes a nav link to a page that only ever existed as a plan row. That is, precisely, "the system linked to something it never built."

The platform already knew: a comment at `load_work_item_actions.go:750` names the success-labelled terminal and prescribes a `mark_no_sections` step to flag it. **That step doesn't exist** — it appears nowhere in the repo but that one comment.

**The standing hypothesis was refuted.** It named `reconcile_site_plan`'s routing table. The table lives in `WriteBuildItemsAction`, and absence from it doesn't drop a page — unknown types fall through to `page-build-handler` by default. The differential resolves too: gamesdesign's guide pages deploy through the *same* handler, because they have sections.

Since we now hold ground truth with citations, I reframed the pilot from a discovery run into a **known-answer benchmark**, with a scoring rubric pre-registered so grading can't drift to fit whatever the loop emits. Blinding is mandatory — the answer now lives in the repo.

## What I built

`diagnosis_artifacts` (F0.1a), applied and idempotent. The bundle write-through inside `DiagnoseAssembleBundleAction` (F0.1b), where a persistence failure degrades to a warning on all four paths — observability must never cost a diagnosis. `090_TRIGGER_needs_diagnosis_v1.sh` (F0.1c), which writes a durable intake record and fires the 084 envelope on the same `correlation_id`. And `diagnose-dispatch-loop` plus its scheduled task (F0.1d), **shipped disabled**.

## Three times the code contradicted the plan

Each of these came from reading the live system rather than trusting a document, and each changed the design:

**Q-B was impossible.** "Null-site allowed" fails twice over — `site_id` is `NOT NULL`, and `LoadWorkItemsAction` queries `WHERE wi.site_id = $1` regardless. The platform had already solved it with the `system.internal` pseudo-site.

**Nothing filters work items by pipeline where it matters.** Not `load_items`, not `find_dispatchable_site`. That accident is the *only* reason the `maintenance` pipeline dispatches at all — so the obvious one-key fix would orphan it. Routed to the builder thread, not fixed. And `triage_detect_items` rewrites `pipeline` to `'build'` while its comment asserts a filter that doesn't exist. Both are the pilot bug's exact family: a rule enforced in one place and not its partner.

**Your challenge on `claimed` found a real bug.** The premise was off — the dispatcher claims, not the handler — but you were pointing at something true one step along. I'd argued for a status no sweep touches, then moved the item into `claimed`, the one status that re-enters that surface. It would have been reset to `triaged` after 40 minutes, and would have blocked build dispatch for `system.internal` for a diagnosis's entire duration. Now it's `diagnosing`, and the loop reaps its own dead runs.

Twice more, a grep beat a recollection: `nullIfEmpty` already existed, and `seed_scope` as a comma string would have silently parsed to nil.

Throughout I verified with negative controls rather than just checking a query returned nothing — flipping an item to `triaged` to prove the guard was the status and not a typo, executing the production SQL with typed parameters rather than reading it. That's what caught the upsert leaving `orchestration_id` stale.

## Where things stand

Blocked on your deploy: no bundle row can appear until `v1.0.1100` is rolled out. Then regenerate the bundle with `-psql` plus `site_plans`/`site_plan_pages` (the original invocation omitted both, and that's where the guides evidence lived), confirm the fixloop doc directory is excluded from the loop's corpus, run 090, and score against the rubric. Only then enable the trigger.

Two flags: your working tree has changes I didn't make (`plan_sections_action.go`, +80 lines, and an untracked `check_image_source_unsatisfiable.go`) — and `plan_sections` sits immediately upstream of `check_has_ready_sections`, so someone may already be in that code before F1 touches it.

---

## Addendum, 2026-07-10 — the council reviewing a plan, then reviewing its revision

*Added so the reports can be shown side by side. These are the actual council
outputs from one fix-proposer run (`e08c5b01`, the CONFIRMED darts diagnosis):
the loop proposed a fix plan, the council sent it back with objections, the
proposer rewrote the plan to answer them, and the council reviewed the rewrite.
Two reviewers sit on the council — an **edit-quality** reviewer (are these the
right changes; do they miss anything the diagnosis found?) and a **pipeline
guardian** (does this damage other pipelines?). A plain rule aggregates them:
any objection → **revise**; a guardian veto → **rejected**; all approve →
**approved**. Both rounds below came back **revise**.*

### How the plan changed between the two reviews

**Round 1 plan — six edits.** Two sound core edits (`defaultSectionsForPage`
gives a section-index page real sections; `applyNewPage` derives the right
`/guides/index.html` URL), plus **four that stray off the causal path** — one of
them an explicit *"NO EDIT"* placeholder submitted as a real edit, and three
adding warning-logging and message-renaming that the diagnosis never asked for.

**Round 2 plan — five edits.** The proposer **kept the two sound core edits
unchanged**, **dropped the no-op** the reviewers flagged hardest, and **tightened
its wording** — but **kept the same trio of off-path observability edits** (now
renumbered 3–5) and **still did not add the one thing both reviewers said was
missing**: a step to re-render the *already-blank* `guides-index` page after its
sections are repaired. Fixing the code stops *future* blank pages; nothing in
either plan un-blanks the page that is blank *today*.

That persistent gap is the honest reason the loop never reached *approved*: the
plan got cleaner each round but kept missing the same remediation, so the council
kept — correctly — sending it back.

### Council report — Round 1 (reviewing the first plan) → **revise**

**Edit-quality reviewer — `object`:**
- *Edit 3 (severity high):* "Edit 3 is explicitly a no-op ('NO EDIT') … every
  edit must change something real. A non-edit masquerading as an edit entry
  pollutes the plan and must be removed."
- *Edit 4 (medium):* adds a warning log to `classifyPagesForNav`, which "the
  diagnosis does not cite … as causing the blank page … an adjacent concern, not
  on the causal path."
- *Edit 5 (medium):* "exists solely to support the warning log in edit 4 … not
  on the causal path to the blank page."
- *Edit 6 (low):* renaming a success message is "a monitoring/observability
  change that is not minimal to fixing the diagnosed defect."
- *Notes:* "Edits 1 and 2 are sound and directly address the diagnosed causal
  path … The plan would be approvable if edit 3 is removed … edits 4 and 5 are
  removed … and edit 6 is either removed or demoted to a separate observability
  ticket. Core fix is correct."

**Pipeline guardian — `object` (no veto):**
- *Edit 3 (medium):* a declared no-op "submitted as an operation:'add' entry …
  creates a false artifact in the change record."
- *Edit 4 (medium):* adding a field to the `pageNavInfo` struct — "Blast radius
  must be confirmed — search for all `pageNavInfo{...}` literal constructions
  across the codebase before merging."
- *Edit 5 (medium):* `loadPagesForNav` "is shared infrastructure: every pipeline
  that triggers a nav rebuild … calls through this path. Adding a JSON unmarshal
  per page row on every invocation is a latency regression on all those
  pipelines for all sites."
- *Missing:* "The plan … does not remediate the current blank page. The two code
  fixes prevent future blank section-index pages but … the fix is forward-only
  and the live blank page remains." Plus: enumerate every `defaultSectionsForPage`
  caller before the new guard is safe.
- *Notes:* "none rise to architecture-change level … all concerns are resolvable
  within this plan's scope … provided the missing remediation path for the
  existing blank page is added."

### Council report — Round 2 (reviewing the reproposed plan) → **revise**

**Edit-quality reviewer — `object`:**
- *Edit 3 (medium):* "exists solely to enable the warning log in Edit 4 … Loading
  sections per-row on every nav rebuild is a real code change with a real cost,
  introduced to support a diagnostic warning that is not on the causal path …
  This is adjacent-path work, not causal-path work."
- *Edit 4 (medium):* "a monitoring addition, not a fix … not on the causal path
  the diagnosis established."
- *Edit 5 (low):* "addresses an observability gap, not the confirmed causal path."
- *Missing:* "the plan fixes the index page sections but contains no edit that
  triggers a rerender of guides-index after sections are populated … the existing
  guides-index page_components row count remains 0 and build_status remains
  planned unless a new work item is enqueued."
- *Notes:* "Edits 1 and 2 are sound … Stripping Edits 3, 4, and 5 would leave a
  more precise plan. The missing rerender trigger is the more serious gap."

**Pipeline guardian — `object` (no veto):**
- *Edit 2 (severity high):* "The ON CONFLICT WHERE guard assumes the sections
  column is jsonb … A text-vs-jsonb mismatch will cause a runtime SQL error on
  every applyNewPage call for all page types across all sites … a latent
  cross-site blast radius that must be resolved before merge."
- *Edit 5 (medium):* adding `output_fields` to a `complete_workflow` step "is a
  wire-format change … does not assess whether any other pipeline … reads
  complete_workflow step output envelopes … Blast radius is not fully mapped."
- *Edit 1 (medium):* component names must be confirmed to "exist in
  content_components.function for dartsonline.com … If none of the returned names
  resolve, the page is still blank and the fix produces no observable improvement."
- *Edit 3 (low)* and *Edit 5 (low):* a quiet `pageNavInfo` contract change; a
  workflow-JSON edit that must carry an explicit owning-pipeline field.
- *Notes:* "No veto. … The architecture boundary is not crossed. The
  high-severity objection on edit 2 … is containable: schema verification is a
  one-query check and must be done before merge. … All objections are resolvable
  without an architecture review."

### What the two reports show, together

The council is not a rubber stamp and it is not a dead end. Round to round the
plan measurably improved — the flagged no-op is gone, the wording is tighter, and
the two correct core edits held steady — while the reviewers stayed on the two
things that actually matter: **stop shipping edits the diagnosis didn't call
for**, and **repair the page that is blank now, not only the ones built next.**
The second never landed, so the loop honestly ran out its revise budget and
stopped at *exhausted* rather than talking itself into *approved*. A caveat for
whoever shows this: these are two rounds on one bug, and the round counting in
the currently-deployed image over-counts when a correlation already carries
review history (fixed in source, rides the next image build) — so a fresh clean
run is the fair benchmark.
