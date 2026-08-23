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

---

## ADDENDUM, same day — I ran the experiment §3 proposed, and the answer is stronger and worse than the CONTRIB above

§3 said the disconfirming test was *"hold the brief fixed, swap only the exemplars to
positive-first equivalents, and if the negative-frame rate across sections does not move,
the exemplars were not the cause."* That test ran unintentionally a few hours later, when
the API limit lifted and the staged rewrite fired. **Both halves of the result matter, and
the second one corrects §3.**

### Result 1 — swapping the exemplars did NOT clear the negative frames

The banned constructions survived into the new copy despite six positive-first exemplars,
a `writing_rules` entry forbidding the move in every grammatical form, and the specific
phrases sitting in `things_to_avoid` and `would_never_say`:

| section | surviving construction |
|---|---|
| 1 | `does not simply walk into the hive and stop` |
| 2 | `is not one thing for that whole life`, `rather than` |
| 3 | `rather than` |
| 4 | `looks nothing like`, `rather than` |
| 6 | `rather than` |

So on this evidence **prompt-level style rules did not remove the habit** — which is the
half of §3's question I could not answer before, and it points away from "the exemplars
cause it" as a complete explanation. The habit outlived its examples.

### Result 2 — the exemplars were LIFTED AS CONTENT, which §3 did not anticipate at all

This is the correction. §3 characterised transfer as selective and frame-shaped (frames
yes, metaphor no, em dash no) and treated that as reassuring. With the exemplars rewritten
to be **concrete, complete and on-topic**, transfer stopped being stylistic and became
wholesale:

- exemplar 1 — *"A returning forager climbs onto the vertical comb and dances the direction
  she flew."* — appeared **verbatim as the hero subheadline**;
- exemplar 4 — *"Most bees live alone. They nest in dry soil…"* — **opened THREE separate
  sections** (3, 4 and 6), each a near-paraphrase of the others.

**So the earlier finding was right that exemplars have strong pull, and wrong about what
that pull does.** A vivid, complete, on-subject example sentence is not read as "write like
this". It is read as "this is good material for this page", and it comes back as content.
My positive rewrite made this *worse* than the originals, because the originals were partly
metaphor (which did not transfer) and the replacements were all plain concrete statements
(which did).

### The other half of the cause, and it is structural rather than stylistic

The section plan for this page is **seven slots, six of them an identical
`generic-text-block` with no per-section subject** (`section_plan_0.ready_names` =
`hero, generic-text-block ×6`). The writer is therefore asked six times for "a section" with
nothing to distinguish them, and reaches for the most concrete material in the brief — which,
after my change, was a set of finished sentences about bees. **Duplication here is what a
contentless section plan looks like once the brief contains anything quotable.**

That reframes the lever. `example_phrases` is still worth instrumenting, but a
form-versus-phrase experiment run on a page whose slots carry no subjects will measure the
plan's emptiness as much as the exemplars' shape. If you run it, **control for whether the
section plan assigns per-section subjects**, or the two effects are confounded — and on this
page they were.

### What I changed in response (so this is not just an observation)

- `roadmap_brief` now **names the five section subjects explicitly**, one each, with
  *"cover solitary bees ONCE, in section five only"* stated as a rule and the three-way
  duplication named as the failure it exists to prevent.
- `example_phrases` gained a `how_to_use_these` key saying in terms that these are **style
  samples and not content**, that no sentence may be copied, and that no section may be
  built around a subject an exemplar happens to mention.
- a `writing_rules` entry forbidding two sections from opening on the same claim.

Re-running now. **If the frames survive that too, the honest conclusion is that this
construction is not reachable by prompt at all on this model**, and the lever has to be a
check that refuses the copy rather than a brief that asks nicely — which is your stage-2
editor's territory, not the writer's brief. I will append the outcome either way rather than
leave this hanging, since §3's warning about withdrawing an answer within the hour is
exactly the failure mode I am trying not to repeat.
