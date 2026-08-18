# NOTES — bugfix 309 (append-only, newest at the bottom)

## 2026-08-18 — session "bugfix bugs_open/297" picks up 309

- Asked to take 297; it is CLOSED (fixed+live 2026-08-17, mig 453, bugs_closed/).
  Swept 298–310 with `who-owns.py` + live-transcript greps: 298–308 owned by active
  lanes; 310's file is untracked in the tree (its filing session f7646672 still
  active on it); 309 filed by the 279/284/290 thread whose session (fca1cedb) fired
  a 090 (corr `df8ca3a1-9cca-474a-88fb-19577e088080`), wrote its handoff, and
  ENDED. 309 is the pick.
- **Bug re-validated at the served page** (~19:25 UTC): 200, 32,594 B, 6 bl-cards,
  0 anchors in each. Matches the filing exactly.
- 090 dispatch not yet visible in orchestration_states (payload query, 0 rows) and
  0 diagnosis_artifacts — consistent with the measured ~29 min publish→start
  latency; did NOT re-fire.
- `bl-card` appears in NO Go file — the markup is `content_components` data.
  The page's listing pc `79d769e4…` → component `blog-listing_pre_037`
  (`4b097683…`), `created_from='generated'`, born 2026-04-08, template stores
  anchors as `{{if .postN_url}}<a href="{{.postN_url}}"…>{{end}}`.
- `content_data` for that pc holds all six posts' titles/dates/excerpts/images and
  **no `postN_url` key at all** (65 keys, none `_url` except image URLs). So the
  template's `{{if}}` ate every anchor. Renderer behaved exactly as told.
- Schema declares `postN_url` `required:true`, `source: site_specs.blog.postN_url`.
  **`SELECT … FROM site_specs WHERE aspect='blog'` → 0 rows fleet-wide, all
  history.** The source has never been resolvable for any site since the component
  was born.
- `plan_sections_action.go`: `on_missing` defaults to `skip_field` (line ~2093) —
  required + skip_field ⇒ field omitted, section builds, structural miss recorded
  (`STRUCTURAL_KEY_CARRY_MISS` in agent_error_log — 28 rows fleet-wide, none for
  this page; the page's last builds may predate the mechanism or have gone through
  the merge-path rerender. Not load-bearing for the fix; the 090 may say).
- **The bug file's control does not control what it seems to**: the working
  `mortgagecalculator.co.uk/investor/index.html` uses a DIFFERENT component
  (`tool-list`), not blog-listing. So "the component is capable" was true of the
  card idiom, not of this code path. `blog-listing_pre_037` has plausibly never
  produced a working link anywhere.
- Second consumer `leopardessconsulting.co.uk/blog`: its pc points at the same
  component but its stored rendered_html (8,712 B) contains **zero bl-cards** — a
  plain link list; some other writer replaced it. Do not assume broken-the-same-way;
  verify before touching. [UNVERIFIED what serves live on that URL]
- **Census** (queries in RUNBOOK): 10 phantom `site_specs.<aspect>` vocabularies
  across 11 active components (blog, categories, inventory, legal, nav ×16 fields,
  pricing, product, search, social, social_proof — aspect exists on NO site).
  Live exposure: this page, leopardess/blog, and 3 pcs on gaswholesalers.com
  (testimonials, social_proof). 8 components dormant. Plus 7 declared `query.*`
  names the resolver's switch does not know (affiliate_products, category,
  category_posts, comparison_filter_types, comparison_results, featured_post,
  bare `pages`) — same silent-skip fate at plan time.
- The store gate (`store_generated_component_action.go`) has checks 1–4 (HTML
  structure, unclosed style, empty schema, legacy dialect) — and no source
  validation. That is the class gap. `recordValidationRejection` already exists as
  the feedback channel for refused generations.
