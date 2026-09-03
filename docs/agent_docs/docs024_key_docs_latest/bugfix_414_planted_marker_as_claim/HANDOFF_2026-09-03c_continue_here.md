# HANDOFF 2026-09-03c — the evidence-register programme, after four owner rulings

**Supersedes `HANDOFF_2026-09-03b_continue_here.md`** (13:00 UTC), which the owner's afternoon
rulings overtook. Read this one. `-03b` and the two before it are kept for the trajectory.

**All times UTC.** The session clock is BST, one hour ahead; every figure below is stamped UTC.
**`bugs_closed/414` itself is finished** — this lane is now the register programme it spun out.

---

## 0. THE ONE-LINE STATE

**Three of the owner's four rulings are built and live; the fourth is waiting on a peer's reply.**
What remains is a *programme* (12 registers to populate) rather than a decision, plus two council
verdicts to read and two clocks that will quietly mislead you if nobody watches them (§3).

---

## 0b. ⚠ WHAT HAPPENED AFTER §0 WAS WRITTEN — read this before anything else

**Two things went differently from the rest of this file, and both need action.**

### 0b(i). D2's repair worked AND cost content. Three pages lost their disclaimer.

739's four items were claimed by `build-dispatch-loop` **7 to 20 minutes** after filing (14:25,
14:29, 14:34, 14:43) and all four completed. **The corrections landed, and landed well** — all
three wrong strings went 1 → 0, and the served jargon-buster now says *"CONC 5A.2.14R(1) makes this
a cumulative limit, not a per-instance one: £15 is the most a lender can charge in default fees
across the whole agreement, however many payments you miss."* Better than the brief asked.

**But the items left `spec.mode` unset, so page-build-handler REGENERATED each section instead of
editing it.** 739's council verdict — **APPROVED** — carried a medium objection naming exactly that,
and I read it after the items had run. `load_current_section_content_action.go` (bugs_open/178) is
explicit: without `mode='edit_live'` the writer "gets the item's guidance text and nothing to work
from, so it fabricates a full replacement section."

`[MEASURED at the served bytes vs a pre-repair crawl]`

| page | bytes | sentences replaced | wrong claim | disclaimer |
|---|---|---|---|---|
| the-payday-loan-price-cap | 84% | **36 of 37** | gone ✓ | **LOST** |
| jargon-buster | 88% | **49 of 50** | gone ✓ | gained ✓ |
| loan-sharks-and-illegal-lending | 66% | 47 | gone ✓ | **LOST** |
| check-your-lender-is-authorised | — | — | gone ✓ | **LOST** |

⚠ **BYTE RETENTION HID THE REWRITE.** 84% and 88% read as mild edits and were near-total
replacement at a similar length. **A length check cannot detect a rewrite; only a sentence-identity
diff can.** The estate's text floors are all length-based, so this blind spot is not only mine.

**The site-identity disclaimer** ("does not lend money, broker loans, or take applications…", plus
not-advice and not-the-FCA) is now **missing from three pages**. It was on 14 of 30 before and is on
12 today. On a finance information site that is a compliance element. On loan-sharks the FSMA-2000
criminal-offence framing (a **registered** fact), the card/passport security prohibition and the
Illegal Money Lending Team's anonymity/free detail also went.

