# HANDOFF 2026-09-04 — the register programme after D1 and D2 landed

**Supersedes `HANDOFF_2026-09-03c_continue_here.md`**, whose §0b described the loancash restoration as
*written, tested, committed and deliberately NOT applied* and D1 as *ready to build*. **Both are now
done and verified at the artefacts.** `-03c` and its predecessors are kept for the trajectory.

**All times UTC.** ⚠ The session clock is **BST, one hour ahead** — I mislabelled two migration notes
"UTC" off it yesterday and had to correct them. Use `date -u`, or better, let a column carry the time.

---

## 0. THE ONE-LINE STATE

**D1, D2, D3's check and D3's first register are all DONE and verified at the served bytes or the
stored row.** What remains is a **programme** — 11 registers to populate — plus one council round in
flight and three things that will quietly mislead you if nobody watches them (§4).

---

## 1. WHAT IS DONE, AND HOW TO RE-PROVE IT IN ONE COMMAND EACH

| ruling | state |
|---|---|
| **D1** *"a register for vetcomparison"* | **DONE.** Migrations 759 / 761 / 763 / 767 / **772**. 21 facts, 6 banned_claims, posture recorded. Council **APPROVED** (`b6cbdcd3`). 772 corrected the 5 PDF facts to keep their citation + `reverifiable:false` + a 23-Sept staleness clock. |
| **D2** *"fix the loancash wrong sentences"* | **DONE.** 739's corrections stand; 743 restored what 739's rewrite dropped. 3 pages, **0 orphaned sentences**. |
| **D3** *"build the missing check and fill the missing data"* | **CHECK LIVE** (CLM-033, mig 742/744). **POPULATION: 1 of 12 done, 11 left.** |
| **D4** *"a register for each site, lower bar for normal ones"* | **DESIGNED, still UNEXERCISED.** vetcomparison took the CITED bar, so the ATTESTED bar has still never been built. |

```bash
# D2 — the three repaired pages (each must print 1 1 1), with a fetch control
for u in check-your-lender-is-authorised loan-sharks-and-illegal-lending the-payday-loan-price-cap; do
  b=$(curl -s "https://loancash.co.uk/guides/$u.html")
  echo "$u $(printf '%s' "$b"|grep -c 'does not lend money, broker loans') \
$(printf '%s' "$b"|grep -c 'We are independent') \
$(printf '%s' "$b"|grep -c 'not the Financial Conduct Authority')"
done
curl -s -o /dev/null -w 'control must be 404: %{http_code}\n' https://loancash.co.uk/guides/zzz-not-real-qx7.html

# D1 — the register, read back through a JOIN on sites (never the uuid you typed, RUNBOOK §8e)
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT s.domain, jsonb_array_length(ss.data->'facts') facts,
         jsonb_array_length(ss.data->'banned_claims') banned, ss.data->'posture'->>'rung' rung
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='vetcomparison.uk';"
# expect: vetcomparison.uk | 21 | 6 | relied_upon
```

---

## 2. THE THREE VETCOMPARISON ERRORS — **REPAIR AUTHORISED AND DISPATCHED 2026-09-04**

> **OWNER INSTRUCTION 2026-09-04: "please fix the vetcomparison copy".** The content freeze is lifted
> for these three named findings — the same shape as D2's hold being lifted for three named findings on
> loancash. **Migration 771 files 9 `content_rewrite` items**, one per affected page, all
> `mode='edit_live'` + `approval_mode='manual'`, with a guard REFUSING to include
> `/guides/animal-health-certificates/` (its body is a hand-attested literal block whose regeneration
> would reintroduce three model-memory falsehoods the vetcomparison lane replaced on 08-27; measured
> 2026-09-04, it carries 0 instances of all three errors anyway). That lane was consulted first and
> confirmed nothing mid-flight.
>
> **STATE AT HANDOFF: `/about.html` released as the canary and NOT YET CLAIMED; the other 8 still
> `triaged`.** This is queue latency, diagnosed not assumed — see §4.4.

