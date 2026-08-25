# 395 — an item closes `complete` while its own stated acceptance test is unmet, and now there is a machine verdict saying so

**Filed:** 2026-08-25 by the `vigilant_designer_offer_analysis` lane.
**Status:** OPEN. Reproducible on demand; one instance caught by the platform's own machinery.
**Not a regression** — nothing on the completion path has ever read an acceptance test.

> **⚠ ON THE 2026-07-31 OWNER RULING (a `bugs_open` file asserting a cross-cutting cause is not
> "filed" until it has been through the `090` diagnosis loop, or the filing session states plainly
> why it substituted equivalent first-hand verification).** Substituted here, and this is the
> statement: I did not infer this mechanism, I **watched it happen on a run I fired myself** and then
> got a machine verdict on the artefact. I read the whole path first-hand
> (`write_audit_findings_action.go` → `site_work_items` → `page-build-handler` → `complete`), the
> item's own predicate was evaluated by the platform's exported evaluator rather than by my reading,
> and the served page was fetched. What is NOT established is the blast radius beyond
> `audit_source='offer-analysis'` — see §5, which is a census nobody has run and is the obvious first
> job for whoever picks this up.

## 1. The symptom, in one worked case

`site_work_items` row (webdesign.co.uk, `content_rewrite`, `audit_source='offer-analysis'`):

| | |
|---|---|
| created | 2026-08-24 22:08:38Z |
| its own `spec.acceptance_test` | *"The index page meta description mentions both the tool-article pairing and the no-account promise, in that order, before any catalogue count."* |
| handler | `page-build-handler` |
| closed | **`complete`**, 22:25:49Z, `result.response.commit_sha = ee88ba3c…`, with a `deploy_result` |
| page rebuilt again | 2026-08-25 11:23:13Z |
| the served `<meta name="description">` NOW | `Sixty-three browser tools for web design and development. No account, no upload, everything runs in your browser.` |

The criterion is not met: there is no mention of the tool-article pairing, and the catalogue count is
the first thing in the string. **The page was rebuilt twice and deployed; the item is closed;
nothing noticed.**

## 2. Why this instance is different from every previous one — the verdict is MECHANICAL

Until 2026-08-24 an acceptance test was free prose and this class could only be found by a human
reading a page. That item carries, in the same spec, a structured predicate the platform emitted
alongside the prose (`CLM-024`, live since 22:08Z):

```json
{"type": "text_order", "page": "index", "field": "meta_description",
 "before": ["paired", "pairing", "article", "guide"], "after": ["$cardinal"],
 "verdict_at_emission": "refutes",
 "evidence_at_emission": "meta_description of \"index\" states none of \"paired\", \"pairing\", \"article\", \"guide\", so nothing can precede \"$cardinal\""}
```

Feeding that predicate and the CURRENT served string to the platform's own exported evaluator
(`actions.EvaluateAcceptancePredicate`) returns **`refutes`** — same verdict as at emission, after a
rebuild, a deploy and a second rebuild. Pinned by
`TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix`.

So this is not "a reviewer thinks the page still reads badly". It is a machine saying the stated
criterion is false, about the exact field the criterion names.

## 3. Root cause, as narrowly as the evidence supports

**`complete` means "the handler reported success", and nothing on that path reads
`spec.acceptance_test`.** `page-build-handler` rebuilt and deployed the page — its own job, done,
truthfully reported. The item's criterion was never an input to the completion decision, so a rebuild
that changes something other than what the test demanded closes the item exactly as a correct one
would.

`complete_work_item_no_change.go`'s own comment states the gap from the other side:

> *"grading the item's own stated acceptance_test is a separate and larger job (that field is free
> LLM prose; 10 of 15 live values name a computed property and 2 contain clauses no probe can assess,
> so it needs a producer-side contract change first)"*

**That producer-side change now exists for one producer** (`CLM-024`), which is what makes this
filable rather than merely true.

⚠ **This is NOT `bugs_open/213` / WII-017 (`noChangeGates`).** That gate refuses a completion whose
handler reports it changed **nothing**. Here the handler changed something real — a rebuild and a
deploy, with a commit sha — and `noChangeGates` is right not to fire. The two are complements: 213
catches "the handler did nothing"; this is "the handler did something else".

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Evaluate the item's own predicate at completion, and refuse a `complete` that still refutes** —
   beside `handlerReportedFailure` / the `noChangeGates` roster, which is the only place that can see
   the handler's report (`verifyBeforeComplete`'s `VerifyTarget` carries the SPEC, not the RESULT, so
   a verifier would grade the row's previous value; 213's comment records why). Opt-in per item_type
   with the unsafe default OFF, per the 2026-08-02 ruling, because a refused completion is a live
   behaviour change on handlers several lanes own. The evaluator is already exported and needs no
   browser, no HTTP probe and no page fetch on the completion path — the standing objection to every
   other option here.
   ⚠ **The honest limit of this candidate:** it can only refuse where a predicate EXISTS. On the
   first live run that was 3 findings of 4, and only for one producer.
