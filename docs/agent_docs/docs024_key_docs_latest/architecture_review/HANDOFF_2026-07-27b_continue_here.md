# HANDOFF — architecture seat, continue here (2026-07-27, evening)

**COLD-START ENTRY POINT. This supersedes `HANDOFF_2026-07-27_continue_here.md`**,
which is still accurate in its landmines section but wrong at the top: the thing it
calls "THE ONE THING OWED" was done, and the two decisions it leaves open have been
ruled. Read this instead; go back to it only for §5's landmine list.

Then, if you want the prose: `SUMMARY_2026-07-27c_rulings_and_the_seat_that_cannot_see.md`
(current state), `README_where_we_are.md` (the owner's plain-English log, append-only,
newest at bottom), `NOTES_architecture_seat.md` (technical log + every misstep),
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` (the decision register,
D1–D11), `RUNBOOK_architecture_seat.md` (every command with its gotcha).

---

## 1. What this workstream is

The owner asked for a process — possibly a council seat — that keeps the architecture
robust, stops it shifting underneath us, and keeps it sufficient for anticipated
plans, knowing those goals conflict. The conservative half already existed (the
guardian seat, sole hard-veto holder). Nothing argued the forward half. This
workstream measured that the imbalance was real (ossification, not churn) and built
the counterweight.

## 2. THE OWNER HAS RULED. Nothing is waiting on him.

| ruling | date | meaning |
|---|---|---|
| **D7(a) the veto survives** | 07-27 | guardian keeps the hard block |
| **D7(b) DO NOT narrow the guardian** | 07-27 eve | it keeps the veto **and** the full remit incl. weighing benefit. **D7 closed both halves** |
| **D9 DO NOT add a fix-lane forward seat** | 07-27 eve | *"we have this new one"* — one forward voice is enough |
| **D11 seats must be able to LOOK THINGS UP** | 07-27 eve | the honesty caveat is an **interim, not a destination** |

**The design is settled: one conservative seat at full remit + veto, one forward seat,
no duplicates, balance struck by the two arguing rather than by trimming either.**

**Do not reopen D7(b) or D9 on a trigger firing.** I left reversal triggers in the
decision text; after a ruling those are the *evidence you would need to ask the owner
to revisit*, which is a higher bar. D10 (landmines as a footprinted corpus) is another
thread's proposal parked for the owner to READ, not a question we are blocked on.

## 3. State — everything LIVE is config; nothing waits on an image

Verified by content against the live DB, 2026-07-27 evening.

| thing | where | state |
|---|---|---|
| Council reads its own minutes (D8a′) | guardian + historians, all 3 councils | **LIVE** |
| Guardian deflection check (D5) | all 3 councils | **LIVE** |
| Generated case index (D8e-1) | both historians | **LIVE** |
| `review_architecture` forward seat | `feature-designer` only | **LIVE** — and has **never spoken** (0 reviews) |
| `CODE INDEX LIMITS` caveat | **all 15** prompts mentioning `code_checks` | **LIVE** |
| **D11 layer 2 (routing)** | `fix-proposer` + `feature-designer` | **LIVE** |
| D11 layer 1 (bodies in the index) | council gate | **round 3 in flight** |

**D11 layer 2, precisely:** `fix-proposer.code_lookup.code_check_fields` 6 → **7**
(`review_prior_art` added — it was the one seat emitting `code_checks` and missing
from the answer list); `feature-designer` **gained a `code_lookup` step**
(`run_checks → code_lookup → repropose`) answering `review_architecture`.
`council-gate` deliberately has none — see §5.

**What that buys today:** `symbol` and `ls` code lookups **work now** for the
architecture seat. `content` lookups still return nothing useful until layer 1 ships.

## 4. What is open, in order

1. **The council verdict on layer 1.** `SUBMISSION_CORR=18fe4035-4fa6-4079-ab44-8541d6e58944`.
   Rounds 1 and 2 were REVISE; **round 3 was in flight at ~19:45**. Check it first:
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='18fe4035-4fa6-4079-ab44-8541d6e58944' AND kind='council_report'
   ORDER BY created_at;
   ```
   Objections come back with the reviewers' own checks already answered. Read them
   with `jsonb_array_elements(body::jsonb->'reviews')` — the RUNBOOK has the query.
   **If APPROVED**: the plan is a schema migration + Go, so it is the first thing on
   this workstream that is NOT config — commit with `Council-Reviewed: 18fe4035-…`,
   then build, roll, and verify **against the running pod**, never git or the tag.
   The submission is committed at
   `architecture_review/SUBMISSION_2026-07-27_code_tier_lookup.json`.
   **If REVISE again**: resubmit with `RESUBMIT_CORR=18fe4035-…` so the trail
   accumulates. Rounds are orchestration-scoped, so a resubmission is judged fresh.

