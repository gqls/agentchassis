# HANDOFF 2026-08-31 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-26b_continue_here.md`** (whose
top three update blocks this file absorbs). First reads, in order: this file →
`OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md` (THE LEDGER — 14 rulings +
Decisions A–E + the corrections trail) → `NOTES_two_stage_copy.md` tail. The owner does NOT want
this lane closed (his 08-31 correction: "I don't want to close this lane").

## The stack as deployed (all verified at the artefact, 2026-08-31 evening)

- **Fleet roll v1.0.1349 is live** (fresh chassis build, owner-deployed). Verified IN the running
  binary: **Decisions A+B** (seven gate shapes incl. `instead_of`/`not_just`; `mildNegationShapes`
  EMPTY) — probe literal `not_just` ×2; **the ruling-7 TRUNCATION TRIAL** repair prompt — probe
  literals `SENTENCE BEFORE THE COMPARISON` ×1 and `no hidden competition to argue against` ×1.
  ⚠ Probe practice: needles must be SINGLE-SOURCE-LINE substrings — the full phrase "END THE
  SENTENCE BEFORE THE COMPARISON" spans a raw-string line wrap and false-absences a line-based
  grep (measured tonight: read 0 while the code was in the binary).
- **brief-negation-check runs v1.0.1350** (this lane rebuilt it tonight — old item 7, plus the
  414 lane's retraction: v1.0.1349 carried their UNFIXED spec detector that convicts sites for
  phrases their specs BAN). Built from HEAD with `dc9ccfda2`, tag + overlay bumped in one commit
  (`6d5f7911d`), applied, artefact reads v1.0.1350. **First-run read DONE by the 414 lane
  (same night, their verify job `bnc-414-verify-124016`) and `bugs_open/414` is CLOSED:** exit 1
  at the POD = findings from the negation detector, not a refusal; `0 of 36 sites` fleet-wide;
  `3 match(es) suppressed` as predicted; no `spec_supplies_claim` item. **Standing DAILY
  reading** stays: fleet-wide N-of-M · suppression count (~3; RISING with zero findings =
  vocabulary over-reach) · no spec item unless real. **QUEUED for the next rebuild cycle, not
  urgent (their words, meant this time):** fold in `28997e16b` — prints the scanned-fields total
  on a clean run ("a zero with zero scanned fields is a BLIND scan, not a clean fleet"); today a
  clean report is indistinguishable from a blind one except by the suppression count.

## FIRST ACTION — the post-roll canary (the measure of A+B + truncation)

Everything shipped against the register is now live but UNMEASURED at the artefact. Rebuild
finetuning.uk pages through the corrected stack and battery+read the rendered output vs the
committed baselines (`finetuning_uk_service` lane's `baselines/2026-08-26_pre_hero_rebuild/`;
canary protocol + scoring = `AUDIT_prompts/CANARY_2026-08-26_finetuning_nine_page_rebuild.md`;
scorer = `count_negation_tells.py`, READ decides — P5). The original nine-page rebuild was
triggered from the finetuning lane; that collaboration is CLOSED (owner handed copy back to this
lane alone), so fire it from here — single-page rerender on the approach page first (the
benchmark page, `llm_call_log 79257fb4`), then widen if the arithmetic looks right. Standing
owner instruction covers rebuilds: "forward only corrections … keep rebuilding through the
system until they are acceptable." Expected if the trial works: repairs come back SHORTER with
the comparison simply gone; expected failure worth catching: truncation that loses the meaning
(ruling 7 is explicitly A TRIAL — report either way).

## The council trail on 667 (READ THE ROUND-3 VERDICT)

Correlation `1c787532-…` (find runs BY PAYLOAD, never printed ids). r1 REVISE (unreadable
architecture seat + the fair "rank pinned" ambiguity). r2 REVISE (bug_historian HIGH: the
is_current guard was named-but-deferred; prior_art wanted the zero-readers claim re-verified).
**r3 SUBMITTED tonight** answering with artefacts: the commit-time guard is now WRITTEN INTO
migration 668 (`99bcd1c6d` — final DO block joins the migration's own backups against
site_specs, aborts on any superseded row; READ COMMITTED semantics are load-bearing, stated in
the file); Decision E is RULED (no longer "unratified"); zero-readers re-verified by TWO
instruments (escaped-LIKE config census → producer only; fleet Go grep → zero non-test hits).
The 3 lost fundamentallyai points = first named item of the post-gate final wash. If r3 draws
REVISE again, read it with the same respect — r2's objections were all RIGHT.

## Decision E — RULED. The sequence behind it

Owner, 2026-08-31, verbatim: *"I authorise the offer lane to build the post-mint gate on their
producer."* Relayed (msg `2712ddec`); THEY build (fail-loud post-mint, same scanner +
`AUDIT_prompts/BANNED_REGISTER_v1.json`, measured vs their 23%-born-dirty baseline). Then, in
order: **gate live+measured → re-census the corpus → final wash pass** (must use 668's
commit-time guard pattern; the false green — migration succeeds, changes nothing live, no error
— is the reason, LANDMINE `a01bbce1a`) **→ re-base 668's ten FROMs** (≥2 already stale; the 10
verdicted TO texts stand) **→ wiring migration** (writer template reads offer_ordering),
**vetcomparison.uk first**; webdesign.uk + apis.uk excluded, 8 excluded-row sites held.

