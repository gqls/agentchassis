# Where we are — bugs_open/390

Plain prose, append-only, newest at the bottom.

## 2026-08-25 — opening the lane

The bug, said simply: when the site checker finds text that is too faint to read, it sends a
repair agent to change the colour. The agent writes a new line of styling into the site's main
stylesheet, saves it, and marks the job done. The job is marked done because something was
written — not because anyone looked to see whether the text got easier to read. Usually it did
not, so the checker finds the same faint text again next time, files it again, and another dead
line of styling gets added. Some sites now carry sixteen of these for the same five bits of text.

Why the repair does not work is worth understanding, because it is not the agent being careless.
Each page carries its own little block of styling, written inside the page itself, and browsers
give the last word to whichever instruction is more specific — and, when two are equally
specific, to whichever comes later. The page's own block is both more specific and later. The
agent is editing a file that structurally cannot win the argument. It is doing exactly what it
was told; what it was told is wrong. The prompt it works from actually says *"repeat the
offending selector exactly as it appears above … so your override wins"*, and on the pages I
measured that sentence produces the losing move every time.

I checked how often this happens rather than assuming. Of the repairs that were marked done and
whose target was a real one, **seventy-five** came straight back, and **ninety-seven** of those
returns recorded byte-for-byte the same two colours as before — the repair had changed nothing at
all. Then I took forty of them at random and went and looked at the actual pages: in
**thirty-three** the deciding instruction was in the page's own block and more specific than what
the agent wrote, and in **none** was it in the file the agent edits. So this is not an agent
writing weak rules. It is the platform sending the repair to the wrong place, and then recording
success.

Two things I got wrong or nearly wrong, worth writing down. The bug report's own worked example
turns out not to demonstrate the bug — that particular site has no stylesheet linked at all, so
the agent refuses and parks rather than pretending, which is an earlier fix doing its job. I found
a genuine example elsewhere and used that. And I briefly announced a "third failure mode" on
another site before checking the guard that governs it; the guard already catches that case, and
I have recorded the check I skipped.

One genuinely new thing did come out of it, and it is uncomfortable. Even a repair that *does*
win gets deleted. When a site's design is regenerated — a palette change, say — the whole
stylesheet is rewritten from scratch and every repair appended to it disappears. I watched that
happen on one site at 12:09 today, wiping five repairs made the evening before, and the work
items still say `complete`. That is behaving as designed; a migration from earlier this month
deliberately made the design agent own that file. Nobody has drawn the consequence for repairs
until now. It is a separate bug and I am filing it separately rather than folding it in here.

You chose the order: fix where the repair is aimed first, and leave the historical records alone.
So the plan is three steps. First, one small database change that stops the prompt instructing the
losing move — that ships without rebuilding anything and helps immediately. Second, the real fix:
teach the page checker to record *which* instruction is actually winning and where it lives, so
the platform can aim the repair at something it can reach — and have the checker prove its own
answer by removing the instruction it thinks is winning and confirming the colour changes. Third,
let the agent use that measurement, and park honestly the one case no stylesheet can ever fix.

I have written down, in advance, what I expect each step to do, including what result would prove
me wrong. That matters here more than usual, because the failure mode of this whole area is work
that reports success without having achieved anything — which is precisely the bug.

## 2026-08-25, later — what has actually shipped, and the two things that went wrong

Three things are now in.

**First, the quick one, and it is live.** The repair agent's instructions used to tell it, in
writing, to do the thing that loses. I replaced that text. It went live at about a quarter to five
and I checked the live record afterwards rather than assuming — the old sentence is gone, the new
one is there, and running it a second time now refuses loudly instead of quietly doing it twice.

**Second, the real fix, which is committed but asleep.** The page checker now works out *which*
styling instruction is actually deciding a colour, and — this is the part I care about — it
**proves** it rather than guessing. It removes the instruction it thinks is winning and looks
again. If the colour changes, that instruction was the winner; if nothing changes, it was not, and
the checker says "I could not tell" instead of inventing an answer. It puts everything back
afterwards and checks that it did. Nothing here takes effect until the next release of two
services, so the pipeline behaves exactly as it did today until then.

**Third, a separate bug filed.** The erasure problem I described earlier is now its own case with
its own number, so it does not get lost inside this one.

Now the two things that went wrong, because those are the useful part.

**The review panel found a real flaw and I had already shipped past it.** I put the quick fix
through the review council, and while it was thinking I applied the change. The panel came back
asking for a revision: my migration assumed there was exactly one copy of the agent's settings, and
never actually checked. It turns out four other agents in the estate *do* have two copies, so the
assumption is not idle — it just happened to be true for this one. I verified that myself rather
than taking their word, added the check, and proved it works by faking a second copy and watching
it refuse.

The second flaw they found is the one worth writing down. I had typed the same sentence into the
file twice — once for the safety check, once for the edit itself. If those two copies ever drifted
apart by a single character, the safety check would pass and the edit would silently do nothing,
while still reporting success. That is the *second* time in this one file that I wrote a value down
twice and created a way for it to disagree with itself; I caught the first one myself. The rule I
am taking from it: never restate a value, derive it. The file now declares each sentence once.

**And a mistake in how I sequenced things.** Because I applied the change before the panel finished,
their automatic checks looked at the world *after* my edit and reported that the text I was
proposing to replace was not there. That reads like something had gone wrong, and nothing had — it
was simply me having already done it. I have written that down where the next person will see it.

Two review rounds are still running. The remaining piece — teaching the repair agent to use the new
measurement — is deliberately waiting until the services roll, so it can be built against real data
rather than my expectations of it.