**→ THE REPAIRS ARE WRITTEN, TESTED, COMMITTED AND DELIBERATELY NOT APPLIED.** Migrations **743**
(two pages) and **745** (the third), both carrying `mode='edit_live'`, the anti-fabrication
constraint bound to the 738 register, `source='manual'`, and `page_id`/`affected_url` derived by
SELECT. Both dry-run clean with mutations killed. **Held pending 743's verdict
(`4718725c-7d23-41ca-a320-17ebbbfb5e02`) because applying before reading a verdict is precisely
what caused this.** Applying is one command each:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/743_loancash_restore_content_dropped_by_the_739_rewrite.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/745_loancash_restore_disclaimer_third_page.sql
# then record both with run-migrations.sh --record-only <file> --note "..."
```

**Do NOT revert 739.** The corrections are good; only the collateral needs restoring.
Full account: `WRONG_CALLS.md`, 2026-09-03 entry.

### 0b(ii). D3's check came back REVISE, and one objection was a real defect — now fixed.

**742's verdict: REVISE** (`0d730d51-a923-4b44-a58f-ab8c898d7e22`). The sharpest objection: the
pre_query scoped on `s.status='deployed'`, but **the estate's liveness convention is
`IN ('active','deployed')`** — the predicate **all four** `site-discovery-rotation` tasks use. A
detector with a narrower liveness predicate than the dispatcher feeding every other check
re-creates its own blind spot one status value over. **Fixed and applied: migration 744.**
Population unchanged at 12 (no `active` sites exist today), asserted by its own verify block —
which is exactly why it was worth taking then rather than later.

Three other objections dispositioned with measurements, one accepted as a real process gap
(I did not survey `discovery_checks/` before building, and implied I had — a per-site discovery
check could have hosted this; the design still stands on cadence and on Go-waits-for-a-roll, but
that was not the reason I gave). All recorded in **CLM-033**. **742 still owes a RESUBMIT** with
`RESUBMIT_CORR=0d730d51-a923-4b44-a58f-ab8c898d7e22` so the trail accumulates.

---

## 1. THE FOUR RULINGS AND WHERE EACH STANDS

Given 2026-09-03 afternoon, recorded verbatim in substance at **RFC_060 §3g** (one of them amends a
previously-ruled table row, so it is struck through in place rather than changed silently).

| | ruling | state |
|---|---|---|
| **D1** | *"we'd want a register for vetcomparison"* | **NOT STARTED — and read §2 before starting.** |
| **D2** | *"fix the loancash wrong sentences"* | **DISPATCHED 14:18:45 UTC.** Migration **739**, 4 items. ⚠ two clocks — §3. |
| **D3** | *"build the missing check and fill the missing data"* | **CHECK LIVE AND OBSERVED RUNNING** (mig **742**, CLM-033). **POPULATION QUEUED, not done: 12 items.** |
| **D4** | *"a register for each site … but the bar can be lower for normal sites somehow"* | **DESIGNED (§3g(i)), UNEXERCISED.** |

### 1a. D4's "somehow" has a code-derived answer — do not re-invent it

The owner's instinct landed on a rung the estate had **already ruled** (Q3, 2026-09-02): the posture
ladder `standard` / `sourced` / `relied_upon`. But its `standard` row said *"today's behaviour,
unchanged"* — i.e. **no register**. D4 changes that row, and RFC_060 §3g(i) sets the new bar:

| | `standard` — the **ATTESTED** register | `sourced`/`relied_upon` — the **CITED** register |
|---|---|---|
| facts carry | `value` + `context_terms` (+ `tolerance`), **no citation** | that, **plus** `source.citation{url,quote,…}` |
| what it arms | `ScanUnregisteredNumbers` — **fully** | that, **plus** the nightly quote re-check |
| nightly fetch | **none, so no `citation_lost` drift risk at all** | every URL re-fetched, every quote re-matched |
| cost | **hours** | **~half a day** |

**Why the asymmetry is correct rather than a concession:** `numberSupported` (`claims.go`) consults
only `Value`, `ContextTerms` and `Tolerance` — **never `Source`**. So an uncited fact arms the
anti-slop scan *exactly as fully* as a verified one. And a `standard` site's claims are about
**itself**: "we have built 40 sites" cannot be checked against a URL, only attested. Demanding a
citation there yields either a fabricated source or an empty register — and an empty register
**disarms the scan entirely** (`if eb == nil … return nil`), which is the failure D4 exists to prevent.

⚠ **The bound, stated so nobody discovers it after twelve registers.** `ScanUnregisteredNumbers` is
gated on `ProseNumbersAreClaims()`, **false** for `editorialPageTypes` — `guide`, `blog-post`, `tool`,
`game`, `news-index` — and for `thirdPartyDataComponents`. **A register does NOT arm the numeric scan
over guide, blog-post or tool bodies.** Coverage falls on `content`, `landing`, `section-index`,
`entity-page`, `report`. **That exclusion is measured and deliberate** (each membership earned by
counted false positives: `blog-post` 46, `tool` 7, `game` 4, `guide` 1+15, `news-index` 1;
`section-index` was measured and **rejected**) — read the doctrine block above `editorialPageTypes`
before touching it. D4 delivers numeric coverage on **marketing surfaces**, plus `banned_claims` at
both rungs, plus the end of disarming-by-emptiness. It does not promise number-checking inside guides.

---

## 2. D1 — vetcomparison: **REPLIED. UNOWNED AND ACCEPTED.** Ready to build.

**The lane answered in full (16:30) and the answer is yes.** The register is **unowned**, that lane
never started one, and they welcomed it — *"your resolveEvidenceSites point means nothing else will
ever surface this."* Their whole handover is recorded in
`docs024_key_docs_latest/vetcomparison/NOTES_vetcomparison.md` (2026-09-03 entry), contributed at
their invitation so it survives whichever session ends first. **Read that before starting.** The
headlines:

- **Rung: the CITED bar.** vetcomparison is RFC_060's own `relied_upon` worked example. Full method,
  quotes verified through the production matcher. ~20 served pages. Half a day.
- **⚠ A LIVE ERROR IS ALREADY IDENTIFIED AND UNFIXED SINCE 08-24:** the guides state the
  **£21 / £12.50 prescription caps as SETTLED FIGURES**; the draft Order carries them as bracketed
  placeholders *"adjusted for inflation before the Order is made"*. That lane fetched the draft and
  flagged it internally; nobody acted. **Same shape as the loancash findings — a provisional figure
  served as settled.**
- **Facts are ready-made, take them verbatim** (calculator item spec `d5163ed3` + their NOTES): the
  draft Order Article 3 compliance table (17 obligations × Large/Small), and three gov.uk pet-travel
  pages attested 08-26.
- **⚠ TAG EVERY CMA FACT DRAFT-VS-FINAL.** Statutory deadline for the substantive Order is
  **23 September**; the day it is made, all of it needs re-verification.
- **Do NOT re-assert third-party prices** (already handled structurally in `business_intel`, with
  consent snapshots), and **do NOT fill the deliberate absences** — no independence claims, no
  unclaimed practice prices, the OV-qualification claim held unpublished on purpose. A register that
  "completes" those publishes something a person chose not to publish.
- **For the vet preset: BAN "proprietary data"** — the site was remediated 2026-07-14 for fabricated
  prices, a false "proprietary data" notice and a fabricated CMA quote
  (`LEGAL_2026-07-15_vetcomparison_factual_record.md`). Their words: *that exact phrase is this
  site's original sin.*
- **And the width warning applies in reverse:** the finance sibling patterns will false-positive
  here (the site legitimately says "we publish no prices yet"). Run every inherited pattern over its
  own served pages, require 0 hits, with a positive control proving it is not inert.

### (superseded) what was asked before the reply arrived

I messaged the live `vetcomparison [0d8f85]` session at ~13:50 with two questions:

1. **Is the register yours or unowned?** (`scripts/who-owns.py` reads commits, so a session mid-work
   is invisible to it — and that lane has been running 8 days.)
2. **Would a register be the wrong tool here?** A comparison site's figures may be **other people's
   claims rather than its own**, which changes the rung and possibly the whole approach.

**No reply had arrived when this was written.** Check for it first — it arrives as a
`<cross-session-message>`; `ListAgents` shows the session. If it is unowned, the site is RFC_060's own
`relied_upon` worked example (§3a/§3b, "carries animal-health claims"), so it takes the **cited** bar,
not the cheap one. Method: `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` **§8**, and read **§8b, §8c and §8e**
first. Budget half a day. Its `missing_evidence_register` item is already queued.

---

## 3. ⚠ THE TWO CLOCKS ON D2 — the part most likely to go wrong unattended

Migration **739** filed four `content_rewrite` items at **14:18:45 UTC** (verified from `created_at`, not estimated), `status='triaged'`,
`handler_agent='page-build-handler'`, on `/guides/the-payday-loan-price-cap.html`,
`/guides/jargon-buster.html`, `/guides/loan-sharks-and-illegal-lending.html` and
`/guides/check-your-lender-is-authorised.html`.

1. **48 HOURS.** `stale-work-item-reaper` flips `triaged` + `pipeline='build'` items with
   `claimed_at IS NULL` and `updated_at` older than 48h to **`unresolved`**, prefixing the summary
   `[stale: triaged 48h+]`. All four are exactly that shape. **Deadline 2026-09-05 14:18:45 UTC** — but the predicate is on `updated_at`, not `created_at`, so ANY write to the row resets the clock (`trg_site_work_items_updated_at` bumps on every write; a periodic write would make an item unreapable for ever, which is its own documented landmine). A
   reaped item reads as *processed*, not *ignored* — that is the trap.
2. **A `page_rerender` was already queued on all four pages at 13:18 UTC**, before the repairs. A
   rerender regenerates from `content_data`, so it **re-ships the wrong wording byte for byte and
   completes successfully.** It is not a repair. Do not read a completed rerender as a fix, and do not
   cancel the repair items because a rerender ran.

**And the acceptance gate for this type is inert:**
`complete_work_item_acceptance_predicate.go` records that **no verifier is registered for
`content_rewrite`** (13 `RegisterVerifier` calls, none naming it). **A `complete` content_rewrite is
not a repaired artefact.** Every item therefore carries its own `acceptance_test` naming the served URL
and the exact string that must or must not appear. **Verify at the served bytes.**

### What the three repairs actually are

1. **The £15 default cap is CUMULATIVE across the whole agreement** — CONC 5A.2.14R(1), "whether in
   relation to one breach or cumulatively in relation to multiple breaches". Two pages call it
   *per missed payment*. **The site UNDERSTATES the protection it exists to explain**: a reader who
   missed two payments would accept a second £15 as lawful. **This is the one that matters.**
2. **The CPA limit is TWO REFUSED requests and there is no £1 threshold in CONC 7.6 at all.** The
   loan-sharks page says "more than one payment attempt of over £1". The site's **own** CPA page is
   correct and **must not be changed** — it is the reference to be made consistent WITH.
3. **Affordability is CONC 5.2A, not "CONC 5A"** (the price-cap chapter, no affordability rule).

---

## 4. D3's CHECK — live, and here is what "working" looked like

**Migration 742**, register **CLM-033**, applied **14:29:02 UTC**. A **CTE-only scheduled task**
(`fire_message=false` — the `pre_query` IS the worker, `cmd/scheduler/main.go:274`), daily, enabled.
DB config, so **live on apply; no roll needed** — which is why it is not Go (RFC_006's ruled shape: a
CI-time check structurally cannot gate live config).

**It fired within a minute and filed 12 items.** Observed at the artefact, not inferred:

```
kafka-scheduler-6b646c4b4c-srnv2  scheduler/main.go:286
"Pre-query task completed (no message fired)" task=evidence-register-absence
pre_query_result={"already_open":"0","filed_new":"12","missing_total":"12"}
```

**Why it reports three numbers on every tick rather than staying quiet when idle.** Returning no rows
would have been shorter — the scheduler treats that as a successful no-op — but it collapses *"no site
is missing a register"* with *"sites are missing one but each already has an item open"*. That is the
exact `omitempty` failure this lane spent the morning on (RFC_060 §3f). **A missing log line now means
the task did not run, and `missing_total=0` is a positive statement of health.**

**Proven before it was enabled** — three arms in one rolled-back transaction, and four killed
mutations: pass 1 `12/12/0`; pass 2 `12/0/12` with **12 rows, not 24** (it will not re-file daily);
pass 3 after cancelling one item `12/1/11` (a dismissed finding returns while the register is still
missing — right for a compliance queue). `loancash.co.uk` was **absent from all three**, which is the
control showing the check discriminates rather than filing on everything.

**Why `NOT EXISTS` and not `ON CONFLICT`:** `idx_swi_dedup` is a PARTIAL unique index, and an
`ON CONFLICT` inference must restate its predicate exactly or raise **42P10**. Guard 2 of the migration
aborts if that predicate stops listing the seven terminal statuses the pre_query hardcodes, and the
verify block trips on the string `ON CONFLICT` appearing in the pre_query at all.

---

## 5. D3's OTHER HALF — the population, which is the real remaining work

**12 `missing_evidence_register` items, `needs_human_review`, empty `handler_agent`** (deliberate: the
posture rung is a **Q4 record a person signs**, and for a cited-bar site the facts need a primary
source read — giving it a handler would manufacture the appearance of an automated fix for
irreducibly human work). Each spec carries the evidence for the decision (page count, distinct page
types) and **both bars**, not a decision.

`advertise.co.uk` 24pp · `cookly.uk` 15 · `cv1.co.uk` 8 · `designblog.co.uk` 17 · `garden-tools.uk` 20
· `homegarden.uk` 27 · `idea.uk` 38 · `lampenkap.com` 13 · `oxenunity.com` 6 · `seotools.co.uk` 42 ·
**`vetcomparison.uk` 23** · `websitepromotion.co.uk` 27

**Do not treat RFC_060 §3g(iii)'s split into likely-`standard` and likely-`sourced` as scope** — it is
`[INFERRED from domain and category, NOT measured per site]` and says so. The rung is decided by
**reading what the site actually asserts, at the served pages**.

**Expect to find errors.** Four lanes out of four have found real mistakes in their own site's live
copy (lendzy 2, loanzy 1, loancalculator 2, loancash 3). Record each as `corrects_site_citation`;
**do not rewrite served copy without the owner** — that hold was lifted once, for three named findings
on one site (D2), and is not general.

---

## 6. TWO COUNCIL VERDICTS TO READ — both were pending at 15:30 UTC

Both commits carry `Council-Submitted:`, which `098` credits automatically once approved — no amend
needed, and forward-only forbids one anyway. **But read them and act on a REVISE: both migrations are
already applied.**

| | correlation | what it reviews |
|---|---|---|
| **739** | `93897fb5-0b73-4b1e-b4aa-c0d2b9d4a87b` | the four copy repairs |
| **742** | `0d730d51-a923-4b44-a58f-ab8c898d7e22` | the absence check + CLM-033 |

```bash
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
 "SET statement_timeout='20s'; SELECT coalesce(metadata->>'decision','(pending)') FROM diagnosis_artifacts
  WHERE correlation_id='<corr>' AND kind='council_report' LIMIT 1;"
