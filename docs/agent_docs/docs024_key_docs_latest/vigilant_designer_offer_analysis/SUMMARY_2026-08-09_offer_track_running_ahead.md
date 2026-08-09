# SUMMARY 2026-08-09 — the offer track runs ahead: B1+B2 witnessed, B3 built

*Read-out for the owner. Current state only; chronology in NOTES and
README_where_we_are. Previous: `SUMMARY_2026-08-08_the_offer_track_jumps_the_queue.md`.*

---

## What we're trying to do

Unchanged: a vigilant designer and an offer-and-benefit analyser, each a standing faculty,
each built to the rule that every detector ships with the thing that acts on it. The offer
side keeps asking one question per site: does this site answer its target market's need, in
a way that pays us — judged against the site's own recorded premise, because we have no
visitor data and an analyser that pretended otherwise would be confidently wrong.

## Where we've come from

Yesterday's summary ended with your decision that the two cheapest offer pieces jump the
queue. Since then both were built as configuration, applied, and — the part that matters —
**witnessed doing their job on real sites**, and the third phase is now written as code.

## What we've done

**B1 — the strategic review reads the premise, proven by a planted marker.** A hidden
marker in webdesign.co.uk's strategy record travelled into the exact prompt the model
received (29,110 characters, no truncation), and the output transformed: the review now
opens from what the site is FOR and every finding is a premise-vs-artefact mismatch — it
caught a "Read the guides" button routing to a tool page, an About page rendering wrong
content, and a hero that buries the site's recorded moat. The blind version saw none of
this.

**B2 — refreshing a premise is safe, proven on a live site.** The strategy agent ran
against loancalculator.co.uk (live, twenty-seven pages, no strategy on record): it wrote
the site's first premise — affiliate model, with all four new questions answered concretely,
down to "commission per completed application, monthly in arrears" — and filed **zero**
rebuild jobs. Before this, that run would have re-planned a live site; that is why nobody
had ever dared run one.

**B3 — the two offer checks now exist as tested code.** `premise_incomplete` finds a
shipped site with no premise (or the old shape with no revenue model) and files the same
strategy request the build pipeline uses — safe precisely because B2's gate shipped first.
`revenue_shape` makes doc 028's rule mechanical against each site's own recorded model: a
tools or advertising site carrying "Start a Project"-class button text gets a per-page fix
item; a lead-generation site with no reachable enquiry form gets a conversion-path item;
an affiliate site files the honest "no machinery exists" roadmap row; and hits in the
shared site chrome file a roadmap row rather than dispatching, because editing shared
chrome edits every site at once. Both checks retract their own findings when they
positively observe them fixed; both new item types carry verifiers that re-run the
detector's own predicate, and the timeout sweep that could have bypassed those verifiers
was excluded — in the declared list and in the live column — before the first item can
ever exist.

**What the machinery caught on the way through, which is the part worth reading twice.**
The commit-time pattern check flagged that both new checks (and, it turned out, B2's live
gate from yesterday) asked "is this page live" as `build_status = 'deployed'` — which
misses every page flagged for rebuild that is still serving. In the gate's case the miss
pointed the dangerous way: a site mid-rebuild would have read as "not deployed" and been
re-planned. All three sites now use the shared shipped-page predicate, the gate fix is
applied live, and the witnessed behaviour was re-verified unchanged first.

**Corrections carried forward honestly:** yesterday's claim that a plan-critical field
never existed was wrong (it lives one level down; the plan was right, and the recovered
reading gave the revenue-shape check its day-one population: ten of seventeen sites record
the consultancy shape). The council refuses config-only changes by scope — so B1/B2
shipped on the lane's config precedent, while B3, being Go, went through the gate
properly (verdict pending at write time).

## Where we are now

- **B1, B2 live and witnessed.** Ledger rows recorded by your hand (340/341 — note both
  numbers are now ambiguous: another lane took the same pair the same evening; resolve by
  slug).
- **B3 code committed and council-submitted; INERT by design** until an image rolls and
  the two check names are added to the quality-discovery agent's list — a name added
  before the roll is fatal, so the order is fixed and written down.
- **Two watch items stand:** the next real greenfield build is the natural proof that B2's
  else-arm still chains (nobody has run one since the gate); and one pre-existing test
  failure in the checks package belongs to another lane's in-flight work, verified
  pre-existing at clean HEAD.

## Where we're going

Next session, in order: read B3's council verdict and act on it; ride or cut an image roll
and pod-verify the two checks; add the names to quality-discovery-agent's array
(observe-only); hand-fire one sweep and read what the checks file against the
10-of-17 `direct_business` population — expecting the first honest offer findings to be
arguments, not certainties. Then B4, the analyser itself, which reads everything B1–B3
built: the premise in the review's context, the four premise questions in every strategy,
and a clean mechanical floor underneath so the LLM analyser only ever judges what a check
cannot.

The open questions from `features_open/030` remain the owner's: which council (if any) the
offer judgement belongs to, which of the two missing correspondence routes matters first
(tool design or the experience loops), and whether premise *quality* ever comes into scope.
