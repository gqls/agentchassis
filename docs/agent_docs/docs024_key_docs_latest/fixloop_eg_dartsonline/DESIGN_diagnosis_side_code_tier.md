# DESIGN — the diagnosis-side code tier (and how it fits the evidence-tool family)

*2026-07-18, diagnosis-fixloop thread. Plans the diagnosis-side equivalent of
the council's code-lookup tier (F2.3b(c)), and places it in the family of
"ask for more evidence" tools both sides of the loop already have. Also records
the decision on bugs_open/016 finding 2 (a related council-plumbing fix).
DRAFT for build; nothing here is built yet.*

---

## 1. The family of evidence tools today

Both halves of the loop — the DIAGNOSER (finds the cause) and the COUNCIL
(reviews the fix) — can ask for more evidence mid-run. Today they have these:

| Tool | Side | Tier | How it works |
|---|---|---|---|
| `data_requests` | diagnosis | **state** | verdict emits read-only SQL → run under containment (lint + READ ONLY tx + EXPLAIN gate) → folded into the next bundle, cited `state` |
| `next_scope` + call-graph follow | diagnosis | **code-navigation** | verdict names symbols → loop re-fetches their bodies and follows the call graph one hop |
| `run_checks` | council | **state** | reviewers emit `checks:[{sql,why}]` → same SQL containment → fed to the repropose |
| `code_lookup` (F2.3b(c)) | council | **code-search** | reviewers emit `code_checks:[{kind,query,why}]` → answered from the `code_symbols` index (symbol/content/ls) → fed to the repropose, cited by commit_sha |

Read the columns: both sides have a **state** tier (SQL). The council has a
**code-search** tier (the thing I built this week). The diagnosis side has
**code-navigation** (follow the call graph) but **no code-search**.

## 2. The gap, and why it matters

Call-graph following answers "what does the code I'm looking at *call*?" — depth
within the scope the evidence already reaches. It cannot answer "does this
mechanism exist **elsewhere**?", "what **references** symbol Y?", or "how many
implementations of Z are there?" — breadth *beyond* the reachable scope. And that
is exactly where cross-cutting causes hide: a symbol used somewhere the current
scope never calls into.

This is not hypothetical. It is the *diagnosis-side* version of the very question
the council's code tier was built to answer. On the fix side, the bug-historian
asked "do other provider adapters exist?" and could not get an answer until the
code tier shipped. On the diagnosis side, the same class of question — "is this
truncation-swallowing pattern in other client adapters?", "does anything else
read this config the shadowed way?" — is unanswerable today unless it happens to
be call-graph-reachable. The diagnoser can follow the trail it is on; it cannot
sweep the codebase for the pattern.

## 3. The plan: reuse, don't rebuild

The council's tier is an **action** (`diagnose_code_lookup`) that answers
`{kind: symbol|content|ls, query, why}` from `code_symbols`, runs in-chassis, and
renders each answer with its commit_sha. The diagnosis-side tier is **the same
action, wired into the diagnosis loop** — three small pieces:

1. **Verdict wire:** add `code_requests: [{kind, query, why}]` to the verdict
   schema (`pkg/diagnose/verdict_wire.go`), a sibling of the existing
   `data_requests`. Same shape as the council's `code_checks`; one convention
   across the whole loop.
2. **Route forwarding:** `diagnose_route` already forwards `data_requests` into
   `route.data_requests` on loop-back; forward `code_requests` into
   `route.code_requests` the same way.
3. **Gather step:** in the loop's gather phase (where `diagnose_load_runtime`
   runs the `route.data_requests` SQL), also run `diagnose_code_lookup` over
   `route.code_requests` and append its `results_text` to the bundle. Cite at a
   `static`/`code` tier so the two-evidence-family guard treats a code-search
   answer as static evidence (it is code, not observed state).

The verdict prompt gains one paragraph, mirroring the council's: *"when your
hypothesis turns on whether a mechanism exists elsewhere — another
implementation, a second call site, anything that references X — emit a
`code_request` rather than guessing or abstaining; it is answered from the
code_symbols index next iteration."*

