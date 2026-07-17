# PILOT — "Tooling & Provenance" council reviewer (stage 3, seat #4 of the extended roster; candidate #10)

**Status: LIVE as of 2026-07-17.** Applied to `clients_db` via
`fixloop_eg_dartsonline/0NN_fix_proposer_v10_tooling_provenance.sql`. Pre-flight
confirmed no active fix-proposer runs; prior row snapshotted (`f9d90a2d-...`).
Verified live: `review_tooling_provenance` present, wired
`review_guidelines → review_tooling_provenance → review_guardian`,
`council_decide` + `escalate` + `run_checks` all carry 6 reviewers, prompt
intact (2,801 chars). **The council is now 6 reviewers.** Footprint added to
the relevance-filter config (`DESIGN_relevance_filter.md` §7) so it auto-gates
when the filter deploys.

---

## 1. Why this seat, and an honest note on sequencing

The user picked candidate #10 from the "ten more" list next: the
documentation/contextkit specialist, anchored in `CTXK-015` — the single
most-rediscovered concept in the entire register (11 independent source
citations). The lesson people keep relearning: **this platform already has
first-class tooling for investigating and documenting a change, and work that
ignores it re-derives lost context or reinvents machinery that already
exists.**

**Sequencing tension, stated plainly:** this is a *specialist* seat, and the
relevance filter built earlier today (`DESIGN_relevance_filter.md`) exists
precisely so specialists run *only when relevant*, not on every decision. The
filter's engine is built and committed but not yet deployed (another thread
leads that deploy). So this seat is applied **always-on for now** — a
deliberate, low-cost interim: the council isn't being exercised on real cases
yet (BUG A's dispatch awaits the owner), so a 6th always-on seat costs
effectively nothing today, and its footprint is added to the filter config so
it **auto-gates the moment the filter deploys**. This seat's narrowness (most
fixes never touch context/doc tooling) makes it the clearest illustration of
why the filter matters — noted so the interim isn't mistaken for abandoning
the gated design.

## 2. Grounding concepts

| Concept | Category | The lesson |
|---|---|---|
| `CTXK-015` | contextkit-toolchain | The platform's investigation tooling (`cmd/bundle`/`contextkit`) and its standing rule: "that is a code question → a bundle" — produce a written VERDICT of the deciding questions before touching code whose supported method is unclear. Plus the recurring trap: resolve an action from the REGISTRY (key → Handler → function), never by filename convention (`execute_llm_prompt` lives in `ai_actions.go`; `validate_page_content.go` lacks the `_action` suffix). |
| `DOC-010` | documentation-system | Travelling documentation: every tool/pipeline carries a living PLAN + NOTES in Postgres (`doc_plans`/`doc_notes`), keyed by `(subject_type, subject_key)`. Agents load them before touching a subject and write NOTES as a byproduct, so fixes build on prior decisions instead of re-deriving lost context. The fix-loop itself ADOPTED this (its `diagnosis_artifacts`/`doc_notes`) rather than building a rival — a live endorsement of the discipline this seat enforces. |
| `DOC-046` | documentation-system | Doc-hygiene tooling (dedup / thin_versions / archiving) already exists — a fix should not reinvent it. |

## 3. Charter

**The tooling & provenance reviewer judges one question: does this fix use the
platform's own investigation and documentation machinery, or reinvent / work
around it?** Concretely, it looks for:
- (a) new ad-hoc context-gathering / bundling / source-parsing code where
  `cmd/bundle`/`contextkit` or an existing action already does the job;
- (b) a fix that touches a tool/pipeline/agent carrying a travelling PLAN/NOTES
  (`doc_plans`/`doc_notes`) without accounting for it — neither loading the
  prior decisions nor leaving a NOTES entry;
- (c) resolving an action/handler by filename convention instead of the
  registry (the `CTXK-015` trap);
- (d) reinventing doc/context tooling that already exists (`DOC-046`).

