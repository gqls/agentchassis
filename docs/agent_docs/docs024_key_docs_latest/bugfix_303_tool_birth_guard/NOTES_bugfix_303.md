# NOTES — bugfix 303 (append-only, newest at the bottom)

## 2026-08-18 — session "303, bugs_open/298" picks the bug up

- Asked to take the next unowned bug, probably 303 or 298.
- **298 is TAKEN**: two live sessions (`612ff4f1`, `cba50c3a`) actively on it — 98/48
  `internal-linker` transcript mentions, both quoting the `load_candidate_pages` SQL.
  Checked via live `.jsonl` transcripts, not who-owns (who-owns is blind to uncommitted
  sessions and reported both bugs "OWNED or recently active" on the strength of the FILING
  commits alone — a false signal for this purpose).
- **303 is FREE**: the only session touching its symbols (`fb6dd47e`) is the filer's lane
  (`webdesign_tool_rebuilds`), whose transcript ends with both tools rebuilt by AVOIDING the
  guard ("the guard-avoidance clause… carried over, so bugs_open/303 didn't fire"). It worked
  around the bug; nobody is fixing the guard. Bug file itself says OPEN, UNOWNED.
- **Bug re-validated at HEAD** (files change beneath us): `toolTemplateValid`
  (plan_sections_action.go:1853-1867) and `balancedPairs` counting
  (component_write_guard.go:146-177, 219-230) are exactly as the bug describes.
- **Queue checked**: no open `site_work_items` targets the guard itself. Three open
  `truncated_component` items exist (gauntlet-round-record-vonc-com, info-card-grid,
  tool-llm-cost-calculator-ai-agent-orchestration-com) — those are the discovery check's
  genuine-casualty findings; my calibration must confirm they STAY flagged under the new
  predicate.
- **Fourth affected surface found, beyond the bug file's three**: the discovery check
  `check_truncated_component.go` (sweep + verifier + newestIntactVersion) uses a hand-mirrored
  copy of the pair list with the same raw counting. Its false-positive mode is worse in one
  way: the VERIFIER can refuse to resolve an item for ever, and `newestIntactVersion` can call
  every intact version damaged.
- Plan written: markup-context tag counting as a shared scanner in `platform/content` (leaf
  package, already imported by the birth path; importable by discovery_checks — kills the
  mirror). See PLAN_2026-08-18.

## 2026-08-18 (later) — built, calibrated, committed; the shared-file dance with the 309 lane

- **Census widened the bug**: grep for the counting idiom found SIX surfaces, five defective —
  the bug file's three, plus `check_tool_completeness_action.go` (advisory, tool-recreation flow,
  ALSO case-sensitive so `<SCRIPT>` escaped it entirely) and `store_generated_component_action.go`
  Check 2 (style-only, no length floor).
- Scanner + tests written first (`platform/content/markup_balance.go`); all tests passed first run.
- **Same-file collision, resolved by sequencing:** `store_generated_component_action.go` carried the
  309 lane's uncommitted wiring (their commit `0df9f1be9` explicitly held it "until 303's
  content.UnbalancedStructuralTags lands at HEAD"). Committing my hunk would have broken HEAD in one
  order, theirs in the other. Landed `platform/content` + all other rewiring in `6d962bcf8` (proven
  against HEAD+only-my-files in a worktree first), messaged their session; they committed the shared
  file at `e21b172f0` with my hunk as declared passenger, built against a clean `git archive HEAD`.
- **Calibration (full detail in the fix record in the bug file):** component_versions 264 rows —
  26 flagged by both, 0 flips; 121 comparative pairs — 1 block both, 0 disagreements;
  content_components 300 rows — new strictly ⊂ old (8 of 11); the 3 cleared rows hand-read, all
  CSS-comment mentions. **Two of them carry OPEN false-alarm `truncated_component` items**
  (`91007600` info-card-grid — its "restore intact v1" advice would have replaced a GOOD template;
  `6e2c9ebf` gauntlet-round-record — its "regenerate" advice risks fabrication per bugs_open/020).
  Left open; their verifier resolves them after the roll.
- The filer's ADDENDUM (output-a-script-tag tools, escaped close ⇒ imbalance BY CONSTRUCTION) is
  covered by the raw-text semantics; pinned as a test in the same package (follow-up commit).
- Council: `Council-Submitted: 70cf0da5-e91a-42f0-8dd6-0cb5710b51dc`. Register: CLC-019.
- **Still owed:** read the council verdict and act on REVISE/REJECTED; after the next chassis roll,
  verify per the fix record (add_tool with angle-bracket description; the two items resolving).

## 2026-08-18 (council round 1) — REVISE, and two of the objections were worth having

