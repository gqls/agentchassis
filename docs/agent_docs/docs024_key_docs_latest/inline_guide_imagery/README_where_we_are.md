# Where we are — pictures inside guide articles

Plain-prose running log, appended to, newest at the bottom. Started 2026-08-31 when the
lane was picked up; the design work behind it dates from 2026-08-14.

---

## 2026-08-31 — picking this up, and the one thing that is actually in the way

The ask, from a fortnight ago and repeated this week: our guide articles should have
pictures **inside** them — beside and between the paragraphs — not just the one banner
image at the top. The example you gave was the darts grip-styles guide, which talks about
ring grip, razor grip and shark grip and shows none of them.

Two things I expected to be the hard part turn out not to be.

**The pictures themselves are no longer a problem.** The current image model draws darts
correctly. Where our images are wrong they are old ones from July, made by the previous
model, and the lane looking after that site has spent today replacing them.

**The building block already exists.** A week ago another thread built exactly the piece
this plan asked for: a page section that is "some prose plus its own illustration", with
the picture held outside the writing so a rewrite of the words cannot take the picture with
it. Another change on the 26th taught the page planner to notice that this block can hold a
picture at all, which it previously could not distinguish from a plain block of text.

**What is actually in the way is smaller and more annoying than either.** The system can
attach *one* picture to a page, and it knows which page a picture belongs to — but it does
not know which *section* of that page a picture belongs to. The information is recorded
(each picture is filed against a page and a position within it), but the part of the code
that hands pictures to sections throws the position away and gives every section on the
page the same picture. So a guide with six small sections would show the same photograph six
times.

There is a second, related gap: when a page has several sections of the same kind, the
safety net that normally preserves what a page already has stops working for them, because
it identifies sections by type rather than by which one they are. Two sections of the same
type look identical to it, so it gives up and protects neither.

Both are small, both are in one file, and neither has ever mattered before — because until
now no page in the estate has had more than one illustrated section that the framework
itself produced. The one page that does have several (a bees homepage on another site) was
put together by hand, and by my reading its six pictures will disappear the next time
anything rewrites that page's words. I have not asserted that as fact: I have put it
through the diagnosis loop for an independent read rather than trusting my own reading of
the code, and will record what comes back.

So the position is: the parts are built, the pictures are good, and the missing piece is a
join between a picture and a section. That is the next thing to fix, and it is the whole
reason the guides still look like walls of text.

## 2026-08-31 (evening) — the join is built; the pictures themselves are the next job

The missing piece I described this afternoon is now written and committed. A picture filed
against "page X, section 3" is now actually handed to section 3, on both of the two routes
that rebuild a page — and it had to be both, because the second one would otherwise have
undone the first the next time anything re-rendered.

Two things are worth saying plainly about how it was done, because they are the parts that
could have gone quietly wrong.

**I proved the fault before fixing it.** It would have been easy to read the code, believe
it, and write the fix. Instead I wrote a test that asks two sections for two different
pictures, ran it against the old code, and watched it hand back the same picture twice.
That is the difference between "I think this is the bug" and "here is the bug". The same
test now passes, and a second one guards every page that already has a single picture, so
we can tell this change apart from one that quietly broke them.

**I sent it to the diagnosis loop first and it did not agree with me.** It came back "not
confirmed", and it was right: I had bundled two related faults into one description and
offered a page as evidence that only demonstrates one of them. That cost a few minutes and
saved a wrong claim from going into the record. The loop's objection is written up in the
technical notes, with my own error named.

The change is committed but will not do anything until the next fleet build — it is Go
code, so it sits inert until then. The review council has it and I will read the verdict.

**What is still missing is the pictures.** Nothing yet asks for one picture per section of a
guide: fleet-wide there are four such requests in total, across three sites, and none on the
darts site. So the next job is the supply — plan a figure for each small section of the
grip-styles guide (the one you pointed at, six subsections, no images at all today),
generate them, and watch one guide come out right end to end. That is the first thing where
you would actually see a difference on a page.
