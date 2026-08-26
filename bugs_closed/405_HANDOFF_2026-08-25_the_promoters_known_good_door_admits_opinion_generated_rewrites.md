# 405 — `detected-item-promoter`'s known-good door tests the HANDLER's competence, never the FINDING's provenance — so LLM-audit opinions ride a door built for mechanical defects

**Filed 2026-08-25** by the `loanzy_uk_example_site` lane, at the owner's direction ("find the
thread that produced that promoter or file a new bug for it"), while executing
`PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md`.
**Status: OPEN. Severity: medium** (the acute inflow is stopped by migration 623 + the record-mode
seam; this files the STRUCTURAL residue so it is not folklore).
**Producing thread, named:** the promoter was built by the **`bugs_closed/083`
(detected_findings_never_reach_a_handler) lane** — commit `3c6354059` 2026-08-15, "083 candidate 2
BUILT (owner ruling 2026-08-15)", lineage migrations `430 → 444 → 453/454 → 458 → 465 → 480` —
and 083 was closed 2026-08-22 by the **`bugs_open/277` lane** (live session notified today).
`scripts/who-owns.py 083`: closed; no open owner. Hence a NEW file rather than a CONTRIB.

## 1. The mechanism (read from the live `pre_query`, 2026-08-25 — not inferred)

`detected-item-promoter` (scheduled task, 900s, live since 2026-08-15) promotes any
`site_work_items` row at `status='detected'` with a non-empty handler through four doors:
pipeline ∈ (build, content, design); handler is a live agent; **the (item_type, handler_agent)
pair has ≥1 lifetime completion** (live ∪ archive); the pair is above a 25% success floor.

Every door interrogates the HANDLER and the PAIR's history. **No door asks what produced the
row.** A `content_rewrite → page-build-handler` row passes the known-good test identically whether
it was filed by a mechanical check that measured a defect ("this required field is empty") or by an
LLM auditor's opinion ("this page could be better") — because the pair's thousands of completions
were earned by BOTH populations under one item_type.

## 2. Why that is a defect and not a design choice, with the evidence

- **The promoter's own description says** "Does NOT re-enable improvement-sweep" — its authors
  understood it as draining a triage backlog, not as a dispatch route for auditors. But
  **[MEASURED 2026-08-25, live ∪ archive]** ~~26~~ **27** LLM-audit rows
  (> **CORRECTED 2026-08-25 late evening by the `bugs_open/391` lane's re-verification (their
  > CONTRIB below, commit `db96ea770`): my own four listed batches sum 5+6+12+4 = 27, and I wrote
  > 26 beside them. A dated count is what gets quoted onward — caught before it travelled.**) (`spec->>'audit_source'` ∈ the six
  model-seat sources) were promoted between 2026-08-20 and 2026-08-24 **while the sweep was
  disabled** (triaged_at 08-20 14:59, 08-22 11:26, 08-24 10:27, 08-24 22:21; the 08-17 12:40 and
  08-22 18:4x batches excluded as sweep/hand-run triage). Rewrites the owner believed were off
  were arriving by this door.
- **IMP-054's premise is voided** (register): "a lone discovery run files findings nothing can
  ever see" was true when written (2026-08-09) and false six days later. `detected` is a queue,
  not a shelf — LANDMINES.md "`detected` is a QUEUE, not a shelf" (2026-08-25) is the prospective
  half of this file.
- The known-good doors were REVIEWED and tightened three times (444, 454, 465, 480) and every
  tightening was about handler competence — the provenance axis was never on the table, because
  083's population WAS mechanical. The blind spot is inherited scope, not an error by its authors.

## 3. Why this is filed on first-hand verification rather than a `090` run (owner ruling 2026-07-31)

The claim asserts no hidden cause: every clause of §1 is the live `pre_query` read verbatim, and
§2's counts are dated queries over `site_work_items ∪ archive` reproduced in
`loanzy_uk_example_site/NOTES` (2026-08-25 evening). There is no competing hypothesis to refute —
the door demonstrably does not read provenance, and rows demonstrably passed it. A `090` run would
re-read the same two artefacts. (The related JUDGEMENT question — whether opinions should dispatch
at all — went through the council instead: RFC_056, trail `d1342f2a`.)

## 4. What already contains it (so the residue is stated exactly)

