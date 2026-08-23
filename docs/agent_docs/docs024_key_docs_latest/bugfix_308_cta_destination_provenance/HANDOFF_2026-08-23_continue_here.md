# HANDOFF — `bugs_open/308`, CTA destination provenance. Continue here.

**Written 2026-08-23 ~13:30Z. Supersedes `HANDOFF_2026-08-22_continue_here.md`**, whose §4 BLOCKED
banner is FALSE (see §0). Lane dir:
`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`

Read this, then `CALIBRATION_2026-08-23_phase_b_widening_report.md` (the evidence everything rests
on), then `NOTES_…` (technical log, newest at the bottom). `README_where_we_are.md` is the owner's
plain-prose log. Newest milestone read-out: `SUMMARY_2026-08-23b_…`.

---

## 0. ⚠ The previous handoff's blocker is RETRACTED

It said the estate's LLM budget was exhausted until **2026-09-01** and that Phase B must not start.
**The cap was on the WRONG ANTHROPIC ACCOUNT** — the console the owner lands on by default is not
the org the fleet's key belongs to, so it read `0% used` while the API refused calls, and
`2026-09-01` was that other account's monthly reset (see
`memory/the-fleet-key-is-not-on-the-default-console-org.md`). Lifted **2026-08-23 10:10:40Z**,
measured at `llm_call_log`. Council rounds run normally again — one was dispatched and answered
inside an hour today.

## 1. State in one paragraph

**Phase A (the provenance record, LNK-035) is LIVE and proven at the artefact. Phase B (the
widening, LNK-036 + LNK-037) is COMMITTED (`7f85aa814`) and INERT UNTIL THE NEXT FLEET ROLL.**
`bugs_open/308` **stays OPEN** — nothing has touched a served page, which is that bug file's own
verification bar #1. Two council correlations are outstanding (§4), and `RFC_047` carries one live
question for the owner (§5).

## 2. What Phase B did

One shared candidate universe — `datahelpers.LoadCTALabelUniverse` — consumed by the misdirected-CTA
**detector** and **both** CTA writers; `candidatesFromHubs` deleted. So the resolver can now mint
`/contact.html`, which is the fix this bug was filed for. Plus `BestLabelMatch` now returns a third
value, `ambiguous`, and refuses a winner decided only by alphabetical order.

**The POSITIONAL pick is NOT widened.** `rank()` still refuses every utility area, and the loaders
keep their third consumer — the site HEADER CTA, which no `content_data` diff could ever show.

**Both writers' keeps changed, and that is the half a reader will miss.** A MINTED utility
destination whose label goes generic used to take no keep at all and fell to the positional pick —
`bugs_open/248`'s clobber arriving through 308's own fix. The invariant that replaces LNK-033's:

> **The positional pick may neither CHOOSE nor DISPLACE a utility destination.** Only a confident
> label match puts one there or moves it away.

## 3. The numbers you must not re-derive from scratch (all [MEASURED 2026-08-23])

| | |
|---|---|
| fleet CTA writes today → after widening → after the ambiguity refusal | **32 → 428 → 291** |
| wide-pool matches decided by ALPHABETICAL ORDER alone | **263 of 1,146** (137 would overwrite a live CTA) |
| this bug's findings (was 200 on 08-22) | **188** — `complete` 63 items/99 findings, `unresolved` 53/86 |
| …repairable by the writers at all | **147 (78%)**; the other **41 can never be** — prose anchors |
| all `misdirected_cta` findings fleet-wide → reachable by the writers | 1,855 → **675 (36%)** |
| live pages that are planned-and-never-deployed | **43 of 764**, and **10** findings named one |

**Three things measurement settled, so do not re-open them without new evidence:**

1. **Do NOT add `about` to `LabelStopwords`** — it suppresses the false matches AND the correct
   `Learn More About Us` → `/about.html`.
2. **Do NOT take the narrow widening** (utility pages only): 108 writes vs 291 and *more* of them
   wrong. A pool that omits the label's real target gives the matcher a monopoly, not a choice.
3. **Do NOT invent a fourth tie-break key.** Token-set-size (2026-08-11), name-tier and path-depth
   (2026-08-23) have all been measured over the fleet and rejected for the same reason.

## 4. Outstanding: two council correlations

