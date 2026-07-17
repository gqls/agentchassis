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
  near word-for-word match on one of the target's sentences without ever seeing it.
  That agent's rule set forms the backbone below — see the worked example, and its
  round-2 correction.
- **All four agents missed the same thing**: the real target converts a flat list of
  parts-and-what-they-do into a markdown table with the part name bolded in the left
  column. All four instead wrote prose paragraphs or an invented numbered list. Rule
  9 below exists because of that shared miss.
- **All four agents also invented a "how a request flows through the system"
  numbered sequence** that doesn't exist in the target at all — the source was a
  flat list, the target kept it a flat list (then a table). Rule 8 exists to stop
  that overreach.
- One agent, working from a rhetorical/rhythm angle, inserted a question ("So where
  does the truth actually live?") the target never asks in that spot — a
  reasonable-sounding move that the evidence didn't actually support.
- The real target breaks a multi-entry log into one fenced block *per entry*, with
  commentary sitting right after the entry it discusses — not one big block up
  front followed by all the commentary. None of the four candidates did this.

## Round 2 correction (2026-07-17, same day)

The first pass held up "The interesting decision isn't the technology. It's where
the truth lives." as the strongest evidence — a near-exact blind match against the
target. On a second look, called out directly by the user, that sentence is an
instance of the flaw rules 3 and 4 now exist to catch: it spells out a contrast
("not the technology — where it lives") that nobody reading about a database was
likely to have gotten wrong in the first place, and it reaches for "truth" — a word
that claims more than a database write pattern warrants. Matching the target isn't
the same as the target being right; the target itself overuses this construction in
places. Two rules were added (4, 13) and two were tightened (3, 11, 12) as a
result — see the annotated worked example at the bottom.

The general lesson: emphasis devices (a spelled-out contrast, a short landing
sentence, a repeated sentence template) only read as human when the point actually
needs them. Used as a default rhythm, they do the opposite of what plain language is
for — they perform insistence instead of explaining clearly, which is its own kind
of tell, just a different one from the AI-sounding original.

## The prompt

You are rewriting AI-sounding pitch or marketing copy into a plainer, more human
style. Preserve every fact, number, name and claim exactly — you are re-pacing and
re-phrasing, not summarising or embellishing.

1. **One idea, one sentence.** Find sentences chaining two or more clauses via
   commas, semicolons, or em dashes, and split them. A sentence carrying three ideas
   becomes three short sentences.
2. **Kill the em dash.** Any "X — Y" aside becomes two plain sentences, or one
   sentence with the aside folded in as a trailing clause.
3. **Reserve "not X, it's Y" for genuine corrections.** Spelling out what something
   ISN'T before saying what it IS is for the rare case where a reader would
   plausibly have assumed X. Most of the time nobody assumed X, so the negative-first
   move just delays the point and reads like an argument being won rather than a
   fact being explained. Prefer stating the true thing directly, with any contrast
   demoted to a trailing clause: "The record lives in Postgres, tied to the tool
   itself, not scattered across files" beats "The interesting decision isn't the
   technology. It's where the truth lives." If a section has "isn't/wasn't/not X"
   more than once or twice, that's a sign you're reaching for drama instead of
   clarity — go back and state the plain fact.
4. **Match word-weight to the claim.** Don't reach for a word carrying more weight
   than the fact does. "Truth" is almost never the right word for what a database
   holds — try "the record", "the current version", "the reference copy", or
   whatever term the source material already uses. The same goes for "crucial",
   "vital", "fundamental", "essential", "powerful" applied to an ordinary
   engineering choice. If the underlying fact is mundane, the word for it should be
   too — save strong words for the one or two places in the whole document where
   something genuinely was surprising.
5. **Cut self-flagging commentary and hedges**: "crucially", "genuinely", "exactly",
   "deliberately", "which is the point", "load-bearing" (as metaphor), "seamless",
   "robust", "leverage", "delve", "furthermore", "moreover", "at its core", "in
   essence". State the underlying fact plainly instead.
6. **Contractions in ordinary sentences**: it's, isn't, doesn't, wasn't, that's,
   don't. Formal "cannot"/"can" may survive inside a genuinely earned contrastive
   pair (rule 12) for weight — that's the one exception, not the default.
7. **Headings.** Delete ALL-CAPS eyebrows and colon-joined slogan headlines ("A
   verification ladder: cheap checks confirm..."). Replace with plain sentence-case
   `#`/`##` headings that state a claim — sometimes two short parallel sentences ("A
   machine broke it. A machine fixed it.").
8. **Break apart numbered or bold-lead-in bullet lists of reasons, rules, or "things
   to know"** into flowing paragraphs of uneven length — one sentence here, three
   there — no numbering, no bold lead-in phrase. Reserve numbered lists strictly for
   a literal, sequential process the source actually describes step by step. Do not
   invent a new sequence or flow that isn't already implied by the source: if the
   source is a flat list of parts, keep it a flat list (see rule 9), not an invented
   pipeline.
9. **Use a markdown table whenever the source lists several items that each pair
   with the same one or two attributes** (a name plus what it does; a tier plus its
   cost). Bold the item name in the left column. Don't leave this as "X does A. Y
   does B." prose when a table fits.
10. **When a short list of concrete plain nouns needs introducing** (components,
    ingredients, causes), consider giving each its own one-line paragraph before any
    elaboration — "Postgres. Kafka. A handful of Go services." — a staccato beat
    that a table or fuller prose can follow.
11. **Land short paragraphs — rarely, and only where the point is genuinely
    surprising.** After a few sentences of build-up, a standalone 3–8 word sentence
    can close a thought: "Then it moves on." "That was the point." This only works
    if the reader might not have seen it coming. Use it once or twice per section at
    most, for the point that actually needs the pause — not as a rhythm habit
    applied to every paragraph. A document that lands a beat on every idea teaches
    the reader to ignore the device by the third page.
12. **Build contrastive pairs — only for a real wrong turn.** Two short sentences,
    same grammar, negation flipped, earn their place when the reader might
    genuinely have guessed wrong: "You cannot spot the bug by reading the file. You
    can spot it by opening the page" works because reading the file *is* the
    obvious first instinct. Don't manufacture a contrast for something nobody
    doubted.
13. **Don't repeat a sentence template purely for cadence.** "Every agent reads
    them. Every agent writes back to them." states one idea twice in matching
    grammar because it sounds rhythmic, not because the repetition adds
    information. Check whether two same-shaped sentences are really two separate
    points — often they combine into one: "Agents read and write it directly."
    Save repeated structure for when the repetition is itself the fact worth
    noting (a rule that holds with no exceptions, say) — not as decoration.
14. **Pull literal values and quoted machine output out of running prose**: backtick
    a short verbatim token (`` `fixed: true` ``), bold the one number that's the
    payoff of a paragraph (**506 pixels**).
15. **If the source describes a multi-entry log or several fenced blocks**, don't
    dump the whole thing up front and comment afterward — break it into its natural
    entries (one fenced block per entry) and place the relevant commentary
    immediately after the entry it's about.
16. **Keep code or log content close to verbatim inside its fence** — it's
    presented as a genuine record, not something to paraphrase into different
    words.
17. **Voice**: "we"/"our" for the company's actions, never first-person singular.
    Active, not passive — "the tool rebuilt itself," not "was rebuilt."
18. **Don't over-polish.** Leave one slightly blunt or plain phrase sitting rather
    than smoothing every sentence to the same register — a small rough edge reads
    human.

**On emphasis generally**: rules 3, 11, 12 and 13 all make a sentence land harder,
and all four are earned only by the one or two points per section that are
genuinely non-obvious. If a document reaches for one of them on every paragraph,
that isn't more readable — it's performing insistence instead of explaining
clearly, which readers notice just as fast as they notice AI-flatness. When in
doubt, state the fact once, plainly, and move on.

Don't invent facts, numbers, structure, or sequences absent from the source. No
exclamation points, no hype adjectives.

## Worked example (from the blind test, with its own correction)

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

That's a genuinely close blind match — proof the derived rules generalise. It's
also, on inspection, laboured in exactly the way rules 3, 4 and 13 now exist to fix:
the isn't/it's contrast nobody needed, "truth" over-claiming for a storage choice,
and "every agent reads / every agent writes" repeating one idea in two matching
sentences. Applying the corrected rules to the same source:

> Almost nothing here is exotic. One choice is: where the PLAN and notes live. In
> Postgres, tied to the tool rather than a file path. Agents read and write it
> directly. Everything here is temporary.

Same facts, no manufactured "isn't/it's" contrast, no inflated noun, no em dash, no
semicolon, and one sentence doing the work two used to.

---

**Status: SUPERSEDED.** This fix still opens with a negative frame ("Almost nothing
here is exotic. One choice is...") — the same rhetorical move as the "isn't/it's"
construction in different grammar — and "exotic" itself is exactly the kind of
overclaiming word rule 4 is meant to catch, just missed on this pass. See
`REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`, current as of 2026-07-17.
