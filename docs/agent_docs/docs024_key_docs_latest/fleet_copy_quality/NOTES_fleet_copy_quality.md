# NOTES — fleet copy quality

Running record, append-only, newest at the bottom. Technical log: evidence, commands,
what the system actually said, and every misstep.

Cross-links: the voice work lives in `portfolio_positioning/VOICE_gentle_explanatory_v1.md`
(session `fffe0948`, quiet since 2026-08-05T10:59Z). The owner's own style prompt is
`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` — **read its
round-2 and round-3 corrections before proposing anything**, they are the prior art that
most constrains this workstream.

---

## 2026-08-06 — opened. Three measurements, three refutations, and the fault found by reading

**Owner brief (2026-08-06):** copywriting is an art and subtler than I had been treating
it. Copy should be readable; humans don't repeat one sentence shape. What deserves
explanation gets it; what is quietly non-obvious gets said plainly. Everything phrased
for the reader's benefit. *"I can only give them what I can truly offer but I don't have
to constantly say what I can and can't do — I can offer what I can do strongly when
needed or softly when a hint is more in order."* Generic, all sites.

### Why measure at all, and why measure first

A style complaint is easy to agree with and hard to act on. The first question is
whether the fault is **machine-visible**: if it is, we can detect, gate and prove a fix;
if it is not, no rule and no automated check will ever hold it and we should stop
building one. So: take the most plausible mechanical explanations and try to CONFIRM
each against live served copy. Each probe carries its own comparison, so it can come out
against me — sentence variance against the known human range, within-page repetition
against a between-page baseline, our copy against a different model's copy written
outside the framework.

Script: `prose_stats.py` (this directory). Pulls served HTML, strips
script/style/nav/header/footer and HTML comments, takes `<p>` text over 40 chars.

### Probe 1 — rhythm. REFUTED, and I nearly published the artefact

Hypothesis: "one idea, one sentence" plus "split anything chaining clauses" yields
uniform sentence length.

**First sample (2 pages, 26 sentences), webdesign.co.uk:** mean 9.9 words, sd 2.4,
**CV 0.25**, min 6 max 14, zero sentences ≥25 words. That is a striking monotony
signature and I began building an argument on it.

**Widened to 6 sites / 880 sentences:**

```
finetuning.uk        240 sents  mean 15.9  CV 0.55  min 2 max 47   >=25w 16.7%
gaswholesalers.com   163        mean 15.6  CV 0.46  min 1 max 46   >=25w 12.9%
idea.uk              289        mean 21.8  CV 0.56  min 1 max 67   >=25w 36.3%
vonc.com             137        mean  9.3  CV 0.67  min 2 max 46   >=25w  1.5%
webdesign.co.uk       53        mean 11.0  CV 0.51  min 2 max 25   >=25w  1.9%
```

**CV 0.46–0.67 is normal human range.** webdesign.co.uk itself went 0.25 → 0.51 on the
same site, purely by sampling six pages instead of two.

> **MISSTEP, caught by widening rather than by thinking.** 26 sentences from 2 pages of
> one site was never enough to characterise fleet prose, and the number it produced was
> exactly the number my hypothesis predicted. **A small sample that agrees with you is
> the most dangerous result available.** The only thing that saved it was widening
> before writing it down. Same family as the standing "your measurement answers the
> question you ENCODED" rule, and the loancalculator lane's own thin-baseline missteps
> two days earlier.

### Probe 2 — self-limiting language. REFUTED as measured (but see the reading)

Hypothesis: the copy constantly announces its own limits.

Regex over ~22k words of served copy for "can be wrong / not advice / no guarantee /
cannot guarantee / always check / seek advice / indicative / estimate only / we don't
offer …", plus a hedge-word count and a you:we ratio.