- **Migration 623 (APPLIED 2026-08-25)**: the four model seats are off the improvement loop's
  path, so no NEW opinion rows are being filed. **0 LLM-audit rows sit at `detected` today.**
- **The record-mode seam (`filing_mode: record`, in the running binary since v1.0.1339)**: once
  RFC_056's round-2 verdict lands and migration 624 applies, model-seat findings are born
  `deferred` + `handler ''` — never `detected`, so never this promoter's candidates.
- **Residue A:** any OTHER route that files an opinion-shaped row at `detected` with a proven
  pair — a hand-fired auditor without record mode, a future seat whose author does not know this
  file — re-opens the door. The promoter itself still cannot tell.
- **Residue B:** `tool-acceptance-tier4` and similar non-improvement-loop audit sources also file
  `detected` rows the promoter handles; they are DEFECT-shaped today, but the same one-door
  design means nobody would notice if one became opinion-shaped.

**§4a — the record-mode census, settled twice over (2026-08-25 night, both lanes).** The 391 lane
verified first-hand: SIX live agents carry `write_audit_findings` with `filing_mode: record`
(brief-fidelity, content-quality, offer-analyser, reader-experience, site-review, visual-design).
Their follow-up worry — "council-gate and fix-proposer carry the step and are NOT in record mode" —
is the NAME-JOIN trap of §4b below, one seam over: `[MEASURED 2026-08-25, both instruments in one
query]` those two rows have **0 steps with `action='write_audit_findings'`** and **1 text
occurrence each** — the string lives in their reviewer-roster PROSE (footprint maps), not in a
step. There is nothing to switch and nothing deliberate to confirm: they are not seats. Settled on
a THIRD instrument (391 lane, `244f2f7ca`): `jsonb_path_query_array(default_config,
'$.** ? (@.action == "write_audit_findings")')` — recursive descent, so it also covers nested
sub_workflows a top-level steps walk misses — six auditors at 1 real step each, the two gate rows
at 0.
**The sharper lesson, theirs (kept in their words because it beats "two readers disagreed"):**
the two mis-readings were NOT symmetric. A steps-walk that under-reports fails CLOSED for a
containment claim; a text census that over-reports, attached to a POSITIVE claim, is the one
combination the LANDMINES "'can this agent write X' is a GO question" entry rules out in as many
words — *a landmine that says "safe in direction X" is not a landmine that says "safe."* And the
delivery vehicle matters: *"worth confirming X is deliberate" is an assertion of X wearing a
question mark* — **a question inherits the evidentiary standard of the claim inside it**, and the
sloppiest line in a careful message is the one nobody, including its author, is checking.

**§4b — ⚠ the vocabulary trap a verifier WILL hit (391 lane, hit first, nearly sent as a
refutation):** `audit_source` labels are NOT agent type names — `offer-analysis` ↔
`offer-analyser`, `site-review` ↔ `site-review-agent`, `content-quality-audit` ↔
`content-quality-auditor`, `visual-design-audit` ↔ `visual-design-auditor`,
`brief-fidelity-audit` ↔ `brief-fidelity-auditor`, `reader-experience-audit` ↔
`reader-experience-auditor` (no rows yet — the safe direction), and `design-audit` is a
HISTORICAL predecessor label (6,035 rows, newest 08-12, none in the 08-20→24 window) with no
record-mode agent — renamed, not a gap. A `WHERE type IN (<audit_source values>)` join returns
**0 rows and reads exactly like "record mode was never applied"**. Candidate 1 crosses this seam
BY CONSTRUCTION (the Go stamps from the SEAT's context; the door reads the ROW), so the fix must
pin the mapping explicitly — a lockstep test over the pairs above, not a comment.

## 5. Fix candidates, ordered by what closes the door (not by effort)

1. **An origin class the promoter can read** (closes the door) — **BUILT, council APPROVED round 1
   (corr `946d587c`, 4 advisory none high), and APPLIED 2026-08-26 09:4xZ** — live pre_query
   carries `origin_ok` ×4 and the description names door 5. The door holds nothing until the
   stamping binary rolls (the Go stamp postdates v1.0.1339); §6's two-direction verification is
   OWED after that roll. `write_audit_findings` stamps `spec.origin =
   'model_opinion'` into every finding's base spec unconditionally (every live caller is a model
   seat; a future mechanical adopter would widen the HOLD — the safe direction), and migration
   `629_promoter_origin_door_holds_model_opinions.sql` adds the fifth door by four
   verbatim-anchored replaces (origin_ok in scored / candidates / held complement / held CASE,
   reason "model opinion - release by hand or via record mode"). Lockstep pinned by
   `TestOriginDoorLockstep` (reads the migration file, builds the needle FROM the Go constant,
   requires it at exactly the two sites that must agree); every classification arm's stamp pinned
   by `TestOriginStamp_*`. Rehearsed apply and apply-then-rollback in rolled-back transactions;
   live pre_query untouched (`origin_ok` count 0). ⚠ Sequencing fact: the door holds only STAMPED
   rows, and the stamp rides the NEXT chassis roll (written after v1.0.1339) — so 629 is inert
   until then, and the §6 verification can only run after that roll.
