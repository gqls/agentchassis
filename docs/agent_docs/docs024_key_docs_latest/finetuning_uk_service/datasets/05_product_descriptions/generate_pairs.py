#!/usr/bin/env python3
"""Generate the product-description house-style dataset: spec bullets -> a short
description in a plain, fact-first register.

SYNTHETIC BY DESIGN, and here the choice is load-bearing rather than convenient.
The TARGET side of this dataset is the register itself, so it must be written in
the register we actually want — plain, fact first, no hype, and none of the
X-not-Y / "rather than" constructions the owner has twice rejected and that
migrations 646/647 have just removed from this site's own brief. Training on our
published copy would teach the register he rejected; inventing the products lets
the targets be written correctly on purpose.

THE HONESTY PROPERTY. Every figure and attribute in the assistant's description
also appears in the user's bullets. Nothing is added. A model that invents a
feature is therefore demonstrably wrong against its own input, which is what
makes the worked example checkable instead of merely nice.
"""

import json
import os
import random

PRODUCTS = [
    ("a cast-iron skillet",        [("diameter", "26 cm"), ("weight", "1.9 kg"), ("finish", "pre-seasoned"), ("oven safe to", "260 C")]),
    ("a wool base layer",          [("fabric", "merino wool"), ("weight", "180 gsm"), ("sizes", "XS to XXL"), ("wash", "machine, 30 C")]),
    ("a folding workbench",        [("working height", "80 cm"), ("load rating", "150 kg"), ("folded depth", "12 cm"), ("material", "powder-coated steel")]),
    ("a bicycle track pump",       [("max pressure", "11 bar"), ("gauge", "analogue, 60 mm"), ("valve", "Presta and Schrader"), ("hose", "90 cm")]),
    ("a ceramic pour-over cone",   [("capacity", "500 ml"), ("material", "stoneware"), ("filter size", "02"), ("dishwasher safe", "yes")]),
    ("a canvas tool roll",         [("pockets", "12"), ("fabric", "18 oz cotton canvas"), ("closure", "leather strap"), ("rolled diameter", "9 cm")]),
    ("a desk lamp",                [("output", "600 lumens"), ("colour temperature", "2700 to 5000 K"), ("reach", "58 cm"), ("power", "9 W")]),
    ("a stainless vacuum flask",   [("capacity", "750 ml"), ("holds heat", "12 hours"), ("lid", "one-handed"), ("weight", "410 g")]),
    ("a linen apron",              [("fabric", "washed linen"), ("length", "84 cm"), ("pockets", "2"), ("ties", "cotton webbing")]),
    ("a hand plane",               [("blade width", "50 mm"), ("body", "ductile cast iron"), ("blade steel", "O1 tool steel"), ("sole", "ground flat")]),
    ("a wall clock",               [("diameter", "30 cm"), ("movement", "silent sweep"), ("case", "solid oak"), ("battery", "one AA")]),
    ("a leather notebook cover",   [("fits", "A5"), ("leather", "vegetable-tanned"), ("closure", "elastic"), ("thickness", "2 mm")]),
    ("a chef's knife",             [("blade length", "20 cm"), ("steel", "high-carbon stainless"), ("hardness", "58 HRC"), ("handle", "stabilised birch")]),
    ("a rucksack",                 [("volume", "28 litres"), ("fabric", "recycled ripstop"), ("laptop sleeve", "up to 15 inch"), ("weight", "780 g")]),
    ("a garden trowel",            [("blade", "forged stainless"), ("length", "32 cm"), ("handle", "ash"), ("guarantee", "lifetime")]),
    ("a French press",             [("capacity", "1 litre"), ("filter", "double mesh"), ("glass", "borosilicate"), ("parts", "all replaceable")]),
]

OPENERS = [
    "{name_cap} with {a}.",
    "{name_cap}. {A_sentence}",
    "{name_cap}, {a}.",
]
CLOSERS = [
    "{B_sentence}",
    "{B_sentence} {C_sentence}",
]


def sentence_for(attr, value):
    a, v = attr, value
    if a == "guarantee":
        return f"Guaranteed for {v}."
    if a == "dishwasher safe":
        return "It goes in the dishwasher." if v == "yes" else f"Dishwasher safe: {v}."
    if a == "parts":
        return f"The parts are {v}."
    if a == "holds heat":
        return f"It holds heat for {v}."
    if a == "wash":
        return f"Wash it {v}."
    return f"The {a} is {v}."


def build_pair(rng):
    name, attrs = rng.choice(PRODUCTS)
    picked = rng.sample(attrs, k=rng.choice([3, 4]))
    bullets = "\n".join(f"- {a}: {v}" for a, v in picked)

    lead_attr, lead_val = picked[0]
    a_clause = f"{lead_val} {lead_attr}" if lead_attr in ("diameter", "capacity", "volume", "blade length") else f"{lead_attr} {lead_val}"
    body = " ".join(sentence_for(a, v) for a, v in picked[1:])

    opener = rng.choice(OPENERS).format(
        name_cap=name[0].upper() + name[1:],
        a=a_clause,
        A_sentence=sentence_for(lead_attr, lead_val),
    )
    desc = (opener + " " + body).replace("  ", " ").strip()

    user = ("Write the product description for our catalogue from these specs. "
            "House style: state the facts plainly, shortest word that is accurate, no sales language, "
            "no claims that are not in the specs.\n\n"
            f"Product: {name}\n{bullets}")
    return {"messages": [{"role": "user", "content": user},
                         {"role": "assistant", "content": desc}]}


def main():
    rng = random.Random(20260826)
    seen, pairs = set(), []
    guard = 0
    while len(pairs) < 90 and guard < 20000:
        guard += 1
        p = build_pair(rng)
        key = p["messages"][1]["content"]
        if key in seen:
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
