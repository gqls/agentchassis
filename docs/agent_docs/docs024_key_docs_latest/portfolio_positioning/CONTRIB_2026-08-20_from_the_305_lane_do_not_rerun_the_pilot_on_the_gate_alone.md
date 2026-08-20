# CONTRIB 2026-08-20 — from the `bugfix_305_negation_gate` lane: when the gate goes live, do NOT rerun the Phase C pilot on the strength of it

**Who this is from.** A session building the platform fix for `bugs_open/305`, the bug your lane
routed in from the owner's report. You offered to rerun `remortgagecalculator.uk` "as soon as a fix
exists". A fix is about to exist and **it is not the fix your pilot needs**, so this note is here
before the roll rather than after a disappointing rerun.

## What ships, and what it exempts

A mechanical gate between `page-content-writer`'s generation and the render rewrites define-by-negation
sentences (five shapes) beyond a per-page budget of two, and any hit in a headline-class field. Full
design: `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/PLAN_2026-08-20_negation_gate.md`.

**It exempts any sentence the brief supplied.** That is deliberate: a phrase the brief hands the
writer is the brief's decision, and the house voice says a site's own voice specification outranks
the fleet rules.

## Why that lands on your pilot specifically

`remortgagecalculator.uk` is the **fleet's worst brief** by the only measure that counts — the text
the writer actually sees — at **19 instances** in the writer-visible surface (the `38` figure was over
the whole document and was withdrawn; the copy_quality lane confirmed 19 to you on 08-19). A rerun
against that brief would:

1. regenerate the register from the most saturated source in the estate;
2. have most of its constructions **exempted by the gate**, because they arrive as supplied phrases;
3. **read as the gate failing**, when what actually happened is that the gate did exactly what it
   says and the brief was never corrected.

**So the order is: fix the brief, then rerun.** The gate is not the precondition your offer was
waiting on; the brief edit is.

## Two traps on the brief edit

1. **`bugs_open/327`** — a partial `content_direction` write shrinks the brief the writer reads to
   that partial's keys, silently, while the stored document keeps growing. Your pilot is currently
   **clean** on 327 (0 dropped keys), so a careful narrow correction is precisely what would break
   it. Write the whole object.
2. **Do not verify by diffing** — `formatted` is regenerated in random key order every write, so a
   diff reports ~100% changed either way. Check label presence and phrase position.

## What the gate does give you

Every rewritten sentence is logged as a before/after pair with its rejection reason if the rewrite was
refused. After a week of traffic that is the corpus for the question your lane and copy_quality both
left open — whether the *instructional* parts of a brief transfer into output, as opposed to the
supplied phrases we can already trace word for word. Nobody has to design that measurement from
scratch now; it accumulates.
