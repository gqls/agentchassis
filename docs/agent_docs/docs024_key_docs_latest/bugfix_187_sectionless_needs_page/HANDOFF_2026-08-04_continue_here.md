# HANDOFF 2026-08-04 — 187 is CLOSED; this file is the cold-start for the NEXT bug thread

Written by the bugfix_187 lane at high token load, per the owner's instruction:
"find the next bug that isn't being worked on and take it on" is a standing
loop — 187 was this thread's iteration; a new chat picks the next one.

## State: bugs_closed/187 — nothing left to do except two watch items

Fixed AND live AND re-proven across two rolls (v1.0.1248 built by this lane;
survived the fleet roll to v1.0.1250, re-grepped both replicas: pos 5/3,
neg 0). Council `e2e87b04` APPROVED r1. Full record:
`bugs_closed/187_HANDOFF_...md` (close-out section) + this dir's
PLAN/NOTES/RUNBOOK/README/SUMMARY. Register: WII-010. Landmine filed
(site_work_items.page_id NULL — join pages BY NAME). Leak measured stopped:
0 new parked no-sections rows since the guard (was ~3/day); raise arm
witnessed live at 08:41Z.

**Watch item 1 (skip arm, nobody owes work unless it FAILS):** next natural
image-landing on a sectionless non-plan page should log
`skipped_sectionless_page` and mint NO row. Check:
`SELECT item_key, status, created_at FROM site_work_items WHERE
 item_type='needs_page' AND created_at > '2026-08-04 08:34Z' ORDER BY 3;`
plus a pod-log grep for `skipped_sectionless_page`. If a sectionless page
row EVER parks again with the no-sections error, the guard has a hole —
reopen against WII-010.

**Watch item 2 (owner call, not code):** 4 robot-hands tool pages are
current-plan members with zero declared sections — they park honestly as
`unknown`. TL-009 is the owner question (should tool pages declare
sections?). Do not "fix" these rows; they are the residue the guard
correctly refuses to guess about.

**Explicitly NOT ours:** the two post-guard `index` rows failing on
`process_sections_loop`/`sections_for_render.sections_ready` are
`bugs_open/192` (filed 08-03, another lane owns it — check who-owns before
touching).

## For the next thread: the bug-picking protocol that worked

1. `ls bugs_open/` for candidates; newest spin-outs filed "at the council's
   direction" with a closed filing lane are the cleanest takes.
2. Ownership is THREE checks, not one: `scripts/who-owns.py <n>` (lagging —
   reads commits); `git log --oneline -3 -- bugs_open/<file>` (a lone
   `file(NNN)` commit = likely unowned); **grep live transcripts**
   `~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl` newest-first
   for the slug — a session with many focused hits owns it uncommitted
   (188 showed 8 hits, 181 showed 14; directory-listing noise is ~1-2 hits).
3. CLAIM EARLY with a visible commit into the bug file ("CLAIMED <date>,
   workstream dir, do not compete") — who-owns reads commits, so this is
   what makes you visible to the next picker.
4. Re-verify the bug against the live DB before designing (187 had GROWN;
   its revalidator claim was FALSE — a registry coverage claim is answered
   by the registry, one grep).
5. 090 the root cause BEFORE asserting it (2026-07-31 ruling) — budget
   ~30 min; UNVERIFIABLE-on-tooling refutes nothing if your first-hand
   substitute is stated (the stale code index makes the loop's code reads
   fail on new symbols — known landmine).
6. Council: submit before/alongside the commit; `plan` is an OBJECT
   (summary/edits/grounded_in/risks), edits carry `symbol`, new file =
   `add` not `create`. Commit with `Council-Submitted:`, upgrade to
   `Council-Reviewed:` only after READING an approved verdict.
7. Before a chassis roll: check for in-flight council rounds (a roll kills
   them) — SQL in this dir's RUNBOOK. After ANY roll: pod-grep positive AND
   negative, every replica. Before ANY deploy: **grep the IMAGE** — an
   image built hours after your commit can still lack it (v1.0.1247 did;
   RUNBOOK recipe; memory `grep-the-image-before-the-deploy`).

## Candidate state as of 2026-08-04 ~09:00 (STALE — re-run the checks)

Taken/owned at last look: 188, 181 (live sessions); 184-markdown, 186, 189
(active filing lanes); 178/179/183/185 (worked); 192 (filed+owned). Never
re-trust this list — transcripts and commits move hourly.

## Traps this session hit that the next one should not

- `ExtractActionInputs` treats a config STRING as a REFERENCE into collected
  data (bugs_open/042 family) — test params carry inputs in CollectedData.
- `date -u -d 'today HH:MM'` parses the -d input in LOCAL time — print a
  deadline before waiting on it (this fired a dispatch inside the 300s
  rebalance window; it survived, but by luck).
- /tmp is a 16G tmpfs shared with concurrent builds — do HEAD-archive test
  builds under $HOME/.cache and delete after.
- A compile-error mutation proves nothing — a mutation must COMPILE to show
  a test bites (`satisfiable && false`, not `false`).
- Opus weekly cap resets Aug 7 — until then, implementation agents on opus
  will die mid-task; plan for sonnet or inline work.
