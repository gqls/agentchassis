# PLAN — direction reach into the build pipeline, and the direction-document drift guard

**2026-07-20. Status: PARTIALLY IMPLEMENTED same day (owner approved D2-as-gate
and the staged R sequencing).** LIVE: R0+R1 (classifier seed, snapshotted),
D1 (`DIRECTION_LEDGER.md`), D2 (`.githooks/commit-msg`, tested on all three
paths), D3 (`100_CHECK_direction_integrity.py`, first run ALL GREEN) + `101`
(R1's consumer). NOT started: R2 (gated on R1's numbers), R3 (needs an image
window), R4 (other councils), D4 (standards table). Owner asked for both plans
after the v19 constitution + mission seats went live on the two review councils.
Everything below is grounded in checks run against the live system on 2026-07-20;
each grounded fact is marked. Companion: `PLAN_concept_register.md` §Direction
gatekeepers (what is already live).

**Grounded facts this plan stands on (verified 2026-07-20):**
- There is **no `standards` table** in clients_db — `CTS-029`'s "destined to become
  `standards` rows" was never built. The flat files are the only source of truth.
- The build pipeline's strategic brain is **`domain-research-classifier`**
  (dispatched from `apply_adoption_plan_action.go:594` and
  `validate_composition_inputs_action.go:260`); its workflow steps include
  `read_site_specs → classify_and_extract → write_classification_spec /
  write_identity_spec / write_content_direction_spec / write_design_intent_spec →
  create_next_item`. A second active `site-classifier` type exists — implementation
  must confirm which types need the injection (likely both).
- All three non-archive constitution copies are **byte-identical** (sha256
  `18453e8cac84…`), and both mission copies are identical (`c6aa949edab8…`) — so
  blessing a canonical copy is a naming decision today, not a reconciliation.
- `.githooks/pre-commit` already runs a **two-tier precedent**: `check-secrets.sh`
  is a REAL GATE (non-zero exit stops the commit — "that is the whole point");
  `commit-scope-report.sh` and `pattern-check.py` are ADVISORY (wrapped, never
  block). A direction-doc gate slots into the real-gate tier with an existing
  justification pattern.
- The consumer trap is real: `bugs_open/023` proved findings routed to
  `needs_human_review` die with **zero consumers**. Every detection lane below
  therefore names its consumer explicitly.

---

## A. Reach — mission (and constitution) into the build pipeline

The mission is literally about site-building, so this is where enforcement pays
most — and where it's most dangerous to get wrong (an over-eager mission check
objecting to legitimate consultancy-shaped sites would be its own drift). The plan
is staged observe-first, mirroring how the claims-verification and imagery lanes
rolled out.

**R0 — the classifier sees the platform mission (config-only, cheap).**
Inject a compact digest of `BIZ-001`/doc 028 (revenue-model-shapes-the-site; no
consultancy default; aspirational items marked `blocked`, never trimmed) into
`classify_and_extract` and `write_classification_spec` prompts. Today those
prompts see the per-site mission aspect (`SPEC-021`) but not the platform mission.
Honest labelling: this is a *complement*, not enforcement — pasted text is exactly
what failed for the constitution for months. Verify by reading the live seeded
prompt before/after; patch-style seed (the config re-seed clobber landmine).

**R1 — mission review of the classifier's output (config-only, the real step).**
A new advisory LLM step in the classifier workflow after
`write_classification_spec`: judge the just-written spec against the mission —
(a) is a revenue model named and argued from the domain's evidence (not
defaulted)? (b) any consultancy-shape default where the signal is absent?
(c) shape-mixing (a tools site with a "Start a Project" CTA)? (d) were
aspirational items marked rather than silently trimmed?
On objection: record the finding, then continue the build unchanged.

> **CORRECTED 2026-07-20 (during implementation, before applying):** this plan
> originally said "write a `mission_review` site_work_item at `status='detected'`;
> consumer = the triage sweep." Reading `triage_detect_items_action.go:91-103`
> refuted that: the triager is **site-scoped and type-blind** — it promotes ALL
> `detected` items into the dispatch pipeline, so a `mission_review` item would
> be swept toward a nonexistent handler: the opposite of observe-only, and a new
> instance of the 023 class. **As built:** objections append a `doc_notes` row
> (categories `['mission-review']` — the same machinery the council gate records
> verdicts in; reuse before recreate), which nothing can dispatch. **Named
> consumer: `101_REPORT_mission_review_findings.sh`** (run weekly alongside 098,
> and before any R2 decision). What caught it: the pre-apply grounding read of
> the consumer's actual code — the exact check the plan's own "named consumer"
> rule demanded.
Cost: one LLM call per classifier run — the classifier already makes several.
Measure ≥1 week: objection rate, and hand-grade a sample for false positives
(remember: a consultancy shape is *legitimate* when the evidence supports it —
the mission bans the *default*, not the shape).

**R2 — promotion to enforcement (decision-gated on R1's numbers).**
If R1's false-positive rate is acceptable (owner reads the week's sample), high-
severity objections flip from observe-only to a revise loop inside the classifier
(re-run `classify_and_extract` with the objection injected, once, then proceed
flagged). This needs small new machinery (the classifier has no revise loop);
design it only after R1 earns it. **Owner decision point.**

