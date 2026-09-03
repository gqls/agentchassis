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

---

## 2026-08-22 — the composition plan is written, by the model you asked for

The plan you wanted Fable to write now exists —
`features_open/035_FEATURE_component_hierarchy.md`. It took no workaround in the
end: this session itself runs Fable, so the fifth attempt was simply to write it
here. Nothing was substituted.

You added a steer as you unblocked it: you don't like the interleaved content and
imagery being produced in one model call, and you want it decomposed — more
control and consistency, and control over versions and design variations of the
same content. That became the plan's centre. In short: every piece of an
interleaved page — a block of prose, a figure, a chart, a pull-quote — becomes
its own addressable row, written by its own call against one shared brief,
individually lockable, individually regenerable, with its own history. Rewriting
one paragraph can no longer destroy a neighbouring figure, because they are no
longer the same thing.

Two useful discoveries came out of checking the ground first. The versioning
machinery is half-built already: the platform has been quietly snapshotting every
template edit — 363 snapshots so far — but nothing ever reads them back. The plan
wires that up, so a piece of a page can pin the design version it renders with,
and the same words can be re-dressed in a different design without being
regenerated. And the database already refuses, loudly, to delete a parent that
still has children — which is exactly the safety behaviour we want, found rather
than built.

The plan is deliberately staged: prove the rendering walk on one of our own
editorial pages first (they are locked and owned, so the blast radius is one
page), then decomposed generation, then versions and variants, then letting the
design agents propose re-arrangements — always through human review, never
applied automatically. The guides on the other sites come last, jointly with the
lane that owns the in-body imagery plan.

Nothing is built yet. The next concrete step is the local rendering proof, then
the first council-gated code phase.

---

## 2026-08-22, later — the rendering proof ran, and the design held

The composition plan's first test was deliberately cheap: prove, on this
machine, with no writes to the live system, that the new "components inside
components" rendering can work **without changing the template engine at all**
— because if it needed engine changes, the design would be wrong and we would
want to know before writing any real code. It passed. Eight checks, including
the important negative ones: a page with no composition renders byte-for-byte
as it does today (so nothing changes for any existing page); a broken
arrangement — a loop, a missing required piece, too-deep nesting — refuses to
render rather than quietly serving a page with pieces missing.

We also checked that the checks themselves can fail, by deliberately breaking
the new code two ways and watching exactly the right check catch each break.
A test that cannot fail proves nothing, and this platform has been caught by
that before.

One real discovery came out: the danger with a loop in the arrangement is not
that rendering spins for ever — it is that the looped pieces silently vanish
from the page while everything else renders fine. So the rule is now "every
piece must appear exactly once, or the whole render refuses", which catches
loops and stranded pieces alike. That went back into the plan.

Next is the first real code phase: the rendering walk in the live system,
proven on one of our own locked editorial pages, through the council gate.

---

## 2026-09-02 — why the boxing articles have no pictures, and the answer is smaller than it looks

You said the profiles have no pictures and there isn't enough imagery on the pages, and
asked us to talk to the imagery and component threads. Here is what we found, and it is
not what either of us expected.

**The pictures were made. They are sitting on the server right now. Nothing shows them.**

Every one of the six boxing articles has its own header image, generated on 1 September,
uploaded, and reachable — we fetched all six and every one came back fine. Yet each
article page displays exactly one image: the logo. The images are not missing, not
failed, not queued. They are finished work that nothing puts on the page.

**The reason is one building block.** Every article page on this platform is made of a
single piece called `article-body`. That piece has one input — the words — and its layout
has no place to put a picture at all. Not a header image, not a picture between
paragraphs, nothing. So an article page cannot show an image no matter how many we make
for it. The same is true of the block that builds the news page, which is why that page
is bare too. The front page and the articles index look better only because the block
they use *does* have a slot for a picture — that is the whole difference between the
surface that works and the two that don't.

**And we contributed to it.** Two weeks ago this thread asked another team to forbid the
writer from putting images into the article text. That request was right — images written
into the text get destroyed the next time the text is rewritten, which has already cost us
figures on other sites. But the instruction we asked for says images "belong to the
component system", and for this block there is no such thing. We closed one door without
checking anyone had opened the other. Written up honestly in our shared log of wrong
calls, because the failure is completely silent: the writer obeys, every job reports
success, and the page simply has no picture in it.

