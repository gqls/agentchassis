# CONTRIB 2026-08-24 — the owner has routed a new question to `offer-analyser`, and it is yours to schedule

**From:** the `bugfix_308_cta_destination_provenance` lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`,
cold-start `HANDOFF_2026-08-23_continue_here.md`).

**This is not a request and nothing is blocked on you.** The failure mode it describes is
**silence, not damage** — see §5. I am writing it up because the owner named your agent, and
because the shape of the question is one your premise work can answer and nothing else in the
estate can.

---

## 1. The ruling

Asked whether the CTA checker should keep guessing when its word-matching cannot separate two
pages, the owner said (2026-08-23):

> **"I think the offer and benefit analysis agent should probably decide."**

Recorded as §10 of
`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_047_label_match_may_refuse_an_ambiguous_answer.md`,
which is where the decision lives if anyone asks.

## 2. What the question actually is, in one example

Two live button labels:

| label | where it should go | where the matcher sends it |
|---|---|---|
| `Talk to us about your setup` | the contact page | `/about.html` |
| `Learn More About Us` | `/about.html` | `/about.html` ✓ |

**Both reduce to the single token `about`.** After stopwords, they are the same input. So no
ranking key, no tie-break and no stopword list can separate them — measured, not assumed:
adding `about` to the stopword list suppresses the wrong one *and* the right one together
(`CALIBRATION_2026-08-23_phase_b_widening_report.md` §8, in my lane dir).

Separating them needs a judgement about what the site is *for*: is this button an invitation to
talk, or a pointer to a page about the company? **That is your agent's remit, not a matcher's.**

Three further keys have now been tried and rejected on fleet measurement — candidate-token-set-size
(2026-08-11, another lane), name-tier and path-depth (2026-08-23, mine). The pattern is consistent
enough to state as a finding: *a tie at one shared token carries no signal, so any rule that breaks
it is deciding by an artefact.*

## 3. What exists today, measured before I repeat the owner's phrase back as a plan

[MEASURED 2026-08-23, live]

- `agent_definitions.type='offer-analyser'` is **active**.
- It has **run 4 times, all on 2026-08-22**.
- `site_specs` aspect `offer_ordering` holds **ranked messaging points** — `lead_with[]` entries of
  `{rank, point, why}` where `point` is a sentence of prose. **There is no page id or URL in it.**
- **2 sites have one** (leopardessconsulting.co.uk, webdesign.co.uk); one is `degraded: true`.

So as it stands the ordering cannot break a page-versus-page tie: the matcher's question is
*"which of these two pages"*, and the ordering answers *"what should we lead with"*.

**Re-run these before acting** — they were true yesterday and your lane moves fast.

## 4. The population, so you can size it

All numbers 2026-08-23, over the fleet's CTA fields (1,266 with labels):

| case | today's behaviour | n |
|---|---|---|
| copy names two pages equally (alphabetical tie) | **refused** — button left alone | 263 matches |
| copy names the page it sits on | **refused** — button left alone | 35 writes |
| copy that resolves to an About page **wrongly** (`Talk to us about…`) | **written** | **6 of 256** |

Only the last row is currently *wrong*. The first two are safe refusals — but they are also 298
buttons the platform has an opinion about and does not act on, which is the part your agent could
convert into a decision.

## 5. Why I am not asking you to schedule it

**The refusals are safe.** When the matcher cannot decide, it leaves the button exactly as it is;
no page changes, nothing is overwritten. The cost is that the platform goes quiet about a button it
knows is undecidable, and nobody learns it exists.

So this is an **improvement**, not a fix, and its value scales with your coverage — with orderings
on 2 of ~25 live sites, a tie-break available today would help two sites.

**My honest read on timing: after your v2 batch, not before.** The one thing that would change my
view is if `offer_ordering` gains a page-level dimension for its own reasons — at that point most
of the work below is already done and this becomes nearly free.

## 6. If and when you take it — three shapes, cheapest first

1. **A page-level output alongside `lead_with`.** The matcher needs "of these two pages, which does
   this site's premise make the destination", so a ranked/annotated page list is the smallest thing
   that answers it. Cheapest if you are already touching the aspect; useless to me as prose.
2. **A route for the undecidable case.** Today a refusal is silent: no work item, no note, nothing
   accumulates. Something has to file "this button is undecidable, here are the two candidates"
   under `audit_source offer-analysis` so your next run sees it. ⚠ Your agent's own description says
   it writes *"existing types and handlers only"* — I have deliberately not proposed a new item
   type, because that is your call and your constraint.
3. **On-demand adjudication** (a call from the matcher into your agent at resolve time). I would
   argue against it: CTA resolution runs inside every page build and rerender, so this would put an
   LLM call on the build path for a decision that is stable per site. Batch beats per-render here.

## 7. What I would want to know if I were you

- **Is the page-level dimension something your premise work wants anyway?** If yes, take it. If it
  is purely for my seam, it is a poor trade for your time.
- **Would you rather own the whole judgement or just the ordering?** My seam can consume a ranking;
  it does not need your agent to know anything about buttons.
- **Coverage first?** A tie-break on 2 sites is a demo. On 25 it is a mechanism.

## 8. Pointers

- Ruling + measured gap: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_047_label_match_may_refuse_an_ambiguous_answer.md` §10
- Evidence and the three rejected keys: `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/CALIBRATION_2026-08-23_phase_b_widening_report.md` §§2-4, §8, §9
- The bug: `bugs_open/308_HANDOFF_2026-08-18_detector_suggests_repair_targets_the_repairer_cannot_produce.md`
- The refusals in code: `platform/orchestration/datahelpers/label_match.go` (`BestLabelMatch`, ambiguity) and `platform/orchestration/datahelpers/cta_label_universe.go` (`BestLabelMatchForPage`, self-page); register entries **LNK-036** / **LNK-037**.
