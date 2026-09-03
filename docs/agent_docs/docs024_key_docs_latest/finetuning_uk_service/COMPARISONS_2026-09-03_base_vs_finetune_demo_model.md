# COMPARISONS 2026-09-03 — the demo model (Phase 0 fine-tune) against its untrained base, on the prompts it was trained for

**Why this file exists.** The owner looked at the live playground (2026-09-03 evening) and asked for
"an example of input and output or two" and "a clearer explanation of what the model is doing for
me", then: *"also publish those comparisons, that will help explain what it's doing."* Before any of
that goes on the page, here is what the model actually does, measured, so the page cannot say more
than the model shows. **Verbatim outputs; nothing edited.** Placeholders like `[NAME]` are the
anonymisation in the training corpus, reproduced by the models.

**Set-up `[MEASURED 2026-09-03 20:45–20:55Z]`.** Base = `smollm2:1.7b` (SmolLM2-1.7B-Instruct, the
model Phase 0 started from), pulled onto the in-cluster `ollama-adapter` for this comparison (not the
visitor path). Fine-tune = `finetuning-demo` (`cd4c8ea62f1d`), the SAME artefact the public demo
serves. Prompts are HELD-OUT rows from the training datasets (never trained on), in the exact shape the
model was trained on ("Write this email in my voice. Situation: …", "Reply to this in my voice.",
"Rewrite this in my voice", "Summarise this in our house format"). Where the held-out row has the
owner's real answer, it is shown as TARGET. Live-demo runs went through the public route; in-cluster
runs through `ollama run` (CPU, ~14 tok/s; the demo box does ~40).

## The honest reading, first

- **The fine-tune did not learn the owner's voice.** 295 rows over six tasks (20-ish pairs per voice
  task) moved a 1.7-billion-parameter model a little on one prompt and not at all, or backwards, on
  the others. Training loss 1.41 → 0.73 (RESULTS 08-15) measured fit to the training rows, not this.
- **Pair A (decline a domain offer) is the only one with a visible improvement:** shorter, British
  "holiday" for the base's "vacation", less gushing, structure closer to the brief. Still an
  off-the-shelf business email, not the target's *"I just can't just yet! … Am on hols at the moment"*.
- **Pair B (late reply) got WORSE:** the fine-tune is longer and stiffer than the base.
- **Pairs C (rewrite in my voice) and Q2 (reply in my voice): the fine-tune ECHOES the input** (C: both
  models echo; Q2: the fine-tune returns the inbound message nearly verbatim). Echoing is a training
  defect signature (the response-marker filter dropped 5/300 rows for truncation; the model may have
  learned "copy the user turn" from rows where the boundary was unclear) — worth the training lane's
  attention before more rows are added.
