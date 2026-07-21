# HANDOFF — CTA / link integrity (bug 023 and its family)

**Purpose.** Single resume point. Read top-to-bottom; this is enough to continue without
re-reading the whole directory. Deeper detail: `PLAN` (defect classes A–H), `NOTES`
(evidence + every misstep, append-only), `RUNBOOK` (R1–R14, every query with its gotcha),
`README_where_we_are.md` (the owner's plain-prose log), `SUMMARY_*` (milestone read-outs).

**Last updated:** 2026-07-21 · **Branch:** `085_debug_and_feature_loops`
**Origin:** owner review of leopardessconsulting.co.uk, 4 broken buttons (2026-07-19).
**Owning workstream:** this one. If CTA/link work surfaces elsewhere, coordinate and
contribute into `bugs_open/023` itself — do **not** start a competing fix.

**All figures below were re-measured live 2026-07-21** unless marked `[INHERITED]` (came
from a concurrent thread's docs, not personally re-checked this session).

---

## 0. The one-paragraph version

A button's label and its URL are unrelated schema fields, and nothing expresses "a label
implies a destination". The four buttons the owner reported are **extinct fleet-wide**. The
structural fix runs as a staged council-reviewed rollout: its first stage (schema-derived
CTA pairing, **observe-only** — logs where the platform's hardcoded map disagrees with the
schema, changes no behaviour) is **live in production**. The remaining work is construction
against a known design, plus a family of sibling bugs (033/039/045) that were split out
because they are owned by different decisions.

---

## 1. What is DONE and verified

- **The 4 owner-reported buttons: EXTINCT fleet-wide.** leopardess3 removed two misplaced
  sections from the two leopardess tool pages (P4); the `bugfix-023` thread completed the
  fleet (finetuning.uk + robot-hands) and **migration 179** fixed the `tool-guide-intro`
  component itself (CTA urls → renderer/optional, anchors gated, `#guide-start` removed).
  Fragment-dead-link class 4→0. Verified 2026-07-21: both leopardess tool pages serve 200
  with **zero** `Start Ranking Free`, no re-plan clobber since.
- **Observe-only stage SHIPPED and LIVE.** Commit `f6b4aea5a`. Verified in-pod 2026-07-21:
  running `docker.io/aqls/agent-chassis:v1.0.1144`, marker `cta derivation delta` present
  (survived the 1140→1144 rolls). Three log streams now emit (RUNBOOK R12/R14):
  `cta derivation delta` (Info, resolver: schema-derived set ≠ `ctaFieldNames`),
  `uncovered cta url field` (Warn), `cta ownership conflict` (Info, at the true loss site —
  the rerender merge). **No organic emissions yet = quiet window, NOT "no gap"** — the gap
  is structural (map covers 5 of 33 functions); it logs on the first CTA-bearing build.
- **Council trail `2525f980` CONCLUDED** — voided(019)→REJECTED→REVISE×4, final round 10
  approve / 2 object. Shipped **without** the `Council-Reviewed` trailer (trailer = APPROVED
  only) on the owner's explicit instruction to proceed and report objections. The two
  residual objections are dispositioned in NOTES (2026-07-20 evening entry): one satisfied
  in code, one converted to a named flip-round item.

## 2. The code that is live (observe-only — eyes, not hands)

- `platform/orchestration/datahelpers/ctafields.go` — `DeriveCTAURLFields` (3 sibling
  forms: `_label`/`_text`/bare stem, + a `site_assets` source guard so `logo_url` never
  derives), `UncoveredCTAURLFields` (the map's silent failure made loud, Warn-only),
  `ParseInputSchemaValue`. Both correction cases (from `bugfix-023`'s doc_notes, commit
  `b6e374fc2`) are pinned as regression tests in `ctafields_test.go`.
- `resolve_internal_links_action.go` — observe-only delta + uncovered logging;
  `sectionInputSchema` + `ctaDerivationDelta` added (`ctaDerivationDelta` has its own
  reachability test proving it is **not** dead code, unlike the sketch it replaced).
- `rerender_page_sections_action.go` — ownership-conflict log at the merge loop (stored
  content_data merges first, fresh `ResolvedData` wins). The earlier planSection placement
  was **dead code** (fresh local map, single-write loop) — the `bugfix-023` thread caught
  it; the comment records why so it is not re-invented.

## 3. THE FLIP ROUND (the next real code change) — a 5-constraint contract

Stage 2 inverts precedence: the schema-derived pairing decides writes, `ctaFieldNames`
retires. It returns to the council gate carrying **5 binding constraints accumulated from 5
seats** across the trail — a materially better-specified change than day-one:

1. Needle-gated **separate apply/verify/rollback** for any jsonb surgery (expected-count
   assert + `RETURNING` post-condition) — debug_historian.
2. The Warn-only uncovered-field guard **replaced by a consumed detection path** — a
   work-item type with a NAMED handler, explicitly **not** `needs_human_review` — or the
   flip does not ship — bug_historian.
3. `guidelines` **re-reviews the `on_missing`/`required` skip branch** specifically before
   it lands (a required field with no fallback must follow the schema's `on_missing`, never
   a bare `continue`).
4. `loadSingleComponentSchema` **converges onto `ParseInputSchemaValue`** (kill the second
   inline parse) — reuse_agent.
5. **Merge-loop ownership INVENTORY** — the observe log covers CTA url fields only; the
   generic stored-vs-fresh overwrite still hits logos/images/every other class silently.
   Enumerate what passes stored-vs-fresh through that merge and decide per class — bug_historian.

**Gate on the DELTA stream (coverage evidence), not conflict-log volume.** Conflict silence
on the 16 unmapped fossils is *expected* (the resolver never wrote them → fresh==stored);
they show up loud in the delta stream. RUNBOOK R12/R14.

## 4. The bug family — 023 is now NARROWED, siblings split out `[INHERITED]`

Other threads rescoped this the evening of 2026-07-20; I have **not** personally re-verified
every figure in this section — it reflects their `bugs_open/` filings (all four files
present, confirmed):

- **`bugs_open/023`** stays OPEN but narrowed to **classes A/B/C/E**. Its verify-criteria
  were rewritten because the old "34 inert items reach a terminal state via a handler"
  criterion was gated on a decision that is not 023's. New criterion 3 is **structural** —
  *no active library component can pair a rendered label with an absent destination* — so it
  cannot be closed by cleaning pages. Class D CLOSED (fragments extinct); class H = the
  council trail (observe stage live).
- **`bugs_open/033`** — the ~292-item write-only human-review queue. **Blocked on an OWNER
  DECISION about intent, not on code.** Fleet-wide, `[INHERITED]` claim: 0 of 119 CTA items
  have ever left `needs_human_review` (oldest 2026-06-22). This is where the "delivery gap"
  (findings die unread) actually lives now — it needs *you* to decide what should consume
  them, not more code from this thread.
- **`bugs_open/045`** — class F split out: the library's ONLY active `hero-tool` component
  is the Bayesian ranker (14 frozen `source:static` labels). ⚠️ **ARMED — verified live
  2026-07-21:** `finetuning.uk/ai-agent-roi-estimator` and
  `ai-agent-orchestration.com/agent-complexity-estimator` are `needs_rebuild` with
  `hero-tool` STILL in `pages.sections`. The 023 cleanup removed *placements*, never
  *plans* — the next rebuild re-adopts the Bayesian panel. **A clean page is NOT evidence;
  a rebuild is what arms it.** Do NOT delete the `_pre_037` row (sole active row for its
  function). Fix = build one neutral `hero-tool` component.
- **`bugs_open/039`** — sibling branch of the same selector: a section name that resolves
  to NOTHING → a hollow stub (045 = resolves to the WRONG component). Same root, different
  arm.

**Mechanism to feed the handler design `[INHERITED, robot-hands]`:** `chooseCTATargets`
(`resolve_internal_links_action.go:319`) picks `ordered[0]`/`ordered[1]` after sorting by
NavOrder/Name and **never reads the button's label** — so every CTA of a kind on a site
converges on the same two destinations. Any repair logic that picks destinations without
reading labels reproduces this. `areasExcludedFromCTA` {about,contact,privacy,terms,legal}
also makes some *correct* pairings (e.g. "Request Integration Support"→/contact.html)
unreachable by design.

## 5. Still open — construction against a known design (023 A/B/C/E)

Re-measured live 2026-07-21 (numbers moved since the memory's 2026-07-20 figures — other
threads gated some anchors and migration 179 dropped some llm urls; use these):

- **P2.1 — gate the ungated anchors.** **68 ungated / 37 components** (was 75/38 on 07-19).
  Wrap CTA anchors in `{{if .x_url}}` so a missing destination renders no button (LNK-005).
  `html_template` is a DB column → live immediately, high blast radius, back up first;
  re-derive the exact list with a real template parse (R9 is a heuristic).
- **P1.5 — deterministic fabrication checks.** (a) email→hostname: a host equal to a known
  contact email with `@`→`.` is fabricated by construction (6 sites share
  `contactforsales.com`). (b) different-TLD: `finetuning.ai` proved the class live — it
  RESOLVES to a third-party page. Both need no network call.
- **schema-lint — ban `source:llm` on URL fields.** **19 llm-sourced url fields / 4
  components** live (the memory's "22/6" is now stale — migration 179 reduced it). Each is
  an instruction to a model to invent a URL it cannot look up.
- **P1.2 — build-time pairing check** (`cta_without_destination`): label resolves non-empty,
  URL empty/absent → finding. Ship as WARNING first, drain, then blocker (30→ empty hrefs
  live; a cold blocker fails the fleet's next rebuild — the LNK-009 staging lesson).
- **Merge-loop inventory** (flip constraint 5, above).

## 6. Owner deliverable, ready to hand over

`PROPOSAL_council_seat_sketch_falsifier.md` — a paragraph proposing a council seat (or
edit-quality amendment) whose job is to **falsify edit sketches against the real source** of
the functions they patch (symbols exist, conditions reachable, invariants quoted from
lines). Founding case: my dead observe-log survived 3 rounds and 10+ seats because every
seat reviewed the *idea* and none traced the *code*. Round 6's editquality started doing
exactly this mid-trail — cite it as in-the-wild evidence the behaviour works and needs a
mandate. Hand to the concept-register / council-gate workstream; seat via fix-proposer +
099 mirror (do not hand-patch the gate).

## 7. Landmines (every one cost real time)

- `slot_name` resolves via `content_components.function`, **not** `.name`; `component_id`
  is unpopulated fleet-wide (joining on it "finds" ~100 false orphans).
- **`bugs_open/001` is FIXED and active in v1.0.1140+** — so `page_components` edits on
  these sites are now durable (verified: no re-plan clobber). That constraint from the early
  plan is lifted.
- Council: seat edit-indexing is inconsistent (read the `problem` text, not the `edit`
  number); `council_report.metadata->>'source_agent'` is empty (partition on
  `reviews[].reviewer`); a truncated reviewer voids the WHOLE round (bugs_open/019) — keep
  submissions lean and tell reviewers to be brief.
- **019's cap is not on the seat config** — no `ai_service` at root or on any `review_*`
  step, so the 8000 is a global default you cannot inspect from the agent.
- The 090 diagnosis trigger now defaults `REF` to the **current branch** and REFUSES
  off-origin refs (I changed it — `main` was 345 commits stale). Verdicts live in
  `orchestration_states.collected_data->'diagnosis'` and
  `diagnosis_artifacts kind='council_report'`; they reach `doc_notes` only if
  `SUBJECT_TYPE`+`SUBJECT_KEY` were set.
- **Coordination:** the memory dir is not under git and was overwritten twice this arc
  (my README, then this workstream's memory file). Contribute into shared bug files with
  merges, not `cat >`.

## 8. How to verify anything here

`RUNBOOK_cta_link_integrity.md` — R1 (fleet dead-control census), R2 (slot→component),
R9 (ungated worklist), R10 (backup rows live), R11 (retrieve a council/diagnosis verdict),
R12/R14 (read the observe streams), plus every council-submission gotcha. Deploy is verified
against the **running pod**, never the tag:
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "cta derivation delta"'`.
