# DESIGN — automated diagnosis loop (the iterative-bundle reasoner)

Status: design only. Sibling to `DESIGN_doc_drift_classifier.md`; shares its
evidence-or-abstain contract and read-only posture. The plumbing it wraps
(`cmd/bundle`, the analyser, `dbcontext`) already exists — the loop is the new
part. NOT built; the prompt contract + convergence guards must be right first.

## 0. What this automates — and the one move that makes it hard

The gamesdesign debug (this session) was a loop of five moves:
1. **Hypothesise** from the symptom report ("sections never reach save").
2. **Gather** evidence — analyse repo, resolve targets, pull schema + runtime.
   ALREADY AUTOMATED: this is `cmd/bundle`.
3. **Test** the hypothesis against the evidence — the logs showed sections DO
   reach save; the regression guard blocks them. Hypothesis FALSIFIED.
4. **Re-scope** from what the evidence revealed — the in-scope code named
   `spawn_content_writer`, so the next bundle scopes there.
5. **Iterate** — narrower bundle, until the cause is pinned or the loop gives up.

Moves 1, 2, 4, 5 are mechanical. **Move 3 — falsification — is the crux and the
risk.** The single most valuable thing in the whole debug was the model SAYING
"my hypothesis is wrong, the evidence contradicts it." Twice. (First "sections
never reach save" → actually the guard blocks them; then "check the persisted
section status" → there is no status column, it's computed at runtime.) LLMs are
BAD at this by default: they rationalise the initial hypothesis and chase ever-
deeper bundles after a phantom. So this tool is not hard because the plumbing is
hard (it's done) — it is hard because the reasoning quality that matters is the
willingness to be wrong, which is what models do worst. Every design choice
below exists to force move 3 to happen explicitly rather than be skipped.

## 1. The loop

```
hypothesis ← symptom report
repeat (capped):
    bundle   ← cmd/bundle(scope, tables, runtime)        # gather (read-only)
    verdict  ← LLM(hypothesis, bundle)                    # cite-or-abstain (§2)
    case verdict:
      CONFIRMED  → stop: report cause + evidence trail
      REFUTED    → hypothesis ← verdict.revised_hypothesis  # the falsification move
                   scope      ← verdict.next_scope          # symbols the evidence named
      UNVERIFIABLE→ scope     ← verdict.needed_evidence      # widen to get what's missing
    guard: stop if not converging (§3)
emit: diagnosis (cause | best-effort | "couldn't pin it") + full evidence trail
```

The output is never a fix. It is a diagnosis WITH its evidence trail, for a human
to act on (§4).

## 2. The prompt contract — cite-or-abstain, falsify-first (load-bearing)

Reused from the doc-drift classifier, adapted to debugging. Per iteration the
model is given the hypothesis + the assembled bundle and must return:

- `CONFIRMED` — and QUOTE the evidence that confirms the hypothesis (the log
  line, the symbol + path, the row). No quote ⇒ not allowed to confirm.
- `REFUTED` — and QUOTE the evidence that CONTRADICTS the hypothesis, state what
  the evidence shows INSTEAD, and give the revised hypothesis + the next scope
  (the specific symbols/files the evidence names — e.g. "the code calls
  `spawn_content_writer`; scope that next").
- `UNVERIFIABLE` — the bundle doesn't settle it; name the SPECIFIC evidence that
  would (a table to query, a log to pull, a symbol to add to scope), which
  becomes the next bundle's gather.

Baked-in rules:
- A verdict without a citation is invalid → treated as `UNVERIFIABLE`. (The
  item-24 dated-claim discipline, as a machine rule.)
- The model is told, explicitly, that REFUTING its own hypothesis is the
  CORRECT and expected outcome when evidence contradicts it — not a failure.
  This is the antidote to confirmation-rationalising; make abandonment cheap.
- The model may NOT propose a code FIX. It diagnoses and cites; the human fixes
  (§4). (classify-don't-merge, applied to debugging.)
- Evidence is tagged with its tier (static / state / runtime — the classifier's
  T1/T2/T3) and, for state/runtime, its freshness, so a verdict resting on stale
  logs is visibly weak.
- A hypothesis may be marked CONFIRMED only on DIRECT evidence, never on
  "consistent with" — the classifier's abstention asymmetry. "The logs are
  consistent with a short generation" is UNVERIFIABLE, not CONFIRMED.

## 3. Convergence guards — a loop can run forever or wander

- **Iteration cap** (e.g. 5). Past it: stop, emit best-effort + what's still open.
- **Scope must NARROW.** Each new scope should be more specific than the last
  (fewer/deeper symbols), not a lateral wander. A loop that keeps widening scope
  is not converging — flag and stop.
- **Evidence must GROW.** If two consecutive iterations add no new grounded
  evidence (same citations, no new symbol/row/log), stop: "couldn't pin it,
  here's what I have." (The recency-heuristic caution from thin_versions, applied
  to reasoning depth — more iterations ≠ better answer.)
- **No hypothesis thrash.** If the loop oscillates between two hypotheses without
  new evidence discriminating them, stop and report BOTH with their evidence —
  let the human discriminate.

## 4. Read-only and human-gated (the boundaries, non-negotiable)

- The loop GATHERS (read-only `cmd/bundle`: analyser, `code_symbols`, `dbcontext`
  `\d`/capped SELECT/existing-log read) and PROPOSES (diagnosis + next scope). It
  does NOT apply fixes and does NOT trigger runs to test a hypothesis — it reads
  what already happened. (The doc-drift classifier's read-only rule; the
  conformance-suite carve-out's never-trigger rule.)
- Terminal output is a diagnosis for a HUMAN: cause (or best-effort), the
  evidence trail (every citation, every bundle scope, every verdict), and the
  suggested fix SURFACE (which file/step), never the fix itself.
- The human is kept at the two points that actually mattered this session:
  deciding the fix, and being the backstop for the model's willingness to
  abandon a wrong hypothesis (if the human sees the loop confirm on weak
  evidence, they override — the loop surfaces its reasoning precisely so this is
  possible).

## 5. Dogfoods everything

- gather = `cmd/bundle` (built); retrieval = analyser / `code_symbols` /
  `resolve_targets` (built, and B4a measures whether semantic retrieval helps
  pick the next scope); contract = the doc-drift classifier's cite-or-abstain
  (designed). The loop is the assembly of parts that already exist or are
  specified.
- It also generates B4a-relevant signal: each REFUTED→re-scope is a real test of
  whether retrieval surfaced the right next symbol.

## 6. Honest limits (so it isn't over-trusted)

- Plumbing-ready ≠ reasoning-reliable. The loop is only as good as move 3
  (falsification), and that is the unproven part. TEST it against KNOWN bugs
  before trusting it to narrow scope unsupervised — there is a growing set: this
  gamesdesign bug (two faults, with the full evidence trail captured), and the
  silent-no-op catalogue in 016 §9. A loop that reproduces those diagnoses —
  INCLUDING the mid-course hypothesis reversals — is one worth trusting; one that
  confirms the first guess on every known bug is not.
- The runtime tier reads EXISTING logs; if the relevant evidence was never
  logged (this session: the rejected ~3k content is unobservable because the
  guard returns before persisting it), the loop will correctly hit UNVERIFIABLE
  and ask for evidence that doesn't exist — which is a signal to ADD logging, not
  to let the loop guess. (Hence the bug runbook's "guard should log rejected
  content" finding.)

## 7. Open questions (decide before building)

- Hypothesis seed: from the symptom report verbatim, or a first cheap T1 pass to
  propose one? Probably: seed from the report, let iteration 1 refine.
- Who picks the initial scope — the human, or `resolve_targets` on the symptom?
  Probably: `resolve_targets` proposes, human confirms (the thin-slice's existing
  propose-then-confirm), THEN the loop runs.
- Single-hypothesis or a small frontier (carry the top-2)? Single is simpler and
  matches how this session ran; a frontier hedges against early wrong commitment
  but costs bundles. Start single; add a frontier only if single-track proves to
  commit too early on the known-bug set.
- Where the evidence trail lives + whether a CONFIRMED diagnosis auto-opens a
  work item — or stays a report a human reads. Recommended: report first, same
  as the classifier.
```