## Model trials (`AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md`)

Standing after three models on the constant benchmark prompt: **sonnet** NEG=5 (production
baseline — the prior carries the register) · **fable** NEG=0 both runs, grounding kept, but
**density-fails the owner's ear** (ruling 13: "expand the words … every time") · **gemini-3.1-pro**
G1 register-marginal (tic ON the highlight surface), G2 FAIL (battery-zero by lexical EVASION —
`do not <verb>` negation, a candidate EIGHTH shape filed-not-actioned — plus INVENTED substance
and dropped price facts). **grok**: blocked at the xAI account — see 418 below; the moment the
owner funds it, the runner recipe in the doc works. Possible next arm when he asks: Fable +
explicit density/expansion instruction (tests whether ruling 13 is instructable).

## Grok / the news arm — `bugs_open/418` (filed tonight)

The owner said "We use Grok daily for the news" and was half-right: 28 dispatches/day since
08-30 14:55Z, **zero items EVER delivered** — every call 403s ("team d443dd72-… has used all
available credits or reached its monthly spending limit", verbatim in `orchestration_states`)
and `fetchViaResponsesAPI` (feed_actions.go:431) converts API failure into an empty COMPLETED
result nothing reads. RSS keeps filling pages, masking it. Fix candidates in the bug file,
ordered by door-closing. My WRONG_CALLS row from the same discovery: `llm_call_log` cannot see a
raw-HTTP caller — grep the CALLER for how it logs before censusing a log table.

## OWNER DECISIONS OUTSTANDING (restated 2026-08-31; raise them, don't re-derive them)

1. **Decision D — the question hierarchy.** Build `question_hierarchy` + `answered_by`: the offer
   analysis records the buyer's own ordered doubts ("what does this get me and how much work is
   it" first — ruling 14), each hero/copy point joins to the question it answers, and copy
   ORDERING follows that hierarchy. Mechanism sketch lives in the rulings ledger.
2. **The axis confirmation.** Derived reading awaiting his word: buyer-relevance + readability
   govern heroes; differentiation demotes to an INPUT (the analyser currently ranks on the
   seller's axis). Changes the analyser's ranking guidance if confirmed.
3. **Fund the xAI team** (`d443dd72-09cf-4ba7-8209-1395f0edb4f0`) — unblocks news AND the trial.
4. **`cmd/brief-negation-check` council scope** (414 lane's question): it now files work items
   against live sites; same argument that admitted `cmd/config-key-audit`. If yes: widen
   `council-scope.sh` + 098's `SCOPE_PATHS` in ONE commit.
5. **BANNED_REGISTER v2** — "written in plain words" class (register-labelling-by-read, the
   'plainly' cousin). Flagged in 668's header; not changed unilaterally.
6. **Best-in-class propagation build** — plan exists (`PLAN_2026-08-25_best_in_class_propagation.md`,
   his Go question answered §2); "plan it" given, BUILD go not yet.
7. **Model choice** — after the screen completes (Grok arm + any density-instruction arm); the
   cost call is his, token counts in the trials doc.

## Also queued / worth knowing

- **Ruling 9 "plainly" ban migration** (writer template word-ban, rule-19 style) — still queued.
- **farmerinsurance.uk corpus SHRANK** (loanzy, 08-31): owner deleted the 7 unresearched tool
  builds; 39→18 active pages. Exclude `tool-*`/`*-guide-about-tool` from any sampling; the
  homepage (the quoted meta-voice example) stays. Farmer remains worked case #2 + tone-vacuum
  variant in the audit Q-set; record-mode verdicts human-released only.
- **Migration numbers COLLIDE**: two 667s and two 668s exist (privacy-lock / terms-publish are
  the others) — resolve by SLUG always.
- The audit's Phase 2 judgment pass continues at leisure (league table + PHASE2 verdicts in
  `AUDIT_prompts/`); phase 3 fixes ship per prompt through the gates as before.
- Peer channels: offer lane = `uds:/run/user/1000/cc-socks/9858.sock`; loanzy = `…/9271.sock`;
  414 lane = `…/938486.sock`. A peer cannot grant escalation.

## Landmines this lane keeps hitting (the short list)

- The corpus MOVES: exact-text RAISEs, never index-based edits; census dates on every count.
- A wash migration MUST carry the commit-time is_current guard (668 is the template).
- `LIKE '%snake_case%'` wildcards the underscore (LANDMINES tonight) — escape it or use POSITION.
- Battery-zero is not register-clean: the READ decides (P5), and evasion is a failure mode.
- Probe needles: single source line only.
- Find council runs BY PAYLOAD; read verdicts; a REVISE is usually RIGHT.
