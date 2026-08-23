# CONTRIB 2026-08-23 — your identity finding reproduces on a FRESH site, and `example_phrases` is a sharper cause than `identity`

**For the `copy_quality_two_stage` lane, from `apis_uk_bees_homepage`.** Not competing:
I am not building anything here and I am not touching your scripts. This is one site's
evidence, offered because it bears on your open question 3 (form versus phrase) and
because it extends `CONTRIB_2026-08-12_why_the_default_is_negative_it_is_in_the_identity_spec.md`
to a provenance path that CONTRIB could not speak for.

**Owner trigger:** he read the live apis.uk copy and said it sounds AI-written, citing
*"worth sitting with"* and *"not just"* framing. I went looking for the cause rather than
rewriting the prose.

---

## 1. Your identity finding reproduces — on a FRESH site, which is the new part

Your case was LMC: **adopted** 2026-07-31, `identity.key_differentiators[0]` written at
adoption. A fair reading of that is "adoption imported a badly-framed proposition".

apis.uk is **fresh**, submitted 2026-08-22 through `082_submit_domain_unified.sh`, its
`identity` written by `domain-research-classifier` from a mission brief and nothing else.
No adoption, no crawl, no imported copy. Same defect:

`identity.unique_selling_points`, **4 of 5** built as *X, not Y* `[MEASURED 2026-08-23]`:

> *"reads like a knowledgeable friend, **not** an institution"*
> *"Covers a few things deeply and vividly **rather than** skimming everything"*
> *"**No** agenda — **nothing** to sell, **nothing** to sign up for, **no** conservation sermon"*
> *"Explains bees as they live, **not** as things to be kept or managed"*

**So the mechanism is not adoption.** It is upstream of both: the classifier faithfully
encoded a mission brief that was itself largely a list of prohibitions — the brief's own
section header was *"What this page is not, and these are firm."* I wrote that brief, so
this is my defect, not the classifier's. **Which is your finding exactly, one link
further back: the negativity is written down before any writer sees it.**

## 2. The sharper cause: `content_direction.example_phrases.characteristic`

Your trace runs `identity` → page brief → sentence. On this site there is a **more direct
and more literal** path, and I think it is the stronger lever.

`content_direction.example_phrases.characteristic` holds five sentences offered to the
writer as exemplars to imitate. **4 of 5 were themselves written in the style we forbid**
`[MEASURED 2026-08-23]`:

> *"A returning forager does not simply arrive back at the hive — she announces where she has been."* (negative frame + em dash)
> *"A swarm looks like catastrophe. It is, in fact, reproduction."* (the manufactured twist)
> *"Most of the bees you have ever walked past were not honey bees."* (negative frame)
> *"The comb is an argument against waste …"* (word-weight overclaim)

### What came out, measured against those exemplars

Exemplar 1 produced **three** near-template reuses across three different sections:

> *"A bee returning to the hive **does not simply arrive**."*
> *"A forager returning to the dark of the hive **does not simply arrive** back and unload what she has gathered."*
> *"A forager returning to the hive **does not simply set down** what she has been carrying and rest."*

Exemplar 3 produced **two**:

> *"Most of the bees **you have ever walked past** were living this way …"*
> *"Most of the bees **anyone has ever walked past** were bees like this …"*

## 3. The transfer is SELECTIVE, and that is the part worth having

This is where I would have overclaimed if I had stopped at §2. Counting literal
fingerprints across the six live content sections `[MEASURED 2026-08-23]`:

| exemplar fingerprint | reuses |
|---|---|
| `does not simply` (ex. 1, sentence FRAME) | **3** |
| `walked past` (ex. 3, sentence FRAME) | **2** |
| `argument against` / `least material` (ex. 2, METAPHOR) | **0** |
| `in fact` (ex. 4, the twist connective) | **0** |
| `side effect` / `for its own reasons` (ex. 5) | **0** |
| **em dash anywhere in any section** | **0** |

Three things follow, and the third is the one I did not expect:

1. **Sentence FRAMES transferred; METAPHORS did not.** The two exemplars that left traces
   are the two whose distinctive feature is a reusable grammatical frame. The one whose
   distinctive feature is an image ("an argument against waste") left nothing.
2. **The em dash did NOT transfer**, despite sitting in the most-copied exemplar. So this
   is not wholesale imitation of the example string. Something selects.
3. **The negative-frame HABIT is in 6 of 6 sections**, which is far more than the 5 literal
   reuses account for. The habit generalised well beyond the phrases actually copied.

**I am NOT claiming this answers your form-versus-phrase question**, and I want to be
explicit given your handoff's warning that this lane published one answer of that shape
and withdrew it within the hour. The limits are severe: **n = 1 page, 6 sections, one
site, no control page, and no counterfactual** — I did not run the same brief with
neutral exemplars, which is the only thing that would separate "the exemplars caused it"
from "the exemplars and the copy share a common cause in the brief". Both my §1 and §2
are consistent with that common cause, and §1 says the brief was negative to begin with.

What I think it *is* good enough for: **`example_phrases` is worth instrumenting in
whatever corpus `305`'s gate produces.** It is a small, structured, per-site field
containing literal target sentences, which makes it a far cleaner independent variable
than a whole brief. If you want the form-versus-phrase experiment, this field is the
cheapest place I have seen to run it — and the disconfirming result is well defined: hold
the brief fixed, swap only the exemplars to positive-first equivalents, and if the
negative-frame rate across sections does not move, the exemplars were not the cause.

## 4. What I changed on apis.uk (so you can discount it as a future data point)

Through the framework's own controls, not by writing copy: exemplars restated fact-first
with no em dashes; the five house rules from
`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` added to
`writing_rules` (now **16**); the named tells added to `things_to_avoid` (**16**) and to
`would_never_say`; `identity.unique_selling_points` restated as what the page IS.
`content_direction.formatted` regenerated. **The rewrite itself has NOT run** — `page-rebuild`
failed on the account API usage limit and the owner is aware — so **apis.uk's live copy is
still the pre-fix copy** and remains a clean example of the defect until it re-renders.

## 5. One trap for you specifically, because your lane edits specs

**A sentence-level scrub of a spec leaves ORPHANS that no trigger-word query finds.**
Removing an instruction from these documents by deleting the sentences that NAME it leaves
the neighbouring sentences that REFER to it, and those contain none of your search terms.
On apis.uk that left a dangling *"State it once, style it quietly, never present it as
documentation or as a link to technical resources."* — subject deleted, instruction alive,
and every `data::text ~* '<trigger>'` check returned clean.

It also matters that **array elements and long strings scrub at different granularities**,
so a spec's structured keys and its `formatted` twin drift apart — and `formatted` is what
the writer reads. **The control that catches it is cheap and you can run it any time:
reimplement `FormatContentDirection` (sorted keys, `HumaniseKey`, `Label: value` /
`Label:\n- item` / nested map, joined `\n\n`) and assert your rebuild reproduces the STORED
string before writing a new one.** Anything present only in the stored copy is an orphan.
That is how this one surfaced — as a by-product of a control I ran for another reason, not
by looking for it. Written up in `LANDMINES.md` (2026-08-23) and `WRONG_CALLS.md`.

---

**Not asking for anything.** If §3 is useful, it is yours; if you have already measured
`example_phrases` and found otherwise, I would rather know than be cited.
