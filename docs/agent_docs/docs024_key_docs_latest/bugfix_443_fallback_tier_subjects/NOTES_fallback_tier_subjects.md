# NOTES — bugfix 443 fallback-tier subjects (append-only, newest at the bottom)

## 2026-09-02 — session "bugs_open/443" resumes the handoff

- **Ownership check:** `who-owns.py 443` → filing lane (finetuning_uk_service) active on the
  SITE, but the bug file is a handoff ("Filed as bugs_open/443 with fix candidates ordered" in
  their 2026-09-02 handoff). No dirty or recent-fix commits on
  `load_page_sections_from_spec_action.go` beyond apis.uk's tier-1 mechanism (`35905c547`).
  Messaged the live `finetuning` session; they confirmed handoff and sent five findings (below).
- **Bug still valid:** finetuning.uk `site_plans` rows = 0 `[MEASURED 2026-09-02]`; deciding
  arm unchanged at :508–515 (both `section_facts` and `section_subjects` gated on
  `specSource == "site_plan_tables"`).
- **The mechanism this rides is LIVE as of today:** 638 applied 2026-08-26; **639 applied
  2026-09-02 by apis.uk, pod-verified** (`plan_sections` config carries both
  `section_facts` and `section_subjects` wirings — read from the live `agent_definitions`
  row, not the seed); 640 applied 2026-09-02 (second attempt; first correctly refused by the
  anchor guard); **641 NOT applied** (owner-read gated) → `sectionPlanItem.Subject` is stamped
  and writer-inert fleet-wide until it lands. Evidence: seed file APPLIED lines + live config
  query in RUNBOOK.
