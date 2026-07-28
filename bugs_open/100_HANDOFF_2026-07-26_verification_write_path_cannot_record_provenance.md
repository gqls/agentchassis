# 100 — The verification write path CANNOT record provenance: it reads the source URL from the LLM

> ## STATUS 2026-07-28 — candidate 1 COMMITTED (`2ebabf2ca`); INERT until the chassis image rolls
>
> Taken by the "bugsearch" thread, bundled with `bugs_open/101` as both files instruct.
> Council submitted: `SUBMISSION_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c`.
>
> **The fetcher was already recording the answer.** Every webscrape provider result
> carries `url` and `captured_at`, set beside the HTTP call
> (`providers/firecrawl.go`). Nothing read it. So this needed no new provenance
> mechanism and no prompt change — only for the writer to stop asking the model a
> question the fetcher had already answered. `datahelpers.ExtractFetchProvenance` is
> the reader, and it is generic across verticals (this file's candidate 3, at
> candidate 1's cost).
>
> **The three model reads are DELETED, not demoted to a fallback.** A fallback would
> have restored the old behaviour the moment a model volunteered a plausible-looking
> URL — which is the failure this file forbids, arriving later and harder to see. A
> model-supplied `source_url` is now logged as **ignored**, so the prompt drifting
> toward self-reported provenance becomes visible rather than silently taking effect.
> Same substitution at both price insert sites, which fed from the same three reads.
>
> **`scraped_data` is appended to the writer's inputs unconditionally**, not added to
> the definition's `input_fields`. Making provenance depend on every caller remembering
> a config key is precisely what produced 2,970 unsourced rows in silence.
>
> **The unrepresentable-state half is SQL `257`, written and deliberately NOT APPLIED.**
> It must follow the image: applied first, it would refuse writes the running binary
> cannot yet satisfy, turning a silent data-quality defect into a hard failure of vet
> verification. Two notes on it:
> - It is a `CHECK`, not `NOT NULL`, because **`source_type` was ALREADY `NOT NULL` and
>   never fired once** — the Go read produced an empty *string*, not a NULL. The empty
>   string is the bad value, so the empty string is what has to be refused. The
>   constraint that was already there is the reason to distrust "there is a constraint"
>   as an answer.
> - `NOT VALID`, so the 2,970 historical rows stay as they are: genuinely unsourced and
>   unpublishable, refused by the publishing rule rather than back-filled with invented
>   provenance — which would be this very bug.
>
> **Blast radius checked, not assumed:** exactly one writer in the tree inserts into
> `data_observations` (`business_intel_actions.go:370`), so the constraint cannot break
> a second path.
>
> **`[UNVERIFIED]` — the live shape of `scraped_data`.** No run carrying one survives
> (collection off since 2026-03-18; `orchestration_states` is on a retention clock), so
> the path was **traced through the code** rather than observed: adapter
> `sendSuccessResponse` → `ResponseBody.Body` → `parseResponseBody` →
> `collected_data[output_field]` = `{data:{url, captured_at}}`, i.e. **`data.url`**. The
> reader accepts six shapes and logs loudly when none matches. **The first real
> verification run settles it**, and the check is this file's own: `source_url`
> non-empty **and** `raw_data ? 'source_url'` still **false**.
>
> **Still open until:** the chassis image rolls, SQL 257 is applied after it, and one
> verification runs green against both columns.

**Filed** 2026-07-26 by session "bugfix 061" (vetcomparison workstream).
**Status** OPEN. Unowned. **Not a correctness emergency** — nothing unsourced is published; the
consequence is that a large body of held data is permanently unpublishable under our own rule.
**Blocks** `PLAN_2026-07-26_site_strength.md` P1 (restarting vet collection) and P2 (ownership).

---

## Symptom

Every row in `business_intel.data_observations` — the table whose entire job is provenance — has
an empty `source_url`, `source_type` and `source_name`. All 2,970 of them. The column exists and
is in the INSERT statement, so this reads like a bug that *sometimes* fails to populate.

It does not sometimes fail. **It cannot ever succeed.**

## Root cause

`StoreBusinessVerificationAction` takes provenance from the **LLM's own output object**:

```go
// platform/orchestration/actions/business_intel_actions.go:322-324
sourceType, _ := verResult["source_type"].(string)
sourceName, _ := verResult["source_name"].(string)
sourceURL,  _ := verResult["source_url"].(string)
```
where `verResult := extracted["verification_result"]` (line 180) — i.e. whatever the model
returned. Three independent facts make those three reads permanently empty:

1. **The prompt never asks for them.** `vet-practice-verifier`'s `extract_and_reconcile` step
   requests exactly six sections — `business`, `vet_details`, `vet_staff`, `prices`,
   `confidence_score`, `extraction_notes`. There is no source/provenance field anywhere in
   `prompt_template`.
2. **The URL that *was* fetched never reaches the writer.**
   `store_results.config.input_fields` = `["business_id","verification_result","task_id"]`.
   The scrape's output field, `scraped_data`, **is not in that list**, so the writer has no access
   to it — even though `scrape_website.config.url_field` names the URL deterministically
   (`business_record.business.website_url`).
3. **Observed, not inferred.** `raw_data` *is* `json.Marshal(verResult)`, so its keys are the
   object's keys:
   ```sql
   SELECT count(*) AS total,
          count(*) FILTER (WHERE raw_data ? 'source_url')  AS has_source_url_key,
          count(*) FILTER (WHERE raw_data ? 'source_type') AS has_source_type_key
   FROM business_intel.data_observations;
   --  total | has_source_url_key | has_source_type_key
   --   2970 |                  0 |                   0
   ```
   The keys are **absent**, not blank. The key set present across all 2,970 rows is exactly the
   prompt's six sections plus model improvisation (`opening_hours` 43, `services` 14, `branches` 5).

This is a **contract mismatch**: the writer reads three fields the producer is never asked to
emit, and the component that genuinely knows the answer is not wired to the writer.

**The same three reads feed `insertPrice` and `insertMedicinePrice`** (lines 267-268, 301-302),
so per-price `source_url` is empty by the same mechanism — which is the standing prerequisite the
RUNBOOK names for re-enabling `vet-batch-verify`, and it is currently unmeetable.

## Why the obvious fix is WRONG

The one-line fix is to add *"and tell us the source_url"* to the prompt. **Do not do this.** It
makes provenance a **model claim about its own evidence** — an assertion, generated by the same
call that generated the facts, with nothing to check it against. This site was remediated in July
precisely for AI-asserted quantitative claims (`bugs_closed/043`, `bugs_closed/061`), and vet
price/ownership data is the most consequential place on the fleet to reintroduce that surface.
Provenance must be **recorded by the component that performed the fetch**, never reported by the
model that read the result.

## Fix candidates — ordered by what closes the door

1. **Make the unsourced state unrepresentable at the write.** Thread the fetched URL from
   `scraped_data` into the writer (add `scraped_data` to `store_results.input_fields`, take the
   URL from it, and record an as-at timestamp), then make `data_observations.source_url` **NOT
   NULL with no default** — plus the same for the price inserts. An observation that cannot say
   where it came from then cannot be stored at all. This is the only candidate under which
   "somebody forgets" stops being possible.
2. **Thread the URL through, leave the column nullable.** Same Go change, no constraint. Works,
   but leaves the current failure mode reachable by any future caller that omits it — the exact
   shape that produced 2,970 empty rows in silence.
3. **Record provenance in the scrape action itself**, writing an observation at fetch time rather
   than at store time. Structurally the cleanest (the fetcher is the only thing that *knows*), and
   generic across verticals — but a larger change touching a shared action.
4. ~~Ask the LLM for `source_url`~~ — rejected above. Listed only so the next reader does not
   re-propose it.

**All of candidates 1-3 are Go changes** to `platform/`, so: council gate → image build → roll.
`input_fields` alone is config and live immediately, but is **not sufficient** — the writer must
also be taught to read it. **Bundle the `bugs_open/101` scrape-config fix into the same round**;
they touch the same step and neither is worth a roll alone.

## How to verify a fix

- Not by the status. Re-verify one practice, then:
  ```sql
  SELECT source_url, source_type, raw_data ? 'source_url' AS llm_claimed_it, collected_at
  FROM business_intel.data_observations ORDER BY collected_at DESC LIMIT 5;
  ```
  `source_url` must be non-empty **and** `llm_claimed_it` must stay **false** — if the LLM is
  supplying it, the fix is the wrong one even though the column is populated. That second column
  is the discriminating check; without it a green `source_url` proves nothing about *where* it
  came from.
- Spawned worker pods run `agent_definitions.image_tag`, **not** the deployed image — pod-grep the
  **spawned** pod for a symbol the fix *creates*, with a negative control in the same command.

## Post-roll triage 2026-07-27 (~15:55 UTC) — unchanged, and the priority is higher than the severity

Fleet rolled to **v1.0.1174** (`2026-07-27T15:11:15Z`). No fix has been written for
this bug, so the roll changes nothing. Re-measured live:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE COALESCE(source_url,'')<>'') AS has_source_url,
       count(*) FILTER (WHERE raw_data ? 'source_url')     AS llm_claimed,
       max(collected_at) AS newest
