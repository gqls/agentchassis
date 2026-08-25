# 288 — the evidence register guards COPY, not CODE: a figure a calculator ENCODES is checked by nothing, and the one mechanism that could was unreachable

**Filed 2026-08-16** as the CLASS behind `bugs_closed/225` (the SDLT calculator that
ran an expired £625,000 first-time-buyer cap for sixteen months). 225 itself is fixed
and live; this file is what 225's own section *"Why no existing check could ever have
caught this"* was pointing at, written down as a tracked defect instead of a remark
inside a closed case.

**Status: PARTIALLY FIXED. Phases 1 and 2 landed 2026-08-24** (commits `995b5fbbe`,
`eecd99b0a`; council corr `67643b47`, **APPROVED at round 1**, both real advisories acted
on rather than banked). Pieces 2+3 landed 2026-08-16 (`cff364b8`, round 3). **All Go is
inert until the next chassis roll.** §5 carries the residual and §6 the two phases still
owed.

⚠ **THREE CLAIMS IN THIS FILE WERE WRONG AND ARE CORRECTED BELOW — read §0 first.**

⚠ **Two REVISE rounds preceded the approval and each found a REAL defect — read the
council section below before trusting anything here.** Round 1: the mechanism was blind
to its own motivating bug. Round 2: **the round-1 fix was inert**, defeated by a
baseline fallback onto register history. A green after one round would have shipped a
check that could never fire.

## §0 — CORRECTIONS, 2026-08-24 (this file asserted three things that were not true)

**C1 — "Validator rule P11 refuses a malformed declaration where it is written" was
FALSE for every document that carries one.** §"What was built" says it below, the lane
PLAN says it, and the round-3 council submission says it. P11 is real, tested and
correct — and it lives in `ValidateExperienceCriteria` (`experience_criteria.go:320-333`),
whose **only** production caller is `write_experience_pattern_action.go:217`, the
*experience-pattern* register. Tool PLANs are written by `WriteDocPlanAction`, which was
a pure supersede-write of a body string with **no criteria validation of any kind**. P11
had never once seen a tool fence. **The lesson is not the fix — it is that three
documents asserted a control's existence and none of them enumerated its callers.**
FIXED 2026-08-24: a tool PLAN whose fence declares `facts` is parsed before the
transaction and refused if unreadable.

**C2 — the parser FAILED OPEN on a fence that DID declare something, and that disarmed a
safeguard this council asked for.** `parseCriteriaFacts` returned `(nil, nil)` on a
whole-fence JSON error — no ids **and no issues** — against `criteria_facts.go`'s own
header contract. It propagated: `planSiteFactDrift`'s zero-rows warning (added at the
`bug_historian` seat's request in round 3) is gated on `issues`/`unresolved` being
non-empty. **So one trailing comma in a fence edit disarmed the declaration AND the
warning whose whole job is to say a declaration was disarmed** — on the ONE live
declaration, which guards a UK stamp-duty tax calculator. FIXED 2026-08-24, and the
finding now files a `fact_declaration_broken` doc_note per (tool, site).

**C3 — §5.5's deferral cites `bugs_open/093`, which CLOSED on 2026-08-17** — the day
after this file was written. It is `bugs_closed/093` now, and **the way it closed is the
template for the work this file defers**: its fix was not a second scanner but a second
SURFACE for the existing one, inside the existing post-deploy audit
(`check_unverified_claims_stats.go`). So the prose half of the class needs re-grounding
on that precedent, not silent inheritance of a blocker that no longer exists.
(`a-closed-blocker-keeps-being-obeyed` — grepping `bugs_open/NNN` while NNN is closed.)

**C4 — §5.4's "the `improve_tool` arm may be unreachable on today's fleet" is too
pessimistic, and the measurement is cheap.** `[MEASURED 2026-08-24]` **91 of 178** tool
pages on register-bearing sites carry a tool-level component, so `isFork()` holds for
just over half the exposed population. The arm is unreachable **for the two SDLT tools**
(both multi-component, both `no_auto_fix`), not for the fleet. Say the former.

## Diagnosis provenance (2026-07-31 owner ruling)

**Not run through 090 — stated substitution, and the substitution is stronger than
usual here because there is no inference step to check.** Every mechanism below is a
single visible clause in code I read directly and quote inline with its file, plus
five counts I measured myself against the live database **with a positive control in
the same breath** (the `?` operator that returned 0 for `artifact_check` returned 61
for `citation`, so the zero is an absence and not a broken query). Same ground as
`bugs_open/281:8-12`. What a diagnosis loop would add — an independent re-read of the
same functions — is exactly what the quotes below already are.

## The class, in one sentence

**A fact in the evidence register constrains what a page may SAY. Nothing connected a
fact to the constants inside a calculator's JavaScript** — so a tool can compute a
number that contradicts the very register the site is governed by, and every check we
own passes it.

