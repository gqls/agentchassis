# Prompt audit — phase 2 verdicts (1 of N): the writer template and the house voice, 2026-08-25

Read in full, against the six questions in `PLAN_2026-08-25_prompt_audit.md` §2. Verdict scale:
**teaches-good / neutral / teaches-AI**, on the owner's question — *"is it contributing to good
readable copy or encouraging AI styles of writing?"* Fix SHAPES are recorded, **not applied**:
prompt text changes ship as migrations through the council gate, and one of them touches text the
owner approved himself.

---

## 1. `page-content-writer` · `generate_content` (sub_workflow) — 14,918 chars, 6,452 calls/30d

**Verdict: TEACHES-AI — not in its bans, which are right, but in what it prescribes INSTEAD.**

Three instructed moves, each quoted from the live template `[MEASURED 2026-08-25]`:

**(a) A "method" section must narrate sourcing.** Twice, in two blocks:
> *"Where the brief asks for method, say what the site does WITH SOURCES: name the manufacturer
> specification, published standard or retailer listing a figure comes from, and date it."*
> (Operating history: NONE RECORDED block — emitted in BOTH branches of its `if`, i.e. always when
> no history is recorded, which is every greenfield site)

> *"If the section calls for a statement about method, say what we DO — and with no recorded
> operating history that means ONLY how the content is sourced: we name our sources and their dates
> so a reader can check them — and, where it fits, say plainly that we can still be wrong."*
> (STRICT RULE — NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE, migration 223 → 599)

→ homegarden.uk about.html: *Sourced and dated · Sources named · Where the detail comes from ·
Timing stated plainly · Why the plain answer matters more than a confident one*. The owner:
*"Who would want to know that, really? (no one reading this site)"*.

**(b) "Honest" was banned as a WORD and re-prescribed as a MOVE.** Rule 19:
> *"Never write the words 'honest', 'honestly' or 'honesty' in page copy … Show it instead, by
> naming the limit, the failure mode, or what the thing cannot do. Say 'we cannot tell you X' rather
> than 'an honest assessment'."*

→ *What this site will not do* (×2 on one page), *What a comparison cannot settle for you*, *"this
site will not tell you…"*; finetuning's *"That won't scale forever, and we'll say plainly if that
changes"*. The owner: *"no one cares"*. This is the displacement the REFRESH documents (§2), in
the fleet's most-read prompt: the 07-26 word ban removed "honest" and installed the
self-limiting register that replaced it. Rule 18 (*"It is ALWAYS better to be honest and general"*)
still names the posture the copy then performs.

**(c) Empty social-proof slots are filled with self-description by instruction.** Rules 16–17:
> *"For testimonial sections: write 2-3 statements in the company's own voice about their values,
> approach, or commitment."* · *"For case study sections: describe the types of problems the company
> solves and the approach they take."*

→ "talking about ourselves", the owner's first objection (08-12: *"we don't want to talk about
ourselves unless it's to their benefit"*). This is partly a PLANNER question (why is a
testimonial slot planned on a site with no testimonials?) — noted for the planner verdicts.

Smaller demonstrations in its own prose: *"frame it plainly as something we could do"* (×2 — the
word the owner said *"people just don't say"*), *"write real structure, not a run of paragraphs"*,
*"EDITING this content, not by replacing it"*. Reachability: maximal. Detector coupling: rule 19
pairs with the voice-tells scanner's "honest" check; the STRICT RULE pairs with the compliance
council seat (223 added both) — the seat's contract (flag overclaimed reliability) is unaffected by
removing the remedy clause.

**What is RIGHT and must stay:** every fabrication ban (rules 5, 8, 13, 14, 15), the
no-overclaiming ban (first half of the STRICT RULE), the "honest" word ban's motive, the
no-markdown/structure rules (9, 10 — 381's fix), the JSON discipline.

**Fix SHAPE (not applied):** keep every ban; delete the three "say this instead" prescriptions and
replace with one positive instruction — *a method or about section is about the READER's subject;
if the brief gives nothing reader-facing to say, say what the reader can do with the page in one or
two sentences and stop* — with ONE on-register exemplar in the owner's homegarden sentence shape
(*"We're hoping you can get a lot of useful tips from this site…"*), guarded as a style sample
(`how_to_use_these` lesson). Rule 19 keeps the ban and drops the substitute phrase. Rules 16–17
prefer the empty string to philosophy statements, and the slot question goes to the planner audit.
**Causal test before any of this is called a fix** (Finding 2's `[INFERRED]`): build one
method/about section with the current template and one with the remedy clauses removed, same
brief, compare headings and self-description count at the served page.

---

## 2. House voice — `voice_style_block` — 6,033 chars, in every `{{.voice_style}}` template

**Verdict: NEUTRAL in what it says, TEACHES-AI in how it says it.** Owner-approved v2 (08-14:
*"I approve v2, I like that much more"*); its content is essentially the v3 style prompt's rules
(em-dash ban, word-weight both directions, contractions, cadence, read-it-aloud) — the right
lineage. Zero "plainly", zero "honest", no methodology instruction, positive exemplars present
(D1 satisfied). But it is the DENSEST single prompt in the fleet for the construction it bans:
**17 demonstrations in 6,033 chars (2.8/1k)** `[MEASURED 2026-08-25]`, including the rule against
the tell written as the tell:

> *"Say what a thing IS rather than what it is not."* · *"apply the reason, not just the letter"* ·
> *"persuaded rather than sold to"* · *"a machine gun rather than a person talking"* · *"Say why it
> matters, not just what is true"* · *"like someone with a point of view …, not like a
> specification being read out"* · *"name the action rather than gesturing at it"* · *"Leave one
> slightly blunt or plain phrase standing rather than smoothing every sentence"*

Two prescriptions with tic potential, both plausible generators of the "methodical" cadence the
owner hears: *"At least one sentence should give the reader a reason to care"* (per section — a
why-it-matters beat every section) and the word-list ban itself (13 word-weight proxy hits — it is
a LIST of the banned words; a proximity detector convicts the denial, 016b :12508; low risk, non-zero).

**Fix SHAPE (not applied — OWNER DECISION, it is his approved text):** re-write the block's own
sentences in the shape it prescribes — state the positive, fold contrast after — without changing
a single rule's MEANING. *"Say what a thing IS. 'The parts that are more judgement than
arithmetic' tells the reader something; a definition by subtraction leaves them to work out the
remainder."* Testable: re-run `audit_prompt_demonstrations.py` (17 → ~0 on population F), then
measure `X, not Y` per rendered call and per shipped section over the following week, refutation
condition written first (the lane has withdrawn one before/after of this shape already —
REFRESH §8). Sequence it AFTER the writer-template change so the two effects are separable.

---

## Reading order for the next verdict file

3. `content_direction.formatted` briefs (top five + homegarden's) — the densest layer per call.
4. `content-gap-planner` `plan_gaps` and `build-site-planner` `plan_site` — the premise layer:
   what an about page is FOR, and why testimonial/method slots get planned with nothing to fill
   them.
5. `llm_guidance` top 30 (the loanandmortgagecalculator `not just` family first).
6. `copy-editor`'s own prompt (this lane's — 6 demonstrations incl. *"THE READER, NOT THE SITE."*).