- `content-listing` (manual, 2025-11-28, active) is the correct article listing:
  `articles` ← `query.blog_posts`, required, `on_missing: skip_section`, range
  loop. fundamentallyai has 9 `page_type='blog-post'` pages (8 active+deployed,
  1 archived) — `query.blog_posts` returns exactly the right 8 and drops the
  archived card-4 target naturally.

### Missteps so far

- My opening `ls bugs_open/ | grep -i 297` returned empty and I nearly concluded
  the file was missing; a combined `ls` of both dirs found it in `bugs_closed/`.
  The first command's empty output remains unexplained — treat a surprising empty
  grep as suspect and re-ask differently (grep-silent traps are a known family).

## 2026-08-18 later — the collision, the real 090, and the division of labour

- **Two sessions worked 309 in parallel for ~70 min.** Session "bugs_open/272"
  (transcript 24fed6ae; redirected onto 309 like this one was) measured the same
  mechanism (bfaf27f75, 19:35), verified fix candidate 1 (c15984e70), and recorded
  the 090 verdict (5b7ab5a66). Discovered via git log at my code commit. Logged in
  WRONG_CALLS (ownership snapshots expire); SendMessage to the session failed (name
  not reachable) so coordination went through the bug file: §8 CONTRIB (1ce6d7808).
- **The df8ca3a1 correlation names a run that NEVER dispatched** (their finding —
  explains my three 0-row polls; I stopped polling it). The REAL run is
  `6e578bf5-778a-4e72-aab2-0531e45c07d8`: verdict **CONFIRMED**, first iteration
  set, independently grounding the same chain (0-row blog aspect, all has_postN_url
  false, onMissing=skip_field). My council submission cites df8ca3a1 as "queued" —
  stale in that one line; the evidence base is unaffected. Correct it if a REVISE
  round happens.
- **Division:** they own the CASE repair (at the owner fork; their candidate 1 =
  migrate blog-listing_pre_037 to query.blog_posts — endorsed, and my guard blesses
  exactly that migration). This lane owns the CLASS guard: committed `0df9f1be9`
  (queryresolve map + IsKnownQueryName/KnownQueryBases + component_source_guard.go
  + 10 tests, all green against a clean `git archive HEAD` + my files).
- **Wiring still HELD:** store_generated_component_action.go carries the 303 lane's
  hunk depending on untracked platform/content/markup_balance.go. Committing the
  file before that symbol lands breaks HEAD. Whoever commits the file takes my
  wiring hunk with it — safe, the guard symbols are at HEAD already.
- Census reconciliation: their "61 silently-dead fields" = my 58 phantom-aspect
  + 3 junk-prefix. Same population, sliced differently. The 7 unknown query names
  are ADDITIONAL and only in my census.

## 2026-08-18 latest — wiring LANDED; the hold resolved by handshake

- 303 lane committed `markup_balance.go` at HEAD (`6d962bcf8`) and messaged the
  all-clear directly (cross-session message — so SendMessage works INBOUND even
  when my outbound by-name send failed earlier). Wiring committed at `e21b172f0`
  with their store-action hunk as a DECLARED passenger (their corr 70cf0da5 named
  in the message body, mine in the trailer). Combined file verified against clean
  `git archive HEAD` before the commit: build + tests green.
- Full fix now on the shared branch: `0df9f1be9` (guard + queryresolve map + 10
  tests) + `e21b172f0` (wiring). Inert until a chassis image rolls — Go changes
  need a build + roll. NOT requesting a roll for this alone (releases are
  whole-fleet, owner runs make release).
- Register updated: CLC-018 status BUILT+COMMITTED+WIRED, index row matched.
- Council `fdb032c6`: review_guardian EXECUTING_STEP as of 18:51:56Z.

## 2026-08-18 latest+1 — round 1 REVISE, answered by code, round 2 in flight

- Council round 1: **REVISE**, gating objection from editquality — "the wiring
  edit is HELD, so approving lands a guard nothing calls". TRUE when submitted;
  resolved in the repo before the verdict even landed (e21b172f0). A lesson in
  submission timing: the submission described the held state, events outran it.
