# HANDOFF 2026-08-26 — bugs_open/390: all three commits live; VERIFY AT THE ARTEFACT, then close

**Cold-start:** read this, then `NOTES_cascade_attribution.md` (append-only, newest at the bottom
— the missteps are the point), then `PLAN_2026-08-25_cascade_attribution.md` for the design
decisions. The bug file is
`bugs_open/390_HANDOFF_2026-08-25_a_correct_contrast_selector_still_loses_the_cascade_so_the_repair_is_authored_and_inert.md`
(its two APPENDED sections carry the corrected mechanism — its own §1 worked case is WRONG, it
parks rather than completes; the real worked case is vonc.com).

---

## 0. STATE in one table [MEASURED 2026-08-26 ~09:00 BST — re-verify, this tree moves in minutes]

| piece | state | evidence |
|---|---|---|
| mig 616 (prompt correction) | **APPLIED + council APPROVED** round 2, corr `ef5f9a0d` | live row: old bullet absent, correction at 1109 |
| commit 2 `ea64845e0` (cascade attribution, Go, both images) | **LIVE + council APPROVED**, corr `058b59b6` | stamp `2fb40a96` on all 3 services; ancestry YES with HEAD as NO-control; chassis binary-probed with absent-sha control |
| mig 635 (gate + park + prompt block) | **APPLIED 08-26, verified at the row**; council verdict **PENDING**, corr `fe5cbe0c` | gate=`check_repair_surface`, park stamps `needs_human_review`/`css_cascade_unreachable_390` before `complete_refused`, block at 133, 0 dangling edges |
| exercised by a real audit? | **NO — zero audits since the roll** | rotation ticked 08-26 07:47, no site past its 3-day window |
| bugs_open/396 (erasure, found by this lane) | FILED, unowned | its §6a records why 616 *raises* its stakes |

**The bug is closed on paper, not at the artefact. Do not close 390 yet** — the bar is fixed AND
live AND *proven at the served page*, and no cascade-attributed row exists anywhere yet.

## 1. NEXT ACTIONS, in order

### (a) Read the 635 council verdict — OWED, verdict was pending at handoff time

```sql
SELECT metadata->>'decision', created_at FROM diagnosis_artifacts
WHERE correlation_id='fe5cbe0c-8a04-47a7-b64d-68fc235da14d' AND kind='council_report' ORDER BY created_at;
-- the human-readable note:
SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%fe5cbe0c%' ORDER BY created_at DESC LIMIT 1;
```
On REVISE: 635 is **already applied** — fix in place and resubmit with `RESUBMIT_CORR=fe5cbe0c`,
exactly as 616's round 2 did (that trail is the worked example, NOTES §(j)). ⚠ Remember the
sequencing lesson: the seats' read-only checks observe the POST-apply world; say so in the
resubmission so their `has_X: false` answers are not read as drift. Same for `ef5f9a0d` (616) and
`058b59b6` — both APPROVED, verdicts read; nothing owed there.

### (b) When the first post-roll render audit fires (expected by ~08-29) — grade predictions P2–P4

The rotation re-audits each site every **3 days** (read `scheduled_tasks` where
`name='site-render-audit-rotation'`, never WII-016's stale "7 days"). Watch for it:

```sql
-- new attributed findings (P4):
SELECT s.domain, w.status, w.spec->>'repair_surface', w.spec->'override_requirement'->>'min_specificity_text',
       w.spec->'override_requirement'->>'strictly_greater', w.created_at
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='contrast_failure' AND w.spec ? 'cascade_scheme' ORDER BY w.created_at DESC;
-- the producer's own accounting (cascade_scheme_present distinguishes "cannot attribute" from "attributed nothing"):
SELECT collected_data->'audit_findings' FROM orchestration_states
WHERE collected_data->'audit_findings' ? 'cascade_attributed' ORDER BY updated_at DESC LIMIT 3;
```

**P4 (from NOTES §(g)):** `repair_surface='theme'` with `strictly_greater=true` on the large
majority; `unreachable` rare. **Disconfirming:** `unattributed` dominating — that means the
remove-and-remeasure proof is rejecting its own attributions (probe blind), NOT that pages are
odd. If so: read `cascade_unverified_by_probe` / `cascade_dirty_pages` in the result map first.

**P2:** the next repair on a site passing 542's gate (theme ≥4096 bytes, unshared: loanzy, vonc,
noted qualify) appends a rule meeting the requirement — for pre-attribution rows, `!important` on
exactly one property. Curl the served stylesheet.