The register pass found three live errors, recorded as four `corrects_site_citation` entries.

1. **THE CMA FINAL REPORT IS DATED NOVEMBER 2024 ON TWO GUIDES. It is 24 March 2026.** NEW this pass.
   `/guides/cma-compliance/` and `/guides/cma-market-investigation/`. The CMA's own case-page timetable
   says *"24 March 2026 Final report published"*; November 2024 is the Inquiry Chair's BVA Congress
   speech, listed on that same page, which is almost certainly the origin.
2. **THE £21 / £12.50 CAPS ARE SERVED AS SETTLED ON SEVEN PAGES.** The draft Order carries them in
   square brackets, inflation-adjustable before the Order is made. Known to the vetcomparison lane
   since **2026-08-24** and still unfixed. Their recommended wording is in the fact record.
3. **"36 SERVICE CATEGORIES" IS 36 SERVICES IN 5 CATEGORIES.** NEW this pass. Draft Schedule 1's own
   heading is *"Service, product, treatment or procedure (36 total)"*; 12+6+6+9+3 = 36.

**If the owner commissions the repair, read §3 FIRST.** A `content_rewrite` item minted without
`spec.mode='edit_live'` will regenerate these pages wholesale, which is exactly what happened to
loancash on 2026-09-03.

---

## 3. THE REPAIR RECIPE THAT WORKS — dearly bought, use it verbatim

`spec.mode='edit_live'` is **not optional**. Without it `page-build-handler`'s writer gets the item's
guidance text and nothing to work from, and fabricates a full replacement section. Verified at both
ends, not assumed: `load_current_section_content_action.go:98` `const editLiveMode = "edit_live"` and
:139 `if inputs.Get("mode") != editLiveMode { return passthrough("not_edit_live") }`, and
page-build-handler's **live** `default_config` carries both the step and the literal.

- **`approval_mode='manual'` parks the item until a human releases it.** `load_work_item_actions.go:802`
  — `AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')`. ⚠ **Left alone it is then
  reaped at 48h into `unresolved`, which reads as PROCESSED.** Release with
  `UPDATE site_work_items SET status='approved' WHERE …`.
- **CANARY ONE PAGE.** 739 fired four at once and lost content on three. Release one, verify, then the
  rest.
- **VERIFY WITH A SENTENCE-IDENTITY DIFF, NOT A LENGTH CHECK.** 739 kept 84–88% of bytes while replacing
  36 of 37 and 49 of 50 sentences. Pre-image: the `op='delete'` row in `page_component_history`
  (join on `page_id`, never `component_id`).

> ### ⚠ AND THE ACCEPTANCE TEST IN `-03c` IS THE WRONG TEST — do not copy it
> It says *"ADDITIONS ONLY, no existing sentence removed or reworded"*. **A restoration that splices a
> clause into an existing sentence CANNOT satisfy that**, so the literal test fails a correct repair.
> Measured on loan-sharks: 5 sentences "removed", every one a splice seam at 0.89–0.98 similarity
> (*"lends money **for profit** without authorisation"*), and **0 orphaned**.
> **The right measure is ORPHANED sentences — a removed sentence with no close survivor.** It was 0 on
> all three pages while "removed" was 5 on one of them.

---

## 4. ⚠ FOUR THINGS THAT WILL MISLEAD YOU IF NOBODY WATCHES THEM

1. **A COUNCIL VERDICT IS ABOUT THE SUBMISSION, NOT THE TREE.** 743's round-2 REVISE was gated on
   `approval_mode='auto'` — true of its **sketch**, false of its **file**, which carried `'manual'` and
   asserted it in its own verify block. Five reviewers across three seats spent objections on a defect
   that did not exist. **This round failed safe; the same staleness pointed the other way draws an
   APPROVAL FOR A FILE THAT IS WRONG, and nothing downstream compares the two** (`098` joins on the
   trailer). **So: generate the sketch FROM the file.** 759/761/767 do, with a self-check asserting the
   file's prefix and suffix match. ~15 lines, and it makes the class impossible.
