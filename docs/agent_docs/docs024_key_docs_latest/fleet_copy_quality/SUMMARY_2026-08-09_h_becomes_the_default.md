# SUMMARY 2026-08-09 — H becomes the default voice

Third summary in this workstream. The first (06th) reported that we could not find the
fault by measuring. The second (08th) reported what happened when we tried to fix it by
tuning rules, and was mostly a record of things that did not work. This one records a
decision, and it is the first summary here that closes a question rather than opening one.
Written to be read aloud.

---

## The decisions, laid bare

**Decided today by the owner:**

1. **Voice H becomes the fleet default.** It stops being the finance pool's house style and
   becomes the house voice that ships in the writer's base prompt — every vertical, every
   future site. This confirms and completes the "wide over contained" choice of 2026-08-05,
   which had been sitting undone for four days.
2. **It is provisional — "for now".** This is explicitly a working default, not a
   settlement. The review trigger is written into "Where we're going" below.

**Entailed by that decision, and therefore also decided:**

3. **The opening rule changes form, from a prescription to a prohibition.** The current
   default says *"Start with the fact."* H says *"never open a section cold with a bare
   assertion… vary how sections open."* H's version wins. This matters more than it looks:
   it is the single point where the two genuinely conflict.
4. **The shared prohibition stays.** Both the old default and H ban opening on a negative
   twist ("X isn't about Y…"). That half is not a change and should not be re-argued.

**Explicitly NOT decided, and still open:**

5. **How it ships** — seven separate edits, or one shared carrier read at prompt-assembly
   time. Both go in the submission; neither gets picked quietly.
6. **What happens to the worked examples.** On this week's evidence this is the part most
   likely to decide whether the change does anything at all.

**Standing decisions this does not disturb:** a site's own voice spec still overrides the
house voice; the expanded copy on loancalculator stays (owner, 08-08).

---

## What we're trying to do

Have the framework write copy that reads as though an intelligent person wrote it, for a
reader with something they are trying to get done. Not salesy. Explaining what deserves
explanation, saying the quiet non-obvious things plainly, and offering what we can
genuinely do without narrating our own limits.

The only mechanism we have for steering that is a prompt: a house style shared across the
fleet, plus a per-site voice specification. Everything below is about which of those two
carries the weight.

## Where we've come from

Two attempts, both rule-based. In July the owner's own writing style was reverse-engineered
into a prompt and refined over three rounds. In August a "gentle explanatory" voice — H —
was developed with him and seeded onto the finance sites. Between the house style and a
site's own spec, a writer could be carrying thirty-odd rules at once.

Then a week of testing the instinct to tune those rules, rather than following it. Three
measurements found nothing across six sites and nearly nine hundred sentences. The owner
read a page and found the fault in about a minute, and has since found three more, none of
them reachable by anything we can measure. A controlled removal of the suspect rules left
the tic untouched and introduced a new fault on a live page.

In parallel, H was rolled out to every one of loancalculator's twenty-six pages, reviewed
by the owner, and corrected twice in response. That site is now the only place H has been
run to completion and judged.

## What we've done

We finished the rollout, took the owner's review, and made the two corrections it asked
for. The second correction is the one that matters here: H's opening rule was originally a
**mandate** — every section had to open a particular way — and it produced an identical tic
across every section it touched. We changed the mandate into a prohibition plus an
instruction to vary, and the tic broke. That is the version being made the default.

We also established what the fleet-wide change actually consists of, which was not obvious.
"The base prompt" is **seven prompts, not one**, and they have already drifted apart — no
two identical. All seven still carry the old opening rule today. One of the seven is
`page-content-writer`, the agent that wrote every page of the site this decision is based
on, so our evidence sits inside the change rather than beside it.

## Where we are now

The decision is made and the blast radius is smaller than it sounds, because the house
voice already defers to a site's own spec. Measured today:

- **21 deployed sites.** 20 carry their own voice spec and are unaffected. **One —
  `cookly.uk` — has no voice spec at all** and is governed entirely by the default.
- Of the 20, **18 say something about openings**; the other two are silent on it and would
  inherit H's opening rule. *(That is a crude proxy — "the spec mentions openings" is not
  the same as "the spec overrides this specific rule" — so treat three-of-twenty-one as an
  upper bound on live disruption, not a finding.)*
- **17 sites sitting in the pool, none with a voice spec.** Every one of them, and every
  site we build from here, gets H.

So this is overwhelmingly a decision about **future** copy, not a rewrite of the estate. It
is cheap to make and expensive to leave half-made, which is the argument for doing it now.

One thing should be said plainly, because it is the strongest evidence pointing the other
way and it would be dishonest to bury it in a footnote. In a controlled two-arm test on the
loancalculator homepage, **the plain default house voice produced the softer, better claim
and H's spec did not** — the over-claim we later had to remove survived H and was killed by
the default. That is one comparison on one page, and it was about claim strength rather
than about openings. It is not a reason to stop; it is the reason the decision is marked
"for now" and the reason the review trigger below is specific.

## Where we're going

Three pieces of work, in order.

**First, decide how it ships.** Seven edits will drift again — they already have, from a
common ancestor, without anyone intending it. The alternative is one shared carrier read at
assembly time, for which we have local precedent. This goes through the council gate with
both options written up, plus a concept-register entry in the same commit, and the other
consumers told rather than merely counted. It is a revision of a live fleet-wide default,
not an addition, and it should be submitted as one.

**Second, change the examples, not just the rule.** This week's clearest finding was that a
writer follows exemplars more reliably than rules — we deleted a rule, left its three
worked examples, and the behaviour continued unchanged. **The example is the instruction;
the rule is commentary.** A base-prompt change that edits rule text and leaves the old
exemplars in place is theatre, and we now have direct evidence of that rather than a
suspicion.

**Third, the problem underneath all of it.** The writer sees one section at a time and never
its siblings, so any instruction is applied by every section that can believe it qualifies.
That is now five instances deep, and on the fifth we followed our own documented remedy —
phrase it conditionally — and it leaked anyway; what held was locking the sections we
weren't targeting. A fleet-wide rule is precisely an instruction applied uniformly by every
section, so this failure mode is not a side issue for this change. It is the change's main
risk.

**When we revisit "for now".** Two triggers, either one sufficient: the owner rejects an
opening on a non-finance site that H produced and the old default would not have; or the
claim-strength result above reproduces on a second page, which would mean H is
systematically weaker than the default on the axis he has judged on twice. Until one of
those fires, H is the default and we build on it.