Verdict at 18:57: **REVISE**, gating objection from editquality — edit 6 (store_generated) was
contingent on the 309 lane's commit at submission time. It landed (`e21b172f0`) eight minutes after
the round went in, so three seats' objections (editquality/guardian/debug_historian, all on the
contingency) were moot by facts, not argument. The round's real products:

- **prior_art_librarian asked whether the repo already imports an HTML tokenizer — it DOES**
  (`golang.org/x/net/html`, used by claims.go and section_visible_text.go). I had not checked; the
  seat was right to ask. The checked answer (those callers parse assumed-WHOLE documents; a spec
  tokenizer's EOF-recovery normalises exactly the malformedness a truncation guard detects — e.g. a
  tag cut mid-open emits NO StartTag, losing a true positive the old counter caught) is now in
  `markup_balance.go`'s header (`0e21e3cf1`) so it is not re-litigated.
- **debug_historian asked for the pod-verification recipe** — added to the bug file's fix record
  with must-hit and must-miss controls.
- editquality's edit-8 point (a test file bundled inside another file's sketch) was declaration
  hygiene, fixed in the r2 plan; the feared build-break window never existed (mirror var and mirror
  test retired in the same commit).

Round 2 resubmitted under the same correlation (`RESUBMIT_CORR`), run `d353d5de`. Lesson upheld:
a REVISE round is cheaper than the defect it finds — two of five objecting threads produced real
improvements, and none required defending the design.

## 2026-08-18 (council round 2) — APPROVED, four advisories, all four acted on tonight

**APPROVED 20:17Z** ("4 advisory objections — none high-severity"). Actions taken:
- bug_historian (load-time drop still quiet): the `componentInfoFromRaw` Warn now carries
  `unbalanced_markup_context` + `ends_cleanly`; noted in the log line itself that the
  truncated_component SWEEP owns the durable finding (the work item the seat asked about exists —
  it is the sweep's, by design).
- editquality (check_tool_completeness change had no stated coverage): true — the action had NO
  test file at all. `check_tool_completeness_test.go` added: mention-tool passes, cut flags
  advisorily without failing the step, uppercase cut caught (the pre-303 case-sensitivity pinned).
- guardian ('<no value>' dismissal asserted not verified): read the deciding arms
  (fix_component_template_action.go:916/926/950/971) — exact-literal artifact counting for slot
  repairability, not markup balance. Dismissal stands, now on a read.
- debug_historian (provenance recipe inoperative on agent-chassis): recipe reordered — stamp
  first (with the scrolls-out-of-range caveat), known-value binary probe with both controls,
  THEN merge-base ancestry. **Caught my own false-miss probe while doing it** (grep the binary for
  my commit sha — misses on every later build that carries the fix): WRONG_CALLS 2026-08-18.

## 2026-08-19 — roll verified at the binary, bug CLOSED; two instrument defects in my own recipe

- Fresh chassis deployed: `v1.0.1314`, pods 07:52Z. The provenance log line had already rotated
  (as the LANDMINE predicts), so the stamp came from known-value binary probes: build commit
  `d3590ca46` (22:17 BST last night), present on BOTH replicas; `git merge-base --is-ancestor`
  ✓ for all three fix commits (6d962bcf8, e21b172f0, d71e8abc7).
- **Live behaviour:** two tools born post-roll (09:26Z, 09:39Z), zero `tool_birth_truncation_blocked`
  rows, 4 tool-generator calls (demand present). Neither newborn exercises the DIFFERENTIAL —
  both pass under the old counter too (measured over their stored templates) — so the differential
  rests on the pinned tests + calibration at the shipped code; stated plainly in the close-out.
- **My own recipe failed twice while I ran it, both instrument defects, both now in WRONG_CALLS:**
  (1) yesterday's false-miss probe (grep the binary for MY commit — misses on any later build);
  (2) the all-zeros must-miss control HITS legitimately in Go binaries — a control that always
  fires, published unmeasured. Recipe corrected in the bug file; LANDMINES entry refined.
  Also re-derived live what LANDMINES line ~8903 already said (grep -aoE splits the sha via
  maximal munch in Go's string table) — I grepped LANDMINES for the guard's symbols but not for
  the probe recipe I was following. Cheap lesson: grep it for the CHECK you run, not just the
  code you touch.
- The two false-alarm items' summaries now carry a `[FALSE ALARM per bugs_closed/303 …]` prefix
  (UPDATE 2 confirmed) so nobody acts on the dangerous original remedy text before completing.
- **Bug MOVED to bugs_closed/** (fixed AND live bar met). Register CLC-019 → LIVE+VERIFIED;
  MEMORY workstream line → CLOSED; LANDMINES entry → historical, workaround retired.
