# Where we are — editorial design uplift

The owner's running log. Plain prose, append-only, newest at the bottom.

---

## 2026-08-20 — the lane opens

You said the editorial feature looks good and can look a great deal better: more
imagery, placed against the parts of the text it belongs to; more graphic and
typographic treatment; more variety in the charts; and more timelines, collecting
the information for them if we have to. And you were right that it needs its own
plan rather than being bolted onto the news work — so this is its own workstream.

**I found both discussions you remembered, and they turned out to be two halves of
the same document** — the durable in-body imagery plan from 13-15 August. Worth
knowing what each actually says, because one of them changes what we should do
first.

**On interleaving copy and pictures:** the hard part is not the layout, it is that
the picture does not survive. An article's whole body lives in a single field that
the writing agent rewrites wholesale, so a figure placed inside the prose works,
looks permanent, and is silently destroyed the next time that article is rewritten.
Somebody measured the loss window on a real darts page. The fix in that plan is to
stop keeping the figure inside the prose at all: declare it in the site's plan,
where the rebuild path re-derives it, and splice it into the rendered page at
display time. That plan belongs to another lane, so I will coordinate with them
rather than fork it.

**On components made of other components:** this is the one that changes the order
of work. The capability is genuinely in the architecture — there is a column for a
component to have a parent, with an index and a foreign key ready. It has **never
been used once**: zero of 1,580 rows on the whole estate, and no code anywhere
reads it. A second mechanism for the same idea exists and is unreachable, because
the function that decides a component's render mode cannot produce the value.

So adopting it is building and proving something new, not switching something on.
**The good news is that we do not need it to make these pages look much better.**
The current page already alternates prose and charts section by section; that is
interleaving, at a coarser grain, and it works today. So the plan proceeds on what
exists, and treats composition as an unlock for later rather than a prerequisite.

**One thing is blocked, and I have not worked around it.** You asked for Fable to
write the component-hierarchy plan. I dispatched it with a full brief; it failed on
a Fable account limit — the fourth time that plan has died the same way. I did not
substitute another model, because you have twice asked specifically for Fable. The
brief is ready and everything it needs to read is catalogued. It needs Fable
capacity, not more thinking.

**What I did do:** rather than write a design plan out of my own taste, I asked the
platform's own design auditor to look at the site — it runs both an automated
check and a visual review, and writes down what it finds. The plan's first phase is
to act on what it says. I would rather the uplift start from findings than from my
opinion about typography.

The plan then goes: typography and editorial furniture first (cheapest, needs no
new mechanism); hero and in-body imagery second; chart variety third; timelines
fourth — and for timelines I want to **start collecting the dated events now**,
with citations, inside the features we are already writing, so that when the
timeline component is built there is honest data to draw. Building the picture
before the data is the one mistake this platform has already made and caught twice.