**What this means for the fix.** It is a much smaller job than "generate more imagery" —
the imagery is already there for the articles. The block needs somewhere to put a picture.
That is a contained change, but the block is used by 297 articles across 30 sites and it
goes live the instant it is written, so it goes through the review council with a test
proving every existing page renders exactly as it does today. We have not written it yet
and have not touched the boxing site — the delivery thread owns that pipeline.

**On the profiles specifically:** that site has no fighter profile pages at all — they
were asked for and never built, and that is already with the diagnosis loop. So there are
no unillustrated profiles; there are missing pages. When they are built they will need
real fighter imagery, and that site's own research is explicit that generic stock photos
are the thing to avoid.

**Your two rulings are recorded and we will work to them:** the cream off-white palette
stays, so we design for a light background; and a logo must not have its background baked
into it. That second one we are treating as a general rule for every site, not a one-off
repair of the boxing logo — the repair itself is with the delivery thread.

---

## 2026-09-02, later — the fix is written and passed review; it needs your go-ahead to switch on

The change that lets an article page show a picture has been written, tested, put through the
review council, and **approved on the second round**. It is not switched on: this kind of change
takes effect on every site the moment it is applied, so that is your call rather than mine.

**What it does.** The building block that every article page is made of gains one optional
picture slot. If a picture exists for that article it appears at the top of the piece; if not,
the page renders exactly as it does today. Every one of the 297 existing articles across 30 sites
is unaffected until a picture is actually available for it.

**Why it is small.** The images already exist and the machinery that finds them already exists —
it was built for precisely this case and has been sitting unused because the building block had
nowhere to put a picture. So this is a missing hole, not a missing pipeline.

**The review was worth having, and it caught me twice.** The first round came back "revise" with
two serious objections, and both were right: I had claimed two things were safe by pointing at
something similar rather than by checking. One of them mattered a great deal — there is a known
trap in this system where adding a non-writing field to a component can stop the writer producing
the article's *text* at all, silently, on every site. It does not apply here, and I can now say
exactly why rather than "the similar case was fine". The second was the worry that all six boxing
articles would end up showing the same picture instead of their own; I proved they each resolve
their own, page by page.

**Two honest caveats, both now written into the change itself.** If a picture is ever deleted, a
later rewrite of the article would quietly drop it with no error — that is how this system treats
all such fields, not something new here. And whoever switches it on must use a rebuild route that
re-reads the picture, not the quick redeploy — otherwise nothing appears and it looks like the
change did nothing.

**What I need from you:** whether to apply it. The command is ready and it is reversible — the
undo was rehearsed, not just written. Separately, making the six boxing articles actually show
their pictures is the delivery thread's job, not mine, and should happen after this is applied.

---

## 2026-09-02, later still — I applied the fix, then found it was the wrong fix, and undid it

You gave the go-ahead and I applied it. Then I took it back out about an hour later, because it
was wrong — not risky, wrong. **Nothing was ever shown to anyone**: the change only takes effect
on a page when that page is next rebuilt, and none were in the window. The system is exactly as it
was this morning.

**What I got wrong.** I proved that the article building block cannot display a picture, and that
six pictures were sitting unused on the boxing site. Both true. What I never checked was what
every *other* article page in the estate already does — and 292 of the 301 of them already have a
separate picture panel at the top of the page, drawing the very same image. So my change would
have shown **the same picture twice on 97% of them**: once in the panel, once again inside the
article. The six boxing pages that started all this turn out to be nine-in-a-hundred oddities that
have no picture panel at all.

**How it came to light.** Another thread mentioned, in passing, an unrelated case where a page was
showing one image twice. I recognised the shape as something my own just-applied change could
cause, went and counted, and undid it. Not my testing, and not the review council — eleven
reviewers looked at it and none of them asked the question either, because they were shown the
change and not the estate around it.

