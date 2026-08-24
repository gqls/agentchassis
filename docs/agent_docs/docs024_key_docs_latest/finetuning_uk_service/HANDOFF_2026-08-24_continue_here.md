> # ⚠ SUPERSEDED same day (evening) — read `HANDOFF_2026-08-24b_continue_here.md`. This file's state lines are stale (the page is LIVE, 377 is ROLLED); kept for the seeding detail, the 377 story and the morning traps.

# HANDOFF 2026-08-24 — register SEEDED + £99 REGISTERED + offer page DISPATCHED; next = verify the built copy, run copy-editor once, report to copy_quality

**COLD-START for the merged finetuning.uk lane.** Supersedes `HANDOFF_2026-08-18_continue_here.md`
(whose 08-19 delta remains the record of how 302 left this lane and what the copy-quality answer
said). Statistics of record: `RESULTS_2026-08-15…`. Market + positioning:
`RESEARCH_2026-08-18_competitive_landscape.md` (owner-ratified principles + the
"[UNVERIFIED AS PROMISE]" correction). Today's full technical log: NOTES 08-24 section.

## State, verified 2026-08-24 (~10:30 UTC)

| thing | state |
|---|---|
| Chassis | fresh build `0b262ed5e` live (both pods, read per-pod from the startup provenance line) |
| Register seeding | **DONE + VERIFIED at the live rows.** `content_direction` + `identity` superseded-and-reinserted whole (`source=operator:finetuning_lane_20260824`): positive-first exemplars + `example_phrases.how_to_use_these` guard, fact-first house rules, em-dash instruction retired, `key_differentiators` gains-framed with the offer lead at `[0]`, `formatted` regenerated after a clean round-trip control. All 4 phrases flagged by `brief_supplies_negation` item `5ff2355f…` are GONE from the current specs |
| Dead voice aspects | `tone_of_voice`, `voice_and_tone` retired (is_current=false, dated note). `voice` KEPT — it feeds `check_voice_tells.go` |
| evidence_base | facts `ft-price-99` (£99, owner 08-18) + `ft-market-anchor` (~$5,000) registered; writer_block names them. Without this the page could not have stated its own price |
| Offer page | `your-own-model` / `/your-own-model.html` created `planned` + `needs_content_page` item `gap_plan_new_your-own-model_<site>` (status was `claimed` 10:22:55 by the dispatch loop when this file was written — **read the item + page row for current truth, this line is a snapshot**). Brief in `spec.suggestion`: one subject per section (6 sections count-matched), bans "a real person checks every run", safe form "run by people, not left to a queue" included, only £99 + $5k anchor may be stated |
| required_links | `pages.content_direction.required_links=["/contact.html"]` set at page creation. [INFERRED] the builder may overwrite `pages.content_direction` — RE-CHECK after build, before copy-editor |
| bugs_open/001 | CLOSED (in `bugs_closed/`) — the 07-31 PLAN's "no new pages" constraint is EXPIRED; do not re-obey it |
| Consultations | copy_quality: answered 08-18 (+ apis.uk CONTRIB 08-23 with the how_to_use_these guard + per-section-subject findings — read it before touching exemplars anywhere). offer-analysis: STILL SILENT; we made the differentiator-[0] call ourselves |
| brief_supplies_negation item | `5ff2355f-de45-49f1-aa11-ba3e3b320f7d` still `needs_human_review` — substance addressed by today's seeding; left for the owner/next sweep to confirm clean. Do NOT hand-close |
| Owner calls open | playground booking shape · sample datasets · Stripe posture (Phase 1 payment link blocked on these) |

