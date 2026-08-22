# HANDOFF — 2026-08-21. `277`'s clause 1 is MET (7 pages repaired, proven at the served bytes), the owed `diagnosis_guardian` fix is APPLIED, and `083`'s soak is at day 4 of ~7

**Supersedes `HANDOFF_2026-08-20b_continue_here.md`.** That file's §2 sequence is now HISTORY — all
four steps ran today. Read this from disk; then `NOTES_required_fields_repair.md` **from the
bottom** (two new entries today, §§1–7 the canary and §8 the seat fix).

> **Written 2026-08-21 ~15:00Z; re-verified after TWO further rolls, latest 2026-08-22 ~09:15Z.
> Deploy facts have a shelf life of hours — re-read, do not quote.**
> Chassis is now **`70e7b4f9c`** (`sha256:b83dc450…`, pods up 2026-08-22 **08:36Z**); it replaced
> `bac189921`/`sha256:68075cf5…` (08-21 16:54Z), which replaced `0483e7f4e`/`sha256:3ed50651…`.
> Each is forward with no revert, and `af0f00bb5` is an ancestor of all three — but **ancestry is not
> the check**, because a commit AFTER mine could delete the code and still leave mine an ancestor.
> Re-probed on the newest binary: `rendered_html_transform` **8**, `code_span_to_code_tag` **5**,
> negative control **0** (grep exit 1). **Nothing in this lane needs another roll**, and three rolls
> have now passed over it without disturbing it.
>
> **The seven repairs still hold at the served bytes** [MEASURED 2026-08-22 09:16Z]: prose backticks
> **0** on cubic-bezier / head-architect / grid-generator, with **4 / 44 / 8** in-script backticks
> intact. That is the durability question answered for two roll boundaries, not just for the hour the
> repair ran.
> Live md5s after today's config changes: `fix-proposer.review_diagnosis_guardian`
> **`99bf2e45…`**, `council-gate.review_diagnosis_guardian` **`347a20cf…`**.

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | **clause 1 MET** — 7 pages repaired, verified at the served bytes (§7) | the **`no_content_data` half — now an OWNER DECISION, not a build task** (§8.4 of the bug file). ⚠ §8.2 CORRECTS this file's own "content-acquisition problem" framing: **15 of 27 are RECOVERY** — the value is already on the page, absent only from `content_data` |
| **`bugs_open/083`** | fix live + artefact-proven; **door soak at day 4 of ~7** | ~**08-24/25**: `444`/`458` sat their week. Close then, with the two statements in §3 below |
| `bugs_open/333` | theirs | **CONTRIB filed today** — the two-strike rule reaches their false *"tried twice"* down a second road, with no refusal loop involved |
| `530`/`531` (council) | **APPLIED + APPROVED r1** (corr `c00fbfd8…`, 14:03:49Z, `point_fix`, no truncation) | **nothing** — verdict read, both medium advisories answered with the queries they asked for (NOTES §10), trailer written. Done |

---

## 1. WHAT HAPPENED TODAY, in the order it mattered

1. **The roll had already happened** (08-20 19:51Z). Proven per service with both controls — see the
   banner. The `build provenance` **startup line was NOT readable**: a chassis pod's log history is
   ~90 seconds under load, and the three "hits" a grep found were the phrase quoted inside a council
   submission's `collected_data`. Use the `buildinfo.GitCommit=` probe of `/proc/1/exe`.
2. **The sweep the canary needed was FOUR DAYS away, and nothing said so.** `literal_markdown` is a
   *quality*-discovery check; only `site-discovery-rotation-quality` runs it, `LIMIT 1` site older
   than 7 days. Every one of the 25 stamped sites was inside its floor (oldest: 5d 01h), so the
   rotation was **idle** while its `last_triggered_at` advanced every 3h. Owner approved forcing it.
   A one-shot task (**no `pre_query`, so the rotation stamp was not consumed** — the natural 08-25
   slot survives) fired 13:19:01Z; the run was confirmed at `orchestration_states`, not at the stamp.
   Recipe + the "when would the rotation reach my site" query: **RUNBOOK**, "Force a discovery sweep
   for ONE site". Also filed as a LANDMINE.
3. **The router filed 8 rows and every one carried the new shape.** No other check filed anything.
4. **One canary promoted by hand 13:21:42Z → `complete` 13:25:03Z → the promoter released the other
   six on its very next tick (13:27) → 7 of 7 `complete`, all `verified`, by 13:37Z.**
5. **Proof at the served bytes on all seven**, with the control that carried the risk: prose
   backticks to **zero** everywhere while `tool-head-architect` kept all **44** of its in-script
   template-literal backticks. Per-page table in NOTES §7; the worked before/after in `277` §7.
6. **The owed `diagnosis_guardian` message became migrations `530`/`531`** — see §2.

---

## 2. `530`/`531` — APPLIED, and what a reviewer of this needs to know

The seat's prompt asserted the coordinator *"reads ONLY `step.config.error_step`"* and that a
step-level one is *"silently inert"*. `routeToErrorStepOrFail` (`coordinator.go:3666-3679`) checks
**step-level FIRST** — its own comment calls that the preferred location — with config-level as the
backward-compatibility fallback. The seat was objecting to authors who did the right thing.

