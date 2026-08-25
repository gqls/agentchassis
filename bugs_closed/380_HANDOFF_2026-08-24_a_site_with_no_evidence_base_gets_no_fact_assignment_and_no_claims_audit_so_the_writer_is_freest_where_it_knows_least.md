# 380 — a site with NO evidence base gets no fact assignment AND no claims audit: three mechanisms degrade to "no constraint" on one condition, so the writer is freest exactly where it knows least

**Filed 2026-08-24** by the `loanzy_uk_example_site` lane (one-shot build route), from the completed
greenfield `garden-tools.uk` build. **Status: OPEN, unowned. Severity: HIGH** — it is the mechanism
behind the `loanzy.uk` invented-credit-broker incident and behind this build's invented review
methodology, and it fails silently and reports success.

> **On the 2026-07-31 owner ruling** (a cross-cutting root cause is not "filed" until it has been
> through `090`, or the filer states why they substituted first-hand verification): **substituted.**
> The cause is read directly from live `agent_definitions` config — the auditor's own branch
> condition and the planner's own prompt template, both quoted verbatim below — plus a live site
> that exhibits all three consequences. There is no hypothesis for a loop to narrow: the config IS
> the mechanism. What `090` could add is estate-wide blast radius, left open in §6.

## 1. Symptom, in the owner's words

A greenfield build produced a 1,486-word `how-we-assess` page describing a product-review
methodology in the present tense — *"We record the metal used…"*, *"We describe the steel, the handle
material, and the grading standard"*, *"Where we can, **we buy the tool at the same price a reader
would pay**, from the same retailers we link to"*, *"Manufacturers do occasionally send review
samples"*, and on `about`, *"**We garden ourselves, and we test what we can get our hands on**"*.

**None of it has happened.** No tool has been tested, no tool has been bought, no manufacturer has
sent anything, and there is no team that gardens. The owner's ruling: *"I agree that we aspire to
these claims but we don't and haven't actually done any of it… we need to stop this sort of
hallucination."*

**The sharpest form of it:** `how-we-assess` is the **largest page on the site**, and it describes
how we assess products on a site with **zero products** — every buying guide, brand directory and
brand profile failed to build (`bugs_open/206`). The methodology page outlived its subject matter.

## 2. The mechanism — three independent degradations, one shared condition

`garden-tools.uk` has **no `evidence_base` spec** `[MEASURED 2026-08-24]`:
```sql
SELECT aspect FROM site_specs WHERE site_id='16784842-…' AND aspect ILIKE '%evidence%';  -- 0 rows
```

Three separate mechanisms key on that, and **all three fail open**:

**(a) Nothing in the greenfield path CREATES one.** `build-site-planner.plan_site` only ever *reads*
it. There is no step, on any agent in the greenfield chain, that mints an evidence base from
research. (Many other sites have one — `apis.uk`, `agritec.uk`, `oufe.com`, `dartsonline.com` and
others — so the artefact is real and populated elsewhere; this route simply never produces it.)

**(b) The planner tells the writer it is unconstrained.** Verbatim from `plan_site`'s prompt
template, the `else` arm that fires when the evidence base is absent:
> *"No verified facts are registered for this site — use plain string section entries and no facts keys."*

The fact-assignment machinery (RULES rule 17: every section carries an explicit `facts` list, each
fact stated in exactly ONE place) is **switched off wholesale**. The writer receives no roster to
stay inside.

**(c) The claims auditor SKIPS, and reports `complete`.** Verbatim, `claims-auditor.check_opted_in`:
```json
{ "action": "conditional_branch",
  "config": { "condition": "evidence_facts.facts_text",
              "then_step": "load_page_text",
              "else_step": "complete" },
  "description": "Skip entirely when the site has no evidence base" }
```
No evidence base → branch straight to `complete`. **Not skipped-with-a-warning; skipped as success.**
No work item, no `doc_notes` row, nothing on the site record. `garden-tools.uk` has **zero**
claims-related work items of any kind, ever.

## 3. Why this is an INVERTED safety property, and not merely a gap

Read the three together: **the amount of claim-checking a site receives is proportional to how much
verified material it already has.**

A site with a rich evidence base — one that *knows* things, and whose writer therefore has true facts
to reach for — gets fact assignment AND an audit. A site with no evidence base — one that knows
**nothing**, and whose writer must therefore invent to fill a page — gets **neither**. The gate is
weakest at exactly the moment invention is certain rather than merely possible.

**That is the whole of the `loanzy.uk` incident too.** A bare domain name produced a regulated credit
broker with a lender panel; the remedy shipped then (`CGV-032`, migration 464) was a classifier-level
rule about *regulated business models*. It did not touch this: the classifier can be right about what
the site should BE and the writer will still invent what the business has DONE. **`CGV-032` gates the
vertical; nothing gates the practice claims.**

