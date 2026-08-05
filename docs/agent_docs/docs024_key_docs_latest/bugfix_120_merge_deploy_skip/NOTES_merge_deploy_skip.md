# NOTES — bugfix 120 (append-only, newest at the bottom)

## 2026-08-05 ~11:20 — bug selection and ownership checks

- Session named "bugfix 201" by the owner; found `bugs_open/201` OWNED: the
  184 lane's live transcript (`b9c69b3e…jsonl`) shows the owner pointed that
  session back at 201 at 10:12:48Z and it replied "Picking this up per my own
  handoff's step 1" — active minutes before this session started. Did not
  compete; picked the next unowned bug instead.
- Backlog ranked by reference-heat over 37 live transcripts (last 4 h,
  excluding self): 120 coldest at 13 refs (next: 197@19, 196@20, 114@21,
  085@26). 197/196 belong to the bugfix_195 lane (fresh continue-here handoff
  today); 114/085 cited heavily by the very active brochure lane. 120's only
  citing lane (webdesign_couk) last touched it 07-29.
- Symbol grep for 120's fix site (`deploy-to-b2|github.event.before|Get changed
  domains`) over live transcripts: only context-level mentions (098-retraction
  lane reading the workflow as background; finetuning.uk hosting session
  describing the deploy chain). Nobody at the fix site. Claimed at ~11:30
  (`8d16cdae6`).

## 2026-08-05 ~11:35 — validity re-verified first-hand

- `gqls/sites@bbd7703a4` (origin/master, fetched today): defect present
  (`fetch-depth: 2`, `HEAD~1 HEAD` diff). Local clone was 1 commit behind and
  carries other sessions' untracked scratch — not swept.
- `gqls/vm-sites@bec162b` (origin/main, fetched today): same line at :42.
  `deploy-targets.json` allowlists only `relojistas.com`.
- Deploy cadence: ≥40 runs/24 h on gqls/sites (gh run list, capped). Pack size
  86.67 MiB (the fetch-depth: 0 one-time cost).
- LANDMINES pre-read for the touched footprints: `--skip-newer` revert trap
  (deletions unaffected — matters for probe cleanup); local `grep` is ugrep and
  does NOT match the workflow's ERE (never preview the pipeline locally);
  `gqls/sites` has no `main` branch (master is default); `sites.github_repo`
  routes some domains to vm-sites (idea.uk, relojistas.com).
