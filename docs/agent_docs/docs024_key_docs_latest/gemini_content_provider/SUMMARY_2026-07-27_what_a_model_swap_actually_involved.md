# What changing model actually involved

Written 2026-07-27, after moving the two content-producing agents from Claude to
Gemini. The first attempt at this was made on 23 July and reversed on 24 July. This
is the account of why, and of everything that turned out to sit underneath a change
that looks like editing one line of config.

## The line of config really is one line

Both agents pick their model from a settings block. Provider, model name, which
environment variable holds the API key. Changing it takes a few seconds.

```yaml
ai_service:
  provider: "gemini"
  model: "gemini-pro-latest"
  api_key_env_var: "GEMINI_API_KEY"
```

We made that change on 23 July. By the afternoon of the 24th we'd changed it back,
and we'd concluded the model couldn't write.

## The number we sent meant something different to the new model

We give every model a limit on how much it may produce. With Claude, that number is
the answer. All of it. With Gemini, the same number covers the model's own reasoning
*and* the answer, and the reasoning is spent first.

Every one of those numbers in our system was chosen years ago, against Claude. Our
code passed them straight through.

So when we asked for a tweet in 100 tokens, we were asking the model to fit its
thinking and its tweet into a space barely big enough for the tweet. Here's what
came back at each of our real settings:

| limit we sent | reasoning tokens spent | tokens of tweet left | result |
|---|---|---|---|
| 100 | 92 | 4 | **zero characters of usable text** |
| 500 | 477 | 19 | about 85 characters, cut off |
| 1,200 | 1,145 | 38 | finished normally |
| 6,000 | 786 | 44 | finished normally |

Read the first row again. The model thought until it ran out of room, then had
nothing left to speak with. That's arithmetic. It isn't a limit on what the model
can write.

## We diagnosed the model, and the fault was ours

The thread investigating this in July got three quarters of the way there. It found
that the newer Gemini models reason before answering, that the reasoning couldn't be
switched off, and that the reasoning was consuming the space the writing needed. All
true.

It concluded the model was therefore unusable at our budgets, and that our only
option was a smaller, cheaper model that's a clear step down in quality. That option
was put to the owner, declined, and the whole switch was abandoned.

The missing quarter is that the reasoning was consuming space **we never gave it**.
Our code set the limit to the caller's number and never mentioned reasoning
anywhere in the file. Fixing it costs almost nothing, because that limit is a
ceiling rather than a purchase. Google bills what's produced. We can ask for
generous headroom and pay only for what gets used.

The failure was invisible for a second reason. Google's response carries a count of
reasoning tokens. Our code read the other counts and threw that one away. So the
tokens doing the damage never appeared anywhere, and the error we logged said only
that the model had run out of room. That reads exactly like a model that wanted to
write more.

Three days later, the same model produced a **264-character** tweet at the same
100-token setting.

## The model catalogue is a property of your key, not of the model

We'd pinned `gemini-2.5-pro`. Google answers that with a 404 saying the model is no
longer available to new users. Our key was issued after Google closed that
generation off. The model exists. Other people can call it. We can't.

The catalogue makes this worse. Asking Google what our key can use returns 42 models
including `gemini-2.5-pro`, and calling it fails. A model appearing in the list
isn't evidence you can reach it. The only proof is a call.

We now refuse the known-dead names at startup with the working alternative in the
message, instead of failing on every request in the middle of a job. We don't
substitute a working model silently. Serving a different model than the one someone
asked for, at a different price, behind unchanged config, is a bug we've filed
before under other names.

## The reasoning can be reduced, not stopped, and not capped

Google offers two controls over how much a model thinks. We assumed, from one
rejected value, that they belonged to different model generations and were mutually
incompatible. We wrote that into five places.

We were wrong, and the probe we'd written for the purpose caught it in about thirty
seconds. Both controls work. Only the value zero is refused, and the API says why:
this model only works in thinking mode.

The useful finding is the one that replaced it. Neither control actually caps the
reasoning. Ask for 128 tokens of it and the model spends 483. Ask for 32,768 and it
spends 783. They're a cost lever, not a limit, so neither can substitute for
provisioning the headroom properly.

