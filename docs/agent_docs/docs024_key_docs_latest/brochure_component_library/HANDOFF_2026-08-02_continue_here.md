# HANDOFF 2026-08-02 — brochure component library / fundamentallyai.com

**Supersedes `HANDOFF_2026-07-30b_continue_here.md`.** Read this one first; that one
is still accurate about everything it describes except TL-035's status.

---

## 1. Where this lane actually is, in one paragraph

TL-035 — *photograph a page that PASSES, not only one that fails* — is **live and PROVEN
at the artefact.** The adapter half went live 07-31; the caller half went live tonight on
**v1.0.1229**; the flag was set by seed `292` in the ordered position. At **19:22** a real
acceptance run on `tool-review-council-simulator` passed 22/22 checks and filed **two**
renders (desktop + mobile) as durable `s3://` URIs on its `acceptance-run` note. The
control is in the data: the *same tool's* 07-31 note has no render line. **What remains is
not engineering — it is that nobody looks.** See §5.

> **CORRECTED 19:30 — this section previously said "armed end to end and not yet
> demonstrated", and §2 was a decision table for an outcome that had not landed.** It was
> written at 19:10 while the run was still queued and was accurate then. The run completed
> twenty minutes later. Left visible rather than silently rewritten, because a cold-start
> doc that quietly changes its own claims teaches the next reader to distrust all of them.

## 2. The artefact proof, and how to re-check it

```sql
SELECT created_at, subject_key, body LIKE '%Rendered:%' AS has_render_line, length(body)
  FROM doc_notes WHERE categories ? 'acceptance-run' ORDER BY created_at DESC LIMIT 4;
--  2026-08-02 19:22 | tool-review-council-simulator  | t | 2176   <- armed
--  2026-07-31 18:08 | tool-ai-vendor-trust-checklist | f | 1008
--  2026-07-31 12:59 | tool-review-council-simulator  | f | 1653   <- SAME TOOL, before
--  2026-07-29 20:51 | smart-contrast                 | f |  521
```

**Two limits, both stated in the register and worth repeating here.**

- **The PNGs were never fetched.** The bucket is private and returns **401 for a key that
  does not exist exactly as for one that does** — verified with a deliberate nonsense key,
  which is the only reason we know the check is worthless rather than reading 401 as
  "present but protected". What stands behind the URI is `screenshots.go:48-51` (`Save`
  returns `("", "", err)` on upload failure) and `extractShotList` (drops any ref with no
  durable URI). Sound, but an inference. Marked `[UNFETCHED]`.
- **If you ever see a note WITHOUT a render line**, do not assume the flag broke. Read
  `collected_data->'request_run'->'response'`: if the payload shows `capture_renders: true`
  the chassis did its half and the fault is downstream. And a run that FAILED its checks
  legitimately has no render for the failing profile — check the verdict first.

> **CORRECTED 2026-08-03 — the query above is RIGHT but INCOMPLETE, and the very next
> run tripped over the gap.** On picking the lane up I ran it and the top row was a
> render-less note *newer* than the armed one (`teaser-reveal-panel`, 08-02 21:53). Both
> exits this section offers are **dead ends for that row**: the run **PASSED** 15/15, and
> the flag was still `true` in `agent_definitions`. The real cause is a **third
> possibility this section did not contemplate — a SECOND CALLER that was never armed.**
> `request_component_browser_run` (the `staged_component_build` lane's component
> acceptance) files into the *same* `acceptance-run` category with the *same*
> `created_by='tool-acceptance-agent'`, so **no column on the note can separate the two.**
> Add the caller to the check — the discriminator is on the orchestration row, not the note:
>
> ```sql
> SELECT o.workflow_plan #>> '{steps,request_run,action}'                   AS action,
>        o.workflow_plan #>> '{steps,request_run,config,capture_renders}'   AS flag
>   FROM orchestration_states o
>  WHERE o.created_at BETWEEN '<note ts>'::timestamptz - interval '5 min'
>                         AND '<note ts>'::timestamptz;
> ```
>
> A NULL `flag` next to a non-null `action` is an **unarmed caller, not a regression**.
> Nothing about TL-035 on the tool path is weakened by this — the 19:22 proof stands
> unchanged. What was wrong was my claim of *coverage*: I armed one of two call sites and
> the re-check could not see the other. Filed fleet-wide in `LANDMINES.md`; the other lane
> has been told (`staged_component_build/CONTRIB_2026-08-03_capture_renders_is_free_on_your_path.md`),
> as the owner ruling of 2026-07-29 §3 requires. **The blast radius of arming a shared
> helper by config is the helper's CALL SITES, not the config rows you edited:**
> `grep -n "dispatchBrowserRun(" platform/orchestration/actions/tool_acceptance_actions.go`
> → `:184` and `:390`. That grep is the check I should have run on 08-02 and did not.

## 3. What was done tonight, and the two things that were done differently from plan

**Pod-verified before writing the key, both replicas, one exec each.** The ordering is
load-bearing (DB config is live instantly, Go is not) and it was honoured rather than
asserted:

```
capture_renders                 1   target
Rendered: full-page screenshot  1   target (added prose)
request_browser_run             8   POSITIVE control
.response.data.screenshots      0   NEGATIVE control
```

> **The negative control is the point, and the RUNBOOK's original instruction was
> weak.** It said to use `judge_acceptance_results` as a control — that is a *positive*
> control, and it reads identically on a binary built before the change. It proves the
> grep works, never that your code shipped. `.response.data.screenshots` was a real
> concatenated literal in the old binary, deleted when four hand-built path strings
> became one `envelopePaths(field, key)`, so **0 is only reachable post-change.**
> Derive candidates with `git diff <before>^ <after> -- <file> | grep '^-'`, then check
> each against HEAD before trusting it — `criteria_json` appeared in that removed-lines
> list and is still present (it moved), so it would have been a *false* negative control.

