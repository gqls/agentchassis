# HANDOFF — the checker-layer remaining items (`bugs_open/149` + the C1 gap)

**Written 2026-07-31 for a fresh thread. Self-contained: you should not need to read
the oufe lane to work any of this.** Source queue is
`bugs_open/149_HANDOFF_2026-07-29_discovery_checker_layer_defect_queue.md` — read its
status block first, then come back here for what to do.

---

## 0. Before you touch anything

1. **Re-run every figure in this file.** Each is stamped with 2026-07-30 or 07-31.
   This tree moves fast and two of 149's own items were wrong *because a figure was
   measured in the morning and written up in the afternoon*. Your own measurement
   from an hour ago is another document by now.
2. **`git status` first.** Many sessions, one tree, one HEAD. Committing is shipping:
   `make build-*` builds from committed HEAD, so anyone's roll ships your commit.
3. **Check the queue before dispatching at a target** —
   `SELECT summary, status FROM site_work_items WHERE status NOT IN ('complete','cancelled','rejected') AND ...`
   and `python3 scripts/who-owns.py <number|slug>`.
4. **Read `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`** and grep it by the
   path/table you are about to touch. Five entries were added on 07-30 from this exact
   work; two of them cost 15 minutes each.

**What is already done and LIVE** (chassis `v1.0.1211`, commit `f61dce806`,
pod-verified both replicas): 149's **B4** in full, **C3** (corrected — not a defect),
and **C1's structural half**. Concept register **CLM-018**. Owner signed off on the
refusing behaviour 2026-07-30.

---

## 1. THE BIG ONE — C1's real gap: the gate cannot recognise the fabrication

**Status: OPEN, unowned, not yet a numbered item. This is the most valuable thing in
this file and it is the one I would start with.**

> **CORRECTED 2026-07-31 — worked, and this section's diagnosis is WRONG in a way that
> makes two of its three candidates inert. Now filed as `bugs_open/161`; read that first.**
>
> The gate did not miss "10,000 Monte Carlo trials per query" because
> `businessClaimContextRe`'s vocabulary is too narrow. **It missed it because that figure
> is a REGISTERED FACT** — `gd-trials` in gamesdesign's `evidence_base`, seeded 2026-07-24
> by `bugs_closed/043`'s own remediation, value 10000, `context_terms` including
> `"monte carlo"`. `numberSupported()` matches it on exact tolerance and skips it. **So the
> 0 findings this section measured were CORRECT** — and the same one row disarms the prose
> scan, `ScanStatClaims` and the 149 C1 persistence floor simultaneously, because all three
> call that one function.
>
> Worse, the fact is **false**: neither drop-rate tool performs any random sampling
> (`Math.random` count **0** in both; the tuner is a closed-form `Math.pow(1-p,k)`, the
> simulator says "binomial" and clamps input with `Math.min(val,10000)`). And the register
> is *also* the writer whitelist — `writer_block`, headed "NUMBERS (state only these…)",
> injected into the page-content-writer prompt. **The model did not invent that specific;
> the platform handed it over as verified truth.** The commit message on `4494162af`
> ("invented supporting specifics … a Monte Carlo trial count") is wrong on that point.
>
> Therefore, against this section's own motivating case:
> - **Candidate 1 (widen `businessClaimContextRe`) is INERT** — the number would then reach
>   `numberSupported` and be correctly skipped. Zero effect, plus the false-positive risk
>   `claims.go:612-617` already records. **Do not spend a council round on it.**
> - **Candidate 2 (a structural rule rather than a lexical one) is INERT and largely
>   already built** — `ScanStatClaims` has **no lexical gate today**; it is already purely
>   structural. It also calls `numberSupported`, so the same row disarms it. On the sharper
>   sub-question this section rightly asked — *did the stat audit run, and if it stayed
>   silent that is more serious* — the answer is it would have stayed silent **and been
>   right to**, for this claim.
> - **Candidate 3 (claim-vs-source diff) survives**, but only for the credential half below.
>
> **What stands, unchanged:** the second bullet — "built **by** a shipped live-service
> designer" is a fabricated *human credential*, non-numeric, matching no banned pattern
> (gamesdesign has **0** `banned_claims`). That is a genuine recognition gap and is not
> what 161 is about. The mechanism-shaped paragraph below also stands and is the best thing
> in this section.
>
> Filed through the 2026-07-31 owner ruling's diagnosis loop: intake
> `93cc6cef-39c6-42b0-9861-ab80a235740e`, run `08ff91c4-dfa7-4226-a039-e80a08e44cc1`.