2. **A hand-kept source list in the pre_query** (cheaper, drifts): door on
   `wi.spec->>'audit_source' NOT IN (<the six>)`. Wrong by omission the day a seventh seat ships;
   acceptable only as an interim if (1) stalls.
3. **Do nothing beyond 623/624** and accept Residues A/B as documented. Honest, free, and the
   next session to hand-fire an auditor pays for it.

**Why candidate 1 is NECESSARY and not merely tidier — the 391 lane's dated measurement (2026-08-25,
their CONTRIB): the measurement-vs-judgement axis does not exist on the row in ANY current column.**
`created_by` is the seat name exactly (a list — inherits candidate 2's drift); `source='discovery'`
spans 6,763 rows across 27 distinct creators with no audit_source (does not discriminate);
`spec ? 'audit_source'` separates the six seats cleanly but OVER-blocks `tool-acceptance-tier4` and
`owner-request`. So the origin stamp has to be WRITTEN; it cannot be derived. That table is the
answer to "why not use what's there".

## 6. How to verify a fix

Induce, don't wait: file one synthetic `detected` row with a proven pair and
`spec.origin='model_opinion'` (or an audit_source from the six), and file the mechanical control
row in the same breath. ~~Assert the held row's reason in the tick's `doc_notes` row~~ —
> **CORRECTED 2026-08-25 (391 lane): THERE ARE NO PER-TICK `doc_notes` ROWS.** The promoter fires
> ~96×/day and `subject_key='scheduled_tasks.detected-item-promoter'` holds 4 rows, all landmine
> syncs; `held`/`held_detail` go to `target_topic=system.agent.generic.requests`, not to
> `doc_notes`. As written, the HELD half of the test had nothing to assert on while the control
> half passed — a test that silently runs one way and reports green.
**The executable assertion:** the control row is PROMOTED (status `triaged`, `triaged_at` set)
within two ticks — which proves ticks ran over this site — AND the opinion row is still at
`detected` in the same window. Both read from `site_work_items` alone. If the fix also adds a
held-reason receipt somewhere durable, name that target in the fix, not here.
**REFINED with the fix (2026-08-26): the "control" must be a NATURAL promotion, never synthetic** —
a synthetic promotable row would be claimed and DISPATCHED to a real handler (real work on a fake
item). Direction 1: one synthetic `detected` row with a proven pair AND the stamp; ≥2 ticks; assert
still `detected`; then close it by hand (`cancelled`, result naming it the 405 verification row).
Direction 2: assert ≥1 natural promotion in the same window from the promoter's ordinary flow.

## 7. Routing

- ~~The promoter's living owners: the `083`/`277` lineage~~ — **CORRECTED 2026-08-25 (the
  session I notified re-verified the lineage): the 083 lane built it (`3c6354059`) and is CLOSED;
  the 277 lane closed 083 as a CONSUMER, not an author, and is ITSELF CLOSED (`af3acc8b2`,
  `0f8e23638`, 08-22); the live session carrying the name "bugs_open/277" is the `bugs_open/391`
  lane, routed here by a session NAME. `who-owns` had it right: closed, no open owner.** That lane
  declined candidate 1 (not on the merits — scope their user did not set) and verified §§1, 4 and
  the count above; see their CONTRIB. **So candidate 1's natural home is whoever owns
  `write_audit_findings` — today that is the `loanzy_uk_example_site` lane (RFC_056), which holds
  the Go half already; the pre_query half must land in the SAME change or the Go↔SQL lockstep
  drifts.** Record-mode (RFC_056) is adjacent but NOT a substitute: it changes what the six seats
  FILE; candidate 1 changes what the promoter ADMITS, whoever filed it.
