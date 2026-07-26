# vetcomparison.uk — CONTINUE HERE (2026-07-26 22:10 BST)

Cold-start entry point for the next session. Written by session "bugfix 61" after closing
`bugs_closed/061`. Read this, then `PLAN_2026-07-26_site_strength.md` (the direction, written
by a *different* session at 19:32 tonight), then `SUMMARY_2026-07-26_readout.md`.

---

## 1. OWNER DECISIONS RECEIVED TONIGHT — act on these

1. **"Yes, please scale the company number crawl past the pilot."** The gate in
   `PLAN_2026-07-26_site_strength.md` §P1 step 4 ("scale to the 2,109 published practices with
   the owner's nod") is **ANSWERED — the nod is given.**
   **This does NOT mean skip the pilot.** Pilot on ~25 first and *report the hit rate*; the
   pilot exists to measure, not to ask permission. Scale straight after without stopping.
2. **The 5 standing review items were presented** (§3 below) and are with the owner.
3. **`/claim-listing` and `/search` destinations** — investigated, findings in §4. Not decided.

## 2. ⚠️ OWNERSHIP — CHECK BEFORE YOU TOUCH P1

`PLAN_2026-07-26_site_strength.md` belongs to **another session**, committed `f804b84ed` at
**19:32 tonight**. At the time of writing it had no follow-on commits, but **uncommitted work is
invisible** — a session mid-crawl would not show in `git log` or `who-owns.py`.

**Before applying seed 082 or starting any crawl:**
```bash
git log --oneline --since="6 hours ago" | grep -iE "vetcomparison|company.?number|ch-company"
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT type, is_active FROM agent_definitions WHERE type='ch-company-scraper' AND deleted_at IS NULL;"
```
If the agent now EXISTS, that session has already applied the seed — **do not re-apply**; find out
how far the crawl got before adding to it. Two sessions crawling 2,109 sites in parallel is the
collision this check exists to prevent.

## 3. THE 5 STANDING REVIEW ITEMS (as presented to the owner)

All `status='needs_human_review'` on `vetcomparison.uk`, all created 2026-07-17. Every one maps
onto a page row that is `planned` with **0 sections** — they are the same fact recorded twice.

| id | item | page row | my recommendation |
|---|---|---|---|
| `e30dc7b9` | **HIGH** — `tool-compliance-deadline-calculator` not_built, "needs owner-aware build, not the generic builder" | `/tools/compliance-deadline-calculator/index.html` *planned, 0 sections* | **BUILD, relative-only.** P4 already answers it. HARD RAIL: every figure and date in the draft Order is a bracketed placeholder (`[£21]`, `[£12.50]`, `[X March 2027]`), so it may say "six months after the Order is made" and must **never** print a calendar date. Publishing a provisional date as fact is the exact thing this site was remediated for |
| `715ec305` | Build **directory-index** (not_built) | `/directory/index.html` *planned, 0 sections* | **BUILD FIRST of the three.** It is the only one whose data is fully publishable today (2,109 practices × 6 fields), and it is where `/search` should point (§4) |
| `3cce980c` | Build **practice** page (not_built) | `/entities/practice.html` *planned, 0 sections* | **HOLD until P1 lands.** Today a practice page would carry name + postcode + outbound link and nothing else. The company-number crawl is what gives it something to say |
| `2f50bfda` | Build **guides-index** (not_built) | `/guides/index.html` *planned, 0 sections* | **BUILD — cheapest of the five.** Three sourced guides already exist and are live; this is an index over content we already publish, no new data and no provenance question |
| `f98dc52e` | `contact-info` section on contact needs "Business contact email address" | `/contact.html` *needs_rebuild, 3 sections* | **ALREADY ANSWERED BY DATA — probably resolvable without the owner.** `sites.email` for vetcomparison.uk now reads `vetcomparison@contactforsales.com` (owner registered these fleet-wide on 07-26; see the 063 case). The page is `needs_rebuild`, so a rebuild should pick the fact up. **Verify the rebuilt page renders the address before resolving the item** — trust the rendered artefact, not the row |

## 4. `/claim-listing` AND `/search` — WHERE THEY SHOULD POINT

Both are **live 404s linked from the homepage**, measured tonight. The full homepage 404 set is
five, not nine: `/claim-listing`, `/search`, `/about-ownership-disclosure`, `/about-pricing`,
`/guides/pet-owner-rights`. Everything else the homepage links resolves 200.

**`/search` → `/directory/index.html`.** This needs no new decision and no new data: the
`entity-directory` page row already exists (`planned`), the directory export already publishes the
six fields it needs, and 2,109 practices are already live-safe. It is review item `715ec305` under
another name. Cheapest real win on the site.

**`/claim-listing` → a claim flow that is already designed and already has its database.**
`business_intel.claim_requests` **exists** (created by this workstream's `010_claim_requests.sql`,
generic across verticals, keyed to `business_intel.businesses`). It snapshots the consent text per
row rather than by reference, precisely so a business can later be shown exactly what it agreed to,
and it logs opt-outs on the same trail. So the destination is not a decision about *what to build*
so much as *how far*:
- **Minimum honest version:** a page explaining what claiming a listing means and what we do with
  it, with a contact route — no form. Costs nothing, kills a 404, tells the truth.
- **Full version:** the form writing to `claim_requests`, which is the licence under which we would
  ever publish a practice's own prices. That is the commercially important one, because a claimed
  listing is the *only* route by which practice prices become publishable under our provenance rule
  (762 held price rows are barred today for having no source URL).
- **Owner call needed on:** whether we solicit claims at all before the CMA Order lands, and how a
  claimant is verified as representing the practice. Do not build the form without answering the
  second — the whole value of the table is that verification is recorded.

## 5. WHAT I DID THIS SESSION (complete, nothing left)

`bugs_closed/061` **CLOSED** — commit `46cec068e`; summary `SUMMARY_2026-07-26_readout.md`
(`3927b2f7a`). Fix was already live in `v1.0.1165`; I proved it by inducing the fault. **The
guard caught a live fabrication:** the AI invented £19.25/£34.99/£68.75 on the original Advocate
page and all three were dropped, storing nothing. Full-table sweep 2,583 OK / 0 PRICE_ABSENT.

**Carry this forward — it is the transferable bit:** I had 16 consecutive post-fix AI calls
returning `[]` and wrote in an approved plan that the guard was redundant and *could not be
induced*. It fabricated on the first attempt. A quiet failing branch is not a removed failure mode.

## 6. OWED / OFFERED, NOT DONE

- **`bugs_open/082` — a concrete trigger, unrecorded.** That bug (BestEffort postgres + 1s
  `pg_isready` probe → killed whenever the node is busy) says its trigger "belongs to whoever owns
  the ollama lane". I found one: **`ollama-adapter` and `postgres-clients-0` are on the same node**
  (`prod-instance-17735925437536833`, verified), and the med-scrape AI fallback held that node's
  CPU for **495 s** tonight while the collector logged `Database ping failed` every 30 s from
  14:55:04 to 15:00:04 — 082's own quoted postgres restart (14:56:59) falls inside that window.
  **[INFERRED]** the CPU-saturation step: no historical node metrics exist for that window, so this
  is co-location + timing + 082's established mechanism, not a measurement.
  **I have not touched 082 or its file** — another session's, deliberately unpatched in production.
  The owner was offered this and has not yet said yes. It also upgrades the `[OBSERVED]` throughput
  note in `bugs_closed/061` §Residual from "slow" to "may knock over the shared database".
- **A fresh chassis build (`v1.0.1171`) is deployed** — the 5 `ch-*` agents already carry it.
  Nothing in 061 depends on it; 061 was proven live on `v1.0.1165` and its code is unchanged since.

## 7. LANDMINES FOR WHOEVER PICKS THIS UP

- **Read seed 082 to the END before applying it.** The plan's own author was wrong about it once
  and corrected it in-file: an earlier seed (`037`) shipped an unblanked domain and became a live
  reinfection vector via `ON CONFLICT DO UPDATE`. Blank anything site-specific; check `image_tag`.
- **Spawned worker pods run `agent_definitions.image_tag`, NOT the deployed image.** Pod-grep the
  *spawned* pod. This is how 061 was verified and it is the trap that makes "we rolled it" false.
- **A seed is DB config — live immediately, no build, no roll.** That is why P1 is cheap.
- **The med scraper keeps running 6-hourly** and its AI fallback still fabricates on category
  pages — that is *fine and expected*, the guard drops them. Do not read a `fidelity guard dropped
  variant` log line as a regression; it is the system working.
- **Never purge med price data by price value** — the model invents fresh values each run, and
  genuine rows share values with fabricated ones. Only the evidence check separates them.
