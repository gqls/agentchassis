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
