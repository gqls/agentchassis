# Compliance read — page-content-writer prompt v4 (seed 330)

**Date:** 2026-08-10. **Performed by:** the fact-assignment front session, at the
owner's explicit direction of today ("Please read the prompt for compliance
yourself (acting as a lawyer)"). **Satisfies:** the compliance seat's round-1 ask,
restated in the Slice B REVISE round (corr `a06ff850-aff6-4ed0-8e0a-93d57b0cbc45`
— see the pinned verdict): *the human read of the v4 plaintext must happen before seed 330
applies, and should explicitly check the three-way branch (scoped / factless /
unscoped) for overclaimed-reliability phrasing, especially what the writer is
told to say when a section has "no verified facts".*

**Provenance, stated plainly:** the seat asked for a *human* read; this read was
performed by the session, under the owner's instruction quoted above. The
operative text is quoted in full in §2 below, so the owner's review of this
document is itself a human read of the load-bearing lines. **Owner countersign
line at the end.** Seed 330's own apply precondition ("a human has read the v4
prompt plaintext") should be treated as satisfied only once that line is
initialled.

## 1. The text under review, and what seed 330 actually changes

- File: `sql/page_content_writer_prompt_v4_2026-08-06.txt` (this directory),
  sha256 `9da490ca8acbf29e5f01f1cbc9adc46bb29f08947b3b0db3f5bb1b3cf29e58cb`.
- Per seed 330's own header, v4 was built **from the live template dumped
  2026-08-06** (not from the stale v3 file) with **one edit**: the Verified
  Facts block becomes a three-way branch. Everything else in the file —
  Edit Mode (bugs_open/178), the `{{.voice_style}}` placeholder (241), the
  house-voice block, and the "NEVER PROMISE ACCURACY" strict rule — is the
  text already live today. `[RELIED ON seed 330's header derivation
  (byte-exact from the live dump, one edit); not independently re-dumped in
  this session.]`
- Consequence for scope: **the three-way branch is the entirety of the change
  being certified.** Findings on the rest of the file are findings about the
  live estate, recorded here because the seat asked for a read of the whole
  plaintext; none of them is introduced or worsened by 330.

## 2. The seat's specific question — the three branches, quoted and read

**Branch A — section has assigned facts** (`facts_scoped` + `assigned_writer_block`):

> ## Verified Facts (assigned to THIS section - the ONLY numbers and named
> entities you may assert here)
> …
> If a fact you want is not in this list: write the capability WITHOUT the
> number, or frame it plainly as something we could do. Never approximate,
> extrapolate, or round a listed number into a different claim. The site's
> OTHER verified facts are stated by other sections - do NOT restate a fact
> that is not in this list. For this site, this list overrides the general
> no-statistics rule: numbers listed here may be stated with their listed
> meaning; any other number about the business is forbidden.

Read: this **narrows** the writer's licence relative to today's site-wide block.
"With their listed meaning" is a genuine anti-distortion control (a true number
recontextualised into a different claim is the classic misleading-statistics
move, and this clause bans it in terms). "Frame it plainly as something we
could do" instructs honest-aspiration framing — capability without track-record
implication — which is the framing advertising regulation wants. **No
instruction here tells the writer to publish any claim about verification or
reliability.** PASS.