2. **The seat still has 0 reviews, and waiting alone will not change that.**
   `review_architecture` exists only on `feature-designer`, which refuses anything
   without an owner-approved `capability_gap` spec (`check_spec_approved` wants BOTH
   `owner_approval` and `code_pointers`). There are 5 such items, 2 approved, and
   **both belong to other threads**. Its first review arrives on the colour-fixer
   thread's **round 4** (owner-directed, instructions already in the spec) or on a
   newly approved spec. **A zero in §4 of the adoption report is a rate limit, not a
   fault.** Do NOT manufacture one by firing at another thread's ticket — see §5.

3. **When it does speak, read it honestly — and be willing to use the kill switch.**
   `./scripts/council-adoption-report.sh`. Section 5 exists to say the seat is
   producing confident noise and should be pulled. That is a real option, not a
   formality.

4. **D11 layer 3 — the dynamic round-trip.** Today a seat emits `code_checks` and gets
   answers **next round**, so it cannot look while reasoning: it must guess, commit to
   a verdict, and be corrected a round later. The owner's word was "dynamic" and this
   is where it points. `[UNSCOPED]` — a tool-use-shaped change to how a seat runs,
   materially larger than layers 1–2, and it should get its own RFC rather than being
   folded into either.

5. **Markdown is still unreachable and is NOT in the layer-1 plan.** 0 of 4,535 index
   rows are markdown, so `WRONG_CALLS.md`, `/bugs_open/` and the concept register
   remain invisible to every seat. Layer 1 builds the mechanism (a `body` column
   populated from stored line spans); markdown additionally needs the `kind` CHECK
   constraint relaxed — a separate migration. Do it **after** layer 1 lands, reusing
   the concept register's own rediscovery-frequency signal for ranking rather than
   inventing one.

## 5. Landmines — the expensive ones, all paid for

**Read `HANDOFF_2026-07-27_continue_here.md` §5 too** — the Go-contract landmines
there (only `{approve,object,veto}` parse; only `{reviewer,verdict,objections,missing,
notes,degraded}` persist; `hard_veto_from` is an audit label) are unchanged and still
the most expensive ones. New since:

- **A work item's `status` is not an ownership signal, and a docs grep is not a
  coverage check.** I nearly fired a probe council at `7b89fb35` because it read
  `deferred` and no doc mentioned it. Its `spec` jsonb contained three completed
  council rounds and `=== ROUND 4 — ONE CHANGE ONLY, owner-directed`. **For a
  `site_work_items` row the ownership trail lives INSIDE the spec, where no repo grep
  reaches, and `who-owns.py` does not cover work items.** Read the row.
- **The 099 mirror forces `fix-proposer` and `council-gate` to carry byte-identical
  prompt text**, while those two lanes *differ* on whether `code_checks` are answered.
  **No per-lane truth can survive the mirror** — that is why the `CODE INDEX LIMITS`
  caveat is lane-agnostic. Plan around it; don't fight it.
