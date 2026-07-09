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