Everything else exists: the action, the index (3,723 symbols, commit-pinned), the
Go-receiver-aware symbol match and dedup (the two fixes the council's first live
run earned), the in-chassis-no-token property.

## 4. How it fits with call-graph following (they are complementary)

- **Call-graph following = depth.** "Follow what this code calls." Navigation
  within the reachable scope. Keep it — it is how the loop reaches a cause down a
  known trail.
- **`code_requests` = breadth.** "Search the whole indexed codebase for a
  pattern." Discovery beyond the reachable scope. New.

A diagnosis uses both: navigation to walk the trail from symptom toward cause,
search to check "is this the only place this happens?" before it confirms. The
second question is precisely what turns a local diagnosis into a platform-wide
one — the thing BUG A and BUG B needed and got only because a human asked.

## 5. One nuance: the diagnoser HAS the real tree

Unlike reviewers (in-chassis, index-only), the diagnoser runs in a spawned pod
with the repo tarball checked out (`analyse_repo_local` → `out.Root`), so it
*could* grep the live tree instead of the index. Trade-off:

- **Index (recommended):** reuses the action verbatim; commit-pinned so staleness
  is visible; symmetric with the council; no per-query tree walk. Cost: the index
  can lag the fetched ref (refreshed by index-orchestrator).
- **Local-tree grep (later option):** always matches the exact fetched ref; no
  staleness. Cost: bespoke code, a second mechanism, not shared with the council.

**Recommendation: index-based first** (reuse, symmetry, one convention). If a
diagnosis is ever misled by index staleness against its fetched ref, add a
`code_requests` mode that greps `out.Root` as a fallback — but only if it proves
needed. Don't build the second mechanism speculatively.

## 6. Related decision — bugs_open/016 finding 2 (the reviser's blind seats)

Confirmed live 2026-07-18: 13 council seats are seeded, but the `repropose`/
`reframe` prompts thread only **6** per-seat refs (`{{.review_editquality}}` …),
so **7 seats' objections are invisible to the reviser** (adoption_guardian,
compliance, debug_historian, llm_reliability, diagnosis_guardian,
improvement_guardian, render_guardian). A revise round cannot see 54% of the
council. The gap arrived by **seat growth** (6→13), not a code defect, so adding
seven more prompt sections is not idempotent — it re-breaks on seat 14.

**Decision (this was mine to make): read the artifact, don't list the seats.**
The `council_report` artifact already carries EVERY reviewer's verdict in one
`reviews: [{reviewer, verdict, notes}]` array. The reviser should read that once
instead of threading each seat through the prompt. Concretely:

- `diagnose_council_decide` already assembles that array to make its decision and
  persists it as the `council_report` artifact. Have it ALSO emit a rendered
  `reviews_text` (or expose the array) into its collected output, and have the
  `repropose`/`reframe` prompts render **that one field** — either
  `{{.council.reviews_text}}` or `{{range .council.reviews}}### {{.reviewer}} —
  {{.verdict}}\n{{.notes}}{{end}}` (the template engine supports `range`).
- Then the reviser sees all 13 seats, and seat 14 needs no prompt change — it
  flows through the array automatically. This is the idempotent, roster-growth-
  proof shape the reasoning-dataset thread asked for.

Needs a small `diagnose_council_decide` change (emit the rendered reviews) + a
prompt change on `repropose`/`reframe` (one ref, not N). Sequence it with the
sonnet-5 councils; patch-style; mirror to the gate via `099` (now config-drift-
aware). This is a distinct build item — planned here, not done in this doc.

## 7. Build order (when the owner greenlights)
1. `code_requests` in the verdict wire + route forwarding + gather wiring (reuse
   `diagnose_code_lookup`). Prove on a real config/cross-cutting diagnosis.
2. The 016-finding-2 reviser fix (read the council_report array). Independent;
   can go first — it is a live correctness bug (the reviser is half-blind now).
3. Both are small because the hard parts (the action, the index, the artifact)
   already exist.
