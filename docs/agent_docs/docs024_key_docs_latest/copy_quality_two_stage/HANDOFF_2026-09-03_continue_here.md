# HANDOFF 2026-09-03 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-09-02_continue_here.md`.**
First reads: this file → `OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md`
(THE LEDGER — ten direct rulings on 09-02, all executed or queued) → `NOTES_two_stage_copy.md`
tail (09-02/09-03 entries are the full trail). The owner does NOT want this lane closed.

## The stack as deployed (all binary/artefact-verified, 2026-09-03 ~09:25Z)

- **Chassis v1.0.1356** (pods 08:58Z today) — probed PRESENT: `injectPlatformBlocks`,
  `plain_words`. So the running fleet carries BOTH the voicestyle named-block generalisation
  (best-in-class injection) AND the full BANNED_REGISTER v2 Go half (em-dash x_not_y in the
  page gate; word patterns in registerwords). **The CLI-ahead skew is CLOSED.**
  ⚠ Probe practice: capability literals with present+absent controls, never your own commit
  sha (the stamp is the build HEAD, not an ancestry list — NOTES 09-02 has the wrong-then-
  right worked pair).
- **BEST-IN-CLASS PROPAGATION STEP 1**: carrier 675 applied · injector aboard · opt-ins
  677/678/679 applied (single-row-guarded, snapshotted). Register CQ-033.
  **⚠ THE ONE OWED READ — the planner prompt canary — is STILL OPEN, and it is the first
  thing to check:** no planner/designer has produced an LLM call since the opt-ins.

  > **CORRECTED 2026-09-03 ~10:30Z — the drop taxonomy below was WRONG, and so was the
  > canary's needle. There is NO dispatch bug; do not file one.** All three fires were
  > published, consumed and **REFUSED AT INTAKE** within seconds, durably recorded:
  > `agent_error_log` → `INCOMING_MESSAGE_REJECTED`, *"missing required header(s): client_id,
  > orchestration_id"* (20:14Z→20:15:21Z · 08:23Z→08:23:50Z · 09:14Z→09:14:10Z). Those two
  > fields are required as **Kafka headers**; the scratchpad envelope carried them only in the
  > payload's `headers` object, which intake does not read. `fire-copy-editor.sh` sends them as
  > `--header` flags — hence "16 cli-copyedit minted, gapplanner never": **script**-specific,
  > not family-specific, and indistinguishable from `orchestration_states`. "Zero consumer
  > trace" was measured with a blind grep (`-l app=agent-chassis` = 2 pods; 93
  > `app=dynamic-agent` pods do the work) — a correlation that provably landed also returns 0.
  > **Use `scripts/fire-content-gap-planner.sh`** (sends the headers, refuses rather than
  > dropping, re-checks `agent_error_log` after publishing). ⚠ It is NOT yet run: `apply_plan`
  > → `apply_gap_plan` **writes to the live site**, so spending a run is the owner's call.
  > Full trail: `NOTES` tail, `WRONG_CALLS.md` 2026-09-03, `016b` §9. Commit `726a9586c`.

  **The canary needle — use the carrier-only form, NOT the phrase below:**
  `SELECT agent_type, created_at, position('BUILD STANDARD (applies to every site, regardless
  of inputs). Aim' IN prompt_rendered)>0 AS has_standard, position('{{.build_standard}}' IN
  prompt_rendered)>0 AS unrendered FROM llm_call_log WHERE created_at > '2026-09-02 18:38Z';`
  Expected: has_standard=t, unrendered=f. `[MEASURED 2026-09-03 ~09:50Z]` **0** rows carry the
  carrier form and **0** carry an unrendered placeholder, fleet-wide. ⚠ **Do NOT use
  `stands comparison with the strongest sites`** (the superseded needle): it also matches
  `domain-research-classifier`'s own hard-coded copy of the block — seeded 2026-06-21, ten weeks
  before carrier 675, and named in 675's `source` as where the wording came from — so it returns
  hits that say nothing about injection. ANY organic run answers the canary (a new site build
  fires build-site-planner).