```

**738's verdict came back APPROVED round 1 with four low objections, and one of them was real** — the
migration bound nothing to the *domain*, so a mistyped UUID would populate another site while every
verify count still passed, being scoped to the same wrong id. That is fixed, written into the shared
method as **RUNBOOK_lendzy §8e**, and 739/742 both carry the guard. Expect the same standard of
objection on these two.

---

## 7. WHAT I GOT WRONG OR NEARLY WROTE — the transferable half

1. **`created_by` on a current row names the last WRITER, not the author of the values.** A refresher
   that supersede-and-reinserts relabels every field it merely preserved — including one it has **no
   code path to author**. I nearly wrote "the daily sweep closed farmerinsurance's gap on its own"; the
   history says migration 713, the previous evening. Read the aspect's **history**, not its current row.
   Now a LANDMINE (verified `87ea5043`).
2. **A hand-transcribed citation quote fails silently and for ever.** The `DISP 2.8.2(2)(b)` quote had
   commas where the source has parentheses. It returned **false** on the production matcher. Shipped, it
   would have read as `citation_lost` drift **every day**. **Paste; never retype.**
3. **A "shared" set can exist in more than one WIDTH.** The finance `banned_claims` set drifted while
   being copied site to site. lendzy's bare `\bno credit checks?\b` fires on loancash's **correct**
   advice about employer salary advances (measured, 1 hit); the narrow variant fires 0. *"Does this site
   carry the shared set?"* returns the same answer either way. Now a LANDMINE.
4. **A guard nobody has watched fail is not a guard** — and a bad mutation is not a passing guard. My
   first attempt at mutating 739's double-file guard **did not actually change its behaviour**, so it
   "passed"; the correct mutation aborted before the insert. If a mutation passes, suspect the mutation.
5. **Migration numbers collide on this tree.** Two other sessions took `740` while I was writing, and a
   third took `739` alongside mine. Filenames are unique and the runner records by filename, so this is
   survivable — but **check the max immediately before you name the file, and resolve by filename, never
   by number.**
6. **A handoff can go stale within the hour** — `-03`'s §4 had three rows wrong 15 minutes after it was
   written. That is why this file leads with state and stamps everything.

---

## 8. HOW TO RE-DERIVE EVERY FIGURE

```bash
# the 12 (or fewer) sites still missing a register — the D3 queue
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT spec->>'domain' AS domain, spec->>'page_count' AS pages, status
  FROM site_work_items WHERE item_type='missing_evidence_register' ORDER BY 1;"

