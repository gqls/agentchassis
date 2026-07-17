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


Please let's work a bit more on the prompt. All versions so far including mine have a sort of competitive or negative framing e.g. "The interesting decision isn't the technology. It's where the truth lives.
That's
> Postgres. Keyed to the tool, not to a file path." - there's a whole load of "not this but that..." in that phrasing. People don't say that unless it's a difficult point to express and they _have_ to spell it
out. Also we'd use something like ground-truth not truth - truth is a word that is very strong, too strong for this context. This following copy is also laboured: "Every agent reads them. Every agent writes
> back to them. Everything else is temporary." readable copy wouldn't need to stress every agent this and that unless it's a point that is not obvious the first time or it needs emphasising, and that sort of
emphasising shouldn't happen too often in the whole document - if it has to emphasise everything all the time it is not explaining things well enough in the first place.

================================
----------------------------------

Updated the prompt document directly at docs/agent_docs/docs024_key_docs_latest/travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT.md. Two changes:

Rule 3 now restricts "not X, it's Y" to genuine corrections — cases where a reader would plausibly have assumed X first. Most of the time nobody assumed X, so spelling it out just delays the point. If a section
leans on "isn't/wasn't/not" more than once or twice, that's the tell.

New rule 4 on word-weight: don't reach for a word heavier than the fact. "Truth" is almost never right for what a database holds — "the record," "the current version," or whatever term the source already uses. Same
for "crucial," "vital," "fundamental" applied to ordinary engineering choices.

Rule 13 (new) targets the "Every agent reads them. Every agent writes back to them." pattern specifically — repeating a sentence template for cadence rather than because the repetition itself is informative.
Usually two such sentences are one point said twice and should just combine.

I also added a closing note tying rules 3, 11, 12, and 13 together: they're all devices for making a sentence land harder, and all four are only earned by the one or two genuinely non-obvious points per section —
not a rhythm to apply throughout.

I kept the worked example but made it honest rather than triumphant: the blind-matched sentence I held up last time as proof the reverse-engineering worked is also an instance of the flaw you just named, so I
annotated it and rewrote it against the corrected rules — then caught myself using an em dash and a semicolon in that very rewrite on review, which I left in as a footnote because it's a fair demonstration that the
rules are easy to state and easy to slip on.