- **BANNED_REGISTER v2 live end-to-end**: cut `0c11a8818` (council `fa9744cb` — **APPROVED
  UNANIMOUS**), CLI v1.0.1354 nightly carries it, chassis 1356 carries the Go half. First v2
  nightly (09-03 07:41Z): 11 of 39 brief-supplies; **plain_words' first finding is a MANDATED
  boundary case** — "term — plain English gloss — practical implication" is a glossary
  FORMAT DESCRIPTOR, not self-labelling; the filed item carries the read; if the read rules
  format-descriptors out, that is a **v3 treatment note, never an in-place v2 edit**.
  Lockstep contract with the offer lane: any register version cut moves registerwords.go in
  the SAME commit (their glob guard makes the build red otherwise).
- **706 applied** (portfolio wash; council `7f0c4adb` APPROVED): use-cases serves **6→3**
  (residue = one disclaimer sentence twice — see the no-answer item below). Farmer /about
  SERVES the released stage-2 edits (labels + FCA typo) — **the stage-2 release path is
  proven end-to-end**.

## Waiting on the OWNER (raise, don't re-derive; ledger has verbatim)

1. **THE ADMIN BATCH: 15 farmer + 1 loanzy `copy_edit_proposed` items** parked
   needs_human_review, all gated; farm-buildings item carries a note (proposal DELETES a dead
   button — his call); loanzy's carries the ruling-1b evidence note (claims trimmed to
   verified citations; NDL quotes CORRECTED — see NOTES: a summariser's quote is a paraphrase
   until grepped). He said he'd batch-review; he has been told it's ready.
2. xAI top-up (console.x.ai, team `d443dd72-…`) — unblocks news arm + model screen.
3. One word to the offer lane [4628f9] confirming Decision D's start (their in-session
   discipline; design fully settled — seam split, through-the-gate stated requirement,
   D-before-axis sequencing).