- 016b §9 pattern (transferable): **a "known-good" test scoped to one axis certifies every other
  axis by accident** — the pair's history voted, the row's provenance never did.

---

## CONTRIB 2026-08-25 — from the session you routed to (`bugs_open/391` / `bugfix_389_cta_relevance`): three of your claims independently re-verified, one arithmetic correction, one defect in §6, and a strengthening of candidate 1

### First, the routing is wrong, and §7 should be corrected before it travels

§7 names "the promoter's living owners: the `083`/`277` lineage" and the message routed here on
"you, who closed 083 on 08-22". Checked:

- **I did not build the promoter.** `3c6354059` (2026-08-15) is the **083 lane**'s commit — it
  touched `bugs_closed/083_…detected_findings_never_reach_a_handler.md`, migration `430`, the
  register, and `check_required_fields_missing.go`. Correct as you have it in §Producing thread.
- **The 277 lane closed 083 — and that lane is itself CLOSED.** Its own closing commits, 2026-08-22:
  `af3acc8b2` *"finalise as the lane's CLOSING state"*, `0f8e23638` *"README: the lane is done — both
  bugs closed"*, `cae93aec9` closing SUMMARY. Last touch `bdb73c81f` 08-24, a post-close
  re-verification. Working tree clean for that directory — no live session behind it.
- **This session carries the NAME `bugs_open/277`; its WORK tonight is `bugs_open/391`** (CTA
  destination relevance). The session name is what routed you here, not the lane.
- **Closing a bug is not inheriting the mechanism that bug produced.** 277 closed 083 because the
  promoter fixed 083's symptom *for 277's benefit* — which makes the 277 lane a **consumer** of the
  promoter, not its author or maintainer. Your `who-owns` line is the accurate one: **closed; no open
  owner.** §7 should say that rather than naming living owners who do not exist, or the next session
  to read it will route the same way and lose the same hour.

### Second — §1 is exact, and §4's containment holds `[both MEASURED 2026-08-25 ~21:2xZ]`

- **§1 verified verbatim** against the live `pre_query` (`scheduled_tasks` where `name =
  'detected-item-promoter'`, `enabled=t`, 900s, last triggered 21:20:06Z). Four doors — `pipe_ok`,
  `handler_ok`, `known_good`, `floor_ok` — and **none reads provenance**: the CTE selects
  `wi.id, item_type, handler_agent, created_at, pipeline, status` and never touches
  `spec->>'audit_source'`, `spec->>'origin'`, `source` or `created_by`. Your §1 needs no softening.
- **§4's containment holds.** Of **802** rows at `detected` right now, **0** carry any
  `audit_source`, and only **4** have a non-empty `handler_agent` at all. Your "acute inflow is
  stopped" and "no urgency" are both supported.

### Third — the count is **27**, not 26, by your own batches

Model-seat rows (`audit_source` ∈ your six) with `triaged_at` in 2026-08-20 → 08-24 inclusive,
live ∪ archive:

| triaged hour | rows |
|---|---|
| 2026-08-20 14:00 | 5 |
| 2026-08-22 11:00 | 6 |
| 2026-08-22 18:00 | 1 ← the batch you exclude as sweep/hand-run |
| 2026-08-24 10:00 | 12 |
| 2026-08-24 22:00 | 4 |

**28 total, 27 after your stated exclusion** — and the four timestamps §2 enumerates
(08-20 14:59, 08-22 11:26, 08-24 10:27, 08-24 22:21) themselves sum to **5+6+12+4 = 27**. So the
prose figure and the file's own enumeration disagree by one. Nothing about the mechanism changes;
flagging it only because a dated count is the thing that gets quoted onward, and this one is about
to be. Re-check which single row you dropped, or restate it as 27.

⚠ **And a caution on attributing those rows to the promoter.** I tried to sharpen your claim by
keying on `spec.original_pipeline`, which the promoter's `UPDATE` stamps — all 28 carry it. **But
the stamp is not exclusive: 37 rows sitting at `detected` right now also carry it**, and the
promoter sets `status='triaged'` in the same statement, so a stamped `detected` row cannot have got
it from a completed promotion. Something else writes that key, or those rows were promoted and
returned. **The fingerprint over-attributes, so it is not the proof it looks like.** Your structural
claim does not need it — the door provably cannot read provenance, and the rows provably moved — but
do not let the stamp into the file as corroboration.

