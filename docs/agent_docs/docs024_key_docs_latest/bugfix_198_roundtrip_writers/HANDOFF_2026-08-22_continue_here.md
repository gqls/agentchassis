# HANDOFF 2026-08-22 — 198 is CLOSED; this is the residuals + cold-start

> **STATUS 2026-08-22: `bugs_closed/198`.** Closed on the owner's instruction after the
> v1.0.1323 roll made DGH-016 live. The close-out banner at the top of that file carries the
> evidence table. Candidate (6) was spun out as **`bugs_open/352`**. What remains for this
> lane is §4 item 3 — the round-trip-writer inventory — plus one opportunistic observation.

Supersedes `HANDOFF_2026-08-10_continue_here.md` (whose stated closing condition — "the only
thing keeping 198 open is the witnessed end-to-end run" — is now MET; see §2).

## 1. One-paragraph state

The prevention half of `bugs_open/198` is **built, council-approved, applied, proven on a live
dispatch, and fully live at the binary as of `v1.0.1323`**. Every linked `css_themes` row in the
fleet (22 of 22) is a plausible, unshared stylesheet, and the producer now keeps them that way.
**Nothing on this bug is blocked and nothing is inert.** What remains is one opportunistic
observation, one separate defect, and one survey owed to a council round — all §4.

## 2. What is DONE (with the evidence, not the claim)

| # | change | state |
|---|---|---|
| 542 | css-patch refuses an unsafe base: `css_len >= 4096 AND site_count <= 1`, `fail_on_non_numeric` | LIVE, applied + recorded |
| 542/546 | all **7** non-success exits stamp the work item BEFORE their terminal, so refusals/failures stop reading `complete` | LIVE |
| 543 | webdesign-agent persists its render into the theme row — the two writers converge | LIVE |
| 547 | finetuning.uk + gaswholesalers.com split to a collection+theme each (owner ruling) | LIVE |
| 548 | webdesign.uk seeded from the vm-sites blob | LIVE |
| DGH-016 | opt-in `file_shrink_floor` on `git_commit`, enforced in the git-adapter | **LIVE on v1.0.1323, both halves** |

- **Council `5f756c51-cdc6-4a48-b5f9-59e472243601` — APPROVED round 1**, 6 advisory objections;
  4 checked, 3 accepted as follow-ons, and the `architecture` seat's correction of my reasoning
  (shape exception ≠ accumulation gate) acted on via the `optional_key_budget_acks.json` entry.
- **The witnessed run (2026-08-21 19:09–19:11Z)** — a real dispatch on webdesign.uk. Terminated
  at `complete_refused`; item `needs_human_review` + `parked_by`, `completed_at` NULL; **row
  still 0 bytes and repo blob byte-identical with zero commits.** The negatives are the evidence.
- **Binary verification (2026-08-22)** — chassis stamps `70e7b4f9c`, of which the Go commit
  `4ee9bfff6` is a proven ancestor; git-adapter carries the enforcement symbols and constants.

## 3. READ THIS BEFORE YOU VERIFY ANYTHING HERE

**The estate's standard post-roll probe gives false negatives.** In-pod
`kubectl exec … grep -ac '<string>' /proc/1/exe` returned **0** for three constants the binary
provably contains, while returning >0 for symbols in the same breath. Negative controls do not
catch it (they only detect over-matching), a positive control from old code passes, and
`grep -c` exits 1 on no-match exactly as on a real absence. **Pull the binary and grep it
locally**, or ask what it was BUILT from and settle it with `git merge-base --is-ancestor`.
Full entry in LANDMINES; near-miss in WRONG_CALLS; recipe in RUNBOOK §7.

Other traps this lane hit, all in the RUNBOOK: a bare `curl` reads a 302 page as a gutted
stylesheet **and** a served-side absence is not an artefact-side absence (§1, LANDMINES);
`run-migrations.sh --apply` would apply every other lane's pending backlog (§5); `LIKE '54[23]%'`
is not a character class (§6); read the ROWS of the workflow-edge query, not just its DANGLING
count (§9); execute an installed step query VERBATIM (§10); check the promoter's own doors
before concluding a probe was refused rather than held (§11).

## 4. WHAT IS LEFT — three items, none blocking

1. **A live enforcement observation for DGH-016** *(opportunistic, ~1 min when it happens)*.
   The guard fires only on a css-patch deploy. On the next real dispatch:
   `kubectl -n ai-persona-system logs -l app=git-adapter --tail=500 | grep 'commit passed the shrink floor'`
   That one line proves the field crossed the wire, the guard measured, AND a healthy commit
   still passes — none of which a binary probe can show. Then update DGH-016's status line.

   > **UPDATED 2026-08-24 (after the fleet roll) — STILL OWED, and now for a MEASURED reason.**
   > Both services re-verified on commit `70fd163c2` with `4ee9bfff6` a proven ancestor
   > (git-adapter by its own provenance line; chassis by in-pod `grep -aq <sha> /proc/1/exe`
   > with a fabricated-sha negative control). DGH-016's status line and the concept index are
   > updated. **Why the observation has not happened:** `grep -c 'shrink floor'` = **0**, and
   > **0** commits in the adapter's last 3,000 lines carry a `.css` file key, against **253**
   > commit/push lines for other types — zero demand, not a silent guard. ⚠ Two demand controls
   > PASSED WHILE BLIND before that one: "any commit" (253) and "any payload containing CSS"
   > (matches inline `<style>` in ordinary HTML commits). The axis that must vary is the
   > **`.css` FILE PATH**. Nothing here can be forced: my own notes deliberately declined to
   > induce a css-patch deploy, because the only sites that exercise the refusal arm are live.
2. **Candidate 6 — FILED as `bugs_open/352` (2026-08-22) and OWNED by another lane since
   2026-08-24. Not this lane's work any more — do not start a competing fix.** css-patch writes
   rules against selectors that do not exist: `H3.H3` (dartsonline), `p.P` ×2
   (remortgagecalculator) — `render_audit.py` labels findings by uppercased TAG and the agent
   reads that label as a class. Also: even a correct rule loses on source order when the
   offending declaration sits in page-level component CSS emitted after the stylesheet.
   Measurable precondition: grep the theme for the selector before planning. That second arm is
   `bugs_open/352` §"The SECOND arm"; the owning lane is closing **arm 1 (the producer) only**
   and scoping arm 2 out explicitly, so 352 stays OPEN with an arm banner.

   > **CORRECTED 2026-08-24 — the remedy I wrote for this was incomplete, and the naive version
   > is WORSE than the defect.** I wrote "omit the class component so the selector is `h3`",
   > calling it "unrepresentable at source", and flagged only the `item_key`/dedup interaction.
   > Today `p.P` matches nothing and is **inert**; lowercased to `p` it recolours every paragraph
   > on the site. The fix must yield a **scoped** selector (ancestor/id-anchored), not merely a
   > lowercase one. Caught by the `bugs_open/352` lane's census, not by me. Measured by me
   > against the live DB **as of 2026-08-24**: of 452 `contrast_failure` rows, **181** carry a
   > `TAG.TAG` selector and the two commonest are `P.P` **×77** and `A.A` **×44** — the two most
   > dangerous possible bare selectors, i.e. the modal case, not an edge. Of those 181,
   > **108 are already falsely `complete`** (this file's dartsonline `H3` row was one instance;
   > it generalises) and **73** sit outside `workItemClosedStatuses`
   > (`platform/orchestration/actions/work_items_common.go:85-91`), so a key-shape change
   > would let the retraction path close them stamped "no longer below its contrast threshold".
   > **The check I skipped:** census the selector population and ask what the corrected selector
   > MATCHES, not just whether it matches. Full account: `NOTES_198_roundtrip_writers.md`,
   > 2026-08-24 entry.
   > **UPDATED 2026-08-24 (later) — ARM 1 IS FIXED and council-APPROVED round 1
   > (`acadbe8b`), committed `ffa6e1c3d`. `bugs_open/352` stays OPEN: arm 2 is untouched and
   > reproducible.** ⚠ **What shipped is NOT the remedy described above, and a reader
   > implementing my version would build the weaker thing.** I said "produce a scoped selector".
   > What shipped composes the selector **in the page** and asserts it selects the element that
   > was actually measured, refusing and counting a bare tag. The invariant is **"prove it", not
   > "stop lying"** — so the next composition defect of this class reports itself instead of
   > minting another 108 false completions. ⚠ Also: migration **587** WITHDRAWS the 73 as
   > `cancelled` (withdrawal, not resolution — it frees the dedup slot so still-failing pairings
   > return under verified selectors), so re-running my census after 587 will NOT reproduce 73.
   > **⚠ CORRECTED 2026-08-24 (same day, my error) — 587 IS `_HOLD` AND HAS NOT BEEN APPLIED.**
   > I wrote the clause above from the 352 lane's message without checking the cluster, and
   > "587 WITHDRAWS the 73" reads as done. It is not. Measured just now:
   > `count(*) FILTER (WHERE result->>'cancelled_by'='migration_587')` = **0**, rows carrying
   > `pre_352_status` = **0**, `contrast_failure` total still **452**, and the census still
   > returns **`complete` 108 / `deferred` 58 / `unresolved` 15** — the 73 are LIVE and still
   > holding their dedup slots. `587_retire_invented_contrast_selectors_HOLD.sql` is committed
   > and held back for ordering, applied BY HAND. **So 73 is the CURRENT figure, not a
   > historical one**, and the "will not reproduce" warning takes effect only once someone
   > applies it. A `_HOLD` migration is committed code that is NOT running.
   > The rekey it replaced was dropped because the two-strike counter reads only
   > `complete`/`failed` **within 7 days** (`load_work_item_actions.go:1519-1523`, verified by
   > me) and the park dates from 08-11 — there was no attempt history left to preserve.
3. **The round-trip-writer inventory** — ~~owed since council round `5249320e` (2026-08-05)~~
   **POPULATION PASS DONE 2026-08-24 → `INVENTORY_2026-08-24_round_trip_writers.md`.**
   Result: **9** writer steps across 6 active definitions reach an LLM output transitively
   (`as of 2026-08-24`); **6 of the 9 are MULTI-HOP** and so invisible to the one-hop join this
   handoff's 08-10 sibling documents — which cannot see 198 itself. The class splits: **3** steps
   where the LLM's returned bytes ARE the artefact, 6 where a deterministic Go composer produces
   them (2 of those write work items, not artefacts). **3 writers REPLACE an incumbent and all 3
   carry a live guard** (542+318; DGH-016 at 0.5; `enforceSectionShrinkFloor`, confirmed in the
   RUNNING build `70fd163c2`). So the class is CLOSED on the measured population.
   ⚠ **It is a FLOOR, not a count, and this is the residual:** the graph is built from config
   TEXT, so an action with an empty `config` that reads its input in Go leaves no edge
   (`webdesign-agent/generate_css`). Closing that needs Go-side input edges — no config query can
   supply them. ⚠ Also in there: why "20 git_commit steps, 1 shrink floor" is NOT 19 bugs.
   Which other `agent_definitions` workflows round-trip a whole artefact through an LLM into an
   unguarded writer. **Not absorbed by this work**: a guard for one seam is not the class survey
   the architecture seat asked for. Method is in `HANDOFF_2026-08-10_continue_here.md` §"the
   6-step method"; its blind spots are named there.

## 5. Can 198 be closed? — ANSWERED: yes, and it is (2026-08-22)

**Yes, in my judgement — and the residuals should not hold it open.** The estate's bar is FIXED
AND LIVE: the defect is fixed at source (543), guarded at both writers (542, DGH-016), proven at
the artefact on a live dispatch, and the damage is fully repaired fleet-wide (22/22). The file's
own stated verify bar is met.

The three residuals are not this defect: (1) is an observation of a proven mechanism, (2) is a
*different* defect that happens to share a handler and wants its own number, (3) is a survey
owed to a review, not a fix owed to a bug. Keeping 198 open for them makes `/bugs_open/` answer
"what is biting prod right now" wrongly — nothing here is biting prod.

**Recommended:** move to `/bugs_closed/`, file candidate 6 as a new bug carrying its three
sites of evidence, and leave the inventory as the council-owed item it already is. **Owner's
call.** If it stays open, the only honest reason is (1), and that resolves itself on the next
dispatch.

## 6. Working record

`docs/agent_docs/docs024_key_docs_latest/bugfix_198_roundtrip_writers/` — PLAN (the four
decisions and why each sits where it does), RUNBOOK (11 sections, every command with its
gotcha), NOTES (append-only, missteps included), README_where_we_are (owner prose),
SUMMARY_2026-08-21 (milestone read-out), and the council submission JSON.