- **Pair D (house-format summary): the fine-tune returns only the title.** Degenerate.
- **Off-distribution prompts** (what visitors actually type: "what should I check before signing a
  lease") get the base model's generic answer, with American spelling.

So: **as a demonstration of "your company's voice, in a model you own", the current demo model
demonstrates the gap, not the product.** Publishing the comparisons is honest and educational only if
the page says that plainly ("this is what a first, small fine-tune does; the effect is modest and
uneven"). Publishing them under a claim of success would be the site's original sin again (08-25).

## Pair A — decline a domain offer (held-out, dataset 01_email_voice) — the one with an improvement

**Prompt:** Write this email in my voice. Situation: A domain broker you buy from has offered you a
domain. You cannot buy right now. Decline, but say honestly what you admired about a site of theirs,
and mention you are away on holiday.

**TARGET (the owner's real email, anonymised):**
> Hi [NAME], Thank you. I just can't just yet! I loved your [DOMAIN] it was a really emotive image and
> good copy. I'll show you some of my sites shortly, not quite as exciting I'm afraid! Thanks [NAME],
> it is a fantastic domain and I would normally have grabbed it with both hands. Am on hols at the
> moment in [LOCATION] and just off to the beach. Happy times!

**BASE (smollm2:1.7b):**
> Subject: Grateful for Your Honesty, but I'm on Vacation Right Now
> 
> Dear [Domain Broker's Name],
> 
> I hope this email finds you well. I wanted to express my sincere gratitude 
> for the domain you recently offered me, [Domain Name]. I am truly impressed
> impressed by your site and the effort you put into maintaining it. Your sit
> site is not only well-designed, but also user-friendly, and the content you
> you provide is consistently high-quality.
> 
> I admire the professionalism and dedication you display in your work, and I
> I'm so glad to see a reputable site in your portfolio. I'm sure your site w
> will continue to be a great source of information for your audience.
> 
> I would like to extend my sincere thanks to you and your team for consideri
> considering me for the domain. Unfortunately, I'm currently on vacation and
> and won't be able to purchase 

**FINE-TUNE (finetuning-demo, in-cluster run):**
> Subject: A Domain We Could Have Bought If Only...
> 
> Dear [Domain Broker's Name],
> 
> I hope this email finds you well. As I'm currently on holiday, I'm unable t
> to respond right now to your offer for the domain you've recently listed.
> 
> I really admired the work you've done on your own site. The design and the 
> content are both top-notch, and I think your site stands out in a great way
> way. It's clear you put a lot of effort into creating a unique, professiona
> professional, and engaging platform.
> 
> I wish we could have had the chance to make a deal, but it seems the timing
> timing didn't work out. Thank you for the opportunity, and I hope to catch 
> up with you once I'm back home.
> 
> Best,
> [Your Name]
> 
> 

**FINE-TUNE, the LIVE demo's own answer to the same prompt (through the public route, 150-token cap):**
> Subject: Re: Consideration for Domain Offer
> 
> Dear [Domain Broker's Name],
> 
> Thank you for extending an offer to the domain that resonated with me. Your domain name, [Domain Name], is both visually appealing and uniquely positioned for the niche market you've targeted. I particularly enjoyed the clear structure and SEO-friendly composition of the site's design, which suggested a well-planned and executed brand identity.
> 
> I appreciate that you understood my current business constraints and opted to keep the domain open for potential acquisition at a later date. Given that I am currently on vacation, I am unable to proceed with an immediate purchase. However, I am eager to explore this opportunity further when I return from my break.
> 
> Thank
> 

## Pair B — a late reply to a domain offer (held-out, 01_email_voice) — the fine-tune is WORSE

**Prompt:** Write this email in my voice. Situation: You are late replying to a domain offer. Decline
for now but ask whether you can come back to them about it later.

**TARGET (owner):** Just seen this, sorry. That's very kind of you [NAME]. Not for now, I'll ask you
again later if that's ok? That's very nice of you. Kind Regards, [NAME]

**BASE:**
> Subject: Re: Late Domain Offer Update
> 
> Hello [Name],
> 
> Thank you for reaching out. I apologize for not responding sooner, as I was
> was delayed with my current tasks. 
> 
> I would love to discuss the domain offer with you when I have a moment. Per
> Perhaps we could schedule a time that works best for both of us? 
> 
> Looking forward to our conversation.
> 
> Best,
> [Your Name]
> 
> 

**FINE-TUNE:**
> Subject: Delay in Domain Reply - Request to Review Your Availability
> 
> Dear [Domain Name Owner],
> 
> I hope this message finds you well. I'm writing to inform you that I've bee
> been diligently working on replying to your domain offer, but unfortunately
> unfortunately, I've encountered a delay that I regret not catching earlier.
> earlier.
> 
> I apologize for any inconvenience this may have caused and appreciate the u
> understanding that you've given us the benefit of the doubt. I'm reaching o
> out to request a moment of your time to reassess my availability at your ea
> earliest convenience, should you wish to revisit this opportunity when thin
> things have settled back into a more predictable routine.
> 
> In

## Pair C — rewrite in my voice (held-out, 03_copy_style) — BOTH models echo the input

**Prompt:** Rewrite this in my voice. Keep every fact. Do not shorten it by dropping content.
"Organisations frequently operate multiple issue-tracking platforms concurrently during transitional
periods. Departmental adoption patterns are typically driven by established working practices rather
than by formal directive, and integration work is commonly deprioritised relative to defect resolution."

**TARGET (owner):** So, back to the office, even the tracking software that's being used is in that
limbo state of transition from say Redmine to Jira. Both are being used by different departments, their
behaviour being more influenced by their ingrained habits than the mild sniping coming from someone
else's project m…

**BASE and FINE-TUNE, identically:** the input paragraph returned word for word.

## Q2 — reply in my voice (held-out, 04_reply_tone) — the LIVE fine-tune echoes the inbound

**Prompt:** Reply to this in my voice. "No problem at all! Though I should say — I don't think that
one was from me. Hope it was nice whoever sent it. Just back from two weeks away, feeling human again."

**TARGET (owner):** Hi [NAME], I think, actually, this one was from someone else, sorry! I'll let you
know when yours arrives, haha! Hope you had a great holiday. Kind Regards

**FINE-TUNE (live demo):**
> No problem at all. Hope it was nice whoever sent it. Just back from two weeks away, feeling human again.

## Pair D — summarise in the house format (held-out, 06_doc_summaries) — the fine-tune degenerates

**Prompt:** Summarise this in our house format. What it is: one line / What it says: three bullets /
What to do: one line… [an article on persuasive presentations, Aristotle's ethos/pathos/logos]

**BASE:**
> You've already summarized the main points:
> 
> 1. Start by asking Aristotle about making presentations convincing.
> 2. He suggested the three modes of persuasion: Ethos (credibility), Pathos 
> (emotions), and Logos (logical reasoning).
> 3. Ethos is based on your credibility and recent successes in the field.
> 4. Pathos is about emotional appeal, understanding the audience's needs, an
> and creating a competitive edge.
> 5. Logos is about using logic and facts to justify your pitch.
> 
> You've also mentioned some additional tips and emphasized the importance of
> of prepping well.
> 
> Next steps could include delving deeper into how to effectively use each mo
> mode of persuasion in your pitch, such as how to dem

**FINE-TUNE:**
> Seduce Your Clients. How to Make Your Presentations Jaw-Dropping.
> 
> 

## Off-distribution, what visitors actually type (live demo, 20:40Z)

- "A customer emails to ask whether we can deliver a kitchen fit-out in three weeks. Draft a short
  reply." → a generic "Dear Customer… Kind Regards, [Your Name]" letter.
- "Write two sentences introducing our accountancy firm to a new client." → "…personalized attention…
  comprehensive accounting solutions tailored to the unique needs of every client."
- "What should I check before signing a lease on a small shop?" → a sensible generic five-point list.

## What this means for the page, and the decision it puts to the owner

1. The page can honestly say: *this is a small model (1.7 billion parameters, a gigabyte) fine-tuned
   on a few hundred examples of one person's writing; ask it to write or reply to an email in that
   voice and you will see a small shift from the untrained model; ask it anything else and you get the
   small model's general answer.* Pair A, base against fine-tune, verbatim, illustrates exactly that.
2. It cannot honestly say the demo shows "your company's voice" — it does not, yet. The product claim
   on `/your-own-model.html` is about a customer's model trained on THEIR documents; the demo is our
   first, thin run. Either the page says so, or the demo model is improved before it carries the claim.
3. The echo behaviour (Q2, C) and the degenerate summary (D) are training defects to fix before the
   next run, not copy problems.
