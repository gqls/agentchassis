# 254 — code-index answers describe the last PUSHED tip, but the commit vocabulary never reaches the answer site, so a model explains staleness with kinds

**STATUS: CLOSED 2026-08-14 — fixed AND live on `v1.0.1297`, council APPROVED.
Moved to `bugs_closed/` under the owner's 2026-08-12 restoration of the fixed-and-live
bar (superseding the 2026-08-06 keep-in-place direction this file's original banner cited).
Closure evidence in the dated section at the bottom.**

Filed by the `bugs_open/223` lane (bugfix_223_index_answerability). Diagnosis route
declared per the OWNER RULING of 2026-07-31: a `090` round ran
(`RUN_CORRELATION_ID=520b2f7e-5473-4655-8f41-9a04b7b9eab1`, work item `3382da26`,
completed 13:18Z) and returned `UNVERIFIABLE` — but not empty-handed: its last
hypothesis refuted the filed premise and named the one deciding check it could not
run. That check (an `llm_call_log.prompt_rendered` census, the loop's own "option
b") was then run first-hand and is quoted below, alongside a controlled both-ways
behavioural proof. The substitution is therefore the loop's own residual question,
answered with the loop's own prescribed query.

## Symptom

A landmine-verifier verdict explained the absence of `ValueDef` (a `struct`, three
siblings from the same file indexed at that moment) as "of kinds not indexed" —
false. The symbol was absent because the commit adding it had not been **pushed**:
`analyse_repo_local` fetches `tarball/<ref>` from the remote, so `code_symbols`
mirrors the last pushed tip. Measured at the time: the index sat **246 commits and
88 changed `.go` files** behind the working tree.

**Proven both ways** (NOTES, 08-11): after the owner pushed and the index refreshed,
the *same entry* flipped `NEEDS_HUMAN_REVIEW` ("of kinds not indexed") →
`STILL_VALID` ("are present"), 2 NOT ANSWERABLE → 0, with the indexed commit the
only variable.

## Root cause — not a missing guard; a guard whose vocabulary sits in the wrong place

The obvious diagnosis ("nothing states the indexed commit") is **wrong**, and was
the 090 symptom's own error (recorded in `WRONG_CALLS.md`, 2026-08-11):

1. `freshnessBanner` (`diagnose_code_lookup_action.go`, shipped `87d0bcf97`,
   2026-07-28, `bugs_closed/108` defect A, in the deployed image) renders the
   indexed commit, ref, age, and "The index mirrors the last pushed tip — local
   unpushed work is never visible." on **every** run.
2. That text **was in the verdict-forming prompt**. `llm_call_log`, all four
   `verify`-step calls in the incident window (09:55, 10:00 ×2, 10:19Z):
   `prompt_rendered LIKE '%index freshness%'` → **t** ×4; `'%never visible%'` →
   **t** ×4; `'%INVISIBLE below%'` (the loud STALE branch) → **f** ×4.
3. The workflow consumes the right field: `landmine-verifier`'s `verify` and
   `verify_unverifiable` steps read `results_text` (banner included);
   `persist_verdict` suffixes `evidence_line`.

What actually failed, two mechanisms in series:

- **The STALE branch is clock-gated** (`codeIndexCommitStaleAfter = 48h`) and the
  indexed commit was ~17.5h old, so the calm FRESH variant rendered — while the
  index was 246 commits behind the tree the questions were about. Staleness on this
  tree is commit **distance**, not wall-clock age, and the pod cannot measure
  distance (it has no checkout; only a reader can diff). No threshold fixes this.
- **The model used the vocabulary rendered beside the empty answer and talked past
  the header.** Its wrong verdict *quoted* the kind census (phase 1's note, rendered
  at the 0-row answer) while the commit caveat sat a screen above. This is the
  file's own stated design rule — "the distinction has to travel WITH the data"
  (`codeIndexScope` doc comment) — observed failing at the one place it was not
  applied: freshness travelled as a header; kinds travelled with the answer; the
  answer-site vocabulary won.
- Additionally, the **persisted** `evidence_line` (suffixed onto every doc_notes
  verdict) carried no commit at all, so a verdict read months later could not be
  dated against the code.

## Fix (this commit)

Same seam as phase 1, three additive renderings, no schema change, no new keys:

- `codeIndexFreshness` now also returns the `indexFreshness` struct it read (one
  query, two renderings); `loadCodeIndexScope` takes it as a **parameter** (not a
  field a caller may forget — a forgotten assignment would silently reproduce the
  header-only state).
- `emptyAnswer` appends `indexedAsOfNote()` to every in-scope empty answer:
  commit, ref, absolute commit time (minute precision — the motivating gap was
  same-day: index 16:27, symbol 23:13), and the reading to block, by name:
  "if the target postdates that commit, this 0 is INDEX STALENESS — not absence,
  not removal, not a rename." Deliberately **not** gated on the 48h clock.
- `codeEvidenceLine` (→ persisted verdicts, both lanes) now ends: "Answers
  describe indexed commit <sha> (ref …), committed <time> — the last pushed tip,
  not the present tree."
- The runtime lane (`diagnose_load_runtime_action.go`) threads the same struct.

Tests: `TestIndexedAsOfNote`, `TestEmptyAnswersCarryTheIndexedCommit` (through
`answerCodeCheck`, all three arms — written call-site-first per the WRONG_CALLS
2026-08-10 mutation lesson), `TestCodeEvidenceLineNamesTheIndexedCommit`. Both
unwiring mutations were run and each fails its test; full `actions` package green.

## How to verify live (after a roll)

1. `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor <this commit> <stamp>`.
2. Fire a landmine entry naming a symbol committed but not pushed
   (`./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' <branch>`): the
   0-row answers must carry `as-of: this answer describes commit …`, and the
   persisted doc_note suffix must end `… not the present tree.]`.
3. The negative control: a symbol that genuinely does not exist anywhere must
   still render the honest-absence wording alongside the as-of note — the note
   qualifies, it does not retract.

## What this does NOT fix

- The index still cannot see unpushed commits (structural: the indexer reads the
  remote). The note makes that legible per-answer; only a push advances the corpus.
- The verifier can still reach a wrong verdict against a caveat it can now read at
  the answer site; `[UNMEASURED]` how often the note flips outcomes fleet-wide —
  the both-ways proof above is n=1 on the motivating entry.
- The 48h clock branches of `freshnessBanner` are untouched — they remain correct
  for what they measure (refresh-pipeline death, wall-clock age).

## Relations

`bugs_open/223` (kind/extension answerability — this is its commit-axis sibling) ·
`bugs_closed/108` (its fix works; this is the residual it leaves) ·
`LANDMINES.md` "The code index is only as fresh as the last PUSH …" (corrected
2026-08-11 — the reader-side check lives there and is unchanged by this fix) ·
`WRONG_CALLS.md` 2026-08-11 (the filed premise this diagnosis corrected) ·
`bugfix_223_index_answerability/NOTES` 2026-08-11 ~18:30Z (full evidence trail).

## CLOSURE EVIDENCE — 2026-08-14, fixed AND live

- **Council:** corr `42afbd67-48c7-4581-915d-2880cd1dc74d` → **approved, "all
  reviewers approve"**, 4 abstained, no blocking objections (editquality's two
  minor trust points were both already disclosed in the submission). The commit
  (`0c880908a`) carries `Council-Submitted:`; 098 credits it at report time.
- **Live, per SERVICE, at the artefact** (three independent checks, each with a
  control): pod imageID digest `2e89958a9b…` = local `v1.0.1297` digest exactly;
  image revision label `3b0ea20ff…` and `git merge-base --is-ancestor 0c880908a
  3b0ea20ff` → yes; pod binary carries the as-of literal (`grep -c -a "as-of:
  this answer describes commit" /proc/1/exe` → **1**, fabricated-needle control
  → 0/exit-1). ⚠ A first probe run returned thirteen clean "absent" rows —
  every one a swallowed `NotFound` from a pod deleted mid-probe behind
  `2>/dev/null`, the exact trap CLAUDE.md's build section names. Re-run against
  live pods with stderr visible before believing any of this section.
- **Behaviourally, on a real run** (landmine-verifier, corr `16f0475d`,
  2026-08-14 07:39Z, the corrected staleness entry itself): the **persisted
  verdict carries the evidence-line commit clause** — `doc_notes` body matches
  `%not the present tree%` — and the verify prompt shows `Answers describe
  indexed commit …`. Verdict STILL_VALID; the fix's own paperwork was the fixture.
- **Residual, stated:** the as-of note renders only on EMPTY answers and that
  run had none, so its first live rendering is still unobserved (build-time it
  is mutation-proven through `answerCodeCheck` on all three arms, and the
  literal is in the running binary). The check, when one occurs:
  `SELECT count(*) FROM llm_call_log WHERE step_name='verify' AND
  prompt_rendered LIKE '%as-of: this answer describes commit%';` — do not
  "verify" this with a zero; an empty result means no empty answer has happened
  yet, not that the note is missing.