## Why every existing check is blind, with the clause that makes it so

Three of these are deliberate design decisions, correct on their own terms. That is
what makes the gap hard to see: nothing is broken.

| # | mechanism | the clause | consequence |
|---|---|---|---|
| 1 | every claims scan | `datahelpers/claims.go` `nonAssertionElements` — `"script": true` | a figure inside `<script>` is never text, so it is never an assertion, so it is never scanned |
| 2 | the numeric arm | `claims.go` `editorialPageTypes` — `"tool": true` | `ScanUnregisteredNumbers` does not run on a tool page's prose at all |
| 3 | prose currency | `claims.go` `isExcludedNumber` drops `£`/`$`/`€` amounts | even `£625,000` stated in FINANCE PROSE is scanned by nothing |
| 4 | the golden | `loanandmortgagecalculator_couk/golden_compare_post.py` | proves CONSISTENCY; a baseline recorded from a never-correct tool certifies the defect for ever (016b §9) |
| 5 | Tier-2 acceptance | `check_tool_acceptance.go` header: static checks *"CONFIRM, never refute"* | structurally cannot fail a page for what it computes |

Nothing in that list is asking whether the answer is RIGHT. The only thing that ever
has is `SQAM-003`, a Python oracle in a docs directory with no scheduler, no work
item and no Go caller — its own register entry says so, re-confirmed 2026-08-15.

## The mechanism that COULD have, and the three reasons it could not

RFC_025's `artifact_check` (`refresh_evidence_base_action.go`) re-proves a fact
against stored `rendered_html` — raw bytes, scripts included. It is the one reader in
the estate on the right surface. It was unreachable for this class:

1. **Zero consumers.** `[MEASURED 2026-08-16, with a positive control]` — of **143**
   current facts across 12 registers, **0** carry `source.artifact_check`; 28 carry a
   plain `artifact` source, which is checked for presence in the register and never
   re-proved. RFC_025 §10 predicted this ("still short of IMPLEMENTED… no real fact
   anywhere has actually been retyped").
   > **RE-MEASURED 2026-08-24, same query, same control:** **294** current facts across
   > **15** registers, still **0** carrying `source.artifact_check`, against **185**
   > carrying `source.citation`. The register more than DOUBLED in eight days while
   > adoption of the one mechanism on the right surface stayed at zero — the census went
   > stale by ADDITION, exactly as the count-carries-its-date rule predicts, and the
   > direction of travel is the finding: **the gap is widening, not closing.**
2. ~~**Unreachable for the facts that hold legislated figures.**~~ **FIXED 2026-08-24
   (Phase 2 / RFC_025 stage 2b, `eecd99b0a`).** The per-fact loop handled
   `source.citation` and `continue`d **before** the `artifact_check` test seventeen lines
   below; every SDLT fact is a citation fact, so an `artifact_check` beside one was never
   evaluated. `[read directly]` The `sql` and `attested_by` arms sat below the same
   `continue` and were equally unreachable for a dual-source fact. Now a **pre-pass**
   evaluates it for every source kind, leaving the existing branches and their `continue`
   untouched. **A secondary check never touches `verified_at` or `changed`**, so the
   RFC_025 raise gate is unchanged for existing callers.
3. ~~**Addressed by a component id that dies.**~~ **FIXED 2026-08-24 (same commit).**
   `artifact_check.component_id` names a `page_components` row, and **225's own component
   (`55682bc8-…`) no longer exists** — the page was decomposed into `prose-0` / `tool-1` /
   `prose-2` — so a binding written when the bug was filed would today resolve to nothing
   and fail closed **for ever, which reads exactly like a working check**. A fact may now
   carry `artifact_check.subject_key` instead, resolved through the platform's own name
   rule every pass (`discovery_checks.ToolSubjectKeyExpr` plus the audit page predicate,
   the same join family as the fan-out). Exactly one address; both, or neither, fails
   closed. Verified viable: `(site_id, name)` is unique across all **178** tool pages
   `[MEASURED 2026-08-24]`.

## What was built 2026-08-16 (Pieces 2+3), and what it does NOT do

Commit `989addb1c`, council-submitted `cff364b8`. Concept register **CLM-022** (+
**TL-045**). This is not a new design: it is Pieces 2 and 3 of
`mortgagecalculator_couk_adoption/PLAN_2026-08-09_facts_into_tool_acceptance.md`,
whose Piece 1 has been live since migration 366 (CLM-021).

- **Piece 2** — a tool's criteria fence may declare `"facts": ["<fact_id>", …]`: ids
  only, never values, resolved against the SWEPT SITE's register (doc_plans has no
  `site_id`). Validator rule P11 refuses a malformed declaration where it is written.