2. **Route a refuted completion to `needs_human_review` instead of refusing it.** Cheaper politically,
   keeps the work visible, and does not block a handler that did its job. Weaker: the queue is where
   items go to wait.
3. **Record only — write a `doc_note` / work item when a closed item's predicate still refutes,
   and change no completion semantics.** This is the "make it visible first" option and is the
   smallest thing that stops the estate learning nothing. It is also the one that risks becoming
   permanent.
4. ~~Have the handler read the acceptance test~~ — rejected: `page-build-handler` serves many
   producers and free prose is not a handler input. That is the design this bug exists because of.

## 5. What is NOT measured, and it is the first job

**The blast radius.** Everything above rests on ONE item, from ONE producer, on ONE site. Nobody has
asked how often a `complete` item's stated criterion is unmet across the estate, because until
2026-08-24 nothing could ask it mechanically. The census that would size this:

```sql
-- items that carry a machine-checkable predicate AND are closed
SELECT s.domain, wi.item_type, wi.status,
       wi.spec->'acceptance_predicate'->>'type'  AS pred,
       wi.spec->>'page_name'                     AS page
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE wi.spec ? 'acceptance_predicate'
   AND wi.status IN ('complete','wont_fix')
 ORDER BY wi.created_at DESC;
```

⚠ **As of 2026-08-25 that returns 3 rows, all from one run, because the producer is one day old** —
so the census is a *plan*, not a finding, and it grows only as `offer-analyser` runs. Do not quote a
small number from it as a low rate. The prose-only population (**37** acceptance tests as of
2026-08-24, of which **3** were found refuted by hand) is the honest current estimate and it was
assembled by reading, not querying.

## 6. How to verify a fix

- **Positive:** re-run the evaluator over every closed item carrying a predicate; a completion gate
  must have refused (or flagged) the webdesign `index` case above.
- **NEGATIVE CONTROL, and it is not optional:** an item whose predicate is genuinely satisfied after
  its fix must still close `complete`. There is **no such row today** — the three live predicates all
  refute — so a fix cannot be called proven until one exists. A gate that has only ever seen failures
  is indistinguishable from one that refuses everything.
- ⚠ **Do not verify at the item's status alone.** Read the served page: this whole class exists
  because a status agreed with a handler instead of with an artefact.

## 7. Pointers

- `CLM-024` (the predicate producer + the exported evaluator) and `CLM-023` in
  `docs/agent_docs/docs026_concept_register/register/claims-verification.md`
- `platform/orchestration/actions/verify_acceptance_predicates_action.go` — `EvaluateAcceptancePredicate`
- `platform/orchestration/actions/complete_work_item_no_change.go` — the completion seam and why a
  verifier cannot do this
- `features_open/030` §10 v2(d) and the lane at
  `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/`
  (cold-start `HANDOFF_2026-08-25_continue_here.md`)
- `bugs_open/213` / WII-017 — the complement, not the same bug

---

## 8. WHAT SHIPPED — candidate 1, at RECORD-ONLY (2026-08-25, same lane)

**Committed `69479bcf6`, council `064841bd-58fc-46a1-a77d-6b0a6309d0ba` APPROVED round 1 (14 seats,
5 advisory objections, none high). Go, so INERT until the next fleet roll.** Register **WII-033**.

⚠ **Do not read "committed" as "running".** Until the chassis carries it, no completion is graded,
and an empty `result->'_verification'->'acceptance_predicate'` means *"this was never switched on"*,
not *"nothing refutes"*.

**Completion gate 1c**, in `complete_work_item_acceptance_predicate.go`, between gate 1b and gate 2.
Opt-in per `item_type`; three-valued (`predicateUndeclared` / `predicateRecords` / `predicateRefuses`,
the zero value refused by the roster test). `content_rewrite` is the only entry and it **RECORDS**.

### 8a. It records rather than refuses, for the reason §6 gives

**There is still no live negative control** — §6's requirement is unmet and shipping did not change
that. The permit arm is proven as a UNIT (`TestAPredicateThatHoldsPermitsTheCompletion`, a real live
predicate against a string edited to satisfy it) and **not in the wild**. The two must not be quoted
for one another. `outcome='permitted'` appearing on a real row after the roll IS the control §6 asks
for, and until then the refusal arm stays unproven.