**R3 — fleet audit: `mission_alignment_check` discovery check (Go + image).**
R1 covers new builds; the existing fleet's drift is invisible to it. A discovery
check in the improvement loop's registry (self-registers via `init()`, enabled
only by being named in the agent's `checks` array, findings at `detected` — the
`IMP-004` contract the improvement-guardian seat now defends) that audits BUILT
sites for mechanical mission-drift signals: shape-mixing detectors (service-CTA
components on tools/content archetypes), consultancy-page sets on non-consultancy
classifications. Go change — inert until an image rolls; observe-only first, same
consumer as R1.

**R4 — the other two councils (config-only, the settled v19 pattern).**
Feature-designer and experience-planner councils get the same two (now three,
with the librarian) always-on seats. Same migration shape as v19; their
councils have their own decide steps, so this is a per-council patch + smoke run.
Cheap; can go any time the owner says.

**Sequencing:** R0+R1 together (one seed migration), R4 next, R3 when an image
window exists, R2 only on R1's evidence. Nothing blocks anything else.

---

## B. The direction-document drift guard

The v19/v20 seats enforce conformance *to* the constitution and mission; nothing
guards the documents themselves, or the seats' embedded copies of them. Three
drift surfaces: (1) the canonical files (anyone can commit an edit), (2) the seat
prompts in `agent_definitions` (any thread can UPDATE them live — this session
watched the render seat's prompt change under it mid-query), (3) file↔file copy
skew (three constitution copies today, identical only by luck of no edits).

**D1 — bless canonical paths + a direction ledger (docs-only).**
Name one canonical file per document. Constitution: verify which copy
`cmd/assembler` actually loads (CTXK-004) and bless THAT path; mission:
`docs024_key_docs_latest/028_platform_mission_and_pipeline_direction(2).md`.
Create `docs/agent_docs/docs026_concept_register/DIRECTION_LEDGER.md`: one row
per blessed doc — path, sha256, date, approver ("uk", the owner), and the sha of
each sanctioned copy. The ledger is the drift reference; updating it is what
"owner sign-off" concretely writes down.

**D2 — a real-gate commit hook (the actual guard).**

> **CORRECTED 2026-07-20 (during implementation):** planned as a `pre-commit`
> chain entry, built as **`.githooks/commit-msg`** — a trailer check needs the
> commit message, which does not exist yet at pre-commit time. Same tier, same
> behaviour.

Real-gate tier (alongside `check-secrets.sh`, whose precedent explicitly allows blocking): if the
staged diff touches a blessed path (or any tracked copy of one) AND the commit
message lacks a `Direction-Approved: <name>` trailer → **block**, printing why and
the exact trailer to add after getting the owner's word. Deliberately narrow: it
fires only on direction-doc edits, so fleet impact is zero until someone touches
the constitution — which is exactly the moment friction is wanted.
**Owner decision point: gate vs advisory.** Recommendation: GATE — "the direction
is fixed" was the ruling, the secrets precedent exists, and the cost is one
trailer line for a legitimate edit. (Advisory fallback: same hook, prints loudly,
never blocks — catches accidents, not intent.)

**D3 — integrity check across all three surfaces (script, 099-family).**
`100_CHECK_direction_integrity.py` (read-only by default, like 099's dry run):
(a) recompute blessed files' sha256 vs the ledger — mismatch = an unsigned edit
got past/around the hook; (b) compare every sanctioned copy against the canonical
— copy skew; (c) assert the constitution/mission/librarian seat prompts in BOTH
councils still carry the blessed rule anchors (needle LIKE checks, the same
technique the seat patches use) — catches live DB edits to the seats.
On mismatch: file a `site_work_item` at `detected` (consumer: triage sweep, as
above) and print loudly. Run: manually after any seat migration (add to the
runbook next to the 099 step), and — if a generic script-runner scheduled task
exists (verify; do not build one for this) — nightly. The 098 coverage report's
weekly run is the fallback cadence.

**D4 — longer-term: the `standards` table (`CTS-029`'s unbuilt destination).**
One migration: `standards` rows (`scope='constitution'`, one rule per row,
versioned per platform convention — `version` + `previous_version_id`, soft-delete
via `deleted_at`); seeded FROM the canonical file; assembler and seats read the
rows. The FILE stays the editing surface (D2 still gates it); a seed script
re-syncs rows from the file under the same `Direction-Approved` discipline. This
collapses drift surface (3) permanently and gives the seats a single live source.
Do this after D1–D3 prove the discipline; it is the structural fix, D2/D3 are the
guards that hold until and after it.

**Sequencing:** D1 (an hour) → D2 (the guard, owner picks gate/advisory) → D3
(same day) → D4 (its own small workstream, council-reviewed since it's platform/).

---

## C. Recorded alongside (from the same owner message)

- **`review_prior_art` (prior-art librarian) is LIVE** on both councils as of
  2026-07-20 (v20 + 099 mirror; 16 seats): verifies rationale existence-claims
  (asserted-absence / dormant-machinery) with mechanical lookups — code_checks
  for the codebase, SQL checks for agent_definitions/run history; explicitly
  states it CANNOT verify the running binary and names the pod-grep instead.
- **Dormant-machinery sweep (complement, not built):** the librarian's four
  lookups inverted, as a discovery check or 099-family report — seeded+active
  agents with zero orchestrations in N days; registered actions referenced by no
  workflow. Turns dormant machinery into a queryable surface instead of folklore
  (would have flagged the section-editor in days, not months). Natural sibling of
  R3; same observe-only + named-consumer rules.