> ## DELTA, same day ~11:00 — the build RAN and was BLOCKED by a validator false positive; that is now `bugs_open/377`, FIXED + council-APPROVED (r1, `8dd767ed`), committed `9094bc65c`, INERT until a post-`9094bc65c` chassis roll
>
> The blocker: `placeholderPatterns`' bare `"your company"` convicting the hero line — 46/46
> of that pattern's recorded firings were ordinary prose, 41 of them this site since 08-03.
> The WRITTEN copy measured well (`llm_call_log 774ca9c5`): owner tells 0, exemplar lift 0/3,
> only £99 stated, unverified promise absent — but `rather than` ×6 + `X, not Y` ×3 survived,
> matching 1:1 the shapes the writer's inputs still demonstrate (an instruction is also an
> example). Round-2 de-demonstration applied to what this lane owns (brief, own voice line,
> unique_selling_points); fleet instructional text left for copy_quality/305 with the counts.
> Outcome CONTRIB delivered:
> `copy_quality_two_stage/CONTRIB_2026-08-24_from_the_finetuning_lane_the_exemplar_seed_outcome_and_the_brief_that_taught_the_tell.md`.
>
> **So step 1 below is now: WAIT for a chassis roll carrying `9094bc65c`** (ancestry check:
> `git merge-base --is-ancestor 9094bc65c <pod's stamped sha>`), **then reset item
> `gap_plan_new_your-own-model_…` from `needs_human_review` to `triaged`** (attempt_count 0
> stands) and let the 60s sweep rebuild. Then run step 1's verification as written — and
> compare the negation-tell counts against build 1 (the controlled test of the
> instruction-as-exemplar finding). The re-block signature if the fix did NOT ship: one
> `placeholder_text/your company` blocker on the hero line, in
> `agent_error_log(CONTENT_VALIDATION_BLOCKER_DETAIL)`.

## The next work, in order

1. **Verify the built offer page** (the watcher/loop may already have the terminal status):
   - status is NOT proof (016b): read `pages.build_status`, then the rendered sections
     (`page_components`), then the served page. bugfix-099 landmine: a FAILED step can show
     COMPLETED with `error` NULL — read `__step_error` in the orchestration row.
   - count negation tells on the BUILT copy (`copy_quality_two_stage/count_negation_tells.py`
     is the ready tool); bugs_open/305 says check the output, never assume the spec suppressed.
   - exemplar-lift check: none of the 3 new `characteristic` exemplars may appear verbatim.
   - claims: £99 present; "$5,000" anchor allowed; NO other figures; "real person checks every
     run" ABSENT; banned patterns in evidence_base regexes absent.
   - date the copy: `SELECT id, created_at FROM llm_call_log WHERE response_text ILIKE '%<a
     distinctive built sentence>%'` (copy-quality's §4 ask — attributable to a run).
2. **Run `copy-editor` ONCE deliberately** (experimental, nothing dispatches it, parks at
   needs_human_review — that is expected). Dispatch recipe is in
   `copy_quality_two_stage/HANDOFF_2026-08-23_continue_here.md` (their apply-path traps too):
   template `scripts/fire-internal-linker.sh`, needs `input_data.page_id` + `input_data.domain`,
   one JSON line to `system.agent.generic.requests`; readiness = deployment rollout.
   Re-check `required_links` first (see above).
3. **Report the exemplar outcome to `copy_quality_two_stage` either way** — CONTRIB into their
   dir, dated via llm_call_log. A seed whose exemplars carried the register is their cleanest
   positive evidence; one that did not is "more useful still" (their words).
4. **Phase 1 remainder**: Stripe Payment Link + concierge (owner calls above), legal/terms
   extension for customer training data. Offer-analysis reply: ingest if it ever lands.
5. NOT ours: 302 follow-ups (semantics ruling, envelope measurement, carrier re-enable).

## Traps current for this lane (fuller set: RUNBOOK §7–§9)

- The dispatcher's pre-query selects on `site_work_items.pipeline='build'` + status `triaged`
  (+ site not locked, attempts < max). `pipeline` has a column default that saved us today;
  `resolution_path` is NOT the field that matters.
- Item briefs travel in `spec.suggestion` (bugs_open/271: `content_guidance` was dead, now
  aliased; `purpose` is NOT read as the brief).
- This site has NO `site_plans` row — reconcile_site_plan is not its birth path; pages are born
  via `pages` row `planned` + `needs_content_page` (mirror `gapPlanWorkItem`: triaged/40/
  page-build-handler/item_key `gap_plan_new_<page>_<site>`).
- Exemplars ARE content to the writer unless `how_to_use_these` says otherwise; a brief must
  name one subject per section and the COUNT must match the slots (apis.uk CONTRIB 08-23 both
  addenda — including the RETRACTION arc; read before generalising from one build).
- evidence_base `writer_block` gates ALL numbers; a new priced page needs its price in facts[]
  FIRST or the claims machinery is against you.
- Chassis rolls often: re-read per-service provenance before ancestry claims; no orchestration
  dispatch within ~300s of a chassis pod start.
- All watchers: foreground-test the filter against a line that EXISTS (WRONG_CALLS 08-15/08-17).