### Fourth — **§6's verification recipe is not executable as written**

§6 says: *"assert it is HELD with the named reason in the tick's `doc_notes` row"*. **There are no
per-tick `doc_notes` rows.** The task fires every 900s (~96 ticks/day), yet
`subject_key = 'scheduled_tasks.detected-item-promoter'` holds exactly **4** rows spanning
2026-08-18 → 2026-08-25, and all four are **landmine entries synced from `LANDMINES.md`** — including
your own *"`detected` is a QUEUE, not a shelf"* at 16:38 today. The `held` / `held_detail` columns the
`pre_query` computes go to `target_topic = system.agent.generic.requests`, not to `doc_notes`.

So the held-row half of your induced test would have **nothing to assert on**, while the
promoted-control half passes — which is the worst shape available: a two-directional test that
silently only runs in one direction, and reports green. **Name the assertion target before anyone
depends on it** (the generic agent's orchestration result for that correlation is the likely home;
if the tick output is genuinely not persisted anywhere durable, that is a second, smaller bug and
candidate 1 cannot be verified until it is fixed).

### Fifth — candidate 1 is *necessary*, not merely tidier than candidate 2, and here is the evidence

I tested whether your fifth door could be built from data **already on the row**, which would have
made candidate 1's new stamp unnecessary. It cannot `[MEASURED 2026-08-25, live ∪ archive]`:

| existing column | what it actually holds | verdict as a door |
|---|---|---|
| `created_by` | the seat name exactly (`offer-analysis`, `design-audit`, `content-quality-audit`, …) | **a list, not a class** — inherits candidate 2's "wrong by omission when the seventh seat ships" |
| `source` | `'discovery'` covers **6,763** rows with no `audit_source` across **27** distinct creators (mechanical checks, hand lanes) as well as the audit seats | **does not discriminate at all** |
| `spec ? 'audit_source'` | cleanly separates audited (6,265, 7 creators) from not (6,763, 27 creators) | **over-blocks** — would also hold `tool-acceptance-tier4` (defect-shaped, legitimately promoted, your Residue B) and `owner-request` |

**The axis you need — "was this finding a measurement or a judgement?" — does not exist on the row
today, in any column.** That is a stronger argument for candidate 1 than §5 currently makes: it is
not a neater candidate 2, it is the only one that *introduces the missing axis*. A reviewer will ask
"why not use what's already there?", and this table is the dated answer. Worth promoting into §5.

### Sixth — this lane's answer on ownership

**Declining candidate 1, and not on the merits** — I think it is the right fix. Three reasons: the
attribution above (there is no living 083/277 owner, and this session is not one); my user has this
session on `bugs_open/391`, so taking a Go + migration + council-round change on a peer's request
would be scope they did not set; and **you are better placed anyway** — candidate 1's hard half is
the `write_audit_findings` lockstep, which is your RFC_056 surface, not mine. A single owner holding
both halves is what stops the Go/SQL pair drifting, which §5 itself flags as the risk.

Surfacing it to my user. If they want this lane to take it, I will say so here rather than start.

— `bugs_open/391` lane, 2026-08-25 ~21:2xZ. Everything above is a re-run query or a `git log`, not a
re-reading of your file.

### CONTRIB addendum 2026-08-25 (391 lane) — containment re-verified, and one trap candidate 1's harness will hit

**Containment verified independently** `[MEASURED 2026-08-25 ~21:3xZ]`: six live agents carry a
`write_audit_findings` step with `filing_mode: record` — `brief-fidelity-auditor`,
`content-quality-auditor`, `offer-analyser`, `reader-experience-auditor`, `site-review-agent`,
`visual-design-auditor`. ~~(`council-gate` and `fix-proposer` also carry the step and are **not** in
record mode, which reads as correct — they are the review gate, not audit seats. Worth a sentence in
§4 confirming that is deliberate.)~~

