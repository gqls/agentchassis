#!/usr/bin/env python3
"""Dataset 6 — internal-doc summarisation: a long document in, a fixed house format out.

⚠ THIS ONE IS DIFFERENT FROM 1/3/4 AND THE DIFFERENCE MATTERS FOR HONESTY.
The TARGET here is a FORMAT, not the owner's voice. The inputs are his own articles
from the approved corpus, but the summaries are written by this lane to a fixed
shape. So the dataset teaches structure and compression — it does NOT teach his
register, and a worked example built from it must not claim to.

That is exactly why the design (DESIGN_2026-08-25 §2) put a non-voice task in the
set: a prospect whose problem is volume rather than tone never sees themselves in a
row of voice demos.

THE HOUSE FORMAT, fixed across every row:
    What it is:      one line
    What it says:    three bullets, most important first
    What to do:      one line, or "Nothing — background reading."
"""

import json
import os

CORPUS_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                          "..", "_owner_voice_corpus", "cleaned")

SUMMARIES = {
 "01_pitching_aristotle.txt": (
   "What it is: an article on preparing a business pitch, framed through Aristotle's three modes of persuasion.\n\n"
   "What it says:\n"
   "- Ethos is your credibility, and the audience has largely formed it before you start speaking.\n"
   "- Pathos is the emotional argument, including your own stake in winning or losing.\n"
   "- Logos is the logical case, and it needs real figures and analysis.\n\n"
   "What to do: brush up your credentials and start writing the pitch down now."),
 "02_agile_teenagers_bedroom.txt": (
   "What it is: a field report on what an Agile office actually looks like after a few years.\n\n"
   "What it says:\n"
   "- Two tracking systems run in parallel during transitions, and integration work loses to bug work.\n"
   "- The backlog becomes a record of abandoned ideas, and that costs morale and staff.\n"
   "- Contractor churn means constant handover and months of reduced output.\n\n"
   "What to do: adopt one tracking tool properly, train the business in it, and track time for consistency over accuracy."),
 "03_jira_scope_creep.txt": (
   "What it is: a warning about organisational dependence on a single tracking tool.\n\n"
   "What it says:\n"
   "- As the tool spreads across departments, maintenance stays with the most time-poor team.\n"
   "- Two real incidents — an expired licence needing a reboot, and a misconfigured cloud migration — took time to diagnose.\n"
   "- A tool that raises productivity lowers it by the same amount when it fails.\n\n"
   "What to do: treat the tracking tool as production infrastructure, not an internal convenience."),
 "04_drupal_hosting.txt": (
   "What it is: an assessment of a content platform, and a hosting offer built around its weaknesses.\n\n"
   "What it says:\n"
   "- The platform is genuinely capable — modules, templating, theming, and an active community.\n"
   "- It is also heavy to host and maintain, especially when inherited from another builder.\n"
   "- The offer covers updates, security patching, caching, backups and migration.\n\n"
   "What to do: nothing unless you host one of these and are maintaining it yourself."),
 "05_ticket_mining.txt": (
   "What it is: a short proposal for mining historical bug-tracker data.\n\n"
   "What it says:\n"
   "- Old tickets across several systems could be categorised by type, area, origin and estimate accuracy.\n"
   "- The result would inform resourcing, budgets and which code to tackle.\n"
   "- It is difficult but not impossible.\n\n"
   "What to do: nothing yet — background reading."),
 "06_pellet_burners.txt": (
   "What it is: a consumer guide to heating rooms with wood pellet stoves.\n\n"
   "What it says:\n"
   "- Pellets are cheaper than wood and competitive with mains gas, and supply is growing.\n"
   "- Most of the saving comes from behaviour: heating one room you are in, not a whole empty house.\n"
   "- Installation is straightforward and the unit costs about half a year's heating bill.\n\n"
   "What to do: nothing — background reading."),
 "07_sites_for_sale_seo.txt": (
   "What it is: observations from reading website sale listings.\n\n"
   "What it says:\n"
   "- Traffic and content quality correlate weakly; thin sites can carry very large audiences.\n"
   "- High-quality backlinks are the clearest driver, and often come from relationships rather than technique.\n"
   "- Design quality appears to matter less than expected above a minimum standard.\n\n"
   "What to do: nothing — background reading."),
 "08_cartoon_ai.txt": (
   "What it is: an experiment turning drawn comic characters into photorealistic images.\n\n"
   "What it says:\n"
   "- Source images were cleaned up by hand and fed to the model front-facing only.\n"
   "- Results arrived in minutes and were recognisably people, if not quite as imagined.\n"
   "- The model handled angled faces and inconsistent eye colour poorly.\n\n"
   "What to do: nothing — background reading."),
 "09_christmas_gifts.txt": (
   "What it is: a suggestion list for keeping staff and clients in touch over a remote Christmas.\n\n"
   "What it says:\n"
   "- The usual office rituals are unavailable, so morale needs something deliberate.\n"
   "- Clients are at home receiving competitors' offers, so a reminder is worth sending.\n"
   "- Personalised items work if the photo is swapped for the company logo.\n\n"
   "What to do: pick something small and send it to staff and top clients."),
 "10_copywriter_marketplace_rules.txt": (
   "What it is: the operating rules for a copywriting marketplace.\n\n"
   "What it says:\n"
   "- Buyers may approach at most two writers at a time, and writers take at most two briefs.\n"
   "- Scope creep is named as the main source of disputes, on both sides.\n"
   "- Quality is hard to measure and price does not reliably predict it.\n\n"
   "What to do: be explicit about the limits of the brief before work starts."),
}


def main():
    pairs = []
    for fname, summary in SUMMARIES.items():
        body = open(os.path.join(CORPUS_DIR, fname), encoding="utf-8").read().strip()
        user = ("Summarise this in our house format.\n\n"
                "What it is: one line\n"
                "What it says: three bullets, most important first\n"
                "What to do: one line, or \"Nothing — background reading.\"\n\n"
                "---\n\n" + body)
        pairs.append({"messages": [{"role": "user", "content": user},
                                   {"role": "assistant", "content": summary}]})
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pairs.jsonl")
    with open(out, "w", encoding="utf-8") as fh:
        for p in pairs:
            fh.write(json.dumps(p, ensure_ascii=False) + "\n")
    print(f"wrote {len(pairs)} pairs")


if __name__ == "__main__":
    main()