2. **`RESUBMIT_CORR` REUSES THE CORRELATION, so `ORDER BY created_at DESC LIMIT 1` returns the OLD
   round.** Bound the query by time or count rows. For the round now in flight (`5d54f835`, resubmitted
   **2026-09-04 11:09**): `… AND created_at > '2026-09-04 11:09'`, or `count(*)` where **1 = round 2 has
   not landed**.
3. **THE ABSENCE CHECK HAS RUN EXACTLY ONCE.** `evidence-register-absence`, `interval_seconds=86400`,
   `last_completed_at = 2026-09-03 14:29:02` — that is its **first and only** tick, and it is **on
   schedule**, not stale (next due ~14:29 today; I briefly misread age-since-last-run as overdue).
   **Consequence: its dedup arm has never been exercised in production.** Pass 2's `12/0/12` — files
   nothing when items are already open — is proven only inside a rolled-back transaction. **Watch the
   next tick**: it should report `missing_total=11, already_open=11, filed_new=0`, and the 12→11
   decrement is the first live evidence the check tracks reality.
   ```bash
   kubectl -n ai-persona-system logs deploy/kafka-scheduler --tail=2000 | grep evidence-register-absence | tail -3
   ```

---

4. **A FRESHLY FILED ITEM IS LAST IN A FLEET-WIDE FIFO — "filed and approved" is not "about to run".**
   `build-pipeline-trigger > find_dispatchable_site` picks **ONE site per 30s tick**, ordered
   `ORDER BY MIN(created_at) ASC LIMIT 1` over eligible items. So a new item waits behind every older
   eligible item on every other site. `[MEASURED 2026-09-04 11:45]` the queue was
   mortgagecalculator (10:40) → gamedesign (10:54) → copyonline (**51 items**, 11:09) →
   **vetcomparison (11:32, mine)** → advertise (11:34). Yesterday loancash was claimed in ~4 minutes
   because its item happened to be the oldest; that was luck, not a service level.
   A site is also excluded while it has any `claimed` item, so sites interleave rather than one
   draining fully — which is why this is latency, not a stall.
   ```sql
   -- where am I in the queue? (reproduce the selector's OWN clauses — see the warning below)
   SELECT step->'config'->>'query' FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS s(k, step)
    WHERE ad.type='build-pipeline-trigger' AND ad.is_active
      AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL AND s.k='find_dispatchable_site';
   ```
   ⚠ **AND DO NOT HAND-REPRODUCE THAT QUERY FROM MEMORY — I DID, AND GOT A CONFIDENTLY WRONG ANSWER.**
   My first reproduction omitted `depends_on` and `governor_admits`, which put
   `websitepromotion.co.uk` at position 1 with items 23 HOURS OLD and had me about to report a
   fleet-wide dispatch stall. There is no stall: all three of those items have **unsatisfied
   `depends_on`**, so the real selector correctly excludes them. **Read the query out of
   `agent_definitions` and run it verbatim** — the same "a mirror is not production" family as §6a's
   probe defect, hit twice in one session.

## 5. THE REMAINING PROGRAMME — 11 registers, and pick the rung by READING the site

`advertise.co.uk` 24pp · `cookly.uk` 15 · `cv1.co.uk` 8 · `designblog.co.uk` 17 · `garden-tools.uk` 20 ·
`homegarden.uk` 27 · `idea.uk` 38 · `lampenkap.com` 13 · `oxenunity.com` 6 · `seotools.co.uk` 42 ·
`websitepromotion.co.uk` 27

Each has a `missing_evidence_register` item at `needs_human_review` carrying **both bars** and the
evidence for the decision, deliberately with no `handler_agent`. **Its acceptance test has TWO halves**
and the second is easy to miss: a register with at least one valued fact, **AND** the posture rung
recorded with who declared it and when. §6 is how vetcomparison satisfied the second.