### What happened

A claims FLOOR now runs at the page-section persistence seam
(`save_sections_claims_guard.go`, called from `SavePageSectionsAction`). Six live
agents persist page sections; before this, two ran any claims check. All six now do.
A banned claim refuses the save; an unregistered number is recorded and allowed.

**And it would not have caught the fabrication another thread witnessed live the same
day.** That case (`4494162af`, contributed into 149): `page-content-writer` wrote four
false claims onto gamesdesign.co.uk's homepage via `page-build-handler` — a path that
*does* go through the floor — and they rendered, compiled, completed and deployed.

I checked this rather than assumed it. Copy through the floor's own engine, against
gamesdesign's own register, `page_type=content` so prose numbers ARE scanned:
**0 findings.**

```bash
go build -o /tmp/claimscan ./cmd/claimscan
printf 'homepage\thero\t%s\tcontent\n' \
  "$(printf '%s' '<p>Our drop-rate simulators run 10,000 Monte Carlo trials per query, built by a shipped live-service designer.</p>' | base64 -w0)" \
  > /tmp/witness.tsv
/tmp/claimscan -evidence <gamesdesign evidence_base json> -components /tmp/witness.tsv
```

### Why it missed, which is the actual finding

- **"10,000 Monte Carlo trials per query"** — the unregistered-number scan only looks
  at a number when `businessClaimContextRe` (`datahelpers/claims.go:339`) matches its
  surrounding window: clients, customers, records, awards, uptime, years of
  experience. *"Monte Carlo trials per query"* is not that vocabulary, so the number
  is never scanned at all. This is **`CLM-003`'s documented blind spot**, not a new
  defect — the concept register already says the scan "is near-inert on non-English
  copy and on finance prose". Nobody had connected that to *technical* prose.
- **"built **by** a shipped live-service designer"** — the previous copy said "built
  **for** live-service and tabletop designers". A fabricated *human credential*,
  matching no banned pattern. The set has no shape for it.

### The mechanism-shaped part, which is worse than a wrong number

From the witnessing thread's own write-up, and it is the bit to design against:
the model took a true sentence, made **one grammatical substitution** (`for` → `by`),
then **invented supporting specifics to justify the new sentence** — a trial count, a
technique name. *A fabrication that arrives with corroborating detail reads as more
researched than the honest copy it replaced.* And it repeated across **four**
components including a headline stat card (`stat2_label: "Monte Carlo Trials"`,
`stat2_value: "10,000"`), so a fix that corrects one mention leaves a corrected
sentence directly above a stat card asserting the same falsehood.

### How to work it

**File it as a new Group C item about the SET, not about the seam.** The seam is
fixed and signed off; do not re-litigate placement. The question is what the engine
can recognise.

Candidates, ordered by what closes the door — but **measure before choosing**:

1. **Widen `businessClaimContextRe`** to cover technical/product vocabulary
   (simulations, trials, queries, iterations, samples, requests…). Cheapest. **Dry-run
   fleet-wide with `cmd/claimscan` BEFORE proposing** — the last person to narrow a
   pattern by reasoning made it match nothing across 919 components, and the last
   person to widen one produced 4 false positives on honest negated sentences.
2. **A structural rule rather than a lexical one.** `bugs_closed/043`'s own argument
   was that a `stat*_value` field is a claim **by construction** — position, not
   vocabulary. The witnessed case landed in a stat card. A rule that says "a number in
   a stat field with no register entry is a claim" needs no vocabulary at all, and
   `validate_page_content_stats.go` already exists for exactly this. **Check first
   whether the stat audit ran on that page and what it said** — if it did and stayed
   silent, that is a different and more serious finding.
3. **A claim-vs-source diff at rewrite time.** The witnessed case is detectable as a
   *change*: a true sentence became a false one. Nothing currently compares a rewrite
   to what it replaced. Expensive, and it is a new mechanism — architecture scope.

**Do not file this as "the floor didn't work".** It works and is signed off; it
enforces what the engine recognises. The gap is recognition.

---

## 2. B4's missing durable record — small, self-contained, my own inconsistency

**Status: OPEN. I introduced this and deliberately did not patch it in after the
council verdict, because adding code post-review is the pattern the guardian seat
objects to. It is yours to do properly.**