**Change 1: it went in as a numbered seed, not a bare `UPDATE`.**
`sql_for_agents/292_acceptance_runs_photograph_a_page_that_passes.sql`. A DB-only write
leaves a key in `default_config` with no provenance whatsoever — no who, no when, no
against-which-binary, no why. Two `DO`/`RAISE` guards; the second asserts a **neighbour**
key (seed 147's `profiles`) because a guard checking only what it just wrote cannot tell
a surgical `jsonb_set` from a write that flattened the whole `config` object. **Both were
induced before the real apply.** Watch for the self-satisfying mutant: `sed`-ing
`'true'::jsonb` globally flips the guard's own expectation too and passes.

**Change 2: it is numbered 292, not 291, and that was my error.** I applied by hand with
`psql` — correct, because `--apply` takes *every* pending file and 17 belong to other
threads — but a hand run writes no `schema_migrations` row, so **the number was never
claimed.** Another session recorded a different `291_` five minutes later. Now recorded
with `run-migrations.sh --record-only … --note '<what was verified>'`. Filed fleet-wide in
`LANDMINES.md`. The two `291:` strings in NOTES are left as printed — that is a
transcript, and tidying it would make it false.

## 4. Why the run is waiting, and why you must not speed it up

Two mechanisms, both working as designed:

1. **`tool_acceptance_due` has a 7-day cooldown per subject.** Every candidate ran 2–4
   days ago, so **nothing fires on its own** and a run must be queued by hand. The
   work-item shape is in the RUNBOOK under *"Dispatching an acceptance run by hand"*.
2. **`build-dispatch-loop` takes items fleet-wide, strictly by priority**, and acceptance
   runs are deliberately **priority 90** so they test the *new* page rather than one about
   to be replaced. At 19:02 there were **19** higher-priority items ahead — other lanes'
   repairs, including the `bugs_open/140` contact-info fix and the `178` content restore.
   It drained to 8 within ten minutes, so this is a wait of tens of minutes, not days.

**Do not raise the priority to jump the queue.** Those are repairs to pages serving wrong
content to real visitors. A verification run is not more urgent than that, and the
ordering is the design working.

## 5. Open, in the order I would take them

1. **NOBODY LOOKS AT THE RENDERS. This is now the top item and it is an owner call.**
   They land as `s3://` URIs inside a technical note — no page, no digest, nothing that
   puts an image in front of a person. A photograph nobody opens is worth the same as no
   photograph, so on the original problem (faults only an eye catches) we have closed the
   half that needs machinery and none of the half that needs eyes. Options range from a
   line in a weekly digest to a contact sheet of the last N passing runs to a review page.
   **Editorial, not technical** — do not pick one unilaterally.
2. ~~The lane SUMMARY~~ — **written**: `SUMMARY_2026-08-02_the_camera_works_and_nobody_looks_yet.md`.
3. **Open design question on the camera itself:** a full-page shot at one width is not what
   a mobile reviewer needs. Should `Renders` carry its viewport? (register `verify-later` (b)).
4. **Not chosen by the owner but still open from the 07-30b handoff:** `llm-cost-calculator`'s
   two dead CTA blocks; the `/tools/decision-record/index.html` 404 stub and the missing
   `/tools.html` (both stale rows, nothing links to them); a guide page for the
   review-council simulator; a `tool-guide-intro` section (note: a `tool-guide-intro`
   component now appears on `llm-cost-calculator` — check what it is before rebuilding it).

## 6. Things that will mislead you on this lane

- **A roll is not evidence your fix shipped.** Grep a string your change ADDED *and* one
  it REMOVED, on **every** replica, same exec. `strings` is absent from several of these
  images — use `grep -acF` on the binary directly, or you get a confident 0 for target and
  control alike.
- **A render is a look, never a verdict.** `Renders` carries no `failing_checks` by
  construction. If anything starts branching on a render's presence, the two-list design
  has been undone and the whole point is lost.
- **`Renders` is on the FAILED note too, and that is not a slip.** The adapter captures per
  (url, profile), so a two-profile run where desktop fails and mobile passes files a shot
  **and** a render.
- **A verify block of `SELECT`s cannot stop a `COMMIT`.** Use `DO`/`RAISE`, and induce it —
  a guard that has never fired is decoration.
- **Check `git log` before assuming HEAD is yours**, and `git ls-files --error-unmatch <path>`
  before reaching for `add`: `git commit <path>` reads the working tree and ignores the
  index, so a *tracked* file never needs `add`, and adding one leaves it in the shared index
  for another session's bare commit to collect. This lane has been bitten by both halves of
  that in one day.

## 7. Commits

- `97dacc5e8` — arm(TL-035): seed 292, pod verification, RUNBOOK/NOTES/README/register
- `02990d88d` — landmine: a migration number is yours when the LEDGER says so

Earlier, for context: `9cc63c775` (caller half), `72463f51e` (r1 objection fix),
`606c5aa3a` (register), `3162f0271` (lane docs), `62e72ff77` (wrong call).
Council `2c895dd1-adae-4f8e-8acf-4592b8ca3981` — APPROVED r1, 11 approve / 1 object; the
one medium objection changed the code and is written up in
`EVIDENCE_2026-07-31b_TL-035_caller_half.md`.