- **Census (mine, independent):** 6 plan-less real sites; 203 deployed pages
  (57/47/41/30/15/13 — reconciles exactly with the bug file's corrected figure); **11 pages
  repeat a component type** (finetuning 4, gaswholesalers 4, ai-agent-orchestration 3; other
  three sites 0). My query used `count(*) FILTER (WHERE EXISTS …)` so the 37 empty-sections
  pages were counted, not dropped (the finetuning session hit the CROSS JOIN LATERAL
  drop-trap and warned me; my shape dodged it by construction — noted, not virtue).
- **Finetuning session's five findings** (their message, folded into the bug file by them):
  (1) use `your-own-model` as the verification canary (verbatim "How it works" ×3, built
  2026-08-27); (2) do NOT link the owner's "very AI sounding" verdict (2026-08-25) to that
  copy (postdates it); (3) their census matches mine 11 = 4+4+3; (4) `section_facts` shares
  the gate — fix must carry both arrays; (5) brief-splitting is an option to TEST, not a
  recommendation (count-match unmeasured fleet-wide; misalignment is the failure the gate
  exists to prevent) — I am not pursuing it: same-guess objection as cross-tier alignment.
- **Their second message:** `our-position-on-ai` repeats with two NON-ADJACENT blocks →
  trigger is per-type count page-wide, not adjacency (my detector design already counts
  per-type, page-wide). copy_quality_two_stage lane independently concluded the fix belongs
  at the subject-publishing tier, not in any prompt. Parked stage-2 copy proposal `8003c51a`
  sits on `your-own-model` — whoever grades it needs the 443 context (their lane carries it).
- **Serve check, 11/11:** curled all 7 previously-unread damaged pages on gaswholesalers.com
  and ai-agent-orchestration.com with per-domain invented-URL controls (both 404 → 200s are
  real). All 7 repeat; verbatim-identical h2 pairs on 5 of 7 (case-study-kafka,
  containment-first, service-areas, supply-terms, wholesale-pricing); subject-level
  (near-verbatim) repetition on enterprise-reference-deployment and pricing-transparency.
  With finetuning's 4/4: **11 of 11 census-flagged pages serve real repetition; ≥8 verbatim.**
  The layout census is a confirmed damage proxy in every case tested.
- **Tier resolution of the 11:** all TIER 3 (aspect exists for all 6 sites but does not name
  these pages with sections; per-page membership query in RUNBOOK — the site-level
  "aspect present" query MISLEADS, first draft of mine did exactly that before the per-page
  join corrected it).
- **Detector bound:** 25 repeat-layout pages fleet-wide (14 plan-carrying + 11 plan-less);
  0 repeat-pages have plan subjects today.
- **Storage decision inputs:** `pages.content_direction` is free-form writer steering fed to
  the prompt (bug 025) — wrong home for a machine array; `page_spec` is admin/purpose
  metadata; object-form entries in `pages.sections` rejected (downstream expects strings —
  validate_plan's own normalise comment; `jsonb_array_elements_text` errors on objects).
  → dedicated sibling columns. 19 candidate writer files of `pages.sections` as of
  2026-09-02 → no CHECK constraint (would error them), no destructive trigger; read-guard +
  visible degrade + 5b repair.
- **No 090 run, substitution declared** (2026-07-31 ruling): the deciding arm is a five-line
  condition read directly, both censuses were derived twice independently (filing session and
  this one, identical numbers), and the predicted symptom was measured at the served artefact
  11/11. A third re-read is what 090 would buy; the council gate reviews the FIX. Queue
  checked: 0 `awaiting_diagnosis`, no open work items on the mechanism.

## 2026-09-02 (later) — implementation, council round, and the advisory answers

- **Council corr `b7c59309-1f70-448f-9d20-1c47ebf64196`: APPROVED round 1**, 11 advisory
  objections (headline "5 advisory — none high"), 8 seats abstained. Answers, each measured
  where measurable:
  - *editquality: does plan_sections even run for fallback pages?* YES — it is the
    page-build-handler step downstream of load_spec_sections regardless of which tier served;
    the demonstrating case is today's playground build itself (tier 3, 6 sections,
    ready_count 6, COMPLETED). The detector sees every build that plans sections.
  - *editquality: equality guard thin in the sketch* — implemented and MUTATION-PROVEN
    (equality→length mutation made the same-length-different-content test fail; recorded in
    the test file header).
  - *reuse_agent: double-fire with the planner-side code on tier-1 pages* — accepted and
    documented in the registry note: the planner code fires at PLAN-WRITE (once, pre-persist),
    mine at BUILD (per build); a tier-1 page with a subjectless repeat can legitimately draw
    both, and they correlate trivially on page_name. Different event, different remedy text —
    kept distinct deliberately.
  - *reuse_agent: why not a discovery check?* — named now: the finding needs the BUILD's own
    in-memory filtered triple (what the writer was actually handed), which a scheduled
    site-level sweep cannot reconstruct; precedent is the loader's own LOCKED_MERGE_SKIPPED
    inline entry. Registered in the finding-code registry either way.
  - *guardian: other consumers of the loader* — `[MEASURED 2026-09-02, live agent_definitions]`
    exactly ONE active agent type carries `load_page_sections_from_spec` in config:
    page-build-handler. (The tool-recreation-handler landmine concerns the SIBLING
    load_page_record path, not this action.)
  - *guardian: tier-1 regression from the guard change* — the 285-lane tests
    (merge + facts alignment on tier 1) and the 7 PBP-049 tests all pass UNCHANGED; they are
    the tier-1-unchanged assertion the seat asked for.
  - *guardian: column-name collision in flight* — `[MEASURED 2026-09-02]` grep across
    sql_for_agents: no other migration defines pages.section_subjects/section_facts (hits are
    the 328/330/362/598 wiring/prompt seeds and SUPERSEDED files).
  - *debug_historian: deploy verification* — pod capability probe added to RUNBOOK
    (present + absent controls, provenance stamp + merge-base).
  - *debug_historian: IF NOT EXISTS* — the actual migration file already uses
    ADD COLUMN IF NOT EXISTS (the sketch predated the file).
  - *architecture (both)* — on record: the seat holds that RFC_022's exception does not cover
    schema additions to a heavily-written table, and that aligned-or-absent + detector
    compensates for, rather than resolves, the writer-coordination gap; both routed into
    RFC_063 (the convergence question), which now also records that this pattern shipped
    first under the estate's after-the-fact review posture (2026-07-29 ruling 2).
- **Same-file passengers, both directions:** my LANDMINES entry rode another session's
  `b8547dfc2` (verified present in HEAD, nothing lost). My plan_sections commit carries
  session bugs_open/444's error-defer hunk (~:2746, their council corr c0990eb3) by their own
  proposal — verified coherent: complete comment + code, builds, full package tests green
  with it in the tree.
- Tests: full `platform/orchestration/actions` green; `cmd/config-key-audit` green (registry
  self-consistency required naming the retention window in the `why` — fixed).

## 2026-09-02 (close of session) — shipped state

- **Committed:** `dbb218a41` (fix + tests + registry + register + bug §8; Council-Reviewed
  b7c59309, APPROVED r1), `c03b82c61` (mig txn wrap), `42463ec61` (RFC_063 + sha stamp +
  owner log). HEAD verified building via verify-head-builds.sh after the main commit.
- **Migration 717 APPLIED 2026-09-02** by this session (hand-apply + --record-only, note in
  ledger; both columns schema-verified nullable jsonb). Column-before-binary satisfied — Go
  half inert until the next chassis roll.
- **Messages sent (all delivered):** bugs_open/444 (passenger committed; their
  LISTING_PAGE_HELD_NO_ITEM_SOURCE registry entry owed in their commit), finetuning (canary
  plan, recipe change, backfill offer), apis.uk (consumers-told, PBP-049 superseded line,
  641 acceptance population grew), copy_quality_two_stage (second subjects source; 8003c51a
  sits on the canary page), bugs_open/114 (RFC_063 pointer, IMG-078 named).
- **Open for the next session of this lane:** (1) post-roll pod probe (RUNBOOK, with
  controls) then Stage A on your-own-model — finetuning may run the backfill themselves,
  check their reply first; (2) Stage B after seed 641 lands (owner-read gated — not ours to
  apply); (3) RFC_063 awaits the owner; (4) watch the detector read-back query after the
  roll — first rows should be the 25-page cohort as their rebuilds occur.
- **Post-roll obligation added (copy_quality_two_stage request, 2026-09-02):** at Stage B,
  SAVE the before/after served HTML pair of `your-own-model` (curl both sides of the rebuild)
  and point to it from the bug file's close-out — their lane wants it as a natural experiment
  (three identical briefs → three subjects) on whether subject-scoping alone reduces AI-tell
  density. They cancelled proposal `8003c51a` (their call, reason on the item) so the canary
  rebuild has nothing pending under it.

## 2026-09-02 (night) — corrections in both directions with the finetuning lane; state as verified

- **Their catch (real, banked):** my RUNBOOK's shipped-probe originally leaned on literals
  that do not discriminate dbb218a41 — `section_subjects` is the RAILS literal (born
  35905c547, already rolled) and `without_subject` pre-exists in write_site_plan. Their pod
  measurement with both controls: `subjects_attached`=0 / `facts_attached`=0 (genuinely new,
  absent at dbb218a41^) are THE probes. RUNBOOK corrected (`d654b8196`). My fix is NOT in the
  running binary — expected, rides the next roll.
- **My corrections back (artefact-verified before asserting):** their "639 NOT applied (still
  _HOLD)" and "gate 1 on 641 not cleared" both invert — the _HOLD suffix never changes on
  apply (state = header APPLIED line + live row; wiring re-measured live minutes before
  replying), and 641's gate 1 gates on the RAILS image + 639, both satisfied (their own probe
  showed the rails literal in the binary). **641 waits only on gate 2 (owner read), which the
  finetuning lane is actively moving — it does NOT need my roll.** New LANDMINES entry for
  the _HOLD-filename trap (fired today, live instance; rode `021906ed8` as a same-file
  passenger — third passenger event today, all verified in HEAD, all benign).