`discovery_checks.go`: an **erroring** check is now appended to `failedChecks` and
reported in the step output map — but **nothing durable records it**. Pod logs roll,
and `collected_data` is pruned at ~24h. So "which check silently stopped working, and
when" is unanswerable a day later.

The claims floor **in the same commit** honours the opposite rule
(`CONTENT_CLAIMS_FLOOR_DETAIL` written to `agent_error_log`), which is the estate's
"a pod log line is not a record" rule from `bugs_open/071` gap 3. B4 does not. That
asymmetry is the defect.

**Fix:** write an `agent_error_log` row per failed check, at `warning` severity, with
its own error code (do NOT reuse an existing one — see `validate_page_content.go:619-624`
for why distinct codes matter). Mirror `writeSaveSectionsRepairSkipLog` in
`save_sections_claims_guard.go` / `save_sections_link_repair.go`; both are short and
both are the house pattern.

**Raised by:** council `2d0dbc2e`, `bug_historian` edit 3, medium.
**Verify:** the table is `agent_error_log`; its timestamp column is **`occurred_at`,
not `created_at`**, and `agent_type` is **NOT NULL** (default it to `"unknown"`).

---

## 3. The three unguarded persistence paths

**Status: OPEN.** The floor covers `save_page_sections`. Three other live Go call
sites persist LLM-authored page prose with no claims check by any route:

| path | addressable surface, measured 2026-07-30 |
|---|---|
| `create_report_page` | **2** components (`page_type='report'`) |
| `rebuild_blog_listing` | **7** components (`blog-index`) |
| `ApplySectionEditAction` | **cannot be bounded from `page_components`** — it edits in place, and the table has **no provenance column** naming the writing action |

The first two are together under 1% of the 949-component surface, so they are cheap
and low-risk. **`ApplySectionEditAction` is the real one** and it belongs to
`bugs_open/136`'s territory — check ownership before starting.

> **CHECKED 2026-07-31 — the "cannot be bounded" claim STANDS, and `page_component_history`
> does not rescue it.** That table has a `source` column which looks exactly like the
> provenance `page_components` lacks. It is a **write-mode** label, not a writer:
> `save_page_sections_overwrite` on **12,386 of 12,416 rows** fleet-wide, every pipeline
> emitting the same literal; the other 30 are hand-typed operator strings, which is what
> makes the column look discriminating. `SELECT source, count(*) FROM
> page_component_history GROUP BY 1 ORDER BY 2 DESC;` — recorded so nobody re-runs it
> hopefully. Now in `LANDMINES.md`.

**`create_report_page` is gated on a question, not on code:** `report-builder` is the
**only** live agent that sets `check_claims: false` on `validate_page_content`. That
is 149's item **C2** — find out whether it is deliberate (a report restating figures
from a cited upstream may legitimately need different handling) before switching
anything on.

**Raised by:** council `2d0dbc2e`, `bug_historian` edit 1 + `architecture` edit 1.

---

## 4. The rest of 149, in the file's own order

Nothing below is started. Each item in 149 carries the query that measured it.

- **B2** — no `nav_drift` item has **ever** been raised by a discovery agent (all 16
  from named sessions or `generic`). Cause `[UNMEASURED]`, three candidates named with
  **different fixes**: dispatch coverage / swallowed check error / dedup suppression.
  **Establish which before changing anything.** Cheap, and it decides whether Group A's
  fixes would even be reached in production. **Note: B4 is now live, so a swallowed
  check error is no longer silent — re-run B2's measurement on the new code first; it
  may have answered itself.**
- **A6 → A2 → A3** — creation-time first (unrepresentable beats detectable): the two
  tool creators each do half the nav write; then stop routing `/tools/` to a handler
  that structurally cannot repair it; then add the missing `orphan_tool_pages` →
  rebuild-listing route (the blog analogue already exists).
- **A4, A5** — `pages.in_header`/`in_footer` **default to `true`**, and two nav
  builders with opposite predicates. Both shared-mechanism changes wanting their own
  council round. Neither blocks the above.
- **B1** — six registered checks are configured in **no** agent. **NEVER RAN, not
  broken** — do not rewrite them on that basis. Each needs a decision: seat it and see
  what it finds, or delete it. **Expect a burst on first run; that is the check
  working, not a regression.** Needs an owner call.
