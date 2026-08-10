# evidence/ — how this site's fact register is built, reproducibly

Added 2026-08-10. `PLAN_2026-08-09_facts_into_tool_acceptance.md` §1(a) recorded
the gap this closes: the first four SDLT facts were seeded straight into the
database and **no `.sql` existed in the repo**, so the register could be read but
not rebuilt or reviewed.

Order of use:

1. **`quotecheck/`** — a ~60-line Go program that fetches a citation URL with the
   same headers as `fetchCitationDocument` and runs the **real** extractor and
   matcher over it (`datahelpers.VisibleTextFromHTML`, `QuoteFoundInText`). It is
   its own module; copy `go.mod.example` to `go.mod` and point the `replace` at
   your repo root, then `go run . <url> dump` or `go run . <url> check "<quote>"…`.
   Using Go here removes the whole class of "my python extraction disagrees with
   the sweep's Go extraction", which is the day-one `citation_lost` trap in
   RUNBOOK §11.
2. **`build_facts.py`** — lifts each quote out of that dumped text **by regex,
   never retyped**, builds the fact objects, and then re-checks every quote
   through `quotecheck` before printing anything. It exits non-zero and emits
   nothing if a single quote fails.
3. **`emit_register_sql.py`** — turns those facts into the supersede-then-insert
   SQL, composing `writer_block` by the same rule as Go's `composeWriterBlock`
   so the next scheduled sweep recomposes byte-identical text. It never parses
   the live register through a typed struct (see RUNBOOK/LANDMINES: that deletes
   every citation, `writer_line`, `unit` and `staleness_days`).
4. **`SEED_2026-08-10_sdlt_facts_per_threshold.sql`** — the output that was
   actually applied on 2026-08-10, taking the register from 4 facts to 13, one
   per band edge and per rate. It ends in a `DO $$ … RAISE EXCEPTION $$` guard
   (fact count, every fact cited, ids unique, exactly one current row) because a
   verify block made of plain `SELECT`s cannot stop a `COMMIT`.

**Induce the red before trusting the green.** Ask `quotecheck` for a quote you
know is wrong (`Up to £126,000 Zero`) in the same run as the real ones: it must
report `NOTFOUND` and exit 2. Thirteen `FOUND`s prove nothing until one
`NOTFOUND` proves the check can fail.
