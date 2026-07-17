# Plain voice v2 — the readable register (owner decision, 2026-07-17)

**What changed and why.** The v1 voice (VOICE_REWRITE_PROMPT.md) killed the marketing tells
and the fabrications, but the copy it produced is dense: long sentences that pack three ideas,
literary turns ("laundered rumour", "unglamorous, and that is the point"), no contractions,
em-dash rhythm. The owner reviewed it against a plainer rewrite and ruled: **move further in
the plain direction.** Matter-of-fact, friendly, readable. Explaining, not performing.

v1's honesty rules all survive. This is a register change, not a truth change.

## The rules (delta from v1)

1. **One idea per sentence.** If a sentence carries two, split it. A sentence past ~20 words
   is suspect; past ~30 is wrong.
2. **Contractions are in.** "It's", "we'd", "you're", "isn't". The old copy's formal
   "we would rather" reads stiff; write like you talk.
3. **Short paragraphs.** One to three sentences. White space is a feature.
4. **Aim near Flesch 80.** Practically: short words, short sentences, active voice. "Use" not
   "utilise". "Help" not "facilitate". Prefer the Anglo-Saxon word over the Latinate one.
5. **No literary moves.** No aphorisms, no compressed elegance, no "and that is the point"
   landings. If a phrase feels quotable, simplify it.
6. **Em-dashes near zero.** Use a full stop or a comma instead.
7. **Talk to the reader.** "You" and "we" throughout. Start with And or But when it's natural.
8. **Friendly is calm, not chummy.** No "You know what?", no "honestly", no rhetorical
   questions as a tic, no forced warmth. Friendliness here means: easy to read, nothing to
   decode, no sales pressure.
9. **Keep the numbers and the specifics.** Concreteness stays. Every number still traces to
   the evidence base. Plain does not mean vague.
10. **Rhythm still varies** — a few very short sentences among the medium ones. But the
    variance lives in a narrow, readable band, not between staccato and baroque.

## What we deliberately did NOT adopt from the reference material

The source material (humanizing-prompt threads) mixes good plain-language advice with
AI-detector-evasion tricks. Rejected, with reasons:

- **Deliberate errors, slang, "mess up the grammar a lil"** — this is a consulting site;
  errors read as carelessness, not humanity.
- **Casual fillers ("You know what?", "Honestly…")** — forced friendliness; v1 was right
  to keep honesty over warmth.
- **Rhetorical questions / playful subheadings as a technique** — becomes its own tell.
- **"Aim to beat AI detectors"** as a goal — our goal is readability and honesty; detector
  scores are not a quality measure.
- **Injected "natural digressions" and tangents** — padding. The no-fluff rules stay.

Adopted: simple words, short sentences, contractions, active voice, no hype vocabulary,
short mixed paragraphs, direct address, "write like you're explaining to a smart friend".

## Worked example (the owner's own pair, homepage hero)

**Too dense (v1-era, live before this change):**
> Most of what we build is unglamorous, and that is the point. A pipeline that checks scraped
> business records against Companies House, and stops to ask a person when it is genuinely
> unsure. A system that reads across news sources and scores what is worth trusting. A website
> that keeps itself current. Each one runs without anybody watching it, and every decision it
> made is written down where you can read it back afterwards.

**Plain v2 (the direction):**
> We build systems that take over repetitive work. Each one has a clear job. It knows when to
> ask a person for help, and it writes down every decision it makes. When it isn't sure, it
> stops and asks. Nothing happens in a black box.

**⚠️ Claims note:** the owner's illustrative rewrite contained "reads news from hundreds of
sources". The evidence base has 18 configured sources (13 news_search + 3 api_news + 1 scrape
+ 1 rss). "Hundreds" was NOT adopted. Plain voice changes the register, never the facts —
every rewrite still passes the claims gate.

## How to apply (mechanics)

Same safe path as all voice work: edit `content_data` fields, dual-branch escalation
pre-check (HANDOFF §9), `section_data_resolved` rerender, verify live. Never the writer
pipeline for rewrites. When the writer generates NEW content, it reads the `voice` spec —
which now encodes these rules (site_specs aspect `voice`, updated 2026-07-17, backup
`bak_voice_spec_leo_20260717`).
