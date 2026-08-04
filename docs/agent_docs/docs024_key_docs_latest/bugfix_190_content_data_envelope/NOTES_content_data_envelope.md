# NOTES — bugs_open/190, content_data stores the raw LLM transport envelope

Running record. Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — session opens: picking the bug, and why this one

Task was "take the next bug in `bugs_open/` that no other thread is working". The picking
itself was most of the first hour, and two of the three methods I tried were bad. Recording
all three because the bad ones look reasonable:

1. **`git log --grep=<number>` over 4 days.** Useless on its own — a bare number matches
   dates, image tags (`v1.0.1181`), migration numbers and other bugs' cross-references. It
   told me `122` had 50 "commits" and `126` had 0; neither figure meant what it looked like.
2. **Grepping live session transcripts for `bugs_open/NNN_`.** Also bad, and worse because
   it looks precise: every session that runs `ls bugs_open/` has the entire bug list in its
   transcript, so *every* bug scored 2–6 "live sessions". A signal that fires for everything
   is not a signal.
3. **What actually worked** — grep the transcripts for the bug file appearing as a
   `"file_path"` in a tool call, i.e. sessions that actually *opened* it. That separated
   ~11 unclaimed bugs from ~35 claimed ones.

Then a piece of luck worth stealing: session `62bc0e8b` was itself running a session→bug
census and had left the output in its transcript — a table of every live session against the
bugs it mentions most. That is a better instrument than anything I built, and it immediately
showed `181` (which I had provisionally picked, and which is *labelled* "OPEN, UNOWNED" in
its own header) was in fact being worked by two live sessions, one of them the `163` lane
with 34 mentions.

> **MISSTEP, and the useful kind.** I had already half-committed to `181` on the strength of
> the words "OPEN, UNOWNED" written in the file on 2026-08-02. **A file's own ownership
> claim is a snapshot of the moment it was typed**, and this tree turns over fast enough
> that two days is long. `who-owns.py` is documented as lagging; what I had not internalised
> is that the *bug file itself* lags in exactly the same way and is more persuasive because
> it is phrased as a fact.

Also checked and rejected, each for a stated reason rather than a vibe:

- **`146` (ported tool pages outside every acceptance tier)** — unclaimed, and the
  structural half is **already fixed**. `discovery_checks/tool_eligibility.go` exists,
  shares one predicate between Tier 2 and Tier 4, and handles ported pages by keying on the
  page rather than `cc.function`. It shipped 2026-07-29 under commit `ac9f75a0c` crediting
  **`bugs_open/084` candidate 3**, not `146` — which is why `146` still reads as open. Worth
  a close-out by someone; not a fix task.
- **`093` (stat audit has one guarded call site)** — the code fix is built and live since
  `v1.0.1172`, and the file's own final triage says it is now blocked on `bugs_open/083`
  (the check has no scheduler). `083` is being actively worked by another session. Nothing
  for me to do that would not collide.
- **`096`, `126`** — both end in an owner/architecture call, not a code change.

Landed on **`190`**: unclaimed by every instrument above, framework-shaped, and its fix
candidate (1) is explicitly a make-the-bad-state-unrepresentable change.

## 2026-08-04 — re-validating 190, and finding two errors in it

The bug is **still live**: 2 envelope rows. I took the denominator in the same query as the
numerator, which this bug file itself recommends (its filer had been burned by an
empty-population count):

```sql
SELECT count(*) FILTER (WHERE content_data ? 'type' AND content_data ? 'result'
                          AND content_data->>'type'='text') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS denom_nonnull,
       count(*) AS denom_all
FROM page_components;   --  2 | 1054 | 1207
```

`site_components`: `0 | 54 | 54`. The bug never checked the sibling table. Clean today, but
it is a `content_data` store and belongs in the guard's scope.

**Error 1 in the filed bug — the row identity moved.** The file names
`25c73a1c-…` for gaswholesalers `how-pricing-works`. That id does not exist. The row serving
that page is `17e7739e-…`, `created_at = updated_at = 2026-08-03 22:35:17` — *after* the bug
was filed at ~21:30Z the same evening.