- **A1** — **not** a handler defect (corrected in the file). The real item is
  recurrence branding: 20 of 24 repeat detections born `unresolved` (terminal,
  non-dispatchable), 16 distinct `item_key`s for 24 rows. Pair with B2 — dedup on a
  repeat `item_key` is one of B2's candidate causes and may be the same mechanism from
  the other side.
- **C2** — see §3 above.

**Every Group B number was measured under the old silent-skip behaviour. B4 is live
now. Re-measure B1/B2/B3 before acting on them.**

---

## 5. Traps that will cost you time on this specific work

All five are in `LANDMINES.md` with footprints; these are the ones that fire here.

- **`grep` in the Claude Code shell is a ugrep wrapper with `-I`.** One non-UTF-8 byte
  in site copy and it returns zero matches and **prints nothing at all** — not even
  `0` for `grep -c`. `LC_ALL=C` does not fix it. Use **`command grep -a`**. The tell is
  a blank where a count should be: real grep always prints a number.
- **`kubectl exec -i` eats a `while read` loop's stdin** — read from fd 3 or
  `< /dev/null`, or you will scan one site of fourteen and the script will exit
  looking successful.
- **`kubectl exec` truncates large exports mid-stream** leaving a well-formed SHORT
  file (`unexpected EOF` on stderr, easy to miss). **Count the rows in the DB first
  and assert the export matches**; one site needed three attempts.
- **`strings` splits a Go literal at every non-ASCII byte**, so a marker containing an
  em dash greps to **0** in an image that contains it — indistinguishable from
  `bugs_open/153`'s tag-bumped-never-rebuilt. Pick ASCII-only markers and run a
  positive **and** a negative control in the same exec.
- **The council runbook's own verdict query returns the most recent note FLEET-WIDE.**
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at
  DESC LIMIT 1` handed me a complete, well-formed REVISE belonging to another session.
  **Key on your own correlation**, or parse `diagnosis_artifacts.body` as JSON
  (`metadata->'reviewers'` is only a COUNT; the per-seat verdicts are in `body`).
- **`jsonb_path_query($.**.checks)` over `agent_definitions` sweeps up
  `maintenance-triage`'s array**, which belongs to `scan_sites_for_maintenance` and is
  a different vocabulary. Only **three** agents call `run_discovery_checks`.
- **A `default_config::text LIKE '%action_name%'` matches PROMPT TEXT.** `council-gate`
  and `fix-proposer` both "contain" `save_page_sections` and neither has the step. Use
  `jsonb_path_query($.**.action)`.

---

## 6. Council practice for this area

- Scope is `platform/`, `internal/`, `pkg/`. Submit **before or alongside** the commit;
  review here is after the fact by design — do not claim an ordering constraint you do
  not have.
- **Answer an evidence objection with a QUERY, never with code.** Round 1 of `2d0dbc2e`
  was gated at HIGH purely on absence claims with no attached SQL; round 2 attached the
  queries, changed no code, and that seat flipped to approve.
- **Read `unreadable`, not `abstained`.** Round 2 of the same correlation returned
  REVISE with `decided_by = "unreadable reviewer(s): …"` and **no high-severity
  objection at all**. That is `bugs_open/119`'s shape — a seat's malformed result
  costing a round — not a judgement on the change.
- Trailer: `Council-Reviewed:` on an **approved verdict you have read**, never
  otherwise. Committing before the verdict → `Council-Submitted:`; `098` resolves it
  at report time, so approval credits with no amend.
- **A SCOPE objection needs a human, not a better-measured resubmit.** The guardian
  seat's objection here was discharged by an owner sign-off on 2026-07-30, not by
  another round.

---

## 7. Where everything is

| what | where |
|---|---|
| the queue itself | `bugs_open/149_HANDOFF_2026-07-29_discovery_checker_layer_defect_queue.md` — **read its status block first** |
| what shipped, and its honest limits | concept register `CLM-018`, `docs026_concept_register/register/claims-verification.md` |
| the witnessed fabrication | commit `4494162af`, contributed into 149 C1 |
| the fleet claims scan procedure, with its gotchas | `docs024_key_docs_latest/oufe/RUNBOOK_oufe.md` |
| the traps | `docs024_key_docs_latest/LANDMINES.md` |
| how we got things wrong here | `docs024_key_docs_latest/WRONG_CALLS.md`, 2026-07-30 |
| the code | `platform/orchestration/actions/save_sections_claims_guard.go`, `discovery_checks.go` |
