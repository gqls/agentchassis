# DESIGN — doc-drift claim classifier (grounded, tiered, read-only)

Status: design only. The prompt contract (§3) is the part that must be right
before any code; everything else follows from it. Dogfoods contextkit.

## 0. What this is — and the two-tool split it must NOT blur

There are two related-but-different jobs, and conflating them sinks both:

- **Doc-drift classifier** (THIS note): a one-off / on-demand pass that, for
  each claim in a doc, decides *current | stale | unverifiable* against the
  real system, emits the evidence, and produces a per-document report. It does
  NOT merge docs and does NOT move files. Output is an annotated map of what's
  current; the human consolidates from it.
- **Standing conformance suite** (CARVED OUT — §6): a continuous "does the live
  system behave as documented" monitor, built on the existing
  `DiscoveryCheckContext`/`CheckResult` pattern, that pages when production
  diverges. Different cadence, owner, failure cost. The evidence tiers below
  feed it later; it is NOT built as part of doc cleanup.

If these merge, the heavyweight always-on thing gets built under the banner of
a cleanup and the cleanup never ships. Keep them separate.

## 1. Claim taxonomy — what is even checkable (carried from item 24)

Per-claim, against the code/system, three buckets:

- **code-checkable** — "X function exists", "table has column Y", "the adapter
  reads action from the body". Mechanically confirmable. The classifier's
  target.
- **superseded-but-not-wrong** — "we chose fork-on-deploy because…". The code
  can confirm the decision still HOLDS; it cannot confirm the rationale is still
  the operative one. Partial signal.
- **code-invisible** — "better for users", "we tried X, too slow", design
  intent, NEGATIVE RESULTS. The code says nothing. No signal — and this bucket
  is disproportionately why old docs are worth keeping.

The method works fully for bucket 1, partially for 2, NOT AT ALL for 3. So the
classifier makes the *factual* layer checkable and leaves the *judgement* layer
exactly as unverifiable as before. That is fine — provided bucket 2/3 reliably
route to *keep untouched* (§3), never to a confident verdict.

## 2. Evidence tiers — depth, with cost and safety

A claim is checked at the SHALLOWEST tier that can settle it; deeper tiers only
when needed and only read-only.

| Tier | Evidence | Source (all exist) | Cost | Safety |
|---|---|---|---|---|
| T1 static | symbol/signature/doc exists; column exists | `code_symbols` (analyser), `\d` / `information_schema` | cheap | inert |
| T2 state | rows actually present/shaped as claimed | read-only `SELECT` via `dbcontext -rows` | cheap | read-only |
| T3 behavioural | what ALREADY happened: recent errors, work-item lifecycle, orchestration outcomes | `dbcontext -runtime-site` (reads `agent_error_log`, `site_work_items`; + `system_events`, `entity_state_log`, `orchestration_states`) | medium | **read-only: inspects existing logs/rows, NEVER triggers a run** |

**T1 vs T3 is the depth you asked for:** T1 confirms code *exists*; T3 confirms
it *ran and did the thing* — by reading the trail it already left, not by
making it run. "The analyser indexes on deploy" is T1-confirmable as code and
T3-confirmable by finding the index rows + the orchestration that wrote them in
the existing logs.

### 2.1 The read-only rule (structural, non-negotiable)

