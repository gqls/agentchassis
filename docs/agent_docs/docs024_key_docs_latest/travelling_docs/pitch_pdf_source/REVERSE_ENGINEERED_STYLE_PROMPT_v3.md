# Reverse-engineered "de-AI-ify this copy" prompt

Built 2026-07-17 by comparing the v1/v2 pitch-deck copy (AI-competent, slightly stiff)
against a hand-edited rewrite the user judged genuinely more readable, then refined
across two further rounds of the user critiquing the prompt's own output. Earlier
versions, kept for comparison: `REVERSE_ENGINEERED_STYLE_PROMPT_v1.md` (first
composite, from a 4-agent blind test), `_v2.md` (round-2 fix for unearned "isn't/
it's" contrasts and word-weight). This file, v3, is current.

## What the blind test found (unchanged from v1/v2 — the discovery phase)

- One agent, working from a "what specifically reads as AI-written" angle, landed a
  near word-for-word match on one of the target's sentences without ever seeing it.
  That agent's rule set forms the backbone below.
- **All four agents missed the same thing**: the real target converts a flat list of
  parts-and-what-they-do into a markdown table with the part name bolded in the left
  column. All four instead wrote prose paragraphs or an invented numbered list. Rule
  9 below exists because of that shared miss.
- **All four agents also invented a "how a request flows through the system"
  numbered sequence** that doesn't exist in the target at all — the source was a
  flat list, the target kept it a flat list (then a table). Rule 8 exists to stop
  that overreach.
- The real target breaks a multi-entry log into one fenced block *per entry*, with
  commentary sitting right after the entry it discusses — not one big block up
  front followed by all the commentary. None of the four candidates did this.

## Round 2 correction

The first pass held up "The interesting decision isn't the technology. It's where
the truth lives." as its best evidence. On review, that sentence spells out a
contrast nobody needed spelled out, and reaches for "truth" — a word claiming more
than a database write pattern warrants. Added word-weight calibration (rule 4) and
restricted the "not X, it's Y" construction to genuine corrections (rule 3, at the
time).

## Round 3 correction (2026-07-17, same day)

The round-2 fix was: "Almost nothing here is exotic. One choice is: where the PLAN
and notes live. In Postgres, tied to the tool rather than a file path. Agents read
and write it directly. Everything here is temporary." The user caught two things
rounds 1 and 2 both missed:

1. **"Exotic" is the wrong word — too strong, and strong in the self-congratulatory
   direction.** "Nothing here is exotic" is still asking the reader to be impressed,
   just via humility instead of grandeur. It's a recognisable trope ("nothing fancy
   here, folks") and reads as performed rather than plain. Rule 4 previously only
   caught words that overclaim *importance* ("truth", "crucial"); it didn't catch
   words that overclaim *ordinariness*. Fixed below.
2. **The sentence was still negatively framed, just in different grammar.** "Nothing
   here is X. One choice is Y" is the same move as "X isn't Y, it's Z" — state a
   negative first, then reveal the real fact as if it were a twist. Round 2's rule 3
   only checked for the literal words "isn't/it's"; it didn't catch the same
   instinct wearing a different sentence shape. The user's own suggested fix says it
   plainly: a person would start with the fact — "The plan and notes are stored in
   the database and are always..." — not with what the stack *isn't*. Rule 3 is now
   written against the underlying move, not the specific grammar, so it should catch
   both surface forms and whatever the next one turns out to be.

## The prompt

You are rewriting AI-sounding pitch or marketing copy into a plainer, more human
style. Preserve every fact, number, name and claim exactly — you are re-pacing and
re-phrasing, not summarising or embellishing.

1. **One idea, one sentence.** Find sentences chaining two or more clauses via
   commas, semicolons, or em dashes, and split them. A sentence carrying three ideas
   becomes three short sentences.
2. **Kill the em dash.** Any "X — Y" aside becomes two plain sentences, or one
   sentence with the aside folded in as a trailing clause.
3. **Don't open a fact with a negative frame.** "The interesting decision isn't the
   technology. It's where the truth lives." and "Nothing here is exotic. One choice
   is..." are the same move in different grammar: state what ISN'T true, or what
   ISN'T the point, before revealing the actual fact — a kind of manufactured twist.
   A person explaining this out loud starts with the fact itself: "The PLAN and
   notes are stored in Postgres, attached to the tool rather than a file path."
   Check every paragraph-opening sentence for this shape, not just ones using the
   literal words "isn't" or "not" — a "Nothing is X. Y is the exception" setup is
   the same problem wearing different grammar. If there's a genuine contrast worth
   keeping, fold it in as a trailing clause after the fact is stated, not as the
   sentence's opening move. And if a section has more than one or two sentences
   built this way — in *any* grammatical form — that's the sign to go back and
   state the plain facts in the order a person would actually say them.
4. **Match word-weight to the claim — including words that dramatize being
   ordinary.** Don't reach for a word carrying more weight than the fact does, in
   either direction. "Truth" is almost never the right word for what a database
   holds — try "the record", "the current version", or whatever term the source
   already uses. And "exotic", "unusual", "remarkable", "surprisingly simple"
   applied to an ordinary tech stack are the same overclaiming move pointed the
   other way: "nothing exotic here" still asks the reader to be impressed, just by
   humility instead of grandeur. Usually the plainest move is to skip commentary on
   ordinariness entirely — just list what's used ("Postgres. Kafka. A browser.")
   without characterising it first. Save any strong word, either direction, for the
   one or two places in the whole document where something genuinely was
   surprising.
5. **Cut self-flagging commentary and hedges**: "crucially", "genuinely", "exactly",
   "deliberately", "which is the point", "the important/interesting/real decision
   is", "what matters here is", "load-bearing" (as metaphor), "seamless", "robust",
   "leverage", "delve", "furthermore", "moreover", "at its core", "in essence".
   Don't tell the reader a fact is important before giving it to them — state it and
   let its placement do that work.
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
    that a table or fuller prose can follow. This is a plain enumeration, not a
    claim — don't attach a characterisation ("nothing exotic") to it; let the list
    speak for itself.
11. **Land short paragraphs — rarely, and only where the point is genuinely
    surprising.** After a few sentences of build-up, a standalone 3–8 word sentence
    can close a thought: "Then it moves on." "That was the point." This only works
    if the reader might not have seen it coming. Use it once or twice per section at
    most — not as a rhythm habit applied to every paragraph. A document that lands a
    beat on every idea teaches the reader to ignore the device by the third page.
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
doubt, state the fact once, plainly, in the order a person would actually say it,
and move on.

Don't invent facts, numbers, structure, or sequences absent from the source. No
exclamation points, no hype adjectives — in either the grand or the falsely-humble
direction.

## Worked example — three rounds on the same sentence

Source (v1, "before"):
> Nothing here is exotic. The interesting decision is where the truth lives: in
> Postgres, keyed to the tool, written by the agents, read by anything that wants to
> know what "working" means.

Real target (from the user's hand-edited rewrite):
> There's no unusual infrastructure here. […] The important decision isn't the
> technology. It's where the truth lives. The PLAN and the notes live in the
> database, attached to the tool itself. Every agent reads them. Every agent writes
> back to them. Everything else is temporary.

**Round 1** (blind agent output, scored a near-exact match against the target
above — the match was real, but matching a flawed target just reproduces the flaw):
> The interesting decision isn't the technology. It's where the truth lives. That's
> Postgres. Keyed to the tool, not to a file path. Written by the agents. Read by
> anything that wants to know what "working" means.

**Round 2** (fixed the "isn't/it's" contrast and "truth" — but kept a negative frame
and added a new overclaiming word):
> Almost nothing here is exotic. One choice is: where the PLAN and notes live. In
> Postgres, tied to the tool rather than a file path. Agents read and write it
> directly. Everything here is temporary.

**Round 3** (states the fact first, no negative frame, no "exotic", no "truth"):
> The PLAN and notes are stored in Postgres, attached to the tool rather than a
> file path. Agents read and write them there directly. Everything else in the
> system is temporary.

Same facts as every prior round. No setup, no reveal, no word doing more work than
the fact underneath it. This is close to how the user phrased the fix directly:
start with what's actually stored, not with what the stack isn't.