- **Do NOT treat RFC_060 §3g(iii)'s likely-`standard`/likely-`sourced` split as scope** — it is
  `[INFERRED from domain and category, NOT measured]` and says so.
- **EXPECT TO FIND ERRORS. Five lanes have run this method; five found errors in their own live copy**
  (lendzy 2, loanzy 1, loancalculator 2, loancash 3, vetcomparison 3). Record as
  `corrects_site_citation`; do not rewrite served copy without the owner.
- **D4's ATTESTED bar is still unbuilt.** vetcomparison was `relied_upon`, so it took the CITED bar.
  The first `standard` site is the cheap path (`value` + `context_terms`, no citation, hours not half
  a day) and nobody has walked it yet.

---

## 6. THE METHOD, PLUS FOUR THINGS THIS RUN ADDED TO IT

Base method: `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` **§8**, and read **§8b, §8c, §8e, §8f, §8g** first.

### 6a. ⚠ ~~A PDF SOURCE CANNOT BE CITED~~ — **RETRACTED 2026-09-04. I had this wrong.**

> **THE CLAIM THIS SECTION USED TO MAKE IS FALSE, and it was in five documents plus a landmine.**
> I wrote that citing a PDF would classify as `citation_lost` drift *"every day, for ever"*.
> **Production never does that.** `refreshCitationFact` → `verifyCitationLiveForRule` →
> `fetchCitationDocument`, which refuses a non-html/xml/text content type
> (`evidence_citations.go:143-148`) and classifies it `fetch_error` → outcome **`error`**. That
> file's header has always said so: *"fetch failed (network, 403, 5xx, unsupported content type) →
> UNKNOWN, not drift … Reported as an error, never as loss"*, and *"PDFs and other non-text content
> are refused rather than half-read"*.
>
> **I asserted a production behaviour from an instrument that is not production.** Full account:
> WRONG_CALLS 2026-09-04.

**What was really broken, and is now fixed:** `cmd/fcaquotecheck` — the probe RUNBOOK §8 step 4 sends
every register author to — did **not** call the production fetch. It did its own bare `http.Get` and
ran `VisibleTextFromHTML` over whatever came back. On a PDF that returns `false` for a real quote AND
`false` for the absent control, which reads exactly like a mistyped quote. §8g's rule is *"never
validate a citation host with an instrument other than the one that will re-check it daily"* and §8
described the tool as *"calls the production fetch + extraction"* — it called the **extraction** only.
**I obeyed the rule, using the tool the rule names, and the tool did not satisfy the rule.**

**Fixed 2026-09-04** (`9d2af8e7f`, council `9ac46a72`): `FetchCitationDocumentForProbe` exports the
production fetch and the probe calls it — one export, not a second implementation, because a mirror
that drifts from production is the failure this package warns about twice in its own comments. Proven
both ways: the CMA PDF now prints `NOT VERIFIABLE UNATTENDED: unsupported content type
"application/pdf"`, and a gov.uk page still returns true-then-false for quote-and-control.

**THE STILL-TRUE, STILL-USEFUL PART — the reading skill, which no gate gives you:**
⚠ **TWO FALSES MEANS A BLIND INSTRUMENT, NOT A WRONG QUOTE.** If your quote AND your deliberately-absent
control both return `false`, the probe discriminated nothing. Always pass the control; always read it.

**And what to do with a PDF source — the platform's own answer, which 759 missed:** keep the
`citation` (URL + verbatim quote, so provenance survives) and set **`"reverifiable": false`** plus
`staleness_days`. `refreshCitationFact` checks `reverifiable` BEFORE fetching and returns early, so
the PDF is never requested. Migration **772** did this for vetcomparison's five CMA facts, with
`staleness_days: 64` anchored on the draft's `published: 2026-07-21` — so they age out on
**2026-09-23**, the statutory deadline, and the refresher then asks for a hand re-attestation.
**The 23 September re-verification no longer depends on anyone reading this handoff.**

