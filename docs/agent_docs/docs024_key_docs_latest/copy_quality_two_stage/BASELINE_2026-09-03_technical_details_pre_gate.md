# BASELINE (pre-gate) — finetuning.uk `technical-details`, captured 2026-09-03 10:12Z

Pinned by the `copy_quality_two_stage` lane because the copy gate **mutates `content_data` in
place** (`negation_content.go`: *"a repair must mutate in place"*), so this state is destroyed
by the rebuild and cannot be re-derived afterwards. This is the BEFORE half of the before/after
pair the 443 repair canary owes this lane; the finetuning lane queued the rebuild as work item
`896bb245` (page-build-handler, content rebuild, NOT a rerender).

Measured on `page_components.content_data` — the field the gate rewrites — not on rendered HTML.
Shape regexes are the live Go vocabulary in `platform/orchestration/datahelpers/negationtells.go`.

| shape | hits in content_data |
|---|---|
| `x_not_y` | **2** |
| `rather_than` | **5** |
| `so_consequence(v3 cand)` | **1** |

**Total: 8** across 2 components.

## The sentences, verbatim (this is what the after-pass must be diffed against)

- **`faq`** · `x_not_y`
  > The point of it is for you to judge whether it sounds like your business, not for us to sell you anything further.", "question": "What actually happens in the playground hour?"}, {"answer": "You do.
- **`generic-text-block`** · `rather_than`
  > {"content": " The base model is a small open-weight model, meaning the company that built it has published the underlying weights for anyone to use rather than locking the model behind an API.
- **`generic-text-block`** · `rather_than`
  > If the plan is to put one of these models to work inside a system where several AI agents pass tasks between each other rather than a single model answering on its own, the licence is only one part of the decision.
- **`generic-text-block`** · `rather_than`
  > {"content": " The base model is a small open-weight model, meaning the company that built it has published the underlying weights for anyone to use rather than locking the model behind an API.
- **`generic-text-block`** · `so_consequence(v3 cand)`
  > If you're weighing up a multi-agent build, our agent architecture complexity estimator walks through agent count, how they're connected, and what compliance and integration work is likely to add, so you can see where the engineering effort tends to get underestimated before you commit to a build decision.
- **`generic-text-block`** · `rather_than`
  > {"content": " The base model is a small open-weight model, meaning the company that built it has published the underlying weights for anyone to use rather than locking the model behind an API.
- **`generic-text-block`** · `x_not_y`
  > When the question is architecture, not licensing Everything above applies to a single fine-tuned model doing one job.
- **`generic-text-block`** · `rather_than`
  > Like any estimate, it can be wrong, so treat it as a starting point for a conversation rather than a final number.

## ⚠ Compare like with like — this page has TWO surfaces and they disagree

The same page measured on its **served HTML** the same morning gave `x_not_y` 2 · `rather_than` **7**;
`content_data` above gives `x_not_y` 2 · `rather_than` **5**. Neither is wrong. The served page
carries shared chrome, nav and site-level content that no `page_components.content_data` row on
this page owns, and sentence boundaries fall differently once the template has wrapped the copy.

**The after-pass must be run against `content_data`, with this file's regexes, or the comparison
is meaningless** — a drop from 7 (HTML, before) to 5 (content_data, after) would look like a 29%
repair and could equally be no repair at all. The gate rewrites `content_data`; measure it there.
A served-HTML pass is still worth running afterwards as a *separate* before/after (its own before
is recorded in `NOTES_two_stage_copy.md`, 2026-09-03 midday entry), but never mixed with this one.

**Also expect the total to MOVE UP as well as down.** The rebuild re-writes the copy, and the
model prior produces `rather than` freely — the register's own note says so, and the repealed
mild-set comment measured it at 71% of all gate rewrites. So the honest post-gate question is not
"is it zero" but "did the gate repair what the writer produced *this* time", which the
`rewrite_negations` rows in `llm_call_log` for the rebuild's correlation answer directly:

```sql
SELECT step_name, created_at, success FROM llm_call_log
 WHERE correlation_id = '<the rebuild correlation>' AND step_name LIKE '%rewrite_negations%';
```

A rebuild with **zero** `rewrite_negations` rows and a non-zero count above means the gate did not
run on those sections — check `llm_field_specs` before concluding the repair failed
(`LANDMINES.md`, "a component field whose `source` is not `llm`…"). Verified 2026-09-03 for this
page: every component carries `llm` fields, so that path is clear and the model WILL be called.