- **`council-gate` has no `code_lookup` and that is DELIBERATE, not a gap to patch.**
  `099_SYNC_gate_roster.py:24-29`: `code_lookup`/`repropose` *"serve the blind
  reproposer, which the gate has no equivalent of (its authors read the objections
  themselves)."* The **same test** is what includes `feature-designer` (it has one).
  The gate's real gap is that code results never reach its verdict note — fix *that*,
  recorded in `bugs_open/108`. **An asymmetry with a comment explaining it is a
  decision; one without is a bug.**
- **`code_symbols.content` does THREE jobs** — trigram search text, the **embedded**
  text, and the input to `content_hash` (the re-embed trigger). Appending bodies to it
  silently **re-embeds all 4,535 rows** and skews the vectors. Bodies go in a
  **separate column**.
- **`code_symbols.kind` has a CHECK constraint** (`func|method|struct|interface|alias|
  type|var|const`) ⇒ **markdown cannot be inserted without a migration.**
  `bugs_open/108` calls indexing markdown "the same question, already settled" — true
  of the mechanism, **not** of the cost.
- **`CREATE INDEX CONCURRENTLY` fails the migration dry-run probe**, which deliberately
  wraps the file in a poisoned transaction (`run-migrations.sh:129-139`). At 4,535 rows
  a plain index is milliseconds. The safe-looking option breaks the safety check.
- **The live migrations home is `docs/agent_docs/sql_for_agents/`**
  (`run-migrations.sh:56`), NOT `platform/database/migrations/`, which stops at 200 and
  is historical. Next free number was **241**.
- **A single-successor walk over a branching workflow reports a FALSE orphan.**
  Traverse `then_step`/`else_step`/`error_step` too, or `conditional` steps make a
  healthy graph look broken.
- **A `ROLLBACK=1` file is a guess unless you prove `live == that file` first.** It
  restores a *file*, not a snapshot; if live has drifted you write a third state.

## 6. Where the measurement stands, and the caveat on it

Baseline (pre-cutover **13:44:56** — take it from `agent_definitions.updated_at`,
never guess): guardian 210 reviews, 90 invoked the stability preference, and **2 of
those 90 cited precedent (2.2%)**.

> **The old "6 of 90" figure was WRONG and is corrected everywhere.** `invoked` and
> `cited` were two **independent** `FILTER`s over all guardian reviews, so the pair was
> never a subset — 4 of the 6 cited precedent *without* invoking the preference.
> `scripts/council-adoption-report.sh` now reports `both_invoked_and_cited` +
> `pct_of_invoked` with `cited_but_did_not_invoke` kept visible.

**The metric is wrong in BOTH directions and the script says so.** It over-counts —
`%deflect%` matches the bare word, and the new prompt itself says "deflected upward",
so a seat echoing its instructions scores a citation. And it under-counts — the 14:18
guardian reasoned correctly about recurrence without quoting a past report and scored
zero. **At low n, read the review text.**

**First post-cutover evidence was genuinely good:** `debug_historian` rejected a "no
sixth path exists" absence claim by citing `WRONG_CALLS.md` **by date** (07-21, 07-24);
`bug_historian` cited `016b §9`, `bugs_open/034`, `109`; the guardian invoked the
stability preference and reasoned its way **out** of deflecting.

## 7. Posture worth carrying

**Eleven wrong calls went into `WRONG_CALLS.md` today**, and the tally is the point:
**three were the same family — a stale or borrowed figure voiced as measurement**
("6 of 90", "3 of 87", "86 of 92 migrations", the last lifted from a code comment).
That check is now worth making reflexive.

**Three of four errors in the council submission were caught by the council, not by
me**, on a plan I had already pre-flighted for quote fidelity, schema and scope. The
sharpest: I proposed *extracting* a shared line-slicer that **already existed** forty
lines below a line I had cited twice — caught by `prior_art_librarian`, a seat that
was answering code questions for the first time **because I had wired it up an hour
earlier**. A plan about unchecked absence claims, containing one.

That is this workstream's own thesis arriving as evidence: **a reviewer with the
written record in front of it caught what the person who wrote that record the same
day did not.** Which is the argument for finishing layer 1 — the record is still
mostly markdown, and markdown is still invisible.