FROM business_intel.data_observations;
--  2970 | 0 | 0 | 2026-03-18 22:09:03
```

Identical to the filing. `newest = 2026-03-18` confirms the file's own note that
collection has been off since March — which is *why* the row count has not moved,
and why this cannot self-correct by waiting.

**The finding that should change how this is prioritised: 100 + 101 together gate a
whole site workstream, and neither is severe on its own.** Read individually, 100 is
"not a correctness emergency" (its own words — nothing unsourced is published) and
`101` is "low severity". Read together they are the reason `vetcomparison`'s P1
(restarting vet collection) **cannot start**: provenance is structurally
unrecordable, so every row collected under a restarted crawl would be born
unpublishable under our own rule, and the fix is a **Go change** — council gate,
image build, roll — not config. A workstream blocked on a Go window is a different
class of problem from a data-quality blemish, and the two bug files' individual
severity lines actively understate it because neither can see the other's half.

**Consequence for sequencing:** these two are the only bugs in this triage cluster
whose fix is on the critical path of another workstream. They should be bundled into
**one** council round and **one** image window (the files already say so, at
`100` §"Fix candidates" and `101` §"Fix candidates" item 2) — and that round is worth
scheduling ahead of work with higher nominal severity but no downstream blockee.

**Ownership:** `who-owns.py 100` names no owning workstream; `who-owns.py 101` names
`vetcomparison` (ACTIVE, 59 commits/14d). Since 101 is the cheaper half of the same
round and vetcomparison is the blocked party, that thread is the natural owner of
both — but it has not claimed 100, so this needs an explicit hand-off rather than an
assumption.

## Related

- `bugs_open/101` — four inert config keys on the same `scrape_website` step (same round).
- `PLAN_2026-07-26_site_strength.md` — P1's branch point; this bug decides it.
- `NOTES_vetcomparison.md` 2026-07-26 §1 and §5 — full evidence trail and the diagnosis verdict.
- Diagnosis loop returned **UNVERIFIABLE** (corr `e6580fe5-…`, ran under dispatch-loop corr
  `38394a85-…`): it corroborated both structural halves it could reach but its bundle truncated
  before the prompt, and there is no runtime trace because collection has been off since March.
  It did **not** refute the mechanism. The two gaps it named are closed by items 1 and 3 above,
  gathered independently before the verdict returned.
