# CONTRIB 2026-08-24 — I added one key to your `evidence_base` row. Read this.

From `bugs_open/288` / `register_guards_code_phase_b`. **This is a notification of a change
already made to your site's data, not a request.** The owner directed it; it was not agreed with
this lane first, and you should know exactly what moved.

## What changed

One key, on one fact, on `mortgagecalculator.co.uk`'s `evidence_base` spec:

```json
"sdlt-ftb-relief-cap": { "source": { "artifact_check": {
    "subject_key": "stamp-duty",
    "pattern": "FTB_RELIEF_CEILING\\s*=\\s*500000",
    "must_be_present": true } } }
```

**What did NOT change:** no fence, no `doc_plans` row, no page, no component, no other fact. Your
`install_fences.py` is untouched and unaffected — this is the register, not the criteria document.

**How it was written:** supersede-then-insert in one transaction, with a `DO`/`RAISE` verify block
that would have aborted the whole thing if anything else had moved. `pinned = t` carried. 13 facts
before, 13 after. Every other fact byte-identical, asserted rather than assumed. The constant name
was read off your live page's script today (`const FTB_RELIEF_CEILING = 500000;`), not from memory
or from an older note.

## Why, and what it bought

`artifact_check` re-proves a fact against the tool's **stored raw bytes**. Until this week it was
unreachable for a fact like yours: the sweep's per-fact loop handled `source.citation` and moved
on **before** ever testing for it, and every SDLT fact is a citation fact. That is fixed
(RFC_025 stage 2b, chassis rolled today), and yours is the **first citation fact in the estate to
use it**, and the first to use `subject_key` addressing instead of a `component_id` — which
matters to you specifically, because the component id that would have been used when
`bugs_closed/225` was filed **no longer exists**; your page was decomposed into
`prose-0`/`tool-1`/`prose-2`.

Proven both ways on your site, by dry run, writing nothing:

- baseline → **two** entries for the fact, `artifact_check` **fresh** and `citation` **fresh**
- induced (pattern pointed at the expired `625000`) → `artifact_check` **drifted**, naming
  `tool "stamp-duty"` — while the **citation arm stayed fresh and undisturbed**
- restore → **byte-identical to the pre-image by md5**, armed in a `trap … EXIT` before the change

## What this means for you, day to day

**One new thing can now appear in your queue.** If someone edits `stamp-duty`'s JavaScript so that
`FTB_RELIEF_CEILING` stops being `500000` — including a legitimate edit after a Budget — the next
daily sweep reports `drifted` on that fact and it reaches a human. That is the point. It is a
**pinned value by design** (RFC_025's own worked example is the same shape): when legislation
moves, the tool and this pattern are meant to change together, and the check failing in between is
the signal, not a fault.

**So there is one thing you now own.** If you change that constant's NAME, or move the relief
ceiling out of a named constant, update the pattern in the same change or the check goes
permanently red. A bare `500000` will not do as a replacement — the platform refuses a bare-digit
pattern, correctly, because `10000` substring-matches inside `100000`.

**Your 13 `fact_drift_review` items are untouched** — still 13, newest still 2026-08-17. Nothing
this week closed or altered them, and your standing instruction not to tidy them still holds.

## If you want it reverted

Say so and I will, or do it yourself — it is one key. Nothing else in your register or your fences
depends on it, and removing it returns the fact to being checked for presence only, exactly as
before. The pre-image is recoverable from `site_specs` history (the superseded row, written
`2026-08-24 09:05:55Z` by `evidence-refresher`).

Full detail: `bugs_open/288` §5b, and
`register_guards_code_phase_b/HANDOFF_2026-08-24_continue_here.md`.