## 4. What the writer produced when unconstrained — the shape, not just examples

The copy is internally inconsistent, which is the tell that no single authority governs it. The same
page hedges honestly in some paragraphs and asserts practice in others:

| honest, present on the page | asserted as practice, on the same page |
|---|---|
| *"no amount of research replaces trying a tool in your own soil, and we can still get a call wrong"* | *"Where we can, we buy the tool at the same price a reader would pay"* |
| *"Where we have not used a tool directly, we say so"* | *"We garden ourselves, and we test what we can get our hands on"* |
| *"Every figure here comes from a manufacturer's stated specification, a published standard, or a retailer's own listing"* | *"We record the metal used… the handle material… and the stated weight"* |

**Both voices are LLM output from the same run.** The writer is not lying so much as writing the page
a real review site would have; nothing tells it which sentences are checkable and which are wishes.
**The FAQs carry the same defect** (owner) — *"Questions about how we test"* answers a question about
testing that has not occurred.

## 5. Fix candidates, ordered by what closes the door

1. **Make the auditor's skip LOUD, not silent (smallest, closes the reporting hole).** `else_step`
   should file a work item / `doc_notes` row — *"claims audit skipped: no evidence base"* — not
   `complete`. Today the absence of an audit is indistinguishable from a passed audit at every
   observation point. **This is the one that makes the other two visible**, and it is a one-branch
   change.
2. **Gate PRACTICE claims independently of the evidence base.** A first-person-plural present-tense
   assertion about what the operator *does* (`we test/buy/record/compare/measure/receive`) is
   checkable against nothing today. It needs its own rule, because it is exactly the class an empty
   evidence base cannot cover — the evidence base holds facts about the WORLD, and these are claims
   about US. Cheap detector, high precision: the seven verbs above in first-person present, on any
   page of a site with no operating history.
3. **Mint an evidence base on the greenfield path, even an empty-but-present one.** An explicit
   *"this site has no verified facts and no operating history"* record would flip (b) and (c) from
   fail-open to fail-closed without changing either. Bigger change; the right destination.
4. **Not a fix: telling the writer to be careful in its prompt.** `content_direction.avoid_phrases`
   exists and is writer-side; this build already carried one and produced the copy above. A prompt
   instruction is not a control on output (house rule: *a doc comment is not an enforcement mechanism*).

## 6. Blast radius — MEASURED for this site, OPEN for the estate

`[MEASURED 2026-08-24]` `garden-tools.uk`: no evidence base, no claims work item ever, and practice
claims on at least 3 of 7 served pages (`how-we-assess`, `about`, `index`).
**SWEPT, and the number is large** `[MEASURED 2026-08-24]`:

```sql
SELECT count(*) FILTER (WHERE ev.site_id IS NULL) AS no_evidence_base, count(*) AS total
FROM sites s LEFT JOIN (SELECT DISTINCT site_id FROM site_specs
                        WHERE aspect ILIKE '%evidence%' AND is_current) ev ON ev.site_id = s.id;
-- 29 no_evidence_base | 19 has_one | 48 total
```

**29 of 48 live sites — 60% — have no evidence base**, and every one of them has had its claims
audit branch to `complete` without reading a page, for its entire life. **The audit's own records
cannot distinguish that from a clean pass**: both are `complete`, both leave no work item. So the
fleet's claims-checking coverage is not 100% with some failures; it is **~40%, silently**, and no
report anywhere says so. Fix candidate 1 (make the skip loud) is what turns this from unknowable
into a number someone can watch.

⚠ **A COUNT OF THINGS CARRIES ITS DATE (owner ruling 2026-08-22)** — the sites-with-an-evidence-base
list above is **as of 2026-08-24** and this class grows by ADDITION with every greenfield build.

## 6a. TAKEN AND PARTLY FIXED THE SAME DAY — and the fixing lane's diagnosis is STRONGER than mine

`597_claims_auditor_runs_cold_and_fails_closed.sql` **applied 2026-08-24 16:50Z**, its header citing
*"bugs_open/380 slice S2"*, alongside `598_build_site_planner_distinct_no_facts_arms.sql`. Verified at
the live config: **`check_opted_in` no longer exists** on `claims-auditor` — the skip-to-`complete`
gate is gone and the auditor is posture-changed to run cold.

**Three corrections to this file from that lane's work, all verified here:**

1. **My predicate was too narrow. Use 33, not 29.** I counted sites with *no evidence spec at all*
   (**29 of 48**). The gate branched on a `string_agg` over `facts[]`, which is **NULL for a site with
   no register AND for one with an empty register** — so the right population is *nothing attested*:
   **33 of 48** `[MEASURED 2026-08-24]`. Both numbers are correct; only theirs answers the question
   the code asks. **Counting the thing I could see rather than the thing the predicate tests.**
