# NOTES — `bugfix_305_negation_gate` (append-only, newest at the bottom)

## 2026-08-19/20, session 1 — research, measurement, and three refuted sub-designs

### Ownership and validity, checked before anything else

- `scripts/who-owns.py 305` → `copy_quality_two_stage` (ACTIVE, 81 commits/14d), 16 mentions across
  their handoffs and NOTES. Read `HANDOFF_2026-08-19_continue_here.md`: **no 305 fix in flight**;
  item 6 parks the brief detector's scheduling as an owner/architecture call; item 3 leaves the briefs
  to the site lanes. Writer-side half open.
- `site_work_items` where `item_type='needs_diagnosis'` and status not terminal: 4 open, none on this
  symptom. No open work item mentions negation/contrast copy.
- Bug still valid `[MEASURED 2026-08-19 ~21:30Z]`:
  - brief still supplies the tagline — `SELECT position('in days, not months' in data->>'formatted')>0`
    on `ai-agent-orchestration.com` `content_direction` `is_current` → **true**, 3,558 visible chars,
    row created 2026-07-24 and never updated;
  - the three pages still serve the quoted copy — `page_components.content_data`, all nine components
    `locked_at IS NULL`, `updated_at 2026-08-17` (a **rerender**, per the bug's §3);
  - the writer has not stopped — same site, 2026-08-19 18:26–18:32Z: *"not a catalogue built to look
    busy"*, *"not from provider marketing pages"*, *"not staging load"*.

### ⚠ MISSTEP 1 (mine, caught by a positive control) — `\b` in Postgres regex

My first distribution census reported `not X, but Y` = **0** and `rather than` = **0** across 1,503
writer calls. Both are false. I had pasted Go-shaped patterns (`\bnot …\b`) into psql, and **Postgres
ARE has no `\b` word boundary — there `\b` is a backspace character** (`\y` is the boundary;
`LANDMINES.md:4219`, and `WRONG_CALLS.md:17787` records another session making the identical mistake).
Caught by running the pattern against a sentence I *knew* matched and getting 0.

**Cheap check, now used for every pattern in this lane: assert the regex against a known-positive and
a known-negative string in the same query before quoting any count from it.** Logged in
`WRONG_CALLS.md`.

Corrected figures (`llm_call_log`, `agent_type='page-content-writer'`, `success`, 2026-08-13..19,
**1,503 calls ≈ sections**):

| shape | sections with ≥1 |
|---|---|
| `x_not_y` (`[a-z)"'],\s+(not\|never)\s+…`) | 631 (42%) — ≥2: 208 (14%) |
| `rather than` | 646 (43%) |
| `not X, but Y` (the only shape the Go detector has) | **23 (1.5%)** |
| negative reveal (`. It doesn't …`) | 168 (11%) |
| headline-class JSON field carrying `x_not_y` | 209 (14%) |

### What the existing machinery is and is not

- `platform/orchestration/datahelpers/voicetells.go` — `ScanVoice` is **site opt-in**
  (`ParseVoiceGate` returns nil without `voice_gate.enabled`): **9 of 43 sites**. `strawmanCommaRe`
  (:151) needs a trailing `, but`. So the estate's only wired detector is blind to both sentences the
  owner quoted. `FindStringIndex`, so at most one strawman finding per block.
- Callers: `discovery_checks/check_voice_tells.go` (post-deploy, files `voice_tells` at
  `needs_human_review`, **no handler** — 45 parked, 1 ever closed), `revalidate_voice_tells.go`
  (retractor), `save_page_meta_description_action.go:296-330` (**the only hard gate**, one sentence,
  and its header explains why it reuses the text-level entry point rather than the page-level check),
  `cmd/voicescan`.
- `page-content-writer` is effectively the whole writing population: **1,516 of 1,519** voice-carrier
  LLM calls in the last 7 days (`copy-editor` 3).
- `execute_llm_prompt`: 66 carriers, **no `ActionInputSpec`**, so `scripts/audit-optional-key-budget.sh`
  lists it under *"NOT COUNTED — the optional surface is UNKNOWABLE, not zero"*. A design that added a
  key there would be adding to the widest shared action in the estate, invisibly to the RFC_022 budget.

### The three refuted sub-designs

1. **Whole-section re-ask + "keep the lower score"** → adopts displacement. Neighbour baselines in the
   same corpus: `instead of` 5.9%, `isn't just/a` 6.4%, `more than (just)` 10.8%, `unlike` 0.3%,
   `without the/a` 4.5%, em dash 0.5%.
2. **"Verbatim in the rendered prompt" exemption** → `rather than` is in every rendered writer prompt
   (house voice ×6 + STRICT RULE 19), silently exempting the 43% arm.
3. **Quoting the house-voice rule in the repair prompt** → that rule's text carries the construction
   and a worked example of it.

Cost, measured rather than assumed: mean writer call **11,009 in / 2,126 out** tokens, **0 cached
prompts** (the template has no `<!--CACHE_BREAKPOINT-->`, and caching is opt-in by marker in
`platform/aiservice/anthropic.go`). Whole-section repair ≈ $0.072/call ≈ $200/month at 215
sections/day; the patch shape ≈ $0.0135.

### Decisions taken

- Scanner in `datahelpers` (pure), annotation default-ON in `render_component`, repair in its **own
  action** with its own input spec, page budget in `CollectedData`, migration held until the image is
  live. Full reasoning in `PLAN_2026-08-20_negation_gate.md`.
- Three lanes told before any code: `copy_quality_two_stage`, `site_ai_agent_orchestration`,
  `portfolio_positioning` (CONTRIB files dated 2026-08-20 in each lane's own directory).

### Council round 1 submitted — `SUBMISSION_CORR=c48b7612-3ecc-4345-912e-5966c079cb91`

8 edits (the schema cap), submission JSON kept in this directory. `DRY_RUN=1` first: admission passed
client-side before any credits were spent. Budget ~30 minutes, not ~2 — the dispatch queues behind the
fleet, and a missing orchestration row is latency, not a dropped dispatch:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c48b7612-3ecc-4345-912e-5966c079cb91';
```

### ⚠ MISSTEP 2 (mine, caught by a test that asserted what a regex MATCHED)

`capTokenRe` was `\b\pLu[\pL'’-]*\b`, intended as "a capitalised token". In Go's regexp the one-letter
form `\pL` takes **exactly one letter of class name**, so `\pLu` parses as *"any letter, then a literal
`u`"* — it matched `running` and `our`, and the invented-name check therefore rejected **every**
proposed rewrite. It reads correctly at a glance and it compiles.

Caught because a debug test printed `capTokenRe.FindAllString(to)` instead of asserting a verdict:
`caps: ["running" "our"]`. The two-letter class needs braces: `\p{Lu}`.

**Cheap check: for any `\p…` class, assert what it MATCHES on a two-arm fixture, never that a function
using it returns the answer you expected.** A wrong class makes the *whole check* fail closed, which
looks like a strict guard rather than a broken regex. Logged in `WRONG_CALLS.md`.

Two other gaps the tests found, both real:
- `x_not_y` required a **letter** after "not", so `"1,600 a day, not 12 a week"` — a number on the Y
  side — did not trip. Now `[\pL\pN]`.
- the exemption core was a single fixed character window, and the writer **extends** a supplied
  phrase: the brief supplies *"deployed to production in days, not months"*, the writer emitted
  *"…in days, not months on Kubernetes."*, and the wide window carried `on Kubernetes`, which the
  brief never said — so the tagline read as NOT supplied. Now word windows at k=4,3,2 with an
  18-character floor on the normalised form (the floor is what stops it narrowing to `s, not m` and
  exempting everything).

### Mutation probes run by hand (all three fail a NAMED test; package green when restored)

| probe | test that fails |
|---|---|
| drop `x_not_y` from `negationShapes` | `TestOwnersOwnSentencesTrip`, `TestShapeVocabularyIsStable` |
| drop the neighbour comparison in `AcceptNegationRewrite` | `TestRewriteRejectsDisplacement` — accepts *"instead of"*, *"more than just"* and an em dash |
| exempt on the matched fragment alone (i.e. prompt-wide) | `TestExemptionIsSentenceScoped` — the house-voice prose exempts a `rather_than` hit, which is the 43% hole |

### ⚠ The shared file was poisoned, and it changed the design (2026-08-20)

I wrote the annotation as two hunks inside `RenderComponentAction` and
`CompilePageSectionsAction`. It built and the local tests passed. Then the HEAD-overlay check
(`go build` over `git archive HEAD` + only my files) failed:

```
platform/orchestration/actions/v3_site_actions.go:6128: undefined: applyWorkItemFailureLadder
platform/orchestration/actions/v3_site_actions.go:6180: undefined: workItemDecisionStatuses
```

`git diff` on that file showed **four** hunks: two mine (2562, 2773) and two another session's
(6041, 6052), and the symbols theirs calls live in `work_item_failure_ladder.go`, which is
**untracked**. So a pathspec commit of `v3_site_actions.go` would have taken their half-finished
change as a same-file passenger and **broken the compile at HEAD** — and `make build-*` builds from
HEAD, so every session's next build breaks. My local build was green precisely *because* their
untracked file was in my tree.

**What I did instead of committing it:** reverted **only my two hunks** (`git diff` → keep my hunk
headers → `git apply -R`), leaving their work in the tree untouched, and moved my edit to a file
nobody else is in — `copy_gate_annotation.go`, wrapping the two handlers at their `registry.go`
entries. Verified by grep that nothing else in the repo calls either action directly, so registry
registration is complete coverage rather than a partial hook. It is also the better structure, so it
stays after the reason expires.

**Cheap check, and it is now this lane's rule before every commit of platform code:**

```bash
rm -rf $SP/head && git archive HEAD | tar -x -C $SP/head   # clean HEAD
# copy ONLY your own files over it, then:
cd $SP/head && go build ./... && go test ./<your packages>/
```

A green build in the working tree says nothing about HEAD on a tree this many sessions share.
Logged in `WRONG_CALLS.md`.

## 2026-08-20 — the build, and the four things that changed the design after it started

### Phase 1 built and committed (inert until the next chassis roll)

`negationtells.go` (scanner + neighbours + exemption + acceptance) · `negation_content.go` (the content-map
walker, with a `Set` closure so a repair writes THROUGH to the map the renderer reads) · `voicetells.go`
(the strawman arm now calls the family; `rather_than` as a density) · `copy_gate_annotation.go` (two
registry wrappers, counting, default ON) · `rewrite_negations_action.go` (the repair, its own
`ActionInputSpec`, registered) · migration `509` **_HOLD** + `_HOLD_ROLLBACK`.

⚠ **The migration is `509`, not `497`.** 497 was taken by another session between writing the plan and
writing the file — and the council's editquality seat flagged exactly that risk in the same round.
**Check `ls docs/agent_docs/sql_for_agents | grep '^NNN'` at WRITE time, not at plan time.**

### Phase 2 built, deployed and PROVEN LIVE

`cmd/brief-negation-check` + CronJob at 07:40 UTC, image `v1.0.1321`. Verified at the artefact, not the
log: 12 findings filed, **2 closed by the close-out arm on positive re-observation**, 10 open, and 4
`doc_notes` rows for 2 runs (two per run, because `backoffLimit: 1` retries a run that exits 1 — exactly
what the design says a red day looks like).

### ⚠ MISSTEP 3 — my inline edit would have broken the compile at HEAD for every session

Covered in full in `WRONG_CALLS.md`. Short version: `v3_site_actions.go` carried another session's
uncommitted work calling `applyWorkItemFailureLadder`, which lives in an **untracked** file. A pathspec
commit takes same-file passengers, so committing my two hunks would have shipped their half-finished
change to HEAD — which `make build-*` builds from. My local build was green *because* their untracked
file was in my tree. Caught by the HEAD-overlay build; fixed by reverting only my hunks (`git apply -R`
on a patch holding just my hunk headers) and moving the change into a file of my own.

### ⚠ MISSTEP 4 — the check flagged a company's policy as a mannerism

The first live fleet run flagged *"we do not offer refunds"*, *"we do not invent figures"* and *"we do not
charge"*. Those are limit statements, which the writer's own STRICT RULE 19 **asks for**. My
`negative_reveal` shape had first-person subjects in it. Fixed: third-person only (`it/this/that/they/
these`), both arms tested, fleet 12 → 10 sites, and the demand control still passes (the complained-of
site still shows its one MANDATED tagline). **The lesson is where it was caught: reading the finished
detector's output against live data, not reading the regex.**

### Council round 1: REVISE — and it was worth far more than the 20 minutes it cost

Gating objection: a `sub_workflow`'s running half is often keyed `substeps`, not `steps`, in which case
the migration inserts a step nothing runs while the RAISE still passes. **The council's own read-only
check answered it** (`has_substeps=false, has_steps=true`) — and the seat was still right that the guard
was checking the wrong thing, so the migration now anchors on the container path.

**The one I would not have found:** compliance, HIGH — nothing scanned the accepted replacement for
**banned claims**. The acceptance test was structural (facts, links, markup, displacement) and could not
tell an honest reframing from an overclaim. "Say what it IS" is precisely the pressure that fills the
affirmative slot with an invented superlative, and no gate downstream inspects a spliced sentence: by
deploy time it IS the page. Now every candidate goes through `checkBannedClaims` + `loadEvidenceBase`
with the fleet arm on, BEFORE the structural test.

Also fixed from that round: truncation (a cut repair splices nothing, both arms); the repair prompt now
forbids new capability/reliability claims explicitly; tag comparison moved from multiset to **sequence**
(`<b><i>x</b></i>` has the same tags as the well-formed version); and un-holding `509` now has TWO
preconditions, the image AND the per-page budget canary.

Answered with evidence rather than changed: `apply_section_edit` needs a `page_component_id` and edits a
**persisted** row, so it cannot be reused at a seam where the section does not exist yet; the
`bugs_open/119` re-ask fires only on an unparseable answer and re-asks the whole section.

### Mutation probes, phase 1 repair (all fail a NAMED test; package green when restored)

| probe | test that fails |
|---|---|
| count exempt hits against the budget | `TestExemptHitsDoNotConsumeTheBudget` |
| make headline hits obey the budget | `TestHeadlineHitIsAlwaysATarget` |
| key `matchTarget` on the FIELD name | `TestMatchTargetIgnoresRenamedField` |
| drop the per-page carry in `CollectedData` | `TestBudgetIsPerPageNotPerSection` |
| drop the `nonProseFieldRe` test in `prosey()` | `TestWalkerSkipsNonProse` |
| drop the `err != nil` early return in the wrapper | `TestAnnotationPassesErrorsThrough` |

### Anchors RE-VERIFIED after the config moved under me (2026-08-20 ~15:40Z)

`page-content-writer`'s row was updated at **10:16:39Z** — after I read it at ~09:20 and after the
council's round-1 check quoted it. Another session's edit. Re-checked every anchor migration `509`
depends on, because a step-chain migration written against a stale read is the failure this whole file
is about:

| anchor | value now |
|---|---|
| `…sub_workflow,steps,generate_content,action` | `execute_llm_prompt` ✓ |
| `…sub_workflow,steps,generate_content,next_step` | `render_section` ✓ |
| `…sub_workflow,steps,render_section,action` | `render_component` ✓ |
| `…sub_workflow,steps,rewrite_negations` | absent ✓ |
| `…sub_workflow,substeps` | **absent** ✓ (the council's gating concern, still answered) |

All hold, so `509` is unaffected. **This is the argument for anchoring and needle-gating rather than a
blind `jsonb_set`**: the migration would have refused (0 rows, loud RAISE) if any of them had moved,
instead of minting an orphan key on a chain that no longer exists. Re-run this table before applying
`509` regardless of what this note says — it was true at 15:40Z and the tree moves hourly.

### Council round 2 (REVISE) — and it found the estate's own trap in my code

compliance and render_guardian flipped to **approve**, so round 1's fixes held. Two new real defects:

1. **`guidelines`: my `item_key` was coarser than my finding.** `brief-negation:<site_id>` against a
   finding that is a LIST of phrases — and a brief is edited by config at any moment. That is the
   documented trap almost word for word: *"the second, DIFFERENT finding hits `ON CONFLICT DO NOTHING`
   and is gone, and the open row goes on describing the first thing it ever saw"* (measured elsewhere
   at four of five open items naming the WRONG facts, with nothing ever erroring). I had even written
   the consequence into the code comment as an acceptable cost, and justified it by the reaper
   landmine — which is the more interesting mistake: **I reasoned about one landmine carefully enough
   to talk myself past another.**
   Fixed: the key carries a sha256 digest of the sorted phrase set. Both landmines now hold — no
   finding is dropped, and no open row is rewritten daily. **Proven in production on the same
   afternoon** (v1.0.1322): one run closed all 12 site-keyed items and refiled 9 under phrase-set keys.
2. **`llm_reliability` (HIGH): my truncation second arm could not fire.** `outTok >= sent` with
   `outTok` defaulting to 0 when a provider reports no usage is always false, so the arm silently
   never ran — indistinguishable from "the answer was complete". The three states are distinct now and
   UNKNOWN is logged. Stated plainly in the code: that arm was never the load-bearing protection.

**The scope objection was ROUTED, not argued.** `guardian` (HIGH) and `architecture` both said the
default-ON annotation on two of the most-invoked actions is a shared-contract change that arrived
inside a single-agent bug fix, and that RFC_022's exception does not cover default-ON. Both are
correct. Per the 2026-07-28 ruling that a scope veto is recorded and routed rather than resubmitted
with better measurements: `architecture_review/RFC_044_default_on_annotation_on_two_shared_render_actions.md`.

**One consequence worth recording, because it looked like a loss and was a gain:** after the
attribution fix, `loanzy.uk` moved from *1 supplied / 1 regulatory* to *0 / 2* and the fleet went 10 →
9 sites. Checked rather than assumed: its phrase is *"not a lender, not a broker"*, and the reveal hit
now attributes to the sentence that actually carries it, which the regulatory rule then correctly
exempts. A better answer, not a dropped one.

Round 3 dispatched on the same correlation.

### The measurement that changed the design last, and it came from an objection I nearly skipped

Round 2's `guardian` had a second objection I did not action in round 3 because it read like a request
for an estimate: *"the plan doesn't estimate the delta in findings volume this produces against the
existing `voice_tells` review queue"*. It is not an estimate request. It is a load-bearing number, and
it lands on **another lane's** queue.

`[MEASURED 2026-08-20]` over the 189 live, unlocked pages of the 9 sites with a voice gate:

| | pages flagged |
|---|---|
| today (the two original shapes) | **14** |
| `x_not_y` as a per-hit finding | **139** |
| + the two-sentence reveal | +46 |
| + `rather than` at >2 | +39 |

A **tenfold flood** into a queue that holds 45 parked `voice_tells` items and has had exactly **one**
closed by a human, ever. A check that flags three quarters of the estate's pages tells nobody anything,
and the cost would have landed on whoever drains that queue rather than on me.

**So the shape changed:** the two original shapes stay per-hit (volume unchanged); the three broad ones
feed a page-level `negation_density`. Its default is **>12**, set from the curve rather than from taste
— it flags 14 of 189, which is exactly today's volume. The full curve is in the code comment
(8 → 43 pages, 5 → 61, 3 → 87, 1 → 150) so a site can lower it deliberately once `bugs_open/033` gives
that queue a working surface.

**The division of labour this produces is the honest one, and it is better than what I had:** the
writer-seam gate enforces the real standard (the house voice's once or twice per page) at the moment
the copy is written, where a repair is automatic and costs no human; the post-deploy check keeps a
higher bar, because every finding there costs a person. ⚠ This landed AFTER the round-3 submission, so
round 3 is reviewing the pre-density version — if the same objection returns, the answer is "measured,
fixed and committed at `7639dacf4`".

### ⚠ MISSTEP 5 — I answered a scope objection with a document, twice, and the council vetoed it

Round 3: **REJECTED**, guardian veto. The reasoning is the correction:

> *"the code under review for edit 4 has not changed since round 2's HIGH objection — only the
> paperwork around it has … Routing a scope objection to architecture review does not license
> deploying the disputed change … 'we wrote it down and routed it' is not the same as 'it was
> contained.'"*

**What I got wrong, precisely.** I read the owner ruling of 2026-07-28 — *"the code stays and the
precedent gets fixed … record it where the change lives, route the seam to architecture review"* — as
a standing rule a session may invoke. It is not. It is an **owner's decision about one case**
(`bugs_closed/124`), and reading a one-off ruling as a general licence is how a precedent becomes
folklore. Two seats flagged the same thing at HIGH for two consecutive rounds and I answered with an
RFC file both times.

**The cheap check, and it is a question not a procedure: after an objection, has the CODE changed, or
only the writing about it?** If a reviewer could re-read the diff and see nothing different, nothing
has been answered — however good the note is. Logged in `WRONG_CALLS.md`.

**What containment cost, so nobody has to rediscover it:** the counting is now opt-in per step,
default OFF, on `page-content-writer` alone. Outside that agent *"the copy improved"* and *"the check
was not wired here"* are the same number again. `RFC_044` carries the question of whether it should go
back, and it is no longer urgent — which is the version of this argument a human should settle.

### The reuse follow-up this lane owes (round 3, reuse_agent, medium)

The truncation three-state (`known-at-ceiling / known-below / UNKNOWN`) is written inline in
`rewrite_negations_action.go`. The seat is right that `output_tokens >= max_tokens` is a named
fleet-wide trap and that this wants to be a shared helper — something like
`aiservice.TruncationState(outTok, maxTok)`. **Deliberately not done in round 4**, because extracting
it means touching other actions' truncation handling inside a round whose whole purpose was to CONTAIN
scope. It is the next reuse step, and whoever takes it should audit the other call sites of that
predicate at the same time rather than moving one.

### Council round 4: APPROVED (`c48b7612`) — REVISE → REVISE → REJECTED → APPROVED

11 of 14 seats approved outright, including `guardian` once the seam was contained. 4 advisories at
medium, none high. Two closed in code the same hour:

- **`compliance`** — the anti-fabrication guard leaned entirely on `checkBannedClaims`, which only
  catches patterns a site has **armed**, and the register is sparse. `AcceptNegationRewrite` now
  rejects a rewrite that INTRODUCES a superlative the original did not carry (`invented_superlative`),
  with the author's own words explicitly allowed through. This is the objection I would least have
  found myself: my guard checked that nothing was LOST and never that something was GAINED.
- **`bug_historian`** — "never returns an error for a style outcome" read as swallowing infrastructure
  failures too. Still never fails the step, but an infra failure is stamped `repair_unavailable` and
  logged at Error, so a census can find runs where the gate was **present and blind**.

Two recorded and deliberately not acted on: `architecture` (RFC_022's exception is still not claimable
even with the default OFF — its third condition is not literally met, because migration `509` names
the key; **contained is not exempt**, and RFC_044 is where that is settled) and `reuse_agent` (the
truncation three-state wants to be a shared `aiservice` helper — the next reuse step, and whoever
takes it should audit the other call sites of that predicate rather than move one).

**All 11 platform commits are credited to the approved correlation automatically** — the
`Council-Submitted:` trailers resolved at report time, with no amend, exactly as that mechanism is
designed. Verified in the `098` report.

## 2026-08-21 — both halves LIVE, and the first live run found the gate was blind

### The roll, probed rather than assumed

Chassis `v1.0.1321`, both replicas: `rewrite_negations` **7**, `copy_gate_annotate` **1**, control
`rewrite_negationz` **0**. ⚠ `invented_superlative` **0** — the build predates commit `1ac9b8890`, so
the hype-superlative guard is not live yet. The **accuracy-claim family is covered anyway**: the
fleet-wide banned-claim set already catches *always accurate*, *definitive*, *guaranteed accurate*,
*every claim is verified*, *never wrong* (`claims_global.go:112-190`). What waits for the next roll is
*industry-leading*, *best-in-class*, *100%*, *flawless*. Bounded, and stated rather than discovered.

Migration `509` applied 10:28Z; `517` at 10:40Z. Both hand-applied and both now in the runner's ledger
via `--record-only`.

### ⚠ MISSTEP 6 — my own precondition was written in an impossible order

`509`'s precondition (2) said the per-page budget canary must pass BEFORE applying. It cannot: the
marker it reads only exists once the step runs. Corrected in the file. What WAS measurable beforehand
and was measured: non-output `CollectedData` keys do reach the durable row (all 13 recent writer
orchestrations carry `agent_config` and `__my_requests_topic__`). **That turned out to be the wrong
inference** — see below.

### ⚠ MISSTEP 7, and it is the one that mattered: THE GATE WENT LIVE BLIND

First page built after `509` (orch `8ce1ebc0`, iter 1, 10:31Z):

```json
{"status":"repair_unavailable","error":"no ai_service configuration resolvable",
 "hits_before":3,"targets":1,"within_budget":2,"rewritten":[],"hits_after":3,"page_hits":3}
```

Detecting correctly. **Repairing nothing.** `resolveAIServiceConfig` reads the agent's ROOT
`ai_service` and `workflow.steps.<currentStep>.config.ai_service`; `page-content-writer` has no root
block (its model sits on `generate_content`) and `currentStep` for a loop substep is
`process_sections_loop_iter_N_rewrite_negations`, which is not a top-level step. Neither lookup could
resolve. Fixed by `517` (declared on the step, not fetched from the sibling — a step whose model comes
from another step's config is what nobody finds when the model changes).

**The thing worth keeping: this was visible in ONE query, and only because the council made me
distinguish an infra failure from a style outcome** (round 4, `bug_historian`, medium — an advisory I
could have banked). Without `repair_unavailable` and its Error log, the gate would have been live,
silent, and returning a clean-looking status while repairing nothing: the armed-but-inert shape this
estate keeps recording. **The advisory earned its keep within a day of being closed.**

### The `__copy_gate` key does NOT reach the durable row — and the budget question is STILL OPEN

Same run: `copy_gate`, `copy_gate_0`, `copy_gate_1` all present (the step's **output_field**, which
`saveStepResultWithRetry` copies) — and `__copy_gate` **absent**. So that function's fresh-state copy
does drop my bare key, exactly as reading it suggested; the `agent_config` evidence was misleading,
because that key survives by being written before an AWAIT, where the full state is persisted.

⚠ **And this run could NOT answer the accumulation question**, which is the discipline this lane keeps
having to re-learn: iteration 0 had **0** hits and iteration 1 had 3, so `page_hits: 3` is consistent
with accumulating AND with resetting. **A measurement that cannot come out either way is not a
measurement.** It needs a page where two sections both carry hits. Until then the honest statement is:
the per-page budget is **unproven**, the fallback (per-section budget, headlines always repaired) is
safe, and the page-level number is counted at `compile_page_sections` regardless.

### 11:06Z — THE REPAIR WORKS ON REAL COPY, and the no-regression control passed too

**The repair** (`mortgagecalculator.co.uk`/`scorecard-simulator`, `mechanism-flow`): 5 hits found, **1
left alone as regulatory**, 2 allowed by the per-section budget, **2 rewritten, 0 rejected**,
`hits_before 5 → hits_after 3`. One call, 447 output tokens. Both rewrites surgical:

- *"The result breaks down by area **rather than giving you one verdict**: your income…"* → *"The result
  breaks down by area: your income…"*
- *"…the simulator shows that kind of trade-off **rather than a flat pass or fail**."* → *"…the
  simulator shows that kind of trade-off."*

Every element of the design fired correctly on its first real page: the exemption, the budget, the
selection, the acceptance test, the splice.

**The no-regression control** (`webdesign.co.uk`/`tool-social-card-guide`): gate ran, `hits_before 0`,
`status clean`, orchestration **COMPLETED**. A page with nothing to fix passes through the new step
untouched and the build finishes — which is the half of "it works" that only a clean page can show.

**⚠ Still owed: artefact-level proof.** The marker is a status. The only page that has tripped the gate
cannot render, for an unrelated reason (`bugs_open/260`'s type gate on `mechanism-flow` —
`steps[N].branches` arrives as prose where the schema declares objects). **Not caused by the gate**, and
the control is documented rather than asserted: on the 10:30 run the repair spliced nothing
(`repair_unavailable`) and the identical failure occurred, and `steps[1].branches` fails too, which the
gate never touched. Told the 260 lane with the census query.

**Traffic is the constraint, not the code:** one writer run in the last hour. Three gate runs total —
clean/COMPLETED, repaired/FAILED-at-260, repair_unavailable/FAILED-at-260.

### 11:52Z — the first MULTI-HIT page found a defect in my own repair, and settled the budget question

`webdesign.co.uk`/`tool-social-card-guide`, iteration 1: `hits_before 8 → hits_after 7`, **`rewritten:
6`**, `rejected: 0`, **1 distinct field**.

**Six accepted, one landed.** Every target of a field carries the same captured original, so each
accepted replacement spliced against *that* and wrote the whole field back — last writer wins. Fixed at
`0eea9e597` (carry each field's text forward), pinned by a test built from this page and a mutation
probe that reproduces the race. Inert until the next roll.

**⚠ MISSTEP 8, and it is the one I am least comfortable about: I nearly confirmed the repair from a
check that could not fail.** My first artefact query asked whether each rewrite's `to` prefix appeared
in the stored content. Five of six said yes — and it was worthless, because `from` and `to` share their
opening, so the prefix matches either way. The honest check is the part that DIFFERS: is the removed
construction still there? It is — *"rather than compete"*, *"rather than trust that…"*, *"rather than
requirement"*, and six `rather than` in the field overall. **This lane has now written the "a
measurement that could not come out otherwise" lesson three times and I still did it under the
pull of a result I wanted.** The marker also truncates `from`/`to` at 160 chars, which is what made the
prefix the convenient thing to test.

### The per-page budget: ANSWERED, and the answer is NO

The same run was the first that could discriminate (its iteration 0 had a hit):

| iter | hits_before | page_hits | within_budget |
|---|---|---|---|
| 0 | 1 | **1** | 0 — the hit was headline-class, repaired regardless |
| 1 | 8 | **8** (not 9) | **2** — a fresh budget |

**Per SECTION.** As `__copy_gate`'s absence from the durable row predicted: `saveStepResultWithRetry`
copies only the step's own keys. The safe fallback is what is live; headline hits are still always
repaired; the page total is still counted at `compile_page_sections`. A true per-page budget needs the
count carried in the step's OUTPUT (`copy_gate_<N-1>`), never a bare state key.

### And the no-regression control, which only a clean page can give

`webdesign.co.uk`/`tool-social-card-guide` iteration 2 and the earlier `tool-social-card-guide` run:
`hits_before 0`, `status clean`, orchestration **COMPLETED**. The new step does not disturb a healthy
build.

### ⚠ MISSTEP 9 — the repair was being made and thrown away, and I had the evidence two hours earlier

Two pages built after the roll, both COMPLETED, both reporting `status: repaired, hits_before 1 →
hits_after 0`. The stored `content_data` was **byte-identical to the pre-repair value**.

The in-place mutation of the writer's content map does not survive the step boundary:
`saveStepResultWithRetry` reloads a fresh state and copies only the CURRENT step's own
`stepName`/`output_field`, so an edit to the PREVIOUS step's output is dropped and the renderer reads
the unpatched map.

**I had already measured the mechanism.** When `__copy_gate` turned out to be absent from the durable
row, I concluded "the page counter will not persist" and stopped. The same sentence explains why an
in-place content edit does not persist either, and I did not follow it through — I had a working
theory for the symptom in front of me and did not ask what else it predicted. **A mechanism you have
just proved is a tool for the next question, not only an answer to this one.**

Ruled out before fixing: `render_component`'s `merge_with` overlay wins conflicts, so it could have
overwritten the patch. It did not — the stored text is the LLM's own original, not a resolved value.

**Fixed:** the step now returns the patched content as its own `result`, and migration `548`
(**HELD**) points `render_section.content_from` at it.

**Precisely where this lane now stands: the gate detects correctly, selects correctly, rewrites well,
and does not yet change pages.** One roll, then `548`, then one page to confirm at the artefact.

### MISSTEP 10 — backticks in a commit message, which is in the fleet memory index

`git commit -m "…returns the patched content as its own \`result\`…"` executed the backticked word;
bash reported `result: command not found` and the message committed with the word missing (`dd9fc619`,
corrected in `99ee9a5e2`). It is a documented trap and I read the line this session. Third documented
trap in two days — **the signal is the rate, not the miss: known traps get hit when the content of a
message feels more important than its mechanics.** Use `git commit -F -` with a quoted heredoc.

## Session 2026-08-22 morning — the roll landed, 548 applied, the wiring is complete

**Step 1 (the roll): DONE, verified with controls.** Both chassis replicas
(`agent-chassis-74ffb74b8d-4qlp7` / `-qp8kk`, pods up ~08:36Z) carry build stamp
`70e7b4f9cabb9676e34131c52d06966b5d62e97e` (`grep -ac` in `/proc/1/exe` = 3 on both; absent-control
`f11851c2…` — committed 09:15Z, after the pods started — = 0 on both).
`git merge-base --is-ancestor dd9fc6197 70e7b4f9c` → **yes**, and the control commit is NOT an
ancestor, so the test discriminates. The migration header's own capability probe also run:
`copy_gate` = 7, `copy_gatez` = 0, both replicas.

**Step 2 (migration 548): APPLIED 2026-08-22 ~09:19Z, recorded 09:20:25Z.** Pre-apply anchor check
read all three needle values on the live row (`content_from = generated_content.result`,
`rewrite_negations` present, `render_section` action `render_component`). Renamed out of `_HOLD`
(the deliberate act, both file and `_ROLLBACK`), hand-applied with `ON_ERROR_STOP`: `UPDATE 1`,
verify `DO` block passed (`548 OK`), recorded via `--record-only`. Live row now reads
`copy_gate.result` (re-read after apply, not inferred).

**Wart, cosmetic:** 548's `snapshot_agent` reason string still says `518_…` — the pre-renumbering
name. The snapshot itself is fine (`5946a27b-38ab-41e8-8b49-7bc1a4b626b8`); anyone reading the
snapshot ledger for "why" should map `518_render_section…` → `548`. Not edited post-review; noted here
instead. Also noted: the number 548 is SHARED with another session's unrelated
`548_seed_webdesign_uk_theme_row_from_deploy_repo.sql` — the ledger keys on full filename, so no
collision in the ledger, but a bare "548" is ambiguous in prose from now on.

**Checked before trusting the new wiring: every exit path of `RewriteNegationsAction` that can precede
`render_section` carries `result`.** The two returns without it are the `initialize` lifecycle call
(not the section pass) and `no_content` — which only fires when `generated_content.result` was empty,
a state the OLD wiring also could not render. The marker sets `result` unconditionally whenever a
content map exists (the comment at `rewrite_negations_action.go:163` exists precisely so a clean page
is not dropped).

**Boundary accounting for anyone reading markers later:** runs COMPLETED before 09:20Z today —
`loanzy.uk/tool-loan-repayment-calculator` 09:10Z (`repaired 6→5`), `loanzy.uk/tool-loan-vs-savings`
09:14Z (`repaired 4→4`), and the overnight `remortgagecalculator.uk/mortgage-lenders` pair — still ran
under `generated_content.result`, so their repairs were made and thrown away (§22 behaviour). Their
markers are honest about the map, wrong about the page. The first run whose SPAWN postdates 09:20:25Z
is the first that can prove anything.

**What remains for this lane: one page, at the artefact.** A post-09:20Z page with `status: repaired`,
then `page_components.content_data` must LACK the removed construction (verify on the part that
DIFFERS, never the rewrite's opening; never `updated_at`). Traffic is live (a `page_rerender` was
claimed at query time; 235 unresolved queued), so this should arrive without prompting.

## 2026-08-22 ~09:40Z — the first post-548 page PROVED the gate through render and compile, and an unrelated guard blocked the save

`loanzy.uk/tool-interest-rate-stress-test`, spawned 09:23:03Z (post-548), COMPLETED 09:24:31Z,
`copy_gate_2: repaired, hits_before 6 → hits_after 5`.

**What is now PROVEN that was not yesterday:**
- The rewrite itself: iteration 2's `generated_content_2.result.content` (2631 chars) vs
  `copy_gate_2.result.content` (2582) — word-diff shows ONE surgical change: *"…is the right next
  step, rather than trying to borrow your way around it."* → *"…is the right next step."* Nothing
  else touched. ⚠ The marker's own `from`/`to` were USELESS here — both truncate at 160 chars and the
  pair shares its opening; the diff of the two durable content fields is the working recipe.
- **`render_section` read the patch**: `section_output_2.rendered_html` carries the rewritten
  sentence. This is the thing §22 proved was NOT happening.
- **`compile_page`/`complete` page_html carries the rewrite and LACKS the removed construction**
  (checked both polarities). The repaired copy survives the writer's entire in-run pipeline.
- **The gate hands clean sections on byte-identical**: `copy_gate_0.result = generated_content_0.result`
  (true, and for iter 1) — so 548 does not perturb untouched content.

**What is still NOT proven: persistence.** `page_components` was not updated — the writer's PARENT
(`7ff636c3`) ended at `complete_error`: `save_page_sections` REFUSED the whole save on the
`bugs_open/253` (framework_rewrite slug) SECTION COMPONENT FLOOR — "hero-tool 12→5 class attributes
(42% kept, floor 50%) … Nothing was written". **NOT caused by the gate**, two controls: (1) the hero's
input from the gate was byte-identical (above); (2) the 5-class hero is today's renderer generally —
three OTHER loanzy pages saved fresh this morning (09:14/09:18/09:21Z) all store 5-class `hero-tool`,
while this page's stored hero is 08-18 vintage with 12. Any resave of an 08-18-vintage loanzy tool
page trips the floor no matter what the copy says. Same shape as §0b's `bugs_open/260` blocker: the
proof page died at the NEXT step, on a pre-existing condition, with a control. Observation contributed
to the 253 file; the work item was marked failed by the parent (`mark_item_failed`), so the immune
sweep sees it.

**So the remaining owed item narrows to: one post-548 repaired page whose save is ACCEPTED** (a fresh
page, or one whose stored structure matches today's renderer). Watch armed for the next terminal
repaired run after 09:25Z.

### The negative control that makes the pending proof exact (2026-08-22 ~09:50Z)

`loanzy.uk/tool-loan-repayment-calculator`, built **09:10Z — before 548 applied at 09:20:25Z** — and
its save was **accepted** (unlike the stress-test page). Ran the new §8 RUNBOOK query:
`gate_changed_something=true, stored_matches_PRE_repair=true, stored_matches_POST_repair=false`.

So: same pipeline, same morning, save accepted, and **the repair was still thrown away**. That is
§22's defect measured a third time and — because this page SAVED — it isolates the variable. The
post-548 proof is now a single inversion of one boolean pair on the same query, not a fresh argument.
Recipe with both failure modes and the parent-check for a refused save: `RUNBOOK` §8.

⚠ Place runs either side of the migration by **`created_at`**, not `updated_at` — the stress-test run
was still writing at 09:24 and its `updated_at` sits after the apply while its whole section pass ran
before it. (I nearly used `updated_at` and it would have mislabelled the boundary.)

**Traffic state at 09:50Z:** the loanzy batch that produced this morning's five writer runs is
finished (4 `needs_page` complete, 1 triaged after the floor refusal); no writer run has started
since 09:23. Loanzy builds are another lane's live work (see `bugfix_311` notes, "loanzy opened: four
pages unarchived, four builds filed"), so this lane should NOT fire its own rerender there and
compete — the proof rides their next build, or any fleet page that trips the gate.

### MISSTEP 11 — my LANDMINES correction was swept into another session's commit, from the other side of the known trap

I edited the copy-gate landmine entry (its "until BOTH have shipped, verify at the artefact"
conditional had expired now that `548` is live) and ran `landmines-verify-dispatch.sh` — 3 entries
dispatched. My `git commit <path> -F -` then failed with *"no changes added to commit"*: between my
Edit and my commit, another session committed `LANDMINES.md` for its own entry (`a93fc3ffd`, the
RELEASE_IMAGES landmine) and **took my edit as a same-file passenger**.

**Nothing is lost** — `git show HEAD:…LANDMINES.md` contains my correction, verified with an
unwrapped grep (`tr '\n' ' '` first, because the file is hard-wrapped and a line-oriented `grep -F`
reports false absences on a long phrase). Forward-only holds; there is no remainder to re-commit.
Recorded because the trail is now wrong in a specific way: **the copy-gate landmine's 2026-08-22
correction is authored by this lane but lives under a commit message about release-image coverage.**
Anyone doing `git log` on that entry will not find this lane. This is the documented same-file
passenger trap experienced from the LOSING side, and no pathspec discipline of mine could have
prevented it — the only lever is the gap between edit and commit, which for a shared append-only file
should be seconds, not the length of a verify-dispatch run. **Commit the file edit FIRST, then run
the dispatch.**

### Blast-radius check on the migration I had just applied — and a third, independent proof the roll landed (2026-08-22 ~10:45Z)

`548` makes `render_section` read `copy_gate.result` for **every site**, so I went looking for the way
that could go wrong rather than assuming it could not. `render_component` resolves `content_from`
through `extractContentWithFallbacks` (`v3_site_actions.go:2248`), and that helper's **second**
candidate is the BARE step key. So a marker with no `result` — non-empty — resolves to the marker
map, and the section renders from `status`/`content_from` keys: wrong, not empty, and nothing errors.

**Measured before deciding anything: 0 of 131 markers all-history carry `no_content` or
`initialized`** (the two `result`-less exits). So it is a coupling to preserve, not a live defect;
recorded as the sixth face in `LANDMINES.md` with the explicit warning not to "fix" it by widening
the fallback (that bare-key branch is what `bugs_open/199`'s envelope normaliser contains).

The same census gave a **behavioural probe that needs no binary access**, and it is a cleaner
instrument than the grep I used this morning: every marker written by the fixed binary carries
`result`, and no earlier one does — **28 of 28 post-roll (18 `clean`, 10 `repaired`) against 0 of
103 pre-roll.** A split with no overlap. It independently confirms the roll AND proves `result` is
set on the CLEAN path, which is the property that stops `548` dropping untouched sections. Recorded
in the landmine entry so the next person checking "is the `result` half live?" does not have to exec
into a pod.

### CORRECTION to misstep 11 — my "nothing was lost" check was too weak to have found a partial loss

I wrote above that I verified my swept LANDMINES correction with `tr '\n' ' ' | grep -c`. A peer
session logged the same technique in `WRONG_CALLS.md` the same hour and is right: collapsing the file
to ONE line makes `grep -c` return **1 whether the phrase appears once or fifty times**, so it
establishes that my FIRST phrase is present and says nothing about whether the rest of the edit
survived. My conclusion was correct; the instrument could not have told me otherwise.

**Re-verified properly, at word level, against the sweeping commit itself** — counts of distinctive
tokens in `a93fc3ffd` vs its parent: `70e7b4f9c` 0→1, `09:20:25Z` 0→2, `tool-interest-rate-stress-test`
0→1, `RUNBOOK_negation_gate.md` 0→1, `created_at` 170→171; and the full closing sentence of my
correction is present at HEAD. So the whole edit rode in, not just its opening.

⚠ And an accidental control worth keeping: my sixth probe token, `EXPIRED`, returned 0→0 — because I
had written the word in lower case. **A test whose tokens can come out zero is the only kind worth
running**; had every token returned a hit I would have learned nothing about the check itself.

## ✅ 2026-08-22 10:00Z — PROVEN AT THE ARTEFACT. The gate changes pages.

`loanzy.uk/tool-interest-rate-stress-test`, rebuilt 09:57:01Z (post-548), COMPLETED, parent
COMPLETED at `complete`, components saved 09:59:25Z.

**Marker:** `copy_gate_2` — `hits_before 8 → hits_after 2`, **6 rewritten, 0 rejected**;
`copy_gate_0`/`copy_gate_1` each 1 hit, both **exempt** (brief-supplied, left alone by design).

**RUNBOOK §8 on it: `gate_changed=true, stored_PRE=false, stored_POST=true`** — the exact inversion
of the pre-548 control measured 90 minutes earlier on `tool-loan-repayment-calculator`
(`stored_PRE=true, stored_POST=false`). Same query, same site, same morning, opposite answer.

**The six removals, and both polarities checked at `page_components.content_data`:**
*"…happens rather than after." → "…happens."*; *"interest rather than paying down the loan itself,"*
→ *"interest,"*; *"years rather than just to next month's bill."* → *"years."*;
*"budget, rather than theirs,"* → *"budget"*; *"point rather than a fixed answer,"* → *"point,"*;
*"service rather than waiting to see how things unfold,"* → *"service,"*.
**0 of 6 removed clauses present in the stored artefact; 6 of 6 replacements present** (the second
count is the demand control — it proves the LIKE could have matched).

**This also closes handoff item 1b at the artefact — the same-field splice race.** All six rewrites
were in ONE field. Before `0eea9e597`, six accepted rewrites landed as one (§20). Six accepted, six
landed, and `hits_after` (2) equals `hits_before − len(rewritten)` (8 − 6) exactly.

⚠ **My stale literals nearly misread this.** I first checked the stored component for the 09:23 run's
sentence and got `false` everywhere — because a rebuild REGENERATES the copy, so the earlier run's
phrasing does not exist in it. §8's query is right precisely because it compares against **this run's
own** two durable fields; a literal from an earlier run is not a probe, it is a different page.

## 2026-08-22 ~19:00 BST — a fresh chassis build (`v1.0.1326`) landed; the gate survived it, checked rather than assumed

**Why this needed checking at all:** a deploy writes to `agent_definitions` (the live writer row's
`updated_at` moved to **15:09:49Z, one minute before the pods started at 15:10Z**, with no migration
behind it), and this lane's config half is exactly two keys on that row. A re-seed that reverted
either would be silent — the gate would go back to being live-but-blind (`§18`) or the repair would
go back to being thrown away (`§22`), and nothing would error.

**Both keys survived, re-read from the live row after the deploy:**
- `render_section.config.content_from` = `copy_gate.result` (migration `548`).
- the `rewrite_negations` step intact, **including its `ai_service` block** (migration `517`, the fix
  for the blind first run) — `claude-sonnet-5`, `max_tokens` 2000, `page_budget` 2,
  `output_field: copy_gate`, `next_step: render_section`.

**The binary carries the gate on both replicas** `[MEASURED 2026-08-22]`: `rewrite_negations` 8,
`copy_gate` 7, control `rewrite_negationz` **0**; and `invented_superlative` **1** with control
`invented_superlativz` **0** — so handoff item (2), "the superlative guard is not in the running
binary", is now **CLOSED**: it is in, on `v1.0.1326`.

**Fleet behaviour between the two builds is healthy** `[MEASURED 2026-08-22, runs 10:00–14:07Z]`:
12 gate runs across **6 domains** (`apis.uk`, `remortgagecalculator.uk`, `ai-agent-orchestration.com`,
`loanandmortgagecalculator.co.uk`, `webdesign.co.uk`, `leopardessconsulting.co.uk`), 5 `repaired`,
7 `clean`, and **`has_result` true on every one of the 12** — the post-548 contract holding across
sites, not just on the page that proved it. No gate run has happened yet on `v1.0.1326` (none since
14:07Z); a watch is armed for the first.

## 2026-08-22 evening — the reuse follow-up shipped, APPROVED round 1, and one objection was right

Handoff item (3), the reuse advisory from the approved round, is DONE:
`aiservice.ClassifyTruncation` (register **MDL-043**), three tests each mutation-proven, adopted at
one call site. Council `a696e2a3-311b-4490-b862-f5cdfc1bc169` — **APPROVED at round 1, all reviewers
approve, 7 abstained**.

**The audit changed the shape of the fix, which is the part worth keeping.** The advisory said
"extract the three-state and audit the other call sites". The audit found the platform ALREADY
detects truncation structurally (`aiservice.TruncatedError` / `IsTruncated`, from the provider's own
stop signal — MDL-038), and `GenerateText` returns it, so **any caller doing `if err != nil` is
already protected without arithmetic.** Shipping a numeric helper without saying that would have
invited callers to reimplement a worse detector. So it ships documented as a BACKSTOP for the only
two cases the structural signal cannot reach.

**Demand evidence that the class is real, found while answering a guardian objection**
`[MEASURED 2026-08-22]`: **3 of 70** `rewrite_negations` calls were cut at exactly
`output_tokens = max_tokens = 2000` (08-21 14:19, 08-21 14:26, 08-22 12:25). All three were caught
**structurally** and logged `success=false` — so the numeric arm has still never been the thing that
saved us, exactly as the code comment claims. That is the design working, not a gap.

### The objection that was right, and what it cost to answer

`reuse_agent`, **medium**: *"'that finding is reported to them' is not a tracked artifact (no work
item, no doc_note) — it can be lost with nothing here to show it was ever raised."* Correct, and it
is the seat's own founding failure mode (two paths, nobody unifies them). Answered by filing
**`bugs_open/366`** for `cmd/reasoningset`'s two sites, which state the decision the owning lane has
to make rather than making it here.

### The two low objections, answered with checks rather than agreement

1. *"Confirm no downstream log parser does strict field-count matching"* on the new `usage_state` zap
   field. **Checked:** nothing anywhere parses that line (grep across `.go`/`.py`/`.sh` returns only
   the file itself), and the token-pressure monitors — `fleet-step-token-pressure`,
   `council-seat-token-pressure`, the only two `scheduled_tasks` reading `llm_call_log` — read the
   **table**, not zap logs.
2. *"The behaviour-equivalence claim is asserted, not test-proven at the call site."* True, and
   `editquality` independently verified the algebra including the failed-type-assertion case. ⚠ **And
   then I changed the substitution AFTER submitting** (see below), so a reviewer re-reading it must
   read the committed version, not the plan's sketch.

### ⚠ A correction to my own change, found minutes after submitting it

Writing the MDL-043 register entry sent me to MDL-042 (another lane, same area, shipped today), which
distinguishes `options["max_tokens"]` from `options["__sent_max_tokens"]`, *"the wire number"*.
Following it up: `ai_actions.go` feeds `llm_call_log.max_tokens` from `__sent_max_tokens`, and gemini
deliberately writes the VISIBLE-text budget there so it stays commensurable with
`__usage_output_tokens` across providers (`bugs_open/110`). **So the ceiling of record is the APPLIED
one, and I was comparing against the REQUESTED one.** Equal on anthropic (our provider — verified:
all 70 rows carry `max_tokens = 2000`), but a caller that lets the client choose its own cap sets no
`max_tokens` at all, and the old read would have reported UNKNOWN over a ceiling the provider knew.
Fixed, with a fallback because ollama records no sent value. The same edit killed an **unchecked type
assertion** (`options["max_tokens"].(int)`) sitting inside an action, where a panic kills the step.

**The transferable bit: writing the register entry is what found it.** The entry forced me to say how
this relates to neighbouring mechanisms, and the neighbour contradicted me. Registering is not
paperwork after the work — on this occasion it *was* the review.

### Known limitation, recorded rather than fixed

MDL-042's escalate-on-truncation retry is wired into `execute_llm_prompt`. This action calls
`client.GenerateText` directly, so **it does not inherit escalation** — a cut repair is still one
lost repair, not a retry at a higher ceiling. That is the correct outcome for a style repair (the
copy stands as written and the marker says `repair_unavailable`), but it should be a decision, not a
surprise: 3 of 70 calls hit the ceiling.