# did the absence check run, and what did it report? (three numbers on EVERY tick)
kubectl -n ai-persona-system logs deploy/kafka-scheduler --tail=2000 | grep evidence-register-absence | tail -3
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT last_completed_at::timestamp(0), enabled FROM scheduled_tasks WHERE name='evidence-register-absence';"

# the four copy repairs — and the 48h reaper clock
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT p.url, w.status, w.claimed_at, w.updated_at::timestamp(0),
         now() - w.updated_at AS age_vs_48h
  FROM site_work_items w JOIN pages p ON p.id=w.page_id
  WHERE w.created_by='loancash_couk_fca_validation lane (migration 739)' ORDER BY p.url;"

# VERIFY A REPAIR AT THE SERVED BYTES, never at item status (no verifier exists for content_rewrite)
curl -s https://loancash.co.uk/guides/the-payday-loan-price-cap.html | grep -c "once per missed payment"   # must become 0
curl -s -o /dev/null -w "%{http_code}\n" https://loancash.co.uk/zzz-not-real-qx7.html                      # control: 404

# loancash's register, read back through a JOIN on sites (never the uuid you typed — RUNBOOK §8e)
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT s.domain, jsonb_array_length(ss.data->'facts') facts,
         jsonb_array_length(ss.data->'banned_claims') banned
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='loancash.co.uk';"