- **Division of labour confirmed:** finetuning lane runs backfill+rebuild (their briefs) and
  reports Stage A as Stage A; I ping them at the roll. apis.uk ack'd (785848be9): PBP-049
  entry corrected 5 ways including a week-stale "untested gap" line; suite is 12 tests now,
  not 7. 114 lane's RFC_063 input anchored (`f49976afc`); copy_quality pair obligation stands.
- **Closed loop (finetuning, night):** both my corrections confirmed at the artefact by their
  own re-measurement; their WRONG_CALLS entry (`5984a84a1`) names the sharper mechanism — the
  instrument was the LEDGER (`schema_migrations` is silent on hand-applied _HOLD files BY
  CONSTRUCTION), and the disproving row had been in their own scrollback for hours. LANDMINES
  entry amended with the ledger arm + attribution at their request. Owner corrected by them:
  **641 now waits on the seven-line read alone** — unblocking tier-1 subject rendering
  fleet-wide, independent of my roll. Standing claim unchanged: dbb218a41 not in the binary;
  their 4 pages wait for the next build; backfill theirs.

## 2026-09-02 (late night) — both owner rulings landed, via the finetuning lane

- **RFC_063 DECIDED: option B** (converge the six sites; hand-insert permitted as a scoped
  exception — closed backfill, no new site can take the old path). Ruling verbatim + what it
  does NOT waive recorded in the RFC appendix (`01a3b96ac`, finetuning lane); status header
  updated to match. **Execution UNASSIGNED** — the first act is the one-site reconciler-skip
  measurement (decision-1 against real rows incl. the needs_rebuild 17 and empty-sections 37),
  NOT the insert. IMG-078 caveats stand (pair cookly with an asset-holder; seed imagery rows
  from `assets`). PBP-051 stays correct post-B: its arms serve any page a plan does not name.