Every evaluated predicate now leaves a verdict on the row **including `holds`** — without that, a
gate that permits is indistinguishable from a gate that never ran, which is this entry's own residual.

**Promotion to `predicateRefuses` is a BUILD FAILURE** until `content_rewrite` joins
`livespec.ClaimedItemTimeoutExclusions` and its migration ships:
`TestClaimTimeoutExclusionCoversBothCompletionGates` now counts a gate-1c entry *only when it
refuses*. The half no test can assert is in the entry's own `PromotionOwes` field, which the roster
test requires on every recording entry.

### 8b. ⚠ §4's REASON for candidate 1's placement was WRONG, and the code records the correction

§4 says a verifier cannot do this because *"`verifyBeforeComplete`'s `VerifyTarget` carries the SPEC,
not the RESULT"*. **That is gate 1b's argument and it does not transfer** — gate 1b needs the
handler's reply; gate 1c needs the spec and the current page row, **both of which a verifier has**.

The candidate is still right; the reasons are different. `GetVerifier` is ONE verifier per
`item_type` — a scarce shared slot on a type many producers file into — and these gates compose.
`[MEASURED 2026-08-25]` no verifier is registered for `content_rewrite` (13 `RegisterVerifier` calls,
none naming it), so gate 2 is inert for it either way.

### 8c. ⚠ THE TRAP THAT WOULD HAVE SHIPPED THIS PERMANENTLY BLIND

**A STORED predicate cannot be fed to `EvaluateAcceptancePredicate`.** The evaluator enforces a
closed key set per type; the emit gate stamps `verdict_at_emission` and `evidence_at_emission`
**after** evaluating. So the stored shape carries two keys the evaluator refuses and **every live
predicate returns `inapplicable`** — a legitimate verdict whose message names a KEY, which reads as a
fault in the model's output rather than in the reader.

**Nothing in the suite caught it.** `TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix` — §2's
own evidence, and the only test over real live predicates — hand-writes them WITHOUT those keys, so
it exercises a shape the database does not contain. The stamp and the strip are now single-sourced
(`storedPredicate` / `predicateForEvaluation`) and pinned by `TestStampAndStripAreInverses`, which
fails when a third provenance key is added to one and not the other. Full entry in `LANDMINES.md`.

### 8d. §5's blast-radius census is now COLLECTIBLE, and here is the corrected denominator

§5's query is still the plan. What shipped adds the other half — a per-completion verdict on the row:

```sql
SELECT s.domain, wi.item_type,
       wi.result->'_verification'->'acceptance_predicate'->>'outcome' AS outcome,
       wi.result->'_verification'->'acceptance_predicate'->>'verdict' AS verdict_now,
       wi.result->'_verification'->'acceptance_predicate'->>'verdict_at_emission' AS at_emission
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE wi.result->'_verification' ? 'acceptance_predicate'
 ORDER BY wi.completed_at DESC;
```

⚠ **`outcome='permitted'` is the row §6 says a fix is not proven without.** A census of nothing but
`recorded_only` means the gate has still only ever seen failures.

`[MEASURED 2026-08-25, live UNION archive]` the honest denominator for `content_rewrite`:
**1,638 completions all-history, of which exactly ONE carries an `acceptance_predicate`** — the
worked case in §1. So the gate is inert on 1,637 of 1,638 and the census grows only as the producer
runs.

### 8e. ⚠ THE COVERAGE LIMIT, measured because the council's guardian seat asked

Gate 1c is in `CompleteWorkItemAction` only. The seat asked whether a LOOP's own `mark_complete`
bypasses it — a good question, since the landmine on `build-dispatch-loop.process_item.mark_complete`
records it REPLACING `site_work_items.result` with spawn bookkeeping.

**It does not bypass it.** That step declares `"action": "complete_work_item"`, i.e. this path.
`[MEASURED 2026-08-25, live UNION archive]` of 1,638 `content_rewrite` completions, **1,600 carry
`handled_by='build-dispatch-loop'`**, including §1's worked case.

⚠ **The tail is real and is 38 rows (2.3%), spread 2026-03-10 → 2026-08-23, with `handled_by` NULL.**
`CompleteWorkItemAction` always stamps `handled_by`, so those were written by something that is
neither completion action — the claimed-item-timeout sweep is the obvious candidate. None carries a
`_verification` key and none has ever carried a predicate, so nothing is lost today. **That 2.3% is
exactly what starts to matter on promotion to refusing**, and it is why the exclusion is in
`PromotionOwes`.