- **Piece 3** — the daily evidence sweep resolves declaring PLANs to pages and files
  one item per (fact, tool): `improve_tool` only where a fixer may safely act (a
  tool-level component owns the code AND the fence has no `no_auto_fix`), the new
  handler-less `fact_drift_review` otherwise — including for **every** evidence-only
  drift — and nothing at all on a fetch error.

**Read the honest limits before quoting this as closed:**

- It answers *"did the register's figure MOVE, and who encodes it"*. It does **not**
  answer *"is the register's figure RIGHT"* — that is Piece 4 (an oracle computing
  expectations from the register), which stays behind its RFC. **A tool and a register
  that are both wrong in the same direction agree, and this mechanism is silent.**
- It needs a human to author every declaration. Today **0 of 90** current tool PLANs
  declare anything `[MEASURED 2026-08-16]`.
- It cannot see a page with no `pages` row. mortgagecalculator's repo-only
  `stamp-duty.html` twin carried the identical defect and stays invisible.
- Latency is up to 24 hours (the daily sweep), and a fence carrying `facts` asserts
  **nothing** at acceptance time — both tiers ignore unknown fence keys, by design.

## §5 — the residual, which is why this file stays OPEN

1. ~~**Not live**~~ → **LIVE 2026-08-17, and the first production sweep ran clean.**

   | check, 2026-08-17 | result |
   |---|---|
   | `fact_drift_review` in `/proc/1/exe`, both replicas | **2** (was 0) |
   | `unreconciled_declaration`, both replicas | **1** |
   | `stale_attestation` (positive control), both replicas | 5 |
   | strings unique to the final approved revision (`6b3b0510e`) — `"may have stopped matching"`, `"DISTINCT ON (dp.subject_key"` | **1 each**, so the binary carries the post-council version, not an earlier one |
   | `evidence-freshness` scheduled task | ran **09:04:14Z** with this code |
   | register revisions written by that run | 8 |
   | **errors from that run** | **0** |
   | `fact_drift` work items fleet-wide | **0** |

   **The 0 items is NOT evidence the mechanism works, and must not be quoted as
   such.** It has no demand behind it: no fence declares `facts`, so there is
   nothing for the fan-out to act on. What the run DOES prove is the no-op case —
   the new query path executed against all 13 register-bearing sites in production
   and broke nothing. The mechanism firing is still unproven live; see §5.2.

   ~~Pre-roll baseline, kept for the record — MEASURED, not assumed.~~ `[2026-08-16, both replicas, with a positive
   control in the same exec]`:

   | replica | `fact_drift_review` | `stale_attestation` (control) |
   |---|---|---|
   | `agent-chassis-…-48lv6` | **0** | 5 |
   | `agent-chassis-…-vtfdx` | **0** | 5 |

   The control is what makes the zero mean something: the probe CAN find a symbol in
   this binary, so `0` is an absence rather than a broken grep. Re-run exactly this
   pair after the next roll and expect the first column to become non-zero. Do not
   verify at the tag, and never by the absence of errors.
2. ~~**No live declaration**~~ → **DECLARED AND PROVEN TO FIRE, 2026-08-17.** The
   `mortgagecalculator_couk_adoption` lane seeded `stamp-duty` with **all 13** SDLT ids
   (not the 225 pair — `verify_criteria.py:load_register_bands()` hard-fails on a missing
   id, so declaring two would understate what the tool encodes), and ran the induced
   proof at 16:17Z in a 14-second window with the restore in a `trap … EXIT`.

   | run | `fact_drift` entries | kind | `new_value` for the changed fact |
   |---|---|---|---|
   | baseline (register 500000) | 13 | `unreconciled_declaration` | **500000** |
   | induced (register 550000) | 13 | `unreconciled_declaration` | **550000** |

   **PROVEN:** the fan-out resolves a declaration on a tool the acceptance ladder cannot
   see (2 components, 0 tool-level — the exact case the fan-out refuses to key
   `toolEligibilityWhere` on), routes all 13 to `fact_drift_review` (correct for a
   `no_auto_fix` fence), and **reads the register at check time** — the two runs differ in
   exactly the one value that changed. Register restored to 500000, `pinned` carried, 0
   items written (a dry run writes nothing, as designed).

   **NOT PROVEN, and structurally uninducible today:** the `value_drift` arm. On a fresh
   declaration every pair is `never_reconciled` and that arm wins; a dry run cannot create
   the baselines that would let `value_drift` win. **It becomes inducible only after one
   REAL sweep files the 13 one-time items.** ⚠ My own RUNBOOK predicted `value_drift` here
   and was wrong — corrected inline, and in `WRONG_CALLS.md`.

