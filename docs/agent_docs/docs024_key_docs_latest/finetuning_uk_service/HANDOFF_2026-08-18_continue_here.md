# HANDOFF 2026-08-18 — Phase 0 CLOSED and costed; commercial direction SET; next work = bugs_open/302 fix + the two consultation answers

> ## ⚠ DELTA 2026-08-19 — READ THIS FIRST; two of the three "next work" items below changed under us overnight (395 fleet commits since this file was written)
>
> **1. `bugs_open/302` is NOT ours to fix any more — it is CLOSED.** Another thread
> (`bugfix_302_design_repair_verification`) took it the same afternoon: fixed ONE arm
> (`743bc1945` — gate 1b's unreadable-payload arm became a per-type declaration;
> `dark_section_audit` now REFUSES a blob completion), council **APPROVED r1**
> (`edfef8cc`), rolled and **proven four-way on `v1.0.1314`**, moved to
> `bugs_closed/302_…`. **My filing was corrected three ways by their measurement pass —
> inherit the corrections, not my text:** the registry holds **13** types not 11; 7 of the
> 11 unreadable payloads were `bugs_closed/287`'s spawn-record shape (fixed+rolled same
> day); and **my re-scoped "working candidate" (gate-2 artefact verifiers for the
> design family) was ALREADY DECIDED AGAINST on the record** in
> `verifier_coverage_test.go`'s `itemTypesWithoutVerifiers` — browser-needed checks are a
> standing objection, `needs_design_review`/`spacing_fix`/`responsive_fix` are
> `catJudgement`, and the 4-producer population (1,296 rows) is exactly `213`'s defect.
> "Whoever still wants it owes an argument against a specific recorded reason."
> **What 302 leaves for THIS lane:** of our 4 evidence rows only `dark_section_audit`
> is now gated — `needs_design_review` ×2 needs a SEMANTICS ruling (a blob may BE the
> deliverable for a review type), `spacing_fix` needs its success envelope MEASURED;
> both are owner/ruling-shaped, not quick fixes. And refused rows are NOT tidied up: the
> design-audit carrier has been `enabled=false` since 08-11 (their follow-up 3).
>
> **2. The copy-quality thread ANSWERED — and the answer INVALIDATES the seeding plan
> below.** Read `CONTRIB_2026-08-18_answers_from_copy_quality_two_stage.md` in this
> directory in full before touching the page. The load-bearing finding [MEASURED]: **the
> `site_specs` `voice` aspect does NOT reach the writer — it feeds the DETECTOR**
> (`check_voice_tells.go`); across 1,338 writer calls `tone_guardrails` rendered 0 times
> vs a 214-hit positive control. Writing "friendly-expansive" there changes what gets
> FLAGGED, not written. **Where register actually lands:** (a)
> `identity.key_differentiators` drives the LEAD — write ours as GAINS (their 08-12
> finding: a differentiator written as a subtraction makes the writer lead with a
> loss); (b) `content_direction.example_phrases.characteristic` — **EXEMPLARS beat
> rules; encode friendly-expansive as 2–3 model sentences, not adjectives**, and ⚠ the
> writer reads `content_direction.formatted`, NOT the array — editing the array alone
> is inert; (c) `strategy.tone`. ⚠ finetuning.uk carries **two DEAD voice aspects**
> (`tone_of_voice`, `voice_and_tone` — zero readers anywhere) — tidy. ⚠ Stage 2
> (`copy-editor`) is site-agnostic and guarantees no figure lost/invented (**£99 cannot
> silently drift**) but has NO notion of price/claim/CTA — claims truth stays with
> `evidence_base`; and **declare `required_links` in the seed or one arm of its gate
> passes vacuously**. `copy-editor` is experimental, undispatched, parks at
> `needs_human_review` (no surface, `bugs_open/033`) — a tool you run on purpose, not a
> backstop. ⚠ Their `bugs_open/305`: the writer still emits define-by-negation at a
> non-zero rate — **plan to CHECK the output, do not assume the spec suppresses tells.**
> They ask for the outcome either way, dated via `llm_call_log`.
>
> **3. Offer-analysis thread has NOT replied** (their 08-18 handoff predates our ask;
> our CONTRIB sits unanswered in their dir). Benefit ordering remains open; do not
> wait on it indefinitely — the copy-quality answer says the LEAD is
> `key_differentiators[0]`, so the ordering question is now concretely "what is
> differentiator [0], written as a gain".
>
> **Fleet:** `v1.0.1314` on both thunder-adapter and chassis (19 Aug); GPU estate still
> paused, vendor `{}`. Nothing of ours was in flight across the rolls.
>
> **Revised next work (supersedes "The next work, in order" below):**
> 1. Seed finetuning.uk's identity + `content_direction` per the copy-quality answer:
>    differentiators as gains, exemplar sentences for register, `required_links`
>    declared, dead voice aspects removed. Through the framework; no hand-built page.
> 2. Run the offer page build; then run `copy-editor` ONCE deliberately; CHECK the
>    output for negation tells (305) and for the unverified "person checks every run"
>    claim; report the exemplar outcome back to `copy_quality_two_stage`.
> 3. Phase 1 payment link + concierge; owner calls still open (booking shape, sample
>    datasets, Stripe posture). Offer-analysis reply: ingest when it comes.
> 4. NOT this lane's: 302's follow-ups (semantics ruling, envelope measurement,
>    carrier re-enable) — they are rulings/other lanes; cite, don't redo.

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
