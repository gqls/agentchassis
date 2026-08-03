# NOTES — bugfix 163, the symbol lookup's path blindness

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-03, opening — picking the bug, and the ownership call I nearly got wrong

Swept `bugs_open/` for something unowned. `scripts/who-owns.py` returns "OWNED or recently
active" for nearly everything, because it counts *mentions*, and every session that runs
`ls bugs_open/` picks up all 60 numbers at once. So I added a second instrument: grep the
session `.jsonl` transcripts modified in the last five hours and look at **concentration** —
a lane actually working a bug mentions it 90–250 times; directory-listing noise is 2–10.

That put 163 in the clear (peak 6 in any session). Then I found this, in
`bugfix_097_content_data_links/NOTES_content_data_links.md:21`, dated 2026-08-02:

```
landmine-verifier    112 in 693556a1  -> bugs_open/163 IS being worked; dropped it
```

**[MISSTEP AVOIDED — and it is the generalisable one.]** I nearly dropped the bug on that
note. The check that saved it: grep the same transcript for the **fix site's symbol**, not
the mechanism's name.

| probe in session `693556a1` | count |
|---|---|
| `landmine-verifier` | 300 |
| `symbolTokenClause` | **0** |
| `bugs_open/163` | 6 (all from a bugs-file staleness sweep iterating every file) |

That session was doing landmine **corpus** work. It never once looked at the lookup. The
note's own stated rule was right — *"the signal is the SYMBOL"* — but it was applied to the
mechanism's name, which every adjacent lane says, rather than to the function only a fixer
would open. 163 then sat unowned for two more days on the strength of it. Logged in
`WRONG_CALLS.md`.

## Validating that the bug is still real

`orchestration_states` could not confirm the bug's own headline figure — retention has rolled
it (32 checks / 5 runs / **0** of kind `symbol`, vs the filed "23 of 23"). So I built the
claim on something retention cannot touch, running the predicate the Go actually constructs:

```sql
SELECT 'path-bearing (as executed today)', count(*) FROM code_symbols
 WHERE symbol ILIKE '%internal%' AND symbol ILIKE '%analysis%' AND symbol ILIKE '%symbolbody%'
   AND symbol ILIKE '%go%' AND symbol ILIKE '%ReadSymbolBody%'      --  0
UNION ALL SELECT 'bare symbol name', count(*) FROM code_symbols
 WHERE symbol ILIKE '%ReadSymbolBody%'                              --  1
UNION ALL SELECT 'correct: path col + symbol col', count(*) FROM code_symbols
 WHERE path ILIKE '%internal/analysis/symbolbody.go%'
   AND symbol ILIKE '%ReadSymbolBody%';                             --  1
```

And the structural version, which is the one that makes it not-a-tuning-problem:

```sql
SELECT count(*) AS n, count(*) FILTER (WHERE symbol LIKE '%/%') AS symbol_has_slash
FROM code_symbols;                                    -- 4992 | 0
```

**Damage, re-measured and worse than filed:** `doc_notes` `categories ? 'landmine-verification'`
→ **20 verdicts, 16 NEEDS_HUMAN_REVIEW, 0 CONFIRMED** (07-31 → 08-03). The bug recorded 5/6
at filing. Nothing has ever been mechanically confirmed by this verifier.

## What the exploration changed about the fix

- **A third query convention exists that nobody modelled.** `LANDMINES.md` has doubled to 211
  entries; 0 of 198 use the colon form the prompt asks for, 48 use the paren form, and **12
  `.go:` footprints are LINE NUMBERS** (`spawn_actions.go:3066`, `run_checks_action.go:773-774`).
  A naive last-colon split sends `2730` to the `symbol` column — the same bug in a new costume.
  The fix has to recognise a line reference and degrade to a path check.
- **The "invented cause" was a generalisation, not a fabrication, and that matters for the
  remedy.** The run-level staleness banner (`:148-176`, fires above 48h; index pinned at
  `d98010e8b`/07-28) legitimately warns *"a 'no matches' answer may mean NOT YET INDEXED"*.
  The model applied that run-level caveat to symbols that **predate** the indexed commit. The
  banner is correct and stays. What was missing is a **per-check fact that can override it** —
  which is why narrating the predicate is part of the fix and not decoration.
- **My fix direction was already written down twice** — 016b §9 `:398-401` and 163 `:185-190`.
  Cite it, do not re-derive it.
