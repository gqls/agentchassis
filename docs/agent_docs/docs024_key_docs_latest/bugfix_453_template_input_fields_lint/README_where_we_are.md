# Where we are — the prompt-template lint

---

## 2026-09-03 — what this is, in plain terms

Our agents write their prompts from templates. A template is a bit of text with holes in it,
like `Company: {{.company_name}}`, and the hole gets filled in from data the step was told to
collect. The step says what to collect in a list called `input_fields`.

The problem is what happens when the template asks for something that is not on that list.
Nothing breaks. The hole just fills with nothing, and the prompt goes off to the model looking
perfectly reasonable — one paragraph shorter than it should be, or one list shorter, with no
error, no warning and nothing in any report. Whoever wrote the template thinks it is working.

This has now bitten us four separate times, and every time it was caught by accident, because
somebody happened to write a test that rendered the thing and looked at the output. That is
not a system. So `bugs_open/453` asked for a checker that goes and looks at every template we
have and asks, mechanically, "can this step actually supply what this template is asking for?"

That is what I have built.

## What it found on the first run

139 live templates across 202 agents. It found **one** case where a template asks for
something its step can never supply, and — this is the part I would not have guessed — it is
on the **same step the bug was filed about**. Somebody fixed that step by hand last week for a
different missing field. This one was sitting right next to it and nobody saw it.

The template on our main page-writing agent has a whole "Research Findings" section, with a
list of sources and citation numbers. It asks for a thing called `research_result`. The step
does not collect `research_result`. So that entire section has always rendered as nothing.

## But I want to be careful about how bad that is, because I got it wrong first

My first instinct was that this was expensive: there is a research step wired up right before
it, it calls out to a research agent with a 90-second budget, and its output is named
`research_result` — so it looked like we were paying for research and throwing it away.

Then I went to measure it, and that is not what is happening. **The research agent has never
run — not once, ever.** The switch that would turn it on (`needs_research` on a component) is
never set to true, anywhere. So of 391 page-writing runs in the last month, zero reached the
research step.

> **⚠ THE SENTENCE ABOVE IS WRONG — see the dated correction at the foot of this file
> (2026-09-03).** Left exactly as written rather than edited away: this log is append-only, and a
> claim that turned out false is evidence about how we get things wrong. Short version: "never,
> ever" came off a table that keeps two days. It last ran on 18 January.

So the honest position is: **nothing is broken today**. Two things are dead, and each one
hides the other. The research never runs, so nobody notices the template block is dead; the
template block is dead, so if the research ever did run its output would vanish silently.

The reason that is worth fixing rather than shrugging at: turning the research on is a
one-word config change. The first person who does it will pay for research calls and get
nothing back, and the prompt will still be telling the model to "include source citations" for
research it never received. And no amount of running the system can reveal that in advance —
which is exactly the kind of thing a checker is for and a test is not.

## The two other things it reports

It also flags 16 cases it is honest about not being able to decide (they depend on data that
only exists when a job is actually running), and 19 cases of the opposite problem: a step
collecting something no template ever reads, which costs us a search through the whole data
tree on every single run for nothing.

I have deliberately made only the first kind fail the check. The other two are reported and
never block anything, because a checker that cries wolf gets switched off.

## Something I need to flag that is not mine

While running the tests I found that `go test` on that whole package already fails at our
committed code, for an unrelated reason — a finding-code registry entry from the
`bugs_open/450` work does not fill in a field its own rule requires. I confirmed it fails on a
clean copy of the committed code with none of my changes in it, so it is not something I
introduced. I have not touched it: it is another lane's entry and their guard is doing its job.
Someone on that lane needs an hour.

## What I have not done

I have not changed the page-writing agent. Making its research block work would be a real
behaviour change to the busiest thing we run, and on today's evidence it is not urgent — the
research it would carry does not exist yet. That is a decision to take deliberately, not a
side effect of building a checker, so I have written it up and left it.


---

## 2026-09-03 (same day) — I have to correct something I told you above

I wrote that the research agent had "never run, not once, ever". **That was wrong**, and I
should not have said it to you before checking it properly.

The table I counted from only keeps two days of history. My query had no date filter on it, so
"no rows" felt like "never" when in fact it meant "not in the last two days". You asked what the
built agent actually was, I went to look, and found 183 stored research results going back to
January — which obviously cannot sit alongside "it has never run".

**The corrected version, which is actually a stronger statement, not a weaker one:** the research
agent last ran on **18 January** — seven and a half months ago. It produced 92 results in a
five-day burst then and nothing since, and **none of those 92 was attached to a page or a page
section**, so none of them was the per-section research the page writer wants. Separately, we keep
a full log of every single call any agent makes to a model — five months of it, 87,822 calls, not
pruned — and the research agent appears in it **zero** times. Its workflow has two model-calling
steps, so if it had run, it would be there.

So the thing I was telling you is still true — the per-section research path has never delivered
anything into a page — but the evidence I gave you for it was not. I have logged the mistake and
the check that would have caught it: ask a table how far back it goes before reading an empty
result as history.

---

## 2026-09-04 — it is live, and proving it nearly went wrong

The new chassis went out overnight and the prompt fix is running in production. I checked that
properly rather than assuming it, and I am glad I did, because the first check said the opposite
of the truth.

I looked in the running system's logs for the new messages my change writes. Nothing. So I
looked for the *old* message that my change removes — also nothing. Two blanks like that read
as "the fix did not ship". But a check that comes back empty for everything is not a check, so
I asked a third question: is the system writing *any* prompts right now? Also nothing — while
the database showed 526 model calls over the same hours. The log command only reads one of the
two running copies. I had been looking through a keyhole.

The two checks that did work: I asked the running program directly what text it contains — my
new wording is in there, the old wording is gone — and I compared the database before and after.
Before the deploy, about a third of prompts contained the fake "no value" token. After it,
**none of 447**. Six slipped through in the thirteen minutes after the new version started,
which is the old copy finishing work it had already picked up.

## The uncomfortable part, which I want you to see clearly

**My fix has made the problem harder to measure.** The database stores the prompt *after* my
change has cleaned it, so the token no longer appears there. That means the "65% of prompts"
figure the other team measured can never be reproduced — I removed the evidence while removing
the symptom.

The holes themselves have not gone away. They were never going to: the fix stops the fake token
reaching the model, it does not fill in the missing data. Measuring it a different way, **267 of
448 prompts since the deploy still have the blank line, inside the block that tells the model
"use only these, do not invent"**. Each of those is now writing an error into a log that nothing
keeps and nobody reads.

The reviewers had already told me this would happen — that turning a quiet warning into a loud
one changes little if nobody is listening. The deploy turned their argument into a number. I
would still not undo the fix, because a fake value inside a "do not invent" instruction is worse
than a lost statistic. But the missing piece is now overdue rather than optional, and it is the
first thing on the handoff.