- **641 gate 2 returned as REDRAFT, not approval — do NOT apply 641 as written.** Owner wants
  positive prompting (no negative instructions) written IN the language expected back, no
  production vocabulary, no specimen answer. Candidates + mechanics-to-keep in
  `finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`; his pick of
  framing pending. ⚠ mechanics list flags the sibling-subject range render as UNTESTED.
  **641 is the apis.uk lane's file (PBP-049)** — the finetuning session mis-addressed it to
  this lane; relayed to apis.uk with the draft pointer. **Stage B now depends on the
  REDRAFTED 641; Stage A unaffected.**
- **641 routing correction confirmed by the finetuning lane** (`d2d64edb1` — their draft doc
  and handoff now name apis.uk as owner, marked as a visible correction). Pipeline agreed all
  sides: owner picks framing → finetuning test-renders (incl. the sibling range) → apis.uk
  writes the final SQL in their file → owner reads the exact final words. All channels closed
  for tonight; this lane's next event is the chassis roll (then: ping finetuning, pod probe
  per RUNBOOK, Stage A).
- **apis.uk closed the 641 relay (`c9c9b75ec`):** DO-NOT-APPLY-AS-WRITTEN stamped in the seed
  header itself (control at the point of action — no session can now apply the voided text on
  hearing "the read landed"), PBP-049 status + handoff corrected, sibling-range render
  accepted as theirs as a pre-apply obligation WITH falsifier (a skipped sibling must DROP
  from the enumeration; missingkey=zero makes both failure arms silent — 443's shape one hop
  later, their own framing). Stage B chain is now safe end-to-end: no path applies old 641.

## 2026-09-03 — THE ROLL LANDED; fix verified in the pod; detector live-firing on sites the census never saw

- **Deploy verified at the artefact** (pod `agent-chassis-75b987cbd7-mqrnj`, fresh):
  `subjects_attached`=1, `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`=1 (the discriminating
  literals), present-control `section_subjects`=3, absent-control 0. The provenance stamp had
  already scrolled (startup-line trap, known) — the capability probe is the deciding evidence
  and the empty-sha guard refused a vacuous ancestry pass rather than faking one.
