# HANDOFF 2026-08-07 — brochure component library / 151 candidate 1 mid-rollout

**Supersedes `HANDOFF_2026-08-05_continue_here.md`.** Written same-morning as the
events it records; liveness claims verified at write time (~09:15 BST). The lane's
centre of gravity has moved: the camera/checker work in the 08-05 handoff is all
still true and DONE — this handoff is dominated by **bug 151 candidate 1**, which
was designed, built, council-vetoed on breadth, re-sliced, and is now HALF-DEPLOYED.

## 1. Where the lane is, in one paragraph

Candidate 1 (facts assigned to sections at plan time — the structural fix for
sibling sections independently restating the same verified facts) is **built,
committed, and live to exactly the Slice A boundary**: chassis `v1.0.1262` carries
all the Go (pod-verified on both replicas, added/removed/control greps), migration
`327` (nullable `site_plan_sections.assigned_fact_ids`) and seed `329` (planner
prompt: fact roster + object-form section entries) are **applied and recorded**
(2026-08-07, ledger notes name this lane and RFC_016). Seeds `328`/`330` — the
consumption half — are **deliberately NOT applied**: renamed `*_HOLD.sql`
(`54f36a9ae`) so a blanket `--apply` cannot ship them, because the council
REJECTED shipping both halves as one slice (hard guardian veto, corr
`902a8563-2200-4771-ac0f-55dab0839a02`; 6/11 seats approved; veto was breadth,
not correctness). The contract, blast radius, veto record and the two decisions
awaiting a human are in `architecture_review/RFC_016`.

## 2. What is LIVE right now (all verified this morning)

- **Chassis `v1.0.1262`, both replicas**: pod-grep `"assigned fact id matches no
  current evidence_base fact"`→1, `"fact assignment composed an empty writer
  block"`→1 (proves BOTH commits `b882d5abf` + `ff515351e`), removed symbol
  `extractSectionNames`→0 (negative control), `diagnose_persist_fix_plan`→10
  (positive control). Note: these greps take ~60-90s each on the binary; budget
  a long timeout, and grep's exit 1 on a 0-count is normal.
- **Migration 327 applied**: `site_plan_sections.assigned_fact_ids` jsonb exists.
  NULL = unscoped (every pre-existing row), `[]` = deliberately factless, array
  = scoped. Recorded via `--record-only`.
- **Seed 329 applied**: live `build-site-planner` prompt carries the "Verified
  Facts (evidence base)" roster block and RULES rule 17 (verified by position()
  read-back). Backup table `agent_definitions_bak_329` + snapshot exist.
  **From now, any replan of a site with a populated evidence_base may write
  assignments.** Nothing consumes them yet — that is the point of Slice A.
- **Slice B held**: `328_*_HOLD.sql` (page-build-handler `section_facts` wiring),
  `330_*_HOLD.sql` (writer prompt v4). The v4 plaintext for the human/compliance
  read is committed: `brochure_component_library/sql/page_content_writer_prompt_v4_2026-08-06.txt`.
- **Everything in the 08-05 handoff** (camera, checker, contact sheet, site
  content) — unchanged, still live. The Monday 08-10 cron first-fire check
  stands. Checker population unchanged as of last night: 7 sites, 7 flag-only
  gaps, 0 deletions.

## 3. The story so far (for a cold session; detail in NOTES 08-06 entries + RFC_016)

1. Designed from measured ground: the writer prompt injects ONE whole-site
   `writer_block` identically per section; neither planner referenced
   evidence_base; only fundamentallyai (9 overlap pairs, pool 15) and
   leopardess (pool 18) are fact-rich — 5 of 7 flagged sites are fact-blind and
   candidate 1 does NOT fix their textual near-duplication. Full design + the
   decisions-with-reasons: `PLAN_2026-08-06_151_candidate_1_fact_assignment.md`.
2. Built across four seams: planner may emit `{"name","facts"}`; validate_plan
   resolves objects and NORMALISES back to string `sections` + aligned
   `section_facts` (because `sync_pages_to_db` serialises the raw array into
   `pages.sections`, read fleet-wide as strings); `write_site_plan` persists;
   `load_page_sections_from_spec` surfaces facts ONLY from the authoritative
   tier; `plan_sections` (opt-in config key) attaches
   `facts_scoped`/`assigned_fact_ids`/`assigned_writer_block` per ready item,
   composed via the existing `composeWriterBlock`. 10 tests, 2 mutation-caught.
3. Council: REJECTED round 1, guardian veto on breadth (three fleet-wide prompt
   changes + 4 Go files in one submission). Acted on, NOT resubmitted:
   RFC_016 written; rollout re-sliced (A = emit-only, observe; B = consume,
   own round, piloted); bug_historian's real defect fixed in code — an
   assignment whose every ID is unknown now degrades to UNSCOPED with a durable
   `agent_error_log` row (`FACT_SCOPING_EMPTY_COMPOSITION`) instead of
   rendering the deliberately-factless branch.
