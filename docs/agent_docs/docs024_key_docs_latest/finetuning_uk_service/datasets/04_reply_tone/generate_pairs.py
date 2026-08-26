#!/usr/bin/env python3
"""Dataset 4 — reply tone: an inbound message in, the owner's actual reply out.

TARGET is verbatim his, from the approved corpus. The INBOUND side is reconstructed
by this lane from what the reply plainly responds to — and reconstruction is the
honest word for it: the original inbound messages were never supplied, so these are
inferred, not quoted. They are written as a correspondent would write them, NOT as a
summary of the reply, because an inbound that restates the reply teaches echoing.

⚠ SMALLER THAN THE DESIGN ASKED FOR (60-200). ~20 usable replies exist. Recorded as
a real limit rather than padded with invented correspondence.
"""

import json
import os

CORPUS = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                      "..", "_owner_voice_corpus", "cleaned", "11_emails_cleaned.txt")

# index in the corpus -> the inbound message this reply is answering
INBOUND = {
 1:  "No problem at all! Though I should say — I don't think that one was from me. Hope it was nice whoever sent it. Just back from two weeks away, feeling human again.",
 3:  "Morning — got a lovely one that might suit you. Good length, clean history, easy to remember. Yours for the usual if you want it.",
 4:  "Another one for you, and I think this is the best I've offered you yet. Genuinely would keep it myself if I had the time.",
 5:  "One more, and then I'll leave you alone I promise. Have a look at least.",
 6:  "Go on. You know you want it.",
 8:  "This one's priced well below what it's worth — I'd move quickly if you're interested.",
 9:  "Did you see the one I sent last week? No pressure, just checking it didn't get lost.",
 11: "Afraid it's a no from them on this occasion. They've gone with another candidate.",
 12: "Nothing specific I'm afraid — they rarely give detail at this stage, it's usually just fit against the other CVs on the pile.",
 13: "New batch in this morning, thought of you for this one.",
 15: "I've got one here I think you'd actually do something with. Have it — no charge, call it a thank you for being easy to deal with.",
 17: "Great — how do you want to do the transfer? I can push it or do it via the registrar, whichever suits.",
 18: "All pushed through, should be showing on your side now. How's work going, by the way — still busy?",
 20: "Two for you to pick between. Genuinely can't decide which is the better buy, so you choose.",
 21: "Just flagging that we seem to have received two payments against the same invoice — could you check your end?",
 22: "Happy to do this in person or over a call, whichever you'd prefer. Let me know and I'll send some times.",
 23: "Booked in, thanks. Where suits you for it?",
 26: "So sorry — I've come down with something and I don't think I'll be much use tomorrow. Would you mind if we moved it?",
 27: "Feeling much better, thank you. When's good for you to pick this back up?",
 30: "Really good to meet you today. I'll have a think about everything we discussed and come back to you.",
}


def main():
    text = open(CORPUS, encoding="utf-8").read()
    emails = [p.strip() for p in text.split("--- email ---") if p.strip()]
    pairs = []
    for idx, inbound in INBOUND.items():
        user = ("Reply to this in my voice.\n\n" + inbound)
        pairs.append({"messages": [{"role": "user", "content": user},
                                   {"role": "assistant", "content": emails[idx]}]})
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pairs.jsonl")
    with open(out, "w", encoding="utf-8") as fh:
        for p in pairs:
            fh.write(json.dumps(p, ensure_ascii=False) + "\n")
    print(f"wrote {len(pairs)} pairs")


if __name__ == "__main__":
    main()
