# HANDOFF 2026-09-03b — `bugfix_414` lane, and the evidence-register programme it spun out

> ## ⚠ SUPERSEDED — read `HANDOFF_2026-09-03c_continue_here.md`
> Written 13:00 UTC and overtaken the same afternoon by four owner rulings (RFC_060 §3g): vetcomparison
> gets a register, loancash's three sentences are to be repaired, build the absence check AND populate,
> and a register for EVERY site at a lower bar for ordinary ones. Three of the four are now built. §5 of
> this file ("three decisions outstanding") is therefore DISCHARGED — kept for the trajectory only.

**Supersedes `HANDOFF_2026-09-03_continue_here.md`** (written 10:34 BST the same day, and stale in
three rows within the hour — that staleness is itself recorded, in NOTES §(q)).
Read this one. The earlier three are kept for the trajectory, not for state.

**All times UTC** (the database clock). The session clock is BST, one hour ahead — this cost real
confusion earlier in the day, so every figure below is stamped UTC.

---

## 0. THE ONE-LINE STATE

**`bugs_closed/414` is CLOSED and finished. The register programme it spun out is now the live
work, and its finance half is COMPLETE — all five sites have registers as of today.** One thing is
genuinely in flight (another lane's demand control, fires at 09:10 UTC tomorrow), one thing is
awaiting a peer's reply (vetcomparison), and **one thing needs an owner decision** (§5).

---

## 1. WHAT THIS SESSION DID, in three strands

### 1a. Closed out 414's last loose end — by discovering it was already being done

The previous handoff's §3 owed one item: a **demand control** proving the new
`invalid_banned_claim_pattern` detector *files* an item rather than merely being present in the
binary. It said the item was owed by the `claims-verification` lane.

It was already planted when I looked. `buytoletcalculator.uk` (`sites.status='test'`, **0 pages**,
site `dc7a8ebf-9c23-45e7-970e-32147615bb12`) has a register written **09:34:48** by
`created_by='claims_verification_probe'`: one `banned_claims` entry, pattern `guaranteed(`
(unterminated group), spec row `623c1de8-6893-4700-b0c6-88f177cb955c`. That is **four minutes** after
this lane's landmine `71b85fcc2` told them non-production sites are marked by `sites.status`, not by
a fake-looking domain. **The landmine channel moves in minutes when the receiving lane is live** —
worth knowing.

I then found and relayed a trap in their path, and **it has since been resolved by the 13:28 build** —
see §3.

### 1b. Populated `loancash.co.uk` (owner-directed) — the last of RFC_060 §1d's five

**Migration 738**, applied **~12:45**, **council-APPROVED `cf7470b7-d922-4e2d-aa84-b7aae489cadd`,
round 1, all reviewers**. `docs/agent_docs/sql_for_agents/738_loancash_evidence_base_price_cap_complaints_and_breathing_space.sql`
(+ `_ROLLBACK`). **19 facts** (11 FCA Handbook rules, 8 statutory) + **6 `banned_claims`**, 3 facts
carrying `corrects_site_citation`. Fleet registers **27 → 28**.

**Why this site was the riskiest of the five, and it is structural:** the other four calculate or
compare; loancash **explains the rules themselves** to people in financial difficulty. `[MEASURED
2026-09-03]` its 30 served pages carry **338 regulatory-shaped sentences** (crawled at the artefact,
invented-URL 404 control confirming it is not a parked domain), of which **three** cite a rule
number. The rest state 0.8%, £15, 100%, 8 weeks, 6 months, 60 days, 3%/month in plain English with
nothing re-checking them.

**It discharges an ask already on file.** `loancash_couk_fca_validation/README_where_we_are.md`
(2026-08-11) verified the three price-cap constants by hand, then wrote: *"nothing is checking …
what actually earns its keep is something that reads the rulebook and shouts if it disagrees with
what is on our page."* That is this mechanism, three weeks late. That lane also named the
complaint-deadline tool as its highest-value unchecked item *because limitation periods move* —
`DISP 2.8.2(1)`, `(2)(a)`, `(2)(b)` and `DISP 1.6.2` are registered for exactly that reason, so its
`[UNMEASURED]` marker is now discharged.

**Three wrong live claims found. THE SERVED COPY IS NOT TOUCHED — that is the owner's call (§5).**

1. **CONC 5A.2.14 — the £15 default cap is CUMULATIVE across the whole agreement**, "whether in
   relation to one breach or cumulatively in relation to multiple breaches". Two pages frame it as
   *per missed payment* (`guides/the-payday-loan-price-cap.html`, `guides/jargon-buster.html`).
   **The site understates the protection it exists to explain**: a reader with two missed payments
   would accept a second £15 as lawful. It is not. **This is the one that matters.**
2. **CONC 7.6.12 — the CPA limit is TWO REFUSED requests, and `£1` appears nowhere in CONC 7.6.**
   `guides/loan-sharks-and-illegal-lending.html` says "cannot take more than one payment attempt of
   over £1". The site's **own** `guides/stopping-payments-the-cpa-rules.html` states it correctly —
   an internal contradiction on one page, not a house error.
3. **CONC 5.2A.4 — affordability is CONC 5.2A, not "CONC 5A"** (the price-cap chapter, which has no
   affordability rule). `guides/check-your-lender-is-authorised.html`. Cited correctly on 3 other
   pages.

**The base rate now holds for a fourth lane: lendzy 2 · loanzy 1 · loancalculator 2 · loancash 3.
Every lane that has read the primary source has found errors in its own site's live copy.**

### 1c. Found that the register-absence gap is structural, not a backlog item

`[MEASURED 2026-09-03 13:58]` **12 of 39 `deployed` sites hold no current `evidence_base`** (13
before loancash): `advertise.co.uk`, `cookly.uk`, `cv1.co.uk`, `designblog.co.uk`, `garden-tools.uk`,
`homegarden.uk`, `idea.uk`, `lampenkap.com`, `oxenunity.com`, `seotools.co.uk`, **`vetcomparison.uk`**,
`websitepromotion.co.uk`.

⚠ **And nothing can ever raise the absence.** `resolveEvidenceSites`
(`platform/orchestration/actions/refresh_evidence_base_action.go` **:281**, fleet query **:291**)
builds the daily sweep's target list as
`SELECT site_id FROM site_specs WHERE aspect='evidence_base' AND is_current` — **it selects the sites
that HAVE a register. The target set is defined by the presence of the very thing whose absence is
the defect.** A register-less site is invisible to the freshness sweep, the fact checks and the new
`invalid_banned_claim_pattern` detector alike, permanently; running the check more often would never
reach one. No `item_type` for a missing register has ever been filed (`%evidence%`/`%register%`/
`%claim%` yields only `claims_unverified` 47, `stale_evidence` 10, `spec_supplies_claim` 2,
`stale_directory_claim` 2 — every one presupposing a register exists). **Q1 requires registers; no
reader enforces or reports the requirement.** Recorded in RFC_060 §1d, not built — it is that RFC's
build and the claims-verification lane's seam.

---

## 2. VETCOMPARISON — asked, awaiting reply, DO NOT START WITHOUT IT

I messaged the live `vetcomparison [0d8f85]` session at ~13:50 asking two things: **(1) is
vetcomparison.uk's register yours or unowned?** and **(2) is there anything about the site that
would make a register wrong for it** (e.g. a comparison site whose numbers are other people's claims
rather than its own), plus what it actually asserts about RCVS / VMD / medicines / pricing.

**No reply had arrived when this handoff was written.** Check for it before doing anything —
`ListAgents` shows the session; a reply arrives as a `<cross-session-message>`.

**Why it is now live rather than backlog:** RFC_060 **Q5 was ruled 2026-09-03 and the ruling
INVERTED the RFC's own recommendation**, because the owner supplied the fact it turned on — *"I will
be extending to vet and legal quite soon."* So sector **presets** are in, veterinary is explicitly
one, and the Q5 build landed today (`939593e4c`: per-site `citation_codes` + named sector presets).
**The presets just approved will land on a site with no register to apply them to.**
vetcomparison.uk is already named at RFC_060 §3b as "none — 0 facts".

**If it comes back unowned and the owner wants it done**, the method is
`docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` **§8** (and §8b/§8c/
**§8e**, the last of which I added today). Budget about half a day. **Do not skip step 4** — see §4.

---

## 3. IN FLIGHT: the 09:10 UTC sweep tomorrow now has TWO arms

A **fresh chassis build rolled at 13:28** — replicaset **`85c4984f77`**, replacing `75b987cbd7`
(pods 13:28:18 / 13:28:43).

`[MEASURED 2026-09-03 13:56]` binary probe of `/proc/1/exe`, **both** replicas, **both** controls:

| symbol | nrqf7 | phgh2 | note |
|---|---|---|---|
| `patterns_checked` | **1** | **1** | **NEW — was 0/exit 1 on the 08:57 build** |
| `invalid_banned_claim_pattern` | 6 | 6 | detector, unchanged |
| `stale_evidence` (must be present) | 6 | 6 | control holds |
| `zzz_not_a_real_symbol_qx7` (must be absent) | 0 | 0 | control holds |

**This reverses an inference and retires the trap I relayed this morning.** `996b40542` added an
always-fired Info line at `refresh_evidence_base_action.go:431`; until 13:28 it was NOT in the
running binary (committed 09:29, pods started 08:57 — arithmetic and artefact agreed). It is now
live, so **its absence genuinely means the code did not reach `:431`**, and its presence with a
non-zero `patterns_checked` proves it read real register data. The CONTRIB in the claims lane's dir
carries a dated correction saying so; both lanes have been told.

**`evidence-freshness` last completed 09:10:23 and has NOT run since** — so no pass has happened
under any binary since the probe was planted at 09:34. **Nothing has been tested yet; nothing has
failed.** `invalid_banned_claim_pattern` items: **0**. Probe spec: still `is_current`.

**What to look for after 09:10 tomorrow:**

- **WRITE path (`buytoletcalculator.uk`, the probe)** → expect `patterns_checked=1, invalid=1` in the
  log **and one `invalid_banned_claim_pattern` work item**. This arm has never been exercised outside
  mocks: `createInvalidBannedClaimPatternItems` is reached only from `:713`, gated on a non-empty
  finding.
- **READ path (`loancash.co.uk`, free, needs no plant)** → expect `patterns_checked=6, invalid=0`.
  **A non-zero count on a site nobody touched is the demand control for "ran and found it clean"** —
  the exact state `omitempty` used to erase (`:216`, `:221`).
- **If loancash logs 6 but the probe files nothing**, the defect is isolated to the write path rather
  than the check — a far narrower thing to debug.

**Owner of that experiment: the `claims-verification` lane.** They also owe reverting the probe spec
once they have an answer — it is harmless (0 pages, nothing served) but is now a live register inside
the fleet sweep, re-read daily until removed. **Not this lane's to run or revert.**

---

## 4. THE FIVE THINGS WORTH CARRYING (superseding the old §5, which stands but is narrower)

1. **A hand-transcribed citation quote fails silently and for ever.** The `DISP 2.8.2(2)(b)` quote
   was first written with commas where the source has parentheses — indistinguishable to a human eye.
   It returned **false** on the production matcher. Shipped, it would have read as `citation_lost`
   drift **every single day**, and that false alarm is indistinguishable from a real one.
   `go run ./cmd/fcaquotecheck <url> "<quote>" "absent control"` caught it in the run that was meant
   to be a formality. **Paste the quote; never retype it. Step 4 is not optional.**
2. **A "shared" set can exist in more than one WIDTH, and a coverage count cannot see it.** The
   finance `banned_claims` set drifted while being copied site to site: lendzy carries a bare
   `\bno credit checks?\b`; loanzy/loancalculator require the product noun. `[MEASURED]` the bare
   variant fires on loancash's **correct** advice that an employer salary advance involves "no
   interest and no credit check" (**1** hit); the narrow one fires **0** across 30 pages. *"Does this
   site carry the shared set?"* returns the same answer either way. **Run any inherited pattern over
   the target's own served pages, require 0 hits, and include a positive control proving it is not
   inert.** Now a LANDMINE.
3. **`created_by` on a current row names the last WRITER, not the author of the values.** A refresher
   that supersede-and-reinserts relabels every field it merely preserved — including
   `banned_claims`, which it has **no code path to author**. Read the aspect's **history**, not its
   current row. This nearly reached a handoff as "the daily sweep fixed farmerinsurance's gap on its
   own"; the truth is migration 713, the previous evening. Now a LANDMINE, verified `87ea5043`.
4. **A guard nobody has watched fail is not a guard.** Both of 738's guards were mutation-tested
   against the live DB before applying (fact count 19→18 aborted; guard predicate flipped aborted
   before the INSERT), then the whole file was run with `COMMIT` swapped for `ROLLBACK`. The council
   still found a real hole neither mutation could reach: **nothing tied the site_id to the DOMAIN**,
   and a mistyped UUID would populate another site while **every count in the verify block still
   passed**, because they are all scoped to the same wrong id. Fix now in RUNBOOK §8e.
5. **A detector whose target list is drawn from the population it checks cannot find what is missing
   from that population.** §1c. This is §3's `omitempty` blindness one turn further out: there the
   instrument erases a *clean result*, here it never sees an *absent subject*, and no cadence change
   would help.

---

## 5. ⚠ OWNER DECISIONS OUTSTANDING — there are THREE

**(1) The three wrong claims on loancash.co.uk: repair the copy, or leave it?**
The findings are recorded in the register (`corrects_site_citation`) and the served pages are
untouched, per the 695 precedent and `bugs_open/320` §15 — rewriting published prose on an automated
finding is authority you withheld. **The £15 default-fee one is the one I would repair first**: the
site currently understates a borrower protection, so a reader who missed two payments would accept a
second £15 charge as lawful when it is not. The other two are a contradicted page and a wrong rule
number. *If you want them fixed, that is a content change through the framework, not a hand-edit.*

**(2) `vetcomparison.uk`'s register — do it, or hold?**
Blocked on the peer reply (§2), but the decision is yours either way, because Q5's presets are built
and veterinary is the sector you named as next. Doing it costs about half a day and, on the base rate
of four lanes out of four, will probably turn up errors in that site's live copy too.

**(3) The register-absence gap: 12 deployed sites, and nothing can detect it (§1c).**
Q1 makes registers required; no mechanism enforces or even reports it, structurally. Two shapes of
answer, and this is the one that needs your steer rather than a lane's:
   - **(a) Populate, site by site** — reliable, roughly half a day each, and every lane so far has
     found real errors, so it has a second payoff beyond compliance.
   - **(b) Build the missing detector** — a check that starts from the list of **live sites** rather
     than the list of registers, and files a work item for a `deployed` site with none. Small, and it
     converts a permanent blind spot into a queue. It is new platform surface, so it wants an RFC
     round or at least a council submission.
   My recommendation is **(b) then (a)**: without (b) this recurs silently on every new site, and (b)
   is the cheaper of the two. But the scoping question inside (b) is yours: **is a register required
   on every deployed site, or only on sites at a compliance tier?** Q1 answered that only for
   finance. Most of the 12 are not finance, so a detector scoped "every deployed site" would file 12
   items today, of which perhaps 2 are real.

**Nothing else waits on you.** 414 is closed; the register programme's finance half is complete; the
09:10 experiment is another lane's.

---

## 6. HOW TO RE-DERIVE EVERY FIGURE ABOVE

```bash
# the register, at the live row, read back through a JOIN on sites (never the uuid you typed)
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SET statement_timeout='25s';
  SELECT s.domain, jsonb_array_length(ss.data->'facts') facts,
         jsonb_array_length(ss.data->'banned_claims') banned, ss.created_by
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='loancash.co.uk';"

# deployed sites with NO register (the §1c population)
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT s.domain FROM sites s
  LEFT JOIN site_specs eb ON eb.site_id=s.id AND eb.aspect='evidence_base' AND eb.is_current
  WHERE s.status='deployed' AND eb.id IS NULL ORDER BY s.domain;"

# the 09:10 experiment: did the sweep run, and did either arm fire?
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SELECT now()::timestamp(0), last_completed_at::timestamp(0) FROM scheduled_tasks WHERE name='evidence-freshness';
  SELECT count(*) FROM site_work_items WHERE item_type='invalid_banned_claim_pattern';
  SELECT count(*) FROM site_specs WHERE id='623c1de8-6893-4700-b0c6-88f177cb955c' AND is_current;"

# is a symbol in the RUNNING binary? ALWAYS both controls, EVERY replica
kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers \
  -o custom-columns=NAME:.metadata.name,RS:.metadata.ownerReferences[0].name,START:.status.startTime
for POD in <each pod>; do for SYM in patterns_checked stale_evidence zzz_not_a_real_symbol_qx7; do
  kubectl -n ai-persona-system exec "$POD" -- grep -ac "$SYM" /proc/1/exe; done; done
# NEVER `strings` (absent from the debian-slim images) and never a DISCOVERY grep for "some 40-hex
# string" — it matches Go's internal digit table and answers the same on every service.

# verify a citation quote the way production will, before you ship it
go run ./cmd/fcaquotecheck "<url>" "<verbatim quote>" "zzz deliberately absent control qx7"
# expect true then false. A false on the real quote means it would be daily citation_lost drift.

# the council verdict for 738
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "
  SET statement_timeout='25s';
  SELECT metadata->>'decision' FROM diagnosis_artifacts
  WHERE correlation_id='cf7470b7-d922-4e2d-aa84-b7aae489cadd' AND kind='council_report';"
```

⚠ **The postgres pod was intermittently timing out `kubectl exec` queries this afternoon** (council
runs + landmine verifiers + other lanes). If a query hangs, `SET statement_timeout='20s'` and retry
rather than concluding anything from the silence.

---

## 7. WHERE THE RECORD LIVES

| what | where |
|---|---|
| the bug | `bugs_closed/414` — CLOSED, fix live, verified at the artefact |
| this lane's technical log | `bugfix_414_planted_marker_as_claim/NOTES_planted_marker_as_claim.md` §(q), §(r) |
| this lane's owner prose | `bugfix_414_planted_marker_as_claim/README_where_we_are.md` |
| the register programme | `architecture_review/RFC_060_…md` §1d (re-measured), **§1e** (loancash + its two traps) |
| loancash's register | `docs/agent_docs/sql_for_agents/738_…sql` + `_ROLLBACK`; submission JSON in the lane dir |
| loancash's lane | `loancash_couk_fca_validation/` — NOTES (method, findings, verdict dispositions) + README (owner prose) |
| **the method, for the next register** | `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8 · §8b · §8c · **§8e (new today)** |
| the relay to the code's owner | `claims_verification/CONTRIB_2026-09-03_from_414_…md` (+ its dated correction) |
| landmines added today | `LANDMINES.md` — *a refreshed spec's `created_by` names the last WRITER* · *a "shared" `banned_claims` set exists in more than one WIDTH* |