It overlaps the reuse-agent (both dislike reinvention) but is distinct: its
lens is specifically the *investigation and provenance tooling*, and it adds
the travelling-docs discipline the reuse-agent has no view of.

**Verdicts: `approve | object`, no `veto`** — same advisory design and reason
as seats #1–3 (any reviewer's veto rejects outright regardless of
`hard_veto_from`).

## 4. Prompt template (matches the existing reviewers' contract)

```
# Council reviewer: TOOLING & PROVENANCE

You judge one thing: does this fix use the platform's own investigation and
documentation machinery, or reinvent / work around it? You change nothing;
you judge.

## The platform's own machinery (use it; don't reinvent it)
- INVESTIGATION: cmd/bundle / contextkit is the platform's tool for reading a
  change's real supported method before touching it — the standing rule is
  "that's a code question -> a bundle." Actions resolve from the REGISTRY
  (key -> Handler symbol -> function), NEVER by filename convention
  (execute_llm_prompt lives in ai_actions.go; some actions lack the _action
  suffix) — a fix that assumes a file name from an action key is repeating a
  documented recurring mistake.
- TRAVELLING DOCS: every tool/pipeline/agent carries a living PLAN + NOTES in
  Postgres (doc_plans / doc_notes, keyed by subject_type + subject_key).
  Fixes are supposed to load the subject's prior decisions before changing it
  and leave a NOTES entry — so the next fix builds on this one instead of
  re-deriving lost context. (The fix loop itself uses exactly this pattern via
  diagnosis_artifacts / doc_notes.)
- DOC HYGIENE tooling (dedup / thin_versions / archiving) already exists.

Judge the plan: (a) does it add new ad-hoc context-gathering / bundling /
source-parsing code where cmd/bundle/contextkit or an existing action already
does it; (b) does it touch a tool/pipeline that has a travelling PLAN/NOTES
without accounting for it; (c) does any edit resolve an action/handler by
filename convention rather than the registry; (d) does it reinvent existing
doc/context tooling. If none apply (most fixes touch no tooling at all),
approve.

Verdicts: approve (uses the platform's machinery, or touches none of it),
object (reinvents or works around existing tooling / ignores a subject's
travelling docs -- name the specific existing mechanism it should use). You do
NOT have a veto -- put a severe concern in objections at "high" severity and
trust the router; note a true architecture-level concern explicitly.

CHECKS: if a verdict hinges on whether a doc_plans/doc_notes row or a registry
entry exists, put that query in checks as {"sql": "SELECT ...", "why": "..."}
-- SELECT/WITH only. Write checks ONLY against the tables/columns in the Schema
section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output -- ONLY this JSON
{"reviewer": "tooling_provenance", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the existing tooling ignored/reinvented", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "notes": "..."}
```

## 5. Exact wiring (extends v9, becomes v10)

Chain: `... → review_guidelines → review_tooling_provenance (NEW) →
review_guardian → council_decide`. Same five-edit shape as v6/v7/v8:
1. `review_guidelines.next_step`: `'review_guardian'` → `'review_tooling_provenance'`
2. New step `review_tooling_provenance`, `next_step: 'review_guardian'`
3. `council_decide.review_fields` + `escalate.review_fields`: add
   `'review_tooling_provenance.result'`
4. `run_checks.check_fields`: add `'review_tooling_provenance.result.checks'`
   (the v9 fix's rule — every reviewer's checks run)
5. `repropose.input_fields` + prompt: add `review_tooling_provenance`

## 6. Filter footprint (added to `DESIGN_relevance_filter.md` §7 config)

```json
"tooling_provenance": ["contextkit","cmd/bundle","bundle","doc_plans","doc_notes","resolve_action","registry.go","docubundle","travelling","dedup","thin_versions"]
```
Narrow by construction — most fix plans touch none of these, so once the filter
is live this seat abstains on the large majority of fixes and only wakes for
ones that actually touch investigation/doc tooling. That is the whole point of
gating it.
