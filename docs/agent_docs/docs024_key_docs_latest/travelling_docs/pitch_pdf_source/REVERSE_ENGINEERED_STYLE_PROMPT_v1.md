# Reverse-engineered "de-AI-ify this copy" prompt

Built 2026-07-17 by comparing the v1/v2 pitch-deck copy (AI-competent, slightly stiff)
against a hand-edited rewrite the user judged genuinely more readable. Four parallel
agents each independently derived a candidate prompt from a different analytical
angle, then blind-tested it on two slides they were never shown the target text for.
I scored all four against the real (withheld) target text and merged the parts that
actually held up. Method and full transcripts: see `worked_example` below and
`~/pitch-build/` on the build host (reverse_engineer_bundle.md,
held_out_ground_truth.md — not copied into the repo, ephemeral).

## What the blind test found

- One agent, working from a "what specifically reads as AI-written" angle, landed a
  near word-for-word match on the target's most load-bearing sentence — "The
  important decision isn't the technology. It's where the truth lives." — without
  ever seeing it. That agent's rule set forms the backbone below.
- **All four agents missed the same thing**: the real target converts a flat list of
  parts-and-what-they-do into a markdown table with the part name bolded in the left
  column. All four instead wrote prose paragraphs or an invented numbered list. Rule
  8 below exists because of that shared miss.
- **All four agents also invented a "how a request flows through the system"
  numbered sequence** that doesn't exist in the target at all — the source was a
  flat list, the target kept it a flat list (then a table). Rule 7 exists to stop
  that overreach.
- One agent, working from a rhetorical/rhythm angle, inserted a question ("So where
  does the truth actually live?") the target never asks in that spot — a
  reasonable-sounding move that the evidence didn't actually support. Kept as a
  caution in rule 3.
- The real target breaks a multi-entry log into one fenced block *per entry*, with
  commentary sitting right after the entry it discusses — not one big block up
  front followed by all the commentary. None of the four candidates did this.

## The prompt

You are rewriting AI-sounding pitch or marketing copy into a plainer, more human
style. Preserve every fact, number, name and claim exactly — you are re-pacing and
re-phrasing, not summarising or embellishing.

1. **One idea, one sentence.** Find sentences chaining two or more clauses via
   commas, semicolons, or em dashes, and split them. A sentence carrying three ideas
   becomes three short sentences.
2. **Kill the em dash.** Any "X — Y" aside or "not X — Y" antithesis becomes two
   plain sentences: "Not X. Y."
3. **Ground abstract subjects.** If a sentence's subject is an abstraction ("the
   interesting decision", "the load-bearing part"), rewrite so a concrete actor does
   something, or state it as two flat declaratives: "The important decision isn't
   the technology. It's where the truth lives." Don't reach for a rhetorical
   question here by default — most sections stay declarative; only use a question
   where the passage already leans on that device elsewhere.
4. **Cut self-flagging commentary and hedges**: "crucially", "genuinely", "exactly",
   "deliberately", "which is the point", "load-bearing" (as metaphor), "seamless",
   "robust", "leverage", "delve", "furthermore", "moreover", "at its core", "in
   essence". State the underlying fact plainly instead.
5. **Contractions in ordinary sentences**: it's, isn't, doesn't, wasn't, that's,
   don't. Formal "cannot"/"can" may survive inside a deliberate contrastive pair
   (rule 11) for weight — that's the one exception, not the default.
6. **Headings.** Delete ALL-CAPS eyebrows and colon-joined slogan headlines ("A
   verification ladder: cheap checks confirm..."). Replace with plain sentence-case
   `#`/`##` headings that state a claim — sometimes two short parallel sentences ("A
   machine broke it. A machine fixed it.").
7. **Break apart numbered or bold-lead-in bullet lists of reasons, rules, or "things
   to know"** into flowing paragraphs of uneven length — one sentence here, three
   there — no numbering, no bold lead-in phrase. Reserve numbered lists strictly for
   a literal, sequential process the source actually describes step by step. Do not
   invent a new sequence or flow that isn't already implied by the source: if the
   source is a flat list of parts, keep it a flat list (see rule 8), not an invented
   pipeline.
8. **Use a markdown table whenever the source lists several items that each pair
   with the same one or two attributes** (a name plus what it does; a tier plus its
   cost). Bold the item name in the left column. Don't leave this as "X does A. Y
   does B." prose when a table fits.
9. **When a short list of concrete plain nouns needs introducing** (components,
   ingredients, causes), consider giving each its own one-line paragraph before any
   elaboration — "Postgres. Kafka. A handful of Go services." — a staccato beat that
   a table or fuller prose can follow.
10. **Land short paragraphs.** After 2–4 sentences of build-up, drop a standalone
    3–8 word sentence as its own paragraph to close the thought: "Then it moves on."
    "That was the point." Use this 2–5 times per section — a pause, not a tic.
11. **Build contrastive pairs** for wrong/right or before/after ideas: two short
    sentences, same grammar, negation flipped. "You cannot spot the bug by reading
    the file. You can spot it by opening the page."
12. **Pull literal values and quoted machine output out of running prose**: backtick
    a short verbatim token (`` `fixed: true` ``), bold the one number that's the
    payoff of a paragraph (**506 pixels**).
13. **If the source describes a multi-entry log or several fenced blocks**, don't
    dump the whole thing up front and comment afterward — break it into its natural
    entries (one fenced block per entry) and place the relevant commentary
    immediately after the entry it's about.
14. **Keep code or log content close to verbatim inside its fence** — it's
    presented as a genuine record, not something to paraphrase into different
    words.
15. **Voice**: "we"/"our" for the company's actions, never first-person singular.
    Active, not passive — "the tool rebuilt itself," not "was rebuilt."
16. **Don't over-polish.** Leave one slightly blunt or plain phrase sitting rather
    than smoothing every sentence to the same register — a small rough edge reads
    human.

Don't invent facts, numbers, structure, or sequences absent from the source. No
exclamation points, no hype adjectives.

## Worked example (from the blind test)

Source (v1, "before"):
> Nothing here is exotic. The interesting decision is where the truth lives: in
> Postgres, keyed to the tool, written by the agents, read by anything that wants to
> know what "working" means.

Real target (withheld from all four agents until scoring):
> There's no unusual infrastructure here. […] The important decision isn't the
> technology. It's where the truth lives. The PLAN and the notes live in the
> database, attached to the tool itself. Every agent reads them. Every agent writes
> back to them. Everything else is temporary.

The agent working from the "what reads as AI-written" angle produced, blind:
> The interesting decision isn't the technology. It's where the truth lives. That's
> Postgres. Keyed to the tool, not to a file path. Written by the agents. Read by
> anything that wants to know what "working" means.

That's rule 3 in action, derived without ever seeing the sentence it was scored
against.

---

**Status: SUPERSEDED.** The "worked example" sentence held up above as the strongest
evidence was itself pointed out — by the user, then confirmed on further review — to
be laboured: an unearned "isn't/it's" contrast, and "truth" overclaiming for what a
database holds. See `REVERSE_ENGINEERED_STYLE_PROMPT_v2.md` for the correction, and
`_v3.md` for a further round addressing the same instinct showing up in a different
grammatical shape ("Nothing here is exotic. One choice is...") plus the word "exotic"
itself. `REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` is current as of 2026-07-17.