While reading both rosters side by side: council-gate's copy carried
`## The author's stated rationale loop's load-bearing disciplines`, produced by
`099_SYNC_gate_roster.py:85`'s **unanchored** `replace("## The diagnosis", …)`. `531` repairs the
live text and the script's substitution is now line-anchored so it cannot recur. **`099`'s OTHER
defect is untouched: its transform predates `377`, so `--apply` stays SUSPENDED.**

Exercised three ways and round-tripped byte-identically before applying — the table is in NOTES §8.
⚠ **`531`'s marker-offset guard is structurally unfirable for this anchor set and its header says
so**; the guard that can fail is the fleet-wide `17 seats / 1 prefix` check, induced and confirmed.

---

## 3. `083` — the close, ~08-24/25, and the two things it MUST say

The mechanism is proven twice over now (today's arc is a second independent instance — bug file,
2026-08-21 entry §1). When the doors have sat their week:

1. **Move the file with BOTH paths on the commit** (`git commit bugs_open/083_… bugs_closed/083_… -m …`)
   and verify **at HEAD**, not at the tree:
   `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 083` → exactly one line
   ⚠ **two unrelated cases share the number 083** — resolve by slug.
2. **State that `479`'s reclaim arm has never fired** — re-measured today, **0 live / 0 archive, all
   history**, while the escalated population **grew to 7** (today's 12:59Z tick added
   `dead_fragment_link` and `missing_conversion_path`). The door opens on schedule and nothing has
   ever come back through it. Do not let the close imply otherwise.
3. And carry §3 of that entry: the **third stranding shape** (two-strike strikes inherited across a
   re-route) is invisible to all three of 083's own instruments, because the row is born terminal.

---

## 4. STILL OWED

- ~~Read the `c00fbfd8` verdict~~ **DONE — APPROVED round 1, advisories answered, NOTES §10. Nothing is owed on the council trail.**
- **`083` close, ~08-24/25** (§3).
- **`277`'s `no_content_data` half — the only thing holding that file open, and it is now ONE OWNER
  DECISION between three costed options** (bug file §8.4): (1) recover the field from the component's
  own `rendered_html` — deterministic for 15 of 27, but it re-enables the regenerate routes on
  exactly the components whose un-regenerability makes today's repair safe, so it owns re-checking
  the seven; (2) accept "parked with the facts" as the terminal state, in which case **277 CLOSES**
  and the residual becomes a data-model debt filed elsewhere; (3) build the finding-to-edit converter
  this file originally proposed — right for genuine acquisition, overkill for the 15.
  **A session must not pick between (1) and (2): option 2 changes what "fixed" means for this bug.**
- **`copy_edit_proposed` exclusion in the promoter's `pre_query`** (owner decision D2, 2026-08-12).
  Still deliberately not done by a session: it changes which rows are dispatched.
- **A disconfirmable prediction to check on ~08-25:** `learn-index`'s next filing should be born
  `detected`, not `unresolved`, because both of its strikes age out of the rolling 7-day window
  first. **Do not hand-flip the row** — the prediction is worth more than the one page.

## 4b. CAN THE LANE CLOSE? — the state of every open thread, 2026-08-21 17:10Z

**Not yet, and exactly two things stand between here and closing it.**

| thread | closeable? | why / what it needs |
|---|---|---|
| `bugs_open/083` | **YES, on ~08-24/25** | The fix is complete, artefact-proven, and demonstrated end to end twice. What remains is the owner's own decision-5 soak week on `444`/`458`, at day 4 today. Close per §3 — both paths on the commit, verified at HEAD, and the two statements it must make. Closing early is the owner's call, not a session's |
| `bugs_open/277` | **ONE OWNER DECISION AWAY, and the decision got easier on 08-22** | Routing delivered and approved; clause 1 met and proven at the served bytes. §8.5 measured the backfill hazard I had asserted and it is **ZERO** for this population (no overlap with the repaired pages). §8.6 measured this file's OWN criterion 2 at the served page and it is **MET** — the worked example serves its full table and every "missing" value. So the residual is a **data-model debt with no visitor-facing symptom**, not a broken page. §8.4's three options stand, with option 1 now cheaper and option 2 now better evidenced |
| `bugs_open/333` | not ours | Theirs. CONTRIB filed 08-21; nothing owed |
| `530`/`531` | **DONE** | Applied, APPROVED r1, advisories answered, trailer written |
| the 08-25 prediction | watch only | `learn-index` should be born `detected`, not `unresolved`. Do not hand-flip it |

So: after 083 closes and the owner answers §8.4, **this lane has no open work of its own** — what
would remain is whatever the owner picks from those three options, which is new work, not this
lane's backlog.

## 5. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `333` · the chassis stamp + `git merge-base --is-ancestor` for anything you think shipped ·
`SELECT status, count(*) FROM site_work_items WHERE item_type='literal_markdown' AND
handler_agent='section-editor' GROUP BY 1;` (expect 7 `complete`, 1 `unresolved` until the 08-25
sweep) · then §4.