**Error 2 — the file's own verification recipe would miss half the population.** It says the
guard "must key on the exact two-key shape". The two live rows do not share a key set:

| row | top-level keys |
|---|---|
| `d2e9644b` finetuning | `{content, result, type}` — **three** |
| `17e7739e` gaswholesalers | `{result, type}` — two |

An "exactly two keys" predicate is silent on `d2e9644b` — the very row the file says is
fully repairable. The discriminator has to be the envelope *signature* (`type == 'text'` and
a string `result`), and `d2e9644b` is the interesting case because real content and envelope
keys coexist in one map.

## 2026-08-04 — the misstep I caught before it reached the bug file's conclusions

`page_component_history` has **65** envelope-shaped rows, all `source =
'save_page_sections_overwrite'`, latest 2026-08-03 22:35:17. My first reading was "the seam
has written 65 envelopes" and I very nearly wrote that down as the headline.

**It is not a write count.** The history INSERT
(`save_page_sections_action.go:586-601`) is `SELECT pc.content_data … WHERE pc.page_id = $1`
executed *before* the DELETE — it archives the state being **replaced**. So 65 counts
*overwrite events on pages that already carried an envelope*.

The cheap check that settled it was reading the SQL in the action, not querying harder. A
data question that is actually a code question will happily return a confident number.

Two things fell out of doing it properly:

- **Blast radius is much larger than the filed "2 rows"**: those 65 events span **25 distinct
  pages across 6 sites**. ⚠ `count(DISTINCT component_id)` returns **0** on the same query —
  a NULL trap, not an absence: the FK is `ON DELETE SET NULL`, so archived rows whose
  component was later deleted carry `component_id = NULL`. Group by `page_id`. Same shape as
  the `distinct_content = 0` trap already in `LANDMINES.md`.
- **Generation has stopped; propagation has not.** Exactly ONE envelope event since
  2026-07-18 (the 08-03 one). The three-tier parse fix worked — no new envelope has been
  minted since mid-July. What the 08-03 event shows is the save seam archiving an envelope
  and 84ms later creating a *new row carrying the same envelope forward*. The live defect is
  that **`save_page_sections` re-persists a transport envelope it is handed, indefinitely**,
  so a poisoned row survives every rebuild instead of being cleaned or refused by it.

That reframing matters for the fix: the guard's job is to stop propagation, not to stop
minting. It also means the urgency is lower than "a writer is actively producing poison" —
and I had written exactly that sentence into the bug file before correcting it there too,
visibly, rather than editing it away.

## 2026-08-04 — constraint noted before any code is written

`save_page_sections_action.go` is **dirty in the shared tree from another session** (the
`bugs_open/194` lane: +62/-2, adding `resolveSectionsMetadataField` and a
`require_sections_metadata` refusal). A pathspec commit still takes a **same-file**
passenger. So the fix must keep its footprint in that file to the smallest possible
call-site edit, with the logic in a new file in the same package — which is exactly the
precedent the 194 lane itself set with `save_sections_metadata_source.go`.

## 2026-08-04 — implementation, and the three things that pushed back

**Design decided by a constraint, not a preference.** Fable's plan and my own check agreed on
where the guard must live, but the deciding fact is mechanical: `actions` imports
`datahelpers`, not the reverse, and the normaliser needs `ParseLLMJSONWithProvenance`, which
is in `actions`. So a `datahelpers` home is an import cycle. The pure predicate *could* live
there; splitting one rule across two packages is the drift class this codebase keeps paying
for, so it is all in `actions/content_data_envelope_guard.go`.

**MUTATION TESTING — four run, four red, and this is the part I would not skip again.** This
guard's happy path is "change nothing", which is also what a completely broken version does,
so a green suite is nearly meaningless on its own. Each mutation was applied to the shipped
code, the named test run, then the file restored and `diff`ed against a pre-mutation copy:

| mutation | test that went red | message |
|---|---|---|
| predicate keyed on `type` alone | `TestLegitimateContentDataPassesByteIdentical` | "legitimate content_data was refused" |
| provenance rule replaced with `if false` | `TestLossyProvenanceIsRefusedNotDecoded` | "a lossy recovery was accepted" |
| predicate keyed on `len(m) == 2` | `TestSupersetEnvelopeIsDetected` | "a three-key envelope was not detected" |
| seam assigns to a copy, not in place | `TestSeamMutatesSectionsInPlace` | "not normalised IN PLACE" |

The third one is the bug file's own recommended predicate. It is worth sitting with that: the
instruction written in the case file, followed faithfully, produces a guard blind to half the
known population — and every test would still have passed, because the fixtures would have
been built from the same wrong assumption.

**The package's own coverage test caught me, and it was right to.** The full suite failed with
`content_data_envelope_guard.go reads the "__truncated" marker but implements no action in
truncationAwareActions`. My file names `__truncated` only in a header comment, as an example of
the `__`-prefixed transport keys a decode drops — but `truncation_guard_test.go` scans package
*sources*, so **my comments are load-bearing** (the shape already in MEMORY as "a source-scan
test makes your COMMENTS load-bearing").

I chose to **exempt rather than register**, and the reasoning is the interesting bit.
Registering `save_page_sections` in `truncationAwareActions` would have been the quick fix and
it would have widened what counts as a truncation-aware consumer for *every workflow
containing that action* — blast radius well beyond this bug, taken silently, which is exactly
what the platform-seams ruling exists to stop. So: an exemption, with its reason stated, and
the reason itself converted into a test (`TestTruncatedEnvelopeCannotReachTheDecodeBranch`)
rather than left as an assertion — a truncated payload either fails to parse or recovers only
via a lossy tier, and both outcomes refuse, so dropping the marker cannot lose a warning
anybody would have acted on.

**Two collisions with other sessions inside one hour, in the same file.** First, the 194
lane's uncommitted work in `save_page_sections_action.go` vanished mid-session — they
committed (`47ee3ebce`) and the passenger risk went away. I edited the file. Then a *third*
session began adding `bugs_open/156` dedup wiring to the same file and, briefly, the package
would not compile at all (`duplicatesCollapsed declared and not used`) — a build break that
was nothing to do with me and stopped a mutation run dead. It resolved on its own minutes
later.

**So the package suite is currently RED in the working tree and that is not my change.** The
failing test is `TestCollapseLeavesNullContentDataWithDifferentHTMLAlone` in
`save_sections_dedup_test.go`, an **untracked** file belonging to that session, referencing
none of my symbols. Proving that took the only instrument that actually settles it:
`git archive HEAD` into a clean directory, copy in *only* my files, run the suite there — green
in 0.492s. That recipe is now in the RUNBOOK, because "the tree is red but not because of me"
is a claim, and on a tree this shared it needs evidence rather than confidence.

**Consequence: I committed four files and held one back.** `save_page_sections_action.go` now
carries both my guard call and their uncommitted dedup wiring, which calls their untracked
file — committing it would break HEAD for every session. The guard, its tests, the truncation
exemption and the section-editor seam went in as `ce675f019`. The save seam's one-line call
site waits for them to land theirs, and until it does **the primary seam is not yet guarded**.
That is stated in the commit message rather than left for someone to discover.

## 2026-08-04 — a ruling I half-satisfied

Shipped `ce675f019` with no concept-register entry, then added PBP-032 in the next commit. The
2026-07-29 narrowing retired condition (1) of the ordering exemption and left
registration-in-the-same-commit as the *whole* of the requirement, so this is a straight miss.

What makes it worth recording is *why* it felt fine: I had reasoned hard about RFC_010 and
written the no-opt-in-field argument into both the commit message and the council submission,
because that was the half I expected to be challenged on. **Satisfying the condition you
rehearsed reads as satisfying the rule.** Nothing in the tooling catches it either — the
`commit-msg` nudge and the `098` report both check the council trailer, which I had. Full entry
in `WRONG_CALLS.md`, and the one-command check is now in the RUNBOOK.