Behavioural verification reads what the system ALREADY did. It never triggers a
build, index, spawn, commit, or any state mutation to test a sentence — those
paths spawn agents, write rows, hit GitHub, cost money; a doc-checker that
mutates state to verify a claim is worse than a stale doc. Deliberately-fenced
integration tests against a throwaway scope belong to the conformance suite
(§6), never the doc pass. (Same shape as the dedup tool's "report-only by
default".)

### 2.2 The two failure directions deepen with the tier

- T1 fails SAFE: can't find the symbol → `unverifiable` → keep. The only
  direction is toward "keep".
- T3 adds MISATTRIBUTION: the probe sees something off in the logs and the LLM
  blames the DOC when the real cause is an unrelated bug, a flaky run, or stale
  log data — a false `stale` that makes you rewrite a CORRECT doc. So:
  **behavioural evidence may support `stale` ONLY when it directly contradicts
  the claim; logs merely *consistent with* staleness are `unverifiable`, not
  `stale`.** Ambiguity routes to keep, never to a verdict. This asymmetry is
  what stops deeper checks manufacturing confident wrong verdicts.

## 3. The prompt contract — evidence-or-abstain (the load-bearing part)

The single inversion that separates a useful classifier from a confident-bullshit
generator: **do not ask the model to judge; require it to cite or abstain.**

Per claim, the model is given the claim text + the assembled evidence (T1
symbols, T2 rows, T3 runtime block as available) and must return EXACTLY one of:

- `current`  — and quote the specific evidence that CONFIRMS it (symbol name +
  path, the `\d` line, the row/count, or the log line). No quote ⇒ not allowed
  to say `current`.
- `stale`    — and quote the specific evidence that CONTRADICTS it, plus state
  what the system actually does now. T3 evidence qualifies only under the §2.2
  direct-contradiction rule.
- `unverifiable` — the evidence cannot settle it (bucket 2/3, or the relevant
  code/rows weren't in scope). Routes to KEEP-AS-IS, untouched.

Hard rules baked into the contract:
- A verdict WITHOUT a citation is invalid — reject and treat as `unverifiable`.
  (This is item 24's `[checked: …]` discipline as a machine rule.)
- The model may NOT propose rewritten doc text. It classifies and cites; the
  human authors. (No generative merge — §5.)
- Each evidence item is tagged with its TIER and, for T2/T3, its FRESHNESS
  (query time / latest log timestamp), so a `stale` verdict resting on
  month-old logs is visibly weak.
- "When unsure, `unverifiable`" is the DEFAULT, stated as such — the model is
  told abstention is the correct, expected, non-failure answer for anything it
  cannot ground.

### 3.1 Per-claim output schema

```json
{
  "claim": "verbatim claim sentence from the doc",
  "verdict": "current | stale | unverifiable",
  "evidence": [
    {"tier": "T1|T2|T3", "ref": "symbol path / \\d line / SELECT result / log line",
     "freshness": "RFC3339 or null", "supports": "current|stale"}
  ],
  "actual_state": "for stale only: what the system does now",
  "note": "optional, short"
}
```
No evidence array (or all-`unverifiable`) ⇒ verdict forced to `unverifiable`.

## 4. Pipeline — and it dogfoods contextkit

Per document:
1. **Extract claims.** Split the doc into discrete factual assertions (one
   checkable statement each). Imperative/aspirational/rationale sentences are
   tagged `non-claim` and skipped — only falsifiable statements enter the pass.
2. **Assemble evidence per claim** using the EXISTING tools:
   - T1: `resolve_targets` + `code_symbols` lookup for the symbols the claim
     names (this IS the B4a retrieval path — see §7);
   - T2: `dbcontext -rows` for a claim about row state;
   - T3: `dbcontext -runtime-site` for a claim about behaviour (read-only).
3. **Classify** each claim via the §3 contract.
4. **Aggregate** into a per-document report: `{current: N, stale: M,
   unverifiable: K}` with the stale ones and their citations listed first.
5. **Triage order** by the date/version signal (§4.1) — newest first to break
   ties, NOT to decide verdicts.

No file is moved or merged. The report is the product.

### 4.1 Date/version as TRIAGE, not truth (your refinement, inverted)

A recent file is more likely current; an old file is NOT more likely *wrong* —
it is more likely *unchecked* (your own observation). So:
- date/version ORDER the queue and BREAK TIES when two docs conflict (newer
  wins as the default to carry forward);
- they NEVER override a code check — a recent doc can confidently state what the
  code outgrew last week ("isn't built yet" was in a doc; the tool-doc-header
  docs went stale within hours). Code check decides; date orders.

## 5. Classify, do NOT merge (the line held firmest)

The grounding makes *checking* tractable. It does NOT make *generative merging*
safe: when an LLM rewrites N docs into one, the failure is SILENT — a dropped
caveat reads as clean prose, and no code-check catches an OMISSION because the
code cannot tell you what a human knew and wrote that isn't in the code. So the
tool's output is the per-claim map; the human consolidates with the citations in
hand. The model finds and cites; the human decides and writes. Every canonical
doc stays human-authored.

## 6. CARVE-OUT — the standing conformance suite (future, separate)

The same evidence tiers, run continuously instead of once, answer "does the live
system behave as documented" — a monitoring product, not a cleanup. It belongs
on the existing discovery-check rails (`DiscoveryCheckContext` → `CheckResult` →
work items), runs on a schedule, and MAY use fenced read-only-or-throwaway
integration probes the doc pass forbids. Built later, on demand, as its own
thing. Noted here so the doc classifier stays light and nobody bolts an
always-on suite onto a one-off pass.

## 7. Relationship to B4a

This is the strongest possible B4a test, for free: every claim's T1 step is a
real retrieval (claim → relevant code), the ground truth is the code itself, and
the classifier's hit/miss on finding the right symbol IS recall measured on
real queries. Build the classifier over a CLEAN index (post-dedup, `-exclude
_archive/`) or its retrieval inherits the duplicate-contamination problem.

## 8. Open questions (decide before building)

- Claim extraction granularity — per sentence, or per paragraph-topic? Too fine
  fragments a claim across rows; too coarse hides a stale clause in a current
  paragraph. Probably: per falsifiable sentence, with the paragraph as context.
- The T3 reach — start T1+T2 only (cheap, fully safe), add T3 once the T1/T2
  classifier is trusted? Recommended: yes, ship T1+T2 first.
- Where the report lives and whether stale verdicts auto-open `update_doc` work
  items (ties to the tier-1 `doc_drift` discovery check already noted) — or stay
  a plain report the human reads. Recommended: plain report first.
