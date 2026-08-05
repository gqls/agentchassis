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

## 2026-08-05 ~11:29 — council submitted (FORCE=1, reason stated)

> **CORRECTED 2026-08-05:** this entry's header first said "~11:50" — an
> unchecked clock guess. The council's own `fix_plan` artifact row
> (`created_at 10:29:26Z` = 11:29 BST) dates the submission. Caught by reading
> the verdict row, not by me.

- `SUBMISSION_CORR = f029e06f-cfe9-4e01-84fd-1940349fc6d5`. FORCE=1 because the
  edited paths live in the sibling deploy repos + docs, outside the
  `platform|internal|pkg` scope regex — this is the fleet's deploy pipeline,
  not site content, which is what the scope rule exists to exclude. Stated in
  the submission rationale too. Commits carry `Council-Submitted:` per the
  2026-07-30 rule; verdict to be read and recorded here (~30 min budget —
  dispatch queues behind the fleet).

## 2026-08-05 ~12:00 — verdict read, fix live, induction PASSED, close

- **APPROVED round 1** (report 10:34:30Z — 5 min council, no queue this time),
  "2 advisory objection(s) — none high-severity", 7 seats abstained. Both
  medium advisories were acted on before close:
  - editquality: suspected 42 zeros in the sketch's guard literal. Shipped
    files byte-counted: **exactly 40** in both (`command grep -o '"0*"' | awk
    length`). The sketch's JSON escaping misled the seat; the code is right.
  - bug_historian: "no census for a third HEAD~1 copy — asserted, not
    verified." Correct objection; census run (see the bug file's close
    section): no third live copy; `scripts/github/deploy_to_b2_action.yaml`
    (a maintained reference carrying the defect) found by that census and
    refreshed in the close commit.
  - prior_art (low): the "existing deploy-all fallback" claim — confirmed by
    quote: `bbd7703a4` and `bec162b` both carry
    `if [ -z "$CHANGED" ]; then CHANGED=$(ls -d */ …)` pre-fix.
- Fix commits (Opus agent, per plan): sites `efe046c6d`, vm-sites `30e447a92`,
  agentchassis 034 doc `b50d45426`. Pushed ~11:35–11:36 BST; the workflow-fix
  pushes triggered no deploy runs (`paths-ignore`), as designed.
- Induction (all evidence in the bug file's close section): merge `750e8ecd4`
  → run `30998151789` named the PUSHER's domain and served the probe; control
  ff run `30998135192` unchanged semantics; vm run `30998426033` proved the
  sibling step with zero VM impact (allowlist skip). Probes cleaned, 404
  confirmed.
- Misstep worth naming (small, uncommitted to any doc, so not WRONG_CALLS by
  its own bar): queried `diagnosis_artifacts.content` before `\d` — the column
  is `body`. Schema-first exists for a reason; cost one query.