### 6b. ⚠ THE UNKNOWN-KEY CENSUS — the answer §8d only half-gave

`[CENSUSED 2026-09-04]` **All three live `evidence_base` writers are MAP-BASED; the typed struct is a
READ/VALIDATION path and is NOT a write path today.** Derived by intersecting every `.go` mentioning
`evidence_base` with those writing `site_specs`:

| writer | shape | keys survive? |
|---|---|---|
| `refresh_evidence_base_action.go:1695` | `eb map[string]interface{}` → `json.Marshal(eb)` | yes |
| `evidence_citations.go:449` | same; facts read back at `:320` `eb["facts"].([]interface{})`, new ones **appended** | yes, incl. per-fact keys |
| `site_admin_handlers.go:283` | writes `body.Data` **raw**; `ParseEvidenceBase` at `:225` only REFUSES a bad save | yes |

**Why this matters far beyond one key:** the landmine register warns that a typed round-trip deletes
every `citation` and `writer_line`. That hazard is real **as a shape** but has **no live instance** — so
today the fleet's registers are safe, and a future writer using `EvidenceBase` would strip citations from
**every register at once**. **⚠ This is a claim about a SET OF CALLERS, which grows by ADDITION and reads
as current for ever.** Re-run before quoting; `git log --since=2026-09-04 --diff-filter=A -- platform/orchestration/actions/`
lists what has been added since. Raised by the council's HIGH on `5d54f835` — I had verified one writer
and called it the set.

### 6c. The posture record lives in TWO places, deliberately

RFC_060's Q4 record has **no built home** (0 Go consumers, measured). vetcomparison carries it as
`site_specs.data->'posture'` (mig 761 — travels with the row it governs) **and** as a `doc_notes`
decision record (mig 767 — `subject_type='decision'`, `subject_key='<domain>'`, the platform's existing
typed, indexed convention, which 761 failed to check for). Cross-referenced, with a guard refusing a
record that disagrees with the register, and the body states the register key wins if they diverge.
**The `site_specs` key is OFFERED to the claims-verification lane, not declared as the fleet convention.**

### 6d. `IS DISTINCT FROM`, never `<>`, in a verify block

Migration 761's first cut asserted `IF nfacts <> 21`. The destructive case (`data = data || …` mutated
to `data = …`) **wipes the register**, so `jsonb_array_length` returns NULL, `NULL <> 21` is NULL not
TRUE, the body never runs, and it printed **`761 OK … register intact at <NULL> facts`**. The word
"intact" was in the output. Caught by mutation-testing my own guard. Now a LANDMINE, and every
comparison in 763/767 uses `IS DISTINCT FROM`.
**And read the mutation's OUTPUT TEXT, not its exit status** — a `NOTICE … OK` containing `<NULL>` is
this bug. `debug_historian` later found the same defect on a line of 761 I had missed (`rung <> …`),
masked only incidentally by the next check.

---

## 7. WHAT I GOT WRONG — the transferable half

1. **I verified ONE writer and called it the writer SET** (§6b). "Editing one file is not knowing the
   package", and the council caught it, not me.
2. **My first mutation test PASSED because the MUTATION was inert** — it edited a different clause of
   the same field while the needle survived earlier in the sentence. *If a mutation passes, suspect the
   mutation.* Make it assert it changed what the guard reads (`assert n==1`).
3. **I nearly recorded a VACUOUS ZERO.** `ScanUnregisteredNumbers` returned 0 on all seven non-editorial
   pages with the register loaded — and **the same 0 with the facts deleted**. Replaced with a
   disconfirmable test (control flags "the threshold is 15 first opinion practice sites", the register
   supports it, an unrelated "4,000 clients" stays flagged in **both**). **The numeric half is therefore
   ARMED BUT UNEXERCISED on this site**, and the migration says so rather than implying coverage.
4. **A single-word grep is not a check.** `grep -c cumulativ` returned 0 on the price-cap page and I read
   it as 739's correction having been lost. The correction is there and better stated than the brief
   asked — *"across the whole agreement"*. **My needle was wrong, not the page.**