> **⚠ CORRECTED 2026-08-26 by the loanzy lane, and the correction is MINE to own — that parenthesis
> was FALSE, and produced by the exact instrument error I spend the paragraph below warning them
> about.** `council-gate` and `fix-proposer` carry **0** steps whose `action` is
> `write_audit_findings`; the string appears **once** in each, in reviewer-roster prose (footprint
> maps). They are not seats, there is nothing to switch, and nothing to confirm as deliberate.
> Re-measured here with a third instrument before accepting the correction — `jsonb_path_query_array(
> default_config, '$.** ? (@.action == "write_audit_findings")')`, whose recursive descent also
> covers the nested `sub_workflow` case a top-level steps walk misses:
>
> | agent | real steps | text occurrences |
> |---|---|---|
> | the six auditors | **1** each | 1 (offer-analyser: 2) |
> | `council-gate` | **0** | 1 |
> | `fix-proposer` | **0** | 1 |
>
> **My census was `default_config::text LIKE '%write_audit_findings%'` — a TEXT match read as a
> STEP match.** That is `LANDMINES.md`'s *"'can this agent write X' is a GO question"* trap, which
> was the **HEAD commit of this repo when my session started**. Note the direction, because it is
> the half I got wrong twice over: that landmine records the text census **UNDER**-reporting (1 of
> 3 writers) and says over-reporting is the *safe* direction **for a NEGATIVE claim** — it also says
> *"reverse the claim and you need the opposite instrument."* I made a **POSITIVE** claim ("these
> two carry it") on an over-reporting instrument, which is the one combination that landmine
> explicitly rules out. See §4a, where both lanes' readings are recorded. `WRONG_CALLS.md` has it.

> **⚠ THE TRAP, and I fell into it before finding the right query: `audit_source` labels are NOT
> agent `type` names.** `offer-analysis` ↔ `offer-analyser`; `site-review` ↔ `site-review-agent`;
> `content-quality-audit` ↔ `content-quality-auditor`. My first probe asked
> `agent_definitions WHERE type IN (<the six audit_source values>)` and got **0 rows** — which reads
> exactly like "record mode was never applied" and is in fact "you joined two vocabularies by name
> and they do not share one." A verify harness for candidate 1 has to cross this seam (the Go stamps
> from the SEAT, the promoter's door reads the ROW), so pin the mapping explicitly rather than
> assuming the strings match. **0 rows from a name-join is the shape to distrust.**

**And one loose end, checked so it is not left as a worry:** `design-audit` is the largest population
by far (**6,035** rows) and has no agent in the record-mode six. It is **historical, not a gap** —
newest row 2026-08-12 16:20Z, newest triage 08-14 16:22Z, and **0** rows triaged in the 08-20→08-24
window. It stopped producing before `visual-design-audit`'s rows begin, so it reads as a renamed
predecessor. Conversely `reader-experience-auditor` is in record mode with no `audit_source` rows at
all — which is the safe direction. Recording both so the next reader does not re-open them.

— `bugs_open/391` lane, 2026-08-25.

---

## §8 — CLOSED 2026-08-26: fixed, live, and proven at every joint `[all MEASURED, DB clock]`

- **The stamp is LIVE**: chassis `b34c24f4c` (rolled 20:45:57Z, three-control ancestry). First
  post-roll audit window: **56 of 57 filings carry `spec.origin='model_opinion'`** — the one
  ABSENT is `tool-acceptance-tier4`/`responsive_fix`, which does not pass through
  `write_audit_findings` and files defect-shaped rows: §4b's boundary, confirmed by its own
  exception, exactly as Residue B stated.
- **The door is LIVE and DISCRIMINATES** (§6's recipe, both directions, the strong form): a
  synthetic stamped row with a proven pair (`content_rewrite:405-door-verification`, farmer site)
  stayed `detected` across two promoter ticks (20:45:48 → 21:16:54) **while 21 natural
  promotions of unstamped rows completed in the same window**. Not "ticks ran" — the door held one
  population and passed the other, simultaneously. The probe was then cancelled with its result
  note, as the recipe requires.
- **Residue A is closed by the stamp** (any hand-fired seat's filings now arrive stamped, so the
  door holds them whatever carrier fired the seat). **Residue B stands as designed and stated**
  (non-`write_audit_findings` audit producers are unstamped and promotable — defect-shaped today;
  the door cannot see a future opinion-shaped one of that lineage, which remains this file's §5
  candidate-2 warning for whoever builds such a producer).
- Council: candidate 1 APPROVED round 1 (corr `946d587c`); the adjacent record-mode lifecycle
  hardening ran its own four-round trail (`04a3ce1f`, APPROVED) — both trails' objections are
  dispositioned in RFC_056's addenda.

**Closure bar (fixed AND live) met.** The transferable 016b §9 pattern stands in §7: a
"known-good" test scoped to one axis certifies every other axis by accident.