A refusal tells you about the value you sent. Only its neighbours tell you about the
parameter.

## The prompt was tuned against the old model

Our page writer carries a 12,570-character prompt, including a house style block
built over three rounds of the owner critiquing its output. That block was written
while Claude was answering it. Nobody knew whether it was describing a style or
steering one particular model.

It transferred. On the first Gemini run the copy scored zero em dashes, zero filler
words, opened with the fact rather than a negative frame, and left one blunt
sentence standing where the style asks for a rough edge. It missed one rule:
contractions. For comparison, Claude on the same style prompt had brought em dashes
down from 19 to 14 rather than to zero.

That's one sample against one sample, on different subject matter, so it doesn't
show one model writing better than the other. It does show the style block survives
a change of model family, which nobody had tested.

## The output has a shape, and shapes break quietly

The page writer asks for JSON. A model that wrapped its answer in a code fence, or
ran out of room mid-string, would hand back something unparseable, and a prose test
would never reveal it. We checked that specifically. Valid JSON, no fence, every
required field.

There's a related trap we fixed before it bit. Gemini returns its reasoning and its
answer in the same list of parts, distinguished by a flag. Our code joined every
part together. Had reasoning ever arrived that way, it would have been pasted into a
published page, and nothing above that layer reads the text closely enough to
notice.

## The swap damaged our own instruments

This is the part we didn't expect, and it's the reason the work took a day rather
than an hour.

Our cost estimates are keyed by model name. Both Gemini entries were the dead pinned
names, so they could never match, and every Gemini call was quietly priced at
Claude's rate. Nothing errored. The number was just wrong.

Then we did it to ourselves. We record what we sent in a column called `max_tokens`,
and we filled it with the new inflated ceiling. For Claude that column means "budget
for the answer". For Gemini it now meant "budget for the answer plus the reasoning".
One column, two meanings, depending on the provider. That's the same defect we'd
spent the day fixing, reproduced one level up by the fix for it.

It also broke a rule we rely on. If output tokens equal the limit, the answer was
cut off. That comparison stops working when the limit includes reasoning the answer
never used.

We caught it with a day to spare, because that table had **zero** Gemini rows at the
time. The blog agent doesn't write to it. The first row would have appeared the
moment the page writer switched over.

And we'd claimed, in a bug file, a council submission and three commit messages,
that the change made reasoning costs "visible to logging". Half of that was true.
The counts are visible in the error message. They're written to no column and read
by no query, so nothing can report on them. A ten-seat review approved that
sentence, and one reviewer discussed those exact fields. Writing a field and being
able to read it look identical in a diff.

## What we'd tell the next person doing this

A model swap is a change of contract, not a change of supplier. The vendor's name
changes in one place. The assumptions change everywhere.

The things that carried hidden assumptions about the old model, in the order they
bit us: the token budgets, the model names, the cost table, the log columns, the
truncation rule, the output format, and the style prompt. Six of those seven look
like plumbing and none of them appears in the config you edit.

The single most useful habit was measuring against the live API before believing
anything, including our own conclusions from three days earlier. The probe script we
wrote for it now answers, in one run, which models the key can reach, how much
reasoning each setting buys, and which controls the model accepts. Those answers had
previously been worked out by hand during an incident and then lost inside a commit
message.

The single most expensive habit would have been trusting the first verdict. "The
model can't write" was measured, written down confidently, and wrong. It closed the
question for three days.

## Where we are, and what's next

Both agents run on `gemini-pro-latest` in production. The blog and social agent is
proven on real jobs, including a 1,292-word article. The page writer is switched
over and its prompt and output format are verified against the live model.

One thing is left. We need to build a real page end to end and read the copy, which
is the only test of the question the owner actually asked. That's running now
against a darts page that has never been published, so a poor result costs nothing.

Two follow-ups are filed. Our log column needs its single meaning restored across
providers, which is written and waiting for the next deployment. And the reasoning
counts need somewhere to live, because reasoning is billed as output and we still
can't report on what it costs at the scale of a whole site.
