#!/usr/bin/env python3
"""Dataset 1 — email copywriting voice.

The TARGET side is the owner's own email, verbatim from the anonymised corpus.
The USER side is a brief I wrote describing the SITUATION and the INTENT — never
the phrasing, because a brief that contains the wording teaches copying rather
than voice (the estate's own "a quoted exemplar in a prompt is copied verbatim"
lesson).

Provenance: owner-supplied 2026-08-26, anonymisation approved by him the same day.
See ../_owner_voice_corpus/README.md for exactly what was removed.
"""

import json
import os

CORPUS = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                      "..", "_owner_voice_corpus", "cleaned", "11_emails_cleaned.txt")

BRIEFS = {
 0:  "Someone sent you an unexpected gift in the post and it has arrived. Thank them.",
 1:  "You thanked someone for a gift and have realised it was actually from a different sender. Correct it without making it awkward, and ask after the holiday they have just been on.",
 2:  "The sender confirmed a gift is still on its way to you. Acknowledge it and say you are looking forward to it.",
 3:  "A domain broker you buy from has offered you a domain. You cannot buy right now. Decline, but say honestly what you admired about a site of theirs, and mention you are away on holiday.",
 4:  "Turn down another domain offer from the same broker. You are deliberately not buying at the moment. Be light about it.",
 5:  "Decline yet another domain offer. Very short. Keep it warm.",
 6:  "Decline a domain offer, joking that you are on a self-imposed buying ban and finding it hard.",
 7:  "A company you deal with sent you a piece of clothing as a gift and you do not know which address to thank. Thank them and say what arrived.",
 8:  "Decline a domain offer, explaining you want to develop and sell one of the ones you already own before buying more. Acknowledge it was good value.",
 9:  "You are late replying to a domain offer. Decline for now but ask whether you can come back to them about it later.",
 10: "A recruiter has asked to put you forward for a security-focused contract. Say yes and set out your relevant experience: running your own company, managing teams, legacy and modern code, a lapsed security clearance, penetration testing project management, and a modern AI agent framework you built. Be specific about the architecture.",
 11: "That application was unsuccessful. Ask the recruiter whether there was any real feedback, so you can adjust next time.",
 12: "The recruiter has explained why there was no feedback. Accept it and wish them luck with their other candidates.",
 13: "Decline another domain offer. Two lines.",
 14: "Reply to a short friendly message about the weather. One or two lines.",
 15: "Accept a domain a broker has offered you as a gift. Explain you are not developing much at the moment, reflect on how AI has changed your work — both that it makes things easy for everyone and that it makes things easier for you — and say you will pay now and that they should feel free to change their mind.",
 16: "A domain you were given has attracted a four-figure offer within a week. Tell the person who gave it to you.",
 17: "Arrange the mechanics of a domain transfer. Give your username and email and say the DNS arrangement suits you.",
 18: "The domain transfer went through. Thank them, and answer their question about your work situation honestly — you are between jobs and the market looks very different from a few years ago.",
 19: "Reply to a friendly, joking message about your reputation and your job hunt. Match the humour. Accept an offer of another domain graciously without pressing for it.",
 20: "You have been offered a choice of two domains. Pick one, give the reason, and acknowledge the generosity.",
 21: "A supplier has flagged a possible duplicate payment. Say you are looking into it and will come back to them.",
 22: "Confirm you are happy to meet in person and will book a time.",
 23: "You have booked a meeting for 9am tomorrow. Ask where to meet and say you can travel.",
 24: "Propose where to meet: their town is actually easier for you to drive to than the alternative, and you have coffee vouchers to use up. Offer the alternative too if it suits them better.",
 25: "Confirm a meeting place in one line.",
 26: "The person you were meeting is unwell and has cancelled. Wish them well and make clear there is no pressure to rearrange quickly.",
 27: "Rearrange the meeting: give the day, the place and the time.",
 28: "New Year message to someone you know socially. Describe your low-key New Year's Eve, note how busy they seem, and ask to be unsubscribed from a newsletter that is going to the wrong address — the social address is the right one.",
 29: "Confirm which of your email addresses is which, respond warmly to a photo they sent, and wish them a good year.",
 30: "Follow up after a meeting that went well. Say so, offer to show them a tool you are building when it is ready, and offer help with their search.",
}


def main():
    text = open(CORPUS, encoding="utf-8").read()
    emails = [p.strip() for p in text.split("--- email ---") if p.strip()]
    pairs = []
    for i, body in enumerate(emails):
        brief = BRIEFS.get(i)
        if not brief:
            continue
        user = ("Write this email in my voice.\n\n"
                f"Situation: {brief}")
        pairs.append({"messages": [{"role": "user", "content": user},
                                   {"role": "assistant", "content": body}]})
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pairs.jsonl")
    with open(out, "w", encoding="utf-8") as fh:
        for p in pairs:
            fh.write(json.dumps(p, ensure_ascii=False) + "\n")
    print(f"wrote {len(pairs)} pairs of {len(emails)} emails")


if __name__ == "__main__":
    main()
