# Where we are — Gemini for the content-producing agents

*Owner's running log. Append-only, newest at the bottom. Plain prose.*

---

## 2026-07-27 — why it was reversed, and why it's worth another go

You asked why we backed out of moving the content agents from Claude to Gemini,
and whether we can get it done this time. Here's the short version.

We tried it over two days, 23rd and 24th July. Two agents were involved and
they're configured in completely different places, which is part of why the story
is muddled. One is the **content creator**, which writes blog posts and social
posts — it's its own service with its settings in a config file. The other is the
**page content writer**, which writes the actual copy on every site we publish —
that one lives in the database. The blog-post one got switched over on the 23rd.
The page-copy one got switched on the afternoon of the 24th and switched back six
minutes later.

Two things went wrong, and both were real.

**First, the model we'd pinned didn't exist for us.** We'd asked for
`gemini-2.5-pro`. Google replies with a flat "no longer available to new users"
for our API key. It turns out Google closes off older model generations to keys
issued after a certain point, so this isn't about the model being retired — it's
about our key being too new. There's a floating name, `gemini-pro-latest`, that
does work, and that's what got used instead. This part was straightforward: a
clear error saying what was wrong.

**Second, and this is the one that killed it, the model barely wrote anything.**
On the short formats it produced nothing at all — literally zero words for a
tweet — and about 85 characters where we wanted a couple of paragraphs. The
thread investigating it found that the newer Gemini models "think" before they
answer, that the thinking couldn't be switched off, and concluded that the
thinking was eating the space where the writing should have been. The only Gemini
model that worked properly was a smaller, cheaper one that's a clear step down in
quality from what we'd picked pro for. You were asked whether to take that
downgrade, said no, and everything went back to Claude. On what was known at the
time, that was the right call.

**Here's what nobody spotted.** The diagnosis was three-quarters right and the
last quarter changes the answer. The thinking *was* eating the space — but it was
eating space *we failed to give it*. When we ask Gemini for a piece of writing we
send a number saying how much output we'll allow. With Claude, that number is
purely the writing. With Gemini, the same number covers the thinking *and* the
writing, and the thinking is spent first. Every one of those numbers in our system
was chosen years ago for Claude. So when we asked for a tweet in 100 tokens, we
were asking the model to fit its reasoning and its tweet into a space barely big
enough for the tweet. It thought, ran out of room, and said nothing. That's not
the model being incapable — it's us handing it a budget with no room in it.

Our own code never mentioned thinking anywhere. It passed the Claude-sized number
straight through, and it didn't even read back the figure that would have shown
what was happening — the count of thinking tokens was sitting in Google's
response the whole time and we threw it away. So the error message just said "ran
out of tokens", which looks exactly like a model that wanted to write more.

**The good news is that fixing it costs nothing.** That output number is a
ceiling, not a purchase — Google charges for what's actually produced. So we can
ask for a generous allowance for thinking on every call, and if the model only
thinks a little, we only pay for a little. I've changed our Gemini code to add
that headroom on top of whatever the caller asked for, and to record the thinking
count so this can never be invisible again. If it ever does run out of room, the
error now says so in plain terms and names the setting to turn up. I also fixed a
couple of related traps while I was in there: a dead model name now fails
immediately with the working alternative in the message rather than failing on
every call at run time, and there was a latent problem where the model's own
reasoning could have been pasted into published page copy, because Google returns
the thinking and the answer in the same list and we were concatenating everything
in it.

**Two things I want to be straight about.**

The first is that I have not proved this works. I've proved our code now sends
sensible numbers, and there are tests pinned to the exact settings that produced
nothing in production. But whether Gemini actually writes well for us is a
separate question, and it needs a real run. It's entirely possible the mechanism
explanation is correct and pro still turns out to be impractical — if it thinks
for tens of thousands of tokens on our prompts, that's expensive regardless. The
first thing to do is measure that, and I've written a script that does exactly
that in one go: which models our key can actually reach, how much thinking versus
writing we get at each of our real settings, and which of Google's thinking
controls this model accepts. Those answers were worked out by hand last time,
mid-incident, and then lost inside a commit message.

The second is that I can't finish it from here. The cluster login has expired in
this session — everything I try against it comes back "Unauthorized". That blocks
the remaining work: I can't build and roll the new code, can't verify it landed,
and can't run the probe, because the Gemini API key only exists inside the
cluster and there's no local copy. Everything up to that point is done and
committed.

**Worth knowing about the page-copy writer specifically:** it was never actually
tested on Gemini. It got switched over at 16:53 on the 24th and switched back at
16:59, and the page rebuild that was supposed to be the test was still sitting in
the backlog when the revert happened. So for the agent that writes our sites —
the one that actually matters commercially — we have no evidence about Gemini in
either direction. That's an open experiment, not a settled question.

**One thing I need from you.** Why did we want Gemini in the first place? The
repo doesn't record it, and it changes what I should configure. If it was about
cost, the cheap model already worked on the 24th and needs none of this headroom
machinery. If it was about quality or about not having everything on one
provider, then pro is the candidate and the headroom fix is what makes it
possible — but pro-with-a-proper-budget has genuinely never been run, so it's a
real experiment. Either way it's now a one-line config change to try it and a
one-line change to go back, so trying pro first costs very little.