**Branch B — scoped, but no facts for this section** (the seat's named concern):

> ## Verified Facts
> No verified facts are assigned to this section. State NO business numbers,
> counts, statistics, or named-entity relationship claims in this section -
> other sections on this site carry them. Write this section's purpose without
> site-specific figures.

Read: what the writer is "told to say when a section has no verified facts" is
— nothing about verification at all. The instruction is a prohibition (no
figures, no counts, no named-entity relationship claims) plus a positive
direction to write the section's *purpose*. The ban's breadth is good: "named-
entity relationship claims" catches the non-numeric fabrications ("regulated by
X", "partnered with Y") that a numbers-only ban misses. The residual risk — a
writer compensating with superlatives — is bounded by strict rule 5 (no
invented achievements), rule 18 (honest-and-general beats specific-and-
fabricated), the house voice (no hype adjectives in either direction), and the
accuracy rule read in §3.1. **PASS on the seat's question.** One factual
contingency on the middle sentence, §3.3.

**Branch C — site not scoped** (fallback): byte-identical to today's site-wide
Verified Facts block, same containment language as Branch A minus the
per-section scoping. PASS (it is the status quo).

**Implicit Branch D — no scoping and no `evidence_base` at all:** no Verified
Facts section renders. The writer is then governed by strict rules 5, 13, 14,
15 (no invented statistics, people, case studies, achievements; rule 14
explicitly closes the "required stat field" loophole: an empty stat renders as
nothing, an invented one publishes a false claim). Adequate; no gap opened by
the branch structure.

## 3. Findings

### 3.1 The prompt's strongest compliance control is already live, and it bounds the "Verified Facts" label — no action
The strict rule "NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE" bans, in terms,
every published claim of the form "our figures are verified/sourced/
authoritative/reliable", requires method statements to describe what we DO
rather than what can be relied on, requires interactive tools to say they can
be wrong, and requires named cases to be presented as illustration. This is the
rule that stops the internal label "Verified Facts" ever leaking into copy as a
published reliability claim. It is live text (not part of 330's delta), and it
is the reason the label is safe *as an instruction* — see 3.2 for what the
label vouches internally.

### 3.2 The label "Verified Facts" vouches for the evidence pipeline — condition, not defect
The prompt tells the writer the listed items are verified and licenses their
publication. The prompt is therefore exactly as honest as `evidence_base`
provenance. **This read certifies the prompt text, not the evidence pipeline**;
the claims-gating on `evidence_base` is a separately maintained control.
Disconfirming condition, stated so this read is falsifiable: if any entry in a
site's `evidence_base` is found to be self-reported or estimated rather than
verified, the label overclaims and must change (e.g. "Approved Facts") — that
would be a defect in this prompt's wording, not only in the data.

### 3.3 Branch B's "other sections on this site carry them" — true only after the §3.5 fix; failure direction is omission, so not a blocker
Under the currently-shipped carry, a section whose `facts` key was dropped or
malformed is indistinguishable from one legitimately assigned none (the Slice B
REVISE's upheld §3.5 hole). In that state this sentence can be false: the facts
are carried by no section. **The direction of the error is omission** — a fact
missing from the whole site — never fabrication; a false premise here licenses
no claim in the section being written. Legally that is under-claiming, a
quality defect rather than a compliance exposure. The `FACT_ASSIGNMENT_ABSENT`
fix (already work-queue item 2) is what makes the sentence honest; it should
land in the same round, and this read records that as a **condition of
quality, not of certification**.

### 3.4 Pre-existing gap, LOW–MEDIUM: invented commitments and guarantees without figures
Rule 14 catches every invented *figure*; rule 5 catches invented *achievements,
metrics, or outcomes that have not actually happened*. A forward-looking
promise is none of those: "money-back guarantee", "free no-obligation
consultation", "we'll always pick up the phone" contain no figure and are not
past outcomes. Nothing squarely bans inventing them. Branch A/B do not cure it
(a guarantee is not a "fact you want" from the list). **Remedy, one clause,
follow-up seed:** extend rule 5 — "…and do NOT invent commitments, guarantees,
warranties, or service promises (response times, refunds, availability) not
stated in this prompt." Not a 330 blocker: the gap is in live text and 330
neither introduces nor widens it.

### 3.5 Pre-existing, MEDIUM but mitigated: Edit Mode preserves legacy non-compliant claims
Edit Mode instructs: preserve existing prose and information, change only what
the guidance asks. A fabricated statistic from an earlier, laxer prompt
generation therefore survives edits — preserving is not inventing, so rules
5/13/14/15 do not bite. Mitigation: the platform's `banned_claims` sweep is the
designed detection path for exactly this class (it, not the prompt, is what
arms detection — bugfix 161's lesson). A prompt-side remedy needs careful
drafting because it cuts against "do NOT discard unrelated material"; suggested
shape for a future revision: "If existing content states a figure or named-
entity claim the rules above forbid and this prompt does not verify, you may
generalise or remove that claim even though the guidance does not name it."
Record only; not 330's delta.

### 3.6 Pre-existing, LOW: testimonial placeholders and rendering
Rule 16 has the writer produce 2–3 statements "in the company's own voice" as
placeholder content for testimonial sections, names and titles left empty (rule
13 already bans invented people and attributed quotes — that is the right
rule). The residual exposure is at the **rendering layer**: if a testimonial
component styles these as customer quotes (quotation marks, review styling), an
unattributed company self-statement could read as a review, which is the shape
UK fake-review rules (DMCCA 2025; CAP Code) care about. Writer-side text is
defensible; the check belongs on the component templates: placeholder
statements must not render with review/quote trade dress, or the section should
render empty. Recorded for the component library's queue.

### 3.7 Drafting note, LOW: Branch A vs rule 14 on research figures about the business
Branch A: "any other number about the business is forbidden." Rule 14 permits
figures "given in THIS prompt (Verified Facts, Research Findings, …)". A
research finding stating a figure *about this business* sits in the overlap.
The specific-over-general reading (Branch A wins for business figures) is the
correct construction and almost certainly the intent; a parenthetical in a
future revision would remove the ambiguity. No action now.

## 4. Verdict

**The three-way branch — the whole of seed 330's change — is certified: it
contains no overclaimed-reliability phrasing, publishes no verification claims,
and strictly narrows the writer's licence relative to the live prompt.** The
factless branch, the seat's named concern, tells the writer to claim nothing
and write purpose without figures — the correct instruction. Findings 3.4–3.7
are pre-existing, recorded with remedies as follow-up seeds/checks; finding 3.3
is a quality condition satisfied by the §3.5 fix already in this round's work
queue. Nothing found blocks 330, which remains gated by its own header's
technical preconditions (Go half live on the pod; own council round concluded).

**Owner countersign** (adopting this read as the human read seed 330 requires):

- [x] Read and adopted — owner, 2026-08-10, in session: "the prompt is fine go
  ahead". Recorded by the session at the owner's word; seed 330's apply
  precondition is satisfied.