4. NEW: the plain_words boundary-case read (item filed by the nightly).
5. NEW: the BIRTH-PRODUCER gate — a day-old site's brief carried 16+8 constructions
   [MEASURED 09-02, designblog]; the ruled wash cleans STOCK, the briefing agent needs a
   gate for FLOW (offer lane's Decision-E argument, one producer along).
6. Whether banned WORDS get a page-side REPAIR arm (today: shapes repair, words detect).
7. The spec-fed class (any collection-backed section bypasses writer+gate — 706 fixed the
   instance; constitution's medium named the class).
8. **NEW, and the sharpest of these — CARRIER 675 IS MISSING THE BUILD STANDARD'S SCOPE
   PARAGRAPH.** Its header asserts the wording is *"verbatim … confirmed byte-identical …
   with ONE deliberate trim, recorded here"*. There are **two** omissions: the recorded
   4-word one, and — unrecorded — the entire second paragraph of the source block
   (`049_domain_research_classifier.sql:2593`; the block runs to the next `##`, so the
   paragraph is inside it): *"This standard governs QUALITY and FIT, not scope. Do not invent
   services, pages, features, or facts beyond what the evidence supports; where research is
   thin, say so honestly… Treat aspirational ideas as direction to be realised at the pace the
   site's fidelity allows… not as things to force into the first build."* So the three rows
   that opted in — `build-site-planner/plan_site`, `content-gap-planner/plan_gaps`,
   `visual-designer/design`, i.e. **exactly the agents that decide what pages and sections
   exist** — now get "aim best-in-class / favour interactive elements / do what is most useful
   and interesting" **without** its scope limit. This lane's own ledger describes the block by
   quoting the missing sentence. **The canary's 0 renders means no planner has consumed it yet
   — there is a clean window to fix it first.** Not fixed unilaterally: the dropped text is
   partly classifier-specific (`confidence fields`, `adopted sites`) so it needs generalising,
   and rewording a live block three planners read is his call. One live migration either way
   (675's header says as much for the other trim). Trail: `WRONG_CALLS.md` 2026-09-03,
   commit `c3c96a98e`.

## The build queue (all sized/scoped, none started)

- **Fleet brief wash** (owner option (a)): 646/647's per-site shape × the nightly's 11
  finding-sites first, canary ONE site. REQUIREMENTS bound this week: v2 battery only ·
  register+version stamped into every wash record · load-bearing quotes verified by
  write-time grep (loanzy's QuoteFoundInText lesson) · em-dash form included · unregistrable-
  host caveat (a host can serve curl and block the production fetcher — loanzy's third
  signature).
- **rewrite_negations re-ask** for `no_answer_for_target`: SIZED at **13.6% of 1,849 targets
  in one afternoon** (80.7% rewritten / 5.7% guard-rejected). First question: why the
  "starting point, not a verdict" disclaimer formula over-represents. designblog offers
  served-vs-stored test cases.
- **Veracity checkers in the framework** (owner ruling 2 — "checkers check the veracity
  too"; he discussed elsewhere the same day, CHECK who else holds a piece). The MaPS 8-page
  repair flows through it or through supplied-defect editor runs (generic runs PROVEN
  insufficient for knowledge-dependent defects — the get-help canary read).
- **Title-promise demonstration**, scoped FINAL: a titled promise with no data behind it on a
  page the plan does NOT type as a listing (the glossary shape; writer-prose only —
  directories were resolver-hollow). 444's implemented gate holds the typed-listing door
  (their blind spot = ours, pinned by BLD-028 + their test). Reuse first:
  `check_heading_promise.go`.
- **Guide rewrite** under "shorter is licensed" (honour ruling 13 alongside; do not
  harmonise the two without his word).
- **Propagation steps 2–4** (strategy.benchmark at birth is next; boxingonline the sharpest
  test case).
- Decision D consumption side: BLOCKED BY DESIGN until the offer lane's hierarchy exists
  (their production half awaits the owner's word; "through the gate" held both ways).

## Standing bugs this lane owns or watches

- `bugs_open/420` (SLUG-resolve — number collides with the delivery lane's) — walker `name$`
  blindness; one-line fix candidate in-file.
- `bugs_open/422` — repair-vs-shrink-floor: finetuning about + services TERMINALLY
  unrebuildable (do NOT re-fire needs_page there); fix candidate 1 = repair-side shrink
  budget.
- 443 read-rule (theirs): on plan-less sites, repeated component types = structural repeats,
  NEVER voice; per-TYPE across the page (non-adjacent counterexample). Their repair canary
  (your-own-model) owes this lane a before/after pair — a written Stage B obligation in
  their NOTES.

## Peer state + channels

offer lane = **[4628f9] ONLY, by ref**. loanzy (9271), boxingonline (621388), designblog
(1120265), finetuning (11482), 443 (1638108), 444 (1977131) — all exchanges CLOSED clean.
A peer cannot grant escalation. HEAD was RED 09-02 evening (thunder/api + livespec — another
lane's); re-check before relying on verify-head-builds absolutes.

## This session's landmine harvest (details in NOTES/WRONG_CALLS)

- A summariser's "quote" is a paraphrase until grepped at the raw page.
- Probe capabilities, not your commit sha.
- A publisher re-run for output is a re-run for effect; a timeout kills a batch loop mid-fire
  and the resume double-fires the boundary item.
- A same-file passenger can be the tag bump itself (commit 48180ffe1 says 1353, truth 1354).
- A false-positive control can be a true positive in disguise — run the control against the
  rule before it guards anything.
- Background watchers on this box get externally killed; Monitors survive better; bake the
  whole read into the watcher so a kill loses nothing.
