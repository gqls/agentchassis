# HANDOFF 2026-08-05 — bug-backlog pickup: 188 and 200 closed; here is how to take the next one

Written by the session that closed `bugs_closed/188` (renders photograph the
driven page) and `bugs_closed/200` (hamburger renders as one bar, fleet-wide).
The working brief it ran under, verbatim from the owner, worth re-running:

> Look at bugs_open and find the next bug that isn't being worked on in another
> thread and take it on. Research the docs for previous discussions and prior
> code. Plan with fable preferring a robust framework-level solution; fix with
> opus. Check the council. Check other threads. Check the bug is still valid.
> Commit what should ride the next build. Close the ticket and move it to
> bugs_closed. Keep the docs updated; missteps to WRONG_CALLS.

## State as of this handoff (all committed; verify freshness before trusting)

- **188 CLOSED 2026-08-04** — the fix predated the pickup (TL-035 (d), another
  session, same morning). Closure = verification: pod-grep with controls,
  behavioural run, fetched renders. Full trail in the bug file §7.
- **200 CLOSED 2026-08-05** — filed from 188's §3 leftover, fixed at four
  layers (live `layouts` rows via seed `314`; 16 seed files; 15 sites' served
  `styles.css`; `check_flexless_hamburger` guard). Full trail in §8, including
  the two publish failures and their diagnoses.
- **Bugs being worked by OTHER sessions as of 2026-08-04 evening** (from live
  transcript greps — STALE by now, re-check): 190 (content_data envelope) and
  156/others were mid-council in two active sessions. `bugs_open/199` was filed
  by someone else on 08-04.

## The lessons that changed how the next pickup should run

1. **`who-owns <n>` is blind twice over** — it reads commits (a mid-fix session
   is invisible) AND a fix shipped under a different name (188's cited TL-035,
   never "188") never touches the bug file. **`git log` the FILE the bug's §2
   names, then grep the live `.jsonl` transcripts** (`~/.claude/projects/
   -home-ant-projects-agentchassis/*.jsonl`, mtime < a few hours) for the bug
   number AND the code path.
2. **"Check the bug is still valid" can invert the whole task.** Both bugs this
   arc handled changed shape on re-verification: 188 was already fixed (task
   became verify-and-close); 200's "possibly a capture artefact" was a real
   fleet-wide defect. Neither outcome was in the file.
3. **Two fresh landmines are now in LANDMINES.md** (synced to doc_notes): the
   layout seed driver's "re-running is safe" header is FALSE (five 07-02 live
   drifts + `tool-portal-light` unseeded — surgical `replace()` instead, seed
   314 is the worked example); and a git-adapter publish with `repo_name:
   "sites"` SUCCEEDS green while deploying NOTHING for `sites.github_repo =
   'vm-sites'` domains (idea.uk, relojistas.com). Also: Cloudflare edge caches
   CSS for 4h — after any stylesheet publish, verify with a cache-busting query
   before suspecting the deploy, and expect the plain URL to lag.
4. **Production publishes are classifier-gated.** The operator ran the two
   batch scripts by hand (`!` prefix). Prepare the exact command + built-in
   verification, and hand it over; don't fight the gate.
5. **Scale-out note:** browser-runner-adapter went 1 → 29 replicas on the
   2026-08-05 afternoon roll (v1.0.1254). Pod-grep one pod per ReplicaSet
   (same image spec); the per-replica rule matters across deployments, not
   within one ReplicaSet.

## Next pickup, mechanically

1. `ls bugs_open/` (45+ files remain) — skim the newest first; 193–199 were
   filed 08-03/08-04 and are least likely to be stale.
2. For the candidate: who-owns + FILE-path git log + live-transcript grep +
   `site_work_items` queue check (the standing four).
3. Re-verify the defect against the live system before planning anything.
4. The rest of the brief as written. Council for `platform/`/`internal/`/`pkg/`
   only; docs and scripts are refused client-side.
