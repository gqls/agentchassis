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
