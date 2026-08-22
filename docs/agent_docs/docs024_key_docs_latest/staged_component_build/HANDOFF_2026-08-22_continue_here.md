# HANDOFF — 2026-08-22, fresh chat starts here: **the lane's deliverable is BUILT, LIVE and FIRING.** All four §1 items are done; what remains is a 48-hour gate (closes Sunday ~08:45Z) and two carried-out items that are no longer this lane's work.

**Supersedes `HANDOFF_2026-08-21_continue_here.md`** (whose §1 table is now all-green; keep it
for the step-1→4 evidence trail and its corrected §5 grounds).
**DOES NOT supersede `HANDOFF_2026-08-20_continue_here.md`** — that file is **the gate's home**
(§2.10 1b terms, §2.11 baseline + boundary + interpretation table + the interim reads). The gate
obligation is deliberately **ownership-independent**: whoever is alive runs it and records the
result THERE.

**Read in this order:** this file → `HANDOFF_2026-08-20_continue_here.md` §2.11 (gate) →
NOTES `## 2026-08-22` entries (five; the day's record) → `bugs_open/353` → `bugs_open/330` §§10–11.

---

## 1. STATE — the lane's four blockers are ALL closed

| # | what | state (as of 2026-08-22 18:0xZ) |
|---|---|---|
| **1** | **THE FLIP** — conflicting whole-tree search → refusal | 🟢 **LIVE and PROVEN FIRING.** `v1.0.1323` 08:37Z, re-verified on `v1.0.1326` 15:10Z. Council APPROVED r3 (`26186633`). **2 real refusals in production** (§3) |
| **2** | Read-side tolerance retirement | 🟢 **LIVE** same rolls. APPROVED r1 (`e05ea6f9`), commits `e5c1b3c15`+`9970eb71c` |
| **3** | `bugs_open/330` — the `related_pages` fix (migration 516) | 🟢 **RESOLVER HALF PROVEN BOTH DIRECTIONS** (absence 08-21, presence 08-22 by owner-approved positive control). Its remaining symptom is **`bugs_open/353`**, a different bug. Candidate 2 (269-pair remainder) still open, deliberately post-gate |
| **4** | A standing form of 537's guard | 🟢 **BUILT AND LIVE** as `WFA-022` (CronJob `commit-sha-exposure-check`, daily 06:45 UTC, `v1.0.1324`), by the parallel session. `bugs_open/334` **CLOSED** |

**So: the lane is CLOSEABLE after Sunday's gate read**, with two items carried out as
separately-owned work (§4). That is a change of state from every prior handoff, which all said
"not closeable".

---

## 2. THE ONE LIVE OBLIGATION — the ≥48 h gate, closes **2026-08-24 ~08:45Z**

Terms and interpretation table: **`HANDOFF_2026-08-20_continue_here.md` §2.10 1b + §2.11.**
Do not re-derive them; do not re-run a second window.

- **Boundary:** rows `< 08:35:00Z` 08-22 = old binary · `08:35–08:45Z` = **AMBIGUOUS, exclude**
  · `>= 08:45:00Z` = new binary. Attribute by `created_at`, **never wall clock** (a pre-change
  run kept emitting old behaviour for 8.5 min during 537's verification).
- **⚠ THE ZERO CANNOT PROVE THE FLIP.** The conflict table was already near-silent before the
  roll (537 and 516 killed both live classes hours earlier). **Never write "0 rows, therefore
  proven."** The behaviour claim rests on the 13 tests that fail on a one-line revert, plus the
  capability probe. The gate is for the two signals below.
- **`1-resolve-and-warn` post-boundary = REGRESSION** (a pod on pre-flip code) → probe the pods.
- **`2-refuse` = a real conflict** → trace the (agent,field) pair to its consumer and give it an
  explicit mapping. **Never revert the flip for one.**
- **⚠ A ROLL CROSSING THE WINDOW DOES NOT RESTART IT — but you must re-verify.** Done twice
  today. Method for a build whose provenance line has scrolled: **capability probe with BOTH
  controls** (`2-refuse` present · a known literal present, proving the probe reads · a
  synthetic literal absent). For the retirement (a pure deletion, no literal):
  `git log -S 'stepKey != key' --since=2026-08-21` must return only the removing commit, and
  HEAD's `v3_site_actions.go` must contain 0 occurrences.

```sql
-- the gate read, three queries, demand control FIRST
SELECT count(*) AS orchestrations, count(DISTINCT owner_agent_type) FROM orchestration_states
 WHERE created_at >= '2026-08-22 08:45:00Z';
SELECT context->>'phase' AS phase, error_code, context->>'field', agent_type, count(*)
  FROM agent_error_log WHERE error_code LIKE 'RESOLVER_%' AND occurred_at >= '2026-08-22 08:45:00Z'
 GROUP BY 1,2,3,4;
SELECT count(*) AS rows_any_class, count(DISTINCT error_code) FROM agent_error_log
 WHERE occurred_at >= '2026-08-22 08:45:00Z';   -- instrument alive
```

---

## 3. WHAT THE GATE HAS ALREADY FOUND (interim read, ~9¼ h in)

**2 × `2-refuse`, 0 × `1-resolve-and-warn`**, against 1,792 orchestrations / 64 agent types of
demand and a demonstrably live recorder (175 rows / 20 classes). **The flip is working in
production and there is no regression signal.**

Both refusals are `site-work-orchestrator` at 10:44:06Z, and both were **predicted** by the
08-20 handoff rather than novel:

1. **`result` — 11 candidate paths.** Eleven agents file output under one bare key
   (`content_writer_agent.result`, `spawn_handler.result`, `reviewer_agent.result`, …); the
   ranking would have taken the first. Refused. This is the canonical shape step 5 exists for.
2. **`commit_sha` — 4 paths**, all `*.response.data.commit_sha` (deploy js_snippets / logo and
   aliases). **537's collision class on a different agent.**

**The work each wants: an explicit `?` wire on `site-work-orchestrator`, as 537 did for bdl.**
⚠ **It is NOT a copy of 537** — the consuming steps are **dynamically generated**
(`fix_items_loop_iter_N_call_handler`), so they are invisible to the static step-config queries
537 used. Solving that addressing problem is the first real task for whoever picks it up.

**A coincident FAILED run was traced and EXCLUDED** (a 2026-04-19 `allow_reinstall`
re-install guard; no refusal row names any style field; the flip only alters *conflicting*
resolutions and every conflict writes a row). ⚠ **But its class shows 6 occurrences, all today,
all post-flip, against retention reaching 2026-07-19.** The benign workload explanation is
`[INFERRED]` and unmeasured — **re-check it at the close-out; do not inherit the conclusion.**

**Consequence control:** 312 of 371 completed items still record `result.commit_sha` since the
boundary — the field was not lost fleet-wide.

---

## 4. CARRIED OUT OF THIS LANE — named, and NOT this lane's to fix

1. **`bugs_open/353`** — new-tool cross-links are withheld at birth and nothing re-emits: 029's
   Guard 2 gates on a `needs_content_page` item that fix 177 (08-03) stopped raising for
   pure-tool pages, and `tool-deployer` has 0 runs in retained history. **32 withholdings / 32
   tools / ~24 domains as of 2026-08-22; 30 of 32 pages now deployed with zero crosslink items
   ever.** Fix is **UNOWNED** — claim it in the file. Ready-made repro + backfill census inside.
2. **The tool-birth refusal rate** — `tool_birth_instance_scope_refused` fired on **6 of 9
   births on 2026-08-22** vs 2 on all of 08-21. **Cleared of the roll** (prover byte-identical
   `bac189921..70e7b4f9c`); generation-side. Unfiled — file it if the rate holds.
3. **`bugs_open/330` candidate 2** — the 269-pair / 75-agent unsampled remainder. Deliberately
   deferred until after the gate **so its probe is designed against observed post-flip
   behaviour, not predicted** — and the gate has now produced exactly the observations it needs
   (§3). This one may legitimately come back to this lane.

---

## 5. TRAPS EARNED TODAY (all cost something)

1. **A control that differs from the treatment in TWO variables is not a weak control — it is
   not a control, and its failure mode is a PASS.** A peer's "free paired control" for 330 used
   two runs that differed in with/without-pages *and* in birth-vs-replace; the replace arm
   returns 270 lines before the emitter, so they could only ever confirm. Split on the arm
   first; `replace_existing` is now a fleet landmine.
2. **A bare bug number routes to the wrong owner.** I cited "029" in `bugs_open/353`; 029 is a
   documented ambiguous number and my notice reached the hung-spawns lane, not the crosslink
   lane whose work I was crediting. **Resolve by slug** (`who-owns.py` warns).
3. **My own §9 insert silently deleted a `## 10` heading** — the deleted-markdown-line landmine
   in its heading costume. Gate on `git diff --numstat` (the count), then read the lines.
4. **A time window that misses the run's own timestamps reads as an absence.** My first skip-row
   query used `> 09:10Z` for a run that executed 09:05–09:06. State the run's timestamps first.
5. **`[INFERRED]` beat a census twice** — "first such run since 177" was refuted by one all-time
   query. Per the owner's 2026-08-22 ruling, **every count now carries the date it was counted**.

---

## 6. SESSION-START CHECKLIST FOR THE NEXT THREAD

1. `kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath=…` — one tag, one
   replicaset, settled? If a new build has rolled, run §2's re-verification **before** trusting
   any gate row.
2. Run §2's three queries. Record the result in `HANDOFF_2026-08-20_continue_here.md` §2.11
   **whether clean or not** — that file is the gate's home, not this one.
3. At/after **2026-08-24 ~08:45Z**: write the close-out. It must say what the zero cannot prove
   (§2), re-check the `allow_reinstall` class (§3), and close the lane with §4's items carried
   out as separately-owned work.
4. Council/ownership: **nothing owed.** Both of this lane's changes are APPROVED and live
   (`26186633`, `e05ea6f9`); trailers are `Council-Submitted:` and resolve automatically.