**P3 — the decisive one:** that pairing is **RETRACTED** at the following audit
(`result ? 'resolved_at'`, `resolved_by='render_audit'`), not re-filed. A re-filing with
byte-identical `fg`/`bg` is the disconfirming result and reopens the design.

### (c) Grade P1 when loancash's parked rows or cv1's detected rows move
cv1's 3 `detected` rows should park at `css_base_integrity_guard_198` (theme is a 1,649-byte
shared seed — RUNBOOK §3 has the gate-vs-row query; getting this wrong is the WRONG_CALLS misstep).

### (d) Census the new park after a fortnight
```sql
SELECT count(*), min(updated_at) FROM site_work_items WHERE result->>'parked_by'='css_cascade_unreachable_390';
```
Expected SMALL (important-inline was 0/40 in the sampled completions). Self-draining by design
(status in neither terminal nor closed list → dedup slot held, WII-016 retraction still closes it).

### (e) Then close 390
Bar: fixed AND live AND proven at the artefact (P3+P4 read). Move the file to `bugs_closed/`,
update `MEMORY_workstreams` if the lane entry exists, write the lane's first SUMMARY (it has
none — a genuine milestone), and consider a 016b §9 entry for the transferable pattern
(*"a repair authored at a surface that cannot govern the artefact completes honestly and does
nothing — assert the outcome, not the method"*).

## 2. THE DESIGN IN FOUR SENTENCES (for whoever grades a dispute)

The audit **attributes** the declaration that decides a failing element's colour and **proves it
by removal** — removing a losing declaration cannot change the computed value, so if the value
moves, the removed one WAS the winner; unprovable ⇒ `verified:false` ⇒ routed `unattributed`,
never a weak yes. The filer computes the winning selector's specificity **with cascadia on the
chassis** (one authority; unparseable ⇒ error, never a zero triple) and writes
`repair_surface` / `winning_rule` / `override_requirement` — gated on the adapter's
`cascade_scheme` declaration so an old adapter's spec is byte-identical to before. The agent
renders the requirement into its prompt (fenced on `override_requirement`, NOT `winning_rule` —
sibling-absent dereference is an execute-time template error) and **parks** the one unbeatable
class (important inline `style=`) before LLM spend. `item_key`, `handler_agent` and the
retraction path are deliberately untouched (VIZ-016's alias-key history is why).

## 3. LANDMINES SPECIFIC TO CONTINUING THIS

- **The bug file's §1 worked case does not demonstrate the bug** (loancash parks). Use vonc.
- **Drift anchors for any future prompt migration are 635's shape now**, not 616's, not 318's.
- **A literal written down twice WILL disagree with itself** — twice in one file here (the `/219`
  and the duplicated `$old$`). One DO block, literals as variables.
- **Applying before the verdict poisons the seats' own checks** — they see the post-apply world.
  If you must, say so in the submission.
- **`agent_snapshots` is a VIEW with `snapshot_taken`**, not `created_at`.
- **Migration numbers go stale within hours** (617–634 taken overnight). Check at write time.
- **css-patch-agent has ONE active row today; four other types have TWO.** Every new migration on
  `agent_definitions` needs the row-count assertion (616-round-2's HIGH objection).
- **`{{if}}`-fence prompt blocks on the field that is always present with its siblings** — see §2.
- The **erasure** mechanism (396) will eat theme-appended repairs at the next design run — when
  grading P3, check `css_themes.updated_at` before concluding the repair "didn't take".

## 4. WHAT THIS LANE DID NOT DO (deliberate, with owners)

- **Completion gate** ("complete only on measured improvement") — owner deferred 08-25, routing
  first. Composes with the committed gate 1b/1c family; the 395 lane's plan is the template.
- **Historic rows** — owner: leave them; the 307 completes / 97 byte-identical re-filings stay as
  evidence.
- **Token/palette repair** — palette lane's; 390 §2 says never blind.
- **bugs_open/396** (erasure) — filed, unowned, waiting for a taker. §5 says the git history of
  `assets/css/styles.css` in the sites repo is the instrument that can size it.

## 5. Correlations, commits, artefacts — the full trail

- Commits: `3956adc06` (lane opens) · `c441b3b8f` (616) · `26572c627` (docs+396) · `dc636bd6c`
  (616 round-2 hardening) · `536906837` (owner log) · `ea64845e0` (commit 2, Go) · `0b796c39d` (635).
- Council: `ef5f9a0d` (616, REVISE→**APPROVED**) · `058b59b6` (commit 2, **APPROVED**) ·
  `fe5cbe0c` (635, **PENDING** — action (a)).
- Register: **VIZ-018** (+ index row; status updated 08-26). Related: VIZ-016, WII-016.
- Fleet: stamp `2fb40a96`, rolled 08-25 ~23:11.
- Docs debt: none known; `README_where_we_are.md` needs an 08-26 entry if the owner asks before
  the next session writes one.

---

## ⚠ DATED CORRECTION 2026-08-26 ~09:15 BST — §1(a) said "verdict PENDING"; the run actually DIED, and the cause is the PROVIDER, not the submission

Checked before signing off rather than leaving "pending" to be believed: the `fe5cbe0c` run
terminated at `complete_invalid` — *"no reviewer produced a readable opinion (6 abstained, 11
unreadable)"*. Eleven seats unreadable at once is systemic, and the cause is on the record:
`agent_error_log` carries fleet-wide `LLM_API_ERROR … AI endpoint unavailable: provider=anthropic`
at 08:46–08:48, exactly while the seats executed. (The `bugfix_243_provider_cap_resilience` lane is
active in the tree — related territory.) The last healthy council verdict fleet-wide is 21:38 on
08-25; mine is the only run since.

**So §1(a) becomes: (a-0) FIRST check the provider is back** (any successful `council_report` after
08-26 09:00, or a clean `execute_llm_prompt` in recent logs), **then resubmit 635 on the SAME
correlation**:

```bash
RESUBMIT_CORR=fe5cbe0c-8a04-47a7-b64d-68fc235da14d \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_390_cascade_attribution/council_submission_635_2026-08-26.json
```

An invalid run produced no verdict, so this is a retry of infrastructure, not a defence of a
REVISE — the submission JSON needs no edit (unless the provider outage outlasts the day; then
note the resubmission delay in it). The commit `0b796c39d` carries `Council-Submitted: fe5cbe0c`,
which stays TRUE — it asserts submission, not a verdict — and 098 resolves it once a real verdict
lands on that correlation. Do NOT dispatch a duplicate on a fresh correlation: that is the
double-round the runbook warns about.

---

## ⚠ DATED UPDATE 2026-08-26 ~11:20 BST — §1(a) is DONE: 635 is APPROVED, and the whole 390 trail is now council-clean

The resubmission on `fe5cbe0c` (after the provider outage) returned **APPROVED at 09:14 UTC with
ZERO objections** (6 abstained). All three 390 commits now carry approved verdicts:
616 (`ef5f9a0d`, r2) · attribution (`058b59b6`) · 635 (`fe5cbe0c`, r2 after an infrastructure-dead
r1). 098 credits the `Council-Submitted:` trailers automatically; no amends.

**The ONLY remaining work on 390 is §1(b)–(e): verification at the artefact.** First post-roll
audit expected at the first hourly rotation tick after ~14:16 BST today (remortgagecalculator.uk
due 13:16 UTC; then garden-tools 17:18, cookly 18:19 UTC). P1 was graded — the observation
confirms 542's gate but it was a POSTDICTION (event preceded registration by 2h; WRONG_CALLS
2026-08-26); P2–P4 remain genuine and pending.

---

## ⚠ DATED CORRECTION 2026-08-26 ~12:35 BST — §1(b)'s accounting query names a key that never exists; the counters land under `write_findings`

The producer-accounting query in §1(b) reads `collected_data->'audit_findings'`. **There is no such
key.** The cascade counters are written into the ACTION RESULT map in
`write_render_audit_findings_action.go:642-651`, and that result lands in `collected_data` under
the STEP name — `write_findings` (verified against all 3 existing `render-audit-agent`
orchestrations: `collected_data ? 'audit_findings'` is false on every one, and the cv1 run's
accounting — `inserted: 3`, `deduped: 5` — sits at `collected_data->'write_findings'`). The
query §1(b) gives would return zero rows for ever and read as "no audit has run yet". Correct form:

```sql
SELECT jsonb_pretty(collected_data->'write_findings') FROM orchestration_states
WHERE collected_data->'write_findings' ? 'cascade_attributed' ORDER BY updated_at DESC LIMIT 3;
```

The keys are written unconditionally, zeros included (the code comment at :634 says so and the
code does), so **presence of `cascade_attributed` = post-roll audit; absence = old code or no
audit** — with `cascade_scheme_present` as the discriminator between "cannot attribute" and
"attributed nothing".
