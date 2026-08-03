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

## 2026-08-03, late evening — fix committed, council submitted, and one same-file passenger

- Fix commit: `c3b02f035` (3 files, pathspec, `Council-Submitted: da1d3a40-8ec1-4ea1-9c65-8c585ec2d013`).
  Three mutations run pre-commit, each caught by its intended test. Live-index proof of all
  four predicate shapes: fixed primary 1 row, absent-symbol 0, moved-file primary 0 with
  name-only fallback 1, line-ref path-only 51.
- Lane docs: `58eeef91d`. Register/016b/WRONG_CALLS/181-disposition: `d4fc4d663`.
- **[OBSERVED, the CLAUDE.md same-file passenger, from the other side]** my `LANDMINES.md`
  append was committed by ANOTHER session's `48fa0ac3a` ("scratch: move session
  scratchpads…") between my append and my own docs commit — their commit touched
  LANDMINES.md, so whoever committed first took both edits. Nothing lost; the entry is in
  HEAD under their message. This is the documented behaviour, experienced rather than read.
- Verifier dispatched against the new entry ON THE OLD BINARY — deliberately: that verdict
  is the BEFORE arm of the A/B proof. Its `answerCodeCheck` symbol check exists at the
  indexed commit; `parseSymbolQuery`/`symbolClauseFor` do not (added today), so after the
  roll those two must STILL return 0 from the index — which is the landmine's own point.

## 2026-08-03, ~21:30Z — LIVE on v1.0.1245, both replicas, three probes

- **The roll that beat me to it:** the 178 lane built+rolled `v1.0.1244` at 21:15-16Z from a
  HEAD cut minutes BEFORE `c3b02f035` — pod-grep showed the added strings ABSENT with the
  positive control PRESENT, i.e. an image that genuinely predates the fix, not a bad grep.
  Bumped to `v1.0.1245` (`015ffacb0`), built from committed HEAD, verified the LOCAL image
  first (docker run + strings: 1/1/1), pushed, deployed.
- **Pod proof, both replicas (`76d455b95d-ftswd`, `-hccrc`, started 21:26Z):**
  `names a LINE, not a symbol` = 1 (added), `the NAME alone matches` = 1 (added),
  `searched the names of` = 1 (positive control), `SplitSymbol_163_shouldnotexist` = 0
  (negative control). 1/1/1/0 on both.
- **The 21:16 roll KILLED five in-flight council review steps — including this lane's own
  submission** (`da1d3a40`, died mid `review_prior_art`; all five rows' `updated_at` predate
  the pod restarts). Resubmitting with `RESUBMIT_CORR` so the trail accumulates; the fix
  commit already carries `Council-Submitted:` and 098 resolves at report time.
- Dispatch discipline: nothing fired at the cluster until ≥300s after the 21:26 pod starts.

## 2026-08-03, 23:00Z — END-TO-END PROVEN, and a correction to my own damage figure

**The A/B proof, on the exact entry whose failed verification produced 163:**

| date | binary | verdict on `LANDMINES.md#readsymbolbody-s-whole-file-branch…` |
|---|---|---|
| 07-31 16:01Z | pre-fix | NEEDS_HUMAN_REVIEW — "`ReadSymbolBody`, `findFile`, `namedScope`, `nextScope`, and `NextScope` returned 0 rows — most likely because the code index predates the bugfix_145 commit" (FALSE cause) |
| 08-03 22:56Z | **v1.0.1245** | **STILL_VALID — "All footprint files and symbols (`ReadSymbolBody`, `findFile`, `scopeFromCodeResults`, `namedScope`, `nextScope`, `route.scope`) confirmed present at their cited locations; the code index predates the bugfix but confirms the pre-fix structure the entry describes"** |

Same entry, same index commit (`d98010e8b` — deliberately, per the RUNBOOK landmine), same
footprint text. The only changed variable is the binary. And the new verdict handles the
stale-index fact HONESTLY — names it, scopes it, invents nothing.

**> CORRECTED 2026-08-03, same session: my "20 verdicts, 16 NEEDS_HUMAN_REVIEW, 0 CONFIRMED"
> figure above is part-wrong.** "CONFIRMED" is not in the verdict vocabulary
(`STILL_VALID|STALE|NEEDS_HUMAN_REVIEW`), so my `LIKE '%CONFIRMED%'` filter counted a string
that cannot occur — a zero that could never have been non-zero. Pre-fix truth: **21 verdicts,
16 NEEDS_HUMAN_REVIEW, 4 STILL_VALID (all via content/ls checks), 1 STALE.** The structural
claim is unchanged — no path-bearing SYMBOL check ever returned a row (23/23, plus today's
live predicate test). Caught by reading the first post-fix verdict instead of the counter;
logged in WRONG_CALLS. The register banner (DOC-069) is corrected in place.
