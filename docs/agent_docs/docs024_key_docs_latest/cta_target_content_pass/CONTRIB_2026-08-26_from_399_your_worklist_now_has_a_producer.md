# CONTRIB 2026-08-26 — from the `bugs_open/399` lane: the worklist this commission has been waiting for

This lane was commissioned by the owner on 2026-08-15 ("vary each page's CTA to its most appropriate
tool") and has never run. Reading the lane docs, the blocker has been scope: which fields, on which
sites, and by what evidence. The 391 lane re-cut it to "~20 fields by query, not 16 sites", and
noted the locked set grows while the pass waits.

**As of `08afad7cd` there is a producer for that query.** A write-time pass records, before persist,
every CTA whose copy names a different page than its destination — `CTA_LABEL_MISMATCH` rows in
`agent_error_log`, carrying both sides plus the page the copy actually names. Inert until the next
roll and until migration `643` applies. Query and reading obligation:
`docs024_key_docs_latest/bugfix_399_cta_label_agreement/RUNBOOK_cta_label_agreement.md`.

## The three buckets, and only one of them is yours

`[MEASURED 2026-08-26]` across 186 live mismatched pairs:

| the copy names | count | who owns it |
|---|---|---|
| exactly one other page | **13** | mechanically repointable, but see below |
| two or more pages equally | **78** | RFC_047 §10: an agent that knows the site's premise — arguably you |
| **no page on the site at all** | **95** | **you** — the copy promises something that does not exist |

I deliberately did **not** build a repoint arm for the 13: it inherits `bugs_open/248`'s clobber (a
CTA repair turned a correct `/contact.html` into a wrong link on 2026-08-24) and reaches 7%. So
**173 of 186 are a copy problem**, which is this lane's remit and nobody else's.

## ⚠ Two cautions before you reword anything

1. **Rewording to match the destination is not automatically an improvement.** If the destination
   itself was chosen badly — `bugs_open/391`'s finding, that ranking is `nav_order` with no topic
   input — then copy naming it **locks the bad pick in**, because the next resolve label-matches the
   new wording straight back. That is why `bugs_open/399`'s own "regenerate the label from the title"
   candidate was rejected. **Read `named_url` in the record**: it tells you which page the copy
   already names, which is usually the better destination to move the *link* to, if that is
   available to you.
2. **The 391 lane measured that an unbounded rewrite destroys authored prose** — 2 of 12 pages (17%)
   lost content blocks to a `content_rewrite` whose spec said "reword ONLY the labels". Prompt text
   did not bound the write. Scope your pass at the field, and verify at the served bytes.

## Sequencing

The 391 lane's step 4 (re-resolving ~60 label-less fields) writes new destinations under old copy
and will generate a burst of exactly these records. Coordinate with them before treating a spike as
your backlog — the ordering they published (retirement → re-resolve → verify) still holds.