**The useful thing that came out of it.** We now know what the real problem is, and it is
somewhere else entirely. Other sites' article pages get their picture from a panel at the top,
fed by the site plan. The six boxing pages have no such panel and no plan entry for one — and the
system had helpfully generated per-article pictures precisely *because* those pages had no picture
of their own, then had nowhere to put them. **So the question is why those pages were built
without a picture panel when nearly every comparable page has one.** That is a question about how
pages are composed, not about the article block, and it is a better question than the one I spent
the day answering.

**Where that leaves things.** The building block is untouched and back to exactly its previous
state. The migration file is marked do-not-apply with the reason written into it. The estate's
shared log of wrong calls has the entry, because the lesson is worth more than the fix was: I
measured the broken cases thoroughly and never asked what the healthy ones do. A remedy has to fit
the population, not just the fault.

---

## 3 September 2026 — the piece of work that was finished but never plugged in

Today was the other half of this lane: not the pictures, the **structure** work (the "035" job — letting
one page section be built out of smaller parts instead of one lump).

Back on 31 August a piece of that was written, reviewed three times by the review council, and
committed. It was meant to do one job: when someone edits a small part of a section, the bigger
section that contains it has to be rebuilt, because the bigger one is what the page actually shows.
Without it, you edit the text, the edit saves, everything reports success — and the page keeps
showing the old words.

**It was never plugged in.** Nothing in the system ever called it. On 2 September a check of the
running program found the piece missing from the build, and correctly worked out it was there in the
source but unreachable — the compiler throws away code nothing calls.

**Today I found out why, and it is a small thing with a large moral.** The function asked for
something its only possible caller does not have. In plain terms: it demanded to be handed a
"transaction" — a way of grouping several database writes so they succeed or fail together — and the
part of the system that was supposed to call it doesn't use one at all. It writes to the database one
statement at a time.

What makes this worth writing down is *why nobody noticed*. The file said, in capital letters, that
this arrangement was **forced** — not a preference, a necessity — and explained at length which reads
had to happen inside that transaction. It read like a measured fact. It was an assumption, and it was
wrong, and **a comment is not checked by anything**. Three rounds of expert review looked at the plan
and did not catch it. One search of the file it referred to caught it in a few seconds.

**And once I tried to actually connect it, three more problems fell out** — none of which any amount
of reviewing the design would have found, because they live a level below the thing everyone was
arguing about:

- The write it makes had **no safety catches on it**. Everywhere else on that path, a write checks
  two things first: is this section retired, and has a human locked it? This one checked neither. That
  matters more here than elsewhere, because this write happens *automatically* as a knock-on effect of
  editing something else — so it reaches sections nobody chose. Without the checks, the one route in
  the whole system that can overwrite a section a human has locked would have been the route nobody
  aimed at it.
- **It could not tell "refused" from "done".** If the database declines the write, the statement still
  reports success. So a locked section would have been recorded as updated while quietly keeping its
  old contents — which is precisely the failure this whole feature exists to prevent, appearing inside
  the fix for it.
- **It did not sign its work.** We keep a history of who changed each section; a write that doesn't
  announce itself gets logged as coming from an anonymous network connection.

All three are fixed, and the fix is pinned by tests that I deliberately tried to break: I made four
separate sabotaging edits and checked each one turned a test red. All four did. Before today that file
had no tests at all.

**Does this change anything customers see? No, and that is on purpose.** No page anywhere is built out
of parts yet — I counted, and it is zero out of 3,229 sections, even though the table grew by about a
thousand rows in the last three days. The new code adds one cheap database lookup per edit and then
does nothing. It is switched off by construction until something opts in, which is how we are supposed
to ship anything that touches shared machinery.

**Where the structure work stands now.** Two of the three pieces are live: the system refuses to render
a parent section on its own (which would silently blank its children), and now it rebuilds parents when
a child changes. The missing third piece is the **reading** side — nothing yet actually renders a page
built out of parts. Until that lands, this feature is not usable, and I have corrected an entry in our
internal reference that was telling everyone the wrong reason for that.

**One thing worth flagging that is not mine.** A change another thread committed at lunchtime today
leaves a test failing for everybody in a neighbouring area of the code. I have left it alone — it is
their fix to make and the test message tells them exactly what to do — but it is worth knowing, because
a red test that isn't yours makes it harder to tell whether your own work is sound.