- Three seats independently demanded a wiring-level test ("a mutation deleting
  the wiring line passes CI"). Closed at `e5c9029dc`: two sqlmock ACTION-level
  tests; the mutation was ACTUALLY RUN (wiring line no-op'd → test FAILED →
  restored → green). bug_historian's outage objection (fail-open = only a Warn)
  closed in the same commit: durable `SOURCE_GUARD_ASPECT_SET_UNAVAILABLE` via
  LogActionFindings, branch exercised by the second test.
- Measured answers (no code): Resolve has exactly 2 callers fleet-wide;
  check_required_fields_missing EXPLICITLY excludes resolver-sourced fields
  (no existing discovery check covers the class); ValidateSitePlan drops by
  component-NAME mismatch so it neither catches nor masks the 58 fields;
  STRUCTURAL_KEY_CARRY_MISS (28 rows) is the live signal for the out-of-scope
  registered-but-empty class per the RFC_009 owner ruling.
- Round 2 resubmitted on the SAME correlation (RESUBMIT_CORR=fdb032c6),
  submission JSON: COUNCIL_RESUBMISSION_2026-08-18_r2.json.

- `[VERIFIED 2026-08-18 ~20:30 local]` closing the earlier [UNVERIFIED]:
  leopardessconsulting.co.uk/blog.html serves 200 with **0 bl-cards and 11
  working hrefs in <main>** — matching its stored rendered_html. The second
  consumer is NOT visibly broken live; some earlier writer replaced its listing
  with a plain link list. Its pc still points at the phantom component, so a
  regeneration would re-enter the broken path — the birth gate (once live)
  refuses that, and the other lane's candidate-1 migration retires it properly.

## 2026-08-18 close — round 2 APPROVED (4 advisories, none high; 5 abstained)

Verdict read in full. Commits already carry `Council-Submitted: fdb032c6` and are
credited automatically by 098. Advisory triage, each answered or routed:

1. **guardian (medium): enumerate dispatchers of store_generated_component.**
   MEASURED: exactly ONE active agent_definition names it — `component-creator`.
   One pipeline, whose failure path has handled pre-store refusals since checks
   1–4 existed. Blast radius of the new refusal = that one lane.
2. **bug_historian (medium): enumerate OTHER writers of content_components
   input_schema.** MEASURED: 14 Go files INSERT/UPDATE content_components; only
   TWO touch input_schema — the gated store action, and deploy_tool_action.go
   whose write is `SELECT … input_schema … FROM content_components WHERE id=$3`
   — a verbatim COPY of an existing row (propagation, cannot introduce a new
   phantom). Introduction paths = generated route (gated) + hand SQL
   (LANDMINES entry; CLC-015-style table check is the escalation).
3. **bug_historian (medium): fail-open still STORES the component.** True and
   deliberate (availability over strictness for a transient read error), now
   with the durable row. If SOURCE_GUARD_ASPECT_SET_UNAVAILABLE ever shows a
   sustained streak, flip the default — one-line change, recorded here.
4. **bug_historian (low): does component-creator swallow the refusal?** NOT
   verified this round — routed as the lane's open follow-up (check the 017
   unhandled-error shape against a live rejection once the guard rolls).
5. **editquality (medium): second wiring test asserts INSERT happens, not row
   content.** Fair; tightening = WithArgs/content match on the finding INSERT.
   Advisory, not done tonight — noted as test-debt.
6. **debug_historian (medium): my stated deploy check led with the provenance
   LOG line, which SCROLLS on agent-chassis.** Corrected in CLC-018's
   verify-later: binary probe with present+absent controls is the primary.
7. **architecture/reuse (low): five bespoke Layer-1 guards, consolidation owed.**
   Already in CLC-018 relations as the known refactor direction.