```
site                 paras  words  selfLimit/1k  hedge/1k  you:we
finetuning.uk          115   5266     0.38        5.32     202:68
gaswholesalers.com      96   3635     0.00        1.65     150:135
idea.uk                136   7314     0.55        3.01     208:49
robot-hands.com        105   3985     0.25        0.25      53:26
vonc.com                21    652     0.00        1.53      35:4
webdesign.co.uk         40    839     0.00        2.38      36:2
```

**0.00–0.55 per 1,000 words**, and the copy is already strongly reader-facing.

⚠ **This refutation is the weakest of the three and should not be over-read.** The
reading below found the behaviour the owner described, three times on one page. The
regex missed it because it searched for **phrases** and the behaviour is a **move**. So
probe 2 refutes "the copy uses disclaimer phrases", NOT "the copy avoids announcing its
limits". Those are different claims and I conflated them when I designed the probe.

### Probe 3 — cross-section repetition. REFUTED, and it took a prediction of mine with it

Confirmed first from live config that the writer has no sibling context —
`page-content-writer.process_sections_loop` is `{loop_var: current_section,
iterate_over: sections_ready, max_iterations: 15}`, no accumulator, and none of
`generate_content`'s ten `input_fields` carries another section's output. **Each section
is written by a call that cannot see any other section on the page.**

I predicted repetition. Measured 6-gram Jaccard between paragraph pairs **within** a
page against a **between**-page baseline:

```
site                 within-page >0.10   between-page baseline
finetuning.uk               0.1%                 0.0%
gaswholesalers.com          0.0%                 0.1%
idea.uk                     0.0%                 0.0%
robot-hands.com             0.0%                 0.0%
```

**Indistinguishable.** The near-duplicate opening I had spotted by eye on idea.uk
(paragraphs 1 and 2 both opening "Most people have ideas but nowhere good to work out
whether one is…") is real but isolated — n=1, not a pattern.

> **MISSTEP: I confirmed a MECHANISM and then over-predicted its CONSEQUENCE.** The
> no-accumulator fact is true and read from live config. The damage I attributed to it
> is not there. Keeping the mechanism in the argument required narrowing the claim to
> what it *prevents* (page-level modulation, which I have not measured and may not be
> able to) rather than what it *causes*. **A verified mechanism does not license its
> predicted symptom** — that is the same shape as `damage-confirmed-is-not-mechanism-confirmed`,
> run backwards.

### Then reading it, which is where the fault actually was

`idea.uk/about.html`, three consecutive paragraphs:

- *"We don't tell you whether your idea will succeed."*
- *"We provide the information for you to evaluate your idea, and we don't deliver
  verdicts on whether it will succeed."*
- *"Our guides reflect current thinking and research, but we can still be wrong. You are
  responsible for making the final call on your own work."*

`finetuning.uk/about.html`:

- *"No preferred platforms. No black boxes. Just systems that fit the work you need
  done."* — a negation pile.
- *"We don't have a large org chart or a department for every service. What we have is a
  tight group of practitioners…"*

**That last one is the construction `REVERSE_ENGINEERED_STYLE_PROMPT_v3` round 3
explicitly banned, live on a site today, with the rule sitting in the writer's prompt.**

### The two mechanisms we think explain it

1. **A rule names a FORM; the failure is an INSTINCT.** v3 round 2 banned "isn't/it's";
   round 3 caught the same move in "Nothing here is exotic"; production now does it as
   "We don't have X. What we have is Y." One habit, three spellings, two patches, still
   shipping.
2. **Checkable rules crowd out judgement rules.** "No em dashes" is unambiguous and
   cheap to satisfy; "explain what deserves explanation" is neither. Under pressure a
   model satisfies the first kind. Every tuning round adds more of the first kind. **The
   work we have been doing on this has been making it worse.**

### Status of the script

`prose_stats.py` has now found nothing three times, and that is its value: it is a
**negative control**. It cannot say the copy is good. It will say if a change makes the
copy mechanically worse — which is worth having in place *before* anyone edits a prompt.
⚠ Do not cite it as evidence a copy change worked.
