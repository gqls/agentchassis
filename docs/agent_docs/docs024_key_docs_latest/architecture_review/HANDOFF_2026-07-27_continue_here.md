# HANDOFF — architecture seat (2026-07-27, late) — ⚠️ SUPERSEDED

> **SUPERSEDED 2026-07-27 evening by `HANDOFF_2026-07-27b_continue_here.md` — start
> there.** This file is wrong at the top and right at the bottom: the thing it calls
> "THE ONE THING OWED" was run, and both decisions it leaves open (D7(b), D9) have
> since been **ruled by the owner**. **§5's landmine list is still current and worth
> reading** — the Go-contract ones especially (only `{approve,object,veto}` parse;
> only `{reviewer,verdict,objections,missing,notes,degraded}` persist). Kept rather
> than rewritten, because the corrections inside it record how the understanding
> moved.

**Cold-start entry point for this workstream.** Read this, then
`SUMMARY_2026-07-27b_architecture_seat_built.md` for the prose state and
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` for the per-decision
table. `RUNBOOK_architecture_seat.md` has every command with its gotcha.

---

## 1. What this workstream is

The owner asked for a process — possibly a council seat — that keeps the
architecture robust, stops it shifting underneath us, and keeps it sufficient for
anticipated plans, knowing those goals conflict. The conservative half already
existed (the guardian seat, sole hard-veto holder, charter clause (d)). Nothing
argued the forward half. This workstream measured whether that imbalance mattered,
found it did, and built the counterweight.

## 2. State — what is LIVE right now

All config; **no image dependency, nothing waits on a roll.** Verified by content
against the live DB at 2026-07-27 ~14:00, after the v1.0.1173 deploy.

| thing | where | state |
|---|---|---|
| Council reads its own minutes (D8a′) | `council-gate`, `fix-proposer` — guardian, both historians, prior_art, reuse_agent | **LIVE** |
| Guardian deflection check (from D5) | same two agents | **LIVE** |
| Generated case index (D8e-1) | both historians on both agents | **LIVE** |
| `review_architecture` forward seat (D1/D2/D3) | `feature-designer` only | **LIVE, but pre-fix prompt** |
| Mechanical RFC trigger (D4) | `scripts/commit-scope-report.sh` | **LIVE** in the commit hook |

Scripts (all committed, all re-runnable, none write to live config):
`scripts/build-historian-index.py`, `scripts/add-architecture-seat.py`,
`scripts/patch-feature-designer-context.py`, `scripts/council-adoption-report.sh`.

## 3. THE ONE THING OWED — ✅ DONE 2026-07-27 late

```
/tmp/acm/APPLY_gap.sh
```

> **RUN AND VERIFIED 2026-07-27 (late session).** Nothing is owed here any more;
> §6 is the live list. Post-apply invariants all held: `live == CONTEXT.json`
> byte-exact after canonical load, step set unchanged, `review_fields` 6,
> `hard_veto_from` `["guardian"]`, `max_rounds` 3. Every guardian on all three
> councils now shows `minutes=t deflection=t`; both historians `case_index=t`.
> Pre-flight worth repeating on any future config push: the structural diff of
> live → payload was **exactly three `prompt_template` strings** and nothing
> else, and `live == SEATED.json` was confirmed `True` first — which is what made
> `ROLLBACK=1` trustworthy rather than merely available. Detail in
> `NOTES_architecture_seat.md`.

Verified against live: prompt text only, step set intact,
6 `review_fields`, `hard_veto_from` unchanged. Rollback `ROLLBACK=1`. It does three
things:

1. gives `feature-designer`'s **guardian** its minutes + the deflection check;
2. gives `feature-designer`'s **bug_historian** the case index;
3. **replaces `review_architecture`'s prompt** so its routing signal lands in the
   first line of `notes` — see §5, this one is load-bearing.

Why it matters: the 099 mirror spans `fix-proposer → council-gate` **only**.
`feature-designer` is mirrored by nothing, so it was left the worst-equipped of the
three — the design-stage council, which D2 argues matters most.

**If `/tmp/acm/` has been cleared**, regenerate:
`python3 scripts/patch-feature-designer-context.py --apply` (writes
`/tmp/acm/feature-designer-CONTEXT.json`), then push it the same way
`APPLY_gap.sh` does — base64 into `convert_from(decode(...))::jsonb`, never a
quoted heredoc.

## 4. The measurement — and why there is nothing to read yet

`scripts/council-adoption-report.sh`. **Baseline, all pre-change:**

| seat | reviews | invoked stability pref. | cited precedent |
|---|---|---|---|
| guardian | 210 | 90 | **6 → see correction: the real intersection is 2** |
| bug_historian | 143 | — | 40 cited a source |
| debug_historian | 178 | — | 23 cited a source |
| architecture | 0 | — | — |

~~**6 of 90 is the number to beat.**~~ That is how often the seat that most needs its
own history referred to it while invoking the preference that needs it most.

> **CORRECTED 2026-07-27 late — the number to beat is `2 of 90` (2.2%), and the
> sentence above describes a subset the query never computed.** `invoked_stability`
> and `cited_precedent` were two **independent** `FILTER`s over all 210 guardian
> reviews, so "6 of 90" was never "6 *of the* 90" — **4 of those 6 cited precedent
> WITHOUT invoking the preference at all.** The intersection, which is what the
> sentence claims, is 2. Fixed in `scripts/council-adoption-report.sh`: the
> headline is now `both_invoked_and_cited` with `pct_of_invoked`, and
> `cited_but_did_not_invoke` is kept visible so the gap cannot hide again.
> Caught by asking why this table said 6/90 while the DECISIONS doc said 3/87 —
> two figures both labelled "pre-change" for a population that was supposedly
> fixed. **Corrected reading: before 2 of 90 (2.2%), after 1 of 2 (50%)** — n=2,
> so no conclusion, but the baseline is now honest and it is *lower* than
> reported, which strengthens rather than weakens the D5 case.

**Post-cutover reviews: ZERO.** Five councils ran today (13:00–13:29) but the true
cutover is **13:44:56**, so every one of them is baseline. Do not read them as
signal. Re-run the report once a few councils have run past that time.

> **UPDATED 2026-07-27 late — first post-cutover evidence is IN, and it is good.**
> 10 post-cutover reviews now exist, all `council-gate`. The council at
> **14:18:19** (`b64141e5`, the 109 render-context fix) is the first past the
> cutover and all three payloads landed:
>
> - **debug_historian cited `WRONG_CALLS.md` BY DATE** to reject the plan's
>   *"no sixth path exists"*: *"exactly the shape of WRONG_CALLS 2026-07-21 …
>   and 2026-07-24"*. Our own logged errors catching a new one — the thesis of
>   the workstream, working.
> - **bug_historian cited `016b §9`, `bugs_open/034`, `bugs_open/109`.**
> - **The guardian invoked the stability preference and reasoned its way OUT of
>   deflecting:** *"The recurrence across three rounds is evidence of a genuinely
>   scattered defect …, not evidence that this fix belongs at a higher layer."*
>
> **But read §2 of the report sceptically — it scored that guardian review
> `cited_precedent=0`.** Qualitatively it is the D5 payload working; the metric
> counts a *precedent citation* and reasoning correctly about recurrence without
> quoting a past report does not match it. So **`cited_precedent` undercounts
> correct behaviour**, and "6 of 90 → n" is not on its own the verdict on D5.
> n=1: no conclusion yet, in either direction.

Section 5 of the report is a deliberate **kill switch**: a high object-rate with no
signal line means `review_architecture` is producing confident noise, and it should
be pulled rather than tolerated.

**Blind spot, stated in the script:** `checks`/`code_checks` are NOT persisted
(0 of 2,138 stored reviews carry either key), so the report cannot see whether a
seat issued a SQL query — only what it chose to write down. Silence is not proof it
did not look.

## 5. Landmines — the expensive ones, all paid for this session

- **A council seat's output schema is a Go contract. Read it before authoring a
  prompt.** `platform/orchestration/actions/diagnose_council_decide_action.go`:
  `:160` recognises **only** `{approve, object, veto}`; `:397` marks anything else
  UNREADABLE; `:446` downgrades an otherwise-approved round to REVISE. An invented
  verdict vocabulary would have forced **every** `feature-designer` round to revise,
  exhaust `max_rounds: 3`, and fail — silently, in the lane the seat was added to help.
- **Custom JSON fields are DISCARDED on persistence.** `councilReview` (`:84-95`)
  marshals only `{reviewer, verdict, objections, missing, notes, degraded}`. A seat's
  signal must live in the first line of `notes` or it is written to nothing. This is
  what §3 item 3 fixes and why that patch is not cosmetic.
- **`hard_veto_from` is only an audit label** (`:13-14`) — ANY reviewer's veto
  rejects the round. A seat is advisory because its prompt never offers `veto`.
- **`feature-designer` is mirrored by nothing.** 099 covers `fix-proposer →
  council-gate` and does **not** copy `load_schema_hint` (099 line 117). A hint
  change is a four-place edit; a prompt change rides the mirror to two of four.
- **Never guess a cutover.** Take it from `agent_definitions.updated_at`. My first
  draft hardcoded 13:00 from memory, 45 minutes early, which would have reclassified
  five pre-change runs as evidence.
- **Beware a case-insensitive check that matches what you are replacing.**
  `ILIKE '%ARCHITECTURE_SIGNAL%'` returns TRUE for the pre-fix prompt's lowercase
  `architecture_signal` field, so it read "already patched" when nothing had changed.
- **A work item's `status` is not an ownership signal, and a docs grep is not a
  coverage check.** (2026-07-27 late — caught before acting, would have cost a
  council round and left confusing artefacts in another thread's trail.) Wanting
  to exercise the new seat, I nearly fired a probe run at `7b89fb35`, which looked
  free on two independent signals that were **both wrong**: `status='deferred'`
  (it is where this lane parks an item *between council rounds*, not abandonment)
  and a `grep -rln` for its subject across `docs/`, `bugs_open/`, `features_open/`
  that returned **no owning workstream doc at all**. Opening the spec body showed
  three completed rounds and `=== ROUND 4 — ONE CHANGE ONLY, owner-directed`.
  **For a `site_work_items` row the ownership evidence often lives INSIDE the
  `spec` jsonb, where no repo grep can reach it** — and `scripts/who-owns.py`
  covers a bug number/slug, not a work item, so this class is not covered by any
  tool. **Read the row you are about to act on.** Both approved specs are owned;
  neither is a probe target.
- **The historian's "frozen seven" were CURATED, not lazy** — selected by
  rediscovery frequency from the concept register, deliberately narrow, with
  broadening named as future work in
  `docs026_concept_register/PILOT_bug_historian_reviewer.md` §1/§3. That same doc is
  where the veto semantics were already written down. **Read docs026 before touching
  a seat.**
- **Config survives a chassis roll, but verify by CONTENT not timestamp.** After
  v1.0.1173 all three agents showed an identical `updated_at`, which looks exactly
  like a re-seed clobber. It was not — but only a content check proves that.

## 6. What is open, in order

1. ~~**Run `/tmp/acm/APPLY_gap.sh`**~~ — ✅ **DONE 2026-07-27 late** (§3).
2. **Let councils run, then re-run the adoption report.** This is the honest test of
   the whole workstream and it cannot be hurried.

   > **CORRECTED 2026-07-27 late — "cannot be hurried" understated it; on the
   > current queue the architecture seat cannot speak AT ALL, so waiting alone
   > produces nothing.** The chain: `review_architecture` exists **only** on
   > `feature-designer`; `feature-designer` refuses anything without an
   > owner-approved spec (`check_spec_approved` — `item_type='capability_gap'`
   > carrying BOTH `owner_approval` and `code_pointers`); and there are **5
   > `capability_gap` items in total, 2 approved, both already owned by other
   > threads** (`9ed684bc` tools-api → Gauntlet; `7b89fb35` colour-fixer remit →
   > an active 4-round design thread). So the first architecture review arrives
   > when the colour thread runs its **round 4** (owner-directed, instructions
   > already written into the spec) or when the owner approves a new spec.
   > **A zero in §4 of the report is a rate limit, not a fault** — do not read it
   > as breakage, and do not fire a probe to manufacture one (see §5).
   >
   > The seat itself is verified **reachable**, so nothing is wrong with it: BFS
   > over the workflow graph including conditional branches gives 24 of 24 steps
   > reachable from `load_spec`, no orphans, chain `review_guidelines →
   > review_architecture → review_guardian → council_decide`. It has **no
   > footprint/relevance gate** (unlike the 16-seat gate's seats) so it fires on
   > every design run clearing the approval gate.
3. **D7(b) — should the guardian weigh benefit at all?** Owner undecided, and
   deliberately so. My view: it should not — it has no instrument for benefit and has
   been overturned every time it was escalated. The honest counter: risk and benefit
   are not separable, and a blast-radius-only seat would block every wide change,
   which is *more* conservative. **Now answerable from evidence** once the new seat
   has said something.

   > **New evidence 2026-07-27 late, bearing on this AND on D1/D2/D3.** On the
   > 14:18 council, `bug_historian` opened its note *"Architecture-level concern
   > for a human"* and asked for a human ruling between a shared
   > render-context-builder refactor and continuing to fix drop points one live
   > test at a time. That is a forward-fitness judgement, raised by a seat not
   > commissioned to make one, on the **fix lane** — which has no
   > `review_architecture` seat because D1/D2/D3 placed it on `feature-designer`
   > only. First live sign that the fix lane also generates these questions and
   > currently routes them to "a human" because nothing there owns them. Not
   > enough to reopen the decision; put it on the table when D7(b) is answered.
   > `[UNMEASURED]` — one instance, noticed by reading, not counted.
4. ~~**D6 — name the asymmetric bar**~~ — ✅ **DONE 2026-07-27 late.** Now a named
   section in `PROCESS_architecture_review.md`: keeping battle-tested code is the
   default and needs no evidence; replacing it must clear four bars — a defect the
   current design *cannot express a fix for* (recurrence alone is explicitly NOT
   sufficient), mechanically-derived blast radius, independently-valuable stages
   that build alone, and a rollback needing no migration. The *why* of the asymmetry
   is stated too (risk is measurable pre-hoc, benefit is a forecast; weighing them
   as like quantities favours whoever already preferred one), as is the
   counter-argument (four bars can be met formally by a wrong plan, and they say
   nothing about a change that is too *small* — that gap is the seat's
   `insufficient` signal, not the document's).
5. **The real remaining design: how does a reviewer query markdown at all?**

   > **UPDATED 2026-07-27 late — most of this is already answered by
   > `bugs_open/108`, and the two should be solved TOGETHER, not separately.**
   > 108's fix candidate 2 ("index bodies — populate a `body` column from the
   > `[line_start, line_end]` span already stored on every row") explicitly notes
   > it *"also answers the schema half of the architecture thread's D8b (indexing
   > markdown so `WRONG_CALLS.md` and `bugs_open/` become reachable) — same
   > question, already settled."* So the mechanism is decided; what is left for
   > this workstream is the **ranking** (reuse the concept register's
   > rediscovery-frequency signal, do not invent one), not the plumbing.
   > **Grounded live 2026-07-27: `code_symbols` = 4,535 rows, 100% Go, `0`
   > markdown, 530 files.** And `content` holds **declarations only** —
   > `max(length(content))` = 451, the longest `func` row is its signature line —
   > which confirms 108 rather than contradicting it. Beware `count(*) FILTER
   > (WHERE content <> '')` = 4,535: it looks like every body is indexed and the
   > column name promises more than it holds.
   >
   > **Consequence for the seat we just shipped, contributed into `bugs_open/108`
   > (unowned; do not fork a second account):** `review_architecture`'s prompt
   > tells it `"content" searches source bodies`, which given 108 is **false**, and
   > the prompt carries no "an empty result is not an absence" warning on the code
   > tier though it does on the SQL tier. **Do not fix that clause in the prompt
   > alone** — three other consumers (`review_prior_art`, `review_reuse_agent`,
   > the diagnosis loop's `lookup_code_symbols`) would still be lying, and a
   > prompt edit is not worth making twice.

   The
   concept register (`docs/agent_docs/docs026_concept_register/`) is most of the
   roadmap artefact I had claimed we lacked — concepts by subject, *including
   abandoned ones*, plus `DIRECTION_LEDGER.md` naming what is fixed (constitution +
   mission, hash-checked, commit-hook enforced). It is markdown, so it is invisible
   for the same reason everything else is. Solving that serves the architecture seat,
   both historians, and the reuse and prior-art seats at once. Reuse the register's
   own rediscovery-frequency signal rather than inventing a new ranking.

## 7. Evidence behind the case, if challenged

- **Ossification is measured, not asserted.** `coordinator.go`/`ProcessResponse` was
  deflected upward across **six distinct submissions in seven days** (07-20→07-26);
  `spawn_actions.go` four; kafka lane two. Four bugs open in that core (075, 086,
  034, 096). Method in the RUNBOOK; counts are **floors** (ILIKE tagging undercounts).
- **Churn refutes the premise it was meant to test.** `platform/orchestration` shows
  366 commits/60d but **348 are the plug-in action registry**; the core moved 55 and
  `coordinator.go` 9, against 2,123 repo-wide. Eleven `platform/*` packages moved zero.
  Quote the split or the headline misleads.
- **The guardian's veto has never been sustained when escalated** (003, 030, 086).
  Not proof it was wrong — a risk that did not land was still a risk — but evidence it
  is uncalibrated.
- **D4 backtest:** the RFC trigger fires on 10 of the last 300 commits (3.3%).

## 8. Session posture worth carrying

Three confident claims of mine were corrected in two days — that no pre-build
council existed, that the council could not read its own verdicts, and the verdict
vocabulary that would have broken the lane. **Every one was caught by reading
something or running a measurement, none by thinking harder**, and twice the answer
was already written down in our own docs. Full entries in
`docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` (2026-07-26 and 2026-07-27).

That is also the argument for the workstream in miniature: a seat with our written
history in front of it would have known the veto behaviour, because we had already
written it down.