5. **I stamped two migration notes "UTC" off the BST clock**, the trap `-03c` opens by naming.
6. **I read `\d doc_notes` and stopped at the line where the CHECK constraints begin**, so my first cut
   used `subject_type='site'`, which is refused. The dry run caught it; review would not have.
7. **I ASSERTED A PRODUCTION BEHAVIOUR FROM A PROBE THAT IS NOT PRODUCTION** (§6a) and wrote it into
   five documents and a landmine. It survived two council rounds because every seat read my rationale
   rather than `evidence_citations.go`. The cheap check was `grep -n "Content-Type"` on the file whose
   behaviour I was characterising. **Before asserting what a pipeline does with your input, read the
   pipeline — a probe's output is evidence about the probe until you have.**
8. **I hand-reproduced a production SQL selector minus two clauses and nearly filed a fleet-stall
   report on the result** (§4.4). Same family as the above, same session.
9. **I committed a platform change and an unrelated migration together** (`9d2af8e7f`) and drew the
   architecture-signal advisory. They are genuinely independent — the Go change is inert until a roll,
   the migration is live and does not depend on it — so no staged order is needed, but they should have
   been two commits.
10. **Migration numbers collided twice in one session** (762 and 765 taken between my max check and my
   write). Filenames are unique so nothing was lost, but **check the max immediately before naming, and
   resolve by FILENAME, never by number.**
   ⚠ **AND THE ORDER MATTERS: RENAME FIRST, RECORD SECOND.** Recording before you renumber leaves a
   `schema_migrations` row naming a file that no longer exists — provenance-by-filename then dead-ends,
   and the row reads as an applied migration nobody can find. `[MEASURED 2026-09-04, all 556 recorded
   rows]` this has happened **exactly once** on this estate: `521_tool_deployer_fork_guard_armed.sql`,
   recorded 2026-08-21 13:45:09 with no file anywhere, re-recorded as `530_…` 56 seconds later with the
   note *"renumbered from a COLLIDING 521 (race with…)"*. Benign only because it WAS re-recorded and the
   SQL self-guards. I recorded after renaming both times, so today's two renumbers left nothing.
   The sweep is worth running after any renumber — **and it must carry `notes`, not just `filename`**:
   ```bash
   kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -At \
     -c "SELECT filename||E'\t'||coalesce(notes,'(no note)') FROM schema_migrations ORDER BY applied_at;" \
     | grep -v '^$' > /tmp/rec.txt
   git ls-tree -r --name-only HEAD -- docs/agent_docs/sql_for_agents/ | sed 's|.*/||' | sort > /tmp/head.txt
   while IFS=$'\t' read -r f note; do
     grep -qxF "$f" /tmp/head.txt || printf 'STALE ROW: %s\n  ITS OWN NOTE SAYS: %s\n' "$f" "$note"
   done < /tmp/rec.txt
   ```
   ⚠ **READ THE NOTE OF ANYTHING THE SWEEP FLAGS BEFORE BELIEVING THE FLAG** — the
   `site_delivery_and_editor` lane's finding, 2026-09-04, and it is the sharpest thing to come out of
   this exchange. Their filename-only sweep flagged 521 as an orphan when **530's own note had said
   *"renumbered from a COLLIDING 521"* since 2026-08-21, three weeks earlier**. `schema_migrations.notes`
   is where sessions explain themselves; a sweep that projects it away destroys the explanation and then
   reports its absence as a finding. My first version of this snippet had the identical defect.
   (The `site_delivery_and_editor` lane owns the instrumented version of this check and its landmine;
   don't fork a second account of it.)

---

## 8. OPEN, IN THE ORDER I WOULD TAKE THEM

1. **⚠ RELEASE THE REMAINING 8 COPY REPAIRS — the canary must be verified at the bytes FIRST.**
   `/about.html` is `approved` and queued (§4.4); the other 8 are `triaged`. When the canary completes:
   ```bash
   curl -s https://vetcomparison.uk/about.html | grep -c '36 service categories'   # must become 0
   # then the sentence-identity diff, pre-image from page_component_history's op='delete' row;
   # ORPHANED sentences must be 0 (NOT "nothing reworded" — §3)
   kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c \
     "UPDATE site_work_items SET status='approved'
       WHERE created_by='bugfix_414 register-programme lane (migration 771)' AND status='triaged';"
   ```
   ⚠ **AND THE 48-HOUR REAPER IS RUNNING ON ALL NINE** — filed 2026-09-04 11:32, so unreleased items
   flip to `unresolved` (which reads as *processed*) around **2026-09-06 11:32**. Do not let them
   silently expire.
2. **READ TWO COUNCIL ROUNDS IN FLIGHT.** `5d54f835` (resubmitted 11:09, answering 761's REVISE with
   the §6b census) and `9ac46a72` (the fcaquotecheck fix). Bound the first by time — `RESUBMIT_CORR`
   reuses the correlation, so `ORDER BY created_at DESC LIMIT 1` returns round 1 (§4.2).
3. **`742` STILL OWES A RESUBMIT** — `RESUBMIT_CORR=0d730d51-a923-4b44-a58f-ab8c898d7e22`. Unchanged
   from `-03c` §0b(ii); its REVISE landed 2026-09-03 14:47 and one objection (the liveness predicate)
   was already fixed by migration 744.
4. **THE fcaquotecheck FIX IS INERT UNTIL A ROLL.** It is Go (`9d2af8e7f`). Until the next fleet build
   the OLD blind probe is what a register author runs — so if anyone builds a register before then,
   tell them to check `curl -sI <url> | grep -i content-type` by hand.
5. **WATCH THE ABSENCE CHECK'S SECOND TICK** (§4.3) — first live evidence its dedup arm works. Expect
   `missing_total=11, already_open=11, filed_new=0`.
6. **THE PRODUCER-SIDE GAP, larger than this lane: nothing enforces `spec.mode` on `content_rewrite`.**
   The next item minted anywhere on the fleet without it hits the identical destructive regeneration.
   `bugs_open/178`. Named as the top follow-up in 743's header; still true, still unowned.
7. **THE 11 REGISTERS** (§5) — and the first `standard`-rung site would finally exercise D4's cheap bar,
   which is still the only part of D4 never walked.
8. **RUNBOOK §8/§8g NEED THE §6a CORRECTION.** They still describe `fcaquotecheck` as calling "the
   production fetch + extraction" (true only since 2026-09-04) and §8g's census should gain the
   two-falses tell. That file belongs to the lendzy lane — contribute, do not fork.

## 9. WHERE THE RECORD LIVES

| what | where |
|---|---|
| this lane's technical log | `bugfix_414_planted_marker_as_claim/NOTES_planted_marker_as_claim.md` **§(s)** |
| the owner's plain-prose history | `bugfix_414_planted_marker_as_claim/README_where_we_are.md` (2026-09-03 evening) |
| the four rulings + tiering design | `architecture_review/RFC_060_…md` §3g, §3g(i)–(iii) |
| vetcomparison's register / posture / fixes | `sql_for_agents/759_`, `761_`, `763_`, `767_` (+ each `_ROLLBACK`) |
| loancash's restoration | `sql_for_agents/743_` (745 is a superseded stub) |
| the vet lane's handover + what we sent back | `vetcomparison/NOTES_vetcomparison.md` (2026-09-03 entries) · `vetcomparison/ATTESTATION_2026-09-03_…md` |
| the method | `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8 · §8b · §8c · §8e · §8f · §8g |
| landmines added | *a PDF source fetches 200 and matches nothing* · *a verify block comparing with `<>` passes the destructive case* |
| wrong calls added | 2026-09-03: the stale sketch · the inert mutation · the vacuous zero · the BST timestamps |
