# HANDOFF 2026-08-25 — continue here

**Lane:** `register_guards_code_phase_b` (`bugs_open/288`, the class behind `bugs_closed/225`).
**Supersedes `HANDOFF_2026-08-24_continue_here.md`.**

**State: the mechanism is LIVE, has RUN FOR REAL, and WORKS. Four phases built, council-approved,
proven at the artefact. The first production sweep filed five binding suggestions across three
sites with zero errors — and found one defect of mine, which is fixed but NOT YET ROLLED.**

## The one thing that is owed before anything else

**`bba8a892d` is committed and not in any running binary.** It fixes a real defect the first
production sweep exposed: the suggester proposed **two fact ids for one constant** on agritec.uk
(two register facts both asserting £100,000). Until it rolls, any new
`fact_binding_suggested` note on a site whose register has duplicate values will carry an
unreconcilable binding in its paste-ready fragment. **Nobody should act on such a note until then.**
Verify after the roll:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in "AMBIGUOUS, NOT PROPOSED" "cannot tell which one the tool uses" \
         stale_attestation ZZZ_must_be_absent; do
  printf '%-36s ' "$s"; kubectl -n ai-persona-system exec "$POD" -- grep -ac "$s" /proc/1/exe
done   # first two >0, third =5 (control), fourth =0 (control)
```

## Open with another lane (agritec_uk) — reviewed 2026-08-25, three findings returned

They built `tool-sfi26-revenue-stacker` as a real tool page with a tool-level component (which is
what makes both the suggester's population predicate AND `subject_key` addressing reach it), wrote
four contextual `artifact_check` entries, and de-duplicated the pair the suggester exposed
(105 → 104 facts). Reviewed against their live rows:

1. **The fence is NOT live.** Zero `artifact_check` on every version of their register, so it never
   landed rather than landing and being clobbered — their migration is written, not applied. Told.
2. **Two of four patterns under-anchored**: `rate:224` matches `rate:2240`, `rate:129` matches
   `rate:1290`; the other two use `\b` correctly. **Their mutation test could not have caught it**
   — they mutated the PATTERN, where the failure mode is a mutation of the ARTEFACT. Their register
   already carries four-digit rates (1920, 1072), so it is not hypothetical.
3. `ATT-sfi26-CHRW2` carries **no value and no unit** where every sibling has both — one of the two
   figures their tool had wrong.

**Their nine value-sharing pairs changed this lane's documentation** (see §5c and the register
landmine): on a rate-table site shared values are normal, so the suggester is quiet by
construction there. Answered their question precisely: `artifact_check` produces `facts` entries,
NOT emissions, so it does **not** feed Phase 3a — only a PLAN-fence `facts` declaration does.

## What is PROVEN, and how (do not re-derive)

| claim | evidence |
|---|---|
| all four phases in the running binary | binary probe, each string first confirmed present in SOURCE, with positive + negative controls (`bugs_open/288` §5b) |
| stage 2b fires on a CITATION fact | the real sweep returned **two** entries for `sdlt-ftb-relief-cap` — `artifact_check` AND `citation`. Before stage 2b the first could not exist |
| `ownsVerifiedAt` both arms | same sweep: `sdlt-ftb-relief-cap`/`artifact_check` has **no** `verified_at` (secondary); `gd-trials`/`artifact_check` **has** it (primary, artifact-only fact) |
| durable `subject_key` addressing | induced drift named `tool "stamp-duty"`, resolving through the name rule, not a component id |
| the drifted arm | induced (pattern → expired `625000`) → `drifted`, while the citation arm stayed `fresh` and undisturbed |
| dry runs write nothing | 13 items before and after; register restored **byte-identical by md5** |
| Phase 4 on real data | 5 notes / 3 sites / 0 errors, incl. **7 correct bindings for `mortgages-stamp-duty`** |

## ⚠ THE COUPLING THAT CHANGES THE ROADMAP

**Phase 3a is starved until Phase 4 is adopted.** The probe annotates *emissions*; a self-quieting
fleet emits nothing; so the first real sweep produced **0 annotations**. The
present/absent/markup-only distribution that is **Phase 3b's entire precondition** accumulates
only when a NEW declaration files its one-time `unreconciled_declaration` batch.

**So the order is: adoption → distribution → 3b.** They are in series. "Run it for a month" was
wrong; the right statement is "3b is unreachable until someone adopts". If LMC take their 7
bindings, that is 7 annotated emissions on the next pass and the measurement begins.

## WHAT IS OWED, in order

1. **Roll `bba8a892d`**, then the probe above.
2. **Chase the five suggestions.** They are notes on subjects owned by three lanes; a note is an
   input, not a work item, so nothing chases them by itself. **`mortgages-stamp-duty` is the
   highest-value one** — the estate's second SDLT calculator, 7 correct bindings, currently
   guarded by nothing. CONTRIB filed to that lane 2026-08-24.
3. **Then the distribution** (§5c), and only then argue Phase 3b in its own council round.
4. **`bugs_open/288` §5.4** — the `improve_tool` arm has still never run in production. Measured:
   reachable on **91 of 178** exposed tool pages, so unexercised rather than dead.
5. **§5.5, the prose half** — deliberately untouched; `bugs_closed/093`'s resolution (a second
   SURFACE for the existing scanner, never a second scanner) is the template.
6. **Piece 4, the oracle** — *is the figure RIGHT* — out of scope, behind its own RFC. A tool and
   register wrong in the same direction still agree silently, and nothing here changes that.

## Landmines this lane has earned (read before touching the probe)

- **Probe SCRIPT TEXT, never the whole page.** The register's own `writer_line` puts the figure in
  the PROSE; a whole-page check would have certified `bugs_closed/225` daily for sixteen months.
- **Never tokenize a `string_agg` of `rendered_html`** — partial fragments; one unbalanced
  `<script>` leaks into the next component's prose. Same certification, one level down.
- **The floor (1000) is MEASURED**: 32.75% / 3.79% / 0.06% / 0.03% / 0.00% false positives at
  1–5+ digits over 161 real tool pages with invented probes. Do not lower it by argument.
- **A trailing comma is a LIST SEPARATOR** — excluding it hid the real band table. RE2 has no
  lookaround, hence hand-rolled byte checks.
- **Delete the CALL, not just the body** — three guards in this lane passed their tests with the
  call site removed.
- **A fixture built to DEMONSTRATE a rule will not TEST it** — assert the premise.
- **grep the file for the word your comment uses.** `factProbeNotProbed` was documented as
  refusing the "ambiguous" case for a fortnight of my own reading; the arm did not exist, and one
  `grep -n ambig` would have said so.

## Register / bugs / docs

`bugs_open/288` §0 (four corrections to the file itself) · §5b (live proof) · §5c (the first real
sweep + the defect) · §6 (what is owed). Concept register **CLM-022** carries every landmine
above. `WRONG_CALLS.md` entries 1–6 for this lane, 2026-08-24/25.