2. **The mechanism is worse than "the gate fails open": the auditor is essentially UNDRIVEN.** Their
   header records no seed file (migration 350 notes it), no schedule, and — at the time they wrote it
   — **one `llm_call_log` row in its entire history** (2026-07-18, returning `[]`), with
   `request_claims_review` never having fired. So this was not a guard that leaked on 60% of sites;
   it was a guard that had **run once, ever**. **I should have checked drive before framing severity**
   — my own memory index carries *"a silent mechanism is usually UNDRIVEN, not missing"* and I did not
   apply it to the mechanism I was filing about.
3. ⚠ **Their own "one row ever" figure is now stale, changed by their own verification.** Re-measured
   here after their migration: `llm_call_log` for `claims-auditor` is **3 rows** (latest
   2026-08-24 16:53:06Z) and `claims_llm%` items are **2**, not 0. The figures were true when written.
   **Anyone re-running those queries to confirm the diagnosis will get different numbers and may
   doubt a sound finding** — the fix's own test populated the table cited as empty. Same family as
   *your action moves you to the back of the selector*: **verifying a fix can erase the evidence for
   the bug.** Quote the figures with their date, or re-derive from `created_at < '2026-08-24'`.

**What remains open in this bug after S2:** the planner's fact-assignment arm (598 addresses the
`no_facts` arms — needs its own verification), and **fix candidate 2 — gating PRACTICE claims
independently of the evidence base.** The invented copy on `garden-tools.uk` is claims about *us*
(*"we buy the tool at the same price a reader would pay"*), not claims about the world, and an
evidence register of world-facts cannot cover them however cold the auditor runs.

## 7. Provenance

Live run `garden-tools.uk`, submitted 2026-08-23 17:17Z with no prompt, no mission and no seed.
Owner review of the served pages 2026-08-24 named the hallucinated claims and the FAQ overclaiming.
Lane record: `docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/NOTES_loanzy_uk_example_site.md`.
Related but distinct: `bugs_open/288` (the evidence register guards COPY, not CODE — a tool encoding a
legislated figure is checked by nothing) is the same register failing on a different surface; this is
the register being **absent entirely** and every consumer failing open.

## 6b. FIX RECORD — `bugfix_380_claims_fail_open` lane, 2026-08-24 (owning session: `bugs_open/380`)

**Taken 2026-08-24 ~15:15 BST.** Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_380_claims_fail_open/`
(PLAN / NOTES / RUNBOOK / README_where_we_are / `TRIGGER_claims_audit.sh`). Owner decisions taken in-session
(PLAN "Decisions"): **D1** no register minting or backfill — absence IS the cold posture (RFC_003 §8 Q2 = NO);
**D2** rotation 3600s/7-day; **D3** the Go practice family at `warning`, never a refusal (RFC_003 Q1 stays
open); **D4** the writer arm waits for the owner's plaintext read (RFC_016 §5.2).

**Corrections to this file, in addition to §6a's:**
- §5.1 as written ("file a work item / doc_notes row on the skip") was NOT the fix: the skip branch is
  DELETED (RFC_017 — audits fail closed; a DB error now FAILS the run) and every run leaves a doc_notes
  RECEIPT (`pipeline`/`claims-audit`) so coverage is a query. Findings still file `claims_unverified`.
- §5.3 ("mint an empty-but-present register flips (b) and (c)") is **wrong as written**: both conditions
  test the FACTS, not the row (`facts_text` is NULL on `[]`; `ParseEvidenceBase` returns nil, CLM-005).
  The fix keys on facts and fails closed on their absence; no shell registers are minted (owner D1).
- §6's "~40%, silently" was `[INFERRED]` from config. `[MEASURED 2026-08-24, pinned before 16:00Z]` the
  LLM auditor had ONE `llm_call_log` row in its life and its work-item step had never fired: the LLM
  layer's coverage was ~0% everywhere, and the auditor was UNDRIVEN (no seed, no schedule, no spawner).

**What is LIVE (config, applied + recorded 2026-08-24):**
- **597** — `check_opted_in` deleted; `load_evidence_facts.error_step` removed (fail closed); cold-register
  prompt arm (practice/possession/track-record/named-relationship classes; do-not-report list for
  could-framed, negated, quoted and industry statements); `ALLOWED ENTITIES` nil-guarded; per-page text
  cap 12,000; `recurrence_expected:false` explicit; per-run doc_notes receipts.
- **598** — the planner's two identical `{{else}}` arms are distinct; both mandate the OBJECT form with
  `"facts": []` on every section (proven end to end with no Go change: `scopeItem` scopes a non-nil empty
  list; every carrier keeps `[]` ≠ NULL; `section_facts` wired live on page-build-handler); the no-register
  arm bans briefing a methodology page as practice; rule 17's contradictory last sentence edited.
- **600** — `claims-audit-rotation` (590's shape; shipped-page predicate; `locked_at IS NULL`; stamp).
- **601** — the extraction defect the proof found: a PostgreSQL ARE regex takes the greediness of its
  FIRST quantifier, so `<style[^>]*>.*?</style>` over an unordered `string_agg` ate most of every page
  (how-we-assess 3,732 → 8,266 chars; `index` was ONE char). Now per-component, lazy, ordered.
- `claims_verification/SEED_claims_auditor.sql` — the agent's first seed, regenerated from the live row.
- **599_HOLD** (writer no-register arm + `## Operating history: NONE RECORDED` block + "say what we DO"
  qualified) — HELD with `brochure_component_library/sql/page_content_writer_prompt_v5_2026-08-24.txt`
  generated from the LIVE template (the committed v4 text was 1,718 chars behind it).