4. This morning: roll verified, Slice A applied, Slice B _HOLD-renamed.

## 4. Open, in the order I would take them

1. **Observe Slice A on a real plan.** The planner writes assignments only when
   a replan runs on a site with facts. The acceptance site is fundamentallyai
   (pool 15, the motivating case). Before firing ANY replan: check open work
   items on the site (CLAUDE.md dispatch rule), and know that a full
   build-site-planner run also runs sync_pages/populate_nav/reconcile — built
   pages are force-preserved (`reconcilePlanWithRealised`), but treat it as a
   real action on the owner's demo site, not a free read. Then:
   `SELECT page_name, ordering, component_name, assigned_fact_ids FROM
   site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id JOIN sites s
   ON s.id=sp.site_id WHERE s.domain='fundamentallyai.com' AND sp.is_current
   ORDER BY page_name, ordering;`
   Judge: are assignments present, spread (no two sections sharing 3+), sane
   per section role, and do factless sections carry `[]` vs unscoped `NULL`?
   Record the read-out in NOTES + RFC_016 (it is Slice B's entry evidence).
2. **Slice B, when Slice A's output has been inspected**: human/compliance read
   of the v4 plaintext → rename the two `_HOLD` files back → council submission
   for the slice (fresh corr; cite RFC_016 §3 and the Slice A observation) →
   apply 328 then 330 → then the full acceptance: rebuild fundamentallyai's
   flagged pages through the normal path and re-run the census — fact-overlap
   pairs must FALL; the five fact-blind sites must NOT move. **Rebuild caution:
   `bugs_open/189`'s config half** (`slot_name_from` on the writer steps —
   owned by bug_backlog_clearing) must be applied before builds touch
   locked-row pages; verify its state before firing builds.
3. **RFC_016 §5 needs a human**: ratify the section-entry rule; approve the
   sliced order. Until then Slice B does not move.
4. **Monday 08-10**: first scheduled contact-sheet cron fire (`refresh.log`) —
   the 08-05 handoff's standing item.
5. **Unclaimed, small**: file the `imagery.sections` positional-keying latent
   defect as its own bug (RFC_016 §1 documents it; grep bugs_open first);
   `tool-guide-intro` stays deliberately absent (08-05 handoff §3.4).

## 5. Traps for a fresh session (beyond CLAUDE.md; the 08-05 handoff's list still applies)

- **The shared tree ran HOT on 08-06 and may still be.** My v3_site_actions.go
  edits reached HEAD inside ANOTHER lane's commit (`cb7b4d759`, bug 208) as a
  same-file passenger — so `git log` that file before assuming authorship. A
  third session's `LogActionEntry` refactor may STILL be uncommitted in
  `plan_sections_action.go` (symbols not at HEAD as of last night): if you must
  commit that file, extract their hunks (`git diff`, hunks matching
  LogActionEntry/agenterrors), `git apply --reverse`, build+test, commit yours
  by pathspec, `git apply` to restore — verified twice last night; patches in
  this session's scratchpad are gone, regenerate fresh.
- **Do NOT run `run-migrations.sh --apply`**: the pending list carries many
  other lanes' files (306-326 range) plus the two _HOLD slices. Apply by
  `psql -f` + `--record-only`, exactly as the ledger notes show.
- **Prompt edits: live row is truth.** The lane's v3 prompt file was STALE
  against live by two later patches (241 voice placeholder, 178 edit-mode
  block); v4 was built from a live dump. Any future writer-prompt work: dump
  live first (`...->'generate_content'->'config'->>'prompt_template'`), never
  trust a file copy. Verify em-dash census = 5 after any reapply.
- **Council practice**: a REJECTED-on-scope verdict is answered by recording +
  RFC + slicing, never by resubmitting with more measurements. A `Council-
  Submitted:` trailer on a commit means verdict UNREAD — whoever reads this
  owes the read (last night's watcher pattern works: poll
  `orchestration_states` by `fix_correlation_id`, then `diagnosis_artifacts`
  kind=`council_report` — column is `body`, not `content`).
- **`agent_definitions_bak_327/328/329/330` backup tables** exist (328/330
  created by their seeds only when applied — currently NOT). `bak_329` is the
  planner-prompt rollback: restore `default_config` from it if the planner
  starts emitting garbage assignments.

## 6. Commit / verdict trail (this arc)

`1fa31ffae` session-open notes · `b882d5abf` the build (13 files, PBP-037,
`Council-Submitted: 902a8563…`) · `cb7b4d759` (208's commit, carries my
v3_site_actions.go half as a passenger) · `db6f81f4c` notes · `d0e82dee4`
README · `ff515351e` verdict response (empty-composition fix + RFC_016 +
visible rollout corrections) · `54f36a9ae` Slice B _HOLD renames ·
(this handoff's commit). Council corr `902a8563-2200-4771-ac0f-55dab0839a02`
REJECTED r1 — full report in `diagnosis_artifacts`, acted-on record in
RFC_016 §4. Register: **PBP-037** (+ index row), status corrected post-verdict.