3. ~~**Still owed: induce `value_drift`**~~ → **DONE 2026-08-17 18:30Z. CLM-022 IS NOW
   FULLY PROVEN — both arms and the self-quieting property.**

   A real sweep filed the 13 items at 18:16Z (`created_by=generic`, all
   `unreconciled_declaration`/`never_reconciled`, severity low, priority 60, **handler
   empty** — the handler-less shape the guardian seat asked for). That recorded the
   baselines, which is what made the last arm reachable.

   | check | result |
   |---|---|
   | **self-quieting** — dry run, baselines set, register unchanged | **0 entries** (13 facts checked). The identical run before the baselines existed produced **13**. |
   | **`value_drift`** — baseline moved 500000→450000, register untouched | **1 entry**, `kind: value_drift`, `old_value: 450000`, `new_value: 500000`, `route: fact_drift_review` |
   | **per-fact isolation** | exactly 1 entry; the other 12 declared facts silent |
   | **nothing written** | 13 items before and after; 0 of kind `value_drift`; register still 500000; item spec restored **byte-equal** to its pre-image |

   **The register was never touched.** The comparison is symmetric, so moving the item's
   recorded baseline exercises the same branch without changing a live tax figure on a
   public site. Restore ran from a `trap … EXIT`.

   ⚠ The `reason` came back `not_a_fork`, not `no_auto_fix` — both are true of this page
   and the fork guard is tested first. My RUNBOOK said the latter; corrected. The first is mortgagecalculator's `stamp-duty`
   fence, handed to that lane as a CONTRIB. **Until it lands this mechanism has never
   fired on real data, and today's clean sweep is a pass from a check with nothing to
   check** (`a-pass-from-a-blind-check-outlives-the-blindness`). Measured 2026-08-17:
   **0 of 90** current tool PLANs declare anything. The mcalc lane is ACTIVE today but
   on another front (guides, imagery, dead links) and has not touched the fence since
   2026-08-10 — so the seed was NOT applied unilaterally while they hold the site.
4. **The `improve_tool` arm is unproven in production and may be unreachable.** Both
   live SDLT fences carry `no_auto_fix` and neither page is a fork, so on today's
   fleet every route is human. Proven by unit test only — say so, do not report the
   auto-fix path as exercised.
