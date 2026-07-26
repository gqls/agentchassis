# Handoff — a FAILED page build leaves the page `deployed`, partially composed, and invisible to the reconciler

> **STATUS: CLOSED 2026-07-24 — fixed at every layer, LIVE (guard v1.0.1146, skip
> persistence v1.0.1155), and behaviourally verified end to end on the original
> page.** dartsonline/index: `deployed`, stamped with the CURRENT plan,
> `suppressed_sections=["product-grid","testimonials"]`, 4 planned-effective = 4
> rendered, honest live page; gaswholesalers/index healed the same day via the
> legacy `site_specs.site_plan` path (`["social_proof"]`, 6=6, stamp NULL-safe).
> Deploy guard: 4 real pre-fix catches, zero false positives, zero gaps. Drop
> cause decomposed — `bugs_closed/041` (bulk), the `on_missing=skip_section` data
> guard working correctly (now durably recorded), `bugs_open/045` (tool trio,
> owned there). Council `164058e6` APPROVED. **Residuals, each pointed at its
> owner, none this defect:** ~25 pre-fix short
> pages = per-site backfill (041's residual, site owners); the 364-byte
> `category-listing` shell = `bugs_open/039` residual with its own `empty_section`
> item; deferral policy (should `needs_human_review` hold the deploy stamp?) =
> flagged to council, owner-level question, current behaviour unchanged.
>
> **CANDIDATE 2 IS LIVE 2026-07-26 (v1.0.1170, re-confirmed v1.0.1171) — but its
> BEHAVIOUR IS NOT YET INDUCED.** Pod-grep is discriminating and passes
> (`no error_message literal` → 1, control `build is short of its plan` → 1, negative
> → 0). No work item has been stamped `failed` by `mark_item_failed` since the roll,
> so the flat blank-error census is **absence of traffic, not evidence** — do not read
> it as verification. An induced-fault probe is built and ready but has failed to
> dispatch five times for reasons unrelated to this fix.
> **CONTINUE HERE →**
> `docs/agent_docs/docs024_key_docs_latest/bugfix_040_partial_build/HANDOFF_2026-07-26_continue_here.md`
> (probe IDs, the fire command, what has been ruled out, next hypotheses, cleanup SQL).
> A verification debt, not a defect; nothing is blocked on it.

> **CANDIDATE 2 DONE 2026-07-25 (`43002d3a4`), INERT until the next image roll.**
> The last unowned residual is built: `update_work_item_status` now records the
> routed step error on the work item when the workflow supplied no literal. See
> "Candidate 2 — built 2026-07-25" below for the census, the rule and how to
> verify it live. Council `2cafc4e0` submitted. The case stays CLOSED — this is
> the reporting half, not the defect; the defect (a partial build stamped
> `deployed`) has been fixed and live since v1.0.1146.

> **NUMBER COLLISION (2026-07-20, same day):** another thread filed a different
> `040` (`kafka_dial_timeouts_fleetwide_intermittent`, which cites itself as
> **040-kafka-dial**). Numbers are never reassigned — resolve by slug
> (`bugs_closed/README.md` duplicate-numbers table). Cite this case as
> **040-partial-build**.
>
> The two are related in substance, not just in number: 040-kafka-dial is a
> plausible *cause* of the spawn timeout that triggered this case (see "What
> happened" below), and `/bugs_open/003` is the mechanism connecting them.

**Filed 2026-07-20.** Found by asking the framework to rebuild a page through its own route
(rather than hand-restoring it) while verifying `/bugs_open/001` — which is exactly what exposed it.

**The work item says `failed`. The page says `deployed`. Nothing reconciles the two**, and because
the page also gets stamped with the current plan version, the reconciler will now *skip* it forever.

---

## FIX APPLIED 2026-07-21 — candidate 1 (structural), LIVE + BEHAVIOURALLY VERIFIED

**Status: FIX LIVE (v1.0.1146, still live v1.0.1149) and VERIFIED FIRING IN PRODUCTION on real
partial builds. 040 STAYS OPEN** — candidate 1 (the reconciler-invisibility defect) is done, but the
section-DROP root cause (the "partially composed" half) is separate and still UNKNOWN. This is the
structural fix: a page must not be stamped `deployed` + `built_from_plan_version` unless every
planned section was written.

> ### VERIFIED LIVE 2026-07-22 (v1.0.1149) — the guard is firing on real partial builds
>
> Four things now proven, end to end:
>
> 1. **Live in the binary** (v1.0.1149): `strings /app/agent-chassis | grep -c "build is short of
>    its plan"` → 1.
> 2. **Fired in production on 4 real partial builds.** Four pages carry the guard's exact DB
>    signature — `build_status='needs_rebuild'` **and** `built_from_plan_version IS NULL` — set since
>    the fix went live, and every one is genuinely partial (`1 ≤ pc_rows < planned`):
>    `fundamentallyai.com/index` 5/6, `gaswholesalers.com/index` 6/7,
>    `finetuning.uk/ai-agent-roi-estimator` 1/4, `ai-agent-orchestration.com/agent-complexity-estimator`
>    1/4. **Attribution is airtight:** only two code sites clear the stamp — the 0-component guard
>    (fires at `pc_rows=0`) and this partial guard; every other `needs_rebuild` writer
>    (`check_unresolved_sections`, `store_generated_component`, `flagPagesForRebuild`) leaves
>    `built_from_plan_version` intact. `pc_rows ≥ 1` on all four rules out the 0-component guard. So
>    these are this guard, and nothing else.
> 3. **The bug, observed whole.** `fundamentallyai.com/index` and `gaswholesalers.com/index` each have
>    a `needs_page` work item marked **`complete`** (the handler reported success on a partial build) —
>    yet the *page* is `needs_rebuild`. That is the entire defect: the item lies, but the page state is
>    now honest, so `decideEmit` returns `not_built`/re-emits instead of `skip_built`. Under the old
>    behaviour the page would be `deployed`+stamped and lost forever.
> 4. **No false positives, no gaps.** The fleet sweep is stable (28 → 27), so the count method did NOT
>    wrongly flip healthy pages (the naming-divergence pages like `gaswholesalers.com/services` 4=4 are
>    untouched); and **zero** deployed-short pages have a `deployed_at` after the fix went live — i.e.
>    no new partial build slipped past the guard to `deployed`. The 27 remaining are all pre-fix
>    backlog, healing as they are next built.
>
> The only thing not captured is the ephemeral log line (pod restarted 1146→1149, logs reset) — but
> the DB signature is the more durable proof and every other writer of it is ruled out above.
> **To re-check:** the guard-signature query is
> `SELECT … FROM pages WHERE build_status='needs_rebuild' AND built_from_plan_version IS NULL AND
> updated_at > '<fix-live time>'`, then confirm each hit has `1 ≤ (count page_components) < (count
> sections − suppressed_sections)`; and the no-gap query is the fleet sweep below with
> `AND deployed_at > '<fix-live time>'` (must return 0 rows).

> **It went live faster than expected — via another session's build, not mine.** I edited
> `v3_site_actions.go`, and before I committed it a concurrent add-all sweep committed it into
> `fe2ba5e52` ("v1.0.1146 - sweep"), which then built and deployed the chassis image. So the Go
> source shipped in **v1.0.1146** without my ever running a build — the exact
> "committed code rides ANYONE's next sweep build, and builds are from HEAD" lesson. **Verified live
> against the running pod binary** (not the tag): `strings /app/agent-chassis | grep -c "build is
> short of its plan"` → 1, `"planned sections rendered"` → 1, positive control `"no rendered
> components"` → 2. This commit carries only the **test + docs**; the Go source is in `fe2ba5e52`.
>
> **What is proven vs not.** LIVE in the binary ✓; decision logic unit-tested ✓
> (`page_section_shortfall_test.go`: refuse 6/5, allow 4/4, allow 5>3, allow 0/0); SQL live-validated
> ✓ (dartsonline → `testimonials` missing; 28 pages fleet-wide; `gaswholesalers.com/services` 4=4 NOT
> flagged). **NOT yet proven: an actual live partial build being refused** — the guard is passive
> until a build with a shortfall reaches the deploy mark, and none has since the 12:15:20Z roll
> (`kubectl logs <chassis pod> | grep "short of its plan"` → 0). This is deployment, not correctness
> (the "verify the failing branch" rule). To catch the first natural firing:
> `kubectl -n ai-persona-system logs -l app=agent-chassis --since=1h | grep -i "short of its plan"`.
> A forced proof (rebuild a known-partial page and assert it flips to `needs_rebuild` with the stamp
> cleared) is deliberately NOT done here — it dispatches into the contended queue (`bugs_open/030`,
> ~30-min latency) and, for a deterministic-drop page, starts the rebuild churn described below.

**Where.** `platform/orchestration/actions/v3_site_actions.go`, `UpdatePageStatusAction` (the
`update_page_status` action — the `update_status` step that every deploy path runs *before* the
git deploy). The existing "Option B" guard already refused a **0-component** deploy; this extends
the same choke point to refuse a **partial** deploy. New helper `pageSectionShortfall` returns
`(planned, rendered)`; the guard flips the page to `needs_rebuild` and clears the stamp when
`rendered < planned`, exactly like the 0-component path. Test:
`platform/orchestration/actions/page_section_shortfall_test.go`.

**The load-bearing decision: count ROWS, not section names.** My first instinct was to match each
planned section name against `page_components.slot_name` / `content_components.function` (the
`/bugs_open/039` Part-1 pattern, and candidate 3's suggestion). **I validated it fleet-wide before
shipping and it was WRONG** — it flagged **74** deployed pages where the count method flags **28**.
The extra ~46 are false positives: `pages.sections` names do **not** reliably equal the live
`slot_name`/`function`. Concrete, from the live DB 2026-07-21:

- `gaswholesalers.com/services` — planned `["services-hero","services-grid","features","call_to_action"]`
  vs live slots `["hero-services","services-grid","features","call-to-action"]`. Two names diverge —
  **word order** (`services-hero`↔`hero-services`) and **underscore vs hyphen**
  (`call_to_action`↔`call-to-action`) — on a page that is complete and serving 4 real components
  (1275/3473/4925/2151 bytes). Per-name matching would refuse it and drive it into a rebuild loop.

So the guard compares **row count** (`count(page_components) < count(sections − suppressed_sections)`),
which is the signal the fleet sweep below already validated. Trade-off, stated plainly: the count
method has a **false negative** the name method would not — a page with a duplicate component *and* a
missing section counts equal and slips through. That is acceptable for a deploy gate, where a false
*positive* (refusing a healthy page) is far more damaging than a false negative (the fleet sweep and
`incomplete_page_group`/`empty_sections` discovery checks are the other layers). See
`016b_debugging_guide` §9 for the transferable pattern.

**Suppressed sections are excluded** from the planned count, so a deliberately-dropped section
(`pages.suppressed_sections`) is never read as a shortfall. `sections=[]` pages (tools/blog-index
rendered by another subsystem — `/bugs_open/050`) have `planned=0` and are never refused.

**Positioning verified.** In all four deploy callers (`page-build-handler`, `page-rerender`,
`section-editor`, `tool-recreation-handler`) a component-writing step (`save_sections`/`apply_edit`)
runs *before* `update_status`, so a shortfall at the guard is real, not a mid-build race. This is the
same precondition the live 0-component guard already relies on.

**Fleet consequence now live.** ~28 already-partial deployed pages (the sweep below) will, on their
next build/edit/rerender (the fix is live as of the 12:15:20Z roll), be refused and flipped to
`needs_rebuild`, so the reconciler re-emits them. That is the intended healing. **Caveat (honest):**
if the underlying section-drop is
*deterministic* for a page (dartsonline dropped `testimonials` on two separate builds — root cause
still UNKNOWN, see the CORRECTED block below), that page will re-enter the build queue each reconcile
cycle rather than converge. That is a **slow drip gated by reconcile cadence + work-item dedup**, and
it is strictly better than a silent permanent five-sixths page — the loop is the visible signal that
the separate section-drop defect needs its own diagnosis. **Candidate 2** (propagate the orchestration
error onto the work item's `error`) is NOT done here — it is now smaller than filed (the error is
already durable in `agent_error_log` since v1.0.1140; it is a join, not a missing record) and is a
cheap independent follow-up.

**How to verify** (against a real build): see "How to verify a fix" below, plus re-run the fleet
sweep and confirm no *healthy* page (e.g. `gaswholesalers.com/services`, 4=4) was flipped to
`needs_rebuild`.

**Residual / next investigation (NOT this fix).** The fix stops a partial build being *recorded* as
a complete, plan-stamped success; it does **not** stop the partial build being *produced*. Why
`testimonials` is dropped by a build that reports complete is still **UNKNOWN** (see the CORRECTED
2026-07-20 block below — three hypotheses tested and refuted). Until that is found, the fix converts a
silent permanent five-sixths page into a visible one that keeps asking to rebuild — which for a
*deterministic*-drop page is a slow reconcile-paced loop. Finding the section-drop cause is now the
highest-value follow-on, and this fix is what makes it visible instead of silent.

> **CORRECTED 2026-07-24 — the drop cause is no longer unknown. It decomposes into three, and
> none of them is a new rendering defect:**
>
> **(1) The bulk was `bugs_closed/041` (section-lookup never normalised), already found, fixed and
> behaviourally verified by that workstream** — snake_case section names (`call_to_action`, 14 of
> this sweep's 39 missing occurrences, plus most `hero` cases) never matched their kebab components,
> fell to Path 3, were deferred, and the page deployed short. Fixed & live v1.0.1146. What remains of
> that damage is a per-site **backfill** (already-short pages need a rebuild), owned per site.
>
> **(2) dartsonline's `testimonials` — the case this file called UNKNOWN — was the
> `on_missing=skip_section` data guard working correctly.** Grounded live 2026-07-24: the
> `testimonials` component's `input_schema` requires `site_specs.social_proof.testimonials`
> (`required:true, min_items:1, on_missing:"skip_section"`), and dartsonline has **no current
> `social_proof` aspect at all** (`SELECT aspect FROM site_specs … is_current` — ten aspects, no
> social_proof). The platform refused to invent customer quotes. Why my 2026-07-20 check found
> nothing in `sections_skipped`: that record lives **only in the orchestration's `collected_data`**,
> and `orchestration_states` is pruned at ~24h (the `bugs_open/060` retention finding) — the record
> had evaporated before I looked. The refutation was of the *record*, not the mechanism.
> `bugs_closed/041` §"What is NOT this bug" had already stated this; I verified rather than assumed.
> Same mechanism grounded for `gaswholesalers/index`'s missing `social_proof` (no aspect) and — as a
> *deferral* variant — `fundamentallyai/index`'s `contact-info` (schema sources
> `site_specs.identity.email`; the live data holds `identity.contact.email` — a **source-path
> mismatch** → optional field, `on_missing:"needs_human_review"` → section deferred).
>
> **(3) The two 1-of-4 tool pages** (`finetuning.uk/ai-agent-roi-estimator`,
> `ai-agent-orchestration.com/agent-complexity-estimator`) hold only the tool widget itself; the
> planned `hero-tool`/`tool-guide-intro`/`tool-cta` trio was never rendered — that is the
> `bugs_open/045` tool-hero library gap (+ TP-004 family), **owned by the hero_tool_component_045
> workstream**; the generic `hero-tool` now exists and this guard has already queued the rebuilds.
> Contributed there, not forked.
>
> **The genuinely new structural finding — filed for diagnosis 2026-07-24, corr `65103331`:**
> a *correct* skip/defer decision is **never durably recorded** — `handleMissingField` sets it
> in-memory, `sections_skipped`/`sections_deferred` land only in prunable `collected_data`, and no
> code path writes `pages.suppressed_sections`. So `pages.sections` permanently over-promises, and
> this file's own guard (which subtracts only `suppressed_sections`) reads every data-gated section
> as a shortfall — **a page whose plan contains one can now never be stamped `deployed`**, however
> many times it rebuilds (gaswholesalers/index and fundamentallyai/index are parked this way today;
> the loop is slow — reconcile-paced + anti-churn — not runaway). The fix direction (persist the
> skip on the page row so the guard can exclude it; decide separately whether a *deferral* should
> hold the page un-stamped) awaits the diagnosis verdict before any code moves.
>
> **Diagnosis outcome (2026-07-24): two runs, then a human completion of the trail — mechanism
> CONFIRMED.** Run 1 (corr `65103331`) completed its 5-iteration evidence gather but **lost its
> verdict to `bugs_open/003`** at `call_diagnoser` (instance appended to 003). Run 2 (corr
> `f9bcee6f`) completed and returned **UNVERIFIABLE** — not a refutation: its `needed_evidence`
> names exactly what its code index could not serve, and the loop's own instruction is "Hand to a
> human with the full trail; do NOT auto-conclude." The three open items, answered from direct
> reads in this thread:
> 1. *"Need `UpdatePageStatusAction`'s body"* — read here: the guard compares
>    `count(page_components)` vs `count(sections − suppressed_sections)` and clears the stamp
>    (`v3_site_actions.go:635–700`, shipped `fe2ba5e52`, live).
> 2. *"`handleMissingField` returned no symbol match"* — it is a **closure** inside
>    `plan_sections_action.go` (`:1312`), invisible to the top-level symbol index (index gap noted
>    to the fixloop workstream). Its body: `case "skip_section": shouldSkip = true` — in-memory;
>    `sections_skipped` reaches only the action result (`:845–846`).
> 3. *"Counterexample: `tools`/`shop-index`/`brands-index` are `needs_rebuild` with a NON-null
>    stamp"* — reconciled: only the two guards in `UpdatePageStatusAction` clear the stamp (grep:
>    exactly two `built_from_plan_version = NULL` writers); the other `needs_rebuild` writers
>    (`check_unresolved_sections`, `flagPagesForRebuild`, `store_generated_component`) leave it.
>    Those pages were flipped by other writers — consistent, not contradictory.
>
> The run also surfaced **new supporting instances**: contact pages planned
> `[hero-contact, contact-info, contact-form]` holding 2 of 3 rows (`contact-info` absent —
> the deferral variant), parked `needs_rebuild` + NULL stamp by this guard. Fix goes to the
> council next (skip-persistence into `suppressed_sections` at the decision point; deferral
> policy flagged as a reviewer question).
>
> **SKIP-PERSISTENCE FIX BUILT & COUNCIL-APPROVED (2026-07-24). Inert until an image roll.**
> Council corr `164058e6` — **APPROVED round 1**, 4 advisory objections, none high-severity.
> Commits: `e3bca5b35` (core: `persistSectionSkips` in `plan_sections` — `suppressed :=
> (suppressed − ready) ∪ skipped`, self-healing, merge SQL validated live in both directions;
> committed before the verdict, so no trailer — the corr is in its message body) and
> `88b8f2af0` (carries the earned `Council-Reviewed:` trailer; responds to bug_historian's
> low objection — a persist failure now escalates durably to `agent_error_log` as
> `SKIP_PERSISTENCE_FAILED` instead of being a bare log line, which would have reproduced the
> vanishing-record defect at the persistence layer itself).
>
> **One advisory deliberately declined, recorded here so it is not silently dropped:**
> bug_historian (medium) proposed provenance-tagged entries (`{"name":…, "source":"auto_skip"}`)
> now, while the column is empty, so a future *manual* suppression can't be silently un-suppressed
> by the auto-merge. Declined because tagged objects break the `jsonb ?` containment all three
> readers use today (each would need rewriting), there is zero manual usage to protect (0/306),
> and the comment contract at both ends names the auto-writer. **If anyone builds a manual
> suppression feature: that is the moment to namespace this column — this paragraph is the
> pointer bug_historian asked for.**
>
> **Verify once live** (after the next image roll — or another session's sweep build; check the
> pod, not the tag): (1) rebuild a data-gated page (gaswholesalers/index) → `suppressed_sections`
> gains `social_proof`'s section name, the 040 guard passes, page stamps `deployed` at N−1;
> (2) the parked pages heal on their next natural build; (3) `check_empty_sections` stops
> re-flagging the suppressed slots; (4) negative control: a page with a *genuinely* dropped
> section (no skip recorded) is still refused.
>
> ### VERIFIED LIVE 2026-07-24 (v1.0.1155) — the golden case, end to end, on the original bug page
>
> v1.0.1155 rolled at 16:29Z carrying both commits (pod-grep: `persistSectionSkips` ×5,
> `SKIP_PERSISTENCE_FAILED` ×1, positive control ×1). dartsonline/index (this file's own headline
> case, parked `needs_rebuild`) was nudged through the platform's own route — a `needs_page` item
> in the reconciler's exact shape (item `3226f2a5`, `created_by='bugfix-040-thread'`, reason
> `manual_verification_bugs_open_040_skip_persistence`; no forced dispatch, the normal loop picked
> it up). Result, ~30 minutes later:
>
> - item `complete`; page **`deployed`**, stamped with the **current** plan (`0fb05b75`);
> - **`suppressed_sections = ["product-grid","testimonials"]`** — the skip decisions, durably
>   recorded for the first time; plan 6 − 2 suppressed = 4 planned-effective, 4 rendered → guard
>   passed **honestly** (arithmetic exact);
> - live page: HTTP 200, `<main>` serves hero / info-card-grid / call-to-action + the 364-byte
>   `category-listing` shell (the `/bugs_open/039` legacy residual, separately tracked by its open
>   `empty_section` item — NOT this bug).
>
> **Both skips are correct, and the second one is a cross-fix composing as designed:**
> `testimonials` = no `social_proof` aspect (this file's known case). `product-grid` = `query.products`
> `required:true, min_items:1, on_missing:skip_section` and the site has no product data — it used to
> render 3,055 bytes of **LLM-invented products** because empty `query.*` results bypassed
> `required`/`min_items` entirely until `bugs_closed/054` (unguarded-range) shipped in v1.0.1149.
> 054's fix **armed the data gate**; this fix lets the honest skip **not wedge the page**. Composed
> result: dartsonline serves an honest homepage, its state says exactly what it is, and the two
> data-gated sections are recorded, discoverable, and self-healing if data ever arrives.
>
> Remaining verify items: gaswholesalers/index nudged the same way (item `b9cf6147` — also exercises
> the legacy `site_specs.site_plan` path, since that site has **zero `site_plans` rows** and the
> reconciler never manages it; stamp COALESCE is NULL-safe there); `check_empty_sections`
> non-flagging observable at its next discovery run; negative control already evidenced by the
> guard's four pre-fix live catches (the refusal branch is unchanged for non-suppressed shortfalls).

---

## What happened (dartsonline.com `index`, 2026-07-20)

A `needs_page` item for `index` was handed to the framework. The build wrote five components, then
its spawned child request timed out and the orchestration ended in `complete_error`:

```sql
SELECT status, current_step, collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id='5c930b26-cf9a-4779-a19d-1215c8ad1de6';
-- COMPLETED | complete_error | Request 57508ea9-…-74821de917b8 timed out after 3 retries
```

The item was correctly marked failed:

```
status        | failed
attempt_count | 1
max_attempts  | 3
error         | (empty)
result        | {"completed_by_step": "mark_item_failed",
                "completed_by_orchestration_id": "5c930b26-…"}
```

**But the page was left deployed anyway:**

```sql
SELECT build_status, sections, built_from_plan_version FROM pages WHERE name='index' AND site_id='5fe8785b-…';
-- deployed | ["hero","product-grid","category-listing","features","call-to-action","testimonials"]
--          | 5d438145-0d2d-4023-9761-98ab4d06318c
```

and only **five of the six** planned sections exist, each marked `build_status='deployed'`:

| position | function | bytes |
|---|---|---|
| 1 | hero | 2466 |
| 2 | product-grid | 3055 |
| 3 | category-listing | **364 — hollow** |
| 4 | features | 4079 |
| 5 | call-to-action | 2347 |
| — | **testimonials — ABSENT** | — |

`testimonials` is not suppressed (`pages.suppressed_sections = []`) and the component exists and is
active (`content_components`: `name='testimonials', function='testimonials', is_active=t`). It was
simply never written — almost certainly the request that timed out.

> ### CORRECTED 2026-07-20 18:00 UTC, after the v1.0.1140 deploy — two of my own claims were wrong
>
> **(a) "almost certainly the request that timed out" is REFUTED.** The page was rebuilt again at
> **15:18:17** by an `image-build-handler` item (*"Re-render index after its image asset landed"*,
> reason `image_landed`) which reported **`status='complete'`** at 15:19:02. That build **also
> produced 5 of 6 sections** — `testimonials` missing again, `category-listing` again a 364-byte
> shell. **A successful build drops the section too**, so the timeout never explained it.
>
> Also, the timeout's failing step was **`deploy_page`**, i.e. *after* component writing. It could
> not have eaten a component that the earlier steps were responsible for producing. I had asserted
> a cause that the step name alone contradicts.
>
> Two other hypotheses tested and also refuted: the `testimonials` template passes
> `sectionTemplateValid` (it contains `</section>`), so it is not being dropped by
> `loadComponentSchemas`' truncation skip; and it is not in `sections_deferred` or
> `sections_skipped`. **The cause of the dropped section is currently UNKNOWN** — that is the honest
> state, and it wants its own investigation rather than another guess.
>
> **(b) "the cause is one join away but not written where a human or a sweep would look" is TOO
> STRONG.** The failure *was* durably recorded, in `agent_error_log`, with the message, the step and
> the work item id:
> ```sql
> SELECT agent_type, step_name, error_message, work_item_id FROM agent_error_log
> WHERE orchestration_id='5c930b26-cf9a-4779-a19d-1215c8ad1de6';
> -- page-build-handler | deploy_page | Request 57508ea9-… timed out after 3 retries | f98331dd-…
> ```
> The narrower true statement: **`site_work_items.error` is empty**, so the work item does not carry
> its own failure, but the platform did record it in a queryable table (32,398 rows since
> 2026-04-02). Fix candidate 2 below is therefore smaller than written — it is a join, not a
> missing record. v1.0.1140 adds `platform/orchestration/agent_error_log.go`, a further writer for
> that table from `routeToErrorStep` / `notifyParentOfFailure`.
>
> ### And the core defect is now DEMONSTRATED rather than predicted
>
> After the 15:18 rebuild the page is `deployed` **and stamped with the CURRENT plan**
> (`built_from_plan_version = dcc7834e`, the second re-plan's id). So `decideEmit` now genuinely
> returns `skip_built` for it. **dartsonline's homepage is a permanent five-sixths page that no
> longer asks to be built**, and this time it got there via a build that reported success. That is
> the whole bug, observed end to end, without needing the failure at all — which makes the title's
> emphasis on "FAILED" too narrow. **A partial build reports success and strands the page either
> way.**

`category-listing`'s 364 bytes are an empty shell (empty `<h2>`, `href="#"`, empty grid) — the
`/bugs_open/039` empty-section shape, and there is a matching `empty_section` work item for this
exact page and section sitting `detected` since 2026-07-14.

## Why it is worse than a lost section

**The page is now invisible to the reconciler.** `decideEmit`
(`reconcile_site_plan_action.go:341`) returns `skip_built` when the page is `deployed` **and**
`built_from_plan_version == planID`. Both are now true. So a reconcile against plan `5d438145` will
skip `index` as finished — a five-sixths page that no longer asks to be built. The failure marked
the *item*, and the item is terminal; nothing will revisit the page.

Note the interaction with `/bugs_open/038`: a *new* plan would flip it back to `stale` and rebuild
it. So the page happens to be recoverable by re-planning the whole site — but only as a side effect
of a different defect, and not by any mechanism that knows this page is incomplete.

**Two secondary observations**, not the main defect but worth checking with it:
- `attempt_count = 1`, `max_attempts = 3`, status terminal `failed` — **no retry happened**. Whether
  the retry budget is honoured on this path is unverified; if it is not, that is a second defect.
- `error` on the work item is **empty**, though the orchestration recorded a clear message. The
  cause is one join away but not written where a human or a sweep would look. Same shape as
  `/bugs_open/034` (validation errors dropped with no durable record).

## Fix candidates

1. **A page must not be `deployed` unless every planned section was written.** At the end of the
   build, compare the written `page_components` against the plan's section list; on a shortfall,
   leave `build_status` as-is (or set `needs_rebuild`) and do not stamp `built_from_plan_version`.
   The stamp is the part that makes the damage permanent, so at minimum: **never stamp on a failed
   or partial build.**
2. **Propagate the orchestration error onto the work item.** `mark_item_failed` has the failing
   orchestration id in `result`; write its `__step_error.message` into `error` so the failure is
   legible without a manual join.
3. **A completeness check that compares plan to artefact.** `pages.sections` vs the page's
   `page_components` (matching on `function` — see `/bugs_open/039` Part 1) would catch this class
   fleet-wide regardless of which build path dropped the section. Worth checking whether
   `complete_work_item_verification.go` is the right home; it exists for saga no-ops of this shape.

Candidate 1 is the structural fix. 2 is cheap and independently worth doing.

## Candidate 2 — BUILT 2026-07-25 (`43002d3a4`), inert until the next image roll

**The reason was never lost — it just was not on the row anyone reads.**
`update_work_item_status` wrote `site_work_items.error` only from a **literal**
`error_message` in step config. A literal can only carry a *static* reason, so the two
steps whose reason is dynamic — `page-build-handler`'s and `image-build-handler`'s
`mark_item_failed`, reached via `error_step` — configure no literal at all, and the
column stayed NULL on exactly the path that actually fails.

**Live census 2026-07-25** (both queries run against `clients_db`):

```sql
SELECT status, count(*) AS items,
       count(*) FILTER (WHERE error IS NULL OR error='') AS blank
FROM site_work_items WHERE status IN ('failed','rejected','cancelled') GROUP BY 1;
-- failed 75 / 21 blank · cancelled 51 / 43 blank · rejected 5 / 0 blank

SELECT count(*) FROM site_work_items swi
WHERE swi.status='failed' AND COALESCE(swi.error,'')=''
  AND EXISTS (SELECT 1 FROM agent_error_log ael WHERE ael.work_item_id = swi.id);
-- 20 of the 21
```

All 20 are `page-build-handler` / `deploy_page` / `call_agent`, message
`Request <uuid> timed out after 3 retries` — **this bug's own shape**, most recently
2026-07-24 16:19Z. The 21st is a `needs_diagnosis` item with no log row at all.

**Why the join was always available:** `routeToErrorStep` (`coordinator.go:3261-3269`)
sets `collected_data["__step_error"] = {failed_step, message}` **and** writes the same
message to `agent_error_log`, in that order, in one call. One source, two destinations —
so reading `__step_error` cannot disagree with the log. `fail_work_item` already prefers
it (`load_work_item_actions.go:877-883`); `update_work_item_status` never learned the
same trick, so the two failure-stamping actions disagreed about whether a failed item
explains itself.

**The rule shipped:** no literal **and** status ≠ `complete` → record
`__step_error.message`, prefixed `step <failed_step> failed: ` unless it already starts
with `step `. A configured literal always wins.

- **Why the step name.** The timeout shape does not say *what* timed out; the action-error
  shape (`step validate_content failed: …`, 58 rows) already does. Prefixing only the
  former converges on the format the column already uses instead of adding a second one.
- **Why `complete` is excluded, and why that is load-bearing.** `__step_error` is never
  cleared once set, so a workflow that recovers from a routed error and *then* stamps the
  item complete would be handed a stale failure. The fleet census below found exactly one
  literal-less `complete` step, so this is a live path, not a hypothetical:

```sql
SELECT COALESCE(s.value->'config'->>'status','complete') AS status,
       (s.value->'config'->>'error_message' IS NOT NULL) AS has_literal,
       count(*), string_agg(DISTINCT ad.type, ', ')
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action'='update_work_item_status'
GROUP BY 1,2;
-- failed             | f | 2 | image-build-handler, page-build-handler   <- the fix
-- needs_human_review | t | 2 | page-build-handler                        <- literal wins
-- complete           | f | 1 | image-build-handler                       <- must NOT inherit
```

**Deliberately NOT done: backfilling the 21 pre-fix rows.** Their `agent_error_log` rows
are the record. Writing a message onto a row the platform did not write it onto at the
time manufactures evidence, and this file already carries one correction for reading a
record's *absence* as a mechanism's absence.

**Verify once live** (the roll makes it active; the test is unit-level so far):
1. `strings /app/agent-chassis | grep -c "no error_message literal"` → 1 in the running pod.
2. Wait for a `mark_item_failed` to fire naturally (or induce one), then re-run the blank
   census above: **the blank count must stop growing.** New `failed` rows should read
   `step deploy_page failed: Request … timed out after 3 retries`.
3. Negative control: a `needs_human_review` park must still read its literal
   (`page-build-handler no-op: …`), not a routed error.

Test: `platform/orchestration/actions/update_work_item_status_error_test.go`, 6 cases,
fixtures taken from the live shapes above.

### VERIFIED LIVE 2026-07-26 21:16Z — induced through a scratch workflow

Steps 1 and 2 above are **discharged**, on v1.0.1171, pod `agent-chassis-5b4456686c-s5fkc`.
A scratch agent (`scratch-cand2-probe`) runs `boom` — an `update_work_item_status` whose
`work_item_id_field` resolves to nothing with `skip_if_missing:false` — routed by
`error_step` to a literal-less `mark_failed`, then to a literal-less `mark_complete`
while `__step_error` is **still set**. Both arms in one run:

```sql
SELECT item_key, status, COALESCE(NULLIF(error,''),'<<BLANK>>') FROM site_work_items
WHERE item_type='scratch_cand2_probe';
-- scratch_cand2:a | failed   | step boom failed: failed to execute action
--                              update_work_item_status: work_item_id not found at
--                              input_data.does_not_exist and skip_if_missing=false
-- scratch_cand2:c | complete | <<BLANK>>
```

Orchestrations `2fe04703-df0d-4cd3-bac2-90f5eca44dce` and
`a0257bd3-8a6b-4709-9798-54194fa8b701`, both `COMPLETED` at step `done`.

- **The fallback fires** — pre-fix `scratch_cand2:a` would have been blank, because
  `mark_failed` configures no literal and the literal was the only writer.
- **The `complete` exclusion holds** — `scratch_cand2:c` ran *after* `__step_error` was
  set and still recorded nothing. This is the guardian's staleness concern, exercised
  rather than argued: without the `newStatus != "complete"` guard it would have inherited
  `boom`'s failure.

**The `step X failed: ` prefix branch is still `[UNVERIFIED LIVE]`.** It did *not* fire
here, and the item's text is misleading about that — the stored
`collected_data->'__step_error'->>'message'` **already** read
`step boom failed: failed to execute action …`, so `HasPrefix(errorMessage, "step ")` was
true and the prefix was correctly skipped; the fallback copied the message verbatim. That
confirms the design note above (action errors already carry the prefix) and narrows what
is left to the bare awaited-request-timeout shape, which needs a real `call_agent`
timeout. **Read `__step_error` in `orchestration_states`, not the work item, to tell the
two apart — the output of a prefix-if-absent branch is indistinguishable from the output
of no branch at all.**

Deployment re-confirmed with both controls in one command (a grep for your own new
literal is unfalsifiable without a positive control beside it):

```
"no error_message literal"        -> 1   # created by this change
"build is short of its plan"      -> 1   # positive control, 040 guard, live since v1.0.1146
"candidate two placeholder xyzzy" -> 0   # negative control
```

**Step 3 (the literal-wins control) — VERIFIED LIVE 2026-07-26 22:00Z.** It needed its own
induced run: the four live rows carrying `page-build-handler no-op: …` all predate the
roll, so they prove a literal was written when no fallback existed, which is not the
question. The probe was extended with a **PROBE B** arm — a `needs_human_review` step
*with* a literal, spliced in after `boom` so it executes while `__step_error` is set —
and all three arms passed in one run (orchestration
`e5907364-7dfa-44be-8f45-9ed85da32df4`, `COMPLETED` at `done`):

```
PROBE A | failed             | step boom failed: failed to execute action update_work_item_status: …
PROBE B | needs_human_review | cand2 probe literal: this park reason was configured, not routed
PROBE C | complete           | <<BLANK>>
```

B recorded its own literal with the routed error sitting in the same `collected_data`
(`__step_error.failed_step = boom`), so the `errorMessage == ""` guard holds: **a
configured literal is never overwritten.** The prefix branch was again correctly skipped.

**All three owed verifications are discharged. Candidate 2 carries no verification debt.**
Scratch harness (`scratch-cand2-probe` + three `scratch_cand2_probe` items) deleted.
Full record:
`docs/agent_docs/docs024_key_docs_latest/bugfix_040_partial_build/NOTES_040_partial_build.md`;
reproducible procedure in `RUNBOOK_040_cand2_probe.md`.

**One check in this list cannot be discharged as written.** Step 2's "the blank count must
stop growing" is not measurable: the `failed` population *shrank* 75 → 52 items between
the 07-25 baseline and 07-26 (blank 21 → 14) without a single new item, and **zero**
`failed` items have been stamped by a real handler since the roll. A falling count cannot
show "stopped growing", and a quiet census is absence of traffic. Record a **rate**, not a
count — or do what was done here and induce the path.

> **The handoff's "unsolved" dispatch problem was never a defect.** Five publishes looked
> dropped because `generic-requests-group` was stalled at a frozen offset of 105196; the
> handoff's own §5 lists the probe messages queued at 105197/105202/105204, immediately
> behind it. They ran normally when the lane cleared. When a dispatch looks eaten, check
> the *outcome* table before theorising about the cause.

### Council: APPROVED round 1 (`66d77d4d`), 8 seats, 2 medium objections — both answered

> The first submission (`2cafc4e0`) never reached a reviewer: it died at
> `persist_submission` in 7 seconds because `risks` must be a **string**, not an
> array, and `operation: "create"` is not in the allowlist (`add` is). I then read
> "queued" off a broken poll for an hour — see `WRONG_CALLS.md` 2026-07-25 and the
> `RUNBOOK_council_gate.md` trap bullet. Resubmitted as `66d77d4d`.

**guardian, medium — "the containment claim is empirical; check it, don't take it on
faith."** Their sharper version of my census: this is a *shared leaf action*, so if any
other active workflow calls it on a non-`complete` status without a literal and **not**
via `error_step`, that workflow could inherit a stale `__step_error` from earlier in the
same run. Run 2026-07-25, over all six edge kinds (`error_step`, `next_step`, `else_step`,
`then_step`, `on_success`, `on_failure`):

```sql
-- affected = update_work_item_status steps with NO literal and status <> 'complete'
-- reachable_via = every edge kind that points at them, fleet-wide
 image-build-handler | mark_work_item_failed | failed | error_step
 page-build-handler  | mark_item_failed      | failed | error_step
(2 rows)
```

Exactly the two steps, each reachable **only** as an `error_step` target — the condition
they said would move them to approve. So `__step_error` is set by construction wherever
the fallback can fire.

**The stronger fix they hinted at, deliberately NOT built:** have `routeToErrorStep`
record `routed_to: <errorStep>` alongside `failed_step`, and fire the fallback only when
`ExecutionContext.StepName == __step_error.routed_to`. That makes staleness *structurally*
impossible instead of *empirically* absent, and would survive a future workflow the census
cannot see. Not done here because it edits `coordinator.go` — shared plumbing, wider blast
radius than the defect — and because the shipped code should be the code the council
approved. **Whoever adds a third such step should build this rather than re-run the
census.**

**debug_historian, medium — no pod-binary verification stated.** True of the *submission*
(its risk note said "verification is unit-level so far"); the check was already step 1 of
"Verify once live" above. Keeping it explicit: verify against the **running pod**, never
git and never the image tag —
`strings /app/agent-chassis | grep -c "no error_message literal"` → 1, with
`grep -c "build is short of its plan"` → 1 as the positive control (the 040 guard, known
present since v1.0.1146; a discriminating grep needs a string the change *created* plus a
control that proves the grep works).

**reuse_agent, low — two inline copies of "prefer the routed error over the literal".**
Fair, and recorded as a follow-up. Worth noting they are not actually identical:
`fail_work_item` takes the message raw, with no prefix and no status gate, and additionally
routes on `isAIUnavailable`. A shared helper would have to be parameterised on all three,
so the extraction is less obviously right than the duplication looks. Owner: whoever next
touches either action.

**editquality, low — the `step X failed: ` prefix exceeds the minimal fix.** Accurate; kept,
because the bare timeout shape does not name what timed out, and 58 existing rows already
use that prefix. Recorded so the choice is visible rather than assumed.

## How to verify a fix

1. Force a page build to fail partway (kill the handler pod mid-build, or point a section at a
   missing component). Assert the page does **not** end `deployed`, and `built_from_plan_version`
   is **not** stamped.
2. Assert the reconciler then re-emits that page rather than skipping it.
3. Assert the work item's `error` is non-empty and names the real failure.
4. Re-run the fleet sweep below and assert it shrinks.

## Fleet sweep — this is NOT an isolated page (run 2026-07-20)

```sql
SELECT s.domain, p.name,
       jsonb_array_length(p.sections) AS planned,
       count(pc.id) AS rendered
FROM pages p JOIN sites s ON s.id=p.site_id
LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.build_status='deployed' AND jsonb_array_length(COALESCE(p.sections,'[]'))>0
GROUP BY 1,2,3 HAVING count(pc.id) < jsonb_array_length(p.sections);
```

**25 deployed pages across 6 sites are short of their plan; 39 sections missing in total; 4 pages
have ZERO components.** Worst offenders: `gaswholesalers.com/wholesale-pricing-explained` (7 planned,
0 rendered), `leopardessconsulting.co.uk/case-study-multi-agent-cost-control-platform` (4, 0), two
`finetuning.uk` blog posts (3, 0 each). `leopardessconsulting.co.uk` accounts for 11 of the 25.

> **Read this before chasing the list.** The query measures the **DB record**, not the live page,
> and the two can diverge — a page deployed months ago still serves its old file even if its
> `page_components` rows were later removed. So a row here is a **signal to triage, not proven live
> damage**. I checked one before believing it, and recommend the same for each:
>
> `gaswholesalers.com/wholesale-pricing-explained.html` — `deployed_at` 2026-05-06, 0 component
> rows. It serves **HTTP 200, 11,872 bytes**, which at first looks like a healthy page and a false
> positive. It is not: the body has **zero `<section>` elements, zero `<h1>`, zero `<h2>`, and its
> `<main>` contains 0 characters of text.** The 11.8KB is entirely header/footer/nav chrome. **A
> live, blank page marked `deployed`.**
>
> So for this one the DB signal and the live artefact agree.

### All 25 checked live (2026-07-20) — the split is 4 broken, 21 degraded

Every row above was fetched and its `<main>` text measured. **The DB signal is a poor predictor of
severity: being short by 2 of 7 sections can mean a totally blank page, or a perfectly serviceable
one.** Do not triage from the DB counts alone.

**Group A — 4 pages, genuinely blank and live** (`<main>` = 0 characters; all are the zero-component
rows). These are unambiguous damage:

| page | sections planned |
|---|---|
| `finetuning.uk/blog/chatgpt-has-your-data-does-that-matter.html` | 3 |
| `finetuning.uk/blog/what-is-rag-and-do-small-businesses-need-it.html` | 3 |
| `gaswholesalers.com/wholesale-pricing-explained.html` | 7 |
| `leopardessconsulting.co.uk/case-study-multi-agent-cost-control-platform.html` | 4 |

Two are blog posts with a title and no article. One is a pricing explainer with nothing in it.

> **CORRECTED 2026-07-20 by the site owner.** I wrote here that the leopardess entry was "a case
> study on a site that has no clients", so blank was the safer failure and it should be deleted.
> **The premise was wrong.** The owner's case studies describe **real work — systems actually built
> and running — they are simply not *client* case studies.** The distinction matters: "no clients"
> was read as "nothing to write about", which is false. That mistaken premise also appears in
> `/bugs_open/001`'s FRESH EVIDENCE section ("this site has no clients and no case studies") and
> should not be carried forward from there either.
>
> Checking the live site after the correction changed the conclusion completely — see below.

**The leopardess case-studies section is NOT broken and must not be deleted.** `/case-studies.html`
is `deployed`, in the header and footer nav, and serves 8,054 characters of honest, specific copy:
h1 *"Systems running in production, not proposals waiting for a budget."*, then four real systems —
Companies House record checking, news trust-ranking, no-code interactive tool generation, and "the
platform that built this website". That is exactly the owner's framing (real work, not client work)
and it reads well.

**What is actually wrong is a set of orphaned detail pages.** Eight `case-study-*` detail pages exist
in `pages`, and:

- **nothing links to any of them** — checked `/`, `/case-studies.html`, `/who-we-help.html`,
  `/use-cases.html`, `/how-it-works.html`, `/insights.html`, `/our-approach.html`: zero `href`s to
  `/case-study-*`. They are also `in_header=false, in_footer=false` and **absent from
  `sitemap.xml`** (27 URLs, none of them).
- **seven of them serve HTTP 200 with a completely empty `<main>`**; the eighth
  (`document-intelligence-pipeline`) 404s.

So the index tells the story inline and never needed detail pages, while seven blank shells sit at
live URLs reachable only by typing them directly. They are a tidy-up, not an outage — and deleting
them is the likely right call precisely *because* the section above them is already complete.
Rebuilding them instead would re-run the content writer over case-study material, which is the
`/bugs_open/029` fabrication path on the one site where that history is worst.

**Do not action the deletion without the owner's word** — this is live outward-facing content.

**Group B — 21 pages, short by one or two sections but serving real content** (`<main>` from 1.6k to
17k characters). e.g. `leopardessconsulting.co.uk/our-approach.html` renders 5 of 6 sections and
11,650 characters. These are degraded, not broken; each needs a per-page judgement about whether the
missing section mattered (a CTA and a FAQ are not the same loss). Eleven of the 21 are leopardess.

Caveat on the method: counting `<section>` tags is unreliable across templates — the four
`gamesdesign.co.uk/games/*` pages report zero `<section>` elements while serving 12k–17k characters
of real content in different markup. `<main>` text length is the trustworthy signal; the section
count is not.

## Related

- `/bugs_open/028` (slug `page_build_noop_reports_complete_and_deploys_borrowed_components`) — the
  closest relative: a no-op reporting `complete` and deploying borrowed components. **Distinct**:
  there the item wrongly said `complete`; here the item correctly says `failed` and the *page* state
  disagrees with it. Same family (page state not derived from build reality), different mechanism.
- `/bugs_open/003` — spawned children losing their response is the timeout that triggered this. That
  bug explains the *failure*; this one is about what the failure leaves behind.
- `/bugs_open/038` — why re-planning the site would mask this by rebuilding the page anyway.
- `/bugs_open/039` — the hollow `category-listing`, and the function/name matching needed for fix 3.
- `/bugs_open/034` — the same "error existed but was never written durably" shape.