# verify a citation quote the way production will, BEFORE shipping it
go run ./cmd/fcaquotecheck "<url>" "<verbatim quote>" "zzz deliberately absent control qx7"   # true, then false
```

⚠ **postgres `exec` queries were intermittently timing out this afternoon.** Use
`SET statement_timeout='20s'` and retry; do not conclude anything from a hang. ⚠ **Never `strings`**
on a pod (absent from the debian-slim images) and never a *discovery* grep for "some 40-hex string".

---

## 9. WHERE THE RECORD LIVES

| what | where |
|---|---|
| the four rulings + the tiering design + build status | `architecture_review/RFC_060_…md` **§3g**, §3g(0), §3g(i), §3g(ii), §3g(iii) |
| the structural finding (why absence was invisible) | RFC_060 **§1d**; the fifth register + its traps **§1e** |
| loancash's register / repairs / absence check | `sql_for_agents/738_…`, `739_…`, `742_…` (+ each `_ROLLBACK`) |
| the new mechanism, registered | concept register **CLM-033** (`claims-verification.md` + index row) |
| the method for the next register | `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8 · §8b · §8c · **§8e** |
| this lane's technical log | `bugfix_414_planted_marker_as_claim/NOTES_planted_marker_as_claim.md` §(q), §(r) |
| loancash lane's log + owner prose | `loancash_couk_fca_validation/NOTES_…md`, `README_where_we_are.md` |
| the owner's plain-prose history | `bugfix_414_planted_marker_as_claim/README_where_we_are.md` |
| landmines added today | *a refreshed spec's `created_by` names the last WRITER* · *a "shared" `banned_claims` set exists in more than one WIDTH* |
| the relay to the detector's owning lane | `claims_verification/CONTRIB_2026-09-03_from_414_…md` (+ its dated correction) |
