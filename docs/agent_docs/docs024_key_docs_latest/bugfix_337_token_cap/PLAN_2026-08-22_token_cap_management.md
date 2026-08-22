# PLAN — bugfix 337: the 16,000-token cap, and threshold management (2026-08-22)

Design, phasing, decisions **and their reasons**. Owner steer (session brief): prefer a
robust framework-wide solution over the individual case; "I suggest increasing the
threshold and have some sort of threshold management".

## The problem in one paragraph

`component-creator/generate_template` (the fleet's component writer) is capped at
`ai_service.max_tokens = 16000`. One section type (`loans-credit-health-check`) needs
more, every time: nine truncations, all cut at `output_tokens=16000` with 46.4–48.8k
chars recovered, three items each burning three full generations, two live pages hollow.
The retries are spent on a deterministic refusal (016b §9, added 2026-08-20 from this
bug). And the cap is tight even for the step's ordinary work: successful calls' p95 is
13,633 (85% of cap), max 15,374 (96%).

## Decisions

### D1 — Ship a framework-wide escalate-on-truncation seam (`max_tokens_ceiling`)

**What:** a step may declare `max_tokens_ceiling` in its `ai_service` block. On a typed
`TruncatedError`, `execute_llm_prompt` retries ONCE with `max_tokens` raised to the
ceiling, logging the cut call as `success=false` with an `ESCALATED (bugs_open/337: …)`
prefix. The escalated call's outcome flows through the existing machinery verbatim.
Code: `platform/orchestration/actions/truncation_escalation.go` + wiring in
`ai_actions.go`; register **MDL-042**.

**Why this shape and not the alternatives:**
- *Why not just raise the cap:* the estate has ruled on this four times
  (`truncation.go:26-29`; 019 declined a raise with a 32,000-cap counterexample; 183
  raised its cap four times and is still open; 119: four seats raised, the failure
  landed on the thirteen left behind). A raise restores today's function; only a
  mechanism catches the next section type.
- *Why not the 119 re-ask:* that path deliberately does NOT raise the cap because it
  serves a *judgement* that can be re-asked shorter. A 47k-char component document
  cannot be asked shorter — its length is the work product. The only honest retry for a
  writer step is a taller one, bounded by an operator-chosen number.
- *Why not `tolerate_truncation`:* wrong for whole-component writers, on the record
  (012's composition rule; 205: "a salvaged partial ends at the last complete value, so
  trailing fields go silently ABSENT — failing loudly beats writing a half-record").
  The 076 guard would rightly refuse it here.
- *Why opt-in, default OFF:* owner ruling 2026-08-02 §2 (RFC_010) — new authority on a
  shared seam ships as an opt-in field with the unsafe side OFF. 67 carriers of
  `execute_llm_prompt` run byte-identically unless they name the key. This is not a
  rot-prone switch: without a ceiling *value* there is nothing to escalate to, so the
  key is the configuration.
- *Why ceiling > sent-cap is required:* an identical-height retry on a deterministic
  cut is exactly the measured waste (nine cap-hits, zero successes); refusing keeps a
  misconfiguration inert rather than doubling spend silently.

### D2 — Resize the step from measurement: 24,000 routine + 32,000 ceiling (migration 549)

- **24,000 routine:** clears the ordinary distribution (p95 13,633) by ~75% and the
  failing section's extrapolated need (~19–22k tokens) with margin. The sibling
  whole-component writers were levelled at 32,000/64,000 (migration 168 lineage,
  012's table) and this step was missed at 16,000 — the 067 sweep argument, and
  generate_template is component-creator's ONLY LLM step, so this is the whole sweep.
- **32,000 ceiling, not more — clock safety:** the chassis does not stream (600s HTTP
  timeout). This step measures 92–127 tok/s on sonnet-4-6 (the nine cut calls: 165–170s
  for 16,000 tokens). 32,000 at worst observed throughput ≈ 349s; at a conservative
  60 tok/s still 533s < 600s. 40,000+ at 60 tok/s crosses the timeout and converts a
  loud truncation into a silent clock death — LCO-007's C-vs-T doctrine says that is
  the strictly worse trade.
- **Why routine ≠ 32,000:** an escalation is a forensic event (the ESCALATED row), so
  demand above 24,000 stays visible and queryable. A step parked at its ceiling has no
  headroom signal left except the next truncation. This is the threshold-management
  half of the sizing: the routine cap is the monitored threshold, the ceiling is the
  bounded insurance.
- Migration follows 415 (assert the RESOLVED value; refuse if a higher-precedence
  top-level key exists — the 413 trap) and 484 (pre-state gate `IS DISTINCT FROM
  '16000'`, double-apply refusal, negative control on the inert `config.max_tokens`
  spelling, snapshot first, rollback sidecar). No ordering constraint vs the roll:
  applied first, the cap half is live and the ceiling waits for the code — staged
  arming, stated in the file.

### D3 — Threshold management: extend what exists; do NOT build a new monitor

**Finding that reshaped this plan:** the fleet-wide headroom monitor already exists —
`fleet-step-token-pressure` (register LCO-007, 6-hourly, C/T/N/P thresholds) — and it
**flagged this exact step from 2026-08-18** ("T generate_template@16000 — n=229, p95
92.4%, peak 100.0%, truncated 9"), writing doc_notes rows that nothing consumed. The
gap is not detection, it is flag→action dispatch (the estate-wide pattern already in
memory as "detection works; schedule and dispatch do not"). So:
- No new CronJob/check is built (a first draft of this plan had one — corrected by the
  prior-art sweep before any code was written).
- The escalation seam itself narrows the cost of the dispatch gap: an opted-in step
  under pressure heals per-call, and its ESCALATED rows are countable demand for a
  resize — visible to LCO-007's existing inputs.
- FIX-058's recorded open question ("should the near-miss threshold scale with the
  cap? Revisit when a 16000 seat first crosses") has now had its trigger: 337 IS the
  first 16000 crossing. Recorded in the bug file for that lane/owner; not answered
  unilaterally here.
- The flag→action gap is recorded in the bug file as the named residual, pointed at
  LCO-007, not duplicated into new machinery.

### D4 — Declined candidates, and why

- **Candidate 1 (bound the brief):** real precedent exists (FIX-059 cut a seat's peak
  from 98.3% to 55% of cap by prompt block alone; migration 484 bounded the term that
  scales). For a component writer the scaling term is the product itself; a
  decomposition rule (markup + extractor-separated script) changes WHAT the step
  produces and needs a measured real generation to design honestly. Left as the named
  follow-up, not smuggled into this change.
- **Candidate 4 (spend fewer attempts):** belongs to the RSH-007/WII-failure-ladder
  contract (bugs 307's owner ruling, one mechanism not three patches). With D1+D2 the
  deterministic burn mostly vanishes; a truncation AT the ceiling still burns the
  item's attempts — accepted, recorded as residual.

## Phasing

1. **Commit** code + tests + migration + register + docs (pathspec commits,
   `Council-Submitted:` trailer). Committed-then-reviewed is this estate's design
   (owner ruling 2026-07-29 §2).
2. **Council** submission (097) covering the platform edits + migration 549.
3. **Apply 549** (config half live immediately; scoped dry-run first per migration
   runner practice).
4. **Chassis roll** (owner's cadence) arms the escalation; verify at the binary
   (provenance stamp / `merge-base --is-ancestor`), never at the tag.
5. **Re-drive** `loans-credit-health-check` on loanzy.uk per RUNBOOK (pre-flight
   `pages.status` — `tool-eligibility-checker` is ARCHIVED, do not spend a build on
   it; pin the served page's before-state first: `grep -c '<input'` = 0).
6. **Verify at the artefact** and record WHICH mechanism won: completion with zero
   `ESCALATED%` rows proves the resize; with one, proves the escalation. Either is a
   pass with different attribution — do not conflate them in the close-out.
7. Also re-drive `tool-credit-roadmap` on loancalculator.co.uk (second live loss).

## Close-out bar (fixed AND live AND healed)

- Migration applied, resolved cap 24000, ceiling present.
- Escalation code live at the binary on the serving chassis.
- Both live pages carry a stored component (`</section>` present) and serve real
  controls (`<input` count > 0 from a pinned before of 0).
- Bug file moved to `bugs_closed/` only then; residuals (flag→action dispatch,
  candidate 1 follow-up, candidate 4, `execute_llm_prompt`'s missing ActionInputSpec)
  stay named in the file.
