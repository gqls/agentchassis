# PLAN — bug 020: tool-recreation invents data when the original was data-backed

*Started 2026-07-21 by the "bugfix 020" thread. Fixing the PLATFORM defect in
`/bugs_open/020` (the vetcomparison thread filed it and contained the site; the
platform path was unfixed). There is an owner HOLD on tool imagery until 020 is
fixed (`3f6f1febf`).*

## The defect, in one line

`tool-recreation-handler` recreates an interactive tool as "self-contained"
HTML/JS. When the original tool's behaviour **is its data** (a directory/search
reading `/data/*.json`), the model reads "self-contained" as "embed the data
too", and — with no rule forbidding it — invents a synthetic record corpus so
search/filter/pagination visibly "work". It shipped fake practices + postcodes to
a live public site. Root cause (case file): (a) no data-dependency contract, (b)
rule 9 scoped to arithmetic, not data.

## Design — two halves, deliberately separated

The case file ranks candidates **(1)+(2) structural, (3) the cheap net**. This
plan ships (1)+(2) as one config change (live now) and (3) as the mechanical gate
(needs an image). The key judgement: **a prompt cannot be the whole fix** — the
model already ignored the old rule 9, and "confidence is not a signal"
(CLAUDE.md). So the prompt REDUCES the rate and the Go gate is the net that does
not rely on obedience. Both are needed.

### Half A — prompt contract (candidates 1+2). LIVE, migration 183.

- `recreate_tool`: a prominent **`## Data Integrity`** section prepended to
  `## Requirements`, stating: never invent/seed/generate records; "self-contained"
  = CODE not DATA; if the original loaded data from a source, load from that SAME
  source unchanged; if no source is reachable, render an honest empty state and
  stop. Rule 9 rewritten from arithmetic-only to bind data.
- `analyze_tool`: JSON spec gains a **`data_source`** object capturing the
  original's fetch/XHR/`/data/*.json`/API target verbatim. It flows into
  `recreate_tool` (which renders the whole analysis JSON via `toJSON`) — so the
  contract is explicit, WITHOUT touching adoption-crawl plumbing. This is
  candidate (1) achieved cheaply: the analysis LLM already receives the raw source
  that contains the fetch; we just make it record it.

Decision & reason: candidate (1) as originally written wanted
`extract_interactive_fingerprint` to capture and carry the fetch target through
adoption. That is a larger, cross-cutting change. But the recreate model ALREADY
receives the full original source (with the fetch) and the analysis JSON — so the
same contract is achievable with a prompt-only change. Chosen for blast-radius.

### Half B — mechanical gate (candidate 3). BUILT + tested, inert until image.

New Go action `check_tool_fabrication` inspects the generated tool and, on a
fabrication signature, routes the item to `needs_human_review` (via the existing
`checkpoint_for_review`) instead of deploying. **Precision is the whole problem**
— a blunt PRNG grep flags every legitimate game. Tiered:

- **Tier A (fires alone):** the model DECLARING synthetic data ("realistic,
  deterministic dataset", "fake records", …); or synthetic-PII generators
  (`makePostcode`/`randomPhone`/…) the recreation INTRODUCED (not in the original).
- **Tier B (needs corroboration = the exact bug-020 signature):** a seeded PRNG /
  corpus builder / ≥2 crossed fragment arrays / large literal record array — AND
  the original was data-backed (a data-ish fetch/XHR, or analysis
  `data_source.has_external_data`) — AND the recreation preserves no fetch of its
  own. The corroboration is what spares a dice game or a name-generator tool.

Wiring: `check_completeness → check_fabrication → route_fabrication`; fabricated →
`request_fabrication_review` (checkpoint) → `complete` (no deploy); else →
`save_training_data` (the original next step). Staged **image-first** because the
step names an action that must be in the pod first.

## Not doing (and why)

- **Candidate 4 (machine-readable no-fabrication site flag).** Real, but broader
  than 020 and overlaps the `evidence_base` machinery already in
  `validate_page_content`. Out of scope for this fix; note for a follow-up.
- **Modifying `extract_interactive_fingerprint`** to carry fetch targets — not
  needed once the analysis captures `data_source` (see Half A decision).
- **Making `validate_tool`'s swallowed `error_step` gate** — the recreate
  workflow routes ALL validation errors to `save_sections` (swallowed), so even
  cross-site contamination deploys. Tempting to fix, but it broadens the change
  and risks blocking legit tools on placeholder/template false positives. The
  dedicated fabrication gate is scoped and independent; leave the swallow as a
  separate, noted concern.

## Phases / status

- [x] **A. Prompt contract** — migration 183 applied out of band + ledger-recorded;
  live-verified. Committed `266f900e5`.
- [x] **B1. Detector + tests** — `check_tool_fabrication_action.go` + 11 tests,
  `go build` + `go test` green. Registered. Committed `61f5fe567`.
- [x] **B2. Wiring staged** — `WIRING_..._APPLY_AFTER_IMAGE.sql` (not applied,
  image-first).
- [~] **B3. Council review** — submitted, `SUBMISSION_CORR 8eef369f`.
- [ ] **B4. Ship** — chassis image roll (owner's call), pod-grep
  `check_tool_fabrication`, apply the wiring, verify a data-backed recreation is
  held not deployed. **020 stays OPEN until this lands.**
