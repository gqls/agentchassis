#!/usr/bin/env python3
"""Generate the copy-STRUCTURE dataset: a raw content dump -> the same content
in a house section order.

WHY THIS TASK IS SYNTHETIC AND THAT IS THE RIGHT CHOICE. This dataset teaches
SHAPE, not voice: the target is the same sentences the input already contains,
reordered and headed. Nothing here has to sound like anybody, so inventing the
material costs nothing that matters — and the alternative (our own published
copy) is the register the owner has twice rejected, so it would be the wrong
teacher for anything voice-shaped. See PROVENANCE.md.

HOW THE VARIETY IS REAL RATHER THAN COSMETIC. Each pair draws a different
business, a different set of facts, a different subset of sections, and a
different shuffle of the dump. The assistant side is composed from the SAME
sentence strings as the user side — so a model that invents new claims is
demonstrably wrong, and a worked example can say so.
"""

import json
import os
import random

# The house section order this dataset teaches. It is the order the site builder
# already uses, so a model trained on it produces plans the pipeline can take.
HOUSE_ORDER = ["What this is", "Who it is for", "How it works", "What it costs", "What happens next"]

BUSINESSES = [
    ("a mobile bike repair service", "riders who cannot get to a shop"),
    ("a small-batch coffee roastery", "cafés buying under 20kg a month"),
    ("a two-person accountancy practice", "sole traders filing their first return"),
    ("a garden design studio", "owners of small urban gardens"),
    ("a commercial cleaning contractor", "managers of single-site offices"),
    ("a wedding photography business", "couples booking under twelve months out"),
    ("a boiler servicing firm", "landlords with three or more properties"),
    ("a dog-walking service", "owners working full days from an office"),
    ("a picture framing workshop", "galleries and individual collectors"),
    ("a translation agency", "exporters sending documents to Germany"),
    ("a driving instructor", "learners retaking a test after a fail"),
    ("a physiotherapy clinic", "runners with recurring injuries"),
    ("a printing company", "small publishers doing runs under 500"),
    ("a domestic electrician", "homeowners rewiring a single room"),
    ("an equipment hire yard", "builders on jobs of a week or less"),
    ("a bookkeeping service", "shops that still keep paper receipts"),
]

WHAT_IT_IS = [
    "We {verb} {thing}, and that is the whole of what we do.",
    "This is {thing} — {verb} properly, by people who do it every day.",
    "{thing_cap}, {verb} to a standard you can check.",
]
HOW = [
    "You get in touch, we look at what you have, and we tell you what it needs.",
    "First a short call, then a written quote, then the work itself.",
    "We come to you, assess it on the spot, and start the same week where we can.",
    "Send us the details, we come back within a day with a price and a date.",
]
COST = [
    "Most jobs land between £{lo} and £{hi}. Anything outside that we say so before starting.",
    "It starts at £{lo}. We will tell you the final figure before any work begins.",
    "£{lo} covers the standard job; £{hi} is the most we have charged this year.",
]
NEXT = [
    "Call, or fill in the form, and we will come back to you the same working day.",
    "Send a message with a rough idea of what you need and we will take it from there.",
    "Book a slot online and we will confirm it within the hour.",
]
VERBS = ["fix", "make", "handle", "look after", "sort out", "service"]
THINGS = ["the work", "the job", "the whole thing", "what needs doing"]


def build_pair(rng):
    biz, audience = rng.choice(BUSINESSES)
    verb, thing = rng.choice(VERBS), rng.choice(THINGS)
    lo = rng.choice([45, 60, 80, 120, 150, 200])
    hi = lo * rng.choice([2, 3, 4])

    sections = {
        "What this is": rng.choice(WHAT_IT_IS).format(
            verb=verb, thing=thing, thing_cap=thing.capitalize()),
        "Who it is for": f"This is for {audience}.",
        "How it works": rng.choice(HOW),
        "What it costs": rng.choice(COST).format(lo=lo, hi=hi),
        "What happens next": rng.choice(NEXT),
    }
    # Drop a section sometimes: a real dump is usually incomplete, and a model
    # that hallucinates the missing one is failing in the way that matters.
    chosen = [s for s in HOUSE_ORDER if s in sections]
    if rng.random() < 0.35:
        chosen.remove(rng.choice(chosen[1:]))

    dumped = chosen[:]
    rng.shuffle(dumped)
    dump = "\n".join(sections[s] for s in dumped)

    ordered = "\n\n".join(f"## {s}\n{sections[s]}" for s in chosen)
    user = (f"Here are some notes for {biz}. Put them into our standard section order "
            f"and add the headings. Do not add anything that is not already here.\n\n{dump}")
    return {"messages": [{"role": "user", "content": user},
                         {"role": "assistant", "content": ordered}]}


def main():
    rng = random.Random(20260826)  # fixed: a worked example must be reproducible
    seen, pairs = set(), []
    while len(pairs) < 90:
        p = build_pair(rng)
        key = p["messages"][1]["content"]
        if key in seen:          # no duplicate targets: they teach nothing twice
            continue
        seen.add(key)
        pairs.append(p)
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pairs.jsonl")
    with open(out, "w", encoding="utf-8") as fh:
        for p in pairs:
            fh.write(json.dumps(p, ensure_ascii=False) + "\n")
    print(f"wrote {len(pairs)} pairs to {out}")


if __name__ == "__main__":
    main()
