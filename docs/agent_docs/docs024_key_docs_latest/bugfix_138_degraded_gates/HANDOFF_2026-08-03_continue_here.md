# HANDOFF 2026-08-03 — bugs_open/138: everything is done and live; one owner decision remains

Supersedes `HANDOFF_2026-07-31_continue_here.md` (which predates induction rounds 2–3).
Cold-start order: this file → `NOTES_degraded_gates.md` (tail, 08-02/08-03 entries) →
RUNBOOK §11 + addendum → `bugs_open/138_HANDOFF_2026-07-29_*.md` (tail).

## State, in one paragraph

All four fix candidates are DONE and LIVE. The induced-fault programme ran three rounds
and is CLOSED: round 2 (cap 120) missed on two independent causes; round 3's offline
rehearsal proved `claude-sonnet-5` **refuses an instructed verdict** (verbatim in NOTES
08-02 evening), which closes the injection route on model integrity, permanently. Round
3 pivoted to a real fix that fell out of the investigation — `repairTruncatedJSON` was
bracket-blind inside string literals — **council-APPROVED** (trail `4d7363d7`: rejected
r2 → approved r3) and **VERIFIED LIVE** on replicaset `6d4b55c546`, both replicas
(08-03). Zero collateral across all rounds, each time proven at the artefact.

## The one open decision (owner's)

Close 138, or keep waiting? The TRUNCATED label string
(`gating TRUNCATED objection from X` + `gated_by_truncation=true`) is the ONLY artefact
never produced in production. Every arm feeding it has been separately
production-witnessed: 3× salvage-success correctly excluded (07-30/31), salvage-failure
→ honest `unreadable` (08-02), tolerate+retry (08-02), veto short-circuit (08-02),
frozen-plan mechanics (08-02). The decision rule composing them is unit-tested. The
remaining honest-induction design (RUNBOOK §11 addendum) needs a genuinely
minor-flawed submission + a cripple + a rehearsed cap, at ~30–50% odds per round.
**Lane recommendation: accept the evidence and close.** If closing: `git mv` to
`bugs_closed/`, COMMIT BOTH PATHS BY NAME (LANDMINES: a pathspec commit of a `git mv`
ships a copy otherwise), verify at HEAD with `git ls-tree`.

## Watches (nobody owes work, but these will fire someday)

- `metadata->>'gated_by_truncation'='true'` on any council_report — the organic true
  branch. If it fires, 138's last artefact is witnessed for free; record it in the bug
  file and close.
- The candidate-2 pressure alert (`council-seat-token-pressure`, every 6h, no credits)
  — if a seat grows into its cap again, expect degraded reviews to return.
- The landmine-verification verdict from 08-01 may still land in `doc_notes` under
  `categories ? 'landmine-verification'`, corr `f0c23b95-9c89-4ab4-9c02-47ab210ed0c2`.

## Handed to other lanes (not ours)

- **140/RFC_009 lane**: their `f48bf3e60` claims RFC_009 B "LIVE on v1.0.1237" but the
  running chassis binary (built 23:17–23:20Z 08-02 by literal-bracket) predates their
  commit (23:20:38Z), and no chassis RS existed in between. Notice with observables in
  their NOTES (08-03). Possible same-tag double-build of v1.0.1237; the 1237 makefile
  bump is still uncommitted.

## The five lessons this lane paid for (all in WRONG_CALLS, memory-indexed)

1. **Cite the arm, not the function** — twice reasoned about a mechanism from its
   name/comment; the exit table (every branch + an input landing in it) is the check.
2. **A measurement discharges a risk only at its operating point** — "thinking ≈ 0"
   measured at cap 8000 cost a round at cap 120; `output_tokens` counts thinking+text.
3. **Rehearse the exact call offline before spending a round** — 3 API calls (~pennies)
   found both the cap error AND the model's refusal; the prompt is in
   `llm_call_log.prompt_rendered`.
4. **A binary contains string literals and symbols, nothing else** — a post-roll grep
   for a local variable or a comment reads 0 in every world. For a change adding no
   string: bracket the build with neighbouring commits' compiled literals + ancestry;
   check the PACKAGE before reading an absence (git-adapter code is not in the chassis).
5. **The model will not lie for you** — a disclosed fixture instructing a verdict is
   answered with `approve` and a note that obeying would violate the reviewer's role.
   Design inductions around honest behaviour or not at all.

## Key artefacts

- Fix commit: `68cc1b4e8` (string-aware `repairTruncatedJSON` + first tests) — LIVE.
- Council trail: corr `4d7363d7-63dc-4b01-a912-ac0bb73c3031` (r1 n/a, r2 REJECTED veto,
  r3 APPROVED 23:08Z 08-02). `Council-Submitted` trailer on the commit; 098 credits it.
- Seat state: `review_adoption_guardian` healthy (cap 8000, prompt md5 `a39c7b70…` via
  SQL `md5()` — hash ONE way; the psql-dump route appends a newline and reads
  `94deeb7c…` for the same bytes).
- Rehearsal artefacts + restore/cripple SQL: session scratchpad (ephemeral — /tmp was
  wiped once already; the RUNBOOK carries the durable versions).
