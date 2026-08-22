# HANDOFF 2026-08-22 — bugs_open/198, continue here

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
2. **Candidate 6 — a separate defect, deserves its own bug file.** css-patch writes rules
   against selectors that do not exist: `H3.H3` (dartsonline), `p.P` ×2 (remortgagecalculator) —
   `render_audit.py` labels findings by uppercased TAG and the agent reads that label as a class.
   Three sites' evidence. Also: even a correct rule loses on source order when the offending
   declaration sits in page-level component CSS emitted after the stylesheet. Measurable
   precondition: grep the theme for the selector before planning.
3. **The round-trip-writer inventory** — owed since council round `5249320e` (2026-08-05).
   Which other `agent_definitions` workflows round-trip a whole artefact through an LLM into an
   unguarded writer. **Not absorbed by this work**: a guard for one seam is not the class survey
   the architecture seat asked for. Method is in `HANDOFF_2026-08-10_continue_here.md` §"the
   6-step method"; its blind spots are named there.

## 5. Can 198 be closed?

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
