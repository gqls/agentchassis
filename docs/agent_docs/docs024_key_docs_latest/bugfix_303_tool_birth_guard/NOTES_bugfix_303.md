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
