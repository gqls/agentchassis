# HANDOFF 2026-08-18 — Phase 0 CLOSED and costed; commercial direction SET; next work = bugs_open/302 fix + the two consultation answers

**COLD-START for the merged finetuning.uk lane (backend + front end, owner
merged 08-18).** Supersedes `HANDOFF_2026-08-17_…`. Statistics of record:
`RESULTS_2026-08-15…` (incl. §7 + invoice-settled rates). Market + positioning:
`RESEARCH_2026-08-18_competitive_landscape.md`. Milestone read-out:
`SUMMARY_2026-08-18_…`.

## State, verified 2026-08-18

| thing | state |
|---|---|
| Phase 0 | **COMPLETE + INVOICE-CONFIRMED**: $1.12 total, matches our estimate to the cent; a6000 $0.35/hr FLAT (no vCPU surcharge), a100xl $1.09/hr |
| GPU estate | paused, vendor `{}`, 0 live; all 3 guardians armed (reaper — 1 real reap proven; orphan-scan; **training-monitor ENABLED 08-18**, FTW-035 deployed) |
| Artefacts in B2 | adapter (68MB) + GGUF (1.06GB), both range-GET verified |
| Price | **£99 start** (owner): business audience, low hundreds credible, reduce-later-not-raise-later; believability via PROCESS DETAIL, not price rises |
| Positioning | owner-ratified, 3 principles in RESEARCH doc: journey-not-machinery · specificity-earns-authority · friendly front door + one honest technical page. ⚠ "a real person checks every run" = **[UNVERIFIED AS PROMISE]** (owner's own flag) — must not reach copy as-is |
| Copy | **NOT written, deliberately** — goes through the framework; register = friendly, EXPANSIVE, glossary. Consultation lodged with `copy_quality_two_stage` (CONTRIB 08-18 in their dir, 3 questions) — no reply yet, their lane is active |
| Benefits | consultation lodged with `vigilant_designer_offer_analysis` (CONTRIB 08-18, benefit ordering + believability + upsell-shadow) — no reply yet |
| Front-end repairs | **GATED on `bugs_open/302`** (filed this lane, 08-18): design-repair item types have NO registered verifier → `verifyBeforeComplete` abstains → no-op completions pass. 090 ran twice first (run 1 FAILED null-error; run 2 UNVERIFIABLE, broad claim REFUTED — narrow claim is code-verified + declared). Evidence fixtures: 4 retained `complete` rows on finetuning.uk |
| 259 guard | still no live firing (owner decision stands, 15b §4) |

## The next work, in order

1. **Fix 302.** Candidates ordered in the bug file. ⚠ **Scope carefully before
   coding**: candidate 1 (missing verifier ⇒ REFUSE completion for repair-shaped
   types) changes what the SHARED completion gate guarantees — that is
   architecture-scope territory per the 07-29 ruling (a change to what a shared
   mechanism GUARANTEES needs an RFC; additive-and-inert does not). Candidate 2
   (register verifiers for the design-repair family) is contained and
   council-gate-only. A defensible sequence: ship candidate 2 now, raise the
   candidate-1 class question as an RFC. Either way: council submit before/with
   commit, `Council-Submitted:` trailer, register the seam if one is added.
2. **Chase/ingest the two consultation replies**; then seed the offer page
   through the framework (082_submit_domain_unified / site specs — NEVER
   hand-build, owner ruling 08-04). The copy-quality answer shapes the identity
   spec BEFORE seeding.
3. **Phase 1**: offer page + £99 Stripe Payment Link (concierge). Owner calls
   still open: playground booking shape, sample datasets, Stripe posture.

## Traps current for this lane (fuller set: RUNBOOK §7–§9, HANDOFF 08-17)

- Chassis rolls repeatedly these days: re-read build provenance per service
  before any ancestry claim; no orchestration dispatch within 300s of a chassis
  pod start.
- unsloth `save_pretrained_gguf(out)` writes `<out>_gguf/` while printing
  success naming `<out>`.
- Boot time is day-variable ~20× — never quote one without its date.
- `cost_usd` books run 5.1× over (flat $1.80/hr, deliberate).
- Presigned anything: size by `Content-Range` on a range GET, never HEAD.
- All watchers: foreground-test the filter against a line that EXISTS (two
  incidents, WRONG_CALLS 08-15 + 08-17).

## Fleet-wide records this lane owns

`bugs_open/302` (new) · `bugs_open/258` closed-pending-move? (258/259 fixes
LIVE + APPROVED — check file locations before citing) · RESULTS/RESEARCH/
SUMMARY series · register FTW-031/032/035 updated · 016b §9 verifier-registry
pattern · LANDMINES + WRONG_CALLS entries per their files.