| what | correlation | state |
|---|---|---|
| Phase A | `e4336931-487b-4db3-b4dc-a4b128b3566c` | **REVISE ×4** (rounds 1-4). Every one a SUBMISSION defect; the code was right each time. |
| Phase B | `00732119-4e24-43c3-bd5e-ba30ced47f15` | dispatched 2026-08-23 ~12:55Z |

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;
```
⚠ Never read a verdict via `doc_notes ORDER BY created_at DESC LIMIT 1` — with ~40 live sessions
that returns whoever finished last. And a resubmit writes another `iteration = 0` row, so key on
the correlation and count rows.

**Round 4's gating objection (editquality, HIGH) is CORRECT and unfixed:** Phase A's edit 2 claimed
`storedCTADestinationIsAuthored` as its symbol but showed a call-site-only sketch, byte-identical to
edit 3's, so the one function whose logic changed was never shown. Fixing that means regenerating
edit 2's sketch from `git show 288ce3e7a -- …resolve_internal_links_action.go` scoped to the
predicate's own hunk, merging edits 2/3, and adding the register + LANDMINES edits the notes claim.
**The code needs no change.**

**Round 4's `bug_historian` objection is ANSWERED** (NOTES §2): `apply_section_edit` does write
`content_data` fields directly and never calls `SetCTAMinted` (`section_editor_actions.go:1025-1026`,
0 mint calls in the file) — and every reachable state is safe *because the record is value-bound*.
A stale record cannot vouch for a value it does not name.

## 5. What to do next, in order

1. **After the next fleet roll, verify at the artefact.** Capability probe with a control
   (`LoadCTALabelUniverse` present, a string the change did not add absent, same exec, every
   replica), then induce a discovery run and a `cta_links_stale` rerender on **finetuning.uk** (55
   of the 188 findings) and load the page. Watch **both** directions: utility-destination CTA rows
   should rise from ~0, and `misdirected_cta` items/day should fall — a fall alone could just mean
   the detector stopped running. ⚠ The detector is not on a reliable schedule (08-22: 40 items,
   08-19: 3, 08-18: 128), so induce rather than wait.
2. **`RFC_047` needs an owner ruling.** Its §9 asks one live question: should the DETECTOR keep
   guessing on a tie where the writers now refuse? Filing a finding is cheaper than a write, so it
   is arguable — but it re-drifts the two halves, which is what 308 and `bugs_open/203` both exist
   to stop. Do not decide it in a lane.
3. **Phase C, which is what the bug file actually asked for and is NOT built:** `suggested_target`
   still has **no consumer** — the detector computes the answer, writes it down, and the repairer
   re-derives it. And there is still no completion verifier, so a repair that changes nothing still
   completes. `VerifyMisdirectedCTAResolved` (re-run the detector's own predicate before a
   `page_rerender` may complete) is what turns "complete and unchanged" into a refusal. Do **not**
   have the repair execute the stored `suggested_target`: a work item's spec is data written by an
   earlier binary.
4. **Migration `555_requeue_misdirected_cta_stock.sql` is Phase C only**, and only after the Phase B
   image is stamp-verified per service. A status flip is live instantly; flipping under the old
   binary re-runs the broken repair and burns strikes. `[UNVERIFIED]` that `unresolved` is
   non-dispatchable and `detected` is — read the dispatch loop's status predicate first.

## 6. Traps this lane hit (full text in `LANDMINES.md` / `WRONG_CALLS.md`)

- **`git diff` prints NOTHING for an untracked file**, so a generator that builds review sketches
  from the diff emits an EMPTY one and the submission still looks complete. Assert
  `hunks_in_sketch == hunks_in_diff` per file and print the sizes before dispatch. New LANDMINE.
- **I measured the MATCHER and called it the WRITER** (435/298 → 428/291): the writers additionally
  gate on `validPages.Contains`. State the predicate in the same sentence as the number.
- **The working tree does not compile cleanly** — `TestUpdateWorkItemStatus_*` fails from another
  session's uncommitted edit to `load_work_item_actions.go`. Verified by running the same tests on
  a clean `git archive HEAD` tree, where they pass. Do all Phase B testing in a scratch tree.
- **`/tmp` is a 16 GB tmpfs and was 100% full**, which fails the Go linker with "no space left on
  device" and reads like a broken build. `export TMPDIR=<scratchpad>` rather than clearing a shared
  tmpfs.