### 8f. What is still open on this bug

1. **The negative control (§6).** Unmet. The gate cannot be called proven until a real row completes
   with `outcome='permitted'`.
2. **The refusal arm has never fired.** Units only — CLM-023's residual in a third place.
3. **The blast radius (§5).** Still a plan, and it grows only as the producer runs.
4. **This does not touch WHY the handler produces content that fails the predicate.** The council's
   `constitution` seat flagged it and is right: gate 1c is detection, not the root-cause fix to the
   content-generation gap. Recorded so it is not mistaken for one.
5. **The accumulation.** Four hand-wired gates on one function — `architecture_review/RFC_055`.

---

## 9. ⚠ CORRECTION 2026-08-25 — THIS IS A RECURRENCE OF `bugs_open/320` §5, AND THE CRITERION WAS NEVER SATISFIABLE

Found by the `bugs_open/395` lane taking §8f item 4 (the untouched root cause) and asking the one
question this file never did: **who WRITES the field the predicate grades?** Verified first-hand by
this lane before acceptance.

**All three live predicates grade `pages.meta_description`, and a NON-EMPTY meta description is
structurally immutable on the route these findings take** `[MEASURED 2026-08-25]`:

| writer | reachable how | can it overwrite? |
|---|---|---|
| `page-build-handler` — **where these items are routed** | — | **NO STEP TOUCHES THE COLUMN AT ALL** |
| `save_page_meta_description_action.go:211` (`SET meta_description = $2`) | ONE live agent, `meta-description-backfiller`, whose pre_query selects `COALESCE(p.meta_description,'')=''` | empty values only |
| `site_db_actions.go:1235` | page upsert | `COALESCE(NULLIF(EXCLUDED,''), existing)` — fills a blank, never overwrites |
| `apply_adoption_plan_action.go:84` | adoption only | same guard |

**So the handler could not have satisfied the criterion whatever it was told.** §6's negative control
is not merely missing — it is **UNREACHABLE** for the current predicate vocabulary against the current
routing, and §8f item 1 was wrong to describe it as a row we were waiting for.

### 9a. It was already on record, six days earlier

**`bugs_open/320` §5, 2026-08-19**, measured **the same page** (webdesign.co.uk `index`, item
`13522562-…`, completed 2026-08-15 19:15Z, column 0 chars afterwards) and concludes: *"filing
`content_rewrite` items is not a workaround — it is the expensive version of doing nothing, and it
leaves a green record saying the opposite."* Its §9 says plainly: ***"Do not file `content_rewrite`
items for them."***

**This lane's producer filed precisely that item on 2026-08-24.** Not through carelessness — §9 is
prose in a bug file and nothing enforces prose. 320 §4 also already carried the check that would have
caught it: ***"ask who owns the field first."***

⚠ **And "grep before you file" did not save us:** the grep run before filing 395 was for the
MECHANISM (completion gating, acceptance tests, verifiers) and correctly returned nothing, because
320 is filed under *meta descriptions are never asked for*. **Grep for the mechanism AND for the
column.** Logged in `WRONG_CALLS.md`.

### 9b. What does NOT change

**Gate 1c stands, and this strengthens the case for it.** 320 §9's prose did not stop the recurrence;
the gate is what caught it, mechanically, with a machine verdict. That is the argument for
completion-time grading rather than against it.

### 9c. What DOES change

- **`PromotionOwes` is rewritten** (commit `864e73d8a`) to say the control is unreachable, with the
  four writers as evidence and the two routes that could make it reachable. Both superseded wordings
  are kept in the field, because the sequence is the record: the question had a worse answer each time
  it was asked.
- **§8f item 1 is superseded by this section.** Do not go looking for the `permitted` row until the
  vocabulary covers a writable field, or such findings are routed at a writer that can overwrite.
- **Routing them at a writer that CAN overwrite is an OWNER DECISION and neither lane may take it:**
  `bugs_open/320` §15 records the owner granting `overwrite_existing: true` for a one-off 681-page
  regeneration and **explicitly withholding it for the standing mechanism**. A standing path that
  rewrites published copy in response to an automated finding is exactly the authority he withheld.
- **The producer-side guard** — refuse to FILE a finding whose criterion grades a field the target
  handler cannot write — is being built by the `bugs_open/395` lane, and makes 320 §9 mechanical
  instead of prose. This lane will take the predicate-gate half if that split is agreed, so a
  predicate over an unwritable field is refused at source too.
