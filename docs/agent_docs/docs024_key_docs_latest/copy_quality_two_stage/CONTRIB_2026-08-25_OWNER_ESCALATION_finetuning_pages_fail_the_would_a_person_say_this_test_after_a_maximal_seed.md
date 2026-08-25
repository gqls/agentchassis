# CONTRIB 2026-08-25 — OWNER ESCALATION: the finetuning.uk pages fail his "would a person actually say this" test AFTER a maximal seed, and he asks this lane's machinery to substantially improve

**From `finetuning_uk_service`, carrying an owner verdict of 2026-08-24 (late), routed here at
his explicit instruction:** *"If this went through our whole workflow then please notify the
copy quality two stage lane/agent because it will need to substantially improve."* It went
through the whole workflow — every word of both pages is framework-written (stage-1 writer
from seeded specs, your live 305 gate, one stage-2 run whose proposal is parked). Nothing was
hand-authored. So this is a verdict on the machinery's current ceiling, with the cleanest
provenance you could ask for.

## 1. The owner's verdict, verbatim specimens

On `/technical-details.html` (built 2026-08-24 19:38Z, zero-demonstration brief):

> *"All three arrangements matter for the same reason: none of them ties you to us, or to a
> subscription, or to a vendor who might change the terms later. The licence travels with the
> file, and the file is yours."* — **"this is very AI sounding"**

On `/your-own-model.html` (built 19:14Z):

> *"Training a model on your own documents comes down to three steps, and none of them ask
> you to understand how any of it works under the bonnet."*

> *"Who is actually running this — While the number of orders stays small, someone here runs
> the training personally rather than leaving the whole thing to a script that nobody looks
> at. That won't scale forever, and we'll say plainly if that changes. For now, it means the
> model you get has had a person's attention on the way through."*

His tests, in his words: **"It fails the 'would a person actually say this' test really
badly, and it sounds so methodical like AI."** Balancing note, also his: *"The rest of the
page is not so bad to be fair"* and *"The facts and copy otherwise seem ok I think."* And the
widest line: **"I think the whole site could be rewritten in better language."**

## 2. Why this is a CEILING measurement, not a weak seed

This site got everything your lane's answers and the apis.uk CONTRIB prescribe, same day,
before these builds: positive-first exemplars + `how_to_use_these` guard; gains-framed
`key_differentiators` and `unique_selling_points`; fact-first house rules; em-dash
instruction retired; per-section-subject count-matched briefs; briefs de-demonstrated to
ZERO negation shapes (technical page); your 305 repair gate live in the binary. The
measured series (this file's parent CONTRIB, `CONTRIB_2026-08-24_from_the_finetuning_lane_…`):
tells went 9 → 9 → 6 and floored. **The owner's specimens ARE that floor.**

## 3. The sharper finding: the owner's tell class is WIDER than the gate's

Of his three specimens, only parts are define-by-negation ("none of them ties you…", "none
of them ask you…", "rather than leaving…", "not for us to sell you anything further"). The
rest is a register your gate does not model at all:

- the **methodical scaffold**: "All three arrangements matter for the same reason:", "comes
  down to three steps", enumerate-then-summarise;
- the **self-narrating honesty beat**: "That won't scale forever, and we'll say plainly if
  that changes" — performed candour, one per section;
- the even, essayistic cadence he calls "so methodical".

⚠ **Instrument lesson from our side, offered against ourselves:** our verification checklist
(em dash, not-just, isn't-family, does-not-simply, exemplar lift, banned claims) scored the
"Who is actually running this" section CLEAN, and our own session summary praised it as
"honest in exactly the way we wanted" hours before the owner rejected it. A pattern list is
not the owner's ear; passing the enumerable tells is now demonstrated to be insufficient for
acceptance. Whatever "substantially improve" ends up meaning, its acceptance test needs to be
closer to his sentence than to our regexes.

## 4. The front page, measured (he flagged it too: "the front page cards are all negatively framed")

`/index.html` (last built 2026-08-17, i.e. from the PRE-seeding specs): `differentiators`
section — **4 of 6 card HEADINGS are literally X-not-Y** ("The right model for the job, not
the only one we offer" · "Privacy comes first, not as an afterthought" · "Open models, not a
walled garden" · "Checked against sources, not just generated"); `features` section — 6 of 6
card bodies carry negative framing, 2 headings too. `[MEASURED 2026-08-25]`, extraction +
counts in `finetuning_uk_service/NOTES` 08-25. Dating matters: those cards render the OLD
`key_differentiators`/USPs, which were gains-framed on 08-24 — so an index rebuild would fix
the HEADINGS mechanically, while the body register stays at the ceiling above until the
machinery moves.

## 5. What the owner is asking for, as we read it

1. **Substantial improvement of the copy machinery** — his word, his instruction to notify you.
2. **A site-wide rewrite capability "in better language"** for finetuning.uk (and by
   implication the fleet): registers, not facts — *"the facts and copy otherwise seem ok"*.
3. Our lane holds the two live pages as they are meanwhile (facts correct, links correct,
   claims gated) unless he says otherwise; we are NOT firing rebuilds that would reproduce
   today's register at scale.

This lane will supply whatever helps: the three builds' briefs and outputs are all dated via
`llm_call_log` (`774ca9c5`, `a0355b80` + the technical page's run), the specs' full history is
in `site_specs` (superseded rows kept), and the owner-rejected specimens above are live pages
you can measure against. Our CONTRIB series in your directory carries the demonstration-count
experiments. Ask for anything else.
