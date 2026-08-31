# HANDOFF 2026-08-26b — bugs_open/390: design proven end-to-end on three sites in one day; grade cookly, grade P4 formally, then P3 closes it

**Cold-start:** read this, then `NOTES_cascade_attribution.md` §(m)–(u) (append-only, newest at
the bottom — today's entries are the evidence trail), then `RUNBOOK_cascade_attribution.md` §8
(the post-roll verification commands, each with its gotcha). The previous handoff
(`HANDOFF_2026-08-26_continue_here.md`) covers the design and the trail up to this morning; its
§1(b) accounting query is WRONG (see its appended correction — the key is
`collected_data->'write_findings'`, never `'audit_findings'`).
Bug file: `bugs_open/390_HANDOFF_2026-08-25_a_correct_contrast_selector_still_loses_the_cascade_so_the_repair_is_authored_and_inert.md`.

## 0. STATE [MEASURED 2026-08-26 ~21:00 UTC — re-verify anything you act on]

| piece | state | evidence |
|---|---|---|
| migs 616 + 635 + commit 2 `ea64845e0` | LIVE + council APPROVED (`ef5f9a0d` r2 · `fe5cbe0c` r2 · `058b59b6`) | 098 credits all trailers "by correlation, via submitted" |
| **mig 655** (one !important instruction, never two) | **APPLIED 18:57:48Z + APPROVED r1** (`ffd6952b`, 2 medium advisories, dispositions in NOTES §(t)) | commit `1e513f5c9`, `Council-Submitted:` trailer — 098 resolves it |
| fresh fleet roll (~20:25Z) | **verified per service**: stamp `b34c24f4`; chassis binary-probed (present) + old-stamp & deadbeef controls (absent); `ea64845e0` still an ancestor | NOTES §(u) |
| P2 (repair meets requirement) | **CONFIRMED** — 6/6 pre-fence !important (the defect), then 8/8 post-fence correct + the 1 unattributed row correctly keeping !important (both fence branches live-proven) | NOTES §(o)(s)(t); llm_call_log split at 18:57:48Z |
| P4 (attribution distribution) | confirmed-so-far: **38 attributed / 6 unattributed / 0 unreachable** across remortgage+garden-tools+cookly; **formal grading = tomorrow's three named sites** | NOTES §(n)(r) |
| P3 (retraction, THE close criterion) | clocks running — see §1(c) | |
| cookly's 5 repairs | **still `triaged`, queue-bound, UNGRADED** | NOTES §(u) |
| remortgage theme | v14, 5 rules, served=git=DB sha | NOTES §(p) |
| garden-tools theme | 13 rules served (5 pre-fence w/ !important, 8 post-fence without), 6.81:1 spot check | NOTES §(t) |

## 1. NEXT ACTIONS, in order

### (a) Grade cookly's 5 repairs (first thing — they were queue-bound at handoff)

All five will render post-655. Expected: NO `!important` on any (all are `theme` with
`needs_important:false`), selector meeting each row's `override_requirement`. Grade with
RUNBOOK §8c (served=git=DB sha triple + own arithmetic) and §8d (fence check per row):

```sql
SELECT left(l.work_item_id::text,8) item, l.created_at,
       position('mark ONLY the single property' in l.prompt_rendered) > 0 AS general_rendered,
       w.status, w.result->'response'->'css_fix'->'result'->>'css_added' AS css_added
FROM site_work_items w JOIN sites s ON s.id=w.site_id
LEFT JOIN llm_call_log l ON l.work_item_id=w.id AND l.step_name='plan_css_fix'
WHERE s.domain='cookly.uk' AND w.item_type='contrast_failure' AND w.created_at > '2026-08-26 18:00+00'
ORDER BY l.created_at NULLS LAST;
```
A row STILL open after ~12h is a finding (the loop ran ~1 site/min all evening); check
`build-pipeline-trigger` + per-site loop activity before assuming (NOTES §(u) has both queries).
An `!important` REAPPEARING post-fence is a real finding, not noise.

### (b) Grade P4 formally — tomorrow's audits of the sites the prediction NAMED

Due (from `last_selected_at + 3 days`, hourly ticks at ~:50): **vonc.com 10:27Z · noted.co.uk
12:28Z · loanzy.uk 15:29Z** (loancash 17:30Z; cv1 08-28 12:39Z; agritec — P5 — 08-27 20:31Z).
P4 passes if `repair_surface='theme'` + `strictly_greater` dominate **per distinct
(selector,fg,bg) pairing, never per row** (chrome repeats per page and over-weights). Expect
inherited-colour footers to land `unattributed` — that is the DESIGNED under-claim (NOTES §(n)),
not probe blindness; the discriminators are `opaque_sheets`, `cascade_dirty_pages`,
`cascade_unverified_by_probe` in `collected_data->'write_findings'`. P5 (agritec): its 5
completed pairings should RE-FILE (the 396 erasure) — new rows, same selector+page.

### (c) Grade P3 — THE close criterion — at the 08-29 re-audits

remortgagecalculator ~13:50Z · garden-tools ~17:50Z · cookly ~18:55Z (first hourly tick after
each site's `due_at`). PASS: today's repaired pairings are **RETRACTED**
(`result->>'resolved_by'='render_audit'`), not re-filed. A re-filing byte-identical in fg AND bg
is the disconfirming result and reopens the design. **Before concluding a repair "didn't take",
check `css_themes.updated_at`** — a design run erases appended rules (bugs_open/396), which fails
P3 for a different reason and must be attributed to 396, not 390. Note garden-tools'
`P.tool-description`: a pre-fence `!important` rule coexists with 3 post-fence plain ones — the
!important one wins; expected retraction regardless.

### (d) Then close 390

Bar: fixed AND live AND proven at the artefact — P3 read on at least remortgage + one other site.
Steps (from the previous handoff §1(e)): move the bug file to `bugs_closed/` (name BOTH paths on
the commit; verify at HEAD with `git ls-tree`), fortnight park census
(`result->>'parked_by'='css_cascade_unreachable_390'` — expect SMALL), write the lane's first
SUMMARY (five headings, new file), 016b §9 entry with BOTH transferable patterns:
*"a repair authored at a surface that cannot govern the artefact completes honestly and does
nothing — assert the outcome, not the method"* and *"two co-present prompt instructions are
adjudicated by the model, not by precedence language — fence them so only one renders"*.
Update `MEMORY_workstreams` (lane entry added 08-26) and the VIZ-018 register status.

## 2. OPEN ITEMS THIS LANE HOLDS (not close-blocking)

- **bug_historian's 655 advisory, real and UNOWNED:** sweep other `agent_definitions` prompts for
  the same shape (a computed specific instruction co-present with a later generic one). Not
  390's criterion; needs a taker.
- **N pages × one chrome pairing ⇒ N redundant rules** (4 footer rules on remortgage, 4
  `P.tool-description` on garden-tools). Structural home = the deferred completion gate
  (owner 08-25, 395 lane's plan is the template). Recorded NOTES §(o)(p).
- **Inherited-colour under-attribution class** (NOTES §(n)): probe extension sketch = when no
  removal moves the value AND el's computed colour equals parent's, run the removal proof one
  level up. Own council round if built. Recorded in VIZ-018.
- **bugs_open/396** (design-run erasure) — filed by this lane, unowned, and it will EAT tonight's
  repairs at each site's next design run. P3's grading must check for it (see (c)).

## 3. TRAPS LEARNED OR RE-ARMED TODAY (all bit, none hypothetical)

- `collected_data->'audit_findings'` does not exist — accounting is under **`write_findings`**
  (the STEP name). A wrong key returns 0 rows for ever and reads as "no audit yet".
- The step config key is **`prompt_template`**, not `prompt` — same trap-shape, silently empty.
- **Run every "no rows yet" watch query against a known-positive first** (that control caught
  both of the above; WRONG_CALLS 08-26).
- `kubectl logs -l app=agent-chassis | grep 'build provenance'` now matches **landmine text about
  build provenance** in logged council payloads (1.9 MB of it). Per-pod logs +
  `grep '"build provenance"'` (structured field), or the binary probe with present+absent
  controls.
- `orchestration_states` retains ~1 day; **css-patch-agent child orchestrations are purged on
  completion** — grade repairs from `site_work_items.result.response.css_fix.result.css_added`,
  `llm_call_log` (by `work_item_id`), and the served file. The item's `result` is the SPAWN
  envelope (287's shape); `attempt_count` reads 0 at `complete`.
- The rotation pre_query **stamps in the same statement it selects** — read
  `last_selected_at + interval '3 days'`, never run the pre_query by hand.
- Key watcher deadlines to the **DB clock** (410 lane's skew entry rides build `b34c24f4`; my
  cookly watcher deadlined early unexplained — NOTES §(u)).
- Drift anchors for any future css-patch prompt migration are **655's shape** (fenced 616 passage
  + shortened else-sentence), not 635's, not 616's.
- Cloning an `agent_definitions` row for a mutation proof needs `version+1` (unique
  `(type,version)`).

## 4. THE FULL TRAIL (today's additions in bold)

- Commits: `3956adc06` · `c441b3b8f` · `26572c627` · `dc636bd6c` · `536906837` · `ea64845e0` ·
  `0b796c39d` · **`513f693c1` (docs+correction) · `0937f0124` · `bcbbac78d` · `ab893f374` ·
  `461e222e2` (RUNBOOK §8) · `ee8f078ae` · `0d36c9c27` (VIZ-018) · `13afd947a` · `1e513f5c9`
  (mig 655) · `8fef7830d` · `09efd16c6` · `593fa660d` · plus this handoff's commit**
- Council: `ef5f9a0d` (616, r2 APPROVED) · `058b59b6` (commit 2, APPROVED) · `fe5cbe0c` (635, r2
  APPROVED) · **`ffd6952b` (655, r1 APPROVED)** — all four credited by 098.
- Register: **VIZ-018** (status updated 08-26: LIVE + EXERCISED, two measured limitations).
- Fleet: stamps `2fb40a96` (08-25 roll) → **`b34c24f4` (08-26 ~20:25Z roll, verified)**.
- Watchers: all session-local and DEAD with the previous session — re-arm from RUNBOOK §8 /
  NOTES §(u) queries if wanted; none is load-bearing (the calendar in §1 is).

---

## ⚠ DATED UPDATE 2026-08-27 ~19:00 UTC — §1(a) and §1(b) are DONE; ONLY §1(c) P3 REMAINS (Friday 08-29); a new bug (416) was found and filed on the way

- **§1(a) DONE:** cookly's 5 repairs graded clean (NOTES §(w)) — 13/13 post-fence without
  `!important`, every selector strictly above its requirement, served=DB sha, 6.27/6.55:1.
- **§1(b) DONE, with a recorded deviation:** vonc and loanzy audits TIMED OUT —
  **`bugs_open/416`** (pre-existing ≥2-week defect: every ≥~25-page site's audit dies at
  `DefaultRequestTimeout=180` while the adapter finishes late; mechanism first-hand verified at
  `timeout_helpers.go:18` + `ConvertStepTimeout`; 090 ran twice — provider-cap death, then
  UNVERIFIABLE on the finished-conclusion shape — substitution declared per the 07-31 ruling).
  **P4: CONFIRMED on the measurable population** — noted 3:2 by pairing + six informal sites,
  ~60 attributed pairings, 0 unreachable, zero probe-blindness anywhere; the named-site
  substitution is recorded as caused-by-416 in NOTES §(z). noted's 16 repairs discriminate 16/16
  at the fence.
- **416's fix is UNCLAIMED by this lane** (deliberate; §5 of the bug file). The config-only
  interim (`timeout_seconds` on the `audit` step) is verified viable; one unknown remains for a
  taker (TimeoutMonitor/topic behaviour on ~25-min awaits).
- **§1(c) P3 is unthreatened**: Friday's three sites are 5/14/15 pages, all under 416's ceiling.
  remortgage ~13:50Z · garden-tools ~17:50Z · cookly ~18:55Z. Retraction of the repaired
  pairings = close 390 per §1(d).

---

## ⚠ DATED UPDATE 2026-08-31 — §1(c) P3 GRADED: PASS on all three sites; §1(d) EXECUTED — **390 IS CLOSED. This lane is finished; no next session is owed.**

- **P3 PASS** (evidence: NOTES §(aa)): all three 08-29 audits fired on their predicted ticks;
  **zero re-filings of any repaired pairing anywhere**; 396 confound ruled out (themes
  untouched, served=DB re-verified 08-31); remortgage's audit positively proven complete (it
  filed 2 rows on a NEW pairing the same minute); garden-tools/cookly proven by
  error-channel bracket + size class ([INFERRED] on `pages_audited`, which is purged after
  ~1 day — grade older audits from `site_work_items ∪ archive` + `agent_error_log`, never
  expect the orchestration row).
- **Criterion clarification for any future reader:** retraction stamps only rows not already
  settled. Rows already `complete` (repaired) pass P3 by *not being re-filed* — expecting
  `resolved_by='render_audit'` on them misreads the mechanism (NOTES §(aa)).
- Park census: **0 rows ever** parked `css_cascade_unreachable_390` (as of 08-31; literal +
  demand controls in NOTES §(aa)).
- Bug file moved to `bugs_closed/390_…`; first lane SUMMARY written
  (`SUMMARY_2026-08-31_cascade_attribution.md`); 016b §9 entry added (both transferable
  patterns); VIZ-018 register status updated; MEMORY_workstreams updated.
- Open items from §2 remain UNOWNED and live where recorded: bug_historian's
  prompt-contradiction sweep (NOTES §(t)), the N-pages-per-pairing redundancy class (§(o)(p),
  structural home = deferred completion gate), the inherited-colour probe extension (VIZ-018),
  and **`bugs_open/396`** (design-run erasure — will eat these repairs at each site's next
  design run; unowned) plus **`bugs_open/416`** (13 more big-site audits timed out over the
  weekend; interim fix spelled out in the bug file; unowned).
