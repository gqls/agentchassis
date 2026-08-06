# PLAN 2026-08-06 — executing RFC_012's second-sitting rulings

Owner rulings (recorded in the RFC, commit `3851e90b5`), three pieces, this lane owns all:

1. **(d) → a STANDING check, online if the framework allows.** Bound by addendum 1
   (13-key routing graph, different-action discriminator; both naive detectors return 0
   on the known bug — a candidate must prove itself against that case). Vehicle to be
   chosen against the RFC_006 precedent: live config is guarded by a scheduled check,
   never a commit-time hook.
2. **The §3(a) reader census — COMMISSIONED.** Deliverable: an artefact in
   `architecture_review/` naming every reader of an awaited step's key (config side from
   live `agent_definitions`; Go side mechanically-findable readers, honestly bounded)
   with a per-reader verdict under merge-not-replace. This is what (a)/(a′) stays gated
   behind; the census does not decide (a).
3. **Option B implementation — ASSIGNED here, now.** A leaf-package shared
   `agent_error_log` writer importable from `actions` (retiring the hand-copied INSERT
   column lists) + a named findings-plus-await helper generalising 098 debt-5b's proven
   direct-DB-row pattern. Through the council gate; concept-register entry in the same
   commit as the seam.

## Decisions and their reasons

- Research delegated to three parallel agents (helper ground truth; online-check
  machinery precedents; the census itself) — this session's context is deep, and the
  census in particular is exactly the shape an agent can run end-to-end with the method
  written into the artefact for re-running.
- Order: census lands independently (its own commit); helper + check are code and go
  through one council round as a coherent task (they share the RFC and the register
  category) unless research shows they are better split.

## Verification (to be completed as research lands)

- Helper: unit tests + the debt-5b shape re-expressed through it; `go build ./...` on
  `git archive HEAD`; the import-cycle claim proven by the build itself.
- Standing check: MUST detect the known-bug case addendum 1 names (the disconfirmable
  half); a no-finding run on today's fleet is only meaningful beside that.
- Census: queries verbatim in the artefact so it can be re-run; totals stated;
  [UNVERIFIED] marked where a grep cannot see.