5. **The prose half of the class is untouched.** Blind spots 2 and 3 above (tool-page
   prose, and currency anywhere) are still blind. A retired-figure scan over raw
   deployed HTML in the post-deploy claims audit would close it, and was deliberately
   NOT built here: it is a second scanner over the same components, which is
   `bugs_open/093`'s shape, and it deserves its own measurement of the false-positive
   rate first (the plan's §3 rejects the blanket form for exactly that reason).
6. ~~**RFC_025 stage 2b is still owed**~~ → **DONE 2026-08-24** (`eecd99b0a`, council
   corr `67643b47` covers Phase 1; stage 2b rides the same lane and is registered under
   CLM-022). Reachable for every source kind, addressable by `subject_key` rather than a
   component id that dies. ⚠ **It still has no consumer: 0 of 294 facts carry
   `artifact_check`** — the mechanism is now *usable* for a legislated fact, which it was
   not, but RFC_025 stays short of IMPLEMENTED until one real fact is retyped. That is
   the first item in §6.

## §5b — LIVE AND PROVEN AT THE ARTEFACT, 2026-08-24 18:49–18:55Z (post-roll)

**All four phases are in the running binary**, each probed with a string first confirmed to
exist in the source (so a zero would have meant something), with both controls:

| probe | count | |
|---|---|---|
| `fact_declaration_broken` (Phase 1) | **4** | |
| `resolves to no stored component HTML`, `carries BOTH component_id`, `component_id or subject_key` (Phase 2) | **1 each** | |
| `present_in_markup_only` (Phase 3a) | **1** | |
| `fact_binding_suggested` (Phase 4) | **4** | |
| `stale_attestation` | **5** | positive control |
| `ZZZ_must_be_absent` | **0** | negative control |

**RFC_025 STAGE 2b IS PROVEN ON LIVE DATA, BOTH DIRECTIONS.** `sdlt-ftb-relief-cap` on
mortgagecalculator.co.uk now carries
`artifact_check{subject_key:"stamp-duty", pattern:"FTB_RELIEF_CEILING\s*=\s*500000"}` — the
constant name read off today's bytes, not from memory, and context-bearing so it clears
`bareNumericPattern`. Written by supersede with a `DO`/`RAISE` verify block that would have
aborted the transaction on any other change; `pinned=t` carried; 13 facts before and after; every
other fact byte-identical.

| dry run | result |
|---|---|
| baseline | **TWO entries for one fact**: `tolerance:"artifact_check"` outcome `fresh`, AND `tolerance:"citation"` outcome `fresh`. **Before stage 2b the first could not exist** — the citation arm `continue`d before the artifact test. |
| the artifact entry's `verified_at` | **absent** — the `ownsVerifiedAt=false` rule holds live: a secondary check does not bump the date the primary arm owns |
| induced (pattern → the EXPIRED `625000`, which the tool does not contain) | `outcome: drifted`, detail *"pattern … no longer found in **tool \"stamp-duty\"**"* — so the durable `subject_key` addressing resolves live |
| the citation arm during the induced run | **`fresh`, undisturbed** — a drifted artifact check does not perturb the primary arm |
| restore | **byte-identical to the pre-image (md5 match)**, restore armed in a `trap … EXIT` before the change |
| written by any dry run | **nothing**: 13 `fact_drift_review` items before and after, newest still 2026-08-17; 1 current spec row; 0 doc_notes |

⚠ **`gd-trials` (gamesdesign.co.uk) also carries an `artifact_check` as of 11:07 today, added by
another lane** — so RFC_025's first consumer is theirs, not this lane's. Ours is the first on a
**citation** fact and the first using **`subject_key`**, which are the two things stage 2b
unlocked. Say it that way; do not claim the first consumer.

## §5c — THE FIRST REAL SWEEP, 2026-08-25 09:06Z. It fired, and it found a defect of mine.

The first non-dry pass with this code. 19 sites, **0 errors**.

**PHASE 4 WORKS ON REAL DATA — five notes across three sites, first pass:**

| site | tool | proposed |
|---|---|---|
| loanandmortgagecalculator.co.uk | **`mortgages-stamp-duty`** | **7** correct SDLT bindings, paste-ready — the estate's second stamp-duty calculator, previously guarded by nothing |
| gamesdesign.co.uk | `drop-rate-simulator`, `tool-combat-balance-comparator`, `tool-spawn-rate-balancer` | 3 |
| agritec.uk | `tool-sfi26-revenue-stacker` | 1 (see the defect below) |

**STAGE 2b IS PROVEN IN PRODUCTION, AND BOTH ARMS OF `ownsVerifiedAt` WITH IT** — the same code,
two facts, two roles, correct in each:

| fact | tolerance | outcome | `verified_at` | |
|---|---|---|---|---|
| `sdlt-ftb-relief-cap` | `artifact_check` | fresh | **absent** | SECONDARY — the citation arm owns the date |
| `sdlt-ftb-relief-cap` | `citation` | fresh | 2026-08-25 | the primary arm |
| `gd-trials` | `artifact_check` | fresh | 2026-08-25 | PRIMARY (artifact-only fact) — so it *does* own the date |

### ⚠ THE DEFECT THE SWEEP FOUND, AND IT WAS MINE

agritec's note proposed **two** fact ids for the single `100000` in the tool's script —
`CIT-3f1b219f15ec6a39` and `CIT-86c4010f7cdf820d`, both asserting the SFI26 annual agreement cap
of £100,000. A human cannot reconcile that binding, and pasting it would declare two facts for one
constant so every later pass owes a reconciliation nobody can perform.

**`factProbeNotProbed`'s own comment already said "refused: no value, below the floor, or
ambiguous". The ambiguity arm had never been written.** A comment promising a guard that does not
exist — the exact class this lane spent two days closing, committed while closing it. Fixed
(`bba8a892d`): a value carried by more than one fact is reported in the note but never proposed
and never reaches the paste-ready fragment. **Not yet rolled.**

### ⚠ A MISPLACED `artifact_check` IS SILENTLY INERT — the Phase 1 defect, one table over

Found 2026-08-25 on agritec.uk, by disagreeing with another lane about their own data.

They wrote four `artifact_check` objects with correct contents, correct patterns, correct
`subject_key`, live in the register — **at the TOP LEVEL of the fact** rather than inside
`source`, where every reader looks (`refresh_evidence_base_action.go:348`, `parseArtifactCheck:768`,
and RFC_025 §2.2's own shape). Read by nothing. No signal anywhere. They had tested the patterns
and reported the fence live; I had queried `f->'source' ? 'artifact_check'`, got zero from a
populated row, and told them it had never landed.

**Neither answer alone was survivable.** Accepting theirs would have closed the loop on a fence
that could never fire; accepting mine would have had them re-apply a migration into the same
silent hole. **The disagreement produced the real answer, and that is the transferable part** —
more than either error is.

**Their placement is a REASONABLE reading and arguably the better data model**: `artifact_check`
describes the FACT, not the source. RFC_025 put it under `source` because it was designed as a
sibling of `citation`/`artifact`/`attested_by`. Two defensible readings, one implemented, and the
wrong one silent — which is exactly the defect Phase 1 closed for criteria fences and which I left
open one table over. Fixed: the sweep now names a fact carrying a top-level `artifact_check` in
`res.MisplacedArtifactChecks` and warns. It does **not** relocate the key — the register is
human-owned and a consumer must not rewrite a human's row (CLM-001).

### ⚠ THE AMBIGUITY RULE MAKES THE SUGGESTER STRUCTURALLY QUIET ON A RATE-TABLE SITE

`bba8a892d` refuses to propose a value carried by more than one fact. That is right, and the
`agritec_uk` lane confirmed the motivating pair WAS a genuine duplicate (they de-duplicated,
105 → 104 facts, having recorded it on 08-22 as *"two near-duplicates — harmless"*; it was not
harmless, it was **latent** — until something had to choose between them, "two true facts" and
"one fact twice" were indistinguishable, and this suggester was the first thing that had to choose).

**But they then measured their register and found NINE MORE pairs sharing a value, none of them
duplicates** — (UPL10, CNUM2)=102, (WBD6, WBD7)=115, (GRH8, OFC1)=187, (OFC5, OFM6)=1920,
(CAHL4, CIGL2)=515, (HEF6, CIPM3)=55, (BFS1, OFM5)=707, (AHW5, WBD3)=765, (CIT-f88b5cd, OFM1)=20.
Every one is **two different SFI26 actions that happen to be paid the same rate.** Collapsing them
would destroy real facts.

**So on a rate-table site, shared values are NORMAL rather than pathological, and the suggester
will be quiet across most of the register for a reason that has nothing to do with anything being
wrong.** `[MEASURED by the agritec lane, 2026-08-25]` **Do not read a low suggestion count on such
a site as "few bindings to make".** The remedy for that population is not the value probe at all —
it is a per-fact contextual `artifact_check`, where the ACTION CODE rather than the rate does the
discriminating (`code:'CSAM3'[^}]*rate:224`). The suggester finds bindings; it was never going to
find all of them, and on a rate table it finds proportionally fewer.

### ⚠ AND A COUPLING NOBODY HAD NAMED: PHASE 3a IS STARVED UNTIL PHASE 4 IS ADOPTED

The sweep produced **0 `fact_drift` entries, and therefore 0 probe annotations**. That is not a
fault — it is the self-quieting property working, the 13 baselines are recorded and the register
has not moved. But the consequence had not been stated anywhere:

**The probe annotates EMISSIONS. A steady-state fleet emits nothing, so it measures nothing.**
Phase 3b's precondition — a measured present/absent/markup-only distribution — therefore cannot
accumulate on a quiet fleet. It accumulates only when a NEW declaration files its one-time
`unreconciled_declaration` batch per (fact, tool). **So Phase 4 (adoption) is the precondition for
Phase 3a's data, which is the precondition for Phase 3b.** They are in series, and 3b is further
away than "run it for a month" implied. If LMC adopt their 7 bindings, that is 7 annotated
emissions on the next pass and the distribution starts.

## §5d — ADOPTION MOVED, and Phase 3a's first real distribution is 100% `not_probed`

2026-08-25, after the `agritec_uk` lane landed a 24-fact declaration on
`tool-sfi26-revenue-stacker`. Verified through the parser's own path, not just that the JSON
parses: the `LIKE '%"facts"%'` prefilter selects it, `facts` is a top-level sibling of
`profiles`/`container`/`checks`, 24 entries, all strings, **0 duplicates**, 5 checks preserved, **no
per-check `facts` key** (which P7 would hold inert), and **all 24 resolve against their register**,
so zero `fact_declaration_broken`.

| | |
|---|---|
| declarations filed | **24**, `unreconciled_declaration` → `fact_drift_review` |
| **probe verdict on all 24** | **`not_probed`** |
| citation arm | **104 fresh, 0 error, 0 drifted** (was 22 of 105 two days earlier) |
| `artifact_check` | 4 fresh |

**THE DISTRIBUTION IS THE POINT, AND THIS IS A ZERO-INFORMATION SAMPLE.** Every SFI rate is below
the measured floor of 1000, so all 24 refuse. Phase 3a is no longer *starved* — it annotated every
emission, which is the mechanism working — but a site whose register is a rate table produces
**24 rows of `not_probed` and no discriminating data at all.**

**So Phase 3b's precondition is narrower than "adoption": it needs adoption BY A SITE WHOSE VALUES
CLEAR THE FLOOR.** Of the estate, that is mortgagecalculator, loanandmortgagecalculator and
gamesdesign. **LMC's seven suggested bindings are still untouched and are now the single highest-
value action on this lane** — seven five- and six-digit SDLT figures on the second stamp-duty
calculator, which is both the biggest unguarded surface AND the distribution's best source.

**Adoption two days on:** `artifact_check` **0 → 6 facts**; tool PLANs declaring **1 → 2**;
`fact_binding_suggested` notes **5**, of which **1 acted on**.

## §6 — what is owed now, in order (2026-08-24)

1. **The roll.** All of Phases 1 and 2 is Go and therefore inert. Verify at the binary,
   never at the tag, with a positive control in the same exec — recipe in the lane
   RUNBOOK. Strings unique to this work: `fact_declaration_broken`,
   `artifactCheckSubjectSurfaceQuery` is a var not a literal so grep
   `subject_key %q resolves to no stored component HTML` instead.
2. ~~**Retype ONE real fact.**~~ **DONE 2026-08-24 at the owner's direction, and induced both
   ways — see §5b.** ⚠ The register edit was made on **another lane's site**
   (`mortgagecalculator_couk_adoption`) on the owner's instruction, not on that lane's
   agreement; they have been told. It touches the `evidence_base` row only — no fence, no
   page, no `doc_plans` row — so their `install_fences.py` is unaffected.
3. **Phase 3a — the probe.** ~~Measure the distribution over a full fleet sweep~~ → **BUILT AND LIVE, but see §5c: it is STARVED until adoption grows.** Prove a declared figure is actually in the
   tool's **script text**, and measure the present/absent/markup-only distribution over a
   full fleet sweep before anything acts on it. **Script text, never whole-page** — see §0
   and `WRONG_CALLS` 2026-08-24: the register's own `writer_line` puts the figure in the
   page's PROSE, so a whole-page check would have passed bug 225 daily for sixteen months.
4. **Phase 3b — arm it**, on 3a's measured numbers, as its own council round.
5. **Phase 4 — adoption.** `[MEASURED 2026-08-24]` **15 of the 27** distinctive (≥5-digit)
   register facts already appear verbatim in a tool on their own site, across 7 sites — so
   declarations are machine-*suggestible* without hand-authoring 178 fences and without the
   blanket JS scan `PLAN_2026-08-09 §3` refused (that scan hunts UNREGISTERED numbers, an
   unbounded set; this probes for REGISTERED values, a finite per-site set already vouched
   for — the direction is inverted, and the submission must say so rather than hope nobody
   notices the resemblance).
6. **The probe's floor and the suggester's population are BOTH narrower than the class, and the
   `agritec_uk` lane's live case falls outside both.** Their SFI Revenue Stacker computes a
   subsidy DEFRA abolished, with two of nine revenue lines correct — 288 with money attached.
   `[MEASURED 2026-08-24 against their live rows]`: **75 of their 105 registered facts are below
   the probe's floor of 1000** (SFI rates are 2–3 digit: 382, 224, 20, 13, 10), and their
   calculator sits on a `page_type='blog-post'` page on a site with **zero** tool-level
   components, so the Phase 4 predicate does not select it — fleet-wide that class is **50 of
   222** script-bearing pages on register-bearing sites. Their own mitigation (require every
   applied constant to be a registered fact AND appear in the tool's visible rate table) is
   therefore **not redundant with this work** and should stay: it covers the displayed-rate
   shape, Phase 3a covers the ≥1000 code-constant shape, and **the residual — an internal
   constant, below 1000, that nothing displays — is reached by neither.** That residual is
   Piece 4's, and the honest per-fact answer available today is a human-authored
   `artifact_check` with a context-bearing pattern, which Phase 2 made reachable for citation
   facts. ⚠ Their tool is not yet in stored `rendered_html` (11 pages, 1 with any `<script>`,
   0 containing `Math.min` or `382`), so neither lane may quote a probe verdict on it yet.
   Told to them 2026-08-24 rather than left to be inferred.
7. **Standing, and not this lane's to close: `bugs_open/033`.** The council's `architecture`
   seat objected (medium, advisory) that Phase 1's doc_note surface is the **second**
   bespoke durable surface invented to route around the dead `needs_human_review` queue,
   after `noteNeedsCriteria` — and that each such bypass makes the eventual triage fix
   harder to land, because by then there will be N ad-hoc doc_notes conventions instead of
   one queue. The seat agreed doc_notes is the right call given measured reality
   (**345+** stuck items) and asked for the objection to be **recorded rather than silently
   absorbed by another well-built bypass**. It is recorded here.

## The council round, and the defect it caught (2026-08-16)

Round 1 came back **REVISE**, gated by the `editquality` seat, and the gating objection
was **right in a way that would have shipped a mechanism blind to its own motivating
bug**. Recorded here because it is the most useful thing in this file:

> the baseline is defined purely from PRIOR REGISTER state … a tool that is wrong from
> the day it opts in, against a register fact whose value has been stable since
> (exactly bug 225's shape: correct register, stale code, no subsequent legislative
> change), produces baseline == current and … emits nothing.

**That is 225 exactly.** The fix: a declaration with no baseline now files one
`unreconciled_declaration` review item per (fact, tool) asking a human to confirm the
tool encodes the registered value. It is self-quieting — the item carries the value,
which becomes the next pass's baseline — and it is banded low, because 13 declared
facts means 13 one-time items and that is a backlog, not 13 alerts.

Two further findings acted on, and three answered by measurement rather than by change:

| seat | finding | outcome |
|---|---|---|
| `compliance` (med) | every human-routed case sat at medium/35, ranking a confirmed move on a `no_auto_fix` tax calculator BELOW an auto-fixable one | **fixed** — three bands; a confirmed move is high/30 whichever route it takes |
| `debug_historian` (high) | `p.status='active'` is the liveness predicate and is wrong for an AUDIT — and the equivalent check had been run for the eligibility predicate but not this join | **fixed** — now uses the shared `NeverDeployedPagePredicateFor`. Measured: both motivating pages are `active`/`deployed`, so it would not have shipped inert; 198 active vs 6 archived tool pages fleet-wide |
| `bug_historian` (high) | is the fact id in the `item_key`, or will a second drifted fact on the same tool be swallowed (bugs_open/091)? | **already correct** — `fact_drift:<fact_id>:<subject_key>:<site_id>`; the sketch had hidden it |
| `bug_historian` (med) | a persistently-403 source means permanent silence | **answered** — `refreshCitationFact` marks a fact drifted once it passes `staleness_days` whether or not the fetch succeeded; that arrives here as evidence_drift. Time-based, not attempt-based |
| `guardian` (med) | how many callers share the modified action? | **measured: one** — `evidence-freshness` |
| `architecture` (med) | does P11 cross the WFA-013 optional-key budget (N=10)? | **measured: no** — `audit-optional-key-budget.sh` reports 0 shared actions over budget |

### Round 2 — REVISE again, and it killed round 1's fix

> `unreconciled_declaration` only fires when baseline is nil, but nil requires BOTH no
> lastItem AND no previousRow … Since the register is re-verified daily … previousRow
> will almost always resolve.

**Correct.** The fix for round 1 was inert on the very site it was built for. Repaired
by deleting the register-history fallback entirely: the baseline is the newest
`fact_drift` finding for (fact, tool) and nothing else, and its absence is the
never-reconciled signal. Two questions had been conflated — *has the VALUE moved*
(register rows) and *has this TOOL been reconciled* (a prior finding) — and only the
second is this mechanism's.

⚠ **Proving it took three mutations; the first two passed and were worthless.** One
mutated a path the struct-literal test bypasses; one read a row without using it. Only
the faithful mutation (declared, populated, consulted) failed. **An induced red is
evidence only when the mutation is the defect** — `WRONG_CALLS.md`, 2026-08-16.

Also this round: reused the exported `discovery_checks.ToolSubjectKeyExpr` instead of a
hand-written equivalent (two seats), verified live to resolve both tools.

### Round 3 — APPROVED, and both advisories acted on

- **`handler_agent = 'human-review'` names an agent that does not exist.** Measured:
  **0** `agent_definitions` rows of that type; the fleet's dominant convention is a NULL
  handler — **544 rows across 21 item types**, including `ported_tool_fix`, the very
  precedent this routing was modelled on — against 22 rows using the string. Copied in
  good faith from the sibling function in the same file. `fact_drift_review` is now
  handler-less.
- **"Zero rows ends the pass with no signal."** A site whose fences declare facts but
  whose pages stop resolving now warns; a site that declares nothing stays silent.

## How to verify (after the roll)

```bash
# 1. the code is actually running (never infer from the tag)
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac fact_drift_review /proc/1/exe   # expect >0
kubectl -n ai-persona-system exec <chassis-pod> -- grep -ac stale_attestation /proc/1/exe   # positive control
```
Then seed the mcalc `stamp-duty` fence and **induce**: supersede
`sdlt-ftb-relief-cap` 500000 → 550000, fire a `refresh_evidence_base` dry run, and
assert it names `stamp-duty` with `kind:value_drift, route:fact_drift_review,
reason:no_auto_fix`; restore. **A dry run that reports nothing after a real change is
the failure**, and it is the only result that distinguishes this from an inert check.

## Related

- `bugs_closed/225` — the case. `bugs_open/224` — the sibling defect family found the
  same way (a 0% rate breaking six of seven calculators).
- `PLAN_2026-08-09_facts_into_tool_acceptance.md` — F1/F2/F3, and Piece 4.
- `bugs_open/161` — the register ratifying a false claim: the same register, the
  opposite failure (there the register was wrong and vouched for the copy; here the
  register was right and could not reach the code).
- Register: **CLM-022**, **TL-045**, CLM-021 (the builder half), SQAM-003 (the oracle
  that actually found 225, still outside the platform).
