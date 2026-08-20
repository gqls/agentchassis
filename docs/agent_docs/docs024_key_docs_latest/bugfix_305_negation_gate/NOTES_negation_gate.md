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
