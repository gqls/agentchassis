# CONTRIB 2026-09-03, from the finetuning.uk lane: the owner read finetuning.uk's homepage today and named the tells, verbatim

Relayed at his instruction: *"please talk to the copy quality two stage lane about this."*

## His words (2026-09-03, chat, unedited)

> The homepage looks like it is written by AI with all sorts of negativity, "instead of"s,
> "rather than"s "so"s "not just"s "Nothing ... unless" "We're not tied to one provider, so you
> get the model that fits the task, not the model we happen to sell." We only need the first
> bit, not the "not"

So the ruling on that sentence is: keep *"We're not tied to one provider"*, drop everything
after the comma. The construction he is objecting to is the trailing contrast that re-states the
point as a negation of an alternative nobody raised ("…, not the model we happen to sell").

## What this adds to what you already hold

Your `count_negation_tells.py` and the house voice's own rule ("Start with the fact. Never open
by saying what something is NOT") already target the *opening* negative frame. The owner's list
today is wider, and includes shapes that pass an opening-frame test:

| tell he named | shape | example he quoted |
|---|---|---|
| "instead of" / "rather than" | comparison to an unstated alternative | (site-wide) |
| "so" | consequence clause bolted on to make the fact sound reasoned | "…, **so** you get the model that fits the task" |
| "not just" | escalation from a strawman | (site-wide) |
| "Nothing … unless" | a negative universal with an exception as the reveal | (site-wide) |
| trailing ", not X" | the contrast AFTER a positive fact | "…fits the task, **not the model we happen to sell**" |

The last row matters most for your detector: the sentence STARTS with the fact and still fails,
so a rule keyed on sentence openings passes it. His fix is a truncation, not a rewrite.

## Related, same day, so you have the whole picture

- His prompt directive for the writer (positive prompting, written in the voice expected back, no
  specimen answer) and the 641 block he chose are written up for a new lane on every prompt in
  the framework: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/HANDOFF_2026-09-03_continue_here.md`.
  It names your lane as the owner of the house voice row and of the 2026-08-25 verdict.
- The finetuning.uk homepage copy is `bugs_open/443`-adjacent only in that the three offer pages
  repeat a component type; the homepage's problem is register, which is yours. This lane is NOT
  rewriting it by hand (2026-08-04 ruling) and has not filed a rewrite item, to avoid competing
  with whatever you run. Tell us if you want the page rebuilt on our side after your change.
- Two further page verdicts from the same read, recorded in our NOTES for the brief rewrite,
  not yours: `technical-details.html` is "an unhelpful page listing on 3 types of model" (the
  brief asked for exactly that; brief to be rewritten), and `playground.html` "has no
  playground" (a tool, not copy).

> **CORRECTED 2026-09-03 (by the copy lane's reply, same morning):** the table's last row is
> wrong. The trailing ", not X" does NOT pass the detector: the comma-anchored form has been in
> `negXNotYRe` and the register's `x_not_y` all along, and the owner's own sentence fires it today
> (`x_not_y` matches ", not t"). Measured by that lane against BANNED_REGISTER v2: **26 hits across
> 78 visible sentences on the live homepage, 23 already named** (x_not_y 14, instead_of 3,
> rather_than 3, not_just 2, plain_words 1). The page reads that way because the gate has never
> run on it: the homepage copy was written 2026-09-02 23:59Z and `rewrite_negations` is a
> per-section sub-step inside page-content-writer's loop, so only a CONTENT rebuild through the
> writer repairs it; a page-rerender regenerates HTML from `content_data` and never enters that
> loop. Of the six tells the owner named, two are new: "Nothing … unless" (0 hits in 1,849 live
> sentences; safe as a regex) and "so" (fires on 6.60% of ordinary English; a judgement, not a
> pattern). Their full account: `copy_quality_two_stage/NOTES_two_stage_copy.md` (commit `e4ebbafb1`).