- **The detector is ALREADY WORKING, and it widened the known damage class:** 7 rows within
  ~2h of the roll (`occurred_at` 23:24Z–01:18Z), across FOUR pages on THREE sites — none of
  them in the plan-less six: leopardessconsulting.co.uk (case-study… ×3 rebuilds), seotools.co.uk
  (keyword-research-tools-compared, technical-seo-crawlers-compared ×2), vetcomparison.uk
  (how-it-works). So plan-CARRYING sites are minting subjectless repeats on fresh builds —
  either pages from pre-640 plans, or tier-2/3 pages on planned sites, or rule 17 not biting.
  `[UNVERIFIED]` which of those three; and serve-state of the four pages NOT yet curled — do
  that before calling them damage (invented-URL control per domain, as ever). Also unexplained:
  case-study… rebuilt 3× in 2h — possibly a retry loop, not investigated.
- **RUNBOOK read-back query was WRONG** (`created_at` — the column is `occurred_at`; errored
  on first live use). Fixed in place with the gotcha attached. That is the runbook doing its
  job one use too late.
- **Finetuning pinged** (owed ping delivered; note their previous session ended — bare-name
  send reached the NEW session under the same name): backfill is go, ~300s dispatch caution
  flagged, Stage A assertions restated.
- **09-03 mid-morning, from finetuning (new session):** backfill 3/4 done (playground,
  your-own-model, technical-details; UPDATE 1 ×3, read-back 6/6; `our-position-on-ai` NULL —
  no brief). **Canary switched to `technical-details`** (owner is reading the £99 page right
  now; your-own-model reserved for Stage B). Their phrasing CONTRIB is in this dir → RUNBOOK
  template gained the "You'll want to know ___" + distinct-subjects rules. **Path pin
  (theirs, load-bearing): only page-build-handler contains load_page_sections_from_spec — a
  spawn_content_writer "rebuild" cannot exercise the fix.** Their Edit-Mode caution CHECKED
  by me at the deciding arms: both loaders are opt-in (`mode=recreate` / `mode=edit_live`),
  so Stage B is sound iff the rebuild item leaves spec.mode unset — pinned in HANDOFF §Stage B.
  641: owner picked framing C; two render findings with apis.uk (sections_for_render missing
  from input_fields → silent empty sibling list; exclusion by subject); redraft now owned by
  the new framework_prompts_positive_voice lane. They also independently re-probed both
  v1.0.1355/1356 pods — my probe corroborated.
- **Canary dispatched — identifiers for Stage A** (finetuning, 09:35Z, their `59c27e9c5`):
  item `896bb245`, spec verified mode-less by BOTH lanes independently (keys: title, reason,
  source, purpose, sections, page_name, page_type, suggestion) ⇒ full regeneration. Queue
  position: behind adversecreditmortgage.co.uk's 22 triaged page-build items (09-01 16:25),
  ~4 min/tick — expect ~1.5h, that is fairness working (their lane's own landmine). THEY hold
  the watch and run the Stage A reads at terminal; if this lane gets there first, the writer
  orchestration row is `input_data.current_page.name = 'technical-details'`, created_at >
  09:30Z. Their RUNBOOK now carries "assert NOT (spec ? 'mode') before dispatching" as a
  standing line. Channel closed pending the build.
- **09-03 ~10:40Z — STAGE A PROVEN AT PLAN TIME, first-hand** (bug file §9 has the full
  read-out): my own query of orchestration row `28610ba3…` — tier `pages_table`, 6/6 subjects
  attached, three distinct subjects on the three gtb slots; detector 0 rows for the canary vs
  7 fleet-wide (demand control in the same breath). finetuning's report corroborated, not
  taken on trust. Remaining: writer-row carry at completion (their watch), then Stage A is
  fully closed and the lane waits only on 641 for Stage B.
- **09-03 10:40Z — STAGE A CLOSED** (their watch, full read-out in bug file §9 completion
  note): writer row `ce514ce0` carried subjects on all 6 slots identical to the plan;
  redeploy 200; served h2s repeat with varied wording as predicted — data-not-prompt split
  demonstrated. Stage B = 641 alone; **applier is now the framework_prompts_positive_voice
  lane (owner's routing)**. ⚠ banked trap: one correlation can cover a whole site's items —
  per-page reads key on writer orchestration_id.
