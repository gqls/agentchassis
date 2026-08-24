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
**Open, and worth a sweep:** how many live sites have **no** `evidence_base`? Every one of them has
had its claims audit silently skipped for its whole life, and the audit's own records cannot tell you
— a skipped run and a clean run are both `complete`. The population is countable in one query and I
have not run it; do that before sizing the fix.

⚠ **A COUNT OF THINGS CARRIES ITS DATE (owner ruling 2026-08-22)** — the sites-with-an-evidence-base
list above is **as of 2026-08-24** and this class grows by ADDITION with every greenfield build.

## 7. Provenance

Live run `garden-tools.uk`, submitted 2026-08-23 17:17Z with no prompt, no mission and no seed.
Owner review of the served pages 2026-08-24 named the hallucinated claims and the FAQ overclaiming.
Lane record: `docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/NOTES_loanzy_uk_example_site.md`.
Related but distinct: `bugs_open/288` (the evidence register guards COPY, not CODE — a tool encoding a
legislated figure is checked by nothing) is the same register failing on a different surface; this is
the register being **absent entirely** and every consumer failing open.