**Proof, on this site, without touching it** (corr `bcf23316`, after 601): the cold audit's first two
findings are §1's own sentences — *"We garden ourselves, and we test what we can get our hands on."* (about,
high) and *"Where we can, we buy the tool at the same price a reader would pay…"* (how-we-assess, high) —
with the owner's framing in the suggestion (*"reframe as aspiration… 'we aim to test'"*). Work item
`claims_llm_garden-tools.uk` (needs_human_review) — the FIRST that step has ever filed. Control at
leopardessconsulting.co.uk (corr `be39ddba`): roster arm; one real drift finding ("22 sites" vs a verified
floor of 25).

**Go (built, tested against HEAD, INERT until an image rolls):** `datahelpers/claims_practice.go` — the
practice-claims family (§5.2), five physical-verb patterns, `Check="practice_claim"`, exempted by an
`operating_history` attestation in `evidence_base`, NOT unioned into the refusing set (mutation test pins
it), recorded at `practice_claims_severity` (default `warning`) in `validate_page_content`; `claimscan`
prints `PRACTICE` lines. Full-corpus dry run 2026-08-24: 12 findings / 1,867 components — 7 on this site
(all invented), 3 true-practice on operating sites (the attestation case), 1 clear false positive
(`idea.uk` "how we test your idea"). Also: `ParseEvidenceBase` keeps an attestation-only base non-nil and
the numeric scan's arming moved to `HasScannableRegister()` at its three call sites (closes CGV-033's
latent hazard; zero live instances).

**Council:** config slice `Council-Submitted: e684fc8d` (verdict pending at write time); Go slice
submitted separately. **Status: FIXED IN CONFIG AND LIVE for (b) planner and (c) auditor; (a) is answered
by design (no register minted — absence is cold); the writer half (599) is HELD for the owner; the Go
family is inert until the next roll.** Stays in `bugs_open/` until 599 applies and the Go half rolls.

**Named follow-ons (not in this fix):** wiring `evidence-researcher` into the greenfield chain ("source
more" — unattended research breaks agritec RUNBOOK §9's mandatory review; owner decision); the LLM item's
missing `spec.page_id` (parks under `spec_no_page_id`); discovery-check wiring of the practice family
(slice 2b); RFC_003 Q1 (escalation to refusal); `bugs_open/033` (the queue the findings feed).

**2026-08-24 evening — 599 APPLIED** after the owner read and approved the v5 plaintext ("approved"). The
writer's no-register / no-operating-history arm is live (verified by needle on the live row). All three
mechanisms are now closed at the source in config; the Go practice family (commit c9cd817d9, council
APPROVED 1d87615f) remains inert until the next chassis image rolls. **Bar for `bugs_closed/`: that roll,
plus one greenfield build showing the writer arm in `llm_call_log.prompt_rendered`.**

## CLOSED 2026-08-25 — fixed AND live (the CLAUDE.md bar), verified at the artefact

Chassis `v1.0.1337` (pods 09:27Z, provenance `4c996e1b5`, `c9cd817d9` is an ancestor; binary probe 1/1/0
with controls). Config slices 597/598/599/600/601 live since 2026-08-24; overnight the rotation audited
15 sites with 15 receipts (10 findings, 4 clean, 1 hand-run); the writer arm rendered on 150/150
`generate_content` calls. Both council slices APPROVED round 1 (e684fc8d, 1d87615f). Verify-later, not a
gap: the first rebuild of a register-less page carrying a practice sentence should show a `practice_claim`
warning in its validation result (the three post-roll builds had none to flag).
Residuals belong to other files: `bugs_open/386` (stale renders of a refreshed fact — amplified by the
rotation), `bugs_open/033` (the queue the findings feed), the skip-as-success census (handed to the 354
lane), the greedy-regex mechanism on four other agents (LANDMINES), slice 2b, the shared visible-text
function, `evidence-researcher` on the greenfield chain (owner decision). Lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_380_claims_fail_open/` (HANDOFF_2026-08-25 is the cold start).
